package typechecker

import (
	"able/interpreter-go/pkg/ast"
	"strings"
	"testing"
)

func TestConstraintArityDiagnostics(t *testing.T) {
	checker := New()

	pairSig := ast.FnSig(
		"pair",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("a", ast.Ty("A")),
			ast.Param("b", ast.Ty("B")),
		},
		ast.Ty("String"),
		nil,
		nil,
		nil,
	)
	pairIface := ast.Iface(
		"Pair",
		[]*ast.FunctionSignature{pairSig},
		[]*ast.GenericParameter{
			ast.GenericParam("A"),
			ast.GenericParam("B"),
		},
		ast.Ty("T"),
		nil,
		nil,
		false,
	)

	boxStruct := ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "value"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)

	pairImpl := ast.Impl(
		"Pair",
		ast.Ty("Box"),
		[]*ast.FunctionDefinition{
			ast.Fn(
				"pair",
				[]*ast.FunctionParameter{
					ast.Param("self", ast.Ty("Box")),
					ast.Param("a", ast.Ty("i32")),
					ast.Param("b", ast.Ty("i32")),
				},
				[]ast.Statement{ast.Ret(ast.Str("ok"))},
				ast.Ty("String"),
				nil,
				nil,
				false,
				false,
			),
		},
		nil,
		nil,
		[]ast.TypeExpression{ast.Ty("i32"), ast.Ty("i32")},
		nil,
		false,
	)

	useMissing := ast.Fn(
		"use_missing",
		[]*ast.FunctionParameter{ast.Param("value", ast.Ty("T"))},
		nil,
		ast.Ty("void"),
		[]*ast.GenericParameter{ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Pair")))},
		nil,
		false,
		false,
	)
	useMismatch := ast.Fn(
		"use_mismatch",
		[]*ast.FunctionParameter{ast.Param("value", ast.Ty("T"))},
		nil,
		ast.Ty("void"),
		[]*ast.GenericParameter{
			ast.GenericParam("T", ast.InterfaceConstr(ast.Gen(ast.Ty("Pair"), ast.Ty("i32")))),
		},
		nil,
		false,
		false,
	)

	boxLiteral := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Int(1), "value"),
		},
		false,
		"Box",
		nil,
		nil,
	)
	callMissing := ast.CallTS("use_missing", []ast.TypeExpression{ast.Ty("Box")}, boxLiteral)
	callMismatch := ast.CallTS("use_mismatch", []ast.TypeExpression{ast.Ty("Box")}, boxLiteral)

	module := ast.NewModule(
		[]ast.Statement{pairIface, boxStruct, pairImpl, useMissing, useMismatch, callMissing, callMismatch},
		nil,
		nil,
	)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected constraint arity diagnostics, got none")
	}
	var missingFound bool
	var mismatchFound bool
	for _, d := range diags {
		if strings.Contains(d.Message, "requires 2 type argument(s) for interface 'Pair'") {
			missingFound = true
		}
		if strings.Contains(d.Message, "expected 2 type argument(s) for interface 'Pair', got 1") {
			mismatchFound = true
		}
	}
	if !missingFound || !mismatchFound {
		t.Fatalf("expected constraint arity diagnostics, got %v", diags)
	}
}
