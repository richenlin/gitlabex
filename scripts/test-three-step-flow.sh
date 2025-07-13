#!/bin/bash

# GitLabEx 三步逻辑流程测试脚本
# 测试：
# 1. init容器配置GitLab授权后，生成授权配置文件
# 2. 配置文件生成后，映射授权配置并启动backend服务
# 3. backend服务启动成功后，重启nginx服务

set -e

echo "🧪 开始GitLabEx三步逻辑流程测试..."
echo "=================================================="

# 清理现有环境
echo "📋 准备测试环境..."
docker-compose down || true
docker volume rm gitlabex_gitlab_oauth_config 2>/dev/null || true

# 启动基础服务
echo "📋 启动基础服务..."
docker-compose up -d postgres redis gitlab onlyoffice frontend

# 等待GitLab健康检查通过
echo "📋 等待GitLab服务就绪..."
echo "这可能需要几分钟时间..."

GITLAB_READY=false
MAX_ATTEMPTS=30
ATTEMPTS=0

while [ $ATTEMPTS -lt $MAX_ATTEMPTS ]; do
    if docker-compose ps gitlab | grep -q "healthy"; then
        echo "✅ GitLab健康检查通过！"
        GITLAB_READY=true
        break
    fi
    
    ATTEMPTS=$((ATTEMPTS + 1))
    echo "等待GitLab健康检查... ($ATTEMPTS/$MAX_ATTEMPTS)"
    sleep 30
done

if [ "$GITLAB_READY" = false ]; then
    echo "❌ GitLab启动超时"
    exit 1
fi

# 启动backend服务（但不会有OAuth配置）
echo "📋 启动Backend服务（无OAuth配置状态）..."
docker-compose up -d backend

# 等待一下确保backend启动
sleep 10

# 启动nginx服务（此时应该无法正常工作）
echo "📋 启动Nginx服务..."
docker-compose up -d nginx

# 检查当前状态（应该有问题）
echo "📋 检查初始状态（预期Backend无OAuth配置）..."
docker logs gitlabex-backend --tail=10 | grep -i oauth || echo "Backend OAuth状态：未配置"

# =====================================================
# 现在运行三步流程测试
# =====================================================
echo ""
echo "🚀 开始执行三步逻辑流程..."
echo "=================================================="

# 运行init容器（执行完整的三步流程）
echo "执行init容器的完整三步流程..."
docker-compose up gitlab-init

# 检查init容器的执行结果
echo ""
echo "📊 检查流程执行结果..."
echo "=================================================="

# 检查步骤1：OAuth配置文件是否生成
echo "🔍 检查步骤1：OAuth配置文件生成"
if sudo test -f "/var/lib/docker/volumes/gitlabex_gitlab_oauth_config/_data/gitlab-oauth.env"; then
    echo "✅ 步骤1验证通过：OAuth配置文件已生成"
    echo "配置文件大小: $(sudo wc -l /var/lib/docker/volumes/gitlabex_gitlab_oauth_config/_data/gitlab-oauth.env)"
else
    echo "❌ 步骤1验证失败：OAuth配置文件未生成"
    exit 1
fi

# 检查步骤2：Backend服务是否重启并加载配置
echo ""
echo "🔍 检查步骤2：Backend服务配置加载"
sleep 5
BACKEND_LOGS=$(docker logs gitlabex-backend --tail=20)

if echo "$BACKEND_LOGS" | grep -q "OAuth config file exists"; then
    echo "✅ 步骤2验证通过：Backend已加载OAuth配置"
else
    echo "❌ 步骤2验证失败：Backend未加载OAuth配置"
    echo "Backend日志："
    echo "$BACKEND_LOGS"
    exit 1
fi

# 读取配置文件中的URL
CONFIG_FILE="config/oauth.env"
FRONTEND_BASE_URL="http://127.0.0.1:8000"  # 默认值

if [ -f "$CONFIG_FILE" ]; then
    REDIRECT_URI=$(grep "GITLAB_OAUTH_REDIRECT_URI" "$CONFIG_FILE" | cut -d'=' -f2 | sed 's/^["'\'']*//g' | sed 's/["'\'']*$//g')
    if [ -n "$REDIRECT_URI" ]; then
        FRONTEND_BASE_URL=$(echo "$REDIRECT_URI" | sed 's|/api/auth/gitlab/callback||')
    fi
fi

# 检查步骤3：Nginx服务是否重启并正常工作
echo ""
echo "🔍 检查步骤3：Nginx服务重启和路由"
sleep 5

API_HEALTH_URL="${FRONTEND_BASE_URL}/api/health"

if curl -f -s "$API_HEALTH_URL" > /dev/null; then
    echo "✅ 步骤3验证通过：Nginx路由正常工作"
    
    # 获取健康检查响应
    HEALTH_RESPONSE=$(curl -s "$API_HEALTH_URL")
    echo "健康检查响应: $HEALTH_RESPONSE"
else
    echo "❌ 步骤3验证失败：Nginx路由不正常"
    
    # 调试信息
    echo "调试信息："
    echo "Nginx状态: $(docker inspect gitlabex-nginx --format='{{.State.Status}}')"
    echo "Backend状态: $(docker inspect gitlabex-backend --format='{{.State.Status}}')"
    exit 1
fi

# 最终验证
echo ""
echo "🏁 最终验证..."
echo "=================================================="

# 验证所有服务状态
echo "📊 服务状态总览："
echo "GitLab:   $(docker inspect gitlabex-gitlab --format='{{.State.Health.Status}}')"
echo "Backend:  $(docker inspect gitlabex-backend --format='{{.State.Health.Status}}')"
echo "Nginx:    $(docker inspect gitlabex-nginx --format='{{.State.Status}}')"
echo "Frontend: $(docker inspect gitlabex-frontend --format='{{.State.Status}}')"

# 验证OAuth功能
echo ""
echo "🔐 OAuth配置验证："
sudo cat /var/lib/docker/volumes/gitlabex_gitlab_oauth_config/_data/gitlab-oauth.env | grep -v CLIENT_SECRET

# 成功总结
echo ""
echo "🎉 三步逻辑流程测试成功！"
echo "=================================================="
echo "✅ 步骤1：init容器成功配置GitLab授权并生成配置文件"
echo "✅ 步骤2：配置文件生成后，Backend服务成功重启并加载配置"
echo "✅ 步骤3：Backend启动后，Nginx服务成功重启并建立路由"
# 读取GitLab外部地址用于显示
GITLAB_EXTERNAL_URL="$FRONTEND_BASE_URL/gitlab"  # 默认值
if [ -f "$CONFIG_FILE" ]; then
    GITLAB_EXT=$(grep "GITLAB_EXTERNAL_URL" "$CONFIG_FILE" | cut -d'=' -f2 | sed 's/^["'\'']*//g' | sed 's/["'\'']*$//g')
    if [ -n "$GITLAB_EXT" ]; then
        GITLAB_EXTERNAL_URL="$GITLAB_EXT"
    fi
fi

echo ""
echo "📝 访问地址："
echo "   - 前端应用: ${FRONTEND_BASE_URL}/"
echo "   - GitLab: ${GITLAB_EXTERNAL_URL}/"
echo "   - 后端API: ${FRONTEND_BASE_URL}/api/"
echo ""
echo "🔐 OAuth认证已配置完成，可以开始使用GitLab登录"
echo "==================================================" 