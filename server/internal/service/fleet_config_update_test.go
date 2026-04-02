package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUpdateFleetConfigPreservesItemSettingsForExistingFitting(t *testing.T) {
	db := newFleetConfigServiceTestDB(t)

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	svc := &FleetConfigService{
		repo:    repository.NewFleetConfigRepository(),
		sdeRepo: repository.NewSdeRepository(),
		http:    &http.Client{Timeout: 30 * time.Second},
	}
	eft := "[TypeID:1001, Test Fit]\nTypeID:2001\n\nTypeID:2002"

	created, err := svc.CreateFleetConfig(42, &CreateFleetConfigRequest{
		Name:        "Shield Doctrine",
		Description: "test",
		Fittings: []FleetConfigFittingReq{{
			FittingName: "Test Fit",
			EFT:         eft,
			SrpAmount:   12_500_000,
		}},
	})
	if err != nil {
		t.Fatalf("create fleet config: %v", err)
	}
	if len(created.Fittings) != 1 {
		t.Fatalf("expected 1 fitting, got %d", len(created.Fittings))
	}

	fittingID := created.Fittings[0].ID
	items, err := svc.repo.ListItemsByFittingIDs([]uint{fittingID})
	if err != nil {
		t.Fatalf("list fitting items: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 fitting items, got %d", len(items))
	}

	optionalItem := mustFindFleetConfigItem(t, items, 2001, "LoSlot0")
	replaceableItem := mustFindFleetConfigItem(t, items, 2002, "MedSlot0")

	err = svc.UpdateFittingItemsSettings(created.ID, fittingID, 42, []string{model.RoleAdmin}, &UpdateFittingItemsSettingsRequest{
		Items: []UpdateItemSettingsReq{
			{
				ID:                 optionalItem.ID,
				Importance:         model.FittingItemOptional,
				Penalty:            model.FittingPenaltyHalf,
				ReplacementPenalty: model.FittingPenaltyNone,
			},
			{
				ID:                 replaceableItem.ID,
				Importance:         model.FittingItemReplaceable,
				Penalty:            model.FittingPenaltyHalf,
				ReplacementPenalty: model.FittingPenaltyHalf,
				Replacements:       []int64{3001},
			},
		},
	})
	if err != nil {
		t.Fatalf("update fitting item settings: %v", err)
	}

	payload, err := json.Marshal(map[string]any{
		"name":        "Shield Doctrine",
		"description": "test",
		"fittings": []map[string]any{{
			"id":           fittingID,
			"fitting_name": "Test Fit",
			"eft":          eft,
			"srp_amount":   12_500_000,
		}},
	})
	if err != nil {
		t.Fatalf("marshal update payload: %v", err)
	}

	var updateReq UpdateFleetConfigRequest
	if err := json.Unmarshal(payload, &updateReq); err != nil {
		t.Fatalf("unmarshal update payload: %v", err)
	}

	if _, err := svc.UpdateFleetConfig(created.ID, 42, []string{model.RoleAdmin}, &updateReq); err != nil {
		t.Fatalf("update fleet config: %v", err)
	}

	updatedFittings, err := svc.repo.ListFittingsByConfigID(created.ID)
	if err != nil {
		t.Fatalf("list updated fittings: %v", err)
	}
	if len(updatedFittings) != 1 {
		t.Fatalf("expected 1 updated fitting, got %d", len(updatedFittings))
	}

	updatedItems, err := svc.repo.ListItemsByFittingIDs([]uint{updatedFittings[0].ID})
	if err != nil {
		t.Fatalf("list updated items: %v", err)
	}
	updatedOptional := mustFindFleetConfigItem(t, updatedItems, 2001, "LoSlot0")
	updatedReplaceable := mustFindFleetConfigItem(t, updatedItems, 2002, "MedSlot0")

	if updatedOptional.Importance != model.FittingItemOptional {
		t.Fatalf("optional item importance = %q, want %q", updatedOptional.Importance, model.FittingItemOptional)
	}
	if updatedOptional.Penalty != model.FittingPenaltyHalf {
		t.Fatalf("optional item penalty = %q, want %q", updatedOptional.Penalty, model.FittingPenaltyHalf)
	}
	if updatedReplaceable.Importance != model.FittingItemReplaceable {
		t.Fatalf("replaceable item importance = %q, want %q", updatedReplaceable.Importance, model.FittingItemReplaceable)
	}
	if updatedReplaceable.ReplacementPenalty != model.FittingPenaltyHalf {
		t.Fatalf(
			"replaceable item replacement penalty = %q, want %q",
			updatedReplaceable.ReplacementPenalty,
			model.FittingPenaltyHalf,
		)
	}

	replacements, err := svc.repo.ListReplacementsByItemIDs([]uint{updatedReplaceable.ID})
	if err != nil {
		t.Fatalf("list updated replacements: %v", err)
	}
	if len(replacements) != 1 {
		t.Fatalf("expected 1 replacement, got %d", len(replacements))
	}
	if replacements[0].TypeID != 3001 {
		t.Fatalf("replacement type_id = %d, want %d", replacements[0].TypeID, 3001)
	}
}

func TestParseEFTToFittingRejectsUnknownItems(t *testing.T) {
	db := newFleetConfigServiceTestDB(t)

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	svc := &FleetConfigService{
		repo:    repository.NewFleetConfigRepository(),
		sdeRepo: repository.NewSdeRepository(),
		http:    &http.Client{Timeout: 30 * time.Second},
	}

	_, _, err := svc.parseEFTToFitting("[TypeID:1001, Test Fit]\nMissing Module")
	if err == nil {
		t.Fatal("expected parseEFTToFitting to reject unresolved module names")
	}
	if !strings.Contains(err.Error(), "未找到装备类型") {
		t.Fatalf("unexpected parse error: %v", err)
	}
}

func TestUpdateFleetConfigPreservesFlagsAcrossBuiltEFTRoundTrip(t *testing.T) {
	db := newFleetConfigServiceTestDB(t)

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	svc := &FleetConfigService{
		repo:    repository.NewFleetConfigRepository(),
		sdeRepo: repository.NewSdeRepository(),
		http:    &http.Client{Timeout: 30 * time.Second},
	}

	config := model.FleetConfig{Name: "Shield Doctrine", Description: "test", CreatedBy: 42}
	if err := db.Create(&config).Error; err != nil {
		t.Fatalf("create config: %v", err)
	}

	fitting := model.FleetConfigFitting{
		FleetConfigID: config.ID,
		ShipTypeID:    1001,
		FittingName:   "Test Fit",
		SrpAmount:     12_500_000,
	}
	if err := db.Create(&fitting).Error; err != nil {
		t.Fatalf("create fitting: %v", err)
	}

	items := []model.FleetConfigFittingItem{
		{
			FleetConfigFittingID: fitting.ID,
			TypeID:               2001,
			Quantity:             1,
			Flag:                 "LoSlot2",
			Importance:           model.FittingItemOptional,
			Penalty:              model.FittingPenaltyHalf,
			ReplacementPenalty:   model.FittingPenaltyNone,
		},
		{
			FleetConfigFittingID: fitting.ID,
			TypeID:               2002,
			Quantity:             1,
			Flag:                 "MedSlot0",
			Importance:           model.FittingItemReplaceable,
			Penalty:              model.FittingPenaltyHalf,
			ReplacementPenalty:   model.FittingPenaltyHalf,
		},
		{
			FleetConfigFittingID: fitting.ID,
			TypeID:               3001,
			Quantity:             1,
			Flag:                 "Cargo",
			Importance:           model.FittingItemRequired,
			Penalty:              model.FittingPenaltyNone,
			ReplacementPenalty:   model.FittingPenaltyNone,
		},
	}
	if err := db.Create(&items).Error; err != nil {
		t.Fatalf("create items: %v", err)
	}
	if err := db.Create(&model.FleetConfigFittingItemReplacement{
		FleetConfigFittingItemID: items[1].ID,
		TypeID:                   4001,
	}).Error; err != nil {
		t.Fatalf("create replacement: %v", err)
	}

	eft := buildEFT("TypeID:1001", fitting.FittingName, []model.EveCharacterFittingItem{
		{TypeID: 2001, Quantity: 1, Flag: "LoSlot2"},
		{TypeID: 2002, Quantity: 1, Flag: "MedSlot0"},
		{TypeID: 3001, Quantity: 1, Flag: "Cargo"},
	}, map[int]string{})

	payload, err := json.Marshal(map[string]any{
		"name":        config.Name,
		"description": config.Description,
		"fittings": []map[string]any{{
			"id":           fitting.ID,
			"fitting_name": fitting.FittingName,
			"eft":          eft,
			"srp_amount":   fitting.SrpAmount,
		}},
	})
	if err != nil {
		t.Fatalf("marshal update payload: %v", err)
	}

	var updateReq UpdateFleetConfigRequest
	if err := json.Unmarshal(payload, &updateReq); err != nil {
		t.Fatalf("unmarshal update payload: %v", err)
	}

	if _, err := svc.UpdateFleetConfig(config.ID, 42, []string{model.RoleAdmin}, &updateReq); err != nil {
		t.Fatalf("update fleet config: %v", err)
	}

	updatedFittings, err := svc.repo.ListFittingsByConfigID(config.ID)
	if err != nil {
		t.Fatalf("list updated fittings: %v", err)
	}
	if len(updatedFittings) != 1 {
		t.Fatalf("expected 1 updated fitting, got %d", len(updatedFittings))
	}

	updatedItems, err := svc.repo.ListItemsByFittingIDs([]uint{updatedFittings[0].ID})
	if err != nil {
		t.Fatalf("list updated items: %v", err)
	}

	updatedLow := mustFindFleetConfigItem(t, updatedItems, 2001, "LoSlot2")
	updatedMed := mustFindFleetConfigItem(t, updatedItems, 2002, "MedSlot0")
	updatedCargo := mustFindFleetConfigItem(t, updatedItems, 3001, "Cargo")

	if updatedLow.Importance != model.FittingItemOptional {
		t.Fatalf("low-slot item importance = %q, want %q", updatedLow.Importance, model.FittingItemOptional)
	}
	if updatedMed.Importance != model.FittingItemReplaceable {
		t.Fatalf("med-slot item importance = %q, want %q", updatedMed.Importance, model.FittingItemReplaceable)
	}
	if updatedCargo.Flag != "Cargo" {
		t.Fatalf("cargo item flag = %q, want Cargo", updatedCargo.Flag)
	}

	replacements, err := svc.repo.ListReplacementsByItemIDs([]uint{updatedMed.ID})
	if err != nil {
		t.Fatalf("list updated replacements: %v", err)
	}
	if len(replacements) != 1 || replacements[0].TypeID != 4001 {
		t.Fatalf("expected replacement type_id 4001 after round trip, got %+v", replacements)
	}
}

func TestParseEFTToFittingClassifiesCountedSectionsByTypeMetadata(t *testing.T) {
	db := newFleetConfigServiceTestDB(t)

	seedSDETypeMetadata(t, db, 5001, "TypeID:5001", "Scout Drone Group", "Drone")
	seedSDETypeMetadata(t, db, 6001, "TypeID:6001", "Light Fighter Group", "Fighter")
	seedSDETypeMetadata(t, db, 7001, "TypeID:7001", "Ammo Group", "Charge")

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	svc := &FleetConfigService{
		repo:    repository.NewFleetConfigRepository(),
		sdeRepo: repository.NewSdeRepository(),
		http:    &http.Client{Timeout: 30 * time.Second},
	}

	_, items, err := svc.parseEFTToFitting(
		"[TypeID:1001, Test Fit]\n\nTypeID:5001 x5\n\nTypeID:6001 x1\n\nTypeID:7001 x100",
	)
	if err != nil {
		t.Fatalf("parse EFT: %v", err)
	}

	mustFindFleetConfigItem(t, items, 5001, "DroneBay")
	mustFindFleetConfigItem(t, items, 6001, "FighterBay")
	mustFindFleetConfigItem(t, items, 7001, "Cargo")
}

func TestUpdateFittingItemsSettingsRejectsItemsOutsideFitting(t *testing.T) {
	db := newFleetConfigServiceTestDB(t)

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	svc := &FleetConfigService{
		repo:    repository.NewFleetConfigRepository(),
		sdeRepo: repository.NewSdeRepository(),
		http:    &http.Client{Timeout: 30 * time.Second},
	}

	config := model.FleetConfig{Name: "Shield Doctrine", Description: "test", CreatedBy: 42}
	if err := db.Create(&config).Error; err != nil {
		t.Fatalf("create config: %v", err)
	}

	fittingA := model.FleetConfigFitting{FleetConfigID: config.ID, ShipTypeID: 1001, FittingName: "Fit A"}
	fittingB := model.FleetConfigFitting{FleetConfigID: config.ID, ShipTypeID: 1002, FittingName: "Fit B"}
	if err := db.Create(&fittingA).Error; err != nil {
		t.Fatalf("create fitting A: %v", err)
	}
	if err := db.Create(&fittingB).Error; err != nil {
		t.Fatalf("create fitting B: %v", err)
	}

	itemA := model.FleetConfigFittingItem{FleetConfigFittingID: fittingA.ID, TypeID: 2001, Quantity: 1, Flag: "LoSlot0"}
	itemB := model.FleetConfigFittingItem{FleetConfigFittingID: fittingB.ID, TypeID: 2002, Quantity: 1, Flag: "LoSlot0", Importance: model.FittingItemReplaceable}
	if err := db.Create(&itemA).Error; err != nil {
		t.Fatalf("create fitting A item: %v", err)
	}
	if err := db.Create(&itemB).Error; err != nil {
		t.Fatalf("create fitting B item: %v", err)
	}
	if err := db.Create(&model.FleetConfigFittingItemReplacement{
		FleetConfigFittingItemID: itemB.ID,
		TypeID:                   4001,
	}).Error; err != nil {
		t.Fatalf("create fitting B replacement: %v", err)
	}

	err := svc.UpdateFittingItemsSettings(config.ID, fittingA.ID, 42, []string{model.RoleAdmin}, &UpdateFittingItemsSettingsRequest{
		Items: []UpdateItemSettingsReq{{
			ID:                 itemB.ID,
			Importance:         model.FittingItemOptional,
			Penalty:            model.FittingPenaltyNone,
			ReplacementPenalty: model.FittingPenaltyNone,
		}},
	})
	if err == nil || !strings.Contains(err.Error(), "不属于该装配") {
		t.Fatalf("expected fitting ownership error, got %v", err)
	}

	replacements, err := svc.repo.ListReplacementsByItemIDs([]uint{itemB.ID})
	if err != nil {
		t.Fatalf("list fitting B replacements: %v", err)
	}
	if len(replacements) != 1 || replacements[0].TypeID != 4001 {
		t.Fatalf("expected fitting B replacements to remain untouched, got %+v", replacements)
	}
}

func TestGetFittingEFTReturnsNotFoundForMissingConfig(t *testing.T) {
	db := newFleetConfigServiceTestDB(t)

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	svc := &FleetConfigService{
		repo:    repository.NewFleetConfigRepository(),
		sdeRepo: repository.NewSdeRepository(),
		http:    &http.Client{Timeout: 30 * time.Second},
	}

	_, err := svc.GetFittingEFT(999, "en")
	if err == nil || !strings.Contains(err.Error(), "舰队配置不存在") {
		t.Fatalf("expected missing config error, got %v", err)
	}
}

func TestUpdateFleetConfigRejectsEmptyFittings(t *testing.T) {
	db := newFleetConfigServiceTestDB(t)

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	svc := &FleetConfigService{
		repo:    repository.NewFleetConfigRepository(),
		sdeRepo: repository.NewSdeRepository(),
		http:    &http.Client{Timeout: 30 * time.Second},
	}

	created, err := svc.CreateFleetConfig(42, &CreateFleetConfigRequest{
		Name:        "Shield Doctrine",
		Description: "test",
		Fittings: []FleetConfigFittingReq{{
			FittingName: "Test Fit",
			EFT:         "[TypeID:1001, Test Fit]\nTypeID:2001",
		}},
	})
	if err != nil {
		t.Fatalf("create fleet config: %v", err)
	}

	name := "Shield Doctrine"
	emptyFittings := []FleetConfigFittingReq{}
	_, err = svc.UpdateFleetConfig(created.ID, 42, []string{model.RoleAdmin}, &UpdateFleetConfigRequest{
		Name:     &name,
		Fittings: &emptyFittings,
	})
	if err == nil || !strings.Contains(err.Error(), "至少保留一个装配") {
		t.Fatalf("expected empty fittings error, got %v", err)
	}
}

func TestUpdateFleetConfigWithoutFittingsPreservesExistingFittingIDs(t *testing.T) {
	db := newFleetConfigServiceTestDB(t)

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	svc := &FleetConfigService{
		repo:    repository.NewFleetConfigRepository(),
		sdeRepo: repository.NewSdeRepository(),
		http:    &http.Client{Timeout: 30 * time.Second},
	}

	created, err := svc.CreateFleetConfig(42, &CreateFleetConfigRequest{
		Name:        "Shield Doctrine",
		Description: "test",
		Fittings: []FleetConfigFittingReq{{
			FittingName: "Test Fit",
			EFT:         "[TypeID:1001, Test Fit]\nTypeID:2001",
		}},
	})
	if err != nil {
		t.Fatalf("create fleet config: %v", err)
	}
	if len(created.Fittings) != 1 {
		t.Fatalf("expected 1 fitting, got %d", len(created.Fittings))
	}

	originalFittingID := created.Fittings[0].ID
	updatedName := "Updated Doctrine"
	updatedDescription := "renamed only"

	updated, err := svc.UpdateFleetConfig(created.ID, 42, []string{model.RoleAdmin}, &UpdateFleetConfigRequest{
		Name:        &updatedName,
		Description: &updatedDescription,
	})
	if err != nil {
		t.Fatalf("update metadata-only fleet config: %v", err)
	}
	if len(updated.Fittings) != 1 {
		t.Fatalf("expected 1 fitting after metadata-only update, got %d", len(updated.Fittings))
	}
	if updated.Fittings[0].ID != originalFittingID {
		t.Fatalf("metadata-only update changed fitting ID from %d to %d", originalFittingID, updated.Fittings[0].ID)
	}
}

func newFleetConfigServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:fleet_config_service_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.FleetConfig{},
		&model.FleetConfigFitting{},
		&model.FleetConfigFittingItem{},
		&model.FleetConfigFittingItemReplacement{},
	); err != nil {
		t.Fatalf("auto migrate fleet config tables: %v", err)
	}
	if err := db.Exec(`CREATE TABLE "invTypes" ("typeID" integer primary key, "typeName" text not null, "groupID" integer, "marketGroupID" integer, "published" integer)`).Error; err != nil {
		t.Fatalf("create invTypes table: %v", err)
	}
	if err := db.Exec(`CREATE TABLE "invGroups" ("groupID" integer primary key, "groupName" text, "categoryID" integer)`).Error; err != nil {
		t.Fatalf("create invGroups table: %v", err)
	}
	if err := db.Exec(`CREATE TABLE "invCategories" ("categoryID" integer primary key, "categoryName" text)`).Error; err != nil {
		t.Fatalf("create invCategories table: %v", err)
	}
	if err := db.Exec(`CREATE TABLE "invMarketGroups" ("marketGroupID" integer primary key, "marketGroupName" text)`).Error; err != nil {
		t.Fatalf("create invMarketGroups table: %v", err)
	}
	if err := db.Exec(`CREATE TABLE "trnTranslations" ("tcID" integer, "keyID" integer, "languageID" text, "text" text)`).Error; err != nil {
		t.Fatalf("create trnTranslations table: %v", err)
	}
	return db
}

func seedSDETypeMetadata(t *testing.T, db *gorm.DB, typeID int64, typeName, groupName, categoryName string) {
	t.Helper()

	groupID := typeID + 10_000
	categoryID := typeID + 20_000
	if err := db.Exec(`INSERT INTO "invCategories" ("categoryID", "categoryName") VALUES (?, ?)`, categoryID, categoryName).Error; err != nil {
		t.Fatalf("insert invCategories row: %v", err)
	}
	if err := db.Exec(`INSERT INTO "invGroups" ("groupID", "groupName", "categoryID") VALUES (?, ?, ?)`, groupID, groupName, categoryID).Error; err != nil {
		t.Fatalf("insert invGroups row: %v", err)
	}
	if err := db.Exec(`INSERT INTO "invTypes" ("typeID", "typeName", "groupID", "published") VALUES (?, ?, ?, 1)`, typeID, typeName, groupID).Error; err != nil {
		t.Fatalf("insert invTypes row: %v", err)
	}
}

func mustFindFleetConfigItem(t *testing.T, items []model.FleetConfigFittingItem, typeID int64, flag string) model.FleetConfigFittingItem {
	t.Helper()

	for _, item := range items {
		if item.TypeID == typeID && item.Flag == flag {
			return item
		}
	}
	t.Fatalf("item not found for type_id=%d flag=%s", typeID, flag)
	return model.FleetConfigFittingItem{}
}
