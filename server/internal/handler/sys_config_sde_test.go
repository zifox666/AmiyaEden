package handler

import (
	"amiya-eden/internal/model"
	"reflect"
	"testing"
)

// TestSDEConfigResponseJSONTags 测试 SDE 配置响应 JSON 标签
func TestSDEConfigResponseJSONTags(t *testing.T) {
	typ := reflect.TypeOf(SDEConfigResponse{})
	fields := []struct {
		name      string
		jsonTag   string
		fieldType string
	}{
		{"APIKey", "api_key", "string"},
		{"Proxy", "proxy", "string"},
		{"DownloadURL", "download_url", "string"},
	}

	for _, f := range fields {
		field, ok := typ.FieldByName(f.name)
		if !ok {
			t.Fatalf("SDEConfigResponse missing field %s", f.name)
		}
		if field.Tag.Get("json") != f.jsonTag {
			t.Fatalf("field %s: json tag = %q, want %q", f.name, field.Tag.Get("json"), f.jsonTag)
		}
		if field.Type.Kind().String() != f.fieldType {
			t.Fatalf("field %s: type = %v, want %s", f.name, field.Type.Kind(), f.fieldType)
		}
	}
}

// TestUpdateSDEConfigRequestJSONTags 测试 SDE 配置更新请求 JSON 标签
func TestUpdateSDEConfigRequestJSONTags(t *testing.T) {
	typ := reflect.TypeOf(UpdateSDEConfigRequest{})
	fields := []struct {
		name      string
		jsonTag   string
		fieldType string
	}{
		{"APIKey", "api_key", "ptr"},
		{"Proxy", "proxy", "ptr"},
		{"DownloadURL", "download_url", "ptr"},
	}

	for _, f := range fields {
		field, ok := typ.FieldByName(f.name)
		if !ok {
			t.Fatalf("UpdateSDEConfigRequest missing field %s", f.name)
		}
		if field.Tag.Get("json") != f.jsonTag {
			t.Fatalf("field %s: json tag = %q, want %q", f.name, field.Tag.Get("json"), f.jsonTag)
		}
		if f.fieldType == "ptr" && field.Type.Kind() != reflect.Ptr {
			t.Fatalf("field %s: type = %v, want ptr", f.name, field.Type.Kind())
		}
	}
}

// TestSDEConfigDefaultValues 测试 SDE 配置默认值
func TestSDEConfigDefaultValues(t *testing.T) {
	tests := []struct {
		name     string
		actual   string
		expected string
	}{
		{name: "APIKey", actual: model.SysConfigDefaultSDEAPIKey, expected: "modify_your_api_key"},
		{name: "Proxy", actual: model.SysConfigDefaultSDEProxy, expected: ""},
		{name: "DownloadURL", actual: model.SysConfigDefaultSDEDownloadURL, expected: "https://api.github.com/repos/garveen/eve-sde-converter/releases/latest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actual != tt.expected {
				t.Fatalf("default %s = %q, want %q", tt.name, tt.actual, tt.expected)
			}
		})
	}
}

// TestSDEConfigKeys 测试 SDE 配置键常量
func TestSDEConfigKeys(t *testing.T) {
	tests := []struct {
		name     string
		actual   string
		expected string
	}{
		{name: "APIKey", actual: model.SysConfigSDEAPIKey, expected: "sde.api_key"},
		{name: "Proxy", actual: model.SysConfigSDEProxy, expected: "sde.proxy"},
		{name: "DownloadURL", actual: model.SysConfigSDEDownloadURL, expected: "sde.download_url"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actual != tt.expected {
				t.Fatalf("config key = %q, want %q", tt.actual, tt.expected)
			}
		})
	}
}
