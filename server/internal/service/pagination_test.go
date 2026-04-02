package service

import "testing"

func TestNormalizePageRequestAppliesStandardBounds(t *testing.T) {
	page := 0
	pageSize := 500

	normalizePageRequest(&page, &pageSize, 20, 100)

	if page != 1 {
		t.Fatalf("page = %d, want 1", page)
	}
	if pageSize != 20 {
		t.Fatalf("pageSize = %d, want 20", pageSize)
	}
}

func TestNormalizeLedgerPageRequestAppliesLedgerBounds(t *testing.T) {
	page := -5
	pageSize := 5000

	normalizeLedgerPageRequest(&page, &pageSize)

	if page != 1 {
		t.Fatalf("page = %d, want 1", page)
	}
	if pageSize != 1000 {
		t.Fatalf("pageSize = %d, want 1000", pageSize)
	}
}
