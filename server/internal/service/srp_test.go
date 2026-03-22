package service

import "testing"

func TestSlotCategory(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "high slot with digit", in: "HiSlot0", want: "HiSlot"},
		{name: "med slot with multiple digits", in: "MedSlot12", want: "MedSlot"},
		{name: "already normalized", in: "Cargo", want: "Cargo"},
		{name: "implant", in: "Implant", want: "Implant"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := slotCategory(tt.in); got != tt.want {
				t.Fatalf("slotCategory(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSlotCategoryNamesContainRequiredLocales(t *testing.T) {
	requiredCategories := []string{"HiSlot", "MedSlot", "LoSlot", "Cargo"}

	for _, category := range requiredCategories {
		names, ok := slotCategoryNames[category]
		if !ok {
			t.Fatalf("missing slotCategoryNames entry for %q", category)
		}
		if names["zh"] == "" {
			t.Fatalf("missing zh name for %q", category)
		}
		if names["en"] == "" {
			t.Fatalf("missing en name for %q", category)
		}
	}
}
