package utils

import (
	"amiya-eden/internal/model"
	"testing"
)

func TestInvalidateAllowCorporationsCache(t *testing.T) {
	t.Run("clears the cache", func(t *testing.T) {
		allowCorporationsCache = []int64{model.SystemCorporationID}

		InvalidateAllowCorporationsCache()

		if allowCorporationsCache != nil {
			t.Fatal("expected cache to be nil after invalidation")
		}
	})

	t.Run("does not panic when cache is already nil", func(t *testing.T) {
		allowCorporationsCache = nil

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("unexpected panic: %v", r)
			}
		}()

		InvalidateAllowCorporationsCache()
	})

	t.Run("clears cache multiple times", func(t *testing.T) {
		allowCorporationsCache = []int64{model.SystemCorporationID, 98000001}
		InvalidateAllowCorporationsCache()

		if allowCorporationsCache != nil {
			t.Fatal("expected cache to be nil after first invalidation")
		}

		InvalidateAllowCorporationsCache()

		if allowCorporationsCache != nil {
			t.Fatal("expected cache to be nil after second invalidation")
		}
	})
}

func TestAllowCorporationsCache(t *testing.T) {
	t.Run("can set cache with multiple values", func(t *testing.T) {
		defer func() { allowCorporationsCache = nil }()

		allowCorporationsCache = []int64{model.SystemCorporationID, 98000001, 98000002}

		if len(allowCorporationsCache) != 3 {
			t.Fatalf("expected 3 items in cache, got %d", len(allowCorporationsCache))
		}
	})

	t.Run("can set cache with single value", func(t *testing.T) {
		defer func() { allowCorporationsCache = nil }()

		allowCorporationsCache = []int64{model.SystemCorporationID}

		if len(allowCorporationsCache) != 1 {
			t.Fatalf("expected 1 item in cache, got %d", len(allowCorporationsCache))
		}
	})

	t.Run("can set cache with empty slice", func(t *testing.T) {
		defer func() { allowCorporationsCache = nil }()

		allowCorporationsCache = []int64{}

		if len(allowCorporationsCache) != 0 {
			t.Fatalf("expected 0 items in cache, got %d", len(allowCorporationsCache))
		}
	})
}

func TestNormalizeAllowCorporations(t *testing.T) {
	t.Run("always includes required corporation id first", func(t *testing.T) {
		got := NormalizeAllowCorporations([]int64{98000001, 98000002})
		want := []int64{model.SystemCorporationID, 98000001, 98000002}

		if len(got) != len(want) {
			t.Fatalf("expected %d values, got %d", len(want), len(got))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("expected value %d at index %d, got %d", want[i], i, got[i])
			}
		}
	})

	t.Run("deduplicates the required corporation id", func(t *testing.T) {
		got := NormalizeAllowCorporations([]int64{model.SystemCorporationID, 98000001, model.SystemCorporationID})
		want := []int64{model.SystemCorporationID, 98000001}

		if len(got) != len(want) {
			t.Fatalf("expected %d values, got %d", len(want), len(got))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("expected value %d at index %d, got %d", want[i], i, got[i])
			}
		}
	})

	t.Run("drops non positive corporation ids", func(t *testing.T) {
		got := NormalizeAllowCorporations([]int64{0, -1, 98000001})
		want := []int64{model.SystemCorporationID, 98000001}

		if len(got) != len(want) {
			t.Fatalf("expected %d values, got %d", len(want), len(got))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("expected value %d at index %d, got %d", want[i], i, got[i])
			}
		}
	})
}

func TestValidateAllowCorporations(t *testing.T) {
	t.Run("accepts positive corporation ids", func(t *testing.T) {
		if err := ValidateAllowCorporations([]int64{model.SystemCorporationID, 98000001}); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("rejects non positive corporation ids", func(t *testing.T) {
		if err := ValidateAllowCorporations([]int64{model.SystemCorporationID, 0}); err == nil {
			t.Fatal("expected validation error for non positive corporation id")
		}
	})
}

func TestGetAllowCorporationsReturnsCopy(t *testing.T) {
	defer func() { allowCorporationsCache = nil }()

	allowCorporationsCache = []int64{model.SystemCorporationID, 98000001}
	got := GetAllowCorporations()
	got[0] = 98000099

	if allowCorporationsCache[0] != model.SystemCorporationID {
		t.Fatalf("expected cache to remain %d, got %d", model.SystemCorporationID, allowCorporationsCache[0])
	}
}
