#!/bin/bash

# GitLab OAuth应用自动初始化脚本

set -e

echo "等待GitLab启动完成..."

# 等待GitLab健康检查通过
MAX_TRIES=60
TRIES=0

while [ $TRIES -lt $MAX_TRIES ]; do
    echo "检查GitLab状态... (尝试 $((TRIES + 1))/$MAX_TRIES)"
    
    # 检查GitLab是否响应
    if curl -f -s http://gitlab/-/health > /dev/null 2>&1; then
        echo "GitLab基础服务已启动"
        
        # 等待GitLab完全就绪 (检查用户界面)
        if curl -f -s http://gitlab/users/sign_in > /dev/null 2>&1; then
            echo "GitLab完全就绪！"
            break
        fi
    fi
    
    TRIES=$((TRIES + 1))
    echo "GitLab还未完全启动，等待10秒..."
    sleep 10
done

if [ $TRIES -eq $MAX_TRIES ]; then
    echo "错误：GitLab启动超时"
    exit 1
fi

echo "GitLab已启动，开始创建OAuth应用..."

# 创建共享目录
mkdir -p /shared

# 等待额外的5秒确保GitLab完全初始化
sleep 5

# 使用GitLab Rails Console执行OAuth应用创建
echo "通过GitLab Rails Console创建OAuth应用..."
docker exec gitlabex-gitlab gitlab-rails runner - < /scripts/init-gitlab-oauth.rb

# 检查配置文件是否创建成功
if [ -f /shared/gitlab-oauth.env ]; then
    echo "OAuth应用配置已成功创建！"
    cat /shared/gitlab-oauth.env
else
    echo "错误：OAuth应用配置创建失败"
    exit 1
fi

echo "GitLab OAuth应用初始化完成！" 