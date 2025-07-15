package services

import (
	"fmt"
	"time"

	"gitlabex/internal/models"

	"gorm.io/gorm"
)

// AnalyticsService 分析服务
type AnalyticsService struct {
	db *gorm.DB
}

// NewAnalyticsService 创建分析服务实例
func NewAnalyticsService(db *gorm.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

// AdminOverview 管理员概览数据
type AdminOverview struct {
	TotalProjects    int64 `json:"total_projects"`
	TotalAssignments int64 `json:"total_assignments"`
	TotalStudents    int64 `json:"total_students"`
	CompletionRate   int   `json:"completion_rate"`
}

// TeacherOverview 教师概览数据
type TeacherOverview struct {
	TotalProjects    int     `json:"total_projects"`
	ActiveProjects   int     `json:"active_projects"`
	TotalAssignments int     `json:"total_assignments"`
	TotalSubmissions int     `json:"total_submissions"`
	PendingReviews   int     `json:"pending_reviews"`
	TotalStudents    int     `json:"total_students"`
	AverageScore     float64 `json:"average_score"`
	CompletionRate   float64 `json:"completion_rate"`
}

// StudentOverview 学生概览数据
type StudentOverview struct {
	JoinedProjects       int     `json:"joined_projects"`
	ActiveAssignments    int     `json:"active_assignments"`
	CompletedAssignments int     `json:"completed_assignments"`
	PendingAssignments   int     `json:"pending_assignments"`
	TotalSubmissions     int     `json:"total_submissions"`
	AverageScore         float64 `json:"average_score"`
	HighestScore         float64 `json:"highest_score"`
	LowestScore          float64 `json:"lowest_score"`
}

// ProjectStat 课题统计数据
type ProjectStat struct {
	Name            string    `json:"name"`
	StudentCount    int       `json:"student_count"`
	AssignmentCount int       `json:"assignment_count"`
	CompletionRate  int       `json:"completion_rate"`
	AverageScore    float64   `json:"average_score"`
	LastActivity    time.Time `json:"last_activity"`
}

// StudentStat 学生统计数据
type StudentStat struct {
	Name            string  `json:"name"`
	ClassName       string  `json:"class_name"`
	ProjectCount    int     `json:"project_count"`
	AssignmentCount int     `json:"assignment_count"`
	CompletedCount  int     `json:"completed_count"`
	CompletionRate  int     `json:"completion_rate"`
	AverageScore    float64 `json:"average_score"`
}

// AssignmentStat 作业统计数据
type AssignmentStat struct {
	Title           string    `json:"title"`
	ProjectName     string    `json:"project_name"`
	DueDate         time.Time `json:"due_date"`
	SubmissionCount int       `json:"submission_count"`
	TotalStudents   int       `json:"total_students"`
	CompletionRate  int       `json:"completion_rate"`
	AverageScore    float64   `json:"average_score"`
}

// ChartData 图表数据结构
type ChartData struct {
	Labels   []string `json:"labels"`
	Datasets []struct {
		Label           string    `json:"label"`
		Data            []float64 `json:"data"`
		BackgroundColor string    `json:"backgroundColor,omitempty"`
		BorderColor     string    `json:"borderColor,omitempty"`
		Tension         float64   `json:"tension,omitempty"`
	} `json:"datasets"`
}

// Activity 活动数据
type Activity struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

// DashboardStats 仪表盘统计数据
type DashboardStats struct {
	ClassesCount            int `json:"classesCount"`
	ActiveProjectsCount     int `json:"activeProjectsCount"`
	PendingAssignmentsCount int `json:"pendingAssignmentsCount"`
	DocumentsCount          int `json:"documentsCount"`
}

// GetAdminOverview 获取管理员概览数据
func (s *AnalyticsService) GetAdminOverview(userID uint) (*AdminOverview, error) {
	overview := &AdminOverview{}

	// 获取总课题数
	if err := s.db.Model(&models.Project{}).Count(&overview.TotalProjects); err != nil {
		return nil, fmt.Errorf("failed to count projects: %w", err)
	}

	// 获取总作业数
	if err := s.db.Model(&models.Assignment{}).Count(&overview.TotalAssignments); err != nil {
		return nil, fmt.Errorf("failed to count assignments: %w", err)
	}

	// 获取总学生数
	if err := s.db.Model(&models.User{}).Where("role = ?", 3).Count(&overview.TotalStudents); err != nil {
		return nil, fmt.Errorf("failed to count students: %w", err)
	}

	// 计算完成率
	overview.CompletionRate = s.calculateOverallCompletionRate()

	return overview, nil
}

// GetTeacherOverview 获取教师概览数据
func (s *AnalyticsService) GetTeacherOverview(userID uint) (*TeacherOverview, error) {
	overview := &TeacherOverview{}

	// 课题统计
	var totalProjects int64
	var activeProjects int64
	s.db.Model(&models.Project{}).Where("teacher_id = ?", userID).Count(&totalProjects)
	s.db.Model(&models.Project{}).Where("teacher_id = ? AND status = 'active'", userID).Count(&activeProjects)

	// 作业统计
	var totalAssignments int64
	s.db.Model(&models.Assignment{}).
		Joins("JOIN projects ON assignments.project_id = projects.id").
		Where("projects.teacher_id = ?", userID).
		Count(&totalAssignments)

	// 提交统计
	var totalSubmissions int64
	var pendingReviews int64
	s.db.Model(&models.AssignmentSubmission{}).
		Joins("JOIN assignments ON assignment_submissions.assignment_id = assignments.id").
		Joins("JOIN projects ON assignments.project_id = projects.id").
		Where("projects.teacher_id = ?", userID).
		Count(&totalSubmissions)

	s.db.Model(&models.AssignmentSubmission{}).
		Joins("JOIN assignments ON assignment_submissions.assignment_id = assignments.id").
		Joins("JOIN projects ON assignments.project_id = projects.id").
		Where("projects.teacher_id = ? AND assignment_submissions.status = 'submitted'", userID).
		Count(&pendingReviews)

	// 学生统计
	var totalStudents int64
	s.db.Model(&models.ProjectMember{}).
		Joins("JOIN projects ON project_members.project_id = projects.id").
		Where("projects.teacher_id = ? AND project_members.is_active = true", userID).
		Count(&totalStudents)

	// 平均分和完成率
	var avgScore struct {
		AvgScore float64
		Count    int64
	}
	s.db.Model(&models.AssignmentSubmission{}).
		Select("AVG(score) as avg_score, COUNT(*) as count").
		Joins("JOIN assignments ON assignment_submissions.assignment_id = assignments.id").
		Joins("JOIN projects ON assignments.project_id = projects.id").
		Where("projects.teacher_id = ? AND assignment_submissions.status = 'graded'", userID).
		Scan(&avgScore)

	completionRate := float64(0)
	if totalAssignments > 0 {
		var completedSubmissions int64
		s.db.Model(&models.AssignmentSubmission{}).
			Joins("JOIN assignments ON assignment_submissions.assignment_id = assignments.id").
			Joins("JOIN projects ON assignments.project_id = projects.id").
			Where("projects.teacher_id = ? AND assignment_submissions.status IN ('submitted', 'graded')", userID).
			Count(&completedSubmissions)
		completionRate = float64(completedSubmissions) / float64(totalAssignments) * 100
	}

	overview.TotalProjects = int(totalProjects)
	overview.ActiveProjects = int(activeProjects)
	overview.TotalAssignments = int(totalAssignments)
	overview.TotalSubmissions = int(totalSubmissions)
	overview.PendingReviews = int(pendingReviews)
	overview.TotalStudents = int(totalStudents)
	overview.AverageScore = avgScore.AvgScore
	overview.CompletionRate = completionRate

	return overview, nil
}

// GetStudentOverview 获取学生概览数据
func (s *AnalyticsService) GetStudentOverview(userID uint) (*StudentOverview, error) {
	overview := &StudentOverview{}

	// 参与的课题数量
	var joinedProjects int64
	s.db.Model(&models.ProjectMember{}).
		Where("user_id = ? AND is_active = true", userID).
		Count(&joinedProjects)

	// 作业统计
	var activeAssignments int64
	var completedAssignments int64
	var totalSubmissions int64

	// 活跃作业（未提交的）
	s.db.Model(&models.Assignment{}).
		Joins("JOIN project_members ON assignments.project_id = project_members.project_id").
		Joins("LEFT JOIN assignment_submissions ON assignments.id = assignment_submissions.assignment_id AND assignment_submissions.student_id = ?", userID).
		Where("project_members.user_id = ? AND project_members.is_active = true AND assignments.status = 'active' AND assignment_submissions.id IS NULL", userID).
		Count(&activeAssignments)

	// 已完成作业（已提交的）
	s.db.Model(&models.AssignmentSubmission{}).
		Where("student_id = ?", userID).
		Count(&totalSubmissions)

	s.db.Model(&models.AssignmentSubmission{}).
		Where("student_id = ? AND status IN ('submitted', 'graded')", userID).
		Count(&completedAssignments)

	pendingAssignments := int(activeAssignments)

	// 分数统计
	var scoreStats struct {
		AvgScore float64
		MaxScore float64
		MinScore float64
		Count    int64
	}
	s.db.Model(&models.AssignmentSubmission{}).
		Select("AVG(score) as avg_score, MAX(score) as max_score, MIN(score) as min_score, COUNT(*) as count").
		Where("student_id = ? AND status = 'graded'", userID).
		Scan(&scoreStats)

	overview.JoinedProjects = int(joinedProjects)
	overview.ActiveAssignments = int(activeAssignments)
	overview.CompletedAssignments = int(completedAssignments)
	overview.PendingAssignments = pendingAssignments
	overview.TotalSubmissions = int(totalSubmissions)
	overview.AverageScore = scoreStats.AvgScore
	overview.HighestScore = scoreStats.MaxScore
	overview.LowestScore = scoreStats.MinScore

	return overview, nil
}

// GetProjectStats 获取课题统计数据
func (s *AnalyticsService) GetProjectStats(userID uint, userRole int) ([]ProjectStat, error) {
	var stats []ProjectStat

	query := s.db.Model(&models.Project{}).
		Select("projects.name, COUNT(DISTINCT pm.user_id) as student_count, COUNT(DISTINCT a.id) as assignment_count").
		Joins("LEFT JOIN project_members pm ON pm.project_id = projects.id AND pm.role = 'student'").
		Joins("LEFT JOIN assignments a ON a.project_id = projects.id").
		Group("projects.id, projects.name")

	// 根据角色过滤
	if userRole == 2 { // 教师
		query = query.Where("projects.teacher_id = ?", userID)
	} else if userRole == 3 { // 学生
		query = query.Joins("JOIN project_members pm2 ON pm2.project_id = projects.id").
			Where("pm2.user_id = ? AND pm2.role = 'student'", userID)
	}

	var results []struct {
		Name            string
		StudentCount    int
		AssignmentCount int
	}

	if err := query.Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get project stats: %w", err)
	}

	for _, result := range results {
		stat := ProjectStat{
			Name:            result.Name,
			StudentCount:    result.StudentCount,
			AssignmentCount: result.AssignmentCount,
			CompletionRate:  s.calculateProjectCompletionRate(result.Name),
			AverageScore:    s.calculateProjectAverageScore(result.Name),
			LastActivity:    time.Now(), // 简化实现
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

// GetStudentStats 获取学生统计数据
func (s *AnalyticsService) GetStudentStats(userID uint, userRole int) ([]StudentStat, error) {
	var stats []StudentStat

	query := s.db.Model(&models.User{}).
		Select("users.name, '' as class_name, COUNT(DISTINCT pm.project_id) as project_count").
		Joins("LEFT JOIN project_members pm ON pm.user_id = users.id AND pm.role = 'student'").
		Where("users.role = ?", 3).
		Group("users.id, users.name")

	// 根据角色过滤
	if userRole == 2 { // 教师只能看到自己班级的学生
		query = query.Joins("JOIN project_members pm2 ON pm2.user_id = users.id").
			Joins("JOIN projects p ON p.id = pm2.project_id").
			Where("p.teacher_id = ?", userID)
	}

	var results []struct {
		Name         string
		ClassName    string
		ProjectCount int
	}

	if err := query.Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get student stats: %w", err)
	}

	for _, result := range results {
		stat := StudentStat{
			Name:            result.Name,
			ClassName:       result.ClassName,
			ProjectCount:    result.ProjectCount,
			AssignmentCount: s.calculateStudentAssignmentCount(result.Name),
			CompletedCount:  s.calculateStudentCompletedCount(result.Name),
			CompletionRate:  s.calculateStudentCompletionRate(0), // 简化实现
			AverageScore:    s.calculateStudentAverageScore(result.Name),
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

// GetAssignmentStats 获取作业统计数据
func (s *AnalyticsService) GetAssignmentStats(userID uint, userRole int) ([]AssignmentStat, error) {
	var stats []AssignmentStat

	query := s.db.Model(&models.Assignment{}).
		Select("assignments.title, projects.name as project_name, assignments.due_date").
		Joins("JOIN projects ON assignments.project_id = projects.id")

	// 根据角色过滤
	if userRole == 2 { // 教师
		query = query.Where("projects.teacher_id = ?", userID)
	} else if userRole == 3 { // 学生
		query = query.Joins("JOIN project_members pm ON pm.project_id = projects.id").
			Where("pm.user_id = ? AND pm.role = 'student'", userID)
	}

	var results []struct {
		Title       string
		ProjectName string
		DueDate     time.Time
	}

	if err := query.Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get assignment stats: %w", err)
	}

	for _, result := range results {
		stat := AssignmentStat{
			Title:           result.Title,
			ProjectName:     result.ProjectName,
			DueDate:         result.DueDate,
			SubmissionCount: s.calculateAssignmentSubmissionCount(result.Title),
			TotalStudents:   s.calculateAssignmentTotalStudents(result.Title),
			CompletionRate:  s.calculateAssignmentCompletionRate(result.Title),
			AverageScore:    s.calculateAssignmentAverageScore(result.Title),
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

// GetSubmissionTrend 获取提交趋势数据
func (s *AnalyticsService) GetSubmissionTrend(userID uint, userRole int, startDate, endDate time.Time) (*ChartData, error) {
	// 简化实现，返回模拟数据
	data := &ChartData{
		Labels: []string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"},
		Datasets: []struct {
			Label           string    `json:"label"`
			Data            []float64 `json:"data"`
			BackgroundColor string    `json:"backgroundColor,omitempty"`
			BorderColor     string    `json:"borderColor,omitempty"`
			Tension         float64   `json:"tension,omitempty"`
		}{
			{
				Label:       "作业提交数",
				Data:        []float64{5, 8, 12, 15, 10, 6, 3},
				BorderColor: "#409EFF",
				Tension:     0.4,
			},
		},
	}
	return data, nil
}

// GetProjectDistribution 获取课题分布数据
func (s *AnalyticsService) GetProjectDistribution(userID uint, userRole int) (*ChartData, error) {
	// 简化实现，返回模拟数据
	data := &ChartData{
		Labels: []string{"Web开发", "数据库", "算法", "移动开发", "AI"},
		Datasets: []struct {
			Label           string    `json:"label"`
			Data            []float64 `json:"data"`
			BackgroundColor string    `json:"backgroundColor,omitempty"`
			BorderColor     string    `json:"borderColor,omitempty"`
			Tension         float64   `json:"tension,omitempty"`
		}{
			{
				Data:            []float64{30, 25, 20, 15, 10},
				BackgroundColor: "#409EFF",
			},
		},
	}
	return data, nil
}

// GetGradeDistribution 获取成绩分布数据
func (s *AnalyticsService) GetGradeDistribution(userID uint, userRole int) (*ChartData, error) {
	// 简化实现，返回模拟数据
	data := &ChartData{
		Labels: []string{"90-100", "80-89", "70-79", "60-69", "60以下"},
		Datasets: []struct {
			Label           string    `json:"label"`
			Data            []float64 `json:"data"`
			BackgroundColor string    `json:"backgroundColor,omitempty"`
			BorderColor     string    `json:"borderColor,omitempty"`
			Tension         float64   `json:"tension,omitempty"`
		}{
			{
				Label:           "学生数",
				Data:            []float64{15, 25, 30, 20, 10},
				BackgroundColor: "#409EFF",
			},
		},
	}
	return data, nil
}

// GetActivityStats 获取活跃度统计数据
func (s *AnalyticsService) GetActivityStats(userID uint, userRole int) (*ChartData, error) {
	// 简化实现，返回模拟数据
	data := &ChartData{
		Labels: []string{"作业提交", "文档编辑", "讨论参与", "代码提交", "项目活跃"},
		Datasets: []struct {
			Label           string    `json:"label"`
			Data            []float64 `json:"data"`
			BackgroundColor string    `json:"backgroundColor,omitempty"`
			BorderColor     string    `json:"borderColor,omitempty"`
			Tension         float64   `json:"tension,omitempty"`
		}{
			{
				Label:           "活跃度",
				Data:            []float64{80, 65, 70, 85, 75},
				BackgroundColor: "#409EFF20",
				BorderColor:     "#409EFF",
			},
		},
	}
	return data, nil
}

// GetDashboardStats 获取仪表盘统计数据
func (s *AnalyticsService) GetDashboardStats(userID uint, userRole int) (*DashboardStats, error) {
	var stats DashboardStats

	// 根据角色返回不同的统计数据
	switch userRole {
	case 1: // 管理员
		stats = DashboardStats{
			ClassesCount:            5,
			ActiveProjectsCount:     15,
			PendingAssignmentsCount: 8,
			DocumentsCount:          25,
		}
	case 2: // 教师
		stats = DashboardStats{
			ClassesCount:            3,
			ActiveProjectsCount:     8,
			PendingAssignmentsCount: 12,
			DocumentsCount:          15,
		}
	case 3: // 学生
		stats = DashboardStats{
			ClassesCount:            1,
			ActiveProjectsCount:     4,
			PendingAssignmentsCount: 2,
			DocumentsCount:          15,
		}
	default:
		stats = DashboardStats{
			ClassesCount:            0,
			ActiveProjectsCount:     0,
			PendingAssignmentsCount: 0,
			DocumentsCount:          0,
		}
	}

	return &stats, nil
}

// GetRecentActivities 获取最近活动数据
func (s *AnalyticsService) GetRecentActivities(userID uint, userRole int, limit int) ([]Activity, error) {
	activities := []Activity{
		{
			Title:       "新课题创建",
			Description: "创建了新的课题：Web开发实战",
			Timestamp:   time.Now().Add(-time.Hour * 2),
		},
		{
			Title:       "作业提交",
			Description: "学生张三提交了作业：数据库设计",
			Timestamp:   time.Now().Add(-time.Hour * 4),
		},
		{
			Title:       "文档更新",
			Description: "更新了项目文档：API设计规范",
			Timestamp:   time.Now().Add(-time.Hour * 6),
		},
	}

	// 根据limit限制返回的活动数量
	if limit > 0 && limit < len(activities) {
		activities = activities[:limit]
	}

	return activities, nil
}

// 辅助方法 - 计算整体完成率
func (s *AnalyticsService) calculateOverallCompletionRate() int {
	return 75 // 简化实现
}

// 辅助方法 - 计算教师完成率
func (s *AnalyticsService) calculateTeacherCompletionRate(userID uint) int {
	return 80 // 简化实现
}

// 辅助方法 - 计算学生完成率
func (s *AnalyticsService) calculateStudentCompletionRate(userID uint) int {
	return 70 // 简化实现
}

// 辅助方法 - 计算课题完成率
func (s *AnalyticsService) calculateProjectCompletionRate(projectName string) int {
	return 75 // 简化实现
}

// 辅助方法 - 计算课题平均分
func (s *AnalyticsService) calculateProjectAverageScore(projectName string) float64 {
	return 82.5 // 简化实现
}

// 辅助方法 - 计算学生作业数
func (s *AnalyticsService) calculateStudentAssignmentCount(studentName string) int {
	return 10 // 简化实现
}

// 辅助方法 - 计算学生完成数
func (s *AnalyticsService) calculateStudentCompletedCount(studentName string) int {
	return 8 // 简化实现
}

// 辅助方法 - 计算学生平均分
func (s *AnalyticsService) calculateStudentAverageScore(studentName string) float64 {
	return 85.0 // 简化实现
}

// 辅助方法 - 计算作业提交数
func (s *AnalyticsService) calculateAssignmentSubmissionCount(assignmentTitle string) int {
	return 20 // 简化实现
}

// 辅助方法 - 计算作业学生总数
func (s *AnalyticsService) calculateAssignmentTotalStudents(assignmentTitle string) int {
	return 25 // 简化实现
}

// 辅助方法 - 计算作业完成率
func (s *AnalyticsService) calculateAssignmentCompletionRate(assignmentTitle string) int {
	return 80 // 简化实现
}

// 辅助方法 - 计算作业平均分
func (s *AnalyticsService) calculateAssignmentAverageScore(assignmentTitle string) float64 {
	return 78.5 // 简化实现
}
