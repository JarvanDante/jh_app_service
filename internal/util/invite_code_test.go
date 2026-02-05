package util

import (
	"testing"
)

func TestEncodeInviteCode(t *testing.T) {
	tests := []struct {
		id       uint
		expected string
	}{
		{0, "000000"},
		{1, "000001"},
		{10, "00000A"},
		{35, "00000Z"},
		{36, "00000a"},
		{61, "00000z"},
		{62, "000010"},
		{100, "00001c"},
		{1000, "0000G8"},
		{12345, "0003D7"},
		{999999, "004C91"},
		{56800235583, "zzzzzz"}, // 最大值（62^6 - 1）
	}

	for _, tt := range tests {
		result := EncodeInviteCode(tt.id)
		if result != tt.expected {
			t.Errorf("EncodeInviteCode(%d) = %s; want %s", tt.id, result, tt.expected)
		}
	}
}

func TestDecodeInviteCode(t *testing.T) {
	tests := []struct {
		code     string
		expected uint
	}{
		{"000000", 0},
		{"000001", 1},
		{"00000A", 10},
		{"00000Z", 35},
		{"00000a", 36},
		{"00000z", 61},
		{"000010", 62},
		{"00001c", 100},
		{"0000G8", 1000},
		{"0003D7", 12345},
		{"004C91", 999999},
		{"zzzzzz", 56800235583},
	}

	for _, tt := range tests {
		result := DecodeInviteCode(tt.code)
		if result != tt.expected {
			t.Errorf("DecodeInviteCode(%s) = %d; want %d", tt.code, result, tt.expected)
		}
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	// 测试编码解码往返
	testIds := []uint{0, 1, 10, 100, 1000, 10000, 100000, 999999, 12345678}

	for _, id := range testIds {
		code := EncodeInviteCode(id)
		decodedId := DecodeInviteCode(code)
		if decodedId != id {
			t.Errorf("Round trip failed for ID %d: encoded to %s, decoded to %d", id, code, decodedId)
		}
	}
}

func TestValidateInviteCode(t *testing.T) {
	tests := []struct {
		code     string
		expected bool
	}{
		{"000001", true},
		{"00001C", true},
		{"0003D7", true},
		{"zzzzzz", true},
		{"ABCDEF", true},
		{"123456", true},
		{"", false},        // 空字符串
		{"00001", false},   // 长度不足
		{"0000001", false}, // 长度过长
		{"00001@", false},  // 包含非法字符
		{"00001#", false},  // 包含非法字符
		{"00001 ", false},  // 包含空格
		{"00001-", false},  // 包含连字符
	}

	for _, tt := range tests {
		result := ValidateInviteCode(tt.code)
		if result != tt.expected {
			t.Errorf("ValidateInviteCode(%s) = %v; want %v", tt.code, result, tt.expected)
		}
	}
}

func TestDecodeInvalidCode(t *testing.T) {
	// 测试无效邀请码
	invalidCodes := []string{
		"",
		"@@@@@",
		"------",
		"      ",
	}

	for _, code := range invalidCodes {
		result := DecodeInviteCode(code)
		if result != 0 {
			t.Errorf("DecodeInviteCode(%s) should return 0 for invalid code, got %d", code, result)
		}
	}
}

// 基准测试
func BenchmarkEncodeInviteCode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EncodeInviteCode(12345)
	}
}

func BenchmarkDecodeInviteCode(b *testing.B) {
	code := "0003D7"
	for i := 0; i < b.N; i++ {
		DecodeInviteCode(code)
	}
}
