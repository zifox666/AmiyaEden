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

func TestImpersonateUserRejectsInvalidPrimaryCharacter(t *testing.T) {
	db := newUserServiceTestDB(t)
	seedImpersonationTargetUser(t, db)

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	_, _, err := NewUserService().ImpersonateUser(1)
	if err == nil || !strings.Contains(err.Error(), "无法模拟登录") {
		t.Fatalf("expected impersonation restriction error, got %v", err)
	}
}

func newUserServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:user_service_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.EveCharacter{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func seedImpersonationTargetUser(t *testing.T, db *gorm.DB) {
	t.Helper()

	user := model.User{
		BaseModel:          model.BaseModel{ID: 1},
		Nickname:           "Amiya",
		Role:               model.RoleUser,
		PrimaryCharacterID: 9001,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	if err := db.Create(&model.EveCharacter{
		CharacterID:   9001,
		CharacterName: "Amiya Prime",
		PortraitURL:   "portrait.png",
		UserID:        user.ID,
		TokenInvalid:  true,
	}).Error; err != nil {
		t.Fatalf("create character: %v", err)
	}
}
