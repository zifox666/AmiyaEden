package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
)

// SkillPlanRepository 军团技能计划数据访问层
type SkillPlanRepository struct{}

func NewSkillPlanRepository() *SkillPlanRepository {
	return &SkillPlanRepository{}
}

// Create 创建技能计划及技能要求
func (r *SkillPlanRepository) Create(plan *model.SkillPlan, skills []model.SkillPlanSkill) error {
	tx := global.DB.Begin()

	if err := tx.Create(plan).Error; err != nil {
		tx.Rollback()
		return err
	}

	for i := range skills {
		skills[i].SkillPlanID = plan.ID
	}
	if len(skills) > 0 {
		if err := tx.Create(&skills).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// GetByID 根据 ID 获取技能计划
func (r *SkillPlanRepository) GetByID(id uint) (*model.SkillPlan, error) {
	var plan model.SkillPlan
	err := global.DB.First(&plan, id).Error
	return &plan, err
}

// List 分页获取技能计划
func (r *SkillPlanRepository) List(page, pageSize int, keyword string) ([]model.SkillPlan, int64, error) {
	var plans []model.SkillPlan
	var total int64

	offset := (page - 1) * pageSize
	db := global.DB.Model(&model.SkillPlan{})

	if keyword != "" {
		pattern := "%" + keyword + "%"
		db = db.Where("title ILIKE ? OR description ILIKE ?", pattern, pattern)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("updated_at DESC, id DESC").Offset(offset).Limit(pageSize).Find(&plans).Error; err != nil {
		return nil, 0, err
	}

	return plans, total, nil
}

// ListAll 获取全部技能计划
func (r *SkillPlanRepository) ListAll() ([]model.SkillPlan, error) {
	var plans []model.SkillPlan
	err := global.DB.Order("updated_at DESC, id DESC").Find(&plans).Error
	return plans, err
}

// Update 更新技能计划及技能要求
func (r *SkillPlanRepository) Update(plan *model.SkillPlan, skills []model.SkillPlanSkill) error {
	tx := global.DB.Begin()

	if err := tx.Save(plan).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("skill_plan_id = ?", plan.ID).Delete(&model.SkillPlanSkill{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	for i := range skills {
		skills[i].ID = 0
		skills[i].SkillPlanID = plan.ID
	}
	if len(skills) > 0 {
		if err := tx.Create(&skills).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// Delete 删除技能计划及技能要求
func (r *SkillPlanRepository) Delete(id uint) error {
	tx := global.DB.Begin()

	if err := tx.Where("skill_plan_id = ?", id).Delete(&model.SkillPlanSkill{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&model.SkillPlan{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// ListSkillsByPlanID 获取技能计划下的所有技能要求
func (r *SkillPlanRepository) ListSkillsByPlanID(planID uint) ([]model.SkillPlanSkill, error) {
	var skills []model.SkillPlanSkill
	err := global.DB.Where("skill_plan_id = ?", planID).Order("sort ASC, id ASC").Find(&skills).Error
	return skills, err
}

// ListSkillsByPlanIDs 批量获取多个技能计划的技能要求
func (r *SkillPlanRepository) ListSkillsByPlanIDs(planIDs []uint) ([]model.SkillPlanSkill, error) {
	var skills []model.SkillPlanSkill
	if len(planIDs) == 0 {
		return skills, nil
	}
	err := global.DB.Where("skill_plan_id IN ?", planIDs).Order("sort ASC, id ASC").Find(&skills).Error
	return skills, err
}

// ListCheckCharacterIDsByUserID 获取用户保存的技能检查角色
func (r *SkillPlanRepository) ListCheckCharacterIDsByUserID(userID uint) ([]int64, error) {
	var rows []model.SkillPlanCheckCharacter
	if err := global.DB.Where("user_id = ?", userID).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]int64, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.CharacterID)
	}
	return result, nil
}

// ReplaceCheckCharacters 替换用户保存的技能检查角色
func (r *SkillPlanRepository) ReplaceCheckCharacters(userID uint, characterIDs []int64) error {
	tx := global.DB.Begin()

	if err := tx.Where("user_id = ?", userID).Delete(&model.SkillPlanCheckCharacter{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(characterIDs) > 0 {
		rows := make([]model.SkillPlanCheckCharacter, 0, len(characterIDs))
		for _, characterID := range characterIDs {
			rows = append(rows, model.SkillPlanCheckCharacter{
				UserID:      userID,
				CharacterID: characterID,
			})
		}
		if err := tx.Create(&rows).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}
