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
