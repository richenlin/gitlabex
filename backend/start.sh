#!/bin/sh

# Backend启动脚本 - 等待OAuth配置后启动

set -e

echo "Backend启动中..."

# 等待OAuth配置文件生成
OAUTH_CONFIG_PATH="${GITLAB_OAUTH_CONFIG_PATH:-/shared/gitlab-oauth.env}"
MAX_WAIT=300  # 最大等待5分钟
WAIT_TIME=0

echo "等待OAuth配置文件: $OAUTH_CONFIG_PATH"

while [ ! -f "$OAUTH_CONFIG_PATH" ] && [ $WAIT_TIME -lt $MAX_WAIT ]; do
    echo "OAuth配置文件不存在，等待中... (${WAIT_TIME}s/${MAX_WAIT}s)"
    sleep 5
    WAIT_TIME=$((WAIT_TIME + 5))
done

if [ -f "$OAUTH_CONFIG_PATH" ]; then
    echo "✅ OAuth配置文件已生成"
    echo "配置文件路径: $OAUTH_CONFIG_PATH"
    echo "配置内容预览:"
    cat "$OAUTH_CONFIG_PATH" | sed 's/GITLAB_CLIENT_SECRET=.*/GITLAB_CLIENT_SECRET=***/'
    
    # 加载OAuth配置到环境变量
    set -a  # 自动导出变量
    source "$OAUTH_CONFIG_PATH"
    set +a  # 关闭自动导出
    
    echo "✅ OAuth配置已加载到环境变量"
    echo "Client ID: ${GITLAB_CLIENT_ID:0:10}..."
    echo "External URL: ${GITLAB_EXTERNAL_URL:-未设置}"
    echo "Internal URL: ${GITLAB_INTERNAL_URL:-未设置}"
    echo "Redirect URI: ${GITLAB_REDIRECT_URI:-未设置}"
else
    echo "⚠️  超时: OAuth配置文件未在${MAX_WAIT}秒内生成"
    echo "Backend将使用默认配置启动(可能导致OAuth功能不可用)"
    echo "请检查config/oauth.env配置文件和gitlab-init容器状态"
fi

# 验证必要的环境变量
echo "验证环境变量配置..."
echo "- DB_HOST: ${DB_HOST:-未设置}"
echo "- REDIS_HOST: ${REDIS_HOST:-未设置}"
echo "- GITLAB_EXTERNAL_URL: ${GITLAB_EXTERNAL_URL:-未设置}"
echo "- GITLAB_INTERNAL_URL: ${GITLAB_INTERNAL_URL:-未设置}"
echo "- GITLAB_REDIRECT_URI: ${GITLAB_REDIRECT_URI:-未设置}"
echo "- GITLAB_CLIENT_ID: ${GITLAB_CLIENT_ID:+已设置}"
echo "- GITLAB_CLIENT_SECRET: ${GITLAB_CLIENT_SECRET:+已设置}"

# 最终检查OAuth配置
if [ -z "$GITLAB_CLIENT_ID" ] || [ -z "$GITLAB_CLIENT_SECRET" ]; then
    echo "⚠️  警告: GitLab OAuth配置不完整"
    echo "   Client ID: ${GITLAB_CLIENT_ID:+已设置}${GITLAB_CLIENT_ID:-未设置}"
    echo "   Client Secret: ${GITLAB_CLIENT_SECRET:+已设置}${GITLAB_CLIENT_SECRET:-未设置}"
    echo "   OAuth功能将不可用，请检查配置文件并重新部署"
    echo ""
fi

# 启动主程序
echo "🚀 启动GitLabEx Backend..."
echo "服务器将在端口 ${SERVER_PORT:-8080} 上启动"
exec ./main 