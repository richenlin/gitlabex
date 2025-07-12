package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"gitlabex/internal/config"
	"gitlabex/internal/handlers"
	"gitlabex/internal/models"
	"gitlabex/internal/services"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库连接
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 初始化Redis连接
	rdb := initRedis(cfg)

	// 初始化服务
	gitlabSvc, err := services.NewGitLabService(cfg, rdb, db)
	if err != nil {
		log.Printf("Warning: Failed to initialize GitLab service: %v", err)
		// 继续运行，但GitLab功能将不可用
	}

	authService := services.NewAuthService(db, cfg)
	userService := services.NewUserService(db, gitlabSvc)
	onlyOfficeService := services.NewOnlyOfficeService(cfg, db)
	documentService := services.NewDocumentService(db, gitlabSvc)
	educationServiceSimplified := services.NewEducationServiceSimplified(gitlabSvc, db)

	// 初始化处理器
	userHandler := handlers.NewUserHandler(userService)
	documentHandler := handlers.NewDocumentHandler(onlyOfficeService)
	wikiHandler := handlers.NewWikiHandler(gitlabSvc, onlyOfficeService, documentService)
	educationTestHandler := handlers.NewEducationTestHandler(educationServiceSimplified)
	educationHandler := handlers.NewEducationHandler(educationServiceSimplified, userService)
	learningProgressHandler := handlers.NewLearningProgressHandler(gitlabSvc)
	notificationHandler := handlers.NewNotificationHandler(gitlabSvc)
	educationReportHandler := handlers.NewEducationReportHandler(gitlabSvc)

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化路由
	router := setupRoutes(authService, userHandler, documentHandler, wikiHandler, educationTestHandler, educationHandler, learningProgressHandler, notificationHandler, educationReportHandler)

	// 启动服务器
	addr := cfg.GetServerAddr()
	log.Printf("Server starting on %s", addr)

	srv := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// initDatabase 初始化数据库连接
func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.GetDatabaseDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 设置连接池参数
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 自动迁移数据库表
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database connected and migrated successfully")
	return db, nil
}

// initRedis 初始化Redis连接
func initRedis(cfg *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 测试连接
	ctx := rdb.Context()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
	} else {
		log.Println("Redis connected successfully")
	}

	return rdb
}

// autoMigrate 自动迁移数据库表
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.DocumentAttachment{},
		&models.Class{},
		&models.ClassMember{},
		&models.Project{},
		&models.ProjectMember{},
		&models.Assignment{},
		&models.AssignmentSubmission{},
		&models.Review{},
		&models.Notification{},
	)
}

// setupRoutes 设置路由
func setupRoutes(authService *services.AuthService, userHandler *handlers.UserHandler, documentHandler *handlers.DocumentHandler, wikiHandler *handlers.WikiHandler, educationTestHandler *handlers.EducationTestHandler, educationHandler *handlers.EducationHandler, learningProgressHandler *handlers.LearningProgressHandler, notificationHandler *handlers.NotificationHandler, educationReportHandler *handlers.EducationReportHandler) *gin.Engine {
	router := gin.New()

	// 加载HTML模板
	router.LoadHTMLGlob("templates/*")

	// 中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// API路由组
	api := router.Group("/api")
	{
		// 健康检查
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "ok",
				"service":   "gitlabex-backend",
				"version":   "1.0.0",
				"timestamp": time.Now().Unix(),
			})
		})

		// 认证相关路由（无需认证）
		auth := api.Group("/auth")
		{
			auth.GET("/gitlab", func(c *gin.Context) {
				state := "random-state-string" // 实际应用中应该生成随机字符串
				url := authService.GetGitLabOAuthURL(state)
				c.JSON(200, gin.H{"url": url})
			})
			auth.GET("/gitlab/callback", authService.HandleGitLabCallback)
			auth.POST("/gitlab/callback", authService.HandleGitLabCallback)
			auth.POST("/logout", authService.Logout)
		}

		// 需要认证的路由
		protected := api.Group("/")
		// protected.Use(authService.AuthMiddleware()) // 暂时注释掉，便于测试
		{
			// 用户相关路由
			users := protected.Group("/users")
			{
				users.GET("/health", userHandler.HealthCheck)
				users.GET("/active", userHandler.ListActiveUsers)
				users.GET("/:id", userHandler.GetUserByID)
				users.GET("/me", authService.GetCurrentUser)
				users.PUT("/me", userHandler.UpdateUser)
				users.GET("/me/dashboard", userHandler.GetUserDashboard)
				users.POST("/sync/:gitlab_id", userHandler.SyncUserFromGitLab)
			}

			// 文档相关路由
			documents := protected.Group("/documents")
			{
				documents.POST("/upload", documentHandler.UploadDocument)
				documents.GET("/test", documentHandler.TestUpload)
				documents.GET("/:id/editor", documentHandler.GetDocumentEditor)
				documents.GET("/:id/config", documentHandler.GetDocumentConfig)
				documents.GET("/:id/content", documentHandler.GetDocumentContent)
				documents.POST("/:id/callback", documentHandler.HandleCallback)
			}

			// Wiki相关路由
			wikiHandler.RegisterRoutes(protected)

			// 教育测试路由
			educationTestHandler.RegisterRoutes(protected)

			// 教育管理路由
			educationHandler.RegisterRoutes(protected)

			// 学习进度跟踪路由
			learningProgressHandler.RegisterRoutes(protected)

			// 通知系统路由
			notificationHandler.RegisterRoutes(protected)

			// 教育报表路由
			educationReportHandler.RegisterRoutes(protected)
		}
	}

	// 根路径 - 返回演示首页
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	return router
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
