package service

import "testing"

func TestSysConfigBatchFormatsTypedValues(t *testing.T) {
	items := newSysConfigBatch(5).
		AddInt64("max_sp", 6_000_000, "max sp").
		AddInt("days", 7, "days").
		AddFloat64("bonus_rate", 20.5, "bonus").
		AddBool("enabled", true, "enabled").
		AddString("url", "https://example.test", "url").
		Items()

	if len(items) != 5 {
		t.Fatalf("expected 5 items, got %d", len(items))
	}

	wantValues := []string{"6000000", "7", "20.5", "true", "https://example.test"}
	for i, want := range wantValues {
		if items[i].Value != want {
			t.Fatalf("unexpected value at index %d: got %q want %q", i, items[i].Value, want)
		}
	}
}
