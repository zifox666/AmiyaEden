package esimodel

import "time"

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
