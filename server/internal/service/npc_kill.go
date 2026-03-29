package service

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
)

// NpcKillService NPC 刷怪报表业务逻辑层
type NpcKillService struct {
	npcKillRepo *repository.NpcKillRepository
	charRepo    *repository.EveCharacterRepository
	sdeRepo     *repository.SdeRepository
}

func NewNpcKillService() *NpcKillService {
	return &NpcKillService{
		npcKillRepo: repository.NewNpcKillRepository(),
		charRepo:    repository.NewEveCharacterRepository(),
		sdeRepo:     repository.NewSdeRepository(),
	}
}

// ─────────────────────────────────────────────
//  请求 & 响应结构
// ─────────────────────────────────────────────

// NpcKillRequest 刷怪报表请求（个人 - 单人物）
type NpcKillRequest struct {
	CharacterID int64  `json:"character_id" binding:"required"`
	StartDate   string `json:"start_date"`           // 格式: 2006-01-02
	EndDate     string `json:"end_date"`             // 格式: 2006-01-02
	Language    string `json:"language"`             // 默认 zh
	Page        int    `json:"page" binding:"min=0"` // 0 表示不分页，返回全部
	PageSize    int    `json:"page_size" binding:"min=0"`
}

// NpcKillAllRequest 刷怪报表请求（个人 - 名下所有人物汇总）
type NpcKillAllRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Language  string `json:"language"`
	Page      int    `json:"page" binding:"min=0"`
	PageSize  int    `json:"page_size" binding:"min=0"`
}

// NpcKillCorpRequest 刷怪报表请求（公司/管理员）
type NpcKillCorpRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Language  string `json:"language"`
	Page      int    `json:"page" binding:"min=0"`
	PageSize  int    `json:"page_size" binding:"min=0"`
}

// NpcKillSummary 总览数据
type NpcKillSummary struct {
	TotalBounty    float64 `json:"total_bounty"`    // 总刷怪赏金（bounty_prizes amount 合计）
	TotalESS       float64 `json:"total_ess"`       // 总 ESS 金额（ess_escrow_transfer amount 合计）
	TotalIncursion float64 `json:"total_incursion"` // 总入侵收入（incursion_payout amount 合计）
	TotalMission   float64 `json:"total_mission"`   // 总任务奖励（agent_mission_reward amount 合计）
	TotalTax       float64 `json:"total_tax"`       // 总交税金额
	ActualIncome   float64 `json:"actual_income"`   // 实际获得 = bounty + ess + incursion + mission + tax
	TotalRecords   int     `json:"total_records"`   // 总记录数（bounty_prizes 条数）
	EstimatedHours float64 `json:"estimated_hours"` // 大致时长（有效记录 * 20min / 60）
}

// NpcKillByNpc 按 NPC 分类统计
type NpcKillByNpc struct {
	NpcID   int     `json:"npc_id"`
	NpcName string  `json:"npc_name"`
	Count   int     `json:"count"`
	Amount  float64 `json:"amount"` // 不直接来自 amount，仅按比例估算
}

// NpcKillBySystem 按地点分类统计
type NpcKillBySystem struct {
	SolarSystemID   int     `json:"solar_system_id"`
	SolarSystemName string  `json:"solar_system_name"`
	Count           int     `json:"count"`
	Amount          float64 `json:"amount"`
}

// NpcKillTrend 时间趋势（按天）
type NpcKillTrend struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
	Count  int     `json:"count"`
}

// NpcKillJournalItem 刷怪流水条目
type NpcKillJournalItem struct {
	ID              int64   `json:"id"`
	CharacterID     int64   `json:"character_id"`
	CharacterName   string  `json:"character_name"`
	Amount          float64 `json:"amount"`
	Tax             float64 `json:"tax"`
	Date            string  `json:"date"`
	RefType         string  `json:"ref_type"`
	SolarSystemID   int     `json:"solar_system_id"`
	SolarSystemName string  `json:"solar_system_name"`
	Reason          string  `json:"reason"`
}

// NpcKillResponse 刷怪报表响应（个人）
type NpcKillResponse struct {
	Summary  NpcKillSummary       `json:"summary"`
	ByNpc    []NpcKillByNpc       `json:"by_npc"`
	BySystem []NpcKillBySystem    `json:"by_system"`
	Trend    []NpcKillTrend       `json:"trend"`
	Journals []NpcKillJournalItem `json:"journals"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

// NpcKillCorpMemberSummary 公司成员刷怪统计
type NpcKillCorpMemberSummary struct {
	CharacterID    int64   `json:"character_id"`
	CharacterName  string  `json:"character_name"`
	TotalBounty    float64 `json:"total_bounty"`
	TotalESS       float64 `json:"total_ess"`
	TotalIncursion float64 `json:"total_incursion"`
	TotalMission   float64 `json:"total_mission"`
	TotalTax       float64 `json:"total_tax"`
	ActualIncome   float64 `json:"actual_income"`
	RecordCount    int     `json:"record_count"`
}

// NpcKillCorpResponse 公司刷怪报表响应（管理员）
type NpcKillCorpResponse struct {
	Summary  NpcKillSummary             `json:"summary"`
	Members  []NpcKillCorpMemberSummary `json:"members"`
	BySystem []NpcKillBySystem          `json:"by_system"`
	Trend    []NpcKillTrend             `json:"trend"`
}

// ─────────────────────────────────────────────
//  业务方法
// ─────────────────────────────────────────────

// GetNpcKills 获取个人刷怪报表
func (s *NpcKillService) GetNpcKills(userID uint, req *NpcKillRequest) (*NpcKillResponse, error) {
	// 校验人物归属
	if err := s.validateCharacterOwnership(userID, req.CharacterID); err != nil {
		return nil, err
	}

	lang := req.Language
	if lang == "" {
		lang = "zh"
	}

	startDate, endDate := parseDateRange(req.StartDate, req.EndDate)

	// 获取所有记录（用于统计）
	allJournals, err := s.npcKillRepo.GetBountyJournals(req.CharacterID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("获取刷怪数据失败: %w", err)
	}

	// 分页获取流水明细
	var journals []model.EVECharacterWalletJournal
	var total int64
	if req.Page > 0 && req.PageSize > 0 {
		journals, total, err = s.npcKillRepo.GetBountyJournalsPaged(req.CharacterID, startDate, endDate, req.Page, req.PageSize)
		if err != nil {
			return nil, fmt.Errorf("获取刷怪流水失败: %w", err)
		}
	} else {
		journals = allJournals
		total = int64(len(allJournals))
	}

	// 构建响应
	resp := &NpcKillResponse{
		Page:     req.Page,
		PageSize: req.PageSize,
		Total:    total,
	}

	// 计算总览
	resp.Summary = s.calcSummary(allJournals)

	// 解析 NPC 统计
	resp.ByNpc = s.calcByNpc(allJournals, lang)

	// 按地点统计
	resp.BySystem = s.calcBySystem(allJournals)

	// 时间趋势
	resp.Trend = s.calcTrend(allJournals)

	// 流水明细
	resp.Journals = s.buildJournalItems(journals, nil)

	return resp, nil
}

// GetAllNpcKills 获取当前用户名下所有人物的汇总刷怪报表
func (s *NpcKillService) GetAllNpcKills(userID uint, req *NpcKillAllRequest) (*NpcKillResponse, error) {
	lang := req.Language
	if lang == "" {
		lang = "zh"
	}

	startDate, endDate := parseDateRange(req.StartDate, req.EndDate)

	// 获取该用户名下的所有人物
	chars, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取人物列表失败: %w", err)
	}

	charIDs := make([]int64, 0, len(chars))
	charNameMap := make(map[int64]string, len(chars))
	for _, c := range chars {
		charIDs = append(charIDs, c.CharacterID)
		charNameMap[c.CharacterID] = c.CharacterName
	}

	// 获取所有记录（用于统计）
	allJournals, err := s.npcKillRepo.GetBountyJournalsByCharacterIDs(charIDs, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("获取刷怪数据失败: %w", err)
	}

	// 分页获取流水明细
	var journals []model.EVECharacterWalletJournal
	var total int64
	if req.Page > 0 && req.PageSize > 0 {
		journals, total, err = s.npcKillRepo.GetBountyJournalsByCharacterIDsPaged(charIDs, startDate, endDate, req.Page, req.PageSize)
		if err != nil {
			return nil, fmt.Errorf("获取刷怪流水失败: %w", err)
		}
	} else {
		journals = allJournals
		total = int64(len(allJournals))
	}

	resp := &NpcKillResponse{
		Page:     req.Page,
		PageSize: req.PageSize,
		Total:    total,
	}

	resp.Summary = s.calcSummary(allJournals)
	resp.ByNpc = s.calcByNpc(allJournals, lang)
	resp.BySystem = s.calcBySystem(allJournals)
	resp.Trend = s.calcTrend(allJournals)
	resp.Journals = s.buildJournalItems(journals, charNameMap)

	return resp, nil
}

// GetCorpNpcKills 获取公司所有成员的刷怪报表（管理员）
func (s *NpcKillService) GetCorpNpcKills(req *NpcKillCorpRequest) (*NpcKillCorpResponse, error) {
	startDate, endDate := parseDateRange(req.StartDate, req.EndDate)

	// 获取所有已绑定的人物
	allChars, err := s.charRepo.ListAllWithToken()
	if err != nil {
		return nil, fmt.Errorf("获取人物列表失败: %w", err)
	}

	charIDs := make([]int64, 0, len(allChars))
	charNameMap := make(map[int64]string, len(allChars))
	for _, c := range allChars {
		charIDs = append(charIDs, c.CharacterID)
		charNameMap[c.CharacterID] = c.CharacterName
	}

	// 获取所有记录
	allJournals, err := s.npcKillRepo.GetBountyJournalsByCharacterIDs(charIDs, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("获取刷怪数据失败: %w", err)
	}

	resp := &NpcKillCorpResponse{}

	// 总览
	resp.Summary = s.calcSummary(allJournals)

	// 按成员统计
	memberMap := make(map[int64]*NpcKillCorpMemberSummary)
	for _, j := range allJournals {
		m, ok := memberMap[j.CharacterID]
		if !ok {
			m = &NpcKillCorpMemberSummary{
				CharacterID:   j.CharacterID,
				CharacterName: charNameMap[j.CharacterID],
			}
			memberMap[j.CharacterID] = m
		}
		switch j.RefType {
		case "bounty_prizes":
			m.TotalBounty += j.Amount
			m.RecordCount++
		case "ess_escrow_transfer":
			m.TotalESS += j.Amount
		case "incursion_payout":
			m.TotalIncursion += j.Amount
		case "agent_mission_reward":
			m.TotalMission += j.Amount
		}
		m.TotalTax += j.Tax
	}
	members := make([]NpcKillCorpMemberSummary, 0, len(memberMap))
	for _, m := range memberMap {
		m.ActualIncome = m.TotalBounty + m.TotalESS + m.TotalIncursion + m.TotalMission + m.TotalTax
		members = append(members, *m)
	}
	// 按实际收入排序（降序）
	sort.Slice(members, func(i, j int) bool {
		return members[i].ActualIncome > members[j].ActualIncome
	})
	resp.Members = members

	// 按地点统计
	resp.BySystem = s.calcBySystem(allJournals)

	// 时间趋势
	resp.Trend = s.calcTrend(allJournals)

	return resp, nil
}

// ─────────────────────────────────────────────
//  内部辅助方法
// ─────────────────────────────────────────────

func (s *NpcKillService) validateCharacterOwnership(userID uint, characterID int64) error {
	chars, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return fmt.Errorf("获取人物列表失败")
	}
	for _, c := range chars {
		if c.CharacterID == characterID {
			return nil
		}
	}
	return fmt.Errorf("该人物不属于当前用户")
}

// calcSummary 计算总览统计
func (s *NpcKillService) calcSummary(journals []model.EVECharacterWalletJournal) NpcKillSummary {
	var summary NpcKillSummary

	// 用于计算平均值以过滤低效记录
	var bountyAmounts []float64

	for _, j := range journals {
		switch j.RefType {
		case "bounty_prizes":
			summary.TotalBounty += j.Amount
			summary.TotalRecords++
			bountyAmounts = append(bountyAmounts, j.Amount)
		case "ess_escrow_transfer":
			summary.TotalESS += j.Amount
		case "incursion_payout":
			summary.TotalIncursion += j.Amount
		case "agent_mission_reward":
			summary.TotalMission += j.Amount
		}
		summary.TotalTax += j.Tax
	}

	summary.ActualIncome = summary.TotalBounty + summary.TotalESS + summary.TotalIncursion + summary.TotalMission + summary.TotalTax

	// 计算大致时长：每条 bounty_prizes 记录 ≈ 20min，但明显低于平均金额的不算
	if len(bountyAmounts) > 0 {
		avg := summary.TotalBounty / float64(len(bountyAmounts))
		threshold := avg * 0.3 // 低于平均值 30% 的不计入时长
		effectiveCount := 0
		for _, a := range bountyAmounts {
			if a >= threshold {
				effectiveCount++
			}
		}
		summary.EstimatedHours = math.Round(float64(effectiveCount)*20.0/60.0*100) / 100
	}

	return summary
}

// calcByNpc 按 NPC 分类统计
func (s *NpcKillService) calcByNpc(journals []model.EVECharacterWalletJournal, lang string) []NpcKillByNpc {
	npcCountMap := make(map[int]int)

	for _, j := range journals {
		if j.RefType != "bounty_prizes" || j.Reason == "" {
			continue
		}
		npcCounts := parseReason(j.Reason)
		for npcID, count := range npcCounts {
			npcCountMap[npcID] += count
		}
	}

	if len(npcCountMap) == 0 {
		return nil
	}

	// 查询 NPC 名称
	npcIDs := make([]int, 0, len(npcCountMap))
	for id := range npcCountMap {
		npcIDs = append(npcIDs, id)
	}

	npcNameMap := make(map[int]string)
	typeInfos, err := s.sdeRepo.GetTypes(npcIDs, nil, lang)
	if err == nil {
		for _, t := range typeInfos {
			npcNameMap[t.TypeID] = t.TypeName
		}
	}

	result := make([]NpcKillByNpc, 0, len(npcCountMap))
	for npcID, count := range npcCountMap {
		name := npcNameMap[npcID]
		if name == "" {
			name = fmt.Sprintf("Unknown NPC #%d", npcID)
		}
		result = append(result, NpcKillByNpc{
			NpcID:   npcID,
			NpcName: name,
			Count:   count,
		})
	}

	// 按数量排序（降序）
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	return result
}

// calcBySystem 按地点分类统计
func (s *NpcKillService) calcBySystem(journals []model.EVECharacterWalletJournal) []NpcKillBySystem {
	systemMap := make(map[int]*NpcKillBySystem)
	solarSystemIDs := make(map[int]bool)

	for _, j := range journals {
		if j.RefType != "bounty_prizes" {
			continue
		}
		sysID := int(j.ContextID)
		if sysID == 0 {
			continue
		}
		solarSystemIDs[sysID] = true

		sys, ok := systemMap[sysID]
		if !ok {
			sys = &NpcKillBySystem{SolarSystemID: sysID}
			systemMap[sysID] = sys
		}
		sys.Count++
		sys.Amount += j.Amount
	}

	// 查询星系名称
	ids := make([]int, 0, len(solarSystemIDs))
	for id := range solarSystemIDs {
		ids = append(ids, id)
	}
	systemNames, _ := s.npcKillRepo.GetSolarSystemNames(ids)

	result := make([]NpcKillBySystem, 0, len(systemMap))
	for _, sys := range systemMap {
		if name, ok := systemNames[sys.SolarSystemID]; ok {
			sys.SolarSystemName = name
		} else {
			sys.SolarSystemName = fmt.Sprintf("Unknown System #%d", sys.SolarSystemID)
		}
		result = append(result, *sys)
	}

	// 按金额排序（降序）
	sort.Slice(result, func(i, j int) bool {
		return result[i].Amount > result[j].Amount
	})

	return result
}

// calcTrend 按天统计时间趋势
func (s *NpcKillService) calcTrend(journals []model.EVECharacterWalletJournal) []NpcKillTrend {
	dayMap := make(map[string]*NpcKillTrend)

	for _, j := range journals {
		switch j.RefType {
		case "bounty_prizes", "incursion_payout", "agent_mission_reward":
			// 这些类型计入趋势；ess_escrow_transfer 不含时间颗粒度的星系上下文，不单独趋势
		default:
			continue
		}
		day := j.Date.Format("2006-01-02")
		t, ok := dayMap[day]
		if !ok {
			t = &NpcKillTrend{Date: day}
			dayMap[day] = t
		}
		t.Amount += j.Amount
		t.Count++
	}

	result := make([]NpcKillTrend, 0, len(dayMap))
	for _, t := range dayMap {
		result = append(result, *t)
	}

	// 按日期升序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Date < result[j].Date
	})

	return result
}

// buildJournalItems 构建流水条目
func (s *NpcKillService) buildJournalItems(journals []model.EVECharacterWalletJournal, charNameMap map[int64]string) []NpcKillJournalItem {
	// 收集星系 ID
	solarSystemIDs := make(map[int]bool)
	for _, j := range journals {
		if j.RefType == "bounty_prizes" && j.ContextID != 0 {
			solarSystemIDs[int(j.ContextID)] = true
		}
	}
	ids := make([]int, 0, len(solarSystemIDs))
	for id := range solarSystemIDs {
		ids = append(ids, id)
	}
	systemNames, _ := s.npcKillRepo.GetSolarSystemNames(ids)

	items := make([]NpcKillJournalItem, 0, len(journals))
	for _, j := range journals {
		item := NpcKillJournalItem{
			ID:          j.ID,
			CharacterID: j.CharacterID,
			Amount:      j.Amount,
			Tax:         j.Tax,
			Date:        j.Date.Format("2006-01-02 15:04:05"),
			RefType:     j.RefType,
			Reason:      j.Reason,
		}

		if j.RefType == "bounty_prizes" {
			sysID := int(j.ContextID)
			item.SolarSystemID = sysID
			if name, ok := systemNames[sysID]; ok {
				item.SolarSystemName = name
			}
		}

		if charNameMap != nil {
			if name, ok := charNameMap[j.CharacterID]; ok {
				item.CharacterName = name
			}
		}

		items = append(items, item)
	}

	return items
}

// parseReason 解析 reason 字段，提取 NPC ID 和数量
// 格式: "npc_id: num,npc_id: num"
func parseReason(reason string) map[int]int {
	result := make(map[int]int)
	if reason == "" {
		return result
	}

	parts := strings.Split(reason, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			continue
		}
		npcID, err := strconv.Atoi(strings.TrimSpace(kv[0]))
		if err != nil {
			continue
		}
		count, err := strconv.Atoi(strings.TrimSpace(kv[1]))
		if err != nil {
			continue
		}
		result[npcID] += count
	}

	return result
}

// parseDateRange 解析日期范围
func parseDateRange(startStr, endStr string) (*time.Time, *time.Time) {
	var start, end *time.Time

	if startStr != "" {
		if t, err := time.Parse("2006-01-02", startStr); err == nil {
			start = &t
		}
	}
	if endStr != "" {
		if t, err := time.Parse("2006-01-02", endStr); err == nil {
			// 结束日期取当天 23:59:59
			endOfDay := t.Add(24*time.Hour - time.Second)
			end = &endOfDay
		}
	}

	return start, end
}
