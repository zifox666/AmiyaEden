package model

import "testing"

func TestNormalizePAPLevel(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		want  string
	}{
		// CTA variants
		{name: "exact CTA", in: "CTA", want: PAPTypeCTA},
		{name: "lowercase cta", in: "cta", want: PAPTypeCTA},
		{name: "cta op", in: "cta op", want: PAPTypeCTA},
		{name: "CTA mixed case", in: "Cta", want: PAPTypeCTA},

		// Strat Op variants
		{name: "Strat Op from API", in: "Strat Op", want: PAPTypeStratOp},
		{name: "strat_op snake case", in: "strat_op", want: PAPTypeStratOp},
		{name: "strategic", in: "strategic", want: PAPTypeStratOp},
		{name: "strategic op", in: "strategic op", want: PAPTypeStratOp},
		{name: "Strat Op mixed case", in: "STRAT OP", want: PAPTypeStratOp},

		// Skirmish fallback — any unrecognised value
		{name: "empty string", in: "", want: PAPTypeSkirmish},
		{name: "unknown level", in: "Random Op", want: PAPTypeSkirmish},
		{name: "other keyword", in: "other", want: PAPTypeSkirmish},

		// Whitespace trimming
		{name: "leading and trailing spaces CTA", in: "  CTA  ", want: PAPTypeCTA},
		{name: "leading space Strat Op", in: " Strat Op", want: PAPTypeStratOp},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizePAPLevel(tt.in); got != tt.want {
				t.Fatalf("NormalizePAPLevel(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
