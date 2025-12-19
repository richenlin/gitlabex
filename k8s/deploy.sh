#!/bin/bash

# GitLabEx Kubernetes 部署脚本
# 用于快速部署 GitLabEx 到 Kubernetes 集群

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

# 检查必要的工具
check_requirements() {
    info "检查必要的工具..."
    
    if ! command -v kubectl &> /dev/null; then
        error "kubectl 未安装，请先安装 kubectl"
        exit 1
    fi
    
    # 检查 kubectl 是否能连接到集群
    if ! kubectl cluster-info &> /dev/null; then
        error "无法连接到 Kubernetes 集群，请检查 kubeconfig 配置"
        exit 1
    fi
    
    info "所有必要工具检查通过"
}

# 创建命名空间
create_namespace() {
    info "创建命名空间..."
    kubectl apply -f namespace.yaml
    info "命名空间创建成功"
}

# 创建 Secrets
create_secrets() {
    info "创建 Secrets..."
    
    if [ ! -f "secrets.yaml" ]; then
        error "secrets.yaml 文件不存在！"
        warn "请先复制 secrets.yaml.example 为 secrets.yaml 并填写实际的配置信息："
        echo "  cp secrets.yaml.example secrets.yaml"
        echo "  vim secrets.yaml"
        exit 1
    fi
    
    kubectl apply -f secrets.yaml
    info "Secrets 创建成功"
}

# 部署数据库服务
deploy_databases() {
    info "部署数据库服务..."
    kubectl apply -f postgres.yaml
    kubectl apply -f redis.yaml
    info "数据库服务部署成功"
}

# 部署存储服务
deploy_storage() {
    info "部署存储服务..."
    kubectl apply -f minio.yaml
    info "存储服务部署成功"
}

# 部署 GitLab
deploy_gitlab() {
    info "部署 GitLab 服务..."
    kubectl apply -f gitlab.yaml
    info "GitLab 服务部署成功"
}

# 部署后端服务
deploy_backend() {
    info "部署后端服务..."
    kubectl apply -f backend.yaml
    info "后端服务部署成功"
}

# 部署前端服务
deploy_frontend() {
    info "部署前端服务..."
    kubectl apply -f frontend.yaml
    info "前端服务部署成功"
}

# 部署 Ingress (可选)
deploy_ingress() {
    if [ -f "ingress.yaml" ]; then
        read -p "是否部署 Ingress 配置？(y/N): " deploy_ing
        if [[ $deploy_ing =~ ^[Yy]$ ]]; then
            info "部署 Ingress..."
            kubectl apply -f ingress.yaml
            info "Ingress 部署成功"
        fi
    fi
}

# 等待 Pod 就绪
wait_for_pods() {
    info "等待所有 Pod 就绪..."
    
    # 等待数据库就绪
    info "等待 PostgreSQL 就绪..."
    kubectl wait --for=condition=ready pod -l app=postgres -n gitlabex --timeout=300s || warn "PostgreSQL 可能未完全就绪"
    
    info "等待 Redis 就绪..."
    kubectl wait --for=condition=ready pod -l app=redis -n gitlabex --timeout=300s || warn "Redis 可能未完全就绪"
    
    info "等待 MinIO 就绪..."
    kubectl wait --for=condition=ready pod -l app=minio -n gitlabex --timeout=300s || warn "MinIO 可能未完全就绪"
    
    # GitLab 启动较慢，需要更长时间
    info "等待 GitLab 就绪（这可能需要5-10分钟）..."
    kubectl wait --for=condition=ready pod -l app=gitlab -n gitlabex --timeout=600s || warn "GitLab 可能未完全就绪"
    
    info "等待后端服务就绪..."
    kubectl wait --for=condition=ready pod -l app=backend -n gitlabex --timeout=300s || warn "后端服务可能未完全就绪"
    
    info "等待前端服务就绪..."
    kubectl wait --for=condition=ready pod -l app=frontend -n gitlabex --timeout=300s || warn "前端服务可能未完全就绪"
}

# 显示访问信息
show_access_info() {
    info "获取服务访问信息..."
    echo ""
    echo "=========================================="
    echo "GitLabEx 部署完成！"
    echo "=========================================="
    echo ""
    
    # 获取 NodePort
    FRONTEND_PORT=$(kubectl get svc gitlabex-frontend -n gitlabex -o jsonpath='{.spec.ports[0].nodePort}')
    BACKEND_PORT=$(kubectl get svc gitlabex-backend -n gitlabex -o jsonpath='{.spec.ports[0].nodePort}')
    GITLAB_PORT=$(kubectl get svc gitlabex-gitlab -n gitlabex -o jsonpath='{.spec.ports[0].nodePort}')
    GITLAB_SSH_PORT=$(kubectl get svc gitlabex-gitlab-ssh -n gitlabex -o jsonpath='{.spec.ports[0].nodePort}')
    MINIO_CONSOLE_PORT=$(kubectl get svc gitlabex-minio-console -n gitlabex -o jsonpath='{.spec.ports[0].nodePort}')
    
    # 获取节点 IP
    NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
    
    echo "服务访问地址 (使用 NodePort)："
    echo "  前端服务:      http://${NODE_IP}:${FRONTEND_PORT}"
    echo "  后端API:       http://${NODE_IP}:${BACKEND_PORT}"
    echo "  GitLab:        http://${NODE_IP}:${GITLAB_PORT}"
    echo "  GitLab SSH:    ssh://git@${NODE_IP}:${GITLAB_SSH_PORT}"
    echo "  MinIO控制台:   http://${NODE_IP}:${MINIO_CONSOLE_PORT}"
    echo ""
    echo "默认凭据:"
    echo "  GitLab:"
    echo "    用户名: root"
    echo "    密码: (参考 secrets.yaml 中的 gitlab-root-password)"
    echo ""
    echo "  MinIO:"
    echo "    用户名: admin"
    echo "    密码: (参考 secrets.yaml 中的 minio-root-password)"
    echo ""
    echo "=========================================="
    echo "下一步操作:"
    echo "=========================================="
    echo ""
    echo "1. 访问 GitLab 并使用 root 用户登录"
    echo "2. 运行 OAuth 配置脚本："
    echo "   ./configure-oauth-k8s.sh"
    echo ""
    echo "3. 查看 Pod 状态："
    echo "   kubectl get pods -n gitlabex"
    echo ""
    echo "4. 查看 Pod 日志："
    echo "   kubectl logs -f <pod-name> -n gitlabex"
    echo ""
    echo "5. 查看服务状态："
    echo "   kubectl get svc -n gitlabex"
    echo ""
}

# 主函数
main() {
    info "开始部署 GitLabEx 到 Kubernetes..."
    echo ""
    
    # 检查环境
    check_requirements
    
    # 显示部署计划
    info "部署计划:"
    echo "  1. 创建命名空间"
    echo "  2. 创建 Secrets"
    echo "  3. 部署数据库服务 (PostgreSQL, Redis)"
    echo "  4. 部署存储服务 (MinIO)"
    echo "  5. 部署 GitLab"
    echo "  6. 部署后端服务"
    echo "  7. 部署前端服务"
    echo "  8. (可选) 部署 Ingress"
    echo ""
    
    read -p "是否继续？(y/N): " confirm
    if [[ ! $confirm =~ ^[Yy]$ ]]; then
        warn "用户取消部署"
        exit 0
    fi
    
    # 执行部署
    create_namespace
    create_secrets
    deploy_databases
    deploy_storage
    deploy_gitlab
    deploy_backend
    deploy_frontend
    deploy_ingress
    
    # 等待服务就绪
    wait_for_pods
    
    # 显示访问信息
    show_access_info
    
    info "部署完成！"
}

# 运行主函数
main
