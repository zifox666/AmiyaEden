package esi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPostJSONWithLimitRejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST request, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, strings.Repeat("x", 6))
	}))
	t.Cleanup(server.Close)

	client := NewClientWithConfig(server.URL, "")

	var dest []map[string]any
	err := client.PostJSONWithLimit(context.Background(), "/characters/affiliation/", "", []int64{90000001}, &dest, 5)
	if err == nil {
		t.Fatal("expected oversized response error")
	}
	if !strings.Contains(err.Error(), "response exceeds 5 bytes") {
		t.Fatalf("expected oversize error, got %v", err)
	}
}
