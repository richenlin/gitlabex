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

#### 实现方案
```go
// 极简用户服务 - 只负责GitLab用户映射
type UserService struct {
    gitlab *gitlab.Client
    cache  *redis.Client
    db     *gorm.DB
}

// 极简用户模型 - 只存储必要的映射信息
type User struct {
    ID          uint      `gorm:"primaryKey"`
    GitLabID    int       `gorm:"unique;not null"`
    Username    string    `gorm:"unique;not null"`
    Email       string    `gorm:"unique;not null"`
    Name        string    `gorm:"not null"`
    Avatar      string
    LastSyncAt  time.Time
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// 教育角色映射 - 基于GitLab Group成员关系
type EducationRole int

const (
    RoleGuest     EducationRole = 10  // GitLab Guest -> 访客
    RoleStudent   EducationRole = 20  // GitLab Reporter -> 学生
    RoleAssistant EducationRole = 30  // GitLab Developer -> 助教
    RoleTeacher   EducationRole = 40  // GitLab Maintainer -> 教师
    RoleAdmin     EducationRole = 50  // GitLab Owner -> 管理员
)

// 从GitLab获取用户角色
func (s *UserService) GetUserRole(userID int, groupID int) (EducationRole, error) {
    member, _, err := s.gitlab.GroupMembers.GetGroupMember(groupID, userID)
    if err != nil {
        return RoleGuest, err
    }
    return s.mapGitLabAccessLevel(member.AccessLevel), nil
}

// GitLab权限映射
func (s *UserService) mapGitLabAccessLevel(level gitlab.AccessLevelValue) EducationRole {
    switch level {
    case gitlab.GuestPermissions:
        return RoleGuest
    case gitlab.ReporterPermissions:
        return RoleStudent
    case gitlab.DeveloperPermissions:
        return RoleAssistant
    case gitlab.MaintainerPermissions:
        return RoleTeacher
    case gitlab.OwnerPermissions:
        return RoleAdmin
    default:
        return RoleGuest
    }
}

// 同步用户信息（从GitLab API）
func (s *UserService) SyncUserFromGitLab(gitlabID int) (*User, error) {
    gitlabUser, _, err := s.gitlab.Users.GetUser(gitlabID)
    if err != nil {
        return nil, err
    }
    
    user := &User{
        GitLabID:   gitlabUser.ID,
        Username:   gitlabUser.Username,
        Email:      gitlabUser.Email,
        Name:       gitlabUser.Name,
        Avatar:     gitlabUser.AvatarURL,
        LastSyncAt: time.Now(),
    }
    
    // 保存到本地数据库（仅作为缓存）
    if err := s.db.Save(user).Error; err != nil {
        return nil, err
    }
    
    return user, nil
}
```

### 2. 团队管理模块 - 使用GitLab Group

#### 功能特性
- ✅ 直接使用GitLab Group作为班级/团队
- ✅ 支持多层级：学校 -> 学院 -> 班级 -> 项目组
- ✅ 成员管理通过GitLab Group Members API
- ✅ 权限管理使用GitLab原生权限系统

#### 实现方案
```go
// 团队服务 - 完全基于GitLab Group API
type TeamService struct {
    gitlab *gitlab.Client
    cache  *redis.Client
}

// 不需要自定义Team模型，直接使用GitLab Group

// 获取用户所属团队
func (s *TeamService) GetUserTeams(userID int) ([]*gitlab.Group, error) {
    groups, _, err := s.gitlab.Groups.ListGroups(&gitlab.ListGroupsOptions{
        AllAvailable: gitlab.Bool(true),
    })
    if err != nil {
        return nil, err
    }
    
    var userGroups []*gitlab.Group
    for _, group := range groups {
        if s.isUserInGroup(userID, group.ID) {
            userGroups = append(userGroups, group)
        }
    }
    
    return userGroups, nil
}

// 创建班级/团队
func (s *TeamService) CreateTeam(name, description string, parentID *int) (*gitlab.Group, error) {
    createOpts := &gitlab.CreateGroupOptions{
        Name:        gitlab.String(name),
        Path:        gitlab.String(strings.ToLower(strings.ReplaceAll(name, " ", "-"))),
        Description: gitlab.String(description),
        Visibility:  gitlab.Visibility(gitlab.PrivateVisibility),
    }
    
    if parentID != nil {
        createOpts.ParentID = gitlab.Int(*parentID)
    }
    
    group, _, err := s.gitlab.Groups.CreateGroup(createOpts)
    return group, err
}

// 添加团队成员
func (s *TeamService) AddTeamMember(groupID, userID int, role EducationRole) error {
    accessLevel := s.mapEducationRoleToGitLab(role)
    _, _, err := s.gitlab.GroupMembers.AddGroupMember(groupID, &gitlab.AddGroupMemberOptions{
        UserID:      gitlab.Int(userID),
        AccessLevel: gitlab.AccessLevel(accessLevel),
    })
    return err
}

// 教育角色映射到GitLab权限
func (s *TeamService) mapEducationRoleToGitLab(role EducationRole) gitlab.AccessLevelValue {
    switch role {
    case RoleStudent:
        return gitlab.ReporterPermissions
    case RoleAssistant:
        return gitlab.DeveloperPermissions
    case RoleTeacher:
        return gitlab.MaintainerPermissions
    case RoleAdmin:
        return gitlab.OwnerPermissions
    default:
        return gitlab.GuestPermissions
    }
}
```

### 3. 权限控制模块 - 基于GitLab权限模型

#### 功能特性
- ✅ 完全使用GitLab的5级权限系统
- ✅ 权限检查通过GitLab API实时获取
- ✅ 支持Group级别和Project级别权限
- ✅ 教育场景权限映射

#### 实现方案
```go
// 权限服务 - 基于GitLab权限模型
type PermissionService struct {
    gitlab *gitlab.Client
    cache  *redis.Client
}

// 权限检查
func (s *PermissionService) CheckPermission(userID int, resourceType string, resourceID int, action string) (bool, error) {
    switch resourceType {
    case "project":
        return s.checkProjectPermission(userID, resourceID, action)
    case "group":
        return s.checkGroupPermission(userID, resourceID, action)
    default:
        return false, fmt.Errorf("unsupported resource type: %s", resourceType)
    }
}

// 检查项目权限
func (s *PermissionService) checkProjectPermission(userID int, projectID int, action string) (bool, error) {
    // 从缓存获取权限
    cacheKey := fmt.Sprintf("perm:project:%d:user:%d", projectID, userID)
    if cached, err := s.cache.Get(cacheKey).Result(); err == nil {
        var level gitlab.AccessLevelValue
        if err := json.Unmarshal([]byte(cached), &level); err == nil {
            return s.hasPermissionForAction(level, action), nil
        }
    }
    
    // 从GitLab API获取权限
    member, _, err := s.gitlab.ProjectMembers.GetProjectMember(projectID, userID)
    if err != nil {
        return false, err
    }
    
    // 缓存权限信息
    levelBytes, _ := json.Marshal(member.AccessLevel)
    s.cache.Set(cacheKey, levelBytes, 5*time.Minute)
    
    return s.hasPermissionForAction(member.AccessLevel, action), nil
}

// 检查动作权限
func (s *PermissionService) hasPermissionForAction(level gitlab.AccessLevelValue, action string) bool {
    switch action {
    case "read":
        return level >= gitlab.GuestPermissions
    case "create_issue":
        return level >= gitlab.ReporterPermissions
    case "push_code":
        return level >= gitlab.DeveloperPermissions
    case "manage_project":
        return level >= gitlab.MaintainerPermissions
    case "delete_project":
        return level >= gitlab.OwnerPermissions
    default:
        return false
    }
}
```

### 4. 教育管理模块 - 基于GitLab Issues/Discussions

#### 功能特性
- ✅ 课题管理（GitLab Issues + 课题标签）
- ✅ 作业管理（GitLab Issues + 作业标签）  
- ✅ 话题讨论（GitLab Discussions）
- ✅ 公告发布（GitLab Issues + 公告标签）
- ✅ 作业提交（GitLab Merge Request）

#### 实现方案
```go
// 教育管理服务 - 基于GitLab Issues和Discussions
type EducationService struct {
    gitlab *gitlab.Client
    cache  *redis.Client
}

// 课题管理 - 使用GitLab Issues
func (s *EducationService) CreateProject(groupID int, title, description string, dueDate *time.Time) (*gitlab.Issue, error) {
    // 在Group下创建或获取课题项目
    project, err := s.getOrCreateEducationProject(groupID, "课题管理")
    if err != nil {
        return nil, err
    }
    
    // 创建课题Issue
    labels := []string{"课题", "project"}
    if dueDate != nil {
        labels = append(labels, "截止日期:"+dueDate.Format("2006-01-02"))
    }
    
    issue, _, err := s.gitlab.Issues.CreateIssue(project.ID, &gitlab.CreateIssueOptions{
        Title:       gitlab.String(title),
        Description: gitlab.String(description),
        Labels:      labels,
        DueDate:     (*gitlab.ISOTime)(dueDate),
    })
    
    return issue, err
}

// 作业管理 - 使用GitLab Issues
func (s *EducationService) CreateAssignment(groupID int, title, description string, dueDate *time.Time) (*gitlab.Issue, error) {
    // 在Group下创建或获取作业项目
    project, err := s.getOrCreateEducationProject(groupID, "作业管理")
    if err != nil {
        return nil, err
    }
    
    // 创建作业Issue
    labels := []string{"作业", "assignment"}
    if dueDate != nil {
        labels = append(labels, "截止日期:"+dueDate.Format("2006-01-02"))
    }
    
    issue, _, err := s.gitlab.Issues.CreateIssue(project.ID, &gitlab.CreateIssueOptions{
        Title:       gitlab.String(title),
        Description: gitlab.String(description),
        Labels:      labels,
        DueDate:     (*gitlab.ISOTime)(dueDate),
    })
    
    return issue, err
}

// 公告发布 - 使用GitLab Issues
func (s *EducationService) CreateAnnouncement(groupID int, title, content string) (*gitlab.Issue, error) {
    // 在Group下创建或获取公告项目
    project, err := s.getOrCreateEducationProject(groupID, "班级公告")
    if err != nil {
        return nil, err
    }
    
    // 创建公告Issue
    issue, _, err := s.gitlab.Issues.CreateIssue(project.ID, &gitlab.CreateIssueOptions{
        Title:       gitlab.String(title),
        Description: gitlab.String(content),
        Labels:      []string{"公告", "announcement"},
    })
    
    return issue, err
}

// 话题讨论 - 使用GitLab Discussions
func (s *EducationService) CreateDiscussion(projectID int, title, content string) (*gitlab.Discussion, error) {
    // 创建Issue作为讨论载体
    issue, _, err := s.gitlab.Issues.CreateIssue(projectID, &gitlab.CreateIssueOptions{
        Title:       gitlab.String(title),
        Description: gitlab.String(content),
        Labels:      []string{"讨论", "discussion"},
    })
    if err != nil {
        return nil, err
    }
    
    // 创建讨论
    discussion, _, err := s.gitlab.Discussions.CreateIssueDiscussion(projectID, issue.IID, &gitlab.CreateIssueDiscussionOptions{
        Body: gitlab.String(content),
    })
    
    return discussion, err
}

// 学生提交作业 - 使用GitLab Merge Request
func (s *EducationService) SubmitAssignment(projectID int, issueID int, studentID int, branchName string) (*gitlab.MergeRequest, error) {
    // 创建作业提交MR
    mr, _, err := s.gitlab.MergeRequests.CreateMergeRequest(projectID, &gitlab.CreateMergeRequestOptions{
        Title:        gitlab.String(fmt.Sprintf("作业提交 - Issue #%d", issueID)),
        Description:  gitlab.String(fmt.Sprintf("关联作业: #%d\n\n提交人: @%d", issueID, studentID)),
        SourceBranch: gitlab.String(branchName),
        TargetBranch: gitlab.String("main"),
        AssigneeID:   gitlab.Int(studentID),
        Labels:       []string{"作业提交", "assignment-submission"},
    })
    
    if err != nil {
        return nil, err
    }
    
    // 自动关联到作业Issue
    _, _, err = s.gitlab.Issues.CreateIssueNote(projectID, issueID, &gitlab.CreateIssueNoteOptions{
        Body: gitlab.String(fmt.Sprintf("学生提交作业: !%d", mr.IID)),
    })
    
    return mr, err
}

// 教师批改作业 - 使用GitLab MR Review
func (s *EducationService) GradeAssignment(projectID int, mrID int, grade float64, feedback string) error {
    // 添加批改评论
    _, _, err := s.gitlab.MergeRequestNotes.CreateMergeRequestNote(projectID, mrID, &gitlab.CreateMergeRequestNoteOptions{
        Body: gitlab.String(fmt.Sprintf("## 作业批改\n\n**成绩**: %.1f分\n\n**反馈**: %s", grade, feedback)),
    })
    if err != nil {
        return err
    }
    
    // 添加成绩标签
    gradeLabel := fmt.Sprintf("成绩:%.1f", grade)
    _, _, err = s.gitlab.MergeRequests.UpdateMergeRequest(projectID, mrID, &gitlab.UpdateMergeRequestOptions{
        Labels: []string{"作业提交", "assignment-submission", gradeLabel},
    })
    
    return err
}

// 获取或创建教育项目
func (s *EducationService) getOrCreateEducationProject(groupID int, projectName string) (*gitlab.Project, error) {
    // 先尝试获取现有项目
    projects, _, err := s.gitlab.Groups.ListGroupProjects(groupID, &gitlab.ListGroupProjectsOptions{
        Search: gitlab.String(projectName),
    })
    if err != nil {
        return nil, err
    }
    
    for _, project := range projects {
        if project.Name == projectName {
            return project, nil
        }
    }
    
    // 创建新项目
    project, _, err := s.gitlab.Projects.CreateProject(&gitlab.CreateProjectOptions{
        Name:        gitlab.String(projectName),
        NamespaceID: gitlab.Int(groupID),
        Description: gitlab.String("教育管理项目 - " + projectName),
        Visibility:  gitlab.Visibility(gitlab.PrivateVisibility),
        // 启用必要功能
        IssuesEnabled:    gitlab.Bool(true),
        MergeRequestsEnabled: gitlab.Bool(true),
        WikiEnabled:     gitlab.Bool(true),
    })
    
    return project, err
}
```

### 5. 文档管理模块 - GitLab Wiki + 文档附件 + OnlyOffice

#### 功能特性
- ✅ 基于GitLab Wiki的文档管理
- ✅ 支持文档附件上传（Word、Excel、PowerPoint等）
- ✅ 具有Wiki权限的成员可以使用OnlyOffice编辑文档附件
- ✅ 文档版本控制使用GitLab原生功能
- ✅ 文档权限完全基于GitLab项目Wiki权限

#### 实现方案
```go
// 文档服务 - GitLab Wiki + 文档附件 + OnlyOffice集成
type DocumentService struct {
    gitlab      *gitlab.Client
    onlyoffice  *OnlyOfficeService
    cache       *redis.Client
}

// 文档附件模型 - 只存储OnlyOffice编辑会话信息
type DocumentAttachment struct {
    ID            uint      `gorm:"primaryKey"`
    ProjectID     int       `gorm:"not null"`
    WikiPageSlug  string    `gorm:"not null"`       // 关联的Wiki页面
    FileName      string    `gorm:"not null"`       // 附件文件名
    FileURL       string    `gorm:"not null"`       // GitLab文件URL
    FileType      string    `gorm:"not null"`       // docx, xlsx, pptx
    OnlyOfficeKey string    `gorm:"unique"`         // OnlyOffice编辑密钥
    LastEditedBy  int       `gorm:"default:null"`   // 最后编辑用户
    LastEditedAt  *time.Time `gorm:"default:null"`  // 最后编辑时间
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

// 创建Wiki页面并上传文档附件
func (s *DocumentService) CreateWikiWithAttachment(projectID int, title, content string, attachmentFile []byte, fileName string) (*gitlab.WikiPage, *DocumentAttachment, error) {
    // 1. 创建GitLab Wiki页面
    wikiSlug := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
    wikiPage, _, err := s.gitlab.Wikis.CreateWikiPage(projectID, &gitlab.CreateWikiPageOptions{
        Title:   gitlab.String(title),
        Content: gitlab.String(content),
        Format:  gitlab.String("markdown"),
    })
    if err != nil {
        return nil, nil, err
    }
    
    // 2. 上传文档附件到GitLab
    uploadResult, _, err := s.gitlab.Projects.UploadFile(projectID, &gitlab.UploadFileOptions{
        Content:  attachmentFile,
        Filename: fileName,
    })
    if err != nil {
        return nil, nil, err
    }
    
    // 3. 更新Wiki页面，添加附件链接
    attachmentMD := fmt.Sprintf("\n\n## 文档附件\n\n- [%s](%s) ([在线编辑](/api/documents/edit/%s))", 
        fileName, uploadResult.URL, fileName)
    updatedContent := content + attachmentMD
    
    _, _, err = s.gitlab.Wikis.EditWikiPage(projectID, wikiSlug, &gitlab.EditWikiPageOptions{
        Content: gitlab.String(updatedContent),
        Format:  gitlab.String("markdown"),
    })
    if err != nil {
        return nil, nil, err
    }
    
    // 4. 创建文档附件记录
    attachment := &DocumentAttachment{
        ProjectID:     projectID,
        WikiPageSlug:  wikiSlug,
        FileName:      fileName,
        FileURL:       uploadResult.URL,
        FileType:      s.getFileType(fileName),
        OnlyOfficeKey: s.generateOnlyOfficeKey(projectID, wikiSlug, fileName),
    }
    
    if err := s.db.Create(attachment).Error; err != nil {
        return nil, nil, err
    }
    
    return wikiPage, attachment, nil
}

// 检查用户是否有Wiki编辑权限
func (s *DocumentService) CheckWikiEditPermission(userID int, projectID int) (bool, error) {
    // 从GitLab API获取用户在项目中的权限
    member, _, err := s.gitlab.ProjectMembers.GetProjectMember(projectID, userID)
    if err != nil {
        return false, err
    }
    
    // Wiki编辑需要Developer及以上权限
    return member.AccessLevel >= gitlab.DeveloperPermissions, nil
}

// 启动OnlyOffice编辑会话
func (s *DocumentService) StartOnlyOfficeSession(attachmentID uint, userID int) (*OnlyOfficeConfig, error) {
    // 1. 获取文档附件信息
    attachment, err := s.getAttachment(attachmentID)
    if err != nil {
        return nil, err
    }
    
    // 2. 检查用户Wiki权限
    hasPermission, err := s.CheckWikiEditPermission(userID, attachment.ProjectID)
    if err != nil {
        return nil, err
    }
    
    if !hasPermission {
        return nil, fmt.Errorf("用户没有Wiki编辑权限")
    }
    
    // 3. 从GitLab下载最新文档内容
    fileContent, err := s.downloadFileFromGitLab(attachment.FileURL)
    if err != nil {
        return nil, err
    }
    
    // 4. 生成OnlyOffice配置
    user, err := s.getUser(userID)
    if err != nil {
        return nil, err
    }
    
    config := &OnlyOfficeConfig{
        DocumentType: s.getDocumentType(attachment.FileType),
        Document: OnlyOfficeDocument{
            Key:      attachment.OnlyOfficeKey,
            Title:    attachment.FileName,
            URL:      s.generateTempDocumentURL(attachment.ID),
            FileType: attachment.FileType,
            Permissions: OnlyOfficePermissions{
                Edit:     true,
                Comment:  true,
                Download: true,
                Print:    true,
            },
        },
        EditorConfig: OnlyOfficeEditor{
            Mode:        "edit",
            CallbackURL: s.getCallbackURL(attachment.ID),
            User: OnlyOfficeUser{
                ID:   fmt.Sprintf("%d", user.ID),
                Name: user.Name,
            },
        },
        Token: s.generateJWT(attachment.OnlyOfficeKey),
    }
    
    // 5. 更新编辑记录
    attachment.LastEditedBy = userID
    now := time.Now()
    attachment.LastEditedAt = &now
    s.db.Save(attachment)
    
    return config, nil
}

// 处理OnlyOffice保存回调
func (s *DocumentService) HandleOnlyOfficeCallback(attachmentID uint, callback *OnlyOfficeCallback) error {
    if callback.Status != 2 { // 只处理文档保存状态
        return nil
    }
    
    attachment, err := s.getAttachment(attachmentID)
    if err != nil {
        return err
    }
    
    // 1. 从OnlyOffice下载更新的文档
    updatedContent, err := s.downloadDocumentFromOnlyOffice(callback.URL)
    if err != nil {
        return err
    }
    
    // 2. 重新上传到GitLab（覆盖原文件）
    uploadResult, _, err := s.gitlab.Projects.UploadFile(attachment.ProjectID, &gitlab.UploadFileOptions{
        Content:  updatedContent,
        Filename: attachment.FileName,
    })
    if err != nil {
        return err
    }
    
    // 3. 更新附件记录
    attachment.FileURL = uploadResult.URL
    now := time.Now()
    attachment.LastEditedAt = &now
    
    return s.db.Save(attachment).Error
}

// 获取Wiki页面的所有可编辑附件
func (s *DocumentService) GetWikiEditableAttachments(projectID int, wikiSlug string) ([]*DocumentAttachment, error) {
    var attachments []*DocumentAttachment
    err := s.db.Where("project_id = ? AND wiki_page_slug = ? AND file_type IN (?)", 
        projectID, wikiSlug, []string{"docx", "xlsx", "pptx"}).Find(&attachments).Error
    return attachments, err
}

// 获取项目所有文档附件列表
func (s *DocumentService) GetProjectDocuments(projectID int, userID int) ([]*DocumentSummary, error) {
    // 检查用户权限
    hasPermission, err := s.CheckWikiEditPermission(userID, projectID)
    if err != nil {
        return nil, err
    }
    
    // 获取Wiki页面列表
    wikiPages, _, err := s.gitlab.Wikis.ListWikiPages(projectID, &gitlab.ListWikiPagesOptions{})
    if err != nil {
        return nil, err
    }
    
    var documents []*DocumentSummary
    for _, page := range wikiPages {
        // 获取每个Wiki页面的附件
        attachments, err := s.GetWikiEditableAttachments(projectID, page.Slug)
        if err != nil {
            continue
        }
        
        for _, attachment := range attachments {
            documents = append(documents, &DocumentSummary{
                ID:           attachment.ID,
                Title:        page.Title,
                FileName:     attachment.FileName,
                FileType:     attachment.FileType,
                WikiSlug:     page.Slug,
                LastEditedBy: attachment.LastEditedBy,
                LastEditedAt: attachment.LastEditedAt,
                CanEdit:      hasPermission,
            })
        }
    }
    
    return documents, nil
}

// 文档摘要信息
type DocumentSummary struct {
    ID           uint       `json:"id"`
    Title        string     `json:"title"`
    FileName     string     `json:"file_name"`
    FileType     string     `json:"file_type"`
    WikiSlug     string     `json:"wiki_slug"`
    LastEditedBy int        `json:"last_edited_by"`
    LastEditedAt *time.Time `json:"last_edited_at"`
    CanEdit      bool       `json:"can_edit"`
}

// OnlyOffice配置结构
type OnlyOfficeConfig struct {
    DocumentType string              `json:"documentType"`
    Document     OnlyOfficeDocument  `json:"document"`
    EditorConfig OnlyOfficeEditor    `json:"editorConfig"`
    Token        string              `json:"token"`
}

type OnlyOfficeDocument struct {
    Key         string                 `json:"key"`
    Title       string                 `json:"title"`
    URL         string                 `json:"url"`
    FileType    string                 `json:"fileType"`
    Permissions OnlyOfficePermissions  `json:"permissions"`
}

type OnlyOfficePermissions struct {
    Edit     bool `json:"edit"`
    Comment  bool `json:"comment"`
    Download bool `json:"download"`
    Print    bool `json:"print"`
}

type OnlyOfficeEditor struct {
    Mode        string          `json:"mode"`
    CallbackURL string          `json:"callbackUrl"`
    User        OnlyOfficeUser  `json:"user"`
}

type OnlyOfficeUser struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type OnlyOfficeCallback struct {
    Key    string `json:"key"`
    Status int    `json:"status"`
    URL    string `json:"url"`
}

// 工具方法
func (s *DocumentService) getFileType(fileName string) string {
    ext := strings.ToLower(filepath.Ext(fileName))
    switch ext {
    case ".docx", ".doc":
        return "docx"
    case ".xlsx", ".xls":
        return "xlsx"
    case ".pptx", ".ppt":
        return "pptx"
    default:
        return "unknown"
    }
}

func (s *DocumentService) getDocumentType(fileType string) string {
    switch fileType {
    case "docx":
        return "text"
    case "xlsx":
        return "spreadsheet"
    case "pptx":
        return "presentation"
    default:
        return "text"
    }
}

func (s *DocumentService) generateOnlyOfficeKey(projectID int, wikiSlug, fileName string) string {
    data := fmt.Sprintf("%d-%s-%s-%d", projectID, wikiSlug, fileName, time.Now().Unix())
    hasher := sha256.New()
    hasher.Write([]byte(data))
    return hex.EncodeToString(hasher.Sum(nil))[:32]
}

func (s *DocumentService) generateTempDocumentURL(attachmentID uint) string {
    return fmt.Sprintf("/api/documents/%d/download", attachmentID)
}

func (s *DocumentService) getCallbackURL(attachmentID uint) string {
    return fmt.Sprintf("/api/documents/%d/callback", attachmentID)
}
```

## 数据库设计 - 极简化

```sql
-- 用户表 - 只存储GitLab用户映射
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    gitlab_id INTEGER UNIQUE NOT NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar VARCHAR(255),
    last_sync_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 文档附件表 - 只存储OnlyOffice编辑会话信息
CREATE TABLE document_attachments (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL,
    wiki_page_slug VARCHAR(255) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_url VARCHAR(500) NOT NULL,
    file_type VARCHAR(50) NOT NULL,
    onlyoffice_key VARCHAR(255) UNIQUE NOT NULL,
    last_edited_by INTEGER REFERENCES users(id),
    last_edited_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, wiki_page_slug, file_name)
);

-- 删除原有的复杂表结构
-- 不再需要：teams, projects, permissions, roles, topics, assignments 等表
-- 所有这些信息都从GitLab API获取或使用GitLab原生功能实现
```

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

### 关键组件实现

#### 1. GitLab数据聚合组件
```vue
<template>
  <div class="education-dashboard">
    <el-row :gutter="20">
      <el-col :span="8">
        <el-card title="我的班级" class="dashboard-card">
          <group-list 
            :groups="userGroups" 
            @select="handleGroupSelect"
            education-mode
          />
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card title="我的项目" class="dashboard-card">
          <project-list 
            :projects="userProjects" 
            @select="handleProjectSelect"
            education-mode
          />
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card title="待办事项" class="dashboard-card">
          <todo-list 
            :issues="assignedIssues"
            :merge-requests="assignedMRs"
            education-mode
          />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useGitLabStore } from '@/stores/gitlab'
import { useEducationStore } from '@/stores/education'

const gitlabStore = useGitLabStore()
const educationStore = useEducationStore()

const userGroups = ref([])
const userProjects = ref([])
const assignedIssues = ref([])
const assignedMRs = ref([])

onMounted(async () => {
  // 加载用户数据（直接从GitLab API）
  await Promise.all([
    loadUserGroups(),
    loadUserProjects(),
    loadAssignedItems()
  ])
})

const loadUserGroups = async () => {
  userGroups.value = await gitlabStore.getUserGroups()
}

const loadUserProjects = async () => {
  userProjects.value = await gitlabStore.getUserProjects()
}

const loadAssignedItems = async () => {
  const [issues, mrs] = await Promise.all([
    gitlabStore.getAssignedIssues(),
    gitlabStore.getAssignedMergeRequests()
  ])
  assignedIssues.value = issues
  assignedMRs.value = mrs
}
</script>
```

#### 2. OnlyOffice文档编辑器集成
```vue
<template>
  <div class="onlyoffice-education-editor">
    <div class="editor-header">
      <breadcrumb :items="breadcrumbItems" />
      <div class="editor-actions">
        <el-button-group>
          <el-button @click="shareDocument">
            <i class="el-icon-share"></i> 分享
          </el-button>
          <el-button @click="viewHistory">
            <i class="el-icon-time"></i> 历史版本
          </el-button>
          <el-button @click="exportDocument">
            <i class="el-icon-download"></i> 导出
          </el-button>
        </el-button-group>
      </div>
    </div>
    
    <div class="editor-content">
      <div ref="editorContainer" class="onlyoffice-container"></div>
      
      <div class="editor-sidebar">
        <collaboration-panel 
          :online-users="onlineUsers"
          :document-id="documentId"
        />
        <gitlab-info-panel 
          :project-id="projectId"
          :wiki-page="wikiPage"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useOnlyOfficeStore } from '@/stores/onlyoffice'
import { useGitLabStore } from '@/stores/gitlab'

const onlyOfficeStore = useOnlyOfficeStore()
const gitlabStore = useGitLabStore()

const props = defineProps({
  documentId: {
    type: [String, Number],
    required: true
  },
  projectId: {
    type: Number,
    required: true
  },
  wikiSlug: {
    type: String,
    required: true
  }
})

const editorContainer = ref(null)
const onlineUsers = ref([])
const wikiPage = ref(null)
let docEditor = null

onMounted(async () => {
  await initializeEditor()
  await loadWikiPageInfo()
})

const initializeEditor = async () => {
  // 获取OnlyOffice配置
  const config = await onlyOfficeStore.getEditorConfig(props.documentId)
  
  // 初始化编辑器
  docEditor = new DocsAPI.DocEditor(editorContainer.value, {
    ...config,
    events: {
      onAppReady: () => {
        console.log('OnlyOffice 编辑器就绪')
      },
      onCollaborativeChanges: (changes) => {
        // 处理协作变更
        handleCollaborativeChanges(changes)
      },
      onDocumentReady: () => {
        // 文档加载完成，同步GitLab Wiki信息
        syncWithGitLabWiki()
      }
    }
  })
}

const loadWikiPageInfo = async () => {
  // 从GitLab API获取Wiki页面信息
  wikiPage.value = await gitlabStore.getWikiPage(props.projectId, props.wikiSlug)
}

const shareDocument = async () => {
  // 使用GitLab的分享机制
  const shareUrl = await gitlabStore.getProjectShareUrl(props.projectId)
  // 显示分享对话框
  showShareDialog(shareUrl)
}

const viewHistory = async () => {
  // 查看GitLab Wiki的版本历史
  const history = await gitlabStore.getWikiPageHistory(props.projectId, props.wikiSlug)
  // 显示历史版本
  showHistoryDialog(history)
}
</script>
```

## 核心服务设计

### 1. GitLab服务封装
```go
// GitLab服务 - 所有GitLab API的统一封装
type GitLabService struct {
    client *gitlab.Client
    cache  *redis.Client
    config *GitLabConfig
}

// GitLab配置
type GitLabConfig struct {
    BaseURL      string `json:"base_url"`
    Token        string `json:"token"`
    ClientID     string `json:"client_id"`
    ClientSecret string `json:"client_secret"`
}

// 教育场景的GitLab API封装
func (s *GitLabService) GetEducationDashboard(userID int) (*EducationDashboard, error) {
    // 并发获取用户相关数据
    var (
        groups   []*gitlab.Group
        projects []*gitlab.Project
        issues   []*gitlab.Issue
        mrs      []*gitlab.MergeRequest
        wg       sync.WaitGroup
        mu       sync.Mutex
        errs     []error
    )
    
    wg.Add(4)
    
    // 获取用户组
    go func() {
        defer wg.Done()
        if userGroups, _, err := s.client.Groups.ListGroups(&gitlab.ListGroupsOptions{
            MinAccessLevel: gitlab.AccessLevel(gitlab.ReporterPermissions),
        }); err != nil {
            mu.Lock()
            errs = append(errs, err)
            mu.Unlock()
        } else {
            groups = userGroups
        }
    }()
    
    // 获取用户项目
    go func() {
        defer wg.Done()
        if userProjects, _, err := s.client.Projects.ListUserProjects(userID, &gitlab.ListProjectsOptions{
            MinAccessLevel: gitlab.AccessLevel(gitlab.ReporterPermissions),
        }); err != nil {
            mu.Lock()
            errs = append(errs, err)
            mu.Unlock()
        } else {
            projects = userProjects
        }
    }()
    
    // 获取分配的Issues
    go func() {
        defer wg.Done()
        if assignedIssues, _, err := s.client.Issues.ListIssues(&gitlab.ListIssuesOptions{
            AssigneeID: gitlab.Int(userID),
            State:      gitlab.String("opened"),
        }); err != nil {
            mu.Lock()
            errs = append(errs, err)
            mu.Unlock()
        } else {
            issues = assignedIssues
        }
    }()
    
    // 获取分配的MR
    go func() {
        defer wg.Done()
        if assignedMRs, _, err := s.client.MergeRequests.ListMergeRequests(&gitlab.ListMergeRequestsOptions{
            AssigneeID: gitlab.Int(userID),
            State:      gitlab.String("opened"),
        }); err != nil {
            mu.Lock()
            errs = append(errs, err)
            mu.Unlock()
        } else {
            mrs = assignedMRs
        }
    }()
    
    wg.Wait()
    
    if len(errs) > 0 {
        return nil, fmt.Errorf("failed to load dashboard data: %v", errs)
    }
    
    return &EducationDashboard{
        Groups:        groups,
        Projects:      projects,
        AssignedIssues: issues,
        AssignedMRs:   mrs,
    }, nil
}

// 教育仪表板数据结构
type EducationDashboard struct {
    Groups         []*gitlab.Group        `json:"groups"`
    Projects       []*gitlab.Project      `json:"projects"`
    AssignedIssues []*gitlab.Issue        `json:"assigned_issues"`
    AssignedMRs    []*gitlab.MergeRequest `json:"assigned_mrs"`
}
```

### 2. 教育服务 - 业务逻辑封装
```go
// 教育服务 - 教育场景的业务逻辑
type EducationService struct {
    gitlab     *GitLabService
    onlyoffice *OnlyOfficeService
    cache      *redis.Client
}

// 创建班级（GitLab Group）
func (s *EducationService) CreateClass(name, description string, teacherID int) (*gitlab.Group, error) {
    // 创建GitLab Group
    group, err := s.gitlab.CreateGroup(&gitlab.CreateGroupOptions{
        Name:        gitlab.String(name),
        Path:        gitlab.String(s.generateClassPath(name)),
        Description: gitlab.String(description),
        Visibility:  gitlab.Visibility(gitlab.PrivateVisibility),
    })
    if err != nil {
        return nil, err
    }
    
    // 设置教师为Group Owner
    _, _, err = s.gitlab.client.GroupMembers.AddGroupMember(group.ID, &gitlab.AddGroupMemberOptions{
        UserID:      gitlab.Int(teacherID),
        AccessLevel: gitlab.AccessLevel(gitlab.OwnerPermissions),
    })
    if err != nil {
        return nil, err
    }
    
    // 初始化班级模板
    if err := s.initializeClassTemplate(group.ID); err != nil {
        return nil, err
    }
    
    return group, nil
}

// 布置作业（GitLab Issue）
func (s *EducationService) CreateAssignment(classGroupID int, title, description string, dueDate *time.Time) (*gitlab.Issue, error) {
    // 在班级Group下创建作业项目（如果不存在）
    assignmentProject, err := s.getOrCreateAssignmentProject(classGroupID)
    if err != nil {
        return nil, err
    }
    
    // 创建作业Issue
    labels := []string{"assignment", "homework"}
    if dueDate != nil {
        labels = append(labels, "due:"+dueDate.Format("2006-01-02"))
    }
    
    issue, _, err := s.gitlab.client.Issues.CreateIssue(assignmentProject.ID, &gitlab.CreateIssueOptions{
        Title:       gitlab.String(title),
        Description: gitlab.String(description),
        Labels:      labels,
        DueDate:     (*gitlab.ISOTime)(dueDate),
    })
    if err != nil {
        return nil, err
    }
    
    // 通知班级成员
    if err := s.notifyClassMembers(classGroupID, issue); err != nil {
        log.Printf("Failed to notify class members: %v", err)
    }
    
    return issue, nil
}

// 学生提交作业（GitLab MR）
func (s *EducationService) SubmitAssignment(projectID int, issueID int, studentID int) (*gitlab.MergeRequest, error) {
    // 为学生创建作业分支
    branchName := fmt.Sprintf("assignment-%d-student-%d", issueID, studentID)
    
    // 创建分支
    _, _, err := s.gitlab.client.Branches.CreateBranch(projectID, &gitlab.CreateBranchOptions{
        Branch: gitlab.String(branchName),
        Ref:    gitlab.String("main"),
    })
    if err != nil {
        return nil, err
    }
    
    // 创建Merge Request
    mr, _, err := s.gitlab.client.MergeRequests.CreateMergeRequest(projectID, &gitlab.CreateMergeRequestOptions{
        Title:        gitlab.String(fmt.Sprintf("作业提交 - Issue #%d", issueID)),
        Description:  gitlab.String(fmt.Sprintf("关联作业: #%d\n\n提交人: Student %d", issueID, studentID)),
        SourceBranch: gitlab.String(branchName),
        TargetBranch: gitlab.String("main"),
        AssigneeID:   gitlab.Int(studentID),
        Labels:       []string{"assignment-submission"},
    })
    if err != nil {
        return nil, err
    }
    
    return mr, nil
}
```

## 部署配置 - 简化版

### Docker Compose - 专注核心服务
```yaml
version: '3.8'

services:
  # GitLab CE - 核心服务
  gitlab:
    image: gitlab/gitlab-ce:latest
    hostname: gitlab.local
    ports:
      - "8081:80"
      - "4431:443"
      - "2222:22"
    volumes:
      - gitlab_config:/etc/gitlab
      - gitlab_logs:/var/log/gitlab
      - gitlab_data:/var/opt/gitlab
    environment:
      GITLAB_OMNIBUS_CONFIG: |
        external_url 'http://localhost'
        gitlab_rails['gitlab_shell_ssh_port'] = 2222
        # 启用必要的功能
        gitlab_rails['gitlab_default_projects_features_wiki'] = true
        gitlab_rails['gitlab_default_projects_features_issues'] = true
        gitlab_rails['gitlab_default_projects_features_merge_requests'] = true
    networks:
      - gitlabex-network

  # OnlyOffice Document Server - 文档协作核心
  onlyoffice:
    image: onlyoffice/documentserver:latest
    ports:
      - "8000:80"
    volumes:
      - onlyoffice_data:/var/www/onlyoffice/Data
      - onlyoffice_logs:/var/log/onlyoffice
    environment:
      - JWT_ENABLED=true
      - JWT_SECRET=gitlabex-jwt-secret-2024
      - JWT_HEADER=Authorization
      - JWT_IN_BODY=true
    networks:
      - gitlabex-network

  # PostgreSQL - 极简业务数据存储
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: gitlabex
      POSTGRES_USER: gitlabex
      POSTGRES_PASSWORD: password123
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    networks:
      - gitlabex-network

  # Redis - GitLab API数据缓存
  redis:
    image: redis:7-alpine
    command: redis-server --requirepass password123
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    networks:
      - gitlabex-network

  # 轻量级后端服务
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
      # 数据库配置
      DATABASE_URL: postgres://gitlabex:password123@postgres:5432/gitlabex
      REDIS_URL: redis://redis:6379
      REDIS_PASSWORD: password123
      
      # GitLab集成配置
      GITLAB_URL: http://gitlab
      GITLAB_CLIENT_ID: ${GITLAB_CLIENT_ID}
      GITLAB_CLIENT_SECRET: ${GITLAB_CLIENT_SECRET}
      
      # OnlyOffice集成配置
      ONLYOFFICE_URL: http://onlyoffice
      ONLYOFFICE_JWT_SECRET: gitlabex-jwt-secret-2024
      
      # 应用配置
      SERVER_PORT: 8080
      JWT_SECRET: gitlabex-app-secret-2024
    networks:
      - gitlabex-network

  # 前端服务
  frontend:
    build: ./frontend
    ports:
      - "3000:80"
    depends_on:
      - backend
    environment:
      - VUE_APP_API_BASE_URL=http://localhost:8080
      - VUE_APP_GITLAB_URL=http://localhost
      - VUE_APP_ONLYOFFICE_URL=http://localhost:8000
    networks:
      - gitlabex-network

volumes:
  gitlab_config:
  gitlab_logs:
  gitlab_data:
  onlyoffice_data:
  onlyoffice_logs:
  postgres_data:
  redis_data:

networks:
  gitlabex-network:
    driver: bridge
```

## 性能优化 - 专注核心

### 1. GitLab API缓存策略
```go
// GitLab API缓存服务
type GitLabCacheService struct {
    redis  *redis.Client
    gitlab *gitlab.Client
}

// 缓存用户权限
func (s *GitLabCacheService) GetUserPermission(userID int, projectID int) (gitlab.AccessLevelValue, error) {
    cacheKey := fmt.Sprintf("gitlab:perm:user:%d:project:%d", userID, projectID)
    
    // 尝试从缓存获取
    if cached, err := s.redis.Get(cacheKey).Result(); err == nil {
        var level gitlab.AccessLevelValue
        if err := json.Unmarshal([]byte(cached), &level); err == nil {
            return level, nil
        }
    }
    
    // 从GitLab API获取
    member, _, err := s.gitlab.ProjectMembers.GetProjectMember(projectID, userID)
    if err != nil {
        return 0, err
    }
    
    // 缓存结果（5分钟）
    levelBytes, _ := json.Marshal(member.AccessLevel)
    s.redis.Set(cacheKey, levelBytes, 5*time.Minute)
    
    return member.AccessLevel, nil
}

// 缓存GitLab Groups
func (s *GitLabCacheService) GetUserGroups(userID int) ([]*gitlab.Group, error) {
    cacheKey := fmt.Sprintf("gitlab:groups:user:%d", userID)
    
    // 尝试从缓存获取
    if cached, err := s.redis.Get(cacheKey).Result(); err == nil {
        var groups []*gitlab.Group
        if err := json.Unmarshal([]byte(cached), &groups); err == nil {
            return groups, nil
        }
    }
    
    // 从GitLab API获取
    groups, _, err := s.gitlab.Groups.ListGroups(&gitlab.ListGroupsOptions{
        MinAccessLevel: gitlab.AccessLevel(gitlab.ReporterPermissions),
    })
    if err != nil {
        return nil, err
    }
    
    // 缓存结果（10分钟）
    groupsBytes, _ := json.Marshal(groups)
    s.redis.Set(cacheKey, groupsBytes, 10*time.Minute)
    
    return groups, nil
}
```

### 2. 并发处理优化
```go
// 并发处理服务
type ConcurrentService struct {
    gitlab     *GitLabService
    maxWorkers int
    semaphore  chan struct{}
}

func NewConcurrentService(gitlab *GitLabService, maxWorkers int) *ConcurrentService {
    return &ConcurrentService{
        gitlab:     gitlab,
        maxWorkers: maxWorkers,
        semaphore:  make(chan struct{}, maxWorkers),
    }
}

// 并发加载用户数据
func (s *ConcurrentService) LoadUserData(userID int) (*EducationDashboard, error) {
    type result struct {
        groups   []*gitlab.Group
        projects []*gitlab.Project
        issues   []*gitlab.Issue
        mrs      []*gitlab.MergeRequest
        err      error
    }
    
    results := make(chan result, 4)
    
    // 并发执行4个任务
    go s.loadWithSemaphore(func() {
        groups, err := s.gitlab.GetUserGroups(userID)
        results <- result{groups: groups, err: err}
    })
    
    go s.loadWithSemaphore(func() {
        projects, err := s.gitlab.GetUserProjects(userID)
        results <- result{projects: projects, err: err}
    })
    
    go s.loadWithSemaphore(func() {
        issues, err := s.gitlab.GetAssignedIssues(userID)
        results <- result{issues: issues, err: err}
    })
    
    go s.loadWithSemaphore(func() {
        mrs, err := s.gitlab.GetAssignedMRs(userID)
        results <- result{mrs: mrs, err: err}
    })
    
    // 收集结果
    var dashboard EducationDashboard
    for i := 0; i < 4; i++ {
        res := <-results
        if res.err != nil {
            return nil, res.err
        }
        
        if res.groups != nil {
            dashboard.Groups = res.groups
        }
        if res.projects != nil {
            dashboard.Projects = res.projects
        }
        if res.issues != nil {
            dashboard.AssignedIssues = res.issues
        }
        if res.mrs != nil {
            dashboard.AssignedMRs = res.mrs
        }
    }
    
    return &dashboard, nil
}

func (s *ConcurrentService) loadWithSemaphore(fn func()) {
    s.semaphore <- struct{}{} // 获取信号量
    defer func() { <-s.semaphore }() // 释放信号量
    fn()
}
```

## 项目实施计划 - 修订版

### 第一阶段（2周）- 环境和基础架构
- ✅ 搭建测试环境（GitLab CE、OnlyOffice、PostgreSQL、Redis）
- 🔄 基础架构搭建（Go后端、Vue前端框架）
- 🔄 GitLab API集成（OAuth、基础API封装）
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

### 第四阶段（2周）- 集成测试和部署
- 基于gitlab现有前端页面进行设计，
- 系统包含登录、主视觉界面；主界面包含顶部栏、左侧菜单、中间内容区域（常见的后台管理风格）
- 登录后主界面，左侧菜单有“项目管理”、“团队管理”、“课题管理”、“作业管理”、“学习进度跟踪”、“通知系统”、“教育报表”等页面
- 根据需求设计，实现核心功能
### 第五阶段（2周）- 集成测试和部署
- 系统集成测试
- 性能优化（缓存、并发）
- 安全加固（GitLab OAuth、权限控制）
- 部署上线（Docker Compose一键部署）

## 总结

### 方案优势
1. **开发效率高** - 减少70%以上的自定义代码，专注核心价值
2. **系统稳定性好** - 基于GitLab成熟的功能，避免重复造轮子
3. **维护成本低** - 最小化自定义逻辑，降低维护复杂度
4. **用户体验佳** - 用户可以无缝使用GitLab的所有功能
5. **扩展性强** - 基于GitLab API，可以轻松扩展功能

### 核心价值
本方案将系统定位为**GitLab教育增强平台**，而非重复造轮子的完整社区系统。通过：
- 最大化复用GitLab的用户管理、团队协作、权限控制、项目管理功能
- 专注于OnlyOffice文档协作集成和教育场景UI优化
- 提供轻量级的GitLab API封装和教育业务逻辑
- 实现了既强大又简洁的教育社区解决方案

这种设计理念确保了系统的可持续发展，让开发团队能够专注于真正的差异化功能，而不是重复实现已有的成熟功能。 

#### 前端集成示例
```vue
<template>
  <div class="wiki-document-manager">
    <!-- Wiki页面内容 -->
    <div class="wiki-content" v-html="wikiContent"></div>
    
    <!-- 可编辑文档附件列表 -->
    <div class="document-attachments" v-if="editableAttachments.length > 0">
      <h3>可编辑文档附件</h3>
      <div class="attachment-list">
        <div 
          v-for="attachment in editableAttachments" 
          :key="attachment.id"
          class="attachment-item"
        >
          <div class="attachment-info">
            <i :class="getFileIcon(attachment.file_type)"></i>
            <span class="file-name">{{ attachment.file_name }}</span>
            <span class="last-edited" v-if="attachment.last_edited_at">
              最后编辑: {{ formatTime(attachment.last_edited_at) }}
            </span>
          </div>
          <div class="attachment-actions">
            <el-button 
              size="small" 
              type="primary" 
              @click="openOnlyOfficeEditor(attachment.id)"
              :disabled="!attachment.can_edit"
            >
              <i class="el-icon-edit"></i> 在线编辑
            </el-button>
            <el-button 
              size="small" 
              @click="downloadAttachment(attachment.file_url)"
            >
              <i class="el-icon-download"></i> 下载
            </el-button>
          </div>
        </div>
      </div>
    </div>
    
    <!-- OnlyOffice编辑器模态框 -->
    <el-dialog 
      v-model="editorVisible" 
      title="文档编辑" 
      width="90%" 
      fullscreen
      :before-close="handleEditorClose"
    >
      <div ref="onlyofficeContainer" style="height: 100vh;"></div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'

const route = useRoute()
const wikiContent = ref('')
const editableAttachments = ref([])
const editorVisible = ref(false)
let currentDocEditor = null

const props = defineProps({
  projectId: {
    type: Number,
    required: true
  },
  wikiSlug: {
    type: String, 
    required: true
  }
})

onMounted(async () => {
  await loadWikiContent()
  await loadEditableAttachments()
})

const loadWikiContent = async () => {
  try {
    const response = await fetch(`/api/projects/${props.projectId}/wiki/${props.wikiSlug}`)
    const data = await response.json()
    wikiContent.value = data.content
  } catch (error) {
    ElMessage.error('加载Wiki内容失败')
  }
}

const loadEditableAttachments = async () => {
  try {
    const response = await fetch(`/api/projects/${props.projectId}/wiki/${props.wikiSlug}/attachments`)
    const data = await response.json()
    editableAttachments.value = data
  } catch (error) {
    ElMessage.error('加载文档附件失败')
  }
}

const openOnlyOfficeEditor = async (attachmentId) => {
  try {
    // 启动OnlyOffice编辑会话
    const response = await fetch(`/api/documents/${attachmentId}/edit`, {
      method: 'POST'
    })
    const config = await response.json()
    
    // 显示编辑器
    editorVisible.value = true
    
    // 等待DOM更新
    await nextTick()
    
    // 初始化OnlyOffice编辑器
    currentDocEditor = new DocsAPI.DocEditor("onlyofficeContainer", {
      documentType: config.documentType,
      document: config.document,
      editorConfig: config.editorConfig,
      token: config.token,
      events: {
        onAppReady: () => {
          console.log('OnlyOffice编辑器已就绪')
        },
        onDocumentStateChange: (event) => {
          console.log('文档状态变更:', event.data)
        }
      }
    })
  } catch (error) {
    ElMessage.error('启动文档编辑器失败')
  }
}

const handleEditorClose = () => {
  if (currentDocEditor) {
    currentDocEditor.destroyEditor()
    currentDocEditor = null
  }
  editorVisible.value = false
  // 重新加载附件信息
  loadEditableAttachments()
}

const getFileIcon = (fileType) => {
  switch (fileType) {
    case 'docx':
      return 'el-icon-document'
    case 'xlsx':
      return 'el-icon-s-grid'
    case 'pptx':
      return 'el-icon-picture-outline'
    default:
      return 'el-icon-document'
  }
}

const formatTime = (time) => {
  return new Date(time).toLocaleString()
}

const downloadAttachment = (fileUrl) => {
  window.open(fileUrl, '_blank')
}
</script>
```

这样的设计完全基于GitLab的Wiki功能，同时利用了OnlyOffice的强大编辑能力，实现了：

1. **完全的GitLab集成** - 使用Wiki作为文档管理基础
2. **权限控制简化** - 直接使用GitLab Wiki权限
3. **版本控制自动化** - 利用GitLab的文件版本管理
4. **无缝的编辑体验** - OnlyOffice与GitLab文件存储的完美结合

## 数据库设计 - 极简化

```sql
-- 用户表 - 只存储GitLab用户映射
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    gitlab_id INTEGER UNIQUE NOT NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar VARCHAR(255),
    last_sync_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 文档附件表 - 只存储OnlyOffice编辑会话信息
CREATE TABLE document_attachments (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL,
    wiki_page_slug VARCHAR(255) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_url VARCHAR(500) NOT NULL,
    file_type VARCHAR(50) NOT NULL,
    onlyoffice_key VARCHAR(255) UNIQUE NOT NULL,
    last_edited_by INTEGER REFERENCES users(id),
    last_edited_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, wiki_page_slug, file_name)
);

-- 删除原有的复杂表结构
-- 不再需要：teams, projects, permissions, roles, topics, assignments 等表
-- 所有这些信息都从GitLab API获取或使用GitLab原生功能实现
```

// ... existing code ... 