package utils

import (
	"fmt"
	"slices"

	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
)

var allowCorporationsCache []int64

func ValidateAllowCorporations(corporationIDs []int64) error {
	for _, corporationID := range corporationIDs {
		if corporationID <= 0 {
			return fmt.Errorf("corporation id must be positive: %d", corporationID)
		}
	}

	return nil
}

func NormalizeAllowCorporations(corporationIDs []int64) []int64 {
	normalized := make([]int64, 0, len(corporationIDs)+1)
	seen := map[int64]struct{}{
		model.SystemCorporationID: {},
	}

	normalized = append(normalized, model.SystemCorporationID)
	for _, corporationID := range corporationIDs {
		if corporationID <= 0 {
			continue
		}
		if _, exists := seen[corporationID]; exists {
			continue
		}
		seen[corporationID] = struct{}{}
		normalized = append(normalized, corporationID)
	}

	return normalized
}

func GetAllowCorporations() []int64 {
	if allowCorporationsCache != nil {
		return slices.Clone(allowCorporationsCache)
	}

	repo := repository.NewSysConfigRepository()
	allowCorps, _ := repo.GetInt64Slice(model.SysConfigAllowCorporations, []int64{})
	allowCorporationsCache = NormalizeAllowCorporations(allowCorps)
	return slices.Clone(allowCorporationsCache)
}

func InvalidateAllowCorporationsCache() {
	allowCorporationsCache = nil
}
