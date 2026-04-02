package service

import internalutils "amiya-eden/internal/utils"

func normalizePage(page int) int {
	return internalutils.NormalizePage(page)
}

func normalizePageRequest(page *int, pageSize *int, defaultPageSize, maxPageSize int) {
	*page = normalizePage(*page)
	*pageSize = normalizePageSize(*pageSize, defaultPageSize, maxPageSize)
}

func normalizePageSize(pageSize, defaultPageSize, maxPageSize int) int {
	return internalutils.NormalizePageSize(pageSize, defaultPageSize, maxPageSize)
}

func normalizeLedgerPageRequest(page *int, pageSize *int) {
	*page = normalizePage(*page)
	*pageSize = normalizeLedgerPageSize(*pageSize)
}

func normalizeLedgerPageSize(pageSize int) int {
	return internalutils.NormalizeLedgerPageSize(pageSize)
}
