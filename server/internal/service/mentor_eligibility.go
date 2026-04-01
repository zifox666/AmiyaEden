package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"time"
)

const (
	MenteeDisqualifiedReasonAccountTooOld      = "account_too_old"
	MenteeDisqualifiedReasonSkillPointsTooHigh = "skill_points_too_high"
	MenteeDisqualifiedReasonNoCharacters       = "no_characters"
)

type MenteeEligibilityResult struct {
	IsEligible         bool   `json:"is_eligible"`
	DisqualifiedReason string `json:"disqualified_reason"`
}

type MenteeEligibilityRules struct {
	MaxCharacterSP    int64
	MaxAccountAgeDays int
}

type MenteeCharacterSnapshot struct {
	CharacterID int64
	TotalSP     int64
}

func EvaluateMenteeEligibility(accountCreatedAt time.Time, characters []MenteeCharacterSnapshot, now time.Time, rules MenteeEligibilityRules) MenteeEligibilityResult {
	if len(characters) == 0 {
		return MenteeEligibilityResult{IsEligible: false, DisqualifiedReason: MenteeDisqualifiedReasonNoCharacters}
	}
	if now.Sub(accountCreatedAt) > time.Duration(rules.MaxAccountAgeDays)*24*time.Hour {
		return MenteeEligibilityResult{IsEligible: false, DisqualifiedReason: MenteeDisqualifiedReasonAccountTooOld}
	}
	for _, character := range characters {
		if character.TotalSP >= rules.MaxCharacterSP {
			return MenteeEligibilityResult{IsEligible: false, DisqualifiedReason: MenteeDisqualifiedReasonSkillPointsTooHigh}
		}
	}
	return MenteeEligibilityResult{IsEligible: true}
}

type MentorEligibilityService struct {
	userRepo  *repository.UserRepository
	charRepo  *repository.EveCharacterRepository
	skillRepo *repository.EveSkillRepository
	cfgRepo   mentorEligibilityConfigStore
	now       func() time.Time
}

type mentorEligibilityConfigStore interface {
	GetInt64(key string, defaultVal int64) int64
	GetInt(key string, defaultVal int) int
}

func NewMentorEligibilityService() *MentorEligibilityService {
	return &MentorEligibilityService{
		userRepo:  repository.NewUserRepository(),
		charRepo:  repository.NewEveCharacterRepository(),
		skillRepo: repository.NewEveSkillRepository(),
		cfgRepo:   repository.NewSysConfigRepository(),
		now:       time.Now,
	}
}

func (s *MentorEligibilityService) GetRules() MenteeEligibilityRules {
	return MenteeEligibilityRules{
		MaxCharacterSP:    s.cfgRepo.GetInt64(model.SysConfigMenteeMaxCharacterSP, model.SysConfigDefaultMenteeMaxCharacterSP),
		MaxAccountAgeDays: s.cfgRepo.GetInt(model.SysConfigMenteeMaxAccountAgeDays, model.SysConfigDefaultMenteeMaxAccountAgeDays),
	}
}

func (s *MentorEligibilityService) EvaluateEligibility(userID uint) (*MenteeEligibilityResult, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	characters, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return nil, err
	}
	characterIDs := make([]int64, 0, len(characters))
	for _, character := range characters {
		characterIDs = append(characterIDs, character.CharacterID)
	}

	skillTotals, err := s.skillRepo.GetSkillTotalsByCharacterIDs(characterIDs)
	if err != nil {
		return nil, err
	}

	snapshots := make([]MenteeCharacterSnapshot, 0, len(characters))
	for _, character := range characters {
		snapshots = append(snapshots, MenteeCharacterSnapshot{
			CharacterID: character.CharacterID,
			TotalSP:     skillTotals[character.CharacterID],
		})
	}

	result := EvaluateMenteeEligibility(user.CreatedAt, snapshots, s.now(), s.GetRules())
	return &result, nil
}
