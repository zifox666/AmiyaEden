package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"

	"gorm.io/gorm"
)

type ATSetRepository struct{}

func NewATSetRepository() *ATSetRepository { return &ATSetRepository{} }

type ATSetSlotBundle struct {
	Slot         model.ATSetSlot          `json:"slot"`
	Fit          *model.ATFit             `json:"fit,omitempty"`
	Replacements []model.ATFitReplacement `json:"replacements"`
}
type ATFitDetail struct {
	Fit          model.ATFit              `json:"fit"`
	Replacements []model.ATFitReplacement `json:"replacements"`
}
type ATSetDetail struct {
	Set   model.ATSet               `json:"set"`
	Slots []ATSetSlotBundle         `json:"slots"`
	Score model.ATScoreRosterResult `json:"score"`
}

func (r *ATSetRepository) Save(set *model.ATSet, bundles []ATSetSlotBundle) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if set.ID == 0 {
			if err := tx.Create(set).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&model.ATSet{}).Where("id = ?", set.ID).Updates(map[string]any{"name": set.Name, "format": set.Format, "rule_version_id": set.RuleVersionID, "status": set.Status}).Error; err != nil {
				return err
			}
			var oldSlotIDs []uint
			if err := tx.Model(&model.ATSetSlot{}).Where("set_id = ?", set.ID).Pluck("id", &oldSlotIDs).Error; err != nil {
				return err
			}
			if len(oldSlotIDs) > 0 {
				if err := tx.Unscoped().Where("set_slot_id IN ?", oldSlotIDs).Delete(&model.ATFitReplacement{}).Error; err != nil {
					return err
				}
			}
			if err := tx.Unscoped().Where("set_id = ?", set.ID).Delete(&model.ATSetSlot{}).Error; err != nil {
				return err
			}
		}
		for i := range bundles {
			bundles[i].Slot.SetID = set.ID
			if err := tx.Create(&bundles[i].Slot).Error; err != nil {
				return err
			}
			for j := range bundles[i].Replacements {
				bundles[i].Replacements[j].SetSlotID = bundles[i].Slot.ID
			}
			if len(bundles[i].Replacements) > 0 {
				if err := tx.Create(&bundles[i].Replacements).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *ATSetRepository) Get(id uint) (*model.ATSet, []ATSetSlotBundle, error) {
	var set model.ATSet
	if err := global.DB.First(&set, id).Error; err != nil {
		return nil, nil, err
	}
	var slots []model.ATSetSlot
	if err := global.DB.Where("set_id = ?", id).Order("slot_order ASC").Find(&slots).Error; err != nil {
		return nil, nil, err
	}
	bundles := make([]ATSetSlotBundle, len(slots))
	for i, slot := range slots {
		bundles[i].Slot = slot
		if slot.FitID != nil {
			var fit model.ATFit
			if err := global.DB.First(&fit, *slot.FitID).Error; err != nil {
				return nil, nil, err
			}
			bundles[i].Fit = &fit
			if err := global.DB.Where("fit_id = ?", fit.ID).Order("id ASC").Find(&bundles[i].Replacements).Error; err != nil {
				return nil, nil, err
			}
		} else if err := global.DB.Where("set_slot_id = ?", slot.ID).Order("id ASC").Find(&bundles[i].Replacements).Error; err != nil {
			return nil, nil, err
		}
	}
	return &set, bundles, nil
}

func (r *ATSetRepository) List() ([]model.ATSet, error) {
	var result []model.ATSet
	return result, global.DB.Order("id DESC").Find(&result).Error
}
func (r *ATSetRepository) Delete(id uint) error { return global.DB.Delete(&model.ATSet{}, id).Error }

func (r *ATSetRepository) SaveFit(fit *model.ATFit, replacements []model.ATFitReplacement) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if fit.ID == 0 {
			if err := tx.Create(fit).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&model.ATFit{}).Where("id = ?", fit.ID).Updates(map[string]any{
				"name": fit.Name, "ship_type_id": fit.ShipTypeID, "ship_name": fit.ShipName,
				"base_eft": fit.BaseEFT, "base_e_war_tags": fit.BaseEWarTags,
				"base_bonus_tags": fit.BaseBonusTags, "status": fit.Status,
			}).Error; err != nil {
				return err
			}
			if err := tx.Unscoped().Where("fit_id = ?", fit.ID).Delete(&model.ATFitReplacement{}).Error; err != nil {
				return err
			}
		}
		for i := range replacements {
			replacements[i].FitID = &fit.ID
			replacements[i].SetSlotID = 0
		}
		if len(replacements) > 0 {
			return tx.Create(&replacements).Error
		}
		return nil
	})
}

func (r *ATSetRepository) GetFit(id uint) (*ATFitDetail, error) {
	var detail ATFitDetail
	if err := global.DB.First(&detail.Fit, id).Error; err != nil {
		return nil, err
	}
	if err := global.DB.Where("fit_id = ?", id).Order("id ASC").Find(&detail.Replacements).Error; err != nil {
		return nil, err
	}
	return &detail, nil
}

func (r *ATSetRepository) ListFits() ([]model.ATFit, error) {
	var result []model.ATFit
	return result, global.DB.Order("ship_name ASC, name ASC, id ASC").Find(&result).Error
}

func (r *ATSetRepository) DeleteFit(id uint) error {
	var count int64
	if err := global.DB.Model(&model.ATSetSlot{}).Where("fit_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return gorm.ErrInvalidData
	}
	return global.DB.Delete(&model.ATFit{}, id).Error
}
