# GitLabEx 部署指南

## 🎯 部署概述

GitLabEx 教育协作平台支持多种部署方式，从开发环境到生产环境的完整部署解决方案。

## 🛠️ 部署架构

### 服务组件

```
┌─────────────────────────────────────────────────────────┐
│                    GitLabEx 平台                         │
├─────────────────────────────────────────────────────────┤
│  前端应用 (Vue 3)    │  后端服务 (Go)    │  GitLab CE     │
│  - 用户界面          │  - API 服务       │  - 版本控制     │
│  - 静态资源          │  - 业务逻辑       │  - 协作功能     │
│  Port: 3000         │  Port: 8080      │  Port: 8081    │
├─────────────────────────────────────────────────────────┤
│           数据存储层                                      │
│  PostgreSQL (5432)  │  Redis (6379)    │  文件存储       │
│  - 业务数据          │  - 缓存数据       │  - 上传文件     │
│  - 用户信息          │  - 会话数据       │  - 静态资源     │
└─────────────────────────────────────────────────────────┘
```

## 🔧 开发环境部署

### 快速启动（推荐）

```bash
# 1. 克隆项目
git clone <repository-url>
cd gitlabex2

# 2. 一键启动
./scripts/start-complete-system.sh

# 3. 验证系统
./scripts/verify-system.sh
```

### 手动部署

```bash
# 1. 启动基础服务
docker-compose -f docker-compose.dev.yml up postgres redis gitlab -d

# 2. 配置 GitLab OAuth
./scripts/configure-oauth.sh

# 3. 启动后端
cd backend
go mod tidy
go build -o bin/main cmd/main.go

# 设置环境变量
export DATABASE_URL="postgres://gitlabex:password123@localhost:5432/gitlabex?sslmode=disable"
export REDIS_URL="redis://:password123@localhost:6379"
export GITLAB_URL="http://localhost:8081"
export GITLAB_CLIENT_ID="your_client_id"
export GITLAB_CLIENT_SECRET="your_client_secret"
export JWT_SECRET="your_jwt_secret"
export FRONTEND_URL="http://localhost:3000"

./bin/main

# 4. 启动前端
cd ../frontend
npm install
npm run dev

# 5. 初始化测试数据
cd ..
./scripts/init-test-data.sh
```

## 🚀 生产环境部署

### 环境要求

**硬件要求:**
- CPU: 4核心以上
- 内存: 8GB以上
- 存储: 100GB以上 SSD
- 网络: 稳定的互联网连接

**软件要求:**
- Docker 20.10+
- Docker Compose 2.0+
- 域名和SSL证书
- 反向代理 (Nginx/Traefik)

### 1. 服务器准备

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装 Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 安装 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 创建应用目录
sudo mkdir -p /opt/gitlabex
sudo chown $USER:$USER /opt/gitlabex
cd /opt/gitlabex
```

### 2. 生产环境配置

创建生产环境配置文件 `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  # 前端服务
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile.prod
    container_name: gitlabex-frontend-prod
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./ssl:/etc/ssl/certs
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - backend

  # 后端服务
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: gitlabex-backend-prod
    restart: unless-stopped
    ports:
      - "8080:8080"
    env_file:
      - ./config/production.env
    depends_on:
      - postgres
      - redis
      - gitlab
    volumes:
      - ./uploads:/app/uploads

  # 数据库服务
  postgres:
    image: postgres:15
    container_name: gitlabex-postgres-prod
    restart: unless-stopped
    environment:
      POSTGRES_DB: gitlabex
      POSTGRES_USER: gitlabex
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./config/init-postgres.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"

  # Redis 服务
  redis:
    image: redis:7-alpine
    container_name: gitlabex-redis-prod
    restart: unless-stopped
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"

  # GitLab 服务
  gitlab:
    image: gitlab/gitlab-ce:latest
    container_name: gitlabex-gitlab-prod
    restart: unless-stopped
    hostname: gitlab.yourdomain.com
    environment:
      GITLAB_OMNIBUS_CONFIG: |
        external_url 'https://gitlab.yourdomain.com'
        gitlab_rails['gitlab_shell_ssh_port'] = 2222
        # SSL 配置
        nginx['ssl_certificate'] = "/etc/ssl/certs/gitlab.crt"
        nginx['ssl_certificate_key'] = "/etc/ssl/private/gitlab.key"
    volumes:
      - gitlab_config:/etc/gitlab
      - gitlab_logs:/var/log/gitlab
      - gitlab_data:/var/opt/gitlab
      - ./ssl:/etc/ssl
    ports:
      - "80:80"
      - "443:443"
      - "2222:22"

volumes:
  postgres_data:
  redis_data:
  gitlab_config:
  gitlab_logs:
  gitlab_data:

networks:
  default:
    name: gitlabex-prod-network
```

### 3. 环境变量配置

创建 `config/production.env`:

```bash
# 应用配置
APP_ENV=production
APP_DEBUG=false
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# 数据库配置
DATABASE_URL=postgres://gitlabex:${POSTGRES_PASSWORD}@postgres:5432/gitlabex?sslmode=require
REDIS_URL=redis://:${REDIS_PASSWORD}@redis:6379

# GitLab 配置
GITLAB_URL=https://gitlab.yourdomain.com
GITLAB_CLIENT_ID=${GITLAB_CLIENT_ID}
GITLAB_CLIENT_SECRET=${GITLAB_CLIENT_SECRET}

# JWT 配置
JWT_SECRET=${JWT_SECRET}
JWT_EXPIRATION_HOURS=24

# 前端配置
FRONTEND_URL=https://yourdomain.com

# API 密钥
SYNC_API_KEY=${SYNC_API_KEY}
THIRD_PARTY_API_KEY=${THIRD_PARTY_API_KEY}

# 文件上传配置
UPLOAD_MAX_SIZE=100MB
UPLOAD_PATH=/app/uploads

# 日志配置
LOG_LEVEL=info
LOG_FORMAT=json
```

### 4. SSL 证书配置

```bash
# 使用 Let's Encrypt 获取免费证书
sudo apt install certbot

# 获取证书
sudo certbot certonly --standalone -d yourdomain.com -d gitlab.yourdomain.com

# 复制证书到项目目录
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem ./ssl/
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem ./ssl/
sudo chown $USER:$USER ./ssl/*
```

### 5. Nginx 反向代理配置

创建 `nginx.conf`:

```nginx
events {
    worker_connections 1024;
}

http {
    upstream backend {
        server backend:8080;
    }

    server {
        listen 80;
        server_name yourdomain.com;
        return 301 https://$server_name$request_uri;
    }

    server {
        listen 443 ssl http2;
        server_name yourdomain.com;

        ssl_certificate /etc/ssl/certs/fullchain.pem;
        ssl_certificate_key /etc/ssl/private/privkey.pem;

        # 前端静态文件
        location / {
            root /usr/share/nginx/html;
            try_files $uri $uri/ /index.html;
        }

        # 后端 API
        location /api/ {
            proxy_pass http://backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # 文件上传
        client_max_body_size 100M;
    }
}
```

### 6. 部署脚本

创建生产部署脚本 `scripts/deploy-production.sh`:

```bash
#!/bin/bash

echo "🚀 GitLabEx 生产环境部署"
echo "======================="

# 检查环境变量
if [ -z "$POSTGRES_PASSWORD" ]; then
    echo "❌ 请设置 POSTGRES_PASSWORD 环境变量"
    exit 1
fi

# 拉取最新代码
git pull origin main

# 构建镜像
docker-compose -f docker-compose.prod.yml build --no-cache

# 备份数据库
echo "📦 备份数据库..."
docker exec gitlabex-postgres-prod pg_dump -U gitlabex gitlabex > backup_$(date +%Y%m%d_%H%M%S).sql

# 停止旧服务
docker-compose -f docker-compose.prod.yml down

# 启动新服务
docker-compose -f docker-compose.prod.yml up -d

# 等待服务启动
sleep 30

# 验证部署
if curl -f https://yourdomain.com/health; then
    echo "✅ 部署成功！"
else
    echo "❌ 部署失败，请检查服务状态"
    exit 1
fi
```

## 📊 监控和维护

### 1. 系统监控

```bash
# 服务状态监控
docker-compose -f docker-compose.prod.yml ps

# 资源使用监控
docker stats

# 日志监控
docker-compose -f docker-compose.prod.yml logs -f
```

### 2. 数据备份

```bash
# 数据库备份脚本
#!/bin/bash
BACKUP_DIR="/opt/backups/gitlabex"
DATE=$(date +%Y%m%d_%H%M%S)

# 创建备份目录
mkdir -p $BACKUP_DIR

# 备份数据库
docker exec gitlabex-postgres-prod pg_dump -U gitlabex gitlabex > $BACKUP_DIR/database_$DATE.sql

# 备份 GitLab 数据
docker exec gitlabex-gitlab-prod gitlab-rake gitlab:backup:create

# 压缩备份文件
tar -czf $BACKUP_DIR/gitlabex_backup_$DATE.tar.gz $BACKUP_DIR/database_$DATE.sql

# 清理旧备份（保留30天）
find $BACKUP_DIR -name "*.sql" -mtime +30 -delete
find $BACKUP_DIR -name "*.tar.gz" -mtime +30 -delete
```

### 3. 性能优化

**数据库优化:**
```sql
-- 创建索引
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_research_projects_status ON research_projects(status);
CREATE INDEX idx_submissions_homework_student ON submissions(homework_id, student_id);

-- 定期维护
VACUUM ANALYZE;
```

**Redis 优化:**
```bash
# Redis 配置优化
echo "maxmemory 2gb" >> redis.conf
echo "maxmemory-policy allkeys-lru" >> redis.conf
```

## 🔒 安全配置

### 1. 防火墙设置

```bash
# 允许必要端口
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw allow 2222/tcp  # GitLab SSH

# 启用防火墙
sudo ufw enable
```

### 2. SSL/TLS 配置

```bash
# 自动续期证书
echo "0 3 * * * certbot renew --quiet && docker-compose -f /opt/gitlabex/docker-compose.prod.yml restart frontend" | sudo crontab -
```

### 3. 安全加固

```bash
# 限制文件权限
chmod 600 config/production.env
chmod 600 config/oauth.env

# 设置安全的环境变量
export JWT_SECRET=$(openssl rand -base64 64)
export POSTGRES_PASSWORD=$(openssl rand -base64 32)
export REDIS_PASSWORD=$(openssl rand -base64 32)
```

## 🔄 CI/CD 部署

### GitHub Actions 配置

创建 `.github/workflows/deploy.yml`:

```yaml
name: Deploy to Production

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup SSH
      uses: webfactory/ssh-agent@v0.7.0
      with:
        ssh-private-key: ${{ secrets.SSH_PRIVATE_KEY }}
    
    - name: Deploy to server
      run: |
        ssh -o StrictHostKeyChecking=no user@your-server.com << 'EOF'
          cd /opt/gitlabex
          git pull origin main
          ./scripts/deploy-production.sh
        EOF
```

### GitLab CI/CD 配置

创建 `.gitlab-ci.yml`:

```yaml
stages:
  - build
  - test
  - deploy

variables:
  DOCKER_DRIVER: overlay2

build-backend:
  stage: build
  script:
    - cd backend
    - go build -o bin/main cmd/main.go

build-frontend:
  stage: build
  script:
    - cd frontend
    - npm install
    - npm run build

test:
  stage: test
  script:
    - go test ./...
    - cd frontend && npm run test

deploy-production:
  stage: deploy
  script:
    - ./scripts/deploy-production.sh
  only:
    - main
  when: manual
```

## 📈 扩容和负载均衡

### 水平扩容

```yaml
# docker-compose.scale.yml
version: '3.8'

services:
  backend:
    # ... 基础配置
    deploy:
      replicas: 3
    
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx-lb.conf:/etc/nginx/nginx.conf
```

### 负载均衡配置

```nginx
# nginx-lb.conf
upstream backend_servers {
    server backend_1:8080;
    server backend_2:8080;
    server backend_3:8080;
}

server {
    listen 80;
    
    location /api/ {
        proxy_pass http://backend_servers;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 🔍 监控和告警

### 1. 健康检查

```bash
# 创建健康检查脚本
#!/bin/bash
# scripts/health-check.sh

services=("frontend" "backend" "postgres" "redis" "gitlab")
for service in "${services[@]}"; do
    if ! docker-compose -f docker-compose.prod.yml ps $service | grep -q "Up"; then
        echo "❌ $service 服务异常"
        # 发送告警通知
        curl -X POST "https://hooks.slack.com/your-webhook" \
             -H 'Content-type: application/json' \
             --data '{"text":"GitLabEx '$service' 服务异常"}'
    fi
done
```

### 2. 日志聚合

```yaml
# 添加日志收集服务
  logging:
    image: grafana/loki:latest
    ports:
      - "3100:3100"
    volumes:
      - ./loki-config.yml:/etc/loki/local-config.yaml

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana
```

### 3. 性能监控

```bash
# 安装监控工具
docker run -d \
  --name=prometheus \
  -p 9090:9090 \
  -v ./prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus

# Prometheus 配置
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'gitlabex-backend'
    static_configs:
      - targets: ['localhost:8080']
```

## 🔄 升级和回滚

### 升级流程

```bash
# 1. 备份数据
./scripts/backup-data.sh

# 2. 拉取新版本
git pull origin main

# 3. 更新服务
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d

# 4. 验证升级
./scripts/verify-system.sh
```

### 回滚流程

```bash
# 1. 停止当前服务
docker-compose -f docker-compose.prod.yml down

# 2. 恢复代码版本
git checkout <previous-commit>

# 3. 恢复数据库
docker exec -i gitlabex-postgres-prod psql -U gitlabex -d gitlabex < backup_file.sql

# 4. 重启服务
docker-compose -f docker-compose.prod.yml up -d
```

## 🛡️ 安全检查清单

### 部署前检查

- [ ] 更新所有默认密码
- [ ] 配置 SSL/TLS 证书
- [ ] 设置防火墙规则
- [ ] 配置安全的环境变量
- [ ] 启用访问日志
- [ ] 配置备份策略
- [ ] 设置监控告警
- [ ] 测试故障恢复流程

### 运行时安全

- [ ] 定期更新系统补丁
- [ ] 监控异常访问行为
- [ ] 检查日志异常记录
- [ ] 验证数据备份完整性
- [ ] 测试灾难恢复计划

## 📞 技术支持

### 故障排除

**常见问题:**

1. **服务无法启动**
   ```bash
   # 检查端口占用
   netstat -tulpn | grep :8080
   
   # 查看服务日志
   docker-compose logs backend
   ```

2. **数据库连接失败**
   ```bash
   # 检查数据库状态
   docker exec gitlabex-postgres-prod pg_isready
   
   # 测试连接
   docker exec -it gitlabex-postgres-prod psql -U gitlabex -d gitlabex
   ```

3. **GitLab OAuth 错误**
   ```bash
   # 检查 OAuth 配置
   cat config/oauth.env
   
   # 重新配置
   ./scripts/configure-oauth.sh
   ```

### 支持联系方式

- 📧 **技术支持**: support@gitlabex.com
- 📖 **文档中心**: https://docs.gitlabex.com
- 💬 **社区论坛**: https://community.gitlabex.com
- 🐛 **问题反馈**: https://github.com/gitlabex/issues

---

**GitLabEx - 让教育协作更简单！** 🎓🚀