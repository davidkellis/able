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

func TestCallFunctionBytecode_DoesNotMutateCallerArgsOnCoercion(t *testing.T) {
	interp := NewBytecode()
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

func TestBytecodePrepareCallArgsWithOptionalReceiver_DoesNotMutateSourceArgsOnStableCopy(t *testing.T) {
	source := []runtime.Value{
		runtime.NewSmallInt(7, runtime.IntegerI32),
	}
	receiver := runtime.StringValue{Val: "recv"}

	prepared := bytecodePrepareCallArgsWithOptionalReceiver(source, true, receiver, true)

	if len(prepared) != 2 {
		t.Fatalf("prepared arg count = %d, want 2", len(prepared))
	}
	if got, ok := prepared[0].(runtime.StringValue); !ok || got.Val != "recv" {
		t.Fatalf("prepared receiver = %#v, want recv", prepared[0])
	}
	if got, ok := prepared[1].(runtime.IntegerValue); !ok || got.TypeSuffix != runtime.IntegerI32 {
		t.Fatalf("prepared arg = %#v, want materialized i32 integer", prepared[1])
	}
	if len(source) != 1 {
		t.Fatalf("source arg count = %d, want 1", len(source))
	}
	if got, ok := source[0].(runtime.IntegerValue); !ok || got.TypeSuffix != runtime.IntegerI32 {
		t.Fatalf("source arg mutated to %#v, want original i32 integer", source[0])
	}
}

func TestBytecodePrepareCallArgsWithOptionalReceiverIntoBuffer_DoesNotMutateSourceArgs(t *testing.T) {
	source := []runtime.Value{
		runtime.NewSmallInt(7, runtime.IntegerI32),
	}
	receiver := runtime.StringValue{Val: "recv"}
	var inline [bytecodeInlinePreparedCallArgStorage]runtime.Value

	prepared, ok := bytecodePrepareCallArgsWithOptionalReceiverIntoBuffer(inline[:], source, false, receiver, true)
	if !ok {
		t.Fatalf("expected inline bytecode injected-receiver arg preparation")
	}
	if len(prepared) != 2 {
		t.Fatalf("prepared arg count = %d, want 2", len(prepared))
	}
	if got, ok := prepared[0].(runtime.StringValue); !ok || got.Val != "recv" {
		t.Fatalf("prepared receiver = %#v, want recv", prepared[0])
	}
	if got, ok := prepared[1].(runtime.IntegerValue); !ok || got.TypeSuffix != runtime.IntegerI32 {
		t.Fatalf("prepared arg = %#v, want materialized i32 integer", prepared[1])
	}
	if len(source) != 1 {
		t.Fatalf("source arg count = %d, want 1", len(source))
	}
	if got, ok := source[0].(runtime.IntegerValue); !ok || got.TypeSuffix != runtime.IntegerI32 {
		t.Fatalf("source arg mutated to %#v, want original i32 integer", source[0])
	}
}

func TestBytecodePrepareCallArgsIntoBuffer_DoesNotMutateSourceArgsOnStableCopy(t *testing.T) {
	source := []runtime.Value{
		runtime.NewSmallInt(7, runtime.IntegerI32),
	}
	var inline [bytecodeInlinePreparedCallArgStorage]runtime.Value

	prepared, ok := bytecodePrepareCallArgsIntoBuffer(inline[:], source, true)
	if !ok {
		t.Fatalf("expected inline bytecode arg preparation")
	}
	if len(prepared) != 1 {
		t.Fatalf("prepared arg count = %d, want 1", len(prepared))
	}
	if got, ok := prepared[0].(runtime.IntegerValue); !ok || got.TypeSuffix != runtime.IntegerI32 {
		t.Fatalf("prepared arg = %#v, want materialized i32 integer", prepared[0])
	}
	if len(source) != 1 {
		t.Fatalf("source arg count = %d, want 1", len(source))
	}
	if got, ok := source[0].(runtime.IntegerValue); !ok || got.TypeSuffix != runtime.IntegerI32 {
		t.Fatalf("source arg mutated to %#v, want original i32 integer", source[0])
	}
}

func TestBytecodePrepareCallArgsIntoBuffer_RequiresStableCopy(t *testing.T) {
	source := []runtime.Value{
		runtime.NewSmallInt(7, runtime.IntegerI32),
	}
	var inline [bytecodeInlinePreparedCallArgStorage]runtime.Value

	if prepared, ok := bytecodePrepareCallArgsIntoBuffer(inline[:], source, false); ok || prepared != nil {
		t.Fatalf("expected non-stable args to skip inline copy, got %#v", prepared)
	}
}
