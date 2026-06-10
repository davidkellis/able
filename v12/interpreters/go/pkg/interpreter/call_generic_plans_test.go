package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestFunctionCallGenericPlanCacheReusesEntry(t *testing.T) {
	interp := New()
	decl := ast.Fn(
		"showT",
		[]*ast.FunctionParameter{ast.Param("value", ast.Ty("T"))},
		[]ast.Statement{ast.Ret(ast.ID("value"))},
		nil,
		[]*ast.GenericParameter{
			ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Show"))),
		},
		nil,
		false,
		false,
	)

	planA := interp.functionCallGenericPlan(decl)
	planB := interp.functionCallGenericPlan(decl)

	if planA == nil || planB == nil {
		t.Fatalf("expected cached function call generic plans")
	}
	if planA != planB {
		t.Fatalf("expected function call generic plan cache reuse")
	}
	if got := len(interp.functionCallGenericPlanCache); got != 1 {
		t.Fatalf("expected one function call generic plan cache entry, got %d", got)
	}
	if planA.expectedCount != 1 {
		t.Fatalf("expected one generic slot, got %d", planA.expectedCount)
	}
	if len(planA.namesByIndex) != 1 || planA.namesByIndex[0] != "T" {
		t.Fatalf("expected cached generic name T, got %#v", planA.namesByIndex)
	}
	if planA.functionName != "showT" || planA.callingCtx != "calling showT" {
		t.Fatalf("unexpected cached function call context: %#v", planA)
	}
	if len(planA.constraints) != 1 {
		t.Fatalf("expected one cached constraint spec, got %d", len(planA.constraints))
	}
	if len(planA.inferenceRelevantParams) != 1 {
		t.Fatalf("expected one inference-relevant param, got %d", len(planA.inferenceRelevantParams))
	}
	if planA.inferenceRelevantParams[0].argIndex != 0 {
		t.Fatalf("expected first inference-relevant arg index 0, got %d", planA.inferenceRelevantParams[0].argIndex)
	}
	if _, ok := planA.genericNames["T"]; !ok {
		t.Fatalf("expected cached generic name set to contain T")
	}
	if slot, ok := planA.genericIndex["T"]; !ok || slot != 0 {
		t.Fatalf("expected cached generic index T=0, got ok=%v slot=%d", ok, slot)
	}
}

func TestMethodSetConstraintPlanCacheReusesEntry(t *testing.T) {
	interp := New()
	methodSet := &runtime.MethodSet{
		TargetType: ast.Gen(ast.Ty("Box"), ast.Ty("T")),
		GenericParams: []*ast.GenericParameter{
			ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Show"))),
		},
	}

	planA := interp.methodSetConstraintPlan(methodSet)
	planB := interp.methodSetConstraintPlan(methodSet)

	if planA == nil || planB == nil {
		t.Fatalf("expected cached method-set constraint plans")
	}
	if planA != planB {
		t.Fatalf("expected method-set constraint plan cache reuse")
	}
	if got := len(interp.methodSetConstraintPlanCache); got != 1 {
		t.Fatalf("expected one method-set constraint plan cache entry, got %d", got)
	}
	if _, ok := planA.genericNames["T"]; !ok {
		t.Fatalf("expected cached method-set generic name T")
	}
	if len(planA.constraints) != 1 {
		t.Fatalf("expected one cached method-set constraint spec, got %d", len(planA.constraints))
	}
}

func TestEnforceGenericConstraintsIfAnyHotPathAvoidsAllocations(t *testing.T) {
	interp := New()
	mustEvalModule(t, interp, mustParseModuleSource(t, `
interface Show {
  fn show(self: Self) -> String
}

impl Show for i32 {
  fn show(self: Self) -> String { "i32" }
}
`))

	decl := ast.Fn(
		"showT",
		[]*ast.FunctionParameter{ast.Param("value", ast.Ty("T"))},
		[]ast.Statement{ast.Ret(ast.ID("value"))},
		ast.Ty("T"),
		[]*ast.GenericParameter{
			ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Show"))),
		},
		nil,
		false,
		false,
	)
	call := ast.CallT(ast.ID("showT"), []ast.TypeExpression{ast.Ty("i32")}, ast.Int(1))

	if err := interp.enforceGenericConstraintsIfAny(decl, call); err != nil {
		t.Fatalf("first constraint check failed: %v", err)
	}
	if err := interp.enforceGenericConstraintsIfAny(decl, call); err != nil {
		t.Fatalf("second constraint check failed: %v", err)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		if err := interp.enforceGenericConstraintsIfAny(decl, call); err != nil {
			panic(err)
		}
	})
	if allocs != 0 {
		t.Fatalf("expected hot constrained generic call check to avoid allocations, got %.2f", allocs)
	}
}

func TestExplicitCallTypeBindingCacheReusesAndInvalidates(t *testing.T) {
	interp := New()
	decl := ast.Fn(
		"showT",
		[]*ast.FunctionParameter{ast.Param("value", nil)},
		[]ast.Statement{ast.Ret(ast.ID("T_type"))},
		ast.Ty("String"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	call := ast.CallT(
		ast.ID("showT"),
		[]ast.TypeExpression{ast.Gen(ast.Ty("Array"), ast.Ty("i32"))},
		ast.Arr(ast.Int(1)),
	)

	envA := runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 4)
	interp.bindTypeArgumentsIfAny(decl, call, envA)
	if got := len(interp.explicitCallTypeBindingCache); got != 1 {
		t.Fatalf("expected one explicit call type binding cache entry after first bind, got %d", got)
	}
	if value, ok := envA.Lookup("T_type"); !ok {
		t.Fatalf("expected T_type binding in first env")
	} else if str, ok := value.(runtime.StringValue); !ok || str.Val != "Array<i32>" {
		t.Fatalf("expected T_type=Array<i32>, got %T (%#v)", value, value)
	}

	envB := runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 4)
	interp.bindTypeArgumentsIfAny(decl, call, envB)
	if got := len(interp.explicitCallTypeBindingCache); got != 1 {
		t.Fatalf("expected explicit call type binding cache reuse, got %d entries", got)
	}
	if value, ok := envB.Lookup("T"); !ok {
		t.Fatalf("expected cached T binding in second env")
	} else if ref, ok := value.(runtime.TypeRefValue); !ok || ref.TypeName != "Array" {
		t.Fatalf("expected cached T=Array<i32> type ref, got %T (%#v)", value, value)
	}

	interp.invalidateMethodCache()
	if got := len(interp.explicitCallTypeBindingCache); got != 0 {
		t.Fatalf("expected invalidateMethodCache to clear explicit call type binding cache, got %d entries", got)
	}

	interp.bindTypeArgumentsIfAny(decl, call, runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 4))
	if got := len(interp.explicitCallTypeBindingCache); got != 1 {
		t.Fatalf("expected explicit call type binding cache to repopulate after invalidation, got %d entries", got)
	}

	interp.RegisterTypeAlias("MyInt", ast.NewTypeAliasDefinition(ast.NewIdentifier("MyInt"), ast.Ty("i32"), nil, nil, false))
	if got := len(interp.explicitCallTypeBindingCache); got != 0 {
		t.Fatalf("expected RegisterTypeAlias to clear explicit call type binding cache, got %d entries", got)
	}
}

func TestExplicitCallTypeBindingCacheTracksInferredCallTypeArgumentVersion(t *testing.T) {
	interp := New()
	decl := ast.Fn(
		"type_name",
		[]*ast.FunctionParameter{ast.Param("value", ast.Ty("T"))},
		[]ast.Statement{ast.Ret(ast.ID("T_type"))},
		ast.Ty("String"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	call := ast.NewFunctionCall(ast.ID("type_name"), []ast.Expression{ast.ID("value")}, nil, false)

	if err := interp.populateCallTypeArguments(decl, call, []runtime.Value{runtime.NewSmallInt(1, runtime.IntegerI32)}); err != nil {
		t.Fatalf("populate i32 call type arguments: %v", err)
	}
	envA := runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 4)
	interp.bindTypeArgumentsIfAny(decl, call, envA)
	if value, ok := envA.Lookup("T_type"); !ok {
		t.Fatalf("expected T_type binding in first env")
	} else if str, ok := value.(runtime.StringValue); !ok || str.Val != "i32" {
		t.Fatalf("expected first T_type=i32, got %T (%#v)", value, value)
	}

	if err := interp.populateCallTypeArguments(decl, call, []runtime.Value{runtime.StringValue{Val: "hello"}}); err != nil {
		t.Fatalf("populate String call type arguments: %v", err)
	}
	envB := runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 4)
	interp.bindTypeArgumentsIfAny(decl, call, envB)
	if value, ok := envB.Lookup("T_type"); !ok {
		t.Fatalf("expected T_type binding in second env")
	} else if str, ok := value.(runtime.StringValue); !ok || str.Val != "String" {
		t.Fatalf("expected second T_type=String, got %T (%#v)", value, value)
	}
	if got := len(interp.explicitCallTypeBindingCache); got != 2 {
		t.Fatalf("expected inferred call type binding cache to track both versions, got %d entries", got)
	}
}

func TestMethodSetConstraintResultCacheReusesAndInvalidates(t *testing.T) {
	interp := New()
	mustEvalModule(t, interp, mustParseModuleSource(t, `
interface Show {
  fn show(self: Self) -> String
}

impl Show for i32 {
  fn show(self: Self) -> String { "i32" }
}
`))

	boxDef := ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("T"), "value")},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	boxValue := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: boxDef},
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(7, runtime.IntegerI32),
		},
		TypeArguments: []ast.TypeExpression{ast.Ty("i32")},
	}
	fn := &runtime.FunctionValue{
		Declaration: ast.Fn(
			"box_show",
			[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Box"), ast.Ty("T")))},
			[]ast.Statement{ast.ID("self")},
			nil,
			nil,
			nil,
			false,
			false,
		),
		MethodSet: &runtime.MethodSet{
			TargetType: ast.Gen(ast.Ty("Box"), ast.Ty("T")),
			GenericParams: []*ast.GenericParameter{
				ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Show"))),
			},
		},
		Closure: interp.GlobalEnvironment(),
	}

	if err := interp.enforceMethodSetConstraints(fn, boxValue); err != nil {
		t.Fatalf("first method-set constraint check failed: %v", err)
	}
	if got := len(interp.methodSetConstraintResultCache); got != 1 {
		t.Fatalf("expected one cached method-set constraint result after first check, got %d", got)
	}
	if err := interp.enforceMethodSetConstraints(fn, boxValue); err != nil {
		t.Fatalf("second method-set constraint check failed: %v", err)
	}
	if got := len(interp.methodSetConstraintResultCache); got != 1 {
		t.Fatalf("expected method-set constraint cache reuse, got %d entries", got)
	}

	interp.invalidateMethodCache()
	if got := len(interp.methodSetConstraintResultCache); got != 0 {
		t.Fatalf("expected invalidateMethodCache to clear method-set constraint result cache, got %d entries", got)
	}

	if err := interp.enforceMethodSetConstraints(fn, boxValue); err != nil {
		t.Fatalf("method-set constraint check after invalidation failed: %v", err)
	}
	if got := len(interp.methodSetConstraintResultCache); got != 1 {
		t.Fatalf("expected method-set constraint result cache to repopulate after invalidation, got %d entries", got)
	}

	interp.RegisterTypeAlias("MyInt", ast.NewTypeAliasDefinition(ast.NewIdentifier("MyInt"), ast.Ty("i32"), nil, nil, false))
	if got := len(interp.methodSetConstraintResultCache); got != 0 {
		t.Fatalf("expected RegisterTypeAlias to clear method-set constraint result cache, got %d entries", got)
	}
}

func TestFunctionCallConstraintResultCacheReusesAndInvalidates(t *testing.T) {
	interp := New()
	mustEvalModule(t, interp, mustParseModuleSource(t, `
interface Show {
  fn show(self: Self) -> String
}

impl Show for i32 {
  fn show(self: Self) -> String { "i32" }
}
`))

	decl := ast.Fn(
		"showT",
		[]*ast.FunctionParameter{ast.Param("value", ast.Ty("T"))},
		[]ast.Statement{ast.Ret(ast.ID("value"))},
		ast.Ty("T"),
		[]*ast.GenericParameter{
			ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Show"))),
		},
		nil,
		false,
		false,
	)
	call := ast.CallT(ast.ID("showT"), []ast.TypeExpression{ast.Ty("i32")}, ast.Int(1))

	if err := interp.enforceGenericConstraintsIfAny(decl, call); err != nil {
		t.Fatalf("first function-call constraint check failed: %v", err)
	}
	if got := len(interp.functionCallConstraintResultCache); got != 1 {
		t.Fatalf("expected one cached function-call constraint result after first check, got %d", got)
	}
	if err := interp.enforceGenericConstraintsIfAny(decl, call); err != nil {
		t.Fatalf("second function-call constraint check failed: %v", err)
	}
	if got := len(interp.functionCallConstraintResultCache); got != 1 {
		t.Fatalf("expected function-call constraint cache reuse, got %d entries", got)
	}

	interp.invalidateMethodCache()
	if got := len(interp.functionCallConstraintResultCache); got != 0 {
		t.Fatalf("expected invalidateMethodCache to clear function-call constraint result cache, got %d entries", got)
	}

	if err := interp.enforceGenericConstraintsIfAny(decl, call); err != nil {
		t.Fatalf("function-call constraint check after invalidation failed: %v", err)
	}
	if got := len(interp.functionCallConstraintResultCache); got != 1 {
		t.Fatalf("expected function-call constraint result cache to repopulate after invalidation, got %d entries", got)
	}

	interp.RegisterTypeAlias("MyInt", ast.NewTypeAliasDefinition(ast.NewIdentifier("MyInt"), ast.Ty("i32"), nil, nil, false))
	if got := len(interp.functionCallConstraintResultCache); got != 0 {
		t.Fatalf("expected RegisterTypeAlias to clear function-call constraint result cache, got %d entries", got)
	}
}

func TestFunctionCallConstraintResultCacheTracksInferredCallTypeArgumentVersion(t *testing.T) {
	interp := New()
	mustEvalModule(t, interp, mustParseModuleSource(t, `
interface Show {
  fn show(self: Self) -> String
}

impl Show for i32 {
  fn show(self: Self) -> String { "i32" }
}

impl Show for String {
  fn show(self: Self) -> String { "String" }
}
`))

	decl := ast.Fn(
		"showT",
		[]*ast.FunctionParameter{ast.Param("value", ast.Ty("T"))},
		[]ast.Statement{ast.Ret(ast.ID("value"))},
		ast.Ty("T"),
		[]*ast.GenericParameter{
			ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Show"))),
		},
		nil,
		false,
		false,
	)
	call := ast.NewFunctionCall(ast.ID("showT"), []ast.Expression{ast.ID("value")}, nil, false)

	if err := interp.populateCallTypeArguments(decl, call, []runtime.Value{runtime.NewSmallInt(1, runtime.IntegerI32)}); err != nil {
		t.Fatalf("populate i32 call type arguments: %v", err)
	}
	if err := interp.enforceGenericConstraintsIfAny(decl, call); err != nil {
		t.Fatalf("i32 constrained generic call failed: %v", err)
	}

	if err := interp.populateCallTypeArguments(decl, call, []runtime.Value{runtime.StringValue{Val: "hello"}}); err != nil {
		t.Fatalf("populate String call type arguments: %v", err)
	}
	if err := interp.enforceGenericConstraintsIfAny(decl, call); err != nil {
		t.Fatalf("String constrained generic call failed: %v", err)
	}
	if got := len(interp.functionCallConstraintResultCache); got != 2 {
		t.Fatalf("expected inferred function-call constraint cache to track both versions, got %d entries", got)
	}
}

func TestInferredCallTypeArgumentCacheIgnoresNonGenericParams(t *testing.T) {
	interp := New()
	decl := ast.Fn(
		"pick",
		[]*ast.FunctionParameter{
			ast.Param("value", ast.Ty("T")),
			ast.Param("other", nil),
		},
		nil,
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	call := ast.NewFunctionCall(ast.ID("pick"), []ast.Expression{ast.ID("value"), ast.ID("other")}, nil, false)

	if err := interp.populateCallTypeArguments(decl, call, []runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
		runtime.StringValue{Val: "hello"},
	}); err != nil {
		t.Fatalf("populate first inferred call type arguments: %v", err)
	}
	if got := typeExpressionToString(call.TypeArguments[0]); got != "i32" {
		t.Fatalf("first inferred type arg = %s, want i32", got)
	}

	if err := interp.populateCallTypeArguments(decl, call, []runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
		runtime.BoolValue{Val: true},
	}); err != nil {
		t.Fatalf("populate second inferred call type arguments: %v", err)
	}
	if got := typeExpressionToString(call.TypeArguments[0]); got != "i32" {
		t.Fatalf("second inferred type arg = %s, want i32", got)
	}
	if got := interp.inferredCallTypeArgumentRuntimeCacheEntryCount(); got != 1 {
		t.Fatalf("expected one inferred runtime type-argument cache entry, got %d", got)
	}
	if got := len(interp.inferredCallTypeArgumentCache); got != 0 {
		t.Fatalf("expected no exact inferred call type-argument cache entries on shallow runtime-key path, got %d", got)
	}
}

func TestPopulateCallTypeArguments_UsesTwoArgRuntimeCacheHotPath(t *testing.T) {
	interp := New()
	decl := ast.Fn(
		"pair",
		[]*ast.FunctionParameter{
			ast.Param("left", ast.Ty("T")),
			ast.Param("right", ast.Ty("U")),
		},
		nil,
		ast.Ty("T"),
		[]*ast.GenericParameter{
			ast.GenericParam("T"),
			ast.GenericParam("U"),
		},
		nil,
		false,
		false,
	)
	call := ast.NewFunctionCall(ast.ID("pair"), []ast.Expression{ast.ID("left"), ast.ID("right")}, nil, false)
	args := []runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
		runtime.StringValue{Val: "hello"},
	}

	if err := interp.populateCallTypeArguments(decl, call, args); err != nil {
		t.Fatalf("populate two-arg inferred call type arguments: %v", err)
	}
	if got := typeExpressionToString(call.TypeArguments[0]); got != "i32" {
		t.Fatalf("first inferred type arg = %s, want i32", got)
	}
	if got := typeExpressionToString(call.TypeArguments[1]); got != "String" {
		t.Fatalf("second inferred type arg = %s, want String", got)
	}
	if got := interp.inferredCallTypeArgumentRuntimeCacheEntryCount(); got != 1 {
		t.Fatalf("expected one inferred runtime type-argument cache entry, got %d", got)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		if err := interp.populateCallTypeArguments(decl, call, args); err != nil {
			panic(err)
		}
	})
	if allocs != 0 {
		t.Fatalf("expected two-arg inferred call-type hot path allocations to be zero, got %.2f", allocs)
	}
}

func TestInferCallTypeArgumentsFromResolvedActualTypes_IndexedHotPathAvoidsAllocations(t *testing.T) {
	interp := New()
	decl := ast.Fn(
		"triple_id",
		[]*ast.FunctionParameter{
			ast.Param("value", ast.Gen(ast.Ty("Triple"), ast.Ty("A"), ast.Ty("B"), ast.Ty("C"))),
		},
		nil,
		ast.Ty("A"),
		[]*ast.GenericParameter{
			ast.GenericParam("A"),
			ast.GenericParam("B"),
			ast.GenericParam("C"),
		},
		nil,
		false,
		false,
	)
	plan := interp.functionCallGenericPlan(decl)
	actualType := interp.cachedTypeExpressionFromInfo(typeInfo{
		name: "Triple",
		typeArgs: interp.cachedTypeExpressionTuple3(
			ast.Ty("i32"),
			ast.Ty("String"),
			ast.Ty("bool"),
		),
	})
	actualTypes := interp.cachedTypeExpressionTuple1(actualType)

	got := interp.inferCallTypeArgumentsFromResolvedActualTypes(plan, actualTypes)
	if len(got) != 3 {
		t.Fatalf("expected three inferred type args, got %#v", got)
	}
	if typeExpressionToString(got[0]) != "i32" || typeExpressionToString(got[1]) != "String" || typeExpressionToString(got[2]) != "bool" {
		t.Fatalf("unexpected inferred type args: %#v", got)
	}

	actualGen, ok := actualType.(*ast.GenericTypeExpression)
	if !ok || actualGen == nil || len(actualGen.Arguments) != 3 {
		t.Fatalf("expected cached generic actual type expression, got %#v", actualType)
	}
	allocs := testing.AllocsPerRun(1000, func() {
		got := interp.inferCallTypeArgumentsFromResolvedActualTypes(plan, actualTypes)
		if len(got) != 3 || got[0] != actualGen.Arguments[0] || got[1] != actualGen.Arguments[1] || got[2] != actualGen.Arguments[2] {
			panic("unexpected inferred indexed type args result")
		}
	})
	if allocs != 0 {
		t.Fatalf("expected indexed inferred call-type hot path allocations to be zero, got %.2f", allocs)
	}
}

func TestPopulateCallTypeArguments_HotRuntimeKeyCacheAvoidsAllocationsForResolvedGenericStructArg(t *testing.T) {
	interp := New()
	decl := ast.Fn(
		"id",
		[]*ast.FunctionParameter{ast.Param("value", ast.Ty("T"))},
		nil,
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	box := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{
			Node: ast.StructDef(
				"Box",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("T"), "value")},
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
	call := ast.NewFunctionCall(ast.ID("id"), []ast.Expression{ast.ID("value")}, nil, false)
	args := []runtime.Value{box}

	if err := interp.populateCallTypeArguments(decl, call, args); err != nil {
		t.Fatalf("populate resolved generic struct call type arguments: %v", err)
	}
	if got := typeExpressionToString(call.TypeArguments[0]); got != "Box<i32>" {
		t.Fatalf("inferred type arg = %s, want Box<i32>", got)
	}
	if got := interp.inferredCallTypeArgumentRuntimeCacheEntryCount(); got != 1 {
		t.Fatalf("expected one inferred runtime type-argument cache entry, got %d", got)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		if err := interp.populateCallTypeArguments(decl, call, args); err != nil {
			panic(err)
		}
	})
	if allocs != 0 {
		t.Fatalf("expected resolved generic-struct inferred call-type hot path allocations to be zero, got %.2f", allocs)
	}
}

func TestPopulateCallTypeArguments_FallsBackToExactCacheForThreeArgGenericStruct(t *testing.T) {
	interp := New()
	decl := ast.Fn(
		"id",
		[]*ast.FunctionParameter{ast.Param("value", ast.Ty("T"))},
		nil,
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	tripleDef := ast.StructDef(
		"Triple",
		[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("A"), "first")},
		ast.StructKindNamed,
		[]*ast.GenericParameter{
			ast.GenericParam("A"),
			ast.GenericParam("B"),
			ast.GenericParam("C"),
		},
		nil,
		false,
	)
	triple := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: tripleDef},
		Fields: map[string]runtime.Value{
			"first": runtime.NewSmallInt(1, runtime.IntegerI32),
		},
		TypeArguments: []ast.TypeExpression{ast.Ty("i32"), ast.Ty("String"), ast.Ty("bool")},
	}
	call := ast.NewFunctionCall(ast.ID("id"), []ast.Expression{ast.ID("value")}, nil, false)

	if err := interp.populateCallTypeArguments(decl, call, []runtime.Value{triple}); err != nil {
		t.Fatalf("populate three-arg generic struct call type arguments: %v", err)
	}
	if got := typeExpressionToString(call.TypeArguments[0]); got != "Triple<i32, String, bool>" {
		t.Fatalf("inferred type arg = %s, want Triple<i32, String, bool>", got)
	}
	if got := interp.inferredCallTypeArgumentRuntimeCacheEntryCount(); got != 0 {
		t.Fatalf("expected unsupported three-arg generic struct to skip runtime key cache, got %d entries", got)
	}
	if got := len(interp.inferredCallTypeArgumentCache); got != 1 {
		t.Fatalf("expected one exact inferred call type-argument cache entry, got %d", got)
	}
}

func TestPopulateCallTypeArguments_ExactCacheHotPathAvoidsAllocationsForThreeArgGenericStruct(t *testing.T) {
	interp := New()
	decl := ast.Fn(
		"id",
		[]*ast.FunctionParameter{ast.Param("value", ast.Ty("T"))},
		nil,
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	tripleDef := ast.StructDef(
		"Triple",
		[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("A"), "first")},
		ast.StructKindNamed,
		[]*ast.GenericParameter{
			ast.GenericParam("A"),
			ast.GenericParam("B"),
			ast.GenericParam("C"),
		},
		nil,
		false,
	)
	triple := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: tripleDef},
		Fields: map[string]runtime.Value{
			"first": runtime.NewSmallInt(1, runtime.IntegerI32),
		},
		TypeArguments: []ast.TypeExpression{ast.Ty("i32"), ast.Ty("String"), ast.Ty("bool")},
	}
	call := ast.NewFunctionCall(ast.ID("id"), []ast.Expression{ast.ID("value")}, nil, false)
	args := []runtime.Value{triple}

	if err := interp.populateCallTypeArguments(decl, call, args); err != nil {
		t.Fatalf("populate three-arg generic struct call type arguments: %v", err)
	}
	if got := typeExpressionToString(call.TypeArguments[0]); got != "Triple<i32, String, bool>" {
		t.Fatalf("inferred type arg = %s, want Triple<i32, String, bool>", got)
	}
	if got := interp.inferredCallTypeArgumentRuntimeCacheEntryCount(); got != 0 {
		t.Fatalf("expected unsupported three-arg generic struct to skip runtime key cache, got %d entries", got)
	}
	if got := len(interp.inferredCallTypeArgumentCache); got != 1 {
		t.Fatalf("expected one exact inferred call type-argument cache entry, got %d", got)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		if err := interp.populateCallTypeArguments(decl, call, args); err != nil {
			panic(err)
		}
	})
	if allocs != 0 {
		t.Fatalf("expected three-arg exact inferred call-type hot path allocations to be zero, got %.2f", allocs)
	}
}
