package utils

const (
	FirstPage             = 1
	LedgerDefaultPageSize = 200
	LedgerMaxPageSize     = 1000
)

func NormalizePage(page int) int {
	if page < FirstPage {
		return FirstPage
	}
	return page
}

func NormalizePageSize(pageSize, defaultPageSize, maxPageSize int) int {
	if defaultPageSize < FirstPage {
		defaultPageSize = FirstPage
	}
	if maxPageSize < defaultPageSize {
		maxPageSize = defaultPageSize
	}
	if pageSize < FirstPage || pageSize > maxPageSize {
		return defaultPageSize
	}
	return pageSize
}

func NormalizeLedgerPageSize(pageSize int) int {
	if pageSize > LedgerMaxPageSize {
		return LedgerMaxPageSize
	}
	return NormalizePageSize(pageSize, LedgerDefaultPageSize, LedgerMaxPageSize)
}
