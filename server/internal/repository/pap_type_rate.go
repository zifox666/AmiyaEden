package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
)

// PAPTypeRateRepository PAP 类型兑换汇率数据访问层
type PAPTypeRateRepository struct{}

func NewPAPTypeRateRepository() *PAPTypeRateRepository {
	return &PAPTypeRateRepository{}
}

// List 返回所有 PAP 类型汇率，缺失的类型用默认值补全并写入数据库
func (r *PAPTypeRateRepository) List() ([]model.PAPTypeRate, error) {
	var existing []model.PAPTypeRate
	if err := global.DB.Find(&existing).Error; err != nil {
		return nil, err
	}

	existingMap := make(map[string]model.PAPTypeRate, len(existing))
	for _, rate := range existing {
		existingMap[rate.PapType] = rate
	}

	// 按默认顺序返回，缺失的类型写入默认值
	result := make([]model.PAPTypeRate, 0, len(model.DefaultPAPTypeRates))
	for _, def := range model.DefaultPAPTypeRates {
		if rate, ok := existingMap[def.PapType]; ok {
			result = append(result, rate)
		} else {
			_ = global.DB.Create(&def).Error
			result = append(result, def)
		}
	}
	return result, nil
}

// Save 批量保存 PAP 类型汇率（upsert）
func (r *PAPTypeRateRepository) Save(rates []model.PAPTypeRate) error {
	return global.DB.Save(&rates).Error
}

// GetRateMap 返回 pap_type -> rate 的映射，供结算时快速查询
func (r *PAPTypeRateRepository) GetRateMap() map[string]float64 {
	rates, err := r.List()
	m := make(map[string]float64, len(model.DefaultPAPTypeRates))
	if err != nil {
		for _, d := range model.DefaultPAPTypeRates {
			m[d.PapType] = d.Rate
		}
		return m
	}
	for _, rate := range rates {
		m[rate.PapType] = rate.Rate
	}
	return m
}
