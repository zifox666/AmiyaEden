package service

import (
	"amiya-eden/internal/repository"
)

type EveSkillService struct {
	skillRepo *repository.EveSkillRepository
	sdeRepo   *repository.SdeRepository
}

func NewEveSkillService() *EveSkillService {
	return &EveSkillService{
		skillRepo: repository.NewEveSkillRepository(),
		sdeRepo:   repository.NewSdeRepository(),
	}
}

type EveCharacterSkill struct {
	SkillID            int   `json:"skill_id"`
	ActiveLevel        int   `json:"active_level"`
	TrainedLevel       int   `json:"trained_level"`
	SkillpointsInSkill int64 `json:"skillpoints_in_skill"`
}

type Total struct {
	GroupID int64 `json:"group_id"`
	Num     int   `json:"num"`
}
type EveSkillResponse struct {
	TotalSP   int64               `json:"total_sp"`
	SkillList []EveCharacterSkill `json:"skill_list"`
	Totals    []Total             `json:"totals"`
}

func (s *EveSkillService) GetEveCharacterSkills(characterID int) (*EveSkillResponse, error) {
	result := &EveSkillResponse{}

	skill, err := s.skillRepo.GetSkill(characterID)
	if err != nil {
		return nil, err
	}
	result.TotalSP = skill.TotalSP

	list, err := s.skillRepo.GetSkillList(characterID)
	if err != nil {
		return nil, err
	}

	skillIDs := make([]int, 0, len(list))
	for _, sk := range list {
		skillIDs = append(skillIDs, sk.SkillID)
		result.SkillList = append(result.SkillList, EveCharacterSkill{
			SkillID:            sk.SkillID,
			ActiveLevel:        sk.ActiveLevel,
			TrainedLevel:       sk.TrainedLevel,
			SkillpointsInSkill: sk.SkillpointsInSkill,
		})
	}

	b := true
	typeInfos, err := s.sdeRepo.GetTypes(skillIDs, &b, "en")
	if err != nil {
		return nil, err
	}

	groups := make(map[int64]int)
	for _, t := range typeInfos {
		groups[int64(t.GroupID)]++
	}

	for gid, num := range groups {
		result.Totals = append(result.Totals, Total{GroupID: gid, Num: num})
	}

	return result, nil
}
