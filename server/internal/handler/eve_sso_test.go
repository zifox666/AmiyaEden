package handler

import (
	"amiya-eden/pkg/response"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSplitCSV(t *testing.T) {
	input := " esi-location.read_location.v1,esi-ui.open_window.v1; esi-wallet.read_character_wallet.v1  publicData "
	want := []string{
		"esi-location.read_location.v1",
		"esi-ui.open_window.v1",
		"esi-wallet.read_character_wallet.v1",
		"publicData",
	}

	got := splitCSV(input)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("splitCSV(%q) = %v, want %v", input, got, want)
	}
}

func TestSplitAny(t *testing.T) {
	tests := []struct {
		name  string
		input string
		seps  string
		want  []string
	}{
		{
			name:  "repeated separators are ignored",
			input: ",,alpha; beta  gamma;;",
			seps:  ",; ",
			want:  []string{"alpha", "beta", "gamma"},
		},
		{
			name:  "leading and trailing separators are ignored",
			input: "  one two ",
			seps:  " ",
			want:  []string{"one", "two"},
		},
		{
			name:  "string without separators stays whole",
			input: "publicData",
			seps:  ",; ",
			want:  []string{"publicData"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitAny(tt.input, tt.seps)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("splitAny(%q, %q) = %v, want %v", tt.input, tt.seps, got, tt.want)
			}
		})
	}
}

func TestLoginRejectsAdminOnlyExtraScopesForPublicRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(
		http.MethodGet,
		"/api/v1/sso/eve/login?scopes=esi-killmails.read_corporation_killmails.v1",
		nil,
	)

	(&EveSSOHandler{}).Login(ctx)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for public corp killmail scope request, got %d", rec.Code)
	}
	var resp response.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code != response.CodeForbidden {
		t.Fatalf("expected response code %d, got %d", response.CodeForbidden, resp.Code)
	}
}

func TestBindLoginRejectsAdminOnlyExtraScopesForNonAdminUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Set("userID", uint(42))
	ctx.Set("roles", []string{"user"})
	ctx.Request = httptest.NewRequest(
		http.MethodGet,
		"/api/v1/sso/eve/bind?scopes=esi-killmails.read_corporation_killmails.v1",
		nil,
	)

	(&EveSSOHandler{}).BindLogin(ctx)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-admin corp killmail scope request, got %d", rec.Code)
	}
	var resp response.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code != response.CodeForbidden {
		t.Fatalf("expected response code %d, got %d", response.CodeForbidden, resp.Code)
	}
}
