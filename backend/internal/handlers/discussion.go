package handlers

import (
	"net/http"
	"strconv"

	"gitlabex/internal/models"
	"gitlabex/internal/services"

	"github.com/gin-gonic/gin"
)

// DiscussionHandler 话题讨论处理器
type DiscussionHandler struct {
	discussionService *services.DiscussionService
	userService       *services.UserService
}

// NewDiscussionHandler 创建话题讨论处理器
func NewDiscussionHandler(discussionService *services.DiscussionService, userService *services.UserService) *DiscussionHandler {
	return &DiscussionHandler{
		discussionService: discussionService,
		userService:       userService,
	}
}

// CreateDiscussion 创建话题
func (h *DiscussionHandler) CreateDiscussion(c *gin.Context) {
	// 获取用户信息
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权访问",
		})
		return
	}
	currentUser := user.(*models.User)

	// 解析请求数据
	var req models.DiscussionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 创建话题
	discussion, err := h.discussionService.CreateDiscussion(&req, currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建话题失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "话题创建成功",
		"discussion": discussion,
	})
}

// GetDiscussionList 获取话题列表
func (h *DiscussionHandler) GetDiscussionList(c *gin.Context) {
	// 获取用户信息
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权访问",
		})
		return
	}
	currentUser := user.(*models.User)

	// 解析查询参数
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "缺少项目ID参数",
		})
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "项目ID格式错误",
		})
		return
	}

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 过滤参数
	category := c.Query("category")
	status := c.Query("status")

	// 获取话题列表
	result, err := h.discussionService.GetDiscussionList(uint(projectID), page, pageSize, category, status, currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取话题列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetDiscussionDetail 获取话题详情
func (h *DiscussionHandler) GetDiscussionDetail(c *gin.Context) {
	// 获取用户信息
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权访问",
		})
		return
	}
	currentUser := user.(*models.User)

	// 解析ID参数
	idStr := c.Param("id")
	discussionID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "话题ID格式错误",
		})
		return
	}

	// 获取话题详情
	result, err := h.discussionService.GetDiscussionDetail(uint(discussionID), currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取话题详情失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateDiscussion 更新话题
func (h *DiscussionHandler) UpdateDiscussion(c *gin.Context) {
	// 获取用户信息
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权访问",
		})
		return
	}
	currentUser := user.(*models.User)

	// 解析ID参数
	idStr := c.Param("id")
	discussionID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "话题ID格式错误",
		})
		return
	}

	// 解析请求数据
	var req models.DiscussionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 更新话题
	discussion, err := h.discussionService.UpdateDiscussion(uint(discussionID), &req, currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "更新话题失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "话题更新成功",
		"discussion": discussion,
	})
}

// DeleteDiscussion 删除话题
func (h *DiscussionHandler) DeleteDiscussion(c *gin.Context) {
	// 获取用户信息
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权访问",
		})
		return
	}
	currentUser := user.(*models.User)

	// 解析ID参数
	idStr := c.Param("id")
	discussionID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "话题ID格式错误",
		})
		return
	}

	// 删除话题
	err = h.discussionService.DeleteDiscussion(uint(discussionID), currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "删除话题失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "话题删除成功",
	})
}

// CreateReply 创建回复
func (h *DiscussionHandler) CreateReply(c *gin.Context) {
	// 获取用户信息
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权访问",
		})
		return
	}
	currentUser := user.(*models.User)

	// 解析ID参数
	idStr := c.Param("id")
	discussionID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "话题ID格式错误",
		})
		return
	}

	// 解析请求数据
	var req models.DiscussionReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 创建回复
	reply, err := h.discussionService.CreateReply(uint(discussionID), &req, currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建回复失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "回复创建成功",
		"reply":   reply,
	})
}

// LikeDiscussion 点赞话题
func (h *DiscussionHandler) LikeDiscussion(c *gin.Context) {
	// 获取用户信息
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权访问",
		})
		return
	}
	currentUser := user.(*models.User)

	// 解析ID参数
	idStr := c.Param("id")
	discussionID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "话题ID格式错误",
		})
		return
	}

	// 点赞话题
	err = h.discussionService.LikeDiscussion(uint(discussionID), currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "点赞失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "点赞成功",
	})
}

// UnlikeDiscussion 取消点赞
func (h *DiscussionHandler) UnlikeDiscussion(c *gin.Context) {
	// 获取用户信息
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权访问",
		})
		return
	}
	currentUser := user.(*models.User)

	// 解析ID参数
	idStr := c.Param("id")
	discussionID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "话题ID格式错误",
		})
		return
	}

	// 取消点赞
	err = h.discussionService.UnlikeDiscussion(uint(discussionID), currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "取消点赞失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "取消点赞成功",
	})
}

// PinDiscussion 置顶话题
func (h *DiscussionHandler) PinDiscussion(c *gin.Context) {
	// 获取用户信息
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权访问",
		})
		return
	}
	currentUser := user.(*models.User)

	// 解析ID参数
	idStr := c.Param("id")
	discussionID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "话题ID格式错误",
		})
		return
	}

	// 置顶话题
	err = h.discussionService.PinDiscussion(uint(discussionID), currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "置顶失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "置顶成功",
	})
}

// GetCategories 获取话题分类
func (h *DiscussionHandler) GetCategories(c *gin.Context) {
	categories := h.discussionService.GetCategories()
	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
	})
}

// SyncFromGitLab 从GitLab同步话题
func (h *DiscussionHandler) SyncFromGitLab(c *gin.Context) {
	// 获取用户信息
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权访问",
		})
		return
	}
	currentUser := user.(*models.User)

	// 只有管理员和教师可以同步
	if currentUser.Role != 1 && currentUser.Role != 2 {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "没有权限执行此操作",
		})
		return
	}

	// 解析项目ID
	projectIDStr := c.Param("project_id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "项目ID格式错误",
		})
		return
	}

	// 同步话题
	err = h.discussionService.SyncFromGitLab(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "同步失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "同步成功",
	})
}

// RegisterRoutes 注册路由
func (h *DiscussionHandler) RegisterRoutes(r *gin.RouterGroup) {
	discussions := r.Group("/discussions")
	{
		discussions.POST("", h.CreateDiscussion)                // 创建话题
		discussions.GET("", h.GetDiscussionList)                // 获取话题列表
		discussions.GET("/:id", h.GetDiscussionDetail)          // 获取话题详情
		discussions.PUT("/:id", h.UpdateDiscussion)             // 更新话题
		discussions.DELETE("/:id", h.DeleteDiscussion)          // 删除话题
		discussions.POST("/:id/replies", h.CreateReply)         // 创建回复
		discussions.POST("/:id/like", h.LikeDiscussion)         // 点赞话题
		discussions.DELETE("/:id/like", h.UnlikeDiscussion)     // 取消点赞
		discussions.POST("/:id/pin", h.PinDiscussion)           // 置顶话题
		discussions.GET("/categories", h.GetCategories)         // 获取分类
		discussions.POST("/sync/:project_id", h.SyncFromGitLab) // 同步GitLab
	}
}
