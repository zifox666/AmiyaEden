package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// AlliancePAPService 联盟 PAP 业务逻辑层
type AlliancePAPService struct {
	repo     *repository.AlliancePAPRepository
	charRepo *repository.EveCharacterRepository
	userRepo *repository.UserRepository
	http     *http.Client
}

func NewAlliancePAPService() *AlliancePAPService {
	return &AlliancePAPService{
		repo:     repository.NewAlliancePAPRepository(),
		charRepo: repository.NewEveCharacterRepository(),
		userRepo: repository.NewUserRepository(),
		http:     &http.Client{Timeout: 30 * time.Second},
	}
}

// ─── 外部 API 响应结构 ───

type alliancePAPAPIResponse struct {
	Fleets         []alliancePAPFleet `json:"fleets"`
	MainCharacter  string             `json:"main_character"`
	Month          string             `json:"month"`
	Year           string             `json:"year"`
	Ranking        alliancePAPRanking `json:"ranking"`
	TotalPap       string             `json:"total_pap"`
	YearlyTotalPap string             `json:"yearly_total_pap"`
}

type alliancePAPFleet struct {
	Character struct {
		CharacterID   string `json:"character_id"`
		CharacterName string `json:"character_name"`
	} `json:"character"`
	EndAt   string `json:"end_at"`
	FleetID string `json:"fleet_id"`
	Level   string `json:"level"`
	Pap     string `json:"pap"`
	Ship    struct {
		GroupID   string `json:"group_id"`
		GroupName string `json:"group_name"`
		TypeID    string `json:"type_id"`
		TypeName  string `json:"type_name"`
	} `json:"ship"`
	StartAt string `json:"start_at"`
	Title   string `json:"title"`
}

type alliancePAPRanking struct {
	CalculatedAt      string `json:"calculated_at"`
	CorporationID     string `json:"corporation_id"`
	GlobalMonthlyRank int    `json:"global_monthly_rank"`
	GlobalYearlyRank  int    `json:"global_yearly_rank"`
	MonthlyRank       int    `json:"monthly_rank"`
	TotalGlobal       int    `json:"total_global"`
	TotalInCorp       int    `json:"total_in_corp"`
	YearlyRank        int    `json:"yearly_rank"`
}

const alliancePAPTimeLayout = "2006-01-02 15:04:05"

// FetchAndStore 拉取指定主角色某月的联盟 PAP 数据并入库
func (s *AlliancePAPService) FetchAndStore(mainChar string, year, month int) error {
	cfg := global.Config.AlliancePAP
	if cfg.BaseURL == "" || cfg.APIKey == "" {
		return fmt.Errorf("alliance_pap 配置不完整（base_url 或 api_key 为空）")
	}

	url := fmt.Sprintf("%s/api/pap/main?main_character=%s&year=%d&month=%d",
		cfg.BaseURL, mainChar, year, month)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("构建请求失败: %w", err)
	}
	req.Header.Set("x-api-key", cfg.APIKey)
	req.Header.Set("User-Agent", "AmiyaEden/1.0")

	resp, err := s.http.Do(req)
	if err != nil {
		return fmt.Errorf("请求联盟 PAP API 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("联盟 PAP API 返回 %d: %s", resp.StatusCode, string(body))
	}

	var apiResp alliancePAPAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("解析联盟 PAP 响应失败: %w", err)
	}

	// 入库舰队明细
	for _, f := range apiResp.Fleets {
		pap, _ := strconv.ParseFloat(f.Pap, 64)

		startAt, _ := time.ParseInLocation(alliancePAPTimeLayout, f.StartAt, time.UTC)
		var endAt *time.Time
		if f.EndAt != "" {
			t, err := time.ParseInLocation(alliancePAPTimeLayout, f.EndAt, time.UTC)
			if err == nil {
				endAt = &t
			}
		}

		rec := &model.AlliancePAPRecord{
			MainCharacter: apiResp.MainCharacter,
			CharacterID:   f.Character.CharacterID,
			CharacterName: f.Character.CharacterName,
			FleetID:       f.FleetID,
			Year:          year,
			Month:         month,
			StartAt:       startAt,
			EndAt:         endAt,
			Title:         f.Title,
			Level:         f.Level,
			Pap:           pap,
			ShipGroupID:   f.Ship.GroupID,
			ShipGroupName: f.Ship.GroupName,
			ShipTypeID:    f.Ship.TypeID,
			ShipTypeName:  f.Ship.TypeName,
		}

		if err := s.repo.UpsertRecord(rec); err != nil {
			global.Logger.Warn("upsert alliance pap record 失败",
				zap.String("fleet_id", f.FleetID),
				zap.String("character_id", f.Character.CharacterID),
				zap.Error(err))
		}
	}

	// 解析汇总
	totalPap, _ := strconv.ParseFloat(apiResp.TotalPap, 64)
	yearlyTotalPap, _ := strconv.ParseFloat(apiResp.YearlyTotalPap, 64)

	var calculatedAt time.Time
	if apiResp.Ranking.CalculatedAt != "" {
		calculatedAt, _ = time.ParseInLocation(alliancePAPTimeLayout, apiResp.Ranking.CalculatedAt, time.UTC)
	}

	summary := &model.AlliancePAPSummary{
		MainCharacter:     apiResp.MainCharacter,
		Year:              year,
		Month:             month,
		CorporationID:     apiResp.Ranking.CorporationID,
		TotalPap:          totalPap,
		YearlyTotalPap:    yearlyTotalPap,
		MonthlyRank:       apiResp.Ranking.MonthlyRank,
		YearlyRank:        apiResp.Ranking.YearlyRank,
		GlobalMonthlyRank: apiResp.Ranking.GlobalMonthlyRank,
		GlobalYearlyRank:  apiResp.Ranking.GlobalYearlyRank,
		TotalInCorp:       apiResp.Ranking.TotalInCorp,
		TotalGlobal:       apiResp.Ranking.TotalGlobal,
		CalculatedAt:      calculatedAt,
	}

	return s.repo.UpsertSummary(summary)
}

// FetchAllUsers 拉取系统中所有用户主角色的当前月 PAP 数据
func (s *AlliancePAPService) FetchAllUsers(year, month int) {
	userIDs, err := s.userRepo.ListAllIDs()
	if err != nil {
		global.Logger.Error("获取用户列表失败", zap.Error(err))
		return
	}

	for _, uid := range userIDs {
		user, err := s.userRepo.GetByID(uid)
		if err != nil || user.PrimaryCharacterID == 0 {
			continue
		}

		char, err := s.charRepo.GetByCharacterID(user.PrimaryCharacterID)
		if err != nil || char.CharacterName == "" {
			continue
		}

		if err := s.FetchAndStore(char.CharacterName, year, month); err != nil {
			global.Logger.Warn("拉取联盟 PAP 失败",
				zap.String("main_char", char.CharacterName),
				zap.Error(err))
		}
	}
}

// ─── 查询接口 ───

// GetMyPAP 获取当前用户的联盟 PAP 数据
type AlliancePAPResult struct {
	Summary *model.AlliancePAPSummary `json:"summary"`
	Fleets  []model.AlliancePAPRecord `json:"fleets"`
}

func (s *AlliancePAPService) GetMyPAP(mainChar string, year, month int) (*AlliancePAPResult, error) {
	summary, err := s.repo.GetSummary(mainChar, year, month)
	if err != nil {
		// no data yet is ok
		summary = nil
	}
	records, err := s.repo.ListRecords(mainChar, year, month)
	if err != nil {
		return nil, err
	}
	return &AlliancePAPResult{Summary: summary, Fleets: records}, nil
}

// GetAllPAP 获取所有成员某月联盟 PAP 汇总（管理员）
func (s *AlliancePAPService) GetAllPAP(year, month int) ([]model.AlliancePAPSummary, error) {
	return s.repo.ListAllSummaries(year, month)
}
