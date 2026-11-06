package esi

import (
	"amiya-eden/internal/model"
	"testing"
)

func TestExcludeCoolingDownCharacters(t *testing.T) {
	const (
		seatOnlyCooling   int64 = 1
		seatOnlyAvailable int64 = 2
		ssoWithCoolingKey int64 = 3
	)
	characters := []model.EveCharacter{
		{CharacterID: seatOnlyCooling, Scopes: "esi-a.v1"},
		{CharacterID: seatOnlyAvailable, Scopes: "esi-a.v1"},
		{CharacterID: ssoWithCoolingKey, RefreshToken: "refresh"},
	}
	isCoolingDown := func(characterID int64) bool {
		return characterID == seatOnlyCooling || characterID == ssoWithCoolingKey
	}

	got := excludeCoolingDownCharacters(characters, isCoolingDown)

	if len(got) != 2 || got[0].CharacterID != seatOnlyAvailable || got[1].CharacterID != ssoWithCoolingKey {
		t.Errorf("got %+v, 只应排除冷却中的 SeAT-only 角色", got)
	}
}
