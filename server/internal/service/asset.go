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

// InfoAssetsRequest 资产请求
type InfoAssetsRequest struct {
	Language string `json:"language"`
}

// AssetLocationNode 前端资产树的「位置节点」
type AssetLocationNode struct {
	LocationID   int64           `json:"location_id"`
	LocationType string          `json:"location_type"` // station / structure / solar_system / other
	LocationName string          `json:"location_name"`
	Items        []AssetItemNode `json:"items"`
}

// AssetItemNode 前端资产树的「物品节点」
type AssetItemNode struct {
	ItemID          int64           `json:"item_id"`
	TypeID          int             `json:"type_id"`
	TypeName        string          `json:"type_name"`
	GroupName       string          `json:"group_name"`
	CategoryID      int             `json:"category_id"`
	Quantity        int             `json:"quantity"`
	LocationFlag    string          `json:"location_flag"`
	IsSingleton     bool            `json:"is_singleton"`
	IsBlueprintCopy *bool           `json:"is_blueprint_copy,omitempty"`
	AssetName       string          `json:"asset_name,omitempty"`
	CharacterID     int64           `json:"character_id"`
	CharacterName   string          `json:"character_name"`
	Children        []AssetItemNode `json:"children,omitempty"`
}

// InfoAssetsResponse 资产响应
type InfoAssetsResponse struct {
	TotalItems int                 `json:"total_items"`
	Locations  []AssetLocationNode `json:"locations"`
}

// ─────────────────────────────────────────────
//  Service
// ─────────────────────────────────────────────

// AssetService 资产业务逻辑
type AssetService struct {
	charRepo  *repository.EveCharacterRepository
	assetRepo *repository.AssetRepository
	sdeRepo   *repository.SdeRepository
	ssoSvc    *EveSSOService
	http      *http.Client
}

func NewAssetService() *AssetService {
	return &AssetService{
		charRepo:  repository.NewEveCharacterRepository(),
		assetRepo: repository.NewAssetRepository(),
		sdeRepo:   repository.NewSdeRepository(),
		ssoSvc:    NewEveSSOService(),
		http:      &http.Client{Timeout: 30 * time.Second},
	}
}

// GetUserAssets 获取用户名下所有人物的资产汇总
func (s *AssetService) GetUserAssets(userID uint, req *InfoAssetsRequest) (*InfoAssetsResponse, error) {
	lang := req.Language
	if lang == "" {
		lang = "zh"
	}

	// 1. 获取用户的所有人物
	chars, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return nil, errors.New("获取人物列表失败")
	}
	if len(chars) == 0 {
		return &InfoAssetsResponse{Locations: []AssetLocationNode{}}, nil
	}

	charIDs := make([]int64, 0, len(chars))
	charNameMap := make(map[int64]string)
	for _, c := range chars {
		charIDs = append(charIDs, c.CharacterID)
		charNameMap[c.CharacterID] = c.CharacterName
	}

	// 2. 获取所有人物的资产
	allAssets, err := s.assetRepo.GetAssetsByCharacterIDs(charIDs)
	if err != nil {
		return nil, errors.New("获取资产数据失败")
	}
	if len(allAssets) == 0 {
		return &InfoAssetsResponse{Locations: []AssetLocationNode{}}, nil
	}

	// 3. 收集所有 typeID 查名称
	typeIDSet := make(map[int]struct{})
	for _, a := range allAssets {
		typeIDSet[a.TypeID] = struct{}{}
	}
	typeIDs := make([]int, 0, len(typeIDSet))
	for id := range typeIDSet {
		typeIDs = append(typeIDs, id)
	}

	typeInfoMap := make(map[int]repository.TypeInfo)
	const typeBatch = 500
	for i := 0; i < len(typeIDs); i += typeBatch {
		end := i + typeBatch
		if end > len(typeIDs) {
			end = len(typeIDs)
		}
		infos, err := s.sdeRepo.GetTypes(typeIDs[i:end], nil, lang)
		if err == nil {
			for _, info := range infos {
				typeInfoMap[info.TypeID] = info
			}
		}
	}

	// 4. 构建 item_id -> asset map 以及 parent-child 关系
	itemMap := make(map[int64]*model.EveCharacterAsset)
	for i := range allAssets {
		itemMap[allAssets[i].ItemID] = &allAssets[i]
	}

	// 5. 建立位置分组
	//    根物品: location_type == station / solar_system / other
	//    子物品: location_type == item (location_id 是父物品的 item_id)

	// 先找所有根位置 ID（非 item 类型的 location）
	rootLocationIDs := make(map[int64]string)                // locationID -> locationType
	childrenMap := make(map[int64][]model.EveCharacterAsset) // parentItemID -> children

	for _, a := range allAssets {
		if a.LocationType == "item" {
			childrenMap[a.LocationID] = append(childrenMap[a.LocationID], a)
		} else {
			rootLocationIDs[a.LocationID] = a.LocationType
		}
	}

	// 6. 解析位置名称
	locationNames := make(map[int64]string)
	for locID, locType := range rootLocationIDs {
		locationNames[locID] = s.resolveLocationName(chars, locID, locType)
	}

	// 7. 按位置分组根物品
	locationItemsMap := make(map[int64][]model.EveCharacterAsset)
	for _, a := range allAssets {
		if a.LocationType != "item" {
			locationItemsMap[a.LocationID] = append(locationItemsMap[a.LocationID], a)
		}
	}

	// 8. 递归构建资产树
	var buildChildren func(parentItemID int64) []AssetItemNode
	buildChildren = func(parentItemID int64) []AssetItemNode {
		children, ok := childrenMap[parentItemID]
		if !ok {
			return nil
		}
		result := make([]AssetItemNode, 0, len(children))
		for _, c := range children {
			tInfo := typeInfoMap[c.TypeID]
			node := AssetItemNode{
				ItemID:          c.ItemID,
				TypeID:          c.TypeID,
				TypeName:        tInfo.TypeName,
				GroupName:       tInfo.GroupName,
				CategoryID:      tInfo.CategoryID,
				Quantity:        c.Quantity,
				LocationFlag:    c.LocationFlag,
				IsSingleton:     c.IsSingleton,
				IsBlueprintCopy: c.IsBlueprintCopy,
				AssetName:       c.AssetName,
				CharacterID:     c.CharacterID,
				CharacterName:   charNameMap[c.CharacterID],
				Children:        buildChildren(c.ItemID),
			}
			result = append(result, node)
		}
		return result
	}

	// 9. 组装最终响应
	locations := make([]AssetLocationNode, 0)
	for locID, items := range locationItemsMap {
		locNode := AssetLocationNode{
			LocationID:   locID,
			LocationType: rootLocationIDs[locID],
			LocationName: locationNames[locID],
			Items:        make([]AssetItemNode, 0, len(items)),
		}
		for _, a := range items {
			tInfo := typeInfoMap[a.TypeID]
			node := AssetItemNode{
				ItemID:          a.ItemID,
				TypeID:          a.TypeID,
				TypeName:        tInfo.TypeName,
				GroupName:       tInfo.GroupName,
				CategoryID:      tInfo.CategoryID,
				Quantity:        a.Quantity,
				LocationFlag:    a.LocationFlag,
				IsSingleton:     a.IsSingleton,
				IsBlueprintCopy: a.IsBlueprintCopy,
				AssetName:       a.AssetName,
				CharacterID:     a.CharacterID,
				CharacterName:   charNameMap[a.CharacterID],
				Children:        buildChildren(a.ItemID),
			}
			locNode.Items = append(locNode.Items, node)
		}
		locations = append(locations, locNode)
	}

	return &InfoAssetsResponse{
		TotalItems: len(allAssets),
		Locations:  locations,
	}, nil
}

// ─────────────────────────────────────────────
//  位置解析
// ─────────────────────────────────────────────

// resolveLocationName 解析位置名称
func (s *AssetService) resolveLocationName(chars []model.EveCharacter, locationID int64, locationType string) string {
	if locationID == 0 {
		return ""
	}

	switch locationType {
	case "station":
		return s.resolveStationName(locationID)
	case "solar_system":
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
		return fmt.Sprintf("System-%d", locationID)
	case "other":
		// 太空中的建筑，尝试从建筑表查询
		return s.resolveStructureName(chars, locationID)
	default:
		// 可能是玩家建筑
		return s.resolveStructureName(chars, locationID)
	}
}

// resolveStationName 查询 NPC 空间站名称
func (s *AssetService) resolveStationName(stationID int64) string {
	// 先查缓存表
	station, err := s.assetRepo.GetStationByID(stationID)
	if err == nil && station.StationName != "" {
		return station.StationName
	}

	// 从 SDE staStations 表查
	var name string
	if err := global.DB.Table(`"staStations"`).
		Select(`"stationName"`).
		Where(`"stationID" = ?`, stationID).
		Scan(&name).Error; err == nil && name != "" {
		// 缓存到 eve_stations 表
		if err := s.assetRepo.UpsertStation(&model.EveStation{
			StationID:   stationID,
			StationName: name,
			UpdateAt:    time.Now().Unix(),
		}); err != nil {
			global.Logger.Warn("[Asset] 缓存空间站信息失败", zap.Int64("station_id", stationID), zap.Error(err))
		}
		return name
	}

	// 从 ESI 获取
	return s.fetchAndCacheStation(stationID)
}

// resolveStructureName 查询玩家建筑名称
func (s *AssetService) resolveStructureName(chars []model.EveCharacter, structureID int64) string {
	// 先查本地缓存
	structure, err := s.assetRepo.GetStructureByID(structureID)
	if err == nil && structure.StructureName != "" {
		return structure.StructureName
	}

	// 尝试用任一人物的 token 从 ESI 获取
	for _, c := range chars {
		accessToken, err := s.ssoSvc.GetValidToken(context.Background(), c.CharacterID)
		if err != nil {
			continue
		}
		name := s.fetchAndCacheStructure(c.CharacterID, structureID, accessToken)
		if name != "" && name != fmt.Sprintf("Structure-%d", structureID) {
			return name
		}
	}

	return fmt.Sprintf("Structure-%d", structureID)
}

// fetchAndCacheStation 从 ESI 获取空间站详情并入库缓存
func (s *AssetService) fetchAndCacheStation(stationID int64) string {
	type stationDetail struct {
		Name          string `json:"name"`
		Owner         int64  `json:"owner"`
		SolarSystemID int64  `json:"system_id"`
		TypeID        int64  `json:"type_id"`
		Position      struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
			Z float64 `json:"z"`
		} `json:"position"`
	}

	var detail stationDetail
	path := fmt.Sprintf("/universe/stations/%d/", stationID)
	if err := s.esiGetPublic(context.Background(), path, &detail); err != nil {
		global.Logger.Warn("[Asset] 获取空间站详情失败",
			zap.Int64("station_id", stationID),
			zap.Error(err),
		)
		return fmt.Sprintf("Station-%d", stationID)
	}

	if err := s.assetRepo.UpsertStation(&model.EveStation{
		StationID:     stationID,
		StationName:   detail.Name,
		OwnerID:       detail.Owner,
		TypeID:        detail.TypeID,
		SolarSystemID: detail.SolarSystemID,
		X:             detail.Position.X,
		Y:             detail.Position.Y,
		Z:             detail.Position.Z,
		UpdateAt:      time.Now().Unix(),
	}); err != nil {
		global.Logger.Warn("[Asset] 缓存空间站信息失败", zap.Int64("station_id", stationID), zap.Error(err))
	}

	return detail.Name
}

// fetchAndCacheStructure 从 ESI 获取建筑详情并入库
func (s *AssetService) fetchAndCacheStructure(characterID, structureID int64, accessToken string) string {
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
	path := fmt.Sprintf("/universe/structures/%d/", structureID)
	if err := s.esiGet(context.Background(), path, accessToken, &detail); err != nil {
		global.Logger.Warn("[Asset] 获取建筑详情失败",
			zap.Int64("structure_id", structureID),
			zap.Error(err),
		)
		return fmt.Sprintf("Structure-%d", structureID)
	}

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
		global.Logger.Warn("[Asset] 缓存建筑信息失败",
			zap.Int64("structure_id", structureID),
			zap.Error(err),
		)
	}

	return detail.Name
}

// ─────────────────────────────────────────────
//  ESI HTTP 辅助
// ─────────────────────────────────────────────

func (s *AssetService) esiGet(ctx context.Context, path, accessToken string, out interface{}) error {
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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			global.Logger.Warn("[Asset] 关闭响应体失败", zap.Error(err))
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ESI GET %s 返回 %d: %s", path, resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

func (s *AssetService) esiGetPublic(ctx context.Context, path string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, global.Config.EveSSO.ESIBaseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			global.Logger.Warn("[Asset] 关闭响应体失败", zap.Error(err))
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ESI GET %s 返回 %d: %s", path, resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
