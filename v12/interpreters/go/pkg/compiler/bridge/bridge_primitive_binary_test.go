package bridge

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestApplyBinaryOperatorWithoutInterpreterHandlesPrimitiveIntegers(t *testing.T) {
	rt := &Runtime{}
	left := runtime.NewSmallInt(10, runtime.IntegerI32)
	right := runtime.NewSmallInt(7, runtime.IntegerI32)

	sum, err := ApplyBinaryOperator(rt, "+", left, right)
	if err != nil {
		t.Fatalf("primitive add failed: %v", err)
	}
	sumInt, ok := sum.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("primitive add returned %T, want runtime.IntegerValue", sum)
	}
	if value, fits := sumInt.ToInt64(); !fits || value != 17 || sumInt.TypeSuffix != runtime.IntegerI32 {
		t.Fatalf("primitive add returned %#v, want i32 17", sumInt)
	}

	greater, err := ApplyBinaryOperator(rt, ">", sum, right)
	if err != nil {
		t.Fatalf("primitive comparison failed: %v", err)
	}
	if value, ok := greater.(runtime.BoolValue); !ok || !value.Val {
		t.Fatalf("primitive comparison returned %#v, want true", greater)
	}
}

func TestApplyBinaryOperatorWithoutInterpreterDoesNotClaimNominalDispatch(t *testing.T) {
	_, err := ApplyBinaryOperator(&Runtime{}, "+", &runtime.StructInstanceValue{}, &runtime.StructInstanceValue{})
	if err == nil || !strings.Contains(err.Error(), "missing interpreter") {
		t.Fatalf("nominal operator error = %v, want missing interpreter", err)
	}
}

func TestApplyBinaryOperatorWithoutInterpreterDoesNotTreatDottedFloatXorAsExponent(t *testing.T) {
	left := runtime.FloatValue{Val: 2, TypeSuffix: runtime.FloatF64}
	right := runtime.FloatValue{Val: 3, TypeSuffix: runtime.FloatF64}
	_, err := ApplyBinaryOperator(&Runtime{}, ".^", left, right)
	if err == nil || !strings.Contains(err.Error(), "missing interpreter") {
		t.Fatalf("dotted float xor error = %v, want missing interpreter", err)
	}
}
