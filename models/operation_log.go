package models

import (
	"time"

	"gorm.io/gorm"
)

type OperationLog struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID   uint   `gorm:"index;not null" json:"user_id"`   // 用户ID
	Username string `gorm:"size:100" json:"username"`         // 用户名（写日志时的快照）
	Nickname string `gorm:"-" json:"nickname"`               // 用户昵称（查询时按 UserID 批量填充，不落库）

	Method      string `gorm:"size:10;not null;index:idx_method_path" json:"method"` // 请求方法 PUT/DELETE/POST
	Path        string `gorm:"size:255;not null;index:idx_method_path" json:"path"`   // 请求路径
	RouteName   string `gorm:"size:100" json:"route_name"`           // 路由名称（权限名称）
	Request     string `gorm:"type:text" json:"request"`               // 请求体（JSON格式）
	Response    string `gorm:"type:text" json:"response"`            // 响应体（JSON格式）
	StatusCode  int    `gorm:"default:200" json:"status_code"`         // HTTP状态码
	IP          string `gorm:"size:50" json:"ip"`                     // 客户端IP
	UserAgent   string `gorm:"size:255" json:"user_agent"`           // 用户代理
	Duration    int64  `gorm:"default:0" json:"duration"`             // 请求耗时（毫秒）
}
