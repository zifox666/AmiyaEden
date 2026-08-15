package service

import (
	"amiya-eden/internal/model"
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateQuestionnaireGeneratesInternalKeys(t *testing.T) {
	input := RecruitmentQuestionnaireInput{
		Title: "AT application",
		Fields: []model.RecruitmentField{
			{Type: "text", Label: "Pilot name"},
			{Type: "radio", Label: "Availability", Description: "Pick one", Options: json.RawMessage(`["yes","no"]`)},
		},
	}

	if err := validateQuestionnaire(input); err != nil {
		t.Fatalf("validateQuestionnaire() error = %v", err)
	}
	if input.Fields[0].FieldKey == "" || input.Fields[1].FieldKey == "" {
		t.Fatal("internal field keys were not generated")
	}
	if input.Fields[0].FieldKey == input.Fields[1].FieldKey {
		t.Fatal("generated field keys must be unique")
	}
	if !strings.HasPrefix(input.Fields[0].FieldKey, "question_") {
		t.Fatalf("unexpected generated key %q", input.Fields[0].FieldKey)
	}
}

func TestValidateQuestionnaireRejectsChoiceWithoutOptions(t *testing.T) {
	input := RecruitmentQuestionnaireInput{
		Title:  "AT application",
		Fields: []model.RecruitmentField{{Type: "checkbox", Label: "Ships", Options: json.RawMessage(`[]`)}},
	}
	if err := validateQuestionnaire(input); err == nil {
		t.Fatal("expected an error for a choice question without options")
	}
}

func TestATGuestMenuIncludesCharacterInformation(t *testing.T) {
	menus := model.ATDefaultRoleMenuMap()[model.RoleGuest]
	want := []string{"UserCenter", "ATEveInfo", "ATEveSkills", "ATEveImplants", "ATEveFittings"}
	for _, name := range want {
		found := false
		for _, menu := range menus {
			if menu == name {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("guest menu is missing %s: %v", name, menus)
		}
	}
}
