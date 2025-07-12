# GitLabEx 配置文件说明

## OAuth 配置文件 (oauth.env)

`oauth.env` 文件用于配置GitLab OAuth应用的相关参数，支持不同的部署环境。

### 配置项说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `GITLAB_INTERNAL_URL` | GitLab内部访问地址，用于后端API调用 | `http://gitlab` |
| `GITLAB_EXTERNAL_URL` | GitLab外部访问地址，用于构建OAuth授权URL | `http://127.0.0.1:8000/gitlab` |
| `GITLAB_OAUTH_REDIRECT_URI` | OAuth回调地址 | `http://127.0.0.1:8000/api/auth/gitlab/callback` |
| `GITLAB_OAUTH_APP_NAME` | OAuth应用名称 | `GitLabEx Education Platform` |
| `GITLAB_OAUTH_SCOPES` | OAuth权限范围 | `read_user read_repository write_repository` |
| `FORCE_RECREATE_OAUTH_APP` | 是否强制重新创建OAuth应用 | `false` |

### 部署场景配置示例

#### 1. 本地Docker开发环境 (默认)

```env
GITLAB_INTERNAL_URL=http://gitlab
GITLAB_EXTERNAL_URL=http://127.0.0.1:8000/gitlab
GITLAB_OAUTH_REDIRECT_URI=http://127.0.0.1:8000/api/auth/gitlab/callback
```

#### 2. 使用外部GitLab服务

```env
GITLAB_INTERNAL_URL=https://gitlab.example.com
GITLAB_EXTERNAL_URL=https://gitlab.example.com
GITLAB_OAUTH_REDIRECT_URI=https://your-app.example.com/api/auth/gitlab/callback
```

#### 3. 生产环境

```env
GITLAB_INTERNAL_URL=http://gitlab.internal
GITLAB_EXTERNAL_URL=https://git.yourcompany.com
GITLAB_OAUTH_REDIRECT_URI=https://gitlabex.yourcompany.com/api/auth/gitlab/callback
```

### 配置流程

1. **编辑配置文件**: 根据您的部署环境修改 `config/oauth.env` 文件
2. **重新部署**: 运行 `docker compose down && docker compose up -d`
3. **OAuth应用自动创建**: `gitlab-init` 容器会自动读取配置并创建GitLab OAuth应用
4. **配置文件生成**: 生成的OAuth凭据会保存到 `/shared/gitlab-oauth.env`

### 注意事项

- **URL配置**: 确保 `GITLAB_EXTERNAL_URL` 是用户浏览器可以访问的地址
- **回调地址**: `GITLAB_OAUTH_REDIRECT_URI` 必须是应用的公网访问地址
- **强制重创**: 设置 `FORCE_RECREATE_OAUTH_APP=true` 会删除现有OAuth应用并重新创建
- **兼容性**: 系统会自动为 `127.0.0.1` 和 `localhost` 创建兼容的回调地址

### 故障排除

#### OAuth授权错误: "The redirect URI included is not valid"

1. 检查 `GITLAB_OAUTH_REDIRECT_URI` 是否与实际访问地址匹配
2. 确保回调地址是完整的URL (包含协议、域名/IP、端口)
3. 设置 `FORCE_RECREATE_OAUTH_APP=true` 重新创建OAuth应用

#### GitLab无法访问

1. 检查 `GITLAB_INTERNAL_URL` 是否正确
2. 确保后端容器能够访问GitLab服务
3. 检查网络连接和防火墙设置

#### OAuth应用创建失败

1. 检查GitLab是否正常启动
2. 查看 `gitlab-init` 容器日志: `docker compose logs gitlab-init`
3. 确保配置文件格式正确 (无语法错误) 