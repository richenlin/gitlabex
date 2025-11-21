package dto

import "time"

// ========== GitLab 分支保护相关 DTO ==========

// ProtectBranchRequest 保护分支请求
type ProtectBranchRequest struct {
	Name                      string `json:"name"`
	PushAccessLevel           int    `json:"push_access_level"`      // 30=Developer, 40=Maintainer, 60=Admin
	MergeAccessLevel          int    `json:"merge_access_level"`     // 30=Developer, 40=Maintainer, 60=Admin
	UnprotectAccessLevel      int    `json:"unprotect_access_level"` // 40=Maintainer, 60=Admin
	AllowForcePush            bool   `json:"allow_force_push"`
	CodeOwnerApprovalRequired bool   `json:"code_owner_approval_required"`
}

// ProtectedBranch 受保护的分支信息
type ProtectedBranch struct {
	ID                        int64         `json:"id"`
	Name                      string        `json:"name"`
	PushAccessLevels          []AccessLevel `json:"push_access_levels"`
	MergeAccessLevels         []AccessLevel `json:"merge_access_levels"`
	UnprotectAccessLevels     []AccessLevel `json:"unprotect_access_levels"`
	AllowForcePush            bool          `json:"allow_force_push"`
	CodeOwnerApprovalRequired bool          `json:"code_owner_approval_required"`
}

// AccessLevel 访问级别
type AccessLevel struct {
	AccessLevel            int64  `json:"access_level"`
	AccessLevelDescription string `json:"access_level_description"`
}

// ========== GitLab 用户相关 DTO ==========

// GitLabAPIUser GitLab用户信息 - 用于API响应解析
type GitLabAPIUser struct {
	ID               int64  `json:"id"`
	Username         string `json:"username"`
	Email            string `json:"email"`
	Name             string `json:"name"`
	Avatar           string `json:"avatar_url"`
	IsAdmin          bool   `json:"is_admin"`
	State            string `json:"state"`
	UserType         string `json:"user_type"`
	External         bool   `json:"external"`
	CanCreateGroup   bool   `json:"can_create_group"`
	CanCreateProject bool   `json:"can_create_project"`
}

// GitLabCreateUserData GitLab创建用户数据结构
type GitLabCreateUserData struct {
	Email            string `json:"email"`
	Username         string `json:"username"`
	Name             string `json:"name"`
	Password         string `json:"password"`
	Admin            bool   `json:"admin,omitempty"`
	SkipConfirmation bool   `json:"skip_confirmation,omitempty"`
}

// GitLabUpdateUserData GitLab更新用户数据结构
type GitLabUpdateUserData struct {
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
	Name     string `json:"name,omitempty"`
	Admin    *bool  `json:"admin,omitempty"`
}

// ========== GitLab 项目相关 DTO ==========

// GitLabProject GitLab项目信息
type GitLabProject struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	Path              string `json:"path"`
	PathWithNamespace string `json:"path_with_namespace"`
	WebURL            string `json:"web_url"`
	SSHURL            string `json:"ssh_url_to_repo"`
	HTTPURL           string `json:"http_url_to_repo"`
	Description       string `json:"description"`
	DefaultBranch     string `json:"default_branch"`
	Visibility        string `json:"visibility"`
	CreatedAt         string `json:"created_at"`
	LastActivityAt    string `json:"last_activity_at"`
	Namespace         struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Path string `json:"path"`
	} `json:"namespace"`
	Owner *GitLabAPIUser `json:"owner,omitempty"`
}

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Name                 string `json:"name"`
	Path                 string `json:"path"` // 必需字段，项目路径
	Description          string `json:"description,omitempty"`
	Visibility           string `json:"visibility,omitempty"` // private, internal, public
	InitializeWithReadme bool   `json:"initialize_with_readme,omitempty"`
	DefaultBranch        string `json:"default_branch,omitempty"`
}

// ProjectMember GitLab项目成员结构
type ProjectMember struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	AvatarURL   string `json:"avatar_url"`
	AccessLevel int    `json:"access_level"`
	State       string `json:"state"`
}

// ========== GitLab 文件相关 DTO ==========

// GitLabFile GitLab文件信息
type GitLabFile struct {
	FileName      string `json:"file_name"`
	FilePath      string `json:"file_path"`
	Size          int    `json:"size"`
	Encoding      string `json:"encoding"`
	Content       string `json:"content"`
	ContentSHA256 string `json:"content_sha256"`
	Ref           string `json:"ref"`
	BlobID        string `json:"blob_id"`
	CommitID      string `json:"commit_id"`
	LastCommitID  string `json:"last_commit_id"`
}

// CreateFileRequest 创建文件请求
type CreateFileRequest struct {
	Branch        string `json:"branch"`
	Content       string `json:"content"`
	CommitMessage string `json:"commit_message"`
	AuthorEmail   string `json:"author_email,omitempty"`
	AuthorName    string `json:"author_name,omitempty"`
}

// ========== GitLab 分支相关 DTO ==========

// GitLabBranch GitLab分支信息
type GitLabBranch struct {
	Name   string `json:"name"`
	Commit struct {
		ID        string `json:"id"`
		ShortID   string `json:"short_id"`
		Title     string `json:"title"`
		Author    string `json:"author_name"`
		Message   string `json:"message"`
		CreatedAt string `json:"created_at"`
	} `json:"commit"`
	Protected bool `json:"protected"`
	Merged    bool `json:"merged"`
}

// CreateBranchRequest 创建分支请求
type CreateBranchRequest struct {
	Branch string `json:"branch"`
	Ref    string `json:"ref"`
}

// ========== GitLab 提交相关 DTO ==========

// GitLabCommit GitLab提交信息
type GitLabCommit struct {
	ID        string         `json:"id"`
	ShortID   string         `json:"short_id"`
	Title     string         `json:"title"`
	Author    *GitLabAPIUser `json:"author"`
	Committer *GitLabAPIUser `json:"committer"`
	Message   string         `json:"message"`
	CreatedAt time.Time      `json:"created_at"`
	WebURL    string         `json:"web_url"`
}

// ========== GitLab Issue 相关 DTO ==========

// GitLabIssue GitLab议题信息
type GitLabIssue struct {
	ID          int64    `json:"id"`
	IID         int64    `json:"iid"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	State       string   `json:"state"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
	Labels      []string `json:"labels"`
	Author      struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	} `json:"author"`
	Upvotes        int    `json:"upvotes"`
	Downvotes      int    `json:"downvotes"`
	UserNotesCount int    `json:"user_notes_count"`
	WebURL         string `json:"web_url"`
}

// GitLabIssueNote GitLab Issue评论结构
type GitLabIssueNote struct {
	ID        int64  `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Author    struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	} `json:"author"`
}

// GitLabAwardEmoji GitLab Award Emoji结构
type GitLabAwardEmoji struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	User struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	} `json:"user"`
}

// ========== GitLab 合并请求相关 DTO ==========

// GitLabMergeRequest GitLab合并请求信息
type GitLabMergeRequest struct {
	ID           int64          `json:"id"`
	IID          int64          `json:"iid"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	State        string         `json:"state"`
	SourceBranch string         `json:"source_branch"`
	TargetBranch string         `json:"target_branch"`
	Author       *GitLabAPIUser `json:"author"`
	Assignee     *GitLabAPIUser `json:"assignee"`
	CreatedAt    string         `json:"created_at"`
	UpdatedAt    string         `json:"updated_at"`
	WebURL       string         `json:"web_url"`
	MergedAt     *string        `json:"merged_at,omitempty"`
}

// ========== GitLab Webhook 相关 DTO ==========

// GitLabWebhookPayload GitLab webhook负载
type GitLabWebhookPayload struct {
	ObjectKind string          `json:"object_kind"`
	EventName  string          `json:"event_name"`
	User       *GitLabAPIUser  `json:"user"`
	Project    *GitLabProject  `json:"project"`
	Commits    []*GitLabCommit `json:"commits"`
	Ref        string          `json:"ref"`
	Before     string          `json:"before"`
	After      string          `json:"after"`
	Repository struct {
		Name        string `json:"name"`
		URL         string `json:"url"`
		Description string `json:"description"`
		Homepage    string `json:"homepage"`
	} `json:"repository"`
}
