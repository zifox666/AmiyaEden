package esimodel

// EveCharacterFitting 人物装配主记录
type EveCharacterFitting struct {
	ID          uint   `gorm:"primarykey"                                      json:"id"`
	FittingID   int64  `gorm:"not null;uniqueIndex:udx_char_fitting"           json:"fitting_id"`
	CharacterID int64  `gorm:"not null;uniqueIndex:udx_char_fitting;index"     json:"character_id"`
	Name        string `gorm:"size:256;not null"                               json:"name"`
	ShipTypeID  int64  `gorm:"not null;index"                                  json:"ship_type_id"`
	Description string `gorm:"type:text"                                       json:"description"`
}

func (EveCharacterFitting) TableName() string { return "eve_character_fitting" }

// EveCharacterFittingItem 装配物品明细
type EveCharacterFittingItem struct {
	ID          uint   `gorm:"primarykey"                                      json:"id"`
	FittingID   int64  `gorm:"not null;index:idx_fitting_item"                 json:"fitting_id"`
	CharacterID int64  `gorm:"not null;index:idx_fitting_item"                json:"character_id"`
	TypeID      int64  `gorm:"not null"                                        json:"type_id"`
	Quantity    int    `gorm:"not null;default:1"                              json:"quantity"`
	Flag        string `gorm:"size:64;not null"                                json:"flag"`
}

func (EveCharacterFittingItem) TableName() string { return "eve_character_fitting_item" }
