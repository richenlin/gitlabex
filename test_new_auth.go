package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	GitLabAccessToken string `json:"gitlab_access_token"`
	GitLabUserID      int64  `json:"gitlab_user_id"`
	jwt.RegisteredClaims
}

func main() {
	// 使用真实的GitLab访问令牌（需要替换为实际的token）
	gitlabAccessToken := "glpat-xxxxxxxxxxxxxxxxxxxx" // 需要替换为真实的GitLab Personal Access Token
	gitlabUserID := int64(1)                          // 需要替换为真实的GitLab用户ID

	// 生成新的JWT token
	jwtSecret := "gitlabex_super_secret_jwt_key_make_it_long_and_random_for_production_use"

	now := time.Now()
	claims := &JWTClaims{
		GitLabAccessToken: gitlabAccessToken,
		GitLabUserID:      gitlabUserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "gitlabex",
			Subject:   fmt.Sprintf("%d", gitlabUserID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Generated JWT Token: %s\n", tokenString)
	fmt.Printf("GitLab User ID: %d\n", gitlabUserID)

	// 测试认证
	req, _ := http.NewRequest("GET", "http://localhost:8080/api/v1/research-projects", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Auth test response: %s\n", resp.Status)

	// 读取响应
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("Response: %+v\n", result)

	// 测试创建项目
	if resp.StatusCode == 200 {
		projectData := map[string]interface{}{
			"name":        "GitLab集成测试课题",
			"description": "这是一个测试GitLab集成的课题",
			"is_public":   true,
			"start_date":  time.Now().Format("2006-01-02T15:04:05Z07:00"),
		}

		jsonData, _ := json.Marshal(projectData)

		createReq, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/research-projects", bytes.NewBuffer(jsonData))
		createReq.Header.Set("Content-Type", "application/json")
		createReq.Header.Set("Authorization", "Bearer "+tokenString)

		createResp, err := client.Do(createReq)
		if err != nil {
			fmt.Printf("Error creating project: %v\n", err)
			return
		}
		defer createResp.Body.Close()

		fmt.Printf("Create project response: %s\n", createResp.Status)

		var createResult map[string]interface{}
		json.NewDecoder(createResp.Body).Decode(&createResult)
		fmt.Printf("Create Response: %+v\n", createResult)
	}
}
