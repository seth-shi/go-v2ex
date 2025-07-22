package pkg

import (
	"encoding/base64"
	"regexp"
	"strings"
	"unicode/utf8"
)

// 配置参数
const (
	maxBase64Length = 1024 // 最大Base64长度
)

var (
	// 最小的高度
	stdBase64Regex = regexp.MustCompile(`\b[A-Za-z0-9+/]{8,}(?:==|=)?\b`)
)

func DetectBase64(content string) map[string]string {
	var results = make(map[string]string)

	// 直接获取匹配的Base64字符串
	for _, b64Str := range stdBase64Regex.FindAllString(content, -1) {
		if len(b64Str) > maxBase64Length {
			continue // 跳过过长的匹配
		}

		if decoded, ok := decodeBase64(b64Str); ok {
			results[b64Str] = decoded
		}
	}

	return results
}

func decodeBase64(b64Str string) (string, bool) {
	if pad := len(b64Str) % 4; pad != 0 {
		b64Str += strings.Repeat("=", 4-pad)
	}

	decoded, err := base64.StdEncoding.DecodeString(b64Str)
	if err != nil {
		return "", false
	}

	decodeContent := string(decoded)
	if !utf8.ValidString(decodeContent) {
		return "", false
	}

	return decodeContent, true
}
