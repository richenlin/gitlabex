package dto

import (
	"time"

	"github.com/google/uuid"
)

// ActivityItem 活动项
type ActivityItem struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`                   // document, topic, homework, comment
	Title       string    `json:"title"`                  // 活动标题
	Description string    `json:"description"`            // 活动描述
	UserName    string    `json:"user_name"`              // 操作用户名
	UserAvatar  string    `json:"user_avatar"`            // 用户头像
	ProjectName string    `json:"project_name,omitempty"` // 所属项目名称
	CreatedAt   time.Time `json:"created_at"`             // 创建时间
	URL         string    `json:"url"`                    // 跳转链接
}
