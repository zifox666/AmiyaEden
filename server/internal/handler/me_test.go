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
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMeHandlerGetMeRejectsInvalidPrimaryCharacterToken(t *testing.T) {
	db := newMeHandlerTestDB(t)
	seedMeHandlerUser(t, db, true)

	originalDB := global.DB
	originalRedis := global.Redis
	global.DB = db
	global.Redis = redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})
	defer func() {
		global.DB = originalDB
		if global.Redis != nil {
			_ = global.Redis.Close()
		}
		global.Redis = originalRedis
	}()

	recorder, result := performGetMeRequest(t, 1)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected http status 401, got %d", recorder.Code)
	}
	if result.Code != response.CodeUnauthorized {
		t.Fatalf("expected unauthorized code, got %#v", result)
	}
}

func TestMeHandlerGetMeReturnsCharacterRestrictionToggleWhenPrimaryIsHealthy(t *testing.T) {
	db := newMeHandlerTestDB(t)
	seedMeHandlerUser(t, db, false)
	seedCharacterESIRestrictionConfig(t, db, false)

	originalDB := global.DB
	originalRedis := global.Redis
	global.DB = db
	global.Redis = redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})
	defer func() {
		global.DB = originalDB
		if global.Redis != nil {
			_ = global.Redis.Close()
		}
		global.Redis = originalRedis
	}()

	recorder, result := performGetMeRequest(t, 1)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected http status 200, got %d", recorder.Code)
	}
	if result.Code != response.CodeOK {
		t.Fatalf("expected success code, got %#v", result)
	}

	var payload map[string]any
	if err := json.Unmarshal(result.Data, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload["enforce_character_esi_restriction"] != false {
		t.Fatalf("expected enforcement flag false, got %#v", payload["enforce_character_esi_restriction"])
	}
	characters, ok := payload["characters"].([]any)
	if !ok || len(characters) != 2 {
		t.Fatalf("expected two characters in payload, got %#v", payload["characters"])
	}
}

func TestMeHandlerGetMeFailsClosedWhenCharacterLoadErrors(t *testing.T) {
	db := newMeHandlerUserOnlyTestDB(t)
	seedMeHandlerUserWithoutCharacters(t, db)

	originalDB := global.DB
	originalRedis := global.Redis
	global.DB = db
	global.Redis = redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})
	defer func() {
		global.DB = originalDB
		if global.Redis != nil {
			_ = global.Redis.Close()
		}
		global.Redis = originalRedis
	}()

	recorder, result := performGetMeRequest(t, 1)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected http status 200, got %d", recorder.Code)
	}
	if result.Code != response.CodeBizError {
		t.Fatalf("expected biz error code, got %#v", result)
	}
}

type meHandlerResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func performGetMeRequest(t *testing.T, userID uint) (*httptest.ResponseRecorder, meHandlerResponse) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	ctx.Set("userID", userID)
	ctx.Set("roles", []string{model.RoleUser})

	NewMeHandler().GetMe(ctx)

	var result meHandlerResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return recorder, result
}

func newMeHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:me_handler_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.EveCharacter{}, &model.SystemConfig{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func newMeHandlerUserOnlyTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:me_handler_user_only_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.SystemConfig{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func seedMeHandlerUser(t *testing.T, db *gorm.DB, primaryTokenInvalid bool) {
	t.Helper()

	user := model.User{
		BaseModel:          model.BaseModel{ID: 1},
		Nickname:           "Amiya",
		QQ:                 "12345",
		Role:               model.RoleUser,
		PrimaryCharacterID: 9001,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	characters := []model.EveCharacter{
		{
			CharacterID:   9001,
			CharacterName: "Amiya Prime",
			PortraitURL:   "portrait.png",
			UserID:        user.ID,
			TokenInvalid:  primaryTokenInvalid,
		},
		{
			CharacterID:   9002,
			CharacterName: "Amiya Alt",
			PortraitURL:   "portrait-alt.png",
			UserID:        user.ID,
			TokenInvalid:  true,
		},
	}
	if err := db.Create(&characters).Error; err != nil {
		t.Fatalf("create characters: %v", err)
	}
}

func seedMeHandlerUserWithoutCharacters(t *testing.T, db *gorm.DB) {
	t.Helper()

	user := model.User{
		BaseModel:          model.BaseModel{ID: 1},
		Nickname:           "Amiya",
		QQ:                 "12345",
		Role:               model.RoleUser,
		PrimaryCharacterID: 9001,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
}

func seedCharacterESIRestrictionConfig(t *testing.T, db *gorm.DB, enabled bool) {
	t.Helper()

	value := "false"
	if enabled {
		value = "true"
	}

	if err := db.Create(&model.SystemConfig{
		Key:   model.SysConfigEnforceCharacterESIRestriction,
		Value: value,
		Desc:  "是否强制限制已失效人物 ESI 账号停留在人物页面",
	}).Error; err != nil {
		t.Fatalf("create system config: %v", err)
	}
}
