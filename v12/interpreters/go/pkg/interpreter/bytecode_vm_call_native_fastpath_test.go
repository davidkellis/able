package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_NativeBoundMethodExactCallInjectsReceiverOnce(t *testing.T) {
	interp := NewBytecode()
	receiver := runtime.NewSmallInt(7, runtime.IntegerI32)
	target := &runtime.StructInstanceValue{
		Fields: map[string]runtime.Value{},
	}
	target.Fields["capture"] = &runtime.NativeBoundMethodValue{
		Receiver: receiver,
		Method: runtime.NativeFunctionValue{
			Name:       "capture",
			Arity:      1,
			BorrowArgs: true,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 2 {
					t.Fatalf("expected receiver plus one arg, got %d args", len(args))
				}
				if !valuesEqual(args[0], receiver) {
					t.Fatalf("receiver mismatch: got=%#v want=%#v", args[0], receiver)
				}
				return args[1], nil
			},
		},
	}
	interp.GlobalEnvironment().Define("box", target)

	module := ast.Mod([]ast.Statement{
		ast.CallExpr(ast.Member(ast.ID("box"), "capture"), ast.Int(11)),
	}, nil, nil)

	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := runtime.NewSmallInt(11, runtime.IntegerI32)
	if !valuesEqual(got, want) {
		t.Fatalf("unexpected result: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_NativeBoundMethodArgsStayStableWhenBorrowDisabled(t *testing.T) {
	interp := NewBytecode()
	receiver := runtime.NewSmallInt(5, runtime.IntegerI32)
	target := &runtime.StructInstanceValue{
		Fields: map[string]runtime.Value{},
	}
	var captured []runtime.Value
	target.Fields["capture_once"] = &runtime.NativeBoundMethodValue{
		Receiver: receiver,
		Method: runtime.NativeFunctionValue{
			Name:  "capture_once",
			Arity: 1,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if captured == nil {
					captured = args
				}
				return runtime.NilValue{}, nil
			},
		},
	}
	interp.GlobalEnvironment().Define("box", target)

	module := ast.Mod([]ast.Statement{
		ast.CallExpr(ast.Member(ast.ID("box"), "capture_once"), ast.Int(41)),
		ast.CallExpr(ast.Member(ast.ID("box"), "capture_once"), ast.Int(42)),
		ast.Int(0),
	}, nil, nil)

	_ = runBytecodeModuleWithInterpreter(t, interp, module)

	if len(captured) != 2 {
		t.Fatalf("expected receiver plus one captured arg, got %d", len(captured))
	}
	if !valuesEqual(captured[0], receiver) {
		t.Fatalf("expected captured receiver to remain %#v, got %#v", receiver, captured[0])
	}
	intVal, ok := captured[1].(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected captured arg to be integer, got %#v", captured[1])
	}
	if intVal.BigInt().Int64() != 41 {
		t.Fatalf("expected first captured arg to remain 41, got %d", intVal.BigInt().Int64())
	}
}

func TestBytecodeVM_NativeExactCallsSkipInlineProbeStats(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	interp := NewBytecode()
	interp.GlobalEnvironment().Define("add_one", runtime.NativeFunctionValue{
		Name:       "add_one",
		Arity:      1,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				t.Fatalf("expected one arg, got %d", len(args))
			}
			return runtime.NewSmallInt(args[0].(runtime.IntegerValue).BigInt().Int64()+1, runtime.IntegerI32), nil
		},
	})

	receiver := runtime.NewSmallInt(7, runtime.IntegerI32)
	target := &runtime.StructInstanceValue{Fields: map[string]runtime.Value{}}
	target.Fields["capture"] = &runtime.NativeBoundMethodValue{
		Receiver: receiver,
		Method: runtime.NativeFunctionValue{
			Name:       "capture",
			Arity:      1,
			BorrowArgs: true,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 2 {
					t.Fatalf("expected receiver plus one arg, got %d", len(args))
				}
				if !valuesEqual(args[0], receiver) {
					t.Fatalf("receiver mismatch: got=%#v want=%#v", args[0], receiver)
				}
				return args[1], nil
			},
		},
	}
	interp.GlobalEnvironment().Define("box", target)

	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("a"), ast.Call("add_one", ast.Int(41))),
		ast.Assign(ast.ID("b"), ast.CallExpr(ast.Member(ast.ID("box"), "capture"), ast.Int(11))),
		ast.Bin("+", ast.ID("a"), ast.ID("b")),
	}, nil, nil)

	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := runtime.NewSmallInt(53, runtime.IntegerI32)
	if !valuesEqual(got, want) {
		t.Fatalf("unexpected result: got=%#v want=%#v", got, want)
	}

	stats := interp.BytecodeStats()
	if stats.CallNameLookups == 0 {
		t.Fatalf("expected native call-by-name site to execute")
	}
	if stats.InlineCallHits != 0 || stats.InlineCallMisses != 0 {
		t.Fatalf("expected exact native call sites to skip inline probe stats, got hits=%d misses=%d", stats.InlineCallHits, stats.InlineCallMisses)
	}
}
