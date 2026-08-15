package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"time"

	"gorm.io/gorm"
)

type ATScoreRepository struct{}

func NewATScoreRepository() *ATScoreRepository { return &ATScoreRepository{} }

func (r *ATScoreRepository) CreateRule(rule *model.ATScoreRuleVersion, ships []model.ATScoreShip, banned []model.ATScoreBannedShip, penalties []model.ATScoreDuplicatePenalty) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(rule).Error; err != nil {
			return err
		}
		for i := range ships {
			ships[i].RuleVersionID = rule.ID
		}
		for i := range banned {
			banned[i].RuleVersionID = rule.ID
		}
		for i := range penalties {
			penalties[i].RuleVersionID = rule.ID
		}
		if len(ships) > 0 {
			if err := tx.Create(&ships).Error; err != nil {
				return err
			}
		}
		if len(banned) > 0 {
			if err := tx.Create(&banned).Error; err != nil {
				return err
			}
		}
		if len(penalties) > 0 {
			if err := tx.Create(&penalties).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ATScoreRepository) Publish(id uint) error {
	now := time.Now()
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.ATScoreRuleVersion{}).Where("status = ?", model.ATScoreRulePublished).Update("status", "archived").Error; err != nil {
			return err
		}
		result := tx.Model(&model.ATScoreRuleVersion{}).Where("id = ? AND status = ?", id, model.ATScoreRuleDraft).Updates(map[string]any{"status": model.ATScoreRulePublished, "published_at": &now})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (r *ATScoreRepository) GetRule(id uint) (*model.ATScoreRuleVersion, error) {
	var result model.ATScoreRuleVersion
	err := global.DB.First(&result, id).Error
	return &result, err
}
func (r *ATScoreRepository) ListRules() ([]model.ATScoreRuleVersion, error) {
	var result []model.ATScoreRuleVersion
	return result, global.DB.Order("id DESC").Find(&result).Error
}
func (r *ATScoreRepository) GetPublishedRule() (*model.ATScoreRuleVersion, error) {
	var result model.ATScoreRuleVersion
	err := global.DB.Where("status = ?", model.ATScoreRulePublished).Order("published_at DESC").First(&result).Error
	return &result, err
}
func (r *ATScoreRepository) GetShips(ruleID uint) ([]model.ATScoreShip, error) {
	var result []model.ATScoreShip
	return result, global.DB.Where("rule_version_id = ?", ruleID).Find(&result).Error
}
func (r *ATScoreRepository) GetBanned(ruleID uint) ([]model.ATScoreBannedShip, error) {
	var result []model.ATScoreBannedShip
	return result, global.DB.Where("rule_version_id = ?", ruleID).Find(&result).Error
}
func (r *ATScoreRepository) GetPenalties(ruleID uint) ([]model.ATScoreDuplicatePenalty, error) {
	var result []model.ATScoreDuplicatePenalty
	return result, global.DB.Where("rule_version_id = ?", ruleID).Find(&result).Error
}
