# GitLabEx 第三方系统集成指南

本文档详细介绍如何将第三方系统与GitLabEx进行集成，实现用户账号同步和统一身份管理。

## 概述

GitLabEx提供了完整的第三方系统集成API，支持：
- **用户账号同步**：将外部系统用户同步到GitLab
- **双密钥认证**：支持不同权限级别的API访问
- **批量操作**：支持批量用户创建和更新
- **外部系统标识**：维护外部系统与GitLab用户的映射关系
- **完整的审计日志**：记录所有同步操作

## API认证机制

### 双密钥设计

GitLabEx采用双密钥设计，支持不同权限级别的API访问：

#### 1. 系统同步密钥 (SYNC_API_KEY)
- **用途**：内部系统使用，拥有最高权限
- **权限**：
  - 可以创建管理员用户
  - 批量操作最多100个用户
  - 可以访问所有外部系统来源
  - 每小时1000次请求限制
  - 可以更新和查询用户信息

#### 2. 第三方密钥 (THIRD_PARTY_API_KEY)
- **用途**：外部系统使用，权限受限
- **权限**：
  - 无法创建管理员用户
  - 批量操作最多20个用户
  - 只能访问指定的外部系统来源
  - 每小时300次请求限制
  - 可以更新和查询用户信息

### 认证方式

所有API请求都需要在请求头中包含API密钥：

```http
X-API-Key: your_api_key_here
```

## API接口

### 基础URL

```
https://your-gitlabex-domain.com/api/v1/sync
```

### 1. 创建用户

**接口**：`POST /users`

**描述**：创建单个用户并返回JWT登录Token

**请求体**：
```json
{
  "username": "john_doe",
  "password": "secure_password",
  "email": "john.doe@example.com",
  "name": "John Doe",
  "role": "student",
  "external_id": "ext_123456",
  "external_source": "student_system",
  "department": "计算机科学系",
  "student_id": "2024001001",
  "phone": "13800138000"
}
```

**响应**：
```json
{
  "success": true,
  "message": "用户创建成功",
  "data": {
    "user": {
      "id": 12345,
      "username": "john_doe",
      "email": "john.doe@example.com",
      "name": "John Doe",
      "role": "学生",
      "avatar_url": "",
      "is_active": true,
      "created_at": "2024-01-15T10:30:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "external_user": {
      "id": 1,
      "external_id": "ext_123456",
      "external_source": "student_system",
      "department": "计算机科学系",
      "student_id": "2024001001",
      "phone": "13800138000"
    }
  }
}
```

### 2. 批量创建用户

**接口**：`POST /users/batch`

**描述**：批量创建多个用户

**请求体**：
```json
{
  "users": [
    {
      "username": "student1",
      "password": "password123",
      "email": "student1@example.com",
      "name": "Student One",
      "role": "student",
      "external_id": "ext_001",
      "external_source": "student_system"
    },
    {
      "username": "teacher1",
      "password": "password456",
      "email": "teacher1@example.com",
      "name": "Teacher One",
      "role": "teacher",
      "external_id": "ext_002",
      "external_source": "teacher_system"
    }
  ]
}
```

**响应**：
```json
{
  "success": true,
  "message": "批量创建完成",
  "total_count": 2,
  "success_count": 2,
  "failure_count": 0,
  "results": [
    {
      "success": true,
      "message": "用户创建成功",
      "data": {
        "user": { /* 用户信息 */ },
        "token": "jwt_token_here"
      }
    },
    {
      "success": true,
      "message": "用户创建成功",
      "data": {
        "user": { /* 用户信息 */ },
        "token": "jwt_token_here"
      }
    }
  ]
}
```

### 3. 更新用户

**接口**：`PUT /users/{external_id}?external_source={source}`

**描述**：根据外部系统ID更新用户信息

**请求体**：
```json
{
  "name": "Updated Name",
  "email": "updated@example.com",
  "role": "researcher",
  "department": "新部门",
  "phone": "13900139000"
}
```

**响应**：
```json
{
  "success": true,
  "message": "用户更新成功",
  "data": {
    "user": {
      "id": 12345,
      "username": "john_doe",
      "email": "updated@example.com",
      "name": "Updated Name",
      "role": "研究员",
      "avatar_url": "",
      "is_active": true
    }
  }
}
```

### 4. 查询用户

**接口**：`GET /users/{external_id}?external_source={source}`

**描述**：根据外部系统ID查询用户信息

**响应**：
```json
{
  "success": true,
  "message": "获取用户信息成功",
  "data": {
    "user": {
      "id": 12345,
      "username": "john_doe",
      "email": "john.doe@example.com",
      "name": "John Doe",
      "role": "学生",
      "avatar_url": "",
      "is_active": true,
      "created_at": "2024-01-15T10:30:00Z"
    },
    "external_user": {
      "id": 1,
      "external_id": "ext_123456",
      "external_source": "student_system",
      "department": "计算机科学系",
      "student_id": "2024001001",
      "phone": "13800138000"
    }
  }
}
```

### 5. 管理接口（需要系统API密钥）

#### 获取统计信息

**接口**：`GET /admin/stats`

**响应**：
```json
{
  "success": true,
  "message": "获取统计信息成功",
  "data": [
    {
      "source": "student_system",
      "total_users": 1500,
      "active_users": 1450,
      "last_sync_at": "2024-01-15T15:30:00Z"
    },
    {
      "source": "teacher_system",
      "total_users": 200,
      "active_users": 195,
      "last_sync_at": "2024-01-15T14:20:00Z"
    }
  ]
}
```

#### 获取同步日志

**接口**：`GET /admin/logs?page=1&page_size=20&external_source=student_system`

**响应**：
```json
{
  "success": true,
  "message": "获取同步日志成功",
  "data": [
    {
      "id": 1,
      "external_user_id": 1,
      "operation": "create",
      "status": "success",
      "api_key": "third_party",
      "client_ip": "192.168.1.100",
      "created_at": "2024-01-15T10:30:00Z",
      "external_user": {
        "external_id": "ext_123456",
        "external_source": "student_system",
        "username": "john_doe"
      }
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 1,
    "total_pages": 1
  }
}
```

## 角色映射

GitLabEx支持将外部系统角色自动映射到GitLab角色：

| 外部系统角色 | GitLab角色 | 权限级别 | 说明 |
|-------------|------------|----------|------|
| admin, administrator, 系统管理员, 管理员 | Owner | 50 | 最高权限，可管理所有资源 |
| teacher, instructor, professor, 教师, 老师, 教授 | Maintainer | 40 | 可管理课题和用户 |
| researcher, research_assistant, 研究员, 研究助理 | Developer | 30 | 可参与开发和研究 |
| student, undergraduate, graduate, 学生, 本科生, 研究生 | Reporter | 20 | 可查看和提交作业 |
| guest, visitor, 游客, 访客 | Guest | 10 | 只读权限 |

## 错误处理

### 常见错误码

| HTTP状态码 | 错误类型 | 说明 |
|-----------|----------|------|
| 400 | Bad Request | 请求参数错误 |
| 401 | Unauthorized | API密钥无效或缺失 |
| 403 | Forbidden | 权限不足 |
| 409 | Conflict | 用户已存在 |
| 429 | Too Many Requests | 请求频率超限 |
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
import json

class GitLabExIntegration:
    def __init__(self, base_url, api_key):
        self.base_url = base_url
        self.headers = {
            'Content-Type': 'application/json',
            'X-API-Key': api_key
        }
    
    def create_user(self, user_data):
        """创建单个用户"""
        url = f"{self.base_url}/api/v1/sync/users"
        response = requests.post(url, headers=self.headers, json=user_data)
        return response.json()
    
    def batch_create_users(self, users_data):
        """批量创建用户"""
        url = f"{self.base_url}/api/v1/sync/users/batch"
        payload = {"users": users_data}
        response = requests.post(url, headers=self.headers, json=payload)
        return response.json()
    
    def update_user(self, external_id, external_source, update_data):
        """更新用户信息"""
        url = f"{self.base_url}/api/v1/sync/users/{external_id}"
        params = {"external_source": external_source}
        response = requests.put(url, headers=self.headers, params=params, json=update_data)
        return response.json()
    
    def get_user(self, external_id, external_source):
        """获取用户信息"""
        url = f"{self.base_url}/api/v1/sync/users/{external_id}"
        params = {"external_source": external_source}
        response = requests.get(url, headers=self.headers, params=params)
        return response.json()

# 使用示例
integration = GitLabExIntegration(
    base_url="https://gitlabex.example.com",
    api_key="your_third_party_api_key_here"
)

# 创建用户
user_data = {
    "username": "john_doe",
    "password": "secure_password",
    "email": "john.doe@example.com",
    "name": "John Doe",
    "role": "student",
    "external_id": "stu_123456",
    "external_source": "student_management_system",
    "department": "计算机科学系",
    "student_id": "2024001001"
}

result = integration.create_user(user_data)
if result['success']:
    print(f"用户创建成功，JWT Token: {result['data']['token']}")
else:
    print(f"用户创建失败: {result['message']}")
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

    async createUser(userData) {
        try {
            const response = await axios.post(
                `${this.baseUrl}/api/v1/sync/users`,
                userData,
                { headers: this.headers }
            );
            return response.data;
        } catch (error) {
            return error.response.data;
        }
    }

    async batchCreateUsers(usersData) {
        try {
            const response = await axios.post(
                `${this.baseUrl}/api/v1/sync/users/batch`,
                { users: usersData },
                { headers: this.headers }
            );
            return response.data;
        } catch (error) {
            return error.response.data;
        }
    }

    async updateUser(externalId, externalSource, updateData) {
        try {
            const response = await axios.put(
                `${this.baseUrl}/api/v1/sync/users/${externalId}`,
                updateData,
                { 
                    headers: this.headers,
                    params: { external_source: externalSource }
                }
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

// 批量创建用户
const users = [
    {
        username: 'student1',
        password: 'password123',
        email: 'student1@example.com',
        name: 'Student One',
        role: 'student',
        external_id: 'stu_001',
        external_source: 'student_system'
    },
    {
        username: 'teacher1',
        password: 'password456',
        email: 'teacher1@example.com',
        name: 'Teacher One',
        role: 'teacher',
        external_id: 'tea_001',
        external_source: 'teacher_system'
    }
];

integration.batchCreateUsers(users)
    .then(result => {
        console.log('批量创建结果:', result);
        console.log(`成功: ${result.success_count}, 失败: ${result.failure_count}`);
    })
    .catch(error => {
        console.error('批量创建失败:', error);
    });
```

## 最佳实践

### 1. 安全建议

- **API密钥管理**：将API密钥存储在安全的配置文件中，不要硬编码
- **HTTPS传输**：生产环境必须使用HTTPS
- **密钥轮换**：定期更换API密钥
- **权限最小化**：根据实际需要选择合适的API密钥类型

### 2. 性能优化

- **批量操作**：尽量使用批量接口减少API调用次数
- **请求限制**：注意API请求频率限制，实现适当的重试机制
- **缓存策略**：对不经常变化的数据实现本地缓存

### 3. 错误处理

- **重试机制**：对临时性错误实现指数退避重试
- **日志记录**：记录所有API调用和错误信息
- **监控告警**：监控同步失败率，设置告警阈值

### 4. 数据一致性

- **幂等性**：确保重复调用不会产生副作用
- **状态同步**：定期检查和同步用户状态
- **冲突解决**：制定数据冲突的解决策略

## 故障排除

### 1. 常见问题

#### API密钥无效
```json
{
  "success": false,
  "message": "无效的API密钥",
  "error": "提供的API密钥无效或已过期"
}
```
**解决方案**：检查API密钥是否正确，确认是否在环境变量中正确设置。

#### 权限不足
```json
{
  "success": false,
  "message": "权限不足",
  "error": "当前API密钥无权创建管理员用户"
}
```
**解决方案**：使用系统API密钥或调整用户角色。

#### 批量限制超出
```json
{
  "success": false,
  "message": "批量创建数量超出限制，当前API密钥最多支持 20 个用户"
}
```
**解决方案**：减少批量操作的用户数量或使用系统API密钥。

### 2. 调试建议

1. **检查请求格式**：确认请求头和请求体格式正确
2. **验证网络连接**：确认能够访问GitLabEx服务
3. **查看日志**：通过管理接口查看同步日志
4. **测试环境**：先在测试环境验证集成逻辑

## 支持和联系

如有集成问题或需要技术支持，请联系：
- 技术支持邮箱：support@gitlabex.com
- 文档更新：https://docs.gitlabex.com
- GitHub Issues：https://github.com/gitlabex/gitlabex/issues

---

*本文档最后更新时间：2024年1月15日*
