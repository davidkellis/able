package typechecker

import (
	"strings"
	"testing"

	"able/interpreter10-go/pkg/ast"
)

func TestIteratorLiteralAnnotationMismatch(t *testing.T) {
	checker := New()
	iter := ast.IteratorLit(
		ast.Assign(ast.ID("value"), ast.Int(1)),
		ast.Yield(ast.ID("value")),
	)
	iter.ElementType = ast.Ty("string")
	module := ast.NewModule([]ast.Statement{iter}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected one diagnostic, got %v", diags)
	}
	if want := "iterator annotation expects elements of type string"; !strings.Contains(diags[0].Message, want) {
		t.Fatalf("expected diagnostic to mention %q, got %q", want, diags[0].Message)
	}
}

func TestIteratorLiteralAnnotationSatisfied(t *testing.T) {
	checker := New()
	iter := ast.IteratorLit(
		ast.Assign(ast.ID("value"), ast.Str("ok")),
		ast.Yield(ast.ID("value")),
	)
	iter.ElementType = ast.Ty("string")
	module := ast.NewModule([]ast.Statement{iter}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestForLoopIteratorTypedPatternMismatch(t *testing.T) {
	checker := New()
	iter := ast.IteratorLit(
		ast.Yield(ast.Int(1)),
	)
	iter.ElementType = ast.Ty("i32")
	loop := ast.ForLoopPattern(
		ast.TypedP(ast.PatternFrom(ast.ID("value")), ast.Ty("string")),
		iter,
		ast.Block(ast.ID("value")),
	)
	module := ast.NewModule([]ast.Statement{loop}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for typed pattern mismatch")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "for-loop pattern expects type string") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected for-loop pattern diagnostic, got %v", diags)
	}
}

func TestForLoopIteratorTypedPatternSatisfied(t *testing.T) {
	checker := New()
	iter := ast.IteratorLit(
		ast.Yield(ast.Int(1)),
	)
	iter.ElementType = ast.Ty("i32")
	loop := ast.ForLoopPattern(
		ast.TypedP(ast.PatternFrom(ast.ID("value")), ast.Ty("i32")),
		iter,
		ast.Block(ast.ID("value")),
	)
	module := ast.NewModule([]ast.Statement{loop}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestIteratorLiteralAllowsImplicitGeneratorBinding(t *testing.T) {
	checker := New()
	iter := ast.IteratorLit(
		ast.ForIn(
			"item",
			ast.Arr(ast.Int(1), ast.Int(2)),
			ast.CallExpr(
				ast.Member(ast.ID("gen"), "yield"),
				ast.ID("item"),
			),
		),
	)
	module := ast.NewModule([]ast.Statement{iter}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestForLoopIterableInterfaceTypedPatternMismatch(t *testing.T) {
	checker := New()
	displayIface := ast.Iface(
		"Display",
		[]*ast.FunctionSignature{
			ast.FnSig(
				"describe",
				[]*ast.FunctionParameter{
					ast.Param("self", ast.Ty("Self")),
				},
				ast.Ty("string"),
				nil,
				nil,
				nil,
			),
		},
		nil,
		nil,
		nil,
		nil,
		false,
	)
	iterSig := ast.FnSig(
		"iterator",
		[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
		ast.Ty("Iterator"),
		nil,
		nil,
		nil,
	)
	iterableIface := ast.Iface(
		"Iterable",
		[]*ast.FunctionSignature{iterSig},
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		nil,
		nil,
		false,
	)

	loop := ast.ForLoopPattern(
		ast.TypedP(ast.PatternFrom(ast.ID("value")), ast.Ty("Display")),
		ast.ID("items"),
		ast.Block(ast.ID("value")),
	)
	consume := ast.Fn(
		"consume",
		[]*ast.FunctionParameter{
			ast.Param("items", ast.Gen(ast.Ty("Iterable"), ast.Ty("string"))),
		},
		[]ast.Statement{
			loop,
			ast.Ret(ast.Str("done")),
		},
		ast.Ty("string"),
		nil,
		nil,
		false,
		false,
	)
	module := ast.NewModule([]ast.Statement{displayIface, iterableIface, consume}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for iterable typed pattern mismatch")
	}
	found := false
	for _, diag := range diags {
		if strings.Contains(diag.Message, "for-loop pattern expects type Display") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected for-loop diagnostic mentioning Display, got %v", diags)
	}
}
