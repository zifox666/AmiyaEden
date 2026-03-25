package repository

import (
	"testing"

	"amiya-eden/internal/model"
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
		if model.SysConfigDefaultCorpID != 1 {
			t.Fatalf("expected default corp ID 1, got %d", model.SysConfigDefaultCorpID)
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
