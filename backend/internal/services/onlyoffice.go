package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gitlabex/internal/config"
	"gitlabex/internal/models"

	"gorm.io/gorm"
)

// OnlyOfficeService OnlyOffice文档编辑器服务
type OnlyOfficeService struct {
	config *config.Config
	db     *gorm.DB
}

// NewOnlyOfficeService 创建OnlyOffice服务
func NewOnlyOfficeService(cfg *config.Config, db *gorm.DB) *OnlyOfficeService {
	return &OnlyOfficeService{
		config: cfg,
		db:     db,
	}
}

// DocumentConfig OnlyOffice文档配置
type DocumentConfig struct {
	Document Document `json:"document"`
	Editor   Editor   `json:"editor"`
	Callback string   `json:"callbackUrl"`
	Token    string   `json:"token"`
	Type     string   `json:"type"`
	Width    string   `json:"width"`
	Height   string   `json:"height"`
	Embedded Embedded `json:"embedded,omitempty"`
}

// Document 文档信息
type Document struct {
	FileType    string      `json:"fileType"`
	Key         string      `json:"key"`
	Title       string      `json:"title"`
	URL         string      `json:"url"`
	Permissions Permissions `json:"permissions"`
}

// Editor 编辑器配置
type Editor struct {
	CallbackURL string `json:"callbackUrl"`
	Lang        string `json:"lang"`
	Mode        string `json:"mode"`
	User        User   `json:"user"`
}

// User 用户信息
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Permissions 权限配置
type Permissions struct {
	Comment      bool `json:"comment"`
	Download     bool `json:"download"`
	Edit         bool `json:"edit"`
	FillForms    bool `json:"fillForms"`
	ModifyFilter bool `json:"modifyFilter"`
	Print        bool `json:"print"`
	Review       bool `json:"review"`
}

// Embedded 嵌入式配置
type Embedded struct {
	SaveURL       string `json:"saveUrl"`
	ShareURL      string `json:"shareUrl"`
	ToolbarDocked string `json:"toolbarDocked"`
}

// CallbackData 回调数据
type CallbackData struct {
	Key        string          `json:"key"`
	Status     int             `json:"status"`
	URL        string          `json:"url"`
	ChangesURL string          `json:"changesurl"`
	History    json.RawMessage `json:"history"`
	Users      []string        `json:"users"`
	Actions    []Action        `json:"actions"`
	Token      string          `json:"token"`
}

// Action 操作信息
type Action struct {
	Type   int    `json:"type"`
	UserID string `json:"userid"`
}

// CreateDocumentSession 创建文档编辑会话
func (s *OnlyOfficeService) CreateDocumentSession(userID int, filename string, fileContent []byte, mode string) (*models.DocumentAttachment, error) {
	// 生成文档key
	key := s.generateDocumentKey(filename)

	// 保存文件到本地
	filePath := filepath.Join("uploads", key+"_"+filename)
	if err := s.saveFile(filePath, fileContent); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// 创建文档记录
	doc := &models.DocumentAttachment{
		UserID:      userID,
		FileName:    filename,
		FileType:    s.getFileType(filename),
		DocumentKey: key,
		FilePath:    filePath,
		EditMode:    mode,
		Status:      "editing",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(doc).Error; err != nil {
		return nil, fmt.Errorf("failed to create document record: %w", err)
	}

	return doc, nil
}

// GetDocumentConfig 获取文档配置
func (s *OnlyOfficeService) GetDocumentConfig(docID int, userID int) (*DocumentConfig, error) {
	var doc models.DocumentAttachment
	if err := s.db.Where("id = ? AND user_id = ?", docID, userID).First(&doc).Error; err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	// 构建文档URL
	documentURL := fmt.Sprintf("http://localhost:8080/api/documents/%d/content", doc.ID)
	callbackURL := fmt.Sprintf("http://localhost:8080/api/documents/%d/callback", doc.ID)

	config := &DocumentConfig{
		Type:   "desktop",
		Width:  "100%",
		Height: "100%",
		Document: Document{
			FileType: doc.FileType,
			Key:      doc.DocumentKey,
			Title:    doc.FileName,
			URL:      documentURL,
			Permissions: Permissions{
				Comment:      true,
				Download:     true,
				Edit:         doc.EditMode == "edit",
				FillForms:    true,
				ModifyFilter: true,
				Print:        true,
				Review:       true,
			},
		},
		Editor: Editor{
			CallbackURL: callbackURL,
			Lang:        "zh-CN",
			Mode:        doc.EditMode,
			User: User{
				ID:   fmt.Sprintf("%d", userID),
				Name: fmt.Sprintf("User_%d", userID),
			},
		},
		Callback: callbackURL,
	}

	// 生成JWT令牌
	if s.config.OnlyOffice.JWTSecret != "" {
		token, err := s.generateJWTToken(config)
		if err != nil {
			return nil, fmt.Errorf("failed to generate JWT token: %w", err)
		}
		config.Token = token
	}

	return config, nil
}

// HandleCallback 处理OnlyOffice回调
func (s *OnlyOfficeService) HandleCallback(docID int, callbackData *CallbackData) error {
	var doc models.DocumentAttachment
	if err := s.db.Where("id = ?", docID).First(&doc).Error; err != nil {
		return fmt.Errorf("document not found: %w", err)
	}

	// 验证JWT令牌
	if s.config.OnlyOffice.JWTSecret != "" {
		if !s.verifyJWTToken(callbackData.Token, callbackData) {
			return fmt.Errorf("invalid JWT token")
		}
	}

	// 处理不同的状态
	switch callbackData.Status {
	case 1: // 编辑中
		doc.Status = "editing"
	case 2: // 准备保存
		doc.Status = "saving"
		if callbackData.URL != "" {
			if err := s.downloadAndSaveDocument(callbackData.URL, doc.FilePath); err != nil {
				return fmt.Errorf("failed to save document: %w", err)
			}
		}
	case 3: // 保存出错
		doc.Status = "error"
	case 4: // 关闭无更改
		doc.Status = "closed"
	case 6: // 正在编辑，但当前文档状态已保存
		doc.Status = "editing"
		if callbackData.URL != "" {
			if err := s.downloadAndSaveDocument(callbackData.URL, doc.FilePath); err != nil {
				return fmt.Errorf("failed to save document: %w", err)
			}
		}
	case 7: // 强制保存时出错
		doc.Status = "error"
	}

	doc.UpdatedAt = time.Now()
	if err := s.db.Save(&doc).Error; err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	return nil
}

// GetDocumentContent 获取文档内容
func (s *OnlyOfficeService) GetDocumentContent(docID int, userID int) ([]byte, string, error) {
	var doc models.DocumentAttachment
	if err := s.db.Where("id = ? AND user_id = ?", docID, userID).First(&doc).Error; err != nil {
		return nil, "", fmt.Errorf("document not found: %w", err)
	}

	content, err := os.ReadFile(doc.FilePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}

	return content, doc.FileType, nil
}

// 辅助方法

// generateDocumentKey 生成文档key
func (s *OnlyOfficeService) generateDocumentKey(filename string) string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// getFileType 获取文件类型
func (s *OnlyOfficeService) getFileType(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".docx":
		return "docx"
	case ".xlsx":
		return "xlsx"
	case ".pptx":
		return "pptx"
	case ".pdf":
		return "pdf"
	default:
		return "txt"
	}
}

// saveFile 保存文件
func (s *OnlyOfficeService) saveFile(filePath string, content []byte) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filePath, content, 0644)
}

// downloadAndSaveDocument 下载并保存文档
func (s *OnlyOfficeService) downloadAndSaveDocument(url, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return s.saveFile(filePath, content)
}

// generateJWTToken 生成JWT令牌
func (s *OnlyOfficeService) generateJWTToken(data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, []byte(s.config.OnlyOffice.JWTSecret))
	mac.Write(jsonData)
	signature := base64.URLEncoding.EncodeToString(mac.Sum(nil))

	return signature, nil
}

// verifyJWTToken 验证JWT令牌
func (s *OnlyOfficeService) verifyJWTToken(token string, data interface{}) bool {
	expectedToken, err := s.generateJWTToken(data)
	if err != nil {
		return false
	}

	return hmac.Equal([]byte(token), []byte(expectedToken))
}
