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

func TestMentorMentorHandlerGetRewardStagesReturnsConfiguredStages(t *testing.T) {
	db := newMentorMentorHandlerTestDB(t)
	seedMentorRewardStages(t, db)

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	result := performMentorRewardStagesRequest(t)
	if result.Code != response.CodeOK {
		t.Fatalf("expected success code, got %#v", result)
	}
	if len(result.Data) != 2 {
		t.Fatalf("expected 2 configured stages, got %#v", result.Data)
	}
	if result.Data[0].StageOrder != 1 {
		t.Fatalf("expected first stage_order to be 1, got %#v", result.Data[0])
	}
	if result.Data[0].Name != "首次达标" {
		t.Fatalf("expected first stage name to be preserved, got %#v", result.Data[0])
	}
	if result.Data[1].ConditionType != model.MentorConditionPapCount {
		t.Fatalf("expected second stage condition type to be preserved, got %#v", result.Data[1])
	}
}

type mentorRewardStagesHandlerResponse struct {
	Code int                       `json:"code"`
	Msg  string                    `json:"msg"`
	Data []model.MentorRewardStage `json:"data"`
}

func performMentorRewardStagesRequest(t *testing.T) mentorRewardStagesHandlerResponse {
	t.Helper()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/mentor/dashboard/reward-stages", nil)
	ctx.Set("userID", uint(1))
	ctx.Set("roles", []string{model.RoleMentor})

	NewMentorMentorHandler().GetRewardStages(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected http status 200, got %d", recorder.Code)
	}

	var result mentorRewardStagesHandlerResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return result
}

func newMentorMentorHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:mentor_mentor_handler_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.MentorRewardStage{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func seedMentorRewardStages(t *testing.T, db *gorm.DB) {
	t.Helper()

	stages := []model.MentorRewardStage{
		{
			StageOrder:    1,
			Name:          "首次达标",
			ConditionType: model.MentorConditionSkillPoints,
			Threshold:     4_000_000,
			RewardAmount:  20,
		},
		{
			StageOrder:    2,
			Name:          "军团活跃",
			ConditionType: model.MentorConditionPapCount,
			Threshold:     10,
			RewardAmount:  30,
		},
	}
	if err := db.Create(&stages).Error; err != nil {
		t.Fatalf("create stages: %v", err)
	}
}
