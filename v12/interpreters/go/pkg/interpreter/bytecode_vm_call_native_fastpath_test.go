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

func TestBytecodeVM_NativeBoundMethodExactCallMaterializesComputedIntegerArg(t *testing.T) {
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
				intVal, ok := args[1].(runtime.IntegerValue)
				if !ok {
					t.Fatalf("expected materialized integer arg, got %T (%#v)", args[1], args[1])
				}
				if intVal.BigInt().Int64() != 42 {
					t.Fatalf("expected computed arg 42, got %d", intVal.BigInt().Int64())
				}
				return args[1], nil
			},
		},
	}
	interp.GlobalEnvironment().Define("box", target)

	module := ast.Mod([]ast.Statement{
		ast.CallExpr(
			ast.Member(ast.ID("box"), "capture"),
			ast.Bin("+", ast.Int(40), ast.Int(2)),
		),
	}, nil, nil)

	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := runtime.NewSmallInt(42, runtime.IntegerI32)
	if !valuesEqual(got, want) {
		t.Fatalf("unexpected result: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_NativeCallNameExactCallMaterializesComputedIntegerArg(t *testing.T) {
	interp := NewBytecode()
	interp.GlobalEnvironment().Define("capture", runtime.NativeFunctionValue{
		Name:       "capture",
		Arity:      1,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				t.Fatalf("expected one arg, got %d", len(args))
			}
			intVal, ok := args[0].(runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected materialized integer arg, got %T (%#v)", args[0], args[0])
			}
			if intVal.BigInt().Int64() != 42 {
				t.Fatalf("expected computed arg 42, got %d", intVal.BigInt().Int64())
			}
			return args[0], nil
		},
	})

	module := ast.Mod([]ast.Statement{
		ast.Call("capture", ast.Bin("+", ast.Int(40), ast.Int(2))),
	}, nil, nil)

	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := runtime.NewSmallInt(42, runtime.IntegerI32)
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

func TestBytecodeVM_NativeExactCallSkipContextPassesNilContext(t *testing.T) {
	interp := NewBytecode()
	called := false

	interp.GlobalEnvironment().Define("capture_ctx", runtime.NativeFunctionValue{
		Name:        "capture_ctx",
		Arity:       0,
		SkipContext: true,
		Impl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			called = true
			if ctx != nil {
				t.Fatalf("expected nil context, got %#v", ctx)
			}
			if len(args) != 0 {
				t.Fatalf("expected no args, got %d", len(args))
			}
			return runtime.NewSmallInt(9, runtime.IntegerI32), nil
		},
	})

	module := ast.Mod([]ast.Statement{
		ast.Call("capture_ctx"),
	}, nil, nil)

	got := runBytecodeModuleWithInterpreter(t, interp, module)
	if !called {
		t.Fatalf("expected native impl to be called")
	}
	want := runtime.NewSmallInt(9, runtime.IntegerI32)
	if !valuesEqual(got, want) {
		t.Fatalf("unexpected result: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeExecExactNativeBoundCallBorrowArgsReusesScratchBuffer(t *testing.T) {
	scratch := &nativeBorrowCallArgScratch{}
	receiver := runtime.StringValue{Val: "self"}
	explicitArgs := []runtime.Value{runtime.NewSmallInt(41, runtime.IntegerI32)}
	var firstArgPtr *runtime.Value
	callCount := 0

	native := runtime.NativeFunctionValue{
		Name:        "capture",
		Arity:       1,
		BorrowArgs:  true,
		SkipContext: true,
		Impl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			callCount++
			if ctx != nil {
				t.Fatalf("expected nil context, got %#v", ctx)
			}
			if len(args) != 2 {
				t.Fatalf("expected receiver plus one arg, got %d", len(args))
			}
			if !valuesEqual(args[0], receiver) {
				t.Fatalf("receiver mismatch: got=%#v want=%#v", args[0], receiver)
			}
			if callCount == 1 {
				firstArgPtr = &args[0]
			} else if firstArgPtr != &args[0] {
				t.Fatalf("expected repeated borrowed exact-native calls to reuse scratch backing")
			}
			return args[1], nil
		},
	}

	for _, want := range []int64{41, 42} {
		explicitArgs[0] = runtime.NewSmallInt(want, runtime.IntegerI32)
		got, err := bytecodeExecExactNativeBoundCall(nil, scratch, native, receiver, explicitArgs)
		if err != nil {
			t.Fatalf("unexpected call error: %v", err)
		}
		if !valuesEqual(got, runtime.NewSmallInt(want, runtime.IntegerI32)) {
			t.Fatalf("unexpected result: got=%#v want=%#v", got, runtime.NewSmallInt(want, runtime.IntegerI32))
		}
	}
}

func TestBytecodeExecExactNativeBoundCallIteratorNextBorrowsArgs(t *testing.T) {
	scratch := &nativeBorrowCallArgScratch{}
	step := 0
	iter := runtime.NewIteratorValue(func() (runtime.Value, bool, error) {
		step++
		return runtime.NewSmallInt(int64(step), runtime.IntegerI32), false, nil
	}, nil)
	native := iteratorNextNativeMethod()
	if !native.BorrowArgs {
		t.Fatalf("iterator.next native method should borrow call args")
	}

	got, err := bytecodeExecExactNativeBoundCall(nil, scratch, native, iter, nil)
	if err != nil {
		t.Fatalf("iterator.next exact-native call failed: %v", err)
	}
	want := runtime.NewSmallInt(1, runtime.IntegerI32)
	if !valuesEqual(got, want) {
		t.Fatalf("unexpected iterator.next result: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_NativeBoundMethodExactCallSkipContextPassesNilContext(t *testing.T) {
	interp := NewBytecode()
	called := false
	receiver := runtime.NewSmallInt(7, runtime.IntegerI32)
	target := &runtime.StructInstanceValue{Fields: map[string]runtime.Value{}}
	target.Fields["capture_ctx"] = &runtime.NativeBoundMethodValue{
		Receiver: receiver,
		Method: runtime.NativeFunctionValue{
			Name:        "capture_ctx",
			Arity:       1,
			BorrowArgs:  true,
			SkipContext: true,
			Impl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				called = true
				if ctx != nil {
					t.Fatalf("expected nil context, got %#v", ctx)
				}
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
		ast.CallExpr(ast.Member(ast.ID("box"), "capture_ctx"), ast.Int(11)),
	}, nil, nil)

	got := runBytecodeModuleWithInterpreter(t, interp, module)
	if !called {
		t.Fatalf("expected native impl to be called")
	}
	want := runtime.NewSmallInt(11, runtime.IntegerI32)
	if !valuesEqual(got, want) {
		t.Fatalf("unexpected result: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_NativeBoundMethodBorrowArgsRemainStableAcrossNestedExactCalls(t *testing.T) {
	interp := NewBytecode()
	outerReceiver := runtime.NewSmallInt(7, runtime.IntegerI32)
	innerReceiver := runtime.NewSmallInt(9, runtime.IntegerI32)
	innerMethod := runtime.NativeFunctionValue{
		Name:        "inner",
		Arity:       1,
		BorrowArgs:  true,
		SkipContext: true,
		Impl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if ctx != nil {
				t.Fatalf("expected nil nested context, got %#v", ctx)
			}
			if len(args) != 2 {
				t.Fatalf("expected nested receiver plus one arg, got %d", len(args))
			}
			if !valuesEqual(args[0], innerReceiver) {
				t.Fatalf("nested receiver mismatch: got=%#v want=%#v", args[0], innerReceiver)
			}
			return args[1], nil
		},
	}
	target := &runtime.StructInstanceValue{Fields: map[string]runtime.Value{}}
	target.Fields["outer"] = &runtime.NativeBoundMethodValue{
		Receiver: outerReceiver,
		Method: runtime.NativeFunctionValue{
			Name:        "outer",
			Arity:       1,
			BorrowArgs:  true,
			SkipContext: true,
			Impl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if ctx != nil {
					t.Fatalf("expected nil outer context, got %#v", ctx)
				}
				if len(args) != 2 {
					t.Fatalf("expected outer receiver plus one arg, got %d", len(args))
				}
				if !valuesEqual(args[0], outerReceiver) {
					t.Fatalf("outer receiver mismatch: got=%#v want=%#v", args[0], outerReceiver)
				}
				scratch := interp.acquireNativeBorrowCallArgScratch()
				defer interp.releaseNativeBorrowCallArgScratch(scratch)
				_, err := bytecodeExecExactNativeBoundCall(nil, scratch, innerMethod, innerReceiver, []runtime.Value{runtime.NewSmallInt(99, runtime.IntegerI32)})
				if err != nil {
					return nil, err
				}
				if !valuesEqual(args[0], outerReceiver) {
					t.Fatalf("outer receiver changed across nested call: got=%#v want=%#v", args[0], outerReceiver)
				}
				if !valuesEqual(args[1], runtime.NewSmallInt(41, runtime.IntegerI32)) {
					t.Fatalf("outer arg changed across nested call: got=%#v", args[1])
				}
				return args[1], nil
			},
		},
	}
	interp.GlobalEnvironment().Define("box", target)

	module := ast.Mod([]ast.Statement{
		ast.CallExpr(ast.Member(ast.ID("box"), "outer"), ast.Int(41)),
	}, nil, nil)

	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := runtime.NewSmallInt(41, runtime.IntegerI32)
	if !valuesEqual(got, want) {
		t.Fatalf("unexpected result: got=%#v want=%#v", got, want)
	}
}
