package repository

import (
	"strings"
	"testing"

	"gorm.io/gorm"
)

func TestBuildPendingBadgeShopOrderCountQueryUsesRequestedStatus(t *testing.T) {
	db := newDryRunPostgresDB(t)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildPendingBadgeShopOrderCountQuery(tx).Count(new(int64))
	})

	if !strings.Contains(sql, `FROM "shop_order"`) {
		t.Fatalf("expected shop_order count query, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `status =`) {
		t.Fatalf("expected requested status filter on shop badge count query, got SQL: %s", sql)
	}
}
