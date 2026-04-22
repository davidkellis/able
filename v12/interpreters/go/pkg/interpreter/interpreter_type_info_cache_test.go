package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestCachedSimpleTypeExpressionCachesDynamicNames(t *testing.T) {
	first := cachedSimpleTypeExpression("CustomDynamicType")
	second := cachedSimpleTypeExpression("CustomDynamicType")

	firstSimple, ok := first.(*ast.SimpleTypeExpression)
	if !ok || firstSimple == nil {
		t.Fatalf("expected simple type expression, got %#v", first)
	}
	secondSimple, ok := second.(*ast.SimpleTypeExpression)
	if !ok || secondSimple == nil {
		t.Fatalf("expected simple type expression, got %#v", second)
	}
	if firstSimple != secondSimple {
		t.Fatalf("expected cached simple type expression identity reuse")
	}
}

func TestTypeExpressionFromValueCachesStructAndHostHandleNames(t *testing.T) {
	interp := New()
	pointDef := &runtime.StructDefinitionValue{
		Node: ast.StructDef("Point", nil, ast.StructKindNamed, nil, nil, false),
	}

	pointFirst := interp.typeExpressionFromValue(pointDef)
	pointSecond := interp.typeExpressionFromValue(pointDef)
	pointFirstSimple, ok := pointFirst.(*ast.SimpleTypeExpression)
	if !ok || pointFirstSimple == nil {
		t.Fatalf("expected simple type expression for struct definition, got %#v", pointFirst)
	}
	pointSecondSimple, ok := pointSecond.(*ast.SimpleTypeExpression)
	if !ok || pointSecondSimple == nil {
		t.Fatalf("expected simple type expression for struct definition, got %#v", pointSecond)
	}
	if pointFirstSimple != pointSecondSimple {
		t.Fatalf("expected struct type expression identity reuse")
	}

	handle := &runtime.HostHandleValue{HandleType: "ProcHandle"}
	handleFirst := interp.typeExpressionFromValue(handle)
	handleSecond := interp.typeExpressionFromValue(handle)
	handleFirstSimple, ok := handleFirst.(*ast.SimpleTypeExpression)
	if !ok || handleFirstSimple == nil {
		t.Fatalf("expected simple type expression for host handle, got %#v", handleFirst)
	}
	handleSecondSimple, ok := handleSecond.(*ast.SimpleTypeExpression)
	if !ok || handleSecondSimple == nil {
		t.Fatalf("expected simple type expression for host handle, got %#v", handleSecond)
	}
	if handleFirstSimple != handleSecondSimple {
		t.Fatalf("expected host-handle type expression identity reuse")
	}
}

func TestTypeExpressionFromValueCachesArrayAndIteratorGenerics(t *testing.T) {
	interp := New()
	arr := &runtime.ArrayValue{
		Elements: []runtime.Value{runtime.NewSmallInt(1, runtime.IntegerI32)},
	}
	arrayFirst := interp.typeExpressionFromValue(arr)
	arraySecond := interp.typeExpressionFromValue(arr)

	arrayFirstGen, ok := arrayFirst.(*ast.GenericTypeExpression)
	if !ok || arrayFirstGen == nil {
		t.Fatalf("expected generic type expression for array, got %#v", arrayFirst)
	}
	arraySecondGen, ok := arraySecond.(*ast.GenericTypeExpression)
	if !ok || arraySecondGen == nil {
		t.Fatalf("expected generic type expression for array, got %#v", arraySecond)
	}
	if arrayFirstGen != arraySecondGen {
		t.Fatalf("expected array generic type expression identity reuse")
	}

	iter := runtime.NewIteratorValue(nil, nil)
	iterFirst := interp.typeExpressionFromValue(iter)
	iterSecond := interp.typeExpressionFromValue(iter)

	iterFirstGen, ok := iterFirst.(*ast.GenericTypeExpression)
	if !ok || iterFirstGen == nil {
		t.Fatalf("expected generic type expression for iterator, got %#v", iterFirst)
	}
	iterSecondGen, ok := iterSecond.(*ast.GenericTypeExpression)
	if !ok || iterSecondGen == nil {
		t.Fatalf("expected generic type expression for iterator, got %#v", iterSecond)
	}
	if iterFirstGen != iterSecondGen {
		t.Fatalf("expected iterator generic type expression identity reuse")
	}
}

func TestCanonicalTypeNamesUsesAliasBaseWithoutASTExpansion(t *testing.T) {
	interp := New()
	interp.typeAliases = map[string]*ast.TypeAliasDefinition{
		"AliasI32":   ast.NewTypeAliasDefinition(ast.ID("AliasI32"), ast.Ty("i32"), nil, nil, false),
		"AliasArray": ast.NewTypeAliasDefinition(ast.ID("AliasArray"), ast.Gen(ast.Ty("Array"), ast.Ty("i32")), nil, nil, false),
		"AliasA":     ast.NewTypeAliasDefinition(ast.ID("AliasA"), ast.Ty("AliasB"), nil, nil, false),
		"AliasB":     ast.NewTypeAliasDefinition(ast.ID("AliasB"), ast.Ty("AliasA"), nil, nil, false),
	}

	aliasI32 := interp.canonicalTypeNames("AliasI32")
	if len(aliasI32) != 2 || aliasI32[0] != "AliasI32" || aliasI32[1] != "i32" {
		t.Fatalf("unexpected canonical alias names for AliasI32: %#v", aliasI32)
	}

	aliasArray := interp.canonicalTypeNames("AliasArray")
	if len(aliasArray) != 2 || aliasArray[0] != "AliasArray" || aliasArray[1] != "Array" {
		t.Fatalf("unexpected canonical alias names for AliasArray: %#v", aliasArray)
	}

	cycle := interp.canonicalTypeNames("AliasA")
	if len(cycle) != 1 || cycle[0] != "AliasA" {
		t.Fatalf("expected cycle alias to return only original name, got %#v", cycle)
	}
}

func TestCachedTypeInfoNameAvoidsRepeatedAllocationsForCommonGenericTypes(t *testing.T) {
	interp := New()
	info := typeInfo{
		name: "Array",
		typeArgs: []ast.TypeExpression{
			ast.Ty("i32"),
		},
	}

	if got := interp.cachedTypeInfoName(info); got != "Array<i32>" {
		t.Fatalf("unexpected cached type info name: got=%q want=%q", got, "Array<i32>")
	}
	allocs := testing.AllocsPerRun(1000, func() {
		_ = interp.cachedTypeInfoName(info)
	})
	if allocs != 0 {
		t.Fatalf("expected cachedTypeInfoName hot path allocations to be zero, got %.2f", allocs)
	}
}

func TestCanonicalizeExpandedTypeExpressionReusesUnchangedNodes(t *testing.T) {
	env := runtime.NewEnvironment(nil)

	generic := ast.Gen(ast.Ty("Array"), ast.Ty("String"))
	if got := canonicalizeExpandedTypeExpression(generic, env); got != generic {
		t.Fatalf("expected generic type expression identity reuse")
	}

	nullable := ast.Nullable(ast.Ty("String"))
	if got := canonicalizeExpandedTypeExpression(nullable, env); got != nullable {
		t.Fatalf("expected nullable type expression identity reuse")
	}

	result := ast.Result(ast.Ty("String"))
	if got := canonicalizeExpandedTypeExpression(result, env); got != result {
		t.Fatalf("expected result type expression identity reuse")
	}

	union := ast.UnionT(ast.Ty("String"), ast.Ty("bool"))
	if got := canonicalizeExpandedTypeExpression(union, env); got != union {
		t.Fatalf("expected union type expression identity reuse")
	}

	fn := ast.FnType([]ast.TypeExpression{ast.Ty("String"), ast.Ty("bool")}, ast.Ty("String"))
	if got := canonicalizeExpandedTypeExpression(fn, env); got != fn {
		t.Fatalf("expected function type expression identity reuse")
	}
}

func TestCanonicalizeExpandedTypeExpressionRebuildsChangedNestedNodes(t *testing.T) {
	env := runtime.NewEnvironment(nil)
	env.Define("Alias", &runtime.StructDefinitionValue{
		Node: ast.StructDef("Target", nil, ast.StructKindNamed, nil, nil, false),
	})

	nullable := ast.Nullable(ast.Ty("Alias"))
	gotNullable := canonicalizeExpandedTypeExpression(nullable, env)
	if gotNullable == nullable {
		t.Fatalf("expected nullable type expression rebuild when inner name changes")
	}
	if gotInner, ok := gotNullable.(*ast.NullableTypeExpression); !ok || gotInner.InnerType == nullable.InnerType {
		t.Fatalf("expected canonicalized nullable inner type to change, got %#v", gotNullable)
	}

	union := ast.UnionT(ast.Ty("Alias"), ast.Ty("String"))
	gotUnion := canonicalizeExpandedTypeExpression(union, env)
	if gotUnion == union {
		t.Fatalf("expected union type expression rebuild when member changes")
	}
	if gotTyped, ok := gotUnion.(*ast.UnionTypeExpression); !ok || len(gotTyped.Members) != 2 || gotTyped.Members[0] == union.Members[0] {
		t.Fatalf("expected canonicalized union member to change, got %#v", gotUnion)
	}
}
