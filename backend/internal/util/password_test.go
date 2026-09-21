package util

import "testing"

func TestEncryptPassword(t *testing.T) {
	// 与 Java DigestUtils.md5DigestAsHex(("tanter"+"password123").getBytes()) 一致。
	got := EncryptPassword("password123")
	want := "85c3c6fbf8b5a059ef351f2818fb0bb1"
	if got != want {
		t.Fatalf("EncryptPassword: got %q, want %q", got, want)
	}
}

func TestEncryptPasswordDeterministic(t *testing.T) {
	// 相同输入应得到相同哈希（固定盐特性）。
	if EncryptPassword("abc12345") != EncryptPassword("abc12345") {
		t.Fatal("相同密码应产生相同哈希")
	}
	if EncryptPassword("abc12345") == EncryptPassword("abc12346") {
		t.Fatal("不同密码应产生不同哈希")
	}
}
