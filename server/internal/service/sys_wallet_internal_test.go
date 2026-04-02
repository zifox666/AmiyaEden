package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreditUserStoresSystemOperatorOnWalletTransaction(t *testing.T) {
	db := newSysWalletServiceTestDB(t)
	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	svc := NewSysWalletService()
	if err := svc.CreditUser(42, 15.5, "shop order", "shop", "order:1"); err != nil {
		t.Fatalf("CreditUser() error = %v", err)
	}

	var txs []model.WalletTransaction
	if err := db.Order("id ASC").Find(&txs).Error; err != nil {
		t.Fatalf("load wallet transactions: %v", err)
	}
	if len(txs) != 1 {
		t.Fatalf("wallet transaction count = %d, want 1", len(txs))
	}

	tx := txs[0]
	if tx.UserID != 42 {
		t.Fatalf("wallet transaction user_id = %d, want 42", tx.UserID)
	}
	if tx.Amount != 15.5 {
		t.Fatalf("wallet transaction amount = %f, want 15.5", tx.Amount)
	}
	if tx.BalanceAfter != 15.5 {
		t.Fatalf("wallet transaction balance_after = %f, want 15.5", tx.BalanceAfter)
	}
	if tx.OperatorID != 0 {
		t.Fatalf("wallet transaction operator_id = %d, want 0", tx.OperatorID)
	}
	if tx.RefType != "shop" || tx.RefID != "order:1" || tx.Reason != "shop order" {
		t.Fatalf("unexpected transaction metadata: %+v", tx)
	}
}

func TestCreditUserTruncatesOverlongWalletReason(t *testing.T) {
	db := newSysWalletServiceTestDB(t)
	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	overlongReason := strings.Repeat("伏", walletTransactionReasonMaxLength+16)
	svc := NewSysWalletService()
	if err := svc.CreditUser(42, 10, overlongReason, "shop", "order:2"); err != nil {
		t.Fatalf("CreditUser() error = %v", err)
	}

	var tx model.WalletTransaction
	if err := db.First(&tx).Error; err != nil {
		t.Fatalf("load wallet transaction: %v", err)
	}
	if got := len([]rune(tx.Reason)); got != walletTransactionReasonMaxLength {
		t.Fatalf("wallet transaction reason length = %d, want %d", got, walletTransactionReasonMaxLength)
	}
}

func TestNormalizeLedgerPageSizeUsesLedgerStandardBounds(t *testing.T) {
	tests := []struct {
		name string
		size int
		want int
	}{
		{name: "defaults when zero", size: 0, want: 200},
		{name: "preserves smaller valid page", size: 20, want: 20},
		{name: "keeps ledger default", size: 200, want: 200},
		{name: "allows larger ledger page", size: 500, want: 500},
		{name: "caps at thousand", size: 5000, want: 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeLedgerPageSize(tt.size); got != tt.want {
				t.Fatalf("normalizeLedgerPageSize(%d) = %d, want %d", tt.size, got, tt.want)
			}
		})
	}
}

func TestGetMyTransactionsIncludesOperatorName(t *testing.T) {
	db := newSysWalletTransactionListTestDB(t)
	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	operatorCharacterID := int64(90000077)
	if err := db.Create(&model.User{
		BaseModel:          model.BaseModel{ID: 77},
		Nickname:           "Officer Fox",
		PrimaryCharacterID: operatorCharacterID,
	}).Error; err != nil {
		t.Fatalf("create operator user: %v", err)
	}
	if err := db.Create(&model.EveCharacter{
		CharacterID:   operatorCharacterID,
		CharacterName: "Operator Main",
		UserID:        77,
	}).Error; err != nil {
		t.Fatalf("create operator character: %v", err)
	}
	if err := db.Create(&model.WalletTransaction{
		UserID:       42,
		Amount:       12.5,
		Reason:       "manual payout",
		RefType:      model.WalletRefManual,
		RefID:        "manual:1",
		BalanceAfter: 88.8,
		OperatorID:   77,
	}).Error; err != nil {
		t.Fatalf("create wallet transaction: %v", err)
	}

	svc := NewSysWalletService()
	records, total, err := svc.GetMyTransactions(42, 1, 20)
	if err != nil {
		t.Fatalf("GetMyTransactions() error = %v", err)
	}
	if total != 1 {
		t.Fatalf("transaction total = %d, want 1", total)
	}
	if len(records) != 1 {
		t.Fatalf("transaction count = %d, want 1", len(records))
	}
	if records[0].OperatorName != "Officer Fox" {
		t.Fatalf("operator_name = %q, want %q", records[0].OperatorName, "Officer Fox")
	}
}

func newSysWalletServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:sys_wallet_service_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.SystemWallet{}, &model.WalletTransaction{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func newSysWalletTransactionListTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:sys_wallet_tx_list_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.EveCharacter{}, &model.WalletTransaction{}); err != nil {
		t.Fatalf("auto migrate transaction list db: %v", err)
	}
	return db
}
