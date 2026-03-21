package model

import "time"

// UserListItem is the system user page DTO. It intentionally exposes the
// active role list only, so list consumers do not rely on the legacy user.role
// mirror column.
type UserListItem struct {
	ID          uint       `json:"id"`
	Nickname    string     `json:"nickname"`
	QQ          string     `json:"qq"`
	DiscordID   string     `json:"discord_id"`
	Avatar      string     `json:"avatar"`
	Status      int8       `json:"status"`
	Roles       []string   `json:"roles"`
	LastLoginAt *time.Time `json:"last_login_at"`
	LastLoginIP string     `json:"last_login_ip"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func NewUserListItem(user User, roleCodes []string) UserListItem {
	return UserListItem{
		ID:          user.ID,
		Nickname:    user.Nickname,
		QQ:          user.QQ,
		DiscordID:   user.DiscordID,
		Avatar:      user.Avatar,
		Status:      user.Status,
		Roles:       NormalizeRoleCodes(roleCodes, user.Role),
		LastLoginAt: user.LastLoginAt,
		LastLoginIP: user.LastLoginIP,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}
