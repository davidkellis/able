package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestApplyInterfaceCalls(t *testing.T) {
	interp := New()
	applyInterface := ast.Iface(
		"Apply",
		[]*ast.FunctionSignature{
			ast.FnSig(
				"apply",
				[]*ast.FunctionParameter{
					ast.Param("self", ast.Ty("Self")),
					ast.Param("args", ast.Ty("Args")),
				},
				ast.Ty("Result"),
				nil,
				nil,
				nil,
			),
		},
		[]*ast.GenericParameter{ast.GenericParam("Args", nil), ast.GenericParam("Result", nil)},
		nil,
		nil,
		nil,
		false,
	)
	multStruct := ast.StructDef(
		"Multiplier",
		[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("i32"), "factor")},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	applyFn := ast.Fn(
		"apply",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("input", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Ret(ast.Bin("*", ast.ImplicitMember("factor"), ast.ID("input"))),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	applyImpl := ast.Impl(
		"Apply",
		ast.Ty("Multiplier"),
		[]*ast.FunctionDefinition{applyFn},
		nil,
		nil,
		[]ast.TypeExpression{ast.Ty("i32"), ast.Ty("i32")},
		nil,
		false,
	)
	assign := ast.Assign(
		ast.ID("m"),
		ast.StructLit(
			[]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(3), "factor")},
			false,
			"Multiplier",
			nil,
			nil,
		),
	)
	call := ast.CallExpr(ast.ID("m"), ast.Int(5))
	module := ast.Mod([]ast.Statement{applyInterface, multStruct, applyImpl, assign, call}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("apply callable failed: %v", err)
	}
	intResult, ok := result.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %#v", result)
	}
	if intResult.Val.Cmp(big.NewInt(15)) != 0 {
		t.Fatalf("expected 15, got %v", intResult.Val)
	}
}
