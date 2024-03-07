package utils

import (
	"unicode/utf8"
)

// countCharacters 返回字符串str的字符数
func countCharacters(str string) int {
	if str == "" {
		return 0
	}
	return utf8.RuneCountInString(str)
}

func SimplifyPostBody(body string) string {
	//	判断 内容 字数
	count := countCharacters(body)
	if count <= 100 {
		return body
	}
	//	太长则截取前 75 位
	return string([]rune(body)[:75]) + "...."
}
