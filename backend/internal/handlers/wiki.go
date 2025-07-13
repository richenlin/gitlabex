package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"gitlabex/internal/services"

	"github.com/gin-gonic/gin"
)

// WikiHandler Wiki文档处理器
type WikiHandler struct {
	gitlabService     *services.GitLabService
	onlyOfficeService *services.OnlyOfficeService
	documentService   *services.DocumentService
}

// NewWikiHandler 创建Wiki处理器
func NewWikiHandler(gitlabService *services.GitLabService, onlyOfficeService *services.OnlyOfficeService, documentService *services.DocumentService) *WikiHandler {
	return &WikiHandler{
		gitlabService:     gitlabService,
		onlyOfficeService: onlyOfficeService,
		documentService:   documentService,
	}
}

// RegisterRoutes 注册路由
func (h *WikiHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Wiki相关路由
	wiki := router.Group("/wiki")
	{
		wiki.GET("/projects", h.GetProjects)
		wiki.GET("/projects/:id", h.GetWikiPages)
		wiki.POST("/projects/:id", h.CreateWikiPage)
		wiki.GET("/projects/:id/:slug/attachments", h.GetWikiAttachments)
		wiki.POST("/projects/:id/:slug/attachments", h.UploadWikiAttachment)
	}

	// 文档编辑相关路由
	documents := router.Group("/documents")
	{
		documents.POST("/:id/edit", h.StartEditSession)
		documents.GET("/:id/download", h.DownloadDocument)
	}
}

// GetProjects 获取项目列表
func (h *WikiHandler) GetProjects(c *gin.Context) {
	// 获取用户ID (暂时使用固定值，后续从JWT获取)
	userID := 1

	// 获取用户有权限的项目列表
	projects, err := h.gitlabService.GetUserProjects(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取项目列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取项目列表成功",
		"data":    projects,
	})
}

// GetWikiPages 获取Wiki页面列表
func (h *WikiHandler) GetWikiPages(c *gin.Context) {
	// 获取项目ID
	projectIDStr := c.Param("id")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "项目ID格式错误",
		})
		return
	}

	// 获取Wiki页面列表
	wikiPages, err := h.gitlabService.GetWikiPages(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取Wiki页面列表失败: " + err.Error(),
		})
		return
	}

	// 增强Wiki页面数据，添加可编辑附件数量
	enhancedPages := make([]map[string]interface{}, 0)
	for _, page := range wikiPages {
		attachments, _ := h.documentService.GetWikiAttachments(projectID, page.Slug)

		enhancedPage := map[string]interface{}{
			"id":                  page.Slug,
			"title":               page.Title,
			"slug":                page.Slug,
			"content":             page.Content,
			"updated_at":          nil,
			"editableAttachments": len(attachments),
		}
		enhancedPages = append(enhancedPages, enhancedPage)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取Wiki页面列表成功",
		"data":    enhancedPages,
	})
}

// CreateWikiPage 创建Wiki页面
func (h *WikiHandler) CreateWikiPage(c *gin.Context) {
	// 获取项目ID
	projectIDStr := c.Param("id")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "项目ID格式错误",
		})
		return
	}

	// 获取用户ID
	userID := 1

	// 获取表单数据
	title := c.PostForm("title")
	content := c.PostForm("content")

	if title == "" || content == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "标题和内容不能为空",
		})
		return
	}

	// 创建Wiki页面
	wikiPage, err := h.gitlabService.CreateWikiPage(projectID, title, content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建Wiki页面失败: " + err.Error(),
		})
		return
	}

	// 处理附件上传
	form, _ := c.MultipartForm()
	attachments := form.File["attachments"]

	for _, attachment := range attachments {
		// 打开文件
		file, err := attachment.Open()
		if err != nil {
			continue
		}
		defer file.Close()

		// 读取文件内容
		fileContent := make([]byte, attachment.Size)
		file.Read(fileContent)

		// 上传到GitLab
		uploadResult, err := h.gitlabService.UploadFile(projectID, attachment.Filename, fileContent)
		if err != nil {
			continue
		}

		// 创建文档附件记录
		docAttachment, err := h.documentService.CreateWikiAttachment(userID, projectID, wikiPage.Slug, attachment.Filename, uploadResult.URL, fileContent)
		if err != nil {
			continue
		}

		// 更新Wiki页面，添加附件链接
		attachmentLink := fmt.Sprintf("\n\n## 文档附件\n\n- [%s](%s) ([在线编辑](/documents/editor/%d))",
			attachment.Filename, uploadResult.URL, docAttachment.ID)

		updatedContent := content + attachmentLink
		h.gitlabService.UpdateWikiPage(projectID, wikiPage.Slug, wikiPage.Title, updatedContent)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Wiki页面创建成功",
		"data":    wikiPage,
	})
}

// GetWikiAttachments 获取Wiki附件列表
func (h *WikiHandler) GetWikiAttachments(c *gin.Context) {
	// 获取项目ID
	projectIDStr := c.Param("id")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "项目ID格式错误",
		})
		return
	}

	// 获取Wiki slug
	wikiSlug := c.Param("slug")

	// 获取用户ID
	userID := 1

	// 获取Wiki附件列表
	attachments, err := h.documentService.GetWikiAttachments(projectID, wikiSlug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取Wiki附件列表失败: " + err.Error(),
		})
		return
	}

	// 检查用户权限
	canEdit, err := h.gitlabService.CheckWikiEditPermission(userID, projectID)
	if err != nil {
		canEdit = false
	}

	// 增强附件数据
	enhancedAttachments := make([]map[string]interface{}, 0)
	for _, attachment := range attachments {
		enhancedAttachment := map[string]interface{}{
			"id":             attachment.ID,
			"file_name":      attachment.FileName,
			"file_type":      attachment.FileType,
			"file_url":       attachment.FileURL,
			"last_edited_at": attachment.LastEditedAt,
			"last_edited_by": attachment.LastEditedBy,
			"can_edit":       canEdit && attachment.IsEditableType(),
		}
		enhancedAttachments = append(enhancedAttachments, enhancedAttachment)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取Wiki附件列表成功",
		"data":    enhancedAttachments,
	})
}

// UploadWikiAttachment 上传Wiki附件
func (h *WikiHandler) UploadWikiAttachment(c *gin.Context) {
	// 获取项目ID
	projectIDStr := c.Param("id")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "项目ID格式错误",
		})
		return
	}

	// 获取Wiki slug
	wikiSlug := c.Param("slug")

	// 获取用户ID
	userID := 1

	// 处理文件上传
	form, _ := c.MultipartForm()
	attachments := form.File["attachments"]

	if len(attachments) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请选择要上传的文件",
		})
		return
	}

	uploadedAttachments := make([]map[string]interface{}, 0)

	for _, attachment := range attachments {
		// 打开文件
		file, err := attachment.Open()
		if err != nil {
			continue
		}
		defer file.Close()

		// 读取文件内容
		fileContent := make([]byte, attachment.Size)
		file.Read(fileContent)

		// 上传到GitLab
		uploadResult, err := h.gitlabService.UploadFile(projectID, attachment.Filename, fileContent)
		if err != nil {
			continue
		}

		// 创建文档附件记录
		docAttachment, err := h.documentService.CreateWikiAttachment(userID, projectID, wikiSlug, attachment.Filename, uploadResult.URL, fileContent)
		if err != nil {
			continue
		}

		uploadedAttachments = append(uploadedAttachments, map[string]interface{}{
			"id":        docAttachment.ID,
			"file_name": docAttachment.FileName,
			"file_type": docAttachment.FileType,
			"file_url":  docAttachment.FileURL,
		})
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "附件上传成功",
		"data":    uploadedAttachments,
	})
}

// StartEditSession 启动编辑会话
func (h *WikiHandler) StartEditSession(c *gin.Context) {
	// 获取文档ID
	docIDStr := c.Param("id")
	docID, err := strconv.Atoi(docIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "文档ID格式错误",
		})
		return
	}

	// 获取用户ID
	userID := 1

	// 启动OnlyOffice编辑会话
	config, err := h.onlyOfficeService.GetDocumentConfig(docID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "启动编辑会话失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, config)
}

// DownloadDocument 下载文档
func (h *WikiHandler) DownloadDocument(c *gin.Context) {
	// 获取文档ID
	docIDStr := c.Param("id")
	docID, err := strconv.Atoi(docIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "文档ID格式错误",
		})
		return
	}

	// 获取用户ID
	userID := 1

	// 获取文档内容
	content, fileType, err := h.onlyOfficeService.GetDocumentContent(docID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取文档内容失败: " + err.Error(),
		})
		return
	}

	// 设置响应头
	switch fileType {
	case "docx":
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	case "xlsx":
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	case "pptx":
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	default:
		c.Header("Content-Type", "application/octet-stream")
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=document_%d.%s", docID, fileType))
	c.Data(http.StatusOK, c.GetHeader("Content-Type"), content)
}
