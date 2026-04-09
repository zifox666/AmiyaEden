package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrNoApprovedUnpaidBatchPayoutApplications = errors.New("no approved unpaid batch payout applications")
	ErrBatchPayoutSelectionChanged             = errors.New("batch payout selection changed")
)

// SrpRepository SRP 数据访问层
type SrpRepository struct{}

func NewSrpRepository() *SrpRepository {
	return &SrpRepository{}
}

func buildPendingBadgeSrpCountQuery(db *gorm.DB) *gorm.DB {
	return db.Model(&model.SrpApplication{}).
		Where("review_status IN ?", []string{model.SrpReviewSubmitted, model.SrpReviewApproved}).
		Where("payout_status = ?", model.SrpPayoutNotPaid)
}

func (r *SrpRepository) CountPendingBadgeApplications() (int64, error) {
	var count int64
	err := buildPendingBadgeSrpCountQuery(global.DB).Count(&count).Error
	return count, err
}

// ─────────────────────────────────────────────
//  SrpShipPrice CRUD
// ─────────────────────────────────────────────

// GetShipPriceByTypeID 按 ship_type_id 查找
func (r *SrpRepository) GetShipPriceByTypeID(shipTypeID int64) (*model.SrpShipPrice, error) {
	var p model.SrpShipPrice
	err := global.DB.Where("ship_type_id = ?", shipTypeID).First(&p).Error
	return &p, err
}

// ListShipPrices 查询所有舰船价格（可按名称模糊搜索）
func (r *SrpRepository) ListShipPrices(keyword string) ([]model.SrpShipPrice, error) {
	var list []model.SrpShipPrice
	db := global.DB.Model(&model.SrpShipPrice{})
	if keyword != "" {
		db = db.Where("ship_name LIKE ?", "%"+keyword+"%")
	}
	err := db.Order("ship_type_id ASC").Find(&list).Error
	return list, err
}

// UpsertShipPrice 创建或更新舰船价格
func (r *SrpRepository) UpsertShipPrice(p *model.SrpShipPrice) error {
	return global.DB.Save(p).Error
}

// DeleteShipPrice 删除舰船价格
func (r *SrpRepository) DeleteShipPrice(id uint) error {
	return global.DB.Delete(&model.SrpShipPrice{}, id).Error
}

// ─────────────────────────────────────────────
//  SrpApplication CRUD
// ─────────────────────────────────────────────

// CreateApplication 创建补损申请
func (r *SrpRepository) CreateApplication(app *model.SrpApplication) error {
	return global.DB.Create(app).Error
}

// GetApplicationByID 按 ID 查询
func (r *SrpRepository) GetApplicationByID(id uint) (*model.SrpApplication, error) {
	var app model.SrpApplication
	err := global.DB.First(&app, id).Error
	return &app, err
}

// GetApplicationByIDForUpdate 在事务中按 ID 查询并加锁
func (r *SrpRepository) GetApplicationByIDForUpdate(tx *gorm.DB, id uint) (*model.SrpApplication, error) {
	var app model.SrpApplication
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&app, id).Error
	return &app, err
}

// ExistsApplicationByKillmail 检查该 killmail 是否已被该人物提交过申请
func (r *SrpRepository) ExistsApplicationByKillmail(killmailID int64, characterID int64) bool {
	var count int64
	global.DB.Model(&model.SrpApplication{}).
		Where("killmail_id = ? AND character_id = ?", killmailID, characterID).
		Count(&count)
	return count > 0
}

// UpdateApplication 更新申请（审批 / 发放）
func (r *SrpRepository) UpdateApplication(app *model.SrpApplication) error {
	return global.DB.Save(app).Error
}

// UpdateApplicationTx 在事务中更新申请
func (r *SrpRepository) UpdateApplicationTx(tx *gorm.DB, app *model.SrpApplication) error {
	return tx.Save(app).Error
}

func buildSubmittedLinkedApplicationsQuery(db *gorm.DB) *gorm.DB {
	return db.Model(&model.SrpApplication{}).
		Where("review_status = ?", model.SrpReviewSubmitted).
		Where("fleet_id IS NOT NULL AND fleet_id <> ''")
}

// ListSubmittedLinkedApplications 查询所有 submitted 且已关联舰队的申请
func (r *SrpRepository) ListSubmittedLinkedApplications() ([]model.SrpApplication, error) {
	var list []model.SrpApplication
	err := buildSubmittedLinkedApplicationsQuery(global.DB).
		Order("id ASC").
		Find(&list).Error
	return list, err
}

// ListSubmittedLinkedApplicationsByFleet 查询指定舰队下 submitted 且已关联舰队的申请
func (r *SrpRepository) ListSubmittedLinkedApplicationsByFleet(fleetID string) ([]model.SrpApplication, error) {
	var list []model.SrpApplication
	err := buildSubmittedLinkedApplicationsQuery(global.DB).
		Where("fleet_id = ?", fleetID).
		Order("id ASC").
		Find(&list).Error
	return list, err
}

// SrpTabType 申请列表 Tab 分类
type SrpTabType string

const (
	SrpTabPending SrpTabType = "pending" // 待处理：pending/approved + unpaid
	SrpTabHistory SrpTabType = "history" // 发放记录：paid OR rejected
)

// SrpApplicationFilter 申请列表筛选条件
type SrpApplicationFilter struct {
	Tab          SrpTabType
	UserID       *uint
	CharacterID  *int64
	FleetID      *string
	ReviewStatus string
	PayoutStatus string
	Keyword      string
}

// SrpBatchPayoutSummaryRow 按用户聚合的待批量发放汇总
type SrpBatchPayoutSummaryRow struct {
	UserID           uint    `json:"user_id"`
	TotalAmount      float64 `json:"total_amount"`
	ApplicationCount int64   `json:"application_count"`
}

func buildSrpApplicationListQuery(db *gorm.DB, filter SrpApplicationFilter) *gorm.DB {
	query := db.Model(&model.SrpApplication{})

	switch filter.Tab {
	case SrpTabPending:
		query = query.Where("review_status IN (?, ?) AND payout_status = ?",
			model.SrpReviewSubmitted, model.SrpReviewApproved, model.SrpPayoutNotPaid)
	case SrpTabHistory:
		query = query.Where("payout_status = ? OR review_status = ?",
			model.SrpPayoutPaid, model.SrpReviewRejected)
	}
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.CharacterID != nil {
		query = query.Where("character_id = ?", *filter.CharacterID)
	}
	if filter.FleetID != nil {
		query = query.Where("fleet_id = ?", *filter.FleetID)
	}
	if filter.ReviewStatus != "" {
		query = query.Where("review_status = ?", filter.ReviewStatus)
	}
	if filter.PayoutStatus != "" {
		query = query.Where("payout_status = ?", filter.PayoutStatus)
	}
	query = applyKeywordLikeFilter(
		query,
		filter.Keyword,
		`EXISTS (SELECT 1 FROM "user" AS applicant_user WHERE applicant_user.id = srp_application.user_id AND LOWER(applicant_user.nickname) LIKE ?)`,
		"LOWER(character_name) LIKE ?")

	return query
}

// ListApplications 分页查询申请列表
func (r *SrpRepository) ListApplications(page, pageSize int, filter SrpApplicationFilter) ([]model.SrpApplication, int64, error) {
	var list []model.SrpApplication
	var total int64

	db := buildSrpApplicationListQuery(global.DB, filter)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// ListMyApplications 查询当前用户的申请（按用户 ID）
func (r *SrpRepository) ListMyApplications(userID uint, page, pageSize int) ([]model.SrpApplication, int64, error) {
	uid := &userID
	return r.ListApplications(page, pageSize, SrpApplicationFilter{UserID: uid})
}

// ListBatchPayoutSummary 查询所有可批量发放的按用户汇总数据
func (r *SrpRepository) ListBatchPayoutSummary() ([]SrpBatchPayoutSummaryRow, error) {
	var list []SrpBatchPayoutSummaryRow
	err := global.DB.Model(&model.SrpApplication{}).
		Select(`
			user_id,
			SUM(final_amount) AS total_amount,
			COUNT(id) AS application_count
		`).
		Where("payout_status = ? AND review_status = ?", model.SrpPayoutNotPaid, model.SrpReviewApproved).
		Group("user_id").
		Order("total_amount DESC, user_id ASC").
		Scan(&list).Error
	return list, err
}

func buildApprovedUnpaidBatchPayoutApplicationsQuery(db *gorm.DB, userID uint) *gorm.DB {
	return db.Model(&model.SrpApplication{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND payout_status = ? AND review_status = ?", userID, model.SrpPayoutNotPaid, model.SrpReviewApproved).
		Order("id ASC")
}

// CheckMaxApprovedUnpaidAmount 检查某用户待发放申请中是否存在超过上限的单笔金额
func (r *SrpRepository) CheckMaxApprovedUnpaidAmount(userID uint, limitISK float64) error {
	var maxAmount float64
	err := global.DB.Model(&model.SrpApplication{}).
		Where("user_id = ? AND payout_status = ? AND review_status = ?", userID, model.SrpPayoutNotPaid, model.SrpReviewApproved).
		Select("COALESCE(MAX(final_amount), 0)").
		Scan(&maxAmount).Error
	if err != nil {
		return err
	}
	if maxAmount > limitISK {
		return fmt.Errorf("该用户存在单笔金额 %.0f ISK 超过 SRP 职权上限 6 亿 ISK 的待发放申请", maxAmount)
	}
	return nil
}

func buildApprovedUnpaidApplicationsForUpdateQuery(db *gorm.DB) *gorm.DB {
	return db.Model(&model.SrpApplication{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("payout_status = ? AND review_status = ?", model.SrpPayoutNotPaid, model.SrpReviewApproved).
		Order("user_id ASC, id ASC")
}

func summarizeBatchPayoutApplications(userID uint, apps []model.SrpApplication) (SrpBatchPayoutSummaryRow, []uint) {
	summary := SrpBatchPayoutSummaryRow{UserID: userID}
	ids := make([]uint, 0, len(apps))
	for _, app := range apps {
		summary.TotalAmount += app.FinalAmount
		summary.ApplicationCount++
		ids = append(ids, app.ID)
	}
	return summary, ids
}

// ListApprovedUnpaidApplicationsForUpdate 查询全部已批准未发放的申请并加锁
func (r *SrpRepository) ListApprovedUnpaidApplicationsForUpdate(tx *gorm.DB) ([]model.SrpApplication, error) {
	var apps []model.SrpApplication
	err := buildApprovedUnpaidApplicationsForUpdateQuery(tx).Find(&apps).Error
	return apps, err
}

func buildBatchPayoutApplicationsUpdateQuery(db *gorm.DB, applicationIDs []uint, payerID uint, paidAt time.Time) *gorm.DB {
	return db.Model(&model.SrpApplication{}).
		Where("id IN ?", applicationIDs).
		Where("payout_status = ? AND review_status = ?", model.SrpPayoutNotPaid, model.SrpReviewApproved).
		Updates(map[string]interface{}{
			"payout_status": model.SrpPayoutPaid,
			"paid_by":       payerID,
			"paid_at":       paidAt,
		})
}

// BatchPayoutApplicationsByUser 将某用户所有已批准且待发放的申请标记为已发放
func (r *SrpRepository) BatchPayoutApplicationsByUser(userID uint, payerID uint, paidAt time.Time) (*SrpBatchPayoutSummaryRow, []model.SrpApplication, error) {
	var summary *SrpBatchPayoutSummaryRow
	var selectedApps []model.SrpApplication

	err := global.DB.Transaction(func(tx *gorm.DB) error {
		var apps []model.SrpApplication
		if err := buildApprovedUnpaidBatchPayoutApplicationsQuery(tx, userID).Find(&apps).Error; err != nil {
			return err
		}
		if len(apps) == 0 {
			return ErrNoApprovedUnpaidBatchPayoutApplications
		}

		selectedSummary, applicationIDs := summarizeBatchPayoutApplications(userID, apps)
		updateTx := buildBatchPayoutApplicationsUpdateQuery(tx, applicationIDs, payerID, paidAt)
		if updateTx.Error != nil {
			return updateTx.Error
		}
		if updateTx.RowsAffected != int64(len(applicationIDs)) {
			return ErrBatchPayoutSelectionChanged
		}

		summary = &selectedSummary
		selectedApps = append(selectedApps[:0], apps...)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return summary, selectedApps, nil
}
