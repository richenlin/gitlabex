package dto

// ========== 用户管理相关 DTO ==========

// CreateUserData 创建用户数据结构
type CreateUserData struct {
	Username    string
	Name        string
	Email       string
	Password    string
	IsAdmin     bool
	DefaultRole string
}

// UpdateUserData 更新用户数据结构
type UpdateUserData struct {
	Username string
	Name     string
	Email    string
	IsAdmin  *bool
}

// UpdateUserRolesData 更新用户角色数据结构
type UpdateUserRolesData struct {
	IsAdmin      *bool
	ProjectRoles []struct {
		ProjectID string
		Role      string
	}
}

// AddSSHKeyReq 添加SSH密钥请求
type AddSSHKeyReq struct {
	Title string `json:"title" binding:"required"`
	Key   string `json:"key" binding:"required"`
}

// ChangePasswordReq 修改密码请求
type ChangePasswordReq struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Users    interface{} `json:"users"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Message  string      `json:"message,omitempty"`
}
