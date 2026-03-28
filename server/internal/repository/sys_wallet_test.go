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

func TestListTransactionsWithCharacterAppliesUserKeywordAcrossNicknameAndCharacterName(t *testing.T) {
	db := newDryRunPostgresDB(t)
	filter := WalletTransactionFilter{UserKeyword: "bee"}

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return applyWalletTransactionUserFilter(
			tx.Table("wallet_transaction wt").
				Joins(`LEFT JOIN "user" u ON wt.user_id = u.id`).
				Joins("LEFT JOIN eve_character ec ON u.primary_character_id = ec.character_id"),
			"wt.user_id",
			"wt.ref_type",
			filter,
		).Find(&[]struct{}{})
	})

	if !strings.Contains(sql, `LEFT JOIN "user" u ON wt.user_id = u.id`) {
		t.Fatalf("expected wallet transaction query to join user table, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `LEFT JOIN eve_character ec ON u.primary_character_id = ec.character_id`) {
		t.Fatalf("expected wallet transaction query to join primary character, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `LOWER(u.nickname) LIKE`) || !strings.Contains(sql, `LOWER(ec.character_name) LIKE`) {
		t.Fatalf("expected nickname and character name keyword search, got SQL: %s", sql)
	}
}
