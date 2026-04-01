package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"amiya-eden/internal/service"

	"github.com/gin-gonic/gin"
)

type fakeMentorAdminSettingsService struct {
	getSettingsResp service.MentorSettings
	updateResp      service.MentorSettings
	updateErr       error
	updateCalls     int
	lastUpdate      service.MentorSettings
}

func (f *fakeMentorAdminSettingsService) GetSettings() service.MentorSettings {
	return f.getSettingsResp
}

func (f *fakeMentorAdminSettingsService) UpdateSettings(cfg service.MentorSettings) (service.MentorSettings, error) {
	f.updateCalls++
	f.lastUpdate = cfg
	if f.updateErr != nil {
		return service.MentorSettings{}, f.updateErr
	}
	return f.updateResp, nil
}

func TestMentorAdminGetSettingsReturnsTypedPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/system/mentor/settings", nil)

	handler := &MentorAdminHandler{
		settingsSvc: &fakeMentorAdminSettingsService{
			getSettingsResp: service.MentorSettings{MaxCharacterSP: 5_000_000, MaxAccountAgeDays: 9},
		},
	}

	handler.GetSettings(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var resp struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data service.MentorSettings `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 200 {
		t.Fatalf("expected response code 200, got %d", resp.Code)
	}
	if resp.Data.MaxCharacterSP != 5_000_000 {
		t.Fatalf("expected max_character_sp 5000000, got %d", resp.Data.MaxCharacterSP)
	}
	if resp.Data.MaxAccountAgeDays != 9 {
		t.Fatalf("expected max_account_age_days 9, got %d", resp.Data.MaxAccountAgeDays)
	}
}

func TestMentorAdminUpdateSettingsRejectsInvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(
		http.MethodPut,
		"/api/v1/system/mentor/settings",
		bytes.NewBufferString(`{"max_character_sp":4000000,"max_account_age_days":0}`),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler := &MentorAdminHandler{settingsSvc: &fakeMentorAdminSettingsService{}}

	handler.UpdateSettings(ctx)

	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 400 {
		t.Fatalf("expected response code 400, got %d", resp.Code)
	}
}

func TestMentorAdminUpdateSettingsReturnsUpdatedSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	settingsSvc := &fakeMentorAdminSettingsService{
		updateResp: service.MentorSettings{MaxCharacterSP: 6_200_000, MaxAccountAgeDays: 11},
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(
		http.MethodPut,
		"/api/v1/system/mentor/settings",
		bytes.NewBufferString(`{"max_character_sp":6200000,"max_account_age_days":11}`),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler := &MentorAdminHandler{settingsSvc: settingsSvc}

	handler.UpdateSettings(ctx)

	if settingsSvc.updateCalls != 1 {
		t.Fatalf("expected one update call, got %d", settingsSvc.updateCalls)
	}
	if settingsSvc.lastUpdate.MaxCharacterSP != 6_200_000 {
		t.Fatalf("expected service to receive max_character_sp 6200000, got %d", settingsSvc.lastUpdate.MaxCharacterSP)
	}
	if settingsSvc.lastUpdate.MaxAccountAgeDays != 11 {
		t.Fatalf("expected service to receive max_account_age_days 11, got %d", settingsSvc.lastUpdate.MaxAccountAgeDays)
	}

	var resp struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data service.MentorSettings `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 200 {
		t.Fatalf("expected response code 200, got %d", resp.Code)
	}
	if resp.Data.MaxCharacterSP != 6_200_000 {
		t.Fatalf("expected response max_character_sp 6200000, got %d", resp.Data.MaxCharacterSP)
	}
	if resp.Data.MaxAccountAgeDays != 11 {
		t.Fatalf("expected response max_account_age_days 11, got %d", resp.Data.MaxAccountAgeDays)
	}
}

func TestMentorAdminUpdateSettingsReturnsBusinessError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(
		http.MethodPut,
		"/api/v1/system/mentor/settings",
		bytes.NewBufferString(`{"max_character_sp":6200000,"max_account_age_days":11}`),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler := &MentorAdminHandler{
		settingsSvc: &fakeMentorAdminSettingsService{updateErr: errors.New("write failed")},
	}

	handler.UpdateSettings(ctx)

	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 500 {
		t.Fatalf("expected response code 500, got %d", resp.Code)
	}
	if resp.Msg != "write failed" {
		t.Fatalf("expected business error message, got %q", resp.Msg)
	}
}
