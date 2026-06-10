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

func TestBytecodeVMCallCallableValueMutableDirectFunction(t *testing.T) {
	interp := NewBytecode()
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

	vm := interp.acquireBytecodeVM(interp.GlobalEnvironment())
	defer interp.releaseBytecodeVM(vm)

	got, err := vm.callCallableValueMutable(addOne, []runtime.Value{
		runtime.NewSmallInt(41, runtime.IntegerI32),
	}, nil)
	if err != nil {
		t.Fatalf("vm call: %v", err)
	}
	intVal, ok := got.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T (%#v)", got, got)
	}
	if intVal.BigInt().Int64() != 42 {
		t.Fatalf("expected 42, got %#v", got)
	}
}

func TestBytecodeVMCallCallableValueMutableDirectFunctionPartial(t *testing.T) {
	interp := NewBytecode()
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

	vm := interp.acquireBytecodeVM(interp.GlobalEnvironment())
	defer interp.releaseBytecodeVM(vm)

	partial, err := vm.callCallableValueMutable(triVal, []runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
	}, nil)
	if err != nil {
		t.Fatalf("vm partial call: %v", err)
	}
	if _, ok := partial.(*runtime.PartialFunctionValue); !ok {
		t.Fatalf("expected partial function, got %T (%#v)", partial, partial)
	}

	got, err := interp.CallFunction(partial, []runtime.Value{
		runtime.NewSmallInt(2, runtime.IntegerI32),
		runtime.NewSmallInt(3, runtime.IntegerI32),
	})
	if err != nil {
		t.Fatalf("complete partial call: %v", err)
	}
	intVal, ok := got.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T (%#v)", got, got)
	}
	if intVal.BigInt().Int64() != 123 {
		t.Fatalf("expected 123, got %#v", got)
	}
}

func TestBytecodeVMCallCallableValueWithInjectedReceiverPartial(t *testing.T) {
	interp := NewBytecode()
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

	vm := interp.acquireBytecodeVM(interp.GlobalEnvironment())
	defer interp.releaseBytecodeVM(vm)

	partial, err := vm.callCallableValueWithInjectedReceiver(
		triVal,
		runtime.NewSmallInt(1, runtime.IntegerI32),
		[]runtime.Value{runtime.NewSmallInt(2, runtime.IntegerI32)},
		nil,
	)
	if err != nil {
		t.Fatalf("vm injected partial call: %v", err)
	}
	partialFn, ok := partial.(*runtime.PartialFunctionValue)
	if !ok || partialFn == nil {
		t.Fatalf("expected partial function, got %T (%#v)", partial, partial)
	}
	if len(partialFn.BoundArgs) != 2 {
		t.Fatalf("partial bound arg count = %d, want 2", len(partialFn.BoundArgs))
	}
	if !valuesEqual(partialFn.BoundArgs[0], runtime.NewSmallInt(1, runtime.IntegerI32)) {
		t.Fatalf("partial receiver = %#v, want 1", partialFn.BoundArgs[0])
	}
	if !valuesEqual(partialFn.BoundArgs[1], runtime.NewSmallInt(2, runtime.IntegerI32)) {
		t.Fatalf("partial explicit arg = %#v, want 2", partialFn.BoundArgs[1])
	}

	got, err := interp.CallFunction(partialFn, []runtime.Value{
		runtime.NewSmallInt(3, runtime.IntegerI32),
	})
	if err != nil {
		t.Fatalf("complete partial call: %v", err)
	}
	intVal, ok := got.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T (%#v)", got, got)
	}
	if intVal.BigInt().Int64() != 123 {
		t.Fatalf("expected 123, got %#v", got)
	}
}

func TestBytecodeVMCallCallableValueWithInjectedReceiverOverload(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()

	intFnDef := ast.Fn(
		"render",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("i32")),
			ast.Param("value", ast.Ty("i32")),
		},
		[]ast.Statement{ast.Str("int")},
		ast.Ty("String"),
		nil,
		nil,
		false,
		false,
	)
	intProgram, err := interp.lowerFunctionDefinitionBytecode(intFnDef)
	if err != nil {
		t.Fatalf("lower int overload: %v", err)
	}
	intFn := &runtime.FunctionValue{Declaration: intFnDef, Closure: env}
	setFunctionBytecodeProgram(intFn, intProgram)

	stringFnDef := ast.Fn(
		"render",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("i32")),
			ast.Param("value", ast.Ty("String")),
		},
		[]ast.Statement{ast.Str("string")},
		ast.Ty("String"),
		nil,
		nil,
		false,
		false,
	)
	stringProgram, err := interp.lowerFunctionDefinitionBytecode(stringFnDef)
	if err != nil {
		t.Fatalf("lower string overload: %v", err)
	}
	stringFn := &runtime.FunctionValue{Declaration: stringFnDef, Closure: env}
	setFunctionBytecodeProgram(stringFn, stringProgram)

	overload := &runtime.FunctionOverloadValue{Overloads: []*runtime.FunctionValue{intFn, stringFn}}

	vm := interp.acquireBytecodeVM(env)
	defer interp.releaseBytecodeVM(vm)

	got, err := vm.callCallableValueWithInjectedReceiver(
		overload,
		runtime.NewSmallInt(7, runtime.IntegerI32),
		[]runtime.Value{runtime.NewSmallInt(4, runtime.IntegerI32)},
		nil,
	)
	if err != nil {
		t.Fatalf("vm injected overload call: %v", err)
	}
	strVal, ok := got.(runtime.StringValue)
	if !ok {
		t.Fatalf("expected string result, got %T (%#v)", got, got)
	}
	if strVal.Val != "int" {
		t.Fatalf("expected int overload, got %#v", got)
	}
}
