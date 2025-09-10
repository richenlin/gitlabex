package models

import (
	"time"

	"github.com/google/uuid"
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
	Title       string         `gorm:"not null;size:200" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	ProjectID   uuid.UUID      `gorm:"not null" json:"project_id"`
	CreatorID   uuid.UUID      `gorm:"not null" json:"creator_id"`
	Status      HomeworkStatus `gorm:"not null;default:draft" json:"status"`
	DueDate     *time.Time     `json:"due_date,omitempty"`
	MaxGrade    int            `gorm:"default:100" json:"max_grade"`
	MinGrade    int            `gorm:"default:0" json:"min_grade"`
	Instructions string        `gorm:"type:text" json:"instructions"`
	TemplateFiles []string    `gorm:"type:text[]" json:"template_files"`
	Requirements []string     `gorm:"type:text[]" json:"requirements"`
	Tags        []string       `gorm:"type:text[]" json:"tags"`
	GitLabBranch string        `gorm:"size:100;default:main" json:"gitlab_branch,omitempty"`
	GitLabPath   string        `gorm:"size:500" json:"gitlab_path,omitempty"`
	
	// 关联关系
	Project   ResearchProject `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Creator   User            `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	Submissions []Submission  `gorm:"foreignKey:HomeworkID" json:"submissions,omitempty"`
}

// Submission 作业提交模型
type Submission struct {
	BaseModel
	HomeworkID uuid.UUID      `gorm:"not null" json:"homework_id"`
	StudentID  uuid.UUID      `gorm:"not null" json:"student_id"`
	Status     SubmissionStatus `gorm:"not null;default:pending" json:"status"`
	Content    string         `gorm:"type:text" json:"content"`
	FilePath   string         `gorm:"size:500" json:"file_path"`
	GitLabCommitSHA string    `gorm:"size:40" json:"gitlab_commit_sha,omitempty"`
	GitLabBranch string      `gorm:"size:100" json:"gitlab_branch,omitempty"`
	SubmittedAt *time.Time  `json:"submitted_at,omitempty"`
	Grade      *int         `json:"grade,omitempty"`
	Feedback   string       `gorm:"type:text" json:"feedback"`
	GradedAt   *time.Time   `json:"graded_at,omitempty"`
	GradedBy   *uuid.UUID   `json:"graded_by,omitempty"`
	
	// 关联关系
	Homework Homework `gorm:"foreignKey:HomeworkID" json:"homework,omitempty"`
	Student  User     `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	Grader   *User    `gorm:"foreignKey:GradedBy" json:"grader,omitempty"`
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
	Changer    User       `gorm:"foreignKey:ChangedBy" json:"changer,omitempty"`
}