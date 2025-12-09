package handlers

import (
	"fmt"
	"gitlabex/internal/models"
	"gitlabex/internal/services"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DocumentHandler 文档处理器
type DocumentHandler struct {
	documentService *services.DocumentService
	gitlabService   *services.GitLabService
	researchService *services.ResearchService
}

// NewDocumentHandler 创建文档处理器
func NewDocumentHandler(documentService *services.DocumentService, gitlabService *services.GitLabService, researchService *services.ResearchService) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
		gitlabService:   gitlabService,
		researchService: researchService,
	}
}

// 辅助函数：安全地从上下文获取值

// getInt64FromContextDoc 安全地从上下文获取int64值
func getInt64FromContextDoc(c *gin.Context, key string) (int64, bool) {
	value, exists := c.Get(key)
	if !exists {
		return 0, false
	}
	if v, ok := value.(int64); ok {
		return v, true
	}
	return 0, false
}

// getStringFromContextDoc 安全地从上下文获取string值
func getStringFromContextDoc(c *gin.Context, key string) (string, bool) {
	value, exists := c.Get(key)
	if !exists {
		return "", false
	}
	if v, ok := value.(string); ok {
		return v, true
	}
	return "", false
}

// getBoolFromContextDoc 安全地从上下文获取bool值
func getBoolFromContextDoc(c *gin.Context, key string) bool {
	value, exists := c.Get(key)
	if !exists {
		return false
	}
	if v, ok := value.(bool); ok {
		return v
	}
	return false
}

// GetDocuments 获取文档列表
func (h *DocumentHandler) GetDocuments(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// 获取搜索参数
	searchQuery := c.Query("search")

	// 获取筛选参数
	filters := make(map[string]interface{})
	// 支持两种参数名称：project_id（下划线）和 projectId（驼峰）
	projectID := c.Query("project_id")
	if projectID == "" {
		projectID = c.Query("projectId")
	}
	if projectID != "" {
		if id, err := uuid.Parse(projectID); err == nil {
			filters["project_id"] = id
		}
	}
	if fileType := c.Query("file_type"); fileType != "" {
		filters["file_type"] = fileType
	}
	if category := c.Query("category"); category != "" {
		filters["category"] = category
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if uploaderID := c.Query("uploader_id"); uploaderID != "" {
		filters["uploader_id"] = uploaderID
	}
	if searchQuery != "" {
		filters["search"] = searchQuery
	}

	documents, total, err := h.documentService.GetDocuments(limit, offset, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文档失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"documents": documents,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetDocumentByID 根据ID获取文档详情
func (h *DocumentHandler) GetDocumentByID(c *gin.Context) {
	documentIDStr := c.Param("id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文档ID"})
		return
	}

	document, err := h.documentService.GetDocumentByID(documentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文档不存在"})
		return
	}

	c.JSON(http.StatusOK, document)
}

// CreateDocument 创建新文档
func (h *DocumentHandler) CreateDocument(c *gin.Context) {
	var req struct {
		ProjectID   string `json:"project_id" binding:"required"`
		Title       string `json:"title"`
		FileName    string `json:"file_name" binding:"required"`
		FilePath    string `json:"file_path" binding:"required"`
		FileSize    int64  `json:"file_size"`
		Description string `json:"description"`
		Category    string `json:"category"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	document := &models.Document{
		ProjectID:   &projectID,
		Title:       req.Title,
		FilePath:    req.FilePath,
		FileSize:    req.FileSize,
		Description: req.Description,
		Category:    req.Category,
		UploaderID:  gitlabUserID.(int64), // 使用gitlab_user_id
		Status:      models.DocumentStatusApproved,
	}

	if err := h.documentService.CreateDocument(document); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建文档失败"})
		return
	}

	c.JSON(http.StatusCreated, document)
}

// UploadDocument 上传文档文件
func (h *DocumentHandler) UploadDocument(c *gin.Context) {
	// 获取项目ID
	projectIDStr := c.PostForm("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目ID不能为空"})
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	// 安全地获取用户ID
	userID, ok := getInt64FromContextDoc(c, "gitlab_user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 权限检查：所有登录用户都可以上传文档
	// 已通过获取userID验证

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取上传文件失败: " + err.Error()})
		return
	}
	defer file.Close()

	// 验证文件类型和大小
	if err := h.validateFile(header); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 保存文件
	filePath, err := h.saveUploadedFile(file, header, projectID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败: " + err.Error()})
		return
	}

	// 获取表单参数
	title := c.PostForm("title")
	if title == "" {
		title = header.Filename
	}
	description := c.PostForm("description")

	// 确定文件类型
	fileType := h.getFileType(header.Filename)
	mimeType := header.Header.Get("Content-Type")

	// 根据文件类型自动确定分类
	category := h.getCategoryFromFileType(fileType)

	// 创建文档记录
	document := &models.Document{
		ProjectID:   &projectID,
		Title:       title,
		Description: description,
		FilePath:    filePath,
		FileSize:    header.Size,
		FileType:    fileType,
		MIMEType:    mimeType,
		Category:    category,
		UploaderID:  userID,
		Status:      models.DocumentStatusApproved, // 默认审核通过，可根据需要修改
	}

	if err := h.documentService.CreateDocument(document); err != nil {
		// 如果创建数据库记录失败，删除已上传的文件
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建文档记录失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "文档上传成功",
		"document": document,
	})
}

// validateFile 验证上传的文件
func (h *DocumentHandler) validateFile(header *multipart.FileHeader) error {
	// 检查文件大小 (限制为50MB)
	const maxFileSize = 50 * 1024 * 1024 // 50MB
	if header.Size > maxFileSize {
		return fmt.Errorf("文件大小不能超过50MB")
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	// 移除点号，只保留扩展名
	ext = strings.TrimPrefix(ext, ".")

	// 从配置中获取允许的文件类型
	allowedFileTypes := strings.Split(h.documentService.Config.Upload.AllowedFileTypes, ",")
	allowedExts := make(map[string]bool)

	for _, fileType := range allowedFileTypes {
		fileType = strings.TrimSpace(fileType)
		allowedExts[fileType] = true
	}

	if !allowedExts[ext] {
		return fmt.Errorf("不支持的文件类型: %s", ext)
	}

	return nil
}

// saveUploadedFile 保存上传的文件
func (h *DocumentHandler) saveUploadedFile(file multipart.File, header *multipart.FileHeader, projectID string) (string, error) {
	// 创建上传目录
	uploadDir := filepath.Join("uploads", "documents", projectID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("创建上传目录失败: %v", err)
	}

	// 生成唯一的文件名
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), uuid.New().String()[:8], ext)
	filePath := filepath.Join(uploadDir, filename)

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("复制文件内容失败: %v", err)
	}

	return filePath, nil
}

// getFileType 根据文件名确定文件类型
func (h *DocumentHandler) getFileType(filename string) models.DocumentType {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".pdf":
		return models.DocumentTypePDF
	case ".doc", ".docx":
		return models.DocumentTypeWord
	case ".xls", ".xlsx", ".csv":
		return models.DocumentTypeExcel
	case ".ppt", ".pptx":
		return models.DocumentTypePPT
	case ".jpg", ".jpeg", ".png", ".gif":
		return models.DocumentTypeImage
	case ".mp4", ".avi", ".mov":
		return models.DocumentTypeVideo
	case ".go", ".js", ".py", ".java", ".cpp", ".c", ".h":
		return models.DocumentTypeCode
	default:
		return models.DocumentTypeOther
	}
}

// DownloadDocument 下载文档
func (h *DocumentHandler) DownloadDocument(c *gin.Context) {
	documentIDStr := c.Param("id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文档ID"})
		return
	}

	// 获取文档信息
	document, err := h.documentService.GetDocumentByID(documentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文档不存在"})
		return
	}

	// 从MINIO获取文件内容
	minioPath := document.MinIOPath
	if minioPath == "" {
		// 如果没有MINIO路径，使用文件路径作为后备
		minioPath = document.FilePath
	}

	_, fileReader, err := h.documentService.MinIOService.GetDocument(minioPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在或无法访问"})
		return
	}
	defer fileReader.Close()

	// 读取文件内容
	fileData, err := io.ReadAll(fileReader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
		return
	}

	// 更新下载次数
	h.documentService.IncrementDownloadCount(documentID)

	// 生成带扩展名的文件名
	fileName := document.Title
	if document.FileType != "" {
		// 如果标题没有扩展名，添加文件类型扩展名
		if !strings.Contains(fileName, ".") {
			fileName = fmt.Sprintf("%s.%s", fileName, document.FileType)
		}
	}

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))

	// 根据文件类型设置正确的Content-Type
	contentType := getContentTypeByFileType(document.FileType)
	c.Header("Content-Type", contentType)

	// 发送文件内容
	c.Data(http.StatusOK, contentType, fileData)
}

// UpdateDocument 更新文档信息
func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	documentIDStr := c.Param("id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文档ID"})
		return
	}

	// 安全地获取用户ID
	userID, ok := getInt64FromContextDoc(c, "gitlab_user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 获取文档信息
	document, err := h.documentService.GetDocumentByID(documentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文档不存在"})
		return
	}

	// 权限检查
	if !h.checkDocumentEditPermission(c, document, userID) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "权限不足",
			"message": "只有管理员、教师和文档所有者可以编辑文档",
		})
		return
	}

	var req struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Category    *string `json:"category"`
		Status      *string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := h.documentService.UpdateDocument(documentID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新文档失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文档更新成功"})
}

// DeleteDocument 删除文档
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	documentIDStr := c.Param("id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文档ID"})
		return
	}

	// 安全地获取用户ID
	userID, ok := getInt64FromContextDoc(c, "gitlab_user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 获取文档信息
	document, err := h.documentService.GetDocumentByID(documentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文档不存在"})
		return
	}

	// 权限检查：只有管理员、教师和文档所有者可以删除
	if !h.checkDocumentDeletePermission(c, document, userID) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "权限不足",
			"message": "只有管理员、教师和文档所有者可以删除文档",
		})
		return
	}

	if err := h.documentService.DeleteDocument(documentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除文档失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文档删除成功"})
}

// GetDocumentCategories 获取文档分类列表
func (h *DocumentHandler) GetDocumentCategories(c *gin.Context) {
	categories, err := h.documentService.GetDocumentCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取分类失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

// SearchDocuments 搜索文档
func (h *DocumentHandler) SearchDocuments(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	documents, total, err := h.documentService.SearchDocuments(keyword, limit, (page-1)*limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索文档失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"documents": documents,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetDocumentStats 获取文档统计信息
func (h *DocumentHandler) GetDocumentStats(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目ID不能为空"})
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	stats, err := h.documentService.GetDocumentStats(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计信息失败"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// SyncDocuments 同步项目文档
func (h *DocumentHandler) SyncDocuments(c *gin.Context) {
	// 安全地获取用户ID
	_, ok := getInt64FromContextDoc(c, "gitlab_user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未登录",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	projectIDStr := c.Param("project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	// 权限检查：只有管理员和教师可以同步文档
	if !h.checkDocumentSyncPermission(c, projectID) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "权限不足",
			"message": "只有管理员和教师可以同步文档",
			"code":    "FORBIDDEN",
		})
		return
	}

	// 获取项目信息以获取GitLab项目ID
	var project models.ResearchProject
	if err := h.documentService.DB.First(&project, projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}

	// 检查项目是否有GitLab项目ID
	if project.GitLabProjectID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该项目未关联GitLab项目"})
		return
	}

	gitLabProjectID := *project.GitLabProjectID

	// 安全地获取GitLab访问令牌
	accessToken, ok := getStringFromContextDoc(c, "gitlab_access_token")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitLab访问令牌未找到"})
		return
	}

	if err := h.documentService.SyncDocumentsFromGitLab(projectID, gitLabProjectID, accessToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "同步文档失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文档同步成功"})
}

// ScanProjectDocuments 扫描项目文档
func (h *DocumentHandler) ScanProjectDocuments(c *gin.Context) {
	// 检查用户登录状态
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未登录",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	// 验证gitlab_user_id是否有效
	if gitlabUserID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户认证信息无效",
			"code":  "INVALID_USER",
		})
		return
	}

	projectIDStr := c.Param("project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	// 获取项目信息以获取GitLab项目ID
	var project models.ResearchProject
	if err := h.documentService.DB.First(&project, projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}

	// 检查项目是否有GitLab项目ID
	if project.GitLabProjectID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该项目未关联GitLab项目"})
		return
	}

	gitLabProjectID := *project.GitLabProjectID

	// 从JWT token中获取GitLab访问令牌
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitLab访问令牌未找到"})
		return
	}

	if err := h.documentService.ScanProjectDocuments(projectID, gitLabProjectID, accessToken.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "扫描文档失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文档扫描成功"})
}

// SubmitEditRequest 提交文档编辑请求
func (h *DocumentHandler) SubmitEditRequest(c *gin.Context) {
	documentIDStr := c.Param("id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文档ID"})
		return
	}

	// 安全地获取当前用户ID
	userID, ok := getInt64FromContextDoc(c, "gitlab_user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	var req struct {
		ProposedChanges struct {
			Title       string   `json:"title"`
			Description string   `json:"description"`
			Category    string   `json:"category"`
			Tags        []string `json:"tags"`
		} `json:"proposed_changes"`
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 构建编辑数据
	editData := make(map[string]interface{})
	if req.ProposedChanges.Title != "" {
		editData["title"] = req.ProposedChanges.Title
	}
	if req.ProposedChanges.Description != "" {
		editData["description"] = req.ProposedChanges.Description
	}
	if req.ProposedChanges.Category != "" {
		editData["category"] = req.ProposedChanges.Category
	}
	if len(req.ProposedChanges.Tags) > 0 {
		editData["tags"] = req.ProposedChanges.Tags
	}

	editRequest, err := h.documentService.SubmitEditRequest(documentID, userID, editData, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, editRequest)
}

// GetEditRequests 获取文档编辑请求列表
func (h *DocumentHandler) GetEditRequests(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	offset := (page - 1) * pageSize

	var reviewerID *int64
	if reviewerIDStr := c.Query("reviewer_id"); reviewerIDStr != "" {
		if id, err := strconv.ParseInt(reviewerIDStr, 10, 64); err == nil {
			reviewerID = &id
		}
	}

	// 安全地获取当前用户ID
	currentUserID, ok := getInt64FromContextDoc(c, "gitlab_user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	isAdmin := getBoolFromContextDoc(c, "is_admin")
	// 注意: CheckUserPermission函数可能不存在,使用简化的检查
	isTeacher := false // 将在后续改进中实现

	requests, total, err := h.documentService.GetEditRequestsWithPermissionFilter(
		status, reviewerID, pageSize, offset, currentUserID, isAdmin, isTeacher)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取编辑请求失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":    requests,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// ReviewEditRequest 审核文档编辑请求
func (h *DocumentHandler) ReviewEditRequest(c *gin.Context) {
	requestIDStr := c.Param("id")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求ID"})
		return
	}

	// 获取当前用户ID
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	reviewerID := gitlabUserID.(int64)

	// 获取编辑请求信息以确定审核权限
	editRequest, err := h.documentService.GetEditRequestByID(requestID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "编辑请求不存在"})
		return
	}

	// 获取关联的文档信息
	document, err := h.documentService.GetDocumentByID(editRequest.DocumentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "关联文档不存在"})
		return
	}

	// 检查审核权限
	var canReview bool
	if document.IsStandalone {
		// 独立文档：创建者和管理员可以审核
		isAdmin := h.CheckUserPermission(c, "admin")
		isCreator := document.UploaderID == reviewerID
		canReview = isAdmin || isCreator

		if !canReview {
			c.JSON(http.StatusForbidden, gin.H{"error": "权限不足，只有文档创建者和管理员可以审核独立文档的编辑请求"})
			return
		}
	} else {
		// 课题文档：只有管理员和教师可以审核编辑请求
		canReview = h.CheckUserPermission(c, "admin") || h.CheckUserPermission(c, "teacher")
		if !canReview {
			c.JSON(http.StatusForbidden, gin.H{"error": "权限不足，只有管理员和教师可以审核编辑请求"})
			return
		}
	}

	var req struct {
		Approved bool   `json:"approved"`
		Comments string `json:"comments"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.documentService.ReviewEditRequest(requestID, reviewerID, req.Approved, req.Comments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "审核完成"})
}

// GetDocumentEditHistory 获取文档编辑历史
func (h *DocumentHandler) GetDocumentEditHistory(c *gin.Context) {
	documentIDStr := c.Param("id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文档ID"})
		return
	}

	history, err := h.documentService.GetDocumentEditHistory(documentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取编辑历史失败"})
		return
	}

	c.JSON(http.StatusOK, history)
}

// SyncProjectDocumentsToMinIO 同步课题文档到MinIO
func (h *DocumentHandler) SyncProjectDocumentsToMinIO(c *gin.Context) {
	// 注意：用户认证由RequireAuth中间件处理，权限检查由前端通过/api/v1/permissions/check接口完成

	projectIDStr := c.Param("project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	gitLabProjectIDStr := c.Query("gitlab_project_id")
	if gitLabProjectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "GitLab项目ID不能为空"})
		return
	}

	gitLabProjectID, err := strconv.ParseInt(gitLabProjectIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的GitLab项目ID"})
		return
	}

	// 获取用户access token，支持多种格式
	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {
		// 尝试从其他地方获取token
		accessToken = c.GetHeader("X-Access-Token")
	}
	if accessToken == "" {
		accessToken = c.Query("access_token")
	}

	// 处理Bearer token格式
	if strings.HasPrefix(accessToken, "Bearer ") {
		accessToken = strings.TrimPrefix(accessToken, "Bearer ")
	}

	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "需要访问令牌",
			"code":  "MISSING_TOKEN",
		})
		return
	}

	if err := h.documentService.SyncProjectDocumentsToMinIO(projectID, gitLabProjectID, accessToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "同步文档到MinIO失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文档同步到MinIO成功"})
}

// CreateStandaloneDocument 创建独立文档
func (h *DocumentHandler) CreateStandaloneDocument(c *gin.Context) {
	// 🔍 调试：打印请求信息
	log.Printf("[CreateStandaloneDocument] 开始处理上传请求")
	log.Printf("[CreateStandaloneDocument] Content-Type: %s", c.Request.Header.Get("Content-Type"))
	log.Printf("[CreateStandaloneDocument] Content-Length: %d", c.Request.ContentLength)

	// 获取用户ID
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		log.Printf("[CreateStandaloneDocument] 错误: 用户未登录")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}
	log.Printf("[CreateStandaloneDocument] 用户ID: %v", gitlabUserID)

	// ✅ 先解析 multipart 表单
	log.Printf("[CreateStandaloneDocument] 尝试解析multipart表单...")
	if err := c.Request.ParseMultipartForm(100 << 20); err != nil { // 100 MB
		log.Printf("[CreateStandaloneDocument] 错误: 解析multipart表单失败: %v", err)
		log.Printf("[CreateStandaloneDocument] 请求Method: %s", c.Request.Method)
		log.Printf("[CreateStandaloneDocument] 请求Headers: %+v", c.Request.Header)
		c.JSON(http.StatusBadRequest, gin.H{"error": "解析multipart表单失败: " + err.Error()})
		return
	}
	log.Printf("[CreateStandaloneDocument] 成功解析multipart表单")
	log.Printf("[CreateStandaloneDocument] MultipartForm.File keys: %+v", c.Request.MultipartForm.File)

	// 获取上传的文件
	log.Printf("[CreateStandaloneDocument] 尝试获取FormFile('file')...")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("[CreateStandaloneDocument] 错误: 获取上传文件失败: %v", err)
		log.Printf("[CreateStandaloneDocument] 可用的文件字段: %+v", c.Request.MultipartForm.File)
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取上传文件失败: " + err.Error()})
		return
	}
	defer file.Close()
	log.Printf("[CreateStandaloneDocument] 成功获取文件: %s, 大小: %d", header.Filename, header.Size)

	// 验证文件类型和大小
	if err := h.validateFile(header); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 读取文件内容
	fileData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件内容失败: " + err.Error()})
		return
	}

	// 获取表单参数
	title := c.PostForm("title")
	if title == "" {
		title = strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	}
	description := c.PostForm("description")
	if description == "" {
		description = title
	}

	// 确定文件类型
	fileType := h.getFileType(header.Filename)

	// 根据文件类型自动确定分类
	category := h.getCategoryFromFileType(fileType)

	// 创建独立文档
	document, err := h.documentService.CreateStandaloneDocument(
		title,
		description,
		category,
		gitlabUserID.(int64),
		fileData,
		fileType,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建独立文档失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "独立文档创建成功",
		"document": document,
	})
}

// GetPredefinedCategories 获取预定义分类
func (h *DocumentHandler) GetPredefinedCategories(c *gin.Context) {
	categories := h.documentService.GetPredefinedCategories()
	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

// GetDocumentDownloadURL 获取文档下载URL
func (h *DocumentHandler) GetDocumentDownloadURL(c *gin.Context) {
	documentIDStr := c.Param("id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文档ID"})
		return
	}

	// 获取过期时间参数，默认24小时
	expiresStr := c.DefaultQuery("expires", "24h")
	expires, err := time.ParseDuration(expiresStr)
	if err != nil {
		expires = 24 * time.Hour
	}

	url, err := h.documentService.GetDocumentURL(documentID, expires)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取下载URL失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"download_url": url,
		"expires_in":   expires.String(),
	})
}

// CheckUserPermission 检查用户权限（基于GitLab角色）
func (h *DocumentHandler) CheckUserPermission(c *gin.Context, requiredRole string) bool {
	// 检查是否为管理员
	isAdmin, exists := c.Get("is_admin")
	if exists && isAdmin.(bool) {
		return true // 管理员拥有所有权限
	}

	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		return false
	}

	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		return false
	}

	// 根据所需角色检查权限
	switch requiredRole {
	case "admin":
		return isAdmin != nil && isAdmin.(bool)
	case "teacher":
		// 检查用户是否在Teachers组中（教师角色）
		hasTeacherRole, err := h.gitlabService.CheckUserInGroup(
			accessToken.(string),
			gitlabUserID.(int64),
			10, // Teachers组ID
		)
		if err != nil {
			fmt.Printf("检查教师权限时出错: %v\n", err)
			return false
		}
		return hasTeacherRole
	case "researcher":
		// 检查用户是否在Teaching Assistants组中（研究员角色）
		hasResearcherRole, err := h.gitlabService.CheckUserInGroup(
			accessToken.(string),
			gitlabUserID.(int64),
			11, // Teaching Assistants组ID
		)
		if err != nil {
			fmt.Printf("检查研究员权限时出错: %v\n", err)
			return false
		}
		return hasResearcherRole
	case "student":
		// 检查用户是否在Students组中（学生角色）
		hasStudentRole, err := h.gitlabService.CheckUserInGroup(
			accessToken.(string),
			gitlabUserID.(int64),
			12, // Students组ID
		)
		if err != nil {
			fmt.Printf("检查学生权限时出错: %v\n", err)
			return false
		}
		return hasStudentRole
	default:
		return false
	}
}

// UpdateDocumentWithPermissionCheck 带权限检查的文档更新
func (h *DocumentHandler) UpdateDocumentWithPermissionCheck(c *gin.Context) {
	documentIDStr := c.Param("id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文档ID"})
		return
	}

	// 获取当前用户ID
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	var req struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Category    *string `json:"category"`
		Status      *string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取文档信息以检查是否为独立文档
	document, err := h.documentService.GetDocumentByID(documentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文档不存在"})
		return
	}

	// 检查用户权限
	var canDirectEdit bool

	if document.IsStandalone {
		// 独立文档：创建者和管理员可以直接修改
		isAdmin := h.CheckUserPermission(c, "admin")
		isCreator := document.UploaderID == gitlabUserID.(int64)
		canDirectEdit = isAdmin || isCreator
	} else {
		// 课题文档：管理员/教师可以直接修改，研究员/学生需要提交审核请求
		canDirectEdit = h.CheckUserPermission(c, "admin") || h.CheckUserPermission(c, "teacher")
	}

	// 如果不能直接编辑，检查是否有提交编辑请求的权限
	if !canDirectEdit {
		hasEditRequestPermission := h.CheckUserPermission(c, "researcher") ||
			h.CheckUserPermission(c, "student") ||
			(document.IsStandalone) // 独立文档的其他用户也可以提交编辑请求

		if !hasEditRequestPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "权限不足，无法修改文档"})
			return
		}
	}

	if canDirectEdit {
		// 直接更新文档
		updates := make(map[string]interface{})
		if req.Title != nil {
			updates["title"] = *req.Title
		}
		if req.Description != nil {
			updates["description"] = *req.Description
		}
		if req.Category != nil {
			updates["category"] = *req.Category
		}
		if req.Status != nil {
			updates["status"] = *req.Status
		}

		if err := h.documentService.UpdateDocument(documentID, updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新文档失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "文档更新成功"})
	} else {
		// 提交编辑请求
		editData := make(map[string]interface{})
		if req.Title != nil {
			editData["title"] = *req.Title
		}
		if req.Description != nil {
			editData["description"] = *req.Description
		}
		if req.Category != nil {
			editData["category"] = *req.Category
		}

		editRequest, err := h.documentService.SubmitEditRequest(
			documentID,
			gitlabUserID.(int64),
			editData,
			"用户请求修改文档属性",
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":      "编辑请求已提交，等待审核",
			"edit_request": editRequest,
		})
	}
}

// getContentTypeByFileType 根据文件类型获取Content-Type
func getContentTypeByFileType(fileType models.DocumentType) string {
	switch fileType {
	case models.DocumentTypePDF:
		return "application/pdf"
	case models.DocumentTypeWord:
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case models.DocumentTypeExcel:
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case models.DocumentTypePPT:
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case models.DocumentTypeText:
		return "text/plain"
	case models.DocumentTypeMarkdown:
		return "text/markdown"
	case models.DocumentTypeCode:
		return "text/plain"
	case models.DocumentTypeImage:
		return "image/jpeg"
	case models.DocumentTypeVideo:
		return "video/mp4"
	default:
		return "application/octet-stream"
	}
}

// getCategoryFromFileType 根据文件类型自动确定分类
func (h *DocumentHandler) getCategoryFromFileType(fileType models.DocumentType) string {
	switch fileType {
	case models.DocumentTypePDF:
		return "pdf"
	case models.DocumentTypeWord:
		return "word"
	case models.DocumentTypeExcel:
		return "excel"
	case models.DocumentTypePPT:
		return "ppt"
	case models.DocumentTypeText, models.DocumentTypeMarkdown:
		return "text"
	case models.DocumentTypeImage:
		return "image"
	case models.DocumentTypeVideo:
		return "video"
	case models.DocumentTypeCode:
		return "code"
	default:
		return "other"
	}
}

// checkDocumentEditPermission 检查文档编辑权限
// 权限规则：
// 1. 管理员可以编辑所有文档
// 2. 教师可以直接编辑所有文档
// 3. 研究员可以编辑自己的文档
// 4. 学生可以编辑自己的文档(需要审核)
func (h *DocumentHandler) checkDocumentEditPermission(c *gin.Context, document *models.Document, userID int64) bool {
	// 检查管理员权限
	if getBoolFromContextDoc(c, "is_admin") {
		return true
	}

	// 检查是否为文档所有者
	if document.UploaderID == userID {
		// 文档所有者可以编辑自己的文档
		return true
	}

	// 检查教师权限 (需要关联项目)
	if document.ProjectID != nil {
		project, err := h.researchService.GetResearchProjectByID(*document.ProjectID)
		if err == nil && project.GitLabProjectID != nil {
			accessToken, ok := getStringFromContextDoc(c, "gitlab_access_token")
			if ok {
				accessLevel, err := h.gitlabService.GetUserProjectAccessLevel(accessToken, *project.GitLabProjectID)
				// Maintainer (40+) 可以编辑所有文档
				if err == nil && accessLevel >= 40 {
					return true
				}
			}
		}
	}

	return false
}

// checkDocumentDeletePermission 检查文档删除权限
// 权限规则：
// 1. 管理员可以删除所有文档
// 2. 教师可以删除所有文档
// 3. 文档所有者可以删除自己的文档
func (h *DocumentHandler) checkDocumentDeletePermission(c *gin.Context, document *models.Document, userID int64) bool {
	// 检查管理员权限
	if getBoolFromContextDoc(c, "is_admin") {
		return true
	}

	// 检查是否为文档所有者
	if document.UploaderID == userID {
		return true
	}

	// 检查教师权限
	if document.ProjectID != nil {
		project, err := h.researchService.GetResearchProjectByID(*document.ProjectID)
		if err == nil && project.GitLabProjectID != nil {
			accessToken, ok := getStringFromContextDoc(c, "gitlab_access_token")
			if ok {
				accessLevel, err := h.gitlabService.GetUserProjectAccessLevel(accessToken, *project.GitLabProjectID)
				// Maintainer (40+) 可以删除所有文档
				if err == nil && accessLevel >= 40 {
					return true
				}
			}
		}
	}

	return false
}

// checkDocumentSyncPermission 检查文档同步权限
// 权限规则：
// 1. 管理员可以同步文档
// 2. 教师可以同步文档
func (h *DocumentHandler) checkDocumentSyncPermission(c *gin.Context, projectID uuid.UUID) bool {
	// 检查管理员权限
	if getBoolFromContextDoc(c, "is_admin") {
		return true
	}

	// 获取用户ID
	userID, ok := getInt64FromContextDoc(c, "gitlab_user_id")
	if !ok {
		return false
	}

	// 获取项目信息
	project, err := h.researchService.GetResearchProjectByID(projectID)
	if err != nil {
		return false
	}

	// 检查是否为课题创建者
	if project.CreatorID == userID {
		return true
	}

	// 检查教师权限
	if project.GitLabProjectID != nil {
		accessToken, ok := getStringFromContextDoc(c, "gitlab_access_token")
		if ok {
			accessLevel, err := h.gitlabService.GetUserProjectAccessLevel(accessToken, *project.GitLabProjectID)
			// Maintainer (40+) 可以同步文档
			if err == nil && accessLevel >= 40 {
				return true
			}
		}
	}

	return false
}

// checkDocumentApprovePermission 检查文档审核权限
// 权限规则：
// 1. 管理员可以审核文档
// 2. 教师可以审核文档
func (h *DocumentHandler) checkDocumentApprovePermission(c *gin.Context, document *models.Document) bool {
	// 检查管理员权限
	if getBoolFromContextDoc(c, "is_admin") {
		return true
	}

	// 获取用户ID
	userID, ok := getInt64FromContextDoc(c, "gitlab_user_id")
	if !ok {
		return false
	}

	// 检查教师权限
	if document.ProjectID != nil {
		project, err := h.researchService.GetResearchProjectByID(*document.ProjectID)
		if err == nil {
			// 检查是否为课题创建者
			if project.CreatorID == userID {
				return true
			}

			// 检查GitLab项目权限
			if project.GitLabProjectID != nil {
				accessToken, ok := getStringFromContextDoc(c, "gitlab_access_token")
				if ok {
					accessLevel, err := h.gitlabService.GetUserProjectAccessLevel(accessToken, *project.GitLabProjectID)
					// Maintainer (40+) 可以审核文档
					if err == nil && accessLevel >= 40 {
						return true
					}
				}
			}
		}
	}

	return false
}
