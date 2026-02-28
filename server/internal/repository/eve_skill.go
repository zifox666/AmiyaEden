package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
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

func (r *EveSkillRepository) GetSkillList(characterID int) (*model.EveCharacterSkills, error) {
	var skills model.EveCharacterSkills
	err := global.DB.Where("character_id = ?", characterID).First(&skills).Error
	if err != nil {
		return nil, err
	}
	return &skills, nil
}

// SumTotalSPByCharacterIDs 汇总多个角色的总技能点
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
