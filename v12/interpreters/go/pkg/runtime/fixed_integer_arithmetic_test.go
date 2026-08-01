package runtime

import (
	"math/big"
	"testing"
)

type fixedIntegerBoundaryCase struct {
	kind   IntegerType
	min    string
	max    string
	signed bool
}

func TestFixedIntegerAlternativeArithmeticCoversEveryWidth(t *testing.T) {
	cases := []fixedIntegerBoundaryCase{
		{IntegerI8, "-128", "127", true},
		{IntegerI16, "-32768", "32767", true},
		{IntegerI32, "-2147483648", "2147483647", true},
		{IntegerI64, "-9223372036854775808", "9223372036854775807", true},
		{IntegerI128, "-170141183460469231731687303715884105728", "170141183460469231731687303715884105727", true},
		{IntegerU8, "0", "255", false},
		{IntegerU16, "0", "65535", false},
		{IntegerU32, "0", "4294967295", false},
		{IntegerU64, "0", "18446744073709551615", false},
		{IntegerU128, "0", "340282366920938463463374607431768211455", false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(string(tc.kind), func(t *testing.T) {
			min := integerValueFromDecimal(t, tc.min, tc.kind)
			max := integerValueFromDecimal(t, tc.max, tc.kind)
			one := NewSmallInt(1, tc.kind)
			two := NewSmallInt(2, tc.kind)

			assertFixedIntegerArithmetic(t, max, one, FixedIntegerAdd, FixedIntegerWrapping, tc.min, true)
			assertFixedIntegerArithmetic(t, min, one, FixedIntegerSub, FixedIntegerWrapping, tc.max, true)
			wrappedProduct := "-2"
			if !tc.signed {
				wrappedProduct = new(big.Int).Sub(max.BigInt(), big.NewInt(1)).String()
			}
			assertFixedIntegerArithmetic(t, max, two, FixedIntegerMul, FixedIntegerWrapping, wrappedProduct, true)

			assertFixedIntegerArithmetic(t, max, one, FixedIntegerAdd, FixedIntegerSaturating, tc.max, true)
			assertFixedIntegerArithmetic(t, min, one, FixedIntegerSub, FixedIntegerSaturating, tc.min, true)
			assertFixedIntegerArithmetic(t, max, two, FixedIntegerMul, FixedIntegerSaturating, tc.max, true)

			assertFixedIntegerArithmetic(t, max, one, FixedIntegerAdd, FixedIntegerChecked, "", false)
			assertFixedIntegerArithmetic(t, min, one, FixedIntegerSub, FixedIntegerChecked, "", false)
			assertFixedIntegerArithmetic(t, max, two, FixedIntegerMul, FixedIntegerChecked, "", false)

			six := NewSmallInt(6, tc.kind)
			assertFixedIntegerArithmetic(t, six, two, FixedIntegerAdd, FixedIntegerChecked, "8", true)
			assertFixedIntegerArithmetic(t, six, two, FixedIntegerSub, FixedIntegerChecked, "4", true)
			assertFixedIntegerArithmetic(t, six, two, FixedIntegerMul, FixedIntegerChecked, "12", true)
		})
	}
}

func TestFixedIntegerAlternativeArithmeticRejectsNonFixedAndMixedTypes(t *testing.T) {
	if _, _, err := FixedIntegerArithmetic(
		NewSmallInt(1, IntegerIsize),
		NewSmallInt(2, IntegerIsize),
		FixedIntegerAdd,
		FixedIntegerWrapping,
	); err == nil {
		t.Fatal("expected isize rejection")
	}
	if _, _, err := FixedIntegerArithmetic(
		NewSmallInt(1, IntegerI32),
		NewSmallInt(2, IntegerI64),
		FixedIntegerAdd,
		FixedIntegerWrapping,
	); err == nil {
		t.Fatal("expected mixed-type rejection")
	}
}

func integerValueFromDecimal(t *testing.T, text string, kind IntegerType) IntegerValue {
	t.Helper()
	value, ok := new(big.Int).SetString(text, 10)
	if !ok {
		t.Fatalf("invalid integer %q", text)
	}
	if value.IsInt64() {
		return NewSmallInt(value.Int64(), kind)
	}
	return NewBigIntValue(value, kind)
}

func assertFixedIntegerArithmetic(
	t *testing.T,
	left IntegerValue,
	right IntegerValue,
	operation FixedIntegerOperation,
	mode FixedIntegerMode,
	want string,
	wantPresent bool,
) {
	t.Helper()
	got, present, err := FixedIntegerArithmetic(left, right, operation, mode)
	if err != nil {
		t.Fatalf("FixedIntegerArithmetic failed: %v", err)
	}
	if present != wantPresent {
		t.Fatalf("present = %v, want %v", present, wantPresent)
	}
	if !wantPresent {
		return
	}
	if got.TypeSuffix != left.TypeSuffix {
		t.Fatalf("suffix = %s, want %s", got.TypeSuffix, left.TypeSuffix)
	}
	if got.BigInt().String() != want {
		t.Fatalf("result = %s, want %s", got.BigInt(), want)
	}
}
