package util

import "testing"

func TestEscapeCSV(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"abc", "abc"},
		{"a,b", `"a,b"`},     // 含逗号需加引号
		{`a"b`, `"a""b"`},    // 含引号需转义并加引号
		{"a\nb", "\"a\nb\""}, // 含换行需加引号
	}
	for _, tc := range cases {
		if got := escapeCSV(tc.in); got != tc.want {
			t.Errorf("escapeCSV(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestIsEmptyRow(t *testing.T) {
	if !isEmptyRow([]string{"", "  ", ""}) {
		t.Error("全空白行应判为空行")
	}
	if isEmptyRow([]string{"", "x"}) {
		t.Error("含非空单元格不应判为空行")
	}
}
