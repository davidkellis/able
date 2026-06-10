package interpreter

import (
	"math"
	"math/big"
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestApplyUnaryOperatorNegateSmallIntegerUsesCachedValue(t *testing.T) {
	interp := New()
	want, ok := bytecodeBoxedIntegerValue(runtime.IntegerI32, -1)
	if !ok {
		t.Fatalf("expected cached boxed -1_i32")
	}

	got, err := interp.applyUnaryOperator("-", runtime.NewSmallInt(1, runtime.IntegerI32))
	if err != nil {
		t.Fatalf("applyUnaryOperator failed: %v", err)
	}
	if got != want {
		t.Fatalf("negated identity = %#v, want cached boxed value %#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, -1)
}

func TestApplyUnaryOperatorNegateSmallIntegerPointer(t *testing.T) {
	interp := New()
	src := runtime.NewSmallInt(3, runtime.IntegerI32)

	got, err := interp.applyUnaryOperator("-", &src)
	if err != nil {
		t.Fatalf("applyUnaryOperator failed: %v", err)
	}
	assertIntValue(t, got, runtime.IntegerI32, -3)
}

func TestApplyUnaryOperatorNegateRawI32SlotValue(t *testing.T) {
	interp := New()

	got, err := interp.applyUnaryOperator("-", bytecodeRawI32SlotCachedValue(4))
	if err != nil {
		t.Fatalf("applyUnaryOperator failed: %v", err)
	}
	assertIntValue(t, got, runtime.IntegerI32, -4)
}

func TestApplyUnaryOperatorNegateSmallIntegerHotPathIsAllocationFree(t *testing.T) {
	interp := New()
	src := runtime.NewSmallInt(1, runtime.IntegerI32)
	want, ok := bytecodeBoxedIntegerValue(runtime.IntegerI32, -1)
	if !ok {
		t.Fatalf("expected cached boxed -1_i32")
	}

	allocs := testing.AllocsPerRun(1000, func() {
		got, err := interp.applyUnaryOperator("-", &src)
		if err != nil {
			panic(err)
		}
		if got != want {
			panic("unexpected negated value")
		}
	})
	if allocs != 0 {
		t.Fatalf("expected small integer negate hot path allocations to be zero, got %.2f", allocs)
	}
}

func TestApplyUnaryOperatorNegateMinInt64I128FallsBackToBigInt(t *testing.T) {
	interp := New()
	src := runtime.NewSmallInt(math.MinInt64, runtime.IntegerI128)

	got, err := interp.applyUnaryOperator("-", src)
	if err != nil {
		t.Fatalf("applyUnaryOperator failed: %v", err)
	}
	intVal, ok := got.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("result type = %T, want runtime.IntegerValue", got)
	}
	if intVal.IsSmall() {
		t.Fatalf("expected big-int result for negated MinInt64 i128")
	}
	want := new(big.Int).Neg(big.NewInt(math.MinInt64))
	if intVal.BigInt().Cmp(want) != 0 {
		t.Fatalf("negated MinInt64 i128 = %v, want %v", intVal.BigInt(), want)
	}
}
