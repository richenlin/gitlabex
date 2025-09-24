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
	if cfg.Server.Debug {
		logLevel = logger.Info
	} else {
		logLevel = logger.Error
	}

	// 连接数据库
	db, err := gorm.Open(postgres.Open(cfg.GetDatabaseURL()), &gorm.Config{
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

	// 首先修复research_projects表的creator_id列类型问题
	if err := fixResearchProjectCreatorID(db); err != nil {
		return fmt.Errorf("failed to fix research_projects creator_id: %w", err)
	}

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
		// 第三方系统集成相关模型
		&models.ExternalUser{},
		&models.ExternalUserSyncLog{},
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

// fixResearchProjectCreatorID 修复research_projects表的creator_id列类型问题
func fixResearchProjectCreatorID(db *gorm.DB) error {
	log.Println("Checking research_projects table structure...")

	// 检查表是否存在
	var tableExists bool
	err := db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'research_projects')").Scan(&tableExists).Error
	if err != nil {
		return fmt.Errorf("failed to check if research_projects table exists: %w", err)
	}

	if !tableExists {
		log.Println("research_projects table does not exist, skipping creator_id fix")
		return nil
	}

	// 检查creator_id列的数据类型
	var dataType string
	err = db.Raw("SELECT data_type FROM information_schema.columns WHERE table_name = 'research_projects' AND column_name = 'creator_id'").Scan(&dataType).Error
	if err != nil {
		// 如果列不存在，GORM的AutoMigrate会创建它，不需要特别处理
		log.Println("creator_id column might not exist, AutoMigrate will handle it")
		return nil
	}

	log.Printf("Current creator_id data type: %s", dataType)

	// 如果是uuid类型，需要转换为bigint
	if dataType == "uuid" {
		log.Println("Converting creator_id from UUID to bigint...")

		// 首先删除现有数据（开发环境）
		log.Println("Dropping existing research_projects data for migration...")
		if err := db.Exec("DELETE FROM research_projects").Error; err != nil {
			return fmt.Errorf("failed to delete research_projects data: %w", err)
		}

		// 然后修改列类型
		if err := db.Exec("ALTER TABLE research_projects ALTER COLUMN creator_id TYPE bigint USING 0").Error; err != nil {
			return fmt.Errorf("failed to alter creator_id column type: %w", err)
		}

		log.Println("Successfully converted creator_id to bigint")
	} else if dataType != "bigint" {
		log.Printf("creator_id is already %s type, no conversion needed", dataType)
	}

	return nil
}
