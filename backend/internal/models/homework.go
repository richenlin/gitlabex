package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// HomeworkStatus 作业状态
type HomeworkStatus string

const (
	HomeworkStatusDraft     HomeworkStatus = "draft"
	HomeworkStatusPublished HomeworkStatus = "published"
	HomeworkStatusArchived  HomeworkStatus = "archived"
)

// SubmissionStatus 提交状态
type SubmissionStatus string

const (
	SubmissionStatusPending   SubmissionStatus = "pending"
	SubmissionStatusSubmitted SubmissionStatus = "submitted"
	SubmissionStatusGraded    SubmissionStatus = "graded"
	SubmissionStatusReturned  SubmissionStatus = "returned"
)

// Homework 作业模型
type Homework struct {
	BaseModel
	Title        string         `gorm:"not null;size:200" json:"title"`
	Description  string         `gorm:"type:text" json:"description"`
	Content      string         `gorm:"type:text" json:"content"` // 作业内容，存储在数据库中
	ProjectID    uuid.UUID      `gorm:"not null" json:"project_id"`
	CreatorID    int64          `gorm:"not null" json:"creator_id"` // GitLab用户ID
	Status       HomeworkStatus `gorm:"not null;default:draft" json:"status"`
	DueDate      *time.Time     `json:"due_date,omitempty"`
	MaxGrade     int            `gorm:"default:100" json:"max_grade"`
	MinGrade     int            `gorm:"default:0" json:"min_grade"`
	Instructions string         `gorm:"type:text" json:"instructions"`
	Requirements pq.StringArray `gorm:"type:text[]" json:"requirements"`
	Tags         pq.StringArray `gorm:"type:text[]" json:"tags"`

	// 关联关系
	Project ResearchProject `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	// 注意：Creator关联已移除，创建者信息从GitLab API获取
	Submissions []Submission `gorm:"foreignKey:HomeworkID" json:"submissions,omitempty"`
}

// Submission 作业提交模型
type Submission struct {
	BaseModel
	HomeworkID      uuid.UUID        `gorm:"not null" json:"homework_id"`
	StudentID       int64            `gorm:"not null" json:"student_id"` // GitLab用户ID
	Status          SubmissionStatus `gorm:"not null;default:pending" json:"status"`
	Content         string           `gorm:"type:text" json:"content"` // 提交内容说明
	GitLabCommitSHA string           `gorm:"size:40" json:"gitlab_commit_sha,omitempty"`
	GitLabBranch    string           `gorm:"size:100" json:"gitlab_branch,omitempty"` // 学生提交的分支名
	SubmittedAt     *time.Time       `json:"submitted_at,omitempty"`
	Grade           *int             `json:"grade,omitempty"`
	Feedback        string           `gorm:"type:text" json:"feedback"` // 老师评语
	GradedAt        *time.Time       `json:"graded_at,omitempty"`
	GradedBy        *int64           `json:"graded_by,omitempty"` // 批改者GitLab用户ID

	// 关联关系
	Homework Homework `gorm:"foreignKey:HomeworkID" json:"homework,omitempty"`
	// 注意：Student和Grader关联已移除，用户信息从GitLab API获取
}

// GradeDistribution 成绩分布模型
type GradeDistribution struct {
	HomeworkID uuid.UUID `gorm:"not null" json:"homework_id"`
	GradeRange string    `gorm:"not null" json:"grade_range"`
	Count      int       `gorm:"not null" json:"count"`
	Percentage float64   `gorm:"not null" json:"percentage"`
}

// SubmissionHistory 提交历史模型
type SubmissionHistory struct {
	BaseModel
	SubmissionID uuid.UUID `gorm:"not null" json:"submission_id"`
	Action       string    `gorm:"not null" json:"action"`
	OldStatus    string    `gorm:"not null" json:"old_status"`
	NewStatus    string    `gorm:"not null" json:"new_status"`
	ChangedBy    uuid.UUID `gorm:"not null" json:"changed_by"`
	Notes        string    `gorm:"type:text" json:"notes"`

	// 关联关系
	Submission Submission `gorm:"foreignKey:SubmissionID" json:"submission,omitempty"`
	// 注意：Changer关联已移除，用户信息从GitLab API获取
}
