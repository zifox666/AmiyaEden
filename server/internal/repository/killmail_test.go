package repository

import (
	"amiya-eden/internal/model"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestListVictimKillmailsQueryFiltersOnCharacterAndVictim(t *testing.T) {
	db := newDryRunPostgresDB(t)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		var list []model.EveCharacterKillmail
		return tx.Where("character_id = ? AND victim = ?", int64(12345), true).Find(&list)
	})

	if !strings.Contains(sql, `FROM "eve_character_killmail"`) {
		t.Fatalf("expected eve_character_killmail table, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `character_id = `) {
		t.Fatalf("expected character_id filter, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `victim = `) {
		t.Fatalf("expected victim filter, got SQL: %s", sql)
	}
}

func TestListKillmailsByIDsSinceQueryAppliesTimeFilterOrderAndLimit(t *testing.T) {
	db := newDryRunPostgresDB(t)
	since := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		var list []model.EveKillmailList
		return tx.Where("kill_mail_id IN ? AND kill_mail_time >= ?", []int64{100, 200}, since).
			Order("kill_mail_time DESC").
			Limit(200).
			Find(&list)
	})

	if !strings.Contains(sql, `FROM "eve_killmail_list"`) {
		t.Fatalf("expected eve_killmail_list table, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `kill_mail_id IN (`) {
		t.Fatalf("expected kill_mail_id IN filter, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `kill_mail_time >= `) {
		t.Fatalf("expected time >= filter, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `ORDER BY kill_mail_time DESC`) {
		t.Fatalf("expected descending time order, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `LIMIT `) {
		t.Fatalf("expected LIMIT clause, got SQL: %s", sql)
	}
}

func TestListKillmailsByIDsInTimeRangeQueryAppliesBetween(t *testing.T) {
	db := newDryRunPostgresDB(t)
	start := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		var list []model.EveKillmailList
		return tx.Where("kill_mail_id IN ? AND kill_mail_time BETWEEN ? AND ?", []int64{100}, start, end).
			Find(&list)
	})

	if !strings.Contains(sql, `FROM "eve_killmail_list"`) {
		t.Fatalf("expected eve_killmail_list table, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `BETWEEN`) {
		t.Fatalf("expected BETWEEN clause for time range, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `kill_mail_id IN (`) {
		t.Fatalf("expected kill_mail_id IN filter, got SQL: %s", sql)
	}
}

func TestListCharacterKillmailsByCharacterIDsQueryUsesINClause(t *testing.T) {
	db := newDryRunPostgresDB(t)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		var list []model.EveCharacterKillmail
		return tx.Where("character_id IN ?", []int64{111, 222, 333}).Find(&list)
	})

	if !strings.Contains(sql, `FROM "eve_character_killmail"`) {
		t.Fatalf("expected eve_character_killmail table, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `character_id IN (`) {
		t.Fatalf("expected character_id IN clause, got SQL: %s", sql)
	}
}

func TestListKillmailItemsQueryFiltersByKillmailID(t *testing.T) {
	db := newDryRunPostgresDB(t)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		var list []model.EveKillmailItem
		return tx.Where("kill_mail_id = ?", int64(99999)).Find(&list)
	})

	if !strings.Contains(sql, `FROM "eve_killmail_item"`) {
		t.Fatalf("expected eve_killmail_item table, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `kill_mail_id = `) {
		t.Fatalf("expected kill_mail_id filter, got SQL: %s", sql)
	}
}
