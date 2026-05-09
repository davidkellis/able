package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_CanonicalArrayNewCallCacheFeedsArrayNewOpcode(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())

	arrayDef := ast.StructDef("Array", nil, ast.StructKindNamed, nil, nil, false)
	interp.SetNodeOrigins(map[ast.Node]string{
		arrayDef: "/tmp/able/v12/kernel/src/kernel.able",
	})
	receiver := &runtime.StructDefinitionValue{Node: arrayDef}
	program := &bytecodeProgram{}
	instr := bytecodeInstruction{op: bytecodeOpCallMemberArrayNew, name: "new", argCount: 0}

	vm.storeCachedCanonicalArrayNewCall(program, 11, instr, receiver)
	vm.arrayNewCallCache = nil
	if !vm.lookupCachedCanonicalArrayNewCall(program, 11, instr, receiver) {
		t.Fatalf("expected hot canonical Array.new call cache hit")
	}

	vm.ip = 11
	vm.stack = []runtime.Value{receiver}
	newProg, err := vm.execCallMemberArrayNew(instr, program)
	if err != nil {
		t.Fatalf("execCallMemberArrayNew cached canonical Array.new failed: %v", err)
	}
	if newProg != nil {
		t.Fatalf("execCallMemberArrayNew returned new program on completed fast call")
	}
	arr, ok := vm.stack[0].(*runtime.ArrayValue)
	if !ok || arr == nil {
		t.Fatalf("Array.new result = %#v, want ArrayValue", vm.stack[0])
	}
	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure Array.new result state: %v", err)
	}
	if len(state.Values) != 0 || state.Capacity != 0 {
		t.Fatalf("Array.new state = len %d cap %d, want empty", len(state.Values), state.Capacity)
	}
	if vm.ip != 12 {
		t.Fatalf("ip after cached Array.new = %d, want 12", vm.ip)
	}
}

func TestBytecodeVM_LoweringEmitsArrayNewCallMemberOpcode(t *testing.T) {
	safeNew := ast.Member(ast.ID("maybe_array_type"), "new")
	safeNew.Safe = true
	module := ast.Mod([]ast.Statement{
		ast.CallExpr(ast.Member(ast.ID("Array"), "new")),
		ast.CallExpr(ast.Member(ast.ID("Array"), "new"), ast.Int(8)),
		ast.CallExpr(safeNew),
	}, nil, nil)

	program, err := NewBytecode().lowerModuleToBytecode(module)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}

	arrayNewCount := 0
	regularNewCount := 0
	safeNewCount := 0
	for _, instr := range program.instructions {
		if instr.name != "new" {
			continue
		}
		switch instr.op {
		case bytecodeOpCallMemberArrayNew:
			arrayNewCount++
			if instr.argCount != 0 || instr.safe {
				t.Fatalf("unexpected Array.new opcode instruction: %#v", instr)
			}
		case bytecodeOpCallMember:
			regularNewCount++
			if instr.safe {
				safeNewCount++
			}
		}
	}
	if arrayNewCount != 1 {
		t.Fatalf("Array.new opcode count = %d, want 1", arrayNewCount)
	}
	if regularNewCount != 2 {
		t.Fatalf("regular new CallMember count = %d, want 2", regularNewCount)
	}
	if safeNewCount != 1 {
		t.Fatalf("safe new CallMember count = %d, want 1", safeNewCount)
	}
}
