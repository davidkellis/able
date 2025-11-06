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
