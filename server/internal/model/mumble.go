package model

// MumbleAccount stores the Mumble login credential owned by an AmiyaEden user.
// The username is always the user ID; DisplayName is used by the ICE authenticator.
type MumbleAccount struct {
	BaseModel
	UserID      uint   `gorm:"uniqueIndex;not null" json:"user_id"`
	Password    string `gorm:"size:64;not null"     json:"password"`
	DisplayName string `gorm:"size:128;not null"    json:"display_name"`
}

func (MumbleAccount) TableName() string { return "mumble_account" }

// VoiceRoleGroupMapping maps AmiyaEden roles to provider-side voice groups.
// Mumble uses these groups from the ICE authenticator; channel ACLs stay in Mumble.
type VoiceRoleGroupMapping struct {
	BaseModel
	Provider  string `gorm:"size:32;not null;uniqueIndex:idx_voice_role_group" json:"provider"`
	RoleCode  string `gorm:"size:50;not null;uniqueIndex:idx_voice_role_group" json:"role_code"`
	GroupName string `gorm:"size:128;not null"                                  json:"group_name"`
	Enabled   bool   `gorm:"default:true"                                       json:"enabled"`
}

func (VoiceRoleGroupMapping) TableName() string { return "voice_role_group_mapping" }

const (
	SysConfigMumbleEnabled    = "mumble.enabled"
	SysConfigMumbleURL        = "mumble.url"
	SysConfigMumblePort       = "mumble.port"
	SysConfigMumbleServerName = "mumble.server_name"
	SysConfigMumbleAuthSecret = "mumble.auth_secret"

	SysConfigDefaultMumbleEnabled    = "false"
	SysConfigDefaultMumbleURL        = ""
	SysConfigDefaultMumblePort       = "64738"
	SysConfigDefaultMumbleServerName = "AmiyaEden Mumble"
	SysConfigDefaultMumbleAuthSecret = ""
)
