# 基于GitLab API + Webhook的教育增强平台解决方案

## 项目概述

本系统是一个**GitLab教育增强平台**，不是重复造轮子的完整社区系统，而是基于GitLab现有能力的轻量级增强方案。采用Go后端 + Vue前端的技术架构，核心目标是：

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

## 功能模块设计

### 1. 用户管理模块 - 完全基于GitLab

#### 功能特性
- ✅ GitLab OAuth2.0登录（无需自定义认证）
- ✅ 用户信息同步（从GitLab API获取）
- ✅ 角色映射（GitLab权限 -> 教育角色）
- ✅ 用户资料展示（GitLab用户资料）


### 2. 团队管理模块 - 使用GitLab Group

#### 功能特性
- ✅ 直接使用GitLab Group作为班级/团队
- ✅ 支持多层级：学校 -> 学院 -> 班级 -> 项目组
- ✅ 成员管理通过GitLab Group Members API
- ✅ 权限管理使用GitLab原生权限系统

### 3. 权限控制模块 - 基于GitLab权限模型

#### 功能特性
- ✅ 完全使用GitLab的5级权限系统
- ✅ 权限检查通过GitLab API实时获取
- ✅ 支持Group级别和Project级别权限
- ✅ 教育场景权限映射


### 4. 教育管理模块 - 基于GitLab Issues/Discussions

#### 功能特性
- ✅ 课题管理（GitLab Issues + 课题标签）
- ✅ 作业管理（GitLab Issues + 作业标签）  
- ✅ 话题讨论（GitLab Discussions）
- ✅ 公告发布（GitLab Issues + 公告标签）
- ✅ 作业提交（GitLab Merge Request）

### 5. 文档管理模块 - GitLab Wiki + 文档附件 + OnlyOffice

#### 功能特性
- ✅ 基于GitLab Wiki的文档管理
- ✅ 支持文档附件上传（Word、Excel、PowerPoint等）
- ✅ 具有Wiki权限的成员可以使用OnlyOffice编辑文档附件
- ✅ 文档版本控制使用GitLab原生功能
- ✅ 文档权限完全基于GitLab项目Wiki权限

## 前端架构设计

### Vue应用结构

```
src/
├── components/           # 通用组件
│   ├── OnlyOfficeEditor/ # OnlyOffice编辑器集成
│   ├── GitLabWidget/     # GitLab组件封装
│   ├── EducationUI/      # 教育场景UI组件
│   └── Common/          # 通用组件
├── views/               # 页面视图
│   ├── Dashboard/       # 仪表板（聚合GitLab数据）
│   ├── Groups/          # 班级管理（GitLab Groups）
│   ├── Projects/        # 项目管理（GitLab Projects）
│   ├── Documents/       # 文档管理（GitLab Wiki + OnlyOffice）
│   └── Assignments/     # 作业管理（GitLab Issues/MR）
├── stores/              # 状态管理
│   ├── gitlab.js        # GitLab API状态
│   ├── onlyoffice.js    # OnlyOffice状态
│   └── education.js     # 教育场景状态
├── services/            # API服务
│   ├── gitlab.js        # GitLab API封装
│   ├── onlyoffice.js    # OnlyOffice API封装
│   └── education.js     # 教育业务逻辑
└── utils/               # 工具函数
    ├── auth.js          # GitLab OAuth认证
    ├── permission.js    # 权限工具
    └── format.js        # 数据格式化
```

## 项目实施计划 - 修订版

### 第一阶段（2周）- 环境和基础架构
- ✅ 搭建测试环境（GitLab CE、OnlyOffice、PostgreSQL、Redis）
- ✅ 基础架构搭建（Go后端、Vue前端框架）
- ✅ GitLab API集成（OAuth、基础API封装）
- 🔄 用户管理模块（直接使用GitLab用户体系）

### 第二阶段（3周）- 核心功能实现
- OnlyOffice集成（文档协作编辑）
- 教育UI优化（基于GitLab数据的友好界面）
- 班级管理（GitLab Groups映射）
- 作业管理（GitLab Issues/MR）

### 第三阶段（3周）- 教育场景完善
- 学习进度跟踪（基于GitLab Activity）
- 通知系统（基于GitLab Webhook）
- 教育报表（GitLab数据分析）
- 权限管理（GitLab权限映射）

### 第四阶段（2周）- 前端页面设计
- 基于gitlab现有前端页面进行设计，
- 系统包含登录、主视觉界面；主界面包含顶部栏、左侧菜单、中间内容区域（常见的后台管理风格）
- 登录后主界面，左侧菜单有“项目管理”、“团队管理”、“课题管理”、“作业管理”、“学习进度跟踪”、“通知系统”、“教育报表”等页面
- 根据需求设计，实现核心功能
### 第五阶段（2周）- 集成测试和部署
- 系统集成测试
- 性能优化（缓存、并发）
- 安全加固（GitLab OAuth、权限控制）
- 部署上线（Docker Compose一键部署）
