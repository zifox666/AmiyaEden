package service

import (
	"amiya-eden/internal/model"
	"encoding/json"
	"testing"
	"time"
)

func TestShouldBlockSelfMentorApplication(t *testing.T) {
	if !shouldBlockSelfMentorApplication(42, 42) {
		t.Fatal("expected self-application to be blocked")
	}
	if shouldBlockSelfMentorApplication(42, 84) {
		t.Fatal("did not expect different mentor to be blocked")
	}
}

func TestFilterMentorCandidateUsersExcludesCurrentUser(t *testing.T) {
	users := []model.User{
		{BaseModel: model.BaseModel{ID: 7}},
		{BaseModel: model.BaseModel{ID: 42}},
		{BaseModel: model.BaseModel{ID: 84}},
	}

	filtered := filterMentorCandidateUsers(42, users)
	if len(filtered) != 2 {
		t.Fatalf("expected 2 users after filtering, got %d", len(filtered))
	}
	if filtered[0].ID != 7 || filtered[1].ID != 84 {
		t.Fatalf("unexpected filtered users: %+v", filtered)
	}
}

func TestCalculateMentorDaysActive(t *testing.T) {
	createdAt := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	lastLogin := createdAt.Add(36 * time.Hour)

	if got := calculateMentorDaysActive(createdAt, &lastLogin); got != 1 {
		t.Fatalf("expected 1 active day, got %d", got)
	}
	if got := calculateMentorDaysActive(createdAt, nil); got != 0 {
		t.Fatalf("expected nil last login to produce 0 days, got %d", got)
	}
	beforeCreate := createdAt.Add(-24 * time.Hour)
	if got := calculateMentorDaysActive(createdAt, &beforeCreate); got != 0 {
		t.Fatalf("expected negative duration to clamp to 0, got %d", got)
	}
}

func TestBuildMentorCandidateIncludesContact(t *testing.T) {
	user := model.User{
		BaseModel: model.BaseModel{ID: 99},
		Nickname:  "Teacher",
		QQ:        "123456",
		DiscordID: "teacher#0001",
	}
	primaryChar := model.EveCharacter{
		CharacterID:   777,
		CharacterName: "Helpful Mentor",
	}

	got := buildMentorCandidate(user, primaryChar, 3)

	if got.MentorQQ != "123456" {
		t.Fatalf("expected mentor QQ to be preserved, got %q", got.MentorQQ)
	}
	if got.MentorDiscordID != "teacher#0001" {
		t.Fatalf("expected mentor Discord ID to be preserved, got %q", got.MentorDiscordID)
	}
	if got.ActiveMenteeCount != 3 {
		t.Fatalf("expected active mentee count to be preserved, got %d", got.ActiveMenteeCount)
	}
	payload, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal mentor candidate: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatalf("unmarshal mentor candidate: %v", err)
	}
	if _, exists := raw["mentor_portrait_url"]; exists {
		t.Fatalf("expected mentor candidate to omit mentor_portrait_url, got %#v", raw["mentor_portrait_url"])
	}
}

func TestBuildMentorRelationshipViewIncludesMentorContact(t *testing.T) {
	rel := model.MentorMenteeRelationship{
		BaseModel:    model.BaseModel{ID: 15},
		MenteeUserID: 7,
		MentorUserID: 9,
		Status:       model.MentorRelationStatusActive,
	}
	appliedAt := time.Date(2026, time.April, 2, 8, 0, 0, 0, time.UTC)
	respondedAt := appliedAt.Add(2 * time.Hour)
	rel.AppliedAt = appliedAt
	rel.RespondedAt = &respondedAt

	mentorUser := model.User{
		BaseModel: model.BaseModel{ID: 9},
		Nickname:  "Teacher",
		QQ:        "123456",
		DiscordID: "teacher#0001",
	}
	menteeUser := model.User{
		BaseModel: model.BaseModel{ID: 7},
		Nickname:  "Student",
	}
	mentorChar := model.EveCharacter{
		CharacterID:   91,
		CharacterName: "Helpful Mentor",
	}
	menteeChar := model.EveCharacter{
		CharacterID:   71,
		CharacterName: "Curious Mentee",
	}

	got := buildMentorRelationshipView(rel, mentorUser, menteeUser, mentorChar, menteeChar)

	if got.MentorQQ != "123456" {
		t.Fatalf("expected mentor QQ to be preserved, got %q", got.MentorQQ)
	}
	if got.MentorDiscordID != "teacher#0001" {
		t.Fatalf("expected mentor Discord ID to be preserved, got %q", got.MentorDiscordID)
	}
	if got.MentorCharacterName != "Helpful Mentor" {
		t.Fatalf("expected mentor name to be preserved, got %q", got.MentorCharacterName)
	}
	payload, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal mentor relationship view: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatalf("unmarshal mentor relationship view: %v", err)
	}
	if _, exists := raw["mentor_portrait_url"]; exists {
		t.Fatalf("expected relationship view to omit mentor_portrait_url, got %#v", raw["mentor_portrait_url"])
	}
	if _, exists := raw["mentee_portrait_url"]; exists {
		t.Fatalf("expected relationship view to omit mentee_portrait_url, got %#v", raw["mentee_portrait_url"])
	}
}
