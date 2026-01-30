package models

import (
	"time"

	"gorm.io/gorm"
)

type DictType struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Code   string `gorm:"uniqueIndex;size:64;not null" json:"code"`   // 类型编码，用于程序引用
	Name   string `gorm:"size:100;not null" json:"name"`             // 类型名称，用于展示
	Remark string `gorm:"size:255" json:"remark"`                    // 备注

	Items []DictItem `gorm:"foreignKey:TypeID" json:"items,omitempty"`
}

type DictItem struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	TypeID uint   `gorm:"index;not null" json:"type_id"`           // 所属字典类型 ID
	Label  string `gorm:"size:100;not null" json:"label"`          // 显示文本
	Value  string `gorm:"size:100;not null" json:"value"`          // 实际值（如 0、1、pending）
	Sort   int    `gorm:"default:0" json:"sort"`                   // 排序，数值越小越靠前
	Status int    `gorm:"default:1" json:"status"`                 // 状态：0=禁用，1=启用
	Remark string `gorm:"size:255" json:"remark"`                 // 备注
}
