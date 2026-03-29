package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// FittingsService 装配业务逻辑层
type FittingsService struct {
	charRepo    *repository.EveCharacterRepository
	fittingRepo *repository.FittingsRepository
	sdeRepo     *repository.SdeRepository
}

func NewFittingsService() *FittingsService {
	return &FittingsService{
		charRepo:    repository.NewEveCharacterRepository(),
		fittingRepo: repository.NewFittingsRepository(),
		sdeRepo:     repository.NewSdeRepository(),
	}
}

// ─────────────────────────────────────────────
//  请求 & 响应结构
// ─────────────────────────────────────────────

// FittingsRequest 装配列表请求
type FittingsRequest struct {
	Language string `json:"language"`
}

// FittingItemResponse 装配物品条目
type FittingItemResponse struct {
	TypeID   int64  `json:"type_id"`
	TypeName string `json:"type_name"`
	Quantity int    `json:"quantity"`
	Flag     string `json:"flag"`
}

// FittingSlotGroup 按槽位分组的装配物品
type FittingSlotGroup struct {
	FlagName string                `json:"flag_name"`
	FlagText string                `json:"flag_text"`
	OrderID  int                   `json:"order_id"`
	Items    []FittingItemResponse `json:"items"`
}

// FittingResponse 单个装配响应
type FittingResponse struct {
	FittingID   int64              `json:"fitting_id"`
	CharacterID int64              `json:"character_id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	ShipTypeID  int64              `json:"ship_type_id"`
	ShipName    string             `json:"ship_name"`
	GroupID     int                `json:"group_id"`
	GroupName   string             `json:"group_name"`
	RaceID      int                `json:"race_id"`
	RaceName    string             `json:"race_name"`
	Slots       []FittingSlotGroup `json:"slots"`
}

// FittingsListResponse 装配列表响应
type FittingsListResponse struct {
	Total    int               `json:"total"`
	Fittings []FittingResponse `json:"fittings"`
}

// SaveFittingRequest 保存装配请求
type SaveFittingRequest struct {
	CharacterID int64  `json:"character_id" binding:"required"`
	FittingID   *int64 `json:"fitting_id"` // 有值则先删后增
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Description string `json:"description" binding:"max=500"`
	ShipTypeID  int64  `json:"ship_type_id" binding:"required"`
	Items       []struct {
		TypeID   int64  `json:"type_id" binding:"required"`
		Quantity int    `json:"quantity" binding:"required,min=1"`
		Flag     string `json:"flag" binding:"required"`
	} `json:"items" binding:"required,min=1"`
}

// ─────────────────────────────────────────────
//  槽位排序映射
// ─────────────────────────────────────────────

var slotOrder = map[string]struct {
	name    string
	text    string
	orderID int
}{
	"HiSlot":      {"HiSlot", "高能量槽", 1},
	"MedSlot":     {"MedSlot", "中能量槽", 2},
	"LoSlot":      {"LoSlot", "低能量槽", 3},
	"RigSlot":     {"RigSlot", "改装件槽", 4},
	"SubSystem":   {"SubSystem", "子系统槽", 5},
	"DroneBay":    {"DroneBay", "无人机舱", 6},
	"FighterBay":  {"FighterBay", "战斗机舱", 7},
	"Cargo":       {"Cargo", "货柜舱", 8},
	"ServiceSlot": {"ServiceSlot", "服务槽", 9},
}

func getFlagGroup(flag string) string {
	for prefix := range slotOrder {
		if len(flag) >= len(prefix) && flag[:len(prefix)] == prefix {
			return prefix
		}
	}
	return flag
}

// ─────────────────────────────────────────────
//  业务方法
// ─────────────────────────────────────────────

// validateCharacterOwnership 校验人物归属
func (s *FittingsService) validateCharacterOwnership(userID uint, characterID int64) (*model.EveCharacter, error) {
	chars, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return nil, errors.New("获取人物列表失败")
	}
	for i := range chars {
		if chars[i].CharacterID == characterID {
			return &chars[i], nil
		}
	}
	return nil, errors.New("该人物不属于当前用户")
}

// GetFittings 获取用户名下所有人物的装配列表
func (s *FittingsService) GetFittings(userID uint, req *FittingsRequest) (*FittingsListResponse, error) {
	lang := req.Language
	if lang == "" {
		lang = "zh"
	}

	// 获取用户所有人物
	chars, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return nil, errors.New("获取人物列表失败")
	}
	if len(chars) == 0 {
		return &FittingsListResponse{Fittings: []FittingResponse{}}, nil
	}
	charIDs := make([]int64, 0, len(chars))
	for _, c := range chars {
		charIDs = append(charIDs, c.CharacterID)
	}

	// 查询所有装配
	fittings, err := s.fittingRepo.ListByCharacterIDs(charIDs)
	if err != nil {
		return nil, err
	}

	// 查询所有装配物品
	allItems, err := s.fittingRepo.GetItemsByCharacterIDs(charIDs)
	if err != nil {
		return nil, err
	}

	// 按 fitting_id+character_id 分组 items
	type fittingKey struct {
		FittingID   int64
		CharacterID int64
	}
	itemMap := make(map[fittingKey][]model.EveCharacterFittingItem)
	for _, item := range allItems {
		key := fittingKey{item.FittingID, item.CharacterID}
		itemMap[key] = append(itemMap[key], item)
	}

	// 收集所有 typeID 用于 SDE 翻译
	typeIDSet := make(map[int]struct{})
	for _, f := range fittings {
		typeIDSet[int(f.ShipTypeID)] = struct{}{}
	}
	for _, item := range allItems {
		typeIDSet[int(item.TypeID)] = struct{}{}
	}
	typeIDs := make([]int, 0, len(typeIDSet))
	for id := range typeIDSet {
		typeIDs = append(typeIDs, id)
	}

	// SDE 翻译
	typeInfoMap := make(map[int]repository.TypeInfo)
	if len(typeIDs) > 0 {
		typeInfos, err := s.sdeRepo.GetTypes(typeIDs, nil, lang)
		if err == nil {
			for _, t := range typeInfos {
				typeInfoMap[t.TypeID] = t
			}
		}
	}

	// 获取种族信息
	races, _ := s.sdeRepo.GetAllRaces()
	raceMap := make(map[int]string)
	for _, rc := range races {
		raceMap[rc.RaceID] = rc.RaceName
	}

	// 获取舰船 raceID（通过 SDE invTypes 的 raceID 字段）
	shipRaceMap := make(map[int]int) // typeID -> raceID
	if len(typeIDs) > 0 {
		ships, err := s.sdeRepo.GetShipsByCategoryID(lang)
		if err == nil {
			for _, sh := range ships {
				shipRaceMap[sh.TypeID] = sh.RaceID
			}
		}
	}

	// 组装响应
	result := &FittingsListResponse{
		Total:    len(fittings),
		Fittings: make([]FittingResponse, 0, len(fittings)),
	}

	for _, f := range fittings {
		shipInfo := typeInfoMap[int(f.ShipTypeID)]
		raceID := shipRaceMap[int(f.ShipTypeID)]

		resp := FittingResponse{
			FittingID:   f.FittingID,
			CharacterID: f.CharacterID,
			Name:        f.Name,
			Description: f.Description,
			ShipTypeID:  f.ShipTypeID,
			ShipName:    shipInfo.TypeName,
			GroupID:     shipInfo.GroupID,
			GroupName:   shipInfo.GroupName,
			RaceID:      raceID,
			RaceName:    raceMap[raceID],
		}

		// 构建槽位分组
		key := fittingKey{f.FittingID, f.CharacterID}
		items := itemMap[key]
		slotGroupMap := make(map[string]*FittingSlotGroup)

		for _, item := range items {
			flagGroup := getFlagGroup(item.Flag)
			sg, ok := slotGroupMap[flagGroup]
			if !ok {
				info := slotOrder[flagGroup]
				if info.name == "" {
					info.name = flagGroup
					info.text = flagGroup
					info.orderID = 99
				}
				sg = &FittingSlotGroup{
					FlagName: info.name,
					FlagText: info.text,
					OrderID:  info.orderID,
					Items:    []FittingItemResponse{},
				}
				slotGroupMap[flagGroup] = sg
			}

			itemInfo := typeInfoMap[int(item.TypeID)]
			sg.Items = append(sg.Items, FittingItemResponse{
				TypeID:   item.TypeID,
				TypeName: itemInfo.TypeName,
				Quantity: item.Quantity,
				Flag:     item.Flag,
			})
		}

		// 排序槽位组
		slots := make([]FittingSlotGroup, 0, len(slotGroupMap))
		for _, sg := range slotGroupMap {
			slots = append(slots, *sg)
		}
		// 简单冒泡排序
		for i := 0; i < len(slots); i++ {
			for j := i + 1; j < len(slots); j++ {
				if slots[j].OrderID < slots[i].OrderID {
					slots[i], slots[j] = slots[j], slots[i]
				}
			}
		}
		resp.Slots = slots
		result.Fittings = append(result.Fittings, resp)
	}

	return result, nil
}

// SaveFitting 保存装配（同步 ESI + 数据库）
func (s *FittingsService) SaveFitting(userID uint, req *SaveFittingRequest) (*FittingResponse, error) {
	char, err := s.validateCharacterOwnership(userID, req.CharacterID)
	if err != nil {
		return nil, err
	}

	if char.AccessToken == "" || char.TokenInvalid {
		return nil, errors.New("人物 Token 不可用，请重新绑定")
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	bgCtx := context.Background()

	// 如果有 fittingID，先删除 ESI 上的旧装配
	if req.FittingID != nil && *req.FittingID > 0 {
		deleteURL := fmt.Sprintf("%s/characters/%d/fittings/%d/", global.Config.EveSSO.ESIBaseURL, req.CharacterID, *req.FittingID)
		delReq, _ := http.NewRequestWithContext(bgCtx, http.MethodDelete, deleteURL, nil)
		delReq.Header.Set("Authorization", "Bearer "+char.AccessToken)
		resp, err := httpClient.Do(delReq)
		if err == nil {
			_ = resp.Body.Close()
		}

		// 删除数据库中的旧记录
		_ = s.fittingRepo.DeleteFitting(*req.FittingID, req.CharacterID)
	}

	// 在 ESI 上创建新装配
	esiItems := make([]map[string]interface{}, 0, len(req.Items))
	for _, item := range req.Items {
		esiItems = append(esiItems, map[string]interface{}{
			"type_id":  item.TypeID,
			"quantity": item.Quantity,
			"flag":     item.Flag,
		})
	}

	esiBody := map[string]interface{}{
		"name":         req.Name,
		"description":  req.Description,
		"ship_type_id": req.ShipTypeID,
		"items":        esiItems,
	}

	bodyBytes, err := json.Marshal(esiBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	postURL := fmt.Sprintf("%s/characters/%d/fittings/", global.Config.EveSSO.ESIBaseURL, req.CharacterID)
	postReq, err := http.NewRequestWithContext(bgCtx, http.MethodPost, postURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("构建 ESI 请求失败: %w", err)
	}
	postReq.Header.Set("Authorization", "Bearer "+char.AccessToken)
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(postReq)
	if err != nil {
		return nil, fmt.Errorf("ESI 创建装配失败: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取 ESI 响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("ESI 创建装配失败 (%d): %s", resp.StatusCode, string(respBody))
	}

	var esiResp struct {
		FittingID int64 `json:"fitting_id"`
	}
	if err := json.Unmarshal(respBody, &esiResp); err != nil {
		return nil, fmt.Errorf("解析 ESI 响应失败: %w", err)
	}

	newFittingID := esiResp.FittingID

	// 同步到数据库
	fitting := &model.EveCharacterFitting{
		FittingID:   newFittingID,
		CharacterID: req.CharacterID,
		Name:        req.Name,
		ShipTypeID:  req.ShipTypeID,
		Description: req.Description,
	}

	items := make([]model.EveCharacterFittingItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, model.EveCharacterFittingItem{
			FittingID:   newFittingID,
			CharacterID: req.CharacterID,
			TypeID:      item.TypeID,
			Quantity:    item.Quantity,
			Flag:        item.Flag,
		})
	}

	if err := s.fittingRepo.SaveFitting(fitting, items); err != nil {
		return nil, fmt.Errorf("保存装配到数据库失败: %w", err)
	}

	return &FittingResponse{
		FittingID:   newFittingID,
		CharacterID: req.CharacterID,
		Name:        req.Name,
		Description: req.Description,
		ShipTypeID:  req.ShipTypeID,
	}, nil
}
