#!/bin/bash

# GitLab OAuth 配置辅助脚本
# 用于引导用户完成OAuth应用配置

set -e

echo "🔧 GitLab OAuth 配置向导"
echo "=========================="
echo ""

# 切换到项目根目录
cd "$(dirname "$0")/.."

# 检查GitLab是否运行
if ! curl -s http://localhost:8081/ > /dev/null 2>&1; then
    echo "❌ GitLab服务未运行，请先启动GitLab:"
    echo "   docker-compose -f docker-compose.dev.yml up -d gitlab"
    exit 1
fi

echo "✅ GitLab服务运行正常"
echo ""

# 显示配置步骤
echo "📋 OAuth应用配置步骤:"
echo ""
echo "1️⃣  访问GitLab管理界面"
echo "   🌐 URL: http://localhost:8081"
echo "   👤 用户名: root"
echo "   🔑 密码: b75hZ0qcwLKD"
echo ""
echo "2️⃣  导航到应用管理"
echo "   📍 Admin Area (左侧菜单的扳手图标)"
echo "   📱 Applications (左侧菜单)"
echo "   ➕ New application (右上角按钮)"
echo ""
echo "3️⃣  填写应用信息"
echo "   📝 Name: GitLabEx Education Platform"
echo "   🔗 Redirect URI: http://localhost:3000/auth/gitlab/callback"
echo "   ☑️  Confidential: 勾选"
echo "   🔐 Scopes: 勾选以下选项"
echo "      ✓ api"
echo "      ✓ read_user"
echo "      ✓ openid"
echo "   💾 Save application"
echo ""
echo "4️⃣  复制应用凭据"
echo "   📋 复制 Application ID"
echo "   🔐 复制 Secret"
echo ""

# 打开GitLab页面（如果可能）
if command -v open > /dev/null 2>&1; then
    read -p "🤔 是否自动打开GitLab管理页面？(y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "🌐 打开GitLab..."
        open "http://localhost:8081/admin/applications"
    fi
elif command -v xdg-open > /dev/null 2>&1; then
    read -p "🤔 是否自动打开GitLab管理页面？(y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "🌐 打开GitLab..."
        xdg-open "http://localhost:8081/admin/applications"
    fi
fi

echo ""
echo "⏳ 请完成上述步骤，然后回到这里..."
echo ""

# 等待用户完成配置
read -p "🔑 请输入 Application ID: " client_id
echo ""
read -p "🔐 请输入 Secret: " client_secret
echo ""

# 验证输入
if [ -z "$client_id" ] || [ -z "$client_secret" ]; then
    echo "❌ Application ID 和 Secret 不能为空"
    exit 1
fi

if [ "$client_id" = "your_client_id_here" ] || [ "$client_secret" = "your_client_secret_here" ]; then
    echo "❌ 请输入真实的 Application ID 和 Secret"
    exit 1
fi

# 备份原配置文件
if [ -f "config/oauth.env" ]; then
    cp config/oauth.env config/oauth.env.bak
    echo "📦 已备份原配置文件到 config/oauth.env.bak"
fi

# 更新配置文件
cat > config/oauth.env << EOF
# GitLab OAuth 应用配置
# 由 configure-oauth.sh 脚本自动生成

GITLAB_URL=http://localhost:8081
REDIRECT_URI=http://localhost:3000/auth/gitlab/callback
APPLICATION_NAME="GitLabEx Education Platform"
SCOPES="api read_user openid"

# OAuth 应用凭据
GITLAB_CLIENT_ID=$client_id
GITLAB_CLIENT_SECRET=$client_secret
EOF

echo "✅ OAuth 配置已更新！"
echo ""

# 检查后端服务是否运行
if docker-compose -f docker-compose.dev.yml ps backend | grep -q "Up"; then
    echo "🔄 检测到后端服务正在运行，需要重启以应用新配置"
    read -p "🤔 是否立即重启后端服务？(Y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        echo "🔄 重启后端服务..."
        docker-compose -f docker-compose.dev.yml restart backend
        
        echo "⏳ 等待后端服务启动..."
        sleep 10
        
        # 检查后端服务健康状态
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            echo "✅ 后端服务重启成功！"
        else
            echo "⚠️  后端服务可能未完全启动，请检查日志:"
            echo "   docker-compose -f docker-compose.dev.yml logs backend"
        fi
    fi
else
    echo "💡 后端服务未运行，配置将在下次启动时生效"
fi

echo ""
echo "🎉 OAuth 配置完成！"
echo ""
echo "📝 配置文件位置: config/oauth.env"
echo "🌐 GitLab: http://localhost:8081"
echo "🔧 后端API: http://localhost:8080"
echo "🚀 前端: 启动后访问 http://localhost:3000"
echo ""
echo "🔧 启动前端开发服务器:"
echo "   cd frontend && npm run dev"
