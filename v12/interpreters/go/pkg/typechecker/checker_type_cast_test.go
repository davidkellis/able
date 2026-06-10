package typechecker

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestExplicitInterfaceCastDefersUnprovenImplementationToRuntime(t *testing.T) {
	checker := New()
	display := ast.Iface("Display", []*ast.FunctionSignature{
		ast.FnSig("render", []*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		}, ast.Ty("String"), nil, nil, nil),
	}, nil, nil, nil, nil, false)
	opaque := ast.StructDef("Opaque", nil, ast.StructKindNamed, nil, nil, false)
	cast := ast.NewTypeCastExpression(ast.StructLit(nil, false, "Opaque", nil, nil), ast.Ty("Display"))

	diags, err := checker.CheckModule(ast.Mod([]ast.Statement{display, opaque, cast}, nil, nil))
	if err != nil {
		t.Fatalf("CheckModule returned error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("explicit interface cast should defer an unproven implementation to runtime, got %v", diags)
	}
	if got := typeName(checker.infer[cast]); got != "Display" {
		t.Fatalf("explicit interface cast inferred %q, want Display", got)
	}
}

func TestExplicitCastStillRejectsUnrelatedNonInterfaceTargets(t *testing.T) {
	checker := New()
	opaque := ast.StructDef("Opaque", nil, ast.StructKindNamed, nil, nil, false)
	cast := ast.NewTypeCastExpression(ast.Str("not opaque"), ast.Ty("Opaque"))

	diags, err := checker.CheckModule(ast.Mod([]ast.Statement{opaque, cast}, nil, nil))
	if err != nil {
		t.Fatalf("CheckModule returned error: %v", err)
	}
	if len(diags) != 1 || !strings.Contains(diags[0].Message, "cannot cast String to Opaque") {
		t.Fatalf("expected incompatible non-interface cast diagnostic, got %v", diags)
	}
}
