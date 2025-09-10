#!/bin/bash

# GitLabEx 测试数据初始化脚本
# 在应用启动后执行，插入演示数据

set -e

echo "🗄️ GitLabEx 测试数据初始化"
echo "============================"

# 配置数据库连接参数
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-gitlabex}"
DB_PASSWORD="${DB_PASSWORD:-password123}"
DB_NAME="${DB_NAME:-gitlabex}"

# 检查数据库连接
echo "🔍 检查数据库连接..."
if ! PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" > /dev/null 2>&1; then
    echo "❌ 无法连接到数据库，请检查数据库服务是否启动"
    exit 1
fi

echo "✅ 数据库连接正常"

# 检查表是否存在
echo "🔍 检查数据表是否存在..."
table_count=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'research_projects', 'topics', 'documents');" | tr -d ' ')

if [ "$table_count" -lt 4 ]; then
    echo "⚠️  数据表尚未创建，请先启动 GitLabEx 后端服务以创建数据表"
    echo "   启动命令: cd backend && ./bin/main"
    exit 1
fi

echo "✅ 数据表已存在"

# 检查是否已有测试数据
echo "🔍 检查是否已有测试数据..."
existing_users=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM users WHERE username = 'admin';" | tr -d ' ')

if [ "$existing_users" -gt 0 ]; then
    echo "⚠️  检测到已存在测试数据"
    read -p "🤔 是否要重新初始化测试数据？这将清除现有数据 (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "📋 测试数据初始化已取消"
        exit 0
    fi
    
    echo "🗑️  清除现有测试数据..."
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" << EOF
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

if PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f config/test-data.sql; then
    echo "✅ 测试数据初始化完成！"
else
    echo "❌ 测试数据初始化失败"
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
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
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
