package upload

import (
	"bytes"
	"image"
	"image/png"
	"testing"
)

func pngBytes(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestDetectImageType(t *testing.T) {
	cases := []struct {
		name, declared string
		data           []byte
		mime, ext      string
		ok             bool
	}{
		{"png", "image/jpeg", pngBytes(t), "image/png", ".png", true},
		{"jpeg", "image/png", []byte("\xFF\xD8\xFF\xE0\x00\x10JFIF\x00"), "image/jpeg", ".jpg", true},
		{"gif", "", []byte("GIF89a\x01\x00\x01\x00"), "image/gif", ".gif", true},
		{"webp", "", []byte("RIFF\x24\x00\x00\x00WEBPVP8 "), "image/webp", ".webp", true},
		{"text as png", "image/png", []byte("hello, not an image"), "", "", false},
		{"html", "image/png", []byte("<html><script>alert(1)</script></html>"), "", "", false},
		{"svg declared", "image/svg+xml", []byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`), "image/svg+xml", ".svg", true},
		{"svg undeclared", "image/png", []byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`), "", "", false},
		{"text contains svg", "image/svg+xml", []byte(`not an image but has <svg inside`), "", "", false},
	}
	for _, c := range cases {
		m, e, ok := detectImageType(c.data, c.declared)
		if ok != c.ok || (c.ok && (m != c.mime || e != c.ext)) {
			t.Errorf("%s: got (%s,%s,%v) want (%s,%s,%v)", c.name, m, e, ok, c.mime, c.ext, c.ok)
		}
	}
}

func TestBuildObjectName(t *testing.T) {
	cases := map[string]string{
		"cover.txt":       "cover_100.png",
		"cover.jpg":       "cover_100.png",
		"a/b/../evil.php": "evil_100.png",
		".hidden":         ".hidden_100.png",
		"":                "image_100.png",
		"noext":           "noext_100.png",
	}
	for in, want := range cases {
		if got := buildObjectName(in, ".png", 100); got != want {
			t.Errorf("%q: got %q want %q", in, got, want)
		}
	}
}

func TestValidateFileType_JpegMatchesJpgConfig(t *testing.T) {
	useEmptyConfig(t)
	s := &sUpload{}
	if err := s.validateFileType("image/jpeg", "mobile_logo"); err != nil {
		t.Errorf("image/jpeg should be allowed: %v", err)
	}
}
