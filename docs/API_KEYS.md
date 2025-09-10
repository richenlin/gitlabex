# GitLabEx API密钥说明文档

## 📋 API密钥概览

GitLabEx系统支持多种API密钥，每种密钥都有不同的用途和权限范围。

## 🔐 密钥类型详解

### 1. SYNC_API_KEY - 系统同步密钥

**用途**: 内部系统或可信第三方系统的数据同步

**适用场景**:
- 🏫 **学校教务系统**: 官方教务管理系统同步师生数据
- 💼 **HR管理系统**: 内部人力资源系统同步员工信息
- 🔄 **数据迁移工具**: 从旧系统批量迁移用户数据
- 🛠️ **管理员工具**: 系统管理员使用的批量操作工具

**权限范围**:
- ✅ 创建用户账号
- ✅ 批量创建用户（最高权限，支持100个/批次）
- ✅ 更新用户信息
- ✅ 设置用户角色（包括管理员角色）
- ✅ 激活/停用用户账号
- ✅ 访问完整的用户管理API

**安全级别**: 🔴 **最高级别**
- 应该只分配给最可信的系统
- 建议定期轮换（每3-6个月）
- 需要严格的网络访问控制

```bash
# 配置示例
SYNC_API_KEY=gitlabex_sync_2024_university_system_v1_secure_key_change_in_prod
```

### 2. THIRD_PARTY_API_KEY - 第三方接入密钥

**用途**: 外部第三方系统或应用的受限访问

**适用场景**:
- 📱 **移动应用**: 校园APP的用户注册功能
- 🌐 **外部LMS**: 第三方学习管理系统
- 🤝 **合作伙伴系统**: 合作机构的用户接入
- 📊 **分析工具**: 第三方数据分析平台

**权限范围**:
- ✅ 创建学生和教师账号
- ✅ 小批量创建用户（限制20个/批次）
- ✅ 更新基本用户信息
- ❌ 无法创建管理员账号
- ❌ 无法批量删除用户
- ❌ 无法访问敏感管理功能

**安全级别**: 🟡 **中等级别**
- 可以分配给可信的第三方
- 建议每月轮换
- 具有使用频率限制

```bash
# 配置示例  
THIRD_PARTY_API_KEY=gitlabex_3rd_party_2024_partner_system_limited_access
```

## 🔧 配置和使用

### 环境变量配置

```bash
# config/backend.env

# 系统同步密钥（最高权限）
SYNC_API_KEY=gitlabex_sync_api_key_2024_secure_change_in_production

# 第三方接入密钥（受限权限）
THIRD_PARTY_API_KEY=gitlabex_third_party_api_key_2024

# 可以配置多个第三方密钥（可选）
PARTNER_A_API_KEY=gitlabex_partner_a_key_2024
PARTNER_B_API_KEY=gitlabex_partner_b_key_2024
```

### 使用示例

#### 使用SYNC_API_KEY（教务系统）
```python
# 教务系统批量导入学生
import requests

headers = {
    "X-API-Key": "gitlabex_sync_api_key_2024_secure",
    "Content-Type": "application/json"
}

# 批量创建100个学生（仅SYNC_API_KEY支持）
students = [
    {
        "username": f"student{i:04d}",
        "password": "temp_password_123",
        "email": f"student{i:04d}@university.edu",
        "name": f"学生{i:04d}",
        "role": "student",
        "student_id": f"2024{i:04d}"
    }
    for i in range(1, 101)  # 100个学生
]

response = requests.post(
    "http://localhost:8080/api/v1/sync/users/batch",
    headers=headers,
    json={"users": students}
)
```

#### 使用THIRD_PARTY_API_KEY（移动应用）
```javascript
// 移动应用用户注册
const registerUser = async (userData) => {
    const headers = {
        'X-API-Key': 'gitlabex_third_party_api_key_2024',
        'Content-Type': 'application/json'
    };
    
    // 只能创建学生/教师，不能创建管理员
    if (userData.role === 'admin') {
        throw new Error('第三方应用无权创建管理员账号');
    }
    
    const response = await fetch('http://localhost:8080/api/v1/sync/users', {
        method: 'POST',
        headers: headers,
        body: JSON.stringify(userData)
    });
    
    return response.json();
};
```

## 🛡️ 权限对比表

| 功能 | SYNC_API_KEY | THIRD_PARTY_API_KEY | 说明 |
|------|--------------|---------------------|------|
| 创建普通用户 | ✅ 无限制 | ✅ 有限制 | 第三方限制频率 |
| 批量创建用户 | ✅ 100个/批次 | ✅ 20个/批次 | 不同的批次限制 |
| 创建管理员 | ✅ 允许 | ❌ 禁止 | 管理员角色受限 |
| 更新用户角色 | ✅ 所有角色 | ✅ 非管理员角色 | 角色提升限制 |
| 删除用户 | ✅ 允许 | ❌ 禁止 | 删除操作受限 |
| 激活/停用用户 | ✅ 允许 | ✅ 允许 | 基本状态管理 |
| 访问敏感信息 | ✅ 完整访问 | ❌ 受限访问 | 数据访问权限 |
| 频率限制 | 🟢 宽松 | 🟡 严格 | 不同的限制策略 |

## 🔄 密钥轮换策略

### 自动轮换（推荐）
```bash
#!/bin/bash
# scripts/rotate-api-keys.sh

# 生成新的API密钥
NEW_SYNC_KEY="gitlabex_sync_$(date +%Y%m%d)_$(openssl rand -hex 16)"
NEW_THIRD_PARTY_KEY="gitlabex_3rd_$(date +%Y%m%d)_$(openssl rand -hex 12)"

# 更新配置文件
sed -i "s/SYNC_API_KEY=.*/SYNC_API_KEY=$NEW_SYNC_KEY/" config/backend.env
sed -i "s/THIRD_PARTY_API_KEY=.*/THIRD_PARTY_API_KEY=$NEW_THIRD_PARTY_KEY/" config/backend.env

# 重启服务应用新密钥
docker-compose -f docker-compose.dev.yml restart backend

echo "API密钥已更新："
echo "SYNC_API_KEY: $NEW_SYNC_KEY"
echo "THIRD_PARTY_API_KEY: $NEW_THIRD_PARTY_KEY"
```

### 手动轮换
```bash
# 1. 生成新密钥
openssl rand -hex 32

# 2. 更新配置文件
vim config/backend.env

# 3. 重启服务
docker-compose restart backend

# 4. 通知相关系统更新密钥
```

## 📊 使用监控

### 密钥使用统计
```bash
# 查看API调用日志
docker-compose logs backend | grep "API call"

# 按密钥统计使用次数
docker-compose logs backend | grep "X-API-Key" | cut -d':' -f3 | sort | uniq -c
```

### 安全告警
- 🚨 **异常调用频率**: 超过正常阈值的API调用
- 🚨 **无效密钥尝试**: 多次使用无效密钥的请求
- 🚨 **权限越界**: 尝试执行超出权限的操作

## 🎯 最佳实践

### 1. 密钥管理
- ✅ **定期轮换**: SYNC_API_KEY每3个月，THIRD_PARTY_API_KEY每月
- ✅ **安全存储**: 使用环境变量或密钥管理服务
- ✅ **最小权限**: 根据实际需要分配最小必要权限
- ✅ **监控日志**: 定期检查API使用日志

### 2. 网络安全
- ✅ **IP白名单**: 限制API密钥的来源IP
- ✅ **HTTPS传输**: 强制使用HTTPS协议
- ✅ **防火墙**: 配置防火墙规则限制访问
- ✅ **速率限制**: 实施API调用频率限制

### 3. 应急处理
```bash
# 紧急撤销密钥（从配置中移除）
sed -i 's/COMPROMISED_KEY//' config/backend.env
docker-compose restart backend

# 查看可疑活动
docker-compose logs backend | grep "COMPROMISED_KEY" | tail -50
```

## 📞 技术支持

如需技术支持或有安全相关问题：
- 🔐 **安全问题**: security@gitlabex.com
- 📧 **技术支持**: support@gitlabex.com
- 📖 **文档中心**: https://docs.gitlabex.com
