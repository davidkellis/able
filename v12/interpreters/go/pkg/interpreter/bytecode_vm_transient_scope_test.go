package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_TransientBlockScopeReturnCleanup(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn("pick", []*ast.FunctionParameter{
			ast.Param("n", ast.Ty("i32")),
		}, []ast.Statement{
			ast.IfExpr(
				ast.Bool(true),
				ast.Block(
					ast.Assign(ast.ID("x"), ast.ID("n")),
					ast.Ret(ast.ID("x")),
				),
			),
			ast.Int(0),
		}, ast.Ty("i32"), nil, nil, false, false),
		ast.Bin("+", ast.Call("pick", ast.Int(1)), ast.Call("pick", ast.Int(2))),
	}, nil, nil)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("transient block return cleanup mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_TransientBlockScopeKeepsEscapingLambdaCapture(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Assign(
			ast.ID("captured"),
			ast.IfExpr(
				ast.Bool(true),
				ast.Block(
					ast.Assign(ast.ID("x"), ast.Int(7)),
					ast.Lam(nil, ast.ID("x")),
				),
			),
		),
		ast.IfExpr(
			ast.Bool(true),
			ast.Block(
				ast.Assign(ast.ID("y"), ast.Int(9)),
				ast.ID("y"),
			),
		),
		ast.CallExpr(ast.ID("captured")),
	}, nil, nil)

	ifExpr, ok := module.Body[0].(*ast.AssignmentExpression)
	if !ok {
		t.Fatalf("expected captured assignment at module body[0]")
	}
	if capturedInit, ok := ifExpr.Right.(*ast.IfExpression); ok && capturedInit != nil {
		capturedInit.ElseBody = ast.Block(ast.Lam(nil, ast.Int(0)))
	} else {
		t.Fatalf("expected if-expression initializer for captured lambda")
	}
	if secondIf, ok := module.Body[1].(*ast.IfExpression); ok && secondIf != nil {
		secondIf.ElseBody = ast.Block(ast.Int(0))
	} else {
		t.Fatalf("expected second if-expression at module body[1]")
	}

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("escaping lambda capture mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVMFinishMinimalSelfFastReturnRestoresEnvAndReleasesTransientScopes(t *testing.T) {
	interp := NewBytecode()
	base := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, base)

	vm.slots = vm.acquireSlotFrame(2)
	vm.slots[0] = runtime.NewSmallInt(4, runtime.IntegerI32)
	vm.slots[1] = &runtime.FunctionValue{}
	if !vm.pushSelfFastSlot0CallFrame(11) {
		t.Fatalf("expected compact self-fast frame push to succeed")
	}

	transient := interp.acquireTransientRuntimeScopeEnv(base)
	transient.Define("x", runtime.NewSmallInt(7, runtime.IntegerI32))
	vm.env = transient
	vm.activeTransientScopeEnvs = append(vm.activeTransientScopeEnvs, transient)

	if !vm.finishMinimalSelfFastReturnNoCoerce(runtime.NewSmallInt(9, runtime.IntegerI32)) {
		t.Fatalf("expected compact self-fast return to complete")
	}
	if vm.env != base {
		t.Fatalf("expected env to restore to caller scope")
	}
	if len(vm.activeTransientScopeEnvs) != 0 {
		t.Fatalf("expected transient scope stack to be empty after return, got %d", len(vm.activeTransientScopeEnvs))
	}
}
