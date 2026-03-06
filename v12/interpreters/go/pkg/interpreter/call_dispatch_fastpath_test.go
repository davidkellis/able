package interpreter

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestCallDispatchPartialChainPreservesBoundArgOrder(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"tri",
			[]*ast.FunctionParameter{
				ast.Param("a", ast.Ty("i32")),
				ast.Param("b", ast.Ty("i32")),
				ast.Param("c", ast.Ty("i32")),
			},
			[]ast.Statement{
				ast.Bin(
					"+",
					ast.Bin(
						"+",
						ast.Bin("*", ast.ID("a"), ast.Int(100)),
						ast.Bin("*", ast.ID("b"), ast.Int(10)),
					),
					ast.ID("c"),
				),
			},
			ast.Ty("i32"),
			nil,
			nil,
			false,
			false,
		),
	}, nil, nil)
	if _, _, err := interp.EvaluateModule(module); err != nil {
		t.Fatalf("evaluate module: %v", err)
	}

	triVal, err := interp.GlobalEnvironment().Get("tri")
	if err != nil {
		t.Fatalf("lookup tri: %v", err)
	}

	p1, err := interp.CallFunction(triVal, []runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
	})
	if err != nil {
		t.Fatalf("partial call 1: %v", err)
	}
	if _, ok := p1.(*runtime.PartialFunctionValue); !ok {
		t.Fatalf("expected first call to return partial, got %T (%#v)", p1, p1)
	}

	p2, err := interp.CallFunction(p1, []runtime.Value{
		runtime.NewSmallInt(2, runtime.IntegerI32),
	})
	if err != nil {
		t.Fatalf("partial call 2: %v", err)
	}
	if _, ok := p2.(*runtime.PartialFunctionValue); !ok {
		t.Fatalf("expected second call to return partial, got %T (%#v)", p2, p2)
	}

	got, err := interp.CallFunction(p2, []runtime.Value{
		runtime.NewSmallInt(3, runtime.IntegerI32),
	})
	if err != nil {
		t.Fatalf("final call: %v", err)
	}
	intVal, ok := got.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T (%#v)", got, got)
	}
	if intVal.BigInt().Int64() != 123 {
		t.Fatalf("expected 123, got %#v", got)
	}
}

func TestCallDispatchSingleOverloadMismatchReportsParameterType(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"add_one",
			[]*ast.FunctionParameter{
				ast.Param("x", ast.Ty("i32")),
			},
			[]ast.Statement{
				ast.Bin("+", ast.ID("x"), ast.Int(1)),
			},
			ast.Ty("i32"),
			nil,
			nil,
			false,
			false,
		),
	}, nil, nil)
	if _, _, err := interp.EvaluateModule(module); err != nil {
		t.Fatalf("evaluate module: %v", err)
	}

	addOne, err := interp.GlobalEnvironment().Get("add_one")
	if err != nil {
		t.Fatalf("lookup add_one: %v", err)
	}

	_, err = interp.CallFunction(addOne, []runtime.Value{runtime.StringValue{Val: "oops"}})
	if err == nil {
		t.Fatalf("expected parameter mismatch error")
	}
	if !strings.Contains(err.Error(), "Parameter type mismatch") {
		t.Fatalf("expected parameter mismatch error, got %v", err)
	}
}
