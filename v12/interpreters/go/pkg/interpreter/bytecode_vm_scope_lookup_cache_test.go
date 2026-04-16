package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_CallNameScopeCacheInvalidatesOnLocalRebind(t *testing.T) {
	mainFn := ast.Fn(
		"main",
		nil,
		[]ast.Statement{
			ast.Assign(
				ast.ID("f"),
				ast.Lam([]*ast.FunctionParameter{ast.Param("x", nil)}, ast.ID("x")),
			),
			ast.Assign(ast.ID("first"), ast.Call("f", ast.Int(1))),
			ast.AssignOp(
				ast.AssignmentAssign,
				ast.ID("f"),
				ast.Lam(
					[]*ast.FunctionParameter{ast.Param("x", nil)},
					ast.Bin("+", ast.ID("x"), ast.Int(10)),
				),
			),
			ast.Assign(ast.ID("second"), ast.Call("f", ast.Int(1))),
			ast.ID("second"),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)

	module := ast.Mod([]ast.Statement{
		mainFn,
		ast.Call("main"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode local callname cache invalidation mismatch: got=%#v want=%#v", got, want)
	}
	if intVal, ok := got.(runtime.IntegerValue); !ok || intVal.BigInt().Int64() != 11 {
		t.Fatalf("expected local rebound call result 11, got %#v", got)
	}
}

func TestBytecodeVM_LoadNameScopeCacheInvalidatesOnLocalAssign(t *testing.T) {
	mainFn := ast.Fn(
		"main",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("noop"), ast.Lam(nil, ast.Int(0))),
			ast.Assign(ast.ID("x"), ast.Int(1)),
			ast.Assign(ast.ID("a"), ast.ID("x")),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("x"), ast.Int(2)),
			ast.Assign(ast.ID("b"), ast.ID("x")),
			ast.ID("b"),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)

	module := ast.Mod([]ast.Statement{
		mainFn,
		ast.Call("main"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode local loadname cache invalidation mismatch: got=%#v want=%#v", got, want)
	}
	if intVal, ok := got.(runtime.IntegerValue); !ok || intVal.BigInt().Int64() != 2 {
		t.Fatalf("expected local reassigned load result 2, got %#v", got)
	}
}

func TestBytecodeVM_LoadNameScopeCacheInvalidatesOnCapturedParentAssign(t *testing.T) {
	mainFn := ast.Fn(
		"main",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("x"), ast.Int(1)),
			ast.Assign(ast.ID("read"), ast.Lam(nil, ast.ID("x"))),
			ast.Assign(ast.ID("first"), ast.Call("read")),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("x"), ast.Int(2)),
			ast.Assign(ast.ID("second"), ast.Call("read")),
			ast.ID("second"),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)

	module := ast.Mod([]ast.Statement{
		mainFn,
		ast.Call("main"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode captured-parent loadname cache invalidation mismatch: got=%#v want=%#v", got, want)
	}
	if intVal, ok := got.(runtime.IntegerValue); !ok || intVal.BigInt().Int64() != 2 {
		t.Fatalf("expected captured load result 2 after parent assign, got %#v", got)
	}
}

func TestBytecodeVM_CallNameDotFallbackScopeCacheInvalidatesOnHeadRebind(t *testing.T) {
	structDef := ast.StructDef(
		"S",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "n"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)

	getFn := ast.Fn(
		"get",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		[]ast.Statement{
			ast.Member(ast.ID("self"), "n"),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	methods := ast.Methods(
		ast.Ty("S"),
		[]*ast.FunctionDefinition{getFn},
		nil,
		nil,
	)

	mainFn := ast.Fn(
		"main",
		nil,
		[]ast.Statement{
			ast.Assign(
				ast.ID("s"),
				ast.StructLit([]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(1), "n"),
				}, false, "S", nil, nil),
			),
			ast.Assign(ast.ID("first"), ast.Call("s.get")),
			ast.AssignOp(
				ast.AssignmentAssign,
				ast.ID("s"),
				ast.StructLit([]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(2), "n"),
				}, false, "S", nil, nil),
			),
			ast.Assign(ast.ID("second"), ast.Call("s.get")),
			ast.ID("second"),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)

	module := ast.Mod([]ast.Statement{
		structDef,
		methods,
		mainFn,
		ast.Call("main"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode dotted callname head cache invalidation mismatch: got=%#v want=%#v", got, want)
	}
	if intVal, ok := got.(runtime.IntegerValue); !ok || intVal.BigInt().Int64() != 2 {
		t.Fatalf("expected rebound dotted callname result 2, got %#v", got)
	}
}

func TestBytecodeVM_CallNameScopeCacheInvalidatesOnCapturedParentRebind(t *testing.T) {
	mainFn := ast.Fn(
		"main",
		nil,
		[]ast.Statement{
			ast.Assign(
				ast.ID("f"),
				ast.Lam([]*ast.FunctionParameter{ast.Param("x", nil)}, ast.ID("x")),
			),
			ast.Assign(ast.ID("invoke"), ast.Lam(nil, ast.Call("f", ast.Int(1)))),
			ast.Assign(ast.ID("first"), ast.Call("invoke")),
			ast.AssignOp(
				ast.AssignmentAssign,
				ast.ID("f"),
				ast.Lam(
					[]*ast.FunctionParameter{ast.Param("x", nil)},
					ast.Bin("+", ast.ID("x"), ast.Int(10)),
				),
			),
			ast.Assign(ast.ID("second"), ast.Call("invoke")),
			ast.ID("second"),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)

	module := ast.Mod([]ast.Statement{
		mainFn,
		ast.Call("main"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode captured-parent callname cache invalidation mismatch: got=%#v want=%#v", got, want)
	}
	if intVal, ok := got.(runtime.IntegerValue); !ok || intVal.BigInt().Int64() != 11 {
		t.Fatalf("expected captured call result 11 after parent rebind, got %#v", got)
	}
}

func TestBytecodeVM_CallNameScopeCacheInvalidatesAcrossDispatchKinds(t *testing.T) {
	byteInterp := NewBytecode()
	byteInterp.GlobalEnvironment().Define("add_one", runtime.NativeFunctionValue{
		Name:       "add_one",
		Arity:      1,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				t.Fatalf("expected one arg, got %d", len(args))
			}
			intVal, ok := args[0].(runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected integer arg, got %#v", args[0])
			}
			return runtime.NewSmallInt(intVal.BigInt().Int64()+1, runtime.IntegerI32), nil
		},
	})

	mainFn := ast.Fn(
		"main",
		nil,
		[]ast.Statement{
			ast.Assign(
				ast.ID("f"),
				ast.Lam([]*ast.FunctionParameter{ast.Param("x", nil)}, ast.ID("x")),
			),
			ast.Assign(ast.ID("first"), ast.Call("f", ast.Int(1))),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("f"), ast.ID("add_one")),
			ast.Assign(ast.ID("second"), ast.Call("f", ast.Int(1))),
			ast.ID("second"),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)

	module := ast.Mod([]ast.Statement{
		mainFn,
		ast.Call("main"),
	}, nil, nil)

	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)
	if intVal, ok := got.(runtime.IntegerValue); !ok || intVal.BigInt().Int64() != 2 {
		t.Fatalf("expected callname rebind across dispatch kinds to produce 2, got %#v", got)
	}
}

func TestBytecodeVM_CallNameDotFallbackUsesMemberMethodCache(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	structDef := ast.StructDef(
		"S",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "n"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)

	getFn := ast.Fn(
		"get",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		[]ast.Statement{
			ast.Member(ast.ID("self"), "n"),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	methods := ast.Methods(
		ast.Ty("S"),
		[]*ast.FunctionDefinition{getFn},
		nil,
		nil,
	)

	callGet := ast.Fn(
		"call_get",
		[]*ast.FunctionParameter{
			ast.Param("s", ast.Ty("S")),
		},
		[]ast.Statement{
			ast.Call("s.get"),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	module := ast.Mod([]ast.Statement{
		structDef,
		methods,
		callGet,
		ast.Assign(
			ast.ID("s"),
			ast.StructLit([]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Int(3), "n"),
			}, false, "S", nil, nil),
		),
		ast.Call("call_get", ast.ID("s")),
		ast.Call("call_get", ast.ID("s")),
	}, nil, nil)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	intVal, ok := got.(runtime.IntegerValue)
	if !ok || intVal.BigInt().Int64() != 3 {
		t.Fatalf("expected dotted callname result 3, got %#v", got)
	}

	stats := interp.BytecodeStats()
	if stats.CallNameDotFallback == 0 {
		t.Fatalf("expected dotted callname fallback to execute")
	}
	if stats.MemberMethodCacheMiss == 0 {
		t.Fatalf("expected member method cache miss on first dotted call")
	}
	if stats.MemberMethodCacheHits == 0 {
		t.Fatalf("expected member method cache hit on repeated dotted call")
	}
}
