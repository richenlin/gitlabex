#!/bin/bash

# GitLabEx 配置向导
# 用于引导用户完成系统配置

set -e

echo "🔧 GitLabEx 配置向导"
echo "========================"
echo ""

# 切换到项目根目录
cd "$(dirname "$0")/.."

# 询问是否要配置 Docker Compose
read -p "🤔 是否要配置 Docker Compose 文件？(y/N): " -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo "🐳 Docker Compose 配置"
    echo "======================"
    
    # 选择要更新的 Docker Compose 文件
    echo "📋 选择要更新的 Docker Compose 文件:"
    echo "   1) 开发环境 (docker-compose.yml)"
    echo "   2) 生产环境 (docker-compose.prod.yml)"
    echo ""
    read -p "请选择 (1-2): " -n 1 -r
    echo

    case $REPLY in
        1)
            compose_file="docker-compose.yml"
            ;;
        2)
            compose_file="docker-compose.prod.yml"
            ;;
        *)
            echo "❌ 无效选择，默认使用开发环境配置"
            compose_file="docker-compose.yml"
            ;;
    esac

    # 备份原文件
    if [ -f "$compose_file" ]; then
        cp "$compose_file" "$compose_file.bak"
        echo "📦 已备份原文件到 $compose_file.bak"
    fi

    echo ""
    echo "1️⃣  GitLab 根密码配置"
    read -p "🔑 请输入 GitLab root 密码 (回车使用默认值 b75hZ0qcwLKD): " gitlab_root_password
    gitlab_root_password=${gitlab_root_password:-"b75hZ0qcwLKD"}

    echo ""
    echo "2️⃣  GitLab 网络配置"
    read -p "🌐 请输入 GitLab 外部访问 URL (格式: http://域名或IP:端口): " gitlab_external_url
    
    # 解析 URL 获取主机和端口
    if [[ "$gitlab_external_url" =~ ^https?://([^:/]+)(:([0-9]+))? ]]; then
        gitlab_host="${BASH_REMATCH[1]}"
        gitlab_port="${BASH_REMATCH[3]:-80}"
        
        # 判断是否使用 HTTPS
        if [[ "$gitlab_external_url" =~ ^https:// ]]; then
            gitlab_https="true"
            if [ "$gitlab_port" = "443" ]; then
                gitlab_port="80"  # 外部访问用443，内部用80
            fi
        else
            gitlab_https="false"
        fi
    else
        echo "❌ URL 格式不正确，使用默认配置"
        gitlab_host="127.0.0.1"
        gitlab_port="8081"
        gitlab_https="false"
    fi

    # echo ""
    # echo "3️⃣  GitLab 数据库配置"
    # read -p "🗄️  请输入数据库主机 (回车使用默认值 gitlabex-postgres): " db_host
    # db_host=${db_host:-"gitlabex-postgres"}
    
    # read -p "🔢 请输入数据库端口 (回车使用默认值 5432): " db_port
    # db_port=${db_port:-"5432"}
    
    # read -p "👤 请输入数据库用户名 (回车使用默认值 gitlab): " db_username
    # db_username=${db_username:-"gitlab"}
    
    # read -p "🔐 请输入数据库密码 (回车使用默认值 password123): " db_password
    # db_password=${db_password:-"password123"}
    
    # read -p "📊 请输入数据库名 (回车使用默认值 gitlab): " db_database
    # db_database=${db_database:-"gitlab"}

    # echo ""
    # echo "4️⃣  GitLab Redis 配置"
    # read -p "🔴 请输入 Redis 主机 (回车使用默认值 gitlabex-redis): " redis_host
    # redis_host=${redis_host:-"gitlabex-redis"}
    
    # read -p "🔢 请输入 Redis 端口 (回车使用默认值 6379): " redis_port
    # redis_port=${redis_port:-"6379"}
    
    # read -p "🔐 请输入 Redis 密码 (回车使用默认值 password123): " redis_password
    # redis_password=${redis_password:-"password123"}

    echo ""
    echo "🔄 正在更新 Docker Compose 文件 $compose_file ..."

    # 更新 GitLab root 密码
    sed -i "s|GITLAB_ROOT_PASSWORD: \${GITLAB_ROOT_PASSWORD:-.*}|GITLAB_ROOT_PASSWORD: \${GITLAB_ROOT_PASSWORD:-$gitlab_root_password}|g" "$compose_file"

    # 更新 GitLab 外部 URL 配置
    if [ -n "$gitlab_external_url" ]; then
        # 添加 external_url 配置（如果不存在）
        if ! grep -q "external_url" "$compose_file"; then
            sed -i "/GITLAB_OMNIBUS_CONFIG: |/a\        external_url '$gitlab_external_url'" "$compose_file"
        else
            sed -i "s|external_url '.*'|external_url '$gitlab_external_url'|g" "$compose_file"
        fi
        
        # 更新 GitLab 主机配置
        sed -i "s|gitlab_rails\['gitlab_host'\] = '.*'|gitlab_rails['gitlab_host'] = '$gitlab_host'|g" "$compose_file"
        sed -i "s|gitlab_rails\['gitlab_port'\] = .*|gitlab_rails['gitlab_port'] = $gitlab_port|g" "$compose_file"
        sed -i "s|gitlab_rails\['gitlab_https'\] = .*|gitlab_rails['gitlab_https'] = $gitlab_https|g" "$compose_file"
    fi

    # 更新数据库配置
    # sed -i "s|gitlab_rails\['db_host'\] = '.*'|gitlab_rails['db_host'] = '$db_host'|g" "$compose_file"
    # sed -i "s|gitlab_rails\['db_port'\] = [0-9]*|gitlab_rails['db_port'] = $db_port|g" "$compose_file"
    # sed -i "s|gitlab_rails\['db_username'\] = '.*'|gitlab_rails['db_username'] = '$db_username'|g" "$compose_file"
    # sed -i "s|gitlab_rails\['db_password'\] = '.*'|gitlab_rails['db_password'] = '$db_password'|g" "$compose_file"
    # sed -i "s|gitlab_rails\['db_database'\] = '.*'|gitlab_rails['db_database'] = '$db_database'|g" "$compose_file"

    # 更新 Redis 配置
    # sed -i "s|gitlab_rails\['redis_host'\] = '.*'|gitlab_rails['redis_host'] = '$redis_host'|g" "$compose_file"
    # sed -i "s|gitlab_rails\['redis_port'\] = [0-9]*|gitlab_rails['redis_port'] = $redis_port|g" "$compose_file"
    # sed -i "s|gitlab_rails\['redis_password'\] = '.*'|gitlab_rails['redis_password'] = '$redis_password'|g" "$compose_file"

    # 更新 PostgreSQL 服务配置
    # sed -i "s|POSTGRES_USER: .*|POSTGRES_USER: $db_username|g" "$compose_file"
    # sed -i "s|POSTGRES_PASSWORD: .*|POSTGRES_PASSWORD: $db_password|g" "$compose_file"
    # sed -i "s|POSTGRES_DB: .*|POSTGRES_DB: $db_database|g" "$compose_file"

    # 更新 Redis 服务配置
    # sed -i "s|command: redis-server --requirepass .*|command: redis-server --requirepass $redis_password|g" "$compose_file"
    # sed -i "s|test: \\[\"CMD\", \"redis-cli\", \"-a\", \".*\", \"ping\"\\]|test: [\"CMD\", \"redis-cli\", \"-a\", \"$redis_password\", \"ping\"]|g" "$compose_file"

    # 更新 PostgreSQL 初始化脚本
    # echo ""
    # echo "🔄 正在更新数据库初始化脚本 config/init-postgres.sql ..."
    
    # 备份原文件
    # if [ -f "config/init-postgres.sql" ]; then
    #     cp "config/init-postgres.sql" "config/init-postgres.sql.bak"
    #     echo "📦 已备份原初始化脚本到 config/init-postgres.sql.bak"
    # fi

    # 更新数据库初始化脚本中的用户名和密码
    # sed -i "s|CREATE USER gitlabex WITH ENCRYPTED PASSWORD '.*'|CREATE USER gitlabex WITH ENCRYPTED PASSWORD '$db_password'|g" "config/init-postgres.sql"
    
    # echo "✅ 数据库初始化脚本已更新！"

    echo "✅ Docker Compose 配置已成功更新到 $compose_file！"
    echo ""
    
    # 显示 Docker Compose 配置摘要
    echo "📋 Docker Compose 配置摘要:"
    echo "==========================="
    echo "📁 配置文件: $compose_file"
    echo "🔑 GitLab Root Password: **********"
    echo "🌐 GitLab External URL: $gitlab_external_url"
    echo "🏠 GitLab Host: $gitlab_host"
    echo "🚪 GitLab Port: $gitlab_port"
    echo "🔒 GitLab HTTPS: $gitlab_https"
    # echo "🗄️  DB Host: $db_host"
    # echo "🔢 DB Port: $db_port"
    # echo "👤 DB Username: $db_username"
    # echo "🔐 DB Password: **********"
    # echo "📊 DB Database: $db_database"
    # echo "🔴 Redis Host: $redis_host"
    # echo "🔢 Redis Port: $redis_port"
    # echo "🔐 Redis Password: **********"
    echo ""
fi

echo "🎉 所有配置完成！"
echo ""
echo "💡 下一步操作:"
echo "   1. 启动 Docker 服务: docker-compose -f $compose_file up -d"
echo "   2. 开发环境手动启动前/后端服务"
echo ""
echo "📚 详细说明请查看 README.md"
