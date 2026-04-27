package types

import "testing"

func TestPhase_String_Roundtrip(t *testing.T) {
	for _, p := range AllPhases() {
		if p.String() == "unknown" {
			t.Errorf("phase %d has no string mapping", int(p))
		}
	}
}

func TestBlockerCode_IsKnown(t *testing.T) {
	tests := []struct {
		code BlockerCode
		want bool
	}{
		{BlockerFKRequired, true},
		{BlockerOffsetPaginationIncompat, true},
		{BlockerSecretInSource, true},
		{BlockerCompileFailed, true},
		{BlockerEnforcerViolation, true},
		{BlockerCode("MADE_UP_CODE"), false},
		{BlockerCode(""), false},
	}
	for _, tt := range tests {
		if got := tt.code.IsKnown(); got != tt.want {
			t.Errorf("BlockerCode(%q).IsKnown() = %v, want %v", tt.code, got, tt.want)
		}
	}
}

func TestAllPhases_CountMatchesPlan(t *testing.T) {
	// Plan §5 specifies 14 phases (0-13).
	if got := len(AllPhases()); got != 14 {
		t.Errorf("AllPhases() returned %d phases, want 14", got)
	}
}
