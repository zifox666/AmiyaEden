package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FleetConfigService 舰队配置业务逻辑层
type FleetConfigService struct {
	repo     *repository.FleetConfigRepository
	charRepo *repository.EveCharacterRepository
	sdeRepo  *repository.SdeRepository
	ssoSvc   *EveSSOService
	http     *http.Client
}

func NewFleetConfigService() *FleetConfigService {
	return &FleetConfigService{
		repo:     repository.NewFleetConfigRepository(),
		charRepo: repository.NewEveCharacterRepository(),
		sdeRepo:  repository.NewSdeRepository(),
		ssoSvc:   NewEveSSOService(),
		http:     &http.Client{Timeout: 30 * time.Second},
	}
}

// ─────────────────────────────────────────────
//  请求 & 响应结构
// ─────────────────────────────────────────────

// FleetConfigFittingReq 创建/更新中的装配条目请求（输入 EFT 文本，后端解析存储）
type FleetConfigFittingReq struct {
	FittingName string  `json:"fitting_name" binding:"required"`
	EFT         string  `json:"eft" binding:"required"` // 英文 EFT 格式，后端解析为 items
	SrpAmount   float64 `json:"srp_amount"`
}

// CreateFleetConfigRequest 创建舰队配置请求
type CreateFleetConfigRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Description string                  `json:"description"`
	Fittings    []FleetConfigFittingReq `json:"fittings" binding:"required,min=1"`
}

// UpdateFleetConfigRequest 更新舰队配置请求
type UpdateFleetConfigRequest struct {
	Name        *string                  `json:"name"`
	Description *string                  `json:"description"`
	Fittings    *[]FleetConfigFittingReq `json:"fittings"`
}

// FleetConfigFittingResp 装配条目响应（不含 EFT，通过专用端点获取）
type FleetConfigFittingResp struct {
	ID            uint    `json:"id"`
	FleetConfigID uint    `json:"fleet_config_id"`
	ShipTypeID    int64   `json:"ship_type_id"`
	FittingName   string  `json:"fitting_name"`
	SrpAmount     float64 `json:"srp_amount"`
}

// FleetConfigResp 舰队配置响应
type FleetConfigResp struct {
	ID          uint                     `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	CreatedBy   uint                     `json:"created_by"`
	CreatedAt   string                   `json:"created_at"`
	UpdatedAt   string                   `json:"updated_at"`
	Fittings    []FleetConfigFittingResp `json:"fittings"`
}

// ImportFittingRequest 从用户装配导入请求
type ImportFittingRequest struct {
	CharacterID int64 `json:"character_id" binding:"required"`
	FittingID   int64 `json:"fitting_id" binding:"required"`
}

// ImportFittingResponse 从用户装配导入响应（返回英文 EFT，供编辑表单预填充）
type ImportFittingResponse struct {
	FittingName string  `json:"fitting_name"`
	EFT         string  `json:"eft"` // 英文名称 EFT，可直接粘贴到编辑表单
	SrpAmount   float64 `json:"srp_amount"`
}

// ExportToESIRequest 导出装配到 ESI 请求
type ExportToESIRequest struct {
	CharacterID   int64 `json:"character_id" binding:"required"`
	FleetConfigID uint  `json:"fleet_config_id" binding:"required"`
	FittingItemID uint  `json:"fitting_item_id" binding:"required"` // fleet_config_fitting.id
}

// FleetConfigEFTFitting 单个装配的 EFT 结果
type FleetConfigEFTFitting struct {
	ID  uint   `json:"id"`
	EFT string `json:"eft"`
}

// FleetConfigEFTResponse GetFittingEFT 响应
type FleetConfigEFTResponse struct {
	Fittings []FleetConfigEFTFitting `json:"fittings"`
}

// ─────────────────────────────────────────────
//  CRUD
// ─────────────────────────────────────────────

// CreateFleetConfig 创建舰队配置
func (s *FleetConfigService) CreateFleetConfig(userID uint, req *CreateFleetConfigRequest) (*FleetConfigResp, error) {
	config := &model.FleetConfig{
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   userID,
	}

	fwis := make([]repository.FittingWithItems, 0, len(req.Fittings))
	for _, f := range req.Fittings {
		shipTypeID, items, err := s.parseEFTToFitting(f.EFT)
		if err != nil {
			return nil, fmt.Errorf("装配「%s」EFT 解析失败: %w", f.FittingName, err)
		}
		fwis = append(fwis, repository.FittingWithItems{
			Fitting: model.FleetConfigFitting{
				ShipTypeID:  shipTypeID,
				FittingName: f.FittingName,
				SrpAmount:   f.SrpAmount,
			},
			Items: items,
		})
	}

	if err := s.repo.Create(config, fwis); err != nil {
		return nil, err
	}

	fittings, _ := s.repo.ListFittingsByConfigID(config.ID)
	return s.buildResp(config, fittings), nil
}

// GetFleetConfig 获取舰队配置详情
func (s *FleetConfigService) GetFleetConfig(id uint) (*FleetConfigResp, error) {
	config, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("舰队配置不存在")
	}
	fittings, err := s.repo.ListFittingsByConfigID(id)
	if err != nil {
		return nil, err
	}
	return s.buildResp(config, fittings), nil
}

// ListFleetConfigs 分页查询舰队配置列表
func (s *FleetConfigService) ListFleetConfigs(page, pageSize int) ([]FleetConfigResp, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	configs, total, err := s.repo.List(page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	configIDs := make([]uint, 0, len(configs))
	for _, c := range configs {
		configIDs = append(configIDs, c.ID)
	}
	allFittings, err := s.repo.ListFittingsByConfigIDs(configIDs)
	if err != nil {
		return nil, 0, err
	}

	fittingMap := make(map[uint][]model.FleetConfigFitting)
	for _, f := range allFittings {
		fittingMap[f.FleetConfigID] = append(fittingMap[f.FleetConfigID], f)
	}

	result := make([]FleetConfigResp, 0, len(configs))
	for _, c := range configs {
		result = append(result, *s.buildResp(&c, fittingMap[c.ID]))
	}
	return result, total, nil
}

// UpdateFleetConfig 更新舰队配置
func (s *FleetConfigService) UpdateFleetConfig(id uint, userID uint, userRole string, req *UpdateFleetConfigRequest) (*FleetConfigResp, error) {
	config, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("舰队配置不存在")
	}
	if !s.canManage(config, userID, userRole) {
		return nil, errors.New("权限不足")
	}

	if req.Name != nil {
		config.Name = *req.Name
	}
	if req.Description != nil {
		config.Description = *req.Description
	}

	var fwis []repository.FittingWithItems
	if req.Fittings != nil {
		fwis = make([]repository.FittingWithItems, 0, len(*req.Fittings))
		for _, f := range *req.Fittings {
			shipTypeID, items, err := s.parseEFTToFitting(f.EFT)
			if err != nil {
				return nil, fmt.Errorf("装配「%s」EFT 解析失败: %w", f.FittingName, err)
			}
			fwis = append(fwis, repository.FittingWithItems{
				Fitting: model.FleetConfigFitting{
					FleetConfigID: id,
					ShipTypeID:    shipTypeID,
					FittingName:   f.FittingName,
					SrpAmount:     f.SrpAmount,
				},
				Items: items,
			})
		}
	} else {
		// Fittings 为 nil 表示不更新装配，保留现有
		existFittings, _ := s.repo.ListFittingsByConfigID(id)
		fittingIDs := make([]uint, len(existFittings))
		for i, f := range existFittings {
			fittingIDs[i] = f.ID
		}
		existItems, _ := s.repo.ListItemsByFittingIDs(fittingIDs)
		itemByFitting := make(map[uint][]model.FleetConfigFittingItem)
		for _, item := range existItems {
			itemByFitting[item.FleetConfigFittingID] = append(itemByFitting[item.FleetConfigFittingID], item)
		}
		for _, ef := range existFittings {
			fwis = append(fwis, repository.FittingWithItems{
				Fitting: ef,
				Items:   itemByFitting[ef.ID],
			})
		}
	}

	if err := s.repo.Update(config, fwis); err != nil {
		return nil, err
	}

	updatedFittings, _ := s.repo.ListFittingsByConfigID(id)
	return s.buildResp(config, updatedFittings), nil
}

// DeleteFleetConfig 删除舰队配置
func (s *FleetConfigService) DeleteFleetConfig(id uint, userID uint, userRole string) error {
	config, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("舰队配置不存在")
	}
	if !s.canManage(config, userID, userRole) {
		return errors.New("权限不足")
	}
	return s.repo.Delete(id)
}

// ─────────────────────────────────────────────
//  从用户装配导入
// ─────────────────────────────────────────────

// ImportFromUserFitting 从用户的 ESI 装配导入为英文 EFT（供编辑表单预填充）
func (s *FleetConfigService) ImportFromUserFitting(userID uint, req *ImportFittingRequest) (*ImportFittingResponse, error) {
	// 验证角色属于当前用户
	char, err := s.charRepo.GetByCharacterID(req.CharacterID)
	if err != nil {
		return nil, errors.New("角色不存在")
	}
	if char.UserID != userID {
		return nil, errors.New("该角色不属于当前用户")
	}

	fittingRepo := repository.NewFittingsRepository()

	// 获取装配列表
	fittings, err := fittingRepo.ListByCharacterIDs([]int64{req.CharacterID})
	if err != nil {
		return nil, errors.New("获取装配失败")
	}

	var target *model.EveCharacterFitting
	for i := range fittings {
		if fittings[i].FittingID == req.FittingID {
			target = &fittings[i]
			break
		}
	}
	if target == nil {
		return nil, errors.New("装配不存在")
	}

	// 获取物品
	items, err := fittingRepo.GetItemsByFittingAndCharacter(req.FittingID, req.CharacterID)
	if err != nil {
		return nil, errors.New("获取装配物品失败")
	}

	// 始终使用英文名称生成 EFT，方便后续 parseEFTToFitting 解析
	typeIDs := []int{int(target.ShipTypeID)}
	for _, item := range items {
		typeIDs = append(typeIDs, int(item.TypeID))
	}
	typeInfoMap := make(map[int]string)
	if typeInfos, sdeErr := s.sdeRepo.GetTypes(typeIDs, nil, "en"); sdeErr == nil {
		for _, t := range typeInfos {
			typeInfoMap[t.TypeID] = t.TypeName
		}
	}

	shipName := typeInfoMap[int(target.ShipTypeID)]
	if shipName == "" {
		shipName = fmt.Sprintf("TypeID:%d", target.ShipTypeID)
	}

	eft := buildEFT(shipName, target.Name, items, typeInfoMap)

	return &ImportFittingResponse{
		FittingName: target.Name,
		EFT:         eft,
		SrpAmount:   0,
	}, nil
}

// ─────────────────────────────────────────────
//  导出到 ESI
// ─────────────────────────────────────────────

// ExportToESI 将配置中的某个装配导出到 ESI（使用存储的 items，无需 EFT 解析）
func (s *FleetConfigService) ExportToESI(userID uint, req *ExportToESIRequest) error {
	// 验证角色属于当前用户
	char, err := s.charRepo.GetByCharacterID(req.CharacterID)
	if err != nil {
		return errors.New("角色不存在")
	}
	if char.UserID != userID {
		return errors.New("该角色不属于当前用户")
	}
	if char.AccessToken == "" || char.TokenInvalid {
		return errors.New("角色 Token 不可用，请重新绑定")
	}

	// 获取目标装配
	target, err := s.repo.GetFittingByID(req.FittingItemID)
	if err != nil {
		return errors.New("装配条目不存在")
	}
	if target.FleetConfigID != req.FleetConfigID {
		return errors.New("装配不属于该舰队配置")
	}

	// 直接使用存储的 items 构建 ESI 请求
	storedItems, err := s.repo.ListItemsByFittingIDs([]uint{target.ID})
	if err != nil {
		return errors.New("获取装配模块失败")
	}

	esiItems := make([]map[string]interface{}, 0, len(storedItems))
	for _, item := range storedItems {
		esiItems = append(esiItems, map[string]interface{}{
			"type_id":  item.TypeID,
			"quantity": item.Quantity,
			"flag":     item.Flag,
		})
	}

	esiBody := map[string]interface{}{
		"name":         target.FittingName,
		"description":  "",
		"ship_type_id": target.ShipTypeID,
		"items":        esiItems,
	}

	bodyBytes, err := json.Marshal(esiBody)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %w", err)
	}

	ctx := context.Background()
	accessToken, err := s.ssoSvc.GetValidToken(ctx, req.CharacterID)
	if err != nil {
		return fmt.Errorf("获取 Token 失败: %w", err)
	}

	postURL := fmt.Sprintf("https://esi.evetech.net/latest/characters/%d/fittings/", req.CharacterID)
	postReq, err := http.NewRequestWithContext(ctx, http.MethodPost, postURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("构建请求失败: %w", err)
	}
	postReq.Header.Set("Authorization", "Bearer "+accessToken)
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Accept", "application/json")

	resp, err := s.http.Do(postReq)
	if err != nil {
		return fmt.Errorf("ESI 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ESI 创建装配失败 (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// ─────────────────────────────────────────────
//  辅助方法
// ─────────────────────────────────────────────

func (s *FleetConfigService) canManage(config *model.FleetConfig, userID uint, userRole string) bool {
	if model.HasRole(userRole, model.RoleAdmin) {
		return true
	}
	if model.HasRole(userRole, model.RoleFC) {
		return true
	}
	if model.HasRole(userRole, model.RoleSRP) {
		return true
	}
	return config.CreatedBy == userID
}

func (s *FleetConfigService) buildResp(config *model.FleetConfig, fittings []model.FleetConfigFitting) *FleetConfigResp {
	fittingResps := make([]FleetConfigFittingResp, 0, len(fittings))
	for _, f := range fittings {
		fittingResps = append(fittingResps, FleetConfigFittingResp{
			ID:            f.ID,
			FleetConfigID: f.FleetConfigID,
			ShipTypeID:    f.ShipTypeID,
			FittingName:   f.FittingName,
			SrpAmount:     f.SrpAmount,
		})
	}
	return &FleetConfigResp{
		ID:          config.ID,
		Name:        config.Name,
		Description: config.Description,
		CreatedBy:   config.CreatedBy,
		CreatedAt:   config.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   config.UpdatedAt.Format(time.RFC3339),
		Fittings:    fittingResps,
	}
}

// ─────────────────────────────────────────────
//  EFT 解析 & 生成
// ─────────────────────────────────────────────

type eftHeader struct {
	ShipType    string
	FittingName string
}

// parseEFTHeader 解析 EFT 格式第一行: [Ship Type, Fitting Name]
func parseEFTHeader(eft string) *eftHeader {
	lines := strings.Split(strings.TrimSpace(eft), "\n")
	if len(lines) == 0 {
		return nil
	}
	first := strings.TrimSpace(lines[0])
	if !strings.HasPrefix(first, "[") || !strings.HasSuffix(first, "]") {
		return nil
	}
	inner := first[1 : len(first)-1]
	parts := strings.SplitN(inner, ",", 2)
	if len(parts) < 2 {
		return nil
	}
	return &eftHeader{
		ShipType:    strings.TrimSpace(parts[0]),
		FittingName: strings.TrimSpace(parts[1]),
	}
}

var countRegex = regexp.MustCompile(`^(.+?)\s+x(\d+)\s*$`)

// parseEFTToESIItems 已废弃 - 由 parseEFTToFitting 替代

// parseEFTToFitting 将英文 EFT 文本解析为 (shipTypeID, []FleetConfigFittingItem)
// EFT 各段以空行分隔，顺序为：低槽、中槽、高槽、钻孔槽、子系统、服务槽；
// 之后的段中若所有行都有 "x N" 后缀则视为无人机/货物。
func (s *FleetConfigService) parseEFTToFitting(eft string) (int64, []model.FleetConfigFittingItem, error) {
	header := parseEFTHeader(eft)
	if header == nil {
		return 0, nil, errors.New("EFT 格式错误：缺少 [舰船, 装配名] 头部")
	}

	lines := strings.Split(strings.TrimSpace(eft), "\n")
	bodyLines := lines[1:] // 跳过头部行

	// 按空行拆分成若干段
	var sections [][]string
	cur := []string{}
	for _, line := range bodyLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if len(cur) > 0 {
				sections = append(sections, cur)
				cur = []string{}
			}
		} else {
			// 跳过 [Empty Xxx Slot] 占位行
			if strings.HasPrefix(trimmed, "[Empty") {
				cur = append(cur, "") // 保留slot计数位置
			} else {
				cur = append(cur, trimmed)
			}
		}
	}
	if len(cur) > 0 {
		sections = append(sections, cur)
	}

	// 收集所有涉及的英文type名称
	nameSet := make(map[string]struct{})
	nameSet[header.ShipType] = struct{}{}
	for _, section := range sections {
		for _, line := range section {
			if line == "" {
				continue
			}
			name := line
			if m := countRegex.FindStringSubmatch(line); m != nil {
				name = strings.TrimSpace(m[1])
			}
			// 处理内联装填物: "Module, Charge"
			if strings.Contains(name, ",") {
				parts := strings.SplitN(name, ",", 2)
				nameSet[strings.TrimSpace(parts[0])] = struct{}{}
				nameSet[strings.TrimSpace(parts[1])] = struct{}{}
			} else {
				nameSet[name] = struct{}{}
			}
		}
	}

	names := make([]string, 0, len(nameSet))
	for n := range nameSet {
		if n != "" {
			names = append(names, n)
		}
	}

	nameToTypeID, err := s.sdeRepo.GetTypeIDsByNames(names)
	if err != nil {
		return 0, nil, fmt.Errorf("SDE 查询失败: %w", err)
	}

	shipTypeID, ok := nameToTypeID[header.ShipType]
	if !ok {
		return 0, nil, fmt.Errorf("未找到舰船类型「%s」，请确认 EFT 使用英文名称", header.ShipType)
	}

	// 定义插槽组顺序
	slotGroups := []struct {
		prefix string
		count  int
	}{
		{"LoSlot", 8},
		{"MedSlot", 8},
		{"HiSlot", 8},
		{"RigSlot", 3},
		{"SubSystemSlot", 4},
		{"ServiceSlot", 8},
	}

	var items []model.FleetConfigFittingItem
	sectionIdx := 0

	for _, section := range sections {
		// 判断是否为无人机/货物段（所有非空行均有 "x N" 后缀）
		isDroneSection := true
		for _, line := range section {
			if line == "" {
				continue
			}
			if countRegex.FindStringSubmatch(line) == nil {
				isDroneSection = false
				break
			}
		}

		if isDroneSection {
			for _, line := range section {
				if line == "" {
					continue
				}
				m := countRegex.FindStringSubmatch(line)
				if m == nil {
					continue
				}
				name := strings.TrimSpace(m[1])
				qty, _ := strconv.Atoi(m[2])
				typeID, exists := nameToTypeID[name]
				if !exists {
					continue
				}
				items = append(items, model.FleetConfigFittingItem{
					TypeID:   typeID,
					Quantity: qty,
					Flag:     "DroneBay",
				})
			}
		} else {
			if sectionIdx < len(slotGroups) {
				group := slotGroups[sectionIdx]
				slotCounter := 0
				for _, line := range section {
					if slotCounter >= group.count {
						break
					}
					if line == "" {
						// [Empty Slot] 占位：消耗一个槽位
						slotCounter++
						continue
					}

					flag := fmt.Sprintf("%s%d", group.prefix, slotCounter)

					// 处理内联装填物 "Module, Charge"
					moduleName := line
					chargeName := ""
					if strings.Contains(line, ",") {
						parts := strings.SplitN(line, ",", 2)
						moduleName = strings.TrimSpace(parts[0])
						chargeName = strings.TrimSpace(parts[1])
					}

					moduleTypeID, exists := nameToTypeID[moduleName]
					if exists {
						items = append(items, model.FleetConfigFittingItem{
							TypeID:   moduleTypeID,
							Quantity: 1,
							Flag:     flag,
						})
					}
					slotCounter++

					if chargeName != "" {
						chargeTypeID, exists := nameToTypeID[chargeName]
						if exists {
							items = append(items, model.FleetConfigFittingItem{
								TypeID:   chargeTypeID,
								Quantity: 1,
								Flag:     "Cargo",
							})
						}
					}
				}
				sectionIdx++
			}
		}
	}

	return shipTypeID, items, nil
}

// GetFittingEFT 返回舰队配置中所有装配的本地化 EFT 文本
func (s *FleetConfigService) GetFittingEFT(configID uint, lang string) (*FleetConfigEFTResponse, error) {
	fittings, err := s.repo.ListFittingsByConfigID(configID)
	if err != nil {
		return nil, err
	}
	if len(fittings) == 0 {
		return &FleetConfigEFTResponse{Fittings: []FleetConfigEFTFitting{}}, nil
	}

	fittingIDs := make([]uint, len(fittings))
	for i, f := range fittings {
		fittingIDs[i] = f.ID
	}

	storedItems, err := s.repo.ListItemsByFittingIDs(fittingIDs)
	if err != nil {
		return nil, err
	}

	// 收集所有 type ID
	typeIDSet := make(map[int]struct{})
	for _, f := range fittings {
		typeIDSet[int(f.ShipTypeID)] = struct{}{}
	}
	for _, item := range storedItems {
		typeIDSet[int(item.TypeID)] = struct{}{}
	}
	typeIDs := make([]int, 0, len(typeIDSet))
	for tid := range typeIDSet {
		typeIDs = append(typeIDs, tid)
	}

	if lang == "" {
		lang = "zh"
	}
	typeInfos, err := s.sdeRepo.GetTypes(typeIDs, nil, lang)
	if err != nil {
		return nil, fmt.Errorf("SDE 查询失败: %w", err)
	}
	typeNames := make(map[int]string, len(typeInfos))
	for _, t := range typeInfos {
		typeNames[t.TypeID] = t.TypeName
	}

	// 按 fitting ID 分组 items
	itemsByFitting := make(map[uint][]model.EveCharacterFittingItem)
	for _, item := range storedItems {
		itemsByFitting[item.FleetConfigFittingID] = append(
			itemsByFitting[item.FleetConfigFittingID],
			model.EveCharacterFittingItem{
				TypeID:   item.TypeID,
				Quantity: item.Quantity,
				Flag:     item.Flag,
			},
		)
	}

	result := &FleetConfigEFTResponse{
		Fittings: make([]FleetConfigEFTFitting, 0, len(fittings)),
	}
	for _, f := range fittings {
		shipName := typeNames[int(f.ShipTypeID)]
		if shipName == "" {
			shipName = fmt.Sprintf("TypeID:%d", f.ShipTypeID)
		}
		eft := buildEFT(shipName, f.FittingName, itemsByFitting[f.ID], typeNames)
		result.Fittings = append(result.Fittings, FleetConfigEFTFitting{
			ID:  f.ID,
			EFT: eft,
		})
	}
	return result, nil
}

// buildEFT 构建 EFT 格式文本
func buildEFT(shipName, fittingName string, items []model.EveCharacterFittingItem, typeNames map[int]string) string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("[%s, %s]\n", shipName, fittingName))

	// 按 flag 分组
	type slotItem struct {
		name     string
		quantity int
		flag     string
	}

	slotGroups := map[string][]slotItem{}

	for _, item := range items {
		typeName := typeNames[int(item.TypeID)]
		if typeName == "" {
			typeName = fmt.Sprintf("TypeID:%d", item.TypeID)
		}
		group := getFlagGroupForEFT(item.Flag)
		slotGroups[group] = append(slotGroups[group], slotItem{
			name:     typeName,
			quantity: item.Quantity,
			flag:     item.Flag,
		})
	}

	// 输出顺序: lo, med, hi, rig, subsystem, service, drone, fighter, cargo
	orderedGroups := []string{"LoSlot", "MedSlot", "HiSlot", "RigSlot", "SubSystem", "ServiceSlot", "DroneBay", "FighterBay", "Cargo"}

	firstSection := true
	for _, group := range orderedGroups {
		items, ok := slotGroups[group]
		if !ok || len(items) == 0 {
			continue
		}
		if !firstSection {
			buf.WriteString("\n")
		}
		firstSection = false

		isDroneOrCargo := group == "DroneBay" || group == "FighterBay" || group == "Cargo"

		for _, item := range items {
			if isDroneOrCargo && item.quantity > 1 {
				buf.WriteString(fmt.Sprintf("%s x%d\n", item.name, item.quantity))
			} else if isDroneOrCargo {
				buf.WriteString(fmt.Sprintf("%s x%d\n", item.name, item.quantity))
			} else {
				buf.WriteString(fmt.Sprintf("%s\n", item.name))
			}
		}
	}

	return buf.String()
}

// getFlagGroupForEFT 将 ESI flag 映射为 EFT 分组
func getFlagGroupForEFT(flag string) string {
	prefixes := []string{"HiSlot", "MedSlot", "LoSlot", "RigSlot", "SubSystem", "DroneBay", "FighterBay", "Cargo", "ServiceSlot"}
	for _, prefix := range prefixes {
		if len(flag) >= len(prefix) && flag[:len(prefix)] == prefix {
			return prefix
		}
	}
	return flag
}

// ─────────────────────────────────────────────
//  装备详情 & 设置
// ─────────────────────────────────────────────

// FittingItemReplacementResp 替代品响应
type FittingItemReplacementResp struct {
	ID       uint   `json:"id"`
	TypeID   int64  `json:"type_id"`
	TypeName string `json:"type_name"`
}

// FittingItemResp 装备物品详情响应
type FittingItemResp struct {
	ID                 uint                         `json:"id"`
	TypeID             int64                        `json:"type_id"`
	TypeName           string                       `json:"type_name"`
	Quantity           int                          `json:"quantity"`
	Flag               string                       `json:"flag"`
	FlagGroup          string                       `json:"flag_group"`
	Importance         string                       `json:"importance"`
	Penalty            string                       `json:"penalty"`
	ReplacementPenalty string                       `json:"replacement_penalty"`
	Replacements       []FittingItemReplacementResp `json:"replacements"`
}

// FittingItemsResponse 装配物品详情响应
type FittingItemsResponse struct {
	FittingID   uint              `json:"fitting_id"`
	FittingName string            `json:"fitting_name"`
	ShipTypeID  int64             `json:"ship_type_id"`
	Items       []FittingItemResp `json:"items"`
}

// GetFittingItems 获取装配物品详情（含重要性、惩罚、替代品）
func (s *FleetConfigService) GetFittingItems(configID, fittingID uint, lang string) (*FittingItemsResponse, error) {
	fitting, err := s.repo.GetFittingByID(fittingID)
	if err != nil {
		return nil, errors.New("装配不存在")
	}
	if fitting.FleetConfigID != configID {
		return nil, errors.New("装配不属于该配置")
	}

	items, err := s.repo.ListItemsByFittingIDs([]uint{fittingID})
	if err != nil {
		return nil, err
	}

	// 收集物品 ID 和替代品 ID
	itemIDs := make([]uint, len(items))
	for i, item := range items {
		itemIDs[i] = item.ID
	}

	replacements, err := s.repo.ListReplacementsByItemIDs(itemIDs)
	if err != nil {
		return nil, err
	}

	// 按 item ID 分组替代品
	repMap := make(map[uint][]model.FleetConfigFittingItemReplacement)
	for _, r := range replacements {
		repMap[r.FleetConfigFittingItemID] = append(repMap[r.FleetConfigFittingItemID], r)
	}

	// 收集所有 type_id 用于 SDE 查询
	if lang == "" {
		lang = "zh"
	}
	typeIDSet := make(map[int]struct{})
	for _, item := range items {
		typeIDSet[int(item.TypeID)] = struct{}{}
	}
	for _, r := range replacements {
		typeIDSet[int(r.TypeID)] = struct{}{}
	}
	typeIDs := make([]int, 0, len(typeIDSet))
	for tid := range typeIDSet {
		typeIDs = append(typeIDs, tid)
	}
	typeNames := make(map[int]string)
	if len(typeIDs) > 0 {
		if infos, sdeErr := s.sdeRepo.GetTypes(typeIDs, nil, lang); sdeErr == nil {
			for _, t := range infos {
				typeNames[t.TypeID] = t.TypeName
			}
		}
	}

	// 构建响应（过滤 DroneBay/FighterBay/Cargo）
	skipGroups := map[string]bool{"DroneBay": true, "FighterBay": true, "Cargo": true}
	respItems := make([]FittingItemResp, 0, len(items))
	for _, item := range items {
		group := getFlagGroupForEFT(item.Flag)
		if skipGroups[group] {
			continue
		}
		reps := repMap[item.ID]
		repResps := make([]FittingItemReplacementResp, len(reps))
		for j, r := range reps {
			repResps[j] = FittingItemReplacementResp{
				ID:       r.ID,
				TypeID:   r.TypeID,
				TypeName: typeNames[int(r.TypeID)],
			}
		}
		respItems = append(respItems, FittingItemResp{
			ID:                 item.ID,
			TypeID:             item.TypeID,
			TypeName:           typeNames[int(item.TypeID)],
			Quantity:           item.Quantity,
			Flag:               item.Flag,
			FlagGroup:          group,
			Importance:         item.Importance,
			Penalty:            item.Penalty,
			ReplacementPenalty: item.ReplacementPenalty,
			Replacements:       repResps,
		})
	}

	return &FittingItemsResponse{
		FittingID:   fitting.ID,
		FittingName: fitting.FittingName,
		ShipTypeID:  fitting.ShipTypeID,
		Items:       respItems,
	}, nil
}

// UpdateItemSettingsReq 单个物品设置更新请求
type UpdateItemSettingsReq struct {
	ID                 uint    `json:"id" binding:"required"`
	Importance         string  `json:"importance" binding:"required,oneof=required optional replaceable"`
	Penalty            string  `json:"penalty" binding:"required,oneof=half none"`
	ReplacementPenalty string  `json:"replacement_penalty" binding:"required,oneof=half none"`
	Replacements       []int64 `json:"replacements"` // replaceable 时的替代 type_id 列表
}

// UpdateFittingItemsSettingsRequest 批量更新装配物品设置请求
type UpdateFittingItemsSettingsRequest struct {
	Items []UpdateItemSettingsReq `json:"items" binding:"required,min=1"`
}

// UpdateFittingItemsSettings 批量更新装配物品的重要性、惩罚和替代品
func (s *FleetConfigService) UpdateFittingItemsSettings(configID, fittingID, userID uint, userRole string, req *UpdateFittingItemsSettingsRequest) error {
	config, err := s.repo.GetByID(configID)
	if err != nil {
		return errors.New("配置不存在")
	}
	if !s.canManage(config, userID, userRole) {
		return errors.New("权限不足")
	}

	fitting, err := s.repo.GetFittingByID(fittingID)
	if err != nil {
		return errors.New("装配不存在")
	}
	if fitting.FleetConfigID != configID {
		return errors.New("装配不属于该配置")
	}

	updates := make([]repository.ItemSettingUpdate, len(req.Items))
	for i, item := range req.Items {
		// replaceable 才允许有替代品
		reps := item.Replacements
		if item.Importance != model.FittingItemReplaceable {
			reps = nil
		}
		updates[i] = repository.ItemSettingUpdate{
			ID:                 item.ID,
			Importance:         item.Importance,
			Penalty:            item.Penalty,
			ReplacementPenalty: item.ReplacementPenalty,
			Replacements:       reps,
		}
	}

	return s.repo.UpdateItemSettings(fittingID, updates)
}
