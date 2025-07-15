package utils

import (
	"time"
)

// DateOnly 自定义日期类型，只处理日期部分
type DateOnly struct {
	time.Time
}

// UnmarshalJSON 实现JSON反序列化
func (d *DateOnly) UnmarshalJSON(data []byte) error {
	// 移除引号
	str := string(data)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	// 解析日期格式 YYYY-MM-DD
	t, err := time.Parse("2006-01-02", str)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

// MarshalJSON 实现JSON序列化
func (d DateOnly) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Time.Format("2006-01-02") + `"`), nil
}

// NewDateOnly 创建DateOnly实例
func NewDateOnly(t time.Time) DateOnly {
	return DateOnly{Time: t}
}

// NewDateOnlyFromString 从字符串创建DateOnly实例
func NewDateOnlyFromString(dateStr string) (DateOnly, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return DateOnly{}, err
	}
	return DateOnly{Time: t}, nil
}

// String 返回日期字符串
func (d DateOnly) String() string {
	return d.Time.Format("2006-01-02")
}
