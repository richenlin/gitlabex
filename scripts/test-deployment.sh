#!/bin/bash

# GitLabEx 容器化部署测试脚本

set -e

echo "🚀 开始GitLabEx容器化部署测试..."

# 清理现有容器
echo "📋 步骤1: 清理现有容器..."
docker-compose down -v || true

# 删除现有卷（可选）
echo "📋 步骤2: 清理Docker卷..."
docker volume rm gitlabex_gitlab_oauth_config 2>/dev/null || true

# 启动基础服务
echo "📋 步骤3: 启动基础服务..."
docker-compose up -d postgres redis

# 等待数据库启动
echo "📋 步骤4: 等待数据库启动..."
sleep 10

# 启动GitLab
echo "📋 步骤5: 启动GitLab服务..."
docker-compose up -d gitlab

# 等待GitLab健康检查通过
echo "📋 步骤6: 等待GitLab健康检查通过..."
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

# 启动其他服务
echo "📋 步骤7: 启动其他服务..."
docker-compose up -d onlyoffice frontend backend

# 等待backend服务启动
echo "📋 步骤8: 等待Backend服务启动..."
sleep 15

# 运行初始化容器
echo "📋 步骤9: 运行OAuth初始化..."
docker-compose up gitlab-init

# 检查初始化结果
if docker-compose logs gitlab-init | grep -q "自动初始化流程完成"; then
    echo "✅ OAuth初始化成功！"
else
    echo "❌ OAuth初始化失败"
    echo "查看详细日志:"
    docker-compose logs gitlab-init
    exit 1
fi

# 启动nginx
echo "📋 步骤10: 启动Nginx服务..."
docker-compose up -d nginx

# 等待nginx启动
sleep 10

# 验证完整服务链
echo "📋 步骤11: 验证完整服务链..."
if curl -f -s http://127.0.0.1:8080/api/health > /dev/null; then
    echo "✅ 完整服务链验证成功！"
    
    echo ""
    echo "🎉 GitLabEx容器化部署测试成功！"
    echo "📝 访问地址："
    echo "   - 前端应用: http://127.0.0.1:3000/"
    echo "   - GitLab: http://127.0.0.1:8081"
    echo "   - 后端API: http://127.0.0.1:8080/"
    echo ""
    echo "🔐 OAuth认证已配置完成，可以开始使用GitLab登录"
else
    echo "❌ 服务链验证失败"
    echo "检查服务状态:"
    docker-compose ps
    exit 1
fi

echo ""
echo "✅ 部署测试完成！" 