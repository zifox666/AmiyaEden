package search

import "strings"

// NormalizeKeyword trims surrounding whitespace and lowercases the keyword so
// both SQL-backed and in-memory fuzzy searches share the same normalization.
func NormalizeKeyword(keyword string) string {
	return strings.ToLower(strings.TrimSpace(keyword))
}

// BuildCaseInsensitiveLikePattern returns a normalized LIKE pattern for fuzzy
// searches, or false when the keyword is empty after normalization.
func BuildCaseInsensitiveLikePattern(keyword string) (string, bool) {
	normalized := NormalizeKeyword(keyword)
	if normalized == "" {
		return "", false
	}
	return "%" + normalized + "%", true
}

// ContainsKeyword applies the same normalization used by SQL filters to a set
// of in-memory values, returning true when any value contains the keyword.
func ContainsKeyword(keyword string, values ...string) bool {
	normalized := NormalizeKeyword(keyword)
	if normalized == "" {
		return true
	}

	for _, value := range values {
		if strings.Contains(strings.ToLower(strings.TrimSpace(value)), normalized) {
			return true
		}
	}

	return false
}
