#!/bin/bash

# 数据库迁移执行脚本
# 用于安全地执行GitLabEx数据库架构更新

set -e  # 遇到错误时退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量（从环境变量或默认值获取）
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-"5432"}
DB_NAME=${DB_NAME:-"gitlabex"}
DB_USER=${DB_USER:-"gitlabex"}
DB_PASSWORD=${DB_PASSWORD:-"password"}

# 迁移文件路径
MIGRATION_DIR="$(dirname "$0")/../migrations"
MIGRATION_FILE="005_remove_class_management.sql"
MIGRATION_PATH="$MIGRATION_DIR/$MIGRATION_FILE"

# 日志文件
LOG_DIR="$(dirname "$0")/../logs"
LOG_FILE="$LOG_DIR/migration_$(date +%Y%m%d_%H%M%S).log"

# 函数：打印带颜色的消息
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}[$(date '+%Y-%m-%d %H:%M:%S')] ${message}${NC}"
}

# 函数：检查数据库连接
check_db_connection() {
    print_message $BLUE "检查数据库连接..."
    
    if command -v psql >/dev/null 2>&1; then
        if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c '\q' >/dev/null 2>&1; then
            print_message $GREEN "✓ 数据库连接成功"
            return 0
        else
            print_message $RED "✗ 无法连接到数据库"
            return 1
        fi
    else
        print_message $RED "✗ 未找到 psql 命令，请安装 PostgreSQL 客户端"
        return 1
    fi
}

# 函数：备份数据库
backup_database() {
    print_message $BLUE "创建数据库备份..."
    
    # 创建备份目录
    BACKUP_DIR="$(dirname "$0")/../backups"
    mkdir -p "$BACKUP_DIR"
    
    # 备份文件名
    BACKUP_FILE="$BACKUP_DIR/gitlabex_backup_$(date +%Y%m%d_%H%M%S).sql"
    
    # 执行备份
    if PGPASSWORD=$DB_PASSWORD pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME > "$BACKUP_FILE" 2>/dev/null; then
        print_message $GREEN "✓ 数据库备份完成: $BACKUP_FILE"
        echo "$BACKUP_FILE"
        return 0
    else
        print_message $RED "✗ 数据库备份失败"
        return 1
    fi
}

# 函数：检查迁移文件
check_migration_file() {
    print_message $BLUE "检查迁移文件..."
    
    if [ ! -f "$MIGRATION_PATH" ]; then
        print_message $RED "✗ 迁移文件不存在: $MIGRATION_PATH"
        return 1
    fi
    
    if [ ! -r "$MIGRATION_PATH" ]; then
        print_message $RED "✗ 迁移文件无法读取: $MIGRATION_PATH"
        return 1
    fi
    
    print_message $GREEN "✓ 迁移文件检查通过"
    return 0
}

# 函数：执行迁移
execute_migration() {
    print_message $BLUE "执行数据库迁移..."
    
    # 创建日志目录
    mkdir -p "$LOG_DIR"
    
    # 执行迁移并记录日志
    if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$MIGRATION_PATH" > "$LOG_FILE" 2>&1; then
        print_message $GREEN "✓ 数据库迁移执行成功"
        print_message $BLUE "详细日志: $LOG_FILE"
        
        # 显示迁移摘要
        if grep -q "班级管理系统移除完成" "$LOG_FILE"; then
            print_message $GREEN "✓ 班级管理系统已成功移除"
        fi
        
        return 0
    else
        print_message $RED "✗ 数据库迁移执行失败"
        print_message $YELLOW "错误日志: $LOG_FILE"
        
        # 显示错误详情
        if [ -f "$LOG_FILE" ]; then
            echo ""
            print_message $RED "错误详情:"
            tail -10 "$LOG_FILE"
        fi
        
        return 1
    fi
}

# 函数：验证迁移结果
verify_migration() {
    print_message $BLUE "验证迁移结果..."
    
    # 检查班级表是否已删除
    if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "\dt" 2>/dev/null | grep -q "classes"; then
        print_message $YELLOW "⚠ 警告: classes 表仍然存在"
    else
        print_message $GREEN "✓ classes 表已成功删除"
    fi
    
    # 检查项目表是否有新字段
    if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "\d projects" 2>/dev/null | grep -q "project_code"; then
        print_message $GREEN "✓ projects 表已添加 project_code 字段"
    else
        print_message $YELLOW "⚠ 警告: projects 表缺少 project_code 字段"
    fi
    
    print_message $GREEN "✓ 迁移验证完成"
}

# 函数：显示使用帮助
show_help() {
    echo "GitLabEx 数据库迁移脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help              显示此帮助信息"
    echo "  -n, --no-backup         跳过数据库备份（不推荐）"
    echo "  -f, --force             强制执行迁移（跳过确认）"
    echo "  -v, --verify-only       仅验证迁移状态，不执行迁移"
    echo ""
    echo "环境变量:"
    echo "  DB_HOST                 数据库主机 (默认: localhost)"
    echo "  DB_PORT                 数据库端口 (默认: 5432)"
    echo "  DB_NAME                 数据库名称 (默认: gitlabex)"
    echo "  DB_USER                 数据库用户 (默认: gitlabex)"
    echo "  DB_PASSWORD             数据库密码 (默认: password)"
    echo ""
    echo "示例:"
    echo "  $0                      正常执行迁移（包含备份）"
    echo "  $0 -n                   执行迁移但跳过备份"
    echo "  $0 -v                   仅验证迁移状态"
    echo ""
}

# 主函数
main() {
    # 参数解析
    BACKUP=true
    FORCE=false
    VERIFY_ONLY=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -n|--no-backup)
                BACKUP=false
                shift
                ;;
            -f|--force)
                FORCE=true
                shift
                ;;
            -v|--verify-only)
                VERIFY_ONLY=true
                shift
                ;;
            *)
                print_message $RED "未知选项: $1"
                print_message $YELLOW "使用 -h 或 --help 查看帮助"
                exit 1
                ;;
        esac
    done
    
    print_message $BLUE "=== GitLabEx 数据库迁移开始 ==="
    print_message $BLUE "迁移内容: 移除班级管理系统，转换为直接课题管理"
    
    # 显示数据库配置
    print_message $BLUE "数据库配置:"
    echo "  主机: $DB_HOST:$DB_PORT"
    echo "  数据库: $DB_NAME"
    echo "  用户: $DB_USER"
    echo ""
    
    # 仅验证模式
    if [ "$VERIFY_ONLY" = true ]; then
        if check_db_connection; then
            verify_migration
        fi
        exit 0
    fi
    
    # 检查数据库连接
    if ! check_db_connection; then
        print_message $RED "请检查数据库配置和连接"
        exit 1
    fi
    
    # 检查迁移文件
    if ! check_migration_file; then
        exit 1
    fi
    
    # 用户确认（除非使用 --force）
    if [ "$FORCE" != true ]; then
        echo ""
        print_message $YELLOW "⚠ 警告: 此迁移将删除班级管理相关的所有数据！"
        print_message $YELLOW "请确保已经备份重要数据。"
        echo ""
        read -p "确定要继续执行迁移吗？(y/N): " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_message $BLUE "迁移已取消"
            exit 0
        fi
    fi
    
    # 执行备份
    if [ "$BACKUP" = true ]; then
        BACKUP_FILE=$(backup_database)
        if [ $? -ne 0 ]; then
            print_message $RED "备份失败，迁移终止"
            exit 1
        fi
    else
        print_message $YELLOW "⚠ 跳过数据库备份"
    fi
    
    # 执行迁移
    if execute_migration; then
        verify_migration
        print_message $GREEN "=== 数据库迁移成功完成 ==="
        
        if [ "$BACKUP" = true ]; then
            print_message $BLUE "备份文件: $BACKUP_FILE"
        fi
        print_message $BLUE "日志文件: $LOG_FILE"
    else
        print_message $RED "=== 数据库迁移失败 ==="
        
        if [ "$BACKUP" = true ] && [ -n "$BACKUP_FILE" ]; then
            print_message $YELLOW "如需恢复，可使用备份文件: $BACKUP_FILE"
            print_message $YELLOW "恢复命令: PGPASSWORD=\$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME < $BACKUP_FILE"
        fi
        
        exit 1
    fi
}

# 脚本入口
main "$@" 