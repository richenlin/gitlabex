package config

import "gitlabex/internal/models"

// 角色常量 - int类型，与User.Role字段对应
const (
	RoleAdmin     = int(models.EduRoleAdmin)     // 50 - 管理员
	RoleTeacher   = int(models.EduRoleTeacher)   // 40 - 教师
	RoleAssistant = int(models.EduRoleAssistant) // 30 - 助教
	RoleStudent   = int(models.EduRoleStudent)   // 20 - 学生
	RoleGuest     = int(models.EduRoleGuest)     // 10 - 访客
)

// 权限常量
const (
	PermissionRead   = "read"
	PermissionWrite  = "write"
	PermissionManage = "manage"
	PermissionDelete = "delete"
	PermissionSubmit = "submit"
	PermissionReview = "review"
)

// IsValidRole 验证角色是否有效
func IsValidRole(role int) bool {
	return role == RoleAdmin || role == RoleTeacher || role == RoleAssistant ||
		role == RoleStudent || role == RoleGuest
}
