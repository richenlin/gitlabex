package models

import (
	"time"
)

// Assignment 作业模型
type Assignment struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`
	ProjectID   uint      `gorm:"not null" json:"project_id"`
	TeacherID   uint      `gorm:"not null" json:"teacher_id"`
	Type        string    `gorm:"default:'homework'" json:"type"` // homework, project, quiz
	Status      string    `gorm:"default:'active'" json:"status"` // active, completed, archived
	DueDate     time.Time `json:"due_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// GitLab 相关字段
	RequiredFiles     []string `gorm:"serializer:json" json:"required_files"`     // 必须提交的文件列表
	SubmissionBranch  string   `json:"submission_branch"`                         // 提交分支前缀
	AutoCreateMR      bool     `gorm:"default:false" json:"auto_create_mr"`       // 是否自动创建合并请求
	MRTitle           string   `json:"mr_title"`                                  // 合并请求标题模板
	MRDescription     string   `json:"mr_description"`                            // 合并请求描述模板
	RequireCodeReview bool     `gorm:"default:false" json:"require_code_review"`  // 是否需要代码审查
	MaxFileSize       int64    `gorm:"default:10485760" json:"max_file_size"`     // 最大文件大小（字节）
	AllowedFileTypes  []string `gorm:"serializer:json" json:"allowed_file_types"` // 允许的文件类型

	// 关联关系
	Project     Project                `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Teacher     User                   `gorm:"foreignKey:TeacherID" json:"teacher,omitempty"`
	Submissions []AssignmentSubmission `gorm:"foreignKey:AssignmentID" json:"submissions,omitempty"`
}

// TableName 指定表名
func (Assignment) TableName() string {
	return "assignments"
}

// AssignmentSubmission 作业提交记录
type AssignmentSubmission struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	AssignmentID uint       `gorm:"not null" json:"assignment_id"`
	StudentID    uint       `gorm:"not null" json:"student_id"`
	Content      string     `json:"content"`
	Status       string     `gorm:"default:'submitted'" json:"status"` // submitted, graded, returned
	Score        float64    `json:"score"`
	MaxScore     float64    `gorm:"default:100" json:"max_score"`
	Feedback     string     `json:"feedback"`
	SubmittedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"submitted_at"`
	GradedAt     *time.Time `json:"graded_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// GitLab 相关字段
	CommitHash       string   `json:"commit_hash"`                                 // 提交的Git哈希
	CommitMessage    string   `json:"commit_message"`                              // 提交消息
	CommitURL        string   `json:"commit_url"`                                  // 提交URL
	BranchName       string   `json:"branch_name"`                                 // 提交分支
	BranchURL        string   `json:"branch_url"`                                  // 分支URL
	MergeRequestID   int      `json:"merge_request_id"`                            // 合并请求ID
	MergeRequestURL  string   `json:"merge_request_url"`                           // 合并请求URL
	FilesSubmitted   []string `gorm:"serializer:json" json:"files_submitted"`      // 提交的文件列表
	FilesSummary     string   `json:"files_summary"`                               // 文件摘要
	CodeReviewStatus string   `gorm:"default:'pending'" json:"code_review_status"` // pending, approved, rejected
	AutoCheckPassed  bool     `gorm:"default:false" json:"auto_check_passed"`      // 自动检查是否通过
	AutoCheckResults string   `json:"auto_check_results"`                          // 自动检查结果详情

	// 关联关系
	Assignment Assignment `gorm:"foreignKey:AssignmentID" json:"assignment,omitempty"`
	Student    User       `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	Reviews    []Review   `gorm:"foreignKey:SubmissionID" json:"reviews,omitempty"`
}

// TableName 指定表名
func (AssignmentSubmission) TableName() string {
	return "assignment_submissions"
}

// Review 评审记录
type Review struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	SubmissionID uint      `gorm:"not null" json:"submission_id"`
	ReviewerID   uint      `gorm:"not null" json:"reviewer_id"`
	Score        float64   `json:"score"`
	Feedback     string    `json:"feedback"`
	Status       string    `gorm:"default:'pending'" json:"status"` // pending, completed
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// GitLab 相关字段
	GitLabReviewID    int    `json:"gitlab_review_id"`   // GitLab评审ID
	GitLabReviewURL   string `json:"gitlab_review_url"`  // GitLab评审URL
	InlineComments    int    `json:"inline_comments"`    // 内联评论数量
	GeneralComments   int    `json:"general_comments"`   // 一般评论数量
	ApprovalsRequired int    `json:"approvals_required"` // 需要的批准数量
	ApprovalsReceived int    `json:"approvals_received"` // 收到的批准数量

	// 关联关系
	Submission AssignmentSubmission `gorm:"foreignKey:SubmissionID" json:"submission,omitempty"`
	Reviewer   User                 `gorm:"foreignKey:ReviewerID" json:"reviewer,omitempty"`
}

// TableName 指定表名
func (Review) TableName() string {
	return "reviews"
}

// AssignmentStats 作业统计信息
type AssignmentStats struct {
	TotalSubmissions    int     `json:"total_submissions"`
	ReviewedSubmissions int     `json:"reviewed_submissions"`
	PendingSubmissions  int     `json:"pending_submissions"`
	AverageScore        float64 `json:"average_score"`
}
