package models

import (
	"time"

	"gorm.io/gorm"
)

type Role struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string `gorm:"uniqueIndex;size:100;not null" json:"name"`
	Description string `gorm:"size:255" json:"description"`

	// 显式指定关联表名，NamingStrategy 的 TablePrefix 会作用到该名称
	Users       []User       `gorm:"many2many:user_roles" json:"users,omitempty"`
	Permissions []Permission `gorm:"many2many:role_permissions" json:"permissions,omitempty"`
}
