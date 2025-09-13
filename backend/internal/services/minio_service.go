package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOService MinIO对象存储服务
type MinIOService struct {
	client          *minio.Client
	endpoint        string
	documentsBucket string
	avatarsBucket   string
	tempBucket      string
	urlExpiresIn    time.Duration
}

// DocumentInfo 文档信息结构
type DocumentInfo struct {
	Key          string            `json:"key"`
	Name         string            `json:"name"`
	Size         int64             `json:"size"`
	ContentType  string            `json:"content_type"`
	LastModified time.Time         `json:"last_modified"`
	URL          string            `json:"url,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// NewMinIOService 创建MinIO服务实例
func NewMinIOService(endpoint, accessKey, secretKey string, useSSL bool, region string) (*MinIOService, error) {
	// 初始化MinIO客户端
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	service := &MinIOService{
		client:          client,
		endpoint:        endpoint,
		documentsBucket: "documents",
		avatarsBucket:   "avatars",
		tempBucket:      "temp",
		urlExpiresIn:    24 * time.Hour,
	}

	// 初始化存储桶
	if err := service.initializeBuckets(); err != nil {
		return nil, fmt.Errorf("failed to initialize buckets: %w", err)
	}

	return service, nil
}

// initializeBuckets 初始化存储桶
func (s *MinIOService) initializeBuckets() error {
	ctx := context.Background()
	buckets := []string{s.documentsBucket, s.avatarsBucket, s.tempBucket}

	for _, bucket := range buckets {
		exists, err := s.client.BucketExists(ctx, bucket)
		if err != nil {
			return fmt.Errorf("failed to check bucket existence: %w", err)
		}

		if !exists {
			err = s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{
				Region: "us-east-1",
			})
			if err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", bucket, err)
			}
			log.Printf("Created bucket: %s", bucket)
		}
	}

	return nil
}

// UploadDocument 上传文档到MinIO
func (s *MinIOService) UploadDocument(key string, data []byte, contentType string, metadata map[string]string) (*DocumentInfo, error) {
	ctx := context.Background()

	// 如果没有提供contentType，尝试从文件扩展名推断
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(key))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	// 准备上传选项
	options := minio.PutObjectOptions{
		ContentType:  contentType,
		UserMetadata: metadata,
	}

	// 上传文件
	reader := bytes.NewReader(data)
	_, err := s.client.PutObject(ctx, s.documentsBucket, key, reader, int64(len(data)), options)
	if err != nil {
		return nil, fmt.Errorf("failed to upload document: %w", err)
	}

	// 获取文档信息
	objInfo, err := s.client.StatObject(ctx, s.documentsBucket, key, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object info: %w", err)
	}

	// 生成下载URL
	downloadURL := fmt.Sprintf("%s/%s/%s", s.endpoint, s.documentsBucket, key)

	return &DocumentInfo{
		Key:          key,
		Name:         filepath.Base(key),
		Size:         objInfo.Size,
		ContentType:  objInfo.ContentType,
		LastModified: objInfo.LastModified,
		URL:          downloadURL,
		Metadata:     objInfo.UserMetadata,
	}, nil
}

// GetDocument 获取文档
func (s *MinIOService) GetDocument(key string) (*DocumentInfo, io.ReadCloser, error) {
	ctx := context.Background()

	// 获取对象信息
	objInfo, err := s.client.StatObject(ctx, s.documentsBucket, key, minio.StatObjectOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get object info: %w", err)
	}

	// 获取对象数据
	reader, err := s.client.GetObject(ctx, s.documentsBucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get object: %w", err)
	}

	docInfo := &DocumentInfo{
		Key:          key,
		Name:         filepath.Base(key),
		Size:         objInfo.Size,
		ContentType:  objInfo.ContentType,
		LastModified: objInfo.LastModified,
		Metadata:     objInfo.UserMetadata,
	}

	return docInfo, reader, nil
}

// GetDocumentURL 获取文档的预签名URL
func (s *MinIOService) GetDocumentURL(key string, expires time.Duration) (string, error) {
	ctx := context.Background()

	if expires == 0 {
		expires = s.urlExpiresIn
	}

	// 生成预签名URL
	url, err := s.client.PresignedGetObject(ctx, s.documentsBucket, key, expires, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// DeleteDocument 删除文档
func (s *MinIOService) DeleteDocument(key string) error {
	ctx := context.Background()

	err := s.client.RemoveObject(ctx, s.documentsBucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// ListDocuments 列出文档
func (s *MinIOService) ListDocuments(prefix string, recursive bool) ([]*DocumentInfo, error) {
	ctx := context.Background()

	var documents []*DocumentInfo

	// 列出对象
	objectCh := s.client.ListObjects(ctx, s.documentsBucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}

		documents = append(documents, &DocumentInfo{
			Key:          object.Key,
			Name:         filepath.Base(object.Key),
			Size:         object.Size,
			ContentType:  object.ContentType,
			LastModified: object.LastModified,
		})
	}

	return documents, nil
}

// CopyDocument 复制文档
func (s *MinIOService) CopyDocument(srcKey, destKey string, metadata map[string]string) error {
	ctx := context.Background()

	// 设置复制源
	src := minio.CopySrcOptions{
		Bucket: s.documentsBucket,
		Object: srcKey,
	}

	// 设置复制目标
	dest := minio.CopyDestOptions{
		Bucket:       s.documentsBucket,
		Object:       destKey,
		UserMetadata: metadata,
	}

	_, err := s.client.CopyObject(ctx, dest, src)
	if err != nil {
		return fmt.Errorf("failed to copy document: %w", err)
	}

	return nil
}

// GetDocumentsByType 根据文件类型获取文档
func (s *MinIOService) GetDocumentsByType(fileTypes []string) ([]*DocumentInfo, error) {
	allDocs, err := s.ListDocuments("", true)
	if err != nil {
		return nil, err
	}

	var filteredDocs []*DocumentInfo
	for _, doc := range allDocs {
		ext := strings.ToLower(filepath.Ext(doc.Key))
		if ext != "" {
			ext = ext[1:] // 移除点号
		}

		for _, fileType := range fileTypes {
			if strings.ToLower(fileType) == ext {
				filteredDocs = append(filteredDocs, doc)
				break
			}
		}
	}

	return filteredDocs, nil
}

// IsDocumentType 检查是否为文档类型
func IsDocumentType(filename string) bool {
	documentTypes := []string{
		"doc", "docx", "pdf", "ppt", "pptx",
		"xls", "xlsx", "txt", "md", "rtf", "csv",
		"odt", "ods", "odp", "csv",
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext != "" {
		ext = ext[1:] // 移除点号
	}

	for _, docType := range documentTypes {
		if docType == ext {
			return true
		}
	}

	return false
}

// GetDocumentStats 获取文档统计信息
func (s *MinIOService) GetDocumentStats() (map[string]interface{}, error) {
	allDocs, err := s.ListDocuments("", true)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_count": len(allDocs),
		"total_size":  int64(0),
		"by_type":     make(map[string]int),
	}

	var totalSize int64
	typeCount := make(map[string]int)

	for _, doc := range allDocs {
		totalSize += doc.Size

		ext := strings.ToLower(filepath.Ext(doc.Key))
		if ext != "" {
			ext = ext[1:] // 移除点号
		}
		if ext == "" {
			ext = "unknown"
		}

		typeCount[ext]++
	}

	stats["total_size"] = totalSize
	stats["by_type"] = typeCount

	return stats, nil
}
