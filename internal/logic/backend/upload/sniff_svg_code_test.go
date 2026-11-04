package upload

import (
	"testing"
)

func TestDetectImageType_EmptyAndPolyglot(t *testing.T) {
	if _, _, ok := detectImageType(nil, "image/png"); ok {
		t.Fatal("empty data must fail")
	}
	if _, _, ok := detectImageType([]byte{}, "image/png"); ok {
		t.Fatal("empty slice must fail")
	}
	// 伪装成 png 的文本 / 类 exe 头
	if _, _, ok := detectImageType([]byte("MZ\x90not-an-image"), "image/png"); ok {
		t.Fatal("exe-like header without image magic must fail")
	}
}

func TestLooksLikeSVG_RootOnly(t *testing.T) {
	if !looksLikeSVG([]byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"/>`)) {
		t.Fatal("xml-declared svg root should pass")
	}
	if !looksLikeSVG([]byte(`<SVG xmlns="http://www.w3.org/2000/svg"/>`)) {
		t.Fatal("uppercase SVG root should pass")
	}
	if looksLikeSVG([]byte(`  <svgfoo xmlns="http://www.w3.org/2000/svg"/>`)) {
		t.Fatal("svgfoo must not count as svg root")
	}
	// DetectContentType 把以 HTML 注释开头的内容判成 text/html，当前实现会拒绝；
	// 这是已知限制，不是纯文本夹带 <svg 的绕过。
	if looksLikeSVG([]byte(`<!-- comment --><svg xmlns="http://www.w3.org/2000/svg"/>`)) {
		t.Fatal("html-comment-prefixed svg currently expected to be rejected by DetectContentType")
	}
}
