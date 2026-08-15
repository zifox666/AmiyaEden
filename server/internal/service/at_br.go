package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/pkg/atkm"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type ATBRService struct {
	repo      *repository.ATBRRepository
	kmRepo    *repository.ATKillmailRepository
	matchRepo *repository.ATMatchRepository
	score     *ATScoreService
	sdeRepo   *repository.SdeRepository
}

func NewATBRService() *ATBRService {
	return &ATBRService{repo: repository.NewATBRRepository(), kmRepo: repository.NewATKillmailRepository(), matchRepo: repository.NewATMatchRepository(), score: NewATScoreService(), sdeRepo: repository.NewSdeRepository()}
}

func (s *ATBRService) Build(actor, roundID uint, revisionReason string, admin bool) (*repository.ATBRDetail, error) {
	round, err := s.matchRepo.GetRound(roundID)
	if err != nil {
		return nil, err
	}
	match, err := s.matchRepo.GetMatch(round.MatchID)
	if err != nil {
		return nil, err
	}
	if !admin && match.FCUserID != actor {
		return nil, errors.New("无权审核该比赛 BR")
	}
	if round.Status != "br_review" {
		return nil, errors.New("该局未处于 BR 审核状态")
	}
	nextVersion, err := s.repo.NextVersion(roundID)
	if err != nil {
		return nil, err
	}
	if nextVersion > 1 && (!admin || strings.TrimSpace(revisionReason) == "") {
		return nil, errors.New("管理员修订 BR 必须填写原因")
	}
	blocking, err := s.kmRepo.CountRoundBlocking(roundID)
	if err != nil {
		return nil, err
	}
	if blocking > 0 {
		return nil, fmt.Errorf("该局仍有 %d 份未解析或冲突 KM", blocking)
	}
	roster, err := s.matchRepo.GetRoster(roundID)
	if err != nil {
		return nil, err
	}
	if err := NewATMatchService().validateRoundRoster(match, roster, true); err != nil {
		return nil, err
	}
	kms, err := s.kmRepo.RoundKillmails(roundID)
	if err != nil {
		return nil, err
	}
	bySide := map[string][]model.ATMatchRoundRoster{"A": {}, "B": {}}
	for _, entry := range roster {
		bySide[entry.Side] = append(bySide[entry.Side], entry)
	}
	participants := make([]model.ATBRParticipantSnapshot, 0, len(roster))
	participantIndex := map[string]int{}
	cleanupIDs := map[int64]bool{}
	var cleanupList []int64
	if err := jsonUnmarshal(match.CleanupShipsSnapshot, &cleanupList); err != nil {
		return nil, err
	}
	for _, id := range cleanupList {
		cleanupIDs[id] = true
	}
	typeIDs := make([]int, 0, len(roster))
	for _, entry := range roster {
		typeIDs = append(typeIDs, int(entry.ShipTypeID))
	}
	types, err := s.sdeRepo.GetTypes(typeIDs, nil, "en")
	if err != nil {
		return nil, err
	}
	shipNames := map[int64]string{}
	for _, item := range types {
		shipNames[int64(item.TypeID)] = item.TypeName
	}
	if len(shipNames) != len(uniqueInts(typeIDs)) {
		return nil, errors.New("阵容包含未解析舰船 TypeID")
	}
	for _, side := range []string{"A", "B"} {
		ids := make([]int64, len(bySide[side]))
		for i, entry := range bySide[side] {
			ids[i] = entry.ShipTypeID
		}
		score, err := s.score.ValidateRoster(match.RuleVersionID, match.Format, ids)
		if err != nil {
			return nil, err
		}
		if !score.Valid {
			return nil, fmt.Errorf("%s 方阵容分数无效: %s", side, strings.Join(score.Errors, "; "))
		}
		for i, entry := range bySide[side] {
			participantIndex[atkm.NormalizeIdentity(entry.CharacterName)] = len(participants)
			participants = append(participants, model.ATBRParticipantSnapshot{Side: side, SlotOrder: entry.SlotOrder, CharacterName: entry.CharacterName, ShipTypeID: entry.ShipTypeID, ShipName: shipNames[entry.ShipTypeID], ShipPoints: score.Slots[i].EffectivePoints, ExcludedCleanup: cleanupIDs[entry.ShipTypeID]})
		}
	}
	teams := map[string]*model.ATBRTeamSummary{"A": {Side: "A"}, "B": {Side: "B"}}
	for _, km := range kms {
		victimIndex, ok := participantIndex[km.Killmail.VictimNameNormalized]
		if !ok {
			return nil, fmt.Errorf("KM %s 的受害者 %s 不在完整阵容中", km.Killmail.KMID, km.Killmail.VictimName)
		}
		victim := &participants[victimIndex]
		if victim.ExcludedCleanup {
			continue
		}
		victim.Dead = true
		teams[victim.Side].EffectiveDeaths++
		enemySide := "A"
		if victim.Side == "A" {
			enemySide = "B"
		}
		teams[enemySide].DestroyedPoints += victim.ShipPoints
		signatureSeen := map[int]bool{}
		for _, attacker := range km.Attackers {
			attackerIndex, exists := participantIndex[atkm.NormalizeIdentity(attacker.CharacterName)]
			if !exists {
				continue
			}
			participant := &participants[attackerIndex]
			if participant.ExcludedCleanup {
				continue
			}
			if participant.Side == victim.Side {
				participant.FriendlyDamage += attacker.DamageDone
				teams[participant.Side].FriendlyDamage += attacker.DamageDone
			} else {
				participant.HostileDamage += attacker.DamageDone
				teams[participant.Side].HostileDamage += attacker.DamageDone
				if !signatureSeen[attackerIndex] {
					participant.Signatures++
					signatureSeen[attackerIndex] = true
				}
				if attacker.FinalBlow {
					participant.FinalBlows++
				}
			}
		}
	}
	for i := range participants {
		teamDamage := teams[participants[i].Side].HostileDamage
		if teamDamage > 0 {
			participants[i].DamagePercentage = float64(participants[i].HostileDamage) * 100 / float64(teamDamage)
		}
	}
	version := &model.ATBRVersion{RoundID: roundID, Version: nextVersion, Status: "confirmed", CreatedBy: actor, RuleSnapshot: match.RuleSnapshot, CleanupSnapshot: match.CleanupShipsSnapshot, RevisionReason: revisionReason}
	teamRows := []model.ATBRTeamSummary{*teams["A"], *teams["B"]}
	if err := s.repo.Create(version, teamRows, participants); err != nil {
		return nil, err
	}
	return s.repo.Get(version.ID)
}

func jsonUnmarshal(data []byte, target any) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, target)
}
func (s *ATBRService) GetLatest(roundID uint) (*repository.ATBRDetail, error) {
	return s.repo.Latest(roundID)
}
func (s *ATBRService) Get(id uint) (*repository.ATBRDetail, error) { return s.repo.Get(id) }
func (s *ATBRService) List() ([]repository.ATBRListItem, error)    { return s.repo.ListLatest() }
func (s *ATBRService) Exists(roundID uint) (bool, error) {
	exists, err := s.repo.Exists(roundID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return exists, err
}
