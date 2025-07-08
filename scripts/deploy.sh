#!/bin/bash

# 部署脚本 - GitLabEx Community System
# 版本: 1.0.0
# 作者: AI Assistant

set -e

echo "🚀 开始部署GitLabEx社区系统..."

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印信息函数
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查Docker是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker未安装，请先安装Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose未安装，请先安装Docker Compose"
        exit 1
    fi
    
    print_success "Docker环境检查通过"
}

# 创建必要目录
create_directories() {
    print_info "创建必要目录..."
    
    mkdir -p logs
    mkdir -p ssl
    mkdir -p config/{nginx,gitlab,onlyoffice}
    mkdir -p data/{postgres,redis,gitlab,onlyoffice}
    
    print_success "目录创建完成"
}

# 生成SSL证书
generate_ssl() {
    if [ ! -f ssl/cert.pem ]; then
        print_info "生成SSL证书..."
        
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout ssl/key.pem -out ssl/cert.pem \
            -subj "/C=CN/ST=Beijing/L=Beijing/O=GitLabEx/CN=localhost" \
            2>/dev/null
        
        print_success "SSL证书生成完成"
    else
        print_info "SSL证书已存在，跳过生成"
    fi
}

# 停止现有服务
stop_services() {
    print_info "停止现有服务..."
    
    docker-compose down --remove-orphans 2>/dev/null || true
    
    print_success "现有服务已停止"
}

# 启动基础服务
start_services() {
    print_info "启动基础服务..."
    
    # 启动PostgreSQL
    print_info "启动PostgreSQL数据库..."
    docker-compose up -d postgres
    
    # 等待PostgreSQL启动
    print_info "等待PostgreSQL启动..."
    for i in {1..30}; do
        if docker-compose exec -T postgres pg_isready -U gitlabex > /dev/null 2>&1; then
            print_success "PostgreSQL已启动"
            break
        fi
        echo -n "."
        sleep 2
    done
    
    # 启动Redis
    print_info "启动Redis缓存..."
    docker-compose up -d redis
    
    # 等待Redis启动
    print_info "等待Redis启动..."
    for i in {1..15}; do
        if docker-compose exec -T redis redis-cli -a password123 ping > /dev/null 2>&1; then
            print_success "Redis已启动"
            break
        fi
        echo -n "."
        sleep 2
    done
    
    # 启动OnlyOffice
    print_info "启动OnlyOffice文档服务器..."
    docker-compose up -d onlyoffice
    
    # 启动GitLab
    print_info "启动GitLab CE..."
    docker-compose up -d gitlab
    
    print_success "基础服务启动完成"
}

# 检查服务状态
check_services() {
    print_info "检查服务状态..."
    
    # 检查PostgreSQL
    print_info "检查PostgreSQL状态..."
    if docker-compose exec -T postgres pg_isready -U gitlabex > /dev/null 2>&1; then
        print_success "✓ PostgreSQL运行正常"
    else
        print_error "✗ PostgreSQL未正常运行"
    fi
    
    # 检查Redis
    print_info "检查Redis状态..."
    if docker-compose exec -T redis redis-cli -a password123 ping > /dev/null 2>&1; then
        print_success "✓ Redis运行正常"
    else
        print_error "✗ Redis未正常运行"
    fi
    
    # 检查OnlyOffice
    print_info "检查OnlyOffice状态..."
    if curl -s http://localhost:8000/healthcheck > /dev/null 2>&1; then
        print_success "✓ OnlyOffice运行正常"
    else
        print_warning "⚠ OnlyOffice可能需要更多时间启动"
    fi
    
    # 检查GitLab
    print_info "检查GitLab状态..."
    if curl -s http://localhost/-/health > /dev/null 2>&1; then
        print_success "✓ GitLab运行正常"
    else
        print_warning "⚠ GitLab可能需要更多时间启动"
    fi
}

# 显示访问信息
show_access_info() {
    echo ""
    echo "🎉 部署完成！"
    echo "=============================="
    echo "访问地址："
    echo "  🌐 GitLab:     http://localhost"
    echo "  📄 OnlyOffice: http://localhost:8000"
    echo "  🗄️  PostgreSQL: localhost:5432"
    echo "  🔴 Redis:      localhost:6379"
    echo ""
    echo "默认账号信息："
    echo "  GitLab管理员: root / password123"
    echo "  数据库: gitlabex / password123"
    echo "  Redis: password123"
    echo ""
    echo "重要提示："
    echo "  - GitLab首次启动需要5-10分钟，请耐心等待"
    echo "  - OnlyOffice需要2-3分钟启动时间"
    echo "  - 如果服务未正常启动，请运行: docker-compose logs [service-name]"
    echo ""
}

# 主函数
main() {
    echo "GitLabEx Community System Deployment"
    echo "====================================="
    
    # 检查环境
    check_docker
    
    # 创建目录
    create_directories
    
    # 生成SSL证书
    generate_ssl
    
    # 停止现有服务
    stop_services
    
    # 启动服务
    start_services
    
    # 等待服务启动
    print_info "等待服务完全启动..."
    sleep 30
    
    # 检查服务状态
    check_services
    
    # 显示访问信息
    show_access_info
}

# 错误处理
trap 'print_error "部署过程中发生错误，请检查日志"; exit 1' ERR

# 执行主函数
main "$@" 