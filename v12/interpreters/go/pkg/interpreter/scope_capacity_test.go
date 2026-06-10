package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestBlockLocalBindingCapacityCountsImmediateDeclarations(t *testing.T) {
	block := ast.Block(
		ast.Assign(ast.ID("x"), ast.Int(1)),
		ast.Assign(ast.TypedP(ast.ID("y"), ast.Ty("i32")), ast.Int(2)),
		ast.Fn("helper", nil, []ast.Statement{ast.Int(0)}, ast.Ty("i32"), nil, nil, false, false),
		ast.StructDef("Point", nil, ast.StructKindNamed, nil, nil, false),
		ast.Bin("+", ast.ID("x"), ast.ID("y")),
	)

	if got := blockLocalBindingCapacity(block); got != 4 {
		t.Fatalf("blockLocalBindingCapacity(...) = %d, want 4", got)
	}
}

func TestAssignmentTargetBindingCapacityCountsNestedPatterns(t *testing.T) {
	target := ast.NewStructPattern(
		[]*ast.StructPatternField{
			ast.NewStructPatternField(ast.ID("x"), ast.ID("left"), nil, nil),
			ast.NewStructPatternField(ast.NewTypedPattern(ast.ID("y"), ast.Ty("i32")), ast.ID("right"), nil, nil),
			ast.NewStructPatternField(ast.NewWildcardPattern(), ast.ID("skip"), ast.ID("alias"), nil),
		},
		false,
		ast.ID("Pair"),
	)

	if got := assignmentTargetBindingCapacity(target); got != 3 {
		t.Fatalf("assignmentTargetBindingCapacity(...) = %d, want 3", got)
	}
}

func TestFunctionLocalBindingCapacityIncludesParamsLocalsAndGenerics(t *testing.T) {
	fn := ast.Fn(
		"map_one",
		[]*ast.FunctionParameter{
			ast.Param("x", ast.Ty("T")),
			ast.NewFunctionParameter(ast.NewTypedPattern(ast.ID("y"), ast.Ty("i32")), ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Assign(ast.ID("sum"), ast.Bin("+", ast.ID("x"), ast.ID("y"))),
			ast.Ret(ast.ID("sum")),
		},
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	call := ast.CallExpr(ast.ID("map_one"), ast.Int(1), ast.Int(2))
	call.TypeArguments = []ast.TypeExpression{ast.Ty("i32")}

	if got := functionLocalBindingCapacity(fn, call); got != 5 {
		t.Fatalf("functionLocalBindingCapacity(...) = %d, want 5", got)
	}
}

func TestFunctionLocalBindingCapacityForLayoutSkipsParamBindingsForSlotBackedCalls(t *testing.T) {
	fn := ast.Fn(
		"map_one",
		[]*ast.FunctionParameter{
			ast.Param("x", ast.Ty("T")),
			ast.Param("y", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Assign(ast.ID("sum"), ast.Bin("+", ast.ID("x"), ast.ID("y"))),
			ast.Ret(ast.ID("sum")),
		},
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	call := ast.CallExpr(ast.ID("map_one"), ast.Int(1), ast.Int(2))
	call.TypeArguments = []ast.TypeExpression{ast.Ty("i32")}

	if got := functionLocalBindingCapacityForLayout(fn, call, &bytecodeFrameLayout{}); got != 3 {
		t.Fatalf("functionLocalBindingCapacityForLayout(...) = %d, want 3", got)
	}
}

func TestExpressionNeedsCurrentScopeBindingIgnoresNestedBlockDeclarations(t *testing.T) {
	expr := ast.IfExpr(
		ast.Bool(true),
		ast.Block(ast.Assign(ast.ID("inner"), ast.Int(1))),
	)
	expr.ElseBody = ast.Block(ast.Int(0))

	if expressionNeedsCurrentScopeBinding(expr) {
		t.Fatalf("expected nested block declaration to stay out of current-scope binding analysis")
	}
}

func TestClauseNeedsLocalScopeTracksPatternGuardAndBodyBindings(t *testing.T) {
	if !clauseNeedsLocalScope(ast.ID("value"), nil, ast.ID("value")) {
		t.Fatalf("expected identifier pattern binding referenced by the clause to require local scope")
	}
	if !clauseNeedsLocalScope(ast.LitP(ast.Int(1)), ast.Assign(ast.ID("guard_value"), ast.Int(7)), ast.ID("guard_value")) {
		t.Fatalf("expected guard declaration to require local scope")
	}
	if !clauseNeedsLocalScope(ast.LitP(ast.Int(1)), nil, ast.Assign(ast.ID("inner"), ast.Int(9))) {
		t.Fatalf("expected direct body declaration to require local scope")
	}
	if clauseNeedsLocalScope(ast.LitP(ast.Int(1)), nil, ast.Block(ast.Assign(ast.ID("inner"), ast.Int(9)))) {
		t.Fatalf("expected nested block declaration to keep its own scope without forcing clause scope")
	}
	if clauseNeedsLocalScope(ast.ID("value"), nil, ast.Int(2)) {
		t.Fatalf("expected unused identifier pattern binding to avoid clause-local scope")
	}
	if clauseNeedsLocalScope(
		ast.StructP([]*ast.StructPatternField{
			ast.FieldP(ast.ID("left"), "x", nil),
			ast.FieldP(ast.ID("right"), "y", nil),
		}, false, "Point"),
		nil,
		ast.Int(3),
	) {
		t.Fatalf("expected unused struct-pattern bindings to avoid clause-local scope")
	}

	plan := clauseLocalScopePlan(ast.ID("value"), nil, ast.Assign(ast.ID("inner"), ast.Int(9)))
	if !plan.needsLocalScope {
		t.Fatalf("expected direct body declaration to require clause-local scope")
	}
	if plan.capturePatternBinding {
		t.Fatalf("expected unused identifier pattern binding to stay dead even when clause-local declarations require scope")
	}
	if plan.localBindingCapacity != 1 {
		t.Fatalf("clauseLocalScopePlan(...).localBindingCapacity = %d, want 1", plan.localBindingCapacity)
	}

	plan = clauseLocalScopePlan(
		ast.StructP([]*ast.StructPatternField{
			ast.FieldP(ast.ID("left"), "x", nil),
			ast.FieldP(ast.ID("right"), "y", nil),
		}, false, "Point"),
		ast.Assign(ast.ID("guard_value"), ast.Int(7)),
		ast.Bin("+", ast.ID("left"), ast.Int(1)),
	)
	if !plan.needsLocalScope || !plan.capturePatternBinding {
		t.Fatalf("expected referenced struct-pattern bindings to stay captured")
	}
	if plan.localBindingCapacity != 3 {
		t.Fatalf("clauseLocalScopePlan(...).localBindingCapacity = %d, want 3", plan.localBindingCapacity)
	}
	if plan.patternBindingCount != 2 {
		t.Fatalf("clauseLocalScopePlan(...).patternBindingCount = %d, want 2", plan.patternBindingCount)
	}
	if !clauseLocalScopePlan(ast.ID("value"), nil, ast.ID("value")).transientSingleBindOK {
		t.Fatalf("expected simple identifier binding clause to allow transient single-binding reuse")
	}
	if !clauseLocalScopePlan(ast.TypedP(ast.ID("value"), ast.Ty("i32")), nil, ast.ID("value")).transientSingleBindOK {
		t.Fatalf("expected typed identifier binding clause to allow transient single-binding reuse")
	}
	parserShorthand := ast.NewStructLiteral(
		[]*ast.StructFieldInitializer{
			ast.NewStructFieldInitializer(nil, ast.ID("path"), true),
		},
		false,
		ast.ID("TempDir"),
		nil,
		nil,
	)
	shorthandPlan := clauseLocalScopePlan(ast.TypedP(ast.ID("path"), ast.Ty("String")), nil, parserShorthand)
	if !shorthandPlan.capturePatternBinding || !shorthandPlan.transientSingleBindOK {
		t.Fatalf("expected struct literal shorthand to reference typed pattern binding: %#v", shorthandPlan)
	}
	if !clauseLocalScopePlan(
		ast.ID("value"),
		nil,
		ast.OrElse(ast.Prop(ast.Block(ast.Raise(ast.Str("x")))), nil, ast.ID("value")),
	).transientSingleBindOK {
		t.Fatalf("expected non-escaping or-else body to allow transient single-binding reuse")
	}
	if !clauseLocalScopePlan(
		ast.StructP([]*ast.StructPatternField{
			ast.FieldP(ast.ID("left"), "x", nil),
			ast.FieldP(ast.ID("right"), "y", nil),
		}, false, "Point"),
		nil,
		ast.Rescue(
			ast.Block(ast.Raise(ast.Str("boom"))),
			ast.Mc(ast.Wc(), ast.Block(ast.ID("left"))),
		),
	).transientBindingSetOK {
		t.Fatalf("expected non-escaping rescue body to allow transient multi-binding reuse")
	}
	noBindingPlan := clauseLocalScopePlan(
		ast.Wc(),
		nil,
		ast.Assign(ast.ID("local"), ast.Int(1)),
	)
	if !noBindingPlan.needsLocalScope || noBindingPlan.capturePatternBinding || !noBindingPlan.transientScopeEnvOK {
		t.Fatalf("expected bindingless clause body locals to allow transient local-scope reuse: %#v", noBindingPlan)
	}
	if clauseLocalScopePlan(ast.ID("value"), nil, ast.Lam(nil, ast.ID("value"))).transientSingleBindOK {
		t.Fatalf("expected lambda body to block transient single-binding reuse")
	}
	if clauseLocalScopePlan(ast.ID("value"), nil, ast.Spawn(ast.ID("value"))).transientSingleBindOK {
		t.Fatalf("expected spawn body to block transient single-binding reuse")
	}
}

func TestMatchExpressionClausePlansCached(t *testing.T) {
	interp := NewWithExecutor(NewSerialExecutor(nil))
	expr := ast.Match(
		ast.ID("subject"),
		ast.Mc(ast.ID("value"), ast.ID("value")),
		ast.Mc(ast.Wc(), ast.Int(0)),
	)
	first := interp.matchExpressionClausePlans(expr)
	second := interp.matchExpressionClausePlans(expr)
	if len(first) != len(expr.Clauses) {
		t.Fatalf("len(first) = %d, want %d", len(first), len(expr.Clauses))
	}
	if len(second) != len(expr.Clauses) {
		t.Fatalf("len(second) = %d, want %d", len(second), len(expr.Clauses))
	}
	if len(first) == 0 || &first[0] != &second[0] {
		t.Fatalf("expected match expression clause plans to reuse cached slice storage")
	}
	if first[0].patternBindingCount != 1 {
		t.Fatalf("first[0].patternBindingCount = %d, want 1", first[0].patternBindingCount)
	}
}

func TestRescueExpressionClausePlansCached(t *testing.T) {
	interp := NewWithExecutor(NewSerialExecutor(nil))
	expr := ast.Rescue(
		ast.ID("subject"),
		ast.Mc(ast.ID("err"), ast.ID("err")),
		ast.Mc(ast.Wc(), ast.Int(0)),
	)
	first := interp.rescueExpressionClausePlans(expr)
	second := interp.rescueExpressionClausePlans(expr)
	if len(first) != len(expr.Clauses) {
		t.Fatalf("len(first) = %d, want %d", len(first), len(expr.Clauses))
	}
	if len(second) != len(expr.Clauses) {
		t.Fatalf("len(second) = %d, want %d", len(second), len(expr.Clauses))
	}
	if len(first) == 0 || &first[0] != &second[0] {
		t.Fatalf("expected rescue expression clause plans to reuse cached slice storage")
	}
	if first[0].patternBindingCount != 1 {
		t.Fatalf("first[0].patternBindingCount = %d, want 1", first[0].patternBindingCount)
	}
}
