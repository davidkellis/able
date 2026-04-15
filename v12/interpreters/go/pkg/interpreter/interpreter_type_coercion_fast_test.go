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
