# GitLab 登录信息

## 🔑 登录凭据

**GitLab访问地址**: http://localhost:8081

**管理员账号**:
- 用户名: `root`
- 密码: `b75hZ0qcwLKD`

sudo docker exec -it gitlabex-gitlab gitlab-rails runner "user = User.new(username: 'root', email: 'admin@example.com', name: 'Administrator', password: 'b75hZ0qcwLKD', password_confirmation: 'Gi
tLab@2024#SecurePass', admin: true);  user.save; puts 'Admin user created successfully'"

## 📋 密码说明

新密码符合GitLab安全要求：
- ✅ 包含大写字母 (G, L, S, P)
- ✅ 包含小写字母 (i, t, a, b, e, c, u, r, e, a, s, s)  
- ✅ 包含数字 (2024)
- ✅ 包含特殊字符 (@, #, !)
- ✅ 长度足够 (22个字符)
- ✅ 不包含常见单词组合

## 🚀 首次登录步骤

1. 访问 http://localhost:8081
2. 使用以下凭据登录：
   - Username: `root`
   - Password: `b75hZ0qcwLKD`
3. 登录成功后，可以进入 Admin Area 配置 OAuth 应用

## ⚙️ 配置 OAuth 应用

登录后按照以下步骤配置 OAuth：

1. 点击顶部菜单的 **Admin Area** (扳手图标)
2. 左侧菜单选择 **Applications**
3. 点击 **New Application**
4. 填写应用信息：
   - **Name**: `GitLabEx`
   - **Redirect URI**: `http://172.17.0.1:8080/api/auth/gitlab/callback`
   - **Scopes**: 选择以下权限
     - `read_user` - 读取用户信息
     - `read_repository` - 读取仓库信息
     - `openid` - OpenID Connect
     - `profile` - 用户配置文件
     - `email` - 电子邮件地址
5. 点击 **Save application**
6. 记录生成的 **Application ID** 和 **Secret**
**Application ID**： 3ebea6c37ff96533279ef2400925790fcd15af5e948fb4949392ad5f33be8f55
**Secret**：gloas-1348e62af7177f9ef6a97b0ee28ba6863c3984b8b3949c8a6b17bb0a9bf39800
## 🔧 更新后端配置

获得 OAuth 应用的 Application ID 和 Secret 后：

1. 编辑 `backend/.env` 文件
2. 更新以下配置：
   ```bash
   GITLAB_CLIENT_ID=your-application-id-here
   GITLAB_CLIENT_SECRET=your-application-secret-here
   ```
3. 重启后端服务：
   ```bash
   cd backend
   go run cmd/main.go
   ```

## 🧪 测试 OAuth 流程

1. 访问前端应用: http://localhost:5173
2. 点击 "使用 GitLab 登录"
3. 应该正确跳转到 GitLab OAuth 授权页面
4. 授权后自动跳转回应用完成登录 