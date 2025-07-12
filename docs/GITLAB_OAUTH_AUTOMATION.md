# GitLab OAuth 应用自动化配置

## 概述

本项目实现了 GitLab OAuth 应用程序的完全自动化配置，消除了手动创建 OAuth 应用的需要。

## 工作原理

### 1. 自动化组件

- **Ruby脚本** (`scripts/init-gitlab-oauth.rb`) - 使用 GitLab Rails Console 自动创建 OAuth 应用
- **Shell脚本** (`scripts/init-gitlab-oauth.sh`) - 等待 GitLab 启动并执行 OAuth 创建
- **初始化服务** (`gitlab-init`) - Docker 容器自动运行初始化脚本
- **配置共享** - 通过 Docker 卷共享 OAuth 配置给后端服务
- **后端自动加载** - 后端自动从共享卷读取 OAuth 配置

### 2. 流程说明

1. 启动 `docker-compose up -d`
2. GitLab 服务启动并初始化
3. `gitlab-init` 服务等待 GitLab 完全就绪
4. 自动执行 Ruby 脚本在 GitLab 中创建 OAuth 应用
5. 将生成的 Client ID 和 Secret 保存到共享卷
6. 后端服务自动加载 OAuth 配置
7. 系统完全配置完成，可以进行 OAuth 认证

## 自动创建的 OAuth 应用配置

- **应用名称**: GitLabEx Education Platform
- **重定向URI**: http://localhost:8000/api/auth/gitlab/callback
- **权限范围**: read_user read_repository write_repository

## 验证步骤

### 1. 检查所有服务状态
```bash
sudo docker-compose ps
```

### 2. 测试 OAuth 端点
```bash
curl -s http://localhost:8000/api/auth/gitlab
```
应该返回包含 client_id 的授权 URL。

### 3. 测试后端健康检查
```bash
curl -s http://localhost:8000/api/health
```
应该返回 JSON 格式的健康状态。

### 4. 检查 OAuth 配置文件
```bash
sudo cat /var/lib/docker/volumes/gitlabex_gitlab_oauth_config/_data/gitlab-oauth.env
```

## 配置文件位置

- **共享卷**: `gitlabex_gitlab_oauth_config`
- **配置文件**: `/shared/gitlab-oauth.env` (容器内)
- **物理路径**: `/var/lib/docker/volumes/gitlabex_gitlab_oauth_config/_data/gitlab-oauth.env`

## 环境变量

后端服务自动从配置文件加载以下环境变量：
- `GITLAB_CLIENT_ID` - OAuth 应用的客户端 ID
- `GITLAB_CLIENT_SECRET` - OAuth 应用的客户端密钥
- `GITLAB_REDIRECT_URI` - OAuth 重定向 URI

## Docker Compose 配置

### gitlab-init 服务
```yaml
gitlab-init:
  image: alpine:latest
  container_name: gitlabex-gitlab-init
  restart: "no"
  depends_on:
    - gitlab
  volumes:
    - ./scripts:/scripts:ro
    - gitlab_oauth_config:/shared
    - /var/run/docker.sock:/var/run/docker.sock
  networks:
    - gitlabex-network
  command: >
    sh -c "
      apk add --no-cache curl docker-cli &&
      sh /scripts/init-gitlab-oauth.sh
    "
```

### 后端服务配置
```yaml
backend:
  # ... 其他配置
  depends_on:
    - gitlab-init  # 确保 OAuth 配置完成后再启动
  environment:
    - GITLAB_OAUTH_CONFIG_PATH=/shared/gitlab-oauth.env
  volumes:
    - gitlab_oauth_config:/shared:ro
```

## 故障排除

### 1. 如果 OAuth 应用创建失败
```bash
# 检查 gitlab-init 服务日志
sudo docker-compose logs gitlab-init

# 手动运行 OAuth 创建脚本
sudo docker exec gitlabex-gitlab gitlab-rails runner /tmp/init-gitlab-oauth.rb
```

### 2. 如果后端无法加载配置
```bash
# 检查后端服务日志
sudo docker-compose logs backend | grep -E "(OAuth|config)"

# 验证配置文件是否存在
sudo docker exec -it gitlabex-backend cat /shared/gitlab-oauth.env
```

### 3. 重新生成 OAuth 配置
```bash
# 停止服务
sudo docker-compose down

# 清理配置卷
sudo docker volume rm gitlabex_gitlab_oauth_config

# 重新启动
sudo docker-compose up -d
```

## 优势

1. **完全自动化** - 无需手动登录 GitLab 创建应用
2. **一致性** - 每次部署都使用相同的配置
3. **可重复性** - 可以在任何环境中重复部署
4. **零干预** - 启动后自动完成所有配置
5. **容错性** - 如果应用已存在，会重用现有配置

## 安全注意事项

- OAuth Client Secret 使用 GitLab 内置的加密格式存储
- 配置文件只有容器内部可访问
- 建议在生产环境中使用更安全的密钥管理方案

## 成功部署验证

当您看到以下输出时，说明自动化配置成功：

```bash
# OAuth 端点测试
$ curl -s http://localhost:8000/api/auth/gitlab
{"url":"http://gitlab/oauth/authorize?client_id=...&redirect_uri=..."}

# 健康检查测试
$ curl -s http://localhost:8000/api/health
{"service":"gitlabex-backend","status":"ok","timestamp":...,"version":"1.0.0"}
```

## 总结

此自动化解决方案完全消除了手动创建 GitLab OAuth 应用程序的需要，使部署过程更加简单和可靠。系统启动后，所有必要的 OAuth 配置都会自动完成，您可以立即开始使用 GitLab 认证功能。 