package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newHallOfFameRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:hall_of_fame_repository_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.HallOfFameCard{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	return db
}

func TestBatchUpdateLayoutFailsWhenACardIsMissing(t *testing.T) {
	db := newHallOfFameRepositoryTestDB(t)
	oldDB := global.DB
	global.DB = db
	t.Cleanup(func() { global.DB = oldDB })

	card := model.HallOfFameCard{
		Name:        "Hero Alpha",
		Width:       220,
		StylePreset: "gold",
		Visible:     true,
	}
	if err := db.Create(&card).Error; err != nil {
		t.Fatalf("create card: %v", err)
	}

	repo := NewHallOfFameRepository()
	err := repo.BatchUpdateLayout([]model.CardLayoutUpdate{
		{ID: card.ID, PosX: 10, PosY: 20, Width: 220, Height: 0, ZIndex: 1},
		{ID: card.ID + 999, PosX: 30, PosY: 40, Width: 260, Height: 0, ZIndex: 2},
	})
	if err == nil {
		t.Fatal("expected missing card update to fail")
	}
}

func TestUpdateCardFieldsFailsWhenCardIsMissing(t *testing.T) {
	db := newHallOfFameRepositoryTestDB(t)
	oldDB := global.DB
	global.DB = db
	t.Cleanup(func() { global.DB = oldDB })

	repo := NewHallOfFameRepository()
	err := repo.UpdateCardFields(999, map[string]interface{}{"title": "Founder"})
	if err == nil {
		t.Fatal("expected missing card field update to fail")
	}
}
