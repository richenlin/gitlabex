package models

import (
	"time"
)

// Project 课题模型
type Project struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	Name                 string    `gorm:"not null" json:"name"`
	Description          string    `json:"description"`
	Code                 string    `gorm:"unique;not null" json:"code"`    // 课题代码，用于学生加入
	TeacherID            uint      `gorm:"not null" json:"teacher_id"`     // 创建课题的老师ID
	Status               string    `gorm:"default:'active'" json:"status"` // active, completed, archived
	Type                 string    `gorm:"default:'practice'" json:"type"` // graduation, research, competition, practice
	StartDate            time.Time `json:"start_date"`
	EndDate              time.Time `json:"end_date"`
	MaxMembers           int       `gorm:"default:10" json:"max_members"`
	TotalAssignments     int       `gorm:"default:0" json:"total_assignments"`
	CompletedAssignments int       `gorm:"default:0" json:"completed_assignments"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`

	// GitLab 相关字段
	GitLabProjectID int    `json:"gitlab_project_id"`                    // GitLab项目ID
	GitLabURL       string `json:"gitlab_url"`                           // GitLab项目URL
	RepositoryURL   string `json:"repository_url"`                       // 仓库URL
	DefaultBranch   string `gorm:"default:'main'" json:"default_branch"` // 默认分支
	ReadmeContent   string `json:"readme_content"`                       // README内容（课题介绍）
	WikiEnabled     bool   `gorm:"default:true" json:"wiki_enabled"`     // 是否启用Wiki
	IssuesEnabled   bool   `gorm:"default:true" json:"issues_enabled"`   // 是否启用Issues
	MREnabled       bool   `gorm:"default:true" json:"mr_enabled"`       // 是否启用合并请求

	// 互动开发相关字段
	CodeEditorEnabled    bool     `gorm:"default:true" json:"code_editor_enabled"`        // 是否启用在线代码编辑器
	MainBranchProtected  bool     `gorm:"default:true" json:"main_branch_protected"`      // 主分支是否受保护
	AllowedFileTypes     []string `gorm:"serializer:json" json:"allowed_file_types"`      // 允许编辑的文件类型
	MaxFileSize          int64    `gorm:"default:10485760" json:"max_file_size"`          // 最大文件大小（字节）
	AutoSaveInterval     int      `gorm:"default:30" json:"auto_save_interval"`           // 自动保存间隔（秒）
	StudentBranchPrefix  string   `gorm:"default:'student'" json:"student_branch_prefix"` // 学生分支前缀
	EnableRealTimeCollab bool     `gorm:"default:false" json:"enable_real_time_collab"`   // 是否启用实时协作
	RequireCommitMessage bool     `gorm:"default:true" json:"require_commit_message"`     // 是否要求提交消息
	EditorTheme          string   `gorm:"default:'vs-dark'" json:"editor_theme"`          // 编辑器主题
	EditorFontSize       int      `gorm:"default:14" json:"editor_font_size"`             // 编辑器字体大小
	EditorTabSize        int      `gorm:"default:4" json:"editor_tab_size"`               // 编辑器Tab大小
	EnableLinting        bool     `gorm:"default:true" json:"enable_linting"`             // 是否启用代码检查
	EnableFormatting     bool     `gorm:"default:true" json:"enable_formatting"`          // 是否启用代码格式化

	// 关联关系 - 仅用于查询填充，无强制外键约束
	Teacher     User            `gorm:"foreignKey:TeacherID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"teacher,omitempty"`
	Members     []ProjectMember `gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"members,omitempty"`
	Students    []User          `gorm:"many2many:project_members;" json:"students,omitempty"`
	Assignments []Assignment    `gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"assignments,omitempty"`
	Files       []ProjectFile   `gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"files,omitempty"`
}

// TableName 指定表名
func (Project) TableName() string {
	return "projects"
}

// ProjectMember 课题成员关系
type ProjectMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uint      `gorm:"not null" json:"project_id"`
	UserID    uint      `gorm:"not null" json:"user_id"`       // 用户ID（可以是老师或学生）
	Role      string    `gorm:"default:'student'" json:"role"` // teacher, student
	JoinedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"joined_at"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`

	// GitLab 相关字段
	GitLabAccessLevel int        `json:"gitlab_access_level"` // GitLab访问级别
	PersonalBranch    string     `json:"personal_branch"`     // 个人分支名
	PersonalBranchURL string     `json:"personal_branch_url"` // 个人分支URL
	LastCommitHash    string     `json:"last_commit_hash"`    // 最后提交的哈希
	LastCommitMessage string     `json:"last_commit_message"` // 最后提交的消息
	LastCommitTime    *time.Time `json:"last_commit_time"`    // 最后提交时间

	// 互动开发相关字段
	BranchCreatedAt   *time.Time `json:"branch_created_at"`                       // 分支创建时间
	LastEditTime      *time.Time `json:"last_edit_time"`                          // 最后编辑时间
	OnlineEditEnabled bool       `gorm:"default:true" json:"online_edit_enabled"` // 是否启用在线编辑
	LocalCloneEnabled bool       `gorm:"default:true" json:"local_clone_enabled"` // 是否启用本地克隆
	FilesModified     int        `gorm:"default:0" json:"files_modified"`         // 修改的文件数
	LinesAdded        int        `gorm:"default:0" json:"lines_added"`            // 添加的行数
	LinesDeleted      int        `gorm:"default:0" json:"lines_deleted"`          // 删除的行数
	EditorPreferences string     `json:"editor_preferences"`                      // 编辑器偏好设置(JSON)
	LastActiveTime    *time.Time `json:"last_active_time"`                        // 最后活跃时间

	// 关联关系 - 仅用于查询填充，无强制外键约束
	Project Project `gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"project,omitempty"`
	User    User    `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
}

// TableName 指定表名
func (ProjectMember) TableName() string {
	return "project_members"
}

// ProjectFile 项目文件记录
type ProjectFile struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ProjectID   uint      `gorm:"not null" json:"project_id"`
	FilePath    string    `gorm:"not null" json:"file_path"`         // 文件路径
	FileName    string    `gorm:"not null" json:"file_name"`         // 文件名
	FileType    string    `json:"file_type"`                         // 文件类型
	FileSize    int64     `json:"file_size"`                         // 文件大小
	Branch      string    `gorm:"default:'main'" json:"branch"`      // 所属分支
	Content     string    `json:"content"`                           // 文件内容（小文件）
	ContentHash string    `json:"content_hash"`                      // 内容哈希
	IsDirectory bool      `gorm:"default:false" json:"is_directory"` // 是否是目录
	CreatedBy   uint      `json:"created_by"`                        // 创建者ID
	UpdatedBy   uint      `json:"updated_by"`                        // 更新者ID
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// GitLab 相关字段
	GitLabFileID   string `json:"gitlab_file_id"`   // GitLab文件ID
	GitLabBlobID   string `json:"gitlab_blob_id"`   // GitLab Blob ID
	GitLabCommitID string `json:"gitlab_commit_id"` // GitLab提交ID
	GitLabFileURL  string `json:"gitlab_file_url"`  // GitLab文件URL

	// 编辑相关字段
	IsEditable      bool       `gorm:"default:true" json:"is_editable"` // 是否可编辑
	LastEditedBy    uint       `json:"last_edited_by"`                  // 最后编辑者ID
	LastEditedAt    *time.Time `json:"last_edited_at"`                  // 最后编辑时间
	EditLockBy      uint       `json:"edit_lock_by"`                    // 编辑锁定者ID
	EditLockAt      *time.Time `json:"edit_lock_at"`                    // 编辑锁定时间
	EditLockExpires *time.Time `json:"edit_lock_expires"`               // 编辑锁定过期时间
	Language        string     `json:"language"`                        // 编程语言
	Encoding        string     `gorm:"default:'utf-8'" json:"encoding"` // 文件编码

	// 关联关系
	Project    Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Creator    User    `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Updater    User    `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
	LastEditor User    `gorm:"foreignKey:LastEditedBy" json:"last_editor,omitempty"`
	EditLocker User    `gorm:"foreignKey:EditLockBy" json:"edit_locker,omitempty"`
}

// TableName 指定表名
func (ProjectFile) TableName() string {
	return "project_files"
}

// ProjectStats 课题统计信息
type ProjectStats struct {
	MemberCount          int `json:"member_count"`
	AssignmentCount      int `json:"assignment_count"`
	CompletedAssignments int `json:"completed_assignments"`
	PendingAssignments   int `json:"pending_assignments"`

	// GitLab 相关统计
	TotalCommits       int `json:"total_commits"`
	TotalIssues        int `json:"total_issues"`
	OpenIssues         int `json:"open_issues"`
	TotalMergeRequests int `json:"total_merge_requests"`
	OpenMergeRequests  int `json:"open_merge_requests"`
	WikiPages          int `json:"wiki_pages"`
	ActiveBranches     int `json:"active_branches"`

	// 互动开发统计
	TotalFiles         int `json:"total_files"`
	EditableFiles      int `json:"editable_files"`
	ActiveEditors      int `json:"active_editors"`
	RecentEdits        int `json:"recent_edits"`
	TotalLinesOfCode   int `json:"total_lines_of_code"`
	StudentsWithBranch int `json:"students_with_branch"`
	OnlineEditSessions int `json:"online_edit_sessions"`
}

// CodeEditSession 代码编辑会话
type CodeEditSession struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	ProjectID    uint       `gorm:"not null" json:"project_id"`
	UserID       uint       `gorm:"not null" json:"user_id"`
	FilePath     string     `gorm:"not null" json:"file_path"`
	Branch       string     `gorm:"not null" json:"branch"`
	SessionID    string     `gorm:"unique;not null" json:"session_id"`
	StartTime    time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	Status       string     `gorm:"default:'active'" json:"status"` // active, ended, expired
	LastPing     time.Time  `json:"last_ping"`
	ChangesCount int        `gorm:"default:0" json:"changes_count"`
	SavedCount   int        `gorm:"default:0" json:"saved_count"`

	// 关联关系 - 仅用于查询填充，无强制外键约束
	Project Project `gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"project,omitempty"`
	User    User    `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
}

// TableName 指定表名
func (CodeEditSession) TableName() string {
	return "code_edit_sessions"
}
