package model

import "time"

// SeatUser SeAT 用户绑定记录，一个本系统用户最多关联一个 SeAT 账号
type SeatUser struct {
	BaseModel
	SeatUserID   string    `gorm:"size:64;uniqueIndex;not null" json:"seat_user_id"`  // SeAT 内部用户 ID (sub)
	SeatUsername string    `gorm:"size:128;not null"            json:"seat_username"` // SeAT 用户名 (nam)
	UserID       uint      `gorm:"uniqueIndex;not null"         json:"user_id"`       // 本系统用户 ID
	MainCharID   int64     `gorm:"default:0"                    json:"main_char_id"`  // SeAT 主角色 character_id (uid)
	AccessToken  string    `gorm:"type:text"                    json:"-"`
	RefreshToken string    `gorm:"type:text"                    json:"-"`
	TokenExpiry  time.Time `gorm:""                             json:"token_expiry"`
	Groups       string    `gorm:"type:text"                    json:"groups"` // JSON 数组字符串
}

func (SeatUser) TableName() string { return "seat_user" }
