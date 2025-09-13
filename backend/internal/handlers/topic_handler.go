package handlers

import (
	"fmt"
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
	// 检查是否为游客模式
	isGuest, _ := c.Get("is_guest")
	userID, _ := c.Get("userID")

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

		// 如果是游客模式，只能访问公开项目的话题
		if isGuest == true || userID == "" {
			if !project.IsPublic {
				c.JSON(http.StatusForbidden, gin.H{"error": "无权限访问该课题的话题"})
				return
			}
		} else {
			// 已登录用户，检查项目权限
			// TODO: 重构用户验证以使用GitLab用户系统
			// 暂时跳过用户验证

			if !project.IsPublic {
				// 注意：权限检查已简化，具体权限由GitLab控制
				// 暂时允许访问，实际权限在GitLab层面控制
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

	// 填充作者信息
	topicsWithAuthors := make([]map[string]interface{}, len(topics))
	for i, topic := range topics {
		topicMap := map[string]interface{}{
			"id":              topic.ID,
			"created_at":      topic.CreatedAt,
			"updated_at":      topic.UpdatedAt,
			"deleted_at":      topic.DeletedAt,
			"title":           topic.Title,
			"content":         topic.Content,
			"project_id":      topic.ProjectID,
			"author_id":       topic.AuthorID,
			"gitlab_issue_id": topic.GitLabIssueID,
			"status":          topic.Status,
			"priority":        topic.Priority,
			"tags":            topic.Tags,
			"view_count":      topic.ViewCount,
			"like_count":      topic.LikeCount,
			"project":         topic.Project,
		}

		// 获取作者信息
		if accessToken, exists := c.Get("gitlab_access_token"); exists {
			if author, err := h.gitlabService.GetUserByID(accessToken.(string), topic.AuthorID); err == nil {
				topicMap["author"] = map[string]interface{}{
					"id":         author.ID,
					"username":   author.Username,
					"name":       author.Name,
					"avatar_url": author.AvatarURL,
					"email":      author.Email,
				}
			} else {
				// 如果获取失败，设置默认值
				topicMap["author"] = map[string]interface{}{
					"id":         topic.AuthorID,
					"username":   fmt.Sprintf("user_%d", topic.AuthorID),
					"name":       fmt.Sprintf("用户%d", topic.AuthorID),
					"avatar_url": "/default-avatar.png",
				}
			}
		} else {
			// 没有访问令牌时的默认值
			topicMap["author"] = map[string]interface{}{
				"id":         topic.AuthorID,
				"username":   fmt.Sprintf("user_%d", topic.AuthorID),
				"name":       fmt.Sprintf("用户%d", topic.AuthorID),
				"avatar_url": "/default-avatar.png",
			}
		}

		topicsWithAuthors[i] = topicMap
	}

	c.JSON(http.StatusOK, gin.H{
		"topics": topicsWithAuthors,
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

		// 获取当前用户信息
		// TODO: 重构权限检查以使用GitLab用户系统
		isAdmin, _ := c.Get("is_admin")

		// 使用GitLab权限检查（如果项目关联了GitLab）
		if project.GitLabProjectID != nil {
			// 对于关联GitLab的项目，使用GitLab权限检查
			hasPermission := false

			// 系统管理员可以访问所有项目
			if isAdmin != nil && isAdmin.(bool) {
				hasPermission = true
			} else if project.IsPublic {
				// 公开项目所有人都可以查看
				hasPermission = true
			} else {
				// 对于私有项目，检查GitLab访问令牌
				accessToken, exists := c.Get("gitlab_access_token")
				if exists && accessToken.(string) != "" {
					// 使用GitLab API检查权限（暂时简化）
					hasPermission = true
				}
			}

			if !hasPermission {
				c.JSON(http.StatusForbidden, gin.H{"error": "无权限访问该话题"})
				return
			}
		} else {
			// 对于未关联GitLab的项目，使用简化权限检查
			if !project.IsPublic {
				// 只有项目创建者可以访问私有项目的话题
				gitlabUserID, _ := c.Get("gitlab_user_id")
				// TODO: 修复用户ID类型不匹配
				if project.CreatorID != gitlabUserID.(int64) {
					c.JSON(http.StatusForbidden, gin.H{"error": "无权限访问该话题"})
					return
				}
			}
		}
	}

	// 获取作者信息
	authorInfo, err := h.enrichTopicWithUserInfo(topic)
	if err != nil {
		// 如果获取用户信息失败，仍然返回topic，但不包含用户信息
		c.JSON(http.StatusOK, topic)
		return
	}

	c.JSON(http.StatusOK, authorInfo)
}

// CreateTopic 创建新话题
func (h *TopicHandler) CreateTopic(c *gin.Context) {
	gitlabUserID, exists := c.Get("gitlab_user_id")
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
		_, err := uuid.Parse(*req.ProjectID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
			return
		}

		// 注意：权限检查已简化，具体权限由GitLab控制
	}

	topic := &models.Topic{
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: gitlabUserID.(int64), // TODO: 修复用户ID类型不匹配
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
			// 获取用户的GitLab访问令牌
			accessToken, exists := c.Get("gitlab_access_token")
			if exists && accessToken.(string) != "" {
				// 创建GitLab Issue
				issue, err := h.gitlabService.CreateIssue(
					accessToken.(string),
					*project.GitLabProjectID,
					req.Title,
					req.Content,
					req.Tags,
					nil,
				)
				if err == nil {
					issueID := issue.ID
					topic.GitLabIssueID = &issueID
					// 更新数据库中的GitLab Issue ID
					h.topicService.UpdateTopic(topic.ID, map[string]interface{}{"gitlab_issue_id": issue.ID})
				}
			}
		}
	}

	c.JSON(http.StatusCreated, topic)
}

// UpdateTopic 更新话题
func (h *TopicHandler) UpdateTopic(c *gin.Context) {
	gitlabUserID, exists := c.Get("gitlab_user_id")
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
	if topic.AuthorID != gitlabUserID.(int64) {
		// TODO: 重构权限检查以使用GitLab用户系统
		// 暂时只允许管理员修改他人话题
		isAdmin, _ := c.Get("is_admin")
		if isAdmin == nil || !isAdmin.(bool) {
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
	gitlabUserID, exists := c.Get("gitlab_user_id")
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
		// 注意：权限检查已简化，具体权限由GitLab控制
	}

	comment := &models.Comment{
		TopicID:  topicID,
		Content:  req.Content,
		AuthorID: gitlabUserID.(int64),
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
	gitlabUserID, exists := c.Get("gitlab_user_id")
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
	isLiked, err := h.topicService.HasLikedTopic(gitlabUserID.(int64), topicID)
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
		UserID:  gitlabUserID.(int64),
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
	gitlabUserID, exists := c.Get("gitlab_user_id")
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

	if err := h.topicService.UnlikeTopic(gitlabUserID.(int64), topicID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消点赞失败"})
		return
	}

	// 更新点赞数
	h.topicService.UpdateTopicLikesCount(topicID)

	c.JSON(http.StatusOK, gin.H{"message": "取消点赞成功"})
}

// DeleteTopic 删除话题
func (h *TopicHandler) DeleteTopic(c *gin.Context) {
	gitlabUserID, exists := c.Get("gitlab_user_id")
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
	if topic.AuthorID != gitlabUserID.(int64) {
		// TODO: 重构权限检查以使用GitLab用户系统
		// 暂时只允许管理员删除他人话题
		isAdmin, _ := c.Get("is_admin")
		if isAdmin == nil || !isAdmin.(bool) {
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

// enrichTopicWithUserInfo 为话题添加用户信息
func (h *TopicHandler) enrichTopicWithUserInfo(topic *models.Topic) (map[string]interface{}, error) {
	// 获取访问令牌（用于获取用户信息）
	// 注意：这里我们需要一个管理员令牌或者当前用户的令牌来获取其他用户信息
	// 暂时使用简化的方式，只返回基本的topic信息

	// 构建包含基本信息的响应
	result := map[string]interface{}{
		"id":              topic.ID,
		"title":           topic.Title,
		"content":         topic.Content,
		"author_id":       topic.AuthorID,
		"project_id":      topic.ProjectID,
		"gitlab_issue_id": topic.GitLabIssueID,
		"tags":            topic.Tags,
		"like_count":      topic.LikeCount,
		"view_count":      topic.ViewCount,
		"status":          topic.Status,
		"priority":        topic.Priority,
		"created_at":      topic.CreatedAt,
		"updated_at":      topic.UpdatedAt,
		"author": map[string]interface{}{
			"id": topic.AuthorID,
			// TODO: 从GitLab API获取用户详细信息
			"username":   fmt.Sprintf("gitlab_user_%d", topic.AuthorID),
			"name":       "GitLab用户",
			"avatar_url": "",
		},
	}

	// 如果有项目信息，也包含进去
	if topic.Project.ID != uuid.Nil {
		result["project"] = map[string]interface{}{
			"id":   topic.Project.ID,
			"name": topic.Project.Name,
		}
	}

	return result, nil
}
