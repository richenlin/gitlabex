# GitLabEx - 基于GitLab深度集成的教育增强平台解决方案

## 项目概述

GitLabEx是一个**GitLab教育增强平台**，基于GitLab现有能力的轻量级增强方案。采用Go后端 + Vue前端的技术架构，核心目标是：

- 🔗 **最大化复用GitLab能力** - 用户管理、团队协作、权限控制、项目管理完全依赖GitLab
- 📚 **提供教育场景优化** - 基于GitLab功能的教育友好界面和工作流
- ✏️ **集成OnlyOffice协作编辑** - 这是我们的核心差异化功能
- 🎯 **简化复杂度** - 减少70%以上的自定义代码，专注核心价值

## 设计理念
- ✅ **GitLab First** - 优先使用GitLab原生功能
- ✅ **教育增强** - 专注GitLab在教育场景的优化
- ✅ **轻量集成** - 最小化自定义逻辑，最大化API复用
- ✅ **核心价值** - 聚焦OnlyOffice集成和教育UI优化

## 技术架构

### 整体架构设计

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Vue.js 前端     │    │  Go 后端服务     │    │   GitLab CE     │
│                 │    │                 │    │                 │
│ - 教育UI门户     │◄──►│ - GitLab API    │◄──►│ - 用户管理       │
│ - OnlyOffice    │    │ - OnlyOffice    │    │ - 团队管理       │
│ - 简化界面       │    │ - 轻量业务逻辑    │    │ - 权限控制       │
└─────────────────┘    └─────────────────┘    │ - 项目管理       │
         │                       │            │ - 代码管理       │
         │                       │            │ - Wiki文档       │
         └───────────────────────┼────────────┘ ────────────────            
                                 │
                    ┌─────────────────┐
                    │   数据层         │
                    │                 │
                    │ - PostgreSQL    │  (仅存储必要的业务数据)
                    │ - Redis         │  (缓存GitLab API数据)
                    │ - OnlyOffice    │  (文档协作服务)
                    └─────────────────┘
```

### 核心技术栈

#### 后端技术
- **语言**: Go 1.21+
- **Web框架**: Gin
- **数据库**: PostgreSQL 15+ (极简化数据模型)
- **缓存**: Redis 7+ (主要缓存GitLab API数据)
- **GitLab集成**: GitLab API v4
- **文档服务**: OnlyOffice Document Server
- **容器化**: Docker & Docker Compose

#### 前端技术
- **框架**: Vue 3.4+
- **构建工具**: Vite
- **状态管理**: Pinia
- **UI组件库**: Element Plus
- **文档编辑器**: OnlyOffice Document Server
- **实时通信**: WebSocket (基于GitLab Webhook)

## 系统流程设计

### OAuth登录流程

系统登录流程完全基于GitLab OAuth2.0，具体步骤如下：

1. **登录态检查**
   - 系统除部分特殊页面（登录、帮助等）外，其他页面均检查登录态
   - 通过JWT Token和GitLab API验证用户身份

2. **登录跳转**
   - 未登录用户自动跳转到login页面
   - 用户点击"通过GitLab登录"按钮，启动OAuth流程

3. **GitLab OAuth流程**
   - 跳转到GitLab登录页面进行认证
   - 用户登录成功后，GitLab回调到系统指定URL
   - 系统获取授权码，换取访问令牌
   - 自动跳转回系统首页（仪表盘）

4. **用户信息同步**
   - 获取GitLab用户信息并同步到本地
   - 建立用户会话，设置JWT Token
   - 缓存用户权限信息

### 权限系统设计

#### 角色定义
- **管理员（Admin）**: 可以看到所有数据，管理系统设置
- **老师（Teacher）**: 可以创建并管理班级和课题，管理学生
- **学生（Student）**: 可以参与班级和课题，提交作业
- **访客（Guest）**: 无法进入系统

#### 权限控制规则

1. **第三方API访问**
   - 增加用户、班级和角色管理的API接口
   - 第三方系统使用OAuth获取token，能够调用这些API
   - 实现基于JWT的API访问控制

2. **教师权限**
   - 可以创建并管理班级（GitLab Group）
   - 可以新建并管理课题
   - 可以添加学生到班级或课题
   - 可以查看和管理班级内学生的学习进度
   - 可以审核学生作业并生成报告
   - 只能看到自己的班级和自己班级的课题

3. **学生权限**
   - 只能加入班级或被老师添加进班级
   - 只能参加课题或被老师添加入课题
   - 可以针对课题完成作业
   - 只能看到自己的班级和自己加入的课题
   - 只能查看自己的作业和成绩

4. **作业流程权限**
   - 学生完成作业后，自动发送通知给课题创建老师
   - 老师通过学习进度跟踪学生完成课题作业的情况
   - 学生完成作业后，老师给出评审并生成作业报告
   - 评审完成后，发送通知给学生

## 功能模块设计

### 1. 用户管理模块 - 完全基于GitLab

#### 功能特性
- ✅ GitLab OAuth2.0登录（无需自定义认证）
- ✅ 用户信息同步（从GitLab API获取）
- ✅ 角色映射（GitLab权限 -> 教育角色）
- ✅ 用户资料展示（GitLab用户资料）
- ✅ 第三方API访问控制

### 2. 课题管理模块 - 基于GitLab项目

#### 功能特性
- ✅ **课题即Git仓库**：系统的课题以git repo形式存在
- ✅ **课题信息管理**：课题名就是repo的标题，课题介绍就是repo的readme
- ✅ **自动仓库创建**：课题创建时自动创建GitLab仓库
- ✅ **学生分支管理**：学生参与课题时自动创建个人分支
- ✅ **权限同步**：课题权限与GitLab项目权限完全同步
- **统计分析**：老师和学生根据自身权限查看统计信息（参与的话题、作业的完成评价、作业分析等等）

#### 学生作业提交流程
1. **分支创建**：学生加入课题时，系统自动在对应的Git仓库下创建学生个人分支
2. **作业提交**：学生提交作业就是在个人分支下提交代码和附件
3. **内容管理**：支持代码文件、文档附件等多种类型文件提交
4. **版本控制**：利用Git原生功能实现作业版本管理

### 3. 协作交流模块 - 基于GitLab原生功能

#### 讨论功能
- ✅ **Issues讨论**：课题可以讨论，讨论以issues方式组织
- ✅ **话题管理**：学生可以新增和评论话题，编辑、删除自己的话题
- ✅ **教师管理**：老师可以管理自己班级学生的话题讨论

#### 通知系统
- ✅ **GitLab集成**：公告系统利用GitLab的项目通知机制
- ✅ **自动通知**：参与repo后相关人员会收到交互通知
- ✅ **事件驱动**：基于GitLab Webhook实现实时通知

### 4. 文档管理模块 - GitLab Wiki + OnlyOffice

#### 功能特性
- ✅ **基于GitLab Wiki**：文档管理使用GitLab的wiki实现
- ✅ **权限控制**：文档权限完全基于GitLab项目Wiki权限
- ✅ **附件支持**：支持文档附件上传（Word、Excel、PowerPoint等）
- ✅ **OnlyOffice集成**：具有Wiki权限的成员可以使用OnlyOffice编辑文档附件
- ✅ **版本控制**：文档版本控制使用GitLab原生功能

#### 权限管理
- **学生权限**：可以新建、管理自己的文档
- **教师权限**：可以管理自己班级学生的文档
- **查看权限**：学生可以查看其他人的文档（基于GitLab权限）
- **班级管理**：仅老师可见，管理班级和学生（基于GitLab Group实现）

## 前端界面设计

### 整体布局
- **顶部栏**：用户菜单、公告通知、系统导航
- **左侧菜单**：功能模块导航
- **中间内容区域**：主要功能展示区域
- **采用通用的后台管理风格**

### 顶部栏设计

#### 公告通知功能
1. **公告入口**：用户菜单左侧增加公告按钮
2. **未读提示**：如果收到新的公告，显示红色角标提示（未读公告数量）
3. **公告管理**：
   - 点击公告菜单，打开公告管理页面
   - 获取公告列表数据
   - 点击列表标题显示公告详情
   - 公告阅读后，状态变成已读
   - 所有公告已读后，顶栏角标消失

### 左侧菜单设计

#### 菜单结构
- **首页（仪表盘）**：作为系统首页，展示概览信息
- **班级管理**：仅老师可见，管理班级和学生
- **课题管理**：老师和学生可见，管理课题项目
- **作业管理**：老师和学生可见，作业提交和评审
- **统计分析**：老师和学生可见，学习数据分析
- **文档管理**：老师和学生可见，Wiki文档管理
- **话题讨论**：老师和学生可见，GitLab讨论功能

### 核心页面设计

#### 1. 班级管理页面
- **功能定位**：展现班级列表，老师专用页面
- **主要功能**：
  - 老师可以管理维护班级信息
  - 新增学生到班级
  - 将现有学生加入到班级
  - 班级成员管理和权限分配

#### 2. 课题管理页面
- **教师视图**：
  - 管理维护自己创建的课题列表
  - 添加学生参与课题
  - 查看课题进度和统计
- **学生视图**：
  - 查看自己参与的课题
  - 访问课题详情页面
- **课题详情**：点击课题查看课题详情（对应GitLab repo页面）

#### 3. 作业管理页面
- **提交功能**：学生提交作业到个人分支
- **审核功能**：老师查看和评审学生作业
- **进度跟踪**：实时显示作业完成状态
- **通知机制**：作业状态变更自动通知相关人员

#### 4. 文档管理页面
- **教师权限**：
  - 管理自己的文档
  - 管理自己班级下所有学生的文档
- **学生权限**：
  - 管理自己的文档
  - 查看其他人的文档（基于权限）
- **编辑功能**：集成OnlyOffice实现在线编辑

#### 5. 话题讨论页面
- **功能定位**：基于GitLab Discussions实现话题讨论
- **保持GitLab Discussions页面设计**
- **界面中文化**：使用中文界面和提示
- **权限控制**：基于用户角色显示不同操作权限

## 前端架构设计

### Vue应用结构

```
src/
├── components/           # 通用组件
│   ├── OnlyOfficeEditor/ # OnlyOffice编辑器集成
│   ├── GitLabWidget/     # GitLab组件封装
│   ├── EducationUI/      # 教育场景UI组件
│   ├── NotificationBell/ # 公告通知组件
│   └── Common/          # 通用组件
├── views/               # 页面视图
│   ├── Dashboard/       # 仪表板（系统首页）
│   ├── Classes/         # 班级管理
│   ├── Projects/        # 课题管理
│   ├── Assignments/     # 作业管理
│   ├── Documents/       # 文档管理（GitLab Wiki + OnlyOffice）
│   ├── Analytics/       # 统计分析
│   ├── Notifications/   # 公告管理
│   └── Discussions/     # 话题讨论
├── stores/              # 状态管理
│   ├── auth.js          # 认证状态管理
│   ├── gitlab.js        # GitLab API状态
│   ├── onlyoffice.js    # OnlyOffice状态
│   ├── notifications.js # 通知状态管理
│   └── education.js     # 教育场景状态
├── services/            # API服务
│   ├── gitlab.js        # GitLab API封装
│   ├── onlyoffice.js    # OnlyOffice API封装
│   ├── auth.js          # 认证服务
│   └── education.js     # 教育业务逻辑
├── router/              # 路由管理
│   ├── index.js         # 主路由配置
│   ├── guards.js        # 路由守卫（权限检查）
│   └── routes.js        # 路由定义
└── utils/               # 工具函数
    ├── auth.js          # GitLab OAuth认证
    ├── permission.js    # 权限工具
    ├── notification.js  # 通知工具
    └── format.js        # 数据格式化
```

## 数据模型设计

### 核心数据模型

```go
// 用户模型 - 同步GitLab用户信息
type User struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    GitLabID    int       `json:"gitlab_id" gorm:"uniqueIndex"`
    Username    string    `json:"username"`
    Email       string    `json:"email"`
    Name        string    `json:"name"`
    Avatar      string    `json:"avatar"`
    Role        string    `json:"role"` // admin, teacher, student, guest
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// 班级模型 - 对应GitLab Group
type Class struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    GitLabGroupID int     `json:"gitlab_group_id"`
    TeacherID   uint      `json:"teacher_id"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// 课题模型 - 对应GitLab Project
type Project struct {
    ID              uint      `json:"id" gorm:"primaryKey"`
    Name            string    `json:"name"`
    Description     string    `json:"description"`
    GitLabProjectID int       `json:"gitlab_project_id"`
    GitLabURL       string    `json:"gitlab_url"`
    RepositoryURL   string    `json:"repository_url"`
    ClassID         uint      `json:"class_id"`
    TeacherID       uint      `json:"teacher_id"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

// 作业模型 - 基于GitLab Issues和Commits
type Assignment struct {
    ID              uint      `json:"id" gorm:"primaryKey"`
    Title           string    `json:"title"`
    Description     string    `json:"description"`
    ProjectID       uint      `json:"project_id"`
    GitLabIssueID   int       `json:"gitlab_issue_id"`
    DueDate         time.Time `json:"due_date"`
    RequiredFiles   string    `json:"required_files"`
    MaxFileSize     int64     `json:"max_file_size"`
    AllowedFileTypes string   `json:"allowed_file_types"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

// 作业提交模型 - 基于GitLab Commits
type AssignmentSubmission struct {
    ID            uint      `json:"id" gorm:"primaryKey"`
    AssignmentID  uint      `json:"assignment_id"`
    StudentID     uint      `json:"student_id"`
    CommitHash    string    `json:"commit_hash"`
    CommitMessage string    `json:"commit_message"`
    BranchName    string    `json:"branch_name"`
    FilesSubmitted string   `json:"files_submitted"`
    Status        string    `json:"status"` // submitted, reviewed, graded
    Score         int       `json:"score"`
    Feedback      string    `json:"feedback"`
    SubmittedAt   time.Time `json:"submitted_at"`
    ReviewedAt    *time.Time `json:"reviewed_at"`
}

// 通知模型
type Notification struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    UserID    uint      `json:"user_id"`
    Type      string    `json:"type"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    IsRead    bool      `json:"is_read"`
    CreatedAt time.Time `json:"created_at"`
}
```

## API接口设计

### 认证接口
```
POST /api/auth/gitlab/login     # GitLab OAuth登录
POST /api/auth/gitlab/callback  # GitLab OAuth回调
POST /api/auth/refresh          # 刷新Token
POST /api/auth/logout           # 退出登录
```

### 用户管理接口
```
GET  /api/users                 # 获取用户列表
GET  /api/users/:id             # 获取用户详情
PUT  /api/users/:id             # 更新用户信息
POST /api/users/:id/role        # 设置用户角色
```

### 班级管理接口
```
GET  /api/classes               # 获取班级列表
POST /api/classes               # 创建班级
PUT  /api/classes/:id           # 更新班级信息
DELETE /api/classes/:id         # 删除班级
POST /api/classes/:id/students  # 添加学生到班级
```

### 课题管理接口
```
GET  /api/projects              # 获取课题列表
POST /api/projects              # 创建课题
PUT  /api/projects/:id          # 更新课题信息
DELETE /api/projects/:id        # 删除课题
POST /api/projects/:id/students # 添加学生到课题
GET  /api/projects/:id/gitlab   # 获取GitLab项目信息
```

### 作业管理接口
```
GET  /api/assignments           # 获取作业列表
POST /api/assignments           # 创建作业
PUT  /api/assignments/:id       # 更新作业信息
DELETE /api/assignments/:id     # 删除作业
POST /api/assignments/:id/submit # 提交作业
GET  /api/assignments/:id/submissions # 获取作业提交列表
```

### 文档管理接口
```
GET  /api/documents             # 获取文档列表
POST /api/documents             # 创建文档
PUT  /api/documents/:id         # 更新文档
DELETE /api/documents/:id       # 删除文档
POST /api/documents/:id/onlyoffice # OnlyOffice编辑接口
```

### 通知管理接口
```
GET  /api/notifications         # 获取通知列表
PUT  /api/notifications/:id/read # 标记通知已读
POST /api/notifications/read-all # 标记全部已读
```

## 项目实施计划

### 阶段一：基础架构搭建（已完成）
- ✅ 搭建测试环境（GitLab CE、OnlyOffice、PostgreSQL、Redis）
- ✅ 基础架构搭建（Go后端、Vue前端框架）
- ✅ GitLab API集成（OAuth、基础API封装）
- ✅ 用户管理模块（GitLab用户体系集成）

### 阶段二：核心功能实现（已完成）
- ✅ GitLab深度集成（项目、分支、提交管理）
- ✅ 权限系统实现（基于GitLab权限模型）
- ✅ 班级管理（GitLab Groups映射）
- ✅ 课题管理（GitLab Projects映射）
- ✅ 作业管理（GitLab Issues/MR/Commits）

### 阶段三：教育场景完善（已完成）
- ✅ 学习进度跟踪（基于GitLab Activity）
- ✅ 通知系统（基于GitLab Webhook）
- ✅ 文档管理（GitLab Wiki集成）
- ✅ 权限同步（GitLab与教育角色映射）

### 阶段四：前端界面优化（进行中）
- 🔄 基于设计规范实现前端界面
- 🔄 公告通知系统前端实现
- 🔄 响应式设计和用户体验优化
- 🔄 OnlyOffice编辑器集成

### 阶段五：集成测试和部署
- 系统集成测试
- 性能优化（缓存、并发）
- 安全加固（GitLab OAuth、权限控制）
- 部署上线（Docker Compose一键部署）
