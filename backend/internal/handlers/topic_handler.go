package handlers

import (
	"gitlabex/internal/models"
	"gitlabex/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TopicHandler 话题处理器
type TopicHandler struct {
	topicService    *services.TopicService
	userService     *services.UserService
	gitlabService   *services.GitLabService
	researchService *services.ResearchService
}

// NewTopicHandler 创建话题处理器
func NewTopicHandler(topicService *services.TopicService, userService *services.UserService, gitlabService *services.GitLabService, researchService *services.ResearchService) *TopicHandler {
	return &TopicHandler{
		topicService:    topicService,
		userService:     userService,
		gitlabService:   gitlabService,
		researchService: researchService,
	}
}

// GetTopics 获取话题列表
func (h *TopicHandler) GetTopics(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	projectIDStr := c.Query("project_id")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var topics []models.Topic
	var total int64
	var err error

	if projectIDStr != "" {
		// 获取特定课题的话题
		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
			return
		}

		// 检查访问权限
		project, err := h.researchService.GetResearchProjectByID(projectID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "课题不存在"})
			return
		}

		_, err = h.userService.GetUserByID(userID.(uuid.UUID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
			return
		}

		if !project.IsPublic {
			isMember, err := h.researchService.IsProjectMember(projectID, userID.(uuid.UUID))
			if err != nil || !isMember {
				c.JSON(http.StatusForbidden, gin.H{"error": "无权限访问该课题的话题"})
				return
			}
		}

		topics, total, err = h.topicService.GetTopicsByProject(projectID, limit, offset)
	} else {
		// 获取所有公开话题
		topics, total, err = h.topicService.GetPublicTopics(limit, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取话题失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"topics": topics,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetTopicByID 根据ID获取话题详情
func (h *TopicHandler) GetTopicByID(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	topicIDStr := c.Param("id")
	topicID, err := uuid.Parse(topicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的话题ID"})
		return
	}

	topic, err := h.topicService.GetTopicByID(topicID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "话题不存在"})
		return
	}

	// 检查访问权限
	if topic.ProjectID != nil {
		project, err := h.researchService.GetResearchProjectByID(*topic.ProjectID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "关联课题不存在"})
			return
		}

		if !project.IsPublic {
			isMember, err := h.researchService.IsProjectMember(*topic.ProjectID, userID.(uuid.UUID))
			if err != nil || !isMember {
				c.JSON(http.StatusForbidden, gin.H{"error": "无权限访问该话题"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, topic)
}

// CreateTopic 创建新话题
func (h *TopicHandler) CreateTopic(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	var req struct {
		Title       string   `json:"title" binding:"required"`
		Content     string   `json:"content" binding:"required"`
		ProjectID   *string  `json:"project_id"`
		Tags        []string `json:"tags"`
		GitLabIssue *int64   `json:"gitlab_issue_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查项目权限
	if req.ProjectID != nil {
		projectID, err := uuid.Parse(*req.ProjectID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
			return
		}

		isMember, err := h.researchService.IsProjectMember(projectID, userID.(uuid.UUID))
		if err != nil || !isMember {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限在该课题下创建话题"})
			return
		}
	}

	topic := &models.Topic{
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: userID.(uuid.UUID),
		Tags:     req.Tags,
	}

	if req.ProjectID != nil {
		projectID, _ := uuid.Parse(*req.ProjectID)
		topic.ProjectID = &projectID
	}

	if err := h.topicService.CreateTopic(topic); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建话题失败"})
		return
	}

	// 如果是关联GitLab Issue，同步创建
	if req.ProjectID != nil {
		projectID, _ := uuid.Parse(*req.ProjectID)
		project, err := h.researchService.GetResearchProjectByID(projectID)
		if err == nil && project.GitLabProjectID != nil {
			// TODO: 实现GitLab Issue创建
			// issue, err := h.gitlabService.CreateIssue(*project.GitLabProjectID, req.Title, req.Content, req.Tags, nil)
			// if err == nil {
			//     issueID := issue.ID
			//     topic.GitLabIssueID = &issueID
			//     h.topicService.UpdateTopic(topic.ID, map[string]interface{}{"gitlab_issue_id": issue.ID})
			// }
		}
	}

	c.JSON(http.StatusCreated, topic)
}

// UpdateTopic 更新话题
func (h *TopicHandler) UpdateTopic(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	topicIDStr := c.Param("id")
	topicID, err := uuid.Parse(topicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的话题ID"})
		return
	}

	var req struct {
		Title   *string  `json:"title"`
		Content *string  `json:"content"`
		Tags    []string `json:"tags"`
		Status  *string  `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取原话题
	topic, err := h.topicService.GetTopicByID(topicID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "话题不存在"})
		return
	}

	// 检查权限 - 只有作者和管理员可以更新
	if topic.AuthorID != userID.(uuid.UUID) {
		currentUser, err := h.userService.GetUserByID(userID.(uuid.UUID))
		if err != nil || currentUser.EduRole < models.EduRoleTeacher {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限修改该话题"})
			return
		}
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Tags != nil {
		updates["tags"] = req.Tags
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := h.topicService.UpdateTopic(topicID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新话题失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "话题更新成功"})
}

// CreateComment 创建话题评论
func (h *TopicHandler) CreateComment(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	topicIDStr := c.Param("id")
	topicID, err := uuid.Parse(topicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的话题ID"})
		return
	}

	var req struct {
		Content string     `json:"content" binding:"required"`
		ReplyTo *uuid.UUID `json:"reply_to"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取话题
	topic, err := h.topicService.GetTopicByID(topicID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "话题不存在"})
		return
	}

	// 检查访问权限
	if topic.ProjectID != nil {
		isMember, err := h.researchService.IsProjectMember(*topic.ProjectID, userID.(uuid.UUID))
		if err != nil || !isMember {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限评论该话题"})
			return
		}
	}

	comment := &models.Comment{
		TopicID:  topicID,
		Content:  req.Content,
		AuthorID: userID.(uuid.UUID),
		ParentID: req.ReplyTo,
	}

	if err := h.topicService.CreateComment(comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建评论失败"})
		return
	}

	// 同步到GitLab讨论
	if topic.GitLabIssueID != nil && topic.ProjectID != nil {
		project, err := h.researchService.GetResearchProjectByID(*topic.ProjectID)
		if err == nil && project.GitLabProjectID != nil {
			// TODO: 实现CreateIssueDiscussion方法
			// h.gitlabService.CreateIssueDiscussion(*project.GitLabProjectID, *topic.GitLabIssueID, req.Content)
		}
	}

	c.JSON(http.StatusCreated, comment)
}

// LikeTopic 点赞话题
func (h *TopicHandler) LikeTopic(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	topicIDStr := c.Param("id")
	topicID, err := uuid.Parse(topicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的话题ID"})
		return
	}

	// 检查是否已经点赞
	isLiked, err := h.topicService.HasLikedTopic(userID.(uuid.UUID), topicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查点赞状态失败"})
		return
	}

	if isLiked {
		c.JSON(http.StatusBadRequest, gin.H{"error": "已经点过赞了"})
		return
	}

	like := &models.TopicLike{
		TopicID: topicID,
		UserID:  userID.(uuid.UUID),
	}

	if err := h.topicService.LikeTopic(like); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "点赞失败"})
		return
	}

	// 更新点赞数
	h.topicService.UpdateTopicLikesCount(topicID)

	c.JSON(http.StatusOK, gin.H{"message": "点赞成功"})
}

// UnlikeTopic 取消点赞
func (h *TopicHandler) UnlikeTopic(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	topicIDStr := c.Param("id")
	topicID, err := uuid.Parse(topicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的话题ID"})
		return
	}

	if err := h.topicService.UnlikeTopic(userID.(uuid.UUID), topicID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消点赞失败"})
		return
	}

	// 更新点赞数
	h.topicService.UpdateTopicLikesCount(topicID)

	c.JSON(http.StatusOK, gin.H{"message": "取消点赞成功"})
}

// DeleteTopic 删除话题
func (h *TopicHandler) DeleteTopic(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	topicIDStr := c.Param("id")
	topicID, err := uuid.Parse(topicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的话题ID"})
		return
	}

	// 获取话题信息
	topic, err := h.topicService.GetTopicByID(topicID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "话题不存在"})
		return
	}

	// 检查权限 - 只有作者和管理员可以删除
	if topic.AuthorID != userID.(uuid.UUID) {
		currentUser, err := h.userService.GetUserByID(userID.(uuid.UUID))
		if err != nil || currentUser.EduRole < models.EduRoleTeacher {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限删除该话题"})
			return
		}
	}

	if err := h.topicService.DeleteTopic(topicID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除话题失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "话题删除成功"})
}
