package services

import (
	"bytes"
	"encoding/base64"
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
	MinIOService  *MinIOService
}

// NewDocumentService 创建文档服务
func NewDocumentService(db *gorm.DB, gitlabService *GitLabService, minioService *MinIOService) *DocumentService {
	return &DocumentService{
		BaseService:   NewBaseService(db, gitlabService.Config),
		GitLabService: gitlabService,
		MinIOService:  minioService,
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
	err := query.Preload("Project").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&documents).Error

	return documents, total, err
}

// GetDocumentByID 根据ID获取文档
func (s *DocumentService) GetDocumentByID(id uuid.UUID) (*models.Document, error) {
	var document models.Document
	err := s.DB.Preload("Project").
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
	// 先获取项目信息以获取默认分支
	project, err := s.GitLabService.GetProject(accessToken, gitLabProjectID)
	if err != nil {
		return fmt.Errorf("获取GitLab项目信息失败: %v", err)
	}

	// 使用项目的默认分支
	defaultBranch := project.DefaultBranch
	if defaultBranch == "" {
		defaultBranch = "main" // 如果默认分支为空，使用main作为后备
	}

	// 获取GitLab项目文件树
	files, err := s.GitLabService.GetRepositoryTree(accessToken, gitLabProjectID, "", defaultBranch)
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

		// 获取文件内容
		fileContent, err := s.GitLabService.GetFileContent(accessToken, gitLabProjectID, filePath, defaultBranch)
		if err != nil {
			// 如果无法获取内容，跳过该文件
			continue
		}

		// 解码base64内容
		contentBytes, err := base64.StdEncoding.DecodeString(fileContent.Content)
		if err != nil {
			// 如果解码失败，跳过该文件
			continue
		}

		// 生成唯一的文件名
		fileExt := filepath.Ext(fileName)
		uniqueFileName := fmt.Sprintf("%s_%s%s",
			strings.TrimSuffix(fileName, fileExt),
			uuid.New().String()[:8],
			fileExt)

		// 将文件内容保存到MINIO
		minioPath := fmt.Sprintf("documents/%s/%s", projectID.String(), uniqueFileName)
		_, err = s.MinIOService.UploadDocument(minioPath, contentBytes, fileContent.Encoding, nil)
		if err != nil {
			// 如果保存到MINIO失败，跳过该文件
			continue
		}

		// 获取文件大小
		fileSize := int64(len(contentBytes))

		// 创建文档记录
		document := &models.Document{
			ProjectID:      &projectID,
			Title:          strings.TrimSuffix(fileName, fileExt),
			Description:    extractDescriptionFromContent(string(contentBytes), fileName),
			FilePath:       filePath, // 使用原始GitLab路径作为文件路径
			FileType:       models.DocumentType(docType),
			Category:       categorizeDocument(filePath, docType),
			Status:         models.DocumentStatusPending,
			FileSize:       fileSize,
			GitLabFilePath: filePath, // 保留原始GitLab路径
			GitLabID:       fmt.Sprintf("%d", gitLabProjectID),
			AutoIndexed:    true,      // 标记为自动索引的文档
			MinIOPath:      minioPath, // 存储MINIO路径
		}

		if err := s.CreateDocument(document); err != nil {
			// 如果创建文档记录失败，删除MINIO中的文件
			s.MinIOService.DeleteDocument(minioPath)
			continue
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

// categorizeDocument 根据文件类型自动分类文档
func categorizeDocument(filePath string, fileType string) string {
	// 根据文件类型分类
	switch fileType {
	case "pdf":
		return "pdf"
	case "doc", "docx":
		return "word"
	case "xls", "xlsx", "csv":
		return "excel"
	case "ppt", "pptx":
		return "powerpoint"
	case "txt":
		return "text"
	case "md":
		return "markdown"
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
		ProjectID:      &projectID,
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

// GetDocumentCategories 获取文档分类列表（基于配置的允许文件类型）
func (s *DocumentService) GetDocumentCategories() ([]string, error) {
	// 从配置中获取允许的文件类型
	allowedFileTypes := strings.Split(s.Config.AllowedFileTypes, ",")

	// 将文件扩展名映射到分类
	categoryMap := make(map[string]bool)

	for _, fileType := range allowedFileTypes {
		fileType = strings.TrimSpace(fileType)
		switch fileType {
		case "pdf":
			categoryMap["pdf"] = true
		case "doc", "docx":
			categoryMap["word"] = true
		case "xls", "xlsx", "csv":
			categoryMap["excel"] = true
		case "ppt", "pptx":
			categoryMap["ppt"] = true
		case "txt", "md":
			categoryMap["text"] = true
		case "jpg", "jpeg", "png", "gif":
			categoryMap["image"] = true
		case "zip", "rar", "7z":
			categoryMap["archive"] = true
		default:
			categoryMap["other"] = true
		}
	}

	// 转换为切片
	var categories []string
	for category := range categoryMap {
		categories = append(categories, category)
	}

	return categories, nil
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
		return "xls"
	case ".ppt", ".pptx":
		return "ppt"
	case ".txt":
		return "txt"
	case ".md":
		return "md"
	default:
		return "other"
	}
}

// GetDocumentByPath 根据GitLab文件路径获取文档
func (s *DocumentService) GetDocumentByPath(projectID uuid.UUID, gitlabFilePath string) (*models.Document, error) {
	var document models.Document
	err := s.DB.Where("project_id = ? AND file_path = ?", projectID, gitlabFilePath).First(&document).Error
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
	err := query.Preload("Project").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&documents).Error

	return documents, total, err
}

// SubmitEditRequest 提交文档编辑请求（学生使用）
func (s *DocumentService) SubmitEditRequest(documentID uuid.UUID, requesterID int64, editData map[string]interface{}, reason string) (*models.DocumentEditRequest, error) {
	// 检查文档是否存在
	var document models.Document
	if err := s.DB.First(&document, documentID).Error; err != nil {
		return nil, err
	}

	// 检查用户权限：只有已登录的GitLab用户才能提交编辑请求
	// requesterID > 0 表示是有效的GitLab用户ID
	if requesterID <= 0 {
		return nil, fmt.Errorf("无效的用户ID")
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
func (s *DocumentService) ReviewEditRequest(requestID uuid.UUID, reviewerID int64, approved bool, comments string) error {
	// TODO: 重构审核者权限检查以使用GitLab用户系统
	// 暂时跳过权限检查

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
		Order("created_at DESC").
		Find(&history).Error

	return history, err
}

// IncrementDownloadCount 增加文档下载次数
func (s *DocumentService) IncrementDownloadCount(documentID uuid.UUID) error {
	return s.DB.Model(&models.Document{}).
		Where("id = ?", documentID).
		UpdateColumn("download_count", gorm.Expr("download_count + ?", 1)).
		Error
}

// GetDocumentByFilePath 根据文件路径获取文档（用于扫描服务）
func (s *DocumentService) GetDocumentByFilePath(path string) (*models.Document, error) {
	var document models.Document
	err := s.DB.Where("file_path = ?", path).First(&document).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("document not found")
		}
		return nil, err
	}
	return &document, nil
}

// UpdateDocumentRecord 更新文档记录（用于扫描服务）
func (s *DocumentService) UpdateDocumentRecord(document *models.Document) error {
	return s.DB.Save(document).Error
}

// UploadToMinIO 上传文档到MinIO
func (s *DocumentService) UploadToMinIO(documentID uuid.UUID, fileData []byte, contentType string) error {
	// 获取文档信息
	document, err := s.GetDocumentByID(documentID)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}

	// 生成MinIO存储键
	minioKey := fmt.Sprintf("documents/%s/%s", documentID.String(), document.Title)
	if document.FilePath != "" {
		minioKey = fmt.Sprintf("documents/%s/%s", documentID.String(), filepath.Base(document.FilePath))
	}

	// 准备元数据
	metadata := map[string]string{
		"document_id": documentID.String(),
		"title":       document.Title,
		"file_type":   string(document.FileType),
		"uploaded_at": time.Now().Format(time.RFC3339),
	}

	// 上传到MinIO
	_, err = s.MinIOService.UploadDocument(minioKey, fileData, contentType, metadata)
	if err != nil {
		return fmt.Errorf("failed to upload to MinIO: %w", err)
	}

	// 更新数据库中的文件路径
	err = s.DB.Model(document).Update("file_path", minioKey).Error
	if err != nil {
		return fmt.Errorf("failed to update document path: %w", err)
	}

	return nil
}

// GetDocumentFromMinIO 从MinIO获取文档
func (s *DocumentService) GetDocumentFromMinIO(documentID uuid.UUID) ([]byte, string, error) {
	// 获取文档信息
	document, err := s.GetDocumentByID(documentID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get document: %w", err)
	}

	// 从MinIO获取文件
	docInfo, reader, err := s.MinIOService.GetDocument(document.FilePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get document from MinIO: %w", err)
	}
	defer reader.Close()

	// 读取文件内容
	var buffer bytes.Buffer
	_, err = buffer.ReadFrom(reader)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read document content: %w", err)
	}

	return buffer.Bytes(), docInfo.ContentType, nil
}

// GetDocumentURL 获取文档的预签名URL
func (s *DocumentService) GetDocumentURL(documentID uuid.UUID, expires time.Duration) (string, error) {
	// 获取文档信息
	document, err := s.GetDocumentByID(documentID)
	if err != nil {
		return "", fmt.Errorf("failed to get document: %w", err)
	}

	// 获取预签名URL
	url, err := s.MinIOService.GetDocumentURL(document.FilePath, expires)
	if err != nil {
		return "", fmt.Errorf("failed to get document URL: %w", err)
	}

	return url, nil
}

// DeleteFromMinIO 从MinIO删除文档
func (s *DocumentService) DeleteFromMinIO(documentID uuid.UUID) error {
	// 获取文档信息
	document, err := s.GetDocumentByID(documentID)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}

	// 从MinIO删除文件
	err = s.MinIOService.DeleteDocument(document.FilePath)
	if err != nil {
		return fmt.Errorf("failed to delete document from MinIO: %w", err)
	}

	return nil
}

// GetAllDocumentStats 获取所有文档统计信息（增强版，包含MinIO）
func (s *DocumentService) GetAllDocumentStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 数据库统计
	var totalCount int64
	s.DB.Model(&models.Document{}).Count(&totalCount)
	stats["total_count"] = totalCount

	// 按文件类型统计
	var typeCounts []struct {
		FileType string
		Count    int64
	}
	s.DB.Model(&models.Document{}).
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
		Select("status, count(*) as count").
		Group("status").
		Find(&statusCounts)

	statusStats := make(map[string]int64)
	for _, sc := range statusCounts {
		statusStats[sc.Status] = sc.Count
	}
	stats["statuses"] = statusStats

	// 如果有MinIO服务，获取存储统计
	if s.MinIOService != nil {
		minioStats, err := s.MinIOService.GetDocumentStats()
		if err == nil {
			stats["storage"] = minioStats
		}
	}

	return stats, nil
}

// SyncDocumentWithMinIO 同步文档到MinIO
func (s *DocumentService) SyncDocumentWithMinIO(documentID uuid.UUID, fileData []byte) error {
	document, err := s.GetDocumentByID(documentID)
	if err != nil {
		return err
	}

	// 生成存储键
	minioKey := fmt.Sprintf("documents/%s/%s", documentID.String(), document.Title)

	// 准备元数据
	metadata := map[string]string{
		"document_id": documentID.String(),
		"title":       document.Title,
		"file_type":   string(document.FileType),
		"synced_at":   time.Now().Format(time.RFC3339),
	}

	// 上传到MinIO
	_, err = s.MinIOService.UploadDocument(minioKey, fileData, "", metadata)
	if err != nil {
		return err
	}

	// 更新数据库记录
	syncTime := time.Now()
	return s.DB.Model(document).Updates(map[string]interface{}{
		"minio_path":     minioKey,
		"file_size":      int64(len(fileData)),
		"last_sync_time": &syncTime,
	}).Error
}

// SyncProjectDocumentsToMinIO 同步课题文档到MinIO
func (s *DocumentService) SyncProjectDocumentsToMinIO(projectID uuid.UUID, gitLabProjectID int64, accessToken string) error {
	// 获取GitLab项目文件树
	files, err := s.GitLabService.GetRepositoryTree(accessToken, gitLabProjectID, "", "master")
	if err != nil {
		return fmt.Errorf("获取GitLab文件树失败: %w", err)
	}

	var processedFiles int
	var errors []string

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
			// 文档已存在，检查是否需要更新
			continue
		}

		// 获取文件内容
		fileContent, err := s.GitLabService.GetFileContent(accessToken, gitLabProjectID, filePath, "master")
		if err != nil {
			errors = append(errors, fmt.Sprintf("获取文件内容失败 %s: %v", filePath, err))
			continue
		}

		// 解码文件内容
		var fileBytes []byte
		if fileContent.Encoding == "base64" {
			// 如果是base64编码，需要解码
			// 这里简化处理，实际应该使用base64解码
			fileBytes = []byte(fileContent.Content)
		} else {
			fileBytes = []byte(fileContent.Content)
		}

		// 获取文件大小
		fileSize := int64(len(fileBytes))

		// 创建文档记录
		document := &models.Document{
			ProjectID:      &projectID,
			Title:          strings.TrimSuffix(fileName, filepath.Ext(fileName)),
			Description:    extractDescriptionFromContent(fileContent.Content, fileName),
			FilePath:       filePath,
			FileType:       models.DocumentType(docType),
			Category:       categorizeDocument(filePath, docType),
			Status:         models.DocumentStatusApproved,
			FileSize:       fileSize,
			GitLabFilePath: filePath,
			GitLabID:       fmt.Sprintf("%d", gitLabProjectID),
			AutoIndexed:    true,
		}

		// 先创建文档记录
		if err := s.CreateDocument(document); err != nil {
			errors = append(errors, fmt.Sprintf("创建文档记录失败 %s: %v", filePath, err))
			continue
		}

		// 同步到MinIO
		if err := s.SyncDocumentWithMinIO(document.ID, fileBytes); err != nil {
			errors = append(errors, fmt.Sprintf("同步到MinIO失败 %s: %v", filePath, err))
			// 删除已创建的文档记录
			s.DeleteDocument(document.ID)
			continue
		}

		processedFiles++
	}

	if len(errors) > 0 {
		return fmt.Errorf("同步过程中出现错误: %v", errors)
	}

	return nil
}

// CreateStandaloneDocument 创建独立文档（不关联课题）
func (s *DocumentService) CreateStandaloneDocument(title, description, category string, uploaderID int64, fileData []byte, fileType models.DocumentType) (*models.Document, error) {
	// 获取或创建独立文档项目
	standaloneProjectID, err := s.getOrCreateStandaloneProject()
	if err != nil {
		return nil, fmt.Errorf("获取独立文档项目失败: %w", err)
	}

	// 创建文档记录
	document := &models.Document{
		Title:        title,
		Description:  description,
		Category:     category,
		UploaderID:   uploaderID,
		FileType:     fileType,
		Status:       models.DocumentStatusApproved,
		IsStandalone: true,
		FileSize:     int64(len(fileData)),
		ProjectID:    &standaloneProjectID,
	}

	// 先创建数据库记录
	if err := s.CreateDocument(document); err != nil {
		return nil, fmt.Errorf("创建文档记录失败: %w", err)
	}

	// 同步到MinIO
	if err := s.SyncDocumentWithMinIO(document.ID, fileData); err != nil {
		// 如果MinIO同步失败，删除数据库记录
		s.DeleteDocument(document.ID)
		return nil, fmt.Errorf("同步到MinIO失败: %w", err)
	}

	return document, nil
}

// GetPredefinedCategories 获取预定义的文档分类
func (s *DocumentService) GetPredefinedCategories() []string {
	return []string{
		"documentation", // 文档
		"tutorial",      // 教程
		"assignment",    // 作业
		"lecture",       // 讲义
		"lab",           // 实验
		"exam",          // 考试
		"notes",         // 笔记
		"pdf",           // PDF文档
		"word",          // Word文档
		"presentation",  // 演示文稿
		"spreadsheet",   // 电子表格
		"code",          // 代码
		"text",          // 文本
		"other",         // 其他
	}
}

// GetDocumentDownloadURL 获取文档下载URL
func (s *DocumentService) GetDocumentDownloadURL(document *models.Document) string {
	if document.MinIOPath == "" {
		return ""
	}

	// 使用配置中的MinIO endpoint构建下载URL
	minioEndpoint := s.Config.MinIOEndpoint
	if !strings.HasPrefix(minioEndpoint, "http") {
		// 如果没有协议前缀，添加http://
		if s.Config.MinIOUseSSL {
			minioEndpoint = "https://" + minioEndpoint
		} else {
			minioEndpoint = "http://" + minioEndpoint
		}
	}

	return fmt.Sprintf("%s/%s", minioEndpoint, document.MinIOPath)
}

// getOrCreateStandaloneProject 获取或创建独立文档项目
func (s *DocumentService) getOrCreateStandaloneProject() (uuid.UUID, error) {
	// 查找名为"独立文档"的项目
	var project models.ResearchProject
	err := s.DB.Where("name = ?", "独立文档").First(&project).Error

	if err == nil {
		// 找到了，返回项目ID
		return project.ID, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// 其他数据库错误
		return uuid.Nil, err
	}

	// 没有找到，创建新的独立文档项目
	standaloneProject := &models.ResearchProject{
		Name:        "独立文档",
		Description: "用于存储独立文档的虚拟项目",
		Status:      "active",
		CreatorID:   1, // 使用系统用户ID
		IsPublic:    true,
	}

	if err := s.DB.Create(standaloneProject).Error; err != nil {
		return uuid.Nil, err
	}

	return standaloneProject.ID, nil
}
