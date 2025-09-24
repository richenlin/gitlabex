# GitLabEx 教育协作平台

## 概述

GitLabEx 是一个基于 GitLab 的教育协作平台，专为教学场景设计的现代化教育管理系统。通过深度集成 GitLab 生态系统，提供完整的教育场景解决方案。

### 核心理念

- 🔗 **最大化复用GitLab能力** - 用户管理、团队协作、权限控制、项目管理完全依赖GitLab
- 📚 **教育场景优化** - 针对教学流程的专业功能设计
- 🎯 **简化操作流程** - 为教师和学生提供直观易用的界面
- 🚀 **企业级安全** - 完整的OAuth认证和API安全机制
- 🌐 **第三方集成** - 完善的第三方系统集成能力

## 项目结构

```
gitlabex/
├── backend/                 # 后端Go服务
│   ├── cmd/                # 应用入口
│   ├── internal/           # 内部包
│   │   ├── config/        # 配置管理
│   │   ├── database/      # 数据库连接
│   │   ├── handlers/      # HTTP处理器
│   │   ├── middleware/    # 中间件
│   │   ├── models/        # 数据模型
│   │   ├── services/      # 业务逻辑
│   │   └── types/         # 类型定义
│   ├── Dockerfile         # Docker构建文件
│   ├── go.mod            # Go模块文件
│   └── go.sum            # 依赖校验
├── frontend/               # 前端Vue应用
│   ├── src/               # 源代码
│   │   ├── assets/        # 静态资源
│   │   ├── components/    # Vue组件
│   │   ├── composables/   # 组合式函数
│   │   ├── router/        # 路由配置
│   │   ├── services/      # API服务
│   │   ├── stores/        # 状态管理
│   │   ├── types/         # 类型定义
│   │   └── views/         # 页面视图
│   ├── package.json       # npm配置
│   └── vite.config.ts    # Vite配置
├── config/                 # 配置文件
│   ├── config.yml         # 开发环境配置
│   ├── config.prod.yml    # 生产环境配置
│   ├── .env.dev           # 开发环境变量
│   ├── .env.prod          # 生产环境变量
│   └── init-postgres.sql  # 数据库初始化脚本
├── scripts/                # 脚本文件
│   ├── start-services-dev.sh  # 开发环境启动脚本
│   ├── configure-oauth.sh    # OAuth配置脚本
│   ├── init-test-data.sh     # 测试数据初始化
│   └── build-images.sh       # Docker镜像构建
├── design/                 # 原型设计
│   ├── css/               # 样式文件
│   ├── js/                # 交互脚本
│   └── *.html            # 原型页面
├── docs/                   # 文档
│   ├── SOLUTION.md        # 需求规格说明书
│   ├── SYNC_USER.md       # 第三方集成指南
│   └── TEST_DATA.md       # 测试数据说明
├── data/                   # 数据目录
│   ├── logs/              # 日志文件
│   └── uploads/           # 上传文件
├── docker-compose.yml      # 生产环境配置
├── docker-compose.dev.yml  # 开发环境配置
├── README.md              # 项目说明
└── .gitignore            # Git忽略文件
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
- Docker 和 Docker Compose（推荐）

### 开发环境快速启动
1. **启动基础服务**
```bash
# 生成compose配置
./scripts/configure-compose.sh
# 使用Docker Compose启动开发环境基础服务
docker-compose up -d
```

2. **配置GitLab OAuth应用**
```bash
# gitlab启动成功后运行配置向导配置Oauth以及system token
./scripts/configure-oauth.sh
```

3. **启动后端服务**
```bash
cd backend
go mod tidy
go run cmd/main.go
```

4. **启动前端服务**
```bash
cd frontend
pnpm install
pnpm run dev
```

### 访问地址

启动成功后访问：
- **前端应用**: http://localhost:3000
- **GitLab管理**: http://localhost:8081 (用户名: root, 密码: b75hZ0qcwLKD)
- **MinIO控制台**: http://localhost:9001 (用户名: admin, 密码: password123)
- **后端API**: http://localhost:8080

### 环境配置

#### 开发环境配置
```bash

# 编辑配置文件
vim config/config.yml
```

#### 生产环境配置
```bash

# 编辑生产环境配置
vim config/config.prod.yml
```

### 生产环境部署

```bash
# 构建镜像
./scripts/build-images.sh

# 推送到生产环境仓库
#docker push gitlabex-backend:latest
#docker push gitlabex-frontend:latest
# 或者导出、导入镜像
docker save -o gitlabex-backend:latest gitlabex-backend.tar
docker save -o gitlabex-frontend:latest gitlabex-frontend.tar
docker load -i gitlabex-backend.tar
docker load -i gitlabex-frontend.tar


# 生成compose配置
./scripts/configure-compose.sh

# 启动生产环境基础服务
docker-compose  -f docker-compose.prod.yml up -d postgres redis minio gitlab

# gitlab启动成功后运行配置向导配置Oauth以及system token
./scripts/configure-oauth.sh

# 启动后端/前端服务
docker-compose -f docker-compose.prod.yml up -d backend frontend
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

### 文档管理
- `GET /api/v1/documents` - 获取文档列表
- `GET /api/v1/documents/:id` - 获取文档详情
- `POST /api/v1/documents` - 上传文档
- `PUT /api/v1/documents/:id` - 更新文档
- `DELETE /api/v1/documents/:id` - 删除文档
- `GET /api/v1/documents/:id/download` - 下载文档

### 作业管理
- `GET /api/v1/homework` - 获取作业列表
- `POST /api/v1/homework` - 创建作业
- `GET /api/v1/homework/:id` - 获取作业详情
- `PUT /api/v1/homework/:id` - 更新作业
- `DELETE /api/v1/homework/:id` - 删除作业
- `POST /api/v1/homework/:id/submissions` - 提交作业
- `PUT /api/v1/submissions/:id` - 批改作业

### 第三方系统集成
- `POST /api/v1/sync/users` - 创建用户（需要第三方API密钥）
- `GET /api/v1/sync/users/:username` - 获取用户信息（需要第三方API密钥）

详细的第三方系统集成文档请查看：[SYNC_USER.md](docs/SYNC_USER.md)

完整的API文档请查看项目源代码中的详细注释。


## 贡献指南

1. **Fork 项目**
2. **创建功能分支** (`git checkout -b feature/amazing-feature`)
3. **提交更改** (`git commit -m 'Add amazing feature'`)
4. **推送到分支** (`git push origin feature/amazing-feature`)
5. **创建Pull Request**

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 更新日志

查看 [CHANGELOG.md](CHANGELOG.md) 了解版本更新历史。