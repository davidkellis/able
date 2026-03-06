package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_SelfCallSlotAvoidsCallNameLookups(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")
	fib := ast.Fn(
		"fib",
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin("<=", ast.ID("n"), ast.Int(2)),
				ast.Block(ast.Ret(ast.Int(1))),
			),
			ast.Bin(
				"+",
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(1))),
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(2))),
			),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	module := ast.Mod([]ast.Statement{fib}, nil, nil)

	byteInterp := NewBytecode()
	runBytecodeModuleWithInterpreter(t, byteInterp, module)
	byteInterp.ResetBytecodeStats()
	fibValue, err := byteInterp.GlobalEnvironment().Get("fib")
	if err != nil {
		t.Fatalf("lookup fib: %v", err)
	}
	arg := runtime.IntegerValue{Val: big.NewInt(20), TypeSuffix: runtime.IntegerI32}
	got, err := byteInterp.CallFunction(fibValue, []runtime.Value{arg})
	if err != nil {
		t.Fatalf("bytecode fib call failed: %v", err)
	}

	treeInterp := New()
	mustEvalModule(t, treeInterp, module)
	treeFib, err := treeInterp.GlobalEnvironment().Get("fib")
	if err != nil {
		t.Fatalf("tree lookup fib: %v", err)
	}
	want, err := treeInterp.CallFunction(treeFib, []runtime.Value{arg})
	if err != nil {
		t.Fatalf("tree fib call failed: %v", err)
	}
	if !valuesEqual(got, want) {
		t.Fatalf("fib result mismatch: got=%#v want=%#v", got, want)
	}

	stats := byteInterp.BytecodeStats()
	if stats.CallNameLookups != 0 {
		t.Fatalf("expected self-recursive bytecode to avoid CallName lookups, got %d", stats.CallNameLookups)
	}
	callSelfOps := stats.OpCounts[int(bytecodeOpCallSelf)] + stats.OpCounts[int(bytecodeOpCallSelfIntSubSlotConst)]
	if callSelfOps == 0 {
		t.Fatalf("expected recursive self calls to execute self-call opcodes")
	}
	if stats.InlineCallHits == 0 {
		t.Fatalf("expected inline call hits for recursive self calls")
	}
}

func TestBytecodeVM_SelfCallSlotDisabledWhenFunctionNameAssigned(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")
	g := ast.Fn(
		"g",
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
		[]ast.Statement{ast.ID("n")},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	f := ast.Fn(
		"f",
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin("<=", ast.ID("n"), ast.Int(0)),
				ast.Block(ast.Ret(ast.Int(0))),
			),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("f"), ast.ID("g")),
			ast.Call("f", ast.Bin("-", ast.ID("n"), ast.Int(1))),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	module := ast.Mod([]ast.Statement{g, f}, nil, nil)

	byteInterp := NewBytecode()
	runBytecodeModuleWithInterpreter(t, byteInterp, module)
	byteInterp.ResetBytecodeStats()
	fValue, err := byteInterp.GlobalEnvironment().Get("f")
	if err != nil {
		t.Fatalf("lookup f: %v", err)
	}
	arg := runtime.IntegerValue{Val: big.NewInt(2), TypeSuffix: runtime.IntegerI32}
	got, err := byteInterp.CallFunction(fValue, []runtime.Value{arg})
	if err != nil {
		t.Fatalf("bytecode f call failed: %v", err)
	}

	treeInterp := New()
	mustEvalModule(t, treeInterp, module)
	treeF, err := treeInterp.GlobalEnvironment().Get("f")
	if err != nil {
		t.Fatalf("tree lookup f: %v", err)
	}
	want, err := treeInterp.CallFunction(treeF, []runtime.Value{arg})
	if err != nil {
		t.Fatalf("tree f call failed: %v", err)
	}
	if !valuesEqual(got, want) {
		t.Fatalf("f reassignment result mismatch: got=%#v want=%#v", got, want)
	}

	stats := byteInterp.BytecodeStats()
	if stats.CallNameLookups == 0 {
		t.Fatalf("expected CallName lookups when function self name is assigned")
	}
}
