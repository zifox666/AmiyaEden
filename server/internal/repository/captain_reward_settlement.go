package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"time"

	"gorm.io/gorm"
)

type CaptainRewardSettlementRepository struct{}

func NewCaptainRewardSettlementRepository() *CaptainRewardSettlementRepository {
	return &CaptainRewardSettlementRepository{}
}

type CaptainRewardSettlementFilter struct {
	CaptainUserID *uint
	Keyword       string
}

func (r *CaptainRewardSettlementRepository) CreateTx(tx *gorm.DB, row *model.CaptainRewardSettlement) error {
	return tx.Create(row).Error
}

func buildCaptainRewardSettlementBaseQuery(db *gorm.DB, filter CaptainRewardSettlementFilter) *gorm.DB {
	query := db.Model(&model.CaptainRewardSettlement{})
	if filter.CaptainUserID != nil && *filter.CaptainUserID > 0 {
		query = query.Where("captain_user_id = ?", *filter.CaptainUserID)
	}
	return applyKeywordLikeFilter(
		query,
		filter.Keyword,
		`EXISTS (SELECT 1 FROM "user" AS captain_user WHERE captain_user.id = captain_reward_settlement.captain_user_id AND LOWER(captain_user.nickname) LIKE ?)`,
		`EXISTS (SELECT 1 FROM eve_character AS captain_character WHERE captain_character.user_id = captain_reward_settlement.captain_user_id AND LOWER(captain_character.character_name) LIKE ?)`)
}

func buildCaptainRewardSettlementListQuery(
	db *gorm.DB,
	filter CaptainRewardSettlementFilter,
	page int,
	pageSize int,
) *gorm.DB {
	return buildCaptainRewardSettlementBaseQuery(db, filter).
		Order("processed_at DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize)
}

func (r *CaptainRewardSettlementRepository) List(
	filter CaptainRewardSettlementFilter,
	page int,
	pageSize int,
) ([]model.CaptainRewardSettlement, int64, error) {
	var rows []model.CaptainRewardSettlement
	var total int64
	db := buildCaptainRewardSettlementBaseQuery(global.DB, filter)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := buildCaptainRewardSettlementListQuery(global.DB, filter, page, pageSize).
		Find(&rows).Error
	return rows, total, err
}

func (r *CaptainRewardSettlementRepository) Summarize(
	filter CaptainRewardSettlementFilter,
) (settlementCount int64, totalCreditedValue float64, lastProcessedAt *time.Time, err error) {
	type row struct {
		SettlementCount    int64
		TotalCreditedValue float64
		LastProcessedAt    *time.Time
	}
	var result row
	err = buildCaptainRewardSettlementBaseQuery(global.DB, filter).
		Select(`
			COUNT(*) AS settlement_count,
			COALESCE(SUM(credited_value), 0) AS total_credited_value,
			MAX(processed_at) AS last_processed_at
		`).
		Scan(&result).Error
	return result.SettlementCount, result.TotalCreditedValue, result.LastProcessedAt, err
}
