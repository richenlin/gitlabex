#!/bin/bash

# GitLabEx OAuth 快速设置脚本
echo "🚀 GitLabEx OAuth 快速设置脚本"
echo "================================"

# 检查是否在项目根目录
if [ ! -f "docker-compose.yml" ]; then
    echo "❌ 请在项目根目录运行此脚本"
    exit 1
fi

# 创建 .env 文件
BACKEND_DIR="backend"
ENV_FILE="$BACKEND_DIR/.env"

echo "📝 创建环境配置文件..."

# 检查是否已存在 .env 文件
if [ -f "$ENV_FILE" ]; then
    echo "⚠️  .env 文件已存在，是否覆盖？(y/N)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        echo "❌ 操作取消"
        exit 1
    fi
fi

# 生成随机密钥
JWT_SECRET=$(openssl rand -base64 32)
ONLYOFFICE_JWT_SECRET=$(openssl rand -base64 32)

# 创建 .env 文件
cat > "$ENV_FILE" << EOF
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

# GitLab OAuth配置 (需要手动配置)
GITLAB_URL=http://localhost:8081
GITLAB_CLIENT_ID=your-gitlab-application-id
GITLAB_CLIENT_SECRET=your-gitlab-application-secret
GITLAB_REDIRECT_URI=http://localhost:8080/api/auth/gitlab/callback

# OnlyOffice配置
ONLYOFFICE_URL=http://localhost:8000
ONLYOFFICE_JWT_SECRET=$ONLYOFFICE_JWT_SECRET
ONLYOFFICE_CALLBACK_URL=http://localhost:8080/api/documents/callback

# JWT配置
JWT_SECRET=$JWT_SECRET
EOF

echo "✅ 环境配置文件创建完成: $ENV_FILE"
echo ""

# 显示需要配置的项目
echo "📋 下一步配置 GitLab OAuth:"
echo "1. 访问 GitLab 管理后台 (http://localhost:8081)"
echo "2. 创建新的 OAuth 应用"
echo "3. 获取 Application ID 和 Secret"
echo "4. 编辑 $ENV_FILE 文件，替换以下变量:"
echo "   - GITLAB_CLIENT_ID=your-gitlab-application-id"
echo "   - GITLAB_CLIENT_SECRET=your-gitlab-application-secret"
echo ""

echo "📖 详细配置指南: backend/GITLAB_OAUTH_SETUP.md"
echo ""

# 检查服务状态
echo "🔍 检查服务状态..."

# 检查后端
if curl -s http://localhost:8080/api/health > /dev/null; then
    echo "✅ 后端服务运行正常"
else
    echo "❌ 后端服务未运行，请启动:"
    echo "   cd backend && go run cmd/main.go"
fi

# 检查前端
if curl -s http://localhost:5173 > /dev/null; then
    echo "✅ 前端服务运行正常"
else
    echo "❌ 前端服务未运行，请启动:"
    echo "   cd frontend && npm run dev"
fi

echo ""
echo "🎉 设置完成！请按照指南配置GitLab OAuth应用后重启服务。"
echo "🌐 访问 http://localhost:5173 测试应用" 