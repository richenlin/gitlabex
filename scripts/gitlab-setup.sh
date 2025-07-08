#!/bin/bash

# GitLabEx - GitLab 服务管理脚本
echo "🔧 GitLab 服务管理脚本"
echo "====================="

# 检查是否在项目根目录
if [ ! -f "docker-compose.yml" ]; then
    echo "❌ 请在项目根目录运行此脚本"
    exit 1
fi

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker未运行，请先启动Docker服务"
    echo "   sudo systemctl start docker"
    exit 1
fi

echo "🚀 启动GitLab服务..."

# 启动Docker服务
docker-compose up -d

echo "⏳ 等待GitLab服务启动..."
echo "   这可能需要几分钟时间..."

# 等待GitLab启动
GITLAB_URL="http://localhost:8081"
TIMEOUT=300  # 5分钟超时
START_TIME=$(date +%s)

while true; do
    CURRENT_TIME=$(date +%s)
    ELAPSED=$((CURRENT_TIME - START_TIME))
    
    if [ $ELAPSED -gt $TIMEOUT ]; then
        echo "❌ 超时：GitLab启动时间过长"
        echo "   请检查系统资源和Docker日志"
        exit 1
    fi
    
    if curl -s -o /dev/null -w "%{http_code}" "$GITLAB_URL/users/sign_in" | grep -q "200"; then
        echo "✅ GitLab服务启动成功！"
        break
    fi
    
    echo "   还在启动中... (已等待 ${ELAPSED}s)"
    sleep 10
done

echo ""
echo "🎉 GitLab服务准备就绪！"
echo "====================="
echo ""
echo "📋 GitLab 访问信息:"
echo "   URL: http://localhost:8081"
echo "   默认管理员用户: root"
echo "   默认密码: 需要重置"
echo ""
echo "🔑 首次登录步骤:"
echo "1. 访问 http://localhost"
echo "2. 点击 'Set password' 或 'Forgot password?'"
echo "3. 使用 root 用户重置密码"
echo "4. 或者查看初始密码："
echo "   docker-compose logs gitlab | grep 'Password:'"
echo ""
echo "⚙️  配置 OAuth 应用:"
echo "1. 以 root 用户登录"
echo "2. 进入 Admin Area → Applications"
echo "3. 创建新应用："
echo "   - Name: GitLabEx"
echo "   - Redirect URI: http://localhost:8080/api/auth/gitlab/callback"
echo "   - Scopes: read_user, read_repository, openid, profile, email"
echo ""
echo "📖 详细配置指南: backend/GITLAB_OAUTH_SETUP.md"
echo ""

# 显示服务状态
echo "🔍 服务状态:"
docker-compose ps 