package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"path/filepath"
	"strings"
	"time"

	"gitlabex/internal/models"

	"github.com/xanzy/go-gitlab"
	"gorm.io/gorm"
)

// InteractiveDevService 互动开发服务
type InteractiveDevService struct {
	db                *gorm.DB
	gitlabService     *GitLabService
	permissionService *PermissionService
}

// NewInteractiveDevService 创建互动开发服务
func NewInteractiveDevService(db *gorm.DB, gitlabService *GitLabService, permissionService *PermissionService) *InteractiveDevService {
	return &InteractiveDevService{
		db:                db,
		gitlabService:     gitlabService,
		permissionService: permissionService,
	}
}

// FileTreeNode 文件树节点
type FileTreeNode struct {
	Name         string          `json:"name"`
	Path         string          `json:"path"`
	Type         string          `json:"type"` // file, directory
	Size         int64           `json:"size"`
	LastModified time.Time       `json:"last_modified"`
	Language     string          `json:"language,omitempty"`
	IsEditable   bool            `json:"is_editable"`
	Children     []*FileTreeNode `json:"children,omitempty"`
}

// CodeEditRequest 代码编辑请求
type CodeEditRequest struct {
	ProjectID uint   `json:"project_id" binding:"required"`
	FilePath  string `json:"file_path" binding:"required"`
	Branch    string `json:"branch" binding:"required"`
	Content   string `json:"content" binding:"required"`
	Message   string `json:"message"`
}

// CreateBranchRequest 创建分支请求
type CreateBranchRequest struct {
	ProjectID    uint   `json:"project_id" binding:"required"`
	BranchName   string `json:"branch_name" binding:"required"`
	SourceBranch string `json:"source_branch"`
}

// FileOperationRequest 文件操作请求
type FileOperationRequest struct {
	ProjectID uint   `json:"project_id" binding:"required"`
	FilePath  string `json:"file_path" binding:"required"`
	Branch    string `json:"branch" binding:"required"`
	Operation string `json:"operation" binding:"required"` // create, delete, rename, move
	NewPath   string `json:"new_path,omitempty"`
	Content   string `json:"content,omitempty"`
	Message   string `json:"message"`
}

// EditSessionRequest 编辑会话请求
type EditSessionRequest struct {
	ProjectID uint   `json:"project_id" binding:"required"`
	FilePath  string `json:"file_path" binding:"required"`
	Branch    string `json:"branch" binding:"required"`
}

// GetProjectFileTree 获取项目文件树
func (s *InteractiveDevService) GetProjectFileTree(projectID uint, branch string, userID uint) (*FileTreeNode, error) {
	// 获取用户信息
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 检查权限
	if !s.permissionService.CanAccessProject(user.ID, projectID, "read") {
		return nil, fmt.Errorf("permission denied")
	}

	// 获取项目信息
	var project models.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	if branch == "" {
		branch = project.DefaultBranch
	}

	// 使用GitLab API获取项目树
	treeOptions := &gitlab.ListTreeOptions{
		Ref:       &branch,
		Recursive: gitlab.Bool(true),
	}

	tree, _, err := s.gitlabService.client.Repositories.ListTree(project.GitLabProjectID, treeOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get project tree: %w", err)
	}

	// 构建文件树
	root := &FileTreeNode{
		Name:     project.Name,
		Path:     "",
		Type:     "directory",
		Children: make([]*FileTreeNode, 0),
	}

	// 转换GitLab树结构
	for _, item := range tree {
		node := &FileTreeNode{
			Name:       item.Name,
			Path:       item.Path,
			Type:       string(item.Type),
			IsEditable: s.isFileEditable(item.Path, project.AllowedFileTypes),
		}

		if item.Type == "blob" {
			node.Language = s.detectLanguage(item.Path)
		}

		root.Children = append(root.Children, node)
	}

	return root, nil
}

// GetFileContent 获取文件内容
func (s *InteractiveDevService) GetFileContent(projectID uint, filePath, branch string, userID uint) (string, error) {
	// 获取用户信息
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	// 检查权限
	if !s.permissionService.CanAccessProject(user.ID, projectID, "read") {
		return "", fmt.Errorf("permission denied")
	}

	// 获取项目信息
	var project models.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return "", fmt.Errorf("project not found: %w", err)
	}

	if branch == "" {
		branch = project.DefaultBranch
	}

	// 使用GitLab API获取文件内容
	fileOptions := &gitlab.GetFileOptions{
		Ref: &branch,
	}

	file, _, err := s.gitlabService.client.RepositoryFiles.GetFile(project.GitLabProjectID, filePath, fileOptions)
	if err != nil {
		return "", fmt.Errorf("failed to get file content: %w", err)
	}

	content := string(file.Content)

	// 更新文件记录
	s.updateFileRecord(projectID, filePath, branch, content, userID)

	return content, nil
}

// SaveFileContent 保存文件内容
func (s *InteractiveDevService) SaveFileContent(req *CodeEditRequest, userID uint) error {
	// 获取用户信息
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 检查权限
	canWrite := s.permissionService.CanAccessProject(user.ID, req.ProjectID, "write")
	if !canWrite {
		return fmt.Errorf("permission denied")
	}

	// 获取项目信息
	var project models.Project
	if err := s.db.First(&project, req.ProjectID).Error; err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	// 检查主分支保护
	if project.MainBranchProtected && req.Branch == project.DefaultBranch {
		// 只有创建者和管理员可以编辑主分支
		if user.Role != 1 && project.TeacherID != user.ID {
			return fmt.Errorf("main branch is protected")
		}
	}

	// 检查文件类型
	if !s.isFileEditable(req.FilePath, project.AllowedFileTypes) {
		return fmt.Errorf("file type not allowed for editing")
	}

	// 检查文件大小
	if int64(len(req.Content)) > project.MaxFileSize {
		return fmt.Errorf("file size exceeds limit")
	}

	// 获取编辑锁
	if err := s.acquireEditLock(req.ProjectID, req.FilePath, req.Branch, userID); err != nil {
		return fmt.Errorf("failed to acquire edit lock: %w", err)
	}
	defer s.releaseEditLock(req.ProjectID, req.FilePath, req.Branch, userID)

	// 保存到GitLab
	commitMessage := req.Message
	if commitMessage == "" {
		commitMessage = fmt.Sprintf("Update %s", req.FilePath)
	}

	// 使用GitLab API更新文件
	updateOptions := &gitlab.UpdateFileOptions{
		Branch:        &req.Branch,
		Content:       &req.Content,
		CommitMessage: &commitMessage,
	}

	_, _, err := s.gitlabService.client.RepositoryFiles.UpdateFile(project.GitLabProjectID, req.FilePath, updateOptions)
	if err != nil {
		return fmt.Errorf("failed to save file to GitLab: %w", err)
	}

	// 更新文件记录
	s.updateFileRecord(req.ProjectID, req.FilePath, req.Branch, req.Content, userID)

	// 更新项目成员统计
	s.updateMemberStats(req.ProjectID, userID)

	return nil
}

// CreateStudentBranch 为学生创建个人分支
func (s *InteractiveDevService) CreateStudentBranch(projectID, studentID uint) error {
	// 获取项目信息
	var project models.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	// 检查学生是否已经有分支
	var member models.ProjectMember
	if err := s.db.Where("project_id = ? AND student_id = ?", projectID, studentID).First(&member).Error; err != nil {
		return fmt.Errorf("student not found in project: %w", err)
	}

	if member.PersonalBranch != "" {
		return fmt.Errorf("student already has a branch")
	}

	// 生成分支名
	var user models.User
	if err := s.db.First(&user, studentID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	branchName := fmt.Sprintf("%s-%s-%d", project.StudentBranchPrefix, user.Username, time.Now().Unix())

	// 使用现有的GitLab服务创建分支
	_, err := s.gitlabService.CreateStudentBranch(project.GitLabProjectID, int(studentID), branchName)
	if err != nil {
		return fmt.Errorf("failed to create branch in GitLab: %w", err)
	}

	// 更新成员记录
	now := time.Now()
	member.PersonalBranch = branchName
	member.PersonalBranchURL = fmt.Sprintf("%s/-/tree/%s", project.GitLabURL, branchName)
	member.BranchCreatedAt = &now
	member.LastActiveTime = &now

	if err := s.db.Save(&member).Error; err != nil {
		return fmt.Errorf("failed to update member record: %w", err)
	}

	return nil
}

// StartEditSession 开始编辑会话
func (s *InteractiveDevService) StartEditSession(req *EditSessionRequest, userID uint) (*models.CodeEditSession, error) {
	// 获取用户信息
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 检查权限
	if !s.permissionService.CanAccessProject(user.ID, req.ProjectID, "write") {
		return nil, fmt.Errorf("permission denied")
	}

	// 生成会话ID
	sessionID := s.generateSessionID()

	// 创建编辑会话
	session := &models.CodeEditSession{
		ProjectID: req.ProjectID,
		UserID:    userID,
		FilePath:  req.FilePath,
		Branch:    req.Branch,
		SessionID: sessionID,
		StartTime: time.Now(),
		Status:    "active",
		LastPing:  time.Now(),
	}

	if err := s.db.Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create edit session: %w", err)
	}

	return session, nil
}

// UpdateEditSession 更新编辑会话
func (s *InteractiveDevService) UpdateEditSession(sessionID string, userID uint) error {
	var session models.CodeEditSession
	if err := s.db.Where("session_id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// 更新心跳时间
	session.LastPing = time.Now()
	session.ChangesCount++

	return s.db.Save(&session).Error
}

// EndEditSession 结束编辑会话
func (s *InteractiveDevService) EndEditSession(sessionID string, userID uint) error {
	var session models.CodeEditSession
	if err := s.db.Where("session_id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// 更新会话状态
	now := time.Now()
	session.Status = "ended"
	session.EndTime = &now

	return s.db.Save(&session).Error
}

// GetActiveEditSessions 获取活跃编辑会话
func (s *InteractiveDevService) GetActiveEditSessions(projectID uint) ([]models.CodeEditSession, error) {
	var sessions []models.CodeEditSession

	// 获取5分钟内有活动的会话
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)

	err := s.db.Where("project_id = ? AND status = 'active' AND last_ping > ?", projectID, fiveMinutesAgo).
		Preload("User").
		Find(&sessions).Error

	return sessions, err
}

// 辅助函数

// generateSessionID 生成会话ID
func (s *InteractiveDevService) generateSessionID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 32

	b := make([]byte, length)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[num.Int64()]
	}

	return string(b)
}

// isFileEditable 检查文件是否可编辑
func (s *InteractiveDevService) isFileEditable(filePath string, allowedTypes []string) bool {
	if len(allowedTypes) == 0 {
		return true // 如果没有限制，所有文件都可编辑
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == "" {
		return true // 无扩展名文件（如README）可编辑
	}

	for _, allowedType := range allowedTypes {
		if ext == allowedType || ext == "."+allowedType {
			return true
		}
	}

	return false
}

// detectLanguage 检测文件语言
func (s *InteractiveDevService) detectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	languageMap := map[string]string{
		".js":         "javascript",
		".ts":         "typescript",
		".jsx":        "javascript",
		".tsx":        "typescript",
		".py":         "python",
		".java":       "java",
		".cpp":        "cpp",
		".c":          "c",
		".h":          "c",
		".hpp":        "cpp",
		".cs":         "csharp",
		".php":        "php",
		".rb":         "ruby",
		".go":         "go",
		".rs":         "rust",
		".kt":         "kotlin",
		".swift":      "swift",
		".html":       "html",
		".css":        "css",
		".scss":       "scss",
		".sass":       "sass",
		".less":       "less",
		".xml":        "xml",
		".json":       "json",
		".yaml":       "yaml",
		".yml":        "yaml",
		".md":         "markdown",
		".sql":        "sql",
		".sh":         "shell",
		".bash":       "shell",
		".zsh":        "shell",
		".ps1":        "powershell",
		".dockerfile": "dockerfile",
		".vue":        "vue",
		".svelte":     "svelte",
	}

	if lang, exists := languageMap[ext]; exists {
		return lang
	}

	return "plaintext"
}

// updateFileRecord 更新文件记录
func (s *InteractiveDevService) updateFileRecord(projectID uint, filePath, branch, content string, userID uint) {
	// 计算内容哈希
	hash := sha256.Sum256([]byte(content))
	contentHash := hex.EncodeToString(hash[:])

	// 查找或创建文件记录
	var file models.ProjectFile
	err := s.db.Where("project_id = ? AND file_path = ? AND branch = ?", projectID, filePath, branch).First(&file).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新文件记录
		file = models.ProjectFile{
			ProjectID:   projectID,
			FilePath:    filePath,
			FileName:    filepath.Base(filePath),
			FileType:    s.detectLanguage(filePath),
			FileSize:    int64(len(content)),
			Branch:      branch,
			Content:     content,
			ContentHash: contentHash,
			Language:    s.detectLanguage(filePath),
			CreatedBy:   userID,
			UpdatedBy:   userID,
		}
		s.db.Create(&file)
	} else {
		// 更新文件记录
		now := time.Now()
		file.Content = content
		file.ContentHash = contentHash
		file.FileSize = int64(len(content))
		file.UpdatedBy = userID
		file.LastEditedBy = userID
		file.LastEditedAt = &now
		s.db.Save(&file)
	}
}

// acquireEditLock 获取编辑锁
func (s *InteractiveDevService) acquireEditLock(projectID uint, filePath, branch string, userID uint) error {
	var file models.ProjectFile
	err := s.db.Where("project_id = ? AND file_path = ? AND branch = ?", projectID, filePath, branch).First(&file).Error

	if err == gorm.ErrRecordNotFound {
		return nil // 文件不存在，无需锁定
	}

	// 检查是否已被其他用户锁定
	if file.EditLockBy != 0 && file.EditLockBy != userID {
		if file.EditLockExpires != nil && time.Now().Before(*file.EditLockExpires) {
			return fmt.Errorf("file is locked by another user")
		}
	}

	// 设置锁定
	now := time.Now()
	expires := now.Add(5 * time.Minute) // 5分钟锁定期
	file.EditLockBy = userID
	file.EditLockAt = &now
	file.EditLockExpires = &expires

	return s.db.Save(&file).Error
}

// releaseEditLock 释放编辑锁
func (s *InteractiveDevService) releaseEditLock(projectID uint, filePath, branch string, userID uint) {
	var file models.ProjectFile
	err := s.db.Where("project_id = ? AND file_path = ? AND branch = ? AND edit_lock_by = ?",
		projectID, filePath, branch, userID).First(&file).Error

	if err == nil {
		file.EditLockBy = 0
		file.EditLockAt = nil
		file.EditLockExpires = nil
		s.db.Save(&file)
	}
}

// updateMemberStats 更新成员统计
func (s *InteractiveDevService) updateMemberStats(projectID, userID uint) {
	var member models.ProjectMember
	err := s.db.Where("project_id = ? AND student_id = ?", projectID, userID).First(&member).Error

	if err == nil {
		now := time.Now()
		member.FilesModified++
		member.LastEditTime = &now
		member.LastActiveTime = &now
		s.db.Save(&member)
	}
}

// GetCurrentMember 获取当前用户的项目成员信息
func (s *InteractiveDevService) GetCurrentMember(projectID, userID uint) (*models.ProjectMember, error) {
	var member models.ProjectMember
	err := s.db.Where("project_id = ? AND student_id = ?", projectID, userID).First(&member).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &member, nil
}
