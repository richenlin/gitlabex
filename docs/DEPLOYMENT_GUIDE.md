# GitLabEx 容器化部署指南

## 部署流程修复说明

### 问题诊断与解决

#### 原始问题
1. **Docker权限问题** - 用户没有访问Docker socket的权限
2. **GitLab初始化脚本失败** - OAuth配置文件生成失败
3. **容器启动顺序混乱** - 缺乏精确的依赖关系控制

#### 解决方案
1. **Docker权限修复**
   ```bash
   sudo usermod -aG docker $USER
   newgrp docker
   ```

2. **GitLab OAuth初始化脚本修复**
   - 创建了 `scripts/init-gitlab-oauth-fixed.rb` 修复版本
   - 解决了明文secret生成问题
   - 改进了错误处理和验证逻辑

3. **容器编排优化**
   - 添加了健康检查机制
   - 优化了依赖关系配置
   - 移除了过时的 `version` 配置

### 修复后的部署流程

#### 阶段一：init容器配置GitLab授权
```yaml
gitlab-init:
  image: alpine:latest
  restart: "no"  # 只运行一次
  depends_on:
    gitlab:
      condition: service_healthy
  command: |
    等待GitLab完全启动 →
    生成GitLab OAuth配置 →
    重启Backend服务 →
    重启Nginx服务
```

#### 阶段二：Backend服务动态配置加载
```yaml
backend:
  restart: unless-stopped
  depends_on:
    postgres: service_started
    redis: service_started
    gitlab: service_healthy
    onlyoffice: service_started
  environment:
    GITLAB_OAUTH_CONFIG_PATH: /shared/gitlab-oauth.env
```

#### 阶段三：Nginx服务路由配置
```yaml
nginx:
  depends_on:
    frontend: service_started
    backend: service_healthy
    gitlab: service_started
    onlyoffice: service_started
```

## 部署方式

### 方式一：自动化部署（推荐）
```bash
# 运行自动化部署测试脚本
./scripts/test-deployment.sh
```

### 方式二：手动分步部署
```bash
# 1. 启动基础服务
docker-compose up -d postgres redis

# 2. 启动GitLab（等待健康检查通过）
docker-compose up -d gitlab

# 3. 启动其他服务
docker-compose up -d onlyoffice frontend backend

# 4. 运行OAuth初始化
docker-compose up gitlab-init

# 5. 启动Nginx
docker-compose up -d nginx
```

### 方式三：一键部署（适用于已配置环境）
```bash
docker-compose up -d
```

## 验证部署结果

### 检查服务状态
```bash
docker-compose ps
```

### 验证OAuth配置
```bash
# 检查配置文件是否生成
sudo ls -la /var/lib/docker/volumes/gitlabex_gitlab_oauth_config/_data/

# 查看配置内容（敏感信息已隐藏）
sudo cat /var/lib/docker/volumes/gitlabex_gitlab_oauth_config/_data/gitlab-oauth.env
```

### 测试服务可用性
```bash
# 测试API健康状态
curl http://127.0.0.1:8000/api/health

# 测试前端页面
curl -I http://127.0.0.1:8000/

# 测试GitLab访问
curl -I http://127.0.0.1:8000/gitlab/
```

## 访问地址

- **前端应用**: http://127.0.0.1:8000/
- **GitLab**: http://127.0.0.1:8000/gitlab/
- **后端API**: http://127.0.0.1:8000/api/
- **文档服务**: http://127.0.0.1:8000/onlyoffice/

## 故障排除

### 常见问题

#### 1. Docker权限问题
```bash
# 症状：permission denied while trying to connect to Docker daemon
# 解决：
sudo usermod -aG docker $USER
newgrp docker
```

#### 2. GitLab启动超时
```bash
# 症状：GitLab健康检查失败
# 解决：增加等待时间，检查系统资源
docker-compose logs gitlab
```

#### 3. OAuth配置文件未生成
```bash
# 症状：backend服务等待OAuth配置超时
# 解决：手动运行修复脚本
docker exec gitlabex-gitlab gitlab-rails runner /scripts/init-gitlab-oauth-fixed.rb
```

#### 4. Backend服务不健康
```bash
# 症状：backend容器状态为unhealthy
# 解决：检查OAuth配置和数据库连接
docker logs gitlabex-backend
```

### 调试命令

```bash
# 查看所有容器状态
docker-compose ps

# 查看特定服务日志
docker-compose logs [service-name]

# 重新运行初始化
docker-compose up gitlab-init

# 重启特定服务
docker-compose restart [service-name]

# 完全重新部署
docker-compose down -v
docker-compose up -d
```

## 技术特性

### 容器健康检查
- **GitLab**: 检查help页面可访问性
- **Backend**: 检查health API端点
- **Nginx**: 检查代理转发功能

### 依赖关系控制
- **精确依赖**: 使用 `condition: service_healthy`
- **启动顺序**: 确保严格的启动顺序
- **重启策略**: 支持服务的动态重启

### 配置管理
- **动态加载**: Backend支持运行时加载OAuth配置
- **共享卷**: 使用Docker卷共享配置文件
- **环境变量**: 支持环境变量覆盖配置

---

*部署指南版本: v1.0*  
*最后更新: 2025年7月13日* 