import { z } from 'zod';

const EChartsSeriesSchema = z
  .object({
    type: z.enum([
      'line',
      'bar',
      'pie',
      'scatter',
      'radar',
      'heatmap',
      'tree',
      'treemap',
      'sunburst',
      'boxplot',
      'candlestick',
      'funnel',
      'gauge',
      'sankey',
      'graph',
    ]),
    data: z.array(z.any()).optional(),
    name: z.string().optional(),
  })
  .passthrough();

const EChartsOptionSchema = z
  .object({
    title: z.any().optional(),
    legend: z.any().optional(),
    grid: z.any().optional(),
    xAxis: z.any().optional(),
    yAxis: z.any().optional(),
    series: z.array(EChartsSeriesSchema).min(1, 'series 至少需要一项'),
    tooltip: z.any().optional(),
    toolbox: z.any().optional(),
    dataZoom: z.any().optional(),
    color: z.array(z.string()).optional(),
    radar: z.any().optional(),
  })
  .passthrough();

export type ValidEChartsOption = z.infer<typeof EChartsOptionSchema>;

export type ParseResult =
  | { success: true; option: ValidEChartsOption }
  | { success: false; error: string; fallbackOption: Record<string, any> | null };

const FALLBACK_OPTION: Record<string, any> = {
  title: { text: '图表解析失败', subtext: '请重试或联系管理员' },
  series: [{ type: 'bar', data: [] }],
};

// 辅助函数：提取 JSON 对象
function extractJsonObject(raw: string): string | null {
  const start = raw.indexOf('{');
  const end = raw.lastIndexOf('}');
  if (start === -1 || end === -1 || end <= start) return null;
  return raw.slice(start, end + 1);
}

const DANGEROUS_KEYS = ['formatter', 'rich', 'graphic', 'bindEvent'];

// 辅助函数：清理危险字段
function sanitize(obj: Record<string, any>): Record<string, any> {
  const json = JSON.stringify(obj, (_key, value) => {
    if (typeof value === 'string' && (value.includes('function') || value.includes('=>'))) {
      return undefined;
    }
    if (typeof _key === 'string' && DANGEROUS_KEYS.includes(_key) && typeof value === 'string') {
      return undefined;
    }
    return value;
  });
  return JSON.parse(json);
}

/**
 * 安全解析 AI 返回的 ECharts option 字符串。
 * 1. JSON.parse（容错：去除首尾非 JSON 字符）
 * 2. Zod 结构校验
 * 3. 敏感字段过滤（移除可能的 JS 注入 / 函数字符串）
 * 4. 标准化处理（强制移除 title 避免与卡片标题冲突等）
 */
export function parseChartOption(raw: string | undefined | null, stripTitle = false): ParseResult {
  if (!raw || raw.trim() === '') {
    return { success: false, error: '图表数据为空', fallbackOption: FALLBACK_OPTION };
  }

  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch {
    const trimmed = extractJsonObject(raw);
    if (!trimmed) {
      return {
        success: false,
        error: 'JSON 解析失败：AI 返回格式异常',
        fallbackOption: FALLBACK_OPTION,
      };
    }
    try {
      parsed = JSON.parse(trimmed);
    } catch {
      return {
        success: false,
        error: 'JSON 解析失败：无法提取有效 JSON',
        fallbackOption: FALLBACK_OPTION,
      };
    }
  }

  const result = EChartsOptionSchema.safeParse(parsed);
  if (!result.success) {
    const issues = result.error.issues.map((i) => `${i.path.join('.')}: ${i.message}`).join('; ');
    return {
      success: false,
      error: `Schema 校验失败：${issues}`,
      fallbackOption:
        typeof parsed === 'object' && parsed !== null
          ? sanitize(parsed as Record<string, any>)
          : FALLBACK_OPTION,
    };
  }

  let option = result.data as Record<string, any>;
  option = sanitize(option);

  if (stripTitle) {
    option.title = undefined;
  }

  return { success: true, option: option as ValidEChartsOption };
}
