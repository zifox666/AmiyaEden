package esimodel

import "time"

// EVECharacterWallet 角色钱包余额
type EVECharacterWallet struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"                      json:"id"`
	CharacterID int64     `gorm:"not null;uniqueIndex:udx_chr_wallet_char"                          json:"character_id"`
	Balance     float64   `gorm:"type:decimal(25,2);not null"                   json:"balance"`
	UpdateTime  time.Time `gorm:"column:update_time;autoUpdateTime"             json:"update_time"`
}

func (EVECharacterWallet) TableName() string { return "eve_character_wallet" }

// EVECharacterWalletJournal 角色钱包交易记录
type EVECharacterWalletJournal struct {
	ID            int64     `gorm:"primarykey"                                    json:"id"`
	CharacterID   int64     `gorm:"not null;index:idx_char_wallet_journal"        json:"character_id"`
	Amount        float64   `gorm:"type:decimal(25,2);not null"                   json:"amount"`
	Balance       float64   `gorm:"type:decimal(25,2);not null"                   json:"balance"`
	ContextID     int64     `gorm:"not null"                                      json:"context_id"`
	ContextIDType string    `gorm:"size:32;not null"                              json:"context_id_type"`
	Date          time.Time `gorm:"not null;index"                                json:"date"`
	Description   string    `gorm:"type:text"                                     json:"description"`
	FirstPartyID  int64     `gorm:"not null"                                      json:"first_party_id"`
	Reason        string    `gorm:"type:text"                                     json:"reason"`
	RefType       string    `gorm:"size:64;not null"                              json:"ref_type"`
	SecondPartyID int64     `gorm:"not null"                                      json:"second_party_id"`
	Tax           float64   `gorm:"type:decimal(25,2);not null"                   json:"tax"`
	TaxReceiverID int64     `gorm:"not null"                                      json:"tax_receiver_id"`
}

func (EVECharacterWalletJournal) TableName() string { return "eve_character_wallet_journal" }

// EVECharacterWalletTransaction 角色钱包市场交易
type EVECharacterWalletTransaction struct {
	TransactionID int64     `gorm:"primarykey"                                    json:"transaction_id"`
	ClientID      int64     `gorm:"not null"                                      json:"client_id"`
	CharacterID   int64     `gorm:"not null;index:idx_char_wallet_transaction"    json:"character_id"`
	Date          time.Time `gorm:"not null;index"                                json:"date"`
	IsBuy         bool      `gorm:"not null"                                      json:"is_buy"`
	IsPersonal    bool      `gorm:"not null"                                      json:"is_personal"`
	JournalRefID  int64     `gorm:"not null"                                      json:"journal_ref_id"`
	LocationID    int64     `gorm:"not null"                                      json:"location_id"`
	Quantity      int       `gorm:"not null"                                      json:"quantity"`
	TypeID        int       `gorm:"not null"                                      json:"type_id"`
	UnitPrice     float64   `gorm:"type:decimal(25,2);not null"                   json:"unit_price"`
}

func (EVECharacterWalletTransaction) TableName() string { return "eve_character_wallet_transaction" }
