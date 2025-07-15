# GitLabEx API 接口文档

## 概述

GitLabEx提供完整的RESTful API接口，支持教育场景下的所有核心功能。所有API都基于JSON格式，并完全集成GitLab的用户体系和权限管理。

## 基础信息

- **Base URL**: `http://127.0.0.1:8000/api`
- **认证方式**: GitLab OAuth 2.0 + JWT Token
- **数据格式**: JSON
- **字符编码**: UTF-8

## 认证接口

### GitLab OAuth登录
```http
GET /api/auth/gitlab
```
重定向到GitLab OAuth认证页面

### OAuth回调处理
```http
GET /api/auth/gitlab/callback?code={code}&state={state}
POST /api/auth/gitlab/callback
```

### 用户登出
```http
POST /api/auth/logout
```

## 用户管理

### 获取当前用户信息
```http
GET /api/users/current
```

**响应示例：**
```json
{
  "message": "success",
  "data": {
    "id": 1,
    "username": "teacher001",
    "name": "张老师",
    "email": "teacher001@example.com",
    "avatar": "https://gitlab.example.com/uploads/user/avatar/1/avatar.png",
    "role": 2,
    "role_name": "教师",
    "dynamic_role": "teacher",
    "dynamic_role_name": "教师",
    "gitlab_id": 123,
    "is_active": true,
    "created_at": "2024-03-15T10:00:00Z"
  }
}
```

### 获取用户仪表板
```http
GET /api/users/dashboard
```

### 获取活跃用户列表
```http
GET /api/users/active
```

### 根据ID获取用户信息
```http
GET /api/users/{id}
```

### 更新当前用户信息
```http
PUT /api/users/current
Content-Type: application/json

{
  "name": "更新后的姓名",
  "avatar": "https://example.com/avatar.jpg"
}
```

### 同步GitLab用户信息
```http
POST /api/users/sync/{gitlab_id}
```

## 权限管理

权限管理完全基于GitLab用户组和权限系统，提供教育场景的角色映射。

### 获取用户权限信息
```http
GET /api/permissions/user/{user_id}
```

**响应示例：**
```json
{
  "message": "success", 
  "data": {
    "user_id": 123,
    "static_role": "teacher",
    "dynamic_role": "teacher",
    "effective_role": "teacher",
    "permissions": ["project_create", "assignment_manage", "student_view"],
    "gitlab_access_level": "maintainer"
  }
}
```

### 检查用户权限
```http
POST /api/permissions/check
Content-Type: application/json

{
  "user_id": 123,
  "resource_type": "project",
  "resource_id": 456,
  "action": "read"
}
```

**响应示例：**
```json
{
  "message": "success",
  "data": {
    "allowed": true,
    "reason": "User has teacher role with project access"
  }
}
```

### 获取角色列表
```http
GET /api/permissions/roles
```

**响应示例：**
```json
{
  "message": "success",
  "data": [
    {
      "role": "admin",
      "name": "管理员",
      "level": 50,
      "description": "系统管理员，拥有所有权限"
    },
    {
      "role": "teacher", 
      "name": "教师",
      "level": 40,
      "description": "可以创建和管理课题、作业"
    },
    {
      "role": "student",
      "name": "学生", 
      "level": 20,
      "description": "可以参与课题，提交作业"
    }
  ]
}
```

## 课题管理

课题管理已简化为教师直接创建和管理课题，学生通过课题代码加入课题。

### 创建课题
```http
POST /api/projects
Content-Type: application/json

{
  "name": "Web开发实战项目",
  "description": "使用现代Web技术栈开发一个完整的Web应用",
  "start_date": "2024-03-01T00:00:00Z",
  "end_date": "2024-06-30T23:59:59Z",
  "max_members": 30,
  "wiki_enabled": true,
  "issues_enabled": true,
  "mr_enabled": true
}
```

**响应示例：**
```json
{
  "message": "Project created successfully",
  "data": {
    "id": 1,
    "name": "Web开发实战项目",
    "description": "使用现代Web技术栈开发一个完整的Web应用",
    "teacher_id": 123,
    "teacher_name": "张老师",
    "project_code": "PROJ2024ABC",
    "gitlab_project_id": 456,
    "repository_url": "https://gitlab.example.com/education/web-project.git",
    "start_date": "2024-03-01T00:00:00Z",
    "end_date": "2024-06-30T23:59:59Z",
    "status": "active",
    "current_members": 0,
    "max_members": 30,
    "created_at": "2024-03-15T10:00:00Z"
  }
}
```

### 获取课题列表
```http
GET /api/projects?page=1&page_size=20
GET /api/projects?teacher_id=123
```

### 获取课题详情
```http
GET /api/projects/{id}
```

### 更新课题信息
```http
PUT /api/projects/{id}
Content-Type: application/json

{
  "name": "更新后的课题名称",
  "description": "更新后的课题描述",
  "max_members": 35,
  "status": "active"
}
```

### 删除课题
```http
DELETE /api/projects/{id}
```

### 学生加入课题
```http
POST /api/projects/join
Content-Type: application/json

{
  "code": "PROJ2024ABC"
}
```

**响应示例：**
```json
{
  "message": "Successfully joined project",
  "data": {
    "project_id": 1,
    "project_name": "Web开发实战项目",
    "teacher_name": "张老师",
    "student_branch": "student-zhangsan-20240315",
    "gitlab_access_token": "glpat-xxxxxxxxxxxx"
  }
}
```

### 添加课题成员
```http
POST /api/projects/{id}/members
Content-Type: application/json

{
  "student_id": 123,
  "role": "developer"
}
```

### 移除课题成员
```http
DELETE /api/projects/{id}/members/{user_id}
```

### 获取课题成员列表
```http
GET /api/projects/{id}/members
```

### 获取课题统计信息
```http
GET /api/projects/{id}/stats
```

**响应示例：**
```json
{
  "message": "success",
  "data": {
    "project_id": 1,
    "total_members": 25,
    "total_assignments": 8,
    "completed_assignments": 6,
    "completion_rate": 75.0,
    "average_score": 85.5,
    "recent_activities": [
      {
        "type": "assignment_submitted",
        "student_name": "李同学",
        "assignment_title": "前端页面开发",
        "created_at": "2024-03-15T14:30:00Z"
      }
    ]
  }
}
```

### 获取GitLab项目信息
```http
GET /api/projects/{id}/gitlab
```

## 作业管理

作业管理系统增强，支持教师管理所有课题作业，学生提交和查看个人作业。

### 创建作业
```http
POST /api/assignments
Content-Type: application/json

{
  "title": "前端页面开发",
  "description": "使用Vue.js开发用户注册登录页面",
  "project_id": 1,
  "due_date": "2024-03-31T23:59:59Z",
  "type": "homework",
  "required_files": ["src/views/Login.vue", "src/views/Register.vue", "README.md"],
  "submission_format": "git_commit",
  "max_score": 100,
  "auto_create_mr": true,
  "grading_criteria": {
    "functionality": 40,
    "code_quality": 30,
    "documentation": 20,
    "ui_design": 10
  }
}
```

**响应示例：**
```json
{
  "message": "Assignment created successfully",
  "data": {
    "id": 1,
    "title": "前端页面开发",
    "description": "使用Vue.js开发用户注册登录页面",
    "project_id": 1,
    "project_name": "Web开发实战项目",
    "teacher_id": 123,
    "teacher_name": "张老师",
    "due_date": "2024-03-31T23:59:59Z",
    "type": "homework",
    "status": "active",
    "max_score": 100,
    "submission_count": 0,
    "created_at": "2024-03-15T10:00:00Z"
  }
}
```

### 获取作业列表
```http
GET /api/assignments?page=1&page_size=20
GET /api/assignments?project_id=1
GET /api/assignments?teacher_id=123
```

**响应示例：**
```json
{
  "data": [
    {
      "id": 1,
      "title": "前端页面开发",
      "project_name": "Web开发实战项目",
      "due_date": "2024-03-31T23:59:59Z",
      "status": "active",
      "submission_count": 15,
      "max_score": 100,
      "average_score": 82.5
    }
  ],
  "total": 8,
  "page": 1,
  "page_size": 20
}
```

### 获取作业详情
```http
GET /api/assignments/{id}
```

### 更新作业信息
```http
PUT /api/assignments/{id}
Content-Type: application/json

{
  "title": "更新后的作业标题",
  "description": "更新后的作业描述",
  "due_date": "2024-04-15T23:59:59Z",
  "status": "active"
}
```

### 删除作业
```http
DELETE /api/assignments/{id}
```

### 提交作业
```http
POST /api/assignments/{id}/submit
Content-Type: application/json

{
  "submission_content": "作业完成说明，包含实现的功能点和遇到的问题",
  "commit_hash": "abc123def456",
  "files": {
    "src/views/Login.vue": "Vue组件代码内容",
    "src/views/Register.vue": "Vue组件代码内容", 
    "README.md": "# 实验报告\n\n## 实现功能\n..."
  },
  "branch_name": "student-zhangsan-20240315"
}
```

**响应示例：**
```json
{
  "message": "Assignment submitted successfully",
  "data": {
    "submission_id": 101,
    "assignment_id": 1,
    "student_id": 456,
    "student_name": "张同学",
    "submitted_at": "2024-03-25T16:30:00Z",
    "status": "submitted",
    "commit_hash": "abc123def456",
    "gitlab_mr_url": "https://gitlab.example.com/education/web-project/-/merge_requests/15"
  }
}
```

### 获取作业提交列表
```http
GET /api/assignments/{id}/submissions
```

**响应示例：**
```json
{
  "message": "success",
  "data": [
    {
      "submission_id": 101,
      "student_id": 456,
      "student_name": "张同学",
      "submitted_at": "2024-03-25T16:30:00Z",
      "status": "reviewed",
      "score": 85,
      "review_status": "completed"
    }
  ],
  "total": 15
}
```

### 获取作业提交详情
```http
GET /api/assignments/submissions/{submission_id}
```

### 评审作业
```http
PUT /api/assignments/submissions/{submission_id}/review
Content-Type: application/json

{
  "score": 85,
  "review_report": {
    "code_quality_score": 80,
    "code_quality_comment": "代码结构清晰，变量命名规范，但缺少部分注释",
    "functionality_score": 90,
    "functionality_comment": "功能实现完整，用户体验良好",
    "documentation_score": 75,
    "documentation_comment": "README文档详细，但缺少API接口说明",
    "ui_design_score": 85,
    "ui_design_comment": "界面设计美观，响应式布局良好"
  },
  "general_comment": "整体完成质量很好，建议加强代码注释和API文档",
  "suggestions": [
    "添加详细的函数注释",
    "完善错误处理机制",
    "添加单元测试"
  ]
}
```

### 获取作业统计信息
```http
GET /api/assignments/{id}/stats
```

### 获取我的提交记录
```http
GET /api/assignments/my-submissions
```

**响应示例：**
```json
{
  "message": "success",
  "data": [
    {
      "assignment_id": 1,
      "assignment_title": "前端页面开发",
      "project_name": "Web开发实战项目",
      "submitted_at": "2024-03-25T16:30:00Z",
      "status": "reviewed",
      "score": 85,
      "due_date": "2024-03-31T23:59:59Z",
      "is_late": false
    }
  ]
}
```

## 数据统计

数据统计系统提供教师和学生不同的统计视图，基于权限显示相应数据。

### 获取教师统计概览
```http
GET /api/analytics/teacher/overview
```

**响应示例：**
```json
{
  "message": "success",
  "data": {
    "total_projects": 5,
    "active_projects": 3,
    "total_assignments": 25,
    "total_submissions": 180,
    "pending_reviews": 12,
    "total_students": 95,
    "average_score": 83.5,
    "completion_rate": 78.2
  }
}
```

### 获取教师课题统计
```http
GET /api/analytics/teacher/projects
```

### 获取教师作业统计
```http
GET /api/analytics/teacher/assignments
```

### 获取学生统计概览
```http
GET /api/analytics/student/overview
```

**响应示例：**
```json
{
  "message": "success",
  "data": {
    "joined_projects": 3,
    "active_assignments": 5,
    "completed_assignments": 12,
    "pending_assignments": 3,
    "total_submissions": 15,
    "average_score": 85.2,
    "highest_score": 95
  }
}
```

### 获取学生作业统计
```http
GET /api/analytics/student/assignments
```

### 获取学生学习进度
```http
GET /api/analytics/student/progress
```

### 获取管理员统计概览
```http
GET /api/analytics/overview
```

### 获取项目统计
```http
GET /api/analytics/project-stats
```

### 获取提交趋势
```http
GET /api/analytics/submission-trend
```

### 获取成绩分布
```http
GET /api/analytics/grade-distribution
```

### 获取活动统计
```http
GET /api/analytics/activity-stats
```

### 获取仪表板统计
```http
GET /api/analytics/dashboard-stats
```

### 获取最近活动
```http
GET /api/analytics/recent-activities
```

## 话题讨论

### 创建话题
```http
POST /api/discussions
Content-Type: application/json

{
  "title": "关于项目架构的讨论",
  "content": "我们需要讨论一下项目的整体架构设计...",
  "project_id": 1,
  "category": "general",
  "tags": "架构,设计,讨论",
  "is_public": true
}
```

### 获取话题列表
```http
GET /api/discussions?project_id=1&page=1&page_size=20
GET /api/discussions?category=question&status=open
```

### 获取话题详情
```http
GET /api/discussions/{id}
```

### 更新话题
```http
PUT /api/discussions/{id}
Content-Type: application/json

{
  "title": "更新后的话题标题",
  "content": "更新后的话题内容",
  "category": "announcement",
  "is_public": false
}
```

### 删除话题
```http
DELETE /api/discussions/{id}
```

### 创建回复
```http
POST /api/discussions/{id}/replies
Content-Type: application/json

{
  "content": "这是一个回复内容",
  "parent_reply_id": 0
}
```

### 点赞话题
```http
POST /api/discussions/{id}/like
```

### 取消点赞
```http
DELETE /api/discussions/{id}/like
```

### 置顶话题
```http
POST /api/discussions/{id}/pin
```

### 获取话题分类
```http
GET /api/discussions/categories
```

### 同步GitLab话题
```http
POST /api/discussions/sync/{project_id}
```

## 通知管理

### 获取通知列表
```http
GET /api/notifications?page=1&page_size=20
```

### 获取未读通知
```http
GET /api/notifications/unread
```

### 获取未读通知数量
```http
GET /api/notifications/count
```

### 获取通知统计信息
```http
GET /api/notifications/stats
```

### 标记通知为已读
```http
PUT /api/notifications/{id}/read
```

### 标记所有通知为已读
```http
PUT /api/notifications/read-all
```

### 删除通知
```http
DELETE /api/notifications/{id}
```

### 删除所有通知
```http
DELETE /api/notifications/all
```

### 创建通知（管理员）
```http
POST /api/notifications
Content-Type: application/json

{
  "user_id": 123,
  "title": "系统维护通知",
  "content": "系统将于今晚进行维护，请注意保存工作",
  "type": "system",
  "target_type": "system",
  "target_id": 0
}
```

### 按类型获取通知
```http
GET /api/notifications/types/{type}
```

## 教育管理

### 获取教育仪表板
```http
GET /api/education/dashboard
```

### 获取教育统计
```http
GET /api/education/stats
```

### 获取推荐课题
```http
GET /api/education/recommendations
```

## Wiki和文档管理

### 获取Wiki页面列表
```http
GET /api/wiki/{project_id}/pages
```

### 创建Wiki页面
```http
POST /api/wiki/{project_id}/pages
Content-Type: application/json

{
  "title": "项目说明",
  "content": "# 项目概述\n\n这是一个Web开发项目..."
}
```

### 获取Wiki页面详情
```http
GET /api/wiki/{project_id}/pages/{slug}
```

### 更新Wiki页面
```http
PUT /api/wiki/{project_id}/pages/{slug}
Content-Type: application/json

{
  "title": "更新后的标题",
  "content": "更新后的内容..."
}
```

### 上传文档到OnlyOffice
```http
POST /api/documents/upload
Content-Type: multipart/form-data

file: [文件内容]
mode: edit
```

### 获取文档编辑器
```http
GET /api/documents/{id}/editor
```

### 获取文档配置
```http
GET /api/documents/{id}/config
```

### 获取文档内容
```http
GET /api/documents/{id}/content
```

### OnlyOffice回调处理
```http
POST /api/documents/{id}/callback
```

## 系统接口

### 健康检查
```http
GET /api/health
```

**响应示例：**
```json
{
  "status": "ok",
  "service": "gitlabex-backend",
  "version": "1.0.0",
  "timestamp": 1710505200
}
```

### 系统信息
```http
GET /
```

**响应示例：**
```json
{
  "message": "GitLabEx API Server",
  "version": "1.0.0",
  "status": "running"
}
```

## 第三方系统API

专为第三方系统调用设计的API接口，支持外部系统集成GitLabEx的核心功能。

### Git仓库管理API

#### 创建Git仓库
```http
POST /api/third-party/repos
Content-Type: application/json

{
  "name": "项目名称",
  "description": "项目描述",
  "visibility": "private",
  "init_repo": true
}
```

#### 获取仓库列表
```http
GET /api/third-party/repos?page=1&page_size=20
```

#### 获取仓库详情
```http
GET /api/third-party/repos/{id}
```

#### 克隆仓库
```http
POST /api/third-party/repos/{id}/clone
Content-Type: application/json

{
  "target_path": "/path/to/clone",
  "branch": "main"
}
```

#### 获取提交记录
```http
GET /api/third-party/repos/{id}/commits
```

#### 创建提交
```http
POST /api/third-party/repos/{id}/commits
Content-Type: application/json

{
  "message": "提交信息",
  "files": {
    "file1.txt": "文件内容"
  },
  "branch": "main"
}
```

#### 获取分支列表
```http
GET /api/third-party/repos/{id}/branches
```

#### 创建分支
```http
POST /api/third-party/repos/{id}/branches
Content-Type: application/json

{
  "name": "feature-branch",
  "from": "main"
}
```

#### 文件管理
```http
GET /api/third-party/repos/{id}/files          # 获取文件列表
POST /api/third-party/repos/{id}/files         # 上传文件
GET /api/third-party/repos/{id}/files/{path}   # 获取文件内容
PUT /api/third-party/repos/{id}/files/{path}   # 更新文件内容
```

### 用户管理API

#### 创建用户
```http
POST /api/third-party/users
Content-Type: application/json

{
  "username": "student001",
  "email": "student001@example.com",
  "name": "学生姓名",
  "role": 3
}
```

#### 获取用户列表
```http
GET /api/third-party/users?page=1&page_size=20&role=3
```

#### 用户管理操作
```http
GET /api/third-party/users/{id}                # 获取用户详情
PUT /api/third-party/users/{id}                # 更新用户信息
DELETE /api/third-party/users/{id}             # 删除用户
POST /api/third-party/users/{id}/sync          # 同步GitLab用户
PUT /api/third-party/users/{id}/role           # 更新用户角色
GET /api/third-party/users/{id}/permissions    # 获取用户权限
```

### 权限管理API

#### 获取所有角色
```http
GET /api/third-party/permissions/roles
```

#### 检查权限
```http
POST /api/third-party/permissions/check
Content-Type: application/json

{
  "user_id": 123,
  "resource_type": "project",
  "resource_id": 456,
  "action": "read"
}
```

#### 权限授予和撤销
```http
POST /api/third-party/permissions/grant        # 授予权限
POST /api/third-party/permissions/revoke       # 撤销权限
```

### 项目管理API

#### 项目操作
```http
POST /api/third-party/projects                 # 创建项目
GET /api/third-party/projects                  # 获取项目列表
GET /api/third-party/projects/{id}             # 获取项目详情
PUT /api/third-party/projects/{id}             # 更新项目信息
DELETE /api/third-party/projects/{id}          # 删除项目
POST /api/third-party/projects/{id}/members    # 添加项目成员
GET /api/third-party/projects/{id}/assignments # 获取项目作业
```

## 第三方API认证与安全

### 🔐 强制OAuth认证

所有第三方API都受到严格的OAuth认证保护，**必须提供有效的认证令牌**：

#### 1. **API Key认证**（推荐用于第三方系统）
```http
Authorization: Bearer YOUR_API_KEY
```

#### 2. **JWT Token认证**（用于Web应用）
```http
Authorization: Bearer YOUR_JWT_TOKEN
```

### 🛡️ 安全特性

- **强制认证**: 所有第三方API端点都需要认证
- **角色权限控制**: 基于用户角色进行精细权限管理
- **API访问日志**: 完整的第三方API调用日志记录
- **跨域保护**: 严格的CORS策略，只允许授权域名
- **请求限流**: 防止API滥用的限流机制
- **Token过期**: API Key 7天有效期，确保安全性

### 📋 获取API Key

#### 生成API Key
```http
POST /api/third-party/auth/api-key
Authorization: Bearer YOUR_JWT_TOKEN
```

响应：
```json
{
  "message": "API Key generated successfully",
  "data": {
    "api_key": "1.1625097600.a1b2c3d4e5f6...",
    "user_id": 123,
    "expires_in": "7 days",
    "scopes": ["read", "write", "manage"]
  }
}
```

#### 验证Token
```http
GET /api/third-party/auth/validate
Authorization: Bearer YOUR_API_KEY
```

响应：
```json
{
  "valid": true,
  "data": {
    "user_id": 123,
    "username": "user001",
    "role": 2,
    "auth_type": "api_key"
  }
}
```

### ⚠️ 重要安全提示

1. **保护API Key**: 
   - 不要在客户端代码中硬编码API Key
   - 使用环境变量存储API Key
   - 定期轮换API Key

2. **网络安全**:
   - 只在HTTPS环境下使用API
   - 配置合适的防火墙规则

3. **权限最小化**:
   - 为不同用途创建不同角色的用户
   - 避免使用管理员权限调用第三方API

### 🔄 API代理架构

第三方API采用**代理模式**，避免重复开发：
- 复用现有的内部API逻辑
- 统一的认证和权限控制
- 标准化的响应格式
- 完整的日志和监控

## 响应格式

### 成功响应
```json
{
  "message": "操作成功",
  "data": {
    // 具体数据
  }
}
```

### 列表响应
```json
{
  "data": [
    // 数据列表
  ],
  "total": 100,
  "page": 1,
  "page_size": 20
}
```

### 错误响应
```json
{
  "error": "错误描述",
  "details": "详细错误信息"
}
```

## 状态码

- `200` - 请求成功
- `201` - 创建成功
- `400` - 请求参数错误
- `401` - 未授权（需要登录）
- `403` - 无权限访问
- `404` - 资源不存在
- `500` - 服务器内部错误

## 用户角色定义

基于GitLab权限的教育角色映射：

- **admin (50)** - 管理员: 系统管理员，拥有所有权限，对应GitLab Owner
- **teacher (40)** - 教师: 可以创建和管理课题、作业，对应GitLab Maintainer
- **assistant (30)** - 助教: 协助教师管理课题，对应GitLab Developer
- **student (20)** - 学生: 可以参与课题，提交作业，对应GitLab Reporter
- **guest (10)** - 访客: 只读权限，对应GitLab Guest

## 通知类型

- `assignment_submitted` - 作业提交
- `assignment_reviewed` - 作业评审
- `assignment_created` - 作业创建
- `project_joined` - 加入课题
- `project_created` - 课题创建
- `gitlab_commit` - GitLab提交
- `merge_request` - 合并请求
- `issue_created` - Issue创建
- `wiki_created` - Wiki创建
- `assignment_due` - 作业截止提醒
- `code_review` - 代码审查
- `gitlab_activity` - GitLab活动

## GitLab集成特性

- **自动仓库创建**: 创建课题时自动创建GitLab项目仓库
- **分支管理**: 学生加入课题时自动创建个人分支
- **权限同步**: 教育角色与GitLab权限级别自动映射
- **作业提交**: 基于GitLab Commits的作业提交流程
- **代码审查**: 集成GitLab Merge Request的代码审查功能
- **Activity监控**: 通过GitLab Webhook实时监控项目活动
- **Wiki集成**: 完全基于GitLab Wiki的文档管理
- **Issues集成**: 支持GitLab Issues的讨论和问题跟踪
- **动态权限**: 实时从GitLab获取用户权限信息

## OnlyOffice集成特性

- **在线编辑**: 支持Word、Excel、PowerPoint文档的实时协作编辑
- **版本控制**: 文档版本管理和历史记录
- **权限控制**: 基于用户角色的文档编辑权限
- **回调处理**: 完整的OnlyOffice Document Server回调机制
- **文件类型支持**: docx、xlsx、pptx、pdf等多种格式

## 系统架构变更说明

### 移除功能
- **班级管理**: 完全移除班级概念，简化为教师直接管理课题
- **基于班级的权限控制**: 改为基于GitLab的权限管理

### 新增功能
- **动态角色获取**: 用户角色从GitLab实时获取，不再前端硬编码
- **增强的作业管理**: 详细的评审报告系统，多维度评分
- **权限管理界面**: 基于GitLab的权限管理功能
- **简化的课题流程**: 学生通过课题代码直接加入课题

### 架构优化
- **服务V2版本**: ProjectServiceV2、AssignmentServiceV2、UserServiceV2
- **权限集成**: 完全基于GitLab权限系统
- **数据统计**: 基于角色的统计视图（教师视图、学生视图）
- **API简化**: 移除班级相关API，简化课题管理API

## 使用示例

### 1. 教师创建课题并管理作业
```bash
# 1. 创建课题
curl -X POST /api/projects \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Java Web开发", "description": "Spring Boot项目实战"}'

# 2. 创建作业
curl -X POST /api/assignments \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "用户管理模块", "project_id": 1, "due_date": "2024-04-30T23:59:59Z"}'

# 3. 查看提交
curl -X GET /api/assignments/1/submissions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 2. 学生加入课题并提交作业
```bash
# 1. 加入课题
curl -X POST /api/projects/join \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"code": "PROJ2024ABC"}'

# 2. 提交作业
curl -X POST /api/assignments/1/submit \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"submission_content": "完成用户注册登录功能", "commit_hash": "abc123"}'

# 3. 查看我的提交
curl -X GET /api/assignments/my-submissions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 3. 权限检查和统计查询
```bash
# 1. 检查权限
curl -X POST /api/permissions/check \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_id": 123, "resource_type": "project", "resource_id": 1, "action": "read"}'

# 2. 获取教师统计
curl -X GET /api/analytics/teacher/overview \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 3. 获取学生统计
curl -X GET /api/analytics/student/overview \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 教育场景优化

- **简化的课题管理**: 教师直接创建课题，学生通过代码加入
- **基于GitLab的权限**: 完全依赖GitLab的用户和权限管理
- **增强的作业系统**: 详细的评审报告和多维度评分
- **动态角色系统**: 用户角色实时从GitLab获取
- **权限管理界面**: 提供GitLab权限的教育化界面
- **统计分析优化**: 基于角色的数据统计和分析功能 