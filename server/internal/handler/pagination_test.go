package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

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

func TestParsePaginationQueryUsesDefaultsForMissingAndInvalidValues(t *testing.T) {
	t.Run("missing values", func(t *testing.T) {
		ctx := newPaginationQueryTestContext("")
		page, pageSize, err := parsePaginationQuery(ctx, 20, 100)
		if err != nil {
			t.Fatalf("parsePaginationQuery() error = %v, want nil", err)
		}

		if page != 1 {
			t.Fatalf("page = %d, want 1", page)
		}
		if pageSize != 20 {
			t.Fatalf("pageSize = %d, want 20", pageSize)
		}
	})

	t.Run("invalid current returns error", func(t *testing.T) {
		ctx := newPaginationQueryTestContext("?current=bad")
		_, _, err := parsePaginationQuery(ctx, 20, 100)
		if err == nil {
			t.Fatal("expected error for invalid current query")
		}
	})

	t.Run("invalid size returns error", func(t *testing.T) {
		ctx := newPaginationQueryTestContext("?current=bad&size=oops")
		_, _, err := parsePaginationQuery(ctx, 20, 100)
		if err == nil {
			t.Fatal("expected error for invalid query")
		}
	})

	t.Run("out of range values", func(t *testing.T) {
		ctx := newPaginationQueryTestContext("?current=0&size=101")
		page, pageSize, err := parsePaginationQuery(ctx, 20, 100)
		if err != nil {
			t.Fatalf("parsePaginationQuery() error = %v, want nil", err)
		}

		if page != 1 {
			t.Fatalf("page = %d, want 1", page)
		}
		if pageSize != 20 {
			t.Fatalf("pageSize = %d, want 20", pageSize)
		}
	})
}

func TestParsePaginationQuerySupportsUnboundedSize(t *testing.T) {
	ctx := newPaginationQueryTestContext("?current=2&size=250")
	page, pageSize, err := parseUnboundedPaginationQuery(ctx, 20)
	if err != nil {
		t.Fatalf("parseUnboundedPaginationQuery() error = %v, want nil", err)
	}

	if page != 2 {
		t.Fatalf("page = %d, want 2", page)
	}
	if pageSize != 250 {
		t.Fatalf("pageSize = %d, want 250", pageSize)
	}
}

func TestParseLedgerPaginationQueryUsesDefaultAndClampsOversizedValues(t *testing.T) {
	t.Run("invalid size returns error", func(t *testing.T) {
		ctx := newPaginationQueryTestContext("?size=oops")
		_, _, err := parseLedgerPaginationQuery(ctx, 20)
		if err == nil {
			t.Fatal("expected error for invalid size query")
		}
	})

	t.Run("oversized value clamps", func(t *testing.T) {
		ctx := newPaginationQueryTestContext("?current=0&size=5000")
		page, pageSize, err := parseLedgerPaginationQuery(ctx, 20)
		if err != nil {
			t.Fatalf("parseLedgerPaginationQuery() error = %v, want nil", err)
		}

		if page != 1 {
			t.Fatalf("page = %d, want 1", page)
		}
		if pageSize != 1000 {
			t.Fatalf("pageSize = %d, want 1000", pageSize)
		}
	})
}

func newPaginationQueryTestContext(rawQuery string) *gin.Context {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest("GET", "/test"+rawQuery, nil)
	return ctx
}
