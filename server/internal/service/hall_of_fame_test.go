package service

import "testing"

func TestBuildHallOfFameCardUpdateMapOnlyIncludesProvidedFields(t *testing.T) {
	name := "Hero Alpha"
	title := "Founder"
	description := "Keeps the fleet together."

	updates, err := buildHallOfFameCardUpdateMap(&UpdateCardRequest{
		Name:        &name,
		Title:       &title,
		Description: &description,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(updates) != 3 {
		t.Fatalf("expected 3 fields, got %d (%v)", len(updates), updates)
	}

	if updates["name"] != name {
		t.Fatalf("expected name %q, got %#v", name, updates["name"])
	}
	if updates["title"] != title {
		t.Fatalf("expected title %q, got %#v", title, updates["title"])
	}
	if updates["description"] != description {
		t.Fatalf("expected description %q, got %#v", description, updates["description"])
	}
	if _, exists := updates["pos_x"]; exists {
		t.Fatalf("did not expect layout fields in partial update map: %#v", updates)
	}
}

func TestBuildHallOfFameCardUpdateMapRejectsInvalidStylePreset(t *testing.T) {
	preset := "platinum"

	_, err := buildHallOfFameCardUpdateMap(&UpdateCardRequest{StylePreset: &preset})
	if err == nil {
		t.Fatal("expected invalid style preset error")
	}
}

func TestBuildHallOfFameCardUpdateMapRejectsLayoutFields(t *testing.T) {
	posX := 120.0
	posY := -8.0
	width := 260
	height := 0
	zIndex := 7

	_, err := buildHallOfFameCardUpdateMap(&UpdateCardRequest{
		PosX:   &posX,
		PosY:   &posY,
		Width:  &width,
		Height: &height,
		ZIndex: &zIndex,
	})
	if err == nil {
		t.Fatal("expected layout fields to be rejected")
	}
}

func TestBuildHallOfFameLayoutUpdatesRejectsMissingHeight(t *testing.T) {
	width := 220

	_, err := buildHallOfFameLayoutUpdates([]CardLayoutUpdateRequest{
		{
			ID:     1,
			PosX:   22,
			PosY:   33,
			Width:  &width,
			ZIndex: 4,
		},
	})
	if err == nil {
		t.Fatal("expected missing height to fail")
	}
}

func TestBuildHallOfFameLayoutUpdatesClampsCoordinatesAndKeepsSize(t *testing.T) {
	width := 260
	height := 0

	updates, err := buildHallOfFameLayoutUpdates([]CardLayoutUpdateRequest{
		{
			ID:     2,
			PosX:   140,
			PosY:   -8,
			Width:  &width,
			Height: &height,
			ZIndex: 9,
		},
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(updates))
	}

	if updates[0].Width != width {
		t.Fatalf("expected width %d, got %d", width, updates[0].Width)
	}
	if updates[0].Height != height {
		t.Fatalf("expected height %d, got %d", height, updates[0].Height)
	}
	if updates[0].PosX != 100 {
		t.Fatalf("expected clamped pos_x 100, got %v", updates[0].PosX)
	}
	if updates[0].PosY != 0 {
		t.Fatalf("expected clamped pos_y 0, got %v", updates[0].PosY)
	}
}
