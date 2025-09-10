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
echo "🔧 启动基础服务 (PostgreSQL, Redis)..."
docker-compose -f docker-compose.dev.yml up -d postgres redis

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

# 构建并启动后端服务
echo "🔨 构建并启动后端服务..."
docker-compose -f docker-compose.dev.yml up -d --build backend

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
    docker-compose -f docker-compose.dev.yml logs --tail=30 backend
    echo ""
    echo "💡 可能的解决方案:"
    echo "   1. 检查OAuth配置是否正确"
    echo "   2. 重启后端服务: docker-compose -f docker-compose.dev.yml restart backend"
    echo "   3. 查看完整日志: docker-compose -f docker-compose.dev.yml logs backend"
fi

# 启动前端服务
echo "🎨 启动前端服务..."

# 检查是否为生产环境
FRONTEND_MODE=${FRONTEND_MODE:-dev}

if [ "$FRONTEND_MODE" = "production" ]; then
    echo "🏭 生产模式启动前端..."
    
    # 检查 Node.js 和 npm
    if ! command -v node > /dev/null 2>&1; then
        echo "❌ Node.js 未安装，请先安装 Node.js"
        exit 1
    fi
    
    if ! command -v npm > /dev/null 2>&1; then
        echo "❌ npm 未安装，请先安装 npm"
        exit 1
    fi
    
    # 进入前端目录并安装依赖
    cd frontend
    echo "📦 安装前端依赖..."
    npm install
    
    # 构建前端
    echo "🔨 构建前端应用..."
    npm run build
    
    # 检查是否有 nginx
    if command -v nginx > /dev/null 2>&1; then
        echo "🌐 使用系统 nginx 启动前端服务..."
        
        # 创建 nginx 配置
        cat > /tmp/gitlabex-frontend.conf << 'EOF'
server {
    listen 3000;
    server_name localhost;
    root /home/richen/Workspace/go/src/gitlabex/frontend/dist;
    index index.html;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
    
    location /api {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF
        
        # 启动 nginx
        sudo nginx -c /tmp/gitlabex-frontend.conf -p /tmp/
        echo "✅ 前端服务已通过 nginx 启动在端口 3000"
    else
        echo "⚠️  nginx 未安装，使用 vite preview 启动..."
        npm run preview &
        FRONTEND_PID=$!
        echo "✅ 前端服务已启动在端口 3000 (PID: $FRONTEND_PID)"
    fi
    
    cd ..
else
    echo "🛠️  开发模式启动前端..."
    
    # 检查 Node.js 和 npm
    if ! command -v node > /dev/null 2>&1; then
        echo "❌ Node.js 未安装，请先安装 Node.js"
        exit 1
    fi
    
    if ! command -v npm > /dev/null 2>&1; then
        echo "❌ npm 未安装，请先安装 npm"
        exit 1
    fi
    
    # 进入前端目录并安装依赖
    cd frontend
    echo "📦 安装前端依赖..."
    npm install
    
    # 启动开发服务器
    echo "🚀 启动前端开发服务器..."
    npm run dev &
    FRONTEND_PID=$!
    echo "✅ 前端开发服务器已启动在端口 3000 (PID: $FRONTEND_PID)"
    
    cd ..
fi

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
    echo "❌ 前端服务启动失败"
    if [ ! -z "$FRONTEND_PID" ]; then
        echo "   前端进程 PID: $FRONTEND_PID"
        echo "   可以使用 kill $FRONTEND_PID 停止前端服务"
    fi
fi

# 检查服务状态
echo "📊 检查服务状态..."
docker-compose -f docker-compose.dev.yml ps

echo ""
echo "🎉 GitLabEx 开发环境启动完成!"
echo ""
echo "📋 服务访问信息:"
echo "   🌐 GitLab:     http://localhost:8081"
echo "   🎨 前端应用:   http://localhost:3000"
echo "   🔧 后端API:    http://localhost:8080"
echo "   🗄️  PostgreSQL: localhost:5432"
echo "   🔴 Redis:      localhost:6379"
echo ""
echo "🔑 GitLab 管理员账号:"
echo "   用户名: root"
echo "   密码:   b75hZ0qcwLKD"
echo ""

# 检查OAuth配置状态并给出相应提示
if grep -q "temp_id\|your_client_id_here" config/oauth.env 2>/dev/null; then
    echo "⚠️  OAuth配置提醒:"
    echo "   📝 请配置GitLab OAuth应用以启用完整登录功能"
    echo "   🛠️  运行配置向导: ./scripts/configure-oauth.sh"
    echo "   🔧 或手动配置后重启: docker-compose -f docker-compose.dev.yml restart backend"
    echo ""
fi

echo "🔧 下一步操作:"
echo "   配置OAuth: ./scripts/configure-oauth.sh"
echo "   初始化测试数据: ./scripts/init-test-data.sh"
echo ""
echo "📊 常用命令:"
echo "   查看服务状态: docker-compose -f docker-compose.dev.yml ps"
echo "   查看服务日志: docker-compose -f docker-compose.dev.yml logs -f [service_name]"
echo "   重启后端服务: docker-compose -f docker-compose.dev.yml restart backend"
echo "   停止所有服务: docker-compose -f docker-compose.dev.yml down"
echo ""
echo "🎨 前端相关命令:"
echo "   生产模式启动: FRONTEND_MODE=production ./scripts/start-services.sh"
echo "   单独启动前端: cd frontend && npm run dev"
echo "   构建前端: cd frontend && npm run build"
echo "   预览构建结果: cd frontend && npm run preview"
echo ""
echo "🐛 故障排除:"
echo "   后端日志: docker-compose -f docker-compose.dev.yml logs backend"
echo "   GitLab日志: docker-compose -f docker-compose.dev.yml logs gitlab"
echo "   重建后端: docker-compose -f docker-compose.dev.yml up -d --build backend"
echo "   配置向导: ./scripts/configure-oauth.sh"
echo "   前端进程管理: 查看上方显示的 PID 并使用 kill 命令停止"
echo ""
echo "📚 更多信息请查看 README.md 和 DEPLOYMENT.md"
