package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestSingletonStructImplMethodAccess(t *testing.T) {
	interp := New()
	iface := ast.Iface(
		"Matcher",
		[]*ast.FunctionSignature{
			ast.FnSig(
				"matches",
				[]*ast.FunctionParameter{
					ast.Param("self", ast.Ty("M")),
					ast.Param("actual", ast.Nullable(ast.Ty("T"))),
				},
				ast.Ty("bool"),
				nil,
				nil,
				nil,
			),
		},
		[]*ast.GenericParameter{ast.GenericParam("T")},
		ast.Ty("M"),
		nil,
		nil,
		false,
	)
	structDef := ast.StructDef("BeNilMatcher", nil, ast.StructKindNamed, nil, nil, false)
	impl := ast.Impl(
		"Matcher",
		ast.Ty("BeNilMatcher"),
		[]*ast.FunctionDefinition{
			ast.Fn(
				"matches",
				[]*ast.FunctionParameter{
					ast.Param("self", ast.Ty("Self")),
					ast.Param("actual", ast.Nullable(ast.Ty("T"))),
				},
				[]ast.Statement{ast.Ret(ast.Bool(true))},
				ast.Ty("bool"),
				nil,
				nil,
				false,
				false,
			),
		},
		nil,
		nil,
		[]ast.TypeExpression{ast.Nullable(ast.Ty("T"))},
		nil,
		false,
	)
	mainFn := ast.Fn(
		"main",
		nil,
		[]ast.Statement{
			ast.Ret(ast.CallExpr(ast.Member(ast.ID("BeNilMatcher"), "matches"), ast.Nil())),
		},
		ast.Ty("bool"),
		nil,
		nil,
		false,
		false,
	)
	module := ast.Mod([]ast.Statement{iface, structDef, impl, mainFn}, nil, nil)
	if _, env, err := interp.EvaluateModule(module); err != nil {
		t.Fatalf("module evaluation failed: %v", err)
	} else {
		callee, err := env.Get("main")
		if err != nil {
			t.Fatalf("missing main: %v", err)
		}
		result, err := interp.callCallableValue(callee, nil, env, nil)
		if err != nil {
			t.Fatalf("calling main failed: %v", err)
		}
		boolVal, ok := result.(runtime.BoolValue)
		if !ok || !boolVal.Val {
			t.Fatalf("expected true, got %#v", result)
		}
	}
}
