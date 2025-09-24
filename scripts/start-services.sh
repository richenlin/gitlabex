#!/bin/bash

# GitLabEx 服务启动脚本
# 用于启动完整的生产环境

set -e

echo "🚀 启动 GitLabEx 生产环境..."

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
docker-compose -f docker-compose.yml down

# 构建并启动基础服务
echo "🔧 启动基础服务 (PostgreSQL, Redis, MinIO)..."
docker-compose -f docker-compose.yml up -d postgres redis minio

# 等待基础服务启动
echo "⏳ 等待基础服务启动..."
sleep 15

# 启动 GitLab
echo "🦊 启动 GitLab 服务..."
docker-compose -f docker-compose.yml up -d gitlab

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
    docker-compose -f docker-compose.yml logs --tail=20 gitlab
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

# 检查必需的镜像是否存在
echo "🔍 检查生产镜像..."

# 检查后端镜像
if ! docker images | grep -q "gitlabex-backend"; then
    echo "❌ 后端镜像 'gitlabex-backend' 不存在"
    echo ""
    echo "请先构建后端镜像："
    echo "   cd backend"
    echo "   docker build -t gitlabex-backend ."
    echo ""
    exit 1
fi

# 检查前端镜像
if ! docker images | grep -q "gitlabex-frontend"; then
    echo "❌ 前端镜像 'gitlabex-frontend' 不存在"
    echo ""
    echo "请先构建前端镜像："
    echo "   cd frontend"
    echo "   docker build -t gitlabex-frontend ."
    echo ""
    exit 1
fi

echo "✅ 生产镜像检查通过"

# 启动后端服务（使用现有镜像）
echo "🚀 启动后端服务..."
docker-compose -f docker-compose.yml up -d backend

# 等待后端服务启动
echo "⏳ 等待后端服务启动..."
sleep 15

# 检查后端服务健康状态
echo "🔍 检查后端服务状态..."
max_attempts=6
attempt=0
while [ $attempt -lt $max_attempts ]; do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo "✅ 后端服务启动成功"
        break
    fi
    attempt=$((attempt + 1))
    echo "   尝试 ${attempt}/${max_attempts}..."
    sleep 5
done

if [ $attempt -ge $max_attempts ]; then
    echo "❌ 后端服务启动失败，查看日志:"
    docker-compose -f docker-compose.yml logs --tail=30 backend
    echo ""
    echo "💡 可能的解决方案:"
    echo "   1. 检查OAuth配置是否正确"
    echo "   2. 重启后端服务: docker-compose -f docker-compose.yml restart backend"
    echo "   3. 查看完整日志: docker-compose -f docker-compose.yml logs backend"
fi

# 启动前端服务（使用现有镜像）
echo "🎨 启动前端服务..."
docker-compose -f docker-compose.yml up -d frontend

# 等待前端服务启动
echo "⏳ 等待前端服务启动..."
sleep 10

# 检查前端服务状态
echo "🔍 检查前端服务状态..."
max_attempts=6
attempt=0
while [ $attempt -lt $max_attempts ]; do
    if curl -s http://localhost:3000/ > /dev/null 2>&1; then
        echo "✅ 前端服务启动成功"
        break
    fi
    attempt=$((attempt + 1))
    echo "   尝试 ${attempt}/${max_attempts}..."
    sleep 5
done

if [ $attempt -ge $max_attempts ]; then
    echo "❌ 前端服务启动失败，查看日志:"
    docker-compose -f docker-compose.yml logs --tail=30 frontend
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
docker-compose -f docker-compose.yml ps

echo ""
echo "🎉 GitLabEx 生产环境启动完成!"
echo ""
echo "📋 服务访问信息:"
echo "   🌐 GitLab:     http://localhost:8081"
echo "   🎨 前端应用:   http://localhost:3000"
echo "   🔧 后端API:    http://localhost:8080"
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
    echo "   🔧 或手动配置后重启: docker-compose -f docker-compose.yml restart backend"
    echo ""
fi

echo "📊 生产环境管理命令:"
echo "   查看服务状态: docker-compose -f docker-compose.yml ps"
echo "   查看服务日志: docker-compose -f docker-compose.yml logs -f [service_name]"
echo "   重启后端服务: docker-compose -f docker-compose.yml restart backend"
echo "   重启前端服务: docker-compose -f docker-compose.yml restart frontend"
echo "   停止所有服务: docker-compose -f docker-compose.yml down"
echo ""
echo "🔧 配置管理:"
echo "   配置OAuth: ./scripts/configure-oauth.sh"
echo "   重新初始化数据: ./scripts/init-test-data.sh"
echo ""
echo "🐛 故障排除:"
echo "   后端日志: docker-compose -f docker-compose.yml logs backend"
echo "   前端日志: docker-compose -f docker-compose.yml logs frontend"
echo "   GitLab日志: docker-compose -f docker-compose.yml logs gitlab"
echo "   重建服务: docker-compose -f docker-compose.yml up -d --force-recreate [service_name]"
echo ""
echo "🏗️  镜像管理:"
echo "   重新构建后端镜像: cd backend && docker build -t gitlabex-backend ."
echo "   重新构建前端镜像: cd frontend && docker build -t gitlabex-frontend ."
echo "   查看镜像: docker images | grep gitlabex"
echo ""
echo "📚 更多信息请查看 README.md 和 DEPLOYMENT.md"
