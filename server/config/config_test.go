package config

import "testing"

func TestApplyDefaults(t *testing.T) {
	t.Run("eve endpoint defaults are applied", func(t *testing.T) {
		cfg := &Config{}
		ApplyDefaults(cfg)
		if cfg.EveSSO.ESIBaseURL != DefaultESIBaseURL {
			t.Fatalf("expected ESIBaseURL %q, got %q", DefaultESIBaseURL, cfg.EveSSO.ESIBaseURL)
		}
		if cfg.EveSSO.ESIAPIPrefix != DefaultESIAPIPrefix {
			t.Fatalf("expected ESIAPIPrefix %q, got %q", DefaultESIAPIPrefix, cfg.EveSSO.ESIAPIPrefix)
		}
		if cfg.EveSSO.SSOAuthorizeURL != DefaultSSOAuthorizeURL {
			t.Fatalf("expected SSOAuthorizeURL %q, got %q", DefaultSSOAuthorizeURL, cfg.EveSSO.SSOAuthorizeURL)
		}
		if cfg.EveSSO.SSOTokenURL != DefaultSSOTokenURL {
			t.Fatalf("expected SSOTokenURL %q, got %q", DefaultSSOTokenURL, cfg.EveSSO.SSOTokenURL)
		}
		if cfg.EveSSO.EVEImagesBaseURL != DefaultEVEImagesBaseURL {
			t.Fatalf("expected EVEImagesBaseURL %q, got %q", DefaultEVEImagesBaseURL, cfg.EveSSO.EVEImagesBaseURL)
		}
	})

	t.Run("eve endpoint explicit values are normalized", func(t *testing.T) {
		cfg := &Config{
			EveSSO: EveSSOConfig{
				ESIBaseURL:       "https://esi.evetech.net///",
				ESIAPIPrefix:     "latest/",
				SSOAuthorizeURL:  "https://login.eveonline.com/v2/oauth/authorize/",
				SSOTokenURL:      "https://login.eveonline.com/v2/oauth/token/",
				EVEImagesBaseURL: "https://images.evetech.net/",
			},
		}
		ApplyDefaults(cfg)
		if cfg.EveSSO.ESIBaseURL != "https://esi.evetech.net" {
			t.Fatalf("expected normalized ESIBaseURL, got %q", cfg.EveSSO.ESIBaseURL)
		}
		if cfg.EveSSO.ESIAPIPrefix != "/latest" {
			t.Fatalf("expected normalized ESIAPIPrefix, got %q", cfg.EveSSO.ESIAPIPrefix)
		}
		if cfg.EveSSO.SSOAuthorizeURL != "https://login.eveonline.com/v2/oauth/authorize" {
			t.Fatalf("expected normalized SSOAuthorizeURL, got %q", cfg.EveSSO.SSOAuthorizeURL)
		}
		if cfg.EveSSO.SSOTokenURL != "https://login.eveonline.com/v2/oauth/token" {
			t.Fatalf("expected normalized SSOTokenURL, got %q", cfg.EveSSO.SSOTokenURL)
		}
		if cfg.EveSSO.EVEImagesBaseURL != "https://images.evetech.net" {
			t.Fatalf("expected normalized EVEImagesBaseURL, got %q", cfg.EveSSO.EVEImagesBaseURL)
		}
	})
}
