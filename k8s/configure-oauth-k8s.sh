#!/bin/bash

# GitLab OAuth 配置脚本 (Kubernetes 版本)
# 用于在 Kubernetes 环境中配置 GitLab OAuth 应用

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

prompt() {
    echo -e "${BLUE}[INPUT]${NC} $1"
}

# 检查 kubectl
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        error "kubectl 未安装，请先安装 kubectl"
        exit 1
    fi
}

# 获取 GitLab Pod
get_gitlab_pod() {
    GITLAB_POD=$(kubectl get pods -n gitlabex -l app=gitlab -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    
    if [ -z "$GITLAB_POD" ]; then
        error "未找到 GitLab Pod，请确保 GitLab 已部署"
        exit 1
    fi
    
    info "找到 GitLab Pod: $GITLAB_POD"
}

# 检查 GitLab 是否就绪
check_gitlab_ready() {
    info "检查 GitLab 是否就绪..."
    
    if ! kubectl get pod "$GITLAB_POD" -n gitlabex -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}' | grep -q "True"; then
        error "GitLab Pod 未就绪，请等待 GitLab 完全启动"
        exit 1
    fi
    
    info "GitLab 已就绪"
}

# 获取服务访问地址
get_service_urls() {
    # 获取 NodePort
    FRONTEND_PORT=$(kubectl get svc gitlabex-frontend -n gitlabex -o jsonpath='{.spec.ports[0].nodePort}')
    GITLAB_PORT=$(kubectl get svc gitlabex-gitlab -n gitlabex -o jsonpath='{.spec.ports[0].nodePort}')
    
    # 获取节点 IP
    NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
    
    # 如果有外部 IP，优先使用外部 IP
    EXTERNAL_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}')
    if [ -n "$EXTERNAL_IP" ]; then
        NODE_IP=$EXTERNAL_IP
    fi
    
    GITLAB_URL="http://${NODE_IP}:${GITLAB_PORT}"
    CALLBACK_URL="http://${NODE_IP}:${FRONTEND_PORT}/auth/callback"
    
    info "检测到的服务地址："
    echo "  GitLab URL: $GITLAB_URL"
    echo "  回调 URL: $CALLBACK_URL"
}

# 配置向导
configuration_wizard() {
    echo ""
    echo "=========================================="
    echo "GitLab OAuth 配置向导"
    echo "=========================================="
    echo ""
    
    get_service_urls
    
    echo ""
    prompt "请按照以下步骤手动配置 OAuth 应用："
    echo ""
    echo "1. 访问 GitLab: $GITLAB_URL"
    echo "2. 使用 root 用户登录"
    echo "3. 进入 Admin Area > Applications"
    echo "4. 创建新的应用，填写以下信息："
    echo "   - Name: GitLabEx"
    echo "   - Redirect URI: $CALLBACK_URL"
    echo "   - Scopes: api, read_user, read_repository"
    echo "5. 保存后获取 Application ID 和 Secret"
    echo ""
    
    read -p "配置完成后按回车继续..."
    
    # 输入 OAuth 信息
    echo ""
    prompt "请输入 OAuth Application ID:"
    read -r CLIENT_ID
    
    prompt "请输入 OAuth Secret:"
    read -r CLIENT_SECRET
    
    # 创建 System Token
    echo ""
    info "正在创建 System Token..."
    
    # 在 GitLab Pod 中创建 Personal Access Token
    SYSTEM_TOKEN=$(kubectl exec -n gitlabex "$GITLAB_POD" -- \
        gitlab-rails runner "token = User.find_by_username('root').personal_access_tokens.create(scopes: [:api, :read_user, :read_repository], name: 'GitLabEx System Token', expires_at: 365.days.from_now); token.set_token('glpat-' + SecureRandom.hex(20)); token.save!; puts token.token" 2>/dev/null)
    
    if [ -z "$SYSTEM_TOKEN" ]; then
        error "System Token 创建失败"
        exit 1
    fi
    
    info "System Token 创建成功"
}

# 更新 Secrets
update_secrets() {
    info "更新 Kubernetes Secrets..."
    
    # Base64 编码
    CLIENT_ID_B64=$(echo -n "$CLIENT_ID" | base64)
    CLIENT_SECRET_B64=$(echo -n "$CLIENT_SECRET" | base64)
    SYSTEM_TOKEN_B64=$(echo -n "$SYSTEM_TOKEN" | base64)
    
    # 更新 Secret
    kubectl patch secret gitlabex-secrets -n gitlabex --type='json' -p="[
        {\"op\": \"replace\", \"path\": \"/data/gitlab-client-id\", \"value\": \"$CLIENT_ID_B64\"},
        {\"op\": \"replace\", \"path\": \"/data/gitlab-client-secret\", \"value\": \"$CLIENT_SECRET_B64\"},
        {\"op\": \"replace\", \"path\": \"/data/gitlab-system-token\", \"value\": \"$SYSTEM_TOKEN_B64\"}
    ]"
    
    info "Secrets 更新成功"
}

# 重启后端服务
restart_backend() {
    info "重启后端服务以应用新配置..."
    
    kubectl rollout restart deployment/backend -n gitlabex
    kubectl rollout status deployment/backend -n gitlabex --timeout=300s
    
    info "后端服务重启完成"
}

# 显示配置摘要
show_summary() {
    echo ""
    echo "=========================================="
    echo "配置完成！"
    echo "=========================================="
    echo ""
    echo "OAuth 配置信息："
    echo "  Client ID: $CLIENT_ID"
    echo "  Client Secret: ${CLIENT_SECRET:0:10}..."
    echo "  System Token: ${SYSTEM_TOKEN:0:20}..."
    echo ""
    echo "这些信息已保存到 Kubernetes Secret 中"
    echo ""
    echo "现在可以访问前端应用并使用 GitLab 登录："
    echo "  $CALLBACK_URL"
    echo ""
}

# 主函数
main() {
    info "开始配置 GitLab OAuth..."
    
    check_kubectl
    get_gitlab_pod
    check_gitlab_ready
    configuration_wizard
    update_secrets
    restart_backend
    show_summary
    
    info "OAuth 配置完成！"
}

# 运行主函数
main
