# GitLabEx 权限系统说明文档

## 概述

GitLabEx 的权限系统基于 GitLab 的角色和访问级别,实现了细粒度的资源访问控制。权限系统遵循最小权限原则,确保用户只能访问其被授权的资源。

## 角色定义

系统定义了以下角色,每个角色对应 GitLab 的访问级别:

| 角色 | GitLab 级别 | 访问级别值 | 教育角色名称 | 描述 |
|------|------------|----------|------------|------|
| Owner | Owner | 50 | 管理员 | 拥有所有权限 |
| Maintainer | Maintainer | 40 | 教师 | 可以管理课题、作业、文档 |
| Developer | Developer | 30 | 研究员 | 可以参与开发和批改作业 |
| Reporter | Reporter | 20 | 学生/普通用户 | 可以查看和提交作业 |
| Guest | Guest | 10 | 访客 | 只读权限 |

## 权限矩阵

### 1. 管理员权限（Owner, Level 50）

管理员拥有系统的所有权限:

- ✅ 新建、编辑、删除课题
- ✅ 新建、编辑、删除课题包含的话题
- ✅ 新建、编辑、删除作业
- ✅ 批改课题作业
- ✅ 新建、编辑、同步、审核文档
- ✅ 管理用户
- ✅ 点赞、评论话题

### 2. 教师权限（Maintainer, Level 40）

教师可以管理课题和教学内容:

**课题管理:**
- ✅ 新建课题
- ✅ 编辑课题
- ✅ 删除课题
- ✅ 管理课题成员

**话题管理:**
- ✅ 新建话题
- ✅ 编辑课题包含的所有话题
- ✅ 删除课题包含的所有话题
- ✅ 点赞、评论话题

**作业管理:**
- ✅ 新建作业
- ✅ 编辑作业
- ✅ 删除作业
- ✅ 批改课题作业
- ✅ 查看所有学生提交

**文档管理:**
- ✅ 新建文档
- ✅ 编辑文档
- ✅ 同步文档
- ✅ 审核文档
- ✅ 删除文档

### 3. 研究员权限（Developer, Level 30）

研究员可以参与课题开发和协助教学:

**课题管理:**
- ✅ 查看所属课题
- ❌ 创建、编辑、删除课题

**话题管理:**
- ✅ 新建话题
- ✅ 编辑自己的话题
- ❌ 编辑他人的话题
- ❌ 删除话题（只能删除自己的）
- ✅ 点赞、评论话题

**作业管理:**
- ✅ 查看课题作业
- ✅ 批改课题作业
- ❌ 创建、编辑、删除作业
- ❌ 提交作业

**文档管理:**
- ✅ 新建文档
- ✅ 编辑自己的文档
- ❌ 编辑他人的文档
- ❌ 同步、审核文档

### 4. 普通用户/学生权限（Reporter, Level 20）

普通用户可以参与学习和讨论:

**课题管理:**
- ✅ 查看所属课题
- ✅ 查看课题包含的话题和作业
- ❌ 创建、编辑、删除课题

**话题管理:**
- ✅ 新建话题
- ✅ 编辑自己的话题
- ❌ 编辑他人的话题
- ❌ 删除话题（只能删除自己的）
- ✅ 点赞、评论话题

**作业管理:**
- ✅ 查看作业
- ✅ 提交作业
- ✅ 查看自己的作业和评语
- ❌ 批改作业
- ❌ 创建、编辑、删除作业

**文档管理:**
- ✅ 新建文档
- ✅ 编辑自己的文档（需要审核）
- ❌ 编辑他人的文档
- ❌ 同步、审核、删除文档

### 5. 访客权限（Guest, Level 10）

访客只有查看权限:

**课题管理:**
- ✅ 查看公开课题列表
- ❌ 查看专有课题
- ❌ 创建、编辑、删除课题

**话题管理:**
- ✅ 查看话题列表
- ✅ 查看话题详情
- ❌ 创建、编辑、删除话题
- ❌ 点赞、评论

**作业管理:**
- ❌ 所有作业相关操作

**文档管理:**
- ✅ 查看文档列表
- ✅ 下载文档
- ❌ 创建、编辑、删除文档

## 资源权限定义

### 课题权限（Project Permissions）

```go
const (
    PermProjectCreate        Permission = "project:create"         // 创建课题
    PermProjectRead          Permission = "project:read"           // 查看课题
    PermProjectUpdate        Permission = "project:update"         // 编辑课题
    PermProjectDelete        Permission = "project:delete"         // 删除课题
    PermProjectManageMembers Permission = "project:manage:members" // 管理课题成员
)
```

### 话题权限（Topic Permissions）

```go
const (
    PermTopicCreate  Permission = "topic:create"  // 创建话题
    PermTopicRead    Permission = "topic:read"    // 查看话题
    PermTopicUpdate  Permission = "topic:update"  // 编辑话题
    PermTopicDelete  Permission = "topic:delete"  // 删除话题
    PermTopicLike    Permission = "topic:like"    // 点赞话题
    PermTopicComment Permission = "topic:comment" // 评论话题
)
```

### 作业权限（Homework Permissions）

```go
const (
    PermHomeworkCreate Permission = "homework:create" // 创建作业
    PermHomeworkRead   Permission = "homework:read"   // 查看作业
    PermHomeworkUpdate Permission = "homework:update" // 编辑作业
    PermHomeworkDelete Permission = "homework:delete" // 删除作业
    PermHomeworkSubmit Permission = "homework:submit" // 提交作业
    PermHomeworkGrade  Permission = "homework:grade"  // 批改作业
)
```

### 文档权限（Document Permissions）

```go
const (
    PermDocumentCreate  Permission = "document:create"  // 创建文档
    PermDocumentRead    Permission = "document:read"    // 查看文档
    PermDocumentUpdate  Permission = "document:update"  // 编辑文档
    PermDocumentDelete  Permission = "document:delete"  // 删除文档
    PermDocumentApprove Permission = "document:approve" // 审核文档
    PermDocumentSync    Permission = "document:sync"    // 同步文档
)
```

## 权限检查流程

### 1. 权限上下文构建

```go
type PermissionContext struct {
    UserID      int64      // 用户ID
    Role        GitLabRole // 用户角色
    IsAdmin     bool       // 是否是管理员
    ProjectID   *int64     // 项目ID（可选）
    ResourceID  string     // 资源ID（可选）
    OwnerID     int64      // 资源所有者ID（可选）
    AccessLevel int        // GitLab访问级别（可选）
}
```

### 2. 权限检查逻辑

1. **管理员检查**: 如果用户是管理员,直接授予所有权限
2. **角色权限检查**: 根据用户在课题中的角色,检查是否有对应的权限
3. **所有者检查**: 检查用户是否为资源的所有者
4. **特殊规则检查**: 应用特定的业务规则(如研究员只能编辑自己的话题)

### 3. 权限检查示例

```go
// 检查用户是否可以编辑话题
func checkTopicEditPermission(ctx *PermissionContext, topic *Topic) (bool, string) {
    // 1. 管理员可以编辑任何话题
    if ctx.IsAdmin {
        return true, "管理员权限"
    }
    
    // 2. 教师可以编辑课题包含的话题
    if ctx.Role == GitLabMaintainer && topic.ProjectID != nil {
        // 检查用户是否为课题的教师
        return true, "课题教师"
    }
    
    // 3. 研究员可以编辑自己的话题
    if ctx.Role == GitLabDeveloper && topic.AuthorID == ctx.UserID {
        return true, "话题作者（研究员）"
    }
    
    // 4. 话题作者可以编辑自己的话题
    if topic.AuthorID == ctx.UserID {
        return true, "话题作者"
    }
    
    return false, "无权编辑该话题"
}
```

## API 使用示例

### 1. 检查用户权限

```bash
POST /api/v1/permissions/check
Authorization: Bearer <token>
Content-Type: application/json

{
  "resource": "topic",
  "action": "update",
  "resource_id": "uuid-of-topic"
}

# 响应
{
  "allowed": true,
  "reason": "话题作者"
}
```

### 2. 获取用户在课题中的权限

```bash
GET /api/v1/permissions/projects/:id?detailed=true
Authorization: Bearer <token>

# 响应
{
  "allowed": true,
  "permissions": {
    "read": true,
    "edit": true,
    "manage": true,
    "create_topic": true,
    "create_homework": true,
    "grade_homework": true
  },
  "roles": ["maintainer", "teacher"],
  "access_level": 40,
  "reason": "权限检查完成"
}
```

### 3. 获取用户的全局权限

```bash
GET /api/v1/permissions/user
Authorization: Bearer <token>

# 响应
{
  "permissions": {
    "can_create_project": true,
    "can_manage_users": false,
    "can_view_admin": false,
    "can_create_topic": true,
    "can_upload_document": true
  },
  "user_id": 123,
  "is_admin": false
}
```

## 中间件使用

### 1. 基于角色的中间件

```go
// 要求教师权限
router.POST("/api/v1/research-projects", 
    middleware.RequireAuth(cfg),
    middleware.RequireProjectMaintainer(gitlabService),
    researchHandler.CreateResearchProject,
)

// 要求研究员权限
router.PUT("/api/v1/homework/:id/grade",
    middleware.RequireAuth(cfg),
    middleware.RequireProjectDeveloper(gitlabService),
    homeworkHandler.GradeHomework,
)
```

### 2. 基于权限的中间件

```go
// 要求特定权限
router.POST("/api/v1/documents/sync/:project_id",
    middleware.RequireAuth(cfg),
    middleware.RequirePermission(models.PermDocumentSync),
    documentHandler.SyncDocuments,
)
```

## 最佳实践

### 1. 权限检查原则

- **最小权限原则**: 默认拒绝,显式授权
- **深度防御**: 在多个层次进行权限检查(中间件 + Handler + Service)
- **清晰的错误信息**: 明确告知用户为何被拒绝

### 2. 代码规范

```go
// ✅ 好的实践: 安全的类型断言
if adminValue, ok := isAdmin.(bool); ok {
    userIsAdmin = adminValue
}

// ❌ 不好的实践: 不安全的类型断言
userIsAdmin := isAdmin.(bool)  // 可能panic
```

### 3. 权限缓存

对于频繁的权限检查,建议使用 Redis 缓存:

```go
// 缓存用户在项目中的访问级别
cacheKey := fmt.Sprintf("user:%d:project:%d:access_level", userID, projectID)
redis.Set(cacheKey, accessLevel, 15*time.Minute)
```

## 测试建议

### 1. 单元测试

为每个权限检查函数编写单元测试,覆盖以下场景:

- 管理员访问
- 教师访问
- 研究员访问
- 普通用户访问
- 访客访问
- 无权限访问
- 资源所有者访问

### 2. 集成测试

测试完整的权限检查流程:

- 创建不同角色的用户
- 测试各种资源的 CRUD 操作
- 验证权限边界情况

## 故障排查

### 常见问题

1. **权限被拒绝但应该有权限**
   - 检查用户在 GitLab 项目中的角色
   - 验证项目ID是否正确
   - 查看权限检查日志

2. **类型断言 panic**
   - 使用安全的类型断言: `value, ok := x.(Type)`
   - 添加 nil 检查

3. **缓存不一致**
   - 在权限变更时清除相关缓存
   - 设置合理的缓存过期时间

## 更新日志

- **2024-12-06**: 重构权限系统,明确角色权限映射
- **2024-12-06**: 修复不安全的类型断言
- **2024-12-06**: 添加权限常量定义

## 参考资料

- [GitLab Permissions](https://docs.gitlab.com/ee/user/permissions.html)
- [RBAC (Role-Based Access Control)](https://en.wikipedia.org/wiki/Role-based_access_control)
