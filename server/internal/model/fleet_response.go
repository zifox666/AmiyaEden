package model

import "time"

// FleetListItem is the fleet list response DTO. It adds display-only data that
// is resolved via joins without extending the persisted fleet table schema.
type FleetListItem struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	StartAt         time.Time `json:"start_at"`
	EndAt           time.Time `json:"end_at"`
	Importance      string    `json:"importance"`
	PapCount        float64   `json:"pap_count"`
	FCUserID        uint      `json:"fc_user_id"`
	FCCharacterID   int64     `json:"fc_character_id"`
	FCCharacterName string    `json:"fc_character_name"`
	FCDisplayName   string    `json:"fc_display_name,omitempty"`
	ESIFleetID      *int64    `json:"esi_fleet_id,omitempty"`
	FleetConfigID   *uint     `json:"fleet_config_id,omitempty"`
	AutoSrpMode     string    `json:"auto_srp_mode"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
