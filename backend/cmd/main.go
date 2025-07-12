package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

	// 初始化数据库
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化服务
	permissionService := services.NewPermissionService(db)
	userService := services.NewUserService(db, permissionService)
	analyticsService := services.NewAnalyticsService(db)
	authService := services.NewAuthService(db, cfg)
	gitlabService, err := services.NewGitLabService(cfg, nil, db)
	if err != nil {
		log.Printf("Failed to initialize GitLab service: %v", err)
	}
	onlyofficeService := services.NewOnlyOfficeService(cfg, db)
	documentService := services.NewDocumentService(db, gitlabService, onlyofficeService)
	educationService := services.NewEducationServiceSimplified(gitlabService, db)

	// 初始化处理器
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService, userService)
	userHandler := handlers.NewUserHandler(userService)
	educationHandler := handlers.NewEducationHandler(educationService, userService)
	wikiHandler := handlers.NewWikiHandler(gitlabService, onlyofficeService, documentService)

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化路由
	router := setupRoutes(authService, analyticsHandler, userHandler, educationHandler, wikiHandler)

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

// autoMigrate 自动迁移数据库表
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
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
func setupRoutes(authService *services.AuthService, analyticsHandler *handlers.AnalyticsHandler, userHandler *handlers.UserHandler, educationHandler *handlers.EducationHandler, wikiHandler *handlers.WikiHandler) *gin.Engine {
	router := gin.New()

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

		// 认证相关路由
		auth := api.Group("/auth")
		{
			auth.GET("/gitlab", func(c *gin.Context) {
				// 生成随机state以防止CSRF攻击
				state := fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int63())

				// 直接使用配置的外部URL生成OAuth URL
				url := authService.GetGitLabOAuthURL(state)

				// 直接重定向到GitLab OAuth页面
				c.Redirect(302, url)
			})
			auth.GET("/gitlab/callback", authService.HandleGitLabCallback)
			auth.POST("/gitlab/callback", authService.HandleGitLabCallback)
			auth.POST("/logout", authService.Logout)
		}

		// 分析统计路由
		analytics := api.Group("/analytics")
		{
			analytics.GET("/overview", analyticsHandler.GetAnalyticsOverview)
			analytics.GET("/project-stats", analyticsHandler.GetProjectStats)
			analytics.GET("/student-stats", analyticsHandler.GetStudentStats)
			analytics.GET("/assignment-stats", analyticsHandler.GetAssignmentStats)
			analytics.GET("/submission-trend", analyticsHandler.GetSubmissionTrend)
			analytics.GET("/project-distribution", analyticsHandler.GetProjectDistribution)
			analytics.GET("/grade-distribution", analyticsHandler.GetGradeDistribution)
			analytics.GET("/activity-stats", analyticsHandler.GetActivityStats)
			analytics.GET("/dashboard-stats", analyticsHandler.GetDashboardStats)
			analytics.GET("/recent-activities", analyticsHandler.GetRecentActivities)
		}

		// 用户管理路由
		users := api.Group("/users")
		{
			users.GET("/active", userHandler.ListActiveUsers)
			users.GET("/current", userHandler.GetCurrentUser)
			users.GET("/:id", userHandler.GetUserByID)
			users.PUT("/current", userHandler.UpdateUser)
			users.GET("/dashboard", userHandler.GetUserDashboard)
			users.POST("/sync/:gitlab_id", userHandler.SyncUserFromGitLab)
		}

		// 教育管理路由
		educationHandler.RegisterRoutes(api)

		// Wiki管理路由
		wikiHandler.RegisterRoutes(api)
	}

	// 根路径
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "GitLabEx API Server",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	return router
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
