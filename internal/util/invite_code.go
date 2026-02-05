package util

import (
	"strings"
)

const (
	// Base62字符集：0-9, A-Z, a-z
	base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	// 邀请码固定长度（6位可支持约568亿用户）
	inviteCodeLength = 6
)

// EncodeInviteCode 将用户ID编码为邀请码
// 使用Base62编码，固定长度6位
// 示例：
//
//	ID: 1      → 000001
//	ID: 100    → 00001C
//	ID: 12345  → 0003D7
//	ID: 999999 → 004C91
func EncodeInviteCode(id uint) string {
	if id == 0 {
		return strings.Repeat("0", inviteCodeLength)
	}

	// Base62编码
	result := ""
	tempId := id
	for tempId > 0 {
		remainder := tempId % 62
		result = string(base62Chars[remainder]) + result
		tempId /= 62
	}

	// 填充到固定长度
	for len(result) < inviteCodeLength {
		result = "0" + result
	}

	return result
}

// DecodeInviteCode 将邀请码解码为用户ID
// 返回用户ID，如果邀请码无效则返回0
func DecodeInviteCode(code string) uint {
	if code == "" {
		return 0
	}

	// 移除前导0（但保留至少一个字符）
	code = strings.TrimLeft(code, "0")
	if code == "" {
		return 0
	}

	var id uint = 0
	for _, char := range code {
		// 查找字符在base62字符集中的位置
		index := strings.IndexByte(base62Chars, byte(char))
		if index == -1 {
			// 无效字符，返回0
			return 0
		}
		id = id*62 + uint(index)
	}

	return id
}

// ValidateInviteCode 验证邀请码格式是否正确
// 返回true表示格式正确，false表示格式错误
func ValidateInviteCode(code string) bool {
	// 检查长度
	if len(code) != inviteCodeLength {
		return false
	}

	// 检查每个字符是否在base62字符集中
	for _, char := range code {
		if !strings.ContainsRune(base62Chars, char) {
			return false
		}
	}

	return true
}
