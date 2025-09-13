# GitLabEx 第三方系统同步API文档

GitLabEx提供了一套完整的RESTful API，用于第三方系统（如学校教务系统、LMS平台等）的用户同步和管理。

## 🔐 认证方式

所有同步API都需要通过API密钥进行认证：

```http
X-API-Key: your_api_key_here
```

### 获取API密钥

API密钥在后端配置文件中设置：

```bash
# config/backend.env
SYNC_API_KEY=your_secure_api_key_here
**用途**: 内部系统或可信第三方系统的数据同步
**权限范围**:
- ✅ 创建用户账号
- ✅ 批量创建用户（最高权限，支持100个/批次）
- ✅ 更新用户信息
- ✅ 设置用户角色（包括管理员角色）
- ✅ 激活/停用用户账号
- ✅ 访问完整的用户管理API


THIRD_PARTY_API_KEY=your_third_party_key_here
**用途**: 外部第三方系统或应用的受限访问
**权限范围**:
- ✅ 创建学生和教师账号
- ✅ 小批量创建用户（限制20个/批次）
- ✅ 更新基本用户信息
- ❌ 无法创建管理员账号
- ❌ 无法批量删除用户
- ❌ 无法访问敏感管理功能
```

## 📋 API 概览

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 创建用户 | POST | `/api/v1/sync/users` | 创建单个用户并返回Token |
| 批量创建用户 | POST | `/api/v1/sync/users/batch` | 批量创建用户 |
| 更新用户 | PUT | `/api/v1/sync/users/{id}` | 更新用户信息 |
| 获取用户 | GET | `/api/v1/sync/users/{id}` | 获取用户信息 |

## 🚀 API 详细说明

### 1. 创建用户

**POST** `/api/v1/sync/users`

创建一个新用户并返回登录Token。

#### 请求头
```http
Content-Type: application/json
X-API-Key: your_api_key_here
```

#### 请求体
```json
{
  "username": "zhangsan",
  "password": "password123",
  "email": "zhangsan@university.edu",
  "name": "张三",
  "role": "student",
  "avatar_url": "https://example.com/avatar.jpg",
  "department": "计算机科学与技术",
  "student_id": "2024001001",
  "external_id": "ext_12345",
  "external_source": "university_system"
}
```

#### 字段说明
- `username` (必填): 用户名，3-50字符
- `password` (必填): 密码，最少6位
- `email` (必填): 邮箱地址
- `name` (必填): 真实姓名，2-100字符
- `role` (必填): 角色类型，可选值：
  - `admin`: 管理员
  - `teacher`: 教师
  - `assistant`: 助教
  - `student`: 学生
- `avatar_url` (可选): 头像URL
- `department` (可选): 院系/部门
- `student_id` (可选): 学号
- `teacher_id` (可选): 教师工号
- `phone` (可选): 电话号码
- `external_id` (可选): 外部系统ID
- `external_source` (可选): 外部系统来源

#### 成功响应 (201)
```json
{
  "success": true,
  "message": "用户创建成功",
  "data": {
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "username": "zhangsan",
      "email": "zhangsan@university.edu",
      "name": "张三",
      "role": "student",
      "avatar_url": "https://example.com/avatar.jpg",
      "is_active": true,
      "created_at": "2024-01-01T12:00:00Z",
      "external_id": "ext_12345"
    },
    "token": "jwt.123e4567-e89b-12d3-a456-426614174000.zhangsan"
  }
}
```

#### 错误响应
```json
{
  "success": false,
  "message": "用户已存在",
  "error": "用户名已被使用"
}
```

### 2. 批量创建用户

**POST** `/api/v1/sync/users/batch`

批量创建多个用户，最多支持100个用户。

#### 请求体
```json
{
  "users": [
    {
      "username": "student1",
      "password": "password123",
      "email": "student1@university.edu",
      "name": "学生一",
      "role": "student"
    },
    {
      "username": "teacher1",
      "password": "password456",
      "email": "teacher1@university.edu",
      "name": "教师一",
      "role": "teacher"
    }
  ]
}
```

#### 成功响应 (200)
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
        "token": "jwt.token.here"
      }
    },
    {
      "success": true,
      "message": "用户创建成功",
      "data": {
        "user": { /* 用户信息 */ },
        "token": "jwt.token.here"
      }
    }
  ]
}
```

### 3. 更新用户

**PUT** `/api/v1/sync/users/{id}`

更新用户信息。`{id}` 可以是用户ID或用户名。

#### 请求体
```json
{
  "email": "newemail@university.edu",
  "name": "新姓名",
  "role": "teacher",
  "is_active": true,
  "avatar_url": "https://example.com/new-avatar.jpg"
}
```

#### 成功响应 (200)
```json
{
  "success": true,
  "message": "用户更新成功",
  "data": {
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "username": "zhangsan",
      "email": "newemail@university.edu",
      "name": "新姓名",
      "role": "teacher",
      "avatar_url": "https://example.com/new-avatar.jpg",
      "is_active": true,
      "created_at": "2024-01-01T12:00:00Z"
    }
  }
}
```

## 📱 使用示例

### Python示例

```python
import requests
import json

# 配置
BASE_URL = "http://localhost:8080"
API_KEY = "your_api_key_here"

headers = {
    "Content-Type": "application/json",
    "X-API-Key": API_KEY
}

# 创建用户
def create_user(user_data):
    url = f"{BASE_URL}/api/v1/sync/users"
    response = requests.post(url, headers=headers, json=user_data)
    return response.json()

# 批量创建用户
def batch_create_users(users_list):
    url = f"{BASE_URL}/api/v1/sync/users/batch"
    data = {"users": users_list}
    response = requests.post(url, headers=headers, json=data)
    return response.json()

# 使用示例
if __name__ == "__main__":
    # 创建单个用户
    user = {
        "username": "student001",
        "password": "securePassword123",
        "email": "student001@university.edu",
        "name": "学生001",
        "role": "student",
        "student_id": "2024001001"
    }
    
    result = create_user(user)
    if result["success"]:
        print(f"用户创建成功，Token: {result['data']['token']}")
    else:
        print(f"用户创建失败: {result['message']}")
    
    # 批量创建用户
    users = [
        {
            "username": f"student{i:03d}",
            "password": "password123",
            "email": f"student{i:03d}@university.edu",
            "name": f"学生{i:03d}",
            "role": "student"
        }
        for i in range(1, 6)  # 创建5个用户
    ]
    
    batch_result = batch_create_users(users)
    print(f"批量创建结果: 成功 {batch_result['success_count']}, 失败 {batch_result['failure_count']}")
```

### JavaScript/Node.js示例

```javascript
const axios = require('axios');

const baseURL = 'http://localhost:8080';
const apiKey = 'your_api_key_here';

const headers = {
    'Content-Type': 'application/json',
    'X-API-Key': apiKey
};

// 创建用户
async function createUser(userData) {
    try {
        const response = await axios.post(`${baseURL}/api/v1/sync/users`, userData, { headers });
        return response.data;
    } catch (error) {
        console.error('创建用户失败:', error.response?.data || error.message);
        return null;
    }
}

// 更新用户
async function updateUser(userId, updateData) {
    try {
        const response = await axios.put(`${baseURL}/api/v1/sync/users/${userId}`, updateData, { headers });
        return response.data;
    } catch (error) {
        console.error('更新用户失败:', error.response?.data || error.message);
        return null;
    }
}

// 使用示例
(async () => {
    // 创建教师
    const teacher = {
        username: 'prof_wang',
        password: 'securePassword456',
        email: 'wang.prof@university.edu',
        name: '王教授',
        role: 'teacher',
        department: '计算机科学与技术学院'
    };
    
    const result = await createUser(teacher);
    if (result && result.success) {
        console.log('教师创建成功:', result.data.user.name);
        console.log('登录Token:', result.data.token);
        
        // 更新用户信息
        const updateResult = await updateUser(result.data.user.id, {
            department: '软件工程学院',
            avatar_url: 'https://example.com/prof-avatar.jpg'
        });
        
        if (updateResult && updateResult.success) {
            console.log('用户信息更新成功');
        }
    }
})();
```

### cURL示例

```bash
# 创建用户
curl -X POST \
  http://localhost:8080/api/v1/sync/users \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your_api_key_here" \
  -d '{
    "username": "testuser",
    "password": "password123",
    "email": "testuser@university.edu",
    "name": "测试用户",
    "role": "student"
  }'

# 批量创建用户
curl -X POST \
  http://localhost:8080/api/v1/sync/users/batch \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your_api_key_here" \
  -d '{
    "users": [
      {
        "username": "batch1",
        "password": "password123",
        "email": "batch1@university.edu",
        "name": "批量用户1",
        "role": "student"
      },
      {
        "username": "batch2",
        "password": "password123",
        "email": "batch2@university.edu",
        "name": "批量用户2",
        "role": "student"
      }
    ]
  }'

# 更新用户
curl -X PUT \
  http://localhost:8080/api/v1/sync/users/testuser \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your_api_key_here" \
  -d '{
    "name": "新的测试用户名",
    "role": "assistant"
  }'
```

## 🔒 安全建议

1. **API密钥管理**
   - 使用强随机生成的API密钥
   - 定期轮换API密钥
   - 不要在代码中硬编码密钥

2. **网络安全**
   - 使用HTTPS传输
   - 配置防火墙限制访问
   - 考虑IP白名单

3. **数据验证**
   - 验证输入数据格式
   - 检查邮箱和用户名唯一性
   - 使用强密码策略

## 📊 错误代码参考

| HTTP状态码 | 错误类型 | 说明 |
|------------|----------|------|
| 400 | 请求参数错误 | 缺少必需字段或格式不正确 |
| 401 | 认证失败 | API密钥无效或缺失 |
| 409 | 资源冲突 | 用户名或邮箱已存在 |
| 500 | 服务器错误 | 内部服务器错误 |

## 🔄 集成建议

1. **批量同步**
   - 使用批量接口提高效率
   - 实现断点续传机制
   - 记录同步日志

2. **数据一致性**
   - 定期全量同步校验
   - 实现增量同步机制
   - 处理数据冲突策略

3. **错误处理**
   - 实现重试机制
   - 记录失败案例
   - 提供数据回滚功能

## 📞 技术支持

如需技术支持或有疑问，请联系：
- 📧 Email: support@gitlabex.com
- 📱 电话: +86-xxx-xxxx-xxxx
- 📖 文档: [GitLabEx 开发者文档](https://docs.gitlabex.com)
