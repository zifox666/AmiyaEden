package model

import "time"

// User 用户模型（仅通过 EVE SSO 登录，无用户名/密码）
type User struct {
	BaseModel
	Nickname           string     `gorm:"size:128"               json:"nickname"`
	Avatar             string     `gorm:"size:512"               json:"avatar"`
	Status             int8       `gorm:"default:1"              json:"status"` // 1:正常 0:禁用
	Role               string     `gorm:"size:32;default:'user'" json:"role"`
	PrimaryCharacterID int64      `gorm:"default:0"              json:"primary_character_id"` // 主角色 EVE Character ID
	LastLoginAt        *time.Time `gorm:""                       json:"last_login_at"`
	LastLoginIP        string     `gorm:"size:64"                json:"last_login_ip"`
}

func (User) TableName() string {
	return "user"
}

// UserListItemDTO 用户列表返回 DTO（含多角色）
type UserListItemDTO struct {
	ID                 uint       `json:"id"`
	Nickname           string     `json:"nickname"`
	Avatar             string     `json:"avatar"`
	Status             int8       `json:"status"`
	Role               string     `json:"role"` // 向后兼容字段
	Roles              []string   `json:"roles"`
	PrimaryCharacterID int64      `json:"primary_character_id"`
	LastLoginAt        *time.Time `json:"last_login_at"`
	LastLoginIP        string     `json:"last_login_ip"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}
