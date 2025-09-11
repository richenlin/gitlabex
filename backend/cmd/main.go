package main

import (
	"fmt"
	"log"
	"net/http"

	"gitlabex/internal/config"
	"gitlabex/internal/database"
	"gitlabex/internal/handlers"
	"gitlabex/internal/middleware"
	"gitlabex/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

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
		log.Fatalf("Failed to initialize Redis service: %v", err)
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

	userService := services.NewUserService(db, gitlabService)
	researchService := services.NewResearchService(db, gitlabService)
	topicService := services.NewTopicService(db, gitlabService)
	documentService := services.NewDocumentService(db, gitlabService, minioService)
	homeworkService := services.NewHomeworkService(db, gitlabService)
	notificationService := services.NewNotificationService(db)
	websocketService := services.NewWebSocketService()

	// 初始化文档扫描服务
	documentScannerService := services.NewDocumentScannerService(minioService, documentService)

	// 初始化活动服务
	activityService := services.NewActivityService(db)

	// 权限服务必须在其他handler之前初始化
	permissionService := services.NewPermissionService(db)
	permissionMiddleware := middleware.NewPermissionMiddleware(permissionService)

	// 初始化处理器
	gitlabHandler := handlers.NewGitLabHandler(gitlabService, userService)
	gitlabWebhookHandler := handlers.NewGitLabWebhookHandler(gitlabService, userService, researchService, homeworkService, notificationService, websocketService, documentScannerService)
	researchHandler := handlers.NewResearchHandler(researchService, userService, gitlabService)
	topicHandler := handlers.NewTopicHandler(topicService, userService, gitlabService, researchService)
	syncHandler := handlers.NewSyncHandler(userService, gitlabService, cfg.JWTSecret)
	websocketHandler := handlers.NewWebSocketHandler(websocketService)
	activityHandler := handlers.NewActivityHandler(activityService)
	// documentSyncHandler 已移除，文档扫描现在通过webhook自动触发

	// 创建Gin路由器
	r := gin.Default()

	// 配置CORS
	corsConfig := cors.New(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL, "http://localhost:3000", "http://127.0.0.1:3000", "http://0.0.0.0:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// 使用CORS中间件
	r.Use(func(c *gin.Context) {
		corsConfig.HandlerFunc(c.Writer, c.Request)
		c.Next()
	})

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
		users.GET("/me", userHandler.GetCurrentUser)
		users.PUT("/me", userHandler.UpdateCurrentUser)
		users.GET("", userHandler.GetUsers)
		users.GET("/:id", userHandler.GetUserByID)
	}

	// 研究课题相关路由
	research := api.Group("/research-projects")
	{
		// 公开访问的路由（游客可访问）
		researchPublic := research.Group("")
		researchPublic.Use(middleware.OptionalAuth(cfg))
		{
			researchPublic.GET("", researchHandler.GetResearchProjects) // 课题列表 - 游客可访问
		}

		// 需要认证的路由
		researchAuth := research.Group("")
		researchAuth.Use(middleware.RequireAuth(cfg))
		{
			researchAuth.GET("/:id", researchHandler.GetResearchProjectByID) // 课题详情
			researchAuth.POST("", permissionMiddleware.RequireProjectPermission(services.ProjectPermissionCreate), researchHandler.CreateResearchProject)
			researchAuth.PUT("/:id", permissionMiddleware.RequireProjectPermission(services.ProjectPermissionEdit), researchHandler.UpdateResearchProject)
			researchAuth.DELETE("/:id", permissionMiddleware.RequireProjectPermission(services.ProjectPermissionDelete), researchHandler.DeleteResearchProject)
		}

		// 成员管理 - 需要认证
		researchAuth.GET("/:id/members", permissionMiddleware.RequireProjectPermission(services.ProjectPermissionView), researchHandler.GetMembers)
		researchAuth.POST("/:id/members", permissionMiddleware.RequireProjectPermission(services.ProjectPermissionManage), researchHandler.AddMember)
		researchAuth.DELETE("/:id/members/:userId", permissionMiddleware.RequireProjectPermission(services.ProjectPermissionManage), researchHandler.RemoveMember)

		// 话题管理（基于GitLab Issues）- 需要认证
		researchAuth.GET("/:id/issues", permissionMiddleware.RequireProjectPermission(services.ProjectPermissionView), researchHandler.GetIssues)
		researchAuth.POST("/:id/issues", permissionMiddleware.RequireProjectPermission(services.ProjectPermissionCreate), researchHandler.CreateIssue)
		researchAuth.GET("/:id/issues/:issueId", permissionMiddleware.RequireProjectPermission(services.ProjectPermissionView), researchHandler.GetIssue)
		researchAuth.GET("/:id/issues/:issueId/discussions", permissionMiddleware.RequireProjectPermission(services.ProjectPermissionView), researchHandler.GetDiscussions)
		researchAuth.POST("/:id/issues/:issueId/discussions", permissionMiddleware.RequireProjectPermission(services.ProjectPermissionCreate), researchHandler.CreateDiscussion)

		// 作业管理 - 需要认证
		researchAuth.GET("/:id/homework", permissionMiddleware.RequireProjectPermission(services.ProjectPermissionView), researchHandler.GetHomework)
		researchAuth.POST("/:id/homework", permissionMiddleware.RequireProjectPermission(services.ProjectPermissionCreate), researchHandler.CreateHomework)
	}

	// 话题相关路由
	topics := api.Group("/topics")
	{
		// 公开访问的路由（游客可访问）
		topicsPublic := topics.Group("")
		topicsPublic.Use(middleware.OptionalAuth(cfg))
		{
			topicsPublic.GET("", topicHandler.GetTopics) // 话题列表 - 游客可访问
		}

		// 需要认证的路由
		topicsAuth := topics.Group("")
		topicsAuth.Use(middleware.RequireAuth(cfg))
		{
			topicsAuth.GET("/:id", topicHandler.GetTopicByID) // 话题详情
			topicsAuth.POST("", topicHandler.CreateTopic)
			topicsAuth.PUT("/:id", permissionMiddleware.RequireTopicPermission(services.ProjectPermissionEdit), topicHandler.UpdateTopic)
			topicsAuth.DELETE("/:id", permissionMiddleware.RequireTopicPermission(services.ProjectPermissionDelete), topicHandler.DeleteTopic)
			topicsAuth.POST("/:id/comments", permissionMiddleware.RequireTopicPermission(services.ProjectPermissionCreate), topicHandler.CreateComment)
			topicsAuth.POST("/:id/like", topicHandler.LikeTopic)
			topicsAuth.DELETE("/:id/like", topicHandler.UnlikeTopic)
		}
	}

	// 文档相关路由
	documents := api.Group("/documents")
	documentHandler := handlers.NewDocumentHandler(documentService)
	{
		// 公开访问的路由（不需要认证）
		documents.GET("", documentHandler.GetDocuments)                     // 文档列表
		documents.GET("/:id", documentHandler.GetDocumentByID)              // 文档详情
		documents.GET("/categories", documentHandler.GetDocumentCategories) // 文档分类
		documents.GET("/search", documentHandler.SearchDocuments)           // 文档搜索

		// 需要认证的路由
		documentsAuth := documents.Group("")
		documentsAuth.Use(middleware.RequireAuth(cfg))
		{
			documentsAuth.GET("/:id/download", documentHandler.DownloadDocument) // 文档下载 - 需要登录
			documentsAuth.PUT("/:id", permissionMiddleware.RequireDocumentPermission(services.ProjectPermissionEdit), documentHandler.UpdateDocument)
			documentsAuth.DELETE("/:id", permissionMiddleware.RequireDocumentPermission(services.ProjectPermissionDelete), documentHandler.DeleteDocument)
			documentsAuth.GET("/stats", documentHandler.GetDocumentStats)
			documentsAuth.POST("", permissionMiddleware.RequireProjectPermission(services.ProjectPermissionUpload), documentHandler.CreateDocument)

			// 文件上传路由
			documentsAuth.POST("/upload", permissionMiddleware.RequireProjectPermission(services.ProjectPermissionUpload), documentHandler.UploadDocument)

			// 自动文档索引路由 - 需要认证
			documentsAuth.POST("/sync/:project_id", permissionMiddleware.RequireProjectPermission(services.ProjectPermissionManage), documentHandler.SyncDocuments)
			documentsAuth.POST("/scan/:project_id", permissionMiddleware.RequireProjectPermission(services.ProjectPermissionManage), documentHandler.ScanProjectDocuments)

			// 文档审核路由 - 需要认证
			documentsAuth.POST("/:id/edit-request", documentHandler.SubmitEditRequest)
			documentsAuth.GET("/edit-requests", documentHandler.GetEditRequests)
			documentsAuth.PUT("/edit-requests/:id/review", documentHandler.ReviewEditRequest)
			documentsAuth.GET("/:id/edit-history", documentHandler.GetDocumentEditHistory)
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
		homework.GET("/:id", permissionMiddleware.RequireHomeworkPermission(services.ProjectPermissionView), homeworkHandler.GetHomeworkByID)
		homework.POST("", permissionMiddleware.RequireProjectPermission(services.ProjectPermissionCreate), homeworkHandler.CreateHomework)
		homework.PUT("/:id", permissionMiddleware.RequireHomeworkPermission(services.ProjectPermissionEdit), homeworkHandler.UpdateHomework)
		homework.DELETE("/:id", permissionMiddleware.RequireHomeworkPermission(services.ProjectPermissionDelete), homeworkHandler.DeleteHomework)
		homework.GET("/:id/submissions", permissionMiddleware.RequireHomeworkPermission(services.ProjectPermissionView), homeworkHandler.GetSubmissions)
		homework.GET("/:id/my-submission", homeworkHandler.GetMySubmission)
		homework.POST("/:id/submissions", permissionMiddleware.RequireHomeworkPermission(services.ProjectPermissionCreate), homeworkHandler.SubmitHomework)
		homework.PUT("/submissions/:id", permissionMiddleware.RequireHomeworkPermission(services.ProjectPermissionManage), homeworkHandler.GradeHomework)
		homework.GET("/:id/grade-distribution", permissionMiddleware.RequireHomeworkPermission(services.ProjectPermissionView), homeworkHandler.GetGradeDistribution)
		homework.GET("/:id/export-grades", permissionMiddleware.RequireHomeworkPermission(services.ProjectPermissionManage), homeworkHandler.ExportGrades)
		homework.GET("/:id/details", permissionMiddleware.RequireHomeworkPermission(services.ProjectPermissionView), homeworkHandler.GetAssignmentDetails)
		homework.POST("/bulk-create", permissionMiddleware.RequireProjectPermission(services.ProjectPermissionCreate), homeworkHandler.BulkCreateHomework)
		homework.PUT("/bulk-update-due-date", permissionMiddleware.RequireProjectPermission(services.ProjectPermissionManage), homeworkHandler.BulkUpdateDueDate)
		homework.PUT("/:id/archive", permissionMiddleware.RequireHomeworkPermission(services.ProjectPermissionManage), homeworkHandler.ArchiveHomework)
		homework.PUT("/:id/restore", permissionMiddleware.RequireHomeworkPermission(services.ProjectPermissionManage), homeworkHandler.RestoreHomework)
		homework.GET("/user-submissions", homeworkHandler.GetUserSubmissions)
		homework.GET("/pending-reviews", homeworkHandler.GetPendingReviews)
		homework.GET("/student-progress", homeworkHandler.GetStudentProgress)
		homework.GET("/homework-stats", homeworkHandler.GetHomeworkStats)
		homework.GET("/generate-report", permissionMiddleware.RequireProjectPermission(services.ProjectPermissionManage), homeworkHandler.GenerateReport)

		// 作业分支管理路由
		homework.POST("/:id/create-branch", homeworkHandler.CreateStudentBranch)
		homework.GET("/:id/branches", permissionMiddleware.RequireHomeworkPermission(services.ProjectPermissionView), homeworkHandler.GetHomeworkBranches)
		homework.POST("/:id/submit-to-branch", homeworkHandler.SubmitHomeworkToBranch)
		homework.GET("/:id/branch-info", homeworkHandler.GetStudentBranchInfo)
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
		activities.GET("/recent", activityHandler.GetRecentActivities) // 最近活动 - 游客可访问

		// 需要认证的路由
		activitiesAuth := activities.Group("")
		activitiesAuth.Use(middleware.RequireAuth(cfg))
		{
			activitiesAuth.GET("/users/:userID", activityHandler.GetUserActivities) // 用户活动
		}
	}

	// GitLab集成相关路由
	gitlab := api.Group("/gitlab")
	gitlab.Use(middleware.RequireAuth(cfg))
	{
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

	// GitLab webhook路由
	webhook := api.Group("/webhooks")
	{
		webhook.POST("/gitlab", gitlabWebhookHandler.HandleWebhook)
		webhook.POST("/gitlab/push", gitlabWebhookHandler.HandlePush)
		webhook.POST("/gitlab/merge-request", gitlabWebhookHandler.HandleMergeRequest)
		webhook.POST("/gitlab/issue", gitlabWebhookHandler.HandleIssue)
		webhook.POST("/gitlab/pipeline", gitlabWebhookHandler.HandlePipeline)
		gitlab.POST("/projects/:id/register-webhook", gitlabWebhookHandler.RegisterWebhook)
		gitlab.GET("/projects/:id/webhooks", gitlabWebhookHandler.ListWebhooks)
		gitlab.DELETE("/projects/:id/webhooks/:webhook_id", gitlabWebhookHandler.DeleteWebhook)
	}

	// WebSocket 实时通知路由
	ws := api.Group("/ws")
	ws.Use(middleware.RequireAuth(cfg))
	{
		ws.GET("/connect", websocketHandler.HandleWebSocket)
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
	log.Printf("Frontend URL: %s", cfg.FrontendURL)
	log.Printf("GitLab URL: %s", cfg.GitLabURL)

	address := fmt.Sprintf("%s:%s", cfg.ServerHost, port)
	if err := r.Run(address); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
