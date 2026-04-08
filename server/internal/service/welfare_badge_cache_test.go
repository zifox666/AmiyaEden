package service

import (
	"testing"
	"time"
)

func TestGetCachedEligibleWelfareBadgeCountExpiresAndDeletesEntries(t *testing.T) {
	eligibleWelfareBadgeCache.mu.Lock()
	original := eligibleWelfareBadgeCache.counts
	eligibleWelfareBadgeCache.counts = map[uint]eligibleWelfareBadgeCacheEntry{
		1: {
			count:     3,
			expiresAt: time.Now().Add(-time.Minute),
		},
	}
	eligibleWelfareBadgeCache.mu.Unlock()
	t.Cleanup(func() {
		eligibleWelfareBadgeCache.mu.Lock()
		eligibleWelfareBadgeCache.counts = original
		eligibleWelfareBadgeCache.mu.Unlock()
	})

	if got := getCachedEligibleWelfareBadgeCount(1); got != 0 {
		t.Fatalf("expected expired badge cache entry to be treated as empty, got %d", got)
	}

	eligibleWelfareBadgeCache.mu.RLock()
	_, exists := eligibleWelfareBadgeCache.counts[1]
	eligibleWelfareBadgeCache.mu.RUnlock()
	if exists {
		t.Fatal("expected expired badge cache entry to be deleted")
	}
}

func TestCacheEligibleWelfareBadgeCountPrunesExpiredEntriesOnWrite(t *testing.T) {
	eligibleWelfareBadgeCache.mu.Lock()
	original := eligibleWelfareBadgeCache.counts
	eligibleWelfareBadgeCache.counts = map[uint]eligibleWelfareBadgeCacheEntry{
		1: {
			count:     2,
			expiresAt: time.Now().Add(-time.Minute),
		},
	}
	eligibleWelfareBadgeCache.mu.Unlock()
	t.Cleanup(func() {
		eligibleWelfareBadgeCache.mu.Lock()
		eligibleWelfareBadgeCache.counts = original
		eligibleWelfareBadgeCache.mu.Unlock()
	})

	cacheEligibleWelfareBadgeCount(2, []EligibleWelfareResp{{CanApplyNow: true}})

	eligibleWelfareBadgeCache.mu.RLock()
	_, expiredExists := eligibleWelfareBadgeCache.counts[1]
	entry, freshExists := eligibleWelfareBadgeCache.counts[2]
	eligibleWelfareBadgeCache.mu.RUnlock()

	if expiredExists {
		t.Fatal("expected write path to prune expired badge cache entries")
	}
	if !freshExists || entry.count != 1 {
		t.Fatalf("expected fresh badge cache entry for user 2, got %+v", entry)
	}
}

func TestUpdateEligibleWelfareBadgeCountAfterApplyDoesNothingWithoutWarmCache(t *testing.T) {
	eligibleWelfareBadgeCache.mu.Lock()
	original := eligibleWelfareBadgeCache.counts
	eligibleWelfareBadgeCache.counts = map[uint]eligibleWelfareBadgeCacheEntry{}
	eligibleWelfareBadgeCache.mu.Unlock()
	t.Cleanup(func() {
		eligibleWelfareBadgeCache.mu.Lock()
		eligibleWelfareBadgeCache.counts = original
		eligibleWelfareBadgeCache.mu.Unlock()
	})

	updateEligibleWelfareBadgeCountAfterApply(3, 1)

	eligibleWelfareBadgeCache.mu.RLock()
	_, exists := eligibleWelfareBadgeCache.counts[3]
	eligibleWelfareBadgeCache.mu.RUnlock()
	if exists {
		t.Fatal("expected cold badge cache to remain untouched")
	}
}

func TestUpdateEligibleWelfareBadgeCountAfterApplyAdjustsWarmCache(t *testing.T) {
	eligibleWelfareBadgeCache.mu.Lock()
	original := eligibleWelfareBadgeCache.counts
	eligibleWelfareBadgeCache.counts = map[uint]eligibleWelfareBadgeCacheEntry{
		4: {
			count:     2,
			expiresAt: time.Now().Add(time.Hour),
		},
	}
	eligibleWelfareBadgeCache.mu.Unlock()
	t.Cleanup(func() {
		eligibleWelfareBadgeCache.mu.Lock()
		eligibleWelfareBadgeCache.counts = original
		eligibleWelfareBadgeCache.mu.Unlock()
	})

	updateEligibleWelfareBadgeCountAfterApply(4, 0)

	eligibleWelfareBadgeCache.mu.RLock()
	entry := eligibleWelfareBadgeCache.counts[4]
	eligibleWelfareBadgeCache.mu.RUnlock()
	if entry.count != 1 {
		t.Fatalf("expected warm badge cache to decrement to 1, got %+v", entry)
	}
}
