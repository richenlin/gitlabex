package handlers

import (
	"fmt"
	"gitlabex/internal/models"
	"gitlabex/internal/services"
	"io"
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
}

// NewDocumentHandler 创建文档处理器
func NewDocumentHandler(documentService *services.DocumentService) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
	}
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

	// 获取筛选参数
	filters := make(map[string]interface{})
	if projectID := c.Query("project_id"); projectID != "" {
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

	// 获取用户ID
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

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
	category := c.PostForm("category")

	// 确定文件类型
	fileType := h.getFileType(header.Filename)
	mimeType := header.Header.Get("Content-Type")

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
		UploaderID:  gitlabUserID.(int64),          // 使用gitlab_user_id
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
	allowedExts := map[string]bool{
		".pdf":  true,
		".doc":  true,
		".docx": true,
		".xls":  true,
		".xlsx": true,
		".ppt":  true,
		".pptx": true,
		".txt":  true,
		".md":   true,
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".zip":  true,
		".rar":  true,
		".7z":   true,
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
	case ".xls", ".xlsx":
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

	// 检查文件是否存在
	if _, err := os.Stat(document.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		return
	}

	// 更新下载次数
	h.documentService.IncrementDownloadCount(documentID)

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", document.Title))
	c.Header("Content-Type", "application/octet-stream")

	// 发送文件
	c.File(document.FilePath)
}

// UpdateDocument 更新文档信息
func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	documentIDStr := c.Param("id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文档ID"})
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

	if err := h.documentService.SyncDocumentsFromGitLab(projectID, gitLabProjectID, accessToken.(string)); err != nil {
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

	// 获取当前用户ID
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	requesterID := gitlabUserID.(int64)

	var req struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Category    string   `json:"category"`
		Tags        []string `json:"tags"`
		Reason      string   `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 构建编辑数据
	editData := make(map[string]interface{})
	if req.Title != "" {
		editData["title"] = req.Title
	}
	if req.Description != "" {
		editData["description"] = req.Description
	}
	if req.Category != "" {
		editData["category"] = req.Category
	}
	if len(req.Tags) > 0 {
		editData["tags"] = req.Tags
	}

	editRequest, err := h.documentService.SubmitEditRequest(documentID, requesterID, editData, req.Reason)
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

	var reviewerID *uuid.UUID
	if reviewerIDStr := c.Query("reviewer_id"); reviewerIDStr != "" {
		if id, err := uuid.Parse(reviewerIDStr); err == nil {
			reviewerID = &id
		}
	}

	requests, total, err := h.documentService.GetEditRequests(status, reviewerID, pageSize, offset)
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
	// 获取用户ID
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

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
	category := c.PostForm("category")
	if category == "" {
		category = "other"
	}

	// 确定文件类型
	fileType := h.getFileType(header.Filename)

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
	// TODO: 实现基于GitLab用户角色的权限检查
	// 这里需要从用户会话中获取GitLab用户信息和角色
	// 暂时返回true，允许所有操作
	return true
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

	// 检查用户权限
	// 管理员/教师可以直接修改，研究员/学生需要提交审核请求
	canDirectEdit := h.CheckUserPermission(c, "admin") || h.CheckUserPermission(c, "teacher")

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
