package typechecker

import (
	"able/interpreter10-go/pkg/ast"
	"strings"
	"testing"
)

func TestBlockExpressionScopesBindings(t *testing.T) {
	checker := New()
	block := ast.Block(
		ast.Assign(ast.ID("inner"), ast.Int(5)),
		ast.ID("inner"),
	)
	module := ast.NewModule([]ast.Statement{block}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[block]
	if !ok {
		t.Fatalf("expected block inference entry")
	}
	if typeName(typ) != "i32" {
		t.Fatalf("expected block to return i32, got %q", typeName(typ))
	}
	if _, exists := checker.global.Lookup("inner"); exists {
		t.Fatalf("expected inner binding to remain scoped to block")
	}
}
func TestIfExpressionMergesBranchTypes(t *testing.T) {
	checker := New()
	ifExpr := ast.IfExpr(
		ast.Bool(true),
		ast.Block(ast.Int(1)),
		ast.OrC(ast.Block(ast.Int(2)), ast.Bool(false)),
		ast.OrC(ast.Block(ast.Int(3)), nil),
	)
	assign := ast.Assign(ast.ID("value"), ifExpr)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[ifExpr]
	if !ok {
		t.Fatalf("expected inference for if expression")
	}
	if typeName(typ) != "i32" {
		t.Fatalf("expected if expression to have type i32, got %q", typeName(typ))
	}
}
func TestIfExpressionConditionMustBeBool(t *testing.T) {
	checker := New()
	ifExpr := ast.IfExpr(
		ast.Int(1),
		ast.Block(ast.Int(2)),
	)
	module := ast.NewModule([]ast.Statement{ifExpr}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostics for non-bool condition")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "condition must be bool") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected condition bool diagnostic, got %v", diags)
	}
}
func TestReturnOutsideFunctionProducesDiagnostic(t *testing.T) {
	checker := New()
	module := ast.NewModule([]ast.Statement{ast.Ret(nil)}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for return outside function")
	}
	if want := "outside function"; !strings.Contains(diags[0].Message, want) {
		t.Fatalf("expected diagnostic mentioning %q, got %v", want, diags)
	}
}
func TestRescueExpressionMergesTypes(t *testing.T) {
	checker := New()
	assign := ast.Assign(ast.ID("value"), ast.Str("ok"))
	rescue := ast.Rescue(
		ast.ID("value"),
		ast.Mc(ast.Wc(), ast.Str("fallback")),
	)
	module := ast.NewModule([]ast.Statement{assign, rescue}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[rescue]
	if !ok {
		t.Fatalf("expected inference entry for rescue expression")
	}
	if typeName(typ) != "string" {
		t.Fatalf("expected rescue expression to infer string, got %q", typeName(typ))
	}
}
func TestRescueGuardMustBeBool(t *testing.T) {
	checker := New()
	assign := ast.Assign(ast.ID("value"), ast.Str("ok"))
	rescue := ast.Rescue(
		ast.ID("value"),
		ast.Mc(ast.Wc(), ast.Str("fallback"), ast.Int(1)),
	)
	module := ast.NewModule([]ast.Statement{assign, rescue}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "rescue guard must evaluate to bool") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected rescue guard diagnostic, got %v", diags)
	}
}
func TestOrElseExpressionMergesTypes(t *testing.T) {
	checker := New()
	assign := ast.Assign(ast.ID("value"), ast.Str("ok"))
	orElse := ast.OrElseBlock(
		ast.ID("value"),
		ast.Block(ast.Str("fallback")),
		nil,
	)
	module := ast.NewModule([]ast.Statement{assign, orElse}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[orElse]
	if !ok {
		t.Fatalf("expected inference entry for or-else expression")
	}
	if typeName(typ) != "string" {
		t.Fatalf("expected or-else expression to infer string, got %q", typeName(typ))
	}
}
func TestOrElseProducesUnionWhenTypesDiffer(t *testing.T) {
	checker := New()
	orElse := ast.OrElseBlock(
		ast.Int(1),
		ast.Block(ast.Str("fallback")),
		nil,
	)
	module := ast.NewModule([]ast.Statement{orElse}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[orElse]
	if !ok {
		t.Fatalf("expected inference entry for or-else expression")
	}
	if typeName(typ) != "i32 | string" {
		t.Fatalf("expected union literal for mismatched or-else types, got %q", typeName(typ))
	}
}
func TestOrElseBindsErrorInHandler(t *testing.T) {
	checker := New()
	errorStruct := ast.StructDef(
		"Error",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("string"), "message"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	assign := ast.Assign(ast.ID("value"), ast.Str("ok"))
	orElse := ast.OrElseBlock(
		ast.ID("value"),
		ast.Block(ast.Member(ast.ID("err"), "message")),
		"err",
	)
	module := ast.NewModule([]ast.Statement{errorStruct, assign, orElse}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for or-else handler, got %v", diags)
	}
}
func TestEnsureExpressionReturnsTryType(t *testing.T) {
	checker := New()
	assign := ast.Assign(ast.ID("value"), ast.Int(1))
	ensure := ast.Ensure(
		ast.ID("value"),
		ast.Assign(ast.ID("cleanup"), ast.Int(0)),
	)
	module := ast.NewModule([]ast.Statement{assign, ensure}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[ensure]
	if !ok {
		t.Fatalf("expected inference entry for ensure expression")
	}
	if typeName(typ) != "i32" {
		t.Fatalf("expected ensure expression to infer i32, got %q", typeName(typ))
	}
}
func TestRaiseAllowsAnyValue(t *testing.T) {
	checker := New()
	module := ast.NewModule([]ast.Statement{ast.Raise(ast.Int(1))}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for raise with non-error value, got %v", diags)
	}
}
func TestRaiseAcceptsErrorStruct(t *testing.T) {
	checker := New()
	errorStruct := ast.StructDef(
		"Error",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("string"), "message"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	raise := ast.Raise(
		ast.StructLit(
			[]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Str("boom"), "message"),
			},
			false,
			"Error",
			nil,
			nil,
		),
	)
	module := ast.NewModule([]ast.Statement{errorStruct, raise}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for valid raise, got %v", diags)
	}
	if len(checker.obligations) != 0 {
		t.Fatalf("expected no obligations for raise check, got %v", checker.obligations)
	}
}
func TestBreakRequiresLoop(t *testing.T) {
	checker := New()
	module := ast.NewModule([]ast.Statement{ast.Brk(nil, nil)}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for break outside loop")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "break statement must appear inside a loop") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing break diagnostic: %v", diags)
	}
}
func TestContinueRequiresLoop(t *testing.T) {
	checker := New()
	module := ast.NewModule([]ast.Statement{ast.Cont(nil)}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for continue outside loop")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "continue statement must appear inside a loop") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing continue diagnostic: %v", diags)
	}
}
func TestWhileConditionMustBeBool(t *testing.T) {
	checker := New()
	loop := ast.Wloop(
		ast.Int(1),
		ast.Block(),
	)
	module := ast.NewModule([]ast.Statement{loop}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for while condition")
	}
	if want := "while condition must be bool"; !strings.Contains(diags[0].Message, want) {
		t.Fatalf("expected message %q, got %q", want, diags[0].Message)
	}
}
func TestForLoopIterableMustBeArrayRangeOrIterator(t *testing.T) {
	checker := New()
	loop := ast.ForIn(ast.ID("value"), ast.Int(5))
	module := ast.NewModule([]ast.Statement{loop}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "for-loop iterable must be array, range, or iterator") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected for-loop iterable diagnostic, got %v", diags)
	}
}
func TestForLoopRangeElementType(t *testing.T) {
	checker := New()
	pattern := ast.ID("value")
	bodyExpr := ast.ID("value")
	loop := ast.ForLoopPattern(
		pattern,
		ast.Range(ast.Int(0), ast.Int(3), false),
		ast.Block(bodyExpr),
	)
	module := ast.NewModule([]ast.Statement{loop}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[bodyExpr]
	if !ok {
		t.Fatalf("expected inference for range loop body identifier")
	}
	if typeName(typ) != "i32" {
		t.Fatalf("expected loop value to have type i32, got %q", typeName(typ))
	}
}
