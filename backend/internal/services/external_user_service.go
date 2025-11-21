package services

import (
	"encoding/json"
	"fmt"
	"time"

	"gitlabex/internal/config"
	"gitlabex/internal/dto"
	"gitlabex/internal/models"

	"gorm.io/gorm"
)

// ExternalUserService 外部用户服务
type ExternalUserService struct {
	*BaseService
	gitlabService *GitLabService
}

// NewExternalUserService 创建外部用户服务实例
func NewExternalUserService(db *gorm.DB, gitlabService *GitLabService, cfg *config.Config) *ExternalUserService {
	return &ExternalUserService{
		BaseService:   NewBaseService(db, cfg),
		gitlabService: gitlabService,
	}
}

// GetExternalUserByID 根据外部系统ID和来源获取用户映射
func (s *ExternalUserService) GetExternalUserByID(externalID, externalSource string) (*models.ExternalUser, error) {
	var externalUser models.ExternalUser
	err := s.DB.Where("external_id = ? AND external_source = ?", externalID, externalSource).First(&externalUser).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询外部用户失败: %v", err)
	}
	return &externalUser, nil
}

// GetExternalUserByGitLabID 根据GitLab用户ID获取外部用户映射
func (s *ExternalUserService) GetExternalUserByGitLabID(gitlabUserID int64) ([]*models.ExternalUser, error) {
	var externalUsers []*models.ExternalUser
	err := s.DB.Where("gitlab_user_id = ? AND is_active = ?", gitlabUserID, true).Find(&externalUsers).Error
	if err != nil {
		return nil, fmt.Errorf("查询外部用户映射失败: %v", err)
	}
	return externalUsers, nil
}

// CreateExternalUserMapping 创建外部用户映射
func (s *ExternalUserService) CreateExternalUserMapping(gitlabUserID int64, externalUser *models.ExternalUser) error {
	externalUser.GitLabUserID = gitlabUserID
	externalUser.IsActive = true
	now := time.Now()
	externalUser.LastSyncAt = &now

	err := s.DB.Create(externalUser).Error
	if err != nil {
		return fmt.Errorf("创建外部用户映射失败: %v", err)
	}

	return nil
}

// UpdateExternalUserMapping 更新外部用户映射
func (s *ExternalUserService) UpdateExternalUserMapping(externalUser *models.ExternalUser) error {
	now := time.Now()
	externalUser.LastSyncAt = &now

	err := s.DB.Save(externalUser).Error
	if err != nil {
		return fmt.Errorf("更新外部用户映射失败: %v", err)
	}

	return nil
}

// SyncExternalUser 同步外部用户到GitLab
func (s *ExternalUserService) SyncExternalUser(adminToken string, userData *dto.ExternalUserSyncData, apiKeyType string, clientIP string) (*models.ExternalUserMapping, error) {
	// 记录同步日志
	syncLog := &models.ExternalUserSyncLog{
		Operation: "create",
		Status:    models.SyncStatusPending,
		APIKey:    apiKeyType,
		ClientIP:  clientIP,
	}

	// 序列化请求数据
	requestData, _ := json.Marshal(userData)
	syncLog.RequestData = string(requestData)

	// 检查是否已存在外部用户映射
	existingUser, err := s.GetExternalUserByID(userData.ExternalID, userData.ExternalSource)
	if err != nil {
		syncLog.Status = models.SyncStatusFailed
		syncLog.ErrorMessage = fmt.Sprintf("查询外部用户失败: %v", err)
		s.DB.Create(syncLog)
		return nil, err
	}

	var gitlabUser *dto.GitLabAPIUser
	var externalUser *models.ExternalUser

	if existingUser != nil {
		// 用户已存在，更新信息
		syncLog.Operation = "update"
		syncLog.ExternalUserID = existingUser.ID

		// 更新GitLab用户信息
		gitlabUpdateData := &dto.GitLabUpdateUserData{
			Username: userData.Username,
			Name:     userData.Name,
			Email:    userData.Email,
		}

		gitlabUser, err = s.gitlabService.UpdateUser(adminToken, existingUser.GitLabUserID, gitlabUpdateData)
		if err != nil {
			syncLog.Status = models.SyncStatusFailed
			syncLog.ErrorMessage = fmt.Sprintf("更新GitLab用户失败: %v", err)
			s.DB.Create(syncLog)
			return nil, fmt.Errorf("更新GitLab用户失败: %v", err)
		}

		// 更新外部用户映射
		existingUser.Username = userData.Username
		existingUser.Email = userData.Email
		existingUser.Name = userData.Name
		existingUser.Role = userData.Role
		existingUser.Department = userData.Department
		existingUser.StudentID = userData.StudentID
		existingUser.TeacherID = userData.TeacherID
		existingUser.Phone = userData.Phone

		err = s.UpdateExternalUserMapping(existingUser)
		if err != nil {
			syncLog.Status = models.SyncStatusFailed
			syncLog.ErrorMessage = fmt.Sprintf("更新外部用户映射失败: %v", err)
			s.DB.Create(syncLog)
			return nil, err
		}

		externalUser = existingUser
	} else {
		// 创建新用户
		syncLog.Operation = "create"

		// 在GitLab中创建用户
		gitlabCreateData := &dto.GitLabCreateUserData{
			Email:            userData.Email,
			Username:         userData.Username,
			Name:             userData.Name,
			Password:         userData.Password,
			SkipConfirmation: true,
		}

		gitlabUser, err = s.gitlabService.CreateUser(adminToken, gitlabCreateData)
		if err != nil {
			syncLog.Status = models.SyncStatusFailed
			syncLog.ErrorMessage = fmt.Sprintf("创建GitLab用户失败: %v", err)
			s.DB.Create(syncLog)
			return nil, fmt.Errorf("创建GitLab用户失败: %v", err)
		}

		// 创建外部用户映射
		externalUser = &models.ExternalUser{
			ExternalID:     userData.ExternalID,
			ExternalSource: userData.ExternalSource,
			Username:       userData.Username,
			Email:          userData.Email,
			Name:           userData.Name,
			Role:           userData.Role,
			Department:     userData.Department,
			StudentID:      userData.StudentID,
			TeacherID:      userData.TeacherID,
			Phone:          userData.Phone,
		}

		err = s.CreateExternalUserMapping(gitlabUser.ID, externalUser)
		if err != nil {
			syncLog.Status = models.SyncStatusFailed
			syncLog.ErrorMessage = fmt.Sprintf("创建外部用户映射失败: %v", err)
			s.DB.Create(syncLog)
			return nil, err
		}

		// 根据角色将用户分配到相应的GitLab用户组
		if err := s.gitlabService.AssignUserToRoleGroup(adminToken, gitlabUser.ID, userData.Role); err != nil {
			// 用户组分配失败不应该阻止用户创建，只记录警告
			fmt.Printf("警告：为用户 %d 分配用户组失败: %v\n", gitlabUser.ID, err)
		}

		syncLog.ExternalUserID = externalUser.ID
	}

	// 转换为内部GitLab用户格式
	internalGitlabUser := &models.GitLabUser{
		ID:        gitlabUser.ID,
		Username:  gitlabUser.Username,
		Email:     gitlabUser.Email,
		Name:      gitlabUser.Name,
		AvatarURL: gitlabUser.Avatar,
		IsAdmin:   gitlabUser.IsAdmin,
		Role:      models.ConvertExternalRoleToGitLab(userData.Role),
	}

	// 记录成功日志
	syncLog.Status = models.SyncStatusSuccess
	responseData, _ := json.Marshal(map[string]interface{}{
		"gitlab_user_id":   gitlabUser.ID,
		"external_user_id": externalUser.ID,
	})
	syncLog.ResponseData = string(responseData)
	s.DB.Create(syncLog)

	return &models.ExternalUserMapping{
		GitLabUser:   internalGitlabUser,
		ExternalUser: externalUser,
	}, nil
}

// BatchSyncExternalUsers 批量同步外部用户
func (s *ExternalUserService) BatchSyncExternalUsers(adminToken string, usersData []dto.ExternalUserSyncData, apiKeyType string, clientIP string, maxBatch int) (*dto.BatchSyncResult, error) {
	if len(usersData) > maxBatch {
		return nil, fmt.Errorf("批量同步数量超出限制，最多支持 %d 个用户", maxBatch)
	}

	result := &dto.BatchSyncResult{
		TotalCount:   len(usersData),
		SuccessCount: 0,
		FailureCount: 0,
		Results:      make([]dto.SyncResult, len(usersData)),
	}

	for i, userData := range usersData {
		mapping, err := s.SyncExternalUser(adminToken, &userData, apiKeyType, clientIP)
		if err != nil {
			result.Results[i] = dto.SyncResult{
				Success:      false,
				ErrorMessage: err.Error(),
				ExternalID:   userData.ExternalID,
			}
			result.FailureCount++
		} else {
			result.Results[i] = dto.SyncResult{
				Success:      true,
				GitLabUserID: mapping.GitLabUser.ID,
				ExternalID:   userData.ExternalID,
				Username:     mapping.GitLabUser.Username,
				Email:        mapping.GitLabUser.Email,
			}
			result.SuccessCount++
		}
	}

	return result, nil
}

// GetExternalSystemStats 获取外部系统统计信息
func (s *ExternalUserService) GetExternalSystemStats() ([]*models.ExternalSystemStats, error) {
	var stats []*models.ExternalSystemStats

	err := s.DB.Model(&models.ExternalUser{}).
		Select("external_source as source, COUNT(*) as total_users, COUNT(CASE WHEN is_active = true THEN 1 END) as active_users, MAX(last_sync_at) as last_sync_at").
		Where("deleted_at IS NULL").
		Group("external_source").
		Find(&stats).Error

	if err != nil {
		return nil, fmt.Errorf("获取外部系统统计失败: %v", err)
	}

	return stats, nil
}

// GetSyncLogs 获取同步日志
func (s *ExternalUserService) GetSyncLogs(page, pageSize int, externalSource string) ([]*models.ExternalUserSyncLog, int64, error) {
	var logs []*models.ExternalUserSyncLog
	var total int64

	query := s.DB.Model(&models.ExternalUserSyncLog{}).Preload("ExternalUser")

	if externalSource != "" {
		query = query.Joins("JOIN external_users ON external_users.id = external_user_sync_logs.external_user_id").
			Where("external_users.external_source = ?", externalSource)
	}

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("获取同步日志总数失败: %v", err)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	err = query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&logs).Error
	if err != nil {
		return nil, 0, fmt.Errorf("获取同步日志失败: %v", err)
	}

	return logs, total, nil
}
