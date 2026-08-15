package model

// ProfileDTO is the authenticated user's editable AT profile plus the
// resolved official-server main-character source.
type ProfileDTO struct {
	QQ                      string `json:"qq"`
	Nickname                string `json:"nickname"`
	TournamentCharacterName string `json:"tournament_character_name"`
	TournamentCharacterID   *int64 `json:"tournament_character_id"`
	UILanguage              string `json:"ui_language"`
	GameLanguage            string `json:"game_language"`
	MainCharacterSource     string `json:"main_character_source"`
	MainCharacterID         int64  `json:"main_character_id"`
	MainCharacterName       string `json:"main_character_name"`
}

type UpdateProfileRequest struct {
	QQ                      string `json:"qq"`
	Nickname                string `json:"nickname"`
	TournamentCharacterName string `json:"tournament_character_name"`
	TournamentCharacterID   *int64 `json:"tournament_character_id"`
	UILanguage              string `json:"ui_language"`
	GameLanguage            string `json:"game_language"`
}
