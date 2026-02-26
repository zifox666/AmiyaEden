package esimodel

import "time"

// EveCharacterClone 角色克隆信息（含植入体 & 跳跃疲劳）
type EveCharacterClone struct {
	ID                    uint       `gorm:"primarykey"                                    json:"id"`
	CharacterID           int64      `gorm:"uniqueIndex;not null"                          json:"character_id"`
	HomeLocationID        int64      `gorm:""                                              json:"home_location_id"`
	HomeLocationType      string     `gorm:"size:32"                                       json:"home_location_type"`
	JumpClonesJSON        string     `gorm:"type:text"                                     json:"jump_clones_json"`
	ImplantsJSON          string     `gorm:"type:text"                                     json:"implants_json"`
	LastCloneJumpDate     *time.Time `gorm:""                                              json:"last_clone_jump_date,omitempty"`
	LastStationChangeDate *time.Time `gorm:""                                              json:"last_station_change_date,omitempty"`
	JumpFatigueExpire     *time.Time `gorm:""                                              json:"jump_fatigue_expire,omitempty"`
	LastJumpDate          *time.Time `gorm:""                                              json:"last_jump_date,omitempty"`
	UpdatedAt             time.Time  `gorm:"autoUpdateTime"                                json:"updated_at"`
}

func (EveCharacterClone) TableName() string { return "eve_character_clone" }
