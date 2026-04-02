package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"strconv"
	"testing"
)

type fakePAPExchangeRateStore struct {
	listedRates []model.PAPTypeRate
	savedRates  []model.PAPTypeRate
	saveErr     error
	listErr     error
}

func (f *fakePAPExchangeRateStore) List() ([]model.PAPTypeRate, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return append([]model.PAPTypeRate(nil), f.listedRates...), nil
}

func (f *fakePAPExchangeRateStore) Save(rates []model.PAPTypeRate) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.savedRates = append([]model.PAPTypeRate(nil), rates...)
	f.listedRates = append([]model.PAPTypeRate(nil), rates...)
	return nil
}

type fakePAPExchangeConfigStore struct {
	fcSalary             float64
	fcSalaryMonthlyLimit int
	setManyCalls         int
	setManyItems         []repository.SysConfigUpsertItem
	setManyErr           error
	hasSalary            bool
	hasLimit             bool
}

func (f *fakePAPExchangeConfigStore) GetFloat(key string, defaultVal float64) float64 {
	if key == model.SysConfigPAPFCSalary && f.hasSalary {
		return f.fcSalary
	}
	return defaultVal
}

func (f *fakePAPExchangeConfigStore) GetInt(key string, defaultVal int) int {
	if key == model.SysConfigPAPFCSalaryLimit && f.hasLimit {
		return f.fcSalaryMonthlyLimit
	}
	return defaultVal
}

func (f *fakePAPExchangeConfigStore) SetMany(items []repository.SysConfigUpsertItem) error {
	if f.setManyErr != nil {
		return f.setManyErr
	}
	f.setManyCalls++
	f.setManyItems = append([]repository.SysConfigUpsertItem(nil), items...)
	for _, item := range items {
		switch item.Key {
		case model.SysConfigPAPFCSalary:
			value, err := strconv.ParseFloat(item.Value, 64)
			if err != nil {
				return err
			}
			f.fcSalary = value
			f.hasSalary = true
		case model.SysConfigPAPFCSalaryLimit:
			value, err := strconv.Atoi(item.Value)
			if err != nil {
				return err
			}
			f.fcSalaryMonthlyLimit = value
			f.hasLimit = true
		}
	}
	return nil
}

func TestPAPExchangeUpdateConfigPersistsSingleBatch(t *testing.T) {
	rateStore := &fakePAPExchangeRateStore{}
	configStore := &fakePAPExchangeConfigStore{}
	svc := &PAPExchangeService{rateRepo: rateStore, configRepo: configStore}
	fcSalary := 5.5
	fcSalaryMonthlyLimit := 3

	updated, err := svc.UpdateConfig(&UpdateConfigRequest{
		Rates:                []SetRateRequest{{PapType: "cta", DisplayName: "CTA", Rate: 1.5}},
		FCSalary:             &fcSalary,
		FCSalaryMonthlyLimit: &fcSalaryMonthlyLimit,
	})
	if err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}
	if configStore.setManyCalls != 1 {
		t.Fatalf("expected exactly one batch write, got %d", configStore.setManyCalls)
	}
	if len(configStore.setManyItems) != 2 {
		t.Fatalf("expected 2 config items, got %d", len(configStore.setManyItems))
	}
	if updated.FCSalary != fcSalary {
		t.Fatalf("expected fc salary %v, got %v", fcSalary, updated.FCSalary)
	}
	if updated.FCSalaryMonthlyLimit != fcSalaryMonthlyLimit {
		t.Fatalf("expected monthly limit %d, got %d", fcSalaryMonthlyLimit, updated.FCSalaryMonthlyLimit)
	}
	if len(updated.Rates) != 1 || updated.Rates[0].Rate != 1.5 {
		t.Fatalf("expected updated PAP rates to round-trip, got %+v", updated.Rates)
	}
}
