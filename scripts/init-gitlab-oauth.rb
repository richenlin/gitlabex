#!/usr/bin/env ruby

# GitLab OAuth应用自动创建脚本 - 修复版本
# 使用GitLab Rails Console创建OAuth应用

begin
  puts "开始GitLab OAuth应用初始化（修复版本）..."

  # 验证OAuth系统是否可用
  unless defined?(Doorkeeper)
    puts "错误：Doorkeeper模块未加载，OAuth系统不可用"
    exit 1
  end

  unless ActiveRecord::Base.connection.table_exists?('oauth_applications')
    puts "错误：oauth_applications表不存在，OAuth系统未初始化"
    exit 1
  end

  puts "OAuth系统验证通过"

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
      # 移除引号
      value = value.gsub(/^['"]|['"]$/, '') if value
      config[key] = value
    end
  end

  puts "配置解析完成: #{config.keys.join(', ')}"

  # 应用配置
  app_name = config['GITLAB_OAUTH_APP_NAME'] || "GitLabEx Education Platform"
  redirect_uri = config['GITLAB_OAUTH_REDIRECT_URI'] || "http://127.0.0.1:8000/api/auth/gitlab/callback"
  scopes = config['GITLAB_OAUTH_SCOPES'] || "api read_user email"
  force_recreate = config['FORCE_RECREATE_OAUTH_APP'] == 'true'

  puts "应用配置:"
  puts "  名称: #{app_name}"
  puts "  回调地址: #{redirect_uri}"
  puts "  权限范围: #{scopes}"
  puts "  强制重建: #{force_recreate}"

  # 初始化变量
  client_id = nil
  client_secret = nil
  
  # 检查是否已存在OAuth应用
  existing_app = Doorkeeper::Application.find_by(name: app_name)
  
  if existing_app && force_recreate
    puts "发现已存在的OAuth应用，删除并重新创建..."
    existing_app.destroy
    existing_app = nil
  elsif existing_app
    puts "发现已存在的OAuth应用，更新配置..."
    existing_app.redirect_uri = redirect_uri
    existing_app.scopes = scopes
    
    if existing_app.save
      client_id = existing_app.uid
      # 对于已存在的应用，我们需要生成新的明文secret
      require 'securerandom'
      plain_secret = SecureRandom.hex(32)
      existing_app.secret = plain_secret
      
      if existing_app.save!
        client_secret = plain_secret
        puts "✅ 成功更新已存在的OAuth应用!"
        puts "Client ID: #{client_id}"
        puts "Client Secret (明文): #{client_secret[0..10]}..."
      else
        puts "错误：无法更新应用的secret"
        exit 1
      end
    else
      puts "错误：无法更新已存在的应用"
      puts existing_app.errors.full_messages
      exit 1
    end
  else
    puts "创建新的OAuth应用..."
    
    # 生成明文secret
    require 'securerandom'
    plain_secret = SecureRandom.hex(32)
    
    # 创建应用
    app = Doorkeeper::Application.new(
      name: app_name,
      redirect_uri: redirect_uri,
      scopes: scopes,
      confidential: true
    )
    
    # 先保存应用（会生成加密的secret）
    if app.save
      puts "应用已创建，设置明文secret..."
      
      # 直接设置明文secret
      app.secret = plain_secret
      
      # 保存应用
      if app.save!
        client_id = app.uid
        client_secret = plain_secret
        puts "✅ 成功创建带明文secret的OAuth应用!"
        puts "Client ID: #{client_id}"
        puts "Client Secret (明文): #{client_secret[0..10]}..."
      else
        puts "错误：无法设置明文secret"
        puts app.errors.full_messages
        exit 1
      end
    else
      puts "错误：无法创建OAuth应用"
      puts app.errors.full_messages
      exit 1
    end
  end

  # 再次验证应用是否存在于数据库中
  verification_app = Doorkeeper::Application.find_by(uid: client_id)
  unless verification_app
    puts "错误：OAuth应用创建后在数据库中找不到"
    exit 1
  end
  puts "✅ 数据库验证成功：应用已正确保存"

  # 创建配置文件内容
  config_content = <<~CONFIG
GITLAB_CLIENT_ID=#{client_id}
GITLAB_CLIENT_SECRET=#{client_secret}
GITLAB_REDIRECT_URI=#{config['GITLAB_OAUTH_REDIRECT_URI']}
GITLAB_EXTERNAL_URL=#{config['GITLAB_EXTERNAL_URL']}
GITLAB_INTERNAL_URL=#{config['GITLAB_INTERNAL_URL']}
GITLAB_SCOPES="#{scopes}"
CONFIG

  puts "配置内容:"
  puts config_content.gsub(client_secret, "***SECRET***")

  # 确保目录存在
  require 'fileutils'
  FileUtils.mkdir_p('/shared')
  
  # 写入配置文件到共享卷
  puts "写入配置文件到 /shared/gitlab-oauth.env..."
  File.write('/shared/gitlab-oauth.env', config_content)
  puts "配置已写入 /shared/gitlab-oauth.env"

  # 验证文件是否写入成功
  unless File.exist?('/shared/gitlab-oauth.env')
    puts "错误：配置文件写入失败"
    exit 1
  end

  content = File.read('/shared/gitlab-oauth.env')
  unless content.include?(client_id)
    puts "错误：配置文件内容验证失败"
    exit 1
  end
  puts "✅ 文件写入验证成功"

  # 最终验证：检查应用总数
  total_apps = Doorkeeper::Application.count
  puts "✅ 当前OAuth应用总数: #{total_apps}"

  puts "✅ GitLab OAuth应用初始化完成!"

rescue => e
  puts "❌ 错误：#{e.message}"
  puts "错误堆栈："
  puts e.backtrace.join("\n")
  exit 1
end 