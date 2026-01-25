package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestRescueTypedPatternUnwrapsErrorPayload(t *testing.T) {
	interp := New()
	assertionStruct := ast.StructDef(
		"AssertionError",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Nullable(ast.Ty("String")), "details"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	raiseExpr := ast.Block(
		ast.Raise(
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Str("boom"), "details"),
				},
				false,
				"AssertionError",
				nil,
				nil,
			),
		),
	)
	rescueExpr := ast.Rescue(
		raiseExpr,
		ast.Mc(
			ast.TypedP(ast.ID("assertion"), ast.Ty("AssertionError")),
			ast.Member(ast.ID("assertion"), "details"),
		),
	)
	mainFn := ast.Fn(
		"main",
		nil,
		[]ast.Statement{ast.Ret(rescueExpr)},
		ast.Nullable(ast.Ty("String")),
		nil,
		nil,
		false,
		false,
	)
	module := ast.Mod([]ast.Statement{assertionStruct, mainFn}, nil, nil)
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
		strVal, ok := result.(runtime.StringValue)
		if !ok || strVal.Val != "boom" {
			t.Fatalf("expected boom, got %#v", result)
		}
	}
}
