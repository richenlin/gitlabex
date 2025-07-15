package services

import (
	"fmt"
	"strings"
	"time"

	"gitlabex/internal/models"

	"github.com/xanzy/go-gitlab"
	"gorm.io/gorm"
)

// DocumentService 文档管理服务
type DocumentService struct {
	db                *gorm.DB
	gitlabService     *GitLabService
	onlyofficeService *OnlyOfficeService
}

// NewDocumentService 创建文档管理服务
func NewDocumentService(db *gorm.DB, gitlabService *GitLabService, onlyofficeService *OnlyOfficeService) *DocumentService {
	return &DocumentService{
		db:                db,
		gitlabService:     gitlabService,
		onlyofficeService: onlyofficeService,
	}
}

// CreateDocumentRequest 创建文档请求
type CreateDocumentRequest struct {
	Title     string `json:"title" binding:"required"`
	Content   string `json:"content"`
	ProjectID uint   `json:"project_id" binding:"required"`
	Type      string `json:"type"`      // wiki, markdown, office
	Format    string `json:"format"`    // markdown, html
	IsPublic  bool   `json:"is_public"` // 是否公开
	ParentID  uint   `json:"parent_id"` // 父文档ID
}

// UpdateDocumentRequest 更新文档请求
type UpdateDocumentRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Format   string `json:"format"`
	IsPublic *bool  `json:"is_public"`
}

// CreateDocument 创建文档
func (s *DocumentService) CreateDocument(userID uint, req *CreateDocumentRequest) (*models.Document, error) {
	// 验证项目是否存在
	var project models.Project
	if err := s.db.First(&project, req.ProjectID).Error; err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// 设置默认值
	if req.Type == "" {
		req.Type = "wiki"
	}
	if req.Format == "" {
		req.Format = "markdown"
	}

	// 创建本地文档记录
	document := &models.Document{
		Title:     req.Title,
		Content:   req.Content,
		ProjectID: req.ProjectID,
		AuthorID:  userID,
		Type:      req.Type,
		Format:    req.Format,
		IsPublic:  req.IsPublic,
		ParentID:  req.ParentID,
		Status:    "active",
		Version:   1,
	}

	if err := s.db.Create(document).Error; err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	// 如果项目启用了GitLab并且是Wiki类型，同步到GitLab
	if project.GitLabProjectID > 0 && req.Type == "wiki" && project.WikiEnabled {
		// 生成Wiki页面的slug
		slug := s.generateWikiSlug(req.Title)

		// 创建GitLab Wiki页面
		_, err := s.gitlabService.CreateWikiPage(project.GitLabProjectID, req.Title, req.Content)
		if err != nil {
			// Wiki创建失败不影响本地文档创建
			fmt.Printf("Warning: Failed to create GitLab wiki page: %v\n", err)
		} else {
			// 更新文档的GitLab信息
			document.GitLabWikiSlug = slug
			document.GitLabWikiURL = fmt.Sprintf("%s/-/wikis/%s", project.GitLabURL, slug)
			s.db.Save(document)
		}
	}

	// 预加载关联数据
	if err := s.db.Preload("Author").Preload("Project").First(document, document.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load document: %w", err)
	}

	return document, nil
}

// GetDocumentByID 根据ID获取文档
func (s *DocumentService) GetDocumentByID(documentID uint) (*models.Document, error) {
	var document models.Document
	err := s.db.Preload("Author").
		Preload("Project").
		Preload("Parent").
		First(&document, documentID).Error

	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	return &document, nil
}

// GetDocumentsByProject 获取项目的文档列表
func (s *DocumentService) GetDocumentsByProject(projectID uint) ([]models.Document, error) {
	var documents []models.Document
	err := s.db.Preload("Author").
		Where("project_id = ? AND status = 'active'", projectID).
		Order("created_at DESC").
		Find(&documents).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get documents: %w", err)
	}

	return documents, nil
}

// GetDocumentsByUser 获取用户的文档列表
func (s *DocumentService) GetDocumentsByUser(userID uint) ([]models.Document, error) {
	var documents []models.Document
	err := s.db.Preload("Author").
		Preload("Project").
		Where("author_id = ? AND status = 'active'", userID).
		Order("created_at DESC").
		Find(&documents).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get documents: %w", err)
	}

	return documents, nil
}

// UpdateDocument 更新文档
func (s *DocumentService) UpdateDocument(documentID uint, req *UpdateDocumentRequest) (*models.Document, error) {
	var document models.Document
	if err := s.db.Preload("Project").First(&document, documentID).Error; err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	// 更新字段
	if req.Title != "" {
		document.Title = req.Title
	}
	if req.Content != "" {
		document.Content = req.Content
	}
	if req.Format != "" {
		document.Format = req.Format
	}
	if req.IsPublic != nil {
		document.IsPublic = *req.IsPublic
	}

	// 版本号递增
	document.Version++
	document.UpdatedAt = time.Now()

	if err := s.db.Save(&document).Error; err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	// 如果是Wiki类型且有GitLab集成，同步更新
	if document.Type == "wiki" && document.GitLabWikiSlug != "" && document.Project.GitLabProjectID > 0 {
		_, err := s.gitlabService.UpdateWikiPage(
			document.Project.GitLabProjectID,
			document.GitLabWikiSlug,
			document.Title,
			document.Content,
		)
		if err != nil {
			fmt.Printf("Warning: Failed to update GitLab wiki page: %v\n", err)
		}
	}

	return &document, nil
}

// DeleteDocument 删除文档
func (s *DocumentService) DeleteDocument(documentID uint) error {
	var document models.Document
	if err := s.db.Preload("Project").First(&document, documentID).Error; err != nil {
		return fmt.Errorf("document not found: %w", err)
	}

	// 软删除
	document.Status = "deleted"
	document.DeletedAt = time.Now()

	if err := s.db.Save(&document).Error; err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	// 如果是Wiki类型且有GitLab集成，需要手动处理（GitLab API可能不支持删除Wiki页面）
	if document.Type == "wiki" && document.GitLabWikiSlug != "" && document.Project.GitLabProjectID > 0 {
		fmt.Printf("Note: GitLab Wiki page should be manually deleted: %s\n", document.GitLabWikiURL)
	}

	return nil
}

// GetProjectWikiPages 获取项目的Wiki页面列表（从GitLab同步）
func (s *DocumentService) GetProjectWikiPages(projectID uint) ([]*gitlab.Wiki, error) {
	var project models.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	if project.GitLabProjectID == 0 {
		return nil, fmt.Errorf("project not linked to GitLab")
	}

	// 获取GitLab Wiki页面
	pages, err := s.gitlabService.GetWikiPages(project.GitLabProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wiki pages: %w", err)
	}

	return pages, nil
}

// SyncWikiFromGitLab 从GitLab同步Wiki页面到本地
func (s *DocumentService) SyncWikiFromGitLab(projectID uint) error {
	var project models.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	if project.GitLabProjectID == 0 {
		return fmt.Errorf("project not linked to GitLab")
	}

	// 获取GitLab Wiki页面
	pages, err := s.gitlabService.GetWikiPages(project.GitLabProjectID)
	if err != nil {
		return fmt.Errorf("failed to get wiki pages: %w", err)
	}

	// 同步每个页面
	for _, page := range pages {
		// 检查本地是否已存在
		var existingDoc models.Document
		if err := s.db.Where("project_id = ? AND gitlab_wiki_slug = ?", projectID, page.Slug).First(&existingDoc).Error; err == nil {
			// 更新现有文档
			existingDoc.Title = page.Title
			existingDoc.Content = page.Content
			existingDoc.Format = string(page.Format)
			existingDoc.UpdatedAt = time.Now()
			s.db.Save(&existingDoc)
		} else {
			// 创建新文档
			newDoc := &models.Document{
				Title:          page.Title,
				Content:        page.Content,
				ProjectID:      projectID,
				Type:           "wiki",
				Format:         string(page.Format),
				IsPublic:       true,
				Status:         "active",
				Version:        1,
				GitLabWikiSlug: page.Slug,
				GitLabWikiURL:  fmt.Sprintf("%s/-/wikis/%s", project.GitLabURL, page.Slug),
				AuthorID:       project.TeacherID, // 默认作者为项目创建者
			}
			s.db.Create(newDoc)
		}
	}

	return nil
}

// CheckWikiEditPermission 检查Wiki编辑权限
func (s *DocumentService) CheckWikiEditPermission(userID uint, projectID uint) (bool, error) {
	var project models.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return false, fmt.Errorf("project not found: %w", err)
	}

	// 获取用户信息
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return false, fmt.Errorf("user not found: %w", err)
	}

	// 检查用户角色
	if user.Role == 1 || user.Role == 2 { // Admin or Teacher
		return true, nil
	}

	// 检查是否是项目成员
	var member models.ProjectMember
	if err := s.db.Where("project_id = ? AND student_id = ? AND status = 'active'", projectID, userID).First(&member).Error; err != nil {
		return false, nil
	}

	// 如果有GitLab集成，检查GitLab权限
	if project.GitLabProjectID > 0 && user.GitLabID > 0 {
		return s.gitlabService.CheckWikiEditPermission(user.GitLabID, project.GitLabProjectID)
	}

	// 默认项目成员可以编辑Wiki
	return true, nil
}

// generateWikiSlug 生成Wiki页面的slug
func (s *DocumentService) generateWikiSlug(title string) string {
	// 简单的slug生成逻辑，可以根据需要优化
	return title
}

// GetDocumentTree 获取文档树结构
func (s *DocumentService) GetDocumentTree(projectID uint) ([]models.Document, error) {
	var documents []models.Document
	err := s.db.Preload("Author").
		Where("project_id = ? AND parent_id = 0 AND status = 'active'", projectID).
		Order("created_at DESC").
		Find(&documents).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get document tree: %w", err)
	}

	// 递归加载子文档
	for i := range documents {
		children, err := s.getDocumentChildren(documents[i].ID)
		if err != nil {
			continue
		}
		documents[i].Children = children
	}

	return documents, nil
}

// getDocumentChildren 递归获取子文档
func (s *DocumentService) getDocumentChildren(parentID uint) ([]models.Document, error) {
	var children []models.Document
	err := s.db.Preload("Author").
		Where("parent_id = ? AND status = 'active'", parentID).
		Order("created_at DESC").
		Find(&children).Error

	if err != nil {
		return nil, err
	}

	// 递归加载子文档的子文档
	for i := range children {
		grandchildren, err := s.getDocumentChildren(children[i].ID)
		if err != nil {
			continue
		}
		children[i].Children = grandchildren
	}

	return children, nil
}

// GetDocumentHistory 获取文档历史版本
func (s *DocumentService) GetDocumentHistory(documentID uint) ([]models.DocumentHistory, error) {
	var history []models.DocumentHistory
	err := s.db.Preload("Author").
		Where("document_id = ?", documentID).
		Order("created_at DESC").
		Find(&history).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get document history: %w", err)
	}

	return history, nil
}

// CreateDocumentHistory 创建文档历史记录
func (s *DocumentService) CreateDocumentHistory(documentID uint, userID uint, content, changeNote string) error {
	history := models.DocumentHistory{
		DocumentID: documentID,
		AuthorID:   userID,
		Content:    content,
		ChangeNote: changeNote,
	}

	return s.db.Create(&history).Error
}

// GetWikiAttachments 获取Wiki页面附件列表
func (s *DocumentService) GetWikiAttachments(projectID int, slug string) ([]models.DocumentAttachment, error) {
	var attachments []models.DocumentAttachment

	// 查找项目
	var project models.Project
	if err := s.db.Where("gitlab_project_id = ?", projectID).First(&project).Error; err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// 查找文档
	var document models.Document
	if err := s.db.Where("project_id = ? AND slug = ?", project.ID, slug).First(&document).Error; err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	// 获取文档的附件
	if err := s.db.Where("document_id = ?", document.ID).Find(&attachments).Error; err != nil {
		return nil, fmt.Errorf("failed to get attachments: %w", err)
	}

	return attachments, nil
}

// CreateWikiAttachment 创建Wiki页面附件
func (s *DocumentService) CreateWikiAttachment(userID int, projectID int, slug string, filename string, url string, content []byte) (*models.DocumentAttachment, error) {
	// 查找项目
	var project models.Project
	if err := s.db.Where("gitlab_project_id = ?", projectID).First(&project).Error; err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// 查找或创建文档
	var document models.Document
	if err := s.db.Where("project_id = ? AND gitlab_wiki_slug = ?", project.ID, slug).First(&document).Error; err != nil {
		// 如果文档不存在，创建一个新的
		document = models.Document{
			Title:          slug,
			Content:        "",
			ProjectID:      project.ID,
			AuthorID:       uint(userID),
			Status:         "active",
			GitLabWikiSlug: slug,
		}
		if err := s.db.Create(&document).Error; err != nil {
			return nil, fmt.Errorf("failed to create document: %w", err)
		}
	}

	// 创建附件
	attachment := models.DocumentAttachment{
		UserID:       userID,
		ProjectID:    &projectID,
		WikiPageSlug: slug,
		FileName:     filename,
		FileURL:      url,
		FilePath:     url,
		FileType:     getFileType(filename),
		DocumentKey:  fmt.Sprintf("%d_%s_%d", projectID, slug, time.Now().Unix()),
		Status:       "active",
	}

	if err := s.db.Create(&attachment).Error; err != nil {
		return nil, fmt.Errorf("failed to create attachment: %w", err)
	}

	return &attachment, nil
}

// getFileType 根据文件名获取文件类型
func getFileType(filename string) string {
	switch {
	case strings.HasSuffix(filename, ".doc") || strings.HasSuffix(filename, ".docx"):
		return "document"
	case strings.HasSuffix(filename, ".xls") || strings.HasSuffix(filename, ".xlsx"):
		return "spreadsheet"
	case strings.HasSuffix(filename, ".ppt") || strings.HasSuffix(filename, ".pptx"):
		return "presentation"
	case strings.HasSuffix(filename, ".pdf"):
		return "pdf"
	default:
		return "file"
	}
}

// 使用已定义的角色常量
