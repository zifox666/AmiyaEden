package service

import (
	"amiya-eden/internal/model"
	"errors"
)

var (
	errUserCharacterListFailed = errors.New("获取人物列表失败")
	errCharacterNotOwned       = errors.New("该人物不属于当前用户")
)

type userCharacterLister interface {
	ListByUserID(userID uint) ([]model.EveCharacter, error)
}

func findCharacterByID(characters []model.EveCharacter, characterID int64) *model.EveCharacter {
	for i := range characters {
		if characters[i].CharacterID == characterID {
			return &characters[i]
		}
	}
	return nil
}

func listOwnedCharacters(repo userCharacterLister, userID uint) ([]model.EveCharacter, error) {
	characters, err := repo.ListByUserID(userID)
	if err != nil {
		return nil, errUserCharacterListFailed
	}
	return characters, nil
}

func findOwnedCharacter(repo userCharacterLister, userID uint, characterID int64) (*model.EveCharacter, error) {
	characters, err := listOwnedCharacters(repo, userID)
	if err != nil {
		return nil, err
	}
	if character := findCharacterByID(characters, characterID); character != nil {
		return character, nil
	}
	return nil, errCharacterNotOwned
}

func requireOwnedCharacter(repo userCharacterLister, userID uint, characterID int64) error {
	_, err := findOwnedCharacter(repo, userID, characterID)
	return err
}
