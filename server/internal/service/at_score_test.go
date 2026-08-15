package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/pkg/xlsx"
	"os"
	"testing"
)

func TestParseATXXIIV14Workbook(t *testing.T) {
	path := os.Getenv("AT_SCORE_TEST_XLSX")
	if path == "" {
		t.Skip("AT_SCORE_TEST_XLSX not set")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	book, err := xlsx.Open(data)
	if err != nil {
		t.Fatal(err)
	}
	version, ships, banned, err := parseATScoreWorkbook(book)
	if err != nil {
		t.Fatal(err)
	}
	if version != "V1.4" || len(ships) != 278 || len(banned) != 0 {
		t.Fatalf("unexpected workbook: version=%q ships=%d banned=%d", version, len(ships), len(banned))
	}
	foundTholos, foundGeri := false, false
	for _, ship := range ships {
		if ship.Name == "Tholos" {
			foundTholos = true
			if ship.Points != 13 || ship.Size != "DE" {
				t.Fatalf("unexpected Tholos rule: %+v", ship)
			}
		}
		if ship.Name == "Geri" {
			foundGeri = true
			if ship.Size != "FR3" || ship.Penalty != 3 {
				t.Fatalf("unexpected Geri inflation rule: %+v", ship)
			}
		}
	}
	if !foundTholos {
		t.Fatal("Tholos missing")
	}
	if !foundGeri {
		t.Fatal("Geri missing")
	}
}

func TestCalculateATRosterDuplicatePenaltyAndBoundaries(t *testing.T) {
	ships := []model.ATScoreShip{
		{TypeID: 1, TypeName: "Cruiser A", BasePoints: 20, SizeCode: "CR"},
		{TypeID: 2, TypeName: "Battleship B", BasePoints: 27, SizeCode: "BS"},
		{TypeID: 3, TypeName: "Frigate C", BasePoints: 10, SizeCode: "FR"},
	}
	penalties := []model.ATScoreDuplicatePenalty{
		{SizeCode: "CR", Penalty: 2}, {SizeCode: "BS", Penalty: 4}, {SizeCode: "FR", Penalty: 0}, {SizeCode: "OTHER", Penalty: 0},
	}
	result, err := CalculateATRoster("7v7", []int64{1, 1, 2, 2, 3}, ships, nil, penalties)
	if err != nil {
		t.Fatal(err)
	}
	// Two cruisers score 22 each, two battleships 31 each, one frigate 10.
	if result.TotalPoints != 116 || !result.Valid {
		t.Fatalf("unexpected result: %+v", result)
	}

	exact := []model.ATScoreShip{{TypeID: 9, TypeName: "Exact", BasePoints: 25, SizeCode: "FR"}}
	result, err = CalculateATRoster("10v10", []int64{9, 9, 9, 9, 9, 9, 9, 9}, exact, nil, penalties)
	if err != nil || result.TotalPoints != 200 || !result.Valid {
		t.Fatalf("200 point boundary failed: %+v, %v", result, err)
	}
}

func TestCalculateATRosterRejectsBannedUnknownAndOverLimit(t *testing.T) {
	ships := []model.ATScoreShip{{TypeID: 1, TypeName: "Banned", BasePoints: 1, SizeCode: "FR"}}
	banned := []model.ATScoreBannedShip{{TypeID: 1, TypeName: "Banned"}}
	ids := []int64{1, 2, 1, 1, 1, 1, 1, 1}
	result, err := CalculateATRoster("7v7", ids, ships, banned, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Valid || len(result.Errors) < 3 {
		t.Fatalf("expected slot, banned and unknown errors: %+v", result)
	}
}
