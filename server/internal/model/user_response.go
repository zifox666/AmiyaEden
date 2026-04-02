package model

import "time"

type UserListCharacter struct {
	CharacterID   int64  `json:"character_id"`
	CharacterName string `json:"character_name"`
	PortraitURL   string `json:"portrait_url"`
	TotalSP       int64  `json:"total_sp"`
	TokenInvalid  bool   `json:"token_invalid"`
}

// UserListItem is the system user page DTO. It intentionally exposes the
// active role list only, so list consumers do not rely on the legacy user.role
// mirror column.
type UserListItem struct {
	ID          uint                `json:"id"`
	Nickname    string              `json:"nickname"`
	QQ          string              `json:"qq"`
	DiscordID   string              `json:"discord_id"`
	Avatar      string              `json:"avatar"`
	Status      int8                `json:"status"`
	Roles       []string            `json:"roles"`
	Characters  []UserListCharacter `json:"characters"`
	LastLoginAt *time.Time          `json:"last_login_at"`
	LastLoginIP string              `json:"last_login_ip"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

func NewUserListCharacter(char EveCharacter, totalSP int64) UserListCharacter {
	return UserListCharacter{
		CharacterID:   char.CharacterID,
		CharacterName: char.CharacterName,
		PortraitURL:   char.PortraitURL,
		TotalSP:       totalSP,
		TokenInvalid:  char.TokenInvalid,
	}
}

func NewUserListItem(user User, roleCodes []string, characters []UserListCharacter) UserListItem {
	if len(characters) == 0 {
		characters = []UserListCharacter{}
	}

	return UserListItem{
		ID:          user.ID,
		Nickname:    user.Nickname,
		QQ:          user.QQ,
		DiscordID:   user.DiscordID,
		Avatar:      user.Avatar,
		Status:      user.Status,
		Roles:       NormalizeRoleCodes(roleCodes, user.Role),
		Characters:  characters,
		LastLoginAt: user.LastLoginAt,
		LastLoginIP: user.LastLoginIP,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}
