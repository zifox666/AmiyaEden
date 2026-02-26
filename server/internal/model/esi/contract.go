package esimodel

import "time"

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
