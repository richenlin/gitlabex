package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"gitlabex/internal/services"

	"github.com/gin-gonic/gin"
)

// DocumentHandler 文档处理器
type DocumentHandler struct {
	onlyOfficeService *services.OnlyOfficeService
}

// NewDocumentHandler 创建文档处理器
func NewDocumentHandler(onlyOfficeService *services.OnlyOfficeService) *DocumentHandler {
	return &DocumentHandler{
		onlyOfficeService: onlyOfficeService,
	}
}

// UploadDocument 上传文档
func (h *DocumentHandler) UploadDocument(c *gin.Context) {
	// 获取用户ID (暂时使用固定值，后续从JWT获取)
	userID := 1

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to get file from request",
		})
		return
	}
	defer file.Close()

	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read file content",
		})
		return
	}

	// 获取编辑模式
	editMode := c.DefaultPostForm("mode", "edit")

	// 创建文档会话
	doc, err := h.onlyOfficeService.CreateDocumentSession(userID, header.Filename, content, editMode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to create document session: %v", err),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"document_id": doc.ID,
		"message":     "Document uploaded successfully",
	})
}

// GetDocumentEditor 获取文档编辑器配置
func (h *DocumentHandler) GetDocumentEditor(c *gin.Context) {
	// 获取文档ID
	docIDStr := c.Param("id")
	docID, err := strconv.Atoi(docIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid document ID",
		})
		return
	}

	// 获取用户ID (暂时使用固定值)
	userID := 1

	// 获取文档配置
	config, err := h.onlyOfficeService.GetDocumentConfig(docID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Document not found: %v", err),
		})
		return
	}

	// 将配置序列化为JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to serialize config: %v", err),
		})
		return
	}

	// 返回HTML页面
	c.HTML(http.StatusOK, "simple.html", gin.H{
		"configJSON": string(configJSON),
		"docID":      docID,
	})
}

// GetDocumentConfig 获取文档配置JSON
func (h *DocumentHandler) GetDocumentConfig(c *gin.Context) {
	// 获取文档ID
	docIDStr := c.Param("id")
	docID, err := strconv.Atoi(docIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid document ID",
		})
		return
	}

	// 获取用户ID (暂时使用固定值)
	userID := 1

	// 获取文档配置
	config, err := h.onlyOfficeService.GetDocumentConfig(docID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Document not found: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, config)
}

// GetDocumentContent 获取文档内容
func (h *DocumentHandler) GetDocumentContent(c *gin.Context) {
	// 获取文档ID
	docIDStr := c.Param("id")
	docID, err := strconv.Atoi(docIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid document ID",
		})
		return
	}

	// 获取用户ID (暂时使用固定值)
	userID := 1

	// 获取文档内容
	content, fileType, err := h.onlyOfficeService.GetDocumentContent(docID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Document not found: %v", err),
		})
		return
	}

	// 设置合适的Content-Type
	switch fileType {
	case "docx":
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	case "xlsx":
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	case "pptx":
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	case "pdf":
		c.Header("Content-Type", "application/pdf")
	default:
		c.Header("Content-Type", "text/plain")
	}

	c.Data(http.StatusOK, c.GetHeader("Content-Type"), content)
}

// HandleCallback 处理OnlyOffice回调
func (h *DocumentHandler) HandleCallback(c *gin.Context) {
	// 获取文档ID
	docIDStr := c.Param("id")
	docID, err := strconv.Atoi(docIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid document ID",
		})
		return
	}

	// 解析回调数据
	var callbackData services.CallbackData
	if err := c.ShouldBindJSON(&callbackData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid callback data: %v", err),
		})
		return
	}

	// 处理回调
	if err := h.onlyOfficeService.HandleCallback(docID, &callbackData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to handle callback: %v", err),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"error": 0,
	})
}

// TestUpload 测试上传接口
func (h *DocumentHandler) TestUpload(c *gin.Context) {
	// 创建一个简单的测试文档
	testContent := `<!DOCTYPE html>
<html>
<head>
    <title>测试文档</title>
</head>
<body>
    <h1>GitLabEx 测试文档</h1>
    <p>这是一个由GitLabEx系统创建的测试文档。</p>
    <p>创建时间：` + fmt.Sprintf("%v", "2024-01-01") + `</p>
</body>
</html>`

	// 创建文档会话
	doc, err := h.onlyOfficeService.CreateDocumentSession(1, "test.docx", []byte(testContent), "edit")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to create test document: %v", err),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"document_id": doc.ID,
		"message":     "Test document created successfully",
		"editor_url":  fmt.Sprintf("/api/documents/%d/editor", doc.ID),
	})
}
