package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"strconv"
)

// PAPExchangeService PAP 兑换汇率业务逻辑层
type PAPExchangeService struct {
	rateRepo   *repository.PAPTypeRateRepository
	configRepo *repository.SysConfigRepository
}

func NewPAPExchangeService() *PAPExchangeService {
	return &PAPExchangeService{
		rateRepo:   repository.NewPAPTypeRateRepository(),
		configRepo: repository.NewSysConfigRepository(),
	}
}

// PAPExchangeConfigResponse PAP 兑换配置响应
type PAPExchangeConfigResponse struct {
	Rates                []model.PAPTypeRate `json:"rates"`
	FCSalary             float64             `json:"fc_salary"`
	FCSalaryMonthlyLimit int                 `json:"fc_salary_monthly_limit"`
}

// GetConfig 获取所有 PAP 类型兑换汇率与 FC 工资
func (s *PAPExchangeService) GetConfig() (*PAPExchangeConfigResponse, error) {
	rates, err := s.rateRepo.List()
	if err != nil {
		return nil, err
	}
	return &PAPExchangeConfigResponse{
		Rates:                rates,
		FCSalary:             s.getFCSalary(),
		FCSalaryMonthlyLimit: s.getFCSalaryMonthlyLimit(),
	}, nil
}

// SetRateRequest 更新 PAP 类型汇率的请求项
type SetRateRequest struct {
	PapType     string  `json:"pap_type"     binding:"required"`
	DisplayName string  `json:"display_name"`
	Rate        float64 `json:"rate"         binding:"required,gt=0"`
}

// UpdateConfigRequest 更新 PAP 兑换配置请求
type UpdateConfigRequest struct {
	Rates                []SetRateRequest `json:"rates" binding:"required"`
	FCSalary             *float64         `json:"fc_salary" binding:"required,gte=0"`
	FCSalaryMonthlyLimit *int             `json:"fc_salary_monthly_limit" binding:"required,gte=0"`
}

// UpdateConfig 批量更新 PAP 类型兑换汇率与 FC 工资
func (s *PAPExchangeService) UpdateConfig(req *UpdateConfigRequest) (*PAPExchangeConfigResponse, error) {
	if err := s.SetRates(req.Rates); err != nil {
		return nil, err
	}
	if err := s.configRepo.Set(model.SysConfigPAPFCSalary, formatFloat(*req.FCSalary), "FC工资"); err != nil {
		return nil, err
	}
	if err := s.configRepo.Set(model.SysConfigPAPFCSalaryLimit, strconv.Itoa(*req.FCSalaryMonthlyLimit), "FC工资每月上限次数"); err != nil {
		return nil, err
	}
	return s.GetConfig()
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

func (s *PAPExchangeService) getFCSalary() float64 {
	return s.configRepo.GetFloat(model.SysConfigPAPFCSalary, model.SysConfigDefaultPAPFCSalary)
}

func (s *PAPExchangeService) getFCSalaryMonthlyLimit() int {
	return s.configRepo.GetInt(model.SysConfigPAPFCSalaryLimit, model.SysConfigDefaultPAPFCSalaryLimit)
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
