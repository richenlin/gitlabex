# GitLabEx 教育工作流测试指南

## 🚀 测试环境准备

### 1. 启动后端服务
```bash
cd backend
./main
```

### 2. 启动前端服务
```bash
cd frontend
npm run dev
```

## 📋 完整教育工作流测试流程

### 测试场景：完整的教育工作流
这个测试模拟了从创建班级到学生提交作业，教师批改的完整流程。

### 第一步：创建测试班级
**API 端点：** `POST /api/education/test/project/1`
**请求体：**
```json
{
  "title": "软件工程班",
  "description": "2024年软件工程专业班级"
}
```

**期望结果：**
- 返回 GitLab 项目信息
- 项目具有 Issues、MR、Wiki 功能
- 状态码：201

### 第二步：创建测试作业
**API 端点：** `POST /api/education/test/assignment/{project_id}`
**请求体：**
```json
{
  "title": "第一次作业：系统设计",
  "description": "请设计一个简单的学生管理系统，包含数据库设计和API设计"
}
```

**期望结果：**
- 返回 GitLab Issue 信息
- Issue 具有 "作业" 和 "assignment" 标签
- 状态码：201

### 第三步：创建班级公告
**API 端点：** `POST /api/education/test/announcement/{project_id}`
**请求体：**
```json
{
  "title": "重要通知：期中考试安排",
  "content": "期中考试将在下周进行，请同学们做好准备。考试内容包括：\n1. 系统设计基础\n2. 数据库设计\n3. API 设计原则"
}
```

**期望结果：**
- 返回 GitLab Issue 信息
- Issue 具有 "公告" 和 "announcement" 标签
- 状态码：201

### 第四步：测试完整工作流
**API 端点：** `POST /api/education/test/workflow/1`
**请求体：** 无

**期望结果：**
- 返回包含 5 个步骤的工作流测试结果
- 每个步骤都应该成功完成
- 状态码：200

**工作流步骤：**
1. 创建项目 ✅
2. 创建作业 ✅
3. 创建公告 ✅
4. 学生提交作业 ✅
5. 教师批改作业 ✅

### 第五步：查看作业列表
**API 端点：** `GET /api/education/test/assignments/{project_id}`

**期望结果：**
- 返回作业 Issue 列表
- 每个作业具有正确的标签
- 状态码：200

### 第六步：查看作业提交列表
**API 端点：** `GET /api/education/test/submissions/{project_id}`

**期望结果：**
- 返回作业提交 MR 列表
- 每个 MR 具有正确的标签
- 状态码：200

### 第七步：查看教育统计数据
**API 端点：** `GET /api/education/test/stats/1`

**期望结果：**
- 返回教育统计数据
- 包含班级数量、项目数量等
- 状态码：200

## 🎯 前端界面测试

### 1. 访问 Wiki 文档管理页面
**URL：** `http://localhost:3000/wiki`

**测试要点：**
- 项目选择器正常工作
- 可以查看 Wiki 页面列表
- 可以创建新的 Wiki 页面
- 可以上传文档附件
- OnlyOffice 编辑器正常启动

### 2. 访问班级管理页面
**URL：** `http://localhost:3000/classes`

**测试要点：**
- 班级列表正常显示
- 可以创建新班级
- 成员管理功能正常
- 统计数据正确显示

### 3. 访问作业管理页面
**URL：** `http://localhost:3000/assignments`

**测试要点：**
- 作业列表正常显示
- 可以创建新作业
- 提交管理功能正常
- 批改功能正常

### 4. 访问项目管理页面
**URL：** `http://localhost:3000/projects`

**测试要点：**
- 项目列表正常显示
- 可以创建新项目
- 进度跟踪功能正常
- 成员管理功能正常

## 📝 测试结果验证

### 成功标准
- [ ] 所有 API 端点返回正确的状态码
- [ ] 前端页面正常加载和交互
- [ ] 工作流测试所有步骤成功完成
- [ ] GitLab 集成正常工作
- [ ] OnlyOffice 编辑器正常启动

### 错误处理验证
- [ ] 无效参数时返回适当的错误信息
- [ ] 权限不足时返回适当的错误信息
- [ ] GitLab 连接失败时有适当的错误处理

## 🔧 故障排除

### 常见问题
1. **GitLab 连接失败**
   - 检查 GitLab 服务是否运行
   - 验证配置文件中的 GitLab URL 和 Token

2. **OnlyOffice 编辑器无法启动**
   - 检查 OnlyOffice 服务是否运行
   - 验证配置文件中的 OnlyOffice URL

3. **前端页面无法加载**
   - 检查后端服务是否正在运行
   - 验证 API 端点是否正确

## 🎉 测试完成

完成所有测试后，你应该能够：
- 创建和管理班级
- 布置和管理作业
- 学生提交作业
- 教师批改作业
- 使用 Wiki 文档管理和 OnlyOffice 编辑器
- 查看各种统计数据

这证明 GitLabEx 教育增强平台的核心功能已经完全实现！ 