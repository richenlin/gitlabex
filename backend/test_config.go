package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("=== 配置加载测试 ===")

	// 1. 测试当前环境变量
	fmt.Printf("初始 GITLAB_INTERNAL_URL: '%s'\n", os.Getenv("GITLAB_INTERNAL_URL"))

	// 2. 加载app.env
	if err := godotenv.Load("../config/app.env"); err == nil {
		fmt.Println("✅ 加载 app.env 成功")
	} else {
		fmt.Printf("❌ 加载 app.env 失败: %v\n", err)
	}

	// 3. 加载oauth.env
	if err := godotenv.Load("../config/oauth.env"); err == nil {
		fmt.Println("✅ 加载 oauth.env 成功")
	} else {
		fmt.Printf("❌ 加载 oauth.env 失败: %v\n", err)
	}

	// 4. 测试最终环境变量
	fmt.Printf("最终 GITLAB_INTERNAL_URL: '%s'\n", os.Getenv("GITLAB_INTERNAL_URL"))
	fmt.Printf("最终 GITLAB_EXTERNAL_URL: '%s'\n", os.Getenv("GITLAB_EXTERNAL_URL"))
	fmt.Printf("最终 GITLAB_CLIENT_ID: '%s'\n", os.Getenv("GITLAB_CLIENT_ID")[:10]+"...")
}
