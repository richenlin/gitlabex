# 基于GitLab API + Webhook的社区系统解决方案

## 项目概述

基于GitLab API + Webhook构建的教育社区系统，采用Go后端 + Vue前端的技术架构，具备知识文档管理、在线协作编辑、话题管理、用户团队管理和代码开发管理等功能。

## 技术架构

### 整体架构设计

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Vue.js 前端    │    │   Go 后端服务    │    │   GitLab CE     │
│                 │    │                 │    │                 │
│ - 用户界面      │◄──►│ - RESTful API   │◄──►│ - Git 仓库      │
│ - 实时编辑器    │    │ - WebSocket     │    │ - Wiki 管理     │
│ - 协作工具      │    │ - 业务逻辑      │    │ - 用户管理      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   数据层        │
                    │                 │
                    │ - PostgreSQL    │
                    │ - Redis         │
                    │ - 文件存储      │
                    └─────────────────┘
```

### 核心技术栈

#### 后端技术
- **语言**: Go 1.21+
- **Web框架**: Gin
- **数据库**: PostgreSQL 15+
- **缓存**: Redis 7+
- **ORM**: GORM
- **WebSocket**: Gorilla WebSocket
- **GitLab集成**: GitLab API v4
- **容器化**: Docker & Docker Compose

#### 前端技术
- **框架**: Vue 3.4+
- **构建工具**: Vite
- **状态管理**: Pinia
- **UI组件库**: Element Plus
- **文档编辑器**: OnlyOffice Document Server
- **代码编辑器**: Monaco Editor
- **实时通信**: WebSocket
- **路由**: Vue Router

#### 基础设施
- **容器编排**: Docker Compose
- **反向代理**: Nginx
- **文档服务**: OnlyOffice Document Server
- **监控**: Prometheus + Grafana
- **日志**: ELK Stack

## 功能模块设计

### 1. 用户管理模块

#### 功能特性
- GitLab OAuth2.0集成登录
- 用户角色管理（管理员、教师、学生）
- 权限控制系统
- 用户资料同步

#### 实现方案
```go
// 用户服务结构
type UserService struct {
    db       *gorm.DB
    gitlab   *gitlab.Client
    redis    *redis.Client
}

// 用户角色枚举
type UserRole int

const (
    RoleStudent UserRole = iota
    RoleTeacher
    RoleAdmin
)

// 用户模型
type User struct {
    ID          uint      `gorm:"primaryKey"`
    GitLabID    int       `gorm:"unique;not null"`
    Username    string    `gorm:"unique;not null"`
    Email       string    `gorm:"unique;not null"`
    Name        string    `gorm:"not null"`
    Role        UserRole  `gorm:"not null;default:0"`
    Avatar      string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### 2. 知识文档管理模块

#### 功能特性
- 基于GitLab Wiki的文档管理
- 文档版本控制
- 访问权限管理
- 文档分类和标签
- 文档搜索和索引

#### 实现方案
```go
// 文档服务
type DocumentService struct {
    gitlab    *gitlab.Client
    db        *gorm.DB
    indexer   *bleve.Index
}

// 文档模型
type Document struct {
    ID          uint      `gorm:"primaryKey"`
    ProjectID   int       `gorm:"not null"`
    Slug        string    `gorm:"not null"`
    Title       string    `gorm:"not null"`
    Content     string    `gorm:"type:text"`
    Format      string    `gorm:"default:markdown"`
    CategoryID  uint
    Tags        []Tag     `gorm:"many2many:document_tags"`
    CreatedBy   uint
    UpdatedBy   uint
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// GitLab Wiki API集成
func (s *DocumentService) CreateDocument(doc *Document) error {
    // 调用GitLab API创建Wiki页面
    _, _, err := s.gitlab.Wikis.CreateWikiPage(doc.ProjectID, &gitlab.CreateWikiPageOptions{
        Title:   &doc.Title,
        Content: &doc.Content,
        Format:  &doc.Format,
    })
    return err
}
```

### 3. 在线协作编辑模块

#### 功能特性
- 基于OnlyOffice的实时多人协作编辑
- 支持Word、Excel、PowerPoint文档
- 实时协作和评论功能
- 版本历史管理
- 文档权限控制
- 离线编辑同步

#### 实现方案
```go
// OnlyOffice协作编辑服务
type OnlyOfficeService struct {
    config        *OnlyOfficeConfig
    db            *gorm.DB
    redis         *redis.Client
    fileStorage   *FileStorage
    gitlab        *gitlab.Client
}

// OnlyOffice配置
type OnlyOfficeConfig struct {
    DocumentServerURL string `json:"document_server_url"`
    JWTSecret         string `json:"jwt_secret"`
    CallbackURL       string `json:"callback_url"`
    MaxFileSize       int64  `json:"max_file_size"`
}

// 文档会话
type DocumentSession struct {
    ID           uint      `gorm:"primaryKey"`
    DocumentID   uint      `gorm:"not null"`
    UserID       uint      `gorm:"not null"`
    Key          string    `gorm:"unique;not null"`
    Mode         string    `gorm:"default:edit"` // edit, view, comment
    CreatedAt    time.Time
    UpdatedAt    time.Time
    ExpiresAt    time.Time
}

// 创建编辑会话
func (s *OnlyOfficeService) CreateEditSession(docID, userID uint, mode string) (*DocumentSession, error) {
    // 生成唯一的文档密钥
    key := s.generateDocumentKey(docID, userID)
    
    session := &DocumentSession{
        DocumentID: docID,
        UserID:     userID,
        Key:        key,
        Mode:       mode,
        ExpiresAt:  time.Now().Add(24 * time.Hour),
    }
    
    if err := s.db.Create(session).Error; err != nil {
        return nil, err
    }
    
    return session, nil
}

// 获取OnlyOffice配置
func (s *OnlyOfficeService) GetEditorConfig(session *DocumentSession) (*OnlyOfficeEditorConfig, error) {
    doc, err := s.getDocument(session.DocumentID)
    if err != nil {
        return nil, err
    }
    
    user, err := s.getUser(session.UserID)
    if err != nil {
        return nil, err
    }
    
    config := &OnlyOfficeEditorConfig{
        DocumentType: s.getDocumentType(doc.FileName),
        Document: OnlyOfficeDocument{
            FileType: s.getFileExtension(doc.FileName),
            Key:      session.Key,
            Title:    doc.Title,
            URL:      s.getDocumentURL(doc.ID),
            Permissions: OnlyOfficePermissions{
                Comment:  session.Mode != "view",
                Edit:     session.Mode == "edit",
                Download: true,
                Print:    true,
            },
        },
        Editor: OnlyOfficeEditor{
            Mode:         session.Mode,
            CallbackURL:  s.config.CallbackURL,
            Lang:         "zh-CN",
            CustomizationConfig: OnlyOfficeCustomization{
                Comments:      true,
                CompactToolbar: false,
                Feedback:      false,
                Help:          true,
                Toolbar:       true,
                Zoom:          100,
            },
        },
        User: OnlyOfficeUser{
            ID:    fmt.Sprintf("%d", user.ID),
            Name:  user.Name,
            Group: user.Role.String(),
        },
    }
    
    // 生成JWT签名
    if s.config.JWTSecret != "" {
        config.Token = s.generateJWT(config)
    }
    
    return config, nil
}

// 处理OnlyOffice回调
func (s *OnlyOfficeService) HandleCallback(callback *OnlyOfficeCallback) error {
    switch callback.Status {
    case 2: // 文档已保存
        return s.handleDocumentSaved(callback)
    case 3: // 文档保存出错
        return s.handleDocumentError(callback)
    case 6: // 文档正在编辑
        return s.handleDocumentEditing(callback)
    }
    return nil
}

// 处理文档保存
func (s *OnlyOfficeService) handleDocumentSaved(callback *OnlyOfficeCallback) error {
    // 下载更新的文档
    documentContent, err := s.downloadDocument(callback.URL)
    if err != nil {
        return err
    }
    
    // 更新文档到GitLab
    doc, err := s.getDocumentByKey(callback.Key)
    if err != nil {
        return err
    }
    
    // 保存到GitLab Wiki
    content := base64.StdEncoding.EncodeToString(documentContent)
    _, _, err = s.gitlab.RepositoryFiles.UpdateFile(doc.ProjectID, doc.FilePath, &gitlab.UpdateFileOptions{
        Branch:        gitlab.String("main"),
        Content:       gitlab.String(content),
        CommitMessage: gitlab.String("Update document via OnlyOffice"),
    })
    
    return err
}

// OnlyOffice配置结构
type OnlyOfficeEditorConfig struct {
    DocumentType string              `json:"documentType"`
    Document     OnlyOfficeDocument  `json:"document"`
    Editor       OnlyOfficeEditor    `json:"editorConfig"`
    User         OnlyOfficeUser      `json:"user"`
    Token        string              `json:"token,omitempty"`
}

type OnlyOfficeDocument struct {
    FileType    string                 `json:"fileType"`
    Key         string                 `json:"key"`
    Title       string                 `json:"title"`
    URL         string                 `json:"url"`
    Permissions OnlyOfficePermissions  `json:"permissions"`
}

type OnlyOfficePermissions struct {
    Comment  bool `json:"comment"`
    Edit     bool `json:"edit"`
    Download bool `json:"download"`
    Print    bool `json:"print"`
}

type OnlyOfficeEditor struct {
    Mode                string                    `json:"mode"`
    CallbackURL         string                    `json:"callbackUrl"`
    Lang                string                    `json:"lang"`
    CustomizationConfig OnlyOfficeCustomization  `json:"customization"`
}

type OnlyOfficeCustomization struct {
    Comments       bool `json:"comments"`
    CompactToolbar bool `json:"compactToolbar"`
    Feedback       bool `json:"feedback"`
    Help           bool `json:"help"`
    Toolbar        bool `json:"toolbar"`
    Zoom           int  `json:"zoom"`
}

type OnlyOfficeUser struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Group string `json:"group"`
}

type OnlyOfficeCallback struct {
    Key        string   `json:"key"`
    Status     int      `json:"status"`
    URL        string   `json:"url,omitempty"`
    Users      []string `json:"users,omitempty"`
    Actions    []string `json:"actions,omitempty"`
    LastSave   string   `json:"lastsave,omitempty"`
    NotModified bool    `json:"notmodified,omitempty"`
}
```

### 4. 话题管理模块

#### 功能特性
- 公告系统
- 课题立项管理
- 作业管理
- 讨论区
- 通知推送

#### 实现方案
```go
// 话题服务
type TopicService struct {
    db          *gorm.DB
    gitlab      *gitlab.Client
    notification *NotificationService
}

// 话题类型
type TopicType int

const (
    TopicAnnouncement TopicType = iota
    TopicProject
    TopicAssignment
    TopicDiscussion
)

// 话题模型
type Topic struct {
    ID          uint        `gorm:"primaryKey"`
    Type        TopicType   `gorm:"not null"`
    Title       string      `gorm:"not null"`
    Content     string      `gorm:"type:text"`
    CreatedBy   uint        `gorm:"not null"`
    AssignedTo  []uint      `gorm:"serializer:json"`
    Status      string      `gorm:"default:active"`
    DueDate     *time.Time
    ProjectID   *int        // 关联的GitLab项目ID
    Comments    []Comment   `gorm:"foreignKey:TopicID"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// 作业提交
type Assignment struct {
    ID          uint      `gorm:"primaryKey"`
    TopicID     uint      `gorm:"not null"`
    StudentID   uint      `gorm:"not null"`
    ProjectID   int       `gorm:"not null"` // GitLab项目ID
    SubmittedAt time.Time
    Grade       *float64
    Feedback    string    `gorm:"type:text"`
}
```

### 5. 代码在线开发管理模块

#### 功能特性
- 在线IDE集成
- 文件管理
- Git版本控制
- 代码审查
- CI/CD集成

#### 实现方案
```go
// 代码管理服务
type CodeService struct {
    gitlab  *gitlab.Client
    db      *gorm.DB
}

// 项目模型
type Project struct {
    ID            uint      `gorm:"primaryKey"`
    GitLabID      int       `gorm:"unique;not null"`
    Name          string    `gorm:"not null"`
    Description   string
    OwnerID       uint      `gorm:"not null"`
    Members       []User    `gorm:"many2many:project_members"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

// 文件操作
func (s *CodeService) CreateFile(projectID int, path, content string) error {
    _, _, err := s.gitlab.RepositoryFiles.CreateFile(projectID, path, &gitlab.CreateFileOptions{
        Branch:        gitlab.String("main"),
        Content:       gitlab.String(content),
        CommitMessage: gitlab.String("Create " + path),
    })
    return err
}

// 获取文件内容
func (s *CodeService) GetFile(projectID int, path string) (string, error) {
    file, _, err := s.gitlab.RepositoryFiles.GetFile(projectID, path, &gitlab.GetFileOptions{
        Ref: gitlab.String("main"),
    })
    if err != nil {
        return "", err
    }
    
    decoded, err := base64.StdEncoding.DecodeString(file.Content)
    return string(decoded), err
}
```

### 6. 团队管理模块

#### 功能特性
- 项目分组
- 成员管理
- 权限控制
- 团队协作

#### 实现方案
```go
// 团队服务
type TeamService struct {
    db     *gorm.DB
    gitlab *gitlab.Client
}

// 团队模型
type Team struct {
    ID          uint      `gorm:"primaryKey"`
    Name        string    `gorm:"not null"`
    Description string
    LeaderID    uint      `gorm:"not null"`
    Members     []User    `gorm:"many2many:team_members"`
    Projects    []Project `gorm:"many2many:team_projects"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// 团队权限
type TeamPermission struct {
    ID          uint   `gorm:"primaryKey"`
    TeamID      uint   `gorm:"not null"`
    UserID      uint   `gorm:"not null"`
    Permission  string `gorm:"not null"`
}
```

## GitLab API集成方案

### 1. 认证与授权

#### OAuth2.0集成
```go
// OAuth配置
type OAuthConfig struct {
    ClientID     string
    ClientSecret string
    RedirectURL  string
    Scopes       []string
}

// 用户认证服务
type AuthService struct {
    config *OAuthConfig
    gitlab *gitlab.Client
}

// 处理OAuth回调
func (s *AuthService) HandleCallback(code string) (*User, error) {
    // 获取访问令牌
    token, err := s.exchangeCodeForToken(code)
    if err != nil {
        return nil, err
    }
    
    // 获取用户信息
    gitlabUser, _, err := s.gitlab.Users.CurrentUser()
    if err != nil {
        return nil, err
    }
    
    // 同步到本地数据库
    user := &User{
        GitLabID: gitlabUser.ID,
        Username: gitlabUser.Username,
        Email:    gitlabUser.Email,
        Name:     gitlabUser.Name,
        Avatar:   gitlabUser.AvatarURL,
    }
    
    return s.createOrUpdateUser(user)
}
```

### 2. Webhook集成

#### Webhook处理器
```go
// Webhook服务
type WebhookService struct {
    db            *gorm.DB
    notification  *NotificationService
    collaboration *CollaborationService
}

// 处理Push事件
func (s *WebhookService) HandlePushEvent(event *gitlab.PushEvent) error {
    // 更新本地项目信息
    project, err := s.db.Where("gitlab_id = ?", event.ProjectID).First(&Project{})
    if err != nil {
        return err
    }
    
    // 通知相关用户
    return s.notification.NotifyProjectUpdate(project.ID, event)
}

// 处理Wiki页面事件
func (s *WebhookService) HandleWikiPageEvent(event *gitlab.WikiPageEvent) error {
    // 同步Wiki页面变更
    doc := &Document{
        ProjectID: event.Project.ID,
        Slug:      event.ObjectAttributes.Slug,
        Title:     event.ObjectAttributes.Title,
        Content:   event.ObjectAttributes.Content,
    }
    
    return s.syncDocumentChanges(doc)
}
```

## 数据库设计

### 核心表结构

```sql
-- 用户表
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    gitlab_id INTEGER UNIQUE NOT NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    role INTEGER NOT NULL DEFAULT 0,
    avatar VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 文档表
CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL,
    slug VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    format VARCHAR(50) DEFAULT 'markdown',
    category_id INTEGER,
    created_by INTEGER REFERENCES users(id),
    updated_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, slug)
);

-- 话题表
CREATE TABLE topics (
    id SERIAL PRIMARY KEY,
    type INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    created_by INTEGER REFERENCES users(id),
    assigned_to JSON,
    status VARCHAR(50) DEFAULT 'active',
    due_date TIMESTAMP,
    project_id INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 项目表
CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    gitlab_id INTEGER UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 团队表
CREATE TABLE teams (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    leader_id INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 前端架构设计

### Vue应用结构

```
src/
├── components/           # 通用组件
│   ├── Editor/          # 编辑器组件
│   ├── FileTree/        # 文件树组件
│   ├── Chat/            # 聊天组件
│   └── Common/          # 通用组件
├── views/               # 页面视图
│   ├── Dashboard/       # 仪表板
│   ├── Documents/       # 文档管理
│   ├── Topics/          # 话题管理
│   ├── Projects/        # 项目管理
│   └── Teams/           # 团队管理
├── stores/              # 状态管理
│   ├── user.js          # 用户状态
│   ├── document.js      # 文档状态
│   └── collaboration.js # 协作状态
├── services/            # API服务
│   ├── api.js           # API配置
│   ├── gitlab.js        # GitLab API
│   └── websocket.js     # WebSocket服务
└── utils/               # 工具函数
    ├── auth.js          # 认证工具
    ├── editor.js        # 编辑器工具
    └── format.js        # 格式化工具
```

### 关键组件实现

#### 1. OnlyOffice文档编辑器组件
```vue
<template>
  <div class="document-editor-container">
    <div class="editor-header">
      <h3>{{ documentTitle }}</h3>
      <div class="editor-actions">
        <el-button @click="saveDocument">保存</el-button>
        <el-button @click="shareDocument">分享</el-button>
        <el-dropdown @command="handleCommand">
          <el-button>
            更多操作<i class="el-icon-arrow-down el-icon--right"></i>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="export">导出</el-dropdown-item>
              <el-dropdown-item command="history">版本历史</el-dropdown-item>
              <el-dropdown-item command="permissions">权限设置</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>
    <div ref="onlyOfficeRef" class="onlyoffice-editor"></div>
    <div class="collaboration-status">
      <div class="online-users">
        <div 
          v-for="user in onlineUsers" 
          :key="user.id"
          class="user-avatar"
          :title="user.name"
        >
          <img :src="user.avatar" :alt="user.name">
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useDocumentStore } from '@/stores/document'
import { ElMessage } from 'element-plus'

const route = useRoute()
const documentStore = useDocumentStore()

const onlyOfficeRef = ref(null)
const documentTitle = ref('')
const onlineUsers = ref([])
let docEditor = null

const props = defineProps({
  documentId: {
    type: [String, Number],
    required: true
  },
  mode: {
    type: String,
    default: 'edit', // edit, view, comment
    validator: (value) => ['edit', 'view', 'comment'].includes(value)
  }
})

const emit = defineEmits(['document-saved', 'document-error'])

onMounted(async () => {
  await initializeOnlyOffice()
})

const initializeOnlyOffice = async () => {
  try {
    // 获取文档编辑配置
    const config = await documentStore.getEditorConfig(props.documentId, props.mode)
    
    documentTitle.value = config.document.title
    
    // 初始化OnlyOffice编辑器
    docEditor = new DocsAPI.DocEditor(onlyOfficeRef.value, {
      documentType: config.documentType,
      document: config.document,
      editorConfig: {
        ...config.editor,
        events: {
          onAppReady: onAppReady,
          onDocumentStateChange: onDocumentStateChange,
          onRequestSaveAs: onRequestSaveAs,
          onRequestInsertImage: onRequestInsertImage,
          onRequestMailMergeRecipients: onRequestMailMergeRecipients,
          onRequestCompareFile: onRequestCompareFile,
          onRequestEditRights: onRequestEditRights,
          onRequestHistory: onRequestHistory,
          onRequestHistoryClose: onRequestHistoryClose,
          onRequestHistoryData: onRequestHistoryData,
          onRequestRestore: onRequestRestore,
          onError: onError
        }
      },
      user: config.user,
      token: config.token
    })
  } catch (error) {
    ElMessage.error('加载文档编辑器失败')
    console.error('OnlyOffice initialization error:', error)
  }
}

// OnlyOffice事件处理
const onAppReady = () => {
  console.log('OnlyOffice应用已就绪')
  loadOnlineUsers()
}

const onDocumentStateChange = (event) => {
  if (event.data) {
    console.log('文档状态已更改')
  }
}

const onRequestSaveAs = (event) => {
  const title = event.data.title
  const url = event.data.url
  // 处理另存为请求
  documentStore.saveAsDocument(props.documentId, title, url)
}

const onRequestInsertImage = (event) => {
  // 处理插入图片请求
  const images = [
    {
      url: 'https://example.com/image1.jpg',
      fileType: 'jpg',
      token: 'image_token'
    }
  ]
  docEditor.insertImage(images)
}

const onRequestEditRights = () => {
  // 处理编辑权限请求
  console.log('用户请求编辑权限')
}

const onRequestHistory = () => {
  // 处理历史版本请求
  documentStore.getDocumentHistory(props.documentId)
    .then(history => {
      docEditor.refreshHistory(history)
    })
}

const onRequestHistoryClose = () => {
  // 关闭历史版本
  docEditor.refreshHistory({
    currentVersion: 1,
    history: []
  })
}

const onRequestHistoryData = (event) => {
  // 请求历史数据
  const version = event.data.version
  documentStore.getHistoryData(props.documentId, version)
    .then(data => {
      docEditor.setHistoryData(data)
    })
}

const onRequestRestore = (event) => {
  // 恢复到指定版本
  const version = event.data.version
  documentStore.restoreVersion(props.documentId, version)
    .then(() => {
      ElMessage.success('文档已恢复到指定版本')
    })
}

const onError = (event) => {
  console.error('OnlyOffice错误:', event.data)
  ElMessage.error('文档编辑器出现错误')
  emit('document-error', event.data)
}

// 工具栏操作
const saveDocument = () => {
  if (docEditor) {
    docEditor.downloadAs()
  }
}

const shareDocument = () => {
  // 分享文档
  documentStore.shareDocument(props.documentId)
}

const handleCommand = (command) => {
  switch (command) {
    case 'export':
      exportDocument()
      break
    case 'history':
      showHistory()
      break
    case 'permissions':
      showPermissions()
      break
  }
}

const exportDocument = () => {
  if (docEditor) {
    docEditor.downloadAs()
  }
}

const showHistory = () => {
  if (docEditor) {
    docEditor.showHistory()
  }
}

const showPermissions = () => {
  // 显示权限设置对话框
  documentStore.showPermissionsDialog(props.documentId)
}

const loadOnlineUsers = () => {
  // 加载在线用户
  documentStore.getOnlineUsers(props.documentId)
    .then(users => {
      onlineUsers.value = users
    })
}

onUnmounted(() => {
  if (docEditor) {
    docEditor.destroyEditor()
  }
})
</script>

<style scoped>
.document-editor-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 20px;
  background-color: #f5f5f5;
  border-bottom: 1px solid #e4e7ed;
}

.editor-header h3 {
  margin: 0;
  font-size: 18px;
  color: #303133;
}

.editor-actions {
  display: flex;
  gap: 10px;
}

.onlyoffice-editor {
  flex: 1;
  width: 100%;
  min-height: 600px;
}

.collaboration-status {
  position: fixed;
  top: 80px;
  right: 20px;
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  padding: 10px;
  z-index: 1000;
}

.online-users {
  display: flex;
  gap: 8px;
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  overflow: hidden;
  border: 2px solid #409eff;
}

.user-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
</style>
```

#### 2. 文件树组件
```vue
<template>
  <div class="file-tree">
    <el-tree
      :data="fileTree"
      :props="treeProps"
      @node-click="handleNodeClick"
      node-key="path"
    >
      <template #default="{ node, data }">
        <div class="file-node">
          <i :class="getFileIcon(data)"></i>
          <span>{{ data.name }}</span>
        </div>
      </template>
    </el-tree>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useProjectStore } from '@/stores/project'

const projectStore = useProjectStore()
const fileTree = ref([])

const treeProps = {
  children: 'children',
  label: 'name'
}

const handleNodeClick = (data) => {
  if (data.type === 'file') {
    // 打开文件
    projectStore.openFile(data.path)
  }
}

const getFileIcon = (data) => {
  if (data.type === 'folder') {
    return 'el-icon-folder'
  }
  
  const ext = data.name.split('.').pop()
  switch (ext) {
    case 'js':
    case 'ts':
      return 'el-icon-document'
    case 'vue':
      return 'el-icon-connection'
    case 'md':
      return 'el-icon-tickets'
    default:
      return 'el-icon-document'
  }
}

onMounted(() => {
  // 加载文件树
  projectStore.loadFileTree()
})
</script>
```

## 部署方案

### Docker容器化部署

#### 1. 后端服务Dockerfile
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/config ./config

EXPOSE 8080
CMD ["./main"]
```

#### 2. 前端Dockerfile
```dockerfile
FROM node:18-alpine AS builder

WORKDIR /app
COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

#### 3. Docker Compose配置
```yaml
version: '3.8'

services:
  # GitLab CE
  gitlab:
    image: gitlab/gitlab-ce:latest
    hostname: gitlab.local
    ports:
      - "80:80"
      - "443:443"
      - "22:22"
    volumes:
      - gitlab_config:/etc/gitlab
      - gitlab_logs:/var/log/gitlab
      - gitlab_data:/var/opt/gitlab
    environment:
      GITLAB_OMNIBUS_CONFIG: |
        external_url 'http://gitlab.local'
        gitlab_rails['gitlab_shell_ssh_port'] = 22
    networks:
      - community-network

  # OnlyOffice Document Server
  onlyoffice:
    image: onlyoffice/documentserver:latest
    stdin_open: true
    tty: true
    restart: always
    ports:
      - "8000:80"
    volumes:
      - onlyoffice_data:/var/www/onlyoffice/Data
      - onlyoffice_logs:/var/log/onlyoffice
      - onlyoffice_cache:/var/lib/onlyoffice/documentserver/App_Data/cache/files
      - onlyoffice_forgotten:/var/lib/onlyoffice/documentserver/App_Data/cache/forgotten
    environment:
      - JWT_ENABLED=true
      - JWT_SECRET=your-secret-key
      - JWT_HEADER=Authorization
      - JWT_IN_BODY=true
      - WOPI_ENABLED=false
      - USE_UNAUTHORIZED_STORAGE=false
    networks:
      - community-network

  # PostgreSQL数据库
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: community
      POSTGRES_USER: community
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - community-network

  # Redis缓存
  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    networks:
      - community-network

  # 后端服务
  backend:
    build: ./backend
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
      - gitlab
      - onlyoffice
    environment:
      DATABASE_URL: postgres://community:password@postgres:5432/community
      REDIS_URL: redis://redis:6379
      GITLAB_URL: http://gitlab
      ONLYOFFICE_URL: http://onlyoffice
      ONLYOFFICE_JWT_SECRET: your-secret-key
    networks:
      - community-network

  # 前端服务
  frontend:
    build: ./frontend
    ports:
      - "3000:80"
    depends_on:
      - backend
    networks:
      - community-network

  # Nginx反向代理
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - frontend
      - backend
    networks:
      - community-network

volumes:
  gitlab_config:
  gitlab_logs:
  gitlab_data:
  onlyoffice_data:
  onlyoffice_logs:
  onlyoffice_cache:
  onlyoffice_forgotten:
  postgres_data:
  redis_data:

networks:
  community-network:
    driver: bridge
```

### 部署脚本

#### 1. 初始化脚本
```bash
#!/bin/bash
# deploy.sh

echo "正在部署GitLab社区系统..."

# 创建必要的目录
mkdir -p logs ssl config

# 生成SSL证书（如果需要）
if [ ! -f ssl/cert.pem ]; then
    echo "生成SSL证书..."
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout ssl/key.pem -out ssl/cert.pem \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=Community/CN=localhost"
fi

# 构建和启动服务
echo "构建Docker镜像..."
docker-compose build

echo "启动服务..."
docker-compose up -d

echo "等待服务启动..."
sleep 30

# 检查服务状态
echo "检查服务状态..."
docker-compose ps

# 初始化数据库
echo "初始化数据库..."
docker-compose exec backend ./main migrate

echo "部署完成！"
echo "访问地址:"
echo "  前端: http://localhost:3000"
echo "  后端API: http://localhost:8080"
echo "  GitLab: http://localhost"
echo "  OnlyOffice: http://localhost:8000"

# 检查OnlyOffice健康状态
echo "检查OnlyOffice健康状态..."
curl -s http://localhost:8000/healthcheck || echo "OnlyOffice可能需要更多时间启动"
```

#### 2. 监控脚本
```bash
#!/bin/bash
# monitor.sh

echo "系统监控报告"
echo "===================="

# 检查容器状态
echo "容器状态:"
docker-compose ps

# 检查资源使用
echo "资源使用情况:"
docker stats --no-stream

# 检查日志
echo "最近的错误日志:"
docker-compose logs --tail=10 backend | grep ERROR
```

## 安全考虑

### 1. 认证与授权
- OAuth2.0集成
- JWT令牌管理
- 角色权限控制
- API访问限制

### 2. 数据安全
- 数据库连接加密
- 敏感数据脱敏
- 备份策略
- 访问日志记录

### 3. 网络安全
- HTTPS部署
- 防火墙配置
- 请求限制
- 恶意攻击防护

## 性能优化

### 1. 缓存策略
- Redis缓存常用数据
- 静态资源CDN
- 数据库查询优化
- 前端资源缓存

### 2. 并发处理
- WebSocket连接池
- 数据库连接池
- 异步任务队列
- 负载均衡

### 3. 监控与告警
- 系统性能监控
- 错误日志收集
- 用户行为分析
- 自动告警机制

## 扩展性设计

### 1. 微服务架构
- 服务模块化
- 独立部署
- 服务发现
- 配置管理

### 2. 插件系统
- 功能扩展接口
- 第三方集成
- 自定义工作流
- 主题定制

### 3. 多租户支持
- 数据隔离
- 权限隔离
- 资源配额
- 独立配置

## 项目实施计划

### 第一阶段（4周）
- 搭建测试环境 (数据库、gitlab、onlyoffice)
- 基础架构搭建
- 用户管理模块直接使用gitlab的体系
- GitLab API集成
- 基础文档管理

### 第二阶段（4周）
- 在线编辑器实现
- 实时协作功能
- WebSocket集成
- 文件管理系统（使用git lfs）

### 第三阶段（4周）
- 话题管理系统
- 作业管理功能
- 通知系统
- 团队管理

### 第四阶段（2周）
- 系统集成测试
- 性能优化
- 安全加固
- 部署上线

## 总结

本方案基于GitLab强大的版本控制和协作功能，结合现代Web技术栈，构建了一个功能完整的教育社区系统。通过GitLab API深度集成，实现了文档管理、代码开发、团队协作等核心功能，同时保证了系统的可扩展性和安全性。

系统采用容器化部署，支持快速扩展和维护，适合中小型教育机构使用。通过合理的架构设计和技术选型，确保了系统的稳定性和性能。 