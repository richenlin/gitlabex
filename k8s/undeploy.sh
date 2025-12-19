#!/bin/bash

# GitLabEx Kubernetes 卸载脚本
# 用于从 Kubernetes 集群中删除 GitLabEx

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印信息函数
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 警告并确认
confirm_deletion() {
    warn "此操作将删除 GitLabEx 的所有资源，包括："
    echo "  - 所有部署的应用 (Frontend, Backend, GitLab, MinIO, PostgreSQL, Redis)"
    echo "  - 所有持久化数据 (数据库数据、文件存储、日志等)"
    echo "  - 所有配置和密钥"
    echo "  - gitlabex 命名空间"
    echo ""
    error "此操作不可逆！所有数据将被永久删除！"
    echo ""
    
    read -p "确定要继续吗？请输入 'yes' 确认: " confirm
    if [ "$confirm" != "yes" ]; then
        warn "用户取消操作"
        exit 0
    fi
}

# 删除应用
delete_applications() {
    info "删除应用服务..."
    
    kubectl delete -f frontend.yaml --ignore-not-found=true
    kubectl delete -f backend.yaml --ignore-not-found=true
    kubectl delete -f gitlab.yaml --ignore-not-found=true
    kubectl delete -f minio.yaml --ignore-not-found=true
    kubectl delete -f redis.yaml --ignore-not-found=true
    kubectl delete -f postgres.yaml --ignore-not-found=true
    
    info "应用服务删除成功"
}

# 删除 Ingress
delete_ingress() {
    if [ -f "ingress.yaml" ]; then
        info "删除 Ingress..."
        kubectl delete -f ingress.yaml --ignore-not-found=true
        info "Ingress 删除成功"
    fi
}

# 删除 Secrets
delete_secrets() {
    info "删除 Secrets..."
    kubectl delete -f secrets.yaml --ignore-not-found=true
    info "Secrets 删除成功"
}

# 删除命名空间
delete_namespace() {
    info "删除命名空间..."
    kubectl delete -f namespace.yaml --ignore-not-found=true
    
    # 等待命名空间完全删除
    info "等待命名空间完全删除..."
    timeout=60
    while kubectl get namespace gitlabex &> /dev/null; do
        echo -n "."
        sleep 2
        timeout=$((timeout - 2))
        if [ $timeout -le 0 ]; then
            warn "命名空间删除超时，可能存在资源无法正常清理"
            break
        fi
    done
    echo ""
    
    info "命名空间删除成功"
}

# 主函数
main() {
    info "开始卸载 GitLabEx..."
    echo ""
    
    # 确认删除
    confirm_deletion
    
    # 执行删除
    delete_ingress
    delete_applications
    delete_secrets
    delete_namespace
    
    info "GitLabEx 卸载完成！"
}

# 运行主函数
main
