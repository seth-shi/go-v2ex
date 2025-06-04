package pkg

func CutString(str string, length int) string {
	if length <= 0 {
		return ""
	}
	// 将字符串转为rune切片（按Unicode字符处理）
	runes := []rune(str)
	// 若原字符串长度小于等于目标长度，直接返回
	if len(runes) <= length {
		return str
	}
	// 截断并添加...
	return string(runes[:length]) + "..."
}
