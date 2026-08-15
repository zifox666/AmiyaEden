package repository

import (
	"errors"

	"amiya-eden/global"
	"amiya-eden/internal/model"

	"gorm.io/gorm"
)

type CombatLogRepository struct{}

func NewCombatLogRepository() *CombatLogRepository { return &CombatLogRepository{} }

type CombatLogReportItem struct {
	model.CombatLogReport
	UserNickname string `json:"user_nickname"`
}

type CombatLogDetail struct {
	Report        CombatLogReportItem         `json:"report"`
	Breakdowns    []model.CombatLogBreakdown  `json:"breakdowns"`
	Timeline      []model.CombatLogTimeline   `json:"timeline"`
	HitQualities  []model.CombatLogHitQuality `json:"hit_qualities"`
	AverageDPSOut float64                     `json:"average_dps_out"`
	AverageDPSIn  float64                     `json:"average_dps_in"`
}

func (r *CombatLogRepository) FindByUserAndHash(userID uint, hash string) (*model.CombatLogReport, error) {
	var report model.CombatLogReport
	err := global.DB.Where("user_id = ? AND file_hash = ?", userID, hash).First(&report).Error
	return &report, err
}

func (r *CombatLogRepository) Create(report *model.CombatLogReport, breakdowns []model.CombatLogBreakdown, timeline []model.CombatLogTimeline, hitQualities []model.CombatLogHitQuality) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(report).Error; err != nil {
			return err
		}
		for i := range breakdowns {
			breakdowns[i].ReportID = report.ID
		}
		for i := range timeline {
			timeline[i].ReportID = report.ID
		}
		for i := range hitQualities {
			hitQualities[i].ReportID = report.ID
		}
		if len(breakdowns) > 0 {
			if err := tx.Create(&breakdowns).Error; err != nil {
				return err
			}
		}
		if len(timeline) > 0 {
			if err := tx.Create(&timeline).Error; err != nil {
				return err
			}
		}
		if len(hitQualities) > 0 {
			if err := tx.Create(&hitQualities).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *CombatLogRepository) ReplaceHitQualities(reportID uint, rows []model.CombatLogHitQuality) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("report_id = ?", reportID).Delete(&model.CombatLogHitQuality{}).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		for i := range rows {
			rows[i].ReportID = reportID
		}
		return tx.Create(&rows).Error
	})
}

func (r *CombatLogRepository) List(userID uint, includeAll bool, page, pageSize int) ([]CombatLogReportItem, int64, error) {
	query := global.DB.Table("combat_log_report r").Where("r.deleted_at IS NULL")
	if !includeAll {
		query = query.Where("r.user_id = ?", userID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := make([]CombatLogReportItem, 0)
	err := query.
		Select("r.*, COALESCE(u.nickname, '') AS user_nickname").
		Joins("LEFT JOIN \"user\" u ON u.id = r.user_id AND u.deleted_at IS NULL").
		Order("r.created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&items).Error
	return items, total, err
}

func (r *CombatLogRepository) Get(id uint) (*CombatLogDetail, error) {
	var item CombatLogReportItem
	err := global.DB.Table("combat_log_report r").
		Select("r.*, COALESCE(u.nickname, '') AS user_nickname").
		Joins("LEFT JOIN \"user\" u ON u.id = r.user_id AND u.deleted_at IS NULL").
		Where("r.id = ? AND r.deleted_at IS NULL", id).
		Scan(&item).Error
	if err != nil {
		return nil, err
	}
	if item.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	detail := &CombatLogDetail{Report: item, Breakdowns: []model.CombatLogBreakdown{}, Timeline: []model.CombatLogTimeline{}, HitQualities: []model.CombatLogHitQuality{}}
	if err := global.DB.Where("report_id = ?", id).Order("category ASC, amount DESC").Find(&detail.Breakdowns).Error; err != nil {
		return nil, err
	}
	if err := global.DB.Where("report_id = ?", id).Order("occurred_at ASC, category ASC").Find(&detail.Timeline).Error; err != nil {
		return nil, err
	}
	if err := global.DB.Where("report_id = ?", id).Order("direction ASC, id ASC").Find(&detail.HitQualities).Error; err != nil {
		return nil, err
	}
	hitTotals := make(map[string]int)
	for _, row := range detail.HitQualities {
		hitTotals[row.Direction] += row.HitCount
	}
	for i := range detail.HitQualities {
		if total := hitTotals[detail.HitQualities[i].Direction]; total > 0 {
			detail.HitQualities[i].Percentage = float64(detail.HitQualities[i].HitCount) * 100 / float64(total)
		}
	}
	duration := item.DurationSeconds
	if duration < 1 {
		duration = 1
	}
	detail.AverageDPSOut = float64(item.DamageOut) / float64(duration)
	detail.AverageDPSIn = float64(item.DamageIn) / float64(duration)
	return detail, nil
}

func IsCombatLogNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }
