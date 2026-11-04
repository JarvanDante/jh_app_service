package upload

import (
	"strings"
	"testing"
)

func TestLooksLikeSVG_RejectEmbeddedString(t *testing.T) {
	if looksLikeSVG([]byte(`hello world <svg xmlns="http://www.w3.org/2000/svg"></svg>`)) {
		t.Fatal("plain text that merely contains <svg must be rejected")
	}
	if !looksLikeSVG([]byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"></svg>`)) {
		t.Fatal("xml-declared svg root should pass")
	}
}

func TestSanitizeSVG(t *testing.T) {
	dirty := []byte(`<svg xmlns="http://www.w3.org/2000/svg" onload="alert(1)">
<script>alert(1)</script>
<a href="javascript:alert(1)">x</a>
<foreignObject><div>x</div></foreignObject>
<circle cx="1" cy="1" r="1"/>
</svg>`)
	clean, err := sanitizeSVG(dirty)
	if err != nil {
		t.Fatalf("sanitize: %v", err)
	}
	lower := strings.ToLower(string(clean))
	for _, bad := range []string{"<script", "onload=", "javascript:", "<foreignobject"} {
		if strings.Contains(lower, bad) {
			t.Errorf("still contains %q in %s", bad, lower)
		}
	}
	if !looksLikeSVG(clean) {
		t.Fatal("cleaned output should still look like svg")
	}

	if _, err := sanitizeSVG([]byte(`prefix <svg xmlns="http://www.w3.org/2000/svg"></svg>`)); err == nil {
		t.Fatal("embedded svg text must be rejected")
	}
}

func TestValidateFileType_GameCoverRejectsSVG(t *testing.T) {
	useEmptyConfig(t)
	s := &sUpload{}
	if err := s.validateFileType("image/svg+xml", "game_cover"); err == nil {
		t.Fatal("game_cover must reject svg")
	}
	if err := s.validateFileType("image/svg+xml", "mobile_logo"); err != nil {
		t.Fatalf("mobile_logo should allow svg type: %v", err)
	}
}
