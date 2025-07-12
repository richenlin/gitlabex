#!/bin/bash

# GitLab OAuth应用配置更新脚本
# 用于修复容器环境中的回调地址问题

echo "🔧 GitLab OAuth配置更新脚本"
echo "================================"

# 配置变量
GITLAB_URL="http://localhost:8081"
APP_ID="375dbd60a3bec327790d2f7f814458a137c83e367f4246138aa2c446afa6da5c"
NEW_CALLBACK_URL="http://172.17.0.1:8080/api/auth/gitlab/callback"

echo "📍 当前配置："
echo "   GitLab URL: $GITLAB_URL"
echo "   应用ID: $APP_ID"
echo "   新回调地址: $NEW_CALLBACK_URL"
echo ""

# 检查GitLab容器是否运行
echo "🔍 检查GitLab容器状态..."
if ! docker ps | grep -q "gitlabex-gitlab"; then
    echo "❌ GitLab容器未运行，请先启动："
    echo "   docker-compose up -d gitlab"
    exit 1
fi

echo "✅ GitLab容器正在运行"

# 检查网络连通性
echo ""
echo "🌐 检查网络连通性..."
if timeout 5 bash -c "</dev/tcp/localhost/8081"; then
    echo "✅ 可以访问GitLab (localhost:8081)"
else
    echo "❌ 无法访问GitLab，请检查端口映射"
    exit 1
fi

# 检查后端端口
if timeout 5 bash -c "</dev/tcp/172.17.0.1/8080" 2>/dev/null; then
    echo "✅ 后端服务在8080端口运行"
else
    echo "⚠️  后端服务未在8080端口运行，请先启动后端"
fi

echo ""
echo "📝 接下来需要手动更新GitLab OAuth应用："
echo ""
echo "1. 访问GitLab管理界面："
echo "   $GITLAB_URL"
echo ""
echo "2. 使用管理员账号登录："
echo "   用户名: root"
echo "   密码: b75hZ0qcwLKD"
echo ""
echo "3. 进入Admin Area → Applications"
echo ""
echo "4. 找到应用ID为以下值的应用："
echo "   $APP_ID"
echo ""
echo "5. 点击编辑，更新Redirect URI为："
echo "   $NEW_CALLBACK_URL"
echo ""
echo "6. 保存更改"
echo ""
echo "🧪 测试OAuth流程："
echo "   1. 启动后端服务: cd backend && go run cmd/main.go"
echo "   2. 访问: http://192.168.0.1:8080/api/auth/gitlab"
echo "   3. 检查重定向是否正常工作"
echo ""
echo "✨ 配置完成后，GitLab容器将能够正确回调到后端服务！" 