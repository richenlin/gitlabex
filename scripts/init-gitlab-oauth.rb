#!/usr/bin/env ruby

# GitLab OAuth应用自动创建脚本
# 使用GitLab Rails Console创建OAuth应用

begin
  puts "开始GitLab OAuth应用初始化..."

  # 读取OAuth配置文件
  config_file = "/config/oauth.env"
  unless File.exist?(config_file)
    puts "错误：找不到OAuth配置文件 #{config_file}"
    exit 1
  end

  puts "加载OAuth配置文件: #{config_file}"
  
  # 解析配置文件
  config = {}
  File.readlines(config_file).each do |line|
    line = line.strip
    next if line.empty? || line.start_with?('#')
    
    if line.include?('=')
      key, value = line.split('=', 2)
      config[key] = value
    end
  end

  # 应用配置
  app_name = config['GITLAB_OAUTH_APP_NAME'] || "GitLabEx Education Platform"
  redirect_uri = config['GITLAB_OAUTH_REDIRECT_URI'] || "http://127.0.0.1:8000/api/auth/gitlab/callback"
  scopes = config['GITLAB_OAUTH_SCOPES'] || "read_user read_repository write_repository"
  force_recreate = config['FORCE_RECREATE_OAUTH_APP'] == 'true'

  # 支持多个回调地址（兼容localhost和127.0.0.1）
  redirect_uri_list = [redirect_uri]
  if redirect_uri.include?('127.0.0.1')
    redirect_uri_list << redirect_uri.gsub('127.0.0.1', 'localhost')
  elsif redirect_uri.include?('localhost')
    redirect_uri_list << redirect_uri.gsub('localhost', '127.0.0.1')
  end
  
  redirect_uri = redirect_uri_list.join("\n")

  puts "应用名称: #{app_name}"
  puts "回调URI: #{redirect_uri_list.join(', ')}"
  puts "权限范围: #{scopes}"
  puts "强制重创: #{force_recreate}"

  # 检查是否已存在应用
  puts "检查是否已存在应用..."
  existing_app = Doorkeeper::Application.find_by(name: app_name)

  if existing_app
    puts "OAuth应用已存在: #{app_name}"
    if force_recreate
      puts "配置为强制重创，删除现有应用..."
      existing_app.destroy!
      puts "现有应用已删除"
    else
      puts "使用现有应用，无需重创"
      client_id = existing_app.uid
      client_secret = existing_app.secret
      puts "现有应用ID: #{client_id}"
      
      # 直接跳到创建配置文件的部分
      config_content = <<~CONFIG
GITLAB_CLIENT_ID=#{client_id}
GITLAB_CLIENT_SECRET=#{client_secret}
GITLAB_REDIRECT_URI=#{config['GITLAB_OAUTH_REDIRECT_URI']}
GITLAB_EXTERNAL_URL=#{config['GITLAB_EXTERNAL_URL']}
GITLAB_INTERNAL_URL=#{config['GITLAB_INTERNAL_URL']}
CONFIG

      puts "配置内容:"
      puts config_content

      # 确保目录存在
      require 'fileutils'
      FileUtils.mkdir_p('/shared')
      
      # 写入配置文件到共享卷
      File.write('/shared/gitlab-oauth.env', config_content)
      puts "配置已写入 /shared/gitlab-oauth.env"

      puts "GitLab OAuth应用初始化完成!"
      exit 0
    end
  end

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

  # 创建配置文件内容
  config_content = <<~CONFIG
GITLAB_CLIENT_ID=#{client_id}
GITLAB_CLIENT_SECRET=#{client_secret}
GITLAB_REDIRECT_URI=#{config['GITLAB_OAUTH_REDIRECT_URI']}
GITLAB_EXTERNAL_URL=#{config['GITLAB_EXTERNAL_URL']}
GITLAB_INTERNAL_URL=#{config['GITLAB_INTERNAL_URL']}
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