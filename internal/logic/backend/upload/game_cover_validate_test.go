package upload

import (
	"testing"

	v1 "jh_app_service/api/backend/upload/v1"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
)

// 使用空配置，确保测试的是 game_cover 的代码内默认值（类型白名单、5MB）。
func useEmptyConfig(t *testing.T) {
	t.Helper()
	a, err := gcfg.NewAdapterContent("{}")
	if err != nil {
		t.Fatalf("config adapter: %v", err)
	}
	g.Cfg().SetAdapter(a)
}

func TestGameCover_FileTypes(t *testing.T) {
	useEmptyConfig(t)
	s := &sUpload{}
	for _, ct := range []string{"image/jpeg", "image/jpg", "image/png", "image/webp", "image/gif"} {
		if err := s.validateFileType(ct, "game_cover"); err != nil {
			t.Errorf("%s should be allowed: %v", ct, err)
		}
	}
	for _, ct := range []string{"image/svg+xml", "text/html", "application/octet-stream", "application/pdf", ""} {
		if err := s.validateFileType(ct, "game_cover"); err == nil {
			t.Errorf("%q should be rejected", ct)
		}
	}
}

func TestGameCover_MaxSize5MB(t *testing.T) {
	useEmptyConfig(t)
	s := &sUpload{}
	if err := s.validateFileSize(5*1024*1024, "game_cover"); err != nil {
		t.Errorf("exactly 5MB should be allowed: %v", err)
	}
	if err := s.validateFileSize(5*1024*1024+1, "game_cover"); err == nil {
		t.Error("5MB+1 should be rejected")
	}
	// 默认 code 仍为 2MB
	if err := s.validateFileSize(2*1024*1024+1, "default"); err == nil {
		t.Error("default code should still be limited to 2MB")
	}
}

func TestGameCover_RequestValidation(t *testing.T) {
	s := &sUpload{}
	ok := &v1.UploadImageReq{FileData: []byte{1, 2, 3}, FileName: "a.png", ContentType: "image/png", FileSize: 3, UploadCode: "game_cover"}
	if err := s.validateUploadRequest(ok); err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}
	bad := []*v1.UploadImageReq{
		{FileName: "a.png", ContentType: "image/png", FileSize: 3},
		{FileData: []byte{1}, ContentType: "image/png", FileSize: 1},
		{FileData: []byte{1}, FileName: "a.png", FileSize: 1},
		{FileData: []byte{1, 2}, FileName: "a.png", ContentType: "image/png", FileSize: 1}, // 声明大小与实际不一致
	}
	for i, r := range bad {
		if err := s.validateUploadRequest(r); err == nil {
			t.Errorf("bad request #%d should be rejected", i)
		}
	}
}
