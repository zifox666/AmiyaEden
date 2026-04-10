package service

import "testing"

func TestBuildHallOfFameCardUpdateMapOnlyIncludesProvidedFields(t *testing.T) {
	name := "Hero Alpha"
	title := "Founder"
	description := "Keeps the fleet together."
	characterID := int64(1387156123)
	badgeImage := "data:image/png;base64,abc"
	titleColor := "#ff8fc7"

	updates, err := buildHallOfFameCardUpdateMap(&UpdateCardRequest{
		Name:        &name,
		Title:       &title,
		Description: &description,
		CharacterID: &characterID,
		BadgeImage:  &badgeImage,
		TitleColor:  &titleColor,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(updates) != 6 {
		t.Fatalf("expected 6 fields, got %d (%v)", len(updates), updates)
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
	if updates["character_id"] != characterID {
		t.Fatalf("expected character_id %d, got %#v", characterID, updates["character_id"])
	}
	if _, exists := updates["avatar"]; exists {
		t.Fatalf("did not expect legacy avatar update, got %#v", updates["avatar"])
	}
	if updates["badge_image"] != badgeImage {
		t.Fatalf("expected badge_image %q, got %#v", badgeImage, updates["badge_image"])
	}
	if updates["title_color"] != titleColor {
		t.Fatalf("expected title_color %q, got %#v", titleColor, updates["title_color"])
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

func TestBuildHallOfFameCardUpdateMapAllowsNewStylePresets(t *testing.T) {
	for _, preset := range []string{"rose", "jade", "midnight"} {
		preset := preset

		t.Run(preset, func(t *testing.T) {
			updates, err := buildHallOfFameCardUpdateMap(&UpdateCardRequest{StylePreset: &preset})
			if err != nil {
				t.Fatalf("expected preset %q to be accepted, got %v", preset, err)
			}

			if updates["style_preset"] != preset {
				t.Fatalf("expected style_preset %q, got %#v", preset, updates["style_preset"])
			}
		})
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

func TestBuildHallOfFameCardUpdateMapAcceptsBorderStyle(t *testing.T) {
	for _, style := range []string{"none", "gilded", "imperial", "neon-circuit", "void-rift", "amarr", "caldari", "minmatar", "gallente"} {
		style := style

		t.Run(style, func(t *testing.T) {
			updates, err := buildHallOfFameCardUpdateMap(&UpdateCardRequest{BorderStyle: &style})
			if err != nil {
				t.Fatalf("expected border style %q to be accepted, got %v", style, err)
			}

			if updates["border_style"] != style {
				t.Fatalf("expected border_style %q, got %#v", style, updates["border_style"])
			}
		})
	}
}

func TestBuildHallOfFameCardUpdateMapRejectsInvalidBorderStyle(t *testing.T) {
	style := "rainbow-sparkle"

	_, err := buildHallOfFameCardUpdateMap(&UpdateCardRequest{BorderStyle: &style})
	if err == nil {
		t.Fatal("expected invalid border style error")
	}
}
