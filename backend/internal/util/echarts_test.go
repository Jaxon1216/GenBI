package util

import (
	"encoding/json"
	"testing"
)

func TestCleanEchartsJSON(t *testing.T) {
	cases := []struct {
		name  string
		input string
		valid bool // 期望结果是否为非 "{}" 的合法 JSON
	}{
		{"空字符串兜底", "", false},
		{"去 option= 与末尾分号", `option = {"title":{"text":"标题"}};`, true},
		{"单引号转双引号", `{'a':'b'}`, true},
		{"带换行", "{\n\"x\":1\n}", true},
		{"非法内容兜底", `not a json`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := CleanEchartsJSON(tc.input)
			if !json.Valid([]byte(out)) {
				t.Fatalf("输出应为合法 JSON, got %q", out)
			}
			if tc.valid && out == "{}" {
				t.Fatalf("期望非空 JSON, 但得到兜底 {}: input=%q", tc.input)
			}
			if !tc.valid && out != "{}" {
				t.Fatalf("期望兜底 {}, got %q", out)
			}
		})
	}
}
