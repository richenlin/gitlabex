#!/bin/bash

# GitLabEx 测试数据初始化脚本
# 使用 docker exec 方式在 PostgreSQL 容器中执行数据导入

set -e

echo "🗄️ GitLabEx 测试数据初始化 (Docker版)"
echo "======================================"

# 配置参数
POSTGRES_CONTAINER="${POSTGRES_CONTAINER:-gitlabex-postgres}"
DB_USER="${DB_USER:-gitlabex}"
DB_NAME="${DB_NAME:-gitlabex}"

# 检查 Docker 容器是否存在并运行
echo "🔍 检查 PostgreSQL 容器状态..."
if ! docker ps | grep -q "$POSTGRES_CONTAINER"; then
    echo "❌ PostgreSQL 容器 '$POSTGRES_CONTAINER' 未运行"
    echo "   请先启动 Docker Compose 服务: docker-compose -f docker-compose.dev.yml up -d postgres"
    exit 1
fi

echo "✅ PostgreSQL 容器运行正常"

# 检查数据库连接
echo "🔍 检查数据库连接..."
if ! docker exec "$POSTGRES_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" > /dev/null 2>&1; then
    echo "❌ 无法连接到数据库，请检查数据库服务是否正常"
    exit 1
fi

echo "✅ 数据库连接正常"

# 检查表是否存在
echo "🔍 检查数据表是否存在..."
table_count=$(docker exec "$POSTGRES_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'research_projects', 'topics', 'documents');" | tr -d ' ')

if [ "$table_count" -lt 4 ]; then
    echo "⚠️  数据表尚未创建，请先启动 GitLabEx 后端服务以创建数据表"
    echo "   启动命令: cd backend && ./bin/main"
    exit 1
fi

echo "✅ 数据表已存在"

# 检查是否已有测试数据
echo "🔍 检查是否已有测试数据..."
existing_users=$(docker exec "$POSTGRES_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM users WHERE username = 'admin';" | tr -d ' ')

if [ "$existing_users" -gt 0 ]; then
    echo "⚠️  检测到已存在测试数据"
    read -p "🤔 是否要重新初始化测试数据？这将清除现有数据 (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "📋 测试数据初始化已取消"
        exit 0
    fi
    
    echo "🗑️  清除现有测试数据..."
    docker exec "$POSTGRES_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" << EOF
-- 按依赖关系顺序删除数据
DELETE FROM notifications;
DELETE FROM submissions;
DELETE FROM homeworks;
DELETE FROM documents;
DELETE FROM topics;
DELETE FROM project_members;
DELETE FROM research_projects;
DELETE FROM announcements;
DELETE FROM users WHERE username IN ('admin', 'prof_wang', 'prof_li', 'prof_zhang', 'ta_chen', 'ta_liu', 'student_001', 'student_002', 'student_003', 'student_004');
EOF
    echo "✅ 现有测试数据已清除"
fi

# 执行测试数据初始化
echo "📊 开始插入测试数据..."
cd "$(dirname "$0")/.."

# 将测试数据文件复制到容器中并执行
echo "📋 复制测试数据文件到容器..."
docker cp config/test-data.sql "$POSTGRES_CONTAINER":/tmp/test-data.sql

echo "💾 在容器中执行测试数据导入..."
if docker exec "$POSTGRES_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -f /tmp/test-data.sql; then
    echo "✅ 测试数据初始化完成！"
    # 清理临时文件
    docker exec "$POSTGRES_CONTAINER" rm -f /tmp/test-data.sql
else
    echo "❌ 测试数据初始化失败"
    # 清理临时文件
    docker exec "$POSTGRES_CONTAINER" rm -f /tmp/test-data.sql
    exit 1
fi

echo ""
echo "🎉 GitLabEx 测试数据初始化成功！"
echo ""
echo "📋 测试账号信息："
echo "   管理员: admin@gitlabex.com"
echo "   教师: prof_wang (wang.prof@university.edu)"
echo "   教师: prof_li (li.prof@university.edu)" 
echo "   助教: ta_chen (chen.ta@university.edu)"
echo "   学生: student_001 (student001@university.edu)"
echo ""
echo "📊 数据统计："
docker exec "$POSTGRES_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -c "
SELECT 
    '研究课题' as 类型, COUNT(*) as 数量 FROM research_projects
UNION ALL
SELECT 
    '用户', COUNT(*) FROM users  
UNION ALL
SELECT 
    '话题', COUNT(*) FROM topics
UNION ALL  
SELECT 
    '文档', COUNT(*) FROM documents
UNION ALL
SELECT 
    '作业', COUNT(*) FROM homeworks
UNION ALL
SELECT 
    '公告', COUNT(*) FROM announcements;
"

echo ""
echo "🌐 现在可以访问以下内容："
echo "   📱 前端应用: http://localhost:3000"
echo "   🔧 后端API: http://localhost:8080"
echo "   🦊 GitLab: http://localhost:8081"
echo ""
echo "💡 游客现在可以无需登录查看："
echo "   • 课题列表和详情"
echo "   • 话题列表和详情" 
echo "   • 文档列表和搜索"
echo ""
echo "🔐 需要登录才能："
echo "   • 创建和编辑内容"
echo "   • 参与项目和提交作业"
echo "   • 管理和审核功能"
