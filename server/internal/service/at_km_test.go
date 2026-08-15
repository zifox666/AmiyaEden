package service

import (
	"testing"
	"time"
)

func TestProspectiveWindowBoundary(t *testing.T) {
	base := time.Date(2026, 8, 14, 20, 0, 0, 0, time.UTC)
	if got := prospectiveWindow([]time.Time{base.Add(15 * time.Minute), base}); got != 15*time.Minute {
		t.Fatalf("15 minute boundary = %v", got)
	}
	if got := prospectiveWindow([]time.Time{base, base.Add(15*time.Minute + time.Second)}); got <= 15*time.Minute {
		t.Fatalf("over-boundary window = %v", got)
	}
}
