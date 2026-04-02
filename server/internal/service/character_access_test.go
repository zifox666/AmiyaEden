package service

import (
	"amiya-eden/internal/model"
	"errors"
	"testing"
)

type stubOwnedCharacterRepo struct {
	chars []model.EveCharacter
	err   error
}

func (s stubOwnedCharacterRepo) ListByUserID(_ uint) ([]model.EveCharacter, error) {
	if s.err != nil {
		return nil, s.err
	}
	return append([]model.EveCharacter(nil), s.chars...), nil
}

func TestListOwnedCharactersWrapsRepositoryError(t *testing.T) {
	_, err := listOwnedCharacters(stubOwnedCharacterRepo{err: errors.New("boom")}, 42)
	if err == nil {
		t.Fatal("expected wrapped ownership list error")
	}
	if err.Error() != "获取人物列表失败" {
		t.Fatalf("expected wrapped error message, got %q", err.Error())
	}
}

func TestFindOwnedCharacterReturnsMatchingRecord(t *testing.T) {
	repo := stubOwnedCharacterRepo{chars: []model.EveCharacter{{CharacterID: 1001}, {CharacterID: 1002}}}

	character, err := findOwnedCharacter(repo, 7, 1002)
	if err != nil {
		t.Fatalf("expected owned character lookup to succeed, got %v", err)
	}
	if character == nil {
		t.Fatal("expected owned character record")
	}
	if character.CharacterID != 1002 {
		t.Fatalf("expected character id 1002, got %d", character.CharacterID)
	}
}

func TestRequireOwnedCharacterReturnsOwnershipErrorWhenMissing(t *testing.T) {
	err := requireOwnedCharacter(stubOwnedCharacterRepo{chars: []model.EveCharacter{{CharacterID: 1001}}}, 7, 1002)
	if err == nil {
		t.Fatal("expected ownership error")
	}
	if err.Error() != "该人物不属于当前用户" {
		t.Fatalf("expected ownership error message, got %q", err.Error())
	}
}
