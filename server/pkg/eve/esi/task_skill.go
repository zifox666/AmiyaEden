package esi

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"

	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

func init() {
	Register(&SkillTask{})
}

type SkillTask struct{}

func (t *SkillTask) Name() string        { return "character_skill" }
func (t *SkillTask) Description() string { return "角色技能信息" }
func (t *SkillTask) Priority() Priority  { return PriorityNormal }

func (t *SkillTask) Interval() RefreshInterval {
	return RefreshInterval{
		Active:   24 * time.Hour,
		Inactive: 7 * 24 * time.Hour,
	}
}

func (t *SkillTask) RequiredScopes() []TaskScope {
	return []TaskScope{
		{Scope: "esi-skills.read_skills.v1", Description: "读取角色技能信息"},
		{Scope: "esi-skills.read_skillqueue.v1", Description: "读取角色技能队列信息"},
	}
}

type SkillQueueEntry struct {
	FinishDate      time.Time `json:"finish_date"`
	FinishedLevel   int       `json:"finished_level"`
	LevelEndSP      int64     `json:"level_end_sp"`
	LevelStartSP    int64     `json:"level_start_sp"`
	QueuePosition   int       `json:"queue_position"`
	SkillID         int       `json:"skill_id"`
	StartDate       time.Time `json:"start_date"`
	TrainingStartSP int64     `json:"training_start_sp"`
}

type Skills struct {
	ActiveSkillLevel   int64 `json:"active_skill_level"`
	SkillID            int   `json:"skill_id"`
	SkillpointsInSkill int64 `json:"skillpoints_in_skill"`
	TrainedSkillLevel  int64 `json:"trained_skill_level"`
}

type SkillInfo struct {
	Skills        []Skills `json:"skills"`
	TotalSP       int64    `json:"total_sp"`
	UnallocatedSP int64    `json:"unallocated_sp"`
}

func (t *SkillTask) Execute(ctx *TaskContext) error {
	bgCtx := context.Background()

	var skillInfo SkillInfo
	path := fmt.Sprintf("/characters/%d/skills", ctx.CharacterID)
	if err := ctx.Client.Get(bgCtx, path, ctx.AccessToken, &skillInfo); err != nil {
		return fmt.Errorf("fetch skill info: %w", err)
	}

	var skillQueue []SkillQueueEntry
	path = fmt.Sprintf("/characters/%d/skillqueue", ctx.CharacterID)
	if err := ctx.Client.Get(bgCtx, path, ctx.AccessToken, &skillQueue); err != nil {
		return fmt.Errorf("fetch skill queue: %w", err)
	}

	tx := global.DB.Begin()
	if err := tx.Model(&model.EveCharacterSkill{}).
		Where("character_id = ?", ctx.CharacterID).
		FirstOrCreate(&model.EveCharacterSkill{
			CharacterID:   ctx.CharacterID,
			TotalSP:       skillInfo.TotalSP,
			UnallocatedSP: skillInfo.UnallocatedSP,
		}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("create or update skill: %w", err)
	}

	for _, skill := range skillInfo.Skills {
		if err := tx.Model(&model.EveCharacterSkills{}).
			Where("character_id = ? AND skill_id = ?", ctx.CharacterID, skill.SkillID).
			FirstOrCreate(&model.EveCharacterSkills{
				CharacterID:        ctx.CharacterID,
				SkillID:            skill.SkillID,
				ActiveLevel:        int(skill.ActiveSkillLevel),
				TrainedLevel:       int(skill.TrainedSkillLevel),
				SkillpointsInSkill: skill.SkillpointsInSkill,
			}).Error; err != nil {
			global.Logger.Warn("[ESI] 创建或更新技能记录失败",
				zap.Int64("character_id", ctx.CharacterID),
				zap.Int("skill_id", skill.SkillID),
				zap.Error(err),
			)
		}
	}

	if err := tx.Where("character_id = ?", ctx.CharacterID).Delete(&model.EveCharacterSkillQueue{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("delete old skill queue: %w", err)
	}

	if len(skillQueue) > 0 {
		var queueRecords []model.EveCharacterSkillQueue
		for _, q := range skillQueue {
			queueRecords = append(queueRecords, model.EveCharacterSkillQueue{
				CharacterID:     ctx.CharacterID,
				QueuePosition:   q.QueuePosition,
				SkillID:         q.SkillID,
				LevelEndSP:      q.LevelEndSP,
				LevelStartSP:    q.LevelStartSP,
				TrainingStartSP: q.TrainingStartSP,
				FinishedLevel:   q.FinishedLevel,
				StartDate:       q.StartDate.Unix(),
				FinishDate:      q.FinishDate.Unix(),
			})
		}
		if err := tx.Create(&queueRecords).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("insert skill queue: %w", err)
		}
	}

	tx.Commit()

	return nil
}
