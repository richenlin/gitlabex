#!/bin/bash

# GitLabEx 系统集成测试脚本
# 测试重构后的系统功能：权限管理、课题管理、作业管理、数据统计

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# 配置变量
BASE_URL=${BASE_URL:-"http://localhost:8000"}
TEST_USER_TOKEN=""
TEST_TEACHER_TOKEN=""
TEST_STUDENT_TOKEN=""

# 测试结果统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 测试数据
TEST_PROJECT_ID=""
TEST_ASSIGNMENT_ID=""
TEST_SUBMISSION_ID=""

# 函数：打印带颜色的消息
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}[$(date '+%H:%M:%S')] ${message}${NC}"
}

# 函数：测试结果
test_result() {
    local test_name=$1
    local success=$2
    local details=$3
    
    ((TOTAL_TESTS++))
    
    if [ "$success" = true ]; then
        ((PASSED_TESTS++))
        print_message $GREEN "✓ PASS: $test_name"
    else
        ((FAILED_TESTS++))
        print_message $RED "✗ FAIL: $test_name"
        if [ -n "$details" ]; then
            print_message $RED "  详情: $details"
        fi
    fi
}

# 函数：HTTP请求
http_request() {
    local method=$1
    local url=$2
    local data=$3
    local token=$4
    local expected_status=${5:-200}
    
    local auth_header=""
    if [ -n "$token" ]; then
        auth_header="Authorization: Bearer $token"
    fi
    
    local response
    local status_code
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" -H "$auth_header" "$BASE_URL$url")
    elif [ "$method" = "POST" ]; then
        response=$(curl -s -w "\n%{http_code}" -X POST -H "$auth_header" -H "Content-Type: application/json" -d "$data" "$BASE_URL$url")
    elif [ "$method" = "PUT" ]; then
        response=$(curl -s -w "\n%{http_code}" -X PUT -H "$auth_header" -H "Content-Type: application/json" -d "$data" "$BASE_URL$url")
    elif [ "$method" = "DELETE" ]; then
        response=$(curl -s -w "\n%{http_code}" -X DELETE -H "$auth_header" "$BASE_URL$url")
    fi
    
    status_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" = "$expected_status" ]; then
        echo "$response_body"
        return 0
    else
        echo "HTTP $status_code: $response_body"
        return 1
    fi
}

# 函数：检查服务器状态
check_server_status() {
    print_message $BLUE "检查服务器状态..."
    
    local response
    if response=$(http_request "GET" "/api/health" "" "" 200); then
        test_result "服务器健康检查" true
        return 0
    else
        test_result "服务器健康检查" false "$response"
        return 1
    fi
}

# 函数：测试用户认证（模拟）
test_authentication() {
    print_message $BLUE "测试用户认证..."
    
    # 注意：这里使用模拟的JWT Token，实际环境需要通过GitLab OAuth获取
    TEST_TEACHER_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJyb2xlIjoidGVhY2hlciIsImV4cCI6OTk5OTk5OTk5OX0.mock_teacher_token"
    TEST_STUDENT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyLCJyb2xlIjoic3R1ZGVudCIsImV4cCI6OTk5OTk5OTk5OX0.mock_student_token"
    
    # 测试获取当前用户信息
    local response
    if response=$(http_request "GET" "/api/users/current" "" "$TEST_TEACHER_TOKEN" 200); then
        test_result "获取教师用户信息" true
    else
        test_result "获取教师用户信息" false "$response"
    fi
    
    if response=$(http_request "GET" "/api/users/current" "" "$TEST_STUDENT_TOKEN" 200); then
        test_result "获取学生用户信息" true
    else
        test_result "获取学生用户信息" false "$response"
    fi
}

# 函数：测试权限管理
test_permission_management() {
    print_message $BLUE "测试权限管理..."
    
    # 测试获取角色列表
    local response
    if response=$(http_request "GET" "/api/permissions/roles" "" "$TEST_TEACHER_TOKEN" 200); then
        test_result "获取角色列表" true
    else
        test_result "获取角色列表" false "$response"
    fi
    
    # 测试权限检查
    local check_data='{"user_id": 1, "resource_type": "project", "resource_id": 1, "action": "read"}'
    if response=$(http_request "POST" "/api/permissions/check" "$check_data" "$TEST_TEACHER_TOKEN" 200); then
        test_result "权限检查API" true
    else
        test_result "权限检查API" false "$response"
    fi
    
    # 测试获取用户权限信息
    if response=$(http_request "GET" "/api/permissions/user/1" "" "$TEST_TEACHER_TOKEN" 200); then
        test_result "获取用户权限信息" true
    else
        test_result "获取用户权限信息" false "$response"
    fi
}

# 函数：测试课题管理
test_project_management() {
    print_message $BLUE "测试课题管理..."
    
    # 测试创建课题（教师）
    local project_data='{
        "name": "测试课题项目",
        "description": "这是一个系统集成测试项目",
        "start_date": "2024-03-15T00:00:00Z",
        "end_date": "2024-06-30T23:59:59Z",
        "max_members": 30
    }'
    
    local response
    if response=$(http_request "POST" "/api/projects" "$project_data" "$TEST_TEACHER_TOKEN" 201); then
        TEST_PROJECT_ID=$(echo "$response" | grep -o '"id":[0-9]*' | cut -d':' -f2)
        test_result "创建课题（教师）" true
    else
        test_result "创建课题（教师）" false "$response"
    fi
    
    # 测试获取课题列表
    if response=$(http_request "GET" "/api/projects" "" "$TEST_TEACHER_TOKEN" 200); then
        test_result "获取课题列表" true
    else
        test_result "获取课题列表" false "$response"
    fi
    
    # 测试获取课题详情
    if [ -n "$TEST_PROJECT_ID" ]; then
        if response=$(http_request "GET" "/api/projects/$TEST_PROJECT_ID" "" "$TEST_TEACHER_TOKEN" 200); then
            local project_code=$(echo "$response" | grep -o '"project_code":"[^"]*"' | cut -d'"' -f4)
            test_result "获取课题详情" true
            
            # 测试学生加入课题
            if [ -n "$project_code" ]; then
                local join_data="{\"code\": \"$project_code\"}"
                if response=$(http_request "POST" "/api/projects/join" "$join_data" "$TEST_STUDENT_TOKEN" 200); then
                    test_result "学生加入课题" true
                else
                    test_result "学生加入课题" false "$response"
                fi
            fi
        else
            test_result "获取课题详情" false "$response"
        fi
    fi
    
    # 测试获取课题成员
    if [ -n "$TEST_PROJECT_ID" ]; then
        if response=$(http_request "GET" "/api/projects/$TEST_PROJECT_ID/members" "" "$TEST_TEACHER_TOKEN" 200); then
            test_result "获取课题成员列表" true
        else
            test_result "获取课题成员列表" false "$response"
        fi
    fi
    
    # 测试获取课题统计
    if [ -n "$TEST_PROJECT_ID" ]; then
        if response=$(http_request "GET" "/api/projects/$TEST_PROJECT_ID/stats" "" "$TEST_TEACHER_TOKEN" 200); then
            test_result "获取课题统计信息" true
        else
            test_result "获取课题统计信息" false "$response"
        fi
    fi
}

# 函数：测试作业管理
test_assignment_management() {
    print_message $BLUE "测试作业管理..."
    
    if [ -z "$TEST_PROJECT_ID" ]; then
        test_result "作业管理测试" false "需要先创建课题"
        return
    fi
    
    # 测试创建作业（教师）
    local assignment_data="{
        \"title\": \"系统集成测试作业\",
        \"description\": \"完成指定的开发任务\",
        \"project_id\": $TEST_PROJECT_ID,
        \"due_date\": \"2024-04-30T23:59:59Z\",
        \"max_score\": 100,
        \"grading_criteria\": {
            \"functionality\": 40,
            \"code_quality\": 30,
            \"documentation\": 20,
            \"ui_design\": 10
        }
    }"
    
    local response
    if response=$(http_request "POST" "/api/assignments" "$assignment_data" "$TEST_TEACHER_TOKEN" 201); then
        TEST_ASSIGNMENT_ID=$(echo "$response" | grep -o '"id":[0-9]*' | cut -d':' -f2)
        test_result "创建作业（教师）" true
    else
        test_result "创建作业（教师）" false "$response"
    fi
    
    # 测试获取作业列表
    if response=$(http_request "GET" "/api/assignments" "" "$TEST_TEACHER_TOKEN" 200); then
        test_result "获取作业列表（教师）" true
    else
        test_result "获取作业列表（教师）" false "$response"
    fi
    
    # 测试学生获取作业列表
    if response=$(http_request "GET" "/api/assignments" "" "$TEST_STUDENT_TOKEN" 200); then
        test_result "获取作业列表（学生）" true
    else
        test_result "获取作业列表（学生）" false "$response"
    fi
    
    # 测试学生提交作业
    if [ -n "$TEST_ASSIGNMENT_ID" ]; then
        local submission_data="{
            \"submission_content\": \"这是测试提交的作业内容\",
            \"commit_hash\": \"abc123def456\",
            \"files\": {
                \"main.js\": \"console.log('Hello World');\",
                \"README.md\": \"# 测试项目\\n\\n这是一个测试项目。\"
            },
            \"branch_name\": \"student-test-branch\"
        }"
        
        if response=$(http_request "POST" "/api/assignments/$TEST_ASSIGNMENT_ID/submit" "$submission_data" "$TEST_STUDENT_TOKEN" 201); then
            TEST_SUBMISSION_ID=$(echo "$response" | grep -o '"submission_id":[0-9]*' | cut -d':' -f2)
            test_result "学生提交作业" true
        else
            test_result "学生提交作业" false "$response"
        fi
    fi
    
    # 测试获取作业提交列表（教师）
    if [ -n "$TEST_ASSIGNMENT_ID" ]; then
        if response=$(http_request "GET" "/api/assignments/$TEST_ASSIGNMENT_ID/submissions" "" "$TEST_TEACHER_TOKEN" 200); then
            test_result "获取作业提交列表（教师）" true
        else
            test_result "获取作业提交列表（教师）" false "$response"
        fi
    fi
    
    # 测试评审作业（教师）
    if [ -n "$TEST_SUBMISSION_ID" ]; then
        local review_data="{
            \"score\": 85,
            \"review_report\": {
                \"code_quality_score\": 80,
                \"code_quality_comment\": \"代码结构清晰，但需要增加注释\",
                \"functionality_score\": 90,
                \"functionality_comment\": \"功能实现完整\",
                \"documentation_score\": 75,
                \"documentation_comment\": \"文档详细度有待提高\",
                \"ui_design_score\": 85,
                \"ui_design_comment\": \"界面设计美观\"
            },
            \"general_comment\": \"整体完成质量良好，继续努力\",
            \"suggestions\": [
                \"增加代码注释\",
                \"完善API文档\"
            ]
        }"
        
        if response=$(http_request "PUT" "/api/assignments/submissions/$TEST_SUBMISSION_ID/review" "$review_data" "$TEST_TEACHER_TOKEN" 200); then
            test_result "评审作业（教师）" true
        else
            test_result "评审作业（教师）" false "$response"
        fi
    fi
    
    # 测试学生获取个人提交记录
    if response=$(http_request "GET" "/api/assignments/my-submissions" "" "$TEST_STUDENT_TOKEN" 200); then
        test_result "获取个人提交记录（学生）" true
    else
        test_result "获取个人提交记录（学生）" false "$response"
    fi
    
    # 测试获取作业统计
    if [ -n "$TEST_ASSIGNMENT_ID" ]; then
        if response=$(http_request "GET" "/api/assignments/$TEST_ASSIGNMENT_ID/stats" "" "$TEST_TEACHER_TOKEN" 200); then
            test_result "获取作业统计信息" true
        else
            test_result "获取作业统计信息" false "$response"
        fi
    fi
}

# 函数：测试数据统计
test_analytics() {
    print_message $BLUE "测试数据统计..."
    
    # 测试教师统计概览
    local response
    if response=$(http_request "GET" "/api/analytics/teacher/overview" "" "$TEST_TEACHER_TOKEN" 200); then
        test_result "教师统计概览" true
    else
        test_result "教师统计概览" false "$response"
    fi
    
    # 测试教师课题统计
    if response=$(http_request "GET" "/api/analytics/teacher/projects" "" "$TEST_TEACHER_TOKEN" 200); then
        test_result "教师课题统计" true
    else
        test_result "教师课题统计" false "$response"
    fi
    
    # 测试教师作业统计
    if response=$(http_request "GET" "/api/analytics/teacher/assignments" "" "$TEST_TEACHER_TOKEN" 200); then
        test_result "教师作业统计" true
    else
        test_result "教师作业统计" false "$response"
    fi
    
    # 测试学生统计概览
    if response=$(http_request "GET" "/api/analytics/student/overview" "" "$TEST_STUDENT_TOKEN" 200); then
        test_result "学生统计概览" true
    else
        test_result "学生统计概览" false "$response"
    fi
    
    # 测试学生作业统计
    if response=$(http_request "GET" "/api/analytics/student/assignments" "" "$TEST_STUDENT_TOKEN" 200); then
        test_result "学生作业统计" true
    else
        test_result "学生作业统计" false "$response"
    fi
    
    # 测试学生学习进度
    if response=$(http_request "GET" "/api/analytics/student/progress" "" "$TEST_STUDENT_TOKEN" 200); then
        test_result "学生学习进度" true
    else
        test_result "学生学习进度" false "$response"
    fi
    
    # 测试管理员统计（如果有管理员token）
    if response=$(http_request "GET" "/api/analytics/overview" "" "$TEST_TEACHER_TOKEN" 200); then
        test_result "管理员统计概览" true
    else
        test_result "管理员统计概览" false "$response"
    fi
}

# 函数：测试第三方API
test_third_party_api() {
    print_message $BLUE "测试第三方API..."
    
    # 测试生成API Key
    local response
    if response=$(http_request "POST" "/api/third-party/auth/api-key" "" "$TEST_TEACHER_TOKEN" 200); then
        local api_key=$(echo "$response" | grep -o '"api_key":"[^"]*"' | cut -d'"' -f4)
        test_result "生成第三方API Key" true
        
        # 测试验证API Key
        if [ -n "$api_key" ]; then
            if response=$(http_request "GET" "/api/third-party/auth/validate" "" "$api_key" 200); then
                test_result "验证第三方API Key" true
            else
                test_result "验证第三方API Key" false "$response"
            fi
        fi
    else
        test_result "生成第三方API Key" false "$response"
    fi
    
    # 测试第三方API项目列表
    if response=$(http_request "GET" "/api/third-party/projects" "" "$TEST_TEACHER_TOKEN" 200); then
        test_result "第三方API获取项目列表" true
    else
        test_result "第三方API获取项目列表" false "$response"
    fi
}

# 函数：测试系统性能
test_system_performance() {
    print_message $BLUE "测试系统性能..."
    
    # 并发请求测试
    local start_time=$(date +%s)
    
    # 模拟5个并发请求
    for i in {1..5}; do
        (http_request "GET" "/api/projects" "" "$TEST_TEACHER_TOKEN" 200 > /dev/null 2>&1) &
    done
    wait
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    if [ $duration -lt 5 ]; then
        test_result "并发请求性能测试" true "5个并发请求耗时${duration}秒"
    else
        test_result "并发请求性能测试" false "5个并发请求耗时${duration}秒，超过预期"
    fi
}

# 函数：清理测试数据
cleanup_test_data() {
    print_message $BLUE "清理测试数据..."
    
    # 删除测试作业（如果存在）
    if [ -n "$TEST_ASSIGNMENT_ID" ]; then
        if http_request "DELETE" "/api/assignments/$TEST_ASSIGNMENT_ID" "" "$TEST_TEACHER_TOKEN" 200 > /dev/null 2>&1; then
            test_result "删除测试作业" true
        else
            test_result "删除测试作业" false
        fi
    fi
    
    # 删除测试课题（如果存在）
    if [ -n "$TEST_PROJECT_ID" ]; then
        if http_request "DELETE" "/api/projects/$TEST_PROJECT_ID" "" "$TEST_TEACHER_TOKEN" 200 > /dev/null 2>&1; then
            test_result "删除测试课题" true
        else
            test_result "删除测试课题" false
        fi
    fi
}

# 函数：显示测试总结
show_test_summary() {
    echo ""
    print_message $PURPLE "=== 测试总结 ==="
    echo "总测试数: $TOTAL_TESTS"
    echo "通过测试: $PASSED_TESTS"
    echo "失败测试: $FAILED_TESTS"
    
    local success_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    echo "通过率: ${success_rate}%"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        print_message $GREEN "🎉 所有测试通过！系统集成测试成功！"
        return 0
    else
        print_message $RED "❌ 存在失败的测试，请检查系统功能"
        return 1
    fi
}

# 函数：显示帮助信息
show_help() {
    echo "GitLabEx 系统集成测试脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help              显示此帮助信息"
    echo "  -u, --url URL           指定API基础URL (默认: http://localhost:8000)"
    echo "  -s, --skip-cleanup      跳过测试数据清理"
    echo "  -t, --test TYPE         运行特定类型的测试"
    echo "                          可选值: auth, permission, project, assignment, analytics, api, performance"
    echo "  -v, --verbose           显示详细输出"
    echo ""
    echo "环境变量:"
    echo "  BASE_URL                API基础URL"
    echo ""
    echo "示例:"
    echo "  $0                      运行完整的集成测试"
    echo "  $0 -u http://test.example.com  使用指定的API URL"
    echo "  $0 -t project           只运行课题管理测试"
    echo "  $0 -s                   运行测试但不清理数据"
    echo ""
}

# 主函数
main() {
    local skip_cleanup=false
    local test_type="all"
    local verbose=false
    
    # 参数解析
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
            -s|--skip-cleanup)
                skip_cleanup=true
                shift
                ;;
            -t|--test)
                test_type="$2"
                shift 2
                ;;
            -v|--verbose)
                verbose=true
                shift
                ;;
            *)
                print_message $RED "未知选项: $1"
                print_message $YELLOW "使用 -h 或 --help 查看帮助"
                exit 1
                ;;
        esac
    done
    
    print_message $BLUE "=== GitLabEx 系统集成测试开始 ==="
    print_message $BLUE "API基础URL: $BASE_URL"
    print_message $BLUE "测试类型: $test_type"
    echo ""
    
    # 检查服务器状态
    if ! check_server_status; then
        print_message $RED "服务器无法访问，测试终止"
        exit 1
    fi
    
    # 运行测试
    case $test_type in
        all)
            test_authentication
            test_permission_management
            test_project_management
            test_assignment_management
            test_analytics
            test_third_party_api
            test_system_performance
            ;;
        auth)
            test_authentication
            ;;
        permission)
            test_authentication
            test_permission_management
            ;;
        project)
            test_authentication
            test_project_management
            ;;
        assignment)
            test_authentication
            test_project_management
            test_assignment_management
            ;;
        analytics)
            test_authentication
            test_analytics
            ;;
        api)
            test_authentication
            test_third_party_api
            ;;
        performance)
            test_system_performance
            ;;
        *)
            print_message $RED "未知的测试类型: $test_type"
            exit 1
            ;;
    esac
    
    # 清理测试数据
    if [ "$skip_cleanup" != true ] && [ "$test_type" = "all" -o "$test_type" = "project" -o "$test_type" = "assignment" ]; then
        cleanup_test_data
    fi
    
    # 显示测试总结
    if show_test_summary; then
        exit 0
    else
        exit 1
    fi
}

# 脚本入口
main "$@" 