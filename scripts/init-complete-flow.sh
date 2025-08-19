#!/bin/sh

# GitLabEx 完整自动化授权流程
# 严格按照三步逻辑执行：
# 1. init容器配置GitLab授权后，生成授权配置文件
# 2. 配置文件生成后，映射授权配置并启动backend服务
# 3. backend服务启动成功后，完成初始化

# 启用shell调试模式
set -x

echo "🚀 开始GitLabEx自动化授权流程..."
echo "📋 按照三步逻辑严格执行"

# 读取全局配置文件
echo "📋 读取全局配置文件..."
CONFIG_FILE="/config/oauth.env"

if [ ! -f "$CONFIG_FILE" ]; then
    echo "❌ 错误：配置文件不存在 $CONFIG_FILE"
    ls -la /config/
    exit 1
fi

# 解析配置文件
GITLAB_EXTERNAL_URL=""
GITLAB_OAUTH_REDIRECT_URI=""

while IFS='=' read -r key value; do
    # 跳过注释和空行
    case $key in \#*) continue ;; esac
    [ -z "$key" ] && continue
    
    # 移除值的引号
    value=$(echo "$value" | sed 's/^["'\'']*//g' | sed 's/["'\'']*$//g')
    
    case $key in
        GITLAB_EXTERNAL_URL)
            GITLAB_EXTERNAL_URL="$value"
            ;;
        GITLAB_OAUTH_REDIRECT_URI)
            GITLAB_OAUTH_REDIRECT_URI="$value"
            ;;
    esac
done < "$CONFIG_FILE"

# 从重定向URI中提取前端地址
if [ -n "$GITLAB_OAUTH_REDIRECT_URI" ]; then
    # 从 http://localhost:8080/api/auth/gitlab/callback 提取 http://localhost:8080
    FRONTEND_BASE_URL=$(echo "$GITLAB_OAUTH_REDIRECT_URI" | sed 's|/api/auth/gitlab/callback||')
else
    echo "❌ 错误：无法从配置文件中读取重定向URI"
    cat "$CONFIG_FILE"
    exit 1
fi

echo "✅ 配置加载完成："
echo "   GitLab外部地址: $GITLAB_EXTERNAL_URL"
echo "   前端基础地址: $FRONTEND_BASE_URL"
echo "   重定向URI: $GITLAB_OAUTH_REDIRECT_URI"

# =====================================================
# 步骤1: init容器配置GitLab授权后，生成授权配置文件
# =====================================================
echo ""
echo "🔐 步骤1: 配置GitLab授权并生成配置文件"
echo "=================================================="

# 安装必要工具
echo "安装必要工具..."
apk add --no-cache curl docker-cli jq

# 确保共享目录存在
echo "检查共享目录..."
if [ ! -d "/shared" ]; then
    echo "❌ 错误：共享目录不存在"
    ls -la /
    exit 1
fi

# 检查共享目录权限
echo "检查共享目录权限..."
if [ ! -w "/shared" ]; then
    echo "❌ 错误：共享目录没有写入权限"
    echo "目录权限："
    ls -la /shared
    echo "当前用户："
    id
    exit 1
fi

# 等待GitLab完全启动
echo "等待GitLab完全启动..."
MAX_WAIT=60
WAIT_COUNT=0

check_gitlab_rails() {
    # 检查 GitLab Rails 状态
    echo "检查GitLab Rails状态..."
    RAILS_STATUS=$(docker exec gitlabex-gitlab /opt/gitlab/bin/gitlab-ctl status puma)
    RAILS_EXIT_CODE=$?
    echo "Rails状态检查退出码: $RAILS_EXIT_CODE"
    echo "Rails状态输出: $RAILS_STATUS"
    
    SIDEKIQ_STATUS=$(docker exec gitlabex-gitlab /opt/gitlab/bin/gitlab-ctl status sidekiq)
    SIDEKIQ_EXIT_CODE=$?
    echo "Sidekiq状态检查退出码: $SIDEKIQ_EXIT_CODE"
    echo "Sidekiq状态输出: $SIDEKIQ_STATUS"
    
    # 检查 Puma (Rails) 状态
    if ! echo "$RAILS_STATUS" | grep -q "run: puma:"; then
        echo "⏳ GitLab Rails (Puma) 未运行"
        return 1
    fi
    
    # 检查 Sidekiq 状态
    if ! echo "$SIDEKIQ_STATUS" | grep -q "run: sidekiq:"; then
        echo "⏳ GitLab Sidekiq 未运行"
        return 1
    fi
    
    # 检查容器状态
    echo "检查GitLab容器状态..."
    GITLAB_STATUS=$(docker inspect gitlabex-gitlab --format='{{.State.Status}}' 2>/dev/null || echo "not_found")
    GITLAB_HEALTH=$(docker inspect gitlabex-gitlab --format='{{.State.Health.Status}}' 2>/dev/null || echo "no_health")
    
    echo "GitLab容器状态: $GITLAB_STATUS"
    echo "GitLab健康状态: $GITLAB_HEALTH"
    
    if [ "$GITLAB_STATUS" != "running" ] || [ "$GITLAB_HEALTH" != "healthy" ]; then
        echo "⏳ GitLab 容器未就绪"
        return 1
    fi
    
    # 所有检查通过
    return 0
}

while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
    echo "检查GitLab状态... (${WAIT_COUNT}/${MAX_WAIT})"
    
    if check_gitlab_rails; then
        echo "✅ GitLab服务已完全就绪"
        # 显示所有服务状态
        echo "GitLab服务状态："
        docker exec gitlabex-gitlab /opt/gitlab/bin/gitlab-ctl status
        break
    else
        echo "⏳ GitLab服务还未就绪..."
        # 每6次循环显示详细状态
        if [ $((WAIT_COUNT % 6)) -eq 0 ]; then
            echo "GitLab服务状态："
            docker exec gitlabex-gitlab /opt/gitlab/bin/gitlab-ctl status || true
        fi
    fi
    
    WAIT_COUNT=$(expr $WAIT_COUNT + 1)
    sleep 10
done

if [ $WAIT_COUNT -eq $MAX_WAIT ]; then
    echo "❌ 错误：GitLab启动超时"
    echo "GitLab服务状态："
    docker exec gitlabex-gitlab /opt/gitlab/bin/gitlab-ctl status || true
    echo "GitLab系统日志："
    docker exec gitlabex-gitlab tail -n 20 /var/log/gitlab/gitlab-rails/production.log || true
    exit 1
fi

# 生成OAuth配置
echo "生成GitLab OAuth配置..."
echo "执行OAuth配置脚本..."
if docker exec gitlabex-gitlab gitlab-rails runner /scripts/init-gitlab-oauth.rb; then
    echo "✅ OAuth应用创建脚本执行成功"
else
    echo "❌ 错误：OAuth应用创建失败"
    echo "GitLab Rails日志："
    docker exec gitlabex-gitlab tail -n 50 /var/log/gitlab/gitlab-rails/production.log || true
    exit 1
fi

# 验证配置文件是否生成
if [ -f "/shared/gitlab-oauth.env" ]; then
    echo "✅ 步骤1完成：OAuth配置文件生成成功"
    echo "配置文件位置: /shared/gitlab-oauth.env"
    echo "配置文件权限："
    ls -l /shared/gitlab-oauth.env
    
    # 显示配置预览（隐藏敏感信息）
    echo "配置预览:"
    sed 's/GITLAB_CLIENT_SECRET=.*/GITLAB_CLIENT_SECRET=***HIDDEN***/' /shared/gitlab-oauth.env | head -5
else
    echo "❌ 错误：OAuth配置文件未生成"
    echo "共享目录内容："
    ls -la /shared/
    exit 1
fi

# =====================================================
# 步骤2: 配置文件生成后，映射授权配置并启动backend服务
# =====================================================
echo ""
echo "🚀 步骤2: 映射授权配置并启动Backend服务"
echo "=================================================="

# 确保backend服务存在并重启
echo "检查Backend服务状态..."
if docker ps -a | grep -q "gitlabex-backend"; then
    echo "重启Backend服务以应用新的OAuth配置..."
    
    if docker restart gitlabex-backend; then
        echo "✅ Backend服务重启成功"
    else
        echo "❌ 错误：Backend服务重启失败"
        exit 1
    fi
else
    echo "❌ 错误：Backend服务不存在"
    exit 1
fi

# 等待Backend服务完全启动
echo "等待Backend服务完全启动..."
BACKEND_WAIT=0
MAX_BACKEND_WAIT=60

while [ $BACKEND_WAIT -lt $MAX_BACKEND_WAIT ]; do
    echo "检查Backend健康状态... (${BACKEND_WAIT}/${MAX_BACKEND_WAIT})"
    
    # 首先检查容器是否在运行
    BACKEND_STATUS=$(docker inspect gitlabex-backend --format='{{.State.Status}}' 2>/dev/null || echo "not_found")
    
    if [ "$BACKEND_STATUS" = "running" ]; then
        # 直接检查API健康状态
        if docker exec gitlabex-backend curl -f -s http://localhost:8080/api/health > /dev/null 2>&1; then
            echo "✅ Backend API健康检查通过"
            break
        else
            echo "⏳ Backend API还未就绪..."
            # 如果API检查失败，显示容器日志
            if [ $((BACKEND_WAIT % 6)) -eq 0 ]; then
                echo "Backend最新日志："
                docker logs gitlabex-backend --tail=5 || true
            fi
        fi
    else
        echo "⏳ Backend容器状态: $BACKEND_STATUS"
    fi
    
    BACKEND_WAIT=$(expr $BACKEND_WAIT + 1)
    sleep 5
done

if [ $BACKEND_WAIT -eq $MAX_BACKEND_WAIT ]; then
    echo "❌ 错误：Backend服务启动超时"
    echo "Backend日志："
    docker logs gitlabex-backend --tail=20 || true
    exit 1
else
    echo "✅ 步骤2完成：Backend服务启动成功"
fi

# =====================================================
# 步骤3: 验证服务可用性
# =====================================================
echo ""
echo "🌐 步骤3: 验证服务可用性"
echo "=================================================="

# 验证完整服务链
echo "验证完整服务链..."
API_HEALTH_URL="${FRONTEND_BASE_URL}/api/health"

if curl -f -s "$API_HEALTH_URL" > /dev/null 2>&1; then
    echo "✅ Backend服务验证成功！"
    
    # 显示服务状态
    echo ""
    echo "📊 服务状态总览："
    echo "GitLab:   $(docker inspect gitlabex-gitlab --format='{{.State.Health.Status}}' 2>/dev/null || echo 'unknown')"
    echo "Backend:  $(docker inspect gitlabex-backend --format='{{.State.Health.Status}}' 2>/dev/null || echo 'unknown')"
    echo "Frontend: $(docker inspect gitlabex-frontend --format='{{.State.Status}}' 2>/dev/null || echo 'unknown')"
    
else
    echo "⚠️  警告：Backend服务验证失败，请检查配置"
    echo "尝试直接访问Backend..."
    if docker exec gitlabex-backend curl -f -s http://localhost:8080/api/health > /dev/null 2>&1; then
        echo "✅ Backend服务内部访问正常"
    else
        echo "❌ Backend服务访问失败"
    fi
fi

echo ""
echo "🎉 GitLabEx自动化授权流程完成！"
echo "=================================================="
echo "📝 访问地址："
echo "   - GitLab: ${GITLAB_EXTERNAL_URL}"
echo "   - Backend API: ${FRONTEND_BASE_URL}/api"
echo "   - Frontend: http://localhost:3000"
echo ""
echo "🔐 OAuth认证已配置完成，可以开始使用GitLab登录"
echo "==================================================" 