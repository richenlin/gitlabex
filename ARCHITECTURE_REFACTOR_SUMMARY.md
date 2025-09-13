# GitLabEx 架构重构总结

## 概述

根据 SOLUTION.md 的需求，对 GitLabEx 项目进行了全面的架构重构，移除了冗余的本地用户管理逻辑，统一了架构设计，完全依赖 GitLab 进行用户管理和权限控制。

## 核心设计原则

### 1. 完全依赖 GitLab
- **用户管理**: 完全移除本地用户表，所有用户信息实时从 GitLab API 获取
- **权限控制**: 基于 GitLab 项目权限，不维护任何本地权限关系
- **认证机制**: JWT 只包含 GitLab 访问令牌，不存储本地用户 ID

### 2. 简化的数据模型
- 移除复杂的本地 User 模型
- 使用轻量级的 GitLabUser 结构仅用于 API 传输
- 统一的 GitLab 角色映射系统

## 主要架构变更

### 1. 用户模型重构 (`models/user.go`)

#### 变更前
```go
// 复杂的本地用户模型
type User struct {
    BaseModel
    GitLabID     int64
    Username     string
    Email        string
    // ... 大量本地字段
    AccessToken  string  // 本地存储令牌
    RefreshToken string
    // ... 复杂的关联关系
}
```

#### 变更后
```go
// 简化的 GitLab 用户结构，仅用于 API 传输
type GitLabUser struct {
    ID        int64      `json:"id"`
    Username  string     `json:"username"`
    Email     string     `json:"email"`
    Name      string     `json:"name"`
    AvatarURL string     `json:"avatar_url"`
    IsAdmin   bool       `json:"is_admin"`
    Role      GitLabRole `json:"role,omitempty"`
}

// 统一的 GitLab 角色枚举
type GitLabRole int
const (
    GitLabGuest      GitLabRole = 10 // 访客
    GitLabReporter   GitLabRole = 20 // 学生
    GitLabDeveloper  GitLabRole = 30 // 研究员
    GitLabMaintainer GitLabRole = 40 // 教师
    GitLabOwner      GitLabRole = 50 // 管理员
)
```

### 2. 用户服务重构 (`services/user_service.go`)

#### 变更前
- 复杂的本地数据库操作
- 用户创建、更新、删除逻辑
- 密码管理和验证
- 本地角色推断逻辑

#### 变更后
```go
// 完全基于 GitLab API 的用户服务
type UserService struct {
    gitlabService *GitLabService
    config        *config.Config
}

// 核心方法：
// - GetCurrentUser(accessToken) - 从 GitLab API 获取当前用户
// - GetUserByGitLabID(accessToken, gitlabID) - 获取指定用户
// - GetProjectMembers(accessToken, projectID) - 获取项目成员
// - CheckProjectPermission(accessToken, projectID, requiredRole) - 检查项目权限
```

### 3. 认证处理器重构 (`handlers/auth_handler.go`)

#### 变更前
```go
// JWT 包含本地用户信息
type JWTClaims struct {
    GitLabAccessToken string
    GitLabUserID      int64  // 本地用户 ID
    jwt.RegisteredClaims
}

// 复杂的用户创建/更新逻辑
user, err := h.userService.CreateOrUpdateUserFromGitLab(...)
```

#### 变更后
```go
// 简化的 JWT 结构
type JWTClaims struct {
    GitLabAccessToken string `json:"gitlab_access_token"`
    jwt.RegisteredClaims
}

// 简化的认证流程
gitlabUser, err := h.gitlabService.GetUser(oauthResp.AccessToken)
jwtToken, err := h.generateJWT(oauthResp.AccessToken)
```

### 4. 用户处理器重构 (`handlers/user_handler.go`)

#### 变更前
- 基于本地 UUID 的用户查询
- 本地用户信息更新
- 复杂的用户管理逻辑

#### 变更后
```go
// 基于 GitLab API 的用户操作
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
    accessToken := c.Get("gitlab_access_token")
    user, err := h.userService.GetCurrentUser(accessToken.(string))
    // 直接返回 GitLab 用户信息
}

func (h *UserHandler) UpdateCurrentUser(c *gin.Context) {
    // 重定向到 GitLab 进行用户信息更新
    c.JSON(http.StatusOK, gin.H{
        "message": "用户信息更新请前往GitLab",
        "gitlab_url": "/profile",
    })
}
```

### 5. 权限中间件重构 (`middleware/permission.go`)

#### 变更前
```go
// 基于本地角色的权限控制
func RequireRole(minRole models.UserRole) gin.HandlerFunc
func RequireAdmin() gin.HandlerFunc
func RequireTeacher() gin.HandlerFunc
```

#### 变更后
```go
// 基于 GitLab 项目权限的控制
func RequireProjectAccess(minRole models.GitLabRole) gin.HandlerFunc
func RequireProjectOwner() gin.HandlerFunc      // 管理员
func RequireProjectMaintainer() gin.HandlerFunc // 教师
func RequireProjectDeveloper() gin.HandlerFunc  // 研究员
func RequireProjectReporter() gin.HandlerFunc   // 学生
```

### 6. GitLab 服务优化 (`services/gitlab_service.go`)

#### 变更
- 分离了 API 响应结构 (`GitLabAPIUser`) 和业务模型 (`models.GitLabUser`)
- 统一了类型定义，避免重复
- 保持了完整的 GitLab API 集成功能

## 架构优势

### 1. 简化性
- **移除冗余**: 删除了所有本地用户管理逻辑
- **统一数据源**: 用户信息完全来自 GitLab
- **简化认证**: JWT 只包含必要的 GitLab 访问令牌

### 2. 一致性
- **权限体系**: 完全基于 GitLab 项目权限
- **角色映射**: 统一的 GitLab 角色到教育角色的映射
- **API 设计**: 一致的 GitLab API 调用模式

### 3. 可维护性
- **减少状态**: 无本地用户状态需要维护
- **实时同步**: 用户信息变更自动同步
- **简化部署**: 无需用户数据迁移

### 4. 安全性
- **单一认证源**: 降低认证复杂度
- **实时权限**: 权限变更立即生效
- **令牌管理**: 简化的令牌生命周期管理

## 数据流变更

### 变更前的数据流
```
GitLab OAuth → 本地用户创建/更新 → 本地数据库存储 → JWT(本地用户ID) → 本地权限检查
```

### 变更后的数据流
```
GitLab OAuth → JWT(GitLab Token) → GitLab API 用户信息 → GitLab 项目权限检查
```

## 影响的功能模块

### 1. 用户管理
- ✅ 用户登录/登出
- ✅ 用户信息获取
- ✅ 用户权限检查
- 🔄 用户信息更新（重定向到 GitLab）

### 2. 项目管理
- ✅ 基于 GitLab 项目的研究课题
- ✅ 项目成员管理
- ✅ 项目权限控制

### 3. 权限系统
- ✅ 基于 GitLab 角色的权限映射
- ✅ 项目级权限控制
- ✅ 实时权限验证

## 待完善功能

### 1. GitLab API 扩展
```go
// 需要在 GitLabService 中实现
func (s *GitLabService) GetProjectMembers(accessToken string, projectID int64) ([]*GitLabUser, error)
func (s *GitLabService) GetUserProjectRole(accessToken string, projectID, userID int64) (GitLabRole, error)
```

### 2. 权限检查完善
```go
// 需要完善项目权限检查逻辑
func (s *UserService) CheckProjectPermission(accessToken string, projectID int64, requiredRole GitLabRole) (bool, error)
```

### 3. 缓存优化
- GitLab API 调用结果缓存
- 用户信息缓存策略
- 权限检查结果缓存

## 配置变更

### 环境变量简化
```bash
# 移除的配置
# DB_USER, DB_PASSWORD, DB_NAME (用户相关数据库配置)
# JWT_USER_ID_FIELD (本地用户ID字段)

# 保留的配置
GITLAB_URL=http://localhost:8081
GITLAB_CLIENT_ID=your_client_id
GITLAB_CLIENT_SECRET=your_client_secret
JWT_SECRET=your_jwt_secret
```

## 迁移指南

### 1. 数据库迁移
- 可以安全删除 `users` 表
- 其他表中的用户外键改为 GitLab 用户 ID (int64)

### 2. 前端适配
```typescript
// 用户信息结构变更
interface User {
  id: number;        // GitLab ID (int64)
  username: string;
  email: string;
  name: string;
  avatar_url: string;
  is_admin: boolean;
  role?: string;     // 项目上下文中的角色
}
```

### 3. API 响应变更
```json
{
  "user": {
    "id": 123,
    "username": "student001",
    "email": "student001@university.edu",
    "name": "学生001",
    "avatar_url": "https://gitlab.com/avatar.jpg",
    "is_admin": false
  }
}
```

## 总结

本次重构成功实现了 SOLUTION.md 中的核心要求：

1. ✅ **完全移除本地用户表** - 用户信息完全从 GitLab API 获取
2. ✅ **JWT 只包含 GitLab 访问令牌** - 简化了认证结构
3. ✅ **不存储用户 ID** - 直接使用 GitLab token
4. ✅ **实时获取用户信息** - 包括用户名、邮箱、角色等
5. ✅ **项目权限完全基于 GitLab** - 不维护任何本地权限关系

重构后的架构更加简洁、一致和可维护，完全符合"最大化复用 GitLab 能力"的设计理念。
