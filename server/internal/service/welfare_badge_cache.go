package service

import (
	"amiya-eden/internal/model"
	"sync"
	"time"
)

const eligibleWelfareBadgeCacheTTL = 24 * time.Hour

type eligibleWelfareBadgeCacheEntry struct {
	count     int64
	expiresAt time.Time
}

var eligibleWelfareBadgeCache = struct {
	mu     sync.RWMutex
	counts map[uint]eligibleWelfareBadgeCacheEntry
}{
	counts: make(map[uint]eligibleWelfareBadgeCacheEntry),
}

func cacheEligibleWelfareBadgeCount(userID uint, eligibleWelfares []EligibleWelfareResp) {
	setEligibleWelfareBadgeCount(userID, countEligibleWelfareBadgeEntries(eligibleWelfares))
}

func setEligibleWelfareBadgeCount(userID uint, count int64) {
	eligibleWelfareBadgeCache.mu.Lock()
	defer eligibleWelfareBadgeCache.mu.Unlock()
	pruneEligibleWelfareBadgeCacheExpiredLocked(time.Now())

	if count <= 0 {
		delete(eligibleWelfareBadgeCache.counts, userID)
		return
	}
	eligibleWelfareBadgeCache.counts[userID] = eligibleWelfareBadgeCacheEntry{
		count:     count,
		expiresAt: time.Now().Add(eligibleWelfareBadgeCacheTTL),
	}
}

func getCachedEligibleWelfareBadgeCount(userID uint) int64 {
	count, _ := getCachedEligibleWelfareBadgeCountEntry(userID)
	return count
}

func getCachedEligibleWelfareBadgeCountEntry(userID uint) (int64, bool) {
	eligibleWelfareBadgeCache.mu.RLock()
	entry, ok := eligibleWelfareBadgeCache.counts[userID]
	eligibleWelfareBadgeCache.mu.RUnlock()
	if !ok {
		return 0, false
	}

	if time.Now().Before(entry.expiresAt) {
		return entry.count, true
	}

	eligibleWelfareBadgeCache.mu.Lock()
	defer eligibleWelfareBadgeCache.mu.Unlock()
	if current, ok := eligibleWelfareBadgeCache.counts[userID]; ok && !time.Now().Before(current.expiresAt) {
		delete(eligibleWelfareBadgeCache.counts, userID)
	}
	return 0, false
}

func updateEligibleWelfareBadgeCountAfterApply(userID uint, contributionAfterApply int64) {
	eligibleWelfareBadgeCache.mu.Lock()
	defer eligibleWelfareBadgeCache.mu.Unlock()
	now := time.Now()
	pruneEligibleWelfareBadgeCacheExpiredLocked(now)

	entry, ok := eligibleWelfareBadgeCache.counts[userID]
	if !ok {
		return
	}
	updatedCount := entry.count - 1 + contributionAfterApply
	if updatedCount <= 0 {
		delete(eligibleWelfareBadgeCache.counts, userID)
		return
	}
	eligibleWelfareBadgeCache.counts[userID] = eligibleWelfareBadgeCacheEntry{
		count:     updatedCount,
		expiresAt: now.Add(eligibleWelfareBadgeCacheTTL),
	}
}

func pruneEligibleWelfareBadgeCacheExpiredLocked(now time.Time) {
	for userID, entry := range eligibleWelfareBadgeCache.counts {
		if !now.Before(entry.expiresAt) {
			delete(eligibleWelfareBadgeCache.counts, userID)
		}
	}
}

func countEligibleWelfareBadgeEntries(eligibleWelfares []EligibleWelfareResp) int64 {
	var count int64
	for _, welfare := range eligibleWelfares {
		if welfare.DistMode == model.WelfareDistModePerCharacter {
			for _, character := range welfare.EligibleCharacters {
				if character.CanApplyNow {
					count++
					break
				}
			}
			continue
		}

		if welfare.CanApplyNow {
			count++
		}
	}

	return count
}
