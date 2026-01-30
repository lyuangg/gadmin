package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Username     string `gorm:"uniqueIndex;size:100;not null" json:"username"`
	Password     string `gorm:"size:255;not null" json:"-"`
	Nickname     string `gorm:"size:100" json:"nickname"`
	Avatar       string `gorm:"size:255" json:"avatar"`
	Type         int    `gorm:"default:0" json:"type"`
	Status       int    `gorm:"default:1" json:"status"`
	TokenVersion uint   `gorm:"default:0;not null" json:"-"`
	Remark       string `gorm:"size:500" json:"remark"`

	Roles []Role `gorm:"many2many:user_roles" json:"roles,omitempty"`
}

type UserRole struct {
	UserID uint `gorm:"primaryKey"`
	RoleID uint `gorm:"primaryKey"`
}
