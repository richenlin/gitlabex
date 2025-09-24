package services

import (
	"fmt"
	"log"
	"time"

	"gitlabex/internal/models"
)

// DocumentScannerService 文档扫描服务
type DocumentScannerService struct {
	minioService    *MinIOService
	documentService *DocumentService
}

// SimpleScanResult 扫描结果
type SimpleScanResult struct {
	TotalDocuments int      `json:"total_documents"`
	NewDocuments   int      `json:"new_documents"`
	UpdatedDocs    int      `json:"updated_docs"`
	Errors         []string `json:"errors,omitempty"`
}

// NewDocumentScannerService 创建文档扫描服务
func NewDocumentScannerService(minioService *MinIOService, documentService *DocumentService) *DocumentScannerService {
	return &DocumentScannerService{
		minioService:    minioService,
		documentService: documentService,
	}
}

// ScanMinIODocuments 扫描MinIO中的文档并同步到数据库
func (s *DocumentScannerService) ScanMinIODocuments() (*SimpleScanResult, error) {
	result := &SimpleScanResult{
		Errors: make([]string, 0),
	}

	// 获取MinIO中的所有文档
	documents, err := s.minioService.ListDocuments("", true)
	if err != nil {
		return nil, fmt.Errorf("failed to list MinIO documents: %w", err)
	}

	result.TotalDocuments = len(documents)
	log.Printf("发现 %d 个MinIO文档", len(documents))

	// 处理每个文档
	for _, doc := range documents {
		// 检查数据库中是否已存在该文档
		existingDoc, err := s.documentService.GetDocumentByFilePath(doc.Key)
		if err != nil && err.Error() != "document not found" {
			result.Errors = append(result.Errors, fmt.Sprintf("检查文档失败 %s: %v", doc.Key, err))
			continue
		}

		if existingDoc == nil {
			// 创建新的文档记录
			document := &models.Document{
				Title:       doc.Name,
				Description: fmt.Sprintf("从MinIO自动同步的文档: %s", doc.Key),
				FilePath:    doc.Key,
				FileSize:    doc.Size,
				FileType:    models.DocumentType(getFileTypeFromKey(doc.Key)),
				Status:      "active",
			}

			if err := s.documentService.CreateDocument(document); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("创建文档记录失败 %s: %v", doc.Key, err))
				continue
			}

			result.NewDocuments++
			log.Printf("新增文档记录: %s", doc.Key)
		} else {
			// 更新现有文档记录
			existingDoc.FileSize = doc.Size

			if err := s.documentService.UpdateDocumentRecord(existingDoc); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("更新文档记录失败 %s: %v", doc.Key, err))
				continue
			}

			result.UpdatedDocs++
			log.Printf("更新文档记录: %s", doc.Key)
		}
	}

	log.Printf("MinIO文档扫描完成: 总计 %d, 新增 %d, 更新 %d, 错误 %d",
		result.TotalDocuments, result.NewDocuments, result.UpdatedDocs, len(result.Errors))

	return result, nil
}

// GetSyncStats 获取同步统计信息
func (s *DocumentScannerService) GetSyncStats() (map[string]interface{}, error) {
	// 获取MinIO统计
	minioStats, err := s.minioService.GetDocumentStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get MinIO stats: %w", err)
	}

	// 获取数据库统计
	dbStats, err := s.documentService.GetAllDocumentStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get database stats: %w", err)
	}

	stats := map[string]interface{}{
		"minio":     minioStats,
		"database":  dbStats,
		"sync_time": time.Now(),
	}

	return stats, nil
}

// getFileTypeFromKey 从MinIO键获取文件类型
func getFileTypeFromKey(key string) string {
	// 使用现有的getFileType函数
	return getFileType(key)
}
