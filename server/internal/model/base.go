package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 公共模型字段
type BaseModel struct {
	ID        uint           `gorm:"primarykey"            json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime"        json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"        json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                 json:"-"`
}
