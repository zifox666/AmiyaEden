package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
)

// SkillPlanRepository 技能规划数据访问层
type SkillPlanRepository struct{}

func NewSkillPlanRepository() *SkillPlanRepository { return &SkillPlanRepository{} }

// Create 创建技能规划（含条目），在事务中操作
func (r *SkillPlanRepository) Create(plan *model.SkillPlan, items []model.SkillPlanItem) error {
	tx := global.DB.Begin()
	if err := tx.Create(plan).Error; err != nil {
		tx.Rollback()
		return err
	}
	for i := range items {
		items[i].SkillPlanID = plan.ID
	}
	if len(items) > 0 {
		if err := tx.Create(&items).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// Update 更新技能规划（删旧条目+建新条目）
func (r *SkillPlanRepository) Update(plan *model.SkillPlan, items []model.SkillPlanItem) error {
	tx := global.DB.Begin()
	if err := tx.Save(plan).Error; err != nil {
		tx.Rollback()
		return err
	}
	// 删除旧条目
	if err := tx.Where("skill_plan_id = ?", plan.ID).Delete(&model.SkillPlanItem{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	// 创建新条目
	for i := range items {
		items[i].SkillPlanID = plan.ID
	}
	if len(items) > 0 {
		if err := tx.Create(&items).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// Delete 删除技能规划及其条目
func (r *SkillPlanRepository) Delete(id uint) error {
	tx := global.DB.Begin()
	if err := tx.Where("skill_plan_id = ?", id).Delete(&model.SkillPlanItem{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&model.SkillPlan{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// GetByID 查询单个技能规划
func (r *SkillPlanRepository) GetByID(id uint) (*model.SkillPlan, error) {
	var plan model.SkillPlan
	err := global.DB.First(&plan, id).Error
	return &plan, err
}

// List 分页查询技能规划
func (r *SkillPlanRepository) List(page, pageSize int) ([]model.SkillPlan, int64, error) {
	var plans []model.SkillPlan
	var total int64
	db := global.DB.Model(&model.SkillPlan{})
	db.Count(&total)
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&plans).Error
	return plans, total, err
}

// ListAll 查询全部技能规划（不分页，用于下拉选项）
func (r *SkillPlanRepository) ListAll() ([]model.SkillPlan, error) {
	var plans []model.SkillPlan
	err := global.DB.Order("id DESC").Find(&plans).Error
	return plans, err
}

// GetItems 获取规划的所有条目
func (r *SkillPlanRepository) GetItems(planID uint) ([]model.SkillPlanItem, error) {
	var items []model.SkillPlanItem
	err := global.DB.Where("skill_plan_id = ?", planID).Find(&items).Error
	return items, err
}

// GetSkillsByCharacterIDs 批量获取指定角色的技能记录
func (r *SkillPlanRepository) GetSkillsByCharacterIDs(characterIDs []int64) ([]model.EveCharacterSkills, error) {
	if len(characterIDs) == 0 {
		return nil, nil
	}
	var skills []model.EveCharacterSkills
	err := global.DB.Where("character_id IN ?", characterIDs).Find(&skills).Error
	return skills, err
}
