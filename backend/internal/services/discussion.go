package services

import (
	"fmt"
	"time"

	"gitlabex/internal/models"

	"github.com/xanzy/go-gitlab"
	"gorm.io/gorm"
)

// DiscussionService 话题讨论服务
type DiscussionService struct {
	db                *gorm.DB
	permissionService *PermissionService
	gitlabService     *GitLabService
}

// NewDiscussionService 创建话题讨论服务
func NewDiscussionService(db *gorm.DB, permissionService *PermissionService, gitlabService *GitLabService) *DiscussionService {
	return &DiscussionService{
		db:                db,
		permissionService: permissionService,
		gitlabService:     gitlabService,
	}
}

// CreateDiscussion 创建话题
func (s *DiscussionService) CreateDiscussion(req *models.DiscussionCreateRequest, authorID uint) (*models.Discussion, error) {
	// 验证项目权限
	if !s.permissionService.CanCreateDiscussion(authorID, req.ProjectID) {
		return nil, fmt.Errorf("没有权限在此项目创建话题")
	}

	// 获取项目信息
	var project models.Project
	if err := s.db.First(&project, req.ProjectID).Error; err != nil {
		return nil, fmt.Errorf("项目不存在: %w", err)
	}

	// 在GitLab中创建Issue
	gitlabIssue, err := s.gitlabService.CreateDiscussion(project.GitLabProjectID, req.Title, req.Content)
	if err != nil {
		return nil, fmt.Errorf("创建GitLab讨论失败: %w", err)
	}

	// 创建本地讨论记录
	discussion := &models.Discussion{
		Title:          req.Title,
		Content:        req.Content,
		ProjectID:      req.ProjectID,
		AuthorID:       authorID,
		GitLabIssueID:  gitlabIssue.ID,
		GitLabIssueURL: gitlabIssue.WebURL,
		Status:         "open",
		Priority:       "normal",
		Category:       s.getValidCategory(req.Category),
		Tags:           req.Tags,
		IsPublic:       req.IsPublic,
		IsPinned:       false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.db.Create(discussion).Error; err != nil {
		return nil, fmt.Errorf("创建讨论记录失败: %w", err)
	}

	// 预加载关联数据
	if err := s.db.Preload("Author").Preload("Project").First(discussion, discussion.ID).Error; err != nil {
		return nil, fmt.Errorf("加载讨论数据失败: %w", err)
	}

	return discussion, nil
}

// GetDiscussionList 获取话题列表
func (s *DiscussionService) GetDiscussionList(projectID uint, page, pageSize int, category, status string, userID uint) (*models.DiscussionListResponse, error) {
	// 验证项目权限
	if !s.permissionService.CanViewProject(userID, projectID) {
		return nil, fmt.Errorf("没有权限查看此项目的话题")
	}

	query := s.db.Model(&models.Discussion{}).Where("project_id = ?", projectID)

	// 非公开话题权限检查
	if !s.permissionService.IsTeacherOrAdmin(userID, projectID) {
		query = query.Where("is_public = ?", true)
	}

	// 分类过滤
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 状态过滤
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("获取话题总数失败: %w", err)
	}

	// 分页查询
	var discussions []models.Discussion
	offset := (page - 1) * pageSize
	err := query.Preload("Author").Preload("Project").
		Order("is_pinned DESC, created_at DESC").
		Limit(pageSize).Offset(offset).
		Find(&discussions).Error
	if err != nil {
		return nil, fmt.Errorf("获取话题列表失败: %w", err)
	}

	return &models.DiscussionListResponse{
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
		Discussions: discussions,
	}, nil
}

// GetDiscussionDetail 获取话题详情
func (s *DiscussionService) GetDiscussionDetail(discussionID uint, userID uint) (*models.DiscussionDetailResponse, error) {
	// 获取话题信息
	var discussion models.Discussion
	if err := s.db.Preload("Author").Preload("Project").First(&discussion, discussionID).Error; err != nil {
		return nil, fmt.Errorf("话题不存在: %w", err)
	}

	// 验证权限
	if !s.permissionService.CanViewProject(userID, discussion.ProjectID) {
		return nil, fmt.Errorf("没有权限查看此话题")
	}

	// 检查是否为非公开话题
	if !discussion.IsPublic && !s.permissionService.IsTeacherOrAdmin(userID, discussion.ProjectID) {
		return nil, fmt.Errorf("没有权限查看此话题")
	}

	// 获取回复列表
	var replies []models.DiscussionReply
	if err := s.db.Preload("Author").Where("discussion_id = ?", discussionID).
		Order("created_at ASC").Find(&replies).Error; err != nil {
		return nil, fmt.Errorf("获取回复失败: %w", err)
	}

	// 检查是否已点赞
	var likeCount int64
	s.db.Model(&models.DiscussionLike{}).Where("discussion_id = ? AND user_id = ?", discussionID, userID).Count(&likeCount)
	isLiked := likeCount > 0

	// 检查编辑权限
	canEdit := s.permissionService.CanEditDiscussion(userID, discussionID)
	canDelete := s.permissionService.CanDeleteDiscussion(userID, discussionID)

	// 记录浏览
	go s.recordView(discussionID, userID)

	return &models.DiscussionDetailResponse{
		Discussion: discussion,
		Replies:    replies,
		IsLiked:    isLiked,
		CanEdit:    canEdit,
		CanDelete:  canDelete,
	}, nil
}

// UpdateDiscussion 更新话题
func (s *DiscussionService) UpdateDiscussion(discussionID uint, req *models.DiscussionUpdateRequest, userID uint) (*models.Discussion, error) {
	// 验证权限
	if !s.permissionService.CanEditDiscussion(userID, discussionID) {
		return nil, fmt.Errorf("没有权限编辑此话题")
	}

	// 获取话题信息
	var discussion models.Discussion
	if err := s.db.First(&discussion, discussionID).Error; err != nil {
		return nil, fmt.Errorf("话题不存在: %w", err)
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Category != "" {
		updates["category"] = s.getValidCategory(req.Category)
	}
	if req.Tags != "" {
		updates["tags"] = req.Tags
	}
	updates["is_public"] = req.IsPublic
	updates["updated_at"] = time.Now()

	// 更新数据库
	if err := s.db.Model(&discussion).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新话题失败: %w", err)
	}

	// 同步到GitLab (如果标题或内容有变化)
	if req.Title != "" || req.Content != "" {
		go s.syncToGitLab(discussion.GitLabIssueID, req.Title, req.Content)
	}

	// 重新加载数据
	if err := s.db.Preload("Author").Preload("Project").First(&discussion, discussionID).Error; err != nil {
		return nil, fmt.Errorf("加载更新后的话题失败: %w", err)
	}

	return &discussion, nil
}

// DeleteDiscussion 删除话题
func (s *DiscussionService) DeleteDiscussion(discussionID uint, userID uint) error {
	// 验证权限
	if !s.permissionService.CanDeleteDiscussion(userID, discussionID) {
		return fmt.Errorf("没有权限删除此话题")
	}

	// 获取话题信息
	var discussion models.Discussion
	if err := s.db.First(&discussion, discussionID).Error; err != nil {
		return fmt.Errorf("话题不存在: %w", err)
	}

	// 软删除本地记录
	if err := s.db.Model(&discussion).Update("deleted_at", time.Now()).Error; err != nil {
		return fmt.Errorf("删除话题失败: %w", err)
	}

	// 关闭GitLab Issue
	go s.closeGitLabIssue(discussion.GitLabIssueID)

	return nil
}

// CreateReply 创建回复
func (s *DiscussionService) CreateReply(discussionID uint, req *models.DiscussionReplyRequest, userID uint) (*models.DiscussionReply, error) {
	// 获取话题信息
	var discussion models.Discussion
	if err := s.db.First(&discussion, discussionID).Error; err != nil {
		return nil, fmt.Errorf("话题不存在: %w", err)
	}

	// 验证权限
	if !s.permissionService.CanReplyDiscussion(userID, discussion.ProjectID) {
		return nil, fmt.Errorf("没有权限回复此话题")
	}

	// 创建GitLab Issue Note
	gitlabNote, err := s.createGitLabNote(discussion.GitLabIssueID, req.Content)
	if err != nil {
		return nil, fmt.Errorf("创建GitLab回复失败: %w", err)
	}

	// 创建本地回复记录
	reply := &models.DiscussionReply{
		DiscussionID:  discussionID,
		AuthorID:      userID,
		Content:       req.Content,
		GitLabNoteID:  gitlabNote.ID,
		GitLabNoteURL: gitlabNote.WebURL,
		ParentReplyID: req.ParentReplyID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.db.Create(reply).Error; err != nil {
		return nil, fmt.Errorf("创建回复记录失败: %w", err)
	}

	// 更新话题回复数
	s.db.Model(&discussion).UpdateColumn("reply_count", gorm.Expr("reply_count + ?", 1))

	// 预加载关联数据
	if err := s.db.Preload("Author").First(reply, reply.ID).Error; err != nil {
		return nil, fmt.Errorf("加载回复数据失败: %w", err)
	}

	return reply, nil
}

// LikeDiscussion 点赞话题
func (s *DiscussionService) LikeDiscussion(discussionID uint, userID uint) error {
	// 检查是否已点赞
	var existing models.DiscussionLike
	if err := s.db.Where("discussion_id = ? AND user_id = ?", discussionID, userID).First(&existing).Error; err == nil {
		return fmt.Errorf("已经点赞过此话题")
	}

	// 创建点赞记录
	like := &models.DiscussionLike{
		DiscussionID: discussionID,
		UserID:       userID,
		CreatedAt:    time.Now(),
	}

	if err := s.db.Create(like).Error; err != nil {
		return fmt.Errorf("点赞失败: %w", err)
	}

	// 更新点赞数
	s.db.Model(&models.Discussion{}).Where("id = ?", discussionID).
		UpdateColumn("like_count", gorm.Expr("like_count + ?", 1))

	return nil
}

// UnlikeDiscussion 取消点赞
func (s *DiscussionService) UnlikeDiscussion(discussionID uint, userID uint) error {
	// 删除点赞记录
	if err := s.db.Where("discussion_id = ? AND user_id = ?", discussionID, userID).
		Delete(&models.DiscussionLike{}).Error; err != nil {
		return fmt.Errorf("取消点赞失败: %w", err)
	}

	// 更新点赞数
	s.db.Model(&models.Discussion{}).Where("id = ?", discussionID).
		UpdateColumn("like_count", gorm.Expr("like_count - ?", 1))

	return nil
}

// PinDiscussion 置顶话题
func (s *DiscussionService) PinDiscussion(discussionID uint, userID uint) error {
	// 获取话题信息
	var discussion models.Discussion
	if err := s.db.First(&discussion, discussionID).Error; err != nil {
		return fmt.Errorf("话题不存在: %w", err)
	}

	// 验证权限（只有教师和管理员可以置顶）
	if !s.permissionService.IsTeacherOrAdmin(userID, discussion.ProjectID) {
		return fmt.Errorf("没有权限置顶此话题")
	}

	// 更新置顶状态
	if err := s.db.Model(&discussion).Update("is_pinned", true).Error; err != nil {
		return fmt.Errorf("置顶失败: %w", err)
	}

	return nil
}

// GetCategories 获取话题分类
func (s *DiscussionService) GetCategories() []string {
	return []string{"general", "question", "announcement", "help", "feedback"}
}

// SyncFromGitLab 从GitLab同步话题数据
func (s *DiscussionService) SyncFromGitLab(projectID uint) error {
	// 获取项目信息
	var project models.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return fmt.Errorf("项目不存在: %w", err)
	}

	// 获取GitLab讨论
	gitlabIssues, err := s.gitlabService.GetDiscussions(project.GitLabProjectID)
	if err != nil {
		return fmt.Errorf("获取GitLab讨论失败: %w", err)
	}

	// 同步到本地数据库
	for _, issue := range gitlabIssues {
		s.syncDiscussionFromGitLab(issue, projectID)
	}

	return nil
}

// 私有方法

// recordView 记录浏览
func (s *DiscussionService) recordView(discussionID uint, userID uint) {
	// 检查是否已记录（24小时内）
	var existing models.DiscussionView
	if err := s.db.Where("discussion_id = ? AND user_id = ? AND viewed_at > ?",
		discussionID, userID, time.Now().Add(-24*time.Hour)).First(&existing).Error; err == nil {
		return
	}

	// 创建浏览记录
	view := &models.DiscussionView{
		DiscussionID: discussionID,
		UserID:       userID,
		ViewedAt:     time.Now(),
	}
	s.db.Create(view)

	// 更新浏览数
	s.db.Model(&models.Discussion{}).Where("id = ?", discussionID).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))
}

// getValidCategory 获取有效的分类
func (s *DiscussionService) getValidCategory(category string) string {
	validCategories := map[string]bool{
		"general":      true,
		"question":     true,
		"announcement": true,
		"help":         true,
		"feedback":     true,
	}

	if validCategories[category] {
		return category
	}
	return "general"
}

// syncToGitLab 同步到GitLab
func (s *DiscussionService) syncToGitLab(issueID int, title, content string) {
	// 这里可以调用GitLab API更新Issue
	// 暂时留空，可以根据需要实现
}

// closeGitLabIssue 关闭GitLab Issue
func (s *DiscussionService) closeGitLabIssue(issueID int) {
	// 这里可以调用GitLab API关闭Issue
	// 暂时留空，可以根据需要实现
}

// createGitLabNote 创建GitLab Note
func (s *DiscussionService) createGitLabNote(issueID int, content string) (*gitlab.Note, error) {
	// 这里需要调用GitLab API创建Note
	// 暂时返回模拟数据
	return &gitlab.Note{
		ID:     int(time.Now().Unix()),
		WebURL: fmt.Sprintf("https://gitlab.com/issues/%d#note_%d", issueID, time.Now().Unix()),
		Body:   content,
	}, nil
}

// syncDiscussionFromGitLab 从GitLab同步单个话题
func (s *DiscussionService) syncDiscussionFromGitLab(issue *gitlab.Issue, projectID uint) {
	// 检查是否已存在
	var existing models.Discussion
	if err := s.db.Where("gitlab_issue_id = ?", issue.ID).First(&existing).Error; err == nil {
		return // 已存在，跳过
	}

	// 获取作者信息
	var author models.User
	if err := s.db.Where("gitlab_id = ?", issue.Author.ID).First(&author).Error; err != nil {
		return // 作者不存在，跳过
	}

	// 创建讨论记录
	discussion := &models.Discussion{
		Title:          issue.Title,
		Content:        issue.Description,
		ProjectID:      projectID,
		AuthorID:       author.ID,
		GitLabIssueID:  issue.ID,
		GitLabIssueURL: issue.WebURL,
		Status:         issue.State,
		CreatedAt:      *issue.CreatedAt,
		UpdatedAt:      *issue.UpdatedAt,
	}

	s.db.Create(discussion)
}
