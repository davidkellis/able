package runtime

import "testing"

func TestHasherSemanticStateRoundTripContract(t *testing.T) {
	const state = uint64(0xfedcba9876543210)
	hasher := NewHasherValueFromState(state)
	if got := hasher.SemanticState(); got != state {
		t.Fatalf("Hasher state = %#x, want %#x", got, state)
	}
	if got := (*HasherValue)(nil).SemanticState(); got != 0 {
		t.Fatalf("nil Hasher state = %#x, want 0", got)
	}
}
