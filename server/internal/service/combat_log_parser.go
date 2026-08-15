package service

import (
	"errors"
	"html"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"amiya-eden/internal/model"
)

const (
	combatCategoryDamageOut      = "damage_out"
	combatCategoryDamageIn       = "damage_in"
	combatCategoryLogisticsOut   = "logistics_out"
	combatCategoryLogisticsIn    = "logistics_in"
	combatCategoryCapTransferOut = "cap_transfer_out"
	combatCategoryCapTransferIn  = "cap_transfer_in"
	combatCategoryCapDamageOut   = "cap_damage_out"
	combatCategoryCapDamageIn    = "cap_damage_in"
	combatCategoryMining         = "mining"
	combatHitBarelyScratches     = "barely_scratches"
	combatHitHits                = "hits"
	combatHitSmashes             = "smashes"
	combatHitPenetrates          = "penetrates"
	combatHitWrecks              = "wrecks"
	combatHitGlancesOff          = "glances_off"
)

var (
	combatTimestampPattern = regexp.MustCompile(`^\[\s*(\d{4}\.\d{2}\.\d{2} \d{2}:\d{2}:\d{2})\s*\]`)
	combatTagPattern       = regexp.MustCompile(`<[^>]*>`)
	combatBoldPattern      = regexp.MustCompile(`(?i)<b[^>]*>\s*([^<]+)`)
	combatPlainAmount      = regexp.MustCompile(`(?i)\(combat\)\s*[+-]?([0-9][0-9 ,.]*[0-9]|[0-9])`)
	miningPlainValues      = regexp.MustCompile(`(?i)\(mining\).*?([0-9][0-9 ,.]*[0-9]|[0-9])\s+units of\s+(.+?)(?:\s+were mined|$)`)
	combatActorPattern     = regexp.MustCompile(`^\s*([^\[\]\(\)]+?)\s*(?:\[[^\]]*\]\s*)?\(([^\)]*)\)\s*-\s*([^-]+)`)
	combatHitQualityTerms  = []struct {
		quality string
		terms   []string
	}{
		{combatHitBarelyScratches, []string{"轻轻擦过", "barely scratches"}},
		{combatHitHits, []string{"命中", "hits"}},
		{combatHitSmashes, []string{"强力一击", "smashes"}},
		{combatHitPenetrates, []string{"穿透", "penetrates"}},
		{combatHitWrecks, []string{"致命一击", "wrecks"}},
		{combatHitGlancesOff, []string{"擦过", "glances off"}},
	}
)

type combatLanguageRules struct {
	name          string
	listenerLabel string
	damageOut     string
	damageIn      string
	logisticsOut  []string
	logisticsIn   []string
	capOut        string
	capIn         string
	neut          string
	nosOut        string
	nosIn         string
}

// These phrases mirror the language rules used by PyEveLiveDPS. The parser
// deliberately classifies the specific logistics/capacitor messages before
// the generic outgoing/incoming damage markers.
var combatLanguages = []combatLanguageRules{
	{name: "english", listenerLabel: "Listener:", damageOut: "to", damageIn: "from", logisticsOut: []string{"remote armor repaired to", "remote hull repaired to", "remote shield boosted to"}, logisticsIn: []string{"remote armor repaired by", "remote hull repaired by", "remote shield boosted by"}, capOut: "remote capacitor transmitted to", capIn: "remote capacitor transmitted by", neut: "energy neutralized", nosOut: "energy drained from", nosIn: "energy drained to"},
	{name: "chinese", listenerLabel: "收听者:", damageOut: "对", damageIn: "来自", logisticsOut: []string{"远程装甲维修量至", "远程结构维修量至", "远程护盾回充增量至"}, logisticsIn: []string{"远程装甲维修量由", "远程结构维修量由", "远程护盾回充增量由"}, capOut: "远程电容传输至", capIn: "远程电容传输量由", neut: "能量中和", nosOut: "被从", nosIn: "被吸取到"},
	{name: "russian", listenerLabel: "Слушатель:", damageOut: "на", damageIn: "из", logisticsOut: []string{"единиц запаса прочности брони отремонтировано", "единиц запаса прочности корпуса отремонтировано", "единиц запаса прочности щитов накачано"}, logisticsIn: []string{"единиц запаса прочности брони получено дистанционным ремонтом", "единиц запаса прочности корпуса получено дистанционным ремонтом", "единиц запаса прочности щитов получено накачкой"}, capOut: "единиц запаса энергии накопителя отправлено", capIn: "единиц запаса энергии накопителя получено", neut: "энергии нейтрализовано", nosOut: "энергии извлечено из", nosIn: "энергии извлечено и передано"},
	{name: "french", listenerLabel: "Auditeur:", damageOut: "à", damageIn: "de", logisticsOut: []string{"points de blindage transférés à distance à", "points de structure transférés à distance à", "points de boucliers transférés à distance à"}, logisticsIn: []string{"points de blindage réparés à distance par", "points de structure réparés à distance par", "points de boucliers transférés à distance par"}, capOut: "points de capaciteur transférés à distance à", capIn: "points de capaciteur transférés à distance par", neut: "d'énergie neutralisée", nosOut: "d'énergie siphonnée aux dépens de", nosIn: "d'énergie siphonnée en faveur de"},
	{name: "german", listenerLabel: "Empfänger:", damageOut: "nach", damageIn: "von", logisticsOut: []string{"panzerungs-fernreparatur zu", "rumpf-fernreparatur zu", "schildfernbooster aktiviert zu"}, logisticsIn: []string{"panzerungs-fernreparatur von", "rumpf-fernreparatur von", "schildfernbooster aktiviert von"}, capOut: "fernenergiespeicher übertragen zu", capIn: "fernenergiespeicher übertragen von", neut: "energie neutralisiert", nosOut: "energie transferiert von", nosIn: "energie transferiert zu"},
	{name: "japanese", listenerLabel: "傍聴者:", damageOut: "対象:", damageIn: "攻撃者:", logisticsOut: []string{"remote armor repaired to", "remote hull repaired to", "remote shield boosted to"}, logisticsIn: []string{"remote armor repaired by", "remote hull repaired by", "remote shield boosted by"}, capOut: "remote capacitor transmitted to", capIn: "remote capacitor transmitted by", neut: "エネルギーニュートラライズ", nosOut: "エネルギードレイン 対象:", nosIn: "エネルギードレイン 攻撃者:"},
}

type combatParsedEvent struct {
	OccurredAt time.Time
	Category   string
	Amount     int64
	PilotName  string
	ShipType   string
	WeaponType string
	HitQuality string
}

type combatParsedLog struct {
	CharacterName string
	Language      string
	StartedAt     time.Time
	EndedAt       time.Time
	Events        []combatParsedEvent
}

func parseCombatLog(data []byte) (*combatParsedLog, error) {
	data = []byte(strings.TrimPrefix(string(data), "\ufeff"))
	if !utf8.Valid(data) {
		return nil, errors.New("日志必须是 UTF-8 编码")
	}
	content := string(data)
	rules, character := detectCombatLanguage(content)
	if rules == nil {
		return nil, errors.New("未识别到 EVE 游戏日志的角色信息")
	}

	parsed := &combatParsedLog{CharacterName: character, Language: rules.name}
	for _, line := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		timeMatch := combatTimestampPattern.FindStringSubmatch(strings.TrimSpace(line))
		if len(timeMatch) != 2 {
			continue
		}
		occurredAt, err := time.Parse("2006.01.02 15:04:05", timeMatch[1])
		if err != nil {
			continue
		}
		events := parseCombatLine(line, occurredAt, *rules)
		parsed.Events = append(parsed.Events, events...)
	}
	if len(parsed.Events) == 0 {
		return nil, errors.New("日志中没有可识别的战斗、后勤、电容或采矿记录")
	}
	sort.Slice(parsed.Events, func(i, j int) bool { return parsed.Events[i].OccurredAt.Before(parsed.Events[j].OccurredAt) })
	parsed.StartedAt = parsed.Events[0].OccurredAt
	parsed.EndedAt = parsed.Events[len(parsed.Events)-1].OccurredAt
	return parsed, nil
}

func detectCombatLanguage(content string) (*combatLanguageRules, string) {
	for i := range combatLanguages {
		label := combatLanguages[i].listenerLabel
		if pos := strings.Index(content, label); pos >= 0 {
			name := content[pos+len(label):]
			if end := strings.IndexAny(name, "\r\n"); end >= 0 {
				name = name[:end]
			}
			name = strings.TrimSpace(combatTagPattern.ReplaceAllString(name, ""))
			if name != "" {
				return &combatLanguages[i], name
			}
		}
	}
	return nil, ""
}

func parseCombatLine(raw string, occurredAt time.Time, rules combatLanguageRules) []combatParsedEvent {
	lowerRaw := strings.ToLower(raw)
	originalPlain := normalizeCombatText(raw)
	plain := strings.ToLower(originalPlain)
	if strings.Contains(plain, "(mining)") {
		amount, ore := parseMiningValues(raw)
		if amount > 0 {
			return []combatParsedEvent{{OccurredAt: occurredAt, Category: combatCategoryMining, Amount: amount, WeaponType: ore}}
		}
		return nil
	}
	if !strings.Contains(plain, "(combat)") {
		return nil
	}
	amount := parseCombatAmount(raw)
	if amount <= 0 {
		return nil
	}

	category, actorMarker := "", ""
	if marker := firstContained(plain, rules.logisticsOut); marker != "" {
		category, actorMarker = combatCategoryLogisticsOut, marker
	} else if marker := firstContained(plain, rules.logisticsIn); marker != "" {
		category, actorMarker = combatCategoryLogisticsIn, marker
	} else if strings.Contains(plain, strings.ToLower(rules.capOut)) {
		category, actorMarker = combatCategoryCapTransferOut, strings.ToLower(rules.capOut)
	} else if strings.Contains(plain, strings.ToLower(rules.capIn)) {
		category, actorMarker = combatCategoryCapTransferIn, strings.ToLower(rules.capIn)
	} else if strings.Contains(plain, strings.ToLower(rules.nosOut)) {
		pilot, ship, weapon := parseCombatActor(originalPlain, strings.ToLower(rules.nosOut))
		return []combatParsedEvent{
			{OccurredAt: occurredAt, Category: combatCategoryCapTransferIn, Amount: amount, PilotName: pilot, ShipType: ship, WeaponType: weapon},
			{OccurredAt: occurredAt, Category: combatCategoryCapDamageOut, Amount: amount, PilotName: pilot, ShipType: ship, WeaponType: weapon},
		}
	} else if strings.Contains(plain, strings.ToLower(rules.nosIn)) {
		category, actorMarker = combatCategoryCapDamageIn, strings.ToLower(rules.nosIn)
	} else if strings.Contains(plain, strings.ToLower(rules.neut)) {
		actorMarker = strings.ToLower(rules.neut)
		if strings.Contains(lowerRaw, "ff7fffff") {
			category = combatCategoryCapDamageOut
		} else if strings.Contains(lowerRaw, "ffe57f7f") {
			category = combatCategoryCapDamageIn
		}
	} else if containsCombatMarker(plain, lowerRaw, rules.damageOut) {
		category, actorMarker = combatCategoryDamageOut, strings.ToLower(rules.damageOut)
	} else if containsCombatMarker(plain, lowerRaw, rules.damageIn) {
		category, actorMarker = combatCategoryDamageIn, strings.ToLower(rules.damageIn)
	}
	if category == "" {
		return nil
	}
	pilot, ship, weapon := parseCombatActor(originalPlain, actorMarker)
	hitQuality := ""
	if category == combatCategoryDamageOut || category == combatCategoryDamageIn {
		hitQuality = parseCombatHitQuality(originalPlain)
	}
	return []combatParsedEvent{{OccurredAt: occurredAt, Category: category, Amount: amount, PilotName: pilot, ShipType: ship, WeaponType: weapon, HitQuality: hitQuality}}
}

func parseCombatHitQuality(plain string) string {
	plain = strings.ToLower(strings.TrimSpace(plain))
	for _, definition := range combatHitQualityTerms {
		for _, term := range definition.terms {
			term = strings.ToLower(term)
			if strings.HasSuffix(plain, term) {
				return definition.quality
			}
		}
	}
	return ""
}

func normalizeCombatText(value string) string {
	value = combatTagPattern.ReplaceAllString(value, "")
	value = html.UnescapeString(value)
	return strings.Join(strings.Fields(value), " ")
}

func parseCombatAmount(raw string) int64 {
	matches := combatBoldPattern.FindAllStringSubmatch(raw, -1)
	for _, match := range matches {
		if len(match) == 2 {
			if amount := parseDigits(match[1]); amount > 0 {
				return amount
			}
		}
	}
	if match := combatPlainAmount.FindStringSubmatch(normalizeCombatText(raw)); len(match) == 2 {
		return parseDigits(match[1])
	}
	return 0
}

func parseMiningValues(raw string) (int64, string) {
	matches := combatBoldPattern.FindAllStringSubmatch(raw, -1)
	var amount int64
	for _, match := range matches {
		if len(match) != 2 {
			continue
		}
		value := strings.TrimSpace(combatTagPattern.ReplaceAllString(match[1], ""))
		if amount == 0 {
			amount = parseDigits(value)
			continue
		}
		if parseDigits(value) == 0 {
			return amount, value
		}
	}
	if match := miningPlainValues.FindStringSubmatch(normalizeCombatText(raw)); len(match) == 3 {
		return parseDigits(match[1]), strings.TrimSpace(match[2])
	}
	return amount, ""
}

func parseDigits(value string) int64 {
	digits := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, value)
	if digits == "" {
		return 0
	}
	amount, _ := strconv.ParseInt(digits, 10, 64)
	return amount
}

func firstContained(plain string, markers []string) string {
	for _, marker := range markers {
		marker = strings.ToLower(marker)
		if strings.Contains(plain, marker) {
			return marker
		}
	}
	return ""
}

func containsCombatMarker(plain, raw, marker string) bool {
	marker = strings.ToLower(marker)
	if len([]rune(marker)) > 2 || strings.IndexFunc(marker, func(r rune) bool { return r > unicode.MaxASCII }) >= 0 {
		return strings.Contains(plain, marker)
	}
	return strings.Contains(" "+plain+" ", " "+marker+" ") || strings.Contains(raw, ">"+marker+"<")
}

func parseCombatActor(plain, marker string) (string, string, string) {
	if marker == "" {
		return "", "", ""
	}
	lowerPlain := strings.ToLower(plain)
	pos := strings.Index(lowerPlain, marker)
	if len([]rune(marker)) <= 2 && strings.IndexFunc(marker, func(r rune) bool { return r > unicode.MaxASCII }) < 0 {
		if spaced := strings.Index(lowerPlain, " "+marker+" "); spaced >= 0 {
			pos = spaced + 1
		}
	}
	if pos < 0 {
		return "", "", ""
	}
	candidate := strings.TrimSpace(plain[pos+len(marker):])
	match := combatActorPattern.FindStringSubmatch(candidate)
	if len(match) != 4 {
		return "", "", ""
	}
	return strings.TrimSpace(match[1]), strings.TrimSpace(match[2]), strings.TrimSpace(match[3])
}

func buildCombatAnalysis(parsed *combatParsedLog, report *model.CombatLogReport) ([]model.CombatLogBreakdown, []model.CombatLogTimeline, []model.CombatLogHitQuality) {
	type breakdownKey struct{ category, pilot, ship, weapon string }
	type timelineKey struct {
		occurredAt time.Time
		category   string
	}
	breakdowns := make(map[breakdownKey]*model.CombatLogBreakdown)
	timeline := make(map[timelineKey]*model.CombatLogTimeline)
	hitQualities := make(map[string]*model.CombatLogHitQuality)
	for _, event := range parsed.Events {
		report.EventCount++
		switch event.Category {
		case combatCategoryDamageOut:
			report.DamageOut += event.Amount
		case combatCategoryDamageIn:
			report.DamageIn += event.Amount
		case combatCategoryLogisticsOut:
			report.LogisticsOut += event.Amount
		case combatCategoryLogisticsIn:
			report.LogisticsIn += event.Amount
		case combatCategoryCapTransferOut:
			report.CapTransferOut += event.Amount
		case combatCategoryCapTransferIn:
			report.CapTransferIn += event.Amount
		case combatCategoryCapDamageOut:
			report.CapDamageOut += event.Amount
		case combatCategoryCapDamageIn:
			report.CapDamageIn += event.Amount
		case combatCategoryMining:
			report.MiningUnits += event.Amount
		}
		if event.HitQuality != "" && (event.Category == combatCategoryDamageOut || event.Category == combatCategoryDamageIn) {
			key := event.Category + "\x00" + event.HitQuality
			if hitQualities[key] == nil {
				hitQualities[key] = &model.CombatLogHitQuality{Direction: event.Category, Quality: event.HitQuality}
			}
			hitQualities[key].HitCount++
			hitQualities[key].DamageAmount += event.Amount
		}

		bk := breakdownKey{event.Category, event.PilotName, event.ShipType, event.WeaponType}
		if breakdowns[bk] == nil {
			breakdowns[bk] = &model.CombatLogBreakdown{Category: event.Category, PilotName: event.PilotName, ShipType: event.ShipType, WeaponType: event.WeaponType}
		}
		breakdowns[bk].Amount += event.Amount
		breakdowns[bk].EventCount++

		tk := timelineKey{event.OccurredAt, event.Category}
		if timeline[tk] == nil {
			timeline[tk] = &model.CombatLogTimeline{OccurredAt: event.OccurredAt, Category: event.Category}
		}
		timeline[tk].Amount += event.Amount
	}

	breakdownRows := make([]model.CombatLogBreakdown, 0, len(breakdowns))
	for _, row := range breakdowns {
		breakdownRows = append(breakdownRows, *row)
	}
	sort.Slice(breakdownRows, func(i, j int) bool {
		if breakdownRows[i].Category == breakdownRows[j].Category {
			return breakdownRows[i].Amount > breakdownRows[j].Amount
		}
		return breakdownRows[i].Category < breakdownRows[j].Category
	})

	timelineRows := make([]model.CombatLogTimeline, 0, len(timeline))
	for _, row := range timeline {
		timelineRows = append(timelineRows, *row)
	}
	sort.Slice(timelineRows, func(i, j int) bool {
		if timelineRows[i].OccurredAt.Equal(timelineRows[j].OccurredAt) {
			return timelineRows[i].Category < timelineRows[j].Category
		}
		return timelineRows[i].OccurredAt.Before(timelineRows[j].OccurredAt)
	})
	hitQualityRows := make([]model.CombatLogHitQuality, 0, len(hitQualities))
	for _, row := range hitQualities {
		hitQualityRows = append(hitQualityRows, *row)
	}
	qualityRank := make(map[string]int, len(combatHitQualityTerms))
	for index, definition := range combatHitQualityTerms {
		qualityRank[definition.quality] = index
	}
	sort.Slice(hitQualityRows, func(i, j int) bool {
		if hitQualityRows[i].Direction == hitQualityRows[j].Direction {
			return qualityRank[hitQualityRows[i].Quality] < qualityRank[hitQualityRows[j].Quality]
		}
		return hitQualityRows[i].Direction < hitQualityRows[j].Direction
	})
	return breakdownRows, timelineRows, hitQualityRows
}
