package handler

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/pkg/response"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBadgeHandlerGetBadgeCountsOmitsUnauthorizedFields(t *testing.T) {
	db := newBadgeHandlerTestDB(t)
	userID := seedBadgeHandlerTestData(t, db)

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	response := performBadgeHandlerRequest(t, userID, []string{model.RoleUser})
	if response.Code != 200 {
		t.Fatalf("expected success code, got %d", response.Code)
	}

	want := map[string]int64{"welfare_eligible": 1}
	if len(response.Data) != len(want) {
		t.Fatalf("expected only permitted badge fields, got %#v", response.Data)
	}
	for field, count := range want {
		if response.Data[field] != count {
			t.Fatalf("expected %s=%d, got %#v", field, count, response.Data)
		}
	}
	if _, exists := response.Data["srp_pending"]; exists {
		t.Fatalf("expected unauthorized srp_pending to be omitted, got %#v", response.Data)
	}
	if _, exists := response.Data["welfare_pending"]; exists {
		t.Fatalf("expected unauthorized welfare_pending to be omitted, got %#v", response.Data)
	}
	if _, exists := response.Data["order_pending"]; exists {
		t.Fatalf("expected unauthorized order_pending to be omitted, got %#v", response.Data)
	}
}

func TestBadgeHandlerGetBadgeCountsIncludesOrderPendingForWelfareRole(t *testing.T) {
	db := newBadgeHandlerTestDB(t)
	userID := seedBadgeHandlerTestData(t, db)

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	response := performBadgeHandlerRequest(t, userID, []string{model.RoleWelfare})
	if response.Code != 200 {
		t.Fatalf("expected success code, got %d", response.Code)
	}
	if response.Data["order_pending"] != 1 {
		t.Fatalf("expected welfare role to receive order_pending badge, got %#v", response.Data)
	}
	if response.Data["welfare_pending"] != 1 {
		t.Fatalf("expected welfare role to receive welfare_pending badge, got %#v", response.Data)
	}
	if response.Data["welfare_eligible"] != 1 {
		t.Fatalf("expected welfare role to receive welfare_eligible badge, got %#v", response.Data)
	}
	if _, exists := response.Data["srp_pending"]; exists {
		t.Fatalf("expected srp_pending to remain omitted for welfare role, got %#v", response.Data)
	}
}

func TestBadgeHandlerGetBadgeCountsIncludesMentorPendingApplicationsForMentorRole(t *testing.T) {
	db := newBadgeHandlerTestDB(t)
	userID := seedBadgeHandlerTestData(t, db)

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	response := performBadgeHandlerRequest(t, userID, []string{model.RoleMentor})
	if response.Code != 200 {
		t.Fatalf("expected success code, got %d", response.Code)
	}
	if response.Data["mentor_pending_applications"] != 1 {
		t.Fatalf("expected mentor role to receive mentor pending applications badge, got %#v", response.Data)
	}
}

func TestBadgeHandlerGetBadgeCountsReturnsSafeErrorMessage(t *testing.T) {
	db := newBadgeHandlerTestDB(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close sql db: %v", err)
	}

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	result := performBadgeHandlerRequest(t, 1, []string{model.RoleUser})
	if result.Code != response.CodeBizError {
		t.Fatalf("expected business error code, got %#v", result)
	}
	if result.Msg != "获取可申请福利数量失败" {
		t.Fatalf("expected transport-safe message, got %#v", result)
	}
}

type badgeHandlerResponse struct {
	Code int              `json:"code"`
	Msg  string           `json:"msg"`
	Data map[string]int64 `json:"data"`
}

func performBadgeHandlerRequest(t *testing.T, userID uint, roles []string) badgeHandlerResponse {
	t.Helper()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/badge-counts", nil)
	ctx.Set("userID", userID)
	ctx.Set("roles", roles)

	NewBadgeHandler().GetBadgeCounts(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected http status 200, got %d", recorder.Code)
	}

	var response badgeHandlerResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return response
}

func seedBadgeHandlerTestData(t *testing.T, db *gorm.DB) uint {
	t.Helper()

	user := model.User{Nickname: "Pilot One", QQ: "12345"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	welfare := model.Welfare{
		Name:      "Starter Pack",
		DistMode:  model.WelfareDistModePerUser,
		Status:    model.WelfareStatusActive,
		CreatedBy: user.ID,
	}
	if err := db.Create(&welfare).Error; err != nil {
		t.Fatalf("create welfare: %v", err)
	}

	requestedUserID := uint(999)
	if err := db.Create(&model.WelfareApplication{
		WelfareID:     welfare.ID,
		UserID:        &requestedUserID,
		CharacterID:   7001,
		CharacterName: "Other Pilot",
		Status:        model.WelfareAppStatusRequested,
	}).Error; err != nil {
		t.Fatalf("create welfare application: %v", err)
	}

	if err := db.Create(&model.SrpApplication{
		UserID:            user.ID,
		CharacterID:       5001,
		CharacterName:     "Pilot One",
		KillmailID:        8001,
		ShipTypeID:        100,
		SolarSystemID:     30000142,
		KillmailTime:      time.Unix(1_700_000_000, 0).UTC(),
		ReviewStatus:      model.SrpReviewSubmitted,
		PayoutStatus:      model.SrpPayoutNotPaid,
		FinalAmount:       10,
		RecommendedAmount: 10,
	}).Error; err != nil {
		t.Fatalf("create srp application: %v", err)
	}

	if err := db.Create(&model.ShopOrder{
		OrderNo:           "ORDER-1",
		UserID:            user.ID,
		MainCharacterName: "Pilot One",
		Nickname:          "Pilot One",
		ProductID:         1,
		ProductName:       "Item",
		ProductType:       model.ProductTypeNormal,
		Quantity:          1,
		UnitPrice:         1,
		TotalPrice:        1,
		Status:            model.OrderStatusRequested,
	}).Error; err != nil {
		t.Fatalf("create shop order: %v", err)
	}

	if err := db.Create(&model.MentorMenteeRelationship{
		MenteeUserID:                    user.ID + 100,
		MenteePrimaryCharacterIDAtStart: 7002,
		MentorUserID:                    user.ID,
		Status:                          model.MentorRelationStatusPending,
		AppliedAt:                       time.Unix(1_700_000_300, 0).UTC(),
	}).Error; err != nil {
		t.Fatalf("create mentor relationship: %v", err)
	}

	return user.ID
}

func newBadgeHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:badge_handler_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.EveCharacter{},
		&model.MentorMenteeRelationship{},
		&model.Welfare{},
		&model.WelfareSkillPlan{},
		&model.WelfareApplication{},
		&model.ShopOrder{},
		&model.SrpApplication{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}
