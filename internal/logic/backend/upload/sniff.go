package upload

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"
)

// sniffedImageExt 内容嗅探识别出的图片类型 -> 存储扩展名
var sniffedImageExt = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// detectImageType 根据文件前 512 字节嗅探真实类型，不信任客户端 Content-Type 与文件名。
// 仅识别 jpeg/png/gif/webp；declaredType 为 image/svg+xml 且内容确为 SVG 时
// 保持原有 svg 支持（是否允许仍由 validateFileType 按 upload code 决定）。
func detectImageType(data []byte, declaredType string) (mimeType, ext string, ok bool) {
	head := data
	if len(head) > 512 {
		head = head[:512]
	}
	detected := http.DetectContentType(head)
	if e, found := sniffedImageExt[detected]; found {
		return detected, e, true
	}
	if strings.EqualFold(strings.TrimSpace(declaredType), "image/svg+xml") && looksLikeSVG(data) {
		return "image/svg+xml", ".svg", true
	}
	return detected, "", false
}

func looksLikeSVG(data []byte) bool {
	head := data
	if len(head) > 1024 {
		head = head[:1024]
	}
	detected := http.DetectContentType(head)
	if !strings.HasPrefix(detected, "text/xml") && !strings.HasPrefix(detected, "text/plain") {
		return false
	}
	return bytes.Contains(bytes.ToLower(head), []byte("<svg"))
}

// buildObjectName 用原文件名主体 + 时间戳 + 检测出的扩展名生成对象名
func buildObjectName(fileName, ext string, unix int64) string {
	base := fileName
	if i := strings.LastIndexAny(base, "/\\"); i >= 0 {
		base = base[i+1:]
	}
	if i := strings.LastIndex(base, "."); i > 0 {
		base = base[:i]
	}
	base = strings.TrimSpace(base)
	if base == "" || base == "." {
		base = "image"
	}
	return base + "_" + strconv.FormatInt(unix, 10) + ext
}
