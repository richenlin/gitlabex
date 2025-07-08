# GitLab OAuth 应用配置指南

## 1. 创建GitLab OAuth应用

### 步骤1: 访问GitLab管理后台
1. 以管理员身份登录GitLab
2. 进入 Admin Area (管理区域)
3. 左侧菜单选择 **Applications** (应用程序)

### 步骤2: 创建新应用
1. 点击 **New Application** (新建应用)
2. 填写应用信息：
   - **Name**: `GitLabEx`
   - **Redirect URI**: `http://localhost:8080/api/auth/gitlab/callback`
   - **Scopes**: 选择以下权限
     - `read_user` - 读取用户信息
     - `read_repository` - 读取仓库信息
     - `openid` - OpenID Connect
     - `profile` - 用户配置文件
     - `email` - 电子邮件地址

### 步骤3: 获取应用凭证
创建完成后，您将获得：
- **Application ID** (客户端ID)
- **Secret** (客户端密钥)

## 2. 配置环境变量

### 创建 `.env` 文件
在 `backend/` 目录下创建 `.env` 文件：

```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=gitlabex
DB_PASSWORD=password123
DB_NAME=gitlabex
DB_SSLMODE=disable

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=password123

# 服务器配置
SERVER_PORT=8080
GIN_MODE=debug

# GitLab OAuth配置 (替换为实际值)
GITLAB_URL=http://localhost
GITLAB_CLIENT_ID=your-gitlab-application-id
GITLAB_CLIENT_SECRET=your-gitlab-application-secret
GITLAB_REDIRECT_URI=http://localhost:8080/api/auth/gitlab/callback

# OnlyOffice配置
ONLYOFFICE_URL=http://localhost:8000
ONLYOFFICE_JWT_SECRET=your-jwt-secret
ONLYOFFICE_CALLBACK_URL=http://localhost:8080/api/documents/callback

# JWT配置
JWT_SECRET=your-jwt-secret-key
```

### 3. 重启服务
配置完成后，重启Go后端服务：
```bash
cd backend
go run cmd/main.go
```

## 4. 测试OAuth流程

### 测试步骤
1. 访问 `http://localhost:5173` (前端)
2. 点击 "使用GitLab登录"
3. 应该跳转到GitLab OAuth授权页面
4. 授权后应该跳转回应用并完成登录

### 验证认证
- 检查后端日志是否有认证相关日志
- 检查是否生成了JWT令牌
- 确认用户信息是否正确同步到本地数据库

## 5. 常见问题

### OAuth回调URL不匹配
- 确保GitLab应用中的Redirect URI与环境变量中的一致
- 检查端口号是否正确

### 权限不足
- 确保GitLab应用有足够的权限 (scopes)
- 检查用户是否有权限访问必要的资源

### 网络连接问题
- 确保GitLab服务运行正常
- 检查网络连接和防火墙设置

## 6. 生产环境配置

### 域名配置
生产环境中需要更新：
- `GITLAB_URL`: 实际的GitLab域名
- `GITLAB_REDIRECT_URI`: 实际的回调URL
- 所有localhost引用都应该更新为实际域名

### 安全配置
- 使用强密码和随机密钥
- 启用HTTPS
- 定期更换密钥 