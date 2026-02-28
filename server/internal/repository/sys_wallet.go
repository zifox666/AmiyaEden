package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"

	"gorm.io/gorm"
)

// SysWalletRepository 系统钱包数据访问层
type SysWalletRepository struct{}

func NewSysWalletRepository() *SysWalletRepository {
	return &SysWalletRepository{}
}

// ─────────────────────────────────────────────
//  钱包 CRUD
// ─────────────────────────────────────────────

// GetOrCreateWallet 获取或创建用户钱包
func (r *SysWalletRepository) GetOrCreateWallet(userID uint) (*model.SystemWallet, error) {
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

// GetWalletByUserID 根据用户 ID 获取钱包（不自动创建）
func (r *SysWalletRepository) GetWalletByUserID(userID uint) (*model.SystemWallet, error) {
	var wallet model.SystemWallet
	err := global.DB.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// UpdateBalance 更新钱包余额
func (r *SysWalletRepository) UpdateBalance(userID uint, newBalance float64) error {
	return global.DB.Model(&model.SystemWallet{}).Where("user_id = ?", userID).
		Update("balance", newBalance).Error
}

// UpdateBalanceTx 在事务中更新钱包余额
func (r *SysWalletRepository) UpdateBalanceTx(tx *gorm.DB, userID uint, newBalance float64) error {
	return tx.Model(&model.SystemWallet{}).Where("user_id = ?", userID).
		Update("balance", newBalance).Error
}

// ─────────────────────────────────────────────
//  钱包流水
// ─────────────────────────────────────────────

// CreateTransaction 创建钱包流水
func (r *SysWalletRepository) CreateTransaction(tx *model.WalletTransaction) error {
	return global.DB.Create(tx).Error
}

// CreateTransactionTx 在事务中创建钱包流水
func (r *SysWalletRepository) CreateTransactionTx(dbTx *gorm.DB, tx *model.WalletTransaction) error {
	return dbTx.Create(tx).Error
}

// WalletTransactionFilter 流水查询筛选条件
type WalletTransactionFilter struct {
	UserID  *uint
	RefType string
}

// ListTransactions 分页查询钱包流水
func (r *SysWalletRepository) ListTransactions(page, pageSize int, filter WalletTransactionFilter) ([]model.WalletTransaction, int64, error) {
	var txs []model.WalletTransaction
	var total int64
	offset := (page - 1) * pageSize

	db := global.DB.Model(&model.WalletTransaction{})
	if filter.UserID != nil {
		db = db.Where("user_id = ?", *filter.UserID)
	}
	if filter.RefType != "" {
		db = db.Where("ref_type = ?", filter.RefType)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&txs).Error; err != nil {
		return nil, 0, err
	}
	return txs, total, nil
}

// ─────────────────────────────────────────────
//  操作日志
// ─────────────────────────────────────────────

// CreateLog 创建操作日志
func (r *SysWalletRepository) CreateLog(log *model.WalletLog) error {
	return global.DB.Create(log).Error
}

// CreateLogTx 在事务中创建操作日志
func (r *SysWalletRepository) CreateLogTx(dbTx *gorm.DB, log *model.WalletLog) error {
	return dbTx.Create(log).Error
}

// WalletLogFilter 日志查询筛选条件
type WalletLogFilter struct {
	OperatorID *uint
	TargetUID  *uint
	Action     string
}

// ListLogs 分页查询操作日志
func (r *SysWalletRepository) ListLogs(page, pageSize int, filter WalletLogFilter) ([]model.WalletLog, int64, error) {
	var logs []model.WalletLog
	var total int64
	offset := (page - 1) * pageSize

	db := global.DB.Model(&model.WalletLog{})
	if filter.OperatorID != nil {
		db = db.Where("operator_id = ?", *filter.OperatorID)
	}
	if filter.TargetUID != nil {
		db = db.Where("target_uid = ?", *filter.TargetUID)
	}
	if filter.Action != "" {
		db = db.Where("action = ?", filter.Action)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

// ─────────────────────────────────────────────
//  管理员：批量查询钱包
// ─────────────────────────────────────────────

// ListWallets 分页查询所有用户钱包
func (r *SysWalletRepository) ListWallets(page, pageSize int) ([]model.SystemWallet, int64, error) {
	var wallets []model.SystemWallet
	var total int64
	offset := (page - 1) * pageSize

	db := global.DB.Model(&model.SystemWallet{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&wallets).Error; err != nil {
		return nil, 0, err
	}
	return wallets, total, nil
}
