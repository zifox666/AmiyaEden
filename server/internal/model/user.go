package model

import "time"

// User 用户模型（仅通过 EVE SSO 登录，无用户名/密码）
type User struct {
	BaseModel
	Nickname    string     `gorm:"size:128"               json:"nickname"`
	Avatar      string     `gorm:"size:512"               json:"avatar"`
	Status      int8       `gorm:"default:1"              json:"status"` // 1:正常 0:禁用
	Role        string     `gorm:"size:32;default:'user'" json:"role"`
	LastLoginAt *time.Time `gorm:""                       json:"last_login_at"`
	LastLoginIP string     `gorm:"size:64"                json:"last_login_ip"`
}

func (User) TableName() string {
	return "user"
}
