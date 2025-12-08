package models

// Permission 权限定义
type Permission string

// 资源类型
const (
	ResourceProject  = "project"  // 课题
	ResourceTopic    = "topic"    // 话题
	ResourceHomework = "homework" // 作业
	ResourceDocument = "document" // 文档
	ResourceUser     = "user"     // 用户
)

// 操作类型
const (
	ActionCreate  = "create"  // 创建
	ActionRead    = "read"    // 查看
	ActionUpdate  = "update"  // 编辑
	ActionDelete  = "delete"  // 删除
	ActionManage  = "manage"  // 管理
	ActionGrade   = "grade"   // 批改
	ActionSubmit  = "submit"  // 提交
	ActionApprove = "approve" // 审核
	ActionSync    = "sync"    // 同步
	ActionLike    = "like"    // 点赞
	ActionComment = "comment" // 评论
)

// 课题权限
const (
	PermProjectCreate        Permission = "project:create"         // 创建课题
	PermProjectRead          Permission = "project:read"           // 查看课题
	PermProjectUpdate        Permission = "project:update"         // 编辑课题
	PermProjectDelete        Permission = "project:delete"         // 删除课题
	PermProjectManageMembers Permission = "project:manage:members" // 管理课题成员
)

// 话题权限
const (
	PermTopicCreate  Permission = "topic:create"  // 创建话题
	PermTopicRead    Permission = "topic:read"    // 查看话题
	PermTopicUpdate  Permission = "topic:update"  // 编辑话题
	PermTopicDelete  Permission = "topic:delete"  // 删除话题
	PermTopicLike    Permission = "topic:like"    // 点赞话题
	PermTopicComment Permission = "topic:comment" // 评论话题
)

// 作业权限
const (
	PermHomeworkCreate Permission = "homework:create" // 创建作业
	PermHomeworkRead   Permission = "homework:read"   // 查看作业
	PermHomeworkUpdate Permission = "homework:update" // 编辑作业
	PermHomeworkDelete Permission = "homework:delete" // 删除作业
	PermHomeworkSubmit Permission = "homework:submit" // 提交作业
	PermHomeworkGrade  Permission = "homework:grade"  // 批改作业
)

// 文档权限
const (
	PermDocumentCreate  Permission = "document:create"  // 创建文档
	PermDocumentRead    Permission = "document:read"    // 查看文档
	PermDocumentUpdate  Permission = "document:update"  // 编辑文档
	PermDocumentDelete  Permission = "document:delete"  // 删除文档
	PermDocumentApprove Permission = "document:approve" // 审核文档
	PermDocumentSync    Permission = "document:sync"    // 同步文档
)

// 用户权限
const (
	PermUserManage Permission = "user:manage" // 管理用户
)

// RolePermissions 角色权限映射
var RolePermissions = map[GitLabRole][]Permission{
	// 管理员拥有所有权限
	GitLabOwner: {
		// 课题权限
		PermProjectCreate, PermProjectRead, PermProjectUpdate, PermProjectDelete, PermProjectManageMembers,
		// 话题权限
		PermTopicCreate, PermTopicRead, PermTopicUpdate, PermTopicDelete, PermTopicLike, PermTopicComment,
		// 作业权限
		PermHomeworkCreate, PermHomeworkRead, PermHomeworkUpdate, PermHomeworkDelete, PermHomeworkSubmit, PermHomeworkGrade,
		// 文档权限
		PermDocumentCreate, PermDocumentRead, PermDocumentUpdate, PermDocumentDelete, PermDocumentApprove, PermDocumentSync,
		// 用户权限
		PermUserManage,
	},

	// 教师权限（Maintainer）
	GitLabMaintainer: {
		// 课题权限 - 可以新建、编辑、删除课题
		PermProjectCreate, PermProjectRead, PermProjectUpdate, PermProjectDelete, PermProjectManageMembers,
		// 话题权限 - 可以新建、编辑、删除话题，可以点赞、评论
		PermTopicCreate, PermTopicRead, PermTopicUpdate, PermTopicDelete, PermTopicLike, PermTopicComment,
		// 作业权限 - 可以创建、编辑作业，可以批改作业
		PermHomeworkCreate, PermHomeworkRead, PermHomeworkUpdate, PermHomeworkDelete, PermHomeworkGrade,
		// 文档权限 - 可以新建、编辑、同步文档，可以审核文档
		PermDocumentCreate, PermDocumentRead, PermDocumentUpdate, PermDocumentDelete, PermDocumentApprove, PermDocumentSync,
	},

	// 研究员权限（Developer）
	GitLabDeveloper: {
		// 课题权限 - 可以查看所属课题
		PermProjectRead,
		// 话题权限 - 可以编辑所属课题的话题，可以发表话题、点赞、评论
		PermTopicCreate, PermTopicRead, PermTopicUpdate, PermTopicLike, PermTopicComment,
		// 作业权限 - 可以查看和批改课题作业
		PermHomeworkRead, PermHomeworkGrade,
		// 文档权限 - 可以新建、编辑自己的文档
		PermDocumentCreate, PermDocumentRead, PermDocumentUpdate,
	},

	// 普通用户/学生权限（Reporter）
	GitLabReporter: {
		// 课题权限 - 可以查看所属课题
		PermProjectRead,
		// 话题权限 - 可以发表话题、点赞、评论
		PermTopicCreate, PermTopicRead, PermTopicLike, PermTopicComment,
		// 作业权限 - 可以查看和提交作业
		PermHomeworkRead, PermHomeworkSubmit,
		// 文档权限 - 可以新建、编辑自己的文档
		PermDocumentCreate, PermDocumentRead, PermDocumentUpdate,
	},

	// 访客权限（Guest）
	GitLabGuest: {
		// 课题权限 - 只能查看公开课题
		PermProjectRead,
		// 话题权限 - 只能查看话题
		PermTopicRead,
		// 作业权限 - 无
		// 文档权限 - 只能查看文档
		PermDocumentRead,
	},
}

// HasPermission 检查角色是否有指定权限
func (r GitLabRole) HasPermission(perm Permission) bool {
	perms, exists := RolePermissions[r]
	if !exists {
		return false
	}

	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}

// GetPermissions 获取角色的所有权限
func (r GitLabRole) GetPermissions() []Permission {
	perms, exists := RolePermissions[r]
	if !exists {
		return []Permission{}
	}
	return perms
}

// PermissionContext 权限上下文
type PermissionContext struct {
	UserID      int64      // 用户ID
	Role        GitLabRole // 用户角色
	IsAdmin     bool       // 是否是管理员
	ProjectID   *int64     // 项目ID（可选）
	ResourceID  string     // 资源ID（可选）
	OwnerID     int64      // 资源所有者ID（可选）
	AccessLevel int        // GitLab访问级别（可选）
}

// CanPerform 检查是否可以执行某个操作
func (ctx *PermissionContext) CanPerform(perm Permission) bool {
	// 管理员拥有所有权限
	if ctx.IsAdmin {
		return true
	}

	// 检查角色权限
	return ctx.Role.HasPermission(perm)
}

// IsOwner 检查是否是资源所有者
func (ctx *PermissionContext) IsOwner() bool {
	return ctx.OwnerID > 0 && ctx.UserID == ctx.OwnerID
}

// HasMinAccessLevel 检查是否满足最小访问级别
func (ctx *PermissionContext) HasMinAccessLevel(minLevel int) bool {
	return ctx.AccessLevel >= minLevel
}
