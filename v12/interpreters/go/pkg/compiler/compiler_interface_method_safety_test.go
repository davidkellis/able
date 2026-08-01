package compiler

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestCompilerFailsClosedForStaticOnlyInterfaceValueMethod(t *testing.T) {
	pair := ast.StructDef(
		"Pair",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("T"), "first"),
			ast.FieldDef(ast.Ty("T"), "second"),
		},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	iface := ast.Iface(
		"DuplicatePair",
		[]*ast.FunctionSignature{
			ast.FnSig(
				"duplicate_pair",
				[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
				ast.Gen(ast.Ty("Pair"), ast.Ty("Self")),
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
	callUnsafe := ast.Fn(
		"call_unsafe",
		[]*ast.FunctionParameter{ast.Param("value", ast.Ty("DuplicatePair"))},
		[]ast.Statement{
			ast.CallExpr(ast.Member(ast.ID("value"), "duplicate_pair")),
		},
		ast.Ty("void"),
		nil,
		nil,
		false,
		false,
	)
	module := ast.NewModule(
		[]ast.Statement{pair, iface, callUnsafe},
		nil,
		ast.NewPackageStatement([]*ast.Identifier{ast.ID("demo")}, false),
	)

	_, err := New(Options{
		PackageName:        "compiled",
		RequireNoFallbacks: true,
	}).Compile(testProgramFromModule("demo", module))
	if err == nil {
		t.Fatal("expected compiler to reject static-only interface-value method")
	}
	if !strings.Contains(err.Error(), "static-only interface method call rejected") {
		t.Fatalf("unexpected compiler error: %v", err)
	}
}
