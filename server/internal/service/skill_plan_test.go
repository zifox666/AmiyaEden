package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"strings"
	"testing"
)

func TestValidateSkillPlanPayload(t *testing.T) {
	validSkills := []SkillPlanSkillReq{
		{SkillTypeID: 3300, RequiredLevel: 5},
		{SkillTypeID: 3301, RequiredLevel: 4},
	}

	if err := validateSkillPlanPayload("Logistics Core", validSkills); err != nil {
		t.Fatalf("expected valid payload, got error: %v", err)
	}

	cases := []struct {
		name   string
		title  string
		skills []SkillPlanSkillReq
	}{
		{
			name:   "empty title",
			title:  "",
			skills: validSkills,
		},
		{
			name:  "no skills",
			title: "Logistics Core",
		},
		{
			name:  "invalid level",
			title: "Logistics Core",
			skills: []SkillPlanSkillReq{
				{SkillTypeID: 3300, RequiredLevel: 6},
			},
		},
		{
			name:  "duplicate skill",
			title: "Logistics Core",
			skills: []SkillPlanSkillReq{
				{SkillTypeID: 3300, RequiredLevel: 5},
				{SkillTypeID: 3300, RequiredLevel: 4},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "duplicate skill" {
				normalized := normalizeSkillPlanRequirements(tc.skills)
				if err := validateSkillPlanPayload(tc.title, normalized); err != nil {
					t.Fatalf("expected duplicate skills to normalize successfully, got: %v", err)
				}
				return
			}
			if err := validateSkillPlanPayload(tc.title, tc.skills); err == nil {
				t.Fatalf("expected validation error for case %q", tc.name)
			}
		})
	}
}

func TestCanManageSkillPlan(t *testing.T) {
	tests := []struct {
		name      string
		createdBy uint
		userID    uint
		userRoles []string
		expected  bool
	}{
		{
			name:      "creator can manage",
			createdBy: 9,
			userID:    9,
			userRoles: []string{model.RoleUser},
			expected:  true,
		},
		{
			name:      "admin can manage",
			createdBy: 9,
			userID:    3,
			userRoles: []string{model.RoleAdmin},
			expected:  true,
		},
		{
			name:      "fc can manage",
			createdBy: 9,
			userID:    4,
			userRoles: []string{model.RoleFC},
			expected:  true,
		},
		{
			name:      "super admin can manage",
			createdBy: 9,
			userID:    5,
			userRoles: []string{model.RoleSuperAdmin},
			expected:  true,
		},
		{
			name:      "other user cannot manage",
			createdBy: 9,
			userID:    6,
			userRoles: []string{model.RoleUser},
			expected:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := canManageSkillPlan(tc.createdBy, tc.userID, tc.userRoles); got != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestNormalizeSkillPlanRequirements(t *testing.T) {
	input := []SkillPlanSkillReq{
		{SkillTypeID: 3300, RequiredLevel: 4},
		{SkillTypeID: 3301, RequiredLevel: 3},
		{SkillTypeID: 3300, RequiredLevel: 5},
	}

	got := normalizeSkillPlanRequirements(input)
	if len(got) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(got))
	}
	if got[0].SkillTypeID != 3300 || got[0].RequiredLevel != 5 {
		t.Fatalf("expected first skill to keep highest level, got %+v", got[0])
	}
	if got[1].SkillTypeID != 3301 || got[1].RequiredLevel != 3 {
		t.Fatalf("unexpected second skill: %+v", got[1])
	}
}

func TestNormalizeOptionalSkillPlanShipTypeID(t *testing.T) {
	t.Run("nil remains nil", func(t *testing.T) {
		if got := normalizeOptionalSkillPlanShipTypeID(nil); got != nil {
			t.Fatalf("expected nil, got %+v", got)
		}
	})

	t.Run("non-positive clears selection", func(t *testing.T) {
		value := 0
		if got := normalizeOptionalSkillPlanShipTypeID(&value); got != nil {
			t.Fatalf("expected nil for zero ship type, got %+v", got)
		}
	})

	t.Run("positive value is preserved", func(t *testing.T) {
		value := 22444
		got := normalizeOptionalSkillPlanShipTypeID(&value)
		if got == nil || *got != value {
			t.Fatalf("expected %d, got %+v", value, got)
		}
	})
}

func TestParseSkillPlanLevelToken(t *testing.T) {
	cases := map[string]int{
		"1":   1,
		"II":  2,
		"iii": 3,
		"4":   4,
		"V":   5,
	}

	for input, expected := range cases {
		got, err := parseSkillPlanLevelToken(input)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", input, err)
		}
		if got != expected {
			t.Fatalf("expected %d for %q, got %d", expected, input, got)
		}
	}

	if _, err := parseSkillPlanLevelToken("VI"); err == nil {
		t.Fatal("expected invalid level error")
	}
}

func TestNormalizeSkillPlanName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "lowercases and trims",
			input: "  Logistics Cruisers  ",
			want:  "logistics cruisers",
		},
		{
			name:  "collapses internal whitespace",
			input: "Capital   Shield\tOperation",
			want:  "capital shield operation",
		},
		{
			name:  "empty stays empty",
			input: " \n\t ",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeSkillPlanName(tt.input); got != tt.want {
				t.Fatalf("normalizeSkillPlanName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCompareSkillPlanRequirements(t *testing.T) {
	plan := model.SkillPlan{ID: 7, Title: "Logistics Core"}
	skills := []model.SkillPlanSkill{
		{SkillPlanID: 7, SkillTypeID: 3300, RequiredLevel: 5},
		{SkillPlanID: 7, SkillTypeID: 3301, RequiredLevel: 4},
	}
	typeInfoMap := map[int]repository.TypeInfo{
		3300: {TypeID: 3300, TypeName: "Shield Emission Systems", GroupName: "Engineering"},
		3301: {TypeID: 3301, TypeName: "Logistics Cruisers", GroupName: "Spaceship Command"},
	}
	levelMap := map[int]int{
		3300: 5,
		3301: 2,
	}

	got := compareSkillPlanRequirements(plan, skills, typeInfoMap, levelMap)
	if got.PlanID != 7 || got.PlanTitle != "Logistics Core" {
		t.Fatalf("unexpected plan identity: %+v", got)
	}
	if got.MatchedSkills != 1 || got.TotalSkills != 2 {
		t.Fatalf("expected 1/2 matched, got %d/%d", got.MatchedSkills, got.TotalSkills)
	}
	if got.FullySatisfied {
		t.Fatal("expected plan to be incomplete")
	}
	if len(got.MissingSkills) != 1 {
		t.Fatalf("expected 1 missing skill, got %d", len(got.MissingSkills))
	}

	missing := got.MissingSkills[0]
	if missing.SkillTypeID != 3301 || missing.SkillName != "Logistics Cruisers" {
		t.Fatalf("unexpected missing skill identity: %+v", missing)
	}
	if missing.RequiredLevel != 4 || missing.CurrentLevel != 2 {
		t.Fatalf("unexpected missing skill levels: %+v", missing)
	}
}

func TestSkillPlanTextLinePattern(t *testing.T) {
	lines := []string{
		"Graviton Physics 5",
		"Tactical Logistics Reconfiguration IV",
		"Capital Shield Operation     4",
	}

	for _, line := range lines {
		matches := skillPlanTextLinePattern.FindStringSubmatch(line)
		if len(matches) != 3 {
			t.Fatalf("expected line %q to match pattern", line)
		}
		if strings.TrimSpace(matches[1]) == "" {
			t.Fatalf("expected line %q to contain a skill name", line)
		}
	}
}
