package runtime

import (
	"math"
	"testing"
)

func TestRawIntegerMaterializePreservesUnsignedHighBit(t *testing.T) {
	want := uint64(math.MaxInt64) + 1
	for _, suffix := range []IntegerType{IntegerU64, IntegerU128, IntegerUsize} {
		raw := NewRawIntegerValue(suffix, int64(want))
		value := raw.Materialize()
		got, ok := value.(IntegerValue)
		if !ok {
			t.Fatalf("Materialize(%s) type = %T, want IntegerValue", suffix, value)
		}
		if got.IsSmall() || got.TypeSuffix != suffix || got.BigInt().Uint64() != want {
			t.Fatalf("Materialize(%s) = %#v, want non-small unsigned %d", suffix, got, want)
		}
	}
}
