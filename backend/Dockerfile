# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装依赖
RUN apk add --no-cache git

# 设置Go代理为中国镜像源
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn

# 复制go.mod和go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# 生产阶段
FROM alpine:latest

WORKDIR /app

# 从构建阶段复制二进制文件和模板文件
COPY --from=builder /app/main .
COPY --from=builder /app/templates ./templates/

# 复制启动脚本
COPY start.sh .
RUN chmod +x start.sh

# 暴露端口
EXPOSE 8080

# 使用启动脚本运行应用
CMD ["./start.sh"] 