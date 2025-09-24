#!/bin/bash

# GitLabEx 配置向导
# 用于引导用户完成系统配置

set -e

echo "🔧 GitLabEx 配置向导"
echo "========================"
echo ""

# 切换到项目根目录
cd "$(dirname "$0")/.."

# 询问用户要更新哪个配置文件
echo "📋 选择要更新的配置文件:"
echo "   1) 开发环境 (config/config.yml)"
echo "   2) 生产环境 (config/config.prod.yml)"
echo ""
read -p "请选择 (1-2): " -n 1 -r
echo

case $REPLY in
    1)
        config_file="config/config.yml"
        ;;
    2)
        config_file="config/config.prod.yml"
        ;;
    *)
        echo "❌ 无效选择，默认使用开发环境配置"
        config_file="config/config.yml"
        ;;
esac

# 备份原配置文件
if [ -f "$config_file" ]; then
    cp "$config_file" "$config_file.bak"
    echo "📦 已备份原配置文件到 $config_file.bak"
fi

echo ""
echo "📝 开始配置 GitLab OAuth 应用"
echo "=============================="

# 显示 GitLab OAuth 配置步骤
echo ""
echo "1️⃣  GitLab OAuth 应用配置:"
echo "   🌐 访问: http://localhost:8081/admin/applications"
echo "   👤 用户名: root"
echo "   🔑 密码: b75hZ0qcwLKD"
echo "   📍 导航到: Admin Area → Applications → New application"
echo "   📝 应用名称: GitLabEx Education Platform"
echo "   🔗 重定向URI: 请输入重定向 URI(格式: http://域名或IP:端口/auth/gitlab/callback, 默认: http://localhost:3000/auth/gitlab/callback)"
echo "   ✅ 权限范围: api, read_api, openid"
echo "   💾 保存应用后复制 Application ID 和 Secret"
echo ""

# 获取 GitLab 配置
read -p "🔗 请输入 GitLab 外部访问 URL (格式: http://域名或IP:端口, 默认: http://localhost:8081): " gitlab_url
gitlab_url=${gitlab_url:-"http://localhost:8081"}

read -p "🔗 请输入 前端服务 外部访问 URL (格式: http://域名或IP:端口, 默认: http://localhost:3000): " frontend_url
frontend_url=${frontend_url:-"http://localhost:3000"}

read -p "🔑 请输入 Application ID: " client_id
read -p "🔐 请输入 Secret: " client_secret

redirect_uri="${frontend_url}/auth/gitlab/callback"

read -p "🎯 请输入权限范围 (默认: api read_api openid): " scopes
scopes=${scopes:-"api read_api openid"}

echo ""
echo "2️⃣  GitLab 个人访问令牌 (system_token)"
echo "   🌐 访问: http://localhost:8081/-/profile/personal_access_tokens"
echo "   👤 使用 root 用户登录"
echo "   📝 令牌名称: GitLabEx System Token"
echo "   ✅ 权限范围: api, read_api, read_user, read_repository, write_repository"
echo "   💾 创建后复制生成的令牌"
echo ""

read -p "🔐 请输入 System Token: " system_token

echo ""
echo "3️⃣  JWT 密钥配置 (用户登录令牌)"
echo "   💡 建议使用强密码，长度至少32字符"
echo "   ⏩ 可直接回车使用随机生成的密钥"
echo ""

read -p "🔒 请输入 JWT Secret (回车使用随机生成): " jwt_secret
if [ -z "$jwt_secret" ]; then
    jwt_secret=$(openssl rand -base64 32 | tr -d '\n')
    echo "🎲 已生成随机 JWT Secret: $jwt_secret"
fi

echo ""
echo "4️⃣  CORS 配置 (前端访问地址)"
echo "   💡 需要将前端访问地址加入允许列表"
# echo "   🌐 请输入 CORS 允许的源(格式: http://域名或IP:端口, 多个用半角逗号分隔, 默认: http://localhost:3000,http://127.0.0.1:3000)"
# echo ""

# read -p "🌐 请输入 CORS 允许的源 (多个用逗号分隔): " cors_origins
cors_origins="${frontend_url},http://localhost:3000,http://127.0.0.1:3000"

echo ""
echo "5️⃣  第三方 API 密钥配置"
echo "   💡 用于第三方系统接入，可为空"
echo "   ⏩ 可直接回车使用随机生成的密钥"
echo ""

read -p "🔑 请输入第三方 API 密钥 (回车使用随机生成): " third_party_key
if [ -z "$third_party_key" ]; then
    third_party_key=$(openssl rand -base64 24 | tr -d '\n' | tr '+/' '_-')
    echo "🎲 已生成随机第三方 API 密钥: $third_party_key"
fi

# 验证必填项
if [ -z "$client_id" ] || [ -z "$client_secret" ] || [ -z "$system_token" ]; then
    echo "❌ Application ID、Secret 和 System Token 不能为空"
    exit 1
fi

# 更新配置文件
echo ""
echo "🔄 正在更新配置文件 $config_file ..."

# 更新 GitLab 配置
sed -i "s|url: \".*\"|url: \"$gitlab_url\"|g" "$config_file"
sed -i "s|client_id: \".*\"|client_id: \"$client_id\"|g" "$config_file"
sed -i "s|client_secret: \".*\"|client_secret: \"$client_secret\"|g" "$config_file"
sed -i "s|redirect_uri: \".*\"|redirect_uri: \"$redirect_uri\"|g" "$config_file"
sed -i "s|scopes: \".*\"|scopes: \"$scopes\"|g" "$config_file"
sed -i "s|system_token: \".*\"|system_token: \"$system_token\"|g" "$config_file"

# 更新 JWT 配置
sed -i "s|secret: \".*\"|secret: \"$jwt_secret\"|g" "$config_file"

# 更新 CORS 配置
sed -i "s|cors_allowed_origins: \".*\"|cors_allowed_origins: \"$cors_origins\"|g" "$config_file"

# 更新 API 密钥配置
sed -i "s|third_party_api_key: \".*\"|third_party_api_key: \"$third_party_key\"|g" "$config_file"

echo "✅ 配置已成功更新到 $config_file！"
echo ""

# 显示配置摘要
echo "📋 配置摘要:"
echo "============"
echo "📁 配置文件: $config_file"
echo "🔗 GitLab URL: $gitlab_url"
echo "🔑 Client ID: $client_id"
echo "🔐 Client Secret: **********"
echo "🔄 Redirect URI: $redirect_uri"
echo "🎯 Scopes: $scopes"
echo "🔐 System Token: **********"
echo "🔒 JWT Secret: **********"
echo "🌐 CORS Origins: $cors_origins"
echo "🔑 Third Party Key: **********"
echo ""

echo "🎉 配置文件更新完成！"
echo ""
