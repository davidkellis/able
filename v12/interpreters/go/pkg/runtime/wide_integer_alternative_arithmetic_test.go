package runtime

import "testing"

func TestWideIntegerAlternativeArithmeticUsesNativeCarriers(t *testing.T) {
	unsignedMax := Uint128{High: ^uint64(0), Low: ^uint64(0)}
	if got := unsignedMax.WrappingAdd(Uint128{Low: 1}); !got.IsZero() {
		t.Fatalf("u128 wrapping add = %#v, want zero", got)
	}
	if got := (Uint128{}).WrappingSub(Uint128{Low: 1}); got != unsignedMax {
		t.Fatalf("u128 wrapping sub = %#v, want %#v", got, unsignedMax)
	}
	if got := unsignedMax.SaturatingMul(Uint128{Low: 2}); got != unsignedMax {
		t.Fatalf("u128 saturating mul = %#v, want %#v", got, unsignedMax)
	}

	signedMax := Int128{High: ^(uint64(1) << 63), Low: ^uint64(0)}
	signedMin := Int128{High: uint64(1) << 63}
	if got := signedMax.WrappingAdd(Int128{Low: 1}); got != signedMin {
		t.Fatalf("i128 wrapping add = %#v, want %#v", got, signedMin)
	}
	if got := signedMin.WrappingSub(Int128{Low: 1}); got != signedMax {
		t.Fatalf("i128 wrapping sub = %#v, want %#v", got, signedMax)
	}
	if got := signedMax.SaturatingMul(Int128{Low: 2}); got != signedMax {
		t.Fatalf("i128 saturating mul = %#v, want %#v", got, signedMax)
	}
	negativeTwo := Int128{High: ^uint64(0), Low: ^uint64(1)}
	if got := signedMax.WrappingMul(Int128{Low: 2}); got != negativeTwo {
		t.Fatalf("i128 wrapping mul = %#v, want %#v", got, negativeTwo)
	}
}
