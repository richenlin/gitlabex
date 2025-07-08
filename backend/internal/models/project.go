package models

import (
	"time"

	"gorm.io/gorm"
)

// Project 项目模型
type Project struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	GitLabID       int            `gorm:"unique;not null" json:"gitlab_id"`
	Name           string         `gorm:"not null;size:255" json:"name"`
	Description    string         `gorm:"type:text" json:"description"`
	Path           string         `gorm:"not null;size:255" json:"path"`
	Namespace      string         `gorm:"not null;size:255" json:"namespace"`
	DefaultBranch  string         `gorm:"size:100;default:main" json:"default_branch"`
	WebURL         string         `gorm:"size:500" json:"web_url"`
	SSHURL         string         `gorm:"size:500" json:"ssh_url"`
	HTTPURL        string         `gorm:"size:500" json:"http_url"`
	Visibility     string         `gorm:"size:20;default:private" json:"visibility"` // private, internal, public
	ForksCount     int            `gorm:"default:0" json:"forks_count"`
	StarsCount     int            `gorm:"default:0" json:"stars_count"`
	IssuesEnabled  bool           `gorm:"default:true" json:"issues_enabled"`
	WikiEnabled    bool           `gorm:"default:true" json:"wiki_enabled"`
	OwnerID        uint           `gorm:"not null;index" json:"owner_id"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	LastActivityAt *time.Time     `json:"last_activity_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Owner     User       `gorm:"foreignKey:OwnerID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"owner,omitempty"`
	Members   []User     `gorm:"many2many:project_members;" json:"members,omitempty"`
	Teams     []Team     `gorm:"many2many:team_projects;" json:"teams,omitempty"`
	Documents []Document `gorm:"foreignKey:ProjectID;references:GitLabID" json:"documents,omitempty"`
}

// TableName 指定表名
func (Project) TableName() string {
	return "projects"
}

// ProjectMember 项目成员
type ProjectMember struct {
	ID        uint              `gorm:"primaryKey" json:"id"`
	ProjectID uint              `gorm:"not null;index" json:"project_id"`
	UserID    uint              `gorm:"not null;index" json:"user_id"`
	Role      ProjectMemberRole `gorm:"not null" json:"role"`
	AddedBy   uint              `gorm:"not null" json:"added_by"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`

	// 关联关系
	Project Project `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	User    User    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
	Adder   User    `gorm:"foreignKey:AddedBy;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"adder,omitempty"`
}

// TableName 指定表名
func (ProjectMember) TableName() string {
	return "project_members"
}

// ProjectMemberRole 项目成员角色枚举
type ProjectMemberRole int

const (
	RoleGuest      ProjectMemberRole = iota + 10 // 访客 (10)
	RoleReporter                                 // 报告者 (20)
	RoleDeveloper                                // 开发者 (30)
	RoleMaintainer                               // 维护者 (40)
	RoleOwner                                    // 所有者 (50)
)

// String 返回角色的字符串表示
func (r ProjectMemberRole) String() string {
	switch r {
	case RoleGuest:
		return "guest"
	case RoleReporter:
		return "reporter"
	case RoleDeveloper:
		return "developer"
	case RoleMaintainer:
		return "maintainer"
	case RoleOwner:
		return "owner"
	default:
		return "unknown"
	}
}

// CanRead 检查是否有读权限
func (r ProjectMemberRole) CanRead() bool {
	return r >= RoleGuest
}

// CanWrite 检查是否有写权限
func (r ProjectMemberRole) CanWrite() bool {
	return r >= RoleDeveloper
}

// CanAdmin 检查是否有管理权限
func (r ProjectMemberRole) CanAdmin() bool {
	return r >= RoleMaintainer
}

// Team 团队模型
type Team struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"not null;size:255" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Avatar      string         `gorm:"size:500" json:"avatar"`
	IsPublic    bool           `gorm:"default:false" json:"is_public"`
	LeaderID    uint           `gorm:"not null;index" json:"leader_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Leader   User      `gorm:"foreignKey:LeaderID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"leader,omitempty"`
	Members  []User    `gorm:"many2many:team_members;" json:"members,omitempty"`
	Projects []Project `gorm:"many2many:team_projects;" json:"projects,omitempty"`
}

// TableName 指定表名
func (Team) TableName() string {
	return "teams"
}

// TeamMember 团队成员
type TeamMember struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	TeamID    uint           `gorm:"not null;index" json:"team_id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Role      TeamMemberRole `gorm:"not null" json:"role"`
	AddedBy   uint           `gorm:"not null" json:"added_by"`
	JoinedAt  time.Time      `json:"joined_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`

	// 关联关系
	Team  Team `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	User  User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
	Adder User `gorm:"foreignKey:AddedBy;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"adder,omitempty"`
}

// TableName 指定表名
func (TeamMember) TableName() string {
	return "team_members"
}

// TeamMemberRole 团队成员角色枚举
type TeamMemberRole int

const (
	TeamRoleMember TeamMemberRole = iota // 成员
	TeamRoleAdmin                        // 管理员
	TeamRoleLeader                       // 负责人
)

// String 返回角色的字符串表示
func (r TeamMemberRole) String() string {
	switch r {
	case TeamRoleMember:
		return "member"
	case TeamRoleAdmin:
		return "admin"
	case TeamRoleLeader:
		return "leader"
	default:
		return "unknown"
	}
}

// TeamProject 团队项目关联
type TeamProject struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TeamID    uint      `gorm:"not null;index" json:"team_id"`
	ProjectID uint      `gorm:"not null;index" json:"project_id"`
	AddedBy   uint      `gorm:"not null" json:"added_by"`
	CreatedAt time.Time `json:"created_at"`

	// 关联关系
	Team    Team    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Project Project `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Adder   User    `gorm:"foreignKey:AddedBy;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"adder,omitempty"`
}

// TableName 指定表名
func (TeamProject) TableName() string {
	return "team_projects"
}

// TeamPermission 团队权限
type TeamPermission struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TeamID     uint      `gorm:"not null;index" json:"team_id"`
	UserID     uint      `gorm:"not null;index" json:"user_id"`
	Permission string    `gorm:"not null;size:20" json:"permission"` // read, write, admin
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// 关联关系
	Team Team `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	User User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
}

// TableName 指定表名
func (TeamPermission) TableName() string {
	return "team_permissions"
}

// ProjectStats 项目统计信息
type ProjectStats struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	ProjectID          uint      `gorm:"not null;uniqueIndex" json:"project_id"`
	CommitsCount       int       `gorm:"default:0" json:"commits_count"`
	BranchesCount      int       `gorm:"default:0" json:"branches_count"`
	TagsCount          int       `gorm:"default:0" json:"tags_count"`
	FilesCount         int       `gorm:"default:0" json:"files_count"`
	LinesOfCode        int       `gorm:"default:0" json:"lines_of_code"`
	DocumentsCount     int       `gorm:"default:0" json:"documents_count"`
	IssuesCount        int       `gorm:"default:0" json:"issues_count"`
	MergeRequestsCount int       `gorm:"default:0" json:"merge_requests_count"`
	LastUpdated        time.Time `json:"last_updated"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`

	// 关联关系
	Project Project `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

// TableName 指定表名
func (ProjectStats) TableName() string {
	return "project_stats"
}
