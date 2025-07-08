#!/bin/bash

# 监控脚本 - GitLabEx Community System
# 版本: 1.0.0

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

# 检查容器状态
check_container_status() {
    echo "📊 容器状态检查"
    echo "===================="
    
    containers=("gitlabex-postgres" "gitlabex-redis" "gitlabex-onlyoffice" "gitlabex-gitlab")
    
    for container in "${containers[@]}"; do
        if docker ps --filter "name=${container}" --filter "status=running" --format "table {{.Names}}" | grep -q "${container}"; then
            print_success "✓ ${container} 运行中"
        else
            print_error "✗ ${container} 未运行"
        fi
    done
    echo ""
}

# 检查服务健康状态
check_service_health() {
    echo "🏥 服务健康状态检查"
    echo "===================="
    
    # 检查PostgreSQL
    if docker-compose exec -T postgres pg_isready -U gitlabex > /dev/null 2>&1; then
        print_success "✓ PostgreSQL 健康"
    else
        print_error "✗ PostgreSQL 不健康"
    fi
    
    # 检查Redis
    if docker-compose exec -T redis redis-cli -a password123 ping > /dev/null 2>&1; then
        print_success "✓ Redis 健康"
    else
        print_error "✗ Redis 不健康"
    fi
    
    # 检查OnlyOffice
    if curl -s http://localhost:8000/healthcheck > /dev/null 2>&1; then
        print_success "✓ OnlyOffice 健康"
    else
        print_warning "⚠ OnlyOffice 可能未完全启动"
    fi
    
    # 检查GitLab
    if curl -s http://localhost/-/health > /dev/null 2>&1; then
        print_success "✓ GitLab 健康"
    else
        print_warning "⚠ GitLab 可能未完全启动"
    fi
    
    echo ""
}

# 检查资源使用情况
check_resource_usage() {
    echo "💾 资源使用情况"
    echo "===================="
    
    # 检查磁盘使用
    echo "磁盘使用："
    df -h | grep -E "^/dev"
    echo ""
    
    # 检查内存使用
    echo "内存使用："
    free -h
    echo ""
    
    # 检查Docker容器资源使用
    echo "Docker容器资源使用："
    docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}" 2>/dev/null || echo "无法获取容器统计信息"
    echo ""
}

# 检查网络连接
check_network_connectivity() {
    echo "🌐 网络连接检查"
    echo "===================="
    
    # 检查端口是否开放
    ports=(5432 6379 8000 80)
    port_names=("PostgreSQL" "Redis" "OnlyOffice" "GitLab")
    
    for i in "${!ports[@]}"; do
        port=${ports[$i]}
        name=${port_names[$i]}
        
        if nc -z localhost $port 2>/dev/null; then
            print_success "✓ $name (端口$port) 可访问"
        else
            print_error "✗ $name (端口$port) 不可访问"
        fi
    done
    echo ""
}

# 检查日志错误
check_logs_for_errors() {
    echo "📋 最近的错误日志"
    echo "===================="
    
    # 检查各服务的错误日志
    services=("postgres" "redis" "onlyoffice" "gitlab")
    
    for service in "${services[@]}"; do
        echo "--- $service 错误日志 ---"
        docker-compose logs --tail=5 $service 2>/dev/null | grep -i "error\|failed\|exception" | head -3 || echo "无错误日志"
        echo ""
    done
}

# 显示访问信息
show_access_info() {
    echo "🔗 访问信息"
    echo "===================="
    echo "服务访问地址："
    echo "  🌐 GitLab:     http://localhost"
    echo "  📄 OnlyOffice: http://localhost:8000"
    echo "  🗄️  PostgreSQL: localhost:5432"
    echo "  🔴 Redis:      localhost:6379"
    echo ""
    echo "管理命令："
    echo "  查看所有容器: docker-compose ps"
    echo "  查看日志:     docker-compose logs [服务名]"
    echo "  重启服务:     docker-compose restart [服务名]"
    echo "  停止服务:     docker-compose down"
    echo ""
}

# 快速检查
quick_check() {
    echo "⚡ 快速健康检查"
    echo "===================="
    
    all_healthy=true
    
    # 检查容器
    containers=("gitlabex-postgres" "gitlabex-redis" "gitlabex-onlyoffice" "gitlabex-gitlab")
    for container in "${containers[@]}"; do
        if ! docker ps --filter "name=${container}" --filter "status=running" --format "table {{.Names}}" | grep -q "${container}"; then
            all_healthy=false
            break
        fi
    done
    
    # 检查服务
    if ! docker-compose exec -T postgres pg_isready -U gitlabex > /dev/null 2>&1; then
        all_healthy=false
    fi
    
    if ! docker-compose exec -T redis redis-cli -a password123 ping > /dev/null 2>&1; then
        all_healthy=false
    fi
    
    if $all_healthy; then
        print_success "✓ 系统运行正常"
    else
        print_error "✗ 系统存在问题，请运行详细检查"
    fi
    echo ""
}

# 主函数
main() {
    echo "GitLabEx Community System Monitor"
    echo "=================================="
    echo ""
    
    case "${1:-full}" in
        "quick")
            quick_check
            ;;
        "containers")
            check_container_status
            ;;
        "health")
            check_service_health
            ;;
        "resources")
            check_resource_usage
            ;;
        "network")
            check_network_connectivity
            ;;
        "logs")
            check_logs_for_errors
            ;;
        "full")
            quick_check
            check_container_status
            check_service_health
            check_network_connectivity
            check_resource_usage
            check_logs_for_errors
            show_access_info
            ;;
        *)
            echo "用法: $0 [quick|containers|health|resources|network|logs|full]"
            echo ""
            echo "选项:"
            echo "  quick      - 快速健康检查"
            echo "  containers - 检查容器状态"
            echo "  health     - 检查服务健康状态"
            echo "  resources  - 检查资源使用情况"
            echo "  network    - 检查网络连接"
            echo "  logs       - 检查错误日志"
            echo "  full       - 完整检查 (默认)"
            ;;
    esac
}

# 执行主函数
main "$@" 