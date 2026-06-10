package interpreter

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_TypedIdentifierDeclarationUsesSlotLowering(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn("f", nil, []ast.Statement{
			ast.Assign(ast.TypedP(ast.ID("x"), ast.Ty("i32")), ast.Int(1)),
			ast.Assign(ast.ID("x"), ast.Bin("+", ast.ID("x"), ast.Int(1))),
			ast.ID("x"),
		}, ast.Ty("i32"), nil, nil, false, false),
		ast.Call("f"),
	}, nil, nil)

	byteInterp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)
	intResult, ok := got.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T (%v)", got, got)
	}
	if val, ok := intResult.ToInt64(); !ok || val != 2 {
		t.Fatalf("unexpected result: got=%v want=2", got)
	}

	fnRaw, err := byteInterp.GlobalEnvironment().Get("f")
	if err != nil {
		t.Fatalf("lookup function f: %v", err)
	}
	fn, ok := fnRaw.(*runtime.FunctionValue)
	if !ok || fn == nil {
		t.Fatalf("expected function value for f, got %T", fnRaw)
	}
	prog, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || prog == nil {
		t.Fatalf("expected bytecode program for f")
	}
	if prog.frameLayout == nil {
		t.Fatalf("expected slot-enabled frame layout for typed identifier declaration")
	}
	sawTypedStoreSlotNew := false
	for _, instr := range prog.instructions {
		if instr.op == bytecodeOpAssignPattern {
			t.Fatalf("typed identifier declaration should not lower through AssignPattern")
		}
		if instr.op == bytecodeOpStoreSlotNew && instr.name == "x" && instr.storeTyped {
			sawTypedStoreSlotNew = true
			if got := typeExpressionToString(instr.typeExpr); got != "i32" {
				t.Fatalf("unexpected typed slot annotation: got=%q want=%q", got, "i32")
			}
		}
		if instr.op == bytecodeOpStoreSlotI32 && instr.name == "x" {
			sawTypedStoreSlotNew = true
			if got := prog.frameLayout.slotKinds[instr.target]; got != bytecodeCellKindI32 {
				t.Fatalf("expected typed i32 slot metadata for StoreSlotI32, got %d", got)
			}
		}
	}
	if !sawTypedStoreSlotNew {
		t.Fatalf("expected typed slot declaration opcode metadata for typed identifier assignment")
	}
}

func TestBytecodeVM_TypedIdentifierAssignImplicitDeclarationUsesGuardedSlotLowering(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn("f", nil, []ast.Statement{
			ast.AssignOp(ast.AssignmentAssign, ast.TypedP(ast.ID("x"), ast.Ty("i32")), ast.Int(1)),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("x"), ast.Bin("+", ast.ID("x"), ast.Int(1))),
			ast.ID("x"),
		}, ast.Ty("i32"), nil, nil, false, false),
		ast.Call("f"),
	}, nil, nil)

	byteInterp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)
	intResult, ok := got.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T (%v)", got, got)
	}
	if val, ok := intResult.ToInt64(); !ok || val != 2 {
		t.Fatalf("unexpected result: got=%v want=2", got)
	}

	prog := bytecodeFunctionProgramForTest(t, byteInterp, "f")
	if prog.frameLayout == nil {
		t.Fatalf("expected slot-enabled frame layout for typed identifier assignment")
	}
	sawTypedStore := false
	sawGuardedLoad := false
	for _, instr := range prog.instructions {
		if instr.name != "x" {
			continue
		}
		switch instr.op {
		case bytecodeOpAssignName, bytecodeOpAssignPattern, bytecodeOpLoadName:
			t.Fatalf("implicit typed local should use guarded slot lowering, saw %s", bytecodeOpName(instr.op))
		case bytecodeOpStoreImplicitSlot:
			if instr.storeTyped {
				sawTypedStore = true
			}
		case bytecodeOpLoadImplicitSlot:
			sawGuardedLoad = true
		}
	}
	if !sawTypedStore {
		t.Fatalf("expected typed guarded slot store for implicit assignment")
	}
	if !sawGuardedLoad {
		t.Fatalf("expected guarded slot load for implicit assignment")
	}
}

func TestBytecodeVM_IdentifierAssignImplicitDeclarationUsesGuardedSlotLowering(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn("f", nil, []ast.Statement{
			ast.AssignOp(ast.AssignmentAssign, ast.ID("x"), ast.Int(1)),
			ast.AssignOp(ast.AssignmentAdd, ast.ID("x"), ast.Int(2)),
			ast.ID("x"),
		}, ast.Ty("i32"), nil, nil, false, false),
		ast.Call("f"),
	}, nil, nil)

	byteInterp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)
	intResult, ok := got.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T (%v)", got, got)
	}
	if val, ok := intResult.ToInt64(); !ok || val != 3 {
		t.Fatalf("unexpected result: got=%v want=3", got)
	}

	prog := bytecodeFunctionProgramForTest(t, byteInterp, "f")
	sawStore := false
	sawCompound := false
	sawLoad := false
	for _, instr := range prog.instructions {
		if instr.name != "x" {
			continue
		}
		switch instr.op {
		case bytecodeOpAssignName, bytecodeOpAssignNameCompound, bytecodeOpLoadName:
			t.Fatalf("implicit local should use guarded slot lowering, saw %s", bytecodeOpName(instr.op))
		case bytecodeOpStoreImplicitSlot:
			sawStore = true
		case bytecodeOpCompoundAssignImplicitSlot:
			sawCompound = true
		case bytecodeOpLoadImplicitSlot:
			sawLoad = true
		}
	}
	if !sawStore || !sawCompound || !sawLoad {
		t.Fatalf("expected guarded store/compound/load, got store=%v compound=%v load=%v", sawStore, sawCompound, sawLoad)
	}
}

func TestBytecodeVM_TypedIdentifierAssignPreservesFutureOuterBinding(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn("f", nil, []ast.Statement{
			ast.AssignOp(ast.AssignmentAssign, ast.TypedP(ast.ID("x"), ast.Ty("i32")), ast.Int(2)),
			ast.ID("x"),
		}, ast.Ty("i32"), nil, nil, false, false),
		ast.AssignOp(ast.AssignmentAssign, ast.ID("x"), ast.Int(1)),
		ast.AssignOp(ast.AssignmentAssign, ast.ID("ignored"), ast.Call("f")),
		ast.ID("x"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	byteInterp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode result mismatch: got=%#v want=%#v", got, want)
	}
	intResult, ok := got.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T (%v)", got, got)
	}
	if val, ok := intResult.ToInt64(); !ok || val != 2 {
		t.Fatalf("unexpected final x: got=%v want=2", got)
	}

	prog := bytecodeFunctionProgramForTest(t, byteInterp, "f")
	sawGuardedStore := false
	sawGuardedLoad := false
	for _, instr := range prog.instructions {
		if instr.name == "x" && instr.op == bytecodeOpStoreImplicitSlot {
			sawGuardedStore = true
		}
		if instr.name == "x" && instr.op == bytecodeOpLoadImplicitSlot {
			sawGuardedLoad = true
		}
	}
	if !sawGuardedStore || !sawGuardedLoad {
		t.Fatalf("expected guarded store/load for future outer binding, got store=%v load=%v", sawGuardedStore, sawGuardedLoad)
	}
}

func TestBytecodeVM_TypedIdentifierAssignPreservesExistingOuterBinding(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.AssignOp(ast.AssignmentAssign, ast.ID("x"), ast.Int(1)),
		ast.Fn("f", nil, []ast.Statement{
			ast.AssignOp(ast.AssignmentAssign, ast.TypedP(ast.ID("x"), ast.Ty("i32")), ast.Int(2)),
			ast.ID("x"),
		}, ast.Ty("i32"), nil, nil, false, false),
		ast.AssignOp(ast.AssignmentAssign, ast.ID("ignored"), ast.Call("f")),
		ast.ID("x"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	byteInterp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode result mismatch: got=%#v want=%#v", got, want)
	}
	intResult, ok := got.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T (%v)", got, got)
	}
	if val, ok := intResult.ToInt64(); !ok || val != 2 {
		t.Fatalf("unexpected final x: got=%v want=2", got)
	}

	prog := bytecodeFunctionProgramForTest(t, byteInterp, "f")
	for _, instr := range prog.instructions {
		if instr.name == "x" && (instr.op == bytecodeOpStoreImplicitSlot || instr.op == bytecodeOpLoadImplicitSlot) {
			t.Fatalf("existing outer binding should not lower to guarded local slot, saw %s", bytecodeOpName(instr.op))
		}
	}
}

func TestBytecodeVM_UntypedSlotStoreDoesNotCacheTypedMetadata(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn("f", nil, []ast.Statement{
			ast.Assign(ast.ID("x"), ast.Int(1)),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("x"), ast.Bin("+", ast.ID("x"), ast.Int(1))),
			ast.ID("x"),
		}, ast.Ty("i32"), nil, nil, false, false),
		ast.Call("f"),
	}, nil, nil)

	byteInterp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)
	intResult, ok := got.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T (%v)", got, got)
	}
	if val, ok := intResult.ToInt64(); !ok || val != 2 {
		t.Fatalf("unexpected result: got=%v want=2", got)
	}

	fnRaw, err := byteInterp.GlobalEnvironment().Get("f")
	if err != nil {
		t.Fatalf("lookup function f: %v", err)
	}
	fn, ok := fnRaw.(*runtime.FunctionValue)
	if !ok || fn == nil {
		t.Fatalf("expected function value for f, got %T", fnRaw)
	}
	prog, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || prog == nil {
		t.Fatalf("expected bytecode program for f")
	}
	foundUntypedStore := false
	for _, instr := range prog.instructions {
		if (instr.op == bytecodeOpStoreSlot || instr.op == bytecodeOpStoreSlotNew) && instr.name == "x" {
			foundUntypedStore = true
			if instr.storeTyped {
				t.Fatalf("untyped slot store should not carry typed slot metadata")
			}
		}
	}
	if !foundUntypedStore {
		t.Fatalf("expected untyped slot store in lowered bytecode")
	}
}

func bytecodeFunctionProgramForTest(t *testing.T, interp *Interpreter, name string) *bytecodeProgram {
	t.Helper()
	fnRaw, err := interp.GlobalEnvironment().Get(name)
	if err != nil {
		t.Fatalf("lookup function %s: %v", name, err)
	}
	fn, ok := fnRaw.(*runtime.FunctionValue)
	if !ok || fn == nil {
		t.Fatalf("expected function value for %s, got %T", name, fnRaw)
	}
	prog, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || prog == nil {
		t.Fatalf("expected bytecode program for %s", name)
	}
	return prog
}

func TestBytecodeVM_ExecStoreSlotUntypedFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = make([]runtime.Value, 1)
	vm.stack = append(vm.stack, runtime.NewSmallInt(7, runtime.IntegerI32))

	instr := &bytecodeInstruction{
		op:         bytecodeOpStoreSlot,
		target:     0,
		name:       "x",
		storeTyped: false,
	}

	if err := vm.execStoreSlot(instr); err != nil {
		t.Fatalf("execStoreSlot returned error: %v", err)
	}
	gotSlot := vm.slots[0]
	kind, raw, ok := bytecodeRawIntegerValueInfo(gotSlot)
	if !ok || kind != runtime.IntegerI32 || raw != 7 {
		t.Fatalf("unexpected raw slot value: got=%#v", gotSlot)
	}
	if boxed := bytecodeMaterializeRawValue(gotSlot); !valuesEqual(boxed, runtime.NewSmallInt(7, runtime.IntegerI32)) {
		t.Fatalf("unexpected materialized slot value: got=%#v", boxed)
	}
	if len(vm.stack) != 1 || !valuesEqual(vm.stack[0], runtime.NewSmallInt(7, runtime.IntegerI32)) {
		t.Fatalf("unexpected stack state after untyped store: %#v", vm.stack)
	}
	if vm.ip != 1 {
		t.Fatalf("expected ip to advance to 1, got %d", vm.ip)
	}
}

func TestBytecodeVM_TypedIdentifierMismatchReturnsErrorValue(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn("f", nil, []ast.Statement{
			ast.Assign(ast.TypedP(ast.ID("x"), ast.Ty("i32")), ast.Str("oops")),
		}, nil, nil, nil, false, false),
		ast.Call("f"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	wantErr, ok := want.(runtime.ErrorValue)
	if !ok {
		t.Fatalf("expected tree-walk result to be error value, got %T (%v)", want, want)
	}
	errVal, ok := got.(runtime.ErrorValue)
	if !ok {
		t.Fatalf("expected error value result, got %T (%v)", got, got)
	}
	if errVal.Message != wantErr.Message {
		t.Fatalf("typed mismatch message mismatch: got=%q want=%q", errVal.Message, wantErr.Message)
	}
	if !strings.Contains(errVal.Message, "Typed pattern mismatch in assignment") {
		t.Fatalf("unexpected typed mismatch message content: %q", errVal.Message)
	}
}
