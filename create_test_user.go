package main

import (
	"fmt"
	"log"
	"time"

	"gitlabex/internal/config"
	"gitlabex/internal/database"
	"gitlabex/internal/models"

	"github.com/google/uuid"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 连接数据库
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 创建测试用户
	userID := "6d1dcb11-ee3c-4811-9502-20ca86922983" // 使用测试中的用户ID
	parsedUserID, _ := uuid.Parse(userID)

	user := models.User{
		BaseModel: models.BaseModel{
			ID: parsedUserID,
		},
		GitLabID:     123456,
		Username:     "test_user",
		Email:        "test@example.com",
		Name:         "Test User",
		Role:         models.RoleTeacher,
		EduRole:      models.EduRoleTeacher,
		IsActive:     true,
		LastLoginAt:  &[]time.Time{time.Now()}[0],
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
	}

	// 检查用户是否已存在
	var existingUser models.User
	result := db.First(&existingUser, "id = ?", parsedUserID)
	if result.Error == nil {
		fmt.Printf("User already exists: %s\n", existingUser.Username)
		return
	}

	// 创建用户
	if err := db.Create(&user).Error; err != nil {
		log.Fatal("Failed to create user:", err)
	}

	fmt.Printf("Created test user: %s (ID: %s)\n", user.Username, user.ID)
}
