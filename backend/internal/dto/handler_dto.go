package dto

// ========== Handler 响应相关 DTO ==========
// 这些 DTO 主要用于 HTTP 请求/响应

// CreateUserRequest 创建用户请求 (Handler层使用)
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Role     string `json:"role" binding:"required"`
}

// CreateUserResponse 创建用户响应
type CreateUserResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Data    *UserData `json:"data,omitempty"`
	Error   string    `json:"error,omitempty"`
}

// UserData 用户数据
type UserData struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

// GetUserResponse 获取用户信息响应
type GetUserResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Data    *UserData `json:"data,omitempty"`
	Error   string    `json:"error,omitempty"`
}
