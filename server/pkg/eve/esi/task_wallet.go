package esi

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"

	"context"
	"fmt"
	"time"
)

func init() {
	Register(&WalletTask{})
}

// WalletTask 角色钱包刷新任务
type WalletTask struct{}

func (t *WalletTask) Name() string        { return "character_wallet" }
func (t *WalletTask) Description() string { return "角色钱包信息" }
func (t *WalletTask) Priority() Priority  { return PriorityNormal }

func (t *WalletTask) Interval() RefreshInterval {
	return RefreshInterval{
		Active:   12 * time.Hour,
		Inactive: 7 * 24 * time.Hour,
	}
}

func (t *WalletTask) RequiredScopes() []TaskScope {
	return []TaskScope{
		{Scope: "esi-wallet.read_character_wallet.v1", Description: "读取角色钱包信息"},
	}
}

// WalletJournalEntry 单条钱包交易记录
type WalletJournalEntry struct {
	Amount        float64   `json:"amount"`
	Balance       float64   `json:"balance"`
	ContextID     int64     `json:"context_id"`
	ContextIDType string    `json:"context_id_type"`
	Date          time.Time `json:"date"`
	Description   string    `json:"description"`
	FirstPartyID  int64     `json:"first_party_id"`
	ID            int64     `json:"id"`
	Reason        string    `json:"reason"`
	RefType       string    `json:"ref_type"`
	SecondPartyID int64     `json:"second_party_id"`
	Tax           float64   `json:"tax"`
	TaxReceiverID int64     `json:"tax_receiver_id"`
}

// WalletJournalResult 钱包交易记录查询结果
type WalletTransaction struct {
	ClientID      int64     `json:"client_id"`
	Date          time.Time `json:"date"`
	IsBuy         bool      `json:"is_buy"`
	IsPersonal    bool      `json:"is_personal"`
	JournalRefID  int64     `json:"journal_ref_id"`
	LocationID    int64     `json:"location_id"`
	Quantity      int       `json:"quantity"`
	TransactionID int64     `json:"transaction_id"`
	TypeID        int       `json:"type_id"`
	UnitPrice     float64   `json:"unit_price"`
}

// WalletJournalResult 钱包交易记录查询结果
type WalletJournalResult []WalletJournalEntry

// Execute 执行钱包数据刷新
func (t *WalletTask) Execute(ctx *TaskContext) error {
	bgCtx := context.Background()

	// 1. 获取钱包余额
	var balance float64
	path := fmt.Sprintf("/characters/%d/wallet/", ctx.CharacterID)
	if err := ctx.Client.Get(bgCtx, path, ctx.AccessToken, &balance); err != nil {
		return fmt.Errorf("fetch wallet balance: %w", err)
	}

	// 2. 获取钱包记录
	var walletJournal WalletJournalResult
	path = fmt.Sprintf("/characters/%d/wallet/journal", ctx.CharacterID)
	if _, err := ctx.Client.GetPaginated(bgCtx, path, ctx.AccessToken, &walletJournal); err != nil {
		return fmt.Errorf("fetch wallet journal: %w", err)
	}

	// 3. 获取钱包市场交易
	var walletTransactions []WalletTransaction
	path = fmt.Sprintf("/characters/%d/wallet/transactions", ctx.CharacterID)
	if _, err := ctx.Client.GetPaginated(bgCtx, path, ctx.AccessToken, &walletTransactions); err != nil {
		return fmt.Errorf("fetch wallet transactions: %w", err)
	}

	// 入库
	tx := global.DB.Begin()
	var count int64
	global.DB.Model(&model.EVECharacterWallet{}).Where("character_id = ?", ctx.CharacterID).Count(&count)
	if count == 0 {
		tx.Create(&model.EVECharacterWallet{
			CharacterID: ctx.CharacterID,
			Balance:     balance, // 直接使用 float64
		})
	} else {
		// 更新余额
		tx.Model(&model.EVECharacterWallet{}).
			Where("character_id = ?", ctx.CharacterID).
			Update("balance", balance)
	}

	var waitingEntries []model.EVECharacterWalletJournal
	for _, entry := range walletJournal {
		global.DB.Model(&model.EVECharacterWalletJournal{}).Where("ID = ?", entry.ID).Count(&count)
		if count == 0 {
			waitingEntries = append(waitingEntries, model.EVECharacterWalletJournal{
				ID:            entry.ID,
				CharacterID:   ctx.CharacterID,
				Amount:        entry.Amount,
				Balance:       entry.Balance,
				ContextID:     entry.ContextID,
				ContextIDType: entry.ContextIDType,
				Date:          entry.Date,
				Description:   entry.Description,
				FirstPartyID:  entry.FirstPartyID,
				Reason:        entry.Reason,
				RefType:       entry.RefType,
				SecondPartyID: entry.SecondPartyID,
				Tax:           entry.Tax,
				TaxReceiverID: entry.TaxReceiverID,
			})
		}
	}
	if len(waitingEntries) > 0 {
		if err := tx.Create(&waitingEntries).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("insert wallet journal: %w", err)
		}
	}

	var waitingTransactions []model.EVECharacterWalletTransaction
	for _, t := range walletTransactions {
		global.DB.Model(&model.EVECharacterWalletTransaction{}).Where("transaction_id = ?", t.TransactionID).Count(&count)
		if count == 0 {
			waitingTransactions = append(waitingTransactions, model.EVECharacterWalletTransaction{
				TransactionID: t.TransactionID,
				CharacterID:   ctx.CharacterID,
				ClientID:      t.ClientID,
				Date:          t.Date,
				IsBuy:         t.IsBuy,
				IsPersonal:    t.IsPersonal,
				JournalRefID:  t.JournalRefID,
				LocationID:    t.LocationID,
				Quantity:      t.Quantity,
				TypeID:        t.TypeID,
				UnitPrice:     t.UnitPrice,
			})
		}
	}
	if len(waitingTransactions) > 0 {
		if err := tx.Create(&waitingTransactions).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("insert wallet transactions: %w", err)
		}
	}
	tx.Commit()

	return nil
}
