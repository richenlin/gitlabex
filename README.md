# GitLabEx 教育协作平台

## 概述

GitLabEx 是一个基于 GitLab 的教育协作平台，专为教学场景设计的现代化教育管理系统。通过深度集成 GitLab 生态系统，提供完整的教育场景解决方案。

### 核心理念

- 🔗 **最大化复用GitLab能力** - 用户管理、团队协作、权限控制、项目管理完全依赖GitLab
- 📚 **教育场景优化** - 针对教学流程的专业功能设计
- 🎯 **简化操作流程** - 为教师和学生提供直观易用的界面
- 🚀 **企业级安全** - 完整的OAuth认证和API安全机制
- 🌐 **第三方集成** - 完善的第三方系统集成能力

## 技术栈

### 前端技术栈
- **Vue 3.4+** - 现代化前端框架
- **TypeScript** - 类型安全的JavaScript
- **Element Plus** - 企业级UI组件库
- **Vite** - 快速的构建工具
- **Pinia** - 状态管理
- **Vue Router** - 路由管理

### 后端技术栈
- **Go 1.23+** - 高性能后端语言
- **Gin** - 轻量级Web框架
- **GORM** - ORM数据库操作
- **PostgreSQL 15+** - 关系型数据库
- **Redis 7+** - 内存数据库
- **JWT** - 身份认证

### 集成服务
- **GitLab CE** - 版本控制和协作
- **Docker** - 容器化部署

## 项目结构

```
gitlabex2/
├── backend/                 # 后端Go服务
│   ├── cmd/                # 应用入口
│   ├── internal/           # 内部包
│   │   ├── config/        # 配置管理
│   │   ├── database/      # 数据库连接
│   │   ├── handlers/      # HTTP处理器
│   │   ├── middleware/    # 中间件
│   │   ├── models/        # 数据模型
│   │   └── services/      # 业务逻辑
│   ├── Dockerfile         # Docker构建文件
│   ├── go.mod            # Go模块文件
│   └── go.sum            # 依赖校验
├── frontend/               # 前端Vue应用
│   ├── src/               # 源代码
│   │   ├── components/    # Vue组件
│   │   ├── views/        # 页面视图
│   │   ├── router/       # 路由配置
│   │   ├── stores/       # 状态管理
│   │   └── services/     # API服务
│   ├── package.json      # npm配置
│   └── vite.config.ts    # Vite配置
├── design/                 # 原型设计
│   ├── css/               # 样式文件
│   ├── js/                # 交互脚本
│   └── *.html            # 原型页面
├── docker-compose.dev.yml  # 开发环境配置
├── README.md              # 项目说明
└── SOLUTION.md           # 详细设计文档
```

## 功能特性

### 核心功能
- **社区首页** - 展示热门课题、快速访问和通知公告
- **研究课题管理** - 基于GitLab项目的课题创建和管理
- **话题讨论** - 基于GitLab Issues的话题讨论功能
- **文档管理** - 自动识别和管理项目中的文档文件
- **作业系统** - 完整的作业发布、提交和批改流程
- **第三方系统集成** - 支持外部系统用户同步和OAuth登录

### 用户角色
- **管理员（Admin）** - 系统管理员，拥有所有权限
- **教师（Teacher）** - 可以创建和管理课题、作业，查看所有学生提交
- **研究员（Assistant）** - 可以参与课题开发，协助教学
- **学生（Student）** - 可以参与课题，提交作业，查看个人统计

### 权限系统
- 基于GitLab权限的自动映射
- 细粒度的资源级权限控制
- 支持公开课题和专有课题

## 快速开始

### 环境要求
- Go 1.23+
- Node.js 18+
- PostgreSQL 15+
- Redis 7+
- GitLab CE/EE

### 安装步骤

1. **克隆项目**
```bash
git clone <repository-url>
cd gitlabex2
```

2. **配置环境变量**
```bash
# 复制环境变量模板
cp .env.example .env

# 编辑配置文件
vim .env
```

3. **启动数据库服务**
```bash
# 使用Docker Compose启动开发环境
docker-compose -f docker-compose.dev.yml up -d
```

4. **启动后端服务**
```bash
cd backend
go mod tidy
go run cmd/main.go
```

5. **启动前端服务**
```bash
cd frontend
npm install
npm run dev
```

### 第三方系统集成快速开始

如果您需要集成外部系统，可以使用我们的第三方API：

1. **配置API密钥**
```bash
# 在环境变量中设置第三方API密钥
export THIRD_PARTY_API_KEY="your_secure_api_key_here"
export GITLAB_SYSTEM_TOKEN="your_gitlab_admin_token"
```

2. **创建用户示例**
```bash
curl -X POST "http://localhost:8080/api/v1/sync/users" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your_secure_api_key_here" \
  -d '{
    "username": "john_doe",
    "password": "SecurePass123!",
    "email": "john.doe@example.com",
    "name": "John Doe",
    "role": "student"
  }'
```

3. **用户登录流程**
   - 用户访问GitLabEx登录页面
   - 点击"通过GitLab登录"
   - 使用创建的用户名和密码完成OAuth认证
   - 自动登录到GitLabEx平台

### 环境配置

创建 `.env` 文件并配置以下环境变量：

```bash
# 服务器配置
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
APP_ENV=development
APP_DEBUG=true

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=gitlab
DB_PASSWORD=password123
DB_NAME=gitlabex

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=password123

# GitLab配置
GITLAB_URL=http://localhost:8081
GITLAB_CLIENT_ID=your_gitlab_client_id
GITLAB_CLIENT_SECRET=your_gitlab_client_secret
GITLAB_REDIRECT_URI=http://localhost:3000/auth/gitlab/callback

# JWT配置
JWT_SECRET=your_jwt_secret_key_here_please_change_in_production
JWT_EXPIRATION_HOURS=24

# 第三方系统集成配置
THIRD_PARTY_API_KEY=your_secure_third_party_api_key_32_chars_minimum
GITLAB_SYSTEM_TOKEN=your_gitlab_admin_token
```

## API文档

### 认证端点
- `GET /api/v1/auth/gitlab` - GitLab OAuth登录
- `GET /api/v1/auth/gitlab/callback` - OAuth回调
- `POST /api/v1/auth/refresh` - 刷新Token
- `POST /api/v1/auth/logout` - 用户登出

### 用户管理
- `GET /api/v1/users/me` - 获取当前用户信息
- `PUT /api/v1/users/me` - 更新当前用户信息
- `GET /api/v1/users` - 获取用户列表
- `GET /api/v1/users/:id` - 获取指定用户信息

### 研究课题
- `GET /api/v1/research-projects` - 获取课题列表
- `POST /api/v1/research-projects` - 创建新课题
- `GET /api/v1/research-projects/:id` - 获取课题详情
- `PUT /api/v1/research-projects/:id` - 更新课题信息
- `DELETE /api/v1/research-projects/:id` - 删除课题

### 话题管理
- `GET /api/v1/topics` - 获取话题列表
- `POST /api/v1/topics` - 创建新话题
- `GET /api/v1/topics/:id` - 获取话题详情
- `PUT /api/v1/topics/:id` - 更新话题
- `DELETE /api/v1/topics/:id` - 删除话题

### 第三方系统集成
- `POST /api/v1/sync/users` - 创建用户（需要第三方API密钥）
- `GET /api/v1/sync/users/:username` - 获取用户信息（需要第三方API密钥）

详细的第三方系统集成文档请查看：[SYNC_USER.md](docs/SYNC_USER.md)

## 部署说明

### Docker部署
```bash
# 构建镜像
docker build -t gitlabex-backend ./backend
docker build -t gitlabex-frontend ./frontend

# 运行容器
docker-compose up -d
```

### 生产环境配置
1. 设置环境变量为生产模式
2. 配置HTTPS证书
3. 设置防火墙规则
4. 配置监控和日志

## 开发指南

### 代码规范
- Go代码遵循gofmt格式化
- TypeScript使用ESLint+Prettier
- 提交信息遵循Conventional Commits

### 测试
```bash
# 后端测试
cd backend
go test ./...

# 前端测试
cd frontend
npm run test
```

## 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 联系方式

- 项目维护者：[维护者名称]
- 邮箱：[邮箱地址]
- 项目主页：[项目URL]

## 更新日志

查看 [CHANGELOG.md](CHANGELOG.md) 了解版本更新历史。
