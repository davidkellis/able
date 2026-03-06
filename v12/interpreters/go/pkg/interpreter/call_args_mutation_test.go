package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestCallFunction_DoesNotMutateCallerArgsOnCoercion(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"id",
			[]*ast.FunctionParameter{
				ast.Param("x", ast.Ty("f64")),
			},
			[]ast.Statement{
				ast.ID("x"),
			},
			ast.Ty("f64"),
			nil,
			nil,
			false,
			false,
		),
	}, nil, nil)

	if _, _, err := interp.EvaluateModule(module); err != nil {
		t.Fatalf("evaluate module: %v", err)
	}

	callee, err := interp.GlobalEnvironment().Get("id")
	if err != nil {
		t.Fatalf("lookup id: %v", err)
	}

	args := []runtime.Value{
		runtime.NewSmallInt(7, runtime.IntegerI32),
	}
	result, err := interp.CallFunction(callee, args)
	if err != nil {
		t.Fatalf("call id: %v", err)
	}

	floatResult, ok := result.(runtime.FloatValue)
	if !ok {
		t.Fatalf("expected float result, got %T (%#v)", result, result)
	}
	if floatResult.TypeSuffix != runtime.FloatF64 {
		t.Fatalf("expected coerced f64 return, got %s", floatResult.TypeSuffix)
	}

	argInt, ok := args[0].(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer arg, got %T (%#v)", args[0], args[0])
	}
	if argInt.TypeSuffix != runtime.IntegerI32 {
		t.Fatalf("expected caller arg suffix to remain i32, got %s", argInt.TypeSuffix)
	}
}

func TestCallCallableValueMutable_DoesNotMutatePartialBoundArgsOnCoercion(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"id",
			[]*ast.FunctionParameter{
				ast.Param("x", ast.Ty("f64")),
			},
			[]ast.Statement{
				ast.ID("x"),
			},
			ast.Ty("f64"),
			nil,
			nil,
			false,
			false,
		),
	}, nil, nil)

	if _, _, err := interp.EvaluateModule(module); err != nil {
		t.Fatalf("evaluate module: %v", err)
	}

	callee, err := interp.GlobalEnvironment().Get("id")
	if err != nil {
		t.Fatalf("lookup id: %v", err)
	}

	partial := &runtime.PartialFunctionValue{
		Target: callee,
		BoundArgs: []runtime.Value{
			runtime.NewSmallInt(9, runtime.IntegerI32),
		},
	}

	result, err := interp.callCallableValueMutable(partial, nil, interp.GlobalEnvironment(), nil)
	if err != nil {
		t.Fatalf("call partial: %v", err)
	}
	floatResult, ok := result.(runtime.FloatValue)
	if !ok {
		t.Fatalf("expected float result, got %T (%#v)", result, result)
	}
	if floatResult.TypeSuffix != runtime.FloatF64 {
		t.Fatalf("expected coerced f64 return, got %s", floatResult.TypeSuffix)
	}

	argInt, ok := partial.BoundArgs[0].(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected partial bound arg to stay integer, got %T (%#v)", partial.BoundArgs[0], partial.BoundArgs[0])
	}
	if argInt.TypeSuffix != runtime.IntegerI32 {
		t.Fatalf("expected partial bound arg suffix to remain i32, got %s", argInt.TypeSuffix)
	}
}
