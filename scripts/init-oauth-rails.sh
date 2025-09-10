#!/bin/bash

# GitLab OAuth 应用自动创建脚本
# 使用 GitLab Rails Console 创建 OAuth 应用

set -e

echo "📝 开始初始化 GitLab OAuth 应用..."

# 等待 GitLab 完全启动
echo "⏳ 等待 GitLab 服务完全启动..."
sleep 30

# 源配置文件
source /config/oauth.env

# 检查配置变量
if [ -z "$GITLAB_URL" ] || [ -z "$REDIRECT_URI" ] || [ -z "$APPLICATION_NAME" ]; then
    echo "❌ 配置文件缺少必要参数"
    exit 1
fi

echo "🔧 配置信息:"
echo "  GitLab URL: $GITLAB_URL"
echo "  Redirect URI: $REDIRECT_URI"
echo "  Application Name: $APPLICATION_NAME"
echo "  Scopes: $SCOPES"

# 创建 Rails Console 脚本
cat > /tmp/create_oauth_app.rb << EOF
# 创建 OAuth 应用的 Rails 脚本
app_name = '$APPLICATION_NAME'
redirect_uri = '$REDIRECT_URI'
scopes = '$SCOPES'

# 检查是否已存在同名应用
existing_app = Doorkeeper::Application.find_by(name: app_name)

if existing_app
  puts "✅ OAuth 应用 '#{app_name}' 已存在"
  puts "Application ID: #{existing_app.uid}"
  puts "Application Secret: #{existing_app.secret}"
  puts "Redirect URI: #{existing_app.redirect_uri}"
else
  # 创建新的 OAuth 应用
  app = Doorkeeper::Application.create!(
    name: app_name,
    redirect_uri: redirect_uri,
    scopes: scopes,
    confidential: true
  )
  
  puts "🎉 OAuth 应用创建成功!"
  puts "Application ID: #{app.uid}"
  puts "Application Secret: #{app.secret}"
  puts "Redirect URI: #{app.redirect_uri}"
  puts "Scopes: #{app.scopes}"
  
  # 保存到共享目录
  File.open('/shared/oauth_credentials.env', 'w') do |file|
    file.puts "GITLAB_CLIENT_ID=#{app.uid}"
    file.puts "GITLAB_CLIENT_SECRET=#{app.secret}"
    file.puts "GITLAB_REDIRECT_URI=#{app.redirect_uri}"
    file.puts "GITLAB_URL=$GITLAB_URL"
  end
  
  puts "💾 OAuth 凭据已保存到 /shared/oauth_credentials.env"
end
EOF

# 执行 Rails Console 命令
echo "🚀 正在创建 OAuth 应用..."
docker exec -i $GITLAB_CONTAINER gitlab-rails console < /tmp/create_oauth_app.rb

# 检查结果
if [ -f "/shared/oauth_credentials.env" ]; then
    echo "✅ OAuth 应用初始化完成!"
    echo "📋 生成的凭据:"
    cat /shared/oauth_credentials.env
else
    echo "⚠️  OAuth 应用可能已存在，请检查 GitLab 管理面板"
fi

echo "🔗 GitLab 访问地址: $GITLAB_URL"
echo "👤 默认管理员账号: root"
echo "🔑 默认密码: $GITLAB_ROOT_PASSWORD"

echo "✨ 初始化完成!"
