package esimodel

// EveCharacterAsset 角色资产
type EveCharacterAsset struct {
	ID              uint   `gorm:"primarykey"                                    json:"id"`
	CharacterID     int64  `gorm:"not null;index:idx_char_asset,unique"          json:"character_id"`
	ItemID          int64  `gorm:"not null;index:idx_char_asset,unique"          json:"item_id"`
	TypeID          int    `gorm:"not null;index"                                json:"type_id"`
	Quantity        int    `gorm:"not null;default:1"                            json:"quantity"`
	LocationID      int64  `gorm:"index"                                         json:"location_id"`
	LocationType    string `gorm:"size:32"                                       json:"location_type"`
	LocationFlag    string `gorm:"size:64"                                       json:"location_flag"`
	IsSingleton     bool   `gorm:"default:false"                                 json:"is_singleton"`
	IsBlueprintCopy *bool  `gorm:""                                             json:"is_blueprint_copy,omitempty"`
}

func (EveCharacterAsset) TableName() string { return "eve_character_asset" }
