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
