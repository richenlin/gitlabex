package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"gitlabex/internal/config"
	"gitlabex/internal/database"
	"gitlabex/internal/handlers"
	"gitlabex/internal/middleware"
	"gitlabex/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
)

func main() {
	// 初始化配置
	cfg := config.Load()

	// 设置Gin模式
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化数据库
	db, err := database.Initialize(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化Redis服务
	redisService, err := services.NewRedisService(cfg.RedisHost, cfg.RedisPort, cfg.RedisPassword)
	if err != nil {
		log.Printf("Warning: Failed to initialize Redis service: %v", err)
		log.Println("Continuing without Redis caching...")
		redisService = nil
	}

	// 初始化服务
	gitlabService := services.NewGitLabService(cfg)

	// 初始化MinIO服务
	minioService, err := services.NewMinIOService(
		cfg.MinIOEndpoint,
		cfg.MinIOAccessKey,
		cfg.MinIOSecretKey,
		cfg.MinIOUseSSL,
		cfg.MinIORegion,
	)
	if err != nil {
		log.Fatalf("Failed to initialize MinIO service: %v", err)
	}

	userService := services.NewUserService(db, gitlabService, cfg)
	researchService := services.NewResearchService(db, gitlabService)
	topicService := services.NewTopicService(db, gitlabService)
	documentService := services.NewDocumentService(db, gitlabService, minioService)
	homeworkService := services.NewHomeworkService(db, gitlabService)
	notificationService := services.NewNotificationService(db)

	// 初始化文档扫描服务
	documentScannerService := services.NewDocumentScannerService(minioService, documentService)

	// 初始化活动服务
	activityService := services.NewActivityService(db)

	// 初始化处理器
	gitlabHandler := handlers.NewGitLabHandler(gitlabService, userService)
	researchHandler := handlers.NewResearchHandler(researchService, userService, gitlabService)
	topicHandler := handlers.NewTopicHandler(gitlabService, researchService, topicService)
	syncHandler := handlers.NewSyncHandler(userService, gitlabService, cfg.JWTSecret)
	activityHandler := handlers.NewActivityHandler(activityService)
	permissionHandler := handlers.NewPermissionHandler(gitlabService, researchService, topicService)

	// 创建Gin路由器
	r := gin.Default()

	// 配置CORS - 从配置文件读取AllowedOrigins
	allowedOrigins := strings.Split(cfg.Security.CORSAllowedOrigins, ",")
	allowedMethods := strings.Split(cfg.Security.CORSAllowedMethods, ",")
	allowedHeaders := strings.Split(cfg.Security.CORSAllowedHeaders, ",")

	corsConfig := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   allowedMethods,
		AllowedHeaders:   allowedHeaders,
		AllowCredentials: cfg.Security.CORSAllowCredentials,
	})

	// 使用CORS中间件
	r.Use(func(c *gin.Context) {
		corsConfig.HandlerFunc(c.Writer, c.Request)
		c.Next()
	})

	// 使用XSS防护中间件
	// 跳过某些不需要XSS检查的路径
	skipPaths := []string{
		"/health",
		"/api/v1/auth/gitlab",
		"/api/v1/auth/gitlab/callback",
	}
	r.Use(middleware.XSSWhitelistMiddleware(skipPaths))

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "gitlabex-backend",
			"version": "1.0.0",
		})
	})

	// API路由组
	api := r.Group("/api/v1")

	// 认证相关路由
	auth := api.Group("/auth")
	authHandler := handlers.NewAuthHandler(userService, gitlabService, cfg, redisService)
	{
		auth.GET("/gitlab", authHandler.GitLabAuth)
		auth.GET("/gitlab/callback", authHandler.GitLabCallback)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/logout", middleware.RequireAuth(cfg), authHandler.Logout)
	}

	// 用户相关路由
	users := api.Group("/users")
	userHandler := handlers.NewUserHandler(userService)
	users.Use(middleware.RequireAuth(cfg))
	{
		// 用户资料 - 不需要缓存（不是高频接口）
		users.GET("/me", userHandler.GetCurrentUser)
		users.PUT("/me", userHandler.UpdateCurrentUser)

		// 个人资料页面统计 - 高频接口，添加缓存
		users.GET("/me/stats", middleware.CacheMiddleware(redisService, 15*time.Minute, "cache:user_stats"), userHandler.GetUserPersonalStats)

		// SSH密钥管理
		users.GET("/me/ssh-keys", userHandler.GetSSHKeys)
		users.POST("/me/ssh-keys", userHandler.AddSSHKey)
		users.DELETE("/me/ssh-keys/:id", userHandler.DeleteSSHKey)

		// 密码管理
		users.PUT("/me/password", userHandler.ChangePassword)

		// 通知获取 - 高频接口，添加缓存
		users.GET("/me/notifications", middleware.CacheMiddleware(redisService, 2*time.Minute, "cache:notifications"), userHandler.GetNotifications)
		users.POST("/me/notifications/:id/read", userHandler.MarkNotificationAsRead)
		users.POST("/me/notifications/read-all", userHandler.MarkAllNotificationsAsRead)

		// 用户列表
		users.GET("", userHandler.GetUsers)
		users.GET("/:id", userHandler.GetUserByID)
	}

	// 管理员用户管理路由
	adminUsers := api.Group("/admin/users")
	adminUserHandler := handlers.NewAdminUserHandler(userService)
	adminUsers.Use(middleware.RequireAuth(cfg))
	{
		adminUsers.GET("", adminUserHandler.GetUsers)                              // 获取用户列表
		adminUsers.POST("", adminUserHandler.CreateUser)                           // 创建用户
		adminUsers.GET("/:id", adminUserHandler.GetUserDetails)                    // 获取用户详情
		adminUsers.PUT("/:id", adminUserHandler.UpdateUser)                        // 更新用户信息
		adminUsers.DELETE("/:id", adminUserHandler.DeleteUser)                     // 删除用户
		adminUsers.PUT("/:id/roles", adminUserHandler.UpdateUserRoles)             // 更新用户角色
		adminUsers.GET("/:id/project-roles", adminUserHandler.GetUserProjectRoles) // 获取用户项目角色
		adminUsers.GET("/stats", adminUserHandler.GetUserStats)                    // 获取用户统计
	}

	// 权限检查相关路由
	permissions := api.Group("/permissions")
	permissions.Use(middleware.RequireAuth(cfg))
	{
		permissions.POST("/check", permissionHandler.CheckPermission)              // 通用权限检查
		permissions.GET("/projects/:id", permissionHandler.CheckProjectPermission) // 项目权限检查
		permissions.GET("/user", permissionHandler.GetUserPermissions)             // 获取用户权限列表
	}

	// 研究课题相关路由
	research := api.Group("/research-projects")
	{
		// 公开访问的路由（游客可访问）
		researchPublic := research.Group("")
		researchPublic.Use(middleware.OptionalAuth(cfg))
		{
			researchPublic.GET("", researchHandler.GetResearchProjects) // 课题列表 - 游客可访问（不是高频接口，移除缓存）
		}

		// 热门项目 - 高频接口，添加缓存
		research.GET("/hot", middleware.CacheMiddleware(redisService, 10*time.Minute, "cache:hot_projects"), researchHandler.GetHotProjects)

		// 需要认证的路由
		researchAuth := research.Group("")
		researchAuth.Use(middleware.RequireAuth(cfg))
		{
			researchAuth.GET("/:id", researchHandler.GetResearchProjectByID) // 课题详情（不是高频接口，移除缓存）
			researchAuth.POST("", researchHandler.CreateResearchProject)     // 创建课题不需要项目权限检查
			researchAuth.PUT("/:id", researchHandler.UpdateResearchProject)
			researchAuth.DELETE("/:id", researchHandler.DeleteResearchProject)
		}

		// 成员管理 - 需要认证
		researchAuth.GET("/:id/members", researchHandler.GetMembers)
		researchAuth.POST("/:id/members", researchHandler.AddMember)
		researchAuth.DELETE("/:id/members/:userId", researchHandler.RemoveMember)

		// 话题管理（基于GitLab Issues）- 需要认证
		researchAuth.GET("/:id/issues", researchHandler.GetIssues)
		researchAuth.POST("/:id/issues", researchHandler.CreateIssue)
		researchAuth.GET("/:id/issues/:issueId", researchHandler.GetIssue)
		researchAuth.GET("/:id/issues/:issueId/discussions", researchHandler.GetDiscussions)
		researchAuth.POST("/:id/issues/:issueId/discussions", researchHandler.CreateDiscussion)

		// GitLab IDE URL
		researchAuth.GET("/:id/ide-url", researchHandler.GetGitLabIDEURL)

		// 作业管理 - 需要认证
		researchAuth.GET("/:id/homework", researchHandler.GetHomework)
		researchAuth.POST("/:id/homework", researchHandler.CreateHomework)
	}

	// 话题相关路由
	topics := api.Group("/topics")
	{
		// 热门话题 - 高频接口，添加缓存
		topics.GET("/hot", middleware.CacheMiddleware(redisService, 10*time.Minute, "cache:hot_topics"), topicHandler.GetHotTopics)

		// 公开访问的路由（游客可访问）
		topicsPublic := topics.Group("")
		topicsPublic.Use(middleware.OptionalAuth(cfg))
		{
			topicsPublic.GET("", topicHandler.GetTopics) // 话题列表 - 游客可访问（不是高频接口，移除缓存）
		}

		// 需要认证的路由
		topicsAuth := topics.Group("")
		topicsAuth.Use(middleware.RequireAuth(cfg))
		{
			topicsAuth.GET("/:id", topicHandler.GetTopicByID) // 话题详情 - 需要登录访问
			topicsAuth.POST("", topicHandler.CreateTopic)
			topicsAuth.POST("/:id/comments", topicHandler.CreateComment)
			// 话题浏览点赞反对回复计数 - 高频接口，添加缓存
			topicsAuth.POST("/:id/like", middleware.CacheMiddleware(redisService, 5*time.Minute, "cache:topic_stats"), topicHandler.LikeTopic)
			topicsAuth.DELETE("/:id/like", middleware.CacheMiddleware(redisService, 5*time.Minute, "cache:topic_stats"), topicHandler.UnlikeTopic)
			topicsAuth.POST("/:id/dislike", middleware.CacheMiddleware(redisService, 5*time.Minute, "cache:topic_stats"), topicHandler.DislikeTopic)
			topicsAuth.DELETE("/:id/dislike", middleware.CacheMiddleware(redisService, 5*time.Minute, "cache:topic_stats"), topicHandler.UndislikeTopic)
		}
	}

	// 文档相关路由
	documents := api.Group("/documents")
	documentHandler := handlers.NewDocumentHandler(documentService)
	{
		// 公开访问的路由（不需要认证）
		documents.GET("", documentHandler.GetDocuments)                     // 文档列表（不是高频接口，移除缓存）
		documents.GET("/:id", documentHandler.GetDocumentByID)              // 文档详情（不是高频接口，移除缓存）
		documents.GET("/categories", documentHandler.GetDocumentCategories) // 文档分类（不是高频接口，移除缓存）
		documents.GET("/search", documentHandler.SearchDocuments)           // 文档搜索

		// 需要认证的路由
		documentsAuth := documents.Group("")
		documentsAuth.Use(middleware.RequireAuth(cfg))
		{
			documentsAuth.GET("/:id/download", documentHandler.DownloadDocument) // 文档下载 - 需要登录
			documentsAuth.PUT("/:id", documentHandler.UpdateDocument)
			documentsAuth.DELETE("/:id", documentHandler.DeleteDocument)
			documentsAuth.GET("/stats", documentHandler.GetDocumentStats)
			documentsAuth.POST("", documentHandler.UploadDocument)

			// 自动文档索引路由 - 需要认证
			documentsAuth.POST("/sync/:project_id", documentHandler.SyncDocuments)
			documentsAuth.POST("/scan/:project_id", documentHandler.ScanProjectDocuments)

			// 文档审核路由 - 需要认证
			documentsAuth.POST("/:id/edit-request", documentHandler.SubmitEditRequest)
			documentsAuth.GET("/edit-requests", documentHandler.GetEditRequests)
			documentsAuth.PUT("/edit-requests/:id/review", documentHandler.ReviewEditRequest)
			documentsAuth.GET("/:id/edit-history", documentHandler.GetDocumentEditHistory)

			// 新增的文档管理路由 - 需要认证
			documentsAuth.POST("/projects/:project_id/sync-to-minio", documentHandler.SyncProjectDocumentsToMinIO)
			documentsAuth.POST("/standalone", documentHandler.CreateStandaloneDocument)
			documentsAuth.GET("/categories/predefined", documentHandler.GetPredefinedCategories)
			documentsAuth.GET("/:id/download-url", documentHandler.GetDocumentDownloadURL)
			documentsAuth.PUT("/:id/with-permission-check", documentHandler.UpdateDocumentWithPermissionCheck)
		}
	}

	// 文档管理路由（只保留查看和管理功能，移除手动扫描）
	documentSync := api.Group("/document-management")
	documentSync.Use(middleware.RequireAuth(cfg))
	{
		// 获取统计信息（管理员查看）
		documentSync.GET("/stats", func(c *gin.Context) {
			stats, err := documentScannerService.GetSyncStats()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "获取统计信息失败",
					"details": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "文档管理统计信息",
				"data":    stats,
			})
		})

		// MinIO文档管理（只读）
		documentSync.GET("/minio/documents", func(c *gin.Context) {
			prefix := c.Query("prefix")
			recursive := c.Query("recursive") == "true"

			documents, err := minioService.ListDocuments(prefix, recursive)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "列出MinIO文档失败",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "MinIO文档列表",
				"data":    documents,
				"count":   len(documents),
			})
		})

		documentSync.GET("/minio/stats", func(c *gin.Context) {
			stats, err := minioService.GetDocumentStats()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "获取MinIO统计失败",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "MinIO存储统计",
				"data":    stats,
			})
		})
	}

	// 作业相关路由
	homework := api.Group("/homework")
	homeworkHandler := handlers.NewHomeworkHandler(homeworkService, userService)
	homework.Use(middleware.RequireAuth(cfg))
	{
		homework.GET("", homeworkHandler.GetHomeworkByProject)
		homework.GET("/:id", homeworkHandler.GetHomeworkByID)
		homework.POST("", homeworkHandler.CreateHomework)
		homework.PUT("/:id", homeworkHandler.UpdateHomework)
		homework.DELETE("/:id", homeworkHandler.DeleteHomework)
		homework.GET("/:id/submissions", homeworkHandler.GetSubmissions)
		homework.GET("/:id/my-submission", homeworkHandler.GetMySubmission)
		homework.POST("/:id/submissions", homeworkHandler.SubmitHomework)
		homework.PUT("/submissions/:id", homeworkHandler.GradeHomework)
		homework.GET("/:id/grade-distribution", homeworkHandler.GetGradeDistribution)
		homework.GET("/:id/export-grades", homeworkHandler.ExportGrades)
		homework.GET("/:id/details", homeworkHandler.GetAssignmentDetails)
		homework.POST("/bulk-create", homeworkHandler.BulkCreateHomework)
		homework.PUT("/bulk-update-due-date", homeworkHandler.BulkUpdateDueDate)
		homework.PUT("/:id/archive", homeworkHandler.ArchiveHomework)
		homework.PUT("/:id/restore", homeworkHandler.RestoreHomework)
		homework.GET("/user-submissions", homeworkHandler.GetUserSubmissions)
		homework.GET("/pending-reviews", homeworkHandler.GetPendingReviews)
		homework.GET("/student-progress", homeworkHandler.GetStudentProgress)
		homework.GET("/homework-stats", homeworkHandler.GetHomeworkStats)
		homework.GET("/generate-report", homeworkHandler.GenerateReport)

		// 作业分支管理路由
		homework.POST("/:id/create-branch", homeworkHandler.CreateStudentBranch)
		homework.GET("/:id/branches", homeworkHandler.GetHomeworkBranches)
		homework.POST("/:id/submit-to-branch", homeworkHandler.SubmitHomeworkToBranch)
		homework.GET("/:id/branch-info", homeworkHandler.GetStudentBranchInfo)

		// 获取作业提交的查看URL
		homework.GET("/submissions/:submissionId/view-url", homeworkHandler.GetSubmissionViewURL)
	}

	// 作业模板相关路由
	templates := api.Group("/assignment-templates")
	templates.Use(middleware.RequireAuth(cfg))
	{
		templates.GET("", homeworkHandler.GetAssignmentTemplates)
		templates.POST("", homeworkHandler.CreateAssignmentTemplate)
		templates.GET("/:id", homeworkHandler.GetHomeworkByID)
		templates.PUT("/:id", homeworkHandler.UpdateAssignmentTemplate)
		templates.DELETE("/:id", homeworkHandler.DeleteAssignmentTemplate)
		templates.POST("/:id/use", homeworkHandler.UseAssignmentTemplate)
	}

	// 通知相关路由
	notifications := api.Group("/notifications")
	notificationHandler := handlers.NewNotificationHandler(notificationService, userService)
	notifications.Use(middleware.RequireAuth(cfg))
	{
		notifications.GET("", notificationHandler.GetNotifications)
		notifications.GET("/unread-count", notificationHandler.GetUnreadCount)
		notifications.PUT("/:id/read", notificationHandler.MarkAsRead)
		notifications.PUT("/mark-all-read", notificationHandler.MarkAllAsRead)
	}

	// 公告相关路由
	announcements := api.Group("/announcements")
	announcements.Use(middleware.RequireAuth(cfg))
	{
		announcements.GET("", notificationHandler.GetAnnouncements)
		announcements.POST("", notificationHandler.CreateAnnouncement)
	}

	// 活动相关路由
	activities := api.Group("/activities")
	{
		// 公开访问的路由（游客可访问最近活动）
		// 最新活动 - 高频接口，保留缓存
		activities.GET("/recent", middleware.CacheMiddleware(redisService, 5*time.Minute, "cache:recent_activities"), activityHandler.GetRecentActivities)

		// 需要认证的路由
		activitiesAuth := activities.Group("")
		activitiesAuth.Use(middleware.RequireAuth(cfg))
		{
			activitiesAuth.GET("/users/:userID", activityHandler.GetUserActivities) // 用户活动（不是高频接口，移除缓存）
		}
	}

	// GitLab集成相关路由
	gitlab := api.Group("/gitlab")
	gitlab.Use(middleware.RequireAuth(cfg))
	{
		gitlab.GET("/config", gitlabHandler.GetGitLabConfig)
		gitlab.GET("/user", gitlabHandler.GetCurrentUser)
		gitlab.GET("/projects", gitlabHandler.GetProjects)
		gitlab.GET("/projects/:id", gitlabHandler.GetProject)
		gitlab.POST("/projects", gitlabHandler.CreateProject)
		gitlab.GET("/projects/:id/branches", gitlabHandler.GetBranches)
		gitlab.POST("/projects/:id/branches", gitlabHandler.CreateBranch)
		gitlab.GET("/projects/:id/files/*path", gitlabHandler.GetFileContent)
		gitlab.POST("/projects/:id/files/*path", gitlabHandler.CreateFile)
		gitlab.PUT("/projects/:id/files/*path", gitlabHandler.UpdateFile)
		gitlab.GET("/projects/:id/commits", gitlabHandler.GetCommits)
		gitlab.GET("/projects/:id/issues", gitlabHandler.GetIssues)
		gitlab.POST("/projects/:id/issues", gitlabHandler.CreateIssue)
		gitlab.GET("/projects/:id/merge-requests", gitlabHandler.GetMergeRequests)
		gitlab.GET("/projects/:id/tree", gitlabHandler.GetRepositoryTree)
		gitlab.GET("/projects/:id/search", gitlabHandler.SearchFiles)
		gitlab.GET("/projects/:id/validate", gitlabHandler.ValidateRepositoryAccess)
	}

	// 第三方系统同步API路由 (需要API密钥认证)
	{
		sync := api.Group("/sync")
		sync.Use(middleware.RequireAPIKey())

		// 用户同步接口
		sync.POST("/users", syncHandler.CreateUser)
		sync.POST("/users/batch", syncHandler.BatchCreateUsers)
		sync.PUT("/users/:id", syncHandler.UpdateUser)
		sync.GET("/users/:id", func(c *gin.Context) {
			// 可以添加获取用户信息的接口
			c.JSON(200, gin.H{"message": "获取用户信息接口 - 待实现"})
		})
	}

	// 启动服务器
	port := cfg.ServerPort
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on %s:%s", cfg.ServerHost, port)
	log.Printf("GitLab URL: %s", cfg.GitLabURL)

	address := fmt.Sprintf("%s:%s", cfg.ServerHost, port)
	if err := r.Run(address); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
