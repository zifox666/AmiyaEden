package repository

import (
	"fmt"
	"strings"
	"testing"
)

func TestSDERequiredSkillFilterSQLIncludesAllPairs(t *testing.T) {
	got := sdeRequiredSkillFilterSQL()

	wantParts := make([]string, 0, len(sdeShipSkillRequirementAttributePairs))
	for _, pair := range sdeShipSkillRequirementAttributePairs {
		wantParts = append(wantParts, fmt.Sprintf("%d", pair.RequiredSkillAttributeID))
	}
	want := strings.Join(wantParts, ", ")

	if got != want {
		t.Fatalf("sdeRequiredSkillFilterSQL() = %q, want %q", got, want)
	}
}

func TestSDERequiredSkillLevelCaseSQLIncludesAllPairs(t *testing.T) {
	got := sdeRequiredSkillLevelCaseSQL(`sk."attributeID"`)

	if !strings.HasPrefix(got, `CASE sk."attributeID" `) {
		t.Fatalf("expected CASE expression to start with attribute column, got %q", got)
	}
	if !strings.HasSuffix(got, " END") {
		t.Fatalf("expected CASE expression to end with END, got %q", got)
	}

	for _, pair := range sdeShipSkillRequirementAttributePairs {
		expectedFragment := fmt.Sprintf("WHEN %d THEN %d", pair.RequiredSkillAttributeID, pair.RequiredLevelAttributeID)
		if !strings.Contains(got, expectedFragment) {
			t.Fatalf("expected CASE expression to contain %q, got %q", expectedFragment, got)
		}
	}
}

func TestSDERequiredSkillAttributePairsAreUnique(t *testing.T) {
	requiredSkillIDs := make(map[int]struct{}, len(sdeShipSkillRequirementAttributePairs))
	requiredLevelIDs := make(map[int]struct{}, len(sdeShipSkillRequirementAttributePairs))

	for _, pair := range sdeShipSkillRequirementAttributePairs {
		if _, exists := requiredSkillIDs[pair.RequiredSkillAttributeID]; exists {
			t.Fatalf("duplicate required skill attribute id %d", pair.RequiredSkillAttributeID)
		}
		requiredSkillIDs[pair.RequiredSkillAttributeID] = struct{}{}

		if _, exists := requiredLevelIDs[pair.RequiredLevelAttributeID]; exists {
			t.Fatalf("duplicate required level attribute id %d", pair.RequiredLevelAttributeID)
		}
		requiredLevelIDs[pair.RequiredLevelAttributeID] = struct{}{}
	}

	if sdeShipSkillRequirementMaxDepth <= 0 {
		t.Fatalf("expected positive recursion depth limit, got %d", sdeShipSkillRequirementMaxDepth)
	}
}
