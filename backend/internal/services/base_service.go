package services

import (
	"gitlabex/internal/config"
	"gorm.io/gorm"
)

// BaseService 基础服务
type BaseService struct {
	DB     *gorm.DB
	Config *config.Config
}

// NewBaseService 创建基础服务
func NewBaseService(db *gorm.DB, cfg *config.Config) *BaseService {
	return &BaseService{
		DB:     db,
		Config: cfg,
	}
}