package handler

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/pkg/response"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUpdateCharacterESIRestrictionConfigRejectsMissingField(t *testing.T) {
	db := newSysConfigHandlerTestDB(t)

	originalDB := global.DB
	originalRedis := global.Redis
	global.DB = db
	global.Redis = redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:0",
		DialTimeout:  10 * time.Millisecond,
		ReadTimeout:  10 * time.Millisecond,
		WriteTimeout: 10 * time.Millisecond,
		PoolTimeout:  10 * time.Millisecond,
		MaxRetries:   0,
	})
	defer func() {
		global.DB = originalDB
		if global.Redis != nil {
			_ = global.Redis.Close()
		}
		global.Redis = originalRedis
	}()

	recorder := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/v1/system/basic-config/character-esi-restriction", bytes.NewBufferString(`{}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("roles", []string{model.RoleSuperAdmin})

	NewSysConfigHandler().UpdateCharacterESIRestrictionConfig(ctx)

	var resp response.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != response.CodeParamError {
		t.Fatalf("response code = %d, want %d", resp.Code, response.CodeParamError)
	}

	var count int64
	if err := db.Model(&model.SystemConfig{}).
		Where("key = ?", model.SysConfigEnforceCharacterESIRestriction).
		Count(&count).Error; err != nil {
		t.Fatalf("count config rows: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no config row to be written, got %d", count)
	}
}

func newSysConfigHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:sys_config_handler_test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.SystemConfig{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}
