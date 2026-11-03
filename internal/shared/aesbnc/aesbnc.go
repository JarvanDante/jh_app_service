// Package aesbnc 对齐 my 项目封面加密: AES-128-ECB + PKCS7，密钥与 AesUtil::decryptRaw 相同。
// 密文对象统一用 .bnc 后缀。
package aesbnc

import (
	"crypto/aes"
	"path/filepath"
	"strings"
)

const DefaultKey = "dafb787c76b5ce8a"
const Ext = ".bnc"

func Encrypt(plain []byte) ([]byte, error) {
	if plain == nil {
		plain = []byte{}
	}
	block, err := aes.NewCipher([]byte(DefaultKey))
	if err != nil {
		return nil, err
	}
	padded := pkcs7Pad(plain, aes.BlockSize)
	out := make([]byte, len(padded))
	for i := 0; i < len(padded); i += aes.BlockSize {
		block.Encrypt(out[i:i+aes.BlockSize], padded[i:i+aes.BlockSize])
	}
	return out, nil
}

func pkcs7Pad(b []byte, n int) []byte {
	pad := n - (len(b) % n)
	out := make([]byte, len(b)+pad)
	copy(out, b)
	for i := len(b); i < len(out); i++ {
		out[i] = byte(pad)
	}
	return out
}

func ToBncKey(key string) string {
	if key == "" {
		return "image" + Ext
	}
	low := strings.ToLower(key)
	if strings.HasSuffix(low, Ext) || strings.HasSuffix(low, ".ceb") {
		return key
	}
	ext := filepath.Ext(key)
	base := strings.TrimSuffix(key, ext)
	if strings.TrimSpace(base) == "" {
		return "image" + Ext
	}
	return base + Ext
}
