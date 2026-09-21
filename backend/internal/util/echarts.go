package util

import (
	"encoding/json"
	"regexp"
	"strings"
)

var (
	reOptionAssign  = regexp.MustCompile(`option\s*=\s*`)
	reTrailingSemi  = regexp.MustCompile(`;$`)
	reTrailingComma = regexp.MustCompile(`,\s*}$`)
)

// CleanEchartsJSON 清洗 AI 返回的 ECharts 配置为合法 JSON 字符串。
// 移植自 Java cleanEchartsJsToJson：
//  1. 去掉 "option = " 赋值
//  2. 去掉末尾分号 / 逗号
//  3. 单引号转双引号
//  4. 去掉换行
//  5. 校验合法性，非法则兜底 "{}"
func CleanEchartsJSON(echartsJS string) string {
	if strings.TrimSpace(echartsJS) == "" {
		return "{}"
	}
	s := reOptionAssign.ReplaceAllString(echartsJS, "")
	s = reTrailingSemi.ReplaceAllString(s, "")
	s = reTrailingComma.ReplaceAllString(s, "}")
	s = strings.ReplaceAll(s, "'", "\"")
	s = strings.NewReplacer("\n", "", "\r", "").Replace(s)
	s = strings.TrimSpace(s)
	if json.Valid([]byte(s)) {
		return s
	}
	return "{}"
}
