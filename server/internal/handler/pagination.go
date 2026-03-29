package handler

import "amiya-eden/internal/utils"

func normalizePage(page int) int {
	return utils.NormalizePage(page)
}

func normalizePageSize(pageSize, defaultPageSize, maxPageSize int) int {
	return utils.NormalizePageSize(pageSize, defaultPageSize, maxPageSize)
}

func normalizeLedgerPageSize(pageSize int) int {
	return utils.NormalizeLedgerPageSize(pageSize)
}

func normalizePagination(page, pageSize, defaultPageSize, maxPageSize int) (int, int) {
	return normalizePage(page), normalizePageSize(pageSize, defaultPageSize, maxPageSize)
}

func normalizeLedgerPagination(page, pageSize int) (int, int) {
	return normalizePage(page), normalizeLedgerPageSize(pageSize)
}
