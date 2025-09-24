# GitLabEx 测试数据初始化指南

## 概述

GitLabEx 提供了两种测试数据初始化方式：

1. **传统SQL方式** (`scripts/init-test-data.sh`) - 直接插入SQL数据，快速但不包含GitLab集成
2. **API驱动方式** (`scripts/init-test-data-api.sh`) - 通过API创建数据，完整的GitLab集成 ⭐**推荐**

## API驱动方式 (推荐)

### 特点

- ✅ **完整GitLab集成** - 自动创建GitLab用户和项目
- ✅ **真实权限管理** - 通过API验证权限和角色
- ✅ **数据一致性** - 确保系统和GitLab数据同步
- ✅ **可扩展性** - 易于添加新的测试数据类型
- ✅ **错误处理** - 完善的错误检查和回滚机制

### 前置条件

#### 1. 系统依赖
```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install curl jq

# CentOS/RHEL
sudo yum install curl jq

# macOS
brew install curl jq
```

#### 2. 服务启动
确保以下服务正在运行：

```bash
# 启动所有服务
docker-compose up -d

# 或分别启动
docker-compose up -d postgres redis gitlab
cd backend && go run cmd/main.go
cd frontend && npm run dev
```

#### 3. GitLab配置

**方式一：自动配置（推荐）**
1. 访问 GitLab: http://localhost:8081
2. 使用root账户登录（初始密码在容器日志中）
3. 创建个人访问令牌：
   - 访问：http://localhost:8081/-/profile/personal_access_tokens
   - 令牌名称：`gitlabex-test-data`
   - 权限范围：`api`, `read_user`, `read_repository`, `write_repository`
   - 点击"Create personal access token"
4. 将生成的令牌保存到配置文件：
   ```bash
   # 编辑配置文件
   vim config/oauth.env
   
   # 设置GITLAB_ACCESS_TOKEN字段
   GITLAB_ACCESS_TOKEN=your_gitlab_token_here
   ```

**方式二：手动输入**
- 如果配置文件中未设置令牌，脚本会自动切换到手动输入模式
- 按照提示输入GitLab个人访问令牌
- 可选择将令牌保存到配置文件中以便下次使用

### 使用方法

#### 基本用法
```bash
# 执行脚本
./scripts/init-test-data-api.sh
```

#### 自定义配置
```bash
# 自定义服务地址
API_BASE_URL=http://localhost:8080 \
GITLAB_URL=http://localhost:8081 \
./scripts/init-test-data-api.sh
```

### 执行流程

1. **依赖检查** - 验证curl和jq是否安装
2. **服务检查** - 确认后端API和GitLab服务可用
3. **令牌输入** - 提示输入GitLab访问令牌
4. **令牌验证** - 验证令牌有效性和权限
5. **创建GitLab用户** - 在GitLab中创建测试用户
6. **创建系统用户** - 通过API在系统中创建用户
7. **创建课题项目** - 创建研究课题并关联GitLab项目
8. **添加项目成员** - 为项目分配成员和角色
9. **创建话题讨论** - 为项目创建讨论话题
10. **创建课程作业** - 为项目创建作业任务
11. **显示统计** - 展示创建的数据统计信息

### 创建的测试数据

#### 用户账户
| 用户名 | 邮箱 | 姓名 | 角色 | 密码 |
|--------|------|------|------|------|
| admin | admin@gitlabex.com | 系统管理员 | admin | Kx9#mP2$vL8@nQ5! |
| prof_wang | wang.prof@university.edu | 王明教授 | teacher | Kx9#mP2$vL8@nQ5! |
| prof_li | li.prof@university.edu | 李华教授 | teacher | Kx9#mP2$vL8@nQ5! |
| prof_zhang | zhang.prof@university.edu | 张伟教授 | teacher | Kx9#mP2$vL8@nQ5! |
| ta_chen | chen.ta@university.edu | 陈小明助教 | assistant | Kx9#mP2$vL8@nQ5! |
| ta_wu | wu.ta@university.edu | 吴晓丽助教 | assistant | Kx9#mP2$vL8@nQ5! |
| student_001 | student001@university.edu | 张三 | student | Kx9#mP2$vL8@nQ5! |
| student_002 | student002@university.edu | 李四 | student | Kx9#mP2$vL8@nQ5! |
| student_003 | student003@university.edu | 王五 | student | Kx9#mP2$vL8@nQ5! |
| student_004 | student004@university.edu | 赵六 | student | Kx9#mP2$vL8@nQ5! |

#### 研究课题
- **航天工程类**
  - Starlink星座分析
  - 卫星通信链路优化
  - 轨道机动计算器
  - 火箭回收技术研究
  - LEO轨道碰撞预警

- **人工智能类**
  - 深度学习图像识别
  - 自然语言处理平台
  - 强化学习游戏AI

- **软件工程类**
  - 微服务架构实践
  - React全栈开发

- **数据科学类**
  - 大数据分析平台
  - 机器学习预测模型

#### 其他数据
- **话题讨论** - 每个项目包含多个技术讨论话题
- **课程作业** - 为主要项目创建相关作业任务
- **项目成员** - 自动分配助教和学生到各个项目

### 故障排除

#### 常见问题

**1. GitLab令牌验证失败**
```
❌ GitLab令牌验证失败，请检查令牌是否正确
```
解决方案：
- 确认GitLab服务正在运行
- 检查令牌权限范围是否包含`api`
- 确认令牌未过期

**2. 后端API服务未启动**
```
❌ 后端API服务未启动，请先启动后端服务
```
解决方案：
```bash
cd backend
go run cmd/main.go
```

**3. GitLab服务未启动**
```
❌ GitLab服务未启动，请先启动GitLab
```
解决方案：
```bash
docker-compose up -d gitlab
```

**4. 用户创建失败**
```
❌ 系统用户 username 创建失败
```
解决方案：
- 检查数据库连接
- 确认用户不存在冲突
- 查看后端日志获取详细错误信息

#### 调试模式

启用详细日志：
```bash
set -x  # 在脚本开头添加此行
./scripts/init-test-data-api.sh
```

查看API响应：
```bash
# 手动测试API
curl -v -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/projects
```

### 清理测试数据

如需清理测试数据：

```bash
# 清理系统数据
./scripts/cleanup-test-data.sh

# 或手动清理GitLab数据
# 1. 访问GitLab管理界面
# 2. 删除测试项目和用户
```

## 传统SQL方式

### 使用方法
```bash
# 执行传统SQL初始化
./scripts/init-test-data.sh
```

### 特点
- ✅ **快速执行** - 直接插入SQL，速度快
- ✅ **离线运行** - 不需要GitLab服务
- ❌ **无GitLab集成** - 不创建GitLab项目和用户
- ❌ **数据不一致** - 可能导致系统和GitLab数据不同步

### 适用场景
- 快速开发测试
- 离线环境
- 不需要GitLab功能的场景

## 对比总结

| 特性 | API驱动方式 | 传统SQL方式 |
|------|-------------|-------------|
| GitLab集成 | ✅ 完整集成 | ❌ 无集成 |
| 数据一致性 | ✅ 保证一致 | ⚠️ 可能不一致 |
| 执行速度 | ⚠️ 较慢 | ✅ 快速 |
| 权限验证 | ✅ 真实验证 | ❌ 无验证 |
| 错误处理 | ✅ 完善 | ⚠️ 基础 |
| 维护性 | ✅ 易维护 | ⚠️ 需手动同步 |

## 推荐使用

**生产环境测试**: 使用API驱动方式
**开发调试**: 根据需要选择合适的方式
**演示展示**: 使用API驱动方式以获得完整功能

---

如有问题，请查看项目文档或提交Issue。
