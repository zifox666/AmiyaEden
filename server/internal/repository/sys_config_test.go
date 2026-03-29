package repository

import (
	"strings"
	"testing"

	"amiya-eden/internal/model"

	"gorm.io/gorm"
)

func TestCacheKey(t *testing.T) {
	t.Run("returns correct cache key", func(t *testing.T) {
		key := cacheKey("test.key")
		expected := "sys_config:test.key"

		if key != expected {
			t.Fatalf("expected cache key %s, got %s", expected, key)
		}
	})

	t.Run("handles empty key", func(t *testing.T) {
		key := cacheKey("")
		expected := "sys_config:"

		if key != expected {
			t.Fatalf("expected cache key %s, got %s", expected, key)
		}
	})
}

func TestSysConfigConstants(t *testing.T) {
	t.Run("has valid allow corporations key", func(t *testing.T) {
		expected := "app.allow_corporations"
		if model.SysConfigAllowCorporations != expected {
			t.Fatalf("expected %s, got %s", expected, model.SysConfigAllowCorporations)
		}
	})

	t.Run("has valid default corp ID", func(t *testing.T) {
		if model.SystemCorporationID <= 0 {
			t.Fatalf("expected positive system corporation ID, got %d", model.SystemCorporationID)
		}
	})
}

func TestSystemConfigTableName(t *testing.T) {
	cfg := model.SystemConfig{}
	tableName := cfg.TableName()

	if tableName != "system_config" {
		t.Fatalf("expected table name 'system_config', got %s", tableName)
	}
}

func TestSysConfigRepository_New(t *testing.T) {
	repo := NewSysConfigRepository()

	if repo == nil {
		t.Fatal("expected non-nil repository")
	}
}

func TestBuildSysConfigBatchUpsertQueryUsesSingleUpsertStatement(t *testing.T) {
	db := newDryRunPostgresDB(t)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildSysConfigBatchUpsertQuery(tx, []SysConfigUpsertItem{
			{Key: model.SysConfigNewbroMaxCharacterSP, Value: "20000000", Desc: "max"},
			{Key: model.SysConfigNewbroBonusRate, Value: "20", Desc: "bonus"},
		})
	})

	if !strings.Contains(sql, `INSERT INTO "system_config"`) {
		t.Fatalf("expected system_config insert, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `VALUES (`) {
		t.Fatalf("expected batched values clause, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `ON CONFLICT ("key") DO UPDATE SET`) {
		t.Fatalf("expected upsert on key conflict, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `"value"="excluded"."value"`) {
		t.Fatalf("expected value column to be updated from excluded row, got SQL: %s", sql)
	}
}
