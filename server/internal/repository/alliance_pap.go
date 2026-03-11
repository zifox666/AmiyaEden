package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// AlliancePAPRepository 联盟 PAP 数据访问层
type AlliancePAPRepository struct{}

func NewAlliancePAPRepository() *AlliancePAPRepository {
	return &AlliancePAPRepository{}
}

// UpsertRecord 插入或更新单条舰队记录（以 fleet_id + character_id 为唯一键）
func (r *AlliancePAPRepository) UpsertRecord(rec *model.AlliancePAPRecord) error {
	var existing model.AlliancePAPRecord
	err := global.DB.
		Where("fleet_id = ? AND character_id = ?", rec.FleetID, rec.CharacterID).
		First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return global.DB.Create(rec).Error
		}
		return err
	}
	rec.ID = existing.ID
	rec.CreatedAt = existing.CreatedAt
	return global.DB.Save(rec).Error
}

// UpsertSummary 插入或更新月度汇总（以 main_character + year + month 为唯一键）
func (r *AlliancePAPRepository) UpsertSummary(s *model.AlliancePAPSummary) error {
	var existing model.AlliancePAPSummary
	err := global.DB.
		Where("main_character = ? AND year = ? AND month = ?", s.MainCharacter, s.Year, s.Month).
		First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return global.DB.Create(s).Error
		}
		return err
	}
	s.ID = existing.ID
	s.CreatedAt = existing.CreatedAt
	return global.DB.Save(s).Error
}

// GetSummary 查询特定主角色某月汇总
func (r *AlliancePAPRepository) GetSummary(mainChar string, year, month int) (*model.AlliancePAPSummary, error) {
	var s model.AlliancePAPSummary
	err := global.DB.
		Where("main_character = ? AND year = ? AND month = ?", mainChar, year, month).
		First(&s).Error
	return &s, err
}

// ListRecords 查询特定主角色某月的所有舰队明细
func (r *AlliancePAPRepository) ListRecords(mainChar string, year, month int) ([]model.AlliancePAPRecord, error) {
	var records []model.AlliancePAPRecord
	err := global.DB.
		Where("main_character = ? AND year = ? AND month = ?", mainChar, year, month).
		Order("start_at DESC").
		Find(&records).Error
	return records, err
}

// ListAllSummaries 查询所有人某月的汇总（管理员视图）
func (r *AlliancePAPRepository) ListAllSummaries(year, month int) ([]model.AlliancePAPSummary, error) {
	var list []model.AlliancePAPSummary
	err := global.DB.
		Where("year = ? AND month = ?", year, month).
		Order("total_pap DESC").
		Find(&list).Error
	return list, err
}

// ListAllSummariesPaged 分页查询所有人某月的汇总（管理员视图）
// corporationIDs 非空时只返回这些军团的数据
func (r *AlliancePAPRepository) ListAllSummariesPaged(year, month, page, pageSize int, corporationIDs []int64) ([]model.AlliancePAPSummary, int64, error) {
	var list []model.AlliancePAPSummary
	var total int64
	offset := (page - 1) * pageSize

	db := global.DB.Model(&model.AlliancePAPSummary{}).Where("year = ? AND month = ?", year, month)
	if len(corporationIDs) > 0 {
		strIDs := make([]string, len(corporationIDs))
		for i, id := range corporationIDs {
			strIDs[i] = fmt.Sprintf("%d", id)
		}
		db = db.Where("corporation_id IN ?", strIDs)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("total_pap DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// ListAllMainCharacters 查询数据库中所有已有记录的主角色名列表
func (r *AlliancePAPRepository) ListAllMainCharacters() ([]string, error) {
	var names []string
	err := global.DB.
		Model(&model.AlliancePAPSummary{}).
		Distinct("main_character").
		Pluck("main_character", &names).Error
	return names, err
}

// MarkArchived 将某月所有记录和汇总标记为已归档
func (r *AlliancePAPRepository) MarkArchived(year, month int) error {
	if err := global.DB.
		Model(&model.AlliancePAPRecord{}).
		Where("year = ? AND month = ?", year, month).
		Update("is_archived", true).Error; err != nil {
		return err
	}
	return global.DB.
		Model(&model.AlliancePAPSummary{}).
		Where("year = ? AND month = ?", year, month).
		Update("is_archived", true).Error
}

// ListSummariesByMainChar 查询指定主角色的月度汇总（最近 N 条）
func (r *AlliancePAPRepository) ListSummariesByMainChar(mainChar string, limit int) ([]model.AlliancePAPSummary, error) {
	var list []model.AlliancePAPSummary
	err := global.DB.Where("main_character = ?", mainChar).
		Order("year DESC, month DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

// ListRecentRecordsByMainChar 查询指定主角色最近的舰队参与记录
func (r *AlliancePAPRepository) ListRecentRecordsByMainChar(mainChar string, limit int) ([]model.AlliancePAPRecord, error) {
	var records []model.AlliancePAPRecord
	err := global.DB.Where("main_character = ?", mainChar).
		Order("start_at DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

// ─────────────────────────────────────────────
//  PAP 月度结算 / 兑换
// ─────────────────────────────────────────────

// ListUnredeemedSummaries 查询某月尚未兑换的汇总列表
// corporationIDs 非空时只返回这些军团的数据
func (r *AlliancePAPRepository) ListUnredeemedSummaries(year, month int, corporationIDs []int64) ([]model.AlliancePAPSummary, error) {
	var list []model.AlliancePAPSummary
	db := global.DB.
		Where("year = ? AND month = ? AND is_redeemed = false AND total_pap > 0", year, month)
	if len(corporationIDs) > 0 {
		strIDs := make([]string, len(corporationIDs))
		for i, id := range corporationIDs {
			strIDs[i] = fmt.Sprintf("%d", id)
		}
		db = db.Where("corporation_id IN ?", strIDs)
	}
	err := db.Find(&list).Error
	return list, err
}

// MarkSummaryRedeemed 将汇总标记为已兑换，并记录发放金额
func (r *AlliancePAPRepository) MarkSummaryRedeemed(id uint, walletIssued float64) error {
	return global.DB.Model(&model.AlliancePAPSummary{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_redeemed":   true,
			"wallet_issued": walletIssued,
		}).Error
}
