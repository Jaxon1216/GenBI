package service

import (
	"strings"
	"testing"
)

func TestBuildPrompt(t *testing.T) {
	got := BuildPrompt("分析增长", "折线图", "日期,用户数\n1,10")
	for _, sub := range []string{"分析目标：分析增长", "图表类型：折线图", "日期,用户数"} {
		if !strings.Contains(got, sub) {
			t.Errorf("prompt 缺少 %q，实际: %q", sub, got)
		}
	}
}

func TestBuildOrder(t *testing.T) {
	cases := []struct {
		field, order, want string
	}{
		{"", "asc", ""}, // 空字段不排序
		{"createTime", "descend", "createTime DESC"},
		{"createTime", "ascend", "createTime ASC"},
		{"name", "asc", "name ASC"},
		{"id; DROP TABLE", "asc", ""}, // 含非法字符，拒绝（防注入）
		{"user`Name", "asc", ""},      // 含反引号，拒绝
	}
	for _, tc := range cases {
		if got := buildOrder(tc.field, tc.order); got != tc.want {
			t.Errorf("buildOrder(%q,%q) = %q, want %q", tc.field, tc.order, got, tc.want)
		}
	}
}

func TestNormalizePage(t *testing.T) {
	if c, s := normalizePage(0, 0); c != 1 || s != 10 {
		t.Errorf("normalizePage(0,0) = (%d,%d), want (1,10)", c, s)
	}
	if c, s := normalizePage(3, 20); c != 3 || s != 20 {
		t.Errorf("normalizePage(3,20) = (%d,%d), want (3,20)", c, s)
	}
}
