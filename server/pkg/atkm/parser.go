package atkm

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Victim struct {
	CharacterName   string
	CorporationName string
	AllianceName    string
	FactionName     string
	ShipName        string
	DamageTaken     int
}
type Attacker struct {
	CharacterName   string
	CorporationName string
	AllianceName    string
	FactionName     string
	ShipName        string
	WeaponName      string
	DamageDone      int
	FinalBlow       bool
}
type Killmail struct {
	Time           time.Time
	Victim         Victim
	Attackers      []Attacker
	SolarSystem    string
	SecurityStatus string
	RawText        string
	NormalizedText string
}

var timestampPattern = regexp.MustCompile(`(?m)^\s*\d{4}\.\d{2}\.\d{2}\s+\d{2}:\d{2}:\d{2}\s*$`)
var tournamentTimezone = time.FixedZone("Asia/Shanghai", 8*60*60)

func Normalize(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.ReplaceAll(text, "：", ":")
	lines := strings.Split(text, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	for len(lines) > 0 && lines[0] == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

func Split(text string) []string {
	locations := timestampPattern.FindAllStringIndex(text, -1)
	if len(locations) == 0 {
		if strings.TrimSpace(text) == "" {
			return nil
		}
		return []string{strings.TrimSpace(text)}
	}
	result := make([]string, 0, len(locations))
	for i, location := range locations {
		end := len(text)
		if i+1 < len(locations) {
			end = locations[i+1][0]
		}
		block := strings.TrimSpace(text[location[0]:end])
		if block != "" {
			result = append(result, block)
		}
	}
	return result
}

func ParseMany(text string) ([]Killmail, error) {
	blocks := Split(text)
	if len(blocks) == 0 {
		return nil, errors.New("未检测到 KM 文本")
	}
	result := make([]Killmail, 0, len(blocks))
	for _, block := range blocks {
		km, err := Parse(block)
		if err != nil {
			return nil, err
		}
		result = append(result, *km)
	}
	return result, nil
}

func Parse(text string) (*Killmail, error) {
	normalized := Normalize(text)
	km := &Killmail{RawText: text, NormalizedText: normalized}
	lines := strings.Split(normalized, "\n")
	section := ""
	var current *Attacker
	for _, line := range lines {
		if line == "" {
			continue
		}
		if timestampPattern.MatchString(line) {
			parsed, err := time.ParseInLocation("2006.01.02 15:04:05", strings.TrimSpace(line), tournamentTimezone)
			if err != nil {
				return nil, errors.New("KM 时间格式无效")
			}
			km.Time = parsed
			continue
		}
		if hasField(line, "受害者") {
			section = "victim"
			km.Victim.CharacterName = field(line, "受害者")
			continue
		}
		if hasField(line, "参与者") {
			if current != nil {
				km.Attackers = append(km.Attackers, *current)
				current = nil
			}
			section = "attackers"
			continue
		}
		if hasField(line, "被毁物品") || hasField(line, "掉落物品") {
			if current != nil {
				km.Attackers = append(km.Attackers, *current)
				current = nil
			}
			section = "items"
			continue
		}
		if section == "items" {
			continue
		}
		if section == "victim" {
			switch {
			case hasField(line, "军团"):
				km.Victim.CorporationName = field(line, "军团")
			case hasField(line, "联盟"):
				km.Victim.AllianceName = field(line, "联盟")
			case hasField(line, "势力"):
				km.Victim.FactionName = field(line, "势力")
			case hasField(line, "被摧毁物"):
				km.Victim.ShipName = field(line, "被摧毁物")
			case hasField(line, "星系"):
				km.SolarSystem = field(line, "星系")
			case hasField(line, "安全等级"):
				km.SecurityStatus = field(line, "安全等级")
			case hasField(line, "所受损伤"):
				km.Victim.DamageTaken = parseNumber(field(line, "所受损伤"))
			}
			continue
		}
		if section == "attackers" {
			if hasField(line, "名称") {
				if current != nil {
					km.Attackers = append(km.Attackers, *current)
				}
				current = &Attacker{}
				name := field(line, "名称")
				if strings.Contains(name, "(给予最后一击)") || strings.Contains(name, "（给予最后一击）") {
					current.FinalBlow = true
					name = strings.ReplaceAll(name, "(给予最后一击)", "")
					name = strings.ReplaceAll(name, "（给予最后一击）", "")
				}
				current.CharacterName = strings.TrimSpace(name)
				continue
			}
			if current == nil {
				continue
			}
			switch {
			case hasField(line, "军团"):
				current.CorporationName = field(line, "军团")
			case hasField(line, "联盟"):
				current.AllianceName = field(line, "联盟")
			case hasField(line, "势力"):
				current.FactionName = field(line, "势力")
			case hasField(line, "舰船"):
				current.ShipName = field(line, "舰船")
			case hasField(line, "武器"):
				current.WeaponName = field(line, "武器")
			case hasField(line, "造成损伤"):
				current.DamageDone = parseNumber(field(line, "造成损伤"))
			}
		}
	}
	if current != nil {
		km.Attackers = append(km.Attackers, *current)
	}
	if km.Time.IsZero() || km.Victim.CharacterName == "" || km.Victim.ShipName == "" {
		return nil, errors.New("KM 缺少时间、受害者或被摧毁舰船")
	}
	return km, nil
}

func hasField(line, prefix string) bool { return strings.HasPrefix(line, prefix+":") }
func field(line, prefix string) string {
	return strings.TrimSpace(strings.TrimPrefix(line, prefix+":"))
}
func parseNumber(value string) int {
	value = strings.ReplaceAll(value, ",", "")
	parts := strings.Fields(value)
	if len(parts) == 0 {
		return 0
	}
	result, _ := strconv.Atoi(parts[0])
	return result
}
func NormalizeIdentity(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), " "))
}
