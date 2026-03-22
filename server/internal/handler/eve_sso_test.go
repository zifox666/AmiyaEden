package handler

import (
	"reflect"
	"testing"
)

func TestSplitCSV(t *testing.T) {
	input := " esi-location.read_location.v1,esi-ui.open_window.v1; esi-wallet.read_character_wallet.v1  publicData "
	want := []string{
		"esi-location.read_location.v1",
		"esi-ui.open_window.v1",
		"esi-wallet.read_character_wallet.v1",
		"publicData",
	}

	got := splitCSV(input)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("splitCSV(%q) = %v, want %v", input, got, want)
	}
}

func TestSplitAny(t *testing.T) {
	tests := []struct {
		name  string
		input string
		seps  string
		want  []string
	}{
		{
			name:  "repeated separators are ignored",
			input: ",,alpha; beta  gamma;;",
			seps:  ",; ",
			want:  []string{"alpha", "beta", "gamma"},
		},
		{
			name:  "leading and trailing separators are ignored",
			input: "  one two ",
			seps:  " ",
			want:  []string{"one", "two"},
		},
		{
			name:  "string without separators stays whole",
			input: "publicData",
			seps:  ",; ",
			want:  []string{"publicData"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitAny(tt.input, tt.seps)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("splitAny(%q, %q) = %v, want %v", tt.input, tt.seps, got, tt.want)
			}
		})
	}
}
