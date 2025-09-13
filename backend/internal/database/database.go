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
	log.Println("Starting database migration...")

	// 定义所有需要迁移的模型
	// 注意：User模型已移除，用户信息完全从GitLab API获取
	models := []interface{}{
		&models.ResearchProject{},
		&models.Document{},
		&models.DocumentEditRequest{},
		&models.Homework{},
		&models.Submission{},
		&models.Topic{},
		&models.TopicLike{},
		&models.Comment{},
		&models.Notification{},
		&models.Announcement{},
		&models.DocumentReview{},
		&models.AssignmentTemplate{},
	}

	// 直接执行AutoMigrate，GORM会自动处理：
	// 1. 创建不存在的表
	// 2. 添加不存在的列
	// 3. 创建索引
	// 4. 不会删除现有列或数据
	log.Println("Running AutoMigrate for all models...")
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database migration completed")
	return nil
}
