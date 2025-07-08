package main

import (
	"fmt"
	"log"
	"os"

	"github.com/xanzy/go-gitlab"
)

func main() {
	// 从环境变量获取GitLab token，如果没有则使用默认的root用户
	token := os.Getenv("GITLAB_TOKEN")
	if token == "" {
		fmt.Println("请设置环境变量 GITLAB_TOKEN，或者访问 http://localhost/-/profile/personal_access_tokens 创建访问令牌")
		fmt.Println("提示：GitLab root用户密码是 password123")
		return
	}

	// 创建GitLab客户端
	git, err := gitlab.NewClient(token, gitlab.WithBaseURL("http://localhost"))
	if err != nil {
		log.Fatalf("创建GitLab客户端失败: %v", err)
	}

	// 测试连接
	fmt.Println("正在测试GitLab API连接...")

	// 获取当前用户信息
	user, _, err := git.Users.CurrentUser()
	if err != nil {
		log.Fatalf("获取当前用户信息失败: %v", err)
	}

	fmt.Printf("✅ GitLab API连接成功！\n")
	fmt.Printf("当前用户: %s (%s)\n", user.Name, user.Username)
	fmt.Printf("用户ID: %d\n", user.ID)
	fmt.Printf("邮箱: %s\n", user.Email)

	// 获取用户的项目列表
	projects, _, err := git.Projects.ListUserProjects(user.ID, &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{PerPage: 5},
	})
	if err != nil {
		log.Printf("获取项目列表失败: %v", err)
	} else {
		fmt.Printf("\n用户项目列表 (最多显示5个):\n")
		if len(projects) == 0 {
			fmt.Printf("  暂无项目\n")
		} else {
			for _, project := range projects {
				fmt.Printf("  - %s (%s)\n", project.Name, project.WebURL)
			}
		}
	}

	// 获取用户的组列表
	groups, _, err := git.Groups.ListGroups(&gitlab.ListGroupsOptions{
		ListOptions: gitlab.ListOptions{PerPage: 5},
	})
	if err != nil {
		log.Printf("获取组列表失败: %v", err)
	} else {
		fmt.Printf("\n用户组列表 (最多显示5个):\n")
		if len(groups) == 0 {
			fmt.Printf("  暂无组\n")
		} else {
			for _, group := range groups {
				fmt.Printf("  - %s (%s)\n", group.Name, group.WebURL)
			}
		}
	}

	// 测试创建一个测试项目
	fmt.Printf("\n正在创建测试项目...\n")
	testProjectName := "gitlabex-test-project"
	testProject, _, err := git.Projects.CreateProject(&gitlab.CreateProjectOptions{
		Name:        gitlab.Ptr(testProjectName),
		Description: gitlab.Ptr("GitLabEx教育系统测试项目"),
		Visibility:  gitlab.Ptr(gitlab.PublicVisibility),
	})
	if err != nil {
		// 如果项目已存在，尝试获取它
		if projects, _, err := git.Projects.ListProjects(&gitlab.ListProjectsOptions{
			Search: gitlab.Ptr(testProjectName),
		}); err == nil && len(projects) > 0 {
			testProject = projects[0]
			fmt.Printf("✅ 测试项目已存在: %s\n", testProject.WebURL)
		} else {
			log.Printf("创建测试项目失败: %v", err)
		}
	} else {
		fmt.Printf("✅ 测试项目创建成功: %s\n", testProject.WebURL)
	}

	if testProject != nil {
		// 测试创建一个Issue
		fmt.Printf("\n正在创建测试Issue...\n")
		issue, _, err := git.Issues.CreateIssue(testProject.ID, &gitlab.CreateIssueOptions{
			Title:       gitlab.Ptr("GitLabEx系统测试Issue"),
			Description: gitlab.Ptr("这是一个由GitLabEx教育系统自动创建的测试Issue。"),
			Labels:      &gitlab.LabelOptions{"测试", "自动化"},
		})
		if err != nil {
			log.Printf("创建测试Issue失败: %v", err)
		} else {
			fmt.Printf("✅ 测试Issue创建成功: %s\n", issue.WebURL)
		}
	}

	fmt.Printf("\n🎉 GitLab API集成测试完成！\n")
}
