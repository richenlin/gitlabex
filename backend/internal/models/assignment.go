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
	DueDate     time.Time `json:"due_date"`
	MaxScore    int       `gorm:"default:100" json:"max_score"`
	Status      string    `gorm:"default:'active'" json:"status"` // active, closed
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联关系
	Project     Project                `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Teacher     User                   `gorm:"foreignKey:TeacherID" json:"teacher,omitempty"`
	Submissions []AssignmentSubmission `gorm:"foreignKey:AssignmentID" json:"submissions,omitempty"`
}

// TableName 指定表名
func (Assignment) TableName() string {
	return "assignments"
}

// AssignmentSubmission 作业提交模型
type AssignmentSubmission struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	AssignmentID uint      `gorm:"not null" json:"assignment_id"`
	StudentID    uint      `gorm:"not null" json:"student_id"`
	Content      string    `json:"content"`   // 作业内容
	FilePath     string    `json:"file_path"` // 附件文件路径
	SubmittedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"submitted_at"`
	Status       string    `gorm:"default:'submitted'" json:"status"` // submitted, reviewed, returned
	Score        int       `json:"score"`                             // 得分
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 关联关系
	Assignment *Assignment `gorm:"foreignKey:AssignmentID" json:"assignment,omitempty"`
	Student    User        `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	Review     *Review     `gorm:"foreignKey:SubmissionID" json:"review,omitempty"`
}

// TableName 指定表名
func (AssignmentSubmission) TableName() string {
	return "assignment_submissions"
}

// Review 评审模型
type Review struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	SubmissionID uint      `gorm:"unique;not null" json:"submission_id"` // 一对一关系
	TeacherID    uint      `gorm:"not null" json:"teacher_id"`
	Score        int       `json:"score"`
	Comment      string    `json:"comment"`
	Feedback     string    `json:"feedback"`                          // 详细反馈
	Status       string    `gorm:"default:'completed'" json:"status"` // completed, draft
	ReviewedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"reviewed_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 关联关系
	Submission AssignmentSubmission `gorm:"foreignKey:SubmissionID" json:"submission,omitempty"`
	Teacher    User                 `gorm:"foreignKey:TeacherID" json:"teacher,omitempty"`
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
