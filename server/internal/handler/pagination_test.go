package handler

import "testing"

func TestNormalizeLedgerPaginationRejectsOversizedPageSize(t *testing.T) {
	page, pageSize := normalizeLedgerPagination(0, 5000)

	if page != 1 {
		t.Fatalf("page = %d, want 1", page)
	}
	if pageSize != 1000 {
		t.Fatalf("pageSize = %d, want 1000", pageSize)
	}
}

func TestNormalizeFleetMembersPaginationUsesFixedPageSize(t *testing.T) {
	page, pageSize := normalizePagination(0, 500, 260, 260)

	if page != 1 {
		t.Fatalf("page = %d, want 1", page)
	}
	if pageSize != 260 {
		t.Fatalf("pageSize = %d, want 260", pageSize)
	}
}

func TestNormalizeStandardPaginationRejectsOversizedPageSize(t *testing.T) {
	page, pageSize := normalizePagination(-3, 101, 20, 100)

	if page != 1 {
		t.Fatalf("page = %d, want 1", page)
	}
	if pageSize != 20 {
		t.Fatalf("pageSize = %d, want 20", pageSize)
	}
}
