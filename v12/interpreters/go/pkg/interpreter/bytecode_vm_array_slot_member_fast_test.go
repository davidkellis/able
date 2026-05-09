package interpreter

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_ArraySlotMemberFastPathDetectsCanonicalKernelMethods(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())

	readKey := bytecodeMemberMethodCacheKey{
		member:        "read_slot",
		preferMethods: true,
		receiverKind:  bytecodeMemberReceiverArray,
	}
	readDef := ast.Fn("read_slot", []*ast.FunctionParameter{
		ast.Param("self", ast.Ty("Self")),
		ast.Param("idx", ast.Ty("i32")),
	}, []ast.Statement{ast.Nil()}, ast.Ty("T"), nil, nil, false, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		readDef: "/tmp/able/v12/kernel/src/kernel.able",
	})
	if got := vm.memberMethodFastPathFor(readKey, &runtime.FunctionValue{Declaration: readDef}); got != bytecodeMemberMethodFastPathArrayReadSlot {
		t.Fatalf("read_slot fast path = %d, want ArrayReadSlot", got)
	}

	readWrongOrigin := ast.Fn("read_slot", nil, []ast.Statement{ast.Nil()}, ast.Ty("T"), nil, nil, false, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		readWrongOrigin: "/tmp/project/kernel.able",
	})
	if got := vm.memberMethodFastPathFor(readKey, &runtime.FunctionValue{Declaration: readWrongOrigin}); got != bytecodeMemberMethodFastPathNone {
		t.Fatalf("custom read_slot should not receive canonical fast path, got %d", got)
	}

	writeKey := bytecodeMemberMethodCacheKey{
		member:        "write_slot",
		preferMethods: true,
		receiverKind:  bytecodeMemberReceiverArray,
	}
	writeDef := ast.Fn("write_slot", []*ast.FunctionParameter{
		ast.Param("self", ast.Ty("Self")),
		ast.Param("idx", ast.Ty("i32")),
		ast.Param("value", ast.Ty("T")),
	}, []ast.Statement{ast.Nil()}, ast.Ty("void"), nil, nil, false, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		writeDef: "/tmp/able/v12/kernel/src/kernel.able",
	})
	if got := vm.memberMethodFastPathFor(writeKey, &runtime.FunctionValue{Declaration: writeDef}); got != bytecodeMemberMethodFastPathArrayWriteSlot {
		t.Fatalf("write_slot fast path = %d, want ArrayWriteSlot", got)
	}

	writeWrongReturn := ast.Fn("write_slot", []*ast.FunctionParameter{
		ast.Param("self", ast.Ty("Self")),
		ast.Param("idx", ast.Ty("i32")),
		ast.Param("value", ast.Ty("T")),
	}, []ast.Statement{ast.Nil()}, ast.Ty("nil"), nil, nil, false, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		writeWrongReturn: "/tmp/able/v12/kernel/src/kernel.able",
	})
	if got := vm.memberMethodFastPathFor(writeKey, &runtime.FunctionValue{Declaration: writeWrongReturn}); got != bytecodeMemberMethodFastPathNone {
		t.Fatalf("write_slot with wrong return type should not receive fast path, got %d", got)
	}
}

func TestBytecodeVM_ArrayReadSlotMemberFastPathSemantics(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	arr := interp.newArrayValue([]runtime.Value{
		runtime.StringValue{Val: "zero"},
		runtime.StringValue{Val: "one"},
	}, 0)

	vm := newBytecodeVM(interp, env)
	vm.stack = []runtime.Value{arr, runtime.NewSmallInt(1, runtime.IntegerI32)}
	_, handled, err := vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathArrayReadSlot,
		bytecodeInstruction{name: "read_slot", argCount: 1},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("array read_slot fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected array read_slot fast path to handle call")
	}
	want := runtime.StringValue{Val: "one"}
	if !valuesEqual(vm.stack[0], want) {
		t.Fatalf("read_slot result = %#v, want %#v", vm.stack[0], want)
	}

	vm = newBytecodeVM(interp, env)
	vm.stack = []runtime.Value{arr, runtime.NewSmallInt(5, runtime.IntegerI32)}
	_, handled, err = vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathArrayReadSlot,
		bytecodeInstruction{name: "read_slot", argCount: 1},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("array read_slot out-of-bounds fast path failed: %v", err)
	}
	if !handled || !isNilRuntimeValue(vm.stack[0]) {
		t.Fatalf("out-of-bounds read_slot result = %#v, want nil", vm.stack[0])
	}

	vm = newBytecodeVM(interp, env)
	vm.stack = []runtime.Value{arr, runtime.NewSmallInt(-1, runtime.IntegerI32)}
	_, handled, err = vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathArrayReadSlot,
		bytecodeInstruction{name: "read_slot", argCount: 1},
		0,
		1,
		nil,
	)
	if !handled {
		t.Fatalf("expected negative read_slot fast path to handle with kernel error")
	}
	if err == nil || !strings.Contains(err.Error(), "array index must be non-negative") {
		t.Fatalf("negative read_slot err = %v, want non-negative index error", err)
	}
}

func TestBytecodeVM_ArrayWriteSlotMemberFastPathSemantics(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	arr := interp.newArrayValue([]runtime.Value{
		runtime.StringValue{Val: "zero"},
	}, 1)

	vm := newBytecodeVM(interp, env)
	vm.stack = []runtime.Value{
		arr,
		runtime.NewSmallInt(0, runtime.IntegerI32),
		runtime.StringValue{Val: "updated"},
	}
	_, handled, err := vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathArrayWriteSlot,
		bytecodeInstruction{name: "write_slot", argCount: 2},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("array write_slot fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected array write_slot fast path to handle call")
	}
	if _, ok := vm.stack[0].(runtime.VoidValue); !ok {
		t.Fatalf("write_slot result = %#v, want void", vm.stack[0])
	}
	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state after write_slot: %v", err)
	}
	if want := (runtime.StringValue{Val: "updated"}); !valuesEqual(state.Values[0], want) {
		t.Fatalf("array write_slot value = %#v, want %#v", state.Values[0], want)
	}

	vm = newBytecodeVM(interp, env)
	vm.stack = []runtime.Value{
		arr,
		runtime.NewSmallInt(3, runtime.IntegerI32),
		runtime.StringValue{Val: "grown"},
	}
	_, handled, err = vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathArrayWriteSlot,
		bytecodeInstruction{name: "write_slot", argCount: 2},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("array write_slot grow fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected array write_slot grow fast path to handle call")
	}
	if len(state.Values) != 4 {
		t.Fatalf("array length after sparse write_slot = %d, want 4", len(state.Values))
	}
	if want := (runtime.StringValue{Val: "grown"}); !valuesEqual(state.Values[3], want) {
		t.Fatalf("array sparse write_slot value = %#v, want %#v", state.Values[3], want)
	}

	vm = newBytecodeVM(interp, env)
	vm.stack = []runtime.Value{
		arr,
		runtime.NewSmallInt(-1, runtime.IntegerI32),
		runtime.StringValue{Val: "bad"},
	}
	_, handled, err = vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathArrayWriteSlot,
		bytecodeInstruction{name: "write_slot", argCount: 2},
		0,
		1,
		nil,
	)
	if !handled {
		t.Fatalf("expected negative write_slot fast path to handle with kernel error")
	}
	if err == nil || !strings.Contains(err.Error(), "array index must be non-negative") {
		t.Fatalf("negative write_slot err = %v, want non-negative index error", err)
	}
}

func TestBytecodeVM_CanonicalArraySlotCallCacheFeedsArraySlotOpcode(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{}
	arr := interp.newArrayValue([]runtime.Value{
		runtime.StringValue{Val: "zero"},
		runtime.StringValue{Val: "one"},
	}, 0)

	readInstr := bytecodeInstruction{op: bytecodeOpCallMemberArraySlot, name: "read_slot", argCount: 1}
	vm.storeCachedCanonicalArraySlotCall(program, 3, readInstr, arr, bytecodeMemberMethodFastPathArrayReadSlot)
	vm.arraySlotCallCache = nil
	vm.ip = 3
	vm.stack = []runtime.Value{arr, runtime.NewSmallInt(1, runtime.IntegerI32)}
	newProg, err := vm.execCallMemberArraySlot(readInstr, program)
	if err != nil {
		t.Fatalf("execCallMemberArraySlot cached read_slot failed: %v", err)
	}
	if newProg != nil {
		t.Fatalf("expected cached read_slot to stay in current program")
	}
	if vm.ip != 4 {
		t.Fatalf("ip after cached read_slot = %d, want 4", vm.ip)
	}
	if want := (runtime.StringValue{Val: "one"}); !valuesEqual(vm.stack[0], want) {
		t.Fatalf("cached read_slot result = %#v, want %#v", vm.stack[0], want)
	}

	writeInstr := bytecodeInstruction{op: bytecodeOpCallMemberArraySlot, name: "write_slot", argCount: 2}
	vm.storeCachedCanonicalArraySlotCall(program, 7, writeInstr, arr, bytecodeMemberMethodFastPathArrayWriteSlot)
	vm.arraySlotCallCache = nil
	vm.ip = 7
	vm.stack = []runtime.Value{
		arr,
		runtime.NewSmallInt(0, runtime.IntegerI32),
		runtime.StringValue{Val: "updated"},
	}
	newProg, err = vm.execCallMemberArraySlot(writeInstr, program)
	if err != nil {
		t.Fatalf("execCallMemberArraySlot cached write_slot failed: %v", err)
	}
	if newProg != nil {
		t.Fatalf("expected cached write_slot to stay in current program")
	}
	if vm.ip != 8 {
		t.Fatalf("ip after cached write_slot = %d, want 8", vm.ip)
	}
	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state after cached write_slot: %v", err)
	}
	if want := (runtime.StringValue{Val: "updated"}); !valuesEqual(state.Values[0], want) {
		t.Fatalf("cached write_slot value = %#v, want %#v", state.Values[0], want)
	}
}

func TestBytecodeVM_LoweringEmitsArraySlotCallMemberOpcode(t *testing.T) {
	safeRead := ast.Member(ast.ID("maybe_arr"), "read_slot")
	safeRead.Safe = true
	module := ast.Mod([]ast.Statement{
		ast.CallExpr(ast.Member(ast.ID("arr"), "read_slot"), ast.Int(0)),
		ast.CallExpr(ast.Member(ast.ID("arr"), "write_slot"), ast.Int(0), ast.Int(1)),
		ast.CallExpr(safeRead, ast.Int(0)),
	}, nil, nil)

	program, err := NewBytecode().lowerModuleToBytecode(module)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}

	var slotCalls int
	var safeRegularCalls int
	for _, instr := range program.instructions {
		if instr.name != "read_slot" && instr.name != "write_slot" {
			continue
		}
		switch instr.op {
		case bytecodeOpCallMemberArraySlot:
			slotCalls++
			if instr.safe {
				t.Fatalf("array slot opcode should not be safe-navigation call: %#v", instr)
			}
		case bytecodeOpCallMember:
			if instr.safe {
				safeRegularCalls++
			}
		}
	}
	if slotCalls != 2 {
		t.Fatalf("array slot opcode count = %d, want 2", slotCalls)
	}
	if safeRegularCalls != 1 {
		t.Fatalf("safe regular read_slot call count = %d, want 1", safeRegularCalls)
	}
}
