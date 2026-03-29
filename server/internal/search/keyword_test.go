package search

import "testing"

func TestBuildCaseInsensitiveLikePatternNormalizesWhitespaceAndCase(t *testing.T) {
	pattern, ok := BuildCaseInsensitiveLikePattern("  AmiYa  ")
	if !ok {
		t.Fatal("expected non-empty keyword to produce a LIKE pattern")
	}
	if pattern != "%amiya%" {
		t.Fatalf("expected normalized pattern %%amiya%%, got %q", pattern)
	}
}

func TestContainsKeywordMatchesAnyNormalizedValue(t *testing.T) {
	if !ContainsKeyword("  BEE ", "Captain Bee", "Other") {
		t.Fatal("expected keyword to match ignoring case and surrounding whitespace")
	}
	if ContainsKeyword("amiya", "Captain Bee", "Other") {
		t.Fatal("expected keyword to miss unrelated values")
	}
	if !ContainsKeyword("", "anything") {
		t.Fatal("expected empty keyword to match all values")
	}
}
