#!/usr/bin/env ruby

# GitLab OAuth应用自动创建脚本
# 使用GitLab Rails Console创建OAuth应用

begin
  puts "开始GitLab OAuth应用初始化..."

  # 应用配置
  app_name = "GitLabEx Education Platform"
  redirect_uri = "http://localhost:8000/api/auth/gitlab/callback"
  scopes = "read_user read_repository write_repository"

  puts "应用名称: #{app_name}"
  puts "回调URI: #{redirect_uri}"
  puts "权限范围: #{scopes}"

  # 检查是否已存在应用
  puts "检查是否已存在应用..."
  existing_app = Doorkeeper::Application.find_by(name: app_name)

  if existing_app
    puts "OAuth应用已存在: #{app_name}"
    client_id = existing_app.uid
    client_secret = existing_app.secret
    puts "使用现有应用: ID=#{client_id}"
  else
    puts "创建新的OAuth应用: #{app_name}"
    
    # 创建OAuth应用
    app = Doorkeeper::Application.create!(
      name: app_name,
      redirect_uri: redirect_uri,
      scopes: scopes,
      confidential: true
    )
    
    client_id = app.uid
    client_secret = app.secret
    
    puts "成功创建OAuth应用!"
    puts "新应用ID: #{client_id}"
  end

  # 创建配置文件内容
  config_content = <<~CONFIG
GITLAB_CLIENT_ID=#{client_id}
GITLAB_CLIENT_SECRET=#{client_secret}
GITLAB_REDIRECT_URI=#{redirect_uri}
CONFIG

  puts "配置内容:"
  puts config_content

  # 确保目录存在
  require 'fileutils'
  FileUtils.mkdir_p('/shared')
  
  # 写入配置文件到共享卷
  File.write('/shared/gitlab-oauth.env', config_content)
  puts "配置已写入 /shared/gitlab-oauth.env"

  # 验证文件是否写入成功
  if File.exist?('/shared/gitlab-oauth.env')
    content = File.read('/shared/gitlab-oauth.env')
    puts "文件验证成功，内容:"
    puts content
  else
    puts "警告：配置文件写入失败"
  end

  puts "GitLab OAuth应用初始化完成!"

rescue => e
  puts "错误：#{e.message}"
  puts "错误堆栈："
  puts e.backtrace.join("\n")
  exit 1
end 