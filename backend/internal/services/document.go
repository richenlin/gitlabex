package services

import (
	"fmt"
	"path/filepath"
	"time"

	"gitlabex/internal/models"

	"gorm.io/gorm"
)

// DocumentService 文档服务 - 管理文档附件和Wiki集成
type DocumentService struct {
	db     *gorm.DB
	gitlab *GitLabService
}

// NewDocumentService 创建文档服务
func NewDocumentService(db *gorm.DB, gitlab *GitLabService) *DocumentService {
	return &DocumentService{
		db:     db,
		gitlab: gitlab,
	}
}

// CreateWikiAttachment 创建Wiki附件
func (s *DocumentService) CreateWikiAttachment(userID int, projectID int, wikiSlug string, filename string, fileURL string, fileContent []byte) (*models.DocumentAttachment, error) {
	// 获取文件类型
	fileType := s.getFileType(filename)

	// 生成文档key
	documentKey := s.generateDocumentKey(filename)

	// 创建文档记录
	doc := &models.DocumentAttachment{
		UserID:       userID,
		ProjectID:    &projectID,
		WikiPageSlug: wikiSlug,
		FileName:     filename,
		FileURL:      fileURL,
		FilePath:     fmt.Sprintf("uploads/wiki/%d/%s/%s", projectID, wikiSlug, filename),
		FileType:     fileType,
		DocumentKey:  documentKey,
		EditMode:     "edit",
		Status:       "ready",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.db.Create(doc).Error; err != nil {
		return nil, fmt.Errorf("创建文档记录失败: %w", err)
	}

	return doc, nil
}

// GetWikiAttachments 获取Wiki附件列表
func (s *DocumentService) GetWikiAttachments(projectID int, wikiSlug string) ([]*models.DocumentAttachment, error) {
	var attachments []*models.DocumentAttachment

	err := s.db.Where("project_id = ? AND wiki_page_slug = ?", projectID, wikiSlug).
		Preload("LastEditor").
		Find(&attachments).Error

	if err != nil {
		return nil, fmt.Errorf("获取Wiki附件失败: %w", err)
	}

	return attachments, nil
}

// GetDocumentByID 根据ID获取文档
func (s *DocumentService) GetDocumentByID(docID int) (*models.DocumentAttachment, error) {
	var doc models.DocumentAttachment

	err := s.db.Where("id = ?", docID).
		Preload("LastEditor").
		First(&doc).Error

	if err != nil {
		return nil, fmt.Errorf("文档不存在: %w", err)
	}

	return &doc, nil
}

// UpdateDocumentEditStatus 更新文档编辑状态
func (s *DocumentService) UpdateDocumentEditStatus(docID int, userID int, status string) error {
	var doc models.DocumentAttachment

	if err := s.db.Where("id = ?", docID).First(&doc).Error; err != nil {
		return fmt.Errorf("文档不存在: %w", err)
	}

	now := time.Now()
	userIDUint := uint(userID)

	doc.Status = status
	doc.LastEditedBy = &userIDUint
	doc.LastEditedAt = &now
	doc.UpdatedAt = now

	if err := s.db.Save(&doc).Error; err != nil {
		return fmt.Errorf("更新文档状态失败: %w", err)
	}

	return nil
}

// DeleteWikiAttachment 删除Wiki附件
func (s *DocumentService) DeleteWikiAttachment(docID int, userID int) error {
	var doc models.DocumentAttachment

	if err := s.db.Where("id = ? AND user_id = ?", docID, userID).First(&doc).Error; err != nil {
		return fmt.Errorf("文档不存在或无权限: %w", err)
	}

	if err := s.db.Delete(&doc).Error; err != nil {
		return fmt.Errorf("删除文档失败: %w", err)
	}

	return nil
}

// GetProjectDocuments 获取项目的所有文档
func (s *DocumentService) GetProjectDocuments(projectID int) ([]*models.DocumentAttachment, error) {
	var attachments []*models.DocumentAttachment

	err := s.db.Where("project_id = ?", projectID).
		Preload("LastEditor").
		Order("updated_at DESC").
		Find(&attachments).Error

	if err != nil {
		return nil, fmt.Errorf("获取项目文档失败: %w", err)
	}

	return attachments, nil
}

// GetUserDocuments 获取用户的所有文档
func (s *DocumentService) GetUserDocuments(userID int) ([]*models.DocumentAttachment, error) {
	var attachments []*models.DocumentAttachment

	err := s.db.Where("user_id = ?", userID).
		Preload("LastEditor").
		Order("updated_at DESC").
		Find(&attachments).Error

	if err != nil {
		return nil, fmt.Errorf("获取用户文档失败: %w", err)
	}

	return attachments, nil
}

// 辅助方法

// getFileType 获取文件类型
func (s *DocumentService) getFileType(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".docx", ".doc":
		return "docx"
	case ".xlsx", ".xls":
		return "xlsx"
	case ".pptx", ".ppt":
		return "pptx"
	case ".pdf":
		return "pdf"
	default:
		return "unknown"
	}
}

// generateDocumentKey 生成文档key
func (s *DocumentService) generateDocumentKey(filename string) string {
	return fmt.Sprintf("wiki_%d_%s", time.Now().UnixNano(), filename)
}

// GetEditableDocuments 获取可编辑的文档
func (s *DocumentService) GetEditableDocuments(projectID int) ([]*models.DocumentAttachment, error) {
	var attachments []*models.DocumentAttachment

	err := s.db.Where("project_id = ? AND file_type IN (?)", projectID, []string{"docx", "xlsx", "pptx"}).
		Preload("LastEditor").
		Order("updated_at DESC").
		Find(&attachments).Error

	if err != nil {
		return nil, fmt.Errorf("获取可编辑文档失败: %w", err)
	}

	return attachments, nil
}

// GetDocumentStats 获取文档统计信息
func (s *DocumentService) GetDocumentStats(userID int) (map[string]interface{}, error) {
	var totalCount int64
	var recentCount int64

	// 获取用户相关文档总数
	s.db.Model(&models.DocumentAttachment{}).Where("user_id = ?", userID).Count(&totalCount)

	// 获取最近一周的文档数
	lastWeek := time.Now().AddDate(0, 0, -7)
	s.db.Model(&models.DocumentAttachment{}).Where("user_id = ? AND created_at > ?", userID, lastWeek).Count(&recentCount)

	return map[string]interface{}{
		"total_documents":  totalCount,
		"recent_documents": recentCount,
	}, nil
}
