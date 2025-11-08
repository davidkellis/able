package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestHasherBuiltinsProduceDeterministicHash(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	if _, err := interp.evaluateExpression(ast.Assign(ast.ID("hasher"), ast.Call("__able_hasher_create")), global); err != nil {
		t.Fatalf("hasher creation failed: %v", err)
	}
	if _, err := interp.evaluateExpression(ast.Call("__able_hasher_write", ast.ID("hasher"), ast.Str("abc")), global); err != nil {
		t.Fatalf("hasher write failed: %v", err)
	}
	val, err := interp.evaluateExpression(ast.Call("__able_hasher_finish", ast.ID("hasher")), global)
	if err != nil {
		t.Fatalf("hasher finish failed: %v", err)
	}
	intVal, ok := val.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %#v", val)
	}
	expected := big.NewInt(0x1a47e90b)
	if intVal.Val.Cmp(expected) != 0 {
		t.Fatalf("expected hash 0x1a47e90c, got %s", intVal.Val.String())
	}
}
