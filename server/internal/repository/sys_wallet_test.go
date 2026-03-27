package repository

import (
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestSysWalletCountTransactionsByUserRefTypeInRangeUsesTimeBounds(t *testing.T) {
	db := newDryRunPostgresDB(t)
	startAt := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	endAt := startAt.AddDate(0, 1, 0)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&struct{}{}).
			Table("wallet_transaction").
			Where("user_id = ? AND ref_type = ? AND created_at >= ? AND created_at < ?", 42, "pap_fc_salary", startAt, endAt).
			Count(new(int64))
	})

	if !strings.Contains(sql, `ref_type = 'pap_fc_salary'`) {
		t.Fatalf("expected ref_type filter in SQL, got %s", sql)
	}
	if !strings.Contains(sql, `created_at >=`) || !strings.Contains(sql, `created_at <`) {
		t.Fatalf("expected month bounds in SQL, got %s", sql)
	}
}

func TestSysWalletTransactionLookupByUserRefTypeRefIDUsesAllFilters(t *testing.T) {
	db := newDryRunPostgresDB(t)
	startAt := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	endAt := startAt.AddDate(0, 1, 0)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Where(
			"user_id = ? AND ref_type = ? AND ref_id = ? AND created_at >= ? AND created_at < ?",
			42, "pap_fc_salary", "fleet-1", startAt, endAt,
		).
			First(&struct{}{})
	})

	if !strings.Contains(sql, `ref_type = 'pap_fc_salary'`) {
		t.Fatalf("expected ref_type filter in SQL, got %s", sql)
	}
	if !strings.Contains(sql, `ref_id = 'fleet-1'`) {
		t.Fatalf("expected ref_id filter in SQL, got %s", sql)
	}
	if !strings.Contains(sql, `created_at >=`) || !strings.Contains(sql, `created_at <`) {
		t.Fatalf("expected month bounds in SQL, got %s", sql)
	}
}
