package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateHomeworkRequest 创建作业请求
type CreateHomeworkRequest struct {
	ProjectID   string     `json:"project_id" binding:"required"`
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"`
	MaxGrade    int        `json:"max_grade"`
}

// UpdateHomeworkRequest 更新作业请求
type UpdateHomeworkRequest struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	DueDate     *time.Time `json:"due_date"`
	MaxGrade    *int       `json:"max_grade"`
	Status      *string    `json:"status"`
}

// GradeHomeworkRequest 评分作业请求
type GradeHomeworkRequest struct {
	Grade    int    `json:"grade" binding:"required"`
	Feedback string `json:"feedback"`
}

// SubmitHomeworkRequest 提交作业请求
type SubmitHomeworkRequest struct {
	Content string   `json:"content" binding:"required"`
	Files   []string `json:"files"`
}

// BulkUpdateDueDateRequest 批量更新截止日期请求
type BulkUpdateDueDateRequest struct {
	HomeworkIDs []string   `json:"homework_ids" binding:"required"`
	DueDate     *time.Time `json:"due_date"`
}

// CreateAssignmentTemplateRequest 创建作业模板请求
type CreateAssignmentTemplateRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	MaxGrade    int       `json:"max_grade"`
	IsPublic    bool      `json:"is_public"`
	Tags        []string  `json:"tags"`
}

// UpdateAssignmentTemplateRequest 更新作业模板请求
type UpdateAssignmentTemplateRequest struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	DueDate     *time.Time `json:"due_date"`
	MaxGrade    *int       `json:"max_grade"`
	IsPublic    *bool      `json:"is_public"`
	Tags        []string   `json:"tags"`
}

// UseAssignmentTemplateRequest 使用模板创建作业请求
type UseAssignmentTemplateRequest struct {
	TemplateID string `json:"template_id" binding:"required"`
	ProjectID  string `json:"project_id" binding:"required"`
}

// BulkCreateHomeworkRequest 批量创建作业项
type BulkCreateHomeworkItem struct {
	ProjectID   string     `json:"project_id" binding:"required"`
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"`
	MaxGrade    int        `json:"max_grade"`
}

// HomeworkResponse 作业响应（包含权限信息）
type HomeworkResponse struct {
	Homework    interface{}            `json:"homework"`
	Stats       map[string]interface{} `json:"stats,omitempty"`
	Permissions map[string]bool        `json:"permissions"`
}

// SubmissionResponse 提交响应（包含学生信息）
type SubmissionResponse struct {
	ID              uuid.UUID              `json:"id"`
	HomeworkID      uuid.UUID              `json:"homework_id"`
	StudentID       int64                  `json:"student_id"`
	Status          string                 `json:"status"`
	Content         string                 `json:"content"`
	GitLabBranch    string                 `json:"gitlab_branch"`
	GitLabCommitSHA string                 `json:"gitlab_commit_sha,omitempty"`
	SubmittedAt     *time.Time             `json:"submitted_at"`
	Grade           *int                   `json:"grade"`
	Feedback        string                 `json:"feedback"`
	GradedAt        *time.Time             `json:"graded_at"`
	GradedBy        *int64                 `json:"graded_by"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Student         map[string]interface{} `json:"student"`
}

// SubmitHomeworkResponse 提交作业响应
type SubmitHomeworkResponse struct {
	Branch      string `json:"branch"`
	Path        string `json:"path"`
	WebIDEURL   string `json:"web_ide_url"`
	ProjectPath string `json:"project_path"`
	Message     string `json:"message"`
}

// ViewSubmissionResponse 查看提交响应
type ViewSubmissionResponse struct {
	ViewURL     string `json:"view_url"`
	Branch      string `json:"branch"`
	Path        string `json:"path"`
	ProjectPath string `json:"project_path"`
	StudentName string `json:"student_name"`
}
