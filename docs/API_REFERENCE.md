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

## 班级管理

### 创建班级
```http
POST /api/classes
Content-Type: application/json

{
  "name": "数据结构与算法",
  "description": "2024春季学期数据结构与算法课程班级",
  "auto_create_gitlab_group": true
}
```

### 获取班级列表
```http
GET /api/classes?page=1&page_size=20
```

### 获取班级详情
```http
GET /api/classes/{id}
```

### 更新班级信息
```http
PUT /api/classes/{id}
Content-Type: application/json

{
  "name": "更新后的班级名称",
  "description": "更新后的班级描述"
}
```

### 删除班级
```http
DELETE /api/classes/{id}
```

### 学生加入班级
```http
POST /api/classes/join
Content-Type: application/json

{
  "code": "CLASS2024"
}
```

### 添加班级成员
```http
POST /api/classes/{id}/members
Content-Type: application/json

{
  "student_id": 123
}
```

### 移除班级成员
```http
DELETE /api/classes/{id}/members/{user_id}
```

### 获取班级成员列表
```http
GET /api/classes/{id}/members
```

### 获取班级统计信息
```http
GET /api/classes/{id}/stats
```

### 同步班级到GitLab
```http
POST /api/classes/{id}/sync
```

## 课题管理

### 创建课题
```http
POST /api/projects
Content-Type: application/json

{
  "name": "Web开发实战项目",
  "description": "使用现代Web技术栈开发一个完整的Web应用",
  "class_id": 1,
  "start_date": "2024-03-01T00:00:00Z",
  "end_date": "2024-06-30T23:59:59Z",
  "wiki_enabled": true,
  "issues_enabled": true,
  "mr_enabled": true
}
```

### 获取课题列表
```http
GET /api/projects?page=1&page_size=20
GET /api/projects?class_id=1
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
  "code": "PROJ2024"
}
```

### 添加课题成员
```http
POST /api/projects/{id}/members
Content-Type: application/json

{
  "student_id": 123,
  "role": "member"
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

### 获取GitLab项目信息
```http
GET /api/projects/{id}/gitlab
```

## 作业管理

### 创建作业
```http
POST /api/assignments
Content-Type: application/json

{
  "title": "数据结构实验一",
  "description": "实现链表的基本操作",
  "project_id": 1,
  "due_date": "2024-03-31T23:59:59Z",
  "type": "homework",
  "required_files": ["main.c", "list.h", "README.md"],
  "submission_branch": "assignment-1",
  "auto_create_mr": true,
  "require_code_review": true,
  "max_file_size": 10485760,
  "allowed_file_types": ["c", "h", "md", "txt"]
}
```

### 获取作业列表
```http
GET /api/assignments?page=1&page_size=20
GET /api/assignments?project_id=1
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
  "content": "作业完成说明",
  "files": {
    "main.c": "#include <stdio.h>\nint main() { return 0; }",
    "README.md": "# 实验报告\n\n## 实验内容\n..."
  },
  "auto_create_mr": true
}
```

### 获取作业提交列表
```http
GET /api/assignments/{id}/submissions
```

### 获取作业提交详情
```http
GET /api/assignments/submissions/{submission_id}
```

### 获取作业统计信息
```http
GET /api/assignments/{id}/stats
```

### 获取我的提交记录
```http
GET /api/assignments/my-submissions
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

## 分析统计

### 获取分析概览
```http
GET /api/analytics/overview
```

### 获取项目统计
```http
GET /api/analytics/project-stats
```

### 获取学生统计
```http
GET /api/analytics/student-stats
```

### 获取作业统计
```http
GET /api/analytics/assignment-stats
```

### 获取提交趋势
```http
GET /api/analytics/submission-trend
```

### 获取项目分布
```http
GET /api/analytics/project-distribution
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

### 系统信息
```http
GET /
```

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

## 通知类型

- `assignment_submitted` - 作业提交
- `assignment_reviewed` - 作业评审
- `assignment_created` - 作业创建
- `project_joined` - 加入课题
- `class_joined` - 加入班级
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

## OnlyOffice集成特性

- **在线编辑**: 支持Word、Excel、PowerPoint文档的实时协作编辑
- **版本控制**: 文档版本管理和历史记录
- **权限控制**: 基于用户角色的文档编辑权限
- **回调处理**: 完整的OnlyOffice Document Server回调机制
- **文件类型支持**: docx、xlsx、pptx、pdf等多种格式

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

### 班级（Group）管理API

#### 创建班级
```http
POST /api/third-party/groups
Content-Type: application/json

{
  "name": "班级名称",
  "description": "班级描述",
  "code": "CLASS2024"
}
```

#### 获取班级列表
```http
GET /api/third-party/groups?page=1&page_size=20
```

#### 获取班级详情
```http
GET /api/third-party/groups/{id}
```

#### 班级成员管理
```http
POST /api/third-party/groups/{id}/members      # 添加成员
DELETE /api/third-party/groups/{id}/members/{user_id}  # 移除成员
GET /api/third-party/groups/{id}/members       # 获取成员列表
PUT /api/third-party/groups/{id}/members/{user_id}     # 更新成员角色
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

## 第三方API响应格式

### 成功响应
```json
{
  "message": "操作成功",
  "data": {
    "id": 123,
    "name": "资源名称",
    "gitlab_id": 456,
    "repository_url": "https://gitlab.example.com/repo.git",
    "created_at": "2024-03-15T10:00:00Z"
  }
}
```

### 列表响应
```json
{
  "data": [
    {
      "id": 123,
      "name": "资源名称"
    }
  ],
  "total": 100,
  "page": 1,
  "page_size": 20
}
```

### 错误响应
```json
{
  "error": "详细错误描述",
  "details": "技术错误信息"
}
```

## 用户角色定义

- **1 - 管理员**: 系统管理员，拥有所有权限
- **2 - 教师**: 可以创建和管理班级、课题、作业
- **3 - 学生**: 可以参与班级和课题，提交作业
- **4 - 访客**: 只读权限

## 权限动作类型

- **read**: 读取权限
- **write**: 写入权限
- **manage**: 管理权限

## 资源类型

- **class**: 班级资源
- **project**: 项目/课题资源
- **assignment**: 作业资源
- **user**: 用户资源

## GitLab集成说明

第三方API完全基于GitLab进行资源管理：

- 创建仓库 → 自动创建GitLab Project
- 创建班级 → 自动创建GitLab Group
- 用户管理 → 同步GitLab用户权限
- 文件操作 → 直接操作GitLab仓库

## 使用示例

### 1. 创建班级并添加学生
```bash
# 创建班级
curl -X POST /api/third-party/groups \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"name": "软件工程2024", "description": "2024年软件工程班级"}'

# 添加学生到班级
curl -X POST /api/third-party/groups/1/members \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"user_id": 123, "role": "student"}'
```

### 2. 创建项目仓库
```bash
# 创建项目仓库
curl -X POST /api/third-party/repos \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"name": "web-project", "description": "Web开发项目", "init_repo": true}'
```

### 3. 检查用户权限
```bash
# 检查用户权限
curl -X POST /api/third-party/permissions/check \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"user_id": 123, "resource_type": "project", "resource_id": 456, "action": "read"}'
```

## 教育场景优化

- **班级管理**: 基于GitLab Group的班级组织架构
- **课题管理**: 课题即Git仓库的项目管理模式
- **作业流程**: 完整的作业创建、提交、评审流程
- **权限控制**: 教师、学生、助教等教育角色的精细权限管理
- **进度跟踪**: 基于GitLab Activity的学习进度监控
- **统计分析**: 丰富的教育数据统计和分析功能 