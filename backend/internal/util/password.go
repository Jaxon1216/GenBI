// Package util 提供通用工具：密码加密、Excel 转 CSV、ECharts 配置清洗。
package util

import (
	"crypto/md5"
	"encoding/hex"
)

// salt 混淆密码用的固定盐，与 Java UserServiceImpl.SALT 保持一致。
const salt = "tanter"

// EncryptPassword 计算 md5Hex(salt + 明文密码)，与 Java DigestUtils.md5DigestAsHex 一致。
func EncryptPassword(raw string) string {
	sum := md5.Sum([]byte(salt + raw))
	return hex.EncodeToString(sum[:])
}
