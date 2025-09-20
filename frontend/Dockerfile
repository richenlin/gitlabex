# 构建阶段
FROM node:18-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 package 文件
COPY package*.json ./
COPY pnpm-lock.yaml* ./

# 安装 pnpm
RUN npm install -g pnpm

# 安装依赖
RUN pnpm install --frozen-lockfile

# 复制源代码
COPY . .

# 设置生产环境变量
ENV NODE_ENV=production

# 支持构建时配置API地址（默认使用相对路径）
ARG VITE_API_BASE_URL=/api
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

# 构建应用
RUN pnpm run build

# 生产阶段
FROM nginx:alpine

# 安装必要的工具
RUN apk --no-cache add curl wget

# 复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制入口脚本
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# 创建应用配置文件
RUN echo 'window._VBEN_ADMIN_PRO_APP_CONF_={"VITE_GLOB_API_URL":"/api"};Object.freeze(window._VBEN_ADMIN_PRO_APP_CONF_);Object.defineProperty(window,"_VBEN_ADMIN_PRO_APP_CONF_",{configurable:false,writable:false,});' > /usr/share/nginx/html/_app.config.js

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

# 暴露端口
EXPOSE 80

# 使用动态入口脚本
ENTRYPOINT ["/docker-entrypoint.sh"]
