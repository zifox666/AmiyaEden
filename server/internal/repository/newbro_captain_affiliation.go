package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"strings"
	"time"

	"gorm.io/gorm"
)

type NewbroCaptainAffiliationRepository struct{}

type AdminAffiliationHistoryFilter struct {
	CaptainSearch       string
	PlayerSearch        string
	ChangeStartedAtFrom *time.Time
	ChangeStartedAtTo   *time.Time
}

const captainEligiblePlayerSortExpr = `COALESCE(NULLIF("user".nickname, ''), primary_character.character_name, '')`

func NewNewbroCaptainAffiliationRepository() *NewbroCaptainAffiliationRepository {
	return &NewbroCaptainAffiliationRepository{}
}

func buildActiveByPlayerUserIDQuery(db *gorm.DB, userID uint) *gorm.DB {
	return db.Where("player_user_id = ? AND ended_at IS NULL", userID).
		Order("started_at DESC, id DESC")
}

func (r *NewbroCaptainAffiliationRepository) GetActiveByPlayerUserID(userID uint) (*model.NewbroCaptainAffiliation, error) {
	return r.GetActiveByPlayerUserIDTx(global.DB, userID)
}

func (r *NewbroCaptainAffiliationRepository) GetActiveByPlayerUserIDTx(
	tx *gorm.DB,
	userID uint,
) (*model.NewbroCaptainAffiliation, error) {
	var row model.NewbroCaptainAffiliation
	err := buildActiveByPlayerUserIDQuery(tx, userID).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *NewbroCaptainAffiliationRepository) ListRecentByPlayerUserID(userID uint, limit int) ([]model.NewbroCaptainAffiliation, error) {
	var rows []model.NewbroCaptainAffiliation
	if limit <= 0 {
		limit = 10
	}
	err := global.DB.Where("player_user_id = ?", userID).
		Order("started_at DESC, id DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *NewbroCaptainAffiliationRepository) ListByPlayerUserIDPaged(userID uint, page, pageSize int) ([]model.NewbroCaptainAffiliation, int64, error) {
	var rows []model.NewbroCaptainAffiliation
	var total int64
	db := global.DB.Where("player_user_id = ?", userID)
	if err := db.Model(&model.NewbroCaptainAffiliation{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("started_at DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&rows).Error
	return rows, total, err
}

func (r *NewbroCaptainAffiliationRepository) Create(row *model.NewbroCaptainAffiliation) error {
	return r.CreateTx(global.DB, row)
}

func (r *NewbroCaptainAffiliationRepository) CreateTx(tx *gorm.DB, row *model.NewbroCaptainAffiliation) error {
	return tx.Create(row).Error
}

func (r *NewbroCaptainAffiliationRepository) EndActiveByPlayerUserID(userID uint, endedAt time.Time) error {
	return r.EndActiveByPlayerUserIDTx(global.DB, userID, endedAt)
}

func (r *NewbroCaptainAffiliationRepository) EndActiveByPlayerUserIDTx(
	tx *gorm.DB,
	userID uint,
	endedAt time.Time,
) error {
	return tx.Model(&model.NewbroCaptainAffiliation{}).
		Where("player_user_id = ? AND ended_at IS NULL", userID).
		Update("ended_at", endedAt).Error
}

func (r *NewbroCaptainAffiliationRepository) GetActiveAt(playerUserID uint, at time.Time) (*model.NewbroCaptainAffiliation, error) {
	var row model.NewbroCaptainAffiliation
	err := global.DB.Where(
		"player_user_id = ? AND started_at <= ? AND (ended_at IS NULL OR ended_at > ?)",
		playerUserID,
		at,
		at,
	).
		Order("started_at DESC, id DESC").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *NewbroCaptainAffiliationRepository) CountActiveByCaptainUserIDs(userIDs []uint) (map[uint]int64, error) {
	result := make(map[uint]int64, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}
	type row struct {
		CaptainUserID uint
		Count         int64
	}
	var rows []row
	err := global.DB.Model(&model.NewbroCaptainAffiliation{}).
		Select("captain_user_id, COUNT(*) AS count").
		Where("captain_user_id IN ? AND ended_at IS NULL", userIDs).
		Group("captain_user_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, item := range rows {
		result[item.CaptainUserID] = item.Count
	}
	return result, nil
}

func (r *NewbroCaptainAffiliationRepository) CountDistinctPlayersByCaptainUserID(captainUserID uint, activeOnly bool) (int64, error) {
	var count int64
	db := global.DB.Model(&model.NewbroCaptainAffiliation{}).
		Where("captain_user_id = ?", captainUserID).
		Distinct("player_user_id")
	if activeOnly {
		db = db.Where("ended_at IS NULL")
	}
	err := db.Count(&count).Error
	return count, err
}

func buildCaptainEligiblePlayerListQuery(db *gorm.DB, captainUserID uint, keyword string) *gorm.DB {
	query := db.Model(&model.User{}).
		Joins(`JOIN newbro_player_state ON newbro_player_state.user_id = "user".id`).
		Joins(`LEFT JOIN eve_character AS primary_character ON primary_character.character_id = "user".primary_character_id`).
		Joins(`LEFT JOIN newbro_captain_affiliation AS current_affiliation ON current_affiliation.player_user_id = "user".id AND current_affiliation.ended_at IS NULL`).
		Where(`newbro_player_state.is_currently_newbro = ?`, true).
		Where(`"user".id <> ?`, captainUserID).
		Where(`(current_affiliation.captain_user_id IS NULL OR current_affiliation.captain_user_id <> ?)`, captainUserID)

	trimmedKeyword := strings.TrimSpace(keyword)
	if trimmedKeyword != "" {
		pattern := "%" + trimmedKeyword + "%"
		query = query.Where(`("user".nickname LIKE ? OR primary_character.character_name LIKE ?)`, pattern, pattern)
	}

	return query
}

func buildCaptainEligiblePlayerListSelectQuery(
	db *gorm.DB,
	captainUserID uint,
	keyword string,
	page int,
	pageSize int,
) *gorm.DB {
	return buildCaptainEligiblePlayerListQuery(db, captainUserID, keyword).
		Select(`DISTINCT "user".*, ` + captainEligiblePlayerSortExpr + ` AS player_sort_name`).
		Order(`"user".last_login_at DESC NULLS LAST, player_sort_name ASC, "user".id ASC`).
		Offset((page - 1) * pageSize).
		Limit(pageSize)
}

func (r *NewbroCaptainAffiliationRepository) ListCaptainEligiblePlayers(
	captainUserID uint,
	keyword string,
	page int,
	pageSize int,
) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	db := buildCaptainEligiblePlayerListQuery(global.DB, captainUserID, keyword)
	if err := db.Distinct(`"user".id`).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := buildCaptainEligiblePlayerListSelectQuery(global.DB, captainUserID, keyword, page, pageSize).
		Find(&users).Error
	return users, total, err
}

func (r *NewbroCaptainAffiliationRepository) ListActiveByPlayerUserIDs(userIDs []uint) ([]model.NewbroCaptainAffiliation, error) {
	var rows []model.NewbroCaptainAffiliation
	if len(userIDs) == 0 {
		return rows, nil
	}
	err := global.DB.Where("player_user_id IN ? AND ended_at IS NULL", userIDs).
		Order("started_at DESC, id DESC").
		Find(&rows).Error
	return rows, err
}

func (r *NewbroCaptainAffiliationRepository) ListByCaptainUserID(
	captainUserID uint,
	status string,
	page int,
	pageSize int,
) ([]model.NewbroCaptainAffiliation, int64, error) {
	var rows []model.NewbroCaptainAffiliation
	var total int64
	db := global.DB.Model(&model.NewbroCaptainAffiliation{}).
		Where("captain_user_id = ?", captainUserID)
	switch status {
	case "active":
		db = db.Where("ended_at IS NULL")
	case "historical":
		db = db.Where("ended_at IS NOT NULL")
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("CASE WHEN ended_at IS NULL THEN 0 ELSE 1 END ASC, started_at DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&rows).Error
	return rows, total, err
}

func buildAdminAffiliationHistoryQuery(db *gorm.DB, filter AdminAffiliationHistoryFilter) *gorm.DB {
	query := db.Model(&model.NewbroCaptainAffiliation{})

	if strings.TrimSpace(filter.CaptainSearch) != "" {
		pattern := "%" + strings.TrimSpace(filter.CaptainSearch) + "%"
		query = query.
			Joins(`LEFT JOIN "user" AS captain_user ON captain_user.id = newbro_captain_affiliation.captain_user_id`).
			Joins(`LEFT JOIN eve_character AS captain_character ON captain_character.character_id = newbro_captain_affiliation.captain_primary_character_id_at_start`).
			Where("(captain_user.nickname ILIKE ? OR captain_character.character_name ILIKE ?)", pattern, pattern)
	}
	if strings.TrimSpace(filter.PlayerSearch) != "" {
		pattern := "%" + strings.TrimSpace(filter.PlayerSearch) + "%"
		query = query.
			Joins(`LEFT JOIN "user" AS player_user ON player_user.id = newbro_captain_affiliation.player_user_id`).
			Joins(`LEFT JOIN eve_character AS player_character ON player_character.character_id = newbro_captain_affiliation.player_primary_character_id_at_start`).
			Where("(player_user.nickname ILIKE ? OR player_character.character_name ILIKE ?)", pattern, pattern)
	}
	if filter.ChangeStartedAtFrom != nil {
		query = query.Where("(started_at >= ? OR ended_at >= ?)", *filter.ChangeStartedAtFrom, *filter.ChangeStartedAtFrom)
	}
	if filter.ChangeStartedAtTo != nil {
		query = query.Where("(started_at <= ? OR ended_at <= ?)", *filter.ChangeStartedAtTo, *filter.ChangeStartedAtTo)
	}

	return query.Order("started_at DESC, id DESC")
}

func (r *NewbroCaptainAffiliationRepository) ListAdminAffiliationHistory(
	filter AdminAffiliationHistoryFilter,
	page int,
	pageSize int,
) ([]model.NewbroCaptainAffiliation, int64, error) {
	var rows []model.NewbroCaptainAffiliation
	var total int64

	db := buildAdminAffiliationHistoryQuery(global.DB, filter)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&rows).Error
	return rows, total, err
}
