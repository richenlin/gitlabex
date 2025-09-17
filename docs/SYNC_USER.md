# GitLabEx 第三方系统集成指南

本文档详细介绍如何将第三方系统与GitLabEx进行集成，实现用户账号同步和统一身份管理。

## 概述

GitLabEx提供了精简高效的第三方系统集成API，支持：
- **用户账号同步**：将外部系统用户直接创建到GitLab
- **标准OAuth登录**：用户通过GitLab OAuth流程登录GitLabEx
- **用户信息查询**：第三方系统可查询GitLab用户信息
- **安全的API认证**：使用第三方API密钥进行认证

## 工作流程

```
第三方系统 → API创建GitLab用户 → 用户通过GitLab OAuth登录 → 访问GitLabEx
```

这是最安全、标准且易于维护的集成方案。

## API认证机制

### API密钥认证

GitLabEx使用单一的第三方API密钥进行认证，简化了权限管理：

#### 第三方API密钥 (THIRD_PARTY_API_KEY)
- **用途**：第三方系统专用
- **权限**：
  - 无法创建管理员用户（安全考虑）
  - 每小时500次请求限制
  - 可以创建和查询用户信息

### 认证方式

所有API请求都需要在请求头中包含API密钥：

```http
X-API-Key: your_third_party_api_key_here
```

## API接口

### 基础URL

```
https://your-gitlabex-domain.com/api/v1/sync
```

### 1. 创建用户

**接口**：`POST /users`

**描述**：创建GitLab用户，如果用户已存在则直接返回用户信息

**请求体**：
```json
{
  "username": "john_doe",
  "password": "secure_password123",
  "email": "john.doe@example.com",
  "name": "John Doe",
  "role": "student"
}
```

**字段说明**：
- `username`: 用户名 (必需，3-50字符)
- `password`: 密码 (必需，最少8字符,包含大小写字母、数字和特殊字符)
- `email`: 邮箱 (必需，有效邮箱格式)
- `name`: 姓名 (必需，2-100字符)
- `role`: 角色 (必需，可选值：teacher, researcher, student, guest)

**成功响应** (用户创建成功 - HTTP 201)：
```json
{
  "success": true,
  "message": "用户创建成功，可通过GitLab OAuth登录GitLabEx",
  "data": {
    "id": 12345,
    "username": "john_doe",
    "email": "john.doe@example.com",
    "name": "John Doe",
    "role": "学生"
  }
}
```

**成功响应** (用户已存在 - HTTP 200)：
```json
{
  "success": true,
  "message": "用户已存在，可通过GitLab OAuth登录GitLabEx",
  "data": {
    "id": 12345,
    "username": "john_doe",
    "email": "john.doe@example.com",
    "name": "John Doe",
    "role": "学生"
  }
}
```

### 2. 获取用户信息

**接口**：`GET /users/{username}`

**描述**：根据用户名获取GitLab用户信息

**路径参数**：
- `username`: 用户名

**成功响应**：
```json
{
  "success": true,
  "message": "获取用户信息成功",
  "data": {
    "id": 12345,
    "username": "john_doe",
    "email": "john.doe@example.com",
    "name": "John Doe",
    "role": "用户"
  }
}
```

## 角色说明

| 角色值 | 中文名称 | GitLab权限 | 说明 |
|-------|----------|-----------|------|
| teacher | 教师 | 普通用户 | 教学人员 |
| researcher | 研究员 | 普通用户 | 研究人员 |
| student | 学生 | 普通用户 | 学生用户 |
| guest | 访客 | 普通用户 | 访客用户 |

**注意**：出于安全考虑，第三方API密钥无法创建管理员用户。

## 用户登录流程

### 标准OAuth登录流程

1. **第三方系统创建用户**：通过API在GitLab中创建用户账号
2. **用户访问GitLabEx**：用户访问GitLabEx登录页面
3. **选择GitLab OAuth登录**：点击"通过GitLab登录"按钮
4. **GitLab认证**：在GitLab登录页面输入用户名和密码
5. **授权确认**：确认授权GitLabEx访问GitLab账号
6. **登录成功**：自动跳转到GitLabEx主页，完成登录

这种方式是**最安全**的，因为：
- 密码不会通过API传输
- 遵循OAuth 2.0标准
- 用户直接与GitLab交互
- 支持GitLab的所有安全特性（2FA等）

## 错误处理

### 常见错误码

| HTTP状态码 | 错误类型 | 说明 |
|-----------|----------|------|
| 400 | Bad Request | 请求参数错误 |
| 401 | Unauthorized | API密钥无效或缺失 |
| 404 | Not Found | 用户不存在 |
| 500 | Internal Server Error | 服务器内部错误 |

### 错误响应格式

```json
{
  "success": false,
  "message": "错误描述",
  "error": "详细错误信息"
}
```

## 集成示例

### Python示例

```python
import requests

class GitLabExIntegration:
    def __init__(self, base_url, api_key):
        self.base_url = base_url
        self.headers = {
            'Content-Type': 'application/json',
            'X-API-Key': api_key
        }
    
    def create_user(self, username, password, email, name, role):
        """创建用户"""
        url = f"{self.base_url}/api/v1/sync/users"
        data = {
            "username": username,
            "password": password,
            "email": email,
            "name": name,
            "role": role
        }
        
        response = requests.post(url, headers=self.headers, json=data)
        return response.json()
    
    def get_user(self, username):
        """获取用户信息"""
        url = f"{self.base_url}/api/v1/sync/users/{username}"
        response = requests.get(url, headers=self.headers)
        return response.json()

# 使用示例
integration = GitLabExIntegration(
    base_url="https://gitlabex.example.com",
    api_key="your_third_party_api_key_here"
)

# 创建用户
result = integration.create_user(
    username="john_doe",
    password="secure_password123",
    email="john.doe@example.com",
    name="John Doe",
    role="student"
)

if result['success']:
    print(f"用户操作成功: {result['data']['username']}")
    print("用户现在可以通过GitLab OAuth登录GitLabEx")
else:
    print(f"操作失败: {result['message']}")

# 获取用户信息
user_info = integration.get_user("john_doe")
if user_info['success']:
    print(f"用户信息: {user_info['data']}")
```

### Node.js示例

```javascript
const axios = require('axios');

class GitLabExIntegration {
    constructor(baseUrl, apiKey) {
        this.baseUrl = baseUrl;
        this.headers = {
            'Content-Type': 'application/json',
            'X-API-Key': apiKey
        };
    }

    async createUser(username, password, email, name, role) {
        try {
            const response = await axios.post(
                `${this.baseUrl}/api/v1/sync/users`,
                { username, password, email, name, role },
                { headers: this.headers }
            );
            return response.data;
        } catch (error) {
            return error.response.data;
        }
    }

    async getUser(username) {
        try {
            const response = await axios.get(
                `${this.baseUrl}/api/v1/sync/users/${username}`,
                { headers: this.headers }
            );
            return response.data;
        } catch (error) {
            return error.response.data;
        }
    }
}

// 使用示例
const integration = new GitLabExIntegration(
    'https://gitlabex.example.com',
    'your_third_party_api_key_here'
);

// 创建用户
integration.createUser(
    'john_doe',
    'secure_password123',
    'john.doe@example.com',
    'John Doe',
    'student'
).then(result => {
    if (result.success) {
        console.log(`用户操作成功: ${result.data.username}`);
        console.log('用户现在可以通过GitLab OAuth登录GitLabEx');
    } else {
        console.log(`操作失败: ${result.message}`);
    }
});

// 获取用户信息
integration.getUser('john_doe').then(result => {
    if (result.success) {
        console.log('用户信息:', result.data);
    }
});
```

### cURL示例

```bash
# 创建用户
curl -X POST "https://gitlabex.example.com/api/v1/sync/users" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your_third_party_api_key_here" \
  -d '{
    "username": "john_doe",
    "password": "secure_password123",
    "email": "john.doe@example.com",
    "name": "John Doe",
    "role": "student"
  }'

# 获取用户信息
curl -X GET "https://gitlabex.example.com/api/v1/sync/users/john_doe" \
  -H "X-API-Key: your_third_party_api_key_here"
```

## 最佳实践

### 1. 安全建议

- **API密钥管理**：将API密钥存储在安全的配置文件中，不要硬编码
- **HTTPS传输**：生产环境必须使用HTTPS
- **密码强度**：确保创建用户时使用强密码
- **定期轮换**：定期更换API密钥

### 2. 用户管理

- **用户名唯一性**：确保用户名在GitLab中唯一
- **邮箱有效性**：使用有效的邮箱地址
- **角色选择**：根据用户实际职能选择合适的角色
- **重复检查**：API会自动处理重复用户创建

### 3. 错误处理

- **重试机制**：对临时性错误实现适当的重试
- **日志记录**：记录所有API调用和错误信息
- **用户反馈**：为用户提供清晰的错误信息

### 4. 性能优化

- **请求频率**：遵守API速率限制（每小时500次）
- **批量操作**：对于大量用户，分批创建避免超时
- **缓存用户信息**：适当缓存用户信息，减少API调用

## 部署配置

### 环境变量设置

```bash
# 设置第三方API密钥
export THIRD_PARTY_API_KEY="your_secure_third_party_api_key_32_chars_minimum"

# 设置GitLab系统Token（用于创建用户）
export GITLAB_SYSTEM_TOKEN="your_gitlab_admin_token"
```

### Docker配置

```yaml
services:
  backend:
    environment:
      - THIRD_PARTY_API_KEY=${THIRD_PARTY_API_KEY}
      - GITLAB_SYSTEM_TOKEN=${GITLAB_SYSTEM_TOKEN}
```

## 故障排除

### 1. 用户创建问题

**问题**：用户创建返回500错误
**解决**：检查GitLab系统Token是否正确配置

**问题**：API密钥认证失败
**解决**：检查THIRD_PARTY_API_KEY环境变量是否正确设置

### 2. OAuth登录问题

**问题**：用户无法通过OAuth登录
**解决**：确保用户在GitLab中创建成功，并且密码正确

**问题**：OAuth回调失败
**解决**：检查GitLab OAuth应用配置和回调URL设置

### 3. 用户信息查询问题

**问题**：获取用户信息返回404
**解决**：确认用户名正确，且用户在GitLab中存在

## API变更日志

### v1.0.0 (当前版本)
- 提供2个核心API接口
- 支持用户创建和信息查询
- 使用标准GitLab OAuth登录流程
- 简化的单密钥认证机制

## 支持和联系

如有问题或需要技术支持，请联系：
- 技术支持邮箱：support@gitlabex.com
- 文档更新：https://docs.gitlabex.com
- 项目地址：https://github.com/gitlabex/gitlabex

---

*本文档最后更新时间：2024年1月15日*