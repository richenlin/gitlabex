package dto

// ========== 外部用户同步相关 DTO ==========

// ExternalUserSyncData 外部用户同步数据
type ExternalUserSyncData struct {
	ExternalID     string `json:"external_id" binding:"required"`
	ExternalSource string `json:"external_source" binding:"required"`
	Username       string `json:"username" binding:"required,min=3,max=50"`
	Password       string `json:"password" binding:"required,min=6"`
	Email          string `json:"email" binding:"required,email"`
	Name           string `json:"name" binding:"required,min=2,max=100"`
	Role           string `json:"role" binding:"required"`
	Department     string `json:"department,omitempty"`
	StudentID      string `json:"student_id,omitempty"`
	TeacherID      string `json:"teacher_id,omitempty"`
	Phone          string `json:"phone,omitempty"`
}

// BatchSyncResult 批量同步结果
type BatchSyncResult struct {
	TotalCount   int          `json:"total_count"`
	SuccessCount int          `json:"success_count"`
	FailureCount int          `json:"failure_count"`
	Results      []SyncResult `json:"results"`
}

// SyncResult 单个同步结果
type SyncResult struct {
	Success      bool   `json:"success"`
	GitLabUserID int64  `json:"gitlab_user_id,omitempty"`
	ExternalID   string `json:"external_id"`
	Username     string `json:"username,omitempty"`
	Email        string `json:"email,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}
