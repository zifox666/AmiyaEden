package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"time"

	"gorm.io/gorm"
)

// FleetRepository 舰队数据访问层
type FleetRepository struct{}

func NewFleetRepository() *FleetRepository {
	return &FleetRepository{}
}

// ─────────────────────────────────────────────
//  Fleet CRUD
// ─────────────────────────────────────────────

// Create 创建舰队
func (r *FleetRepository) Create(fleet *model.Fleet) error {
	return global.DB.Create(fleet).Error
}

// GetByID 根据 ID 查询舰队
func (r *FleetRepository) GetByID(id string) (*model.Fleet, error) {
	var fleet model.Fleet
	err := global.DB.Where("id = ? AND deleted_at IS NULL", id).First(&fleet).Error
	return &fleet, err
}

// Update 更新舰队信息
func (r *FleetRepository) Update(fleet *model.Fleet) error {
	return global.DB.Save(fleet).Error
}

// SoftDelete 软删除舰队
func (r *FleetRepository) SoftDelete(id string) error {
	return global.DB.Model(&model.Fleet{}).Where("id = ?", id).
		Update("deleted_at", global.DB.NowFunc()).Error
}

// FleetFilter 舰队列表筛选条件
type FleetFilter struct {
	Importance string
	FCUserID   *uint
}

// List 分页查询舰队列表
func (r *FleetRepository) List(page, pageSize int, filter FleetFilter) ([]model.Fleet, int64, error) {
	var fleets []model.Fleet
	var total int64

	offset := (page - 1) * pageSize
	db := global.DB.Model(&model.Fleet{}).Where("deleted_at IS NULL")

	if filter.Importance != "" {
		db = db.Where("importance = ?", filter.Importance)
	}
	if filter.FCUserID != nil {
		db = db.Where("fc_user_id = ?", *filter.FCUserID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("start_at DESC").Offset(offset).Limit(pageSize).Find(&fleets).Error; err != nil {
		return nil, 0, err
	}
	return fleets, total, nil
}

// ─────────────────────────────────────────────
//  Fleet Members
// ─────────────────────────────────────────────

// AddMember 添加舰队成员（如已存在则忽略）
func (r *FleetRepository) AddMember(member *model.FleetMember) error {
	return global.DB.Where("fleet_id = ? AND character_id = ?", member.FleetID, member.CharacterID).
		Assign(member).FirstOrCreate(member).Error
}

// ListMembers 查询舰队成员列表
func (r *FleetRepository) ListMembers(fleetID string) ([]model.FleetMember, error) {
	var members []model.FleetMember
	err := global.DB.Where("fleet_id = ?", fleetID).Order("joined_at ASC").Find(&members).Error
	return members, err
}

// RemoveMember 移除舰队成员
func (r *FleetRepository) RemoveMember(fleetID string, characterID int64) error {
	return global.DB.Where("fleet_id = ? AND character_id = ?", fleetID, characterID).
		Delete(&model.FleetMember{}).Error
}

// ─────────────────────────────────────────────
//  PAP Log
// ─────────────────────────────────────────────

// DeletePapLogsByFleet 删除某舰队的所有 PAP 记录（用于重新发放）
func (r *FleetRepository) DeletePapLogsByFleet(fleetID string) error {
	return global.DB.Where("fleet_id = ?", fleetID).Delete(&model.FleetPapLog{}).Error
}

// CreatePapLogs 批量创建 PAP 记录
func (r *FleetRepository) CreatePapLogs(logs []model.FleetPapLog) error {
	if len(logs) == 0 {
		return nil
	}
	return global.DB.Create(&logs).Error
}

// ListPapLogsByFleet 查询某舰队的 PAP 发放记录
func (r *FleetRepository) ListPapLogsByFleet(fleetID string) ([]model.FleetPapLog, error) {
	var logs []model.FleetPapLog
	err := global.DB.Where("fleet_id = ?", fleetID).Find(&logs).Error
	return logs, err
}

// UserPapLog 用户 PAP 记录（附带舰队类型）
type UserPapLog struct {
	ID           uint      `json:"id"`
	FleetID      string    `json:"fleet_id"`
	CharacterID  int64     `json:"character_id"`
	UserID       uint      `json:"user_id"`
	PapCount     float64   `json:"pap_count"`
	IssuedBy     uint      `json:"issued_by"`
	IssuedByName string    `json:"issued_by_name"`
	Importance   string    `json:"importance"`
	CreatedAt    time.Time `json:"created_at"`
}

// ListPapLogsByUser 查询某用户的所有 PAP 记录
func (r *FleetRepository) ListPapLogsByUser(userID uint) ([]UserPapLog, error) {
	var logs []UserPapLog
	err := global.DB.Model(&model.FleetPapLog{}).
		Select(`
			fleet_pap_log.id,
			fleet_pap_log.fleet_id,
			fleet_pap_log.character_id,
			fleet_pap_log.user_id,
			fleet_pap_log.pap_count,
			fleet_pap_log.issued_by,
			COALESCE(issuer_main.character_name, pap_user.nickname, '') as issued_by_name,
			COALESCE(fleet.importance, '') as importance,
			fleet_pap_log.issued_at as created_at
		`).
		Joins("LEFT JOIN fleet ON fleet.id = fleet_pap_log.fleet_id").
		Joins(`LEFT JOIN "user" pap_user ON pap_user.id = fleet_pap_log.issued_by`).
		Joins("LEFT JOIN eve_character issuer_main ON issuer_main.character_id = pap_user.primary_character_id").
		Where("fleet_pap_log.user_id = ?", userID).
		Order("fleet_pap_log.issued_at DESC").
		Scan(&logs).Error
	return logs, err
}

// ListFleetsByMemberUserID 查询用户参与过的舰队（通过成员记录关联）
func (r *FleetRepository) ListFleetsByMemberUserID(userID uint, limit int) ([]model.Fleet, error) {
	var fleets []model.Fleet
	subQuery := global.DB.Model(&model.FleetMember{}).Select("DISTINCT fleet_id").Where("user_id = ?", userID)
	err := global.DB.Where("id IN (?) AND deleted_at IS NULL", subQuery).
		Order("start_at DESC").Limit(limit).Find(&fleets).Error
	return fleets, err
}

// MonthlyPapStat 月度 PAP 汇总
type MonthlyPapStat struct {
	Year     int     `json:"year"`
	Month    int     `json:"month"`
	TotalPap float64 `json:"total_pap"`
}

// FleetPapSummaryFilter PAP 汇总筛选条件
type FleetPapSummaryFilter struct {
	StartAt *time.Time
	EndAt   *time.Time
}

// CorporationPapSummaryRow 军团 PAP 汇总行
type CorporationPapSummaryRow struct {
	UserID       uint    `json:"user_id"`
	StratOpPaps  float64 `json:"strat_op_paps"`
	SkirmishPaps float64 `json:"skirmish_paps"`
	TotalPaps    float64 `json:"-"`
}

// SumPapByUserGroupedByMonth 按月汇总用户 PAP
func (r *FleetRepository) SumPapByUserGroupedByMonth(userID uint) ([]MonthlyPapStat, error) {
	var stats []MonthlyPapStat
	err := global.DB.Model(&model.FleetPapLog{}).
		Select("YEAR(issued_at) as year, MONTH(issued_at) as month, COALESCE(SUM(pap_count), 0) as total_pap").
		Where("user_id = ?", userID).
		Group("YEAR(issued_at), MONTH(issued_at)").
		Order("year DESC, month DESC").
		Limit(12).
		Scan(&stats).Error
	return stats, err
}

// ListCorporationPapSummary 分页查询军团 PAP 汇总
func (r *FleetRepository) ListCorporationPapSummary(page, pageSize int, filter FleetPapSummaryFilter) ([]CorporationPapSummaryRow, int64, error) {
	var rows []CorporationPapSummaryRow
	var total int64

	offset := (page - 1) * pageSize
	baseDB := r.applyPapLogDateFilter(
		global.DB.Model(&model.FleetPapLog{}).
			Joins("LEFT JOIN fleet ON fleet.id = fleet_pap_log.fleet_id"),
		filter,
	)

	countSubQuery := baseDB.Select("user_id").Group("user_id")
	if err := global.DB.Table("(?) as pap_users", countSubQuery).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := baseDB.
		Select(`
			fleet_pap_log.user_id,
			COALESCE(SUM(CASE WHEN fleet.importance IN ('cta', 'strat_op') THEN fleet_pap_log.pap_count ELSE 0 END), 0) as strat_op_paps,
			COALESCE(SUM(CASE WHEN fleet.importance = 'other' THEN fleet_pap_log.pap_count ELSE 0 END), 0) as skirmish_paps,
			COALESCE(SUM(fleet_pap_log.pap_count), 0) as total_paps
		`).
		Group("user_id").
		Order("total_paps DESC, fleet_pap_log.user_id ASC").
		Offset(offset).
		Limit(pageSize).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

// ListCorporationPapSummaryAll 查询全部军团 PAP 汇总结果
func (r *FleetRepository) ListCorporationPapSummaryAll(filter FleetPapSummaryFilter) ([]CorporationPapSummaryRow, error) {
	var rows []CorporationPapSummaryRow

	err := r.applyPapLogDateFilter(
		global.DB.Model(&model.FleetPapLog{}).
			Joins("LEFT JOIN fleet ON fleet.id = fleet_pap_log.fleet_id"),
		filter,
	).
		Select(`
			fleet_pap_log.user_id,
			COALESCE(SUM(CASE WHEN fleet.importance IN ('cta', 'strat_op') THEN fleet_pap_log.pap_count ELSE 0 END), 0) as strat_op_paps,
			COALESCE(SUM(CASE WHEN fleet.importance = 'other' THEN fleet_pap_log.pap_count ELSE 0 END), 0) as skirmish_paps,
			COALESCE(SUM(fleet_pap_log.pap_count), 0) as total_paps
		`).
		Group("user_id").
		Order("total_paps DESC, fleet_pap_log.user_id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// SumPapTotal 汇总某一时间范围内的 PAP 总量
func (r *FleetRepository) SumPapTotal(filter FleetPapSummaryFilter) (float64, error) {
	var result struct {
		Total float64 `json:"total"`
	}

	err := r.applyPapLogDateFilter(global.DB.Model(&model.FleetPapLog{}), filter).
		Select("COALESCE(SUM(pap_count), 0) as total").
		Scan(&result).Error
	if err != nil {
		return 0, err
	}

	return result.Total, nil
}

func (r *FleetRepository) applyPapLogDateFilter(db *gorm.DB, filter FleetPapSummaryFilter) *gorm.DB {
	if filter.StartAt != nil {
		db = db.Where("issued_at >= ?", *filter.StartAt)
	}
	if filter.EndAt != nil {
		db = db.Where("issued_at < ?", *filter.EndAt)
	}
	return db
}

// ─────────────────────────────────────────────
//  Fleet Invite
// ─────────────────────────────────────────────

// CreateInvite 创建邀请链接
func (r *FleetRepository) CreateInvite(invite *model.FleetInvite) error {
	return global.DB.Create(invite).Error
}

// GetInviteByCode 根据邀请码查询
func (r *FleetRepository) GetInviteByCode(code string) (*model.FleetInvite, error) {
	var invite model.FleetInvite
	err := global.DB.Where("code = ?", code).First(&invite).Error
	return &invite, err
}

// DeactivateInvite 禁用邀请链接
func (r *FleetRepository) DeactivateInvite(id uint) error {
	return global.DB.Model(&model.FleetInvite{}).Where("id = ?", id).Update("active", false).Error
}

// ListInvitesByFleet 查询某舰队的邀请链接
func (r *FleetRepository) ListInvitesByFleet(fleetID string) ([]model.FleetInvite, error) {
	var invites []model.FleetInvite
	err := global.DB.Where("fleet_id = ?", fleetID).Order("created_at DESC").Find(&invites).Error
	return invites, err
}
