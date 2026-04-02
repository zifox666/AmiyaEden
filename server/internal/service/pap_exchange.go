package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
)

// PAPExchangeService PAP 兑换汇率业务逻辑层
type PAPExchangeService struct {
	rateRepo   papExchangeRateStore
	configRepo papExchangeConfigStore
}

type papExchangeRateStore interface {
	List() ([]model.PAPTypeRate, error)
	Save(rates []model.PAPTypeRate) error
}

type papExchangeConfigStore interface {
	GetFloat(key string, defaultVal float64) float64
	GetInt(key string, defaultVal int) int
	SetMany(items []repository.SysConfigUpsertItem) error
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
	items := newSysConfigBatch(2).
		AddFloat64(model.SysConfigPAPFCSalary, *req.FCSalary, "FC工资").
		AddInt(model.SysConfigPAPFCSalaryLimit, *req.FCSalaryMonthlyLimit, "FC工资每月上限次数").
		Items()
	if err := s.configRepo.SetMany(items); err != nil {
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
