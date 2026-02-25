package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const esiBaseURL = "https://esi.evetech.net/latest"

// FleetService 舰队业务逻辑层
type FleetService struct {
	repo     *repository.FleetRepository
	charRepo *repository.EveCharacterRepository
	ssoSvc   *EveSSOService
	http     *http.Client
}

func NewFleetService() *FleetService {
	return &FleetService{
		repo:     repository.NewFleetRepository(),
		charRepo: repository.NewEveCharacterRepository(),
		ssoSvc:   NewEveSSOService(),
		http:     &http.Client{Timeout: 30 * time.Second},
	}
}

// ─────────────────────────────────────────────
//  舰队 CRUD
// ─────────────────────────────────────────────

// CreateFleetRequest 创建舰队请求
type CreateFleetRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description"`
	StartAt     string  `json:"start_at" binding:"required"` // RFC3339
	EndAt       string  `json:"end_at" binding:"required"`   // RFC3339
	Importance  string  `json:"importance" binding:"required,oneof=strat_op cta other"`
	PapCount    float64 `json:"pap_count"`
	CharacterID int64   `json:"character_id" binding:"required"` // FC 角色 ID
}

// CreateFleet 创建舰队
func (s *FleetService) CreateFleet(userID uint, req *CreateFleetRequest) (*model.Fleet, error) {
	// 验证角色属于当前用户
	char, err := s.charRepo.GetByCharacterID(req.CharacterID)
	if err != nil {
		return nil, errors.New("角色不存在")
	}
	if char.UserID != userID {
		return nil, errors.New("该角色不属于当前用户")
	}

	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		return nil, errors.New("起始时间格式错误（需 RFC3339）")
	}
	endAt, err := time.Parse(time.RFC3339, req.EndAt)
	if err != nil {
		return nil, errors.New("结束时间格式错误（需 RFC3339）")
	}

	if endAt.Before(startAt) {
		return nil, errors.New("结束时间不能早于起始时间")
	}

	fleet := &model.Fleet{
		ID:              uuid.New().String(),
		Title:           req.Title,
		Description:     req.Description,
		StartAt:         startAt,
		EndAt:           endAt,
		Importance:      req.Importance,
		PapCount:        req.PapCount,
		FCUserID:        userID,
		FCCharacterID:   req.CharacterID,
		FCCharacterName: char.CharacterName,
	}

	// 自动从 ESI 拉取当前 ESI 舰队 ID
	ctx := context.Background()
	if accessToken, tokenErr := s.ssoSvc.GetValidToken(ctx, req.CharacterID); tokenErr == nil {
		path := fmt.Sprintf("/characters/%d/fleet/", req.CharacterID)
		var info CharacterFleetInfo
		if esiErr := s.esiGet(ctx, path, accessToken, &info); esiErr == nil {
			fleet.ESIFleetID = &info.FleetID
		} else {
			global.Logger.Warn("CreateFleet: 拉取 ESI fleet_id 失败", zap.Error(esiErr))
		}
	} else {
		global.Logger.Warn("CreateFleet: 获取 Token 失败，跳过 ESI fleet_id", zap.Error(tokenErr))
	}

	if err := s.repo.Create(fleet); err != nil {
		return nil, err
	}

	// 确保 FC 用户有钱包
	_, _ = s.repo.GetOrCreateWallet(userID)

	return fleet, nil
}

// UpdateFleetRequest 更新舰队请求
type UpdateFleetRequest struct {
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	StartAt     *string  `json:"start_at"`
	EndAt       *string  `json:"end_at"`
	Importance  *string  `json:"importance"`
	PapCount    *float64 `json:"pap_count"`
	CharacterID *int64   `json:"character_id"`
	ESIFleetID  *int64   `json:"esi_fleet_id"`
}

// UpdateFleet 更新舰队信息
func (s *FleetService) UpdateFleet(fleetID string, userID uint, userRole string, req *UpdateFleetRequest) (*model.Fleet, error) {
	fleet, err := s.repo.GetByID(fleetID)
	if err != nil {
		return nil, errors.New("舰队不存在")
	}

	if !s.canManageFleet(fleet, userID, userRole) {
		return nil, errors.New("权限不足")
	}

	if req.Title != nil {
		fleet.Title = *req.Title
	}
	if req.Description != nil {
		fleet.Description = *req.Description
	}
	if req.StartAt != nil {
		t, err := time.Parse(time.RFC3339, *req.StartAt)
		if err != nil {
			return nil, errors.New("起始时间格式错误")
		}
		fleet.StartAt = t
	}
	if req.EndAt != nil {
		t, err := time.Parse(time.RFC3339, *req.EndAt)
		if err != nil {
			return nil, errors.New("结束时间格式错误")
		}
		fleet.EndAt = t
	}
	if req.Importance != nil {
		fleet.Importance = *req.Importance
	}
	if req.PapCount != nil {
		fleet.PapCount = *req.PapCount
	}
	if req.CharacterID != nil {
		char, err := s.charRepo.GetByCharacterID(*req.CharacterID)
		if err != nil {
			return nil, errors.New("角色不存在")
		}
		if char.UserID != userID && !model.HasRole(userRole, model.RoleAdmin) {
			return nil, errors.New("该角色不属于当前用户")
		}
		fleet.FCCharacterID = *req.CharacterID
		fleet.FCCharacterName = char.CharacterName
	}
	if req.ESIFleetID != nil {
		fleet.ESIFleetID = req.ESIFleetID
	}

	if err := s.repo.Update(fleet); err != nil {
		return nil, err
	}
	return fleet, nil
}

// DeleteFleet 删除舰队
func (s *FleetService) DeleteFleet(fleetID string, userID uint, userRole string) error {
	fleet, err := s.repo.GetByID(fleetID)
	if err != nil {
		return errors.New("舰队不存在")
	}
	if !s.canManageFleet(fleet, userID, userRole) {
		return errors.New("权限不足")
	}
	return s.repo.SoftDelete(fleetID)
}

// GetFleet 获取舰队详情
func (s *FleetService) GetFleet(fleetID string) (*model.Fleet, error) {
	return s.repo.GetByID(fleetID)
}

// RefreshESIFleetID 从 ESI 刷新舰队的 esi_fleet_id 并持久化
func (s *FleetService) RefreshESIFleetID(fleetID string, userID uint, userRole string) (*model.Fleet, error) {
	fleet, err := s.repo.GetByID(fleetID)
	if err != nil {
		return nil, errors.New("舰队不存在")
	}
	if !s.canManageFleet(fleet, userID, userRole) {
		return nil, errors.New("权限不足")
	}

	ctx := context.Background()
	accessToken, err := s.ssoSvc.GetValidToken(ctx, fleet.FCCharacterID)
	if err != nil {
		return nil, fmt.Errorf("获取 Token 失败: %w", err)
	}

	path := fmt.Sprintf("/characters/%d/fleet/", fleet.FCCharacterID)
	var info CharacterFleetInfo
	if err := s.esiGet(ctx, path, accessToken, &info); err != nil {
		return nil, fmt.Errorf("ESI 查询失败: %w", err)
	}

	fleet.ESIFleetID = &info.FleetID
	if err := s.repo.Update(fleet); err != nil {
		return nil, err
	}
	return fleet, nil
}

// ListFleets 分页查询舰队列表
func (s *FleetService) ListFleets(page, pageSize int, filter repository.FleetFilter) ([]model.Fleet, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.repo.List(page, pageSize, filter)
}

// ─────────────────────────────────────────────
//  舰队成员
// ─────────────────────────────────────────────

// GetMembers 获取舰队成员列表
func (s *FleetService) GetMembers(fleetID string) ([]model.FleetMember, error) {
	return s.repo.ListMembers(fleetID)
}

// JoinFleet 通过邀请码加入舰队
func (s *FleetService) JoinFleet(code string, userID uint, characterID int64) error {
	invite, err := s.repo.GetInviteByCode(code)
	if err != nil {
		return errors.New("邀请链接无效")
	}
	if !invite.Active {
		return errors.New("邀请链接已失效")
	}
	if time.Now().After(invite.ExpiresAt) {
		return errors.New("邀请链接已过期")
	}

	// 验证角色属于当前用户
	char, err := s.charRepo.GetByCharacterID(characterID)
	if err != nil {
		return errors.New("角色不存在")
	}
	if char.UserID != userID {
		return errors.New("该角色不属于当前用户")
	}

	fleet, err := s.repo.GetByID(invite.FleetID)
	if err != nil {
		return errors.New("舰队不存在")
	}

	// 记录成员
	member := &model.FleetMember{
		FleetID:       fleet.ID,
		CharacterID:   characterID,
		CharacterName: char.CharacterName,
		UserID:        userID,
	}
	if err := s.repo.AddMember(member); err != nil {
		return err
	}

	// 尝试通过 ESI 邀请角色加入游戏内舰队
	if fleet.ESIFleetID != nil {
		go s.esiInviteMember(fleet, char)
	}

	return nil
}

// esiInviteMember 通过 ESI 邀请角色加入游戏内舰队
func (s *FleetService) esiInviteMember(fleet *model.Fleet, char *model.EveCharacter) {
	ctx := context.Background()

	accessToken, err := s.ssoSvc.GetValidToken(ctx, fleet.FCCharacterID)
	if err != nil {
		global.Logger.Warn("[Fleet] 获取 FC Token 失败",
			zap.String("fleet_id", fleet.ID),
			zap.Error(err),
		)
		return
	}

	path := fmt.Sprintf("/fleets/%d/members/", *fleet.ESIFleetID)
	body := map[string]interface{}{
		"character_id": char.CharacterID,
		"role":         "squad_member",
	}

	if err := s.esiPost(ctx, path, accessToken, body); err != nil {
		global.Logger.Warn("[Fleet] ESI 邀请成员失败",
			zap.String("fleet_id", fleet.ID),
			zap.Int64("character_id", char.CharacterID),
			zap.Error(err),
		)
	}
}

// ─────────────────────────────────────────────
//  PAP 发放
// ─────────────────────────────────────────────

// IssuePap 发放 PAP 到舰队所有成员
func (s *FleetService) IssuePap(fleetID string, userID uint, userRole string) error {
	fleet, err := s.repo.GetByID(fleetID)
	if err != nil {
		return errors.New("舰队不存在")
	}
	if !s.canManageFleet(fleet, userID, userRole) {
		return errors.New("权限不足")
	}
	if fleet.PapCount <= 0 {
		return errors.New("PAP 数量必须大于 0")
	}

	members, err := s.repo.ListMembers(fleetID)
	if err != nil {
		return err
	}
	if len(members) == 0 {
		return errors.New("舰队中没有成员")
	}

	// 删除旧的 PAP 记录（允许多次发放，只保留最后一次）
	if err := s.repo.DeletePapLogsByFleet(fleetID); err != nil {
		return err
	}

	// 创建新的 PAP 记录
	logs := make([]model.FleetPapLog, 0, len(members))
	for _, m := range members {
		logs = append(logs, model.FleetPapLog{
			FleetID:     fleetID,
			CharacterID: m.CharacterID,
			UserID:      m.UserID,
			PapCount:    fleet.PapCount,
			IssuedBy:    userID,
		})
	}
	if err := s.repo.CreatePapLogs(logs); err != nil {
		return err
	}

	// 尝试更新 ESI 舰队 MOTD
	if fleet.ESIFleetID != nil {
		go s.updateFleetMotd(fleet)
	}

	return nil
}

// updateFleetMotd 在 ESI 舰队 MOTD 中追加 PAP 发放记录
func (s *FleetService) updateFleetMotd(fleet *model.Fleet) {
	ctx := context.Background()

	accessToken, err := s.ssoSvc.GetValidToken(ctx, fleet.FCCharacterID)
	if err != nil {
		global.Logger.Warn("[Fleet] 获取 FC Token 失败（更新 MOTD）",
			zap.String("fleet_id", fleet.ID),
			zap.Error(err),
		)
		return
	}

	// 先获取当前 MOTD
	fleetPath := fmt.Sprintf("/fleets/%d/", *fleet.ESIFleetID)
	var fleetInfo struct {
		Motd string `json:"motd"`
	}
	if err := s.esiGet(ctx, fleetPath, accessToken, &fleetInfo); err != nil {
		global.Logger.Warn("[Fleet] 获取舰队信息失败",
			zap.String("fleet_id", fleet.ID),
			zap.Error(err),
		)
		return
	}

	// 追加 PAP 记录
	papNote := fmt.Sprintf("\n- %.1f PAP 已发放 %s -", fleet.PapCount, time.Now().Format("2006-01-02 15:04"))
	newMotd := fleetInfo.Motd + papNote

	body := map[string]interface{}{
		"motd": newMotd,
	}
	if err := s.esiPut(ctx, fleetPath, accessToken, body); err != nil {
		global.Logger.Warn("[Fleet] 更新舰队 MOTD 失败",
			zap.String("fleet_id", fleet.ID),
			zap.Error(err),
		)
	}
}

// GetPapLogs 获取舰队 PAP 发放记录
func (s *FleetService) GetPapLogs(fleetID string) ([]model.FleetPapLog, error) {
	return s.repo.ListPapLogsByFleet(fleetID)
}

// GetUserPapLogs 获取用户的 PAP 记录
func (s *FleetService) GetUserPapLogs(userID uint) ([]model.FleetPapLog, error) {
	return s.repo.ListPapLogsByUser(userID)
}

// ─────────────────────────────────────────────
//  从 ESI 拉取舰队成员并记录
// ─────────────────────────────────────────────

// ESIFleetMember ESI 舰队成员响应
type ESIFleetMember struct {
	CharacterID   int64  `json:"character_id"`
	JoinTime      string `json:"join_time"`
	Role          string `json:"role"`
	RoleName      string `json:"role_name"`
	ShipTypeID    int64  `json:"ship_type_id"`
	SolarSystemID int64  `json:"solar_system_id"`
	SquadID       int64  `json:"squad_id"`
	WingID        int64  `json:"wing_id"`
}

// SyncESIMembers 从 ESI 获取当前舰队成员并记录到数据库
func (s *FleetService) SyncESIMembers(fleetID string, userID uint, userRole string) ([]ESIFleetMember, error) {
	fleet, err := s.repo.GetByID(fleetID)
	if err != nil {
		return nil, errors.New("舰队不存在")
	}
	if !s.canManageFleet(fleet, userID, userRole) {
		return nil, errors.New("权限不足")
	}
	if fleet.ESIFleetID == nil {
		return nil, errors.New("未设置 ESI 舰队 ID")
	}

	ctx := context.Background()
	accessToken, err := s.ssoSvc.GetValidToken(ctx, fleet.FCCharacterID)
	if err != nil {
		return nil, fmt.Errorf("获取 FC Token 失败: %w", err)
	}

	path := fmt.Sprintf("/fleets/%d/members/", *fleet.ESIFleetID)
	var esiMembers []ESIFleetMember
	if err := s.esiGet(ctx, path, accessToken, &esiMembers); err != nil {
		return nil, fmt.Errorf("获取 ESI 舰队成员失败: %w", err)
	}

	// 将 ESI 成员记录到数据库
	for _, em := range esiMembers {
		char, err := s.charRepo.GetByCharacterID(em.CharacterID)
		if err != nil {
			// 角色不在系统中，跳过
			continue
		}
		shipTypeID := em.ShipTypeID
		solarSystemID := em.SolarSystemID
		member := &model.FleetMember{
			FleetID:       fleetID,
			CharacterID:   em.CharacterID,
			CharacterName: char.CharacterName,
			UserID:        char.UserID,
			ShipTypeID:    &shipTypeID,
			SolarSystemID: &solarSystemID,
		}
		_ = s.repo.AddMember(member)
	}

	return esiMembers, nil
}

// ─────────────────────────────────────────────
//  邀请链接
// ─────────────────────────────────────────────

// CreateInvite 创建舰队邀请链接
func (s *FleetService) CreateInvite(fleetID string, userID uint, userRole string) (*model.FleetInvite, error) {
	fleet, err := s.repo.GetByID(fleetID)
	if err != nil {
		return nil, errors.New("舰队不存在")
	}
	if !s.canManageFleet(fleet, userID, userRole) {
		return nil, errors.New("权限不足")
	}

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	code := hex.EncodeToString(b)

	// 邮证链接过期时间：取舰队结束时间，但至少保证 24 小时内有效
	expiresAt := fleet.EndAt
	if expiresAt.Before(time.Now().Add(24 * time.Hour)) {
		expiresAt = time.Now().Add(24 * time.Hour)
	}

	invite := &model.FleetInvite{
		FleetID:   fleetID,
		Code:      code,
		Active:    true,
		ExpiresAt: expiresAt,
	}

	if err := s.repo.CreateInvite(invite); err != nil {
		return nil, err
	}
	return invite, nil
}

// GetInvites 获取舰队邀请链接列表
func (s *FleetService) GetInvites(fleetID string) ([]model.FleetInvite, error) {
	return s.repo.ListInvitesByFleet(fleetID)
}

// DeactivateInvite 禁用邀请链接
func (s *FleetService) DeactivateInvite(inviteID uint, userID uint, userRole string) error {
	// 简单处理：admin 和 fc 都可以禁用
	if !model.HasRole(userRole, model.RoleFC) {
		return errors.New("权限不足")
	}
	return s.repo.DeactivateInvite(inviteID)
}

// ─────────────────────────────────────────────
//  钱包
// ─────────────────────────────────────────────

// GetWallet 获取用户钱包
func (s *FleetService) GetWallet(userID uint) (*model.SystemWallet, error) {
	return s.repo.GetOrCreateWallet(userID)
}

// GetWalletTransactions 获取用户钱包流水
func (s *FleetService) GetWalletTransactions(userID uint, page, pageSize int) ([]model.WalletTransaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.repo.ListWalletTransactions(userID, page, pageSize)
}

// ─────────────────────────────────────────────
//  权限判断
// ─────────────────────────────────────────────

// canManageFleet 判断用户是否有权管理该舰队（admin 或创建者）
func (s *FleetService) canManageFleet(fleet *model.Fleet, userID uint, userRole string) bool {
	if model.HasRole(userRole, model.RoleAdmin) {
		return true
	}
	return fleet.FCUserID == userID
}

// ─────────────────────────────────────────────
//  ESI: 获取角色当前舰队信息
// ─────────────────────────────────────────────

// CharacterFleetInfo 角色当前舰队信息
type CharacterFleetInfo struct {
	FleetID     int64  `json:"fleet_id"`
	FleetBossID int64  `json:"fleet_boss_id"`
	Role        string `json:"role"`
	SquadID     int64  `json:"squad_id"`
	WingID      int64  `json:"wing_id"`
}

// GetCharacterFleetInfo 获取角色当前所在的 ESI 舰队信息
func (s *FleetService) GetCharacterFleetInfo(userID uint, characterID int64) (*CharacterFleetInfo, error) {
	char, err := s.charRepo.GetByCharacterID(characterID)
	if err != nil {
		return nil, errors.New("角色不存在")
	}
	if char.UserID != userID {
		return nil, errors.New("该角色不属于当前用户")
	}

	ctx := context.Background()
	accessToken, err := s.ssoSvc.GetValidToken(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("获取 Token 失败: %w", err)
	}

	path := fmt.Sprintf("/characters/%d/fleet/", characterID)
	var info CharacterFleetInfo
	if err := s.esiGet(ctx, path, accessToken, &info); err != nil {
		return nil, fmt.Errorf("获取舰队信息失败: %w", err)
	}

	return &info, nil
}

// ─────────────────────────────────────────────
//  ESI HTTP 辅助方法（避免循环依赖 esi 包）
// ─────────────────────────────────────────────

// esiGet GET 请求并解析 JSON 响应
func (s *FleetService) esiGet(ctx context.Context, path, accessToken string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, esiBaseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ESI GET %s 返回 %d: %s", path, resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

// esiPost POST 请求（不期望响应体）
func (s *FleetService) esiPost(ctx context.Context, path, accessToken string, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, esiBaseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ESI POST %s 返回 %d: %s", path, resp.StatusCode, string(respBody))
	}
	return nil
}

// esiPut PUT 请求（不期望响应体）
func (s *FleetService) esiPut(ctx context.Context, path, accessToken string, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, esiBaseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ESI PUT %s 返回 %d: %s", path, resp.StatusCode, string(respBody))
	}
	return nil
}
