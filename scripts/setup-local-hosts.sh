#!/bin/bash

# GitLabEx 本地Hosts文件配置脚本
# 用于跨服务器部署测试

echo "🚀 配置本地Hosts文件以支持跨服务器部署测试..."

# 备份原始hosts文件
if [ -f /etc/hosts ]; then
    sudo cp /etc/hosts /etc/hosts.backup.$(date +%Y%m%d_%H%M%S)
    echo "✅ 已备份原始hosts文件"
fi

# 检查是否已存在GitLabEx的配置
if grep -q "GitLabEx 跨服务器部署测试" /etc/hosts; then
    echo "⚠️  发现已存在的GitLabEx配置，正在更新..."
    # 删除旧的配置
    sudo sed -i '/GitLabEx 跨服务器部署测试/,/^$/d' /etc/hosts
fi

# 添加新的hosts配置
echo "
# GitLabEx 跨服务器部署测试
127.0.0.1   db.test.cn
127.0.0.1   redis.test.cn
127.0.0.1   minio.test.cn
127.0.0.1   gitlab.test.cn
127.0.0.1   backend.test.cn
127.0.0.1   frontend.test.cn
" | sudo tee -a /etc/hosts > /dev/null

echo "✅ Hosts文件配置完成！"

# 验证配置
echo "🔍 验证hosts配置..."
for domain in db.test.cn redis.test.cn minio.test.cn gitlab.test.cn backend.test.cn frontend.test.cn; do
    if ping -c 1 $domain > /dev/null 2>&1; then
        echo "✅ $domain -> 127.0.0.1 配置成功"
    else
        echo "❌ $domain 配置失败"
    fi
done

echo ""
echo "📝 当前hosts文件内容："
echo "----------------------------------------"
grep -A 10 "GitLabEx 跨服务器部署测试" /etc/hosts
echo "----------------------------------------"

echo ""
echo "🎯 下一步操作："
echo "1. 使用域名测试各个服务的连通性："
echo "   - PostgreSQL: telnet db.test.cn 5432"
echo "   - Redis: redis-cli -h redis.test.cn -a password123 ping"
echo "   - MinIO: curl http://minio.test.cn:9000/minio/health/live"
echo "   - GitLab: curl http://gitlab.test.cn:8081/-/readiness"
echo "   - Backend: curl http://backend.test.cn:8080/health"
echo "   - Frontend: curl http://frontend.test.cn:3000"
echo ""
echo "2. 启动跨服务器测试环境："
echo "   docker-compose -f docker-compose.cross-server.yml up -d"
echo ""
echo "3. 测试OAuth授权流程："
echo "   访问 http://frontend.test.cn:3000 并尝试GitLab登录"
