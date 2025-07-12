#!/bin/bash

# GitLab OAuth自动化配置验证脚本

echo "=== GitLab OAuth自动化配置验证 ==="

# 1. 检查GitLab是否运行
echo "1. 检查GitLab服务状态..."
if docker exec gitlabex-gitlab curl -f -s http://localhost/-/health > /dev/null 2>&1; then
    echo "   ✅ GitLab服务正常运行"
else
    echo "   ❌ GitLab服务未运行或不健康"
    exit 1
fi

# 2. 检查OAuth配置文件是否存在
echo "2. 检查OAuth配置文件..."
if [ -f "/var/lib/docker/volumes/gitlabex_gitlab_oauth_config/_data/gitlab-oauth.env" ]; then
    echo "   ✅ OAuth配置文件存在"
    echo "   配置内容："
    cat "/var/lib/docker/volumes/gitlabex_gitlab_oauth_config/_data/gitlab-oauth.env" | sed 's/CLIENT_SECRET=.*/CLIENT_SECRET=***/'
else
    echo "   ❌ OAuth配置文件不存在"
    echo "   尝试查找配置文件位置..."
    docker volume inspect gitlabex_gitlab_oauth_config
    exit 1
fi

# 3. 检查backend是否能访问配置
echo "3. 检查backend配置加载..."
if docker exec gitlabex-backend test -f /shared/gitlab-oauth.env; then
    echo "   ✅ Backend可以访问OAuth配置文件"
    echo "   Backend中的配置："
    docker exec gitlabex-backend cat /shared/gitlab-oauth.env | sed 's/CLIENT_SECRET=.*/CLIENT_SECRET=***/'
else
    echo "   ❌ Backend无法访问OAuth配置文件"
    echo "   检查共享卷挂载..."
    docker exec gitlabex-backend ls -la /shared/
    exit 1
fi

# 4. 检查GitLab中的OAuth应用
echo "4. 检查GitLab中的OAuth应用..."
oauth_apps=$(docker exec gitlabex-gitlab gitlab-rails runner "puts Doorkeeper::Application.where(name: 'GitLabEx Education Platform').count")
if [ "$oauth_apps" -gt 0 ]; then
    echo "   ✅ GitLab中存在OAuth应用"
    # 显示应用详情
    docker exec gitlabex-gitlab gitlab-rails runner "
    app = Doorkeeper::Application.find_by(name: 'GitLabEx Education Platform')
    if app
        puts '   应用名称: ' + app.name
        puts '   客户端ID: ' + app.uid[0..10] + '...'
        puts '   回调URI: ' + app.redirect_uri
        puts '   权限范围: ' + app.scopes.to_s
    end
    "
else
    echo "   ❌ GitLab中不存在OAuth应用"
    exit 1
fi

# 5. 测试backend API端点
echo "5. 测试backend OAuth端点..."
response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8000/api/auth/gitlab/login)
if [ "$response" = "302" ] || [ "$response" = "200" ]; then
    echo "   ✅ Backend OAuth登录端点响应正常 (HTTP $response)"
else
    echo "   ❌ Backend OAuth登录端点异常 (HTTP $response)"
    echo "   尝试查看backend日志..."
    docker logs gitlabex-backend --tail 10
fi

# 6. 测试完整的OAuth URL
echo "6. 测试OAuth授权URL生成..."
auth_url=$(curl -s http://localhost:8000/api/auth/gitlab/login | grep -o 'http://[^"]*' | head -1)
if [[ "$auth_url" =~ http://localhost:8000/gitlab/oauth/authorize ]]; then
    echo "   ✅ OAuth授权URL格式正确"
    echo "   授权URL: $auth_url"
else
    echo "   ❌ OAuth授权URL格式错误或无法获取"
    echo "   获取到的URL: $auth_url"
fi

echo ""
echo "=== 验证完成 ==="
echo "如果所有检查都通过，OAuth自动化配置已正确工作！"
echo "如果有任何失败项，请检查相应的配置和日志。" 