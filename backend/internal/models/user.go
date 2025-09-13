package models

// GitLabRole GitLab角色枚举 - 基于GitLab权限映射
type GitLabRole int

const (
	GitLabGuest      GitLabRole = 10 // GitLab Guest -> 访客
	GitLabReporter   GitLabRole = 20 // GitLab Reporter -> 学生
	GitLabDeveloper  GitLabRole = 30 // GitLab Developer -> 研究员
	GitLabMaintainer GitLabRole = 40 // GitLab Maintainer -> 教师
	GitLabOwner      GitLabRole = 50 // GitLab Owner -> 管理员
)

// GitLabUser GitLab用户信息 - 不存储在本地数据库，仅用于API传输
type GitLabUser struct {
	ID        int64      `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	AvatarURL string     `json:"avatar_url"`
	IsAdmin   bool       `json:"is_admin"`
	Role      GitLabRole `json:"role,omitempty"` // 在项目上下文中的角色
}

// GetEducationRole 获取教育角色名称
func (r GitLabRole) GetEducationRole() string {
	switch r {
	case GitLabOwner:
		return "管理员"
	case GitLabMaintainer:
		return "教师"
	case GitLabDeveloper:
		return "研究员"
	case GitLabReporter:
		return "学生"
	case GitLabGuest:
		return "访客"
	default:
		return "未知"
	}
}

// GetRoleString 获取角色字符串
func (r GitLabRole) GetRoleString() string {
	switch r {
	case GitLabOwner:
		return "owner"
	case GitLabMaintainer:
		return "maintainer"
	case GitLabDeveloper:
		return "developer"
	case GitLabReporter:
		return "reporter"
	case GitLabGuest:
		return "guest"
	default:
		return "guest"
	}
}

// ParseGitLabRole 解析GitLab角色
func ParseGitLabRole(accessLevel int) GitLabRole {
	switch accessLevel {
	case 50:
		return GitLabOwner
	case 40:
		return GitLabMaintainer
	case 30:
		return GitLabDeveloper
	case 20:
		return GitLabReporter
	case 10:
		return GitLabGuest
	default:
		return GitLabGuest
	}
}

// 注意：完全移除本地User模型，用户信息完全从GitLab API获取
// 所有用户相关的外键关系将使用GitLab用户ID (int64)而不是本地UUID
