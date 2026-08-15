package utils

import (
	"net/url"
	"testing"
)

func TestAppendRedirectQueryHashRouterWithExistingQuery(t *testing.T) {
	target, err := AppendRedirectQuery(
		"http://127.0.0.1:3007/#/auth/callback?redirect=%2F",
		url.Values{"token": {"test-token"}, "provider": {"seat"}},
	)
	if err != nil {
		t.Fatalf("AppendRedirectQuery returned error: %v", err)
	}
	const expected = "http://127.0.0.1:3007/#/auth/callback?provider=seat&redirect=%2F&token=test-token"
	if target != expected {
		t.Fatalf("unexpected target: got %q, want %q", target, expected)
	}

	parsed, err := url.Parse(target)
	if err != nil {
		t.Fatalf("parse result: %v", err)
	}
	if parsed.RawQuery != "" {
		t.Fatalf("outer query must stay empty, got %q", parsed.RawQuery)
	}
	if parsed.Fragment != "/auth/callback?provider=seat&redirect=/&token=test-token" {
		t.Fatalf("unexpected fragment: %q", parsed.Fragment)
	}
}

func TestAppendRedirectQueryRegularURL(t *testing.T) {
	target, err := AppendRedirectQuery(
		"https://example.test/auth/callback?redirect=%2Fmatches",
		url.Values{"token": {"test-token"}},
	)
	if err != nil {
		t.Fatalf("AppendRedirectQuery returned error: %v", err)
	}

	parsed, err := url.Parse(target)
	if err != nil {
		t.Fatalf("parse result: %v", err)
	}
	if parsed.Query().Get("redirect") != "/matches" || parsed.Query().Get("token") != "test-token" {
		t.Fatalf("unexpected query: %q", parsed.RawQuery)
	}
}

func TestAppendRedirectQueryHashRouterWithoutExistingQuery(t *testing.T) {
	target, err := AppendRedirectQuery(
		"https://example.test/#/auth/callback",
		url.Values{"error": {"authorization denied"}},
	)
	if err != nil {
		t.Fatalf("AppendRedirectQuery returned error: %v", err)
	}
	if target != "https://example.test/#/auth/callback?error=authorization+denied" {
		t.Fatalf("unexpected target: %q", target)
	}
}
