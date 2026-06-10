package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestNamedStructPatternPlanCachedReusesPlanForSameDefinition(t *testing.T) {
	interp := New()
	def := ast.StructDef(
		"Point",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "x"),
			ast.FieldDef(ast.Ty("i32"), "y"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	pattern := ast.StructP([]*ast.StructPatternField{
		ast.FieldP(ast.ID("left"), "x", nil),
		ast.FieldP(ast.ID("right"), "y", nil),
	}, false, "Point")

	first, ok := interp.namedStructPatternPlanCached(pattern, def)
	if !ok {
		t.Fatalf("expected first pattern plan")
	}
	second, ok := interp.namedStructPatternPlanCached(pattern, def)
	if !ok {
		t.Fatalf("expected second pattern plan")
	}
	if len(first.fieldOrder) != 2 || len(second.fieldOrder) != 2 {
		t.Fatalf("unexpected field orders: %#v %#v", first.fieldOrder, second.fieldOrder)
	}
	if &first.fieldOrder[0] != &second.fieldOrder[0] {
		t.Fatalf("expected cached pattern plan field order backing to be reused")
	}
}

func TestCollectPatternBindingsNamedStructPatternMatchesPositionalNamedStruct(t *testing.T) {
	interp := New()
	def := ast.StructDef(
		"Point",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "x"),
			ast.FieldDef(ast.Ty("i32"), "y"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	defVal := &runtime.StructDefinitionValue{Node: def}
	value := &runtime.StructInstanceValue{
		Definition: defVal,
		Positional: []runtime.Value{
			runtime.NewSmallInt(3, runtime.IntegerI32),
			runtime.NewSmallInt(5, runtime.IntegerI32),
		},
	}
	pattern := ast.StructP([]*ast.StructPatternField{
		ast.FieldP(ast.ID("left"), "x", nil),
		ast.FieldP(ast.ID("right"), "y", nil),
	}, false, "Point")

	bindings, err := interp.collectPatternBindings(pattern, value, runtime.NewEnvironment(nil), nil)
	if err != nil {
		t.Fatalf("collectPatternBindings: %v", err)
	}
	if len(bindings) != 2 {
		t.Fatalf("expected 2 bindings, got %d", len(bindings))
	}
	if !valuesEqual(bindings[0].Value, runtime.NewSmallInt(3, runtime.IntegerI32)) {
		t.Fatalf("binding 0 value = %#v, want 3", bindings[0].Value)
	}
	if !valuesEqual(bindings[1].Value, runtime.NewSmallInt(5, runtime.IntegerI32)) {
		t.Fatalf("binding 1 value = %#v, want 5", bindings[1].Value)
	}
}
