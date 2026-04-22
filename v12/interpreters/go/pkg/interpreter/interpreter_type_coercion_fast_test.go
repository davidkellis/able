package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestInterpreterCoerceValueToTypeWouldBeNoOp(t *testing.T) {
	interp := New()
	interp.interfaces["Box"] = &runtime.InterfaceDefinitionValue{}

	if !interp.coerceValueToTypeWouldBeNoOp(ast.Gen(ast.Ty("Array"), ast.Ty("i32"))) {
		t.Fatalf("expected Array i32 coercion to be a no-op")
	}
	if interp.coerceValueToTypeWouldBeNoOp(ast.Gen(ast.Ty("Box"), ast.Ty("i32"))) {
		t.Fatalf("expected generic interface coercion to require runtime work")
	}
	if interp.coerceValueToTypeWouldBeNoOp(ast.Ty("i32")) {
		t.Fatalf("expected simple primitive coercion to remain active")
	}
}

func TestInterpreterCastValueToCanonicalSimpleTypeFast(t *testing.T) {
	intVal := runtime.NewSmallInt(7, runtime.IntegerI32)

	casted, ok, err := castValueToCanonicalSimpleTypeFast("i32", intVal)
	if err != nil {
		t.Fatalf("unexpected same-type cast error: %v", err)
	}
	if !ok {
		t.Fatalf("expected i32 fast cast to handle same-type integer")
	}
	if !valuesEqual(casted, intVal) {
		t.Fatalf("unexpected same-type cast result: got=%#v want=%#v", casted, intVal)
	}

	floatCasted, ok, err := castValueToCanonicalSimpleTypeFast("f64", intVal)
	if err != nil {
		t.Fatalf("unexpected integer-to-float cast error: %v", err)
	}
	if !ok {
		t.Fatalf("expected f64 fast cast to handle integer input")
	}
	floatVal, ok := floatCasted.(runtime.FloatValue)
	if !ok {
		t.Fatalf("expected float result, got %T", floatCasted)
	}
	if floatVal.TypeSuffix != runtime.FloatF64 || floatVal.Val != 7 {
		t.Fatalf("unexpected float cast result: got=%#v", floatVal)
	}
}

func TestInterpreterCastValueToCanonicalSimpleTypeFast_SmallIntegerWraps(t *testing.T) {
	negVal := runtime.NewSmallInt(-1, runtime.IntegerI16)

	u8Casted, ok, err := castValueToCanonicalSimpleTypeFast("u8", negVal)
	if err != nil {
		t.Fatalf("unexpected u8 cast error: %v", err)
	}
	if !ok {
		t.Fatalf("expected u8 fast cast to handle small integer input")
	}
	u8Val, ok := u8Casted.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T", u8Casted)
	}
	if !u8Val.IsSmall() || u8Val.TypeSuffix != runtime.IntegerU8 || u8Val.Int64Fast() != 255 {
		t.Fatalf("unexpected u8 cast result: %#v", u8Val)
	}

	u64Casted, ok, err := castValueToCanonicalSimpleTypeFast("u64", negVal)
	if err != nil {
		t.Fatalf("unexpected u64 cast error: %v", err)
	}
	if !ok {
		t.Fatalf("expected u64 fast cast to handle small integer input")
	}
	u64Val, ok := u64Casted.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T", u64Casted)
	}
	if u64Val.IsSmall() {
		t.Fatalf("expected wrapped u64 cast to fall back to big integer storage")
	}
	if got := u64Val.BigInt().String(); got != "18446744073709551615" {
		t.Fatalf("unexpected u64 wrap result: %s", got)
	}
}

func TestInterpreterCastValueToCanonicalSimpleTypeFast_SmallIntegerWrapsWithoutAlloc(t *testing.T) {
	negVal := runtime.NewSmallInt(-1, runtime.IntegerI16)
	allocs := testing.AllocsPerRun(1000, func() {
		casted, ok, err := castValueToCanonicalSimpleTypeFast("u8", negVal)
		if err != nil {
			t.Fatalf("unexpected u8 cast error: %v", err)
		}
		if !ok {
			t.Fatalf("expected u8 fast cast to handle small integer input")
		}
		u8Val, ok := casted.(runtime.IntegerValue)
		if !ok || !u8Val.IsSmall() || u8Val.Int64Fast() != 255 {
			t.Fatalf("unexpected u8 cast result: %#v", casted)
		}
	})
	if allocs > 1 {
		t.Fatalf("expected repeated small u8 casts to stay at or below one allocation, got %.2f", allocs)
	}
}
