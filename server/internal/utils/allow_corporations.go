package utils

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
)

var allowCorporationsCache []int64

func GetAllowCorporations() []int64 {
	if allowCorporationsCache != nil {
		return allowCorporationsCache
	}

	repo := repository.NewSysConfigRepository()
	allowCorps, _ := repo.GetInt64Slice(model.SysConfigAllowCorporations, []int64{})
	allowCorporationsCache = allowCorps
	return allowCorps
}

func InvalidateAllowCorporationsCache() {
	allowCorporationsCache = nil
}
