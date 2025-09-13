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

	// 检查数据库是否为空（没有任何表）
	var tableCount int64
	err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tableCount)
	if err != nil {
		log.Printf("Warning: Could not check table count: %v", err)
	}

	if tableCount == 0 {
		log.Println("Database is empty, creating all tables...")
		return db.AutoMigrate(models...)
	}

	// 数据库不为空，检查每个表是否存在，只迁移缺失的表
	log.Printf("Found %d existing tables, checking for missing tables...", tableCount)

	// 逐个检查和迁移模型
	for _, model := range models {
		if !db.Migrator().HasTable(model) {
			log.Printf("Creating missing table for model: %T", model)
			if err := db.AutoMigrate(model); err != nil {
				return fmt.Errorf("failed to migrate model %T: %w", model, err)
			}
		}
	}

	// 检查现有表的列更新（GORM会安全地添加缺失的列）
	log.Println("Checking for column updates...")
	for _, model := range models {
		if db.Migrator().HasTable(model) {
			if err := db.AutoMigrate(model); err != nil {
				// 如果是权限错误，记录警告但不中断
				log.Printf("Warning: Could not update table for model %T: %v", model, err)
			}
		}
	}

	log.Println("Database migration completed")
	return nil
}
