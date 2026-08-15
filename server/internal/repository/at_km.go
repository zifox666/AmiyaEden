package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ATKillmailRepository struct{}

func NewATKillmailRepository() *ATKillmailRepository { return &ATKillmailRepository{} }

type ATKillmailDetail struct {
	Killmail  model.ATKillmail           `json:"killmail"`
	Attackers []model.ATKillmailAttacker `json:"attackers"`
}

func (r *ATKillmailRepository) GetByHash(hash string) (*model.ATKillmail, error) {
	var item model.ATKillmail
	err := global.DB.Where("content_hash = ?", hash).First(&item).Error
	return &item, err
}
func (r *ATKillmailRepository) Create(item *model.ATKillmail, attackers []model.ATKillmailAttacker) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(item).Error; err != nil {
			return err
		}
		for i := range attackers {
			attackers[i].KMID = item.KMID
		}
		if len(attackers) > 0 {
			return tx.Create(&attackers).Error
		}
		return nil
	})
}
func (r *ATKillmailRepository) Get(id string) (*ATKillmailDetail, error) {
	var detail ATKillmailDetail
	if err := global.DB.First(&detail.Killmail, "km_id = ?", id).Error; err != nil {
		return nil, err
	}
	if err := global.DB.Where("km_id = ?", id).Order("id ASC").Find(&detail.Attackers).Error; err != nil {
		return nil, err
	}
	return &detail, nil
}
func (r *ATKillmailRepository) ListPending() ([]model.ATKillmail, error) {
	var result []model.ATKillmail
	return result, global.DB.Where("assigned_round_id IS NULL OR resolution_status IN ?", []string{"unresolved", "conflict"}).Order("killmail_time ASC").Find(&result).Error
}
func (r *ATKillmailRepository) ListAll() ([]model.ATKillmail, error) {
	var result []model.ATKillmail
	return result, global.DB.Order("killmail_time DESC, created_at DESC").Find(&result).Error
}
func (r *ATKillmailRepository) RoundKillmails(roundID uint) ([]ATKillmailDetail, error) {
	var kms []model.ATKillmail
	if err := global.DB.Where("assigned_round_id = ? AND resolution_status = ?", roundID, "resolved").Order("killmail_time ASC").Find(&kms).Error; err != nil {
		return nil, err
	}
	result := make([]ATKillmailDetail, len(kms))
	for i, km := range kms {
		result[i].Killmail = km
		if err := global.DB.Where("km_id = ?", km.KMID).Find(&result[i].Attackers).Error; err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (r *ATKillmailRepository) CountRoundBlocking(roundID uint) (int64, error) {
	var count int64
	err := global.DB.Model(&model.ATKillmail{}).Where("assigned_round_id = ? AND resolution_status IN ?", roundID, []string{"unresolved", "conflict"}).Count(&count).Error
	return count, err
}

type ATKMAssignmentWindow struct {
	First *time.Time
	Last  *time.Time
}

func (r *ATKillmailRepository) RoundWindow(roundID uint) (ATKMAssignmentWindow, error) {
	var result ATKMAssignmentWindow
	err := global.DB.Model(&model.ATKillmail{}).Where("assigned_round_id = ? AND resolution_status != ?", roundID, "excluded").Select("MIN(killmail_time) AS first, MAX(killmail_time) AS last").Scan(&result).Error
	return result, err
}
func (r *ATKillmailRepository) VictimInRound(roundID uint, normalizedVictim string, excludeKMID string) (int64, error) {
	var count int64
	err := global.DB.Model(&model.ATKillmail{}).Where("assigned_round_id = ? AND victim_name_normalized = ? AND km_id != ? AND resolution_status != ?", roundID, normalizedVictim, excludeKMID, "excluded").Count(&count).Error
	return count, err
}
func (r *ATKillmailRepository) Assign(kmID string, matchID, roundID, actor uint, status, action, reason string) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.ATKillmail{}).Where("km_id = ?", kmID).Updates(map[string]any{"assigned_round_id": roundID, "resolution_status": status})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return tx.Create(&model.ATKillmailAssignmentEvent{KMID: kmID, MatchID: matchID, RoundID: roundID, ActorUserID: actor, Action: action, Reason: reason}).Error
	})
}
func (r *ATKillmailRepository) ResolveDeathConflict(roundID, matchID, actor uint, victim, validKMID, reason string) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		var kms []model.ATKillmail
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("assigned_round_id = ? AND victim_name_normalized = ?", roundID, victim).Find(&kms).Error; err != nil {
			return err
		}
		for _, km := range kms {
			status := "excluded"
			action := "exclude_conflict"
			if km.KMID == validKMID {
				status = "resolved"
				action = "resolve_conflict"
			}
			if err := tx.Model(&model.ATKillmail{}).Where("km_id = ?", km.KMID).Update("resolution_status", status).Error; err != nil {
				return err
			}
			if err := tx.Create(&model.ATKillmailAssignmentEvent{KMID: km.KMID, MatchID: matchID, RoundID: roundID, ActorUserID: actor, Action: action, Reason: reason}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

type ATKMSuggestion struct {
	MatchID   uint       `json:"match_id"`
	RoundID   uint       `json:"round_id"`
	RoundNo   int        `json:"round_no"`
	StartedAt *time.Time `json:"started_at"`
}

func (r *ATKillmailRepository) Suggestions(at time.Time) ([]ATKMSuggestion, error) {
	var result []ATKMSuggestion
	start := at.Add(-15 * time.Minute)
	end := at.Add(15 * time.Minute)
	err := global.DB.Table("at_match_round r").Select("r.match_id,r.id AS round_id,r.round_no,r.started_at").Joins("JOIN at_match m ON m.id = r.match_id").Where("m.status = ? AND r.status IN ? AND (r.started_at IS NULL OR r.started_at BETWEEN ? AND ?)", "in_progress", []string{"in_progress", "br_review"}, start, end).Scan(&result).Error
	return result, err
}
