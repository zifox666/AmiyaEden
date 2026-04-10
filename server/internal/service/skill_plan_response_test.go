package service

import (
	"encoding/json"
	"testing"
)

func TestSkillPlanCheckCharacterResponseOmitsPortraitURL(t *testing.T) {
	payload, err := json.Marshal(SkillPlanCheckCharacterResp{
		CharacterID:   9001,
		CharacterName: "Amiya Prime",
	})
	if err != nil {
		t.Fatalf("marshal skill plan character response: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatalf("unmarshal skill plan character response: %v", err)
	}

	if _, exists := raw["portrait_url"]; exists {
		t.Fatalf("expected skill plan character response to omit portrait_url, got %#v", raw["portrait_url"])
	}
}
