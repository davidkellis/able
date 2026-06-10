package bridge

import (
	"math/big"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestStructNamedFieldValueReadsMapAndPositionalNamedStorage(t *testing.T) {
	def := &runtime.StructDefinitionValue{
		Node: ast.StructDef("Record", []*ast.StructFieldDefinition{
			ast.NewStructFieldDefinition(ast.Ty("i32"), ast.ID("value")),
		}, ast.StructKindNamed, nil, nil, false),
		NamedFieldIndices: map[string]int{"value": 0},
	}
	want := runtime.IntegerValue{Val: big.NewInt(42), TypeSuffix: runtime.IntegerI32}

	for _, inst := range []*runtime.StructInstanceValue{
		{Definition: def, Fields: map[string]runtime.Value{"value": want}},
		{Definition: def, Positional: []runtime.Value{want}},
	} {
		got, ok := structNamedFieldValue(inst, "value")
		if !ok {
			t.Fatal("value field not found")
		}
		if value, ok := got.(runtime.IntegerValue); !ok || value.Val.Cmp(want.Val) != 0 {
			t.Fatalf("value = %#v, want %#v", got, want)
		}
	}
}

func TestRuntimeValueToHostReadsPositionalNamedStorage(t *testing.T) {
	def := &runtime.StructDefinitionValue{
		Node: ast.StructDef("Box", []*ast.StructFieldDefinition{
			ast.NewStructFieldDefinition(ast.Ty("i32"), ast.ID("value")),
		}, ast.StructKindNamed, nil, nil, false),
		NamedFieldIndices: map[string]int{"value": 0},
	}
	inst := &runtime.StructInstanceValue{
		Definition: def,
		Positional: []runtime.Value{
			runtime.IntegerValue{Val: big.NewInt(42), TypeSuffix: runtime.IntegerI32},
		},
	}

	got, err := RuntimeValueToHost[any](ast.Ty("Box"), inst)
	if err != nil {
		t.Fatalf("RuntimeValueToHost: %v", err)
	}
	fields, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("host value = %T, want map[string]any", got)
	}
	if value, ok := fields["value"].(int32); !ok || value != 42 {
		t.Fatalf("host Box.value = %#v, want int32(42)", fields["value"])
	}
}
