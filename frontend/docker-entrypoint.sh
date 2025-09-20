#!/bin/sh

# GitLabEx Frontend Docker Entrypoint Script
# 动态生成nginx配置，支持环境变量

set -e

# 默认后端URL
BACKEND_URL=${BACKEND_URL:-gitlabex-backend:8080}

echo "🚀 启动GitLabEx前端服务..."
echo "📡 后端API地址: $BACKEND_URL"

# 生成动态的nginx配置
cat > /etc/nginx/conf.d/default.conf << EOF
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html;

    # 启用 gzip 压缩
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied expired no-cache no-store private auth;
    gzip_types text/plain text/css text/xml text/javascript application/javascript application/xml+rss application/json;

    # 处理前端路由
    location / {
        try_files \$uri \$uri/ /index.html;
        
        # 缓存静态资源
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)\$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }
    }

    # API 代理 - 支持多机部署
    location /api/ {
        # 使用环境变量设置的后端地址
        proxy_pass ${BACKEND_URL}/api/;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        
        # WebSocket 支持
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # 缓冲设置
        proxy_buffering off;
        proxy_request_buffering off;
    }

    # 安全头
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # 错误页面
    error_page 404 /index.html;
}
EOF

echo "✅ Nginx配置生成完成"

# 启动nginx
exec nginx -g 'daemon off;'
