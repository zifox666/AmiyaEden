package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/pkg/atkm"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ATKillmailService struct {
	repo      *repository.ATKillmailRepository
	matchRepo *repository.ATMatchRepository
	sdeRepo   *repository.SdeRepository
}

func NewATKillmailService() *ATKillmailService {
	return &ATKillmailService{repo: repository.NewATKillmailRepository(), matchRepo: repository.NewATMatchRepository(), sdeRepo: repository.NewSdeRepository()}
}

type ATKillmailSubmitItem struct {
	KMID             string   `json:"km_id"`
	Duplicate        bool     `json:"duplicate"`
	ResolutionStatus string   `json:"resolution_status"`
	UnresolvedNames  []string `json:"unresolved_names"`
}
type ATKillmailSubmitResult struct {
	Items []ATKillmailSubmitItem `json:"items"`
}

func (s *ATKillmailService) Submit(userID uint, text string) (*ATKillmailSubmitResult, error) {
	if len(text) > 2*1024*1024 {
		return nil, errors.New("单次 KM 文本不能超过 2 MiB")
	}
	parsed, err := atkm.ParseMany(text)
	if err != nil {
		return nil, err
	}
	if len(parsed) > 100 {
		return nil, errors.New("单次最多提交 100 份 KM")
	}
	names := map[string]bool{}
	for _, km := range parsed {
		names[km.Victim.ShipName] = true
		for _, attacker := range km.Attackers {
			if attacker.ShipName != "" {
				names[attacker.ShipName] = true
			}
			if attacker.WeaponName != "" {
				names[attacker.WeaponName] = true
			}
		}
	}
	nameList := make([]string, 0, len(names))
	for name := range names {
		nameList = append(nameList, name)
	}
	typeIDs, err := s.sdeRepo.GetTypeIDsByNames(nameList)
	if err != nil {
		return nil, err
	}
	result := &ATKillmailSubmitResult{}
	for _, km := range parsed {
		hashBytes := sha256.Sum256([]byte(km.NormalizedText))
		hash := hex.EncodeToString(hashBytes[:])
		existing, lookupErr := s.repo.GetByHash(hash)
		if lookupErr == nil {
			result.Items = append(result.Items, ATKillmailSubmitItem{KMID: existing.KMID, Duplicate: true, ResolutionStatus: existing.ResolutionStatus})
			continue
		}
		if !errors.Is(lookupErr, gorm.ErrRecordNotFound) {
			return nil, lookupErr
		}
		id, idErr := uuid.NewV7()
		if idErr != nil {
			return nil, idErr
		}
		unresolved := []string{}
		victimTypeID := typeIDPointer(typeIDs[km.Victim.ShipName])
		if victimTypeID == nil {
			unresolved = append(unresolved, km.Victim.ShipName)
		}
		item := &model.ATKillmail{KMID: id.String(), SubmittedBy: userID, RawText: km.RawText, NormalizedText: km.NormalizedText, ContentHash: hash, KillmailTime: km.Time, VictimName: km.Victim.CharacterName, VictimNameNormalized: atkm.NormalizeIdentity(km.Victim.CharacterName), VictimCorporation: km.Victim.CorporationName, VictimAlliance: km.Victim.AllianceName, VictimFaction: km.Victim.FactionName, VictimShipName: km.Victim.ShipName, VictimShipTypeID: victimTypeID, SolarSystem: km.SolarSystem, SecurityStatus: km.SecurityStatus, DamageTaken: km.Victim.DamageTaken, ResolutionStatus: "resolved"}
		attackers := make([]model.ATKillmailAttacker, len(km.Attackers))
		for i, attacker := range km.Attackers {
			shipID := typeIDPointer(typeIDs[attacker.ShipName])
			weaponID := typeIDPointer(typeIDs[attacker.WeaponName])
			if attacker.ShipName != "" && shipID == nil {
				unresolved = append(unresolved, attacker.ShipName)
			}
			if attacker.WeaponName != "" && weaponID == nil {
				unresolved = append(unresolved, attacker.WeaponName)
			}
			attackers[i] = model.ATKillmailAttacker{CharacterName: attacker.CharacterName, CorporationName: attacker.CorporationName, AllianceName: attacker.AllianceName, FactionName: attacker.FactionName, ShipName: attacker.ShipName, ShipTypeID: shipID, WeaponName: attacker.WeaponName, WeaponTypeID: weaponID, DamageDone: attacker.DamageDone, FinalBlow: attacker.FinalBlow}
		}
		unresolved = uniqueStrings(unresolved)
		if len(unresolved) > 0 {
			item.ResolutionStatus = "unresolved"
		}
		if err := s.repo.Create(item, attackers); err != nil {
			return nil, err
		}
		result.Items = append(result.Items, ATKillmailSubmitItem{KMID: item.KMID, ResolutionStatus: item.ResolutionStatus, UnresolvedNames: unresolved})
	}
	return result, nil
}

func typeIDPointer(value int64) *int64 {
	if value <= 0 {
		return nil
	}
	copy := value
	return &copy
}
func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, value := range values {
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result
}
func (s *ATKillmailService) Pending() ([]model.ATKillmail, error) { return s.repo.ListPending() }
func (s *ATKillmailService) List() ([]model.ATKillmail, error)    { return s.repo.ListAll() }
func (s *ATKillmailService) Get(id string) (*repository.ATKillmailDetail, error) {
	return s.repo.Get(id)
}
func (s *ATKillmailService) Suggestions(id string) ([]repository.ATKMSuggestion, error) {
	detail, err := s.repo.Get(id)
	if err != nil {
		return nil, err
	}
	return s.repo.Suggestions(detail.Killmail.KillmailTime)
}

func (s *ATKillmailService) Assign(actor uint, kmID string, matchID, roundID uint, admin bool) (string, error) {
	match, err := s.matchRepo.GetMatch(matchID)
	if err != nil {
		return "", err
	}
	if !admin && match.FCUserID != actor {
		return "", errors.New("无权管理该比赛")
	}
	round, err := s.matchRepo.GetRound(roundID)
	if err != nil || round.MatchID != matchID {
		return "", errors.New("比赛与局不匹配")
	}
	detail, err := s.repo.Get(kmID)
	if err != nil {
		return "", err
	}
	window, err := s.repo.RoundWindow(roundID)
	if err != nil {
		return "", err
	}
	first, last := detail.Killmail.KillmailTime, detail.Killmail.KillmailTime
	if window.First != nil && window.First.Before(first) {
		first = *window.First
	}
	if window.Last != nil && window.Last.After(last) {
		last = *window.Last
	}
	if last.Sub(first) > 15*time.Minute {
		return "", errors.New("该局首份至末份 KM 超过 15 分钟")
	}
	duplicateDeaths, err := s.repo.VictimInRound(roundID, detail.Killmail.VictimNameNormalized, kmID)
	if err != nil {
		return "", err
	}
	status := detail.Killmail.ResolutionStatus
	action := "assign"
	reason := ""
	if duplicateDeaths > 0 {
		status = "conflict"
		action = "assign_conflict"
		reason = "同一角色在同一局出现多次死亡"
	}
	if err := s.repo.Assign(kmID, matchID, roundID, actor, status, action, reason); err != nil {
		return "", err
	}
	return status, nil
}
func (s *ATKillmailService) ResolveConflict(actor, roundID uint, validKMID, reason string, admin bool) error {
	if strings.TrimSpace(reason) == "" {
		return errors.New("冲突处理必须填写原因")
	}
	round, err := s.matchRepo.GetRound(roundID)
	if err != nil {
		return err
	}
	match, err := s.matchRepo.GetMatch(round.MatchID)
	if err != nil {
		return err
	}
	if !admin && match.FCUserID != actor {
		return errors.New("无权管理该比赛")
	}
	detail, err := s.repo.Get(validKMID)
	if err != nil {
		return err
	}
	if detail.Killmail.AssignedRoundID == nil || *detail.Killmail.AssignedRoundID != roundID {
		return errors.New("有效 KM 不属于该局")
	}
	return s.repo.ResolveDeathConflict(roundID, match.ID, actor, detail.Killmail.VictimNameNormalized, validKMID, reason)
}

func prospectiveWindow(times []time.Time) time.Duration {
	if len(times) < 2 {
		return 0
	}
	sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })
	return times[len(times)-1].Sub(times[0])
}
