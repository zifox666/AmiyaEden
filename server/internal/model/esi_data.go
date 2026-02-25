package model

import "time"

// ─────────────────────────────────────────────
//  ESI 数据模型 — 由 ESI 刷新任务写入
// ─────────────────────────────────────────────

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
	IsBlueprintCopy *bool  `gorm:""                                              json:"is_blueprint_copy,omitempty"`
}

func (EveCharacterAsset) TableName() string { return "eve_character_asset" }

// EveCharacterNotification 角色通知
type EveCharacterNotification struct {
	ID             uint      `gorm:"primarykey"                                       json:"id"`
	CharacterID    int64     `gorm:"not null;index:idx_char_notif,unique"             json:"character_id"`
	NotificationID int64     `gorm:"not null;index:idx_char_notif,unique"             json:"notification_id"`
	SenderID       int64     `gorm:"not null"                                         json:"sender_id"`
	SenderType     string    `gorm:"size:32;not null"                                 json:"sender_type"`
	Text           *string   `gorm:"type:text"                                        json:"text,omitempty"`
	Timestamp      time.Time `gorm:"not null;index"                                   json:"timestamp"`
	Type           string    `gorm:"size:128;not null"                                json:"type"`
	IsRead         *bool     `gorm:""                                                 json:"is_read,omitempty"`
}

func (EveCharacterNotification) TableName() string { return "eve_character_notification" }

// EveCharacterTitle 角色军团头衔
type EveCharacterTitle struct {
	ID          uint   `gorm:"primarykey"                                    json:"id"`
	CharacterID int64  `gorm:"not null;index:idx_char_title,unique"          json:"character_id"`
	TitleID     int    `gorm:"not null;index:idx_char_title,unique"          json:"title_id"`
	Name        string `gorm:"size:256"                                      json:"name"`
}

func (EveCharacterTitle) TableName() string { return "eve_character_title" }

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

// EveCharacterContract 角色合同
type EveCharacterContract struct {
	ID                  uint       `gorm:"primarykey"                                              json:"id"`
	CharacterID         int64      `gorm:"not null;uniqueIndex:idx_char_contract"                  json:"character_id"`
	ContractID          int64      `gorm:"not null;uniqueIndex:idx_char_contract"                  json:"contract_id"`
	AcceptorID          int64      `gorm:""                                                        json:"acceptor_id"`
	AssigneeID          int64      `gorm:""                                                        json:"assignee_id"`
	Availability        string     `gorm:"size:32"                                                 json:"availability"`
	Buyout              *float64   `gorm:""                                                        json:"buyout,omitempty"`
	Collateral          *float64   `gorm:""                                                        json:"collateral,omitempty"`
	DateAccepted        *time.Time `gorm:""                                                        json:"date_accepted,omitempty"`
	DateCompleted       *time.Time `gorm:""                                                        json:"date_completed,omitempty"`
	DateExpired         time.Time  `gorm:"not null"                                                json:"date_expired"`
	DateIssued          time.Time  `gorm:"not null;index"                                          json:"date_issued"`
	DaysToComplete      *int       `gorm:""                                                        json:"days_to_complete,omitempty"`
	EndLocationID       *int64     `gorm:""                                                        json:"end_location_id,omitempty"`
	ForCorporation      bool       `gorm:"default:false"                                           json:"for_corporation"`
	IssuerCorporationID int64      `gorm:""                                                        json:"issuer_corporation_id"`
	IssuerID            int64      `gorm:""                                                        json:"issuer_id"`
	Price               *float64   `gorm:""                                                        json:"price,omitempty"`
	Reward              *float64   `gorm:""                                                        json:"reward,omitempty"`
	StartLocationID     *int64     `gorm:""                                                        json:"start_location_id,omitempty"`
	Status              string     `gorm:"size:32;not null;index"                                  json:"status"`
	Title               *string    `gorm:"size:256"                                                json:"title,omitempty"`
	Type                string     `gorm:"size:32;not null"                                        json:"type"`
	Volume              *float64   `gorm:""                                                        json:"volume,omitempty"`
	BidsJSON            *string    `gorm:"type:text"                                               json:"bids_json,omitempty"`
	ItemsJSON           *string    `gorm:"type:text"                                               json:"items_json,omitempty"`
}

func (EveCharacterContract) TableName() string { return "eve_character_contract" }
