package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"strings"
	"time"

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

// GetOrCreateWalletTx 在事务内获取或创建用户钱包
func (r *SysWalletRepository) GetOrCreateWalletTx(tx *gorm.DB, userID uint) (*model.SystemWallet, error) {
	var wallet model.SystemWallet
	err := tx.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		wallet = model.SystemWallet{UserID: userID, Balance: 0}
		if err := tx.Create(&wallet).Error; err != nil {
			return nil, err
		}
	}
	return &wallet, nil
}

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

// ExistsTransactionByRefID 检查是否已存在指定 RefID 的流水记录
func (r *SysWalletRepository) ExistsTransactionByRefID(refID string) (bool, error) {
	var count int64
	err := global.DB.Model(&model.WalletTransaction{}).Where("ref_id = ?", refID).Count(&count).Error
	return count > 0, err
}

// GetTransactionByUserRefTypeRefIDInRange 根据用户、流水类型、关联 ID 和时间范围获取单条钱包流水
func (r *SysWalletRepository) GetTransactionByUserRefTypeRefIDInRange(userID uint, refType, refID string, startAt, endAt time.Time) (*model.WalletTransaction, error) {
	var tx model.WalletTransaction
	err := global.DB.Where(
		"user_id = ? AND ref_type = ? AND ref_id = ? AND created_at >= ? AND created_at < ?",
		userID, refType, refID, startAt, endAt,
	).First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// CountTransactionsByUserRefTypeInRange 统计某用户在指定时间范围内的指定流水类型数量
func (r *SysWalletRepository) CountTransactionsByUserRefTypeInRange(userID uint, refType string, startAt, endAt time.Time) (int64, error) {
	var count int64
	err := global.DB.Model(&model.WalletTransaction{}).
		Where("user_id = ? AND ref_type = ? AND created_at >= ? AND created_at < ?", userID, refType, startAt, endAt).
		Count(&count).Error
	return count, err
}

// WalletTransactionFilter 流水查询筛选条件
type WalletTransactionFilter struct {
	UserID      *uint
	UserKeyword string
	RefType     string
}

func applyWalletTransactionUserFilter(db *gorm.DB, userIDColumn string, refTypeColumn string, filter WalletTransactionFilter) *gorm.DB {
	if filter.UserID != nil {
		db = db.Where(userIDColumn+" = ?", *filter.UserID)
	}
	if strings.TrimSpace(filter.UserKeyword) != "" {
		pattern := "%" + strings.ToLower(strings.TrimSpace(filter.UserKeyword)) + "%"
		db = db.Where(
			userIDColumn+` IN (
				SELECT DISTINCT u.id
				FROM "user" u
				LEFT JOIN eve_character ec ON ec.user_id = u.id
				WHERE LOWER(u.nickname) LIKE ? OR LOWER(ec.character_name) LIKE ?
			)`,
			pattern, pattern,
		)
	}
	if filter.RefType != "" {
		db = db.Where(refTypeColumn+" = ?", filter.RefType)
	}
	return db
}

// ListTransactions 分页查询钱包流水
func (r *SysWalletRepository) ListTransactions(page, pageSize int, filter WalletTransactionFilter) ([]model.WalletTransaction, int64, error) {
	var txs []model.WalletTransaction
	var total int64
	offset := (page - 1) * pageSize

	db := applyWalletTransactionUserFilter(global.DB.Model(&model.WalletTransaction{}), "user_id", "ref_type", filter)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&txs).Error; err != nil {
		return nil, 0, err
	}
	return txs, total, nil
}

// ListTransactionsWithCharacter 分页查询钱包流水（附带用户主人物名）
func (r *SysWalletRepository) ListTransactionsWithCharacter(page, pageSize int, filter WalletTransactionFilter) ([]model.TransactionWithCharacter, int64, error) {
	var results []model.TransactionWithCharacter
	var total int64
	offset := (page - 1) * pageSize

	countDB := applyWalletTransactionUserFilter(global.DB.Model(&model.WalletTransaction{}), "user_id", "ref_type", filter)
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	queryDB := global.DB.Table("wallet_transaction wt").
		Select(`wt.*,
			COALESCE(ec.character_name, '') AS character_name,
			COALESCE(NULLIF(u.nickname, ''), ec.character_name, '') AS nickname,
			CASE
				WHEN wt.operator_id = 0 THEN ''
				ELSE COALESCE(NULLIF(operator_u.nickname, ''), operator_ec.character_name, '')
			END AS operator_name`).
		Joins(`LEFT JOIN "user" u ON wt.user_id = u.id`).
		Joins("LEFT JOIN eve_character ec ON u.primary_character_id = ec.character_id").
		Joins(`LEFT JOIN "user" operator_u ON wt.operator_id = operator_u.id`).
		Joins("LEFT JOIN eve_character operator_ec ON operator_u.primary_character_id = operator_ec.character_id")
	queryDB = applyWalletTransactionUserFilter(queryDB, "wt.user_id", "wt.ref_type", filter)
	if err := queryDB.Order("wt.created_at DESC").Offset(offset).Limit(pageSize).Scan(&results).Error; err != nil {
		return nil, 0, err
	}
	return results, total, nil
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

// ListLogsWithCharacter 分页查询操作日志（附带操作人和目标用户主人物名）
func (r *SysWalletRepository) ListLogsWithCharacter(page, pageSize int, filter WalletLogFilter) ([]model.LogWithCharacter, int64, error) {
	var results []model.LogWithCharacter
	var total int64
	offset := (page - 1) * pageSize

	countDB := global.DB.Model(&model.WalletLog{})
	if filter.OperatorID != nil {
		countDB = countDB.Where("operator_id = ?", *filter.OperatorID)
	}
	if filter.TargetUID != nil {
		countDB = countDB.Where("target_uid = ?", *filter.TargetUID)
	}
	if filter.Action != "" {
		countDB = countDB.Where("action = ?", filter.Action)
	}
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	queryDB := global.DB.Table("wallet_log wl").
		Select(`wl.*,
			COALESCE(t_ec.character_name, '') AS target_character_name,
			COALESCE(o_ec.character_name, '') AS operator_character_name`).
		Joins(`LEFT JOIN "user" t_u ON wl.target_uid = t_u.id`).
		Joins("LEFT JOIN eve_character t_ec ON t_u.primary_character_id = t_ec.character_id").
		Joins(`LEFT JOIN "user" o_u ON wl.operator_id = o_u.id`).
		Joins("LEFT JOIN eve_character o_ec ON o_u.primary_character_id = o_ec.character_id")
	if filter.OperatorID != nil {
		queryDB = queryDB.Where("wl.operator_id = ?", *filter.OperatorID)
	}
	if filter.TargetUID != nil {
		queryDB = queryDB.Where("wl.target_uid = ?", *filter.TargetUID)
	}
	if filter.Action != "" {
		queryDB = queryDB.Where("wl.action = ?", filter.Action)
	}
	if err := queryDB.Order("wl.created_at DESC").Offset(offset).Limit(pageSize).Scan(&results).Error; err != nil {
		return nil, 0, err
	}
	return results, total, nil
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

// ListWalletsWithCharacter 分页查询所有用户钱包（附带主人物名）
func (r *SysWalletRepository) ListWalletsWithCharacter(page, pageSize int) ([]model.WalletWithCharacter, int64, error) {
	var results []model.WalletWithCharacter
	var total int64
	offset := (page - 1) * pageSize

	db := global.DB.Model(&model.SystemWallet{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := global.DB.Table("system_wallet sw").
		Select("sw.*, COALESCE(ec.character_name, '') AS character_name").
		Joins(`LEFT JOIN "user" u ON sw.user_id = u.id`).
		Joins("LEFT JOIN eve_character ec ON u.primary_character_id = ec.character_id").
		Order("sw.updated_at DESC").
		Offset(offset).Limit(pageSize).
		Scan(&results).Error
	if err != nil {
		return nil, 0, err
	}
	return results, total, nil
}
