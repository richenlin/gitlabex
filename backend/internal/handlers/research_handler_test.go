package handlers

import (
	"testing"
)

func TestGenerateProjectPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "全中文输入",
			input:    "机器学习研究",
			expected: "jiqixuexiyanjiu",
		},
		{
			name:     "中英文混合",
			input:    "AI人工智能研究",
			expected: "airengongzhinengyanjiu",
		},
		{
			name:     "包含空格的中文",
			input:    "深度 学习 项目",
			expected: "shendu-xuexi-xiangmu",
		},
		{
			name:     "纯英文",
			input:    "Machine Learning",
			expected: "machine-learning",
		},
		{
			name:     "包含特殊字符",
			input:    "项目@2024#测试",
			expected: "xiangmu2024ceshi",
		},
		{
			name:     "包含数字",
			input:    "实验123项目",
			expected: "shiyan123xiangmu",
		},
		{
			name:     "连续特殊字符",
			input:    "测试---项目",
			expected: "ceshi-xiangmu",
		},
		{
			name:     "空字符串",
			input:    "",
			expected: "research-project",
		},
		{
			name:     "单个字符",
			input:    "测",
			expected: "research-project",
		},
		{
			name:     "以连字符开头",
			input:    "-测试项目",
			expected: "ceshixiangmu",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateProjectPath(tt.input)
			if result != tt.expected {
				t.Errorf("generateProjectPath(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
