#!/bin/bash

# 设置日志颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
  echo -e "[$(date '+%Y-%m-%d %H:%M:%S')] ${GREEN}[INFO]${NC} $1"
}

log_warn() {
  echo -e "[$(date '+%Y-%m-%d %H:%M:%S')] ${YELLOW}[WARN]${NC} $1"
}

log_error() {
  echo -e "[$(date '+%Y-%m-%d %H:%M:%S')] ${RED}[ERROR]${NC} $1"
}

# 获取配置文件路径
CONFIG_FILE=${OAUTH_CONFIG_FILE:-"/config/oauth.env"}
SHARED_DIR=${SHARED_DIR:-"/shared"}
GITLAB_CONTAINER=${GITLAB_CONTAINER:-"gitlabex-gitlab"}

# 检查配置文件是否存在
if [ ! -f "$CONFIG_FILE" ]; then
  log_error "配置文件不存在: $CONFIG_FILE"
  exit 1
fi

# 加载配置
log_info "加载配置文件: $CONFIG_FILE"
source "$CONFIG_FILE"

# 验证必要的配置项
REQUIRED_CONFIGS=("GITLAB_INTERNAL_URL" "GITLAB_EXTERNAL_URL" "GITLAB_OAUTH_REDIRECT_URI" "GITLAB_OAUTH_APP_NAME")
for config in "${REQUIRED_CONFIGS[@]}"; do
  if [ -z "${!config}" ]; then
    log_error "缺少必需的配置项: $config"
    exit 1
  fi
done

# 等待GitLab服务就绪
wait_for_gitlab() {
  local max_attempts=60
  local attempt=0
  local wait_time=10
  
  log_info "等待GitLab服务就绪..."
  
  # 初始等待，给GitLab一些启动时间
  sleep 30
  
  while [ $attempt -lt $max_attempts ]; do
    if docker exec $GITLAB_CONTAINER gitlab-rake gitlab:check > /dev/null 2>&1; then
      log_info "GitLab服务已就绪"
      return 0
    fi
    
    attempt=$((attempt + 1))
    log_warn "尝试 $attempt/$max_attempts: GitLab服务未就绪，等待 $wait_time 秒后重试..."
    sleep $wait_time
  done
  
  log_error "GitLab服务未在预期时间内就绪"
  return 1
}

# 创建OAuth应用
create_oauth_application() {
  log_info "创建OAuth应用: $GITLAB_OAUTH_APP_NAME"
  
  # 构建Ruby代码
  read -r -d '' RUBY_CODE << EOF
# 查找或创建OAuth应用
app = Doorkeeper::Application.find_or_create_by(name: '$GITLAB_OAUTH_APP_NAME') do |a|
  a.redirect_uri = '$GITLAB_OAUTH_REDIRECT_URI'
  a.scopes = 'api read_user email'
  a.confidential = true
  a.owner = User.find_by(username: 'root')
end

# 输出应用信息
puts "APPLICATION_ID=#{app.uid}"
puts "SECRET=#{app.secret}"
puts "CREATED_AT=#{app.created_at}"
puts "UPDATED_AT=#{app.updated_at}"
EOF

  # 执行Ruby代码
  local result
  result=$(docker exec $GITLAB_CONTAINER gitlab-rails runner "$RUBY_CODE" 2>&1)
  
  # 检查是否成功
  if [ $? -ne 0 ]; then
    log_error "创建OAuth应用失败"
    log_error "$result"
    return 1
  fi
  
  # 提取应用信息
  APPLICATION_ID=$(echo "$result" | grep "APPLICATION_ID=" | cut -d'=' -f2)
  SECRET=$(echo "$result" | grep "SECRET=" | cut -d'=' -f2)
  
  if [ -z "$APPLICATION_ID" ] || [ -z "$SECRET" ]; then
    log_error "无法提取应用信息"
    log_error "$result"
    return 1
  fi
  
  log_info "成功创建OAuth应用"
  log_info "应用ID: $APPLICATION_ID"
  log_info "Client Secret: ${SECRET:0:10}..."
  
  return 0
}

# 保存OAuth配置
save_oauth_config() {
  # 创建输出目录
  mkdir -p "$SHARED_DIR"
  
  # 配置文件路径
  local output_file="$SHARED_DIR/gitlab-oauth.env"
  
  # 生成配置内容
  cat > "$output_file" << EOF
GITLAB_CLIENT_ID=$APPLICATION_ID
GITLAB_CLIENT_SECRET=$SECRET
GITLAB_REDIRECT_URI=$GITLAB_OAUTH_REDIRECT_URI
GITLAB_EXTERNAL_URL=$GITLAB_EXTERNAL_URL
GITLAB_INTERNAL_URL=$GITLAB_INTERNAL_URL
GITLAB_SCOPES="api read_user email"
EOF

  # 设置权限
  chmod 644 "$output_file"
  
  log_info "配置文件已写入: $output_file"
  log_info "文件权限："
  ls -l "$output_file"
  
  return 0
}

# 主程序
main() {
  log_info "开始GitLab OAuth应用初始化（Rails版本）..."
  
  # 等待GitLab就绪
  wait_for_gitlab || exit 1
  
  # 创建OAuth应用
  create_oauth_application || exit 1
  
  # 保存OAuth配置
  save_oauth_config || exit 1
  
  log_info "GitLab OAuth应用初始化完成"
}

# 执行主程序
main
