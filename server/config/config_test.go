package config

import "testing"

func TestApplyDefaults(t *testing.T) {
	t.Run("empty allow corporations falls back to fuxi legion", func(t *testing.T) {
		cfg := &Config{}

		ApplyDefaults(cfg)

		if len(cfg.App.AllowCorporations) != 1 {
			t.Fatalf("expected one default corporation, got %v", cfg.App.AllowCorporations)
		}
		if cfg.App.AllowCorporations[0] != DefaultAllowCorporationID {
			t.Fatalf("expected default corporation %d, got %d", DefaultAllowCorporationID, cfg.App.AllowCorporations[0])
		}
	})

	t.Run("explicit allow corporations are preserved", func(t *testing.T) {
		cfg := &Config{}
		cfg.App.AllowCorporations = []int64{123, 456}

		ApplyDefaults(cfg)

		if len(cfg.App.AllowCorporations) != 2 || cfg.App.AllowCorporations[0] != 123 || cfg.App.AllowCorporations[1] != 456 {
			t.Fatalf("expected explicit allow corporations to be preserved, got %v", cfg.App.AllowCorporations)
		}
	})
}
