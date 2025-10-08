package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func keysOf(m map[string]runtime.Value) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func setupShowPoint(t *testing.T, interp *Interpreter) {
	t.Helper()
	show := ast.Iface(
		"Show",
		[]*ast.FunctionSignature{
			ast.FnSig(
				"to_string",
				[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
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
	if _, _, err := interp.EvaluateModule(ast.Mod([]ast.Statement{show}, nil, nil)); err != nil {
		t.Fatalf("interface evaluation failed: %v", err)
	}
	point := ast.StructDef(
		"Point",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "x"),
			ast.FieldDef(ast.Ty("i32"), "y"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	if _, _, err := interp.EvaluateModule(ast.Mod([]ast.Statement{point}, nil, nil)); err != nil {
		t.Fatalf("struct evaluation failed: %v", err)
	}
	toString := ast.Fn(
		"to_string",
		[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Point"))},
		[]ast.Statement{
			ast.Ret(
				ast.Interp(
					ast.Str("Point("),
					ast.Member(ast.ID("self"), "x"),
					ast.Str(", "),
					ast.Member(ast.ID("self"), "y"),
					ast.Str(")"),
				),
			),
		},
		ast.Ty("string"),
		nil,
		nil,
		false,
		false,
	)
	acceptShow := ast.Fn(
		"accept_show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Point")),
			ast.Param("x", nil),
		},
		[]ast.Statement{
			ast.Ret(ast.CallExpr(ast.Member(ast.ID("x"), "to_string"))),
		},
		ast.Ty("string"),
		[]*ast.GenericParameter{
			ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Show"))),
		},
		nil,
		false,
		false,
	)
	methods := ast.Methods(
		ast.Ty("Point"),
		[]*ast.FunctionDefinition{toString, acceptShow},
		nil,
		nil,
	)
	if _, _, err := interp.EvaluateModule(ast.Mod([]ast.Statement{methods}, nil, nil)); err != nil {
		t.Fatalf("methods evaluation failed: %v", err)
	}
}

func bigInt(v int64) *big.Int {
	return big.NewInt(v)
}
