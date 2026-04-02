package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAdminDeliverOrderAttemptsInGameMailButIgnoresMailErrors(t *testing.T) {
	db := newShopServiceTestDB(t)
	useShopServiceTestDB(t, db)

	product := &model.ShopProduct{
		Name:      "Navy Omen",
		Price:     10,
		Stock:     -1,
		Type:      model.ProductTypeNormal,
		Status:    model.ProductStatusOnSale,
		SortOrder: 1,
	}
	if err := db.Create(product).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}

	order := &model.ShopOrder{
		OrderNo:           "ORDER001",
		UserID:            42,
		MainCharacterName: "Pilot One",
		Nickname:          "Pilot",
		ProductID:         product.ID,
		ProductName:       product.Name,
		ProductType:       product.Type,
		Quantity:          2,
		UnitPrice:         product.Price,
		TotalPrice:        product.Price * 2,
		Status:            model.OrderStatusRequested,
	}
	if err := db.Create(order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	svc := NewShopService()
	mailAttempted := false
	svc.orderDeliveryMailSender = func(ctx context.Context, operatorID uint, deliveredOrder *model.ShopOrder) error {
		mailAttempted = true
		if operatorID != 77 {
			t.Fatalf("operatorID = %d, want 77", operatorID)
		}
		if deliveredOrder.ID != order.ID {
			t.Fatalf("order id = %d, want %d", deliveredOrder.ID, order.ID)
		}
		return errors.New("mail failed")
	}

	deliveredOrder, mailWarning, err := svc.AdminDeliverOrder(order.ID, 77, "contract issued")
	if err != nil {
		t.Fatalf("AdminDeliverOrder() error = %v", err)
	}
	if !mailAttempted {
		t.Fatal("expected deliver to attempt in-game mail after successful delivery")
	}
	if !strings.Contains(mailWarning, "mail failed") {
		t.Fatalf("mailWarning = %q, want to contain %q", mailWarning, "mail failed")
	}
	if deliveredOrder.Status != model.OrderStatusDelivered {
		t.Fatalf("status = %q, want %q", deliveredOrder.Status, model.OrderStatusDelivered)
	}

	var updated model.ShopOrder
	if err := db.First(&updated, order.ID).Error; err != nil {
		t.Fatalf("reload order: %v", err)
	}
	if updated.Status != model.OrderStatusDelivered {
		t.Fatalf("status = %q, want %q", updated.Status, model.OrderStatusDelivered)
	}
	if updated.ReviewedBy == nil || *updated.ReviewedBy != 77 {
		t.Fatalf("reviewed_by = %v, want 77", updated.ReviewedBy)
	}
}

func TestBuildShopOrderDeliveryMailContentIncludesBilingualOfficerNotice(t *testing.T) {
	subject, body := buildShopOrderDeliveryMailContent("ORD-20260403", "Navy Omen", 2, "Amiya")

	if !strings.Contains(subject, "订单发放通知") || !strings.Contains(subject, "Order Delivery Notice") {
		t.Fatalf("unexpected subject: %q", subject)
	}
	if !strings.Contains(body, "你的订单已由 Amiya 发放") {
		t.Fatalf("expected Chinese body to mention order item and officer nickname, got %q", body)
	}
	if !strings.Contains(body, "订单编号：ORD-20260403") {
		t.Fatalf("expected Chinese body to include order number, got %q", body)
	}
	if !strings.Contains(body, "订单内容：Navy Omen") {
		t.Fatalf("expected Chinese body to include order item, got %q", body)
	}
	if !strings.Contains(body, "数量：2") {
		t.Fatalf("expected Chinese body to include quantity, got %q", body)
	}
	if !strings.Contains(body, "发放官员：Amiya") {
		t.Fatalf("expected Chinese body to include officer detail, got %q", body)
	}
	if !strings.Contains(body, "请检查你的钱包或合同") {
		t.Fatalf("expected Chinese body to mention wallet or contract, got %q", body)
	}
	if !strings.Contains(body, "感谢你的耐心等待。") {
		t.Fatalf("expected Chinese body to include a more professional tone, got %q", body)
	}
	if !strings.Contains(body, "Your shop order has been delivered by Amiya.") {
		t.Fatalf("expected English body to mention order item and officer nickname, got %q", body)
	}
	if !strings.Contains(body, "Order No: ORD-20260403") {
		t.Fatalf("expected English body to include order number, got %q", body)
	}
	if !strings.Contains(body, "Item: Navy Omen") {
		t.Fatalf("expected English body to include order item, got %q", body)
	}
	if !strings.Contains(body, "Quantity: 2") {
		t.Fatalf("expected English body to include quantity, got %q", body)
	}
	if !strings.Contains(body, "Delivered by: Amiya") {
		t.Fatalf("expected English body to include officer detail, got %q", body)
	}
	if !strings.Contains(body, "Please check your wallet or contract.") {
		t.Fatalf("expected English body to mention wallet or contract, got %q", body)
	}
	if !strings.Contains(body, "Thank you for your patience.") {
		t.Fatalf("expected English body to include a more professional tone, got %q", body)
	}
}

func newShopServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:shop_service_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.ShopProduct{},
		&model.ShopOrder{},
		&model.ShopRedeemCode{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func useShopServiceTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()

	previous := global.DB
	global.DB = db
	t.Cleanup(func() {
		global.DB = previous
	})
}

func TestBuildShopOrderResponsesIncludesReviewerNickname(t *testing.T) {
	reviewerID := uint(77)
	createdAt := time.Date(2026, time.April, 3, 8, 0, 0, 0, time.UTC)

	orders := []model.ShopOrder{
		{
			BaseModel:  model.BaseModel{ID: 1, CreatedAt: createdAt},
			OrderNo:    "ORDER-1",
			Status:     model.OrderStatusDelivered,
			ReviewedBy: &reviewerID,
		},
		{
			BaseModel: model.BaseModel{ID: 2, CreatedAt: createdAt},
			OrderNo:   "ORDER-2",
			Status:    model.OrderStatusRequested,
		},
	}

	got := buildShopOrderResponses(orders, map[uint]string{reviewerID: "Logistics Fox"})

	if len(got) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(got))
	}
	if got[0].ReviewerName != "Logistics Fox" {
		t.Fatalf("expected reviewer nickname to be included, got %q", got[0].ReviewerName)
	}
	if got[1].ReviewerName != "" {
		t.Fatalf("expected empty reviewer nickname for unreviewed order, got %q", got[1].ReviewerName)
	}
}
