package handlers

import (
	"fmt"
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

// GetTopics 获取话题列表 (从GitLab Issues获取)
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
		// 获取所有项目的话题
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

// CreateTopic 创建话题 (在GitLab中创建Issue)
func (h *TopicHandler) CreateTopic(c *gin.Context) {
	_, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	var req struct {
		Title     string   `json:"title" binding:"required"`
		Content   string   `json:"content" binding:"required"`
		ProjectID string   `json:"project_id" binding:"required"`
		Labels    []string `json:"labels"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 解析项目ID
	projectID, err := uuid.Parse(req.ProjectID)
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
		req.Title,
		req.Content,
		req.Labels,
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
		},
	})
}

// GetTopicByID 获取话题详情和回复列表
func (h *TopicHandler) GetTopicByID(c *gin.Context) {
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

// CreateComment 创建话题回复 (创建GitLab Issue Note)
func (h *TopicHandler) CreateComment(c *gin.Context) {
	_, exists := c.Get("gitlab_user_id")
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

	var req struct {
		Content   string `json:"content" binding:"required"`
		ProjectID string `json:"project_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从查询参数或请求体获取项目ID
	projectIDStr := req.ProjectID
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
		req.Content,
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

// LikeTopic 点赞话题 (添加👍表情反应)
func (h *TopicHandler) LikeTopic(c *gin.Context) {
	h.toggleEmojiReaction(c, "thumbsup", "点赞")
}

// UnlikeTopic 取消点赞话题 (移除👍表情反应)
func (h *TopicHandler) UnlikeTopic(c *gin.Context) {
	h.removeEmojiReaction(c, "thumbsup", "取消点赞")
}

// DislikeTopic 反对话题 (添加👎表情反应)
func (h *TopicHandler) DislikeTopic(c *gin.Context) {
	h.toggleEmojiReaction(c, "thumbsdown", "反对")
}

// UndislikeTopic 取消反对话题 (移除👎表情反应)
func (h *TopicHandler) UndislikeTopic(c *gin.Context) {
	h.removeEmojiReaction(c, "thumbsdown", "取消反对")
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
