package service

import internalutils "amiya-eden/internal/utils"

func normalizePage(page int) int {
	return internalutils.NormalizePage(page)
}

func normalizePageSize(pageSize, defaultPageSize, maxPageSize int) int {
	return internalutils.NormalizePageSize(pageSize, defaultPageSize, maxPageSize)
}

func normalizeLedgerPageSize(pageSize int) int {
	return internalutils.NormalizeLedgerPageSize(pageSize)
}
