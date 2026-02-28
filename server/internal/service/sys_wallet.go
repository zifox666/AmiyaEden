package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"errors"
	"fmt"
)

// SysWalletService 系统钱包业务逻辑层
type SysWalletService struct {
	repo *repository.SysWalletRepository
}

func NewSysWalletService() *SysWalletService {
	return &SysWalletService{
		repo: repository.NewSysWalletRepository(),
	}
}

// ─────────────────────────────────────────────
//  用户端
// ─────────────────────────────────────────────

// GetMyWallet 获取当前用户钱包
func (s *SysWalletService) GetMyWallet(userID uint) (*model.SystemWallet, error) {
	return s.repo.GetOrCreateWallet(userID)
}

// GetMyTransactions 获取当前用户流水
func (s *SysWalletService) GetMyTransactions(userID uint, page, pageSize int) ([]model.WalletTransaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	filter := repository.WalletTransactionFilter{UserID: &userID}
	return s.repo.ListTransactions(page, pageSize, filter)
}

// ─────────────────────────────────────────────
//  管理员端
// ─────────────────────────────────────────────

// AdminAdjustRequest 管理员调整钱包请求
type AdminAdjustRequest struct {
	TargetUID uint    `json:"target_uid" binding:"required"` // 目标用户 ID
	Action    string  `json:"action" binding:"required,oneof=add deduct set"`
	Amount    float64 `json:"amount" binding:"required,gt=0"` // 操作金额（必须正数）
	Reason    string  `json:"reason" binding:"required"`      // 操作原因
}

// AdminAdjust 管理员调整用户钱包余额
func (s *SysWalletService) AdminAdjust(operatorID uint, req *AdminAdjustRequest) (*model.SystemWallet, error) {
	// 获取或创建目标用户钱包
	wallet, err := s.repo.GetOrCreateWallet(req.TargetUID)
	if err != nil {
		return nil, fmt.Errorf("获取用户钱包失败: %w", err)
	}

	oldBalance := wallet.Balance
	var newBalance float64

	switch req.Action {
	case model.WalletActionAdd:
		newBalance = oldBalance + req.Amount
	case model.WalletActionDeduct:
		newBalance = oldBalance - req.Amount
		if newBalance < 0 {
			return nil, errors.New("余额不足，无法扣减")
		}
	case model.WalletActionSet:
		newBalance = req.Amount
	default:
		return nil, errors.New("无效的操作类型")
	}

	// 事务：更新余额 + 写流水 + 写日志
	tx := global.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 更新余额
	if err := s.repo.UpdateBalanceTx(tx, req.TargetUID, newBalance); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新余额失败: %w", err)
	}

	// 2. 写流水
	var txAmount float64
	switch req.Action {
	case model.WalletActionAdd:
		txAmount = req.Amount
	case model.WalletActionDeduct:
		txAmount = -req.Amount
	case model.WalletActionSet:
		txAmount = newBalance - oldBalance
	}

	walletTx := &model.WalletTransaction{
		UserID:       req.TargetUID,
		Amount:       txAmount,
		Reason:       req.Reason,
		RefType:      model.WalletRefAdminAdjust,
		RefID:        fmt.Sprintf("admin:%d", operatorID),
		BalanceAfter: newBalance,
		OperatorID:   operatorID,
	}
	if err := s.repo.CreateTransactionTx(tx, walletTx); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("写入流水失败: %w", err)
	}

	// 3. 写操作日志
	log := &model.WalletLog{
		OperatorID: operatorID,
		TargetUID:  req.TargetUID,
		Action:     req.Action,
		Amount:     req.Amount,
		Before:     oldBalance,
		After:      newBalance,
		Reason:     req.Reason,
	}
	if err := s.repo.CreateLogTx(tx, log); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("写入操作日志失败: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	wallet.Balance = newBalance
	return wallet, nil
}

// AdminListWallets 管理员查询所有钱包
func (s *SysWalletService) AdminListWallets(page, pageSize int) ([]model.SystemWallet, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListWallets(page, pageSize)
}

// AdminGetWallet 管理员查看指定用户钱包
func (s *SysWalletService) AdminGetWallet(userID uint) (*model.SystemWallet, error) {
	return s.repo.GetOrCreateWallet(userID)
}

// AdminListTransactions 管理员查询流水（可按用户/类型筛选）
func (s *SysWalletService) AdminListTransactions(page, pageSize int, filter repository.WalletTransactionFilter) ([]model.WalletTransaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListTransactions(page, pageSize, filter)
}

// AdminListLogs 管理员查询操作日志
func (s *SysWalletService) AdminListLogs(page, pageSize int, filter repository.WalletLogFilter) ([]model.WalletLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListLogs(page, pageSize, filter)
}

// ─────────────────────────────────────────────
//  内部调用（供其他 Service 调用）
// ─────────────────────────────────────────────

// CreditUser 给用户加钱（内部调用，如 PAP 奖励、SRP 发放等）
func (s *SysWalletService) CreditUser(userID uint, amount float64, reason, refType, refID string) error {
	if amount <= 0 {
		return errors.New("金额必须大于 0")
	}

	wallet, err := s.repo.GetOrCreateWallet(userID)
	if err != nil {
		return fmt.Errorf("获取用户钱包失败: %w", err)
	}

	newBalance := wallet.Balance + amount

	tx := global.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.repo.UpdateBalanceTx(tx, userID, newBalance); err != nil {
		tx.Rollback()
		return err
	}

	walletTx := &model.WalletTransaction{
		UserID:       userID,
		Amount:       amount,
		Reason:       reason,
		RefType:      refType,
		RefID:        refID,
		BalanceAfter: newBalance,
		OperatorID:   0, // 系统操作
	}
	if err := s.repo.CreateTransactionTx(tx, walletTx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// DebitUser 扣减用户余额（内部调用，如商城购买）
func (s *SysWalletService) DebitUser(userID uint, amount float64, reason, refType, refID string) error {
	if amount <= 0 {
		return errors.New("金额必须大于 0")
	}

	wallet, err := s.repo.GetOrCreateWallet(userID)
	if err != nil {
		return fmt.Errorf("获取用户钱包失败: %w", err)
	}

	if wallet.Balance < amount {
		return errors.New("余额不足")
	}

	newBalance := wallet.Balance - amount

	tx := global.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.repo.UpdateBalanceTx(tx, userID, newBalance); err != nil {
		tx.Rollback()
		return err
	}

	walletTx := &model.WalletTransaction{
		UserID:       userID,
		Amount:       -amount,
		Reason:       reason,
		RefType:      refType,
		RefID:        refID,
		BalanceAfter: newBalance,
		OperatorID:   0, // 系统操作
	}
	if err := s.repo.CreateTransactionTx(tx, walletTx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
