package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ATMatchRepository struct{}

func NewATMatchRepository() *ATMatchRepository { return &ATMatchRepository{} }

func (r *ATMatchRepository) SaveSeason(season *model.ATSeason) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if season.IsCurrent {
			if err := tx.Model(&model.ATSeason{}).Where("is_current = true").Update("is_current", false).Error; err != nil {
				return err
			}
		}
		if season.ID == 0 {
			return tx.Create(season).Error
		}
		result := tx.Model(&model.ATSeason{}).Where("id = ?", season.ID).Updates(map[string]any{
			"name": season.Name, "starts_at": season.StartsAt, "ends_at": season.EndsAt, "is_current": season.IsCurrent,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}
func (r *ATMatchRepository) ListSeasons() ([]model.ATSeason, error) {
	var result []model.ATSeason
	return result, global.DB.Order("starts_at DESC").Find(&result).Error
}
func (r *ATMatchRepository) GetSeason(id uint) (*model.ATSeason, error) {
	var result model.ATSeason
	err := global.DB.First(&result, id).Error
	return &result, err
}
func (r *ATMatchRepository) SaveMatchType(item *model.ATMatchType) error {
	if item.ID == 0 {
		return global.DB.Create(item).Error
	}
	result := global.DB.Model(&model.ATMatchType{}).Where("id = ?", item.ID).Updates(map[string]any{
		"name": item.Name, "contribution_value": item.ContributionValue, "active": item.Active,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
func (r *ATMatchRepository) GetMatchType(id uint) (*model.ATMatchType, error) {
	var result model.ATMatchType
	err := global.DB.First(&result, id).Error
	return &result, err
}
func (r *ATMatchRepository) ListMatchTypes() ([]model.ATMatchType, error) {
	var result []model.ATMatchType
	return result, global.DB.Order("id ASC").Find(&result).Error
}
func (r *ATMatchRepository) SetCleanupShips(seasonID uint, ships []model.ATCleanupShip) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		// The season/type unique key must be reusable when an administrator
		// removes and later re-adds the same cleanup ship.
		if err := tx.Unscoped().Where("season_id = ?", seasonID).Delete(&model.ATCleanupShip{}).Error; err != nil {
			return err
		}
		for i := range ships {
			ships[i].SeasonID = seasonID
		}
		if len(ships) > 0 {
			return tx.Create(&ships).Error
		}
		return nil
	})
}

func (r *ATMatchRepository) SearchContributionUsers(keyword string, limit int) ([]model.User, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	query := global.DB.Model(&model.User{}).
		Where("status = ?", 1).
		Order("nickname ASC, id ASC").
		Limit(limit)
	if keyword != "" {
		pattern := "%" + keyword + "%"
		query = query.Where(
			"CAST(id AS TEXT) = ? OR nickname ILIKE ? OR tournament_char_name ILIKE ?",
			keyword, pattern, pattern,
		)
	}
	var result []model.User
	return result, query.Find(&result).Error
}
func (r *ATMatchRepository) GetCleanupShips(seasonID uint) ([]model.ATCleanupShip, error) {
	var result []model.ATCleanupShip
	return result, global.DB.Where("season_id = ?", seasonID).Find(&result).Error
}

func (r *ATMatchRepository) CreateMatch(match *model.ATMatch) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(match).Error; err != nil {
			return err
		}
		return tx.Create(&model.ATMatchStateEvent{MatchID: match.ID, ActorUserID: match.FCUserID, ToStatus: match.Status, Reason: "create"}).Error
	})
}
func (r *ATMatchRepository) GetMatch(id uint) (*model.ATMatch, error) {
	var result model.ATMatch
	err := global.DB.First(&result, id).Error
	return &result, err
}
func (r *ATMatchRepository) GetMatchByUUID(id string) (*model.ATMatch, error) {
	var result model.ATMatch
	err := global.DB.Where("uuid = ?", id).First(&result).Error
	return &result, err
}
func (r *ATMatchRepository) ListMatches() ([]model.ATMatch, error) {
	var result []model.ATMatch
	return result, global.DB.Order("id DESC").Find(&result).Error
}
func (r *ATMatchRepository) SetMatchUUID(id uint, value string) error {
	return global.DB.Model(&model.ATMatch{}).Where("id = ? AND (uuid IS NULL OR uuid = '')", id).Update("uuid", value).Error
}
func (r *ATMatchRepository) UpdateMatchStatus(id, actor uint, from, to, reason string, updates map[string]any) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		updates["status"] = to
		result := tx.Model(&model.ATMatch{}).Where("id = ? AND status = ?", id, from).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrInvalidData
		}
		return tx.Create(&model.ATMatchStateEvent{MatchID: id, ActorUserID: actor, FromStatus: from, ToStatus: to, Reason: reason}).Error
	})
}
func (r *ATMatchRepository) UpdateWinsAndStatus(match *model.ATMatch, actor uint, from string) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.ATMatch{}).Where("id = ?", match.ID).Updates(map[string]any{"team_a_wins": match.TeamAWins, "team_b_wins": match.TeamBWins, "status": match.Status}).Error; err != nil {
			return err
		}
		if from != match.Status {
			return tx.Create(&model.ATMatchStateEvent{MatchID: match.ID, ActorUserID: actor, FromStatus: from, ToStatus: match.Status, Reason: "series result"}).Error
		}
		return nil
	})
}

func (r *ATMatchRepository) UpsertRegistration(registration *model.ATMatchRegistration) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if registration.SlotOrder > 0 && registration.Status != "rejected" {
			var count int64
			if err := tx.Model(&model.ATMatchRegistration{}).
				Where("match_id = ? AND slot_order = ? AND user_id != ? AND status != ?", registration.MatchID, registration.SlotOrder, registration.UserID, "rejected").
				Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return gorm.ErrDuplicatedKey
			}
		}
		return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "match_id"}, {Name: "user_id"}}, DoUpdates: clause.AssignmentColumns([]string{"status", "slot_order", "e_war_tags", "bonus_tags", "updated_at"})}).Create(registration).Error
	})
}
func (r *ATMatchRepository) ListRegistrations(matchID uint) ([]model.ATMatchRegistration, error) {
	var result []model.ATMatchRegistration
	return result, global.DB.Where("match_id = ?", matchID).Order("id ASC").Find(&result).Error
}

func (r *ATMatchRepository) CreateRound(round *model.ATMatchRound, roster []model.ATMatchRoundRoster, actor uint) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(round).Error; err != nil {
			return err
		}
		for i := range roster {
			roster[i].RoundID = round.ID
		}
		if len(roster) > 0 {
			if err := tx.Create(&roster).Error; err != nil {
				return err
			}
		}
		return tx.Create(&model.ATMatchStateEvent{MatchID: round.MatchID, RoundID: &round.ID, ActorUserID: actor, ToStatus: round.Status, Reason: "round create"}).Error
	})
}
func (r *ATMatchRepository) GetRound(id uint) (*model.ATMatchRound, error) {
	var result model.ATMatchRound
	err := global.DB.First(&result, id).Error
	return &result, err
}
func (r *ATMatchRepository) ListRounds(matchID uint) ([]model.ATMatchRound, error) {
	var result []model.ATMatchRound
	return result, global.DB.Where("match_id = ?", matchID).Order("round_no ASC, attempt ASC").Find(&result).Error
}
func (r *ATMatchRepository) GetRoster(roundID uint) ([]model.ATMatchRoundRoster, error) {
	var result []model.ATMatchRoundRoster
	return result, global.DB.Where("round_id = ?", roundID).Order("side ASC, slot_order ASC").Find(&result).Error
}
func (r *ATMatchRepository) ReplaceRoster(roundID uint, roster []model.ATMatchRoundRoster) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("round_id = ?", roundID).Delete(&model.ATMatchRoundRoster{}).Error; err != nil {
			return err
		}
		for i := range roster {
			roster[i].RoundID = roundID
		}
		if len(roster) > 0 {
			return tx.Create(&roster).Error
		}
		return nil
	})
}
func (r *ATMatchRepository) UpdateRound(round *model.ATMatchRound, actor uint, from string) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(round).Error; err != nil {
			return err
		}
		return tx.Create(&model.ATMatchStateEvent{MatchID: round.MatchID, RoundID: &round.ID, ActorUserID: actor, FromStatus: from, ToStatus: round.Status, Reason: round.VoidReason}).Error
	})
}
func (r *ATMatchRepository) VoidAllRounds(matchID uint, actor uint, reason string) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		var rounds []model.ATMatchRound
		if err := tx.Where("match_id = ? AND status != ?", matchID, "void").Find(&rounds).Error; err != nil {
			return err
		}
		for _, round := range rounds {
			if err := tx.Model(&model.ATMatchRound{}).Where("id = ?", round.ID).Updates(map[string]any{"status": "void", "void_reason": reason}).Error; err != nil {
				return err
			}
			if err := tx.Create(&model.ATMatchStateEvent{MatchID: matchID, RoundID: &round.ID, ActorUserID: actor, FromStatus: round.Status, ToStatus: "void", Reason: reason}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (r *ATMatchRepository) ParticipatingUserIDs(matchID uint) ([]uint, error) {
	var ids []uint
	err := global.DB.Table("at_match_round_roster rr").Distinct("rr.user_id").Joins("JOIN at_match_round r ON r.id = rr.round_id").Where("r.match_id = ? AND r.status = ? AND rr.user_id IS NOT NULL AND rr.side = ?", matchID, "confirmed", "A").Pluck("rr.user_id", &ids).Error
	return ids, err
}

func (r *ATMatchRepository) InsertLedger(entries []model.ATContributionLedger) error {
	if len(entries) == 0 {
		return nil
	}
	return global.DB.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "source_key"}}, DoNothing: true}).Create(&entries).Error
}
func (r *ATMatchRepository) MatchLedger(matchID uint) ([]model.ATContributionLedger, error) {
	var result []model.ATContributionLedger
	return result, global.DB.Where("match_id = ? AND source_type = ?", matchID, "match").Find(&result).Error
}

func (r *ATMatchRepository) UserLedger(userID uint, seasonID *uint) ([]model.ATContributionLedger, error) {
	var result []model.ATContributionLedger
	query := global.DB.Where("user_id = ?", userID).Order("created_at DESC, id DESC")
	if seasonID != nil {
		query = query.Where("season_id = ?", *seasonID)
	}
	return result, query.Find(&result).Error
}

type ATContributionLeaderboardRow struct {
	UserID                  uint    `json:"user_id"`
	Nickname                string  `json:"nickname"`
	TournamentCharacterName string  `json:"tournament_character_name"`
	Total                   float64 `json:"total"`
}

func (r *ATMatchRepository) Leaderboard(seasonID *uint) ([]ATContributionLeaderboardRow, error) {
	var result []ATContributionLeaderboardRow
	query := global.DB.Table("at_contribution_ledger ledger").
		Select("ledger.user_id, u.nickname, u.tournament_char_name AS tournament_character_name, SUM(ledger.amount) AS total").
		Joins(`JOIN "user" u ON u.id = ledger.user_id AND u.deleted_at IS NULL`).
		Group("ledger.user_id, u.nickname, u.tournament_char_name").Order("total DESC")
	if seasonID != nil {
		query = query.Where("ledger.season_id = ?", *seasonID)
	}
	return result, query.Scan(&result).Error
}
