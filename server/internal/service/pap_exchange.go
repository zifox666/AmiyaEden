package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
)

// PAPExchangeService PAP 兑换汇率业务逻辑层
type PAPExchangeService struct {
	rateRepo *repository.PAPTypeRateRepository
}

func NewPAPExchangeService() *PAPExchangeService {
	return &PAPExchangeService{
		rateRepo: repository.NewPAPTypeRateRepository(),
	}
}

// GetRates 获取所有 PAP 类型兑换汇率
func (s *PAPExchangeService) GetRates() ([]model.PAPTypeRate, error) {
	return s.rateRepo.List()
}

// SetRateRequest 更新 PAP 类型汇率的请求项
type SetRateRequest struct {
	PapType     string  `json:"pap_type"     binding:"required"`
	DisplayName string  `json:"display_name"`
	Rate        float64 `json:"rate"         binding:"required,gt=0"`
}

// SetRates 批量更新 PAP 类型兑换汇率
func (s *PAPExchangeService) SetRates(items []SetRateRequest) error {
	rates := make([]model.PAPTypeRate, 0, len(items))
	for _, item := range items {
		rates = append(rates, model.PAPTypeRate{
			PapType:     item.PapType,
			DisplayName: item.DisplayName,
			Rate:        item.Rate,
		})
	}
	return s.rateRepo.Save(rates)
}
