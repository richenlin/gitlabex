#!/bin/bash

# GitLabEx API驱动的测试数据初始化脚本
# 通过调用系统API创建测试数据，包括GitLab项目集成

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置参数
API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
GITLAB_URL="${GITLAB_URL:-http://localhost:8081}"
FRONTEND_URL="${FRONTEND_URL:-http://localhost:3000}"

# 测试数据配置
declare -A TEST_USERS=(
    ["gitlabex_admin"]="admin@gitlabex.com:系统管理员:admin"
    ["prof_wang"]="wang.prof@university.edu:王明教授:teacher"
    ["prof_li"]="li.prof@university.edu:李华教授:teacher"
    ["prof_zhang"]="zhang.prof@university.edu:张伟教授:teacher"
    ["ta_chen"]="chen.ta@university.edu:陈小明助教:assistant"
    ["ta_wu"]="wu.ta@university.edu:吴晓丽助教:assistant"
    ["student_001"]="student001@university.edu:张三:student"
    ["student_002"]="student002@university.edu:李四:student"
    ["student_003"]="student003@university.edu:王五:student"
    ["student_004"]="student004@university.edu:赵六:student"
)

# 全局变量
GITLAB_TOKEN=""
ADMIN_JWT_TOKEN=""
USER_TOKENS=()

# 工具函数
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖工具..."
    
    if ! command -v curl &> /dev/null; then
        log_error "curl 未安装，请先安装 curl"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        log_error "jq 未安装，请先安装 jq"
        exit 1
    fi
    
    log_success "依赖检查完成"
}

# 检查服务状态
check_services() {
    log_info "检查服务状态..."
    
    # 检查后端API
    if ! curl -s "$API_BASE_URL/health" > /dev/null 2>&1; then
        log_error "后端API服务未启动，请先启动后端服务"
        log_info "启动命令: cd backend && go run cmd/main.go"
        exit 1
    fi
    
    # 检查GitLab
    if ! curl -s "$GITLAB_URL" > /dev/null 2>&1; then
        log_error "GitLab服务未启动，请先启动GitLab"
        log_info "启动命令: docker-compose up -d gitlab"
        exit 1
    fi
    
    log_success "服务状态检查完成"
}

# 从配置文件加载GitLab令牌
load_gitlab_token_from_config() {
    local config_file="config/oauth.env"
    
    if [ ! -f "$config_file" ]; then
        log_error "配置文件 $config_file 不存在"
        return 1
    fi
    
    # 读取配置文件中的令牌
    if [ -f "$config_file" ]; then
        source "$config_file"
        GITLAB_TOKEN="$GITLAB_ACCESS_TOKEN"
    fi
    
    if [ -z "$GITLAB_TOKEN" ]; then
        log_warning "配置文件中未找到GitLab访问令牌"
        return 1
    fi
    
    return 0
}

# 获取GitLab访问令牌
get_gitlab_token() {
    echo
    log_info "GitLab访问令牌配置"
    
    # 首先尝试从配置文件加载
    if load_gitlab_token_from_config; then
        log_info "从配置文件 config/oauth.env 加载GitLab访问令牌"
        
        # 验证令牌
        log_info "验证GitLab令牌..."
        response=$(curl -s -H "Authorization: Bearer $GITLAB_TOKEN" "$GITLAB_URL/api/v4/user")
        
        if echo "$response" | jq -e '.id' > /dev/null 2>&1; then
            username=$(echo "$response" | jq -r '.username')
            log_success "GitLab令牌验证成功，用户: $username"
            return 0
        else
            log_error "配置文件中的GitLab令牌验证失败"
            log_info "错误信息: $(echo "$response" | jq -r '.message // "未知错误"')"
            log_warning "将切换到手动输入模式"
        fi
    else
        log_info "配置文件中未配置GitLab访问令牌，切换到手动输入模式"
    fi
    
    # 手动输入模式
    echo ""
    echo "⚠️  注意：这里需要的是GitLab的个人访问令牌（Personal Access Token），不是JWT令牌！"
    echo ""
    echo "🔑 GitLab个人访问令牌获取步骤："
    echo "1. 访问 $GITLAB_URL/-/profile/personal_access_tokens"
    echo "2. 使用root账户登录GitLab（初始密码在docker日志中）"
    echo "3. 点击'Add new token'"
    echo "4. 令牌名称：gitlabex-test-data"
    echo "5. 权限范围选择: api, read_user, read_repository, write_repository"
    echo "6. 点击'Create personal access token'"
    echo "7. 复制生成的令牌（格式类似：glpat-xxxxxxxxxxxxxxxxxxxx）"
    echo ""
    echo "💡 提示：您也可以将令牌保存到 config/oauth.env 文件的 GITLAB_ACCESS_TOKEN 字段中"
    echo ""
    echo "❌ 错误示例：JWT令牌（以eyJ开头的长字符串）"
    echo "✅ 正确示例：glpat-xxxxxxxxxxxxxxxxxxxx"
    echo ""
    
    while true; do
        read -p "请输入GitLab个人访问令牌: " -s GITLAB_TOKEN
        echo
        
        if [ -z "$GITLAB_TOKEN" ]; then
            log_warning "令牌不能为空，请重新输入"
            continue
        fi
        
        # 检查令牌格式
        if [[ "$GITLAB_TOKEN" == eyJ* ]]; then
            log_error "您输入的是JWT令牌，这里需要GitLab个人访问令牌！"
            log_info "JWT令牌是系统内部使用的，GitLab个人访问令牌通常以'glpat-'开头"
            echo "请按照上述步骤获取正确的GitLab个人访问令牌"
            continue
        fi
        
        # 验证令牌
        log_info "验证GitLab令牌..."
        response=$(curl -s -H "Authorization: Bearer $GITLAB_TOKEN" "$GITLAB_URL/api/v4/user")
        
        if echo "$response" | jq -e '.id' > /dev/null 2>&1; then
            username=$(echo "$response" | jq -r '.username')
            log_success "GitLab令牌验证成功，用户: $username"
            
            # 询问是否保存到配置文件
            echo ""
            read -p "是否将此令牌保存到配置文件 config/oauth.env 中？(y/N): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                save_token_to_config "$GITLAB_TOKEN"
            fi
            
            break
        else
            log_error "GitLab令牌验证失败，请检查令牌是否正确"
            log_info "错误信息: $(echo "$response" | jq -r '.message // "未知错误"')"
            echo ""
            echo "🔧 故障排除："
            echo "1. 确认GitLab服务正在运行：$GITLAB_URL"
            echo "2. 确认令牌权限包含'api'范围"
            echo "3. 确认令牌未过期"
            echo "4. 确认使用的是管理员账户创建的令牌"
            echo ""
        fi
    done
}

# 保存令牌到配置文件
save_token_to_config() {
    local token="$1"
    local config_file="config/oauth.env"
    
    if [ ! -f "$config_file" ]; then
        log_error "配置文件 $config_file 不存在"
        return 1
    fi
    
    # 使用sed替换GITLAB_ACCESS_TOKEN的值
    if sed -i "s/^GITLAB_ACCESS_TOKEN=.*/GITLAB_ACCESS_TOKEN=$token/" "$config_file"; then
        log_success "GitLab访问令牌已保存到配置文件"
    else
        log_warning "保存令牌到配置文件失败"
    fi
}

# 创建GitLab用户
create_gitlab_users() {
    log_info "创建GitLab用户..."
    
    for username in "${!TEST_USERS[@]}"; do
        IFS=':' read -r email name role <<< "${TEST_USERS[$username]}"
        
        log_info "创建GitLab用户: $username ($name)"
        
        # 检查用户是否已存在
        existing_user=$(curl -s -H "Authorization: Bearer $GITLAB_TOKEN" \
            "$GITLAB_URL/api/v4/users?username=$username")
        
        if echo "$existing_user" | jq -e '.[0].id' > /dev/null 2>&1; then
            log_warning "GitLab用户 $username 已存在，跳过创建"
            continue
        fi
        
        # 创建用户 - 使用高度随机的复杂密码满足GitLab安全策略
        user_data=$(cat <<EOF
{
    "email": "$email",
    "username": "$username", 
    "name": "$name",
    "password": "Kx9#mP2$vL8@nQ5!",
    "skip_confirmation": true
}
EOF
)
        
        response=$(curl -s -X POST \
            -H "Authorization: Bearer $GITLAB_TOKEN" \
            -H "Content-Type: application/json" \
            -d "$user_data" \
            "$GITLAB_URL/api/v4/users")
        
        if echo "$response" | jq -e '.id' > /dev/null 2>&1; then
            gitlab_id=$(echo "$response" | jq -r '.id')
            log_success "GitLab用户 $username 创建成功 (ID: $gitlab_id)"
        else
            log_error "GitLab用户 $username 创建失败"
            log_info "错误信息: $(echo "$response" | jq -r '.message // "未知错误"')"
        fi
    done
}

# 生成系统JWT令牌
generate_system_tokens() {
    log_info "为GitLab用户生成系统JWT令牌..."
    
    for username in "${!TEST_USERS[@]}"; do
        IFS=':' read -r email name role <<< "${TEST_USERS[$username]}"
        
        log_info "为用户 $username 生成JWT令牌"
        
        # 获取GitLab用户信息
        gitlab_user=$(curl -s -H "Authorization: Bearer $GITLAB_TOKEN" \
            "$GITLAB_URL/api/v4/users?username=$username")
        
        if ! echo "$gitlab_user" | jq -e '.[0].id' > /dev/null 2>&1; then
            log_error "无法获取GitLab用户 $username 的信息"
            continue
        fi
        
        # 生成正确的JWT令牌
        # JWT Header
        header='{"alg":"HS256","typ":"JWT"}'
        header_b64=$(echo -n "$header" | base64 -w 0 | tr -d '=' | tr '/+' '_-')
        
        # JWT Payload
        current_time=$(date +%s)
        exp_time=$((current_time + 86400))  # 24小时后过期
        
        payload=$(cat <<EOF
{
    "gitlab_access_token": "$GITLAB_TOKEN",
    "iss": "gitlabex",
    "exp": $exp_time,
    "nbf": $current_time,
    "iat": $current_time
}
EOF
)
        payload_b64=$(echo -n "$payload" | base64 -w 0 | tr -d '=' | tr '/+' '_-')
        
        # 生成签名（使用正确的JWT密钥）
        signature_data="${header_b64}.${payload_b64}"
        signature=$(echo -n "$signature_data" | openssl dgst -sha256 -hmac "gitlabex_super_secret_jwt_key_make_it_long_and_random_for_production_use" -binary | base64 -w 0 | tr -d '=' | tr '/+' '_-')
        
        jwt_token="${header_b64}.${payload_b64}.${signature}"
        
        log_success "为用户 $username 生成JWT令牌成功"
        
        # 保存管理员令牌
        if [ "$username" = "gitlabex_admin" ]; then
            ADMIN_JWT_TOKEN="$jwt_token"
        fi
        
        # 保存用户令牌
        USER_TOKENS+=("$username:$jwt_token")
    done
}

# 创建课题项目
create_projects() {
    log_info "创建课题项目..."
    
    if [ -z "$ADMIN_JWT_TOKEN" ]; then
        log_error "管理员令牌未获取，无法创建项目"
        return 1
    fi
    
    # 课题数据
    declare -a projects=(
        "Starlink星座分析:SpaceX Starlink卫星星座的轨道分析和性能评估研究:true:prof_wang"
        "卫星通信链路优化:针对LEO/MEO/GEO不同轨道高度卫星通信链路的优化算法研究:true:prof_li"
        "轨道机动计算器:卫星轨道机动计算与仿真工具开发:true:prof_wang"
        "火箭回收技术研究:SpaceX猎鹰9号式火箭垂直回收技术的建模与仿真分析:true:prof_zhang"
        "LEO轨道碰撞预警:低轨卫星碰撞风险评估与预警系统:false:prof_li"
        "深度学习图像识别:基于卷积神经网络的卫星图像目标识别系统:true:prof_wang"
        "自然语言处理平台:多语言文本分析和情感分析平台开发:true:prof_li"
        "强化学习游戏AI:基于深度Q网络的游戏AI智能体开发:true:prof_zhang"
        "微服务架构实践:基于Spring Cloud的微服务架构设计与实现:true:prof_zhang"
        "React全栈开发:基于React + Node.js的全栈Web应用开发:true:prof_wang"
        "大数据分析平台:基于Hadoop和Spark的大数据处理平台构建:true:prof_li"
        "机器学习预测模型:时间序列预测和回归分析模型开发:true:prof_zhang"
    )
    
    for project_info in "${projects[@]}"; do
        IFS=':' read -r name description is_public creator <<< "$project_info"
        
        log_info "创建课题: $name"
        
        # 获取创建者令牌
        creator_token=""
        for user_token in "${USER_TOKENS[@]}"; do
            IFS=':' read -r token_username token <<< "$user_token"
            if [ "$token_username" = "$creator" ]; then
                creator_token="$token"
                break
            fi
        done
        
        if [ -z "$creator_token" ]; then
            log_error "无法获取用户 $creator 的令牌"
            continue
        fi
        
        # 创建项目
        project_data=$(cat <<EOF
{
    "name": "$name",
    "description": "$description",
    "is_public": $is_public,
    "tags": ["测试数据", "自动生成"]
}
EOF
)
        
        response=$(curl -s -X POST \
            -H "Authorization: Bearer $creator_token" \
            -H "Content-Type: application/json" \
            -d "$project_data" \
            "$API_BASE_URL/api/v1/research-projects")
        
        # 检查响应是否为有效JSON
        if echo "$response" | jq . > /dev/null 2>&1; then
            # 检查是否有错误信息
            if echo "$response" | jq -e '.error' > /dev/null 2>&1; then
                log_error "课题 $name 创建失败"
                log_info "错误信息: $(echo "$response" | jq -r '.error // .message // "未知错误"')"
            elif echo "$response" | jq -e '.id' > /dev/null 2>&1; then
                # 项目创建成功，直接返回项目对象
                project_id=$(echo "$response" | jq -r '.id')
                log_success "课题 $name 创建成功 (ID: $project_id)"
                
                # 添加项目成员
                add_project_members "$project_id" "$creator_token"
            else
                log_error "课题 $name 创建失败"
                log_info "API响应格式异常: $response"
            fi
        else
            log_error "课题 $name 创建失败"
            log_info "API响应不是有效的JSON: $response"
            log_info "可能的原因: 认证失败、服务器错误或API端点不存在"
        fi
    done
}

# 添加项目成员
add_project_members() {
    local project_id="$1"
    local creator_token="$2"
    
    # 随机添加一些成员
    local members=("ta_chen" "ta_wu" "student_001" "student_002" "student_003")
    local roles=("maintainer" "developer" "developer" "reporter" "reporter")
    
    for i in "${!members[@]}"; do
        local member="${members[$i]}"
        local role="${roles[$i]}"
        
        log_info "添加成员 $member 到项目 (角色: $role)"
        
        member_data=$(cat <<EOF
{
    "username": "$member",
    "access_level": 30
}
EOF
)
        
        response=$(curl -s -X POST \
            -H "Authorization: Bearer $creator_token" \
            -H "Content-Type: application/json" \
            -d "$member_data" \
            "$API_BASE_URL/api/v1/research-projects/$project_id/members")
        
        if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
            log_success "成员 $member 添加成功"
        else
            log_warning "成员 $member 添加失败: $(echo "$response" | jq -r '.error // "未知错误"')"
        fi
    done
}

# 创建话题讨论
create_topics() {
    log_info "创建话题讨论..."
    
    # 获取项目列表
    response=$(curl -s -H "Authorization: Bearer $ADMIN_JWT_TOKEN" \
        "$API_BASE_URL/api/v1/research-projects?page=1&limit=20")
    
    if ! echo "$response" | jq -e '.projects' > /dev/null 2>&1; then
        log_error "无法获取项目列表"
        return 1
    fi
    
    # 话题模板
    declare -a topics=(
        "轨道参数分析方法讨论:大家对于卫星轨道参数分析有什么好的方法推荐吗？"
        "性能优化策略分享:在项目开发中遇到的性能问题和解决方案"
        "算法实现讨论:关于核心算法的实现细节和优化思路"
        "技术难点求助:项目中遇到的技术难点，希望大家帮忙解决"
        "经验分享:项目开发过程中的经验和教训分享"
    )
    
    # 为每个项目创建话题
    echo "$response" | jq -r '.projects[].id' | head -5 | while read -r project_id; do
        for topic_info in "${topics[@]}"; do
            IFS=':' read -r title content <<< "$topic_info"
            
            log_info "为项目 $project_id 创建话题: $title"
            
            # 随机选择一个用户作为作者
            author_tokens=("${USER_TOKENS[@]}")
            random_index=$((RANDOM % ${#author_tokens[@]}))
            IFS=':' read -r author_username author_token <<< "${author_tokens[$random_index]}"
            
            topic_data=$(cat <<EOF
{
    "title": "$title",
    "content": "$content",
    "project_id": "$project_id",
    "labels": ["讨论", "技术"]
}
EOF
)
            
            response=$(curl -s -X POST \
                -H "Authorization: Bearer $author_token" \
                -H "Content-Type: application/json" \
                -d "$topic_data" \
                "$API_BASE_URL/api/v1/topics")
            
            # 检查响应是否为有效JSON
            if echo "$response" | jq . > /dev/null 2>&1; then
                # 检查是否有错误信息
                if echo "$response" | jq -e '.error' > /dev/null 2>&1; then
                    log_warning "话题创建失败: $(echo "$response" | jq -r '.error // .message // "未知错误"')"
                elif echo "$response" | jq -e '.id' > /dev/null 2>&1; then
                    # 话题创建成功，直接返回话题对象
                    topic_id=$(echo "$response" | jq -r '.id')
                    log_success "话题创建成功 (ID: $topic_id)"
                else
                    log_warning "话题创建失败: API响应格式异常"
                fi
            else
                log_warning "话题创建失败: API响应不是有效的JSON: $response"
            fi
        done
    done
}

# 创建作业
create_homeworks() {
    log_info "创建课程作业..."
    
    # 获取项目列表
    response=$(curl -s -H "Authorization: Bearer $ADMIN_JWT_TOKEN" \
        "$API_BASE_URL/api/v1/research-projects?page=1&limit=10")
    
    if ! echo "$response" | jq -e '.projects' > /dev/null 2>&1; then
        log_error "无法获取项目列表"
        return 1
    fi
    
    # 作业模板
    declare -a homeworks=(
        "项目分析报告:完成项目的详细分析报告，包含技术方案和实现细节"
        "算法实现作业:实现项目中的核心算法，并提交源代码和测试结果"
        "系统设计文档:设计完整的系统架构，包含模块划分和接口定义"
        "性能测试报告:对系统进行性能测试，分析瓶颈并提出优化方案"
    )
    
    # 为前几个项目创建作业
    echo "$response" | jq -r '.projects[].id' | head -3 | while read -r project_id; do
        for homework_info in "${homeworks[@]}"; do
            IFS=':' read -r title description <<< "$homework_info"
            
            log_info "为项目 $project_id 创建作业: $title"
            
            # 使用教师令牌创建作业
            teacher_token=""
            for user_token in "${USER_TOKENS[@]}"; do
                IFS=':' read -r token_username token <<< "$user_token"
                if [[ "$token_username" == prof_* ]]; then
                    teacher_token="$token"
                    break
                fi
            done
            
            if [ -z "$teacher_token" ]; then
                log_warning "无法获取教师令牌，跳过作业创建"
                continue
            fi
            
            # 设置截止日期（30天后）
            due_date=$(date -d "+30 days" -Iseconds)
            
            homework_data=$(cat <<EOF
{
    "title": "$title",
    "description": "$description",
    "project_id": "$project_id",
    "deadline": "$due_date",
    "max_grade": 100
}
EOF
)
            
            response=$(curl -s -X POST \
                -H "Authorization: Bearer $teacher_token" \
                -H "Content-Type: application/json" \
                -d "$homework_data" \
                "$API_BASE_URL/api/v1/homework")
            
            # 检查响应是否为有效JSON
            if echo "$response" | jq . > /dev/null 2>&1; then
                # 检查是否有错误信息
                if echo "$response" | jq -e '.error' > /dev/null 2>&1; then
                    log_warning "作业创建失败: $(echo "$response" | jq -r '.error // .message // "未知错误"')"
                elif echo "$response" | jq -e '.id' > /dev/null 2>&1; then
                    # 作业创建成功，直接返回作业对象
                    homework_id=$(echo "$response" | jq -r '.id')
                    log_success "作业创建成功 (ID: $homework_id)"
                else
                    log_warning "作业创建失败: API响应格式异常"
                fi
            else
                log_warning "作业创建失败: API响应不是有效的JSON: $response"
            fi
        done
    done
}

# 显示统计信息
show_statistics() {
    log_info "获取数据统计..."
    
    # 获取各类数据统计
    users_count=$(curl -s -H "Authorization: Bearer $ADMIN_JWT_TOKEN" \
        "$API_BASE_URL/api/v1/admin/users?page=1&limit=1" | jq -r '.pagination.total // 0')
    
    projects_count=$(curl -s -H "Authorization: Bearer $ADMIN_JWT_TOKEN" \
        "$API_BASE_URL/api/v1/research-projects?page=1&limit=1" | jq -r '.pagination.total // 0')
    
    topics_count=$(curl -s -H "Authorization: Bearer $ADMIN_JWT_TOKEN" \
        "$API_BASE_URL/api/v1/topics?page=1&limit=1" | jq -r '.pagination.total // 0')
    
    echo
    echo "=============================================="
    log_success "GitLabEx API测试数据初始化完成！"
    echo "=============================================="
    echo "📊 数据统计："
    echo "   用户: $users_count 个"
    echo "   课题: $projects_count 个" 
    echo "   话题: $topics_count 个"
    echo ""
    echo "🔐 测试账号信息："
    echo "   管理员: gitlabex_admin / Kx9#mP2$vL8@nQ5!"
    echo "   教师: prof_wang / Kx9#mP2$vL8@nQ5!"
    echo "   教师: prof_li / Kx9#mP2$vL8@nQ5!"
    echo "   助教: ta_chen / Kx9#mP2$vL8@nQ5!"
    echo "   学生: student_001 / Kx9#mP2$vL8@nQ5!"
    echo ""
    echo "🌐 访问地址："
    echo "   前端应用: $FRONTEND_URL"
    echo "   后端API: $API_BASE_URL"
    echo "   GitLab: $GITLAB_URL"
    echo ""
    echo "💡 功能特性："
    echo "   • 完整的GitLab项目集成"
    echo "   • 真实的用户权限管理"
    echo "   • 丰富的测试数据内容"
    echo "   • 支持课题、话题、作业等功能"
    echo "=============================================="
}

# 主函数
main() {
    echo "🚀 GitLabEx API驱动测试数据初始化"
    echo "======================================"
    
    # 检查依赖和服务
    check_dependencies
    check_services
    
    # 获取GitLab令牌
    get_gitlab_token
    
    # 创建用户
    create_gitlab_users
    generate_system_tokens
    
    # 创建内容
    create_projects
    sleep 2  # 等待项目创建完成
    create_topics
    sleep 2  # 等待话题创建完成
    create_homeworks
    
    # 显示结果
    show_statistics
    
    log_success "测试数据初始化完成！"
}

# 错误处理
trap 'log_error "脚本执行失败，请检查错误信息"; exit 1' ERR

# 执行主函数
main "$@"
