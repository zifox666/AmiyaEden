package handler

import (
	"bytes"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestParseOptionalNewbroDateEndOfDayCoversFullLastSecond(t *testing.T) {
	got, err := parseOptionalNewbroDate("2026-03-27", true)
	if err != nil {
		t.Fatalf("expected valid date, got error %v", err)
	}
	if got == nil {
		t.Fatal("expected parsed date")
	}

	want := time.Date(2026, 3, 27, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
	if !got.Equal(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestParseOptionalNewbroDateRejectsInvalidDate(t *testing.T) {
	got, err := parseOptionalNewbroDate("2026-13-27", false)
	if err == nil {
		t.Fatal("expected invalid date to return an error")
	}
	if got != nil {
		t.Fatalf("expected invalid date to return nil time, got %v", got)
	}
}

func TestParseOptionalUintQueryParamRejectsInvalidInput(t *testing.T) {
	if _, err := parseOptionalUintQueryParam("player_user_id", "not-a-number"); err == nil {
		t.Fatal("expected invalid player_user_id to return an error")
	}
}

func TestParseOptionalUintQueryParamRejectsOverflow(t *testing.T) {
	overflow := strconv.FormatUint(uint64(math.MaxUint32)+1, 10)
	if _, err := parseOptionalUintQueryParam("player_user_id", overflow); err == nil {
		t.Fatal("expected overflow player_user_id to return an error")
	}
}

func TestUpdateNewbroSettingsRequestAllowsZeroBonusRate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/newbro/settings", bytes.NewBufferString(`{
		"max_character_sp": 20000000,
		"multi_character_sp": 10000000,
		"multi_character_threshold": 3,
		"refresh_interval_days": 7,
		"bonus_rate": 0
	}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	var req UpdateNewbroSettingsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		t.Fatalf("expected zero bonus_rate to bind successfully, got %v", err)
	}
	if req.BonusRate == nil || *req.BonusRate != 0 {
		t.Fatalf("expected bonus_rate pointer to preserve explicit zero, got %#v", req.BonusRate)
	}
}
