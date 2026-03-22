package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm/clause"
)

// ─────────────────────────────────────────────
//  请求 & 响应结构
// ─────────────────────────────────────────────

// InfoImplantsRequest 克隆体/植入体信息请求
type InfoImplantsRequest struct {
	CharacterID int64  `json:"character_id" binding:"required"`
	Language    string `json:"language"`
}

// InfoImplantsResponse 克隆体/植入体信息响应
type InfoImplantsResponse struct {
	// 基本克隆信息
	HomeLocation          *ImplantLocation `json:"home_location"`
	LastCloneJumpDate     *time.Time       `json:"last_clone_jump_date,omitempty"`
	LastStationChangeDate *time.Time       `json:"last_station_change_date,omitempty"`

	// 跳跃疲劳
	JumpFatigueExpire *time.Time `json:"jump_fatigue_expire,omitempty"`
	LastJumpDate      *time.Time `json:"last_jump_date,omitempty"`

	// 当前活跃植入体
	ActiveImplants []ImplantItem `json:"active_implants"`

	// 跳跃克隆体列表
	JumpClones []JumpCloneInfo `json:"jump_clones"`
}

// ImplantLocation 位置信息
type ImplantLocation struct {
	LocationID   int64  `json:"location_id"`
	LocationType string `json:"location_type"`
	LocationName string `json:"location_name"`
}

// ImplantItem 植入体条目
type ImplantItem struct {
	ImplantID   int    `json:"implant_id"`
	ImplantName string `json:"implant_name"`
}

// JumpCloneInfo 跳跃克隆体信息
type JumpCloneInfo struct {
	JumpCloneID int64           `json:"jump_clone_id"`
	Location    ImplantLocation `json:"location"`
	Implants    []ImplantItem   `json:"implants"`
}

// ─────────────────────────────────────────────
//  Service
// ─────────────────────────────────────────────

// CloneService 克隆体/植入体业务逻辑
type CloneService struct {
	charRepo  *repository.EveCharacterRepository
	cloneRepo *repository.CloneRepository
	sdeRepo   *repository.SdeRepository
	ssoSvc    *EveSSOService
	http      *http.Client
}

func NewCloneService() *CloneService {
	return &CloneService{
		charRepo:  repository.NewEveCharacterRepository(),
		cloneRepo: repository.NewCloneRepository(),
		sdeRepo:   repository.NewSdeRepository(),
		ssoSvc:    NewEveSSOService(),
		http:      &http.Client{Timeout: 30 * time.Second},
	}
}

// validateCharacterOwnership 校验角色归属
func (s *CloneService) validateCharacterOwnership(userID uint, characterID int64) error {
	chars, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return errors.New("获取角色列表失败")
	}
	for _, c := range chars {
		if c.CharacterID == characterID {
			return nil
		}
	}
	return errors.New("该角色不属于当前用户")
}

// GetCharacterImplants 获取角色克隆体/植入体信息
func (s *CloneService) GetCharacterImplants(userID uint, req *InfoImplantsRequest) (*InfoImplantsResponse, error) {
	// 校验角色归属
	if err := s.validateCharacterOwnership(userID, req.CharacterID); err != nil {
		return nil, err
	}

	lang := req.Language
	if lang == "" {
		lang = "zh"
	}

	result := &InfoImplantsResponse{}

	// 1. 获取克隆基础信息
	baseInfo, err := s.cloneRepo.GetCloneBaseInfo(req.CharacterID)
	if err != nil {
		return nil, errors.New("未找到克隆体信息，请等待数据刷新")
	}

	result.LastCloneJumpDate = baseInfo.LastCloneJumpDate
	result.LastStationChangeDate = baseInfo.LastStationChangeDate
	result.JumpFatigueExpire = baseInfo.JumpFatigueExpire
	result.LastJumpDate = baseInfo.LastJumpDate

	// 2. 解析基底空间站位置
	if baseInfo.HomeLocationID != 0 {
		result.HomeLocation = &ImplantLocation{
			LocationID:   baseInfo.HomeLocationID,
			LocationType: baseInfo.HomeLocationType,
		}
		result.HomeLocation.LocationName = s.resolveLocationName(req.CharacterID, baseInfo.HomeLocationID, baseInfo.HomeLocationType)
	}

	// 3. 获取所有植入体记录
	implants, err := s.cloneRepo.GetImplants(req.CharacterID)
	if err != nil {
		return nil, err
	}

	// 按 JumpCloneID 分组
	activeImplantIDs := make([]int, 0)
	jumpCloneMap := make(map[int64]*JumpCloneInfo)
	var jumpCloneOrder []int64

	for _, imp := range implants {
		if imp.JumpCloneID == 0 {
			// 当前活跃植入体（ImplantID=0 时跳过，表示无植入体）
			if imp.ImplantID != 0 {
				activeImplantIDs = append(activeImplantIDs, imp.ImplantID)
			}
		} else {
			jc, ok := jumpCloneMap[imp.JumpCloneID]
			if !ok {
				jc = &JumpCloneInfo{
					JumpCloneID: imp.JumpCloneID,
					Location: ImplantLocation{
						LocationID:   imp.LocationID,
						LocationType: imp.LocationType,
					},
					Implants: make([]ImplantItem, 0),
				}
				jumpCloneMap[imp.JumpCloneID] = jc
				jumpCloneOrder = append(jumpCloneOrder, imp.JumpCloneID)
			}
			// ImplantID=0 是占位行，表示该克隆无植入体，不加入列表
			if imp.ImplantID != 0 {
				jc.Implants = append(jc.Implants, ImplantItem{
					ImplantID: imp.ImplantID,
				})
			}
		}
	}

	// 4. 批量查询植入体名称（跳过 ImplantID=0 占位行）
	allImplantIDs := make([]int, 0)
	for _, imp := range implants {
		if imp.ImplantID != 0 {
			allImplantIDs = append(allImplantIDs, imp.ImplantID)
		}
	}
	implantNameMap := s.resolveImplantNames(allImplantIDs, lang)

	// 5. 组装当前活跃植入体
	result.ActiveImplants = make([]ImplantItem, 0, len(activeImplantIDs))
	for _, id := range activeImplantIDs {
		result.ActiveImplants = append(result.ActiveImplants, ImplantItem{
			ImplantID:   id,
			ImplantName: implantNameMap[id],
		})
	}

	// 6. 组装跳跃克隆体
	result.JumpClones = make([]JumpCloneInfo, 0, len(jumpCloneOrder))
	for _, jcID := range jumpCloneOrder {
		jc := jumpCloneMap[jcID]
		// 解析位置名称
		jc.Location.LocationName = s.resolveLocationName(req.CharacterID, jc.Location.LocationID, jc.Location.LocationType)
		// 填充植入体名称
		for i := range jc.Implants {
			jc.Implants[i].ImplantName = implantNameMap[jc.Implants[i].ImplantID]
		}
		result.JumpClones = append(result.JumpClones, *jc)
	}

	return result, nil
}

// resolveImplantNames 批量查询植入体名称
func (s *CloneService) resolveImplantNames(typeIDs []int, lang string) map[int]string {
	result := make(map[int]string)
	if len(typeIDs) == 0 {
		return result
	}

	// 去重
	unique := make(map[int]struct{})
	deduped := make([]int, 0)
	for _, id := range typeIDs {
		if _, ok := unique[id]; !ok {
			unique[id] = struct{}{}
			deduped = append(deduped, id)
		}
	}

	published := true
	types, err := s.sdeRepo.GetTypes(deduped, &published, lang)
	if err != nil {
		return result
	}
	for _, t := range types {
		result[t.TypeID] = t.TypeName
	}
	return result
}

// resolveLocationName 解析位置名称（建筑/空间站）
func (s *CloneService) resolveLocationName(characterID, locationID int64, locationType string) string {
	if locationID == 0 {
		return ""
	}

	if locationType == "structure" {
		// 先从本地缓存查（15天内有效）
		structure, err := s.cloneRepo.GetStructureByID(locationID)
		if err == nil && structure.StructureName != "" {
			return structure.StructureName
		}

		// 缓存未命中，从 ESI 获取
		return s.fetchAndCacheStructure(characterID, locationID)
	}

	// station（NPC 空间站）— 通过 SDE 查询翻译
	if locationType == "station" {
		names, err := s.sdeRepo.GetNames(map[string][]int{
			"solar_system": {int(locationID)},
		}, "zh")
		if err == nil {
			if solarNames, ok := names["solar_system"]; ok {
				if name, ok := solarNames[int(locationID)]; ok {
					return name
				}
			}
		}
		// station ID 通常在 60000000–61000000 范围，尝试通过 staStations 表
		return s.resolveStationName(locationID)
	}

	return fmt.Sprintf("Unknown-%d", locationID)
}

// resolveStationName 查询 NPC 空间站名称
func (s *CloneService) resolveStationName(stationID int64) string {
	var name string
	err := global.DB.Table(`"staStations"`).
		Select(`"stationName"`).
		Where(`"stationID" = ?`, stationID).
		Scan(&name).Error
	if err != nil || name == "" {
		return fmt.Sprintf("Station-%d", stationID)
	}
	return name
}

// fetchAndCacheStructure 从 ESI 获取建筑详情并入库
func (s *CloneService) fetchAndCacheStructure(characterID, structureID int64) string {
	// 获取角色的 access token
	accessToken, err := s.ssoSvc.GetValidToken(context.Background(), characterID)
	if err != nil {
		global.Logger.Warn("[Clone] 获取 access token 失败",
			zap.Int64("character_id", characterID),
			zap.Error(err),
		)
		return fmt.Sprintf("Structure-%d", structureID)
	}

	// 从 ESI 获取建筑详情
	type structureDetail struct {
		Name     string `json:"name"`
		OwnerID  int64  `json:"owner_id"`
		Position struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
			Z float64 `json:"z"`
		} `json:"position"`
		SolarSystemID int64 `json:"solar_system_id"`
		TypeID        int64 `json:"type_id"`
	}

	var detail structureDetail
	structurePath := fmt.Sprintf("/universe/structures/%d/", structureID)
	if err := s.esiGet(context.Background(), structurePath, accessToken, &detail); err != nil {
		global.Logger.Warn("[Clone] 获取建筑详情失败",
			zap.Int64("structure_id", structureID),
			zap.Error(err),
		)
		return fmt.Sprintf("Structure-%d", structureID)
	}

	// 入库缓存
	record := &model.EveStructure{
		StructureID:   structureID,
		StructureName: detail.Name,
		OwnerID:       detail.OwnerID,
		TypeID:        detail.TypeID,
		SolarSystemID: detail.SolarSystemID,
		X:             detail.Position.X,
		Y:             detail.Position.Y,
		Z:             detail.Position.Z,
		UpdateAt:      time.Now().Unix(),
	}
	if err := global.DB.Clauses(clause.OnConflict{UpdateAll: true}).Create(record).Error; err != nil {
		global.Logger.Warn("[Clone] 缓存建筑信息失败",
			zap.Int64("structure_id", structureID),
			zap.Error(err),
		)
	}

	return detail.Name
}

// ─────────────────────────────────────────────
//  ESI HTTP 辅助方法（避免循环依赖 esi 包）
// ─────────────────────────────────────────────

func (s *CloneService) esiGet(ctx context.Context, path, accessToken string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, global.Config.EveSSO.ESIBaseURL+path, nil)
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
