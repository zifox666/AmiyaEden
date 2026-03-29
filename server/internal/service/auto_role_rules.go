package service

import (
	"amiya-eden/internal/model"
	"strings"
)

func isAllowedCorporation(corporationID int64, allowCorpSet map[int64]struct{}) bool {
	if corporationID == 0 || len(allowCorpSet) == 0 {
		return false
	}
	_, ok := allowCorpSet[corporationID]
	return ok
}

func hasAllowedPrimaryCharacter(primaryCharacterID int64, chars []model.EveCharacter, allowCorpSet map[int64]struct{}) bool {
	if primaryCharacterID == 0 {
		return false
	}
	for _, char := range chars {
		if char.CharacterID == primaryCharacterID {
			return isAllowedCorporation(char.CorporationID, allowCorpSet)
		}
	}
	return false
}

func hasAnyAllowedCharacter(chars []model.EveCharacter, allowCorpSet map[int64]struct{}) bool {
	for _, char := range chars {
		if isAllowedCorporation(char.CorporationID, allowCorpSet) {
			return true
		}
	}
	return false
}

func shouldAutoPromoteGuestToUser(currentCodes []string, chars []model.EveCharacter, allowCorpSet map[int64]struct{}) bool {
	return !model.HasNonGuestRole(currentCodes) && hasAnyAllowedCharacter(chars, allowCorpSet)
}

func isDirectorCorpRole(name string) bool {
	return strings.EqualFold(strings.TrimSpace(name), "Director")
}

func shouldAutoAssignAdminFromDirector(corporationID int64, corpRole string) bool {
	return corporationID == model.SystemCorporationID && isDirectorCorpRole(corpRole)
}
