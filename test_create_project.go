package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func main() {
	// 生成JWT token
	jwtSecret := "gitlabex_super_secret_jwt_key_make_it_long_and_random_for_production_use"
	userID := uuid.New().String()

	now := time.Now()
	claims := &JWTClaims{
		UserID:   userID,
		Username: "test_user",
		Role:     "teacher",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "gitlabex",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Generated JWT Token: %s\n", tokenString)
	fmt.Printf("User ID: %s\n", userID)

	// 创建项目
	projectData := map[string]interface{}{
		"name":        "测试课题",
		"description": "这是一个测试课题",
		"is_public":   true,
		"start_date":  time.Now().Format("2006-01-02T15:04:05Z07:00"),
	}

	jsonData, _ := json.Marshal(projectData)

	// 发送创建项目请求
	req, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/research-projects", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenString)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error creating project: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Create project response: %s\n", resp.Status)

	// 读取响应
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("Response: %+v\n", result)

	if projectID, ok := result["id"].(string); ok {
		fmt.Printf("Created project ID: %s\n", projectID)

		// 测试编辑项目
		updateData := map[string]interface{}{
			"name": "更新的测试课题",
		}
		updateJSON, _ := json.Marshal(updateData)

		updateReq, _ := http.NewRequest("PUT", "http://localhost:8080/api/v1/research-projects/"+projectID, bytes.NewBuffer(updateJSON))
		updateReq.Header.Set("Content-Type", "application/json")
		updateReq.Header.Set("Authorization", "Bearer "+tokenString)

		updateResp, err := client.Do(updateReq)
		if err != nil {
			fmt.Printf("Error updating project: %v\n", err)
			return
		}
		defer updateResp.Body.Close()

		fmt.Printf("Update project response: %s\n", updateResp.Status)

		var updateResult map[string]interface{}
		json.NewDecoder(updateResp.Body).Decode(&updateResult)
		fmt.Printf("Update Response: %+v\n", updateResult)
	}
}
