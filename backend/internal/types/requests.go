package types

// AdminCreateUserRequest 管理员创建用户请求
type AdminCreateUserRequest struct {
	Username    string `json:"username" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	IsAdmin     bool   `json:"is_admin"`
	DefaultRole string `json:"default_role"`
}

// AdminUpdateUserRequest 管理员更新用户请求
type AdminUpdateUserRequest struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	IsAdmin  *bool  `json:"is_admin"`
}

// AdminUpdateUserRolesRequest 管理员更新用户角色请求
type AdminUpdateUserRolesRequest struct {
	IsAdmin      *bool `json:"is_admin"`
	ProjectRoles []struct {
		ProjectID string `json:"project_id"`
		Role      string `json:"role"`
	} `json:"project_roles"`
}

// SyncCreateUserRequest 同步创建用户请求
type SyncCreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Role     string `json:"role" binding:"required,oneof=admin teacher assistant student"`

	// 可选字段
	AvatarURL  string `json:"avatar_url,omitempty"`
	Department string `json:"department,omitempty"`
	StudentID  string `json:"student_id,omitempty"`
	TeacherID  string `json:"teacher_id,omitempty"`
	Phone      string `json:"phone,omitempty"`

	// 第三方系统标识
	ExternalID     string `json:"external_id,omitempty"`
	ExternalSource string `json:"external_source,omitempty"`
}

// SyncUpdateUserRequest 同步更新用户请求
type SyncUpdateUserRequest struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	IsAdmin  *bool  `json:"is_admin"`
}

// BatchCreateUsersRequest 批量创建用户请求
type BatchCreateUsersRequest struct {
	Users []SyncCreateUserRequest `json:"users" binding:"required"`
}

// PermissionRequest 权限检查请求
type PermissionRequest struct {
	Action     string `json:"action" binding:"required"`
	Resource   string `json:"resource" binding:"required"`
	ResourceID string `json:"resource_id,omitempty"`
}
