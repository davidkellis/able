package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestMatchExpressionWithIdentifierAndLiteral(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Match(
			ast.Int(2),
			ast.Mc(ast.LitP(ast.Int(1)), ast.Int(10)),
			ast.Mc(ast.ID("x"), ast.Bin("+", ast.ID("x"), ast.Int(5))),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.BigInt().Cmp(bigInt(7)) != 0 {
		t.Fatalf("expected integer 7, got %#v", result)
	}
}

func TestMatchExpressionStructGuard(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Point",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "x"),
				ast.FieldDef(ast.Ty("i32"), "y"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Match(
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(1), "x"),
					ast.FieldInit(ast.Int(2), "y"),
				},
				false,
				"Point",
				nil,
				nil,
			),
			ast.Mc(
				ast.StructP([]*ast.StructPatternField{
					ast.FieldP(ast.ID("a"), "x", nil),
					ast.FieldP(ast.ID("b"), "y", nil),
				}, false, "Point"),
				ast.Bin("+", ast.ID("a"), ast.ID("b")),
				ast.Bin(">", ast.ID("b"), ast.ID("a")),
			),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.BigInt().Cmp(bigInt(3)) != 0 {
		t.Fatalf("expected integer 3, got %#v", result)
	}
}

func TestMatchExpressionLiteralClauseKeepsLocalBindingsScoped(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Match(
			ast.Int(1),
			ast.Mc(
				ast.LitP(ast.Int(1)),
				ast.Block(
					ast.Assign(ast.ID("inner"), ast.Int(9)),
					ast.ID("inner"),
				),
			),
		),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.BigInt().Cmp(bigInt(9)) != 0 {
		t.Fatalf("expected integer 9, got %#v", result)
	}
	if _, err := env.Get("inner"); err == nil {
		t.Fatalf("literal match clause binding leaked into outer scope")
	}
}

func TestMatchExpressionLiteralClauseDirectAssignmentKeepsBindingScoped(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Match(
			ast.Int(1),
			ast.Mc(
				ast.LitP(ast.Int(1)),
				ast.Assign(ast.ID("inner"), ast.Int(9)),
			),
		),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.BigInt().Cmp(bigInt(9)) != 0 {
		t.Fatalf("expected integer 9, got %#v", result)
	}
	if _, err := env.Get("inner"); err == nil {
		t.Fatalf("direct assignment in literal match clause leaked into outer scope")
	}
}

func TestMatchExpressionGuardDeclarationRemainsVisibleToBody(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Match(
			ast.Int(1),
			ast.Mc(
				ast.LitP(ast.Int(1)),
				ast.ID("guard_value"),
				ast.Assign(ast.ID("guard_value"), ast.Int(7)),
			),
		),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.BigInt().Cmp(bigInt(7)) != 0 {
		t.Fatalf("expected integer 7, got %#v", result)
	}
	if _, err := env.Get("guard_value"); err == nil {
		t.Fatalf("guard-local binding leaked into outer scope")
	}
}

func TestMatchExpressionTypedIdentifierPatternBindsValue(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Match(
			ast.Int(4),
			ast.Mc(
				ast.TypedP(ast.ID("value"), ast.Ty("i32")),
				ast.Bin("+", ast.ID("value"), ast.Int(6)),
			),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.BigInt().Cmp(bigInt(10)) != 0 {
		t.Fatalf("expected integer 10, got %#v", result)
	}
}

func TestMatchPatternForClauseReusesBaseForBindinglessPureClause(t *testing.T) {
	interp := New()
	base := runtime.NewEnvironment(nil)

	clauseEnv, matched := interp.matchPatternForClause(
		ast.LitP(ast.Int(1)),
		runtime.NewSmallInt(1, runtime.IntegerI32),
		base,
		clauseLocalScopePlan(ast.LitP(ast.Int(1)), nil, ast.Int(2)),
	)
	if !matched {
		t.Fatalf("expected literal clause to match")
	}
	if clauseEnv != base {
		t.Fatalf("expected bindingless pure clause to reuse base env")
	}
}

func TestMatchPatternForClauseKeepsFreshScopeForDirectClauseDeclaration(t *testing.T) {
	interp := New()
	base := runtime.NewEnvironment(nil)

	clauseEnv, matched := interp.matchPatternForClause(
		ast.LitP(ast.Int(1)),
		runtime.NewSmallInt(1, runtime.IntegerI32),
		base,
		clauseLocalScopePlan(ast.LitP(ast.Int(1)), nil, ast.Assign(ast.ID("inner"), ast.Int(9))),
	)
	if !matched {
		t.Fatalf("expected literal clause to match")
	}
	if clauseEnv == base {
		t.Fatalf("expected direct clause declaration to keep a fresh env")
	}
}

func TestMatchPatternForClauseReusesBaseForUnusedIdentifierBinding(t *testing.T) {
	interp := New()
	base := runtime.NewEnvironment(nil)

	clauseEnv, matched := interp.matchPatternForClause(
		ast.ID("value"),
		runtime.NewSmallInt(1, runtime.IntegerI32),
		base,
		clauseLocalScopePlan(ast.ID("value"), nil, ast.Int(2)),
	)
	if !matched {
		t.Fatalf("expected identifier clause to match")
	}
	if clauseEnv != base {
		t.Fatalf("expected unused identifier binding to reuse base env")
	}
	if _, ok := base.Lookup("value"); ok {
		t.Fatalf("unused identifier binding should not materialize in base env")
	}
}

func TestMatchPatternForClauseReusesBaseForUnusedStructBindings(t *testing.T) {
	interp := New()
	base := runtime.NewEnvironment(nil)
	pointDef := ast.StructDef(
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
	pointVal := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: pointDef},
		Fields: map[string]runtime.Value{
			"x": runtime.NewSmallInt(1, runtime.IntegerI32),
			"y": runtime.NewSmallInt(2, runtime.IntegerI32),
		},
	}
	pattern := ast.StructP([]*ast.StructPatternField{
		ast.FieldP(ast.ID("left"), "x", nil),
		ast.FieldP(ast.ID("right"), "y", nil),
	}, false, "Point")

	clauseEnv, matched := interp.matchPatternForClause(
		pattern,
		pointVal,
		base,
		clauseLocalScopePlan(pattern, nil, ast.Int(3)),
	)
	if !matched {
		t.Fatalf("expected struct clause to match")
	}
	if clauseEnv != base {
		t.Fatalf("expected unused struct bindings to reuse base env")
	}
	if _, ok := base.Lookup("left"); ok {
		t.Fatalf("unused struct binding left should not materialize in base env")
	}
	if _, ok := base.Lookup("right"); ok {
		t.Fatalf("unused struct binding right should not materialize in base env")
	}
}

func TestMatchPatternForClauseCapturesStructBindingsInClauseEnv(t *testing.T) {
	interp := New()
	base := runtime.NewEnvironment(nil)
	pointDef := ast.StructDef(
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
	pointVal := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: pointDef},
		Fields: map[string]runtime.Value{
			"x": runtime.NewSmallInt(3, runtime.IntegerI32),
			"y": runtime.NewSmallInt(4, runtime.IntegerI32),
		},
	}
	pattern := ast.StructP([]*ast.StructPatternField{
		ast.FieldP(ast.ID("left"), "x", nil),
		ast.FieldP(ast.ID("right"), "y", nil),
	}, false, "Point")

	clauseEnv, matched := interp.matchPatternForClause(
		pattern,
		pointVal,
		base,
		clauseLocalScopePlan(pattern, nil, ast.ID("left")),
	)
	if !matched {
		t.Fatalf("expected struct clause to match")
	}
	if clauseEnv == nil || clauseEnv == base {
		t.Fatalf("expected struct clause to materialize a clause env")
	}
	left, ok := clauseEnv.Lookup("left")
	if !ok {
		t.Fatalf("expected left binding in clause env")
	}
	right, ok := clauseEnv.Lookup("right")
	if !ok {
		t.Fatalf("expected right binding in clause env")
	}
	assertIntValue(t, left, runtime.IntegerI32, 3)
	assertIntValue(t, right, runtime.IntegerI32, 4)
	if _, ok := base.Lookup("left"); ok {
		t.Fatalf("struct clause binding left leaked into base env")
	}
	if _, ok := base.Lookup("right"); ok {
		t.Fatalf("struct clause binding right leaked into base env")
	}
}

func TestMatchPatternForClauseCapturesArrayRestBindingsInClauseEnv(t *testing.T) {
	interp := New()
	base := runtime.NewEnvironment(nil)
	pattern := ast.ArrP(
		[]ast.Pattern{
			ast.PatternFrom("first"),
			ast.PatternFrom("second"),
		},
		ast.PatternFrom("rest"),
	)
	value := &runtime.ArrayValue{
		Elements: []runtime.Value{
			runtime.NewSmallInt(1, runtime.IntegerI32),
			runtime.NewSmallInt(2, runtime.IntegerI32),
			runtime.NewSmallInt(3, runtime.IntegerI32),
			runtime.NewSmallInt(4, runtime.IntegerI32),
		},
	}

	clauseEnv, matched := interp.matchPatternForClause(
		pattern,
		value,
		base,
		clauseLocalScopePlan(pattern, nil, ast.ID("first")),
	)
	if !matched {
		t.Fatalf("expected array clause to match")
	}
	if clauseEnv == nil || clauseEnv == base {
		t.Fatalf("expected array clause to materialize a clause env")
	}
	first, ok := clauseEnv.Lookup("first")
	if !ok {
		t.Fatalf("expected first binding in clause env")
	}
	second, ok := clauseEnv.Lookup("second")
	if !ok {
		t.Fatalf("expected second binding in clause env")
	}
	rest, ok := clauseEnv.Lookup("rest")
	if !ok {
		t.Fatalf("expected rest binding in clause env")
	}
	assertIntValue(t, first, runtime.IntegerI32, 1)
	assertIntValue(t, second, runtime.IntegerI32, 2)
	restArray, ok := rest.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected rest binding to be an array, got %#v", rest)
	}
	if len(restArray.Elements) != 2 {
		t.Fatalf("expected rest length 2, got %d", len(restArray.Elements))
	}
	assertIntValue(t, restArray.Elements[0], runtime.IntegerI32, 3)
	assertIntValue(t, restArray.Elements[1], runtime.IntegerI32, 4)
}

func TestMatchPatternForClauseSkipsUnusedIdentifierBindingWhenBodyNeedsOwnScope(t *testing.T) {
	interp := New()
	base := runtime.NewEnvironment(nil)

	clauseEnv, matched := interp.matchPatternForClause(
		ast.ID("value"),
		runtime.NewSmallInt(1, runtime.IntegerI32),
		base,
		clauseLocalScopePlan(ast.ID("value"), nil, ast.Assign(ast.ID("inner"), ast.Int(9))),
	)
	if !matched {
		t.Fatalf("expected identifier clause to match")
	}
	if clauseEnv == base {
		t.Fatalf("expected direct clause declaration to keep a fresh env")
	}
	if _, ok := clauseEnv.Lookup("value"); ok {
		t.Fatalf("unused identifier binding should stay dead when clause-local declarations force scope")
	}
}

func TestMatchTypedPatternExactPrimitiveFastPath(t *testing.T) {
	value := runtime.NewSmallInt(7, runtime.IntegerU8)
	got, ok := matchTypedPatternExactPrimitive(ast.Ty("u8"), value)
	if !ok {
		t.Fatalf("expected exact u8 typed pattern to use primitive fast path")
	}
	if got != value {
		t.Fatalf("expected exact primitive fast path to preserve value, got %#v", got)
	}
	if _, ok := matchTypedPatternExactPrimitive(ast.Ty("u8"), runtime.NewSmallInt(7, runtime.IntegerI32)); ok {
		t.Fatalf("expected non-exact integer kind to use generic typed-pattern path")
	}
	if got, ok := matchTypedPatternExactPrimitive(ast.Ty("bool"), runtime.BoolValue{Val: true}); !ok || got != (runtime.BoolValue{Val: true}) {
		t.Fatalf("expected exact bool typed pattern to use primitive fast path, got %#v ok=%v", got, ok)
	}
}

func TestMatchTypedPatternExactNamedStructFastPath(t *testing.T) {
	nodeDef := ast.StructDef(
		"Node",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "value"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	nodeVal := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: nodeDef},
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(7, runtime.IntegerI32),
		},
	}
	got, ok := matchTypedPatternExactNamedStruct(ast.Ty("Node"), nodeVal)
	if !ok {
		t.Fatalf("expected exact named struct typed pattern to use fast path")
	}
	if got != nodeVal {
		t.Fatalf("expected named struct fast path to preserve value, got %#v", got)
	}
	if _, ok := matchTypedPatternExactNamedStruct(ast.Ty("Other"), nodeVal); ok {
		t.Fatalf("expected mismatched struct name to miss fast path")
	}
}

func TestMatchTypedPatternExactSingletonStructFastPath(t *testing.T) {
	pendingDef := &runtime.StructDefinitionValue{
		Node: ast.StructDef("Pending", nil, ast.StructKindNamed, nil, nil, false),
	}
	got, ok := matchTypedPatternExactNamedStruct(ast.Ty("Pending"), pendingDef)
	if !ok {
		t.Fatalf("expected singleton struct typed pattern to use fast path")
	}
	if got != pendingDef {
		t.Fatalf("expected singleton struct fast path to preserve value, got %#v", got)
	}
}

func TestMatchTypedPatternSimpleIntegerCoercionFastPath(t *testing.T) {
	interp := New()
	value := runtime.NewSmallInt(7, runtime.IntegerI32)

	got, ok := interp.matchTypedPatternValue(ast.Ty("u64"), value)
	if !ok {
		t.Fatalf("expected typed pattern to match fitting integer via shared simple coercion fast path")
	}
	assertIntValue(t, got, runtime.IntegerU64, 7)
}

func TestMatchTypedPatternGenericSimpleTypeKeepsRawIntegerSnapshot(t *testing.T) {
	interp := New()
	source := &bytecodeRawI64SlotCell{Val: 7}

	got, ok := interp.matchTypedPatternValue(ast.Ty("T"), source)
	if !ok {
		t.Fatalf("expected generic simple typed pattern to match")
	}
	source.Val = 99
	kind, raw, ok := bytecodeRawIntegerValueInfo(got)
	if !ok || kind != runtime.IntegerI64 || raw != 7 {
		t.Fatalf("typed pattern result = %#v, want raw i64 snapshot 7", got)
	}
	if _, boxed := got.(runtime.IntegerValue); boxed {
		t.Fatalf("typed pattern result = %#v, wanted raw carrier", got)
	}
}

func TestMatchesTypeGenericSimpleTypeAvoidsRawIntegerMaterialization(t *testing.T) {
	interp := New()
	typeExpr := ast.Ty("T")
	source := &bytecodeRawI64SlotCell{Val: 7}
	if !interp.matchesType(typeExpr, source) {
		t.Fatalf("expected generic simple type to match")
	}

	allocs := testing.AllocsPerRun(100, func() {
		if !interp.matchesType(typeExpr, source) {
			t.Fatalf("expected generic simple type to match")
		}
	})
	if allocs != 0 {
		t.Fatalf("expected generic simple type raw match to allocate zero, got %.2f", allocs)
	}
}

func TestMatchesTypeRawI64ExactPrimitiveAvoidsMaterialization(t *testing.T) {
	interp := New()
	typeExpr := ast.Ty("i64")
	source := &bytecodeRawI64SlotCell{Val: 7}
	if !interp.matchesType(typeExpr, source) {
		t.Fatalf("expected raw i64 to match i64")
	}

	allocs := testing.AllocsPerRun(100, func() {
		if !interp.matchesType(typeExpr, source) {
			t.Fatalf("expected raw i64 to match i64")
		}
	})
	if allocs != 0 {
		t.Fatalf("expected raw i64 exact primitive match to allocate zero, got %.2f", allocs)
	}
}

func TestMatchesTypeRawI64KnownNonPrimitiveMissAvoidsMaterialization(t *testing.T) {
	interp := New()
	typeExpr := ast.Ty("IteratorEnd")
	source := &bytecodeRawI64SlotCell{Val: 7}
	if interp.matchesType(typeExpr, source) {
		t.Fatalf("expected raw i64 not to match IteratorEnd")
	}

	allocs := testing.AllocsPerRun(100, func() {
		if interp.matchesType(typeExpr, source) {
			t.Fatalf("expected raw i64 not to match IteratorEnd")
		}
	})
	if allocs != 0 {
		t.Fatalf("expected raw i64 known non-primitive miss to allocate zero, got %.2f", allocs)
	}
}

func TestDestructuringAssignmentArrayPattern(t *testing.T) {
	interp := New()
	patternWithRest := ast.ArrP([]ast.Pattern{ast.PatternFrom("first"), ast.PatternFrom("second")}, ast.PatternFrom("rest"))
	patternNoRest := ast.ArrP([]ast.Pattern{ast.PatternFrom("first"), ast.PatternFrom("second")}, nil)
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("arr"), ast.Arr(ast.Int(1), ast.Int(2), ast.Int(3))),
		ast.Assign(patternWithRest, ast.ID("arr")),
		ast.AssignOp(ast.AssignmentAssign, patternNoRest, ast.Arr(ast.Int(4), ast.Int(5))),
		ast.ID("rest"),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	first, err := env.Get("first")
	if err != nil {
		t.Fatalf("expected binding for first: %v", err)
	}
	firstInt, ok := first.(runtime.IntegerValue)
	if !ok || firstInt.BigInt().Cmp(bigInt(4)) != 0 {
		t.Fatalf("expected first == 4, got %#v", first)
	}
	second, err := env.Get("second")
	if err != nil {
		t.Fatalf("expected binding for second: %v", err)
	}
	secondInt, ok := second.(runtime.IntegerValue)
	if !ok || secondInt.BigInt().Cmp(bigInt(5)) != 0 {
		t.Fatalf("expected second == 5, got %#v", second)
	}
	if _, err := env.Get("rest"); err != nil {
		t.Fatalf("expected binding for rest: %v", err)
	}
	restVal, ok := result.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected rest array, got %#v", result)
	}
	if len(restVal.Elements) != 1 {
		t.Fatalf("expected rest length 1, got %d", len(restVal.Elements))
	}
	if restElem, ok := restVal.Elements[0].(runtime.IntegerValue); !ok || restElem.BigInt().Cmp(bigInt(3)) != 0 {
		t.Fatalf("expected rest element 3, got %#v", restVal.Elements[0])
	}
}

func TestAssignmentEqualsDeclaresBindingWhenMissing(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.AssignOp(ast.AssignmentAssign, ast.ID("fresh"), ast.Int(42)),
		ast.ID("fresh"),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.BigInt().Cmp(bigInt(42)) != 0 {
		t.Fatalf("expected final value 42, got %#v", result)
	}
	if _, err := env.Get("fresh"); err != nil {
		t.Fatalf("expected binding for fresh: %v", err)
	}
}

func TestAssignmentDeclareRequiresNewBinding(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.AssignOp(ast.AssignmentDeclare, ast.ID("dup"), ast.Int(1)),
		ast.AssignOp(ast.AssignmentDeclare, ast.ID("dup"), ast.Int(2)),
	}, nil, nil)

	if _, _, err := interp.EvaluateModule(module); err == nil {
		t.Fatalf("expected error for redeclaring dup in same scope")
	}
}

func TestTypedAssignmentWidenIntegerValues(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("value"), ast.Int(5)),
		ast.Assign(ast.TypedP(ast.ID("wide"), ast.Ty("i64")), ast.ID("value")),
		ast.ID("wide"),
	}, nil, nil)
	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.TypeSuffix != runtime.IntegerI64 || intVal.BigInt().Cmp(bigInt(5)) != 0 {
		t.Fatalf("expected widened i64 value, got %#v", result)
	}
}

func TestDestructuringDeclareRequiresNewBinding(t *testing.T) {
	interp := New()
	pat := ast.ArrP([]ast.Pattern{ast.PatternFrom("left"), ast.PatternFrom("right")}, nil)
	module := ast.Mod([]ast.Statement{
		ast.AssignOp(ast.AssignmentDeclare, pat, ast.Arr(ast.Int(1), ast.Int(2))),
		ast.AssignOp(ast.AssignmentDeclare, pat, ast.Arr(ast.Int(3), ast.Int(4))),
	}, nil, nil)
	if _, _, err := interp.EvaluateModule(module); err == nil {
		t.Fatalf("expected error when := pattern introduces no new bindings")
	}
}

func TestDestructuringAssignmentEqualsDeclaresBindings(t *testing.T) {
	interp := New()
	pat := ast.ArrP([]ast.Pattern{ast.PatternFrom("first"), ast.PatternFrom("second")}, nil)
	module := ast.Mod([]ast.Statement{
		ast.AssignOp(ast.AssignmentAssign, pat, ast.Arr(ast.Int(9), ast.Int(8))),
		ast.ID("second"),
	}, nil, nil)
	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.BigInt().Cmp(bigInt(8)) != 0 {
		t.Fatalf("expected result 8, got %#v", result)
	}
	if _, err := env.Get("first"); err != nil {
		t.Fatalf("expected binding for first: %v", err)
	}
}

func TestForLoopArrayPattern(t *testing.T) {
	interp := New()
	pattern := ast.ArrP([]ast.Pattern{ast.PatternFrom("x"), ast.PatternFrom("y")}, nil)
	pairs := ast.Arr(
		ast.Arr(ast.Int(1), ast.Int(2)),
		ast.Arr(ast.Int(3), ast.Int(4)),
	)
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("pairs"), pairs),
		ast.Assign(ast.ID("sum"), ast.Int(0)),
		ast.ForLoopPattern(pattern, ast.ID("pairs"), ast.Block(
			ast.AssignOp(ast.AssignmentAssign, ast.ID("sum"), ast.Bin("+", ast.ID("sum"), ast.ID("x"))),
		)),
		ast.ID("sum"),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sum, ok := result.(runtime.IntegerValue)
	if !ok || sum.BigInt().Cmp(bigInt(4)) != 0 {
		t.Fatalf("expected sum 4, got %#v", result)
	}
}

func TestForLoopStructDestructuring(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Point",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "x"),
				ast.FieldDef(ast.Ty("i32"), "y"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Assign(
			ast.ID("points"),
			ast.Arr(
				ast.StructLit(
					[]*ast.StructFieldInitializer{
						ast.FieldInit(ast.Int(1), "x"),
						ast.FieldInit(ast.Int(2), "y"),
					},
					false,
					"Point",
					nil,
					nil,
				),
				ast.StructLit(
					[]*ast.StructFieldInitializer{
						ast.FieldInit(ast.Int(3), "x"),
						ast.FieldInit(ast.Int(4), "y"),
					},
					false,
					"Point",
					nil,
					nil,
				),
			),
		),
		ast.Assign(ast.ID("sum"), ast.Int(0)),
		ast.ForLoopPattern(
			ast.StructP(
				[]*ast.StructPatternField{
					ast.FieldP(ast.PatternFrom("x"), "x", nil),
					ast.FieldP(ast.PatternFrom("y"), "y", nil),
				},
				false,
				"Point",
			),
			ast.ID("points"),
			ast.Block(
				ast.AssignOp(
					ast.AssignmentAssign,
					ast.ID("sum"),
					ast.Bin(
						"+",
						ast.ID("sum"),
						ast.Bin("+", ast.ID("x"), ast.ID("y")),
					),
				),
			),
		),
		ast.ID("sum"),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.BigInt().Cmp(bigInt(10)) != 0 {
		t.Fatalf("expected sum 10, got %#v", result)
	}
}

func TestForLoopContinueSkipsElements(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("sum"), ast.Int(0)),
		ast.ForLoopPattern(
			ast.ID("x"),
			ast.Arr(ast.Int(1), ast.Int(2), ast.Int(3)),
			ast.Block(
				ast.Iff(
					ast.Bin("==", ast.ID("x"), ast.Int(2)),
					ast.Block(ast.Cont(nil)),
				),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("sum"), ast.Bin("+", ast.ID("sum"), ast.ID("x"))),
			),
		),
		ast.ID("sum"),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.BigInt().Cmp(bigInt(4)) != 0 {
		t.Fatalf("expected 4 from continue loop, got %#v", result)
	}
}

func TestBreakpointLabeledBreak(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("sum"), ast.Int(0)),
		ast.Breakpoint(
			"exit",
			ast.Block(
				ast.ForLoopPattern(
					ast.ID("n"),
					ast.Range(ast.Int(1), ast.Int(5), true),
					ast.Block(
						ast.AssignOp(ast.AssignmentAssign, ast.ID("sum"), ast.Bin("+", ast.ID("sum"), ast.ID("n"))),
						ast.Iff(
							ast.Bin("==", ast.ID("n"), ast.Int(3)),
							ast.Block(ast.Brk("exit", ast.Str("done"))),
						),
					),
				),
				ast.Str("fallthrough"),
			),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "done" {
		t.Fatalf("expected 'done', got %#v", result)
	}
}
