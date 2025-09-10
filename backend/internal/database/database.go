package database

import (
	"fmt"
	"log"

	"gitlabex/internal/config"
	"gitlabex/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Initialize 初始化数据库连接
func Initialize(cfg *config.Config) (*gorm.DB, error) {
	// 配置GORM日志级别
	var logLevel logger.LogLevel
	if cfg.Debug {
		logLevel = logger.Info
	} else {
		logLevel = logger.Error
	}

	// 连接数据库
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取SQL数据库连接以配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	// 自动迁移数据库表
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database connected and migrated successfully")
	return db, nil
}

// autoMigrate 自动迁移数据库表
func autoMigrate(db *gorm.DB) error {
	// 检查是否已经有表存在
	var count int64
	err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users'").Scan(&count)
	if err != nil {
		log.Printf("Warning: Could not check if tables exist: %v", err)
		// 如果检查失败，继续尝试迁移，让GORM处理
	} else if count > 0 {
		log.Println("Tables already exist, skipping AutoMigrate to avoid conflicts")
		// 表已存在，只进行必要的列更新而不创建表
		// 这里可以添加特定的迁移逻辑如果需要
		return nil
	}

	log.Println("Creating database tables...")
	return db.AutoMigrate(
		&models.User{},
		&models.ResearchProject{},
		&models.ProjectMember{},
		&models.Document{},
		&models.Homework{},
		&models.Submission{},
		&models.Topic{},
		&models.TopicLike{},
		&models.Comment{},
		&models.Notification{},
		&models.Announcement{},
		&models.DocumentReview{},
		&models.AssignmentTemplate{},
	)
}
