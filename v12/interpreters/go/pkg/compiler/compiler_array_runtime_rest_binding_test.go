package compiler

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func newArrayRestBindingTestContext() *compileContext {
	temps := 0
	return &compileContext{
		locals:      make(map[string]paramInfo),
		packageName: "demo",
		temps:       &temps,
	}
}

func TestCompilerRuntimeArrayMatchRestBindingUsesGenericArrayCarrier(t *testing.T) {
	gen := newGenerator(Options{PackageName: "demo"})
	gen.ensureBuiltinArrayStruct()
	ctx := newArrayRestBindingTestContext()
	pattern := ast.ArrP([]ast.Pattern{
		ast.LitP(ast.Int(1)),
		ast.LitP(ast.Int(2)),
	}, ast.ID("tail"))

	lines, ok := gen.compileRuntimeArrayPatternBindings(ctx, pattern, "subject")
	if !ok {
		t.Fatalf("expected runtime array rest binding to compile, got reason %q", ctx.reason)
	}
	joined := strings.Join(lines, "\n")
	for _, fragment := range []string{
		"var tail *Array =",
		"&Array{Elements:",
		"__able_struct_Array_sync(",
	} {
		if !strings.Contains(joined, fragment) {
			t.Fatalf("expected runtime array rest binding lowering to contain %q:\n%s", fragment, joined)
		}
	}
	if strings.Contains(joined, "runtime.ArrayValue") {
		t.Fatalf("expected runtime array rest binding to avoid runtime.ArrayValue materialization:\n%s", joined)
	}

	binding, ok := ctx.lookup("tail")
	if !ok {
		t.Fatalf("expected tail binding to be recorded in compile context")
	}
	if binding.GoType != "*Array" {
		t.Fatalf("expected tail binding GoType=*Array, got %q", binding.GoType)
	}
	if got := typeExpressionToString(binding.TypeExpr); got != "Array<<?>>" {
		t.Fatalf("expected tail binding type expr Array<<?>>, got %q", got)
	}
}

func TestCompilerRuntimeArrayPatternAssignmentRestBindingUsesGenericArrayCarrier(t *testing.T) {
	gen := newGenerator(Options{PackageName: "demo"})
	gen.ensureBuiltinArrayStruct()
	ctx := newArrayRestBindingTestContext()
	pattern := ast.ArrP([]ast.Pattern{
		ast.LitP(ast.Int(1)),
		ast.LitP(ast.Int(2)),
	}, ast.ID("tail"))

	lines, ok := gen.compileRuntimeArrayPatternAssignmentBindings(ctx, pattern, "subject", patternBindingMode{
		declare:  true,
		newNames: map[string]struct{}{"tail": {}},
	})
	if !ok {
		t.Fatalf("expected runtime array pattern assignment rest binding to compile, got reason %q", ctx.reason)
	}
	joined := strings.Join(lines, "\n")
	for _, fragment := range []string{
		"var tail *Array =",
		"&Array{Elements:",
		"__able_struct_Array_sync(",
	} {
		if !strings.Contains(joined, fragment) {
			t.Fatalf("expected runtime assignment rest binding lowering to contain %q:\n%s", fragment, joined)
		}
	}
	if strings.Contains(joined, "runtime.ArrayValue") {
		t.Fatalf("expected runtime assignment rest binding to avoid runtime.ArrayValue materialization:\n%s", joined)
	}

	binding, ok := ctx.lookup("tail")
	if !ok {
		t.Fatalf("expected tail binding to be recorded in compile context")
	}
	if binding.GoType != "*Array" {
		t.Fatalf("expected tail binding GoType=*Array, got %q", binding.GoType)
	}
	if got := typeExpressionToString(binding.TypeExpr); got != "Array<<?>>" {
		t.Fatalf("expected tail binding type expr Array<<?>>, got %q", got)
	}
}
