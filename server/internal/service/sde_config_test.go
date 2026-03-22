package service

import (
	"amiya-eden/internal/model"
	"testing"
)

// TestSdeConfigDefaults 测试 SDE 配置默认值常量
func TestSdeConfigDefaults(t *testing.T) {
	tests := []struct {
		name     string
		actual   string
		expected string
	}{
		{
			name:     "默认 API Key",
			actual:   model.SysConfigDefaultSDEAPIKey,
			expected: "modify_your_api_key",
		},
		{
			name:     "默认代理为空",
			actual:   model.SysConfigDefaultSDEProxy,
			expected: "",
		},
		{
			name:     "默认下载地址",
			actual:   model.SysConfigDefaultSDEDownloadURL,
			expected: "https://api.github.com/repos/garveen/eve-sde-converter/releases/latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actual != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, tt.actual)
			}
		})
	}
}

// TestSdeConfigKeys 测试 SDE 配置键常量
func TestSdeConfigKeys(t *testing.T) {
	tests := []struct {
		name     string
		actual   string
		expected string
	}{
		{name: "API Key 配置键", actual: model.SysConfigSDEAPIKey, expected: "sde.api_key"},
		{name: "代理配置键", actual: model.SysConfigSDEProxy, expected: "sde.proxy"},
		{name: "下载地址配置键", actual: model.SysConfigSDEDownloadURL, expected: "sde.download_url"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actual != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, tt.actual)
			}
		})
	}
}
