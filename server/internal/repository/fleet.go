package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
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

// ListPapLogsByUser 查询某用户的所有 PAP 记录
func (r *FleetRepository) ListPapLogsByUser(userID uint) ([]model.FleetPapLog, error) {
	var logs []model.FleetPapLog
	err := global.DB.Where("user_id = ?", userID).Order("issued_at DESC").Find(&logs).Error
	return logs, err
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

// ─────────────────────────────────────────────
//  System Wallet
// ─────────────────────────────────────────────

// GetOrCreateWallet 获取或创建用户钱包
func (r *FleetRepository) GetOrCreateWallet(userID uint) (*model.SystemWallet, error) {
	var wallet model.SystemWallet
	err := global.DB.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		wallet = model.SystemWallet{UserID: userID, Balance: 0}
		if err := global.DB.Create(&wallet).Error; err != nil {
			return nil, err
		}
	}
	return &wallet, nil
}

// UpdateWalletBalance 更新钱包余额
func (r *FleetRepository) UpdateWalletBalance(userID uint, newBalance float64) error {
	return global.DB.Model(&model.SystemWallet{}).Where("user_id = ?", userID).
		Update("balance", newBalance).Error
}

// CreateWalletTransaction 创建钱包流水
func (r *FleetRepository) CreateWalletTransaction(tx *model.WalletTransaction) error {
	return global.DB.Create(tx).Error
}

// ListWalletTransactions 查询用户钱包流水
func (r *FleetRepository) ListWalletTransactions(userID uint, page, pageSize int) ([]model.WalletTransaction, int64, error) {
	var txs []model.WalletTransaction
	var total int64
	offset := (page - 1) * pageSize

	db := global.DB.Model(&model.WalletTransaction{}).Where("user_id = ?", userID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&txs).Error; err != nil {
		return nil, 0, err
	}
	return txs, total, nil
}
