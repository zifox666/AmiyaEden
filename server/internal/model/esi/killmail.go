package esimodel

import "time"

// EveKillmailList 击杀邮件主记录（受害者视角，全局去重）
type EveKillmailList struct {
	ID            uint      `gorm:"primarykey"                                    json:"id"`
	KillmailID    int64     `gorm:"column:kill_mail_id;uniqueIndex"               json:"killmail_id"`
	KillmailHash  string    `gorm:"column:kill_mail_hash;size:200"                json:"killmail_hash"`
	KillmailTime  time.Time `gorm:"column:kill_mail_time;index"                   json:"killmail_time"`
	SolarSystemID int64     `gorm:"column:solar_system_id"                        json:"solar_system_id"`
	ShipTypeID    int64     `gorm:"column:ship_type_id"                           json:"ship_type_id"`
	CharacterID   int64     `gorm:"column:character_id"                           json:"character_id"`
	CorporationID int64     `gorm:"column:corporation_id"                         json:"corporation_id"`
	AllianceID    int64     `gorm:"column:alliance_id"                            json:"alliance_id"`
	JaniceAmount  *float64  `gorm:"column:janice_amount;type:decimal(20,2)"       json:"janice_amount"`
	CreateTime    time.Time `gorm:"column:create_time;autoCreateTime"             json:"create_time"`
}

func (EveKillmailList) TableName() string { return "eve_killmail_list" }

// EveKillmailItem 击杀邮件中被摧毁 / 掉落的物品
type EveKillmailItem struct {
	ID         uint      `gorm:"primarykey"                                    json:"id"`
	KillmailID int64     `gorm:"column:kill_mail_id;index"                     json:"killmail_id"`
	ItemID     int       `gorm:"column:item_id"                                json:"item_id"`
	ItemNum    int64     `gorm:"column:item_num"                               json:"item_num"`
	DropType   *bool     `gorm:"column:drop_type"                              json:"drop_type"`
	Flag       int       `gorm:"column:flag"                                   json:"flag"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime"             json:"create_time"`
}

func (EveKillmailItem) TableName() string { return "eve_killmail_item" }

// EveCharacterKillmail 角色-击杀邮件关联表
type EveCharacterKillmail struct {
	ID          uint      `gorm:"primarykey"                                    json:"id"`
	CharacterID int64     `gorm:"not null;uniqueIndex:idx_char_km"              json:"character_id"`
	KillmailID  int64     `gorm:"not null;uniqueIndex:idx_char_km;index"        json:"killmail_id"`
	Srped       bool      `gorm:"default:0"                                     json:"srped"`
	Victim      bool      `gorm:"default:0"                                     json:"victim"`
	CreateTime  time.Time `gorm:"column:create_time;autoCreateTime"             json:"create_time"`
}

func (EveCharacterKillmail) TableName() string { return "eve_character_killmail" }
