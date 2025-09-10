package handlers

import (
	"gitlabex/internal/models"
	"gitlabex/internal/services"
	"net/http"
	"strconv"

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

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	document := &models.Document{
		ProjectID:   projectID,
		Title:       req.Title,
		FilePath:    req.FilePath,
		FileSize:    req.FileSize,
		Description: req.Description,
		Category:    req.Category,
		UploaderID:  userID.(uuid.UUID),
		Status:      models.DocumentStatusApproved,
	}

	if err := h.documentService.CreateDocument(document); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建文档失败"})
		return
	}

	c.JSON(http.StatusCreated, document)
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
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	projectIDStr := c.Param("project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	// 获取项目信息
	// 这里需要调用research service来获取GitLab项目ID
	// 暂时使用查询参数
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

	// 获取用户access token (这里需要从用户会话中获取)
	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "需要访问令牌"})
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
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

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

	// 获取用户access token
	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "需要访问令牌"})
		return
	}

	if err := h.documentService.ScanProjectDocuments(projectID, gitLabProjectID, accessToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "扫描文档失败"})
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
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	requesterID := userID.(uuid.UUID)

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
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	reviewerID := userID.(uuid.UUID)

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
