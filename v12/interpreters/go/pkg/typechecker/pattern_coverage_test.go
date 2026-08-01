package typechecker

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestPatternCoverageUniversalPatternsAreExhaustive(t *testing.T) {
	tests := []struct {
		name    string
		pattern ast.Pattern
	}{
		{name: "wildcard", pattern: ast.Wc()},
		{name: "binding", pattern: ast.ID("value")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := ast.Match(ast.Int(1), ast.Mc(tt.pattern, ast.Int(2)))
			checker := checkPatternCoverageModule(t, match)
			assertPatternCoverage(t, checker, match, true)
		})
	}
}

func TestPatternCoverageGuardedUniversalIsNotExhaustive(t *testing.T) {
	match := ast.Match(
		ast.Int(1),
		ast.Mc(ast.Wc(), ast.Int(2), ast.Bool(true)),
	)
	checker := checkPatternCoverageModule(t, match)
	assertPatternCoverage(t, checker, match, false)
}

func TestPatternCoverageClosedNullableUnionIsExhaustive(t *testing.T) {
	match := ast.Match(
		ast.ID("subject"),
		ast.Mc(ast.TypedP(ast.ID("value"), ast.Ty("i32")), ast.Int(1)),
		ast.Mc(ast.LitP(ast.Nil()), ast.Int(0)),
	)
	fn := ast.Fn(
		"classify",
		[]*ast.FunctionParameter{ast.Param("subject", ast.Nullable(ast.Ty("i32")))},
		[]ast.Statement{ast.Ret(match)},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	checker := checkPatternCoverageModule(t, fn)
	assertPatternCoverage(t, checker, match, true)
}

func TestPatternCoverageTypedBindingCoversOpenInterface(t *testing.T) {
	match := ast.Match(
		ast.ID("subject"),
		ast.Mc(ast.TypedP(ast.ID("caught"), ast.Ty("Error")), ast.Int(1)),
	)
	fn := ast.Fn(
		"classify",
		[]*ast.FunctionParameter{ast.Param("subject", ast.Ty("Error"))},
		[]ast.Statement{ast.Ret(match)},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	checker := checkPatternCoverageModule(
		t,
		ast.Iface("Error", nil, nil, nil, nil, nil, false),
		fn,
	)
	assertPatternCoverage(t, checker, match, true)
}

func TestPatternCoverageResultSyntaxCoversSuccessAndOpenError(t *testing.T) {
	match := ast.Match(
		ast.ID("subject"),
		ast.Mc(ast.TypedP(ast.ID("value"), ast.Ty("i32")), ast.Int(1)),
		ast.Mc(ast.TypedP(ast.ID("failure"), ast.Ty("Error")), ast.Int(0)),
	)
	fn := ast.Fn(
		"classify",
		[]*ast.FunctionParameter{ast.Param("subject", ast.Result(ast.Ty("i32")))},
		[]ast.Statement{ast.Ret(match)},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	checker := checkPatternCoverageModule(
		t,
		ast.Iface("Error", nil, nil, nil, nil, nil, false),
		fn,
	)
	assertPatternCoverage(t, checker, match, true)
}

func TestPatternCoverageConcreteArmDoesNotCloseOpenInterface(t *testing.T) {
	match := ast.Match(
		ast.ID("subject"),
		ast.Mc(ast.TypedP(ast.ID("caught"), ast.Ty("ParseError")), ast.Int(1)),
	)
	fn := ast.Fn(
		"classify",
		[]*ast.FunctionParameter{ast.Param("subject", ast.Ty("Error"))},
		[]ast.Statement{ast.Ret(match)},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	checker := checkPatternCoverageModule(
		t,
		ast.Iface("Error", nil, nil, nil, nil, nil, false),
		ast.StructDef("ParseError", nil, ast.StructKindSingleton, nil, nil, false),
		fn,
	)
	assertPatternCoverage(t, checker, match, false)
}

func TestPatternCoverageTypedErrorRescueIsExhaustive(t *testing.T) {
	rescue := ast.Rescue(
		ast.Int(1),
		ast.Mc(ast.TypedP(ast.ID("caught"), ast.Ty("Error")), ast.Int(2)),
	)
	checker := checkPatternCoverageModule(
		t,
		ast.Iface("Error", nil, nil, nil, nil, nil, false),
		rescue,
	)
	assertPatternCoverage(t, checker, rescue, true)
}

func TestPatternCoverageGuardedErrorRescueIsNotExhaustive(t *testing.T) {
	rescue := ast.Rescue(
		ast.Int(1),
		ast.Mc(
			ast.TypedP(ast.ID("caught"), ast.Ty("Error")),
			ast.Int(2),
			ast.Bool(true),
		),
	)
	checker := checkPatternCoverageModule(
		t,
		ast.Iface("Error", nil, nil, nil, nil, nil, false),
		rescue,
	)
	assertPatternCoverage(t, checker, rescue, false)
}

func TestPatternCoverageIrrefutableStructPatternIsExhaustive(t *testing.T) {
	match := ast.Match(
		ast.ID("point"),
		ast.Mc(
			ast.StructP(
				[]*ast.StructPatternField{
					ast.FieldP(ast.ID("x"), "x", nil),
				},
				false,
				"Point",
			),
			ast.Int(1),
		),
	)
	fn := ast.Fn(
		"classify",
		[]*ast.FunctionParameter{ast.Param("point", ast.Ty("Point"))},
		[]ast.Statement{ast.Ret(match)},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	checker := checkPatternCoverageModule(
		t,
		ast.StructDef(
			"Point",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("i32"), "x")},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		fn,
	)
	assertPatternCoverage(t, checker, match, true)
}

func TestPatternCoverageArrayPatternRemainsRefutable(t *testing.T) {
	match := ast.Match(
		ast.ID("items"),
		ast.Mc(ast.ArrP([]ast.Pattern{ast.ID("first")}, nil), ast.Int(1)),
	)
	fn := ast.Fn(
		"classify",
		[]*ast.FunctionParameter{ast.Param("items", ast.Gen(ast.Ty("Array"), ast.Ty("i32")))},
		[]ast.Statement{ast.Ret(match)},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	checker := checkPatternCoverageModule(t, fn)
	assertPatternCoverage(t, checker, match, false)
}

func checkPatternCoverageModule(t *testing.T, statements ...ast.Statement) *Checker {
	t.Helper()
	checker := New()
	diags, err := checker.CheckModule(ast.NewModule(statements, nil, nil))
	if err != nil {
		t.Fatalf("CheckModule: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	return checker
}

func assertPatternCoverage(t *testing.T, checker *Checker, node ast.Node, want bool) {
	t.Helper()
	fact, ok := checker.PatternCoverage()[node]
	got := ok && fact.Exhaustive
	if got != want {
		t.Fatalf("exhaustive fact = %t, want %t (present=%t)", got, want, ok)
	}
}
