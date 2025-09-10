package services

import (
	"errors"
	"fmt"
	"gitlabex/internal/models"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DocumentService 文档服务
type DocumentService struct {
	*BaseService
	GitLabService *GitLabService
}

// NewDocumentService 创建文档服务
func NewDocumentService(db *gorm.DB, gitlabService *GitLabService) *DocumentService {
	return &DocumentService{
		BaseService:   NewBaseService(db, gitlabService.Config),
		GitLabService: gitlabService,
	}
}

// GetDocuments 获取文档列表
func (s *DocumentService) GetDocuments(limit, offset int, filters map[string]interface{}) ([]models.Document, int64, error) {
	var documents []models.Document
	var total int64

	query := s.DB.Model(&models.Document{})

	// 应用筛选条件
	if projectID, ok := filters["project_id"]; ok {
		query = query.Where("project_id = ?", projectID)
	}
	if fileType, ok := filters["file_type"]; ok {
		query = query.Where("file_type = ?", fileType)
	}
	if category, ok := filters["category"]; ok {
		query = query.Where("category = ?", category)
	}
	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if uploaderID, ok := filters["uploader_id"]; ok {
		query = query.Where("uploader_id = ?", uploaderID)
	}

	// 计算总数
	query.Count(&total)

	// 获取分页数据
	err := query.Preload("Project").Preload("Uploader").Preload("Reviewer").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&documents).Error

	return documents, total, err
}

// GetDocumentByID 根据ID获取文档
func (s *DocumentService) GetDocumentByID(id uuid.UUID) (*models.Document, error) {
	var document models.Document
	err := s.DB.Preload("Project").Preload("Uploader").Preload("Reviewer").
		First(&document, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}

// CreateDocument 创建文档
func (s *DocumentService) CreateDocument(document *models.Document) error {
	// 自动识别文件类型
	fileName := strings.TrimSpace(document.FilePath)
	if fileName != "" {
		document.FileType = models.DocumentType(getFileType(fileName))

		// 如果标题为空，使用文件名作为标题
		if document.Title == "" {
			document.Title = strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
		}

		// 如果描述为空，使用标题作为描述
		if document.Description == "" {
			document.Description = document.Title
		}
	}

	return s.DB.Create(document).Error
}

// UpdateDocument 更新文档
func (s *DocumentService) UpdateDocument(id uuid.UUID, updates map[string]interface{}) error {
	// 如果更新了文件名，重新识别文件类型
	if fileName, ok := updates["file_name"]; ok {
		updates["file_type"] = getFileType(fileName.(string))
	}

	return s.DB.Model(&models.Document{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteDocument 删除文档
func (s *DocumentService) DeleteDocument(id uuid.UUID) error {
	return s.DB.Delete(&models.Document{}, "id = ?", id).Error
}

// UpdateDocumentStatus 更新文档状态
func (s *DocumentService) UpdateDocumentStatus(id uuid.UUID, status string, reviewerID *uuid.UUID, reviewNote string) error {
	updates := map[string]interface{}{
		"status":      status,
		"review_note": reviewNote,
	}

	if reviewerID != nil {
		updates["reviewer_id"] = reviewerID
		updates["reviewed_at"] = time.Now()
	}

	return s.DB.Model(&models.Document{}).Where("id = ?", id).Updates(updates).Error
}

// ScanProjectDocuments 扫描项目中的文档
func (s *DocumentService) ScanProjectDocuments(projectID uuid.UUID, gitLabProjectID int64, accessToken string) error {
	// 获取GitLab项目文件树
	files, err := s.GitLabService.GetRepositoryTree(accessToken, gitLabProjectID, "", "master")
	if err != nil {
		return err
	}

	var processedFiles int
	for _, fileData := range files {
		// 从map中提取文件信息
		fileType, ok := fileData["type"].(string)
		if !ok || fileType != "blob" {
			continue // 只处理文件，跳过目录
		}

		fileName, ok := fileData["name"].(string)
		if !ok {
			continue
		}

		filePath, ok := fileData["path"].(string)
		if !ok {
			continue
		}

		// 检查文件扩展名，只处理支持的文档类型
		docType := getFileType(fileName)
		if docType == "other" {
			continue
		}

		// 检查是否已经存在该文档
		existingDoc, err := s.GetDocumentByPath(projectID, filePath)
		if err == nil && existingDoc != nil {
			// 文档已存在，跳过
			continue
		}

		// 获取文件内容用于提取描述
		fileContent, err := s.GitLabService.GetFileContent(accessToken, gitLabProjectID, filePath, "master")
		var content string
		if err != nil {
			// 如果无法获取内容，仍然创建文档记录
			content = ""
		} else {
			content = fileContent.Content
		}

		// 获取文件大小（如果可用）
		var fileSize int64
		if sizeInterface, exists := fileData["size"]; exists {
			if sizeFloat, ok := sizeInterface.(float64); ok {
				fileSize = int64(sizeFloat)
			}
		}

		// 创建文档记录
		document := &models.Document{
			ProjectID:      projectID,
			Title:          strings.TrimSuffix(fileName, filepath.Ext(fileName)),
			Description:    extractDescriptionFromContent(content, fileName),
			FilePath:       filePath,
			FileType:       models.DocumentType(docType),
			Category:       categorizeDocument(filePath, docType),
			Status:         models.DocumentStatusPending,
			FileSize:       fileSize,
			GitLabFilePath: filePath,
			GitLabID:       fmt.Sprintf("%d", gitLabProjectID),
		}

		if err := s.CreateDocument(document); err != nil {
			continue // 跳过失败的文档，继续处理其他文件
		}
		processedFiles++
	}

	return nil
}

// extractDescriptionFromContent 从文件内容中提取描述
func extractDescriptionFromContent(content string, filename string) string {
	if content == "" {
		return fmt.Sprintf("自动从 %s 提取的文档", filename)
	}

	// 对于文本文件，提取前200个字符作为描述
	lines := strings.Split(content, "\n")
	description := ""
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			description = line
			break
		}
	}

	if len(description) > 200 {
		description = description[:200] + "..."
	}

	if description == "" {
		description = fmt.Sprintf("自动从 %s 提取的文档", filename)
	}

	return description
}

// categorizeDocument 根据文件路径和类型自动分类文档
func categorizeDocument(filePath string, fileType string) string {
	// 根据路径关键词分类
	path := strings.ToLower(filePath)

	if strings.Contains(path, "docs/") || strings.Contains(path, "documentation/") {
		return "documentation"
	}
	if strings.Contains(path, "tutorial/") || strings.Contains(path, "guide/") {
		return "tutorial"
	}
	if strings.Contains(path, "assignment/") || strings.Contains(path, "homework/") {
		return "assignment"
	}
	if strings.Contains(path, "lecture/") || strings.Contains(path, "slide/") {
		return "lecture"
	}
	if strings.Contains(path, "lab/") || strings.Contains(path, "experiment/") {
		return "lab"
	}
	if strings.Contains(path, "exam/") || strings.Contains(path, "test/") {
		return "exam"
	}
	if strings.Contains(path, "note/") || strings.Contains(path, "notes/") {
		return "notes"
	}

	// 根据文件类型分类
	switch fileType {
	case "pdf":
		return "pdf"
	case "doc", "docx":
		return "word"
	case "ppt", "pptx":
		return "presentation"
	case "excel":
		return "spreadsheet"
	case "python", "java", "cpp", "c", "go", "rust", "javascript", "typescript":
		return "code"
	case "markdown", "text":
		return "text"
	default:
		return "other"
	}
}

// SyncDocumentsFromGitLab 从GitLab同步所有文档
func (s *DocumentService) SyncDocumentsFromGitLab(projectID uuid.UUID, gitLabProjectID int64, accessToken string) error {
	// 先删除该项目现有的自动索引文档
	s.DB.Where("project_id = ? AND auto_indexed = ?", projectID, true).Delete(&models.Document{})

	// 重新扫描并索引文档
	return s.ScanProjectDocuments(projectID, gitLabProjectID, accessToken)
}

// UpdateDocumentFromGitLab 更新单个文档信息
func (s *DocumentService) UpdateDocumentFromGitLab(projectID uuid.UUID, gitLabProjectID int64, filePath string) error {
	// 获取最新文件信息
	// 需要添加accessToken参数
	return fmt.Errorf("syncDocumentFromGitLab method needs access token parameter")
}

// createDocumentFromGitLabFile 从GitLab文件创建文档记录
func (s *DocumentService) createDocumentFromGitLabFile(projectID uuid.UUID, gitLabProjectID int64, filePath string, accessToken string) error {
	// 获取文件信息
	fileContent, err := s.GitLabService.GetFileContent(accessToken, gitLabProjectID, filePath, "master")
	if err != nil {
		return err
	}

	fileName := filepath.Base(filePath)
	fileType := getFileType(fileName)

	// 跳过不支持的文件类型
	if fileType == "other" {
		return nil
	}

	syncTime := time.Now()
	document := &models.Document{
		ProjectID:      projectID,
		Title:          strings.TrimSuffix(fileName, filepath.Ext(fileName)),
		Description:    extractDescriptionFromContent(fileContent.Content, fileName),
		FilePath:       filePath,
		FileType:       models.DocumentType(fileType),
		Category:       categorizeDocument(filePath, fileType),
		Status:         models.DocumentStatusPending,
		FileSize:       int64(len(fileContent.Content)),
		GitLabFilePath: filePath,
		GitLabID:       fmt.Sprintf("%d", gitLabProjectID),
		AutoIndexed:    true,
		LastSyncTime:   &syncTime,
	}

	return s.CreateDocument(document)
}

// GetDocumentCategories 获取文档分类列表
func (s *DocumentService) GetDocumentCategories() ([]string, error) {
	var categories []string
	err := s.DB.Model(&models.Document{}).Distinct().Pluck("category", &categories).Error
	return categories, err
}

// GetDocumentStats 获取文档统计信息
func (s *DocumentService) GetDocumentStats(projectID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总文档数量
	var totalCount int64
	s.DB.Model(&models.Document{}).Where("project_id = ?", projectID).Count(&totalCount)
	stats["total_count"] = totalCount

	// 按文件类型统计
	var typeCounts []struct {
		FileType string
		Count    int64
	}
	s.DB.Model(&models.Document{}).
		Where("project_id = ?", projectID).
		Select("file_type, count(*) as count").
		Group("file_type").
		Find(&typeCounts)

	fileTypeStats := make(map[string]int64)
	for _, tc := range typeCounts {
		fileTypeStats[tc.FileType] = tc.Count
	}
	stats["file_types"] = fileTypeStats

	// 按状态统计
	var statusCounts []struct {
		Status string
		Count  int64
	}
	s.DB.Model(&models.Document{}).
		Where("project_id = ?", projectID).
		Select("status, count(*) as count").
		Group("status").
		Find(&statusCounts)

	statusStats := make(map[string]int64)
	for _, sc := range statusCounts {
		statusStats[sc.Status] = sc.Count
	}
	stats["statuses"] = statusStats

	return stats, nil
}

// getFileType 根据文件扩展名获取文件类型
func getFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pdf":
		return "pdf"
	case ".doc", ".docx":
		return "doc"
	case ".xls", ".xlsx":
		return "excel"
	case ".ppt", ".pptx":
		return "ppt"
	case ".txt":
		return "text"
	case ".md":
		return "markdown"
	case ".py":
		return "python"
	case ".js", ".jsx":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".java":
		return "java"
	case ".cpp", ".cc", ".cxx":
		return "cpp"
	case ".c":
		return "c"
	case ".go":
		return "go"
	case ".rs":
		return "rust"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	case ".json":
		return "json"
	case ".xml":
		return "xml"
	case ".zip", ".rar", ".7z":
		return "archive"
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg":
		return "image"
	default:
		return "other"
	}
}

// GetDocumentByPath 根据文件路径获取文档
func (s *DocumentService) GetDocumentByPath(projectID uuid.UUID, filePath string) (*models.Document, error) {
	var document models.Document
	err := s.DB.Where("project_id = ? AND file_path = ?", projectID, filePath).First(&document).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}

// SearchDocuments 搜索文档
func (s *DocumentService) SearchDocuments(keyword string, limit, offset int) ([]models.Document, int64, error) {
	var documents []models.Document
	var total int64

	query := s.DB.Model(&models.Document{}).
		Where("title ILIKE ? OR description ILIKE ? OR file_name ILIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")

	// 计算总数
	query.Count(&total)

	// 获取分页数据
	err := query.Preload("Project").Preload("Uploader").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&documents).Error

	return documents, total, err
}

// SubmitEditRequest 提交文档编辑请求（学生使用）
func (s *DocumentService) SubmitEditRequest(documentID uuid.UUID, requesterID uuid.UUID, editData map[string]interface{}, reason string) (*models.DocumentEditRequest, error) {
	// 检查用户权限 - 只有学生可以提交编辑请求
	var requester models.User
	if err := s.DB.First(&requester, requesterID).Error; err != nil {
		return nil, err
	}

	if requester.EduRole != models.EduRoleStudent {
		return nil, errors.New("只有学生可以提交文档编辑请求")
	}

	// 检查文档是否存在
	var document models.Document
	if err := s.DB.First(&document, documentID).Error; err != nil {
		return nil, err
	}

	// 创建编辑请求
	editRequest := &models.DocumentEditRequest{
		DocumentID:  documentID,
		RequesterID: requesterID,
		Reason:      reason,
		Status:      models.DocumentStatusPending,
	}

	// 设置要修改的字段
	if title, ok := editData["title"].(string); ok {
		editRequest.Title = title
	}
	if description, ok := editData["description"].(string); ok {
		editRequest.Description = description
	}
	if category, ok := editData["category"].(string); ok {
		editRequest.Category = category
	}
	if tags, ok := editData["tags"].([]string); ok {
		editRequest.Tags = tags
	}

	if err := s.DB.Create(editRequest).Error; err != nil {
		return nil, err
	}

	// 预加载关联数据
	s.DB.Preload("Document").Preload("Requester").First(editRequest, editRequest.ID)

	return editRequest, nil
}

// GetEditRequests 获取文档编辑请求列表
func (s *DocumentService) GetEditRequests(status string, reviewerID *uuid.UUID, limit, offset int) ([]models.DocumentEditRequest, int64, error) {
	var requests []models.DocumentEditRequest
	var total int64

	query := s.DB.Model(&models.DocumentEditRequest{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if reviewerID != nil {
		query = query.Where("reviewer_id = ?", *reviewerID)
	}

	// 计算总数
	query.Count(&total)

	// 获取分页数据
	err := query.Preload("Document").Preload("Requester").Preload("Reviewer").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&requests).Error

	return requests, total, err
}

// ReviewEditRequest 审核文档编辑请求
func (s *DocumentService) ReviewEditRequest(requestID uuid.UUID, reviewerID uuid.UUID, approved bool, comments string) error {
	// 检查审核者权限
	var reviewer models.User
	if err := s.DB.First(&reviewer, reviewerID).Error; err != nil {
		return err
	}

	if reviewer.EduRole != models.EduRoleTeacher && reviewer.EduRole != models.EduRoleAdmin {
		return errors.New("只有教师和管理员可以审核文档编辑请求")
	}

	// 获取编辑请求
	var editRequest models.DocumentEditRequest
	if err := s.DB.Preload("Document").First(&editRequest, requestID).Error; err != nil {
		return err
	}

	if editRequest.Status != models.DocumentStatusPending {
		return errors.New("只能审核待处理的编辑请求")
	}

	// 开始事务
	tx := s.DB.Begin()

	// 更新编辑请求状态
	now := time.Now()
	editRequest.ReviewerID = &reviewerID
	editRequest.ReviewComments = comments
	editRequest.ReviewedAt = &now

	if approved {
		editRequest.Status = models.DocumentStatusApproved

		// 应用修改到原文档
		updates := make(map[string]interface{})
		if editRequest.Title != "" {
			updates["title"] = editRequest.Title
		}
		if editRequest.Description != "" {
			updates["description"] = editRequest.Description
		}
		if editRequest.Category != "" {
			updates["category"] = editRequest.Category
		}
		if len(editRequest.Tags) > 0 {
			updates["tags"] = editRequest.Tags
		}

		if len(updates) > 0 {
			if err := tx.Model(&editRequest.Document).Updates(updates).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	} else {
		editRequest.Status = models.DocumentStatusRejected
	}

	if err := tx.Save(&editRequest).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetDocumentEditHistory 获取文档编辑历史
func (s *DocumentService) GetDocumentEditHistory(documentID uuid.UUID) ([]models.DocumentEditRequest, error) {
	var history []models.DocumentEditRequest

	err := s.DB.Where("document_id = ?", documentID).
		Preload("Requester").Preload("Reviewer").
		Order("created_at DESC").
		Find(&history).Error

	return history, err
}
