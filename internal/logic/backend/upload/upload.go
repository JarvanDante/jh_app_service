package upload

import (
	"bytes"
	"context"
	"fmt"
	v1 "jh_app_service/api/backend/upload/v1"
	"jh_app_service/internal/dao"
	"jh_app_service/internal/middleware"
	"jh_app_service/internal/service/backend"
	"jh_app_service/internal/tracing"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type (
	sUpload struct{}
)

func init() {
	backend.RegisterUpload(&sUpload{})
}

// UploadImage 上传图片
func (s *sUpload) UploadImage(ctx context.Context, req *v1.UploadImageReq) (*v1.UploadImageRes, error) {
	// 创建Jaeger span
	ctx, span := tracing.StartSpan(ctx, "upload.UploadImage", trace.WithAttributes(
		attribute.String("file_name", req.FileName),
		attribute.String("content_type", req.ContentType),
		attribute.Int64("file_size", req.FileSize),
		attribute.String("upload_code", req.UploadCode),
		attribute.String("method", "UploadImage"),
	))
	defer span.End()

	middleware.LogWithTrace(ctx, "info", "上传图片请求 - 文件名: %s, 大小: %d, 类型: %s",
		req.FileName, req.FileSize, req.ContentType)

	// 验证参数
	if err := s.validateUploadRequest(req); err != nil {
		tracing.AddSpanEvent(span, "validation_failed", attribute.String("reason", err.Error()))
		middleware.LogWithTrace(ctx, "error", "上传参数验证失败: %v", err)
		return nil, err
	}

	// 获取上传配置
	uploadCode := req.UploadCode
	if uploadCode == "" {
		uploadCode = "default"
	}

	// 验证文件类型
	if err := s.validateFileType(req.ContentType, uploadCode); err != nil {
		tracing.AddSpanEvent(span, "file_type_invalid",
			attribute.String("content_type", req.ContentType),
			attribute.String("upload_code", uploadCode))
		middleware.LogWithTrace(ctx, "error", "文件类型验证失败: %v", err)
		return nil, err
	}

	// 验证文件大小
	if err := s.validateFileSize(req.FileSize, uploadCode); err != nil {
		tracing.AddSpanEvent(span, "file_size_invalid",
			attribute.Int64("file_size", req.FileSize),
			attribute.String("upload_code", uploadCode))
		middleware.LogWithTrace(ctx, "error", "文件大小验证失败: %v", err)
		return nil, err
	}

	// 生成存储路径
	storagePath := s.generateStoragePath(ctx, req.FileName)
	tracing.SetSpanAttributes(span, attribute.String("storage_path", storagePath))

	// 上传到 MinIO
	ctx, uploadSpan := tracing.StartSpan(ctx, "minio.upload", trace.WithAttributes(
		attribute.String("storage.type", "minio"),
		attribute.String("storage.path", storagePath),
	))

	imageUrl, err := s.uploadToMinio(ctx, req.FileData, storagePath, req.ContentType)
	uploadSpan.End()

	if err != nil {
		tracing.SetSpanError(span, err)
		tracing.SetSpanError(uploadSpan, err)
		middleware.LogWithTrace(ctx, "error", "上传到MinIO失败: %v", err)
		return nil, fmt.Errorf("上传失败: %v", err)
	}

	tracing.AddSpanEvent(span, "upload_success",
		attribute.String("file_name", req.FileName),
		attribute.String("image_url", imageUrl),
	)
	tracing.SetSpanAttributes(span, attribute.Bool("success", true))

	middleware.LogWithTrace(ctx, "info", "上传图片成功 - 文件名: %s, URL: %s", req.FileName, imageUrl)

	res := &v1.UploadImageRes{
		ImageUrl: imageUrl,
		FilePath: storagePath,
		FileSize: req.FileSize,
	}

	return res, nil
}

// validateUploadRequest 验证上传请求参数
func (s *sUpload) validateUploadRequest(req *v1.UploadImageReq) error {
	if len(req.FileData) == 0 {
		return fmt.Errorf("文件数据不能为空")
	}

	if req.FileName == "" {
		return fmt.Errorf("文件名不能为空")
	}

	if req.ContentType == "" {
		return fmt.Errorf("文件类型不能为空")
	}

	if req.FileSize <= 0 {
		return fmt.Errorf("文件大小无效")
	}

	// 验证文件数据大小与声明大小是否一致
	if int64(len(req.FileData)) != req.FileSize {
		return fmt.Errorf("文件数据大小与声明大小不一致")
	}

	return nil
}

// validateFileType 验证文件类型
func (s *sUpload) validateFileType(contentType, uploadCode string) error {
	// 获取允许的文件类型配置
	allowedTypes := s.getAllowedFileTypes(uploadCode)

	for _, allowedType := range allowedTypes {
		if contentType == allowedType {
			return nil
		}
	}

	return fmt.Errorf("不支持的文件类型: %s", contentType)
}

// validateFileSize 验证文件大小
func (s *sUpload) validateFileSize(fileSize int64, uploadCode string) error {
	maxSize := s.getMaxFileSize(uploadCode)
	maxSizeBytes := maxSize * 1024 // 配置中是KB，转换为字节

	if fileSize > maxSizeBytes {
		return fmt.Errorf("文件大小不能超过 %dKB", maxSize)
	}

	return nil
}

// getAllowedFileTypes 获取允许的文件类型
func (s *sUpload) getAllowedFileTypes(uploadCode string) []string {
	ctx := context.Background()

	// 默认配置
	defaultTypes := []string{"image/jpg", "image/jpeg", "image/gif", "image/png", "image/webp"}

	// 从配置文件获取
	configKey := fmt.Sprintf("upload.%s.imgType", uploadCode)
	if types := g.Cfg().MustGet(ctx, configKey); !types.IsEmpty() {
		if typeSlice := types.Strings(); len(typeSlice) > 0 {
			return typeSlice
		}
	}

	// 如果是 mobile_logo，使用特殊配置（支持 svg/webp/png/jpg/jpeg）
	if uploadCode == "mobile_logo" {
		return []string{"image/svg+xml", "image/webp", "image/png", "image/jpg", "image/jpeg"}
	}

	return defaultTypes
}

// getMaxFileSize 获取最大文件大小 (KB)
func (s *sUpload) getMaxFileSize(uploadCode string) int64 {
	ctx := context.Background()

	// 默认大小 2MB (2048KB)
	defaultSize := int64(2048)

	// 从配置文件获取
	configKey := fmt.Sprintf("upload.%s.maxSize", uploadCode)
	if size := g.Cfg().MustGet(ctx, configKey); !size.IsEmpty() {
		return size.Int64()
	}

	// 如果是 mobile_logo，使用特殊配置
	if uploadCode == "mobile_logo" {
		return 100
	}

	return defaultSize
}

// generateStoragePath 生成存储路径
func (s *sUpload) generateStoragePath(ctx context.Context, fileName string) string {
	// 获取站点ID (这里需要根据实际情况获取，可能从上下文或配置中获取)
	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		middleware.LogWithTrace(ctx, "error", "未获取站点ID")
		return ""
	}
	siteCode, err := dao.Site.GetCodeById(ctx, siteId)
	if err != nil {
		return ""
	}

	// 生成路径: /site_code/YYYY/MM/filename
	now := time.Now()
	datePath := now.Format("2006/01")

	// 生成唯一文件名
	ext := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, ext)
	uniqueName := fmt.Sprintf("%s_%d%s", baseName, now.Unix(), ext)

	return fmt.Sprintf("%s/%s/%s", siteCode, datePath, uniqueName)
}

// uploadToMinio 上传文件到MinIO
func (s *sUpload) uploadToMinio(ctx context.Context, fileData []byte, objectPath, contentType string) (string, error) {
	// 获取MinIO配置
	endpoint := g.Cfg().MustGet(ctx, "minio.endpoint", "localhost:19000").String()
	accessKey := g.Cfg().MustGet(ctx, "minio.accessKey", "minioadmin").String()
	secretKey := g.Cfg().MustGet(ctx, "minio.secretKey", "minioadmin123").String()
	bucketName := g.Cfg().MustGet(ctx, "minio.bucket", "uploads").String()
	useSSL := g.Cfg().MustGet(ctx, "minio.useSSL", false).Bool()

	// 创建MinIO客户端
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return "", fmt.Errorf("创建MinIO客户端失败: %v", err)
	}

	// 检查bucket是否存在，不存在则创建
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return "", fmt.Errorf("检查bucket失败: %v", err)
	}

	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return "", fmt.Errorf("创建bucket失败: %v", err)
		}
		middleware.LogWithTrace(ctx, "info", "创建MinIO bucket: %s", bucketName)
	}

	// 上传文件
	reader := bytes.NewReader(fileData)
	_, err = minioClient.PutObject(ctx, bucketName, objectPath, reader, int64(len(fileData)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("上传文件到MinIO失败: %v", err)
	}

	// 生成访问URL
	baseURL := g.Cfg().MustGet(ctx, "minio.publicURL", fmt.Sprintf("http://%s", endpoint)).String()
	imageURL := fmt.Sprintf("%s/%s/%s", baseURL, bucketName, objectPath)

	return imageURL, nil
}
