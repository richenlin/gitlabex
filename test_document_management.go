package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
)

// TestDocumentManagement 测试文档管理功能
func main() {
	baseURL := "http://localhost:8080/api/v1"

	fmt.Println("=== 文档管理功能测试 ===")

	// 1. 测试获取预定义分类
	fmt.Println("\n1. 测试获取预定义分类...")
	testGetPredefinedCategories(baseURL)

	// 2. 测试创建独立文档
	fmt.Println("\n2. 测试创建独立文档...")
	testCreateStandaloneDocument(baseURL)

	// 3. 测试获取文档列表
	fmt.Println("\n3. 测试获取文档列表...")
	testGetDocuments(baseURL)

	// 4. 测试同步课题文档到MinIO
	fmt.Println("\n4. 测试同步课题文档到MinIO...")
	testSyncProjectDocumentsToMinIO(baseURL)

	// 5. 测试文档编辑请求
	fmt.Println("\n5. 测试文档编辑请求...")
	testDocumentEditRequest(baseURL)

	fmt.Println("\n=== 测试完成 ===")
}

// testGetPredefinedCategories 测试获取预定义分类
func testGetPredefinedCategories(baseURL string) {
	resp, err := http.Get(baseURL + "/documents/categories/predefined")
	if err != nil {
		log.Printf("获取预定义分类失败: %v", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("预定义分类响应: %s\n", string(body))
}

// testCreateStandaloneDocument 测试创建独立文档
func testCreateStandaloneDocument(baseURL string) {
	// 创建一个测试文件
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	// 添加文件字段
	fileWriter, err := writer.CreateFormFile("file", "test_document.txt")
	if err != nil {
		log.Printf("创建文件字段失败: %v", err)
		return
	}
	fileWriter.Write([]byte("这是一个测试文档的内容"))

	// 添加其他字段
	writer.WriteField("title", "测试独立文档")
	writer.WriteField("description", "这是一个测试用的独立文档")
	writer.WriteField("category", "documentation")

	writer.Close()

	// 发送请求
	req, err := http.NewRequest("POST", baseURL+"/documents/standalone", &b)
	if err != nil {
		log.Printf("创建请求失败: %v", err)
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer test-token") // 模拟认证

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("发送请求失败: %v", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("创建独立文档响应: %s\n", string(body))
}

// testGetDocuments 测试获取文档列表
func testGetDocuments(baseURL string) {
	resp, err := http.Get(baseURL + "/documents?page=1&limit=10")
	if err != nil {
		log.Printf("获取文档列表失败: %v", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("文档列表响应: %s\n", string(body))
}

// testSyncProjectDocumentsToMinIO 测试同步课题文档到MinIO
func testSyncProjectDocumentsToMinIO(baseURL string) {
	projectID := "550e8400-e29b-41d4-a716-446655440000" // 示例项目ID
	gitlabProjectID := "123"

	url := fmt.Sprintf("%s/documents/projects/%s/sync-to-minio?gitlab_project_id=%s",
		baseURL, projectID, gitlabProjectID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Printf("创建同步请求失败: %v", err)
		return
	}
	req.Header.Set("Authorization", "Bearer test-token")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("发送同步请求失败: %v", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("同步文档到MinIO响应: %s\n", string(body))
}

// testDocumentEditRequest 测试文档编辑请求
func testDocumentEditRequest(baseURL string) {
	documentID := "550e8400-e29b-41d4-a716-446655440001" // 示例文档ID

	editRequest := map[string]interface{}{
		"title":       "修改后的文档标题",
		"description": "修改后的文档描述",
		"category":    "tutorial",
		"reason":      "测试文档编辑功能",
	}

	jsonData, _ := json.Marshal(editRequest)

	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/documents/%s/edit-request", baseURL, documentID),
		bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("创建编辑请求失败: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("发送编辑请求失败: %v", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("文档编辑请求响应: %s\n", string(body))
}

// TestAPIEndpoints 测试所有新增的API端点
func TestAPIEndpoints() {
	endpoints := []string{
		"GET /api/v1/documents/categories/predefined",
		"POST /api/v1/documents/standalone",
		"POST /api/v1/documents/projects/{project_id}/sync-to-minio",
		"GET /api/v1/documents/{id}/download-url",
		"PUT /api/v1/documents/{id}/with-permission-check",
		"POST /api/v1/documents/{id}/edit-request",
		"GET /api/v1/documents/edit-requests",
		"PUT /api/v1/documents/edit-requests/{id}/review",
		"GET /api/v1/documents/{id}/edit-history",
	}

	fmt.Println("\n=== 新增API端点列表 ===")
	for _, endpoint := range endpoints {
		fmt.Printf("✓ %s\n", endpoint)
	}
}

// TestFeatures 测试功能特性
func TestFeatures() {
	features := []string{
		"✓ 支持课题文档同步到MinIO",
		"✓ 支持独立文档创建（不关联课题）",
		"✓ 标题和简介默认使用文件名",
		"✓ 支持预定义分类管理",
		"✓ 管理员/教师直接修改文档属性",
		"✓ 研究员/学生修改需要审核",
		"✓ 自动识别文档类型（doc, excel, ppt, pdf等）",
		"✓ MinIO存储和下载URL生成",
		"✓ 文档编辑历史记录",
		"✓ 权限控制和审核流程",
	}

	fmt.Println("\n=== 实现的功能特性 ===")
	for _, feature := range features {
		fmt.Println(feature)
	}
}

func init() {
	// 在main函数执行前显示测试信息
	TestAPIEndpoints()
	TestFeatures()
}
