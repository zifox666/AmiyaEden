package atkm

import (
	"strings"
	"testing"
)

func TestParseManyAndFullWidthColon(t *testing.T) {
	one := "2026.01.01 10:00:00\r\n受害者：Pilot A\r\n被摧毁物：Rifter\r\n所受损伤：1,200\r\n参与者：\r\n名称：Pilot B（给予最后一击）\r\n舰船：Thrasher\r\n武器：280mm Howitzer Artillery II\r\n造成损伤：1200\r\n掉落物品：\r\nIgnored"
	result, err := ParseMany(one + "\n\n" + one)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("count=%d", len(result))
	}
	if !result[0].Attackers[0].FinalBlow || result[0].Victim.DamageTaken != 1200 {
		t.Fatalf("unexpected parse: %+v", result[0])
	}
	if !strings.Contains(result[0].RawText, "受害者：") || !strings.Contains(result[0].RawText, "\r\n") {
		t.Fatalf("raw input was not preserved: %q", result[0].RawText)
	}
}

func TestParseManyRejectsMissingRequiredFields(t *testing.T) {
	if _, err := ParseMany("2026.01.01 10:00:00\n受害者：Pilot A"); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestParseIgnoresDestroyedAndDroppedItems(t *testing.T) {
	text := "2026.01.01 10:00:00\n受害者: Pilot A\n被摧毁物: Rifter\n所受损伤: 10\n参与者:\n名称: Pilot B\n舰船: Rifter\n武器: 200mm AutoCannon II\n造成损伤: 10\n被毁物品:\n舰船: Fake Ship\n武器: Fake Weapon\n掉落物品:\n舰船: Another Fake"
	result, err := ParseMany(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || len(result[0].Attackers) != 1 || result[0].Attackers[0].ShipName != "Rifter" {
		t.Fatalf("item section leaked into attackers: %+v", result)
	}
}
