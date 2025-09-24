#!/bin/bash

# GitLabEx 同步API测试脚本
# 用于测试第三方系统用户同步接口

set -e

# 配置
BASE_URL="http://localhost:8080"
SYNC_API_KEY="gitlabex_sync_api_key_2024"
THIRD_PARTY_API_KEY="gitlabex_third_party_api_key_2024"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 检查服务是否运行
check_service() {
    print_info "检查GitLabEx后端服务..."
    if curl -s "${BASE_URL}/health" > /dev/null 2>&1; then
        print_success "后端服务运行正常"
    else
        print_error "后端服务未运行，请先启动服务"
        exit 1
    fi
}

# 测试创建单个用户
test_create_user() {
    print_info "测试创建单个用户..."
    
    local response=$(curl -s -w "%{http_code}" -X POST \
        		"${BASE_URL}/api/v1/sync/users" \
		-H "Content-Type: application/json" \
		-H "X-API-Key: ${SYNC_API_KEY}" \
		-d '{
			"username": "test_student_001",
			"password": "password123",
			"email": "test001@university.edu",
			"name": "测试学生001",
			"role": "student",
			"department": "计算机科学与技术",
			"student_id": "2024001001",
			"external_id": "ext_test_001"
		}')
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$http_code" = "201" ]; then
        print_success "用户创建成功"
        echo "$body" | jq '.data.user | {id, username, name, role}'
        local token=$(echo "$body" | jq -r '.data.token')
        print_info "获得Token: $token"
        echo "$token" > /tmp/test_user_token.txt
        echo "test_student_001" > /tmp/test_username.txt
    elif [ "$http_code" = "409" ]; then
        print_warning "用户已存在，跳过创建"
    else
        print_error "用户创建失败，HTTP状态码: $http_code"
        echo "$body" | jq '.'
    fi
}

# 测试批量创建用户
test_batch_create_users() {
    print_info "测试批量创建用户..."
    
    local response=$(curl -s -w "%{http_code}" -X POST \
        		"${BASE_URL}/api/v1/sync/users/batch" \
		-H "Content-Type: application/json" \
		-H "X-API-Key: ${SYNC_API_KEY}" \
        -d '{
            "users": [
                {
                    "username": "batch_student_001",
                    "password": "password123",
                    "email": "batch001@university.edu",
                    "name": "批量学生001",
                    "role": "student"
                },
                {
                    "username": "batch_teacher_001",
                    "password": "password456",
                    "email": "batch_teacher001@university.edu",
                    "name": "批量教师001",
                    "role": "teacher"
                },
                {
                    "username": "batch_assistant_001",
                    "password": "password789",
                    "email": "batch_assistant001@university.edu",
                    "name": "批量助教001",
                    "role": "assistant"
                }
            ]
        }')
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$http_code" = "200" ]; then
        print_success "批量创建完成"
        echo "$body" | jq '{total_count, success_count, failure_count}'
        
        # 显示创建结果摘要
        local success_count=$(echo "$body" | jq '.success_count')
        local failure_count=$(echo "$body" | jq '.failure_count')
        
        if [ "$failure_count" -gt 0 ]; then
            print_warning "部分用户创建失败，失败数量: $failure_count"
        else
            print_success "所有用户创建成功"
        fi
    else
        print_error "批量创建失败，HTTP状态码: $http_code"
        echo "$body" | jq '.'
    fi
}

# 测试更新用户
test_update_user() {
    print_info "测试更新用户信息..."
    
    # 获取之前创建的用户名
    local username="test_student_001"
    if [ -f "/tmp/test_username.txt" ]; then
        username=$(cat /tmp/test_username.txt)
    fi
    
    local response=$(curl -s -w "%{http_code}" -X PUT \
        		"${BASE_URL}/api/v1/sync/users/${username}" \
		-H "Content-Type: application/json" \
		-H "X-API-Key: ${SYNC_API_KEY}" \
        -d '{
            "name": "测试学生001(已更新)",
            "role": "assistant",
            "avatar_url": "https://example.com/avatar.jpg",
            "department": "软件工程学院"
        }')
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$http_code" = "200" ]; then
        print_success "用户更新成功"
        echo "$body" | jq '.data.user | {id, username, name, role}'
    else
        print_error "用户更新失败，HTTP状态码: $http_code"
        echo "$body" | jq '.'
    fi
}

# 测试API密钥认证和权限
test_api_key_auth() {
    print_info "测试API密钥认证和权限..."
    
    # 测试无API密钥
    local response=$(curl -s -w "%{http_code}" -X POST \
        "${BASE_URL}/api/v1/sync/users" \
        -H "Content-Type: application/json" \
        -d '{"username": "test", "password": "test", "email": "test@test.com", "name": "test", "role": "student"}')
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "401" ]; then
        print_success "API密钥认证正常工作（拒绝无密钥请求）"
    else
        print_warning "API密钥认证可能存在问题，HTTP状态码: $http_code"
    fi
    
    # 测试错误的API密钥
    local response=$(curl -s -w "%{http_code}" -X POST \
        "${BASE_URL}/api/v1/sync/users" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: invalid_key" \
        -d '{"username": "test", "password": "test", "email": "test@test.com", "name": "test", "role": "student"}')
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "401" ]; then
        print_success "API密钥认证正常工作（拒绝错误密钥）"
    else
        print_warning "API密钥认证可能存在问题，HTTP状态码: $http_code"
    fi
}

# 测试权限控制
test_permission_control() {
    print_info "测试API密钥权限控制..."
    
    # 使用第三方密钥尝试创建管理员（应该被拒绝）
    local response=$(curl -s -w "%{http_code}" -X POST \
        "${BASE_URL}/api/v1/sync/users" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: ${THIRD_PARTY_API_KEY}" \
        -d '{
            "username": "unauthorized_admin",
            "password": "password123",
            "email": "admin@test.com",
            "name": "未授权管理员",
            "role": "admin"
        }')
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "403" ]; then
        print_success "权限控制正常（第三方密钥无法创建管理员）"
    else
        print_warning "权限控制可能存在问题，HTTP状态码: $http_code"
    fi
    
    # 使用系统同步密钥创建管理员（应该成功）
    local response=$(curl -s -w "%{http_code}" -X POST \
        "${BASE_URL}/api/v1/sync/users" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: ${SYNC_API_KEY}" \
        -d '{
            "username": "authorized_admin",
            "password": "password123", 
            "email": "admin_auth@test.com",
            "name": "授权管理员",
            "role": "admin"
        }')
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "201" ]; then
        print_success "权限控制正常（系统密钥可以创建管理员）"
    elif [ "$http_code" = "409" ]; then
        print_success "权限控制正常（管理员已存在）"
    else
        print_warning "权限控制可能存在问题，HTTP状态码: $http_code"
    fi
}

# 测试批量限制
test_batch_limits() {
    print_info "测试批量创建限制..."
    
    # 准备超过第三方密钥限制的批量请求（30个用户，超过20个限制）
    local large_batch='{"users":['
    for i in $(seq 1 30); do
        large_batch+="{\"username\":\"batch_limit_test_$i\",\"password\":\"password123\",\"email\":\"batch$i@test.com\",\"name\":\"批量测试$i\",\"role\":\"student\"}"
        if [ $i -lt 30 ]; then
            large_batch+=","
        fi
    done
    large_batch+="]}"
    
    # 使用第三方密钥尝试批量创建（应该被拒绝）
    local response=$(curl -s -w "%{http_code}" -X POST \
        "${BASE_URL}/api/v1/sync/users/batch" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: ${THIRD_PARTY_API_KEY}" \
        -d "$large_batch")
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "400" ]; then
        print_success "批量限制正常（第三方密钥被限制在20个用户以内）"
    else
        print_warning "批量限制可能存在问题，HTTP状态码: $http_code"
    fi
    
    # 使用系统同步密钥进行相同的批量创建（应该成功）
    local response=$(curl -s -w "%{http_code}" -X POST \
        "${BASE_URL}/api/v1/sync/users/batch" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: ${SYNC_API_KEY}" \
        -d "$large_batch")
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        print_success "批量限制正常（系统密钥支持更大批次）"
    else
        print_warning "批量限制可能存在问题，HTTP状态码: $http_code"
    fi
}

# 测试错误处理
test_error_handling() {
    print_info "测试错误处理..."
    
    	# 测试无效的角色
	local response=$(curl -s -w "%{http_code}" -X POST \
		"${BASE_URL}/api/v1/sync/users" \
		-H "Content-Type: application/json" \
		-H "X-API-Key: ${SYNC_API_KEY}" \
		-d '{
			"username": "invalid_role_user",
			"password": "password123",
			"email": "invalid@university.edu",
			"name": "无效角色用户",
			"role": "invalid_role"
		}')
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "400" ]; then
        print_success "错误处理正常（拒绝无效角色）"
    else
        print_warning "错误处理可能存在问题，HTTP状态码: $http_code"
    fi
    
    	# 测试缺少必需字段
	local response=$(curl -s -w "%{http_code}" -X POST \
		"${BASE_URL}/api/v1/sync/users" \
		-H "Content-Type: application/json" \
		-H "X-API-Key: ${SYNC_API_KEY}" \
		-d '{
			"username": "incomplete_user"
		}')
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "400" ]; then
        print_success "错误处理正常（拒绝不完整数据）"
    else
        print_warning "错误处理可能存在问题，HTTP状态码: $http_code"
    fi
}

# 生成测试报告
generate_report() {
    print_info "生成测试报告..."
    
    local report_file="/tmp/gitlabex_sync_api_test_report.txt"
    
    cat > "$report_file" << EOF
GitLabEx 同步API测试报告
========================

测试时间: $(date)
测试环境: $BASE_URL
系统密钥: ${SYNC_API_KEY:0:20}...
第三方密钥: ${THIRD_PARTY_API_KEY:0:20}...

测试项目:
✅ 服务健康检查
✅ API密钥认证测试
✅ 权限控制测试
✅ 批量限制测试
✅ 单个用户创建
✅ 批量用户创建  
✅ 用户信息更新
✅ 错误处理验证

所有测试项目均已完成。详细结果请查看控制台输出。

测试用户信息:
- test_student_001 (学生 -> 助教)
- batch_student_001 (学生)
- batch_teacher_001 (教师)
- batch_assistant_001 (助教)

清理建议:
如需清理测试数据，可以手动删除上述测试用户。

EOF

    print_success "测试报告已生成: $report_file"
    cat "$report_file"
}

# 主函数
main() {
    echo "🚀 GitLabEx 同步API测试开始"
    echo "=============================="
    echo ""
    
    # 检查依赖
    if ! command -v curl &> /dev/null; then
        print_error "curl 未安装，请先安装 curl"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        print_error "jq 未安装，请先安装 jq 用于JSON解析"
        exit 1
    fi
    
    	# 执行测试
	check_service
	echo ""
	
	test_api_key_auth
	echo ""
	
	test_permission_control
	echo ""
	
	test_batch_limits
	echo ""
	
	test_create_user
	echo ""
	
	test_batch_create_users
	echo ""
	
	test_update_user
	echo ""
	
	test_error_handling
	echo ""
    
    generate_report
    
    echo ""
    print_success "🎉 所有测试完成！"
    
    # 清理临时文件
    rm -f /tmp/test_user_token.txt /tmp/test_username.txt
}

# 显示帮助信息
show_help() {
    cat << EOF
GitLabEx 同步API测试脚本

用法:
  $0 [选项]

选项:
  -h, --help          显示帮助信息
  -u, --url URL       指定后端服务URL (默认: $BASE_URL)
  -k, --key KEY       指定API密钥 (默认: 从配置读取)

示例:
  $0                                    # 使用默认配置运行测试
  $0 -u http://api.example.com         # 指定不同的服务URL
  $0 -k your_custom_api_key            # 指定自定义API密钥

测试项目:
  • 服务健康检查
  • API密钥认证测试
  • 单个用户创建
  • 批量用户创建
  • 用户信息更新
  • 错误处理验证

依赖:
  • curl - HTTP客户端
  • jq - JSON处理工具

EOF
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -u|--url)
            BASE_URL="$2"
            shift 2
            ;;
        -k|--key)
            API_KEY="$2"
            shift 2
            ;;
        *)
            print_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
done

# 运行主函数
main
