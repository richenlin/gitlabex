package handlers

import (
	"fmt"
	"gitlabex/internal/dto"
	"gitlabex/internal/models"
	"gitlabex/internal/services"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TopicHandler 话题处理器
type TopicHandler struct {
	gitlabService   *services.GitLabService
	researchService *services.ResearchService
	topicService    *services.TopicService
}

// NewTopicHandler 创建话题处理器
func NewTopicHandler(gitlabService *services.GitLabService, researchService *services.ResearchService, topicService *services.TopicService) *TopicHandler {
	return &TopicHandler{
		gitlabService:   gitlabService,
		researchService: researchService,
		topicService:    topicService,
	}
}

// GetTopics 获取话题列表 (同时获取独立话题和GitLab Issues)
func (h *TopicHandler) GetTopics(c *gin.Context) {
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

	// 获取GitLab访问令牌，支持游客访问
	var tokenToUse string
	if accessToken, exists := c.Get("gitlab_access_token"); exists {
		fmt.Printf("DEBUG: Using user token\n")
		tokenToUse = accessToken.(string)
	} else {
		// 如果没有用户token，尝试使用系统token（用于游客访问）
		fmt.Printf("DEBUG: No user token, trying system token\n")
		systemToken := h.gitlabService.GetSystemToken()
		if systemToken == "" {
			fmt.Printf("DEBUG: No system token available\n")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少GitLab访问令牌"})
			return
		}
		fmt.Printf("DEBUG: Using system token\n")
		tokenToUse = systemToken
	}

	var allTopics []map[string]interface{}

	if projectIDStr != "" {
		// 获取特定项目的话题
		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
			return
		}

		project, err := h.researchService.GetResearchProjectByID(projectID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "课题不存在"})
			return
		}

		if project.GitLabProjectID != nil {
			topics, err := h.getProjectTopics(tokenToUse, project, page, limit, c)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "获取话题失败",
					"details": err.Error(),
				})
				return
			}
			allTopics = topics
		}
	} else {
		// 获取所有话题（包括独立话题和项目话题）

		// 1. 获取独立话题（从数据库）
		standaloneTopics, err := h.getStandaloneTopics(tokenToUse, page, limit, c)
		if err == nil {
			allTopics = append(allTopics, standaloneTopics...)
		}

		// 2. 获取所有项目的话题（从GitLab）
		projects, _, err := h.researchService.GetAllProjects(1000, 0, false, true) // 获取所有项目，包括私有项目
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "获取项目列表失败",
				"details": err.Error(),
			})
			return
		}

		// 使用并发方式获取所有项目的话题
		type projectTopicsResult struct {
			topics []map[string]interface{}
		}

		resultChan := make(chan projectTopicsResult, len(projects))
		semaphore := make(chan struct{}, 10) // 限制并发数为10

		for _, project := range projects {
			if project.GitLabProjectID != nil {
				go func(proj models.ResearchProject) {
					semaphore <- struct{}{}        // 获取信号量
					defer func() { <-semaphore }() // 释放信号量

					topics, err := h.getProjectTopics(tokenToUse, &proj, 1, limit, c)
					if err != nil {
						// 如果某个项目获取失败，返回空列表
						resultChan <- projectTopicsResult{topics: []map[string]interface{}{}}
						return
					}

					// 为每个话题添加项目信息
					for i := range topics {
						topics[i]["project_id"] = proj.ID.String()
						topics[i]["is_standalone"] = false
						topics[i]["project"] = map[string]interface{}{
							"id":   proj.ID.String(),
							"name": proj.Name,
						}
					}

					resultChan <- projectTopicsResult{topics: topics}
				}(project)
			}
		}

		// 收集所有结果
		projectCount := 0
		for _, project := range projects {
			if project.GitLabProjectID != nil {
				projectCount++
			}
		}

		for i := 0; i < projectCount; i++ {
			result := <-resultChan
			allTopics = append(allTopics, result.topics...)
		}
		close(resultChan)

		// 按创建时间排序（最新的在前面）
		sort.Slice(allTopics, func(i, j int) bool {
			timeI, _ := time.Parse(time.RFC3339, allTopics[i]["created_at"].(string))
			timeJ, _ := time.Parse(time.RFC3339, allTopics[j]["created_at"].(string))
			return timeI.After(timeJ)
		})

		// 应用分页
		start := (page - 1) * limit
		end := start + limit
		if start >= len(allTopics) {
			allTopics = []map[string]interface{}{}
		} else if end > len(allTopics) {
			allTopics = allTopics[start:]
		} else {
			allTopics = allTopics[start:end]
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"topics": allTopics,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": len(allTopics),
			"pages": (len(allTopics) + limit - 1) / limit,
		},
	})
}

// getStandaloneTopics 获取独立话题列表（从数据库）
func (h *TopicHandler) getStandaloneTopics(accessToken string, page, limit int, c *gin.Context) ([]map[string]interface{}, error) {
	// 从数据库获取独立话题
	offset := (page - 1) * limit
	topics, _, err := h.topicService.GetPublicTopics(limit, offset)
	if err != nil {
		return nil, err
	}

	// 转换为前端需要的格式
	result := make([]map[string]interface{}, 0, len(topics))

	// 检查是否需要获取用户的点赞/反对状态
	gitlabUserID, hasUserID := c.Get("gitlab_user_id")

	for _, topic := range topics {
		// 获取作者信息
		author, err := h.gitlabService.GetUserByID(accessToken, topic.AuthorID)
		if err != nil {
			// 如果获取作者信息失败，使用默认值
			author = &dto.GitLabAPIUser{
				ID:       topic.AuthorID,
				Username: "unknown",
				Name:     "Unknown User",
			}
		}

		// 获取评论数量
		comments, _ := h.topicService.GetCommentsByTopic(topic.ID)
		commentsCount := len(comments)

		// 检查用户是否点赞/反对
		userLiked := false
		userDisliked := false
		if hasUserID {
			userLiked, _ = h.topicService.HasLikedTopic(gitlabUserID.(int64), topic.ID)
			userDisliked, _ = h.topicService.HasDislikedTopic(gitlabUserID.(int64), topic.ID)
		}

		topicMap := map[string]interface{}{
			"id":             fmt.Sprintf("standalone-%s", topic.ID.String()),
			"topic_id":       topic.ID.String(),
			"title":          topic.Title,
			"content":        topic.Content,
			"status":         topic.Status,
			"labels":         topic.Tags,
			"created_at":     topic.CreatedAt.Format(time.RFC3339),
			"updated_at":     topic.UpdatedAt.Format(time.RFC3339),
			"like_count":     topic.LikeCount,
			"dislike_count":  topic.DislikeCount,
			"comments_count": commentsCount,
			"user_liked":     userLiked,
			"user_disliked":  userDisliked,
			"is_standalone":  true,
			"author": map[string]interface{}{
				"id":         author.ID,
				"username":   author.Username,
				"name":       author.Name,
				"avatar_url": author.Avatar,
			},
		}

		result = append(result, topicMap)
	}

	return result, nil
}

// getProjectTopics 获取单个项目的话题列表
func (h *TopicHandler) getProjectTopics(accessToken string, project *models.ResearchProject, page, limit int, c *gin.Context) ([]map[string]interface{}, error) {
	// 从GitLab获取Issues
	issues, err := h.gitlabService.GetProjectIssues(accessToken, *project.GitLabProjectID, page, limit)
	if err != nil {
		return nil, err
	}

	// 转换为前端需要的格式
	topics := make([]map[string]interface{}, len(issues))

	// 检查是否需要获取用户的表情反应
	gitlabUserID, hasUserID := c.Get("gitlab_user_id")

	// 如果需要获取用户表情反应，使用并发方式批量获取
	type emojiResult struct {
		index        int
		userLiked    bool
		userDisliked bool
	}

	if hasUserID {
		// 使用channel收集结果
		resultChan := make(chan emojiResult, len(issues))
		semaphore := make(chan struct{}, 5) // 限制并发数为5

		for i, issue := range issues {
			go func(idx int, issueIID int64) {
				semaphore <- struct{}{}        // 获取信号量
				defer func() { <-semaphore }() // 释放信号量

				userLiked, userDisliked := false, false
				emojis, err := h.gitlabService.GetIssueAwardEmojis(accessToken, *project.GitLabProjectID, issueIID)
				if err == nil {
					for _, emoji := range emojis {
						if emoji.User.ID == gitlabUserID.(int64) {
							if emoji.Name == "thumbsup" {
								userLiked = true
							} else if emoji.Name == "thumbsdown" {
								userDisliked = true
							}
						}
					}
				}

				resultChan <- emojiResult{
					index:        idx,
					userLiked:    userLiked,
					userDisliked: userDisliked,
				}
			}(i, issue.IID)
		}

		// 收集所有结果
		emojiResults := make(map[int]emojiResult)
		for i := 0; i < len(issues); i++ {
			result := <-resultChan
			emojiResults[result.index] = result
		}
		close(resultChan)

		// 构建话题列表
		for i, issue := range issues {
			result := emojiResults[i]
			topics[i] = map[string]interface{}{
				"id":             fmt.Sprintf("%d", issue.IID),
				"gitlab_id":      issue.ID,
				"gitlab_iid":     issue.IID,
				"title":          issue.Title,
				"content":        issue.Description,
				"status":         issue.State,
				"labels":         issue.Labels,
				"created_at":     issue.CreatedAt,
				"updated_at":     issue.UpdatedAt,
				"like_count":     issue.Upvotes,
				"dislike_count":  issue.Downvotes,
				"comments_count": issue.UserNotesCount,
				"user_liked":     result.userLiked,
				"user_disliked":  result.userDisliked,
				"author": map[string]interface{}{
					"id":         issue.Author.ID,
					"username":   issue.Author.Username,
					"name":       issue.Author.Name,
					"avatar_url": issue.Author.AvatarURL,
				},
			}
		}
	} else {
		// 游客访问，不需要获取用户表情反应
		for i, issue := range issues {
			topics[i] = map[string]interface{}{
				"id":             fmt.Sprintf("%d", issue.IID),
				"gitlab_id":      issue.ID,
				"gitlab_iid":     issue.IID,
				"title":          issue.Title,
				"content":        issue.Description,
				"status":         issue.State,
				"labels":         issue.Labels,
				"created_at":     issue.CreatedAt,
				"updated_at":     issue.UpdatedAt,
				"like_count":     issue.Upvotes,
				"dislike_count":  issue.Downvotes,
				"comments_count": issue.UserNotesCount,
				"user_liked":     false,
				"user_disliked":  false,
				"author": map[string]interface{}{
					"id":         issue.Author.ID,
					"username":   issue.Author.Username,
					"name":       issue.Author.Name,
					"avatar_url": issue.Author.AvatarURL,
				},
			}
		}
	}

	return topics, nil
}

// CreateTopic 创建话题 (支持创建独立话题或关联课题的话题)
func (h *TopicHandler) CreateTopic(c *gin.Context) {
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	var req struct {
		Title     string   `json:"title" binding:"required"`
		Content   string   `json:"content" binding:"required"`
		ProjectID string   `json:"project_id"` // 可选：不提供则创建独立话题
		Labels    []string `json:"labels"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 如果没有提供 project_id，创建独立话题
	if req.ProjectID == "" {
		h.createStandaloneTopic(c, req.Title, req.Content, req.Labels, gitlabUserID.(int64))
		return
	}

	// 有 project_id，创建关联课题的话题
	h.createProjectTopic(c, req.Title, req.Content, req.ProjectID, req.Labels)
}

// createStandaloneTopic 创建独立话题（存储在数据库中）
func (h *TopicHandler) createStandaloneTopic(c *gin.Context, title, content string, labels []string, authorID int64) {
	// 获取GitLab访问令牌（用于获取作者信息）
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少GitLab访问令牌"})
		return
	}

	// 获取作者信息
	author, err := h.gitlabService.GetUser(accessToken.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取用户信息失败",
			"details": err.Error(),
		})
		return
	}

	// 创建话题记录
	topic := &models.Topic{
		Title:    title,
		Content:  content,
		AuthorID: authorID,
		Status:   "active",
		Priority: "normal",
		Tags:     labels,
	}

	if err := h.topicService.CreateTopic(topic); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建话题失败",
			"details": err.Error(),
		})
		return
	}

	// 返回创建的话题信息
	c.JSON(http.StatusCreated, gin.H{
		"message": "话题创建成功",
		"topic": map[string]interface{}{
			"id":             fmt.Sprintf("standalone-%s", topic.ID.String()),
			"topic_id":       topic.ID.String(),
			"title":          topic.Title,
			"content":        topic.Content,
			"status":         topic.Status,
			"labels":         topic.Tags,
			"created_at":     topic.CreatedAt,
			"updated_at":     topic.UpdatedAt,
			"like_count":     topic.LikeCount,
			"dislike_count":  topic.DislikeCount,
			"comments_count": 0,
			"is_standalone":  true,
			"author": map[string]interface{}{
				"id":         author.ID,
				"username":   author.Username,
				"name":       author.Name,
				"avatar_url": author.Avatar,
			},
		},
	})
}

// createProjectTopic 创建关联课题的话题（存储在GitLab中）
func (h *TopicHandler) createProjectTopic(c *gin.Context, title, content, projectIDStr string, labels []string) {
	// 解析项目ID
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	// 获取项目信息
	project, err := h.researchService.GetResearchProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "课题不存在"})
		return
	}

	// 检查项目是否关联了GitLab
	if project.GitLabProjectID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该课题未关联GitLab项目"})
		return
	}

	// 获取GitLab访问令牌
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少GitLab访问令牌"})
		return
	}

	// 在GitLab中创建Issue
	issue, err := h.gitlabService.CreateProjectIssue(
		accessToken.(string),
		*project.GitLabProjectID,
		title,
		content,
		labels,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建话题失败",
			"details": err.Error(),
		})
		return
	}

	// 返回创建的话题信息
	c.JSON(http.StatusCreated, gin.H{
		"message": "话题创建成功",
		"topic": map[string]interface{}{
			"id":             fmt.Sprintf("%d", issue.IID),
			"gitlab_id":      issue.ID,
			"gitlab_iid":     issue.IID,
			"title":          issue.Title,
			"content":        issue.Description,
			"status":         issue.State,
			"labels":         issue.Labels,
			"created_at":     issue.CreatedAt,
			"updated_at":     issue.UpdatedAt,
			"like_count":     issue.Upvotes,
			"dislike_count":  issue.Downvotes,
			"comments_count": issue.UserNotesCount,
			"is_standalone":  false,
			"project_id":     projectIDStr,
		},
	})
}

// GetTopicByID 获取话题详情和回复列表（支持独立话题和项目话题）
func (h *TopicHandler) GetTopicByID(c *gin.Context) {
	topicIDStr := c.Param("id")

	// 检查是否是独立话题（格式：standalone-{uuid}）
	if len(topicIDStr) > 11 && topicIDStr[:11] == "standalone-" {
		h.getStandaloneTopicByID(c, topicIDStr[11:])
		return
	}

	// 否则按GitLab Issue处理
	h.getProjectTopicByID(c, topicIDStr)
}

// getStandaloneTopicByID 获取独立话题详情
func (h *TopicHandler) getStandaloneTopicByID(c *gin.Context, topicUUIDStr string) {
	// 解析UUID
	topicID, err := uuid.Parse(topicUUIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的话题ID"})
		return
	}

	// 获取GitLab访问令牌
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少GitLab访问令牌"})
		return
	}

	// 从数据库获取话题
	topic, err := h.topicService.GetTopicByID(topicID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "话题不存在"})
		return
	}

	// 获取作者信息
	author, err := h.gitlabService.GetUserByID(accessToken.(string), topic.AuthorID)
	if err != nil {
		author = &dto.GitLabAPIUser{
			ID:       topic.AuthorID,
			Username: "unknown",
			Name:     "Unknown User",
		}
	}

	// 获取评论列表
	dbComments, _ := h.topicService.GetCommentsByTopic(topic.ID)
	comments := make([]map[string]interface{}, 0, len(dbComments))

	for _, comment := range dbComments {
		// 获取评论作者信息
		commentAuthor, err := h.gitlabService.GetUserByID(accessToken.(string), comment.AuthorID)
		if err != nil {
			commentAuthor = &dto.GitLabAPIUser{
				ID:       comment.AuthorID,
				Username: "unknown",
				Name:     "Unknown User",
			}
		}

		comments = append(comments, map[string]interface{}{
			"id":         comment.ID.String(),
			"content":    comment.Content,
			"created_at": comment.CreatedAt.Format(time.RFC3339),
			"updated_at": comment.UpdatedAt.Format(time.RFC3339),
			"author": map[string]interface{}{
				"id":         commentAuthor.ID,
				"username":   commentAuthor.Username,
				"name":       commentAuthor.Name,
				"avatar_url": commentAuthor.Avatar,
			},
		})
	}

	// 检查用户是否点赞/反对
	userLiked := false
	userDisliked := false
	if gitlabUserID, exists := c.Get("gitlab_user_id"); exists {
		userLiked, _ = h.topicService.HasLikedTopic(gitlabUserID.(int64), topic.ID)
		userDisliked, _ = h.topicService.HasDislikedTopic(gitlabUserID.(int64), topic.ID)
	}

	// 返回话题详情
	c.JSON(http.StatusOK, gin.H{
		"topic": map[string]interface{}{
			"id":             fmt.Sprintf("standalone-%s", topic.ID.String()),
			"topic_id":       topic.ID.String(),
			"title":          topic.Title,
			"content":        topic.Content,
			"status":         topic.Status,
			"labels":         topic.Tags,
			"created_at":     topic.CreatedAt.Format(time.RFC3339),
			"updated_at":     topic.UpdatedAt.Format(time.RFC3339),
			"like_count":     topic.LikeCount,
			"dislike_count":  topic.DislikeCount,
			"comments_count": len(comments),
			"user_liked":     userLiked,
			"user_disliked":  userDisliked,
			"is_standalone":  true,
			"author": map[string]interface{}{
				"id":         author.ID,
				"username":   author.Username,
				"name":       author.Name,
				"avatar_url": author.Avatar,
			},
		},
		"comments": comments,
	})
}

// getProjectTopicByID 获取项目话题详情（GitLab Issue）
func (h *TopicHandler) getProjectTopicByID(c *gin.Context, topicIDStr string) {
	topicIID, err := strconv.ParseInt(topicIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的话题ID"})
		return
	}

	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少项目ID"})
		return
	}

	// 解析项目ID
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	// 获取项目信息
	project, err := h.researchService.GetResearchProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "课题不存在"})
		return
	}

	// 检查项目是否关联了GitLab
	if project.GitLabProjectID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该课题未关联GitLab项目"})
		return
	}

	// 获取GitLab访问令牌
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少GitLab访问令牌"})
		return
	}

	// 从GitLab获取Issue详情
	issue, err := h.gitlabService.GetIssue(accessToken.(string), *project.GitLabProjectID, topicIID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取话题详情失败",
			"details": err.Error(),
		})
		return
	}

	// 获取Issue的回复列表
	notes, err := h.gitlabService.GetIssueNotes(accessToken.(string), *project.GitLabProjectID, topicIID, 1, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取回复列表失败",
			"details": err.Error(),
		})
		return
	}

	// 获取当前用户对该Issue的表情反应
	userLiked, userDisliked := false, false
	if gitlabUserID, exists := c.Get("gitlab_user_id"); exists {
		emojis, _ := h.gitlabService.GetIssueAwardEmojis(accessToken.(string), *project.GitLabProjectID, issue.IID)
		for _, emoji := range emojis {
			if emoji.User.ID == gitlabUserID.(int64) {
				if emoji.Name == "thumbsup" {
					userLiked = true
				} else if emoji.Name == "thumbsdown" {
					userDisliked = true
				}
			}
		}
	}

	// 转换回复为前端需要的格式
	comments := make([]map[string]interface{}, len(notes))
	for i, note := range notes {
		comments[i] = map[string]interface{}{
			"id":         note.ID,
			"content":    note.Body,
			"created_at": note.CreatedAt,
			"updated_at": note.UpdatedAt,
			"author": map[string]interface{}{
				"id":         note.Author.ID,
				"username":   note.Author.Username,
				"name":       note.Author.Name,
				"avatar_url": note.Author.AvatarURL,
			},
		}
	}

	// 返回话题详情和回复列表
	c.JSON(http.StatusOK, gin.H{
		"topic": map[string]interface{}{
			"id":             fmt.Sprintf("%d", issue.IID),
			"gitlab_id":      issue.ID,
			"gitlab_iid":     issue.IID,
			"title":          issue.Title,
			"content":        issue.Description,
			"status":         issue.State,
			"labels":         issue.Labels,
			"created_at":     issue.CreatedAt,
			"updated_at":     issue.UpdatedAt,
			"like_count":     issue.Upvotes,
			"dislike_count":  issue.Downvotes,
			"comments_count": issue.UserNotesCount,
			"user_liked":     userLiked,
			"user_disliked":  userDisliked,
			"is_standalone":  false,
			"project_id":     projectIDStr,
			"author": map[string]interface{}{
				"id":         issue.Author.ID,
				"username":   issue.Author.Username,
				"name":       issue.Author.Name,
				"avatar_url": issue.Author.AvatarURL,
			},
		},
		"comments": comments,
	})
}

// CreateComment 创建话题回复（支持独立话题和项目话题）
func (h *TopicHandler) CreateComment(c *gin.Context) {
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	topicIDStr := c.Param("id")

	var req struct {
		Content   string `json:"content" binding:"required"`
		ProjectID string `json:"project_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查是否是独立话题
	if len(topicIDStr) > 11 && topicIDStr[:11] == "standalone-" {
		h.createStandaloneComment(c, topicIDStr[11:], req.Content, gitlabUserID.(int64))
		return
	}

	// 否则创建项目话题评论
	h.createProjectComment(c, topicIDStr, req.Content, req.ProjectID)
}

// createStandaloneComment 创建独立话题评论
func (h *TopicHandler) createStandaloneComment(c *gin.Context, topicUUIDStr, content string, authorID int64) {
	// 解析UUID
	topicID, err := uuid.Parse(topicUUIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的话题ID"})
		return
	}

	// 获取GitLab访问令牌（用于获取作者信息）
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少GitLab访问令牌"})
		return
	}

	// 验证话题是否存在
	_, err = h.topicService.GetTopicByID(topicID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "话题不存在"})
		return
	}

	// 创建评论
	comment := &models.Comment{
		Content:  content,
		TopicID:  topicID,
		AuthorID: authorID,
	}

	if err := h.topicService.CreateComment(comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建回复失败",
			"details": err.Error(),
		})
		return
	}

	// 获取作者信息
	author, err := h.gitlabService.GetUser(accessToken.(string))
	if err != nil {
		author = &dto.GitLabAPIUser{
			ID:       authorID,
			Username: "unknown",
			Name:     "Unknown User",
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "回复创建成功",
		"comment": map[string]interface{}{
			"id":         comment.ID.String(),
			"content":    comment.Content,
			"created_at": comment.CreatedAt.Format(time.RFC3339),
			"updated_at": comment.UpdatedAt.Format(time.RFC3339),
			"author": map[string]interface{}{
				"id":         author.ID,
				"username":   author.Username,
				"name":       author.Name,
				"avatar_url": author.Avatar,
			},
		},
	})
}

// createProjectComment 创建项目话题评论（GitLab Issue Note）
func (h *TopicHandler) createProjectComment(c *gin.Context, topicIDStr, content, projectIDStr string) {
	topicIID, err := strconv.ParseInt(topicIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的话题ID"})
		return
	}

	// 从查询参数或请求体获取项目ID
	if projectIDStr == "" {
		projectIDStr = c.Query("project_id")
	}

	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少项目ID"})
		return
	}

	// 解析项目ID
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	// 获取项目信息
	project, err := h.researchService.GetResearchProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "课题不存在"})
		return
	}

	if project.GitLabProjectID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该课题未关联GitLab项目"})
		return
	}

	// 获取GitLab访问令牌
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少GitLab访问令牌"})
		return
	}

	// 在GitLab中创建Issue Note
	note, err := h.gitlabService.CreateIssueNote(
		accessToken.(string),
		*project.GitLabProjectID,
		topicIID,
		content,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建回复失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "回复创建成功",
		"comment": map[string]interface{}{
			"id":         note.ID,
			"content":    note.Body,
			"created_at": note.CreatedAt,
			"updated_at": note.UpdatedAt,
			"author": map[string]interface{}{
				"id":         note.Author.ID,
				"username":   note.Author.Username,
				"name":       note.Author.Name,
				"avatar_url": note.Author.AvatarURL,
			},
		},
	})
}

// LikeTopic 点赞话题（支持独立话题和项目话题）
func (h *TopicHandler) LikeTopic(c *gin.Context) {
	topicIDStr := c.Param("id")
	if len(topicIDStr) > 11 && topicIDStr[:11] == "standalone-" {
		h.likeStandaloneTopic(c, topicIDStr[11:])
		return
	}
	h.toggleEmojiReaction(c, "thumbsup", "点赞")
}

// UnlikeTopic 取消点赞话题（支持独立话题和项目话题）
func (h *TopicHandler) UnlikeTopic(c *gin.Context) {
	topicIDStr := c.Param("id")
	if len(topicIDStr) > 11 && topicIDStr[:11] == "standalone-" {
		h.unlikeStandaloneTopic(c, topicIDStr[11:])
		return
	}
	h.removeEmojiReaction(c, "thumbsup", "取消点赞")
}

// DislikeTopic 反对话题（支持独立话题和项目话题）
func (h *TopicHandler) DislikeTopic(c *gin.Context) {
	topicIDStr := c.Param("id")
	if len(topicIDStr) > 11 && topicIDStr[:11] == "standalone-" {
		h.dislikeStandaloneTopic(c, topicIDStr[11:])
		return
	}
	h.toggleEmojiReaction(c, "thumbsdown", "反对")
}

// UndislikeTopic 取消反对话题（支持独立话题和项目话题）
func (h *TopicHandler) UndislikeTopic(c *gin.Context) {
	topicIDStr := c.Param("id")
	if len(topicIDStr) > 11 && topicIDStr[:11] == "standalone-" {
		h.undislikeStandaloneTopic(c, topicIDStr[11:])
		return
	}
	h.removeEmojiReaction(c, "thumbsdown", "取消反对")
}

// likeStandaloneTopic 点赞独立话题
func (h *TopicHandler) likeStandaloneTopic(c *gin.Context, topicUUIDStr string) {
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	topicID, err := uuid.Parse(topicUUIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的话题ID"})
		return
	}

	// 检查是否已经点赞
	hasLiked, _ := h.topicService.HasLikedTopic(gitlabUserID.(int64), topicID)
	if hasLiked {
		c.JSON(http.StatusBadRequest, gin.H{"error": "已经点赞过了"})
		return
	}

	// 如果已经反对，先取消反对
	hasDisliked, _ := h.topicService.HasDislikedTopic(gitlabUserID.(int64), topicID)
	if hasDisliked {
		h.topicService.UndislikeTopic(gitlabUserID.(int64), topicID)
	}

	// 添加点赞
	like := &models.TopicLike{
		TopicID: topicID,
		UserID:  gitlabUserID.(int64),
	}

	if err := h.topicService.LikeTopic(like); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "点赞失败",
			"details": err.Error(),
		})
		return
	}

	// 更新点赞数
	h.topicService.UpdateTopicLikesCount(topicID)

	c.JSON(http.StatusOK, gin.H{"message": "点赞成功"})
}

// unlikeStandaloneTopic 取消点赞独立话题
func (h *TopicHandler) unlikeStandaloneTopic(c *gin.Context, topicUUIDStr string) {
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	topicID, err := uuid.Parse(topicUUIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的话题ID"})
		return
	}

	// 检查是否已经点赞
	hasLiked, _ := h.topicService.HasLikedTopic(gitlabUserID.(int64), topicID)
	if !hasLiked {
		c.JSON(http.StatusBadRequest, gin.H{"error": "尚未点赞"})
		return
	}

	// 取消点赞
	if err := h.topicService.UnlikeTopic(gitlabUserID.(int64), topicID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "取消点赞失败",
			"details": err.Error(),
		})
		return
	}

	// 更新点赞数
	h.topicService.UpdateTopicLikesCount(topicID)

	c.JSON(http.StatusOK, gin.H{"message": "取消点赞成功"})
}

// dislikeStandaloneTopic 反对独立话题
func (h *TopicHandler) dislikeStandaloneTopic(c *gin.Context, topicUUIDStr string) {
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	topicID, err := uuid.Parse(topicUUIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的话题ID"})
		return
	}

	// 检查是否已经反对
	hasDisliked, _ := h.topicService.HasDislikedTopic(gitlabUserID.(int64), topicID)
	if hasDisliked {
		c.JSON(http.StatusBadRequest, gin.H{"error": "已经反对过了"})
		return
	}

	// 添加反对
	dislike := &models.TopicDislike{
		TopicID: topicID,
		UserID:  gitlabUserID.(int64),
	}

	if err := h.topicService.DislikeTopic(dislike); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "反对失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "反对成功"})
}

// undislikeStandaloneTopic 取消反对独立话题
func (h *TopicHandler) undislikeStandaloneTopic(c *gin.Context, topicUUIDStr string) {
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	topicID, err := uuid.Parse(topicUUIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的话题ID"})
		return
	}

	// 检查是否已经反对
	hasDisliked, _ := h.topicService.HasDislikedTopic(gitlabUserID.(int64), topicID)
	if !hasDisliked {
		c.JSON(http.StatusBadRequest, gin.H{"error": "尚未反对"})
		return
	}

	// 取消反对
	if err := h.topicService.UndislikeTopic(gitlabUserID.(int64), topicID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "取消反对失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "取消反对成功"})
}

// toggleEmojiReaction 切换表情反应的通用方法
func (h *TopicHandler) toggleEmojiReaction(c *gin.Context, emojiName, actionName string) {
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	topicIDStr := c.Param("id")
	topicIID, err := strconv.ParseInt(topicIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的话题ID"})
		return
	}

	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少项目ID"})
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	project, err := h.researchService.GetResearchProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "课题不存在"})
		return
	}

	if project.GitLabProjectID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该课题未关联GitLab项目"})
		return
	}

	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少GitLab访问令牌"})
		return
	}

	// 检查是否已经有该表情反应
	existingEmoji, err := h.gitlabService.FindUserAwardEmoji(
		accessToken.(string),
		*project.GitLabProjectID,
		topicIID,
		emojiName,
		gitlabUserID.(int64),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   fmt.Sprintf("检查%s状态失败", actionName),
			"details": err.Error(),
		})
		return
	}

	if existingEmoji != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("已经%s过了", actionName)})
		return
	}

	// 如果是点赞，先检查并移除反对；如果是反对，先检查并移除点赞
	var oppositeEmoji string
	if emojiName == "thumbsup" {
		oppositeEmoji = "thumbsdown"
	} else if emojiName == "thumbsdown" {
		oppositeEmoji = "thumbsup"
	}

	if oppositeEmoji != "" {
		// 查找并移除相反的表情反应
		existingOpposite, err := h.gitlabService.FindUserAwardEmoji(
			accessToken.(string),
			*project.GitLabProjectID,
			topicIID,
			oppositeEmoji,
			gitlabUserID.(int64),
		)
		if err == nil && existingOpposite != nil {
			// 移除相反的表情反应
			h.gitlabService.RemoveIssueAwardEmoji(
				accessToken.(string),
				*project.GitLabProjectID,
				topicIID,
				existingOpposite.ID,
			)
		}
	}

	// 添加表情反应
	_, err = h.gitlabService.AddIssueAwardEmoji(
		accessToken.(string),
		*project.GitLabProjectID,
		topicIID,
		emojiName,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   fmt.Sprintf("%s失败", actionName),
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%s成功", actionName)})
}

// removeEmojiReaction 移除表情反应的通用方法
func (h *TopicHandler) removeEmojiReaction(c *gin.Context, emojiName, actionName string) {
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	topicIDStr := c.Param("id")
	topicIID, err := strconv.ParseInt(topicIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的话题ID"})
		return
	}

	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少项目ID"})
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	project, err := h.researchService.GetResearchProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "课题不存在"})
		return
	}

	if project.GitLabProjectID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该课题未关联GitLab项目"})
		return
	}

	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少GitLab访问令牌"})
		return
	}

	// 查找用户的表情反应
	existingEmoji, err := h.gitlabService.FindUserAwardEmoji(
		accessToken.(string),
		*project.GitLabProjectID,
		topicIID,
		emojiName,
		gitlabUserID.(int64),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   fmt.Sprintf("查找%s状态失败", actionName),
			"details": err.Error(),
		})
		return
	}

	if existingEmoji == nil {
		// 根据actionName确定未执行的操作
		var baseAction string
		if actionName == "取消点赞" {
			baseAction = "点赞"
		} else if actionName == "取消反对" {
			baseAction = "反对"
		} else {
			baseAction = actionName
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("尚未%s", baseAction)})
		return
	}

	// 移除表情反应
	err = h.gitlabService.RemoveIssueAwardEmoji(
		accessToken.(string),
		*project.GitLabProjectID,
		topicIID,
		existingEmoji.ID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   fmt.Sprintf("%s失败", actionName),
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%s成功", actionName)})
}

// GetHotTopics 获取热门话题
func (h *TopicHandler) GetHotTopics(c *gin.Context) {
	// 获取限制参数
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	// 获取GitLab访问令牌（可选，支持游客访问）
	accessToken, _ := c.Get("gitlab_access_token")

	// 如果没有访问令牌，尝试使用系统配置的令牌（用于游客访问）
	var tokenToUse string
	if accessToken == nil {
		// 尝试从GitLab服务获取系统令牌
		systemToken := h.gitlabService.GetSystemToken()
		if systemToken == "" {
			// 如果没有系统令牌，返回空列表
			c.JSON(http.StatusOK, gin.H{
				"topics": []interface{}{},
				"count":  0,
			})
			return
		}
		tokenToUse = systemToken
	} else {
		tokenToUse = accessToken.(string)
	}

	// 获取所有项目的话题并按热度排序
	projects, _, err := h.researchService.GetAllProjects(100, 0, true, false) // 只获取公开项目
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取项目列表失败",
			"details": err.Error(),
		})
		return
	}

	var allTopics []map[string]interface{}

	// 使用并发方式获取所有公开项目的话题
	type projectTopicsResult struct {
		topics []map[string]interface{}
	}

	resultChan := make(chan projectTopicsResult, len(projects))
	semaphore := make(chan struct{}, 10) // 限制并发数为10

	for _, project := range projects {
		if project.GitLabProjectID != nil {
			go func(proj models.ResearchProject) {
				semaphore <- struct{}{}        // 获取信号量
				defer func() { <-semaphore }() // 释放信号量

				topics, err := h.getProjectTopics(tokenToUse, &proj, 1, 20, c)
				if err != nil {
					// 如果某个项目获取失败，返回空列表
					resultChan <- projectTopicsResult{topics: []map[string]interface{}{}}
					return
				}

				// 为每个话题添加项目信息
				for i := range topics {
					topics[i]["project_id"] = proj.ID.String()
					topics[i]["project"] = map[string]interface{}{
						"id":   proj.ID.String(),
						"name": proj.Name,
					}
				}

				resultChan <- projectTopicsResult{topics: topics}
			}(project)
		}
	}

	// 收集所有结果
	projectCount := 0
	for _, project := range projects {
		if project.GitLabProjectID != nil {
			projectCount++
		}
	}

	for i := 0; i < projectCount; i++ {
		result := <-resultChan
		allTopics = append(allTopics, result.topics...)
	}
	close(resultChan)

	// 按热度排序（点赞数 + 评论数）
	sort.Slice(allTopics, func(i, j int) bool {
		scoreI := getTopicHotScore(allTopics[i])
		scoreJ := getTopicHotScore(allTopics[j])
		if scoreI == scoreJ {
			// 如果热度相同，按创建时间排序（最新的在前）
			timeI, _ := time.Parse(time.RFC3339, allTopics[i]["created_at"].(string))
			timeJ, _ := time.Parse(time.RFC3339, allTopics[j]["created_at"].(string))
			return timeI.After(timeJ)
		}
		return scoreI > scoreJ
	})

	// 应用限制
	if len(allTopics) > limit {
		allTopics = allTopics[:limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"topics": allTopics,
		"count":  len(allTopics),
	})
}

// getTopicHotScore 计算话题热度分数
func getTopicHotScore(topic map[string]interface{}) int {
	likeCount, _ := topic["like_count"].(int)
	commentsCount, _ := topic["comments_count"].(int)

	// 热度分数 = 点赞数 * 2 + 评论数
	return likeCount*2 + commentsCount
}
