package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"time"
)

type EveSkillRepository struct{}

func NewEveSkillRepository() *EveSkillRepository {
	return &EveSkillRepository{}
}

func (r *EveSkillRepository) GetSkill(characterID int) (*model.EveCharacterSkill, error) {
	var skill model.EveCharacterSkill
	err := global.DB.Where("character_id = ?", characterID).First(&skill).Error
	if err != nil {
		return nil, err
	}
	return &skill, nil
}

func (r *EveSkillRepository) GetSkillList(characterID int) ([]model.EveCharacterSkills, error) {
	var skills []model.EveCharacterSkills
	err := global.DB.Where("character_id = ?", characterID).Find(&skills).Error
	if err != nil {
		return nil, err
	}
	return skills, nil
}

// GetSkillQueue 获取人物技能队列
func (r *EveSkillRepository) GetSkillQueue(characterID int) ([]model.EveCharacterSkillQueue, error) {
	var queue []model.EveCharacterSkillQueue
	err := global.DB.Where("character_id = ?", characterID).
		Order("queue_position ASC").Find(&queue).Error
	if err != nil {
		return nil, err
	}
	return queue, nil
}

// SumTotalSPByCharacterIDs 汇总多个人物的总技能点
func (r *EveSkillRepository) SumTotalSPByCharacterIDs(characterIDs []int64) (int64, error) {
	if len(characterIDs) == 0 {
		return 0, nil
	}
	var total int64
	err := global.DB.Model(&model.EveCharacterSkill{}).
		Where("character_id IN ?", characterIDs).
		Select("COALESCE(SUM(total_sp), 0)").
		Scan(&total).Error
	return total, err
}

func (r *EveSkillRepository) GetSkillTotalsByCharacterIDs(characterIDs []int64) (map[int64]int64, error) {
	result := make(map[int64]int64, len(characterIDs))
	if len(characterIDs) == 0 {
		return result, nil
	}
	type row struct {
		CharacterID int64
		TotalSP     int64
	}
	var rows []row
	err := global.DB.Model(&model.EveCharacterSkill{}).
		Select("character_id, total_sp").
		Where("character_id IN ?", characterIDs).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.CharacterID] = row.TotalSP
	}
	return result, nil
}

func (r *EveSkillRepository) GetLatestSkillUpdateTimeByCharacterIDs(characterIDs []int64) (*time.Time, error) {
	if len(characterIDs) == 0 {
		return nil, nil
	}
	var unixSeconds *int64
	err := global.DB.Model(&model.EveCharacterSkill{}).
		Where("character_id IN ?", characterIDs).
		Select("MAX(updated_time)").
		Scan(&unixSeconds).Error
	if err != nil {
		return nil, err
	}
	if unixSeconds == nil || *unixSeconds == 0 {
		return nil, nil
	}
	value := time.Unix(*unixSeconds, 0)
	return &value, nil
}
