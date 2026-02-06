package bridge

import (
	"math/big"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestAsFloatAcceptsInteger(t *testing.T) {
	val := runtime.IntegerValue{Val: big.NewInt(42), TypeSuffix: runtime.IntegerI64}
	out, err := AsFloat(val)
	if err != nil {
		t.Fatalf("AsFloat error: %v", err)
	}
	if out != 42 {
		t.Fatalf("AsFloat = %v, want 42", out)
	}
}

func TestAsStringAcceptsStringStruct(t *testing.T) {
	byteArr := &runtime.ArrayValue{
		Elements: []runtime.Value{
			runtime.IntegerValue{Val: big.NewInt(72), TypeSuffix: runtime.IntegerU8},
			runtime.IntegerValue{Val: big.NewInt(105), TypeSuffix: runtime.IntegerU8},
		},
	}
	definition := &runtime.StructDefinitionValue{
		Node: ast.StructDef("String", nil, ast.StructKindNamed, nil, nil, false),
	}
	inst := &runtime.StructInstanceValue{
		Definition: definition,
		Fields: map[string]runtime.Value{
			"bytes":     byteArr,
			"len_bytes": runtime.IntegerValue{Val: big.NewInt(2), TypeSuffix: runtime.IntegerI32},
		},
	}
	out, err := AsString(inst)
	if err != nil {
		t.Fatalf("AsString error: %v", err)
	}
	if out != "Hi" {
		t.Fatalf("AsString = %q, want %q", out, "Hi")
	}
}
