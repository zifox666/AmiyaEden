package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"fmt"
	"reflect"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBadgeServiceGetBadgeCountsReturnsOnlyPermittedNonZeroFields(t *testing.T) {
	db := newBadgeServiceTestDB(t)
	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

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
		t.Fatalf("create submitted srp application: %v", err)
	}
	if err := db.Create(&model.SrpApplication{
		UserID:            user.ID,
		CharacterID:       5002,
		CharacterName:     "Pilot One Alt",
		KillmailID:        8002,
		ShipTypeID:        101,
		SolarSystemID:     30000142,
		KillmailTime:      time.Unix(1_700_000_100, 0).UTC(),
		ReviewStatus:      model.SrpReviewApproved,
		PayoutStatus:      model.SrpPayoutNotPaid,
		FinalAmount:       20,
		RecommendedAmount: 20,
	}).Error; err != nil {
		t.Fatalf("create approved srp application: %v", err)
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

	svc := NewBadgeService()
	tests := []struct {
		name  string
		roles []string
		want  BadgeCounts
	}{
		{
			name:  "ordinary user only sees welfare eligible count",
			roles: []string{model.RoleUser},
			want: BadgeCounts{
				BadgeCountWelfareEligible: 1,
			},
		},
		{
			name:  "srp reviewer sees srp pending count",
			roles: []string{model.RoleSRP},
			want: BadgeCounts{
				BadgeCountWelfareEligible: 1,
				BadgeCountSrpPending:      2,
			},
		},
		{
			name:  "welfare officer sees welfare and order pending counts",
			roles: []string{model.RoleWelfare},
			want: BadgeCounts{
				BadgeCountWelfareEligible: 1,
				BadgeCountWelfarePending:  1,
				BadgeCountOrderPending:    1,
			},
		},
		{
			name:  "admin sees every non zero count",
			roles: []string{model.RoleAdmin},
			want: BadgeCounts{
				BadgeCountWelfareEligible: 1,
				BadgeCountSrpPending:      2,
				BadgeCountWelfarePending:  1,
				BadgeCountOrderPending:    1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.GetBadgeCounts(user.ID, tt.roles)
			if err != nil {
				t.Fatalf("GetBadgeCounts() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("GetBadgeCounts() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestBadgeServiceGetBadgeCountsCountsPerCharacterWelfareOnce(t *testing.T) {
	db := newBadgeServiceTestDB(t)
	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	user := model.User{Nickname: "Pilot Two", QQ: "67890"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	characters := []model.EveCharacter{
		{CharacterID: 9001, CharacterName: "Alpha", UserID: user.ID},
		{CharacterID: 9002, CharacterName: "Beta", UserID: user.ID},
	}
	if err := db.Create(&characters).Error; err != nil {
		t.Fatalf("create characters: %v", err)
	}

	if err := db.Create(&model.Welfare{
		Name:      "Per Character Pack",
		DistMode:  model.WelfareDistModePerCharacter,
		Status:    model.WelfareStatusActive,
		CreatedBy: user.ID,
	}).Error; err != nil {
		t.Fatalf("create welfare: %v", err)
	}

	svc := NewBadgeService()
	got, err := svc.GetBadgeCounts(user.ID, []string{model.RoleUser})
	if err != nil {
		t.Fatalf("GetBadgeCounts() error = %v", err)
	}

	want := BadgeCounts{BadgeCountWelfareEligible: 1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetBadgeCounts() = %#v, want %#v", got, want)
	}
}

func TestBadgeServiceGetBadgeCountsOmitsZeroCounts(t *testing.T) {
	db := newBadgeServiceTestDB(t)
	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	user := model.User{Nickname: "Pilot Zero", QQ: "11111"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	svc := NewBadgeService()
	got, err := svc.GetBadgeCounts(user.ID, []string{model.RoleUser})
	if err != nil {
		t.Fatalf("GetBadgeCounts() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected zero counts to be omitted, got %#v", got)
	}
}

func newBadgeServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:badge_service_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.EveCharacter{},
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
