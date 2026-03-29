package esimodel

import "time"

// EveCharacterCloneBaseInfo 人物克隆信息
type EveCharacterCloneBaseInfo struct {
	ID                    uint       `gorm:"primarykey"                                    json:"id"`
	CharacterID           int64      `gorm:"uniqueIndex;not null"                          json:"character_id"`
	HomeLocationID        int64      `gorm:""                                              json:"home_location_id"`
	HomeLocationType      string     `gorm:"size:32"                                       json:"home_location_type"`
	LastCloneJumpDate     *time.Time `gorm:""                                              json:"last_clone_jump_date,omitempty"`
	LastStationChangeDate *time.Time `gorm:""                                              json:"last_station_change_date,omitempty"`
	JumpFatigueExpire     *time.Time `gorm:""                                              json:"jump_fatigue_expire,omitempty"`
	LastJumpDate          *time.Time `gorm:""                                              json:"last_jump_date,omitempty"`
	UpdatedAt             time.Time  `gorm:"autoUpdateTime"                                json:"updated_at"`
}

func (EveCharacterCloneBaseInfo) TableName() string { return "eve_character_clone_base_info" }

// EveCharacterImplants 人物植入体信息
type EveCharacterImplants struct {
	ID           uint      `gorm:"primarykey"                                     json:"id"`
	JumpCloneID  int64     `gorm:"index:idx_character_implants;not null"          json:"jump_clone_id"`
	CharacterID  int64     `gorm:"index:idx_character_implants;not null"          json:"character_id"`
	ImplantID    int       `gorm:""                                               json:"implant_id"`
	LocationID   int64     `gorm:""                                               json:"location_id"`
	LocationType string    `gorm:"size:32"                                        json:"location_type"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (EveCharacterImplants) TableName() string { return "eve_character_implants" }
