package esimodel

// EveCharacterTitle 角色军团头衔
type EveCharacterTitle struct {
	ID          uint   `gorm:"primarykey"                                    json:"id"`
	CharacterID int64  `gorm:"not null;index:idx_char_title,unique"          json:"character_id"`
	TitleID     int    `gorm:"not null;index:idx_char_title,unique"          json:"title_id"`
	Name        string `gorm:"size:256"                                      json:"name"`
}

func (EveCharacterTitle) TableName() string { return "eve_character_title" }
