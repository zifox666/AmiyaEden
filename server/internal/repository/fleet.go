package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"time"
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

// MemberWithPap 舰队成员 + PAP 信息
type MemberWithPap struct {
	model.FleetMember
	PapCount *float64   `json:"pap_count"`
	IssuedAt *time.Time `json:"issued_at"`
}

// ListMembersWithPap 分页查询舰队成员（左连接 PAP 记录）
func (r *FleetRepository) ListMembersWithPap(fleetID string, page, pageSize int) ([]MemberWithPap, int64, error) {
	var results []MemberWithPap
	var total int64
	offset := (page - 1) * pageSize

	base := global.DB.Table("fleet_member").Where("fleet_member.fleet_id = ?", fleetID)
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := global.DB.Table("fleet_member").
		Select("fleet_member.*, fleet_pap_log.pap_count, fleet_pap_log.issued_at").
		Joins("LEFT JOIN fleet_pap_log ON fleet_pap_log.fleet_id = fleet_member.fleet_id AND fleet_pap_log.character_id = fleet_member.character_id").
		Where("fleet_member.fleet_id = ?", fleetID).
		Order("fleet_member.joined_at ASC").
		Offset(offset).Limit(pageSize).
		Scan(&results).Error
	return results, total, err
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

// ListPapLogsByUser 查询某用户的所有 PAP 记录
func (r *FleetRepository) ListPapLogsByUser(userID uint) ([]model.FleetPapLog, error) {
	var logs []model.FleetPapLog
	err := global.DB.Where("user_id = ?", userID).Order("issued_at DESC").Find(&logs).Error
	return logs, err
}

// PapLogDetail PAP 记录（含角色名、FC 名称、舰队信息）
type PapLogDetail struct {
	model.FleetPapLog
	CharacterName   string `json:"character_name"`
	FleetTitle      string `json:"fleet_title"`
	FleetStartAt    string `json:"fleet_start_at"`
	FCCharacterName string `json:"fc_character_name"`
	FleetImportance string `json:"fleet_importance"`
	ShipTypeID      *int64 `json:"ship_type_id"`
}

// ListPapLogsDetailByUser 查询某用户的 PAP 记录（JOIN 舰队、角色信息）
func (r *FleetRepository) ListPapLogsDetailByUser(userID uint) ([]PapLogDetail, error) {
	var results []PapLogDetail
	err := global.DB.Table("fleet_pap_log p").
		Select(`p.*,
			COALESCE(ec.character_name, '') AS character_name,
			COALESCE(f.title, '') AS fleet_title,
			COALESCE(CAST(f.start_at AS TEXT), '') AS fleet_start_at,
			COALESCE(f.fc_character_name, '') AS fc_character_name,
			COALESCE(f.importance, '') AS fleet_importance,
			fm.ship_type_id`).
		Joins("LEFT JOIN eve_character ec ON ec.character_id = p.character_id").
		Joins("LEFT JOIN fleet f ON f.id = p.fleet_id AND f.deleted_at IS NULL").
		Joins("LEFT JOIN fleet_member fm ON fm.fleet_id = p.fleet_id AND fm.character_id = p.character_id").
		Where("p.user_id = ?", userID).
		Order("p.issued_at DESC").
		Scan(&results).Error
	return results, err
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
