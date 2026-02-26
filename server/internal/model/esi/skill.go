package esimodel

type EveCharacterSkill struct {
	ID            uint  `gorm:"primaryKey;autoIncrement"                      json:"id"`
	CharacterID   int64 `gorm:"not null;uniqueIndex:udx_chr_skill_char"        json:"character_id"`
	TotalSP       int64 `gorm:"not null;default:0"                             json:"total_sp"`
	UnallocatedSP int64 `gorm:"not null;default:0"                             json:"unallocated_sp"`
	UpdatedTime   int64 `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

type EveCharacterSkills struct {
	ID                 uint  `gorm:"primaryKey;autoIncrement"                          json:"id"`
	CharacterID        int64 `gorm:"not null;uniqueIndex:udx_chr_skills_char_skill"     json:"character_id"`
	SkillID            int   `gorm:"not null;uniqueIndex:udx_chr_skills_char_skill"     json:"skill_id"`
	ActiveLevel        int   `gorm:"not null;default:0"                                json:"active_level"`
	TrainedLevel       int   `gorm:"not null;default:0"                                json:"trained_level"`
	SkillpointsInSkill int64 `gorm:"not null;default:0"                                json:"skillpoints_in_skill"`
	UpdatedTime        int64 `gorm:"autoUpdateTime"                                    json:"updated_at"`
}

type EveCharacterSkillQueue struct {
	ID              uint  `gorm:"primaryKey;autoIncrement"                           json:"id"`
	CharacterID     int64 `gorm:"not null;uniqueIndex:udx_chr_skill_queue_pos"        json:"character_id"`
	QueuePosition   int   `gorm:"not null;uniqueIndex:udx_chr_skill_queue_pos"        json:"queue_position"`
	SkillID         int   `gorm:"not null"                                           json:"skill_id"`
	LevelEndSP      int64 `gorm:"not null;default:0"                                 json:"level_end_sp"`
	LevelStartSP    int64 `gorm:"not null;default:0"                                 json:"level_start_sp"`
	TrainingStartSP int64 `gorm:"not null;default:0"                                 json:"training_start_sp"`
	FinishedLevel   int   `gorm:"not null;default:0"                                 json:"finished_level"`
	StartDate       int64 `gorm:"not null;default:0"                                 json:"start_date"`
	FinishDate      int64 `gorm:"not null;default:0"                                 json:"finish_date"`
	UpdatedTime     int64 `gorm:"autoUpdateTime"                                     json:"updated_at"`
}
