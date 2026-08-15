package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

type ATSetService struct {
	repo      *repository.ATSetRepository
	score     *ATScoreService
	scoreRepo *repository.ATScoreRepository
	fitting   *FleetConfigService
	sdeRepo   *repository.SdeRepository
}

func NewATSetService() *ATSetService {
	return &ATSetService{repo: repository.NewATSetRepository(), score: NewATScoreService(), scoreRepo: repository.NewATScoreRepository(), fitting: NewFleetConfigService(), sdeRepo: repository.NewSdeRepository()}
}

type ATSetInput struct {
	Name          string                       `json:"name"`
	Format        string                       `json:"format"`
	RuleVersionID uint                         `json:"rule_version_id"`
	Status        string                       `json:"status"`
	Slots         []repository.ATSetSlotBundle `json:"slots"`
}

type ATFitInput struct {
	BaseEFT       string                   `json:"base_eft"`
	BaseEWarTags  []string                 `json:"base_ewar_tags"`
	BaseBonusTags []string                 `json:"base_bonus_tags"`
	Replacements  []model.ATFitReplacement `json:"replacements"`
}

type ATFitDetailDTO struct {
	Fit          model.ATFit              `json:"fit"`
	Replacements []model.ATFitReplacement `json:"replacements"`
	Parsed       ATParsedEFT              `json:"parsed"`
	EWarTags     []string                 `json:"ewar_tags"`
	BonusTags    []string                 `json:"bonus_tags"`
}

func normalizeTagList(values []string) ([]string, error) {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, errors.New("标签不能为空")
		}
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result, nil
}

func fittingDisplayName(shipName string, ewar, bonus []string) string {
	labels := append(append([]string{}, ewar...), bonus...)
	if len(labels) == 0 {
		return shipName
	}
	return fmt.Sprintf("%s · %s", shipName, strings.Join(labels, "+"))
}

func (s *ATSetService) SaveFit(userID uint, id uint, input ATFitInput, admin bool) (*ATFitDetailDTO, error) {
	parsed, err := s.ParseEFT(input.BaseEFT)
	if err != nil {
		return nil, err
	}
	ewar, err := normalizeTagList(input.BaseEWarTags)
	if err != nil {
		return nil, err
	}
	bonus, err := normalizeTagList(input.BaseBonusTags)
	if err != nil {
		return nil, err
	}
	var fit *model.ATFit
	if id > 0 {
		existing, getErr := s.repo.GetFit(id)
		if getErr != nil {
			return nil, getErr
		}
		if !admin && existing.Fit.CreatedBy != userID {
			return nil, errors.New("无权修改该装配")
		}
		fit = &existing.Fit
	} else {
		fit = &model.ATFit{CreatedBy: userID, Status: "active"}
	}
	ewarJSON, _ := json.Marshal(ewar)
	bonusJSON, _ := json.Marshal(bonus)
	fit.Name = fittingDisplayName(parsed.ShipName, ewar, bonus)
	fit.ShipTypeID, fit.ShipName, fit.BaseEFT = parsed.ShipTypeID, parsed.ShipName, strings.TrimSpace(input.BaseEFT)
	fit.BaseEWarTags, fit.BaseBonusTags = ewarJSON, bonusJSON
	for i := range input.Replacements {
		r := &input.Replacements[i]
		if r.TagType != "ewar" && r.TagType != "bonus" {
			return nil, errors.New("标签类型仅支持 ewar 或 bonus")
		}
		if strings.TrimSpace(r.Tag) == "" {
			return nil, errors.New("替换规则标签不能为空")
		}
	}
	items := make([]model.FleetConfigFittingItem, len(parsed.Items))
	for i, item := range parsed.Items {
		items[i] = model.FleetConfigFittingItem{TypeID: item.TypeID, Quantity: item.Quantity, Flag: item.Flag}
	}
	validationSlot := &model.ATSetSlot{EWarTags: ewarJSON, BonusTags: bonusJSON}
	// Validation accepts every replacement tag as an available variant while
	// base tags identify the state already represented by the imported EFT.
	allEwar, allBonus := append([]string{}, ewar...), append([]string{}, bonus...)
	for _, r := range input.Replacements {
		if r.TagType == "ewar" {
			allEwar = append(allEwar, r.Tag)
		} else {
			allBonus = append(allBonus, r.Tag)
		}
	}
	validationSlot.EWarTags, _ = json.Marshal(allEwar)
	validationSlot.BonusTags, _ = json.Marshal(allBonus)
	if err := validateReplacementTags(validationSlot, input.Replacements, items); err != nil {
		return nil, err
	}
	if err := s.repo.SaveFit(fit, input.Replacements); err != nil {
		return nil, err
	}
	return s.GetFit(fit.ID)
}

func (s *ATSetService) GetFit(id uint) (*ATFitDetailDTO, error) {
	detail, err := s.repo.GetFit(id)
	if err != nil {
		return nil, err
	}
	parsed, err := s.ParseEFT(detail.Fit.BaseEFT)
	if err != nil {
		return nil, err
	}
	var ewar, bonus []string
	_ = json.Unmarshal(detail.Fit.BaseEWarTags, &ewar)
	_ = json.Unmarshal(detail.Fit.BaseBonusTags, &bonus)
	for _, r := range detail.Replacements {
		if r.TagType == "ewar" {
			ewar = append(ewar, r.Tag)
		} else {
			bonus = append(bonus, r.Tag)
		}
	}
	ewar, _ = normalizeTagList(ewar)
	bonus, _ = normalizeTagList(bonus)
	return &ATFitDetailDTO{Fit: detail.Fit, Replacements: detail.Replacements, Parsed: *parsed, EWarTags: ewar, BonusTags: bonus}, nil
}

func (s *ATSetService) ListFits() ([]model.ATFit, error) { return s.repo.ListFits() }
func (s *ATSetService) DeleteFit(userID, id uint, admin bool) error {
	detail, err := s.repo.GetFit(id)
	if err != nil {
		return err
	}
	if !admin && detail.Fit.CreatedBy != userID {
		return errors.New("无权删除该装配")
	}
	if err := s.repo.DeleteFit(id); err != nil {
		return errors.New("装配已被阵容 Set 使用，不能删除")
	}
	return nil
}

func (s *ATSetService) Create(userID uint, input ATSetInput) (*repository.ATSetDetail, error) {
	set := &model.ATSet{Name: strings.TrimSpace(input.Name), Format: input.Format, RuleVersionID: input.RuleVersionID, Status: input.Status, CreatedBy: userID}
	if set.Status == "" {
		set.Status = "draft"
	}
	score, err := s.validate(set, input.Slots)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(set, input.Slots); err != nil {
		return nil, err
	}
	return &repository.ATSetDetail{Set: *set, Slots: input.Slots, Score: *score}, nil
}

func (s *ATSetService) Update(userID, id uint, input ATSetInput, admin bool) (*repository.ATSetDetail, error) {
	existing, _, err := s.repo.Get(id)
	if err != nil {
		return nil, err
	}
	if !admin && existing.CreatedBy != userID {
		return nil, errors.New("无权修改该 Set")
	}
	existing.Name, existing.Format, existing.RuleVersionID, existing.Status = strings.TrimSpace(input.Name), input.Format, input.RuleVersionID, input.Status
	score, err := s.validate(existing, input.Slots)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(existing, input.Slots); err != nil {
		return nil, err
	}
	return &repository.ATSetDetail{Set: *existing, Slots: input.Slots, Score: *score}, nil
}

func (s *ATSetService) validate(set *model.ATSet, bundles []repository.ATSetSlotBundle) (*model.ATScoreRosterResult, error) {
	if set.Name == "" {
		return nil, errors.New("Set 名称不能为空")
	}
	rule, err := s.scoreRepo.GetRule(set.RuleVersionID)
	if err != nil {
		return nil, errors.New("分数规则不存在")
	}
	if rule.Status != model.ATScoreRulePublished {
		return nil, errors.New("Set 只能引用已发布的分数规则")
	}
	typeIDs := make([]int64, len(bundles))
	orders := map[int]bool{}
	for i := range bundles {
		slot := &bundles[i].Slot
		if slot.SlotOrder <= 0 || orders[slot.SlotOrder] {
			return nil, errors.New("槽位顺序必须为不重复的正整数")
		}
		orders[slot.SlotOrder] = true
		if slot.FitID != nil {
			fit, fitErr := s.repo.GetFit(*slot.FitID)
			if fitErr != nil {
				return nil, fmt.Errorf("槽位 %d 的装配不存在", slot.SlotOrder)
			}
			slot.ShipTypeID, slot.BaseEFT = fit.Fit.ShipTypeID, fit.Fit.BaseEFT
			bundles[i].Fit = &fit.Fit
			bundles[i].Replacements = fit.Replacements
		}
		shipTypeID, items, parseErr := s.fitting.parseEFTToFitting(slot.BaseEFT)
		if parseErr != nil {
			return nil, fmt.Errorf("槽位 %d: %w", slot.SlotOrder, parseErr)
		}
		if shipTypeID != slot.ShipTypeID {
			return nil, fmt.Errorf("槽位 %d 的 EFT 舰船与 ship_type_id 不一致", slot.SlotOrder)
		}
		selectedEwar, _ := decodeTags(slot.EWarTags)
		selectedBonus, _ := decodeTags(slot.BonusTags)
		activeReplacements := make([]model.ATFitReplacement, 0, len(bundles[i].Replacements))
		for _, replacement := range bundles[i].Replacements {
			if (replacement.TagType == "ewar" && selectedEwar[replacement.Tag]) || (replacement.TagType == "bonus" && selectedBonus[replacement.Tag]) {
				activeReplacements = append(activeReplacements, replacement)
			}
		}
		if err := validateReplacementTags(slot, activeReplacements, items); err != nil {
			return nil, fmt.Errorf("槽位 %d: %w", slot.SlotOrder, err)
		}
		typeIDs[i] = slot.ShipTypeID
	}
	result, err := s.score.ValidateRoster(set.RuleVersionID, set.Format, typeIDs)
	if err != nil {
		return nil, err
	}
	if !result.Valid {
		return nil, fmt.Errorf("Set 校验失败: %s", strings.Join(result.Errors, "; "))
	}
	return result, nil
}

func decodeTags(raw json.RawMessage) (map[string]bool, error) {
	if len(raw) == 0 {
		return map[string]bool{}, nil
	}
	var tags []string
	if err := json.Unmarshal(raw, &tags); err != nil {
		return nil, errors.New("标签必须为字符串数组")
	}
	result := map[string]bool{}
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			return nil, errors.New("标签不能为空")
		}
		result[tag] = true
	}
	return result, nil
}

func validateReplacementTags(slot *model.ATSetSlot, replacements []model.ATFitReplacement, items []model.FleetConfigFittingItem) error {
	ewar, err := decodeTags(slot.EWarTags)
	if err != nil {
		return err
	}
	bonus, err := decodeTags(slot.BonusTags)
	if err != nil {
		return err
	}
	itemCounts := map[int64]int{}
	for _, item := range items {
		itemCounts[item.TypeID] += item.Quantity
	}
	targetResults := map[int64]string{}
	for _, replacement := range replacements {
		if replacement.TagType != "ewar" && replacement.TagType != "bonus" {
			return errors.New("标签类型仅支持 ewar 或 bonus")
		}
		if (replacement.TagType == "ewar" && !ewar[replacement.Tag]) || (replacement.TagType == "bonus" && !bonus[replacement.Tag]) {
			return fmt.Errorf("替换规则引用了未选择的标签 %s", replacement.Tag)
		}
		if replacement.Operation != "replace" && replacement.Operation != "add" && replacement.Operation != "remove" {
			return errors.New("替换操作仅支持 replace/add/remove")
		}
		if replacement.Quantity <= 0 {
			return errors.New("替换数量必须为正整数")
		}
		if replacement.Operation != "add" && itemCounts[replacement.TargetTypeID] < replacement.Quantity {
			return fmt.Errorf("目标 TypeID %d 数量不足", replacement.TargetTypeID)
		}
		if replacement.Operation != "remove" && replacement.ReplacementTypeID <= 0 {
			return errors.New("替换或新增物品必须指定 replacement_type_id")
		}
		if replacement.TargetTypeID > 0 {
			result := fmt.Sprintf("%s:%d:%d", replacement.Operation, replacement.ReplacementTypeID, replacement.Quantity)
			if previous, exists := targetResults[replacement.TargetTypeID]; exists && previous != result {
				return fmt.Errorf("多个标签对 TypeID %d 产生冲突结果", replacement.TargetTypeID)
			}
			targetResults[replacement.TargetTypeID] = result
		}
	}
	return nil
}

func (s *ATSetService) Get(id uint) (*repository.ATSetDetail, error) {
	set, slots, err := s.repo.Get(id)
	if err != nil {
		return nil, err
	}
	typeIDs := make([]int64, len(slots))
	for i := range slots {
		typeIDs[i] = slots[i].Slot.ShipTypeID
		if slots[i].Fit == nil && strings.TrimSpace(slots[i].Slot.BaseEFT) != "" {
			if parsed, parseErr := s.ParseEFT(slots[i].Slot.BaseEFT); parseErr == nil {
				slots[i].Fit = &model.ATFit{Name: fittingDisplayName(parsed.ShipName, nil, nil), ShipTypeID: parsed.ShipTypeID, ShipName: parsed.ShipName, BaseEFT: slots[i].Slot.BaseEFT, BaseEWarTags: slots[i].Slot.EWarTags, BaseBonusTags: slots[i].Slot.BonusTags, Status: "legacy"}
			}
		}
	}
	score, err := s.score.ValidateRoster(set.RuleVersionID, set.Format, typeIDs)
	if err != nil {
		return nil, err
	}
	return &repository.ATSetDetail{Set: *set, Slots: slots, Score: *score}, nil
}
func (s *ATSetService) List() ([]model.ATSet, error) { return s.repo.List() }
func (s *ATSetService) Delete(userID, id uint, admin bool) error {
	set, _, err := s.repo.Get(id)
	if err != nil {
		return err
	}
	if !admin && set.CreatedBy != userID {
		return errors.New("无权删除该 Set")
	}
	return s.repo.Delete(id)
}

type ATSetEFTOutput struct {
	SetID    uint           `json:"set_id"`
	Slots    []ATSetSlotEFT `json:"slots"`
	Combined string         `json:"combined"`
}

type ATParsedEFT struct {
	ShipTypeID  int64             `json:"ship_type_id"`
	ShipName    string            `json:"ship_name"`
	FittingName string            `json:"fitting_name"`
	Items       []ATParsedEFTItem `json:"items"`
}

type ATParsedEFTItem struct {
	TypeID   int64  `json:"type_id"`
	TypeName string `json:"type_name"`
	Quantity int    `json:"quantity"`
	Flag     string `json:"flag"`
}

func (s *ATSetService) ParseEFT(eft string) (*ATParsedEFT, error) {
	header := parseEFTHeader(eft)
	if header == nil {
		return nil, errors.New("EFT 格式错误：缺少 [舰船, 装配名] 头部")
	}
	shipTypeID, parsed, err := s.fitting.parseEFTToFitting(eft)
	if err != nil {
		return nil, err
	}
	typeIDs := make([]int, 0, len(parsed)+1)
	typeIDs = append(typeIDs, int(shipTypeID))
	for _, item := range parsed {
		typeIDs = append(typeIDs, int(item.TypeID))
	}
	types, err := s.sdeRepo.GetTypes(typeIDs, nil, "en")
	if err != nil {
		return nil, err
	}
	names := make(map[int64]string, len(types))
	for _, item := range types {
		names[int64(item.TypeID)] = item.TypeName
	}
	result := &ATParsedEFT{ShipTypeID: shipTypeID, ShipName: names[shipTypeID], FittingName: header.FittingName, Items: make([]ATParsedEFTItem, 0, len(parsed))}
	for _, item := range parsed {
		result.Items = append(result.Items, ATParsedEFTItem{TypeID: item.TypeID, TypeName: names[item.TypeID], Quantity: item.Quantity, Flag: item.Flag})
	}
	return result, nil
}

type ATSetSlotEFT struct {
	SlotID    uint   `json:"slot_id"`
	SlotOrder int    `json:"slot_order"`
	EFT       string `json:"eft"`
}

func (s *ATSetService) GenerateEFT(id uint) (*ATSetEFTOutput, error) {
	detail, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	if !detail.Score.Valid {
		return nil, errors.New("Set 当前校验未通过")
	}
	output := &ATSetEFTOutput{SetID: id}
	for _, bundle := range detail.Slots {
		shipID, parsed, err := s.fitting.parseEFTToFitting(bundle.Slot.BaseEFT)
		if err != nil {
			return nil, err
		}
		items := make([]model.EveCharacterFittingItem, len(parsed))
		for i, item := range parsed {
			items[i] = model.EveCharacterFittingItem{TypeID: item.TypeID, Quantity: item.Quantity, Flag: item.Flag}
		}
		selectedEwar, _ := decodeTags(bundle.Slot.EWarTags)
		selectedBonus, _ := decodeTags(bundle.Slot.BonusTags)
		baseEwar, baseBonus := map[string]bool{}, map[string]bool{}
		if bundle.Fit != nil {
			baseEwar, _ = decodeTags(bundle.Fit.BaseEWarTags)
			baseBonus, _ = decodeTags(bundle.Fit.BaseBonusTags)
		}
		for _, replacement := range bundle.Replacements {
			selected := selectedEwar[replacement.Tag]
			base := baseEwar[replacement.Tag]
			if replacement.TagType == "bonus" {
				selected, base = selectedBonus[replacement.Tag], baseBonus[replacement.Tag]
			}
			if !selected || base {
				continue
			}
			items, err = applyATReplacement(items, replacement)
			if err != nil {
				return nil, err
			}
		}
		ids := []int{int(shipID)}
		for _, item := range items {
			ids = append(ids, int(item.TypeID))
		}
		types, err := s.sdeRepo.GetTypes(ids, nil, "en")
		if err != nil {
			return nil, err
		}
		names := map[int]string{}
		for _, item := range types {
			names[item.TypeID] = item.TypeName
		}
		if len(names) != len(uniqueInts(ids)) {
			return nil, errors.New("EFT 包含未解析 TypeID")
		}
		var selectedTags []string
		for tag := range selectedEwar {
			selectedTags = append(selectedTags, tag)
		}
		for tag := range selectedBonus {
			selectedTags = append(selectedTags, tag)
		}
		sort.Strings(selectedTags)
		fitName := fittingDisplayName(names[int(shipID)], selectedTags, nil)
		eft := buildEFT(names[int(shipID)], fitName, items, names)
		output.Slots = append(output.Slots, ATSetSlotEFT{SlotID: bundle.Slot.ID, SlotOrder: bundle.Slot.SlotOrder, EFT: eft})
	}
	sort.Slice(output.Slots, func(i, j int) bool { return output.Slots[i].SlotOrder < output.Slots[j].SlotOrder })
	parts := make([]string, len(output.Slots))
	for i := range output.Slots {
		parts[i] = output.Slots[i].EFT
	}
	output.Combined = strings.Join(parts, "\n\n---\n\n")
	return output, nil
}

func (s *ATSetService) GeneratePositionEFT(setID uint, slotOrder int, ewarTags, bonusTags []string) (*ATSetSlotEFT, error) {
	detail, err := s.Get(setID)
	if err != nil {
		return nil, err
	}
	var selected *repository.ATSetSlotBundle
	for i := range detail.Slots {
		if detail.Slots[i].Slot.SlotOrder == slotOrder {
			selected = &detail.Slots[i]
			break
		}
	}
	if selected == nil {
		return nil, errors.New("阵容位置不存在")
	}
	if len(ewarTags) > 0 {
		selected.Slot.EWarTags, _ = json.Marshal(ewarTags)
	}
	if len(bonusTags) > 0 {
		selected.Slot.BonusTags, _ = json.Marshal(bonusTags)
	}
	shipID, parsed, err := s.fitting.parseEFTToFitting(selected.Slot.BaseEFT)
	if err != nil {
		return nil, err
	}
	items := make([]model.EveCharacterFittingItem, len(parsed))
	for i, item := range parsed {
		items[i] = model.EveCharacterFittingItem{TypeID: item.TypeID, Quantity: item.Quantity, Flag: item.Flag}
	}
	selectedEwar, _ := decodeTags(selected.Slot.EWarTags)
	selectedBonus, _ := decodeTags(selected.Slot.BonusTags)
	baseEwar, baseBonus := map[string]bool{}, map[string]bool{}
	if selected.Fit != nil {
		baseEwar, _ = decodeTags(selected.Fit.BaseEWarTags)
		baseBonus, _ = decodeTags(selected.Fit.BaseBonusTags)
	}
	for _, replacement := range selected.Replacements {
		active, base := selectedEwar[replacement.Tag], baseEwar[replacement.Tag]
		if replacement.TagType == "bonus" {
			active, base = selectedBonus[replacement.Tag], baseBonus[replacement.Tag]
		}
		if !active || base {
			continue
		}
		items, err = applyATReplacement(items, replacement)
		if err != nil {
			return nil, err
		}
	}
	ids := []int{int(shipID)}
	for _, item := range items {
		ids = append(ids, int(item.TypeID))
	}
	types, err := s.sdeRepo.GetTypes(ids, nil, "en")
	if err != nil {
		return nil, err
	}
	names := map[int]string{}
	for _, item := range types {
		names[item.TypeID] = item.TypeName
	}
	if len(names) != len(uniqueInts(ids)) {
		return nil, errors.New("EFT 包含未解析 TypeID")
	}
	var tags []string
	for tag := range selectedEwar {
		tags = append(tags, tag)
	}
	for tag := range selectedBonus {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	result := &ATSetSlotEFT{SlotID: selected.Slot.ID, SlotOrder: slotOrder, EFT: buildEFT(names[int(shipID)], fittingDisplayName(names[int(shipID)], tags, nil), items, names)}
	return result, nil
}

func applyATReplacement(items []model.EveCharacterFittingItem, replacement model.ATFitReplacement) ([]model.EveCharacterFittingItem, error) {
	remaining := replacement.Quantity
	flag := "Cargo"
	if replacement.Operation != "add" {
		for i := range items {
			if items[i].TypeID == replacement.TargetTypeID && remaining > 0 {
				flag = items[i].Flag
				take := items[i].Quantity
				if take > remaining {
					take = remaining
				}
				items[i].Quantity -= take
				remaining -= take
			}
		}
		if remaining > 0 {
			return nil, fmt.Errorf("TypeID %d 数量不足", replacement.TargetTypeID)
		}
	}
	filtered := items[:0]
	for _, item := range items {
		if item.Quantity > 0 {
			filtered = append(filtered, item)
		}
	}
	items = filtered
	if replacement.Operation == "remove" {
		return items, nil
	}
	for i := range items {
		if items[i].TypeID == replacement.ReplacementTypeID && items[i].Flag == flag {
			items[i].Quantity += replacement.Quantity
			return items, nil
		}
	}
	return append(items, model.EveCharacterFittingItem{TypeID: replacement.ReplacementTypeID, Quantity: replacement.Quantity, Flag: flag}), nil
}

func uniqueInts(values []int) map[int]bool {
	result := map[int]bool{}
	for _, value := range values {
		result[value] = true
	}
	return result
}
