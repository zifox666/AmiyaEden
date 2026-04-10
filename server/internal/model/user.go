package model

import (
	"strings"
	"time"
)

// User 用户模型（仅通过 EVE SSO 登录，无用户名/密码）
type User struct {
	BaseModel
	Nickname           string     `gorm:"size:128"               json:"nickname"`
	QQ                 string     `gorm:"size:20"                json:"qq"`
	DiscordID          string     `gorm:"size:20"                json:"discord_id"`
	Status             int8       `gorm:"default:1"              json:"status"` // 1:正常 0:禁用
	Role               string     `gorm:"size:32;default:'user'" json:"role"`
	PrimaryCharacterID int64      `gorm:"default:0"              json:"primary_character_id"` // 主人物 EVE Character ID
	LastLoginAt        *time.Time `gorm:""                       json:"last_login_at"`
	LastLoginIP        string     `gorm:"size:64"                json:"last_login_ip"`
}

func (User) TableName() string {
	return "user"
}

func (u User) HasNickname() bool {
	return strings.TrimSpace(u.Nickname) != ""
}

func (u User) HasRequiredContact() bool {
	return strings.TrimSpace(u.QQ) != "" || strings.TrimSpace(u.DiscordID) != ""
}

func (u User) ProfileComplete() bool {
	return u.HasNickname() && u.HasRequiredContact()
}
