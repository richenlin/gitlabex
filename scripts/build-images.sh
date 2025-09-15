#!/bin/bash

# GitLabEx 镜像构建脚本
# 用于构建生产环境所需的 Docker 镜像

set -e

echo "🏗️  开始构建 GitLabEx 生产镜像..."

# 检查 Docker 是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker 未运行，请先启动 Docker"
    exit 1
fi

# 切换到项目根目录
cd "$(dirname "$0")/.."

# 构建标志
BUILD_BACKEND=true
BUILD_FRONTEND=true
FORCE_REBUILD=false
API_URL="/api"

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --backend-only)
            BUILD_BACKEND=true
            BUILD_FRONTEND=false
            shift
            ;;
        --frontend-only)
            BUILD_BACKEND=false
            BUILD_FRONTEND=true
            shift
            ;;
        --force)
            FORCE_REBUILD=true
            shift
            ;;
        --api-url)
            API_URL="$2"
            shift 2
            ;;
        -h|--help)
            echo "用法: $0 [选项]"
            echo ""
            echo "选项:"
            echo "  --backend-only       只构建后端镜像"
            echo "  --frontend-only      只构建前端镜像"
            echo "  --force              强制重新构建（不使用缓存）"
            echo "  --api-url <url>      设置前端API地址（默认: /api）"
            echo "  -h, --help           显示此帮助信息"
            echo ""
            echo "示例:"
            echo "  $0                           # 构建所有镜像"
            echo "  $0 --backend-only            # 只构建后端镜像"
            echo "  $0 --force                   # 强制重新构建所有镜像"
            echo "  $0 --api-url http://api.example.com  # 设置自定义API地址"
            exit 0
            ;;
        *)
            echo "未知选项: $1"
            echo "使用 -h 或 --help 查看帮助信息"
            exit 1
            ;;
    esac
done

# 设置构建参数
BUILD_ARGS=""
if [ "$FORCE_REBUILD" = true ]; then
    BUILD_ARGS="--no-cache"
    echo "🔄 强制重新构建模式（不使用缓存）"
fi

# 构建后端镜像
if [ "$BUILD_BACKEND" = true ]; then
    echo ""
    echo "🔨 构建后端镜像..."
    
    # 检查后端目录
    if [ ! -d "backend" ]; then
        echo "❌ 后端目录不存在"
        exit 1
    fi
    
    if [ ! -f "backend/Dockerfile" ]; then
        echo "❌ 后端 Dockerfile 不存在"
        exit 1
    fi
    
    echo "   📦 构建 gitlabex-backend:latest..."
    if docker build $BUILD_ARGS -t gitlabex-backend:latest backend/; then
        echo "   ✅ 后端镜像构建成功"
        
        # 显示镜像信息
        echo "   📊 镜像信息:"
        docker images gitlabex-backend:latest --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"
    else
        echo "   ❌ 后端镜像构建失败"
        exit 1
    fi
fi

# 构建前端镜像
if [ "$BUILD_FRONTEND" = true ]; then
    echo ""
    echo "🎨 构建前端镜像..."
    
    # 检查前端目录
    if [ ! -d "frontend" ]; then
        echo "❌ 前端目录不存在"
        exit 1
    fi
    
    if [ ! -f "frontend/Dockerfile" ]; then
        echo "❌ 前端 Dockerfile 不存在"
        exit 1
    fi
    
    echo "   📦 构建 gitlabex-frontend:latest..."
    if docker build $BUILD_ARGS -t gitlabex-frontend:latest frontend/; then
        echo "   ✅ 前端镜像构建成功"
        
        # 显示镜像信息
        echo "   📊 镜像信息:"
        docker images gitlabex-frontend:latest --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"
    else
        echo "   ❌ 前端镜像构建失败"
        exit 1
    fi
fi

echo ""
echo "🎉 镜像构建完成!"
echo ""

# 显示所有 GitLabEx 镜像
echo "📋 GitLabEx 镜像列表:"
if docker images | grep -q "gitlabex"; then
    docker images | head -1  # 显示表头
    docker images | grep "gitlabex"
else
    echo "   没有找到 GitLabEx 镜像"
fi

echo ""
echo "🚀 下一步操作:"
echo "   启动生产环境: ./scripts/start-services.sh"
echo "   查看镜像详情: docker inspect gitlabex-backend:latest"
echo "   删除镜像: docker rmi gitlabex-backend:latest gitlabex-frontend:latest"
echo ""

# 显示镜像总大小
TOTAL_SIZE=$(docker images gitlabex-backend:latest gitlabex-frontend:latest --format "{{.Size}}" 2>/dev/null | sed 's/[^0-9.]//g' | awk '{sum += $1} END {print sum}')
if [ ! -z "$TOTAL_SIZE" ] && [ "$TOTAL_SIZE" != "0" ]; then
    echo "📊 镜像统计:"
    echo "   总大小: ${TOTAL_SIZE}MB (约)"
fi

echo "📚 更多信息请查看 README.md 和 DEPLOYMENT.md"
