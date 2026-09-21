package util

import (
	"errors"
	"io"
	"mime/multipart"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ErrUnsupportedFile 文件格式不支持。
var ErrUnsupportedFile = errors.New("仅支持 .xlsx 和 .csv 格式文件")

// ExcelToCSV 将上传文件转成带表头的 CSV 字符串。
// 支持 .xlsx（excelize 读首个 sheet）与 .csv（原样返回）。
// 不支持老式 .xls（BIFF 格式，excelize 不支持）。
func ExcelToCSV(fileHeader *multipart.FileHeader) (string, error) {
	name := strings.ToLower(fileHeader.Filename)
	switch {
	case strings.HasSuffix(name, ".xlsx"):
		return xlsxToCSV(fileHeader)
	case strings.HasSuffix(name, ".csv"):
		return csvPassthrough(fileHeader)
	default:
		return "", ErrUnsupportedFile
	}
}

// xlsxToCSV 读取 .xlsx 首个 sheet 的所有行并拼成 CSV。
func xlsxToCSV(fileHeader *multipart.FileHeader) (string, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	f, err := excelize.OpenReader(src)
	if err != nil {
		return "", err
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return "", errors.New("Excel 无可用 sheet")
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, row := range rows {
		if isEmptyRow(row) {
			continue
		}
		for i, cell := range row {
			if i > 0 {
				sb.WriteByte(',')
			}
			sb.WriteString(escapeCSV(cell))
		}
		sb.WriteByte('\n')
	}
	return sb.String(), nil
}

// csvPassthrough 直接读取 .csv 文本内容。
func csvPassthrough(fileHeader *multipart.FileHeader) (string, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	data, err := io.ReadAll(src)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func isEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

// escapeCSV 对包含逗号、引号、换行的字段做 CSV 转义。
func escapeCSV(value string) string {
	if value == "" {
		return ""
	}
	if strings.ContainsAny(value, ",\"\n\r") {
		return "\"" + strings.ReplaceAll(value, "\"", "\"\"") + "\""
	}
	return value
}
