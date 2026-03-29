package repository

import (
	internalsearch "amiya-eden/internal/search"
	"strings"

	"gorm.io/gorm"
)

func applyKeywordLikeFilter(query *gorm.DB, keyword string, predicates ...string) *gorm.DB {
	pattern, ok := internalsearch.BuildCaseInsensitiveLikePattern(keyword)
	if !ok {
		return query
	}

	conditions := make([]string, 0, len(predicates))
	args := make([]any, 0, len(predicates))
	for _, predicate := range predicates {
		if strings.TrimSpace(predicate) == "" {
			continue
		}
		conditions = append(conditions, predicate)
		args = append(args, pattern)
	}

	if len(conditions) == 0 {
		return query
	}

	return query.Where("("+strings.Join(conditions, " OR ")+")", args...)
}
