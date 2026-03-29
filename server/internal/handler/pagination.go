package handler

import (
	"amiya-eden/internal/utils"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

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

func parsePaginationQuery(c *gin.Context, defaultPageSize, maxPageSize int) (int, int, error) {
	page, err := parseIntQuery(c, "current", utils.FirstPage)
	if err != nil {
		return 0, 0, err
	}

	pageSize, err := parseIntQuery(c, "size", defaultPageSize)
	if err != nil {
		return 0, 0, err
	}

	page, pageSize = normalizePagination(page, pageSize, defaultPageSize, maxPageSize)
	return page, pageSize, nil
}

func parseUnboundedPaginationQuery(c *gin.Context, defaultPageSize int) (int, int, error) {
	page, err := parseIntQuery(c, "current", utils.FirstPage)
	if err != nil {
		return 0, 0, err
	}

	pageSize, err := parseIntQuery(c, "size", defaultPageSize)
	if err != nil {
		return 0, 0, err
	}

	page = normalizePage(page)
	if pageSize < utils.FirstPage {
		pageSize = defaultPageSize
	}

	return page, pageSize, nil
}

func parseLedgerPaginationQuery(c *gin.Context, defaultPageSize int) (int, int, error) {
	page, err := parseIntQuery(c, "current", utils.FirstPage)
	if err != nil {
		return 0, 0, err
	}

	pageSize, err := parseIntQuery(c, "size", defaultPageSize)
	if err != nil {
		return 0, 0, err
	}

	page = normalizePage(page)
	if pageSize < utils.FirstPage {
		pageSize = defaultPageSize
	} else {
		pageSize = normalizeLedgerPageSize(pageSize)
	}

	return page, pageSize, nil
}

func parseIntQuery(c *gin.Context, key string, defaultValue int) (int, error) {
	raw := c.Query(key)
	if raw == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s query parameter: expected integer", key)
	}

	return value, nil
}
