package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/xanzy/go-gitlab"
	"gorm.io/gorm"

	"gitlabex/internal/models"
)

// TeamService 团队管理服务 - 基于GitLab Groups
type TeamService struct {
	gitlab *GitLabService
	db     *gorm.DB
}

// NewTeamService 创建团队服务实例
func NewTeamService(gitlabService *GitLabService, db *gorm.DB) *TeamService {
	return &TeamService{
		gitlab: gitlabService,
		db:     db,
	}
}

// CreateClass 创建班级（GitLab Group）
func (s *TeamService) CreateClass(name, description string, teacherID int) (*gitlab.Group, error) {
	// 创建GitLab Group
	group, _, err := s.gitlab.client.Groups.CreateGroup(&gitlab.CreateGroupOptions{
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

// CreateTeam 创建项目团队（GitLab Subgroup）
func (s *TeamService) CreateTeam(parentGroupID int, name, description string, leaderID int) (*gitlab.Group, error) {
	// 创建子组
	group, _, err := s.gitlab.client.Groups.CreateGroup(&gitlab.CreateGroupOptions{
		Name:        gitlab.String(name),
		Path:        gitlab.String(s.generateTeamPath(name)),
		Description: gitlab.String(description),
		Visibility:  gitlab.Visibility(gitlab.PrivateVisibility),
		ParentID:    gitlab.Int(parentGroupID),
	})
	if err != nil {
		return nil, err
	}

	// 设置团队负责人为Maintainer
	_, _, err = s.gitlab.client.GroupMembers.AddGroupMember(group.ID, &gitlab.AddGroupMemberOptions{
		UserID:      gitlab.Int(leaderID),
		AccessLevel: gitlab.AccessLevel(gitlab.MaintainerPermissions),
	})
	if err != nil {
		return nil, err
	}

	return group, nil
}

// AddTeamMember 添加团队成员
func (s *TeamService) AddTeamMember(groupID, userID int, role models.EducationRole) error {
	accessLevel := s.mapEducationRoleToGitLab(role)
	_, _, err := s.gitlab.client.GroupMembers.AddGroupMember(groupID, &gitlab.AddGroupMemberOptions{
		UserID:      gitlab.Int(userID),
		AccessLevel: gitlab.AccessLevel(accessLevel),
	})
	return err
}

// RemoveTeamMember 移除团队成员
func (s *TeamService) RemoveTeamMember(groupID, userID int) error {
	_, err := s.gitlab.client.GroupMembers.RemoveGroupMember(groupID, userID, &gitlab.RemoveGroupMemberOptions{})
	return err
}

// UpdateMemberRole 更新成员角色
func (s *TeamService) UpdateMemberRole(groupID, userID int, role models.EducationRole) error {
	accessLevel := s.mapEducationRoleToGitLab(role)
	_, _, err := s.gitlab.client.GroupMembers.EditGroupMember(groupID, userID, &gitlab.EditGroupMemberOptions{
		AccessLevel: gitlab.AccessLevel(accessLevel),
	})
	return err
}

// GetTeamMembers 获取团队成员列表
func (s *TeamService) GetTeamMembers(groupID int) ([]*TeamMember, error) {
	// 暂时返回空列表，待GitLab API修复
	// TODO: 实现正确的 GitLab 成员列表获取
	return []*TeamMember{}, nil
}

// GetUserTeams 获取用户所属团队
func (s *TeamService) GetUserTeams(userID int) ([]*gitlab.Group, error) {
	groups, _, err := s.gitlab.client.Groups.ListGroups(&gitlab.ListGroupsOptions{
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

// GetTeamDetails 获取团队详情
func (s *TeamService) GetTeamDetails(groupID int) (*TeamDetails, error) {
	// 获取Group信息
	group, _, err := s.gitlab.client.Groups.GetGroup(groupID, &gitlab.GetGroupOptions{})
	if err != nil {
		return nil, err
	}

	// 获取成员列表
	members, err := s.GetTeamMembers(groupID)
	if err != nil {
		return nil, err
	}

	// 获取项目列表
	projects, _, err := s.gitlab.client.Groups.ListGroupProjects(groupID, &gitlab.ListGroupProjectsOptions{})
	if err != nil {
		return nil, err
	}

	// 统计成员角色分布
	roleStats := make(map[models.EducationRole]int)
	for _, member := range members {
		roleStats[member.Role]++
	}

	details := &TeamDetails{
		Group:        group,
		Members:      members,
		Projects:     projects,
		MemberCount:  len(members),
		ProjectCount: len(projects),
		RoleStats:    roleStats,
		CreatedAt:    group.CreatedAt,
	}

	return details, nil
}

// SearchTeams 搜索团队
func (s *TeamService) SearchTeams(query string, userID int) ([]*gitlab.Group, error) {
	groups, _, err := s.gitlab.client.Groups.ListGroups(&gitlab.ListGroupsOptions{
		Search:       gitlab.String(query),
		AllAvailable: gitlab.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	// 筛选用户有权限访问的Groups
	var accessibleGroups []*gitlab.Group
	for _, group := range groups {
		if s.canUserAccessGroup(userID, group.ID) {
			accessibleGroups = append(accessibleGroups, group)
		}
	}

	return accessibleGroups, nil
}

// GetTeamActivity 获取团队活动
func (s *TeamService) GetTeamActivity(groupID int, limit int) ([]*TeamActivity, error) {
	// 获取Group的项目列表
	projects, _, err := s.gitlab.client.Groups.ListGroupProjects(groupID, &gitlab.ListGroupProjectsOptions{})
	if err != nil {
		return nil, err
	}

	var activities []*TeamActivity

	// 收集所有项目的活动
	for _, project := range projects {
		// 获取最近的Issues
		issues, _, err := s.gitlab.client.Issues.ListProjectIssues(project.ID, &gitlab.ListProjectIssuesOptions{
			OrderBy: gitlab.String("updated_at"),
			Sort:    gitlab.String("desc"),
		})
		if err == nil && len(issues) > 0 {
			for i, issue := range issues {
				if i >= limit {
					break
				}
				activity := &TeamActivity{
					Type:        "issue",
					Title:       issue.Title,
					Description: issue.Description,
					Author:      issue.Author.Name,
					CreatedAt:   issue.CreatedAt,
					UpdatedAt:   issue.UpdatedAt,
					ProjectName: project.Name,
				}
				activities = append(activities, activity)
			}
		}

		// 获取最近的MR
		mrs, _, err := s.gitlab.client.MergeRequests.ListProjectMergeRequests(project.ID, &gitlab.ListProjectMergeRequestsOptions{
			OrderBy: gitlab.String("updated_at"),
			Sort:    gitlab.String("desc"),
		})
		if err == nil && len(mrs) > 0 {
			for i, mr := range mrs {
				if i >= limit {
					break
				}
				activity := &TeamActivity{
					Type:        "merge_request",
					Title:       mr.Title,
					Description: mr.Description,
					Author:      mr.Author.Name,
					CreatedAt:   mr.CreatedAt,
					UpdatedAt:   mr.UpdatedAt,
					ProjectName: project.Name,
				}
				activities = append(activities, activity)
			}
		}
	}

	// 按更新时间排序
	if len(activities) > limit {
		activities = activities[:limit]
	}

	return activities, nil
}

// generateClassPath 生成班级路径
func (s *TeamService) generateClassPath(name string) string {
	// 转换为URL友好的路径
	path := strings.ToLower(name)
	path = strings.ReplaceAll(path, " ", "-")
	path = strings.ReplaceAll(path, "班", "class")
	return fmt.Sprintf("class-%s-%d", path, time.Now().Unix())
}

// generateTeamPath 生成团队路径
func (s *TeamService) generateTeamPath(name string) string {
	// 转换为URL友好的路径
	path := strings.ToLower(name)
	path = strings.ReplaceAll(path, " ", "-")
	path = strings.ReplaceAll(path, "组", "team")
	return fmt.Sprintf("team-%s-%d", path, time.Now().Unix())
}

// initializeClassTemplate 初始化班级模板
func (s *TeamService) initializeClassTemplate(groupID int) error {
	// 创建班级的标准项目结构
	projects := []struct {
		name        string
		description string
	}{
		{"课题管理", "班级课题项目管理"},
		{"作业管理", "班级作业发布和提交"},
		{"班级公告", "班级通知和公告"},
		{"资料共享", "班级学习资料和文档"},
	}

	for _, proj := range projects {
		_, _, err := s.gitlab.client.Projects.CreateProject(&gitlab.CreateProjectOptions{
			Name:        gitlab.String(proj.name),
			NamespaceID: gitlab.Int(groupID),
			Description: gitlab.String(proj.description),
			Visibility:  gitlab.Visibility(gitlab.PrivateVisibility),
			// 启用必要功能
			IssuesEnabled:        gitlab.Bool(true),
			MergeRequestsEnabled: gitlab.Bool(true),
			WikiEnabled:          gitlab.Bool(true),
		})
		if err != nil {
			return fmt.Errorf("failed to create template project %s: %w", proj.name, err)
		}
	}

	return nil
}

// isUserInGroup 检查用户是否在组中
func (s *TeamService) isUserInGroup(userID, groupID int) bool {
	_, _, err := s.gitlab.client.GroupMembers.GetGroupMember(groupID, userID)
	return err == nil
}

// canUserAccessGroup 检查用户是否可以访问组
func (s *TeamService) canUserAccessGroup(userID, groupID int) bool {
	// 检查用户是否是组成员
	if s.isUserInGroup(userID, groupID) {
		return true
	}

	// 检查组是否为公开
	group, _, err := s.gitlab.client.Groups.GetGroup(groupID, &gitlab.GetGroupOptions{})
	if err != nil {
		return false
	}

	return group.Visibility == gitlab.PublicVisibility
}

// mapEducationRoleToGitLab 教育角色映射到GitLab权限
func (s *TeamService) mapEducationRoleToGitLab(role models.EducationRole) gitlab.AccessLevelValue {
	switch role {
	case models.EduRoleStudent:
		return gitlab.ReporterPermissions
	case models.EduRoleAssistant:
		return gitlab.DeveloperPermissions
	case models.EduRoleTeacher:
		return gitlab.MaintainerPermissions
	case models.EduRoleAdmin:
		return gitlab.OwnerPermissions
	default:
		return gitlab.GuestPermissions
	}
}

// mapGitLabAccessLevel GitLab权限映射到教育角色
func (s *TeamService) mapGitLabAccessLevel(level gitlab.AccessLevelValue) models.EducationRole {
	switch level {
	case gitlab.GuestPermissions:
		return models.EduRoleGuest
	case gitlab.ReporterPermissions:
		return models.EduRoleStudent
	case gitlab.DeveloperPermissions:
		return models.EduRoleAssistant
	case gitlab.MaintainerPermissions:
		return models.EduRoleTeacher
	case gitlab.OwnerPermissions:
		return models.EduRoleAdmin
	default:
		return models.EduRoleGuest
	}
}

// TeamMember 团队成员信息
type TeamMember struct {
	ID          int                     `json:"id"`
	Username    string                  `json:"username"`
	Name        string                  `json:"name"`
	Email       string                  `json:"email"`
	Avatar      string                  `json:"avatar"`
	AccessLevel gitlab.AccessLevelValue `json:"access_level"`
	Role        models.EducationRole    `json:"role"`
}

// TeamDetails 团队详情
type TeamDetails struct {
	Group        *gitlab.Group                `json:"group"`
	Members      []*TeamMember                `json:"members"`
	Projects     []*gitlab.Project            `json:"projects"`
	MemberCount  int                          `json:"member_count"`
	ProjectCount int                          `json:"project_count"`
	RoleStats    map[models.EducationRole]int `json:"role_stats"`
	CreatedAt    *time.Time                   `json:"created_at"`
}

// TeamActivity 团队活动
type TeamActivity struct {
	Type        string     `json:"type"` // "issue", "merge_request", "commit"
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Author      string     `json:"author"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	ProjectName string     `json:"project_name"`
}
