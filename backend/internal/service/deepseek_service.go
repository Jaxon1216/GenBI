package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// systemPrompt 沿用 Java DeepSeekServiceImpl 的系统提示词。
const systemPrompt = `你是专业智能BI数据分析师，严格按以下规则处理CSV数据：
1. 输入是带\n换行的CSV字符串，第一行是表头（必须完整读取，如日期、用户数），后续是数据行，逗号分隔列
2. 所有分析100%基于输入数据，严禁编造、外推，只基于表头字段和数据值
请根据这两部分内容，按照以下指定格式生成内容（此外不要输出任何多余的开头、结尾、注释）：
1. 输出结构：{"genChart": "标准ECharts JSON配置字符串", "genResult": ["分析结论", "分析建议"]}
2. genChart 必须是：
- 纯JSON格式的ECharts配置（无option=、无分号、无单引号、无注释）
- 所有字符串用双引号
- 直接是对象，比如 {"title":{"text":"标题"},"xAxis":...}
3. genResult 是数组格式，每条结论一句话（20-50字）
4. 不要返回任何JS代码、不要加markdown、不要加解释文字
`

// AiResult AI 分析结果（已从 content 解析出的原始字段）。
type AiResult struct {
	GenChart  string // ECharts 配置字符串（未清洗）
	GenResult string // 分析结论（数组会被序列化为字符串）
}

// DeepSeekService 调用 DeepSeek Chat Completions 分析 CSV。
type DeepSeekService struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewDeepSeekService 构造。
func NewDeepSeekService(apiKey, baseURL string) *DeepSeekService {
	return &DeepSeekService{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

// AnalyzeCsv 调用 DeepSeek 分析（prompt 已由调用方拼好），返回解析后的结果。
//
// ctx 用于超时/取消控制：通过 http.NewRequestWithContext 接入请求，使 AI 调用
// 可被 ctx 超时/取消「真正打断」（底层连接会被关闭），而非仅靠 client 的 60s 兜底。
// 当 ctx 超时/取消时，返回的 error 用 context.DeadlineExceeded / context.Canceled
// 包装，保证上层 errors.Is 可识别为「生成超时 / 已取消」。
func (s *DeepSeekService) AnalyzeCsv(ctx context.Context, prompt string) (*AiResult, error) {
	if ctx == nil {
		// 兜底：Phase 0 调用方均传入非 nil ctx，此处仅防御 http.NewRequestWithContext 对 nil 的 panic。
		ctx = context.Background()
	}
	if s.apiKey == "" {
		return nil, errors.New("DeepSeek API key 未配置")
	}

	reqBody := map[string]any{
		"model":       "deepseek-chat",
		"temperature": 0.1,
		"max_tokens":  1024,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": "分析以下CSV数据：\n" + prompt},
		},
	}
	payload, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/v1/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		// ctx 超时/取消时，client.Do 返回包装了 context.DeadlineExceeded / context.Canceled 的 error；
		// 归一化为「可识别」错误，确保上层 errors.Is(err, context.DeadlineExceeded) 为真。
		if ctxErr := wrapCtxError(ctx, err); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, err
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		// 读取响应体期间 ctx 被打断，同样归一化。
		if ctxErr := wrapCtxError(ctx, err); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, err
	}

	// 解析 choices[0].message.content
	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBytes, &chatResp); err != nil {
		return nil, err
	}
	if len(chatResp.Choices) == 0 {
		return nil, errors.New("DeepSeek 返回内容为空")
	}
	content := chatResp.Choices[0].Message.Content

	// content 本身是一段 JSON：{"genChart": "...", "genResult": [...]}
	// genResult 可能是数组或字符串，统一转成字符串。
	var parsed struct {
		GenChart  string          `json:"genChart"`
		GenResult json.RawMessage `json:"genResult"`
	}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, err
	}
	return &AiResult{
		GenChart:  parsed.GenChart,
		GenResult: rawToString(parsed.GenResult),
	}, nil
}

// rawToString 把 genResult 归一化为字符串：若本身是 JSON 字符串则取其值，否则原样序列化。
func rawToString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return str
	}
	return string(raw)
}

// wrapCtxError 在请求出错时判断是否由 ctx 超时/取消引起，若是则返回「可识别」的中文错误：
//   - 超时：fmt.Errorf("生成超时: %w", context.DeadlineExceeded)，errors.Is(err, context.DeadlineExceeded) 为真；
//   - 取消：fmt.Errorf("生成已取消: %w", context.Canceled)，errors.Is(err, context.Canceled) 为真。
//
// 既优先依据 ctx.Err()（最可靠），也兜底识别 client.Do 返回的包装错误，返回 nil 表示非 ctx 原因。
func wrapCtxError(ctx context.Context, err error) error {
	switch ctx.Err() {
	case context.DeadlineExceeded:
		return fmt.Errorf("生成超时: %w", context.DeadlineExceeded)
	case context.Canceled:
		return fmt.Errorf("生成已取消: %w", context.Canceled)
	}
	// ctx.Err() 为空的边界情况：直接检查底层 error 是否包装了 ctx 错误。
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("生成超时: %w", context.DeadlineExceeded)
	}
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("生成已取消: %w", context.Canceled)
	}
	return nil
}
