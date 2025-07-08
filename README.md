# GitLabEx Community System

基于GitLab API + Webhook构建的教育社区系统，采用Go后端 + Vue前端的技术架构。

## 项目概述

本系统提供以下核心功能：
- 📚 **知识文档管理** - 基于GitLab Wiki的文档管理系统
- ✏️ **在线协作编辑** - 集成OnlyOffice的实时文档编辑
- 💬 **话题管理** - 公告、课题、作业、讨论管理
- 👥 **用户团队管理** - 完整的用户权限和团队协作
- 💻 **代码开发管理** - 基于GitLab的代码管理和审查

## 技术架构

### 后端技术栈
- **语言**: Go 1.21+
- **框架**: Gin
- **数据库**: PostgreSQL 15+
- **缓存**: Redis 7+
- **ORM**: GORM

### 前端技术栈
- **框架**: Vue 3.4+
- **构建工具**: Vite
- **UI组件**: Element Plus
- **文档编辑**: OnlyOffice Document Server

### 基础设施
- **版本控制**: GitLab CE
- **容器化**: Docker & Docker Compose
- **文档服务**: OnlyOffice Document Server

## 快速开始

### 环境要求

- Docker 20.10+
- Docker Compose 2.0+
- 至少 4GB 内存
- 至少 10GB 可用磁盘空间

### 部署步骤

#### 1. 克隆项目
```bash
git clone <repository-url>
cd gitlabex
```

#### 2. 启动测试环境
```bash
# 使用部署脚本启动所有服务
./scripts/deploy.sh
```

#### 3. 等待服务启动
- GitLab 首次启动需要 5-10 分钟
- OnlyOffice 需要 2-3 分钟
- PostgreSQL 和 Redis 通常在 1 分钟内启动

#### 4. 访问服务
- **GitLab**: http://localhost
- **OnlyOffice**: http://localhost:8000
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

### 默认账号

| 服务 | 用户名 | 密码 |
|------|--------|------|
| GitLab | root | password123 |
| PostgreSQL | gitlabex | password123 |
| Redis | - | password123 |

## 管理命令

### 监控系统状态
```bash
# 完整系统检查
./scripts/monitor.sh

# 快速健康检查
./scripts/monitor.sh quick

# 检查容器状态
./scripts/monitor.sh containers

# 检查服务健康
./scripts/monitor.sh health

# 检查资源使用
./scripts/monitor.sh resources
```

### Docker Compose 命令
```bash
# 查看所有服务状态
docker-compose ps

# 查看服务日志
docker-compose logs [service-name]

# 重启特定服务
docker-compose restart [service-name]

# 停止所有服务
docker-compose down

# 重新构建并启动
docker-compose up --build -d
```

### 常用服务操作
```bash
# 进入PostgreSQL
docker-compose exec postgres psql -U gitlabex -d gitlabex

# 进入Redis
docker-compose exec redis redis-cli -a password123

# 查看GitLab日志
docker-compose logs gitlab

# 查看OnlyOffice日志
docker-compose logs onlyoffice
```

## 开发环境配置

### 后端开发
```bash
cd backend

# 安装依赖
go mod tidy

# 运行开发服务器
go run main.go

# 运行测试
go test ./...
```

### 前端开发
```bash
cd frontend

# 安装依赖
npm install

# 运行开发服务器
npm run dev

# 构建生产版本
npm run build
```

## 配置说明

### 环境变量配置
配置文件位于 `config/app.env`，包含以下主要配置：

```bash
# 服务器配置
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# 数据库配置
DATABASE_URL=postgres://gitlabex:password123@localhost:5432/gitlabex

# GitLab配置
GITLAB_URL=http://localhost
GITLAB_CLIENT_ID=your-gitlab-client-id
GITLAB_CLIENT_SECRET=your-gitlab-client-secret

# OnlyOffice配置
ONLYOFFICE_URL=http://localhost:8000
ONLYOFFICE_JWT_SECRET=gitlabex-jwt-secret-2024
```

### GitLab 集成设置

1. 登录 GitLab (http://localhost, root/password123)
2. 创建新的应用程序：
   - 进入 Admin Area → Applications
   - 创建新应用，设置回调URL: `http://localhost:8080/auth/callback`
   - 获取 Client ID 和 Client Secret
3. 更新配置文件中的 GitLab 凭据

## 故障排除

### 常见问题

#### 1. GitLab 启动缓慢
GitLab 首次启动需要初始化，请耐心等待 5-10 分钟。

#### 2. OnlyOffice 无法访问
检查容器是否正常启动：
```bash
docker-compose logs onlyoffice
```

#### 3. 数据库连接失败
确保 PostgreSQL 容器正常运行：
```bash
docker-compose exec postgres pg_isready -U gitlabex
```

#### 4. 端口冲突
如果遇到端口冲突，修改 `docker-compose.yml` 中的端口映射。

### 日志查看
```bash
# 查看所有服务日志
docker-compose logs

# 查看特定服务日志
docker-compose logs [service-name]

# 实时查看日志
docker-compose logs -f [service-name]
```

### 数据备份
```bash
# 备份PostgreSQL数据
docker-compose exec postgres pg_dump -U gitlabex gitlabex > backup.sql

# 备份GitLab数据
docker-compose exec gitlab gitlab-backup create
```

## 项目结构

```
gitlabex/
├── backend/                 # Go后端代码
│   ├── cmd/                # 应用入口
│   ├── internal/           # 内部模块
│   │   ├── models/        # 数据模型
│   │   ├── services/      # 业务逻辑
│   │   ├── handlers/      # HTTP处理器
│   │   └── config/        # 配置管理
│   └── pkg/               # 公共包
├── frontend/               # Vue前端代码
│   ├── src/
│   │   ├── components/    # 组件
│   │   ├── views/         # 页面
│   │   ├── stores/        # 状态管理
│   │   └── services/      # API服务
│   └── public/
├── config/                 # 配置文件
├── scripts/                # 部署脚本
├── docs/                   # 文档
└── docker-compose.yml      # Docker编排文件
```

## 开发计划

### 第一阶段 ✅
- [x] 搭建测试环境
- [x] 基础架构搭建
- [ ] GitLab API集成
- [ ] 基础文档管理

### 第二阶段
- [ ] 在线编辑器实现
- [ ] 实时协作功能
- [ ] WebSocket集成
- [ ] 文件管理系统

### 第三阶段
- [ ] 话题管理系统
- [ ] 作业管理功能
- [ ] 通知系统
- [ ] 团队管理

### 第四阶段
- [ ] 系统集成测试
- [ ] 性能优化
- [ ] 安全加固
- [ ] 部署上线

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证。详情请参见 [LICENSE](LICENSE) 文件。

## 联系方式

如有问题或建议，请通过以下方式联系：
- 提交 Issue
- 发起 Discussion
- 邮件联系：[your-email@example.com]
