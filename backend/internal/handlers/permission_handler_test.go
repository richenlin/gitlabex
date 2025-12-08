package handlers

import (
	"gitlabex/internal/models"
	"testing"
)

// TestRolePermissions 测试角色权限映射
func TestRolePermissions(t *testing.T) {
	tests := []struct {
		name       string
		role       models.GitLabRole
		permission models.Permission
		expected   bool
	}{
		// 管理员权限测试
		{
			name:       "管理员可以创建课题",
			role:       models.GitLabOwner,
			permission: models.PermProjectCreate,
			expected:   true,
		},
		{
			name:       "管理员可以删除课题",
			role:       models.GitLabOwner,
			permission: models.PermProjectDelete,
			expected:   true,
		},
		{
			name:       "管理员可以批改作业",
			role:       models.GitLabOwner,
			permission: models.PermHomeworkGrade,
			expected:   true,
		},
		{
			name:       "管理员可以审核文档",
			role:       models.GitLabOwner,
			permission: models.PermDocumentApprove,
			expected:   true,
		},

		// 教师权限测试
		{
			name:       "教师可以创建课题",
			role:       models.GitLabMaintainer,
			permission: models.PermProjectCreate,
			expected:   true,
		},
		{
			name:       "教师可以编辑课题",
			role:       models.GitLabMaintainer,
			permission: models.PermProjectUpdate,
			expected:   true,
		},
		{
			name:       "教师可以删除课题",
			role:       models.GitLabMaintainer,
			permission: models.PermProjectDelete,
			expected:   true,
		},
		{
			name:       "教师可以创建作业",
			role:       models.GitLabMaintainer,
			permission: models.PermHomeworkCreate,
			expected:   true,
		},
		{
			name:       "教师可以批改作业",
			role:       models.GitLabMaintainer,
			permission: models.PermHomeworkGrade,
			expected:   true,
		},
		{
			name:       "教师可以审核文档",
			role:       models.GitLabMaintainer,
			permission: models.PermDocumentApprove,
			expected:   true,
		},
		{
			name:       "教师可以同步文档",
			role:       models.GitLabMaintainer,
			permission: models.PermDocumentSync,
			expected:   true,
		},

		// 研究员权限测试
		{
			name:       "研究员不能创建课题",
			role:       models.GitLabDeveloper,
			permission: models.PermProjectCreate,
			expected:   false,
		},
		{
			name:       "研究员可以查看课题",
			role:       models.GitLabDeveloper,
			permission: models.PermProjectRead,
			expected:   true,
		},
		{
			name:       "研究员不能删除课题",
			role:       models.GitLabDeveloper,
			permission: models.PermProjectDelete,
			expected:   false,
		},
		{
			name:       "研究员可以创建话题",
			role:       models.GitLabDeveloper,
			permission: models.PermTopicCreate,
			expected:   true,
		},
		{
			name:       "研究员可以编辑话题",
			role:       models.GitLabDeveloper,
			permission: models.PermTopicUpdate,
			expected:   true,
		},
		{
			name:       "研究员可以批改作业",
			role:       models.GitLabDeveloper,
			permission: models.PermHomeworkGrade,
			expected:   true,
		},
		{
			name:       "研究员不能创建作业",
			role:       models.GitLabDeveloper,
			permission: models.PermHomeworkCreate,
			expected:   false,
		},
		{
			name:       "研究员不能提交作业",
			role:       models.GitLabDeveloper,
			permission: models.PermHomeworkSubmit,
			expected:   false,
		},
		{
			name:       "研究员可以创建文档",
			role:       models.GitLabDeveloper,
			permission: models.PermDocumentCreate,
			expected:   true,
		},
		{
			name:       "研究员不能审核文档",
			role:       models.GitLabDeveloper,
			permission: models.PermDocumentApprove,
			expected:   false,
		},

		// 普通用户/学生权限测试
		{
			name:       "学生不能创建课题",
			role:       models.GitLabReporter,
			permission: models.PermProjectCreate,
			expected:   false,
		},
		{
			name:       "学生可以查看课题",
			role:       models.GitLabReporter,
			permission: models.PermProjectRead,
			expected:   true,
		},
		{
			name:       "学生可以创建话题",
			role:       models.GitLabReporter,
			permission: models.PermTopicCreate,
			expected:   true,
		},
		{
			name:       "学生可以点赞话题",
			role:       models.GitLabReporter,
			permission: models.PermTopicLike,
			expected:   true,
		},
		{
			name:       "学生可以评论话题",
			role:       models.GitLabReporter,
			permission: models.PermTopicComment,
			expected:   true,
		},
		{
			name:       "学生可以提交作业",
			role:       models.GitLabReporter,
			permission: models.PermHomeworkSubmit,
			expected:   true,
		},
		{
			name:       "学生不能批改作业",
			role:       models.GitLabReporter,
			permission: models.PermHomeworkGrade,
			expected:   false,
		},
		{
			name:       "学生可以创建文档",
			role:       models.GitLabReporter,
			permission: models.PermDocumentCreate,
			expected:   true,
		},
		{
			name:       "学生不能审核文档",
			role:       models.GitLabReporter,
			permission: models.PermDocumentApprove,
			expected:   false,
		},

		// 访客权限测试
		{
			name:       "访客可以查看课题",
			role:       models.GitLabGuest,
			permission: models.PermProjectRead,
			expected:   true,
		},
		{
			name:       "访客不能创建课题",
			role:       models.GitLabGuest,
			permission: models.PermProjectCreate,
			expected:   false,
		},
		{
			name:       "访客可以查看话题",
			role:       models.GitLabGuest,
			permission: models.PermTopicRead,
			expected:   true,
		},
		{
			name:       "访客不能创建话题",
			role:       models.GitLabGuest,
			permission: models.PermTopicCreate,
			expected:   false,
		},
		{
			name:       "访客不能点赞话题",
			role:       models.GitLabGuest,
			permission: models.PermTopicLike,
			expected:   false,
		},
		{
			name:       "访客可以查看文档",
			role:       models.GitLabGuest,
			permission: models.PermDocumentRead,
			expected:   true,
		},
		{
			name:       "访客不能创建文档",
			role:       models.GitLabGuest,
			permission: models.PermDocumentCreate,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.role.HasPermission(tt.permission)
			if result != tt.expected {
				t.Errorf("角色 %s 对权限 %s 的检查结果: got %v, want %v",
					tt.role.GetEducationRole(), tt.permission, result, tt.expected)
			}
		})
	}
}

// TestPermissionContext 测试权限上下文
func TestPermissionContext(t *testing.T) {
	tests := []struct {
		name       string
		ctx        *models.PermissionContext
		permission models.Permission
		expected   bool
	}{
		{
			name: "管理员拥有所有权限",
			ctx: &models.PermissionContext{
				UserID:  1,
				IsAdmin: true,
			},
			permission: models.PermProjectDelete,
			expected:   true,
		},
		{
			name: "教师可以创建课题",
			ctx: &models.PermissionContext{
				UserID:      2,
				IsAdmin:     false,
				Role:        models.GitLabMaintainer,
				AccessLevel: 40,
			},
			permission: models.PermProjectCreate,
			expected:   true,
		},
		{
			name: "研究员不能删除课题",
			ctx: &models.PermissionContext{
				UserID:      3,
				IsAdmin:     false,
				Role:        models.GitLabDeveloper,
				AccessLevel: 30,
			},
			permission: models.PermProjectDelete,
			expected:   false,
		},
		{
			name: "学生可以提交作业",
			ctx: &models.PermissionContext{
				UserID:      4,
				IsAdmin:     false,
				Role:        models.GitLabReporter,
				AccessLevel: 20,
			},
			permission: models.PermHomeworkSubmit,
			expected:   true,
		},
		{
			name: "学生不能批改作业",
			ctx: &models.PermissionContext{
				UserID:      4,
				IsAdmin:     false,
				Role:        models.GitLabReporter,
				AccessLevel: 20,
			},
			permission: models.PermHomeworkGrade,
			expected:   false,
		},
		{
			name: "访客只能查看",
			ctx: &models.PermissionContext{
				UserID:      5,
				IsAdmin:     false,
				Role:        models.GitLabGuest,
				AccessLevel: 10,
			},
			permission: models.PermTopicRead,
			expected:   true,
		},
		{
			name: "访客不能创建",
			ctx: &models.PermissionContext{
				UserID:      5,
				IsAdmin:     false,
				Role:        models.GitLabGuest,
				AccessLevel: 10,
			},
			permission: models.PermTopicCreate,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ctx.CanPerform(tt.permission)
			if result != tt.expected {
				t.Errorf("权限检查失败: got %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestParseGitLabRole 测试 GitLab 角色解析
func TestParseGitLabRole(t *testing.T) {
	tests := []struct {
		accessLevel int
		expected    models.GitLabRole
	}{
		{50, models.GitLabOwner},
		{40, models.GitLabMaintainer},
		{30, models.GitLabDeveloper},
		{20, models.GitLabReporter},
		{10, models.GitLabGuest},
		{0, models.GitLabGuest},  // 未知级别默认为 Guest
		{99, models.GitLabGuest}, // 未知级别默认为 Guest
	}

	for _, tt := range tests {
		result := models.ParseGitLabRole(tt.accessLevel)
		if result != tt.expected {
			t.Errorf("ParseGitLabRole(%d) = %v, want %v", tt.accessLevel, result, tt.expected)
		}
	}
}

// TestGetEducationRole 测试获取教育角色名称
func TestGetEducationRole(t *testing.T) {
	tests := []struct {
		role     models.GitLabRole
		expected string
	}{
		{models.GitLabOwner, "管理员"},
		{models.GitLabMaintainer, "教师"},
		{models.GitLabDeveloper, "研究员"},
		{models.GitLabReporter, "学生"},
		{models.GitLabGuest, "访客"},
	}

	for _, tt := range tests {
		result := tt.role.GetEducationRole()
		if result != tt.expected {
			t.Errorf("GetEducationRole(%v) = %s, want %s", tt.role, result, tt.expected)
		}
	}
}

// TestGetRoleString 测试获取角色字符串
func TestGetRoleString(t *testing.T) {
	tests := []struct {
		role     models.GitLabRole
		expected string
	}{
		{models.GitLabOwner, "owner"},
		{models.GitLabMaintainer, "maintainer"},
		{models.GitLabDeveloper, "developer"},
		{models.GitLabReporter, "reporter"},
		{models.GitLabGuest, "guest"},
	}

	for _, tt := range tests {
		result := tt.role.GetRoleString()
		if result != tt.expected {
			t.Errorf("GetRoleString(%v) = %s, want %s", tt.role, result, tt.expected)
		}
	}
}

// TestIsOwner 测试资源所有者检查
func TestIsOwner(t *testing.T) {
	tests := []struct {
		name     string
		ctx      *models.PermissionContext
		expected bool
	}{
		{
			name: "用户是资源所有者",
			ctx: &models.PermissionContext{
				UserID:  123,
				OwnerID: 123,
			},
			expected: true,
		},
		{
			name: "用户不是资源所有者",
			ctx: &models.PermissionContext{
				UserID:  123,
				OwnerID: 456,
			},
			expected: false,
		},
		{
			name: "资源没有所有者",
			ctx: &models.PermissionContext{
				UserID:  123,
				OwnerID: 0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ctx.IsOwner()
			if result != tt.expected {
				t.Errorf("IsOwner() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestHasMinAccessLevel 测试最小访问级别检查
func TestHasMinAccessLevel(t *testing.T) {
	tests := []struct {
		name        string
		ctx         *models.PermissionContext
		minLevel    int
		expected    bool
		description string
	}{
		{
			name: "教师满足Maintainer级别",
			ctx: &models.PermissionContext{
				AccessLevel: 40,
			},
			minLevel:    40,
			expected:    true,
			description: "教师访问级别40,要求40",
		},
		{
			name: "研究员不满足Maintainer级别",
			ctx: &models.PermissionContext{
				AccessLevel: 30,
			},
			minLevel:    40,
			expected:    false,
			description: "研究员访问级别30,要求40",
		},
		{
			name: "管理员满足所有级别",
			ctx: &models.PermissionContext{
				AccessLevel: 50,
			},
			minLevel:    40,
			expected:    true,
			description: "管理员访问级别50,要求40",
		},
		{
			name: "学生满足Reporter级别",
			ctx: &models.PermissionContext{
				AccessLevel: 20,
			},
			minLevel:    20,
			expected:    true,
			description: "学生访问级别20,要求20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ctx.HasMinAccessLevel(tt.minLevel)
			if result != tt.expected {
				t.Errorf("%s: HasMinAccessLevel(%d) = %v, want %v",
					tt.description, tt.minLevel, result, tt.expected)
			}
		})
	}
}
