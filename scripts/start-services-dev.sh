#!/bin/bash

# GitLabEx 服务启动脚本
# 用于启动完整的开发环境

set -e

echo "🚀 启动 GitLabEx 开发环境..."

# 检查 Docker 是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker 未运行，请先启动 Docker"
    exit 1
fi

# 切换到项目根目录
cd "$(dirname "$0")/.."

echo "📦 检查配置文件..."

# 检查必要的配置文件
if [ ! -f "config/backend.env" ]; then
    echo "❌ 缺少 config/backend.env 文件"
    exit 1
fi

if [ ! -f "config/oauth.env" ]; then
    echo "❌ 缺少 config/oauth.env 文件"
    exit 1
fi

echo "✅ 配置文件检查完成"

# 停止可能已经运行的服务
echo "🛑 停止现有服务..."
docker-compose -f docker-compose.dev.yml down

# 构建并启动基础服务
echo "🔧 启动基础服务 (PostgreSQL, Redis, MinIO)..."
docker-compose -f docker-compose.dev.yml up -d postgres redis minio

# 等待基础服务启动
echo "⏳ 等待基础服务启动..."
sleep 15

# 启动 GitLab
echo "🦊 启动 GitLab 服务..."
docker-compose -f docker-compose.dev.yml up -d gitlab

# 等待 GitLab 启动
echo "⏳ 等待 GitLab 启动 (这可能需要几分钟)..."
timeout=300  # 5分钟超时
elapsed=0
while [ $elapsed -lt $timeout ]; do
    if curl -s http://localhost:8081/ > /dev/null 2>&1; then
        echo "✅ GitLab 启动完成"
        break
    fi
    sleep 10
    elapsed=$((elapsed + 10))
    echo "   等待中... (${elapsed}/${timeout}秒)"
done

if [ $elapsed -ge $timeout ]; then
    echo "⚠️  GitLab 启动超时，请检查服务状态"
    docker-compose -f docker-compose.dev.yml logs --tail=20 gitlab
    exit 1
fi

# 检查OAuth配置
echo "🔍 检查OAuth配置..."
if grep -q "temp_id\|your_client_id_here" config/oauth.env 2>/dev/null; then
    echo ""
    echo "⚠️  检测到OAuth配置需要更新！"
    echo ""
    echo "📋 请按照以下步骤配置GitLab OAuth应用："
    echo ""
    echo "1️⃣  打开GitLab管理界面："
    echo "   🌐 访问: http://localhost:8081"
    echo "   👤 用户名: root"
    echo "   🔑 密码: b75hZ0qcwLKD"
    echo ""
    echo "2️⃣  创建OAuth应用："
    echo "   📍 导航到: Admin Area → Applications"
    echo "   ➕ 点击 'New application'"
    echo "   📝 应用名称: GitLabEx Education Platform"
    echo "   🔗 重定向URI: http://localhost:3000/auth/gitlab/callback"
    echo "   ✅ 权限范围: api, read_user, openid"
    echo "   💾 保存应用"
    echo ""
    echo "3️⃣  复制应用凭据："
    echo "   📋 复制 Application ID"
    echo "   🔐 复制 Secret"
    echo ""
    echo "4️⃣  更新配置文件:"
    echo "   📄 编辑 config/oauth.env"
    echo "   🔄 将 GITLAB_CLIENT_ID 替换为你的 Application ID"
    echo "   🔄 将 GITLAB_CLIENT_SECRET 替换为你的 Secret"
    echo ""
    
    # 提供配置模板
    echo "📝 配置文件模板 (config/oauth.env):"
    echo "----------------------------------------"
    echo "GITLAB_URL=http://localhost:8081"
    echo "REDIRECT_URI=http://localhost:3000/auth/gitlab/callback"
    echo "APPLICATION_NAME=\"GitLabEx Education Platform\""
    echo "SCOPES=\"api read_user openid\""
    echo "GITLAB_CLIENT_ID=你的_Application_ID"
    echo "GITLAB_CLIENT_SECRET=你的_Secret"
    echo "----------------------------------------"
    echo ""
    
    # 询问用户是否要配置或已完成配置
    echo "🤔 选择操作:"
    echo "   1) 现在配置OAuth (推荐)"
    echo "   2) 稍后手动配置"
    echo "   3) 已完成配置，继续启动"
    echo ""
    read -p "请选择 (1-3): " -n 1 -r
    echo
    
    case $REPLY in
        1)
            echo "🔧 启动OAuth配置向导..."
            ./scripts/configure-oauth.sh
            if [ $? -eq 0 ]; then
                echo "✅ OAuth配置完成，继续启动后端服务"
            else
                echo "❌ OAuth配置失败，以临时配置启动"
            fi
            ;;
        2)
            echo ""
            echo "⏸️  将以临时配置启动后端服务"
            echo "   稍后可以运行以下命令配置OAuth："
            echo "   ./scripts/configure-oauth.sh"
            echo ""
            echo "⏳ 10秒后继续启动（或按Ctrl+C取消）..."
            sleep 10
            ;;
        3)
            echo "✅ 继续启动后端服务"
            ;;
        *)
            echo "无效选择，以临时配置启动"
            ;;
    esac
else
    echo "✅ OAuth配置检查通过"
fi

# 初始化测试数据
echo "📊 初始化测试数据..."
if [ -f "./scripts/init-test-data.sh" ]; then
    ./scripts/init-test-data.sh
else
    echo "⚠️  测试数据初始化脚本不存在，跳过"
fi

# 检查服务状态
echo "📊 检查服务状态..."
docker-compose -f docker-compose.dev.yml ps

echo ""
echo "🎉 GitLabEx 开发环境基础服务启动完成!"
echo ""
echo "📋 服务访问信息:"
echo "   🌐 GitLab:     http://localhost:8081"
echo "   🗄️  PostgreSQL: localhost:5432"
echo "   🔴 Redis:      localhost:6379"
echo "   📦 MinIO API:  http://localhost:9000"
echo "   🎛️  MinIO Console: http://localhost:9001"
echo ""
echo "🔑 管理员账号:"
echo "   GitLab:"
echo "     用户名: root"
echo "     密码:   b75hZ0qcwLKD"
echo "   MinIO:"
echo "     用户名: admin"
echo "     密码:   password123"
echo ""

# 检查OAuth配置状态并给出相应提示
if grep -q "temp_id\|your_client_id_here" config/oauth.env 2>/dev/null; then
    echo "⚠️  OAuth配置提醒:"
    echo "   📝 请配置GitLab OAuth应用以启用完整登录功能"
    echo "   🛠️  运行配置向导: ./scripts/configure-oauth.sh"
    echo ""
fi

echo "🚀 手动启动开发服务:"
echo ""
echo "📦 后端服务启动:"
echo "   cd backend"
echo "   go mod tidy"
echo "   go run cmd/main.go"
echo "   或者: go build -o bin/gitlabex cmd/main.go && ./bin/gitlabex"
echo ""
echo "🎨 前端服务启动:"
echo "   cd frontend"
echo "   npm install  # 或 pnpm install / yarn install"
echo "   npm run dev  # 或 pnpm dev / yarn dev"
echo ""
echo "📊 启动后访问地址:"
echo "   🎨 前端应用:   http://localhost:3000"
echo "   🔧 后端API:    http://localhost:8080"
echo ""
echo "📊 常用命令:"
echo "   查看容器状态: docker-compose -f docker-compose.dev.yml ps"
echo "   查看容器日志: docker-compose -f docker-compose.dev.yml logs -f [service_name]"
echo "   停止基础服务: docker-compose -f docker-compose.dev.yml down"
echo ""
echo "🔧 配置和初始化:"
echo "   配置OAuth: ./scripts/configure-oauth.sh"
echo "   初始化测试数据: ./scripts/init-test-data.sh"
echo ""
echo "🐛 故障排除:"
echo "   GitLab日志: docker-compose -f docker-compose.dev.yml logs gitlab"
echo "   PostgreSQL日志: docker-compose -f docker-compose.dev.yml logs postgres"
echo "   Redis日志: docker-compose -f docker-compose.dev.yml logs redis"
echo ""
echo "📚 更多信息请查看 README.md 和 DEPLOYMENT.md"
