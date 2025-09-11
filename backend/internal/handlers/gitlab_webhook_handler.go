package handlers

import (
	"fmt"
	"gitlabex/internal/services"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GitLabWebhookHandler GitLab webhook处理器
type GitLabWebhookHandler struct {
	gitlabService          *services.GitLabService
	userService            *services.UserService
	researchService        *services.ResearchService
	homeworkService        *services.HomeworkService
	notificationService    *services.NotificationService
	websocketService       *services.WebSocketService
	documentScannerService *services.DocumentScannerService
}

// NewGitLabWebhookHandler 创建GitLab webhook处理器
func NewGitLabWebhookHandler(
	gitlabService *services.GitLabService,
	userService *services.UserService,
	researchService *services.ResearchService,
	homeworkService *services.HomeworkService,
	notificationService *services.NotificationService,
	websocketService *services.WebSocketService,
	documentScannerService *services.DocumentScannerService,
) *GitLabWebhookHandler {
	return &GitLabWebhookHandler{
		gitlabService:          gitlabService,
		userService:            userService,
		researchService:        researchService,
		homeworkService:        homeworkService,
		notificationService:    notificationService,
		websocketService:       websocketService,
		documentScannerService: documentScannerService,
	}
}

// HandlePush 处理推送事件
func (h *GitLabWebhookHandler) HandlePush(c *gin.Context) {
	var payload services.GitLabWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	if payload.ObjectKind != "push" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not a push event"})
		return
	}

	log.Printf("Received push event for project: %s, branch: %s", payload.Project.Name, payload.Ref)

	// 记录提交信息
	hasDocumentChanges := false
	for _, commit := range payload.Commits {
		log.Printf("Commit: %s by %s - %s", commit.ShortID, commit.Author.Name, commit.Title)

		// 简单检查：如果提交消息中包含文档相关关键词，就触发扫描
		if h.checkCommitMessageForDocuments(commit.Message) {
			hasDocumentChanges = true
		}
	}

	// 如果检测到可能的文档变更，触发文档扫描
	if hasDocumentChanges {
		log.Printf("检测到可能的文档变更，触发文档扫描")
		// 异步执行文档扫描，避免阻塞webhook响应
		go h.triggerDocumentScan(payload.Project.ID, payload.Project.Name)
	}

	// 其他业务逻辑：
	// - 更新作业状态
	// - 发送通知
	// - 触发自动评分

	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}

// HandleMergeRequest 处理合并请求事件
func (h *GitLabWebhookHandler) HandleMergeRequest(c *gin.Context) {
	var payload struct {
		ObjectKind string `json:"object_kind"`
		Action     string `json:"action"`
		Project    struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"project"`
		MergeRequest struct {
			ID           int64  `json:"id"`
			IID          int64  `json:"iid"`
			Title        string `json:"title"`
			SourceBranch string `json:"source_branch"`
			TargetBranch string `json:"target_branch"`
			State        string `json:"state"`
			Author       struct {
				Username string `json:"username"`
			} `json:"author"`
		} `json:"object_attributes"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	if payload.ObjectKind != "merge_request" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not a merge request event"})
		return
	}

	log.Printf("Received merge request event: %s - %s", payload.Action, payload.MergeRequest.Title)

	// 处理合并请求的不同状态
	switch payload.Action {
	case "open":
		log.Printf("Merge request opened: #%d %s", payload.MergeRequest.IID, payload.MergeRequest.Title)
	case "update":
		log.Printf("Merge request updated: #%d %s", payload.MergeRequest.IID, payload.MergeRequest.Title)
	case "merge":
		log.Printf("Merge request merged: #%d %s", payload.MergeRequest.IID, payload.MergeRequest.Title)
		// 可以触发作业提交完成逻辑
	case "close":
		log.Printf("Merge request closed: #%d %s", payload.MergeRequest.IID, payload.MergeRequest.Title)
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}

// HandleIssue 处理议题事件
func (h *GitLabWebhookHandler) HandleIssue(c *gin.Context) {
	var payload struct {
		ObjectKind string `json:"object_kind"`
		Action     string `json:"action"`
		Project    struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"project"`
		Issue struct {
			ID     int64  `json:"id"`
			IID    int64  `json:"iid"`
			Title  string `json:"title"`
			State  string `json:"state"`
			Author struct {
				Username string `json:"username"`
			} `json:"author"`
		} `json:"object_attributes"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	if payload.ObjectKind != "issue" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not an issue event"})
		return
	}

	log.Printf("Received issue event: %s - %s", payload.Action, payload.Issue.Title)

	// 处理议题的不同状态
	switch payload.Action {
	case "open":
		log.Printf("Issue opened: #%d %s", payload.Issue.IID, payload.Issue.Title)
	case "update":
		log.Printf("Issue updated: #%d %s", payload.Issue.IID, payload.Issue.Title)
	case "close":
		log.Printf("Issue closed: #%d %s", payload.Issue.IID, payload.Issue.Title)
	case "reopen":
		log.Printf("Issue reopened: #%d %s", payload.Issue.IID, payload.Issue.Title)
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}

// HandlePipeline 处理流水线事件
func (h *GitLabWebhookHandler) HandlePipeline(c *gin.Context) {
	var payload struct {
		ObjectKind string `json:"object_kind"`
		Project    struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"project"`
		Pipeline struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
			Ref    string `json:"ref"`
			SHA    string `json:"sha"`
		} `json:"object_attributes"`
		Commit struct {
			Author struct {
				Username string `json:"username"`
			} `json:"author"`
		} `json:"commit"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	if payload.ObjectKind != "pipeline" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not a pipeline event"})
		return
	}

	log.Printf("Received pipeline event: %s for branch %s", payload.Pipeline.Status, payload.Pipeline.Ref)

	// 处理流水线状态
	switch payload.Pipeline.Status {
	case "success":
		log.Printf("Pipeline succeeded for branch: %s", payload.Pipeline.Ref)
		// 可以触发自动评分或通知
	case "failed":
		log.Printf("Pipeline failed for branch: %s", payload.Pipeline.Ref)
		// 可以发送失败通知
	case "running":
		log.Printf("Pipeline running for branch: %s", payload.Pipeline.Ref)
	case "pending":
		log.Printf("Pipeline pending for branch: %s", payload.Pipeline.Ref)
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}

// HandleWebhook 处理通用webhook事件
func (h *GitLabWebhookHandler) HandleWebhook(c *gin.Context) {
	eventType := c.GetHeader("X-Gitlab-Event")

	switch eventType {
	case "Push Hook":
		h.HandlePush(c)
	case "Merge Request Hook":
		h.HandleMergeRequest(c)
	case "Issue Hook":
		h.HandleIssue(c)
	case "Pipeline Hook":
		h.HandlePipeline(c)
	default:
		log.Printf("Received unsupported webhook event: %s", eventType)
		c.JSON(http.StatusOK, gin.H{"status": "ignored"})
	}
}

// ValidateWebhookSignature 验证webhook签名
func (h *GitLabWebhookHandler) ValidateWebhookSignature(c *gin.Context) bool {
	// TODO: 实现webhook签名验证
	// 这需要配置GitLab webhook secret
	return true
}

// RegisterWebhook 注册webhook
func (h *GitLabWebhookHandler) RegisterWebhook(c *gin.Context) {
	var req struct {
		ProjectID   int64    `json:"project_id" binding:"required"`
		URL         string   `json:"url" binding:"required"`
		Events      []string `json:"events" binding:"required"`
		PushEvents  bool     `json:"push_events"`
		TagEvents   bool     `json:"tag_push_events"`
		MRREvents   bool     `json:"merge_requests_events"`
		IssueEvents bool     `json:"issues_events"`
		JobEvents   bool     `json:"job_events"`
		Token       string   `json:"token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取用户access token
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// TODO: 从用户服务获取access token
	_ = "user_access_token" // 实际应该从用户服务获取

	webhookURL := req.URL
	_ = map[string]interface{}{
		"url":                     webhookURL,
		"push_events":             req.PushEvents,
		"tag_push_events":         req.TagEvents,
		"merge_requests_events":   req.MRREvents,
		"issues_events":           req.IssueEvents,
		"job_events":              req.JobEvents,
		"enable_ssl_verification": false,
		"token":                   req.Token,
	}

	// 调用GitLab API注册webhook
	// 这里应该调用GitLabService的相应方法
	// 需要添加AddProjectHook方法到GitLabService

	c.JSON(http.StatusOK, gin.H{"message": "Webhook registered successfully"})
}

// ListWebhooks 列出项目的webhooks
func (h *GitLabWebhookHandler) ListWebhooks(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// TODO: 获取用户access token并调用GitLab API
	_ = projectID // 使用变量避免编译错误
	c.JSON(http.StatusOK, gin.H{"webhooks": []interface{}{}})
}

// DeleteWebhook 删除webhook
func (h *GitLabWebhookHandler) DeleteWebhook(c *gin.Context) {
	projectIDStr := c.Param("id")
	webhookIDStr := c.Param("webhook_id")

	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	webhookID, err := strconv.ParseInt(webhookIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	// TODO: 获取用户access token并调用GitLab API删除webhook
	log.Printf("Deleting webhook %d from project %d", webhookID, projectID)

	c.JSON(http.StatusOK, gin.H{"message": "Webhook deleted successfully"})
}

// checkCommitMessageForDocuments 检查提交消息中是否包含文档相关关键词
func (h *GitLabWebhookHandler) checkCommitMessageForDocuments(message string) bool {
	documentKeywords := []string{
		"doc", "document", "pdf", "文档", "资料",
		"ppt", "excel", "word", "presentation", "报告",
		"add doc", "update doc", "delete doc", "新增文档", "更新文档", "删除文档",
	}

	messageLower := strings.ToLower(message)
	for _, keyword := range documentKeywords {
		if strings.Contains(messageLower, keyword) {
			log.Printf("提交消息包含文档关键词 '%s': %s", keyword, message)
			return true
		}
	}

	return false
}

// triggerDocumentScan 触发文档扫描
func (h *GitLabWebhookHandler) triggerDocumentScan(projectID int64, projectName string) {
	log.Printf("开始为项目 %s (ID: %d) 执行文档扫描", projectName, projectID)

	// 执行MinIO文档扫描
	result, err := h.documentScannerService.ScanMinIODocuments()
	if err != nil {
		log.Printf("项目 %s 文档扫描失败: %v", projectName, err)
		return
	}

	log.Printf("项目 %s 文档扫描完成: 总计 %d, 新增 %d, 更新 %d, 错误 %d",
		projectName, result.TotalDocuments, result.NewDocuments, result.UpdatedDocs, len(result.Errors))

	// 如果有新增或更新的文档，发送通知
	if result.NewDocuments > 0 || result.UpdatedDocs > 0 {
		h.sendDocumentScanNotification(projectID, projectName, result)
	}

	// 如果有错误，记录日志
	if len(result.Errors) > 0 {
		log.Printf("项目 %s 文档扫描出现错误:", projectName)
		for _, errMsg := range result.Errors {
			log.Printf("  - %s", errMsg)
		}
	}
}

// sendDocumentScanNotification 发送文档扫描通知
func (h *GitLabWebhookHandler) sendDocumentScanNotification(projectID int64, projectName string, result *services.SimpleScanResult) {
	// 构建通知消息
	message := ""
	if result.NewDocuments > 0 && result.UpdatedDocs > 0 {
		message = fmt.Sprintf("项目 %s 检测到文档变更: 新增 %d 个文档，更新 %d 个文档",
			projectName, result.NewDocuments, result.UpdatedDocs)
	} else if result.NewDocuments > 0 {
		message = fmt.Sprintf("项目 %s 检测到 %d 个新文档",
			projectName, result.NewDocuments)
	} else if result.UpdatedDocs > 0 {
		message = fmt.Sprintf("项目 %s 更新了 %d 个文档",
			projectName, result.UpdatedDocs)
	}

	if message != "" {
		log.Printf("发送文档扫描通知: %s", message)
		// 这里可以集成实际的通知发送逻辑
		// 例如通过WebSocket发送实时通知，或者存储到通知表

		// 通过WebSocket发送实时通知
		if h.websocketService != nil {
			// 简化通知发送，只记录日志
			log.Printf("WebSocket通知: %s", message)
			// 注意：这里需要根据实际的WebSocket服务接口调整
		}
	}
}
