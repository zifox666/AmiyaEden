package service

import (
	"testing"
	"time"

	"amiya-eden/internal/model"
)

func TestParseCombatLogEnglish(t *testing.T) {
	log := `------------------------------------------------------------
  Gamelog
  Listener: Capsuleer One
  Session Started: 2026.08.15 10:00:00
------------------------------------------------------------
[ 2026.08.15 10:00:01 ] (combat) <color=0xff00ffff><b>1,200</b></color><hint=target>to</hint><color=0xffffffff><b>Target Pilot[CORP] (Rifter)</b></color> - 425mm AutoCannon II - Hits
[ 2026.08.15 10:00:02 ] (combat) 300 from Enemy Pilot[RED] (Caracal) - Heavy Missile - Hits
[ 2026.08.15 10:00:03 ] (combat) <color=0xff00ffff><b>250</b></color> remote shield boosted to <color=0xffffffff><b>Friendly Pilot[BLUE] (Scimitar)</b></color> - Large Shield Transporter II
[ 2026.08.15 10:00:04 ] (combat) <color=0xff7fffff><b>90</b></color> energy neutralized <color=0xffffffff><b>Enemy Pilot[RED] (Caracal)</b></color> - Medium Energy Neutralizer II
[ 2026.08.15 10:00:05 ] (mining) Your Miner II completes a cycle and <color=0xff00aa00><b>166</b> units of <color=0xffffffff><b>Veldspar</b> were mined`

	parsed, err := parseCombatLog([]byte(log))
	if err != nil {
		t.Fatalf("parseCombatLog() error = %v", err)
	}
	if parsed.CharacterName != "Capsuleer One" || parsed.Language != "english" {
		t.Fatalf("unexpected identity: %#v", parsed)
	}
	if len(parsed.Events) != 5 {
		t.Fatalf("event count = %d, want 5", len(parsed.Events))
	}
	if parsed.Events[0].Amount != 1200 || parsed.Events[0].PilotName != "Target Pilot" || parsed.Events[0].WeaponType != "425mm AutoCannon II" {
		t.Fatalf("unexpected outgoing damage event: %#v", parsed.Events[0])
	}
	if parsed.Events[4].Category != combatCategoryMining || parsed.Events[4].WeaponType != "Veldspar" {
		t.Fatalf("unexpected mining event: %#v", parsed.Events[4])
	}

	report := &model.CombatLogReport{}
	breakdowns, timeline, hitQualities := buildCombatAnalysis(parsed, report)
	if report.DamageOut != 1200 || report.DamageIn != 300 || report.LogisticsOut != 250 || report.CapDamageOut != 90 || report.MiningUnits != 166 {
		t.Fatalf("unexpected totals: %#v", report)
	}
	if len(breakdowns) != 5 || len(timeline) != 5 {
		t.Fatalf("breakdowns=%d timeline=%d", len(breakdowns), len(timeline))
	}
	if len(hitQualities) != 2 || hitQualities[0].Quality != combatHitHits || hitQualities[0].HitCount != 1 || hitQualities[1].Quality != combatHitHits || hitQualities[1].HitCount != 1 {
		t.Fatalf("unexpected hit qualities: %#v", hitQualities)
	}
}

func TestParseCombatLogChineseNosProducesTwoMetrics(t *testing.T) {
	log := `游戏日志
收听者: 测试飞行员
进程开始: 2026.08.15 10:00:00
[ 2026.08.15 10:00:01 ] (combat) <color=0xff00ffff><b>88</b>被从<color=0xffffffff><b>敌对飞行员[TEST] (毒蜥级)</b> - 中型掠能器 II - 命中
[ 2026.08.15 10:00:02 ] (combat) <color=0xff00ffff><b>500</b>对<color=0xffffffff><b>目标[TEST] (渡鸦级)</b> - 巡航导弹发射器 II - 命中`
	parsed, err := parseCombatLog([]byte(log))
	if err != nil {
		t.Fatalf("parseCombatLog() error = %v", err)
	}
	if len(parsed.Events) != 3 {
		t.Fatalf("event count = %d, want 3", len(parsed.Events))
	}
	report := &model.CombatLogReport{}
	buildCombatAnalysis(parsed, report)
	if report.CapTransferIn != 88 || report.CapDamageOut != 88 || report.DamageOut != 500 {
		t.Fatalf("unexpected totals: %#v", report)
	}
}

func TestParseCombatLogChineseHitQualityDistribution(t *testing.T) {
	log := `游戏日志
收听者: 测试飞行员
[ 2026.08.15 10:00:01 ] (combat) 10 对 目标一[TEST] (护卫舰) - 火炮 - 轻轻擦过
[ 2026.08.15 10:00:02 ] (combat) 20 对 目标一[TEST] (护卫舰) - 火炮 - 命中
[ 2026.08.15 10:00:03 ] (combat) 30 对 目标一[TEST] (护卫舰) - 火炮 - 强力一击
[ 2026.08.15 10:00:04 ] (combat) 40 对 目标一[TEST] (护卫舰) - 火炮 - 穿透
[ 2026.08.15 10:00:05 ] (combat) 50 对 目标一[TEST] (护卫舰) - 火炮 - 致命一击
[ 2026.08.15 10:00:06 ] (combat) 60 对 目标一[TEST] (护卫舰) - 火炮 - 擦过
[ 2026.08.15 10:00:07 ] (combat) 70 来自 敌人[TEST] (巡洋舰) - 激光炮 - 擦过`

	parsed, err := parseCombatLog([]byte(log))
	if err != nil {
		t.Fatalf("parseCombatLog() error = %v", err)
	}
	_, _, rows := buildCombatAnalysis(parsed, &model.CombatLogReport{})
	if len(rows) != 7 {
		t.Fatalf("hit quality rows = %d, want 7: %#v", len(rows), rows)
	}
	want := []string{combatHitBarelyScratches, combatHitHits, combatHitSmashes, combatHitPenetrates, combatHitWrecks, combatHitGlancesOff}
	for index, quality := range want {
		if rows[index+1].Direction != combatCategoryDamageOut || rows[index+1].Quality != quality || rows[index+1].HitCount != 1 {
			t.Fatalf("outgoing row %d = %#v, want quality %s", index, rows[index+1], quality)
		}
	}
	if rows[0].Direction != combatCategoryDamageIn || rows[0].Quality != combatHitGlancesOff || rows[0].DamageAmount != 70 {
		t.Fatalf("unexpected incoming row: %#v", rows[0])
	}
}

func TestParseCombatLogRejectsUnsupportedContent(t *testing.T) {
	if _, err := parseCombatLog([]byte("not an eve log")); err == nil {
		t.Fatal("expected unsupported content error")
	}
	log := "Listener: Pilot\n[ 2026.08.15 10:00:00 ] (notify) nothing"
	if _, err := parseCombatLog([]byte(log)); err == nil {
		t.Fatal("expected no recognized event error")
	}
}

func TestCombatLogSingleEventDurationIsUsable(t *testing.T) {
	parsed := &combatParsedLog{
		StartedAt: time.Date(2026, 8, 15, 1, 2, 3, 0, time.UTC),
		EndedAt:   time.Date(2026, 8, 15, 1, 2, 3, 0, time.UTC),
		Events: []combatParsedEvent{{
			OccurredAt: time.Date(2026, 8, 15, 1, 2, 3, 0, time.UTC),
			Category:   combatCategoryDamageOut,
			Amount:     100,
		}},
	}
	report := &model.CombatLogReport{DurationSeconds: 1}
	buildCombatAnalysis(parsed, report)
	if report.DamageOut != 100 || report.EventCount != 1 {
		t.Fatalf("unexpected report: %#v", report)
	}
}
