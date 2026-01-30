package models

import (
	"time"

	"gorm.io/gorm"
)

type Permission struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Path        string `gorm:"size:255;not null" json:"path"`    // 接口路径
	Method      string `gorm:"size:20;not null" json:"method"`   // 请求方法 GET/POST/PUT/DELETE等
	Name        string `gorm:"size:100" json:"name"`             // 权限名称
	Group       string `gorm:"size:50" json:"group"`             // 权限分组名称
	Description string `gorm:"size:255" json:"description"`      // 权限描述
	AutoImport  bool   `gorm:"default:false" json:"auto_import"` // 是否自动导入

	// 显式指定关联表名，NamingStrategy 的 TablePrefix 会作用到该名称
	Roles []Role `gorm:"many2many:role_permissions" json:"roles,omitempty"`
}
