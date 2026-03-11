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
	repo       *repository.AlliancePAPRepository
	charRepo   *repository.EveCharacterRepository
	userRepo   *repository.UserRepository
	walletRepo *repository.SysWalletRepository
	cfgRepo    *repository.SysConfigRepository
	http       *http.Client
}

func NewAlliancePAPService() *AlliancePAPService {
	return &AlliancePAPService{
		repo:       repository.NewAlliancePAPRepository(),
		charRepo:   repository.NewEveCharacterRepository(),
		userRepo:   repository.NewUserRepository(),
		walletRepo: repository.NewSysWalletRepository(),
		cfgRepo:    repository.NewSysConfigRepository(),
		http:       &http.Client{Timeout: 30 * time.Second},
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

// ─── 修改接口 ───
type PAPImportInfo struct {
	PrimaryCharacterName string `json:"primary_character_name" binding:"required"`
	MonthlyPAP float64 `json:"monthly_pap,default=0" binding:"gte=0"`
	CalculatedAt string `json:"calculated_at" binding:"required"`
}

// ImportAlliancePAP 导入联盟 PAP 数据
func (s *AlliancePAPService) ImportAlliancePAP(year, month int, data *PAPImportInfo, mainChar *model.EveCharacter) error {
	existingSummary, err := s.repo.GetSummary(mainChar.CharacterName, year, month)
	if err != nil {
		existingSummary = nil
	}
	
	var totalPap float64 = data.MonthlyPAP
	var yearlyTotalPap float64 = data.MonthlyPAP
	var monthlyRank int = 1
	var yearlyRank int = 1
	var globalMonthlyRank int = 1
	var globalYearlyRank int = 1
	var totalInCorp int = 0
	var totalGlobal int = 0
	calculatedAt, err := time.ParseInLocation(alliancePAPTimeLayout, data.CalculatedAt, time.UTC)

	if err != nil {
		return err
	}

	if existingSummary != nil {
		delta := data.MonthlyPAP - existingSummary.TotalPap
		totalPap = existingSummary.TotalPap + delta
		yearlyTotalPap = existingSummary.YearlyTotalPap + delta
		monthlyRank = existingSummary.MonthlyRank
		yearlyRank = existingSummary.YearlyRank
		globalMonthlyRank = existingSummary.GlobalMonthlyRank
		globalYearlyRank = existingSummary.GlobalYearlyRank
		totalInCorp = existingSummary.TotalInCorp
		totalGlobal = existingSummary.TotalGlobal
	}

	corporationID := strconv.FormatInt(mainChar.CorporationID, 10)
	
	summary := &model.AlliancePAPSummary{
		MainCharacter:     data.PrimaryCharacterName,
		Year:              year,
		Month:             month,
		CorporationID:     corporationID,
		TotalPap:          totalPap,
		YearlyTotalPap:    yearlyTotalPap,
		MonthlyRank:       monthlyRank,
		YearlyRank:        yearlyRank,
		GlobalMonthlyRank: globalMonthlyRank,
		GlobalYearlyRank:  globalYearlyRank,
		TotalInCorp:       totalInCorp,
		TotalGlobal:       totalGlobal,
		CalculatedAt:      calculatedAt,
	}

	if err := s.repo.UpsertSummary(summary); err != nil {
		global.Logger.Warn("upsert alliance pap summary 失败",
			zap.String("main_char", data.PrimaryCharacterName),
			zap.Error(err))
		return err
	}

	return nil
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

// GetAllPAPPaged 分页获取所有成员某月联盟 PAP 汇总（管理员）
// corporationIDs 非空时只返回这些军团的数据
func (s *AlliancePAPService) GetAllPAPPaged(year, month, page, pageSize int, corporationIDs []int64) ([]model.AlliancePAPSummary, int64, error) {
	return s.repo.ListAllSummariesPaged(year, month, page, pageSize, corporationIDs)
}

// ─── PAP 兑换配置 ───

// PAPExchangeConfigDTO PAP 兑换配置视图对象
type PAPExchangeConfigDTO struct {
	WalletPerPAP float64 `json:"wallet_per_pap"`
	Enabled      bool    `json:"enabled"`
}

// GetExchangeConfig 从 system_config 表读取 PAP 兑换配置
func (s *AlliancePAPService) GetExchangeConfig() (*PAPExchangeConfigDTO, error) {
	return &PAPExchangeConfigDTO{
		WalletPerPAP: s.cfgRepo.GetFloat(model.SysConfigPAPWalletPerPAP, 1),
		Enabled:      s.cfgRepo.GetBool(model.SysConfigPAPExchangeEnabled, true),
	}, nil
}

// SetExchangeConfigRequest 更新兑换配置的请求结构
type SetExchangeConfigRequest struct {
	WalletPerPAP float64 `json:"wallet_per_pap" binding:"required,gt=0"`
	Enabled      bool    `json:"enabled"`
}

// SetExchangeConfig 将 PAP 兑换配置写入 system_config 表（含缓存刷新）
func (s *AlliancePAPService) SetExchangeConfig(req *SetExchangeConfigRequest) (*PAPExchangeConfigDTO, error) {
	if err := s.cfgRepo.Set(model.SysConfigPAPWalletPerPAP,
		fmt.Sprintf("%g", req.WalletPerPAP), "每 1 PAP 兑换的系统钱包数量"); err != nil {
		return nil, err
	}
	enabledStr := "false"
	if req.Enabled {
		enabledStr = "true"
	}
	if err := s.cfgRepo.Set(model.SysConfigPAPExchangeEnabled, enabledStr, "PAP 兑换开关"); err != nil {
		return nil, err
	}
	return &PAPExchangeConfigDTO{WalletPerPAP: req.WalletPerPAP, Enabled: req.Enabled}, nil
}

// ─── 月度归档 + PAP 兑换到系统钱包 ───

// SettleMonthResult 月度结算结果
type SettleMonthResult struct {
	Year         int     `json:"year"`
	Month        int     `json:"month"`
	TotalUsers   int     `json:"total_users"`   // 本次参与结算的用户数
	SkippedUsers int     `json:"skipped_users"` // 跳过（已兑换或找不到用户）
	TotalWallet  float64 `json:"total_wallet"`  // 本次共发放系统钱包数量
}

// SettleMonth 归档某月并将 PAP 批量兑换为系统钱包
// 如果 walletConvert=true，则同时兑换；否则仅归档
// corporationIDs 非空时只结算这些军团的数据
func (s *AlliancePAPService) SettleMonth(year, month int, walletConvert bool, operatorID uint, corporationIDs []int64) (*SettleMonthResult, error) {
	// 1. 归档
	if err := s.repo.MarkArchived(year, month); err != nil {
		return nil, fmt.Errorf("归档失败: %w", err)
	}

	result := &SettleMonthResult{Year: year, Month: month}
	if !walletConvert {
		return result, nil
	}

	// 2. 获取兑换配置（来自 system_config，带缓存）
	walletPerPAP := s.cfgRepo.GetFloat(model.SysConfigPAPWalletPerPAP, 1)
	enabled := s.cfgRepo.GetBool(model.SysConfigPAPExchangeEnabled, true)
	if !enabled {
		return nil, fmt.Errorf("PAP 兑换功能已关闭，请在设置中开启")
	}

	// 3. 查找该月所有未兑换且 PAP > 0 的汇总
	summaries, err := s.repo.ListUnredeemedSummaries(year, month, corporationIDs)
	if err != nil {
		return nil, fmt.Errorf("查询未兑换 PAP 失败: %w", err)
	}

	// 4. 逐条兑换
	for _, summary := range summaries {
		// 通过主角色名找到角色
		char, err := s.charRepo.GetByCharacterName(summary.MainCharacter)
		if err != nil || char.CharacterID == 0 {
			global.Logger.Warn("PAP 结算：找不到角色",
				zap.String("main_char", summary.MainCharacter))
			result.SkippedUsers++
			continue
		}

		// 找到绑定该主角色的用户
		user, err := s.userRepo.GetByPrimaryCharacterID(char.CharacterID)
		if err != nil || user == nil {
			global.Logger.Warn("PAP 结算：找不到对应用户",
				zap.String("main_char", summary.MainCharacter),
				zap.Int64("character_id_int", int64(char.CharacterID)))
			result.SkippedUsers++
			continue
		}

		// 计算应发钱包数
		walletAmount := summary.TotalPap * walletPerPAP
		if walletAmount <= 0 {
			result.SkippedUsers++
			continue
		}

		// 事务：获取钱包 → 加余额 → 写流水
		wallet, err := s.walletRepo.GetOrCreateWallet(user.ID)
		if err != nil {
			global.Logger.Warn("PAP 结算：获取钱包失败",
				zap.Uint("user_id", user.ID), zap.Error(err))
			result.SkippedUsers++
			continue
		}

		newBalance := wallet.Balance + walletAmount
		tx := global.DB.Begin()

		if err := s.walletRepo.UpdateBalanceTx(tx, user.ID, newBalance); err != nil {
			tx.Rollback()
			global.Logger.Warn("PAP 结算：更新余额失败",
				zap.Uint("user_id", user.ID), zap.Error(err))
			result.SkippedUsers++
			continue
		}

		walletTx := &model.WalletTransaction{
			UserID:       user.ID,
			Amount:       walletAmount,
			Reason:       fmt.Sprintf("%d年%d月联盟PAP兑换（%.2f PAP × %.2f）", year, month, summary.TotalPap, walletPerPAP),
			RefType:      model.WalletRefPapConvert,
			RefID:        fmt.Sprintf("pap:%d:%d:%s", year, month, summary.MainCharacter),
			BalanceAfter: newBalance,
			OperatorID:   operatorID,
		}
		if err := s.walletRepo.CreateTransactionTx(tx, walletTx); err != nil {
			tx.Rollback()
			global.Logger.Warn("PAP 结算：写入流水失败",
				zap.Uint("user_id", user.ID), zap.Error(err))
			result.SkippedUsers++
			continue
		}

		if err := tx.Commit().Error; err != nil {
			global.Logger.Warn("PAP 结算：提交事务失败",
				zap.Uint("user_id", user.ID), zap.Error(err))
			result.SkippedUsers++
			continue
		}

		// 标记汇总已兑换
		if err := s.repo.MarkSummaryRedeemed(summary.ID, walletAmount); err != nil {
			global.Logger.Warn("PAP 结算：标记兑换状态失败",
				zap.Uint("summary_id", summary.ID), zap.Error(err))
		}

		result.TotalUsers++
		result.TotalWallet += walletAmount
	}

	return result, nil
}
