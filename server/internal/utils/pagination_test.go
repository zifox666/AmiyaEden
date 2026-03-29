package utils

import "testing"

func TestNormalizePageStartsAtOne(t *testing.T) {
	if got := NormalizePage(0); got != 1 {
		t.Fatalf("NormalizePage(0) = %d, want 1", got)
	}
	if got := NormalizePage(-5); got != 1 {
		t.Fatalf("NormalizePage(-5) = %d, want 1", got)
	}
	if got := NormalizePage(3); got != 3 {
		t.Fatalf("NormalizePage(3) = %d, want 3", got)
	}
}

func TestNormalizePageSizeUsesDefaultForOutOfRangeValues(t *testing.T) {
	if got := NormalizePageSize(0, 20, 100); got != 20 {
		t.Fatalf("NormalizePageSize(0, 20, 100) = %d, want 20", got)
	}
	if got := NormalizePageSize(200, 20, 100); got != 20 {
		t.Fatalf("NormalizePageSize(200, 20, 100) = %d, want 20", got)
	}
	if got := NormalizePageSize(50, 20, 100); got != 50 {
		t.Fatalf("NormalizePageSize(50, 20, 100) = %d, want 50", got)
	}
}

func TestNormalizeLedgerPageSizeKeepsLedgerBounds(t *testing.T) {
	if got := NormalizeLedgerPageSize(0); got != LedgerDefaultPageSize {
		t.Fatalf("NormalizeLedgerPageSize(0) = %d, want %d", got, LedgerDefaultPageSize)
	}
	if got := NormalizeLedgerPageSize(5000); got != LedgerMaxPageSize {
		t.Fatalf("NormalizeLedgerPageSize(5000) = %d, want %d", got, LedgerMaxPageSize)
	}
	if got := NormalizeLedgerPageSize(500); got != 500 {
		t.Fatalf("NormalizeLedgerPageSize(500) = %d, want 500", got)
	}
}
