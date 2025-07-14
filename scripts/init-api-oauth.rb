#!/usr/bin/env ruby
require 'net/http'
require 'json'
require 'uri'
require 'fileutils'
require 'logger'
require 'yaml'
require 'date'
require 'openssl'

# 设置日志记录器
$logger = Logger.new(STDOUT)
$logger.level = Logger::INFO
$logger.formatter = proc do |severity, datetime, progname, msg|
  "#{datetime.strftime('%Y-%m-%d %H:%M:%S')} [#{severity}] #{msg}\n"
end

# 配置验证类
class ConfigValidator
  REQUIRED_CONFIGS = [
    'GITLAB_INTERNAL_URL',
    'GITLAB_EXTERNAL_URL',
    'GITLAB_OAUTH_REDIRECT_URI',
    'GITLAB_OAUTH_APP_NAME'
  ]

  def self.validate(config)
    missing = REQUIRED_CONFIGS.select { |key| config[key].nil? || config[key].empty? }
    if missing.any?
      $logger.error("❌ 错误：缺少必需的配置项：#{missing.join(', ')}")
      return false
    end
    true
  end
end

# 重试机制
def with_retries(max_retries = 3, delay = 5, operation = '')
  attempts = 0
  begin
    attempts += 1
    yield
  rescue => e
    if attempts < max_retries
      $logger.warn("#{operation} - 尝试 #{attempts}/#{max_retries} 失败: #{e.message}")
      sleep delay
      retry
    else
      $logger.error("#{operation} - 最终失败: #{e.message}")
      raise
    end
  end
end

# GitLab 健康检查
class GitLabHealthCheck
  def initialize(base_url)
    @base_url = base_url.chomp('/')  # 移除末尾的斜杠
    @uri = URI(base_url)
    @http = Net::HTTP.new(@uri.host, @uri.port)
    @http.use_ssl = @uri.scheme == 'https'
    @http.verify_mode = OpenSSL::SSL::VERIFY_NONE if @http.use_ssl?
    @http.read_timeout = 30
    @http.open_timeout = 30
  end

  def check_api_version
    uri = URI("#{@base_url}/api/v4/version")
    $logger.info("检查API版本: #{uri}")
    
    # 创建 HTTP 请求
    http = Net::HTTP.new(uri.host, uri.port)
    http.use_ssl = uri.scheme == 'https'
    http.verify_mode = OpenSSL::SSL::VERIFY_NONE if http.use_ssl?
    http.read_timeout = 30
    http.open_timeout = 30
    
    # 使用 GET 请求
    request = Net::HTTP::Get.new(uri)
    
    # 尝试不使用认证
    response = http.request(request)
    
    if response.is_a?(Net::HTTPSuccess)
      version = JSON.parse(response.body)['version']
      $logger.info("✅ GitLab API版本: #{version}")
      return true
    elsif response.code == '401'
      $logger.info("API需要认证，但API端点可用")
      return true
    else
      $logger.warn("❌ API版本检查失败: #{response.code} - #{response.body}")
      return false
    end
  end

  def check_health
    uri = URI("#{@base_url}/-/health")
    $logger.info("检查健康状态: #{uri}")
    response = Net::HTTP.get_response(uri)
    
    if response.is_a?(Net::HTTPSuccess)
      health_data = JSON.parse(response.body)
      $logger.info("✅ GitLab 健康状态:")
      health_data.each do |key, value|
        status = value ? '✅' : '❌'
        $logger.info("#{status} #{key}: #{value}")
      end
      health_data.values.all?
    else
      $logger.warn("❌ 健康检查失败: #{response.code} - #{response.body}")
      false
    end
  end

  def check_readiness
    uri = URI("#{@base_url}/-/readiness")
    $logger.info("检查就绪状态: #{uri}")
    response = Net::HTTP.get_response(uri)
    
    if response.is_a?(Net::HTTPSuccess)
      ready_data = JSON.parse(response.body)
      $logger.info("✅ GitLab 就绪状态:")
      ready_data.each do |key, value|
        status = value ? '✅' : '❌'
        $logger.info("#{status} #{key}: #{value}")
      end
      ready_data.values.all?
    else
      $logger.warn("❌ 就绪检查失败: #{response.code} - #{response.body}")
      false
    end
  end

  def wait_for_ready(max_attempts = 60)  # 增加最大尝试次数
    $logger.info("等待GitLab服务就绪...")
    
    # 添加初始等待时间，等待GitLab服务启动
    initial_wait = 60  # 增加初始等待时间
    $logger.info("初始等待 #{initial_wait} 秒...")
    sleep initial_wait
    
    attempts = 0
    while attempts < max_attempts
      begin
        # 只检查API版本，不检查健康和就绪状态
        if check_api_version
          $logger.info("✅ GitLab服务已就绪")
          return true
        end
      rescue => e
        $logger.warn("检查服务状态失败: #{e.message}")
      end
      
      attempts += 1
      wait_time = [10 + attempts * 2, 30].min  # 渐进式增加等待时间，最大30秒
      $logger.warn("尝试 #{attempts}/#{max_attempts}: GitLab服务未就绪，等待 #{wait_time} 秒后重试...")
      sleep wait_time
    end
    
    $logger.error("❌ 错误：GitLab服务未在预期时间内就绪")
    false
  end
end

# OAuth应用管理类
class OAuthManager
  def initialize(base_url)
    @base_url = base_url.chomp('/')
    @uri = URI(base_url)
    @http = Net::HTTP.new(@uri.host, @uri.port)
    @http.use_ssl = @uri.scheme == 'https'
    @http.verify_mode = OpenSSL::SSL::VERIFY_NONE if @http.use_ssl?
    @http.read_timeout = 30
    @http.open_timeout = 30
    @cookies = nil
  end

  def login(username, password)
    uri = URI("#{@base_url}/api/v4/session")
    $logger.info("登录GitLab获取会话: #{uri}")
    
    request = Net::HTTP::Post.new(uri)
    request.content_type = 'application/json'
    
    login_params = {
      login: username,
      password: password
    }
    
    request.body = login_params.to_json
    
    $logger.info("尝试使用用户名: #{username} 登录...")
    
    response = with_retries(3, 5, '登录GitLab') do
      @http.request(request)
    end
    
    if response.is_a?(Net::HTTPSuccess)
      user_data = JSON.parse(response.body)
      @cookies = response['Set-Cookie']
      $logger.info("✅ 成功登录GitLab")
      $logger.info("用户ID: #{user_data['id']}")
      $logger.info("用户名: #{user_data['username']}")
      true
    else
      $logger.error("❌ 错误：无法登录GitLab")
      $logger.error("状态码: #{response.code}")
      $logger.error("响应: #{response.body}")
      false
    end
  end

  def create_access_token(username, password)
    # 先尝试登录获取会话cookie
    return nil unless login(username, password)
    
    # 尝试两个可能的API端点
    endpoints = [
      "/api/v4/personal_access_tokens",
      "/api/v4/user/personal_access_tokens"
    ]
    
    token = nil
    endpoints.each do |endpoint|
      token = try_create_token(endpoint)
      break if token
    end
    
    # 如果上述方法都失败，尝试使用另一种方式
    if token.nil?
      token = try_create_token_via_user_api(username, password)
    end
    
    token
  end
  
  def try_create_token_via_user_api(username, password)
    $logger.info("尝试通过用户API创建访问令牌...")
    
    # 尝试使用GitLab用户API创建令牌
    uri = URI("#{@base_url}/api/v4/users")
    request = Net::HTTP::Get.new(uri)
    request['Cookie'] = @cookies if @cookies
    
    response = with_retries(3, 5, '获取用户列表') do
      @http.request(request)
    end
    
    if response.is_a?(Net::HTTPSuccess)
      users = JSON.parse(response.body)
      root_user = users.find { |user| user['username'] == 'root' }
      
      if root_user
        user_id = root_user['id']
        $logger.info("找到root用户，ID: #{user_id}")
        
        # 使用用户ID创建令牌
        uri = URI("#{@base_url}/api/v4/users/#{user_id}/personal_access_tokens")
        request = Net::HTTP::Post.new(uri)
        request['Cookie'] = @cookies if @cookies
        request.content_type = 'application/json'
        
        token_params = {
          name: 'GitLabEx Init Token',
          scopes: ['api'],
          expires_at: (Date.today + 1).to_s
        }
        
        request.body = token_params.to_json
        
        response = with_retries(3, 5, '通过用户API创建令牌') do
          @http.request(request)
        end
        
        if response.is_a?(Net::HTTPSuccess)
          token = JSON.parse(response.body)['token']
          $logger.info("✅ 成功通过用户API创建访问令牌")
          return token
        else
          $logger.warn("通过用户API创建令牌失败")
          $logger.warn("状态码: #{response.code}")
          $logger.warn("响应: #{response.body}")
        end
      else
        $logger.warn("未找到root用户")
      end
    else
      $logger.warn("获取用户列表失败")
      $logger.warn("状态码: #{response.code}")
      $logger.warn("响应: #{response.body}")
    end
    
    nil
  end
  
  def try_create_token(endpoint)
    uri = URI("#{@base_url}#{endpoint}")
    $logger.info("尝试创建访问令牌: #{uri}")
    
    request = Net::HTTP::Post.new(uri)
    request['Cookie'] = @cookies if @cookies
    request.content_type = 'application/json'
    
    token_params = {
      name: 'GitLabEx Init Token',
      scopes: ['api'],
      expires_at: (Date.today + 1).to_s
    }
    
    request.body = token_params.to_json
    
    response = with_retries(3, 5, '创建访问令牌') do
      @http.request(request)
    end
    
    if response.is_a?(Net::HTTPSuccess)
      token = JSON.parse(response.body)['token']
      $logger.info("✅ 成功创建访问令牌")
      token
    else
      $logger.warn("尝试端点 #{endpoint} 失败")
      $logger.warn("状态码: #{response.code}")
      $logger.warn("响应: #{response.body}")
      nil
    end
  end

  def create_oauth_application(token, app_name, redirect_uri)
    uri = URI("#{@base_url}/api/v4/applications")
    request = Net::HTTP::Post.new(uri)
    request['PRIVATE-TOKEN'] = token
    request.content_type = 'application/json'
    
    app_params = {
      name: app_name,
      redirect_uri: redirect_uri,
      scopes: 'api read_user email',
      confidential: true
    }
    
    request.body = app_params.to_json
    
    response = with_retries(3, 5, '创建OAuth应用') do
      @http.request(request)
    end
    
    if response.is_a?(Net::HTTPSuccess)
      app_data = JSON.parse(response.body)
      $logger.info("✅ 成功创建OAuth应用")
      $logger.info("应用ID: #{app_data['application_id']}")
      $logger.info("Client ID: #{app_data['application_id']}")
      $logger.info("Client Secret: #{app_data['secret'][0..10]}...")
      app_data
    else
      $logger.error("❌ 错误：无法创建OAuth应用")
      $logger.error("状态码: #{response.code}")
      $logger.error("响应: #{response.body}")
      nil
    end
  end

  def create_admin_application(password, app_name, redirect_uri)
    uri = URI("#{@base_url}/api/v4/applications")
    request = Net::HTTP::Post.new(uri)
    request.content_type = 'application/json'
    
    # 使用管理员API创建应用，需要管理员令牌
    # 假设管理员令牌已通过某种方式获取，例如硬编码或环境变量
    # 这里为了简化，假设管理员令牌已设置为环境变量 GITLAB_ADMIN_TOKEN
    admin_token = ENV['GITLAB_ADMIN_TOKEN']
    if admin_token.nil? || admin_token.empty?
      $logger.error("❌ 错误：未设置 GITLAB_ADMIN_TOKEN 环境变量，无法使用管理员API创建应用")
      return nil
    end
    request['PRIVATE-TOKEN'] = admin_token

    app_params = {
      name: app_name,
      redirect_uri: redirect_uri,
      scopes: 'api read_user email',
      confidential: true
    }
    
    request.body = app_params.to_json
    
    response = with_retries(3, 5, '创建OAuth应用 (管理员)') do
      @http.request(request)
    end
    
    if response.is_a?(Net::HTTPSuccess)
      app_data = JSON.parse(response.body)
      $logger.info("✅ 成功创建OAuth应用 (管理员)")
      $logger.info("应用ID: #{app_data['application_id']}")
      $logger.info("Client ID: #{app_data['application_id']}")
      $logger.info("Client Secret: #{app_data['secret'][0..10]}...")
      app_data
    else
      $logger.error("❌ 错误：无法创建OAuth应用 (管理员)")
      $logger.error("状态码: #{response.code}")
      $logger.error("响应: #{response.body}")
      nil
    end
  end
end

# 配置文件管理类
class ConfigManager
  def initialize(input_file, output_dir)
    @input_file = input_file
    @output_dir = output_dir
  end

  def load_config
    unless File.exist?(@input_file)
      $logger.error("❌ 错误：找不到配置文件 #{@input_file}")
      return nil
    end

    config = {}
    File.readlines(@input_file).each do |line|
      line = line.strip
      next if line.empty? || line.start_with?('#')
      
      if line.include?('=')
        key, value = line.split('=', 2)
        value = value.gsub(/^['"]|['"]$/, '') if value
        config[key] = value
      end
    end
    config
  end

  def save_oauth_config(app_data, config)
    config_content = <<~CONFIG
GITLAB_CLIENT_ID=#{app_data['application_id']}
GITLAB_CLIENT_SECRET=#{app_data['secret']}
GITLAB_REDIRECT_URI=#{config['GITLAB_OAUTH_REDIRECT_URI']}
GITLAB_EXTERNAL_URL=#{config['GITLAB_EXTERNAL_URL']}
GITLAB_INTERNAL_URL=#{config['GITLAB_INTERNAL_URL']}
GITLAB_SCOPES="api read_user email"
CONFIG

    unless File.directory?(@output_dir)
      FileUtils.mkdir_p(@output_dir)
    end

    output_file = File.join(@output_dir, 'gitlab-oauth.env')
    
    begin
      File.open(output_file, 'w', 0644) do |f|
        f.write(config_content)
      end
      FileUtils.chmod(0644, output_file)
      
      $logger.info("✅ 配置文件已写入: #{output_file}")
      $logger.info("文件权限：")
      system("ls -l #{output_file}")
      true
    rescue => e
      $logger.error("❌ 错误：无法写入配置文件")
      $logger.error(e.message)
      false
    end
  end
end

begin
  $logger.info("开始GitLab OAuth应用初始化（API版本）...")
  
  # 获取配置文件路径
  config_file = ENV['OAUTH_CONFIG_FILE'] || '/config/oauth.env'
  shared_dir = ENV['SHARED_DIR'] || '/shared'
  
  # 初始化配置管理器
  config_manager = ConfigManager.new(config_file, shared_dir)
  config = config_manager.load_config
  
  unless config && ConfigValidator.validate(config)
    exit 1
  end
  
  # 从环境变量获取GitLab root密码
  root_password = ENV['GITLAB_ROOT_PASSWORD']
  if root_password.nil? || root_password.empty?
    $logger.error("❌ 错误：未设置GITLAB_ROOT_PASSWORD环境变量")
    exit 1
  end
  
  # 设置基本参数
  base_url = config['GITLAB_INTERNAL_URL']
  redirect_uri = config['GITLAB_OAUTH_REDIRECT_URI']
  app_name = config['GITLAB_OAUTH_APP_NAME']
  
  # 等待GitLab就绪
  health_check = GitLabHealthCheck.new(base_url)
  unless health_check.wait_for_ready
    exit 1
  end
  
  # 创建OAuth管理器
  oauth_manager = OAuthManager.new(base_url)
  
  # 创建访问令牌
  $logger.info("开始创建访问令牌...")
  token = oauth_manager.create_access_token('root', root_password)
  unless token
    $logger.error("❌ 错误：无法创建访问令牌，尝试使用备用方法...")
    # 尝试直接使用管理员API创建应用
    app = oauth_manager.create_admin_application(root_password, app_name, redirect_uri)
    unless app
      $logger.error("❌ 错误：所有尝试都失败，无法创建OAuth应用")
      exit 1
    end
  else
    # 创建OAuth应用
    app = oauth_manager.create_oauth_application(token, app_name, redirect_uri)
    unless app
      exit 1
    end
  end
  
  # 保存OAuth配置
  unless config_manager.save_oauth_config(app, config)
    exit 1
  end
  
  $logger.info("✅ GitLab OAuth应用初始化完成")
  
rescue => e
  $logger.error("❌ 错误：#{e.message}")
  $logger.error(e.backtrace.join("\n"))
  exit 1
end 