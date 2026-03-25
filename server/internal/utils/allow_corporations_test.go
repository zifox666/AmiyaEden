package utils

import (
	"testing"
)

func TestInvalidateAllowCorporationsCache(t *testing.T) {
	t.Run("clears the cache", func(t *testing.T) {
		allowCorporationsCache = []int64{98185110}

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
		allowCorporationsCache = []int64{98185110, 98000001}
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

		allowCorporationsCache = []int64{98185110, 98000001, 98000002}

		if len(allowCorporationsCache) != 3 {
			t.Fatalf("expected 3 items in cache, got %d", len(allowCorporationsCache))
		}
	})

	t.Run("can set cache with single value", func(t *testing.T) {
		defer func() { allowCorporationsCache = nil }()

		allowCorporationsCache = []int64{98185110}

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
