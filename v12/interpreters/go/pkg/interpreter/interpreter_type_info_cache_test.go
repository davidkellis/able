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

func TestIsPrimitiveNameDoesNotAllocateForGenericNames(t *testing.T) {
	if isPrimitiveName("T") {
		t.Fatalf("expected generic type variable not to be primitive")
	}
	allocs := testing.AllocsPerRun(1000, func() {
		if isPrimitiveName("T") {
			panic("unexpected primitive match")
		}
	})
	if allocs != 0 {
		t.Fatalf("expected generic-name primitive check to allocate zero, got %.2f", allocs)
	}
}

func TestIsKnownTypeNameCachesMissAndUpdatesOnPackageRegistration(t *testing.T) {
	interp := New()
	name := "CachedLookupType"
	if interp.isKnownTypeName(name) {
		t.Fatalf("expected type name to start unknown")
	}
	if known, ok := interp.knownTypeNameCache[name]; !ok || known {
		t.Fatalf("expected cached unknown type-name lookup, got known=%v ok=%v", known, ok)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		if interp.isKnownTypeName(name) {
			panic("unexpected known type before registration")
		}
	})
	if allocs != 0 {
		t.Fatalf("expected cached unknown type-name lookup to allocate zero, got %.2f", allocs)
	}

	interp.RegisterPackageSymbol("cache.test", name, &runtime.StructDefinitionValue{
		Node: ast.StructDef(name, nil, ast.StructKindNamed, nil, nil, false),
	})
	if !interp.isKnownTypeName(name) {
		t.Fatalf("expected registered struct type name to be known")
	}
	if known, ok := interp.knownTypeNameCache[name]; !ok || !known {
		t.Fatalf("expected registered type name to update cache, got known=%v ok=%v", known, ok)
	}

	interp.RegisterPackageSymbol("cache.other", name, runtime.StringValue{Val: "not a type"})
	if !interp.isKnownTypeName(name) {
		t.Fatalf("expected non-type symbol in another package not to hide registered type")
	}

	interp.RegisterPackageSymbol("cache.test", name, runtime.StringValue{Val: "not a type"})
	if interp.isKnownTypeName(name) {
		t.Fatalf("expected replacing final type symbol with non-type to clear known status")
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

func TestExpandTypeAliasesCachedReusesAndInvalidates(t *testing.T) {
	interp := New()
	interp.RegisterTypeAlias(
		"AliasBox",
		ast.NewTypeAliasDefinition(
			ast.ID("AliasBox"),
			ast.Gen(ast.Ty("Array"), ast.Ty("T")),
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			false,
		),
	)
	expr := ast.Gen(ast.Ty("AliasBox"), ast.Ty("String"))

	first := interp.expandTypeAliasesCached(expr)
	second := interp.expandTypeAliasesCached(expr)
	if first != second {
		t.Fatalf("expected alias expansion cache to reuse expanded type expression identity")
	}
	if len(interp.typeAliasExpansionCache) == 0 {
		t.Fatalf("expected alias expansion cache to store expanded expression")
	}
	if got := typeExpressionToString(first); got != "Array<String>" {
		t.Fatalf("unexpected first alias expansion: got=%q want=%q", got, "Array<String>")
	}

	interp.RegisterTypeAlias(
		"AliasBox",
		ast.NewTypeAliasDefinition(ast.ID("AliasBox"), ast.Ty("String"), nil, nil, false),
	)
	if len(interp.typeAliasExpansionCache) != 0 {
		t.Fatalf("expected alias expansion cache to clear after alias registration")
	}
	updated := interp.expandTypeAliasesCached(expr)
	if updated == first {
		t.Fatalf("expected alias expansion after invalidation to produce a fresh result")
	}
	if got := typeExpressionToString(updated); got != "String" {
		t.Fatalf("unexpected updated alias expansion: got=%q want=%q", got, "String")
	}
}

func TestExpandTypeAliasesCachedSeedsReferenceCacheForNegativeResult(t *testing.T) {
	interp := New()
	interp.RegisterTypeAlias(
		"AliasBox",
		ast.NewTypeAliasDefinition(ast.ID("AliasBox"), ast.Ty("String"), nil, nil, false),
	)
	expr := ast.Result(ast.Ty("PlainType"))

	if got := interp.expandTypeAliasesCached(expr); got != expr {
		t.Fatalf("expected non-alias expansion to return original expression")
	}
	if cached, ok := interp.typeAliasReferenceCache[expr]; !ok || cached {
		t.Fatalf("expected expandTypeAliasesCached to seed negative alias-reference cache entry, got ok=%v cached=%v", ok, cached)
	}
}

func TestTypeExpressionReferencesAliasCachedReusesAndInvalidates(t *testing.T) {
	interp := New()
	expr := ast.Result(ast.Ty("AliasBox"))
	interp.RegisterTypeAlias(
		"OtherAlias",
		ast.NewTypeAliasDefinition(ast.ID("OtherAlias"), ast.Ty("String"), nil, nil, false),
	)

	if got := interp.typeExpressionReferencesAliasCached(expr); got {
		t.Fatalf("expected no alias reference before registration")
	}
	if len(interp.typeAliasReferenceCache) == 0 {
		t.Fatalf("expected alias reference cache to store negative result")
	}

	interp.RegisterTypeAlias(
		"AliasBox",
		ast.NewTypeAliasDefinition(ast.ID("AliasBox"), ast.Ty("String"), nil, nil, false),
	)
	if len(interp.typeAliasReferenceCache) != 0 {
		t.Fatalf("expected alias reference cache to clear after alias registration")
	}

	if got := interp.typeExpressionReferencesAliasCached(expr); !got {
		t.Fatalf("expected alias reference after registration")
	}
	if len(interp.typeAliasReferenceCache) == 0 {
		t.Fatalf("expected alias reference cache to store positive result")
	}
}

func TestMatchesType_GenericStructAndAliasRemainExact(t *testing.T) {
	interp := New()
	boxDef := ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("T"), "value"),
		},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	box := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: boxDef},
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(1, runtime.IntegerI32),
		},
		TypeArguments: []ast.TypeExpression{ast.Ty("i32")},
	}

	if !interp.matchesType(ast.Gen(ast.Ty("Box"), ast.Ty("i32")), box) {
		t.Fatalf("expected Box<i32> to match exact generic struct value")
	}
	if interp.matchesType(ast.Gen(ast.Ty("Box"), ast.Ty("String")), box) {
		t.Fatalf("expected Box<String> not to match Box<i32> value")
	}

	interp.RegisterTypeAlias(
		"IntBox",
		ast.NewTypeAliasDefinition(ast.ID("IntBox"), ast.Gen(ast.Ty("Box"), ast.Ty("i32")), nil, nil, false),
	)
	if !interp.matchesType(ast.Ty("IntBox"), box) {
		t.Fatalf("expected alias-expanded Box<i32> target to match exact generic struct value")
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

func TestCachedTypeExpressionFromInfoReusesGenericExpressions(t *testing.T) {
	interp := New()
	info := typeInfo{
		name: "Box",
		typeArgs: []ast.TypeExpression{
			cachedSimpleTypeExpression("i32"),
		},
	}

	first := interp.cachedTypeExpressionFromInfo(info)
	second := interp.cachedTypeExpressionFromInfo(info)

	firstGen, ok := first.(*ast.GenericTypeExpression)
	if !ok || firstGen == nil {
		t.Fatalf("expected generic type expression, got %#v", first)
	}
	secondGen, ok := second.(*ast.GenericTypeExpression)
	if !ok || secondGen == nil {
		t.Fatalf("expected generic type expression, got %#v", second)
	}
	if firstGen != secondGen {
		t.Fatalf("expected cached generic type expression identity reuse")
	}

	allocs := testing.AllocsPerRun(1000, func() {
		if got := interp.cachedTypeExpressionFromInfo(info); got != first {
			panic("unexpected cached type expression result")
		}
	})
	if allocs != 0 {
		t.Fatalf("expected cachedTypeExpressionFromInfo hot path allocations to be zero, got %.2f", allocs)
	}
}

func TestTypeExpressionFromValueCachesGenericStructExpressions(t *testing.T) {
	interp := New()
	typeArg := cachedSimpleTypeExpression("i32")
	inst := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{
			Node: ast.StructDef(
				"Box",
				[]*ast.StructFieldDefinition{
					ast.FieldDef(ast.Ty("T"), "value"),
				},
				ast.StructKindNamed,
				[]*ast.GenericParameter{ast.GenericParam("T")},
				nil,
				false,
			),
		},
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(1, runtime.IntegerI32),
		},
		TypeArguments: []ast.TypeExpression{typeArg},
	}

	first := interp.typeExpressionFromValue(inst)
	second := interp.typeExpressionFromValue(inst)

	firstGen, ok := first.(*ast.GenericTypeExpression)
	if !ok || firstGen == nil {
		t.Fatalf("expected generic type expression, got %#v", first)
	}
	secondGen, ok := second.(*ast.GenericTypeExpression)
	if !ok || secondGen == nil {
		t.Fatalf("expected generic type expression, got %#v", second)
	}
	if firstGen != secondGen {
		t.Fatalf("expected generic struct type expression identity reuse")
	}
	if got := typeExpressionToString(first); got != "Box<i32>" {
		t.Fatalf("unexpected generic struct type expression: got=%q want=%q", got, "Box<i32>")
	}

	allocs := testing.AllocsPerRun(1000, func() {
		if got := interp.typeExpressionFromValue(inst); got != first {
			panic("unexpected generic struct type expression result")
		}
	})
	if allocs != 0 {
		t.Fatalf("expected generic struct type-expression hot path allocations to be zero, got %.2f", allocs)
	}
}

func TestTypeExpressionFromValueHandlesRecursiveGenericStruct(t *testing.T) {
	interp := New()
	def := ast.StructDef(
		"Node",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("T"), "value"),
			ast.FieldDef(ast.Gen(ast.Ty("Node"), ast.Ty("T")), "next"),
		},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	inst := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: def},
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(1, runtime.IntegerI32),
		},
		TypeArguments: []ast.TypeExpression{ast.Ty("T")},
	}
	inst.Fields["next"] = inst

	first := interp.typeExpressionFromValue(inst)
	second := interp.typeExpressionFromValue(inst)

	if got := typeExpressionToString(first); got != "Node<i32>" {
		t.Fatalf("unexpected recursive generic struct type expression: got=%q want=%q", got, "Node<i32>")
	}
	if first != second {
		t.Fatalf("expected recursive generic struct type expression identity reuse")
	}
}

func TestStructGenericInferencePlanCacheReusesEntry(t *testing.T) {
	interp := New()
	def := ast.StructDef(
		"Record",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "count"),
			ast.FieldDef(ast.Ty("T"), "value"),
			ast.FieldDef(ast.Gen(ast.Ty("Array"), ast.Ty("U")), "items"),
		},
		ast.StructKindNamed,
		[]*ast.GenericParameter{
			ast.GenericParam("T"),
			ast.GenericParam("U"),
		},
		nil,
		false,
	)

	planA := interp.structGenericInferencePlan(def)
	planB := interp.structGenericInferencePlan(def)

	if planA == nil || planB == nil {
		t.Fatalf("expected cached struct generic inference plans")
	}
	if planA != planB {
		t.Fatalf("expected struct generic inference plan cache reuse")
	}
	if planA.expectedCount != 2 {
		t.Fatalf("expected two generic slots, got %d", planA.expectedCount)
	}
	if len(planA.fields) != 2 {
		t.Fatalf("expected only generic-relevant fields in plan, got %d", len(planA.fields))
	}
	if planA.fields[0].fieldIndex != 1 || planA.fields[0].fieldName != "value" {
		t.Fatalf("unexpected first generic-relevant field: %#v", planA.fields[0])
	}
	if planA.fields[1].fieldIndex != 2 || planA.fields[1].fieldName != "items" {
		t.Fatalf("unexpected second generic-relevant field: %#v", planA.fields[1])
	}
	if _, ok := planA.genericNames["T"]; !ok {
		t.Fatalf("expected cached generic name T")
	}
	if _, ok := planA.genericNames["U"]; !ok {
		t.Fatalf("expected cached generic name U")
	}
	if slot, ok := planA.genericIndex["T"]; !ok || slot != 0 {
		t.Fatalf("expected cached generic index T=0, got ok=%v slot=%d", ok, slot)
	}
	if slot, ok := planA.genericIndex["U"]; !ok || slot != 1 {
		t.Fatalf("expected cached generic index U=1, got ok=%v slot=%d", ok, slot)
	}
}

func TestTypeExpressionFromValueCachesConcreteThreeArgStructExpressions(t *testing.T) {
	interp := New()
	inst := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{
			Node: ast.StructDef(
				"Triple",
				[]*ast.StructFieldDefinition{
					ast.FieldDef(ast.Ty("A"), "first"),
					ast.FieldDef(ast.Ty("B"), "second"),
					ast.FieldDef(ast.Ty("C"), "third"),
					ast.FieldDef(ast.Ty("i32"), "count"),
				},
				ast.StructKindNamed,
				[]*ast.GenericParameter{
					ast.GenericParam("A"),
					ast.GenericParam("B"),
					ast.GenericParam("C"),
				},
				nil,
				false,
			),
		},
		Fields: map[string]runtime.Value{
			"first":  runtime.NewSmallInt(1, runtime.IntegerI32),
			"second": runtime.StringValue{Val: "x"},
			"third":  runtime.BoolValue{Val: true},
			"count":  runtime.NewSmallInt(3, runtime.IntegerI32),
		},
		TypeArguments: []ast.TypeExpression{ast.Ty("i32"), ast.Ty("String"), ast.Ty("bool")},
	}

	first := interp.typeExpressionFromValue(inst)
	second := interp.typeExpressionFromValue(inst)

	firstGen, ok := first.(*ast.GenericTypeExpression)
	if !ok || firstGen == nil {
		t.Fatalf("expected generic type expression, got %#v", first)
	}
	secondGen, ok := second.(*ast.GenericTypeExpression)
	if !ok || secondGen == nil {
		t.Fatalf("expected generic type expression, got %#v", second)
	}
	if firstGen != secondGen {
		t.Fatalf("expected three-arg generic struct type expression identity reuse")
	}
	if got := typeExpressionToString(first); got != "Triple<i32, String, bool>" {
		t.Fatalf("unexpected three-arg generic struct type expression: got=%q want=%q", got, "Triple<i32, String, bool>")
	}

	allocs := testing.AllocsPerRun(1000, func() {
		if got := interp.typeExpressionFromValue(inst); got != first {
			panic("unexpected three-arg generic struct type expression result")
		}
	})
	if allocs != 0 {
		t.Fatalf("expected three-arg generic struct type-expression hot path allocations to be zero, got %.2f", allocs)
	}
}

func TestInferStructTypeArgumentsWithSeenReusesCachedWildcardExpression(t *testing.T) {
	interp := New()
	def := ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("T"), "value"),
		},
		ast.StructKindNamed,
		[]*ast.GenericParameter{
			ast.GenericParam("T"),
			ast.GenericParam("U"),
		},
		nil,
		false,
	)

	typeArgs := interp.inferStructTypeArgumentsWithSeen(
		def,
		map[string]runtime.Value{"value": runtime.NewSmallInt(1, runtime.IntegerI32)},
		nil,
		nil,
	)
	if len(typeArgs) != 2 {
		t.Fatalf("expected two type arguments, got %#v", typeArgs)
	}
	if got := typeExpressionToString(typeArgs[0]); got != "i32" {
		t.Fatalf("expected inferred first type argument i32, got %q", got)
	}
	if typeArgs[1] != cachedWildcardTypeExpression {
		t.Fatalf("expected cached wildcard reuse for unbound generic, got %#v", typeArgs[1])
	}
}

func TestInferStructTypeArgumentsWithSeenReusesCachedTupleForSingleGeneric(t *testing.T) {
	interp := New()
	def := ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("T"), "value"),
		},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)

	first := interp.inferStructTypeArgumentsWithSeen(
		def,
		map[string]runtime.Value{"value": runtime.NewSmallInt(1, runtime.IntegerI32)},
		nil,
		nil,
	)
	second := interp.inferStructTypeArgumentsWithSeen(
		def,
		map[string]runtime.Value{"value": runtime.NewSmallInt(2, runtime.IntegerI32)},
		nil,
		nil,
	)
	if len(first) != 1 || len(second) != 1 {
		t.Fatalf("expected one inferred type arg in both runs, got first=%#v second=%#v", first, second)
	}
	if got := typeExpressionToString(first[0]); got != "i32" {
		t.Fatalf("expected first inferred arg i32, got %q", got)
	}
	if got := typeExpressionToString(second[0]); got != "i32" {
		t.Fatalf("expected second inferred arg i32, got %q", got)
	}
	if &first[0] != &second[0] {
		t.Fatalf("expected cached tuple reuse for single generic inference")
	}
}

func TestInferStructTypeArgumentsWithSeenReusesCachedTupleForTwoGenerics(t *testing.T) {
	interp := New()
	def := ast.StructDef(
		"Pair",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("K"), "first"),
			ast.FieldDef(ast.Ty("V"), "second"),
		},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("K"), ast.GenericParam("V")},
		nil,
		false,
	)

	first := interp.inferStructTypeArgumentsWithSeen(
		def,
		map[string]runtime.Value{
			"first":  runtime.NewSmallInt(1, runtime.IntegerI32),
			"second": runtime.StringValue{Val: "x"},
		},
		nil,
		nil,
	)
	second := interp.inferStructTypeArgumentsWithSeen(
		def,
		map[string]runtime.Value{
			"first":  runtime.NewSmallInt(2, runtime.IntegerI32),
			"second": runtime.StringValue{Val: "y"},
		},
		nil,
		nil,
	)
	if len(first) != 2 || len(second) != 2 {
		t.Fatalf("expected two inferred type args in both runs, got first=%#v second=%#v", first, second)
	}
	if got := typeExpressionToString(first[0]); got != "i32" {
		t.Fatalf("expected first tuple arg K=i32, got %q", got)
	}
	if got := typeExpressionToString(first[1]); got != "String" {
		t.Fatalf("expected first tuple arg V=String, got %q", got)
	}
	if &first[0] != &second[0] {
		t.Fatalf("expected cached tuple reuse for two-generic inference")
	}
}

func TestStructTypeArgsConcreteForDefinitionChecksTopLevelSelfGenerics(t *testing.T) {
	def := ast.StructDef(
		"Pair",
		nil,
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("K"), ast.GenericParam("V")},
		nil,
		false,
	)

	if !structTypeArgsConcreteForDefinition(def, []ast.TypeExpression{ast.Ty("i32"), ast.Ty("String")}) {
		t.Fatalf("expected concrete type args to be accepted")
	}
	if structTypeArgsConcreteForDefinition(def, []ast.TypeExpression{ast.Ty("K"), ast.Ty("String")}) {
		t.Fatalf("expected top-level K reference to require inference")
	}
	if structTypeArgsConcreteForDefinition(def, []ast.TypeExpression{ast.Ty("i32"), ast.Ty("V")}) {
		t.Fatalf("expected top-level V reference to require inference")
	}
	if structTypeArgsConcreteForDefinition(def, []ast.TypeExpression{cachedWildcardTypeExpression, ast.Ty("String")}) {
		t.Fatalf("expected wildcard type arg to require inference")
	}
	if !structTypeArgsConcreteForDefinition(def, []ast.TypeExpression{ast.Gen(ast.Ty("Array"), ast.Ty("K")), ast.Ty("String")}) {
		t.Fatalf("expected nested generic references to preserve existing concrete predicate semantics")
	}
}

func TestTypeInfoFromStructInstanceConcreteTypeArgsHotPathIsAllocationFree(t *testing.T) {
	interp := New()
	inst := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{
			Node: ast.StructDef(
				"Box",
				[]*ast.StructFieldDefinition{
					ast.FieldDef(ast.Ty("T"), "value"),
				},
				ast.StructKindNamed,
				[]*ast.GenericParameter{ast.GenericParam("T")},
				nil,
				false,
			),
		},
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(1, runtime.IntegerI32),
		},
		TypeArguments: []ast.TypeExpression{ast.Ty("i32")},
	}

	info, ok := interp.typeInfoFromStructInstance(inst)
	if !ok {
		t.Fatalf("expected type info")
	}
	if got := typeInfoToString(info); got != "Box<i32>" {
		t.Fatalf("unexpected concrete type info: got=%q want=%q", got, "Box<i32>")
	}
	if got := len(interp.structGenericInferencePlanCache); got != 0 {
		t.Fatalf("expected concrete type-info hot path not to seed struct inference plan cache, got %d entries", got)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		info, ok := interp.typeInfoFromStructInstance(inst)
		if !ok || info.name != "Box" || len(info.typeArgs) != 1 || info.typeArgs[0] != inst.TypeArguments[0] {
			panic("unexpected type info result")
		}
	})
	if allocs != 0 {
		t.Fatalf("expected concrete type-info hot path allocations to be zero, got %.2f", allocs)
	}
}

func TestTypeInfoFromStructInstanceInfersMissingArgsWithCachedWildcard(t *testing.T) {
	interp := New()
	inst := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{
			Node: ast.StructDef(
				"Box",
				[]*ast.StructFieldDefinition{
					ast.FieldDef(ast.Ty("T"), "value"),
				},
				ast.StructKindNamed,
				[]*ast.GenericParameter{
					ast.GenericParam("T"),
					ast.GenericParam("U"),
				},
				nil,
				false,
			),
		},
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(1, runtime.IntegerI32),
		},
		TypeArguments: []ast.TypeExpression{ast.Ty("T"), cachedWildcardTypeExpression},
	}

	info, ok := interp.typeInfoFromStructInstance(inst)
	if !ok {
		t.Fatalf("expected type info")
	}
	if len(info.typeArgs) != 2 {
		t.Fatalf("expected two type arguments, got %#v", info.typeArgs)
	}
	if got := typeExpressionToString(info.typeArgs[0]); got != "i32" {
		t.Fatalf("expected inferred first type argument i32, got %q", got)
	}
	if info.typeArgs[1] != cachedWildcardTypeExpression {
		t.Fatalf("expected cached wildcard reuse for inferred missing arg, got %#v", info.typeArgs[1])
	}
}

func TestTypeInfoFromStructInstanceMemoizesInferredArgsForHotPath(t *testing.T) {
	interp := New()
	inst := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{
			Node: ast.StructDef(
				"Box",
				[]*ast.StructFieldDefinition{
					ast.FieldDef(ast.Ty("T"), "value"),
				},
				ast.StructKindNamed,
				[]*ast.GenericParameter{ast.GenericParam("T")},
				nil,
				false,
			),
		},
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(1, runtime.IntegerI32),
		},
		TypeArguments: []ast.TypeExpression{ast.Ty("T")},
	}

	info, ok := interp.typeInfoFromStructInstance(inst)
	if !ok {
		t.Fatalf("expected type info")
	}
	if got := typeInfoToString(info); got != "Box<i32>" {
		t.Fatalf("unexpected inferred type info: got=%q want=%q", got, "Box<i32>")
	}
	if len(inst.TypeArguments) != 1 || typeExpressionToString(inst.TypeArguments[0]) != "i32" {
		t.Fatalf("expected struct instance type args to memoize inferred result, got %#v", inst.TypeArguments)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		info, ok := interp.typeInfoFromStructInstance(inst)
		if !ok || info.name != "Box" || len(info.typeArgs) != 1 || info.typeArgs[0] != inst.TypeArguments[0] {
			panic("unexpected type info result")
		}
	})
	if allocs != 0 {
		t.Fatalf("expected memoized inferred type-info hot path allocations to be zero, got %.2f", allocs)
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
