package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"sync"
	"time"
)

// DashboardService 工作台业务逻辑层
type DashboardService struct {
	charRepo      *repository.EveCharacterRepository
	walletRepo    *repository.EveWalletRepository
	skillRepo     *repository.EveSkillRepository
	sysWalletRepo *repository.SysWalletRepository
	fleetRepo     *repository.FleetRepository
	papRepo       *repository.AlliancePAPRepository
	srpRepo       *repository.SrpRepository
	userRepo      *repository.UserRepository
}

func NewDashboardService() *DashboardService {
	return &DashboardService{
		charRepo:      repository.NewEveCharacterRepository(),
		walletRepo:    repository.NewEveWalletRepository(),
		skillRepo:     repository.NewEveSkillRepository(),
		sysWalletRepo: repository.NewSysWalletRepository(),
		fleetRepo:     repository.NewFleetRepository(),
		papRepo:       repository.NewAlliancePAPRepository(),
		srpRepo:       repository.NewSrpRepository(),
		userRepo:      repository.NewUserRepository(),
	}
}

// ─────────────────────────────────────────────
//  响应结构
// ─────────────────────────────────────────────

// DashboardCards 卡片数据
type DashboardCards struct {
	EveWalletBalance    float64 `json:"eve_wallet_balance"`
	EveSkillPoints      int64   `json:"eve_skill_points"`
	SystemWalletBalance float64 `json:"system_wallet_balance"`
	AlliancePap         float64 `json:"alliance_pap"`
}

// DashboardFleetItem 统一舰队参与记录
type DashboardFleetItem struct {
	Source        string     `json:"source"` // "internal" | "alliance"
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	StartAt       time.Time  `json:"start_at"`
	EndAt         *time.Time `json:"end_at,omitempty"`
	Importance    string     `json:"importance,omitempty"`
	PapCount      float64    `json:"pap_count"`
	ShipTypeName  string     `json:"ship_type_name,omitempty"`
	CharacterName string     `json:"character_name,omitempty"`
}

// DashboardPapMonthly 月度 PAP 统计
type DashboardPapMonthly struct {
	Year     int     `json:"year"`
	Month    int     `json:"month"`
	TotalPap float64 `json:"total_pap"`
}

// DashboardPapStats PAP 统计
type DashboardPapStats struct {
	Alliance []DashboardPapMonthly `json:"alliance"`
	Internal []DashboardPapMonthly `json:"internal"`
}

// DashboardSrpItem 补损列表项
type DashboardSrpItem struct {
	ID                uint      `json:"id"`
	CharacterName     string    `json:"character_name"`
	ShipName          string    `json:"ship_name"`
	SolarSystemName   string    `json:"solar_system_name"`
	KillmailTime      time.Time `json:"killmail_time"`
	RecommendedAmount float64   `json:"recommended_amount"`
	FinalAmount       float64   `json:"final_amount"`
	ReviewStatus      string    `json:"review_status"`
	PayoutStatus      string    `json:"payout_status"`
	CreatedAt         time.Time `json:"created_at"`
}

// DashboardResult 工作台完整响应
type DashboardResult struct {
	Cards    DashboardCards       `json:"cards"`
	Fleets   []DashboardFleetItem `json:"fleets"`
	PapStats DashboardPapStats    `json:"pap_stats"`
	SrpList  []DashboardSrpItem   `json:"srp_list"`
}

// ─────────────────────────────────────────────
//  业务方法
// ─────────────────────────────────────────────

// GetDashboard 获取工作台所有数据
func (s *DashboardService) GetDashboard(userID uint) (*DashboardResult, error) {
	result := &DashboardResult{}

	// 获取用户所有人物
	characters, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		characters = []model.EveCharacter{}
	}

	characterIDs := make([]int64, 0, len(characters))
	for _, c := range characters {
		characterIDs = append(characterIDs, c.CharacterID)
	}

	// 获取用户信息（用于主人物查询联盟 PAP）
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// 获取主人物名称（后续多个查询依赖）
	var mainCharName string
	if user.PrimaryCharacterID != 0 {
		if primaryChar, err := s.charRepo.GetByCharacterID(user.PrimaryCharacterID); err == nil {
			mainCharName = primaryChar.CharacterName
		}
	}

	// ── 并行获取所有数据 ──

	var (
		wg              sync.WaitGroup
		walletBalance   float64
		skillPoints     int64
		sysWalletBal    float64
		alliancePap     float64
		internalFleets  []model.Fleet
		allianceRecords []model.AlliancePAPRecord
		allianceSums    []model.AlliancePAPSummary
		internalStats   []repository.MonthlyPapStat
		srpApps         []model.SrpApplication
	)

	// 1. EVE 钱包余额
	wg.Add(1)
	go func() {
		defer wg.Done()
		walletBalance, _ = s.walletRepo.SumBalanceByCharacterIDs(characterIDs)
	}()

	// 2. EVE 技能点
	wg.Add(1)
	go func() {
		defer wg.Done()
		skillPoints, _ = s.skillRepo.SumTotalSPByCharacterIDs(characterIDs)
	}()

	// 3. 系统钱包余额
	wg.Add(1)
	go func() {
		defer wg.Done()
		if w, err := s.sysWalletRepo.GetOrCreateWallet(userID); err == nil {
			sysWalletBal = w.Balance
		}
	}()

	// 4. 当月联盟 PAP
	wg.Add(1)
	go func() {
		defer wg.Done()
		if mainCharName != "" {
			now := time.Now()
			if summary, err := s.papRepo.GetSummary(mainCharName, now.Year(), int(now.Month())); err == nil {
				alliancePap = summary.TotalPap
			}
		}
	}()

	// 5. 内部舰队
	wg.Add(1)
	go func() {
		defer wg.Done()
		internalFleets, _ = s.fleetRepo.ListFleetsByMemberUserID(userID, 20)
	}()

	// 6. 联盟 PAP 记录
	wg.Add(1)
	go func() {
		defer wg.Done()
		if mainCharName != "" {
			allianceRecords, _ = s.papRepo.ListRecentRecordsByMainChar(mainCharName, 20)
		}
	}()

	// 7. 联盟 PAP 月度汇总
	wg.Add(1)
	go func() {
		defer wg.Done()
		if mainCharName != "" {
			allianceSums, _ = s.papRepo.ListSummariesByMainChar(mainCharName, 12)
		}
	}()

	// 8. 内部 PAP 月度统计
	wg.Add(1)
	go func() {
		defer wg.Done()
		internalStats, _ = s.fleetRepo.SumPapByUserGroupedByMonth(userID)
	}()

	// 9. SRP 列表
	wg.Add(1)
	go func() {
		defer wg.Done()
		srpApps, _, _ = s.srpRepo.ListMyApplications(userID, 1, 10)
	}()

	wg.Wait()

	// ── 组装响应 ──

	result.Cards.EveWalletBalance = walletBalance
	result.Cards.EveSkillPoints = skillPoints
	result.Cards.SystemWalletBalance = sysWalletBal
	result.Cards.AlliancePap = alliancePap

	// 舰队列表
	fleets := make([]DashboardFleetItem, 0, len(internalFleets)+len(allianceRecords))
	for _, f := range internalFleets {
		endAt := f.EndAt
		fleets = append(fleets, DashboardFleetItem{
			Source:     "internal",
			ID:         f.ID,
			Title:      f.Title,
			StartAt:    f.StartAt,
			EndAt:      &endAt,
			Importance: f.Importance,
			PapCount:   f.PapCount,
		})
	}
	for _, r := range allianceRecords {
		fleets = append(fleets, DashboardFleetItem{
			Source:        "alliance",
			ID:            r.FleetID,
			Title:         r.Title,
			StartAt:       r.StartAt,
			EndAt:         r.EndAt,
			PapCount:      r.Pap,
			ShipTypeName:  r.ShipTypeName,
			CharacterName: r.CharacterName,
		})
	}
	result.Fleets = fleets

	// 联盟 PAP 月度
	alliancePapStats := make([]DashboardPapMonthly, 0, len(allianceSums))
	for _, s := range allianceSums {
		alliancePapStats = append(alliancePapStats, DashboardPapMonthly{
			Year:     s.Year,
			Month:    s.Month,
			TotalPap: s.TotalPap,
		})
	}
	result.PapStats.Alliance = alliancePapStats

	// 内部 PAP 月度
	internalPapStats := make([]DashboardPapMonthly, 0, len(internalStats))
	for _, stat := range internalStats {
		internalPapStats = append(internalPapStats, DashboardPapMonthly{
			Year:     stat.Year,
			Month:    stat.Month,
			TotalPap: stat.TotalPap,
		})
	}
	result.PapStats.Internal = internalPapStats

	// SRP 列表
	srpList := make([]DashboardSrpItem, 0, len(srpApps))
	for _, app := range srpApps {
		srpList = append(srpList, DashboardSrpItem{
			ID:                app.ID,
			CharacterName:     app.CharacterName,
			ShipName:          app.ShipName,
			SolarSystemName:   app.SolarSystemName,
			KillmailTime:      app.KillmailTime,
			RecommendedAmount: app.RecommendedAmount,
			FinalAmount:       app.FinalAmount,
			ReviewStatus:      app.ReviewStatus,
			PayoutStatus:      app.PayoutStatus,
			CreatedAt:         app.CreatedAt,
		})
	}
	result.SrpList = srpList

	return result, nil
}
