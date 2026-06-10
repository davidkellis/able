package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestCallCallableValue_DirectCompiledThunkKeepsPayloadStateLazy(t *testing.T) {
	interp := New()
	payload := &asyncContextPayload{}
	env := runtime.NewEnvironment(nil)
	env.SetRuntimeData(payload)
	fn := &runtime.FunctionValue{
		Declaration: ast.Fn(
			"const_one",
			nil,
			[]ast.Statement{ast.Int(1)},
			ast.Ty("i32"),
			nil,
			nil,
			false,
			false,
		),
		Closure: env,
		Bytecode: CompiledThunk(func(env *runtime.Environment, args []runtime.Value) (runtime.Value, error) {
			if env == nil {
				t.Fatalf("compiled thunk env is nil")
			}
			if len(args) != 0 {
				t.Fatalf("compiled thunk args = %d, want 0", len(args))
			}
			return runtime.NewSmallInt(1, runtime.IntegerI32), nil
		}),
	}

	if payload.state != nil {
		t.Fatalf("payload state initialized before call")
	}

	got, err := interp.callCallableValue(fn, nil, env, nil)
	if err != nil {
		t.Fatalf("callCallableValue failed: %v", err)
	}
	intVal, ok := got.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T (%#v)", got, got)
	}
	if intVal.BigInt().Int64() != 1 {
		t.Fatalf("expected 1, got %#v", got)
	}
	if payload.state != nil {
		t.Fatalf("expected payload state to remain nil for direct compiled thunk call")
	}
}

func TestCallFunction_CompiledThunkWithoutImplicitReceiverKeepsRootStateLazy(t *testing.T) {
	interp := New()
	payload := &asyncContextPayload{}
	env := runtime.NewEnvironment(nil)
	env.SetRuntimeData(payload)
	fn := &runtime.FunctionValue{
		Declaration: ast.Fn(
			"const_one",
			nil,
			[]ast.Statement{ast.Int(1)},
			ast.Ty("i32"),
			nil,
			nil,
			false,
			false,
		),
		Closure: env,
		Bytecode: CompiledThunk(func(env *runtime.Environment, args []runtime.Value) (runtime.Value, error) {
			if env == nil {
				t.Fatalf("compiled thunk env is nil")
			}
			if len(args) != 0 {
				t.Fatalf("compiled thunk args = %d, want 0", len(args))
			}
			return runtime.NewSmallInt(1, runtime.IntegerI32), nil
		}),
	}

	if payload.state != nil {
		t.Fatalf("payload state initialized before call")
	}

	got, err := interp.CallFunctionIn(fn, nil, env)
	if err != nil {
		t.Fatalf("CallFunction failed: %v", err)
	}
	intVal, ok := got.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T (%#v)", got, got)
	}
	if intVal.BigInt().Int64() != 1 {
		t.Fatalf("expected 1, got %#v", got)
	}
	if payload.state != nil {
		t.Fatalf("expected payload state to remain nil for call without diagnostics or implicit receiver")
	}
}

func TestCallCallableValue_NativeReceivesRuntimeDataWithoutInitializingPayloadState(t *testing.T) {
	interp := New()
	payload := &asyncContextPayload{}
	env := runtime.NewEnvironment(nil)
	env.SetRuntimeData(payload)
	called := false

	got, err := interp.callCallableValue(runtime.NativeFunctionValue{
		Name:  "capture_ctx",
		Arity: 0,
		Impl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			called = true
			if ctx == nil {
				t.Fatalf("native call context is nil")
			}
			if ctx.Env != env {
				t.Fatalf("native ctx env mismatch: got=%p want=%p", ctx.Env, env)
			}
			if ctx.State != payload {
				t.Fatalf("native ctx state mismatch: got=%#v want=%#v", ctx.State, payload)
			}
			if len(args) != 0 {
				t.Fatalf("expected no args, got %d", len(args))
			}
			return runtime.NewSmallInt(1, runtime.IntegerI32), nil
		},
	}, nil, env, nil)
	if err != nil {
		t.Fatalf("callCallableValue native failed: %v", err)
	}
	if !called {
		t.Fatalf("expected native impl to be called")
	}
	intVal, ok := got.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T (%#v)", got, got)
	}
	if intVal.BigInt().Int64() != 1 {
		t.Fatalf("expected 1, got %#v", got)
	}
	if payload.state != nil {
		t.Fatalf("expected payload state to remain nil for native call without diagnostics")
	}
}

func TestCallCallableValue_CompiledThunkMaterializesRawIntegerArgs(t *testing.T) {
	interp := NewBytecode()
	env := runtime.NewEnvironment(nil)
	fn := &runtime.FunctionValue{
		Declaration: ast.Fn(
			"id_i32",
			[]*ast.FunctionParameter{ast.Param("x", ast.Ty("i32"))},
			[]ast.Statement{ast.ID("x")},
			ast.Ty("i32"),
			nil,
			nil,
			false,
			false,
		),
		Closure: env,
		Bytecode: CompiledThunk(func(env *runtime.Environment, args []runtime.Value) (runtime.Value, error) {
			if env == nil {
				t.Fatalf("compiled thunk env is nil")
			}
			if len(args) != 1 {
				t.Fatalf("compiled thunk args = %d, want 1", len(args))
			}
			intVal, ok := args[0].(runtime.IntegerValue)
			if !ok || intVal.TypeSuffix != runtime.IntegerI32 || intVal.Int64Fast() != int64(bytecodeRawI32SlotCacheMax+17) {
				t.Fatalf("compiled thunk arg = %#v, want boxed i32", args[0])
			}
			return args[0], nil
		}),
	}

	got, err := interp.callCallableValue(fn, []runtime.Value{(&bytecodeRawI32StackCell{Val: int32(bytecodeRawI32SlotCacheMax + 17)})}, env, nil)
	if err != nil {
		t.Fatalf("callCallableValue failed: %v", err)
	}
	assertIntValue(t, got, runtime.IntegerI32, int64(bytecodeRawI32SlotCacheMax+17))
}
