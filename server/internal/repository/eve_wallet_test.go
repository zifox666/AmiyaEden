package repository

import (
	"amiya-eden/internal/model"
	"strings"
	"testing"

	"gorm.io/gorm"
)

func TestGetWalletJournalsAppliesRefTypeFilterWhenProvided(t *testing.T) {
	db := newDryRunPostgresDB(t)
	repo := &EveWalletRepository{}

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return repo.getWalletJournalsQuery(tx, 90000001, []string{"bounty_prizes", "ess_escrow_transfer"}).
			Order("date DESC").
			Offset(0).
			Limit(20).
			Find(&[]model.EVECharacterWalletJournal{})
	})

	if !strings.Contains(sql, `WHERE character_id =`) {
		t.Fatalf("expected character filter, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `ref_type IN (`) {
		t.Fatalf("expected ref_type filter, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `'bounty_prizes'`) || !strings.Contains(sql, `'ess_escrow_transfer'`) {
		t.Fatalf("expected caller supplied ref types, got SQL: %s", sql)
	}
}

func TestListWalletJournalRefTypesUsesDistinctRefTypeQuery(t *testing.T) {
	db := newDryRunPostgresDB(t)
	repo := &EveWalletRepository{}

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return repo.getWalletJournalRefTypesQuery(tx, 90000001).
			Pluck("ref_type", &[]string{})
	})

	if !strings.Contains(sql, `SELECT DISTINCT`) {
		t.Fatalf("expected distinct query, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `WHERE character_id =`) {
		t.Fatalf("expected character filter, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `ORDER BY ref_type ASC`) {
		t.Fatalf("expected ref_type ordering, got SQL: %s", sql)
	}
}
