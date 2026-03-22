package service

import "testing"

func TestParseReason(t *testing.T) {
	result := parseReason("123: 2, 456:3, 123:4, invalid, 789: x, :1")

	if result[123] != 6 {
		t.Fatalf("expected npc 123 total 6, got %d", result[123])
	}
	if result[456] != 3 {
		t.Fatalf("expected npc 456 total 3, got %d", result[456])
	}
	if _, ok := result[789]; ok {
		t.Fatalf("did not expect invalid npc 789 entry to be present")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 valid npc entries, got %d", len(result))
	}
}

func TestParseDateRange(t *testing.T) {
	start, end := parseDateRange("2026-03-01", "2026-03-31")
	if start == nil || start.Format("2006-01-02 15:04:05") != "2026-03-01 00:00:00" {
		t.Fatalf("unexpected start: %v", start)
	}
	if end == nil || end.Format("2006-01-02 15:04:05") != "2026-03-31 23:59:59" {
		t.Fatalf("unexpected end: %v", end)
	}

	start, end = parseDateRange("bad", "")
	if start != nil {
		t.Fatalf("expected nil start for invalid input, got %v", start)
	}
	if end != nil {
		t.Fatalf("expected nil end for empty input, got %v", end)
	}
}
