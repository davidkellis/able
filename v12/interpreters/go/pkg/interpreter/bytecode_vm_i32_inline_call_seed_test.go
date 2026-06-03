package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_I32RegisterFramePreseedsInlineCallNameDirectCallee(t *testing.T) {
	interp := NewBytecode()
	calleeDef := ast.Fn(
		"id",
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("i32"))},
		[]ast.Statement{ast.ID("x")},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	calleeProgram, err := interp.lowerFunctionDefinitionBytecode(calleeDef)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if calleeProgram.frameLayout == nil || !calleeProgram.frameLayout.i32RegisterFrame {
		t.Fatalf("expected direct callee to keep i32 register frame layout")
	}

	calleeEnv := runtime.NewEnvironment(nil)
	calleeFn := &runtime.FunctionValue{
		Declaration: calleeDef,
		Closure:     calleeEnv,
	}
	setFunctionBytecodeProgram(calleeFn, calleeProgram)

	callNode := ast.NewFunctionCall(ast.ID("id"), nil, nil, false)
	lookup := bytecodeResolvedIdentifierLookup{
		value: calleeFn,
		env:   calleeEnv,
		owner: calleeEnv,
	}
	entry := bytecodeBuildCallNameCacheEntry("id", lookup, calleeFn, 1, callNode)
	if !entry.inlineDirect {
		t.Fatalf("expected direct inline cache entry")
	}

	callerProgram := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			slotCount:        1,
			slotKinds:        []bytecodeCellKind{bytecodeCellKindI32},
			i32RegisterFrame: true,
		},
	}
	callerSlots := []runtime.Value{bytecodeBoxedIntegerI32Value(9)}
	vm := newBytecodeVM(interp, calleeEnv)
	vm.currentProgram = callerProgram
	vm.slots = callerSlots
	vm.ip = 7
	vm.activateI32RegisterFrame(callerProgram)
	vm.stack = append(vm.stack, bytecodeBoxedIntegerI32Value(41))

	newProg, handled, err := vm.tryInlineCachedCallNameDirectFromStack(&entry, 0, 1, callNode, callerProgram)
	if err != nil {
		t.Fatalf("inline direct call failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected inline direct call to handle cache entry")
	}
	if newProg != calleeProgram {
		t.Fatalf("inline direct target program = %p, want %p", newProg, calleeProgram)
	}
	if vm.i32RegisterProgram != calleeProgram {
		t.Fatalf("callee register frame not preseeded before switch")
	}
	if got, ok := vm.i32RegisterRaw(0); !ok || got != 41 {
		t.Fatalf("callee raw slot 0 = (%d, %t), want (41, true)", got, ok)
	}
	if len(vm.callFrames) != 1 {
		t.Fatalf("call frame count = %d, want 1", len(vm.callFrames))
	}
	if vm.callFrames[0].i32RegisterProgram != callerProgram {
		t.Fatalf("saved caller register frame program = %p, want %p", vm.callFrames[0].i32RegisterProgram, callerProgram)
	}
	if len(vm.callFrames[0].i32Registers) != 1 || !vm.callFrames[0].i32RegisterValid[0] || vm.callFrames[0].i32Registers[0] != 9 {
		t.Fatalf("saved caller raw frame = (%v, %v), want slot0=9 valid", vm.callFrames[0].i32Registers, vm.callFrames[0].i32RegisterValid)
	}

	// Corrupt the boxed slot copy to prove switchRunProgram keeps the preseeded
	// callee raw frame instead of rescanning vm.slots.
	vm.slots[0] = nil
	program := callerProgram
	instructions := callerProgram.instructions
	validatedIntConsts := vm.validatedIntegerConstSlots(callerProgram)
	slotConstIntImmTable := vm.slotConstImmediateTable(callerProgram)
	vm.switchRunProgram(&program, &instructions, &validatedIntConsts, &slotConstIntImmTable, newProg)
	if vm.currentProgram != calleeProgram {
		t.Fatalf("current program = %p, want %p", vm.currentProgram, calleeProgram)
	}
	if got, ok := vm.i32RegisterRaw(0); !ok || got != 41 {
		t.Fatalf("callee raw slot 0 after switch = (%d, %t), want (41, true)", got, ok)
	}

	returnIP, returnProgram, returnSlots, _, _, _, _, _, ok := vm.popCallFrameFields()
	if !ok {
		t.Fatalf("expected saved caller frame")
	}
	if returnIP != 8 {
		t.Fatalf("return ip = %d, want 8", returnIP)
	}
	if returnProgram != callerProgram {
		t.Fatalf("return program = %p, want %p", returnProgram, callerProgram)
	}
	if !sameSlotFrame(returnSlots, callerSlots) {
		t.Fatalf("returned slot frame was not restored to caller")
	}
	if got, ok := vm.i32RegisterRaw(0); !ok || got != 9 {
		t.Fatalf("restored caller raw slot 0 = (%d, %t), want (9, true)", got, ok)
	}
}

func TestBytecodeVM_I32RegisterFramePreseedsInlineCallNameDirectSlotArgsFromCallerSlots(t *testing.T) {
	interp := NewBytecode()
	calleeDef := ast.Fn(
		"id",
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("i32"))},
		[]ast.Statement{ast.ID("x")},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	calleeProgram, err := interp.lowerFunctionDefinitionBytecode(calleeDef)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}

	calleeEnv := runtime.NewEnvironment(nil)
	calleeFn := &runtime.FunctionValue{
		Declaration: calleeDef,
		Closure:     calleeEnv,
	}
	setFunctionBytecodeProgram(calleeFn, calleeProgram)

	callNode := ast.NewFunctionCall(ast.ID("id"), nil, nil, false)
	lookup := bytecodeResolvedIdentifierLookup{
		value: calleeFn,
		env:   calleeEnv,
		owner: calleeEnv,
	}
	entry := bytecodeBuildCallNameCacheEntry("id", lookup, calleeFn, 1, callNode)
	if !entry.inlineDirect {
		t.Fatalf("expected direct inline cache entry")
	}

	callerProgram := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			slotCount:        1,
			slotKinds:        []bytecodeCellKind{bytecodeCellKindI32},
			i32RegisterFrame: true,
		},
	}
	vm := newBytecodeVM(interp, calleeEnv)
	vm.currentProgram = callerProgram
	vm.slots = []runtime.Value{nil}
	vm.ip = 3
	vm.activateI32RegisterFrame(callerProgram)
	if !vm.setI32RegisterRaw(0, 55) {
		t.Fatalf("expected caller raw lane to accept slot 0")
	}
	sentinel := runtime.NewSmallInt(77, runtime.IntegerI32)
	vm.stack = append(vm.stack, sentinel)

	instr := bytecodeInstruction{
		op:       bytecodeOpCallName,
		name:     "id",
		argCount: 1,
		target:   0,
		slotArgs: true,
		node:     callNode,
	}
	newProg, handled, err := vm.tryInlineCachedCallNameDirectFromSlots(&entry, instr, callNode, callerProgram)
	if err != nil {
		t.Fatalf("inline direct slot-arg call failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected inline direct slot-arg helper to handle cache entry")
	}
	if newProg != calleeProgram {
		t.Fatalf("inline direct slot-arg target program = %p, want %p", newProg, calleeProgram)
	}
	if len(vm.stack) != 1 || vm.stack[0] != sentinel {
		t.Fatalf("slot-arg direct inline should leave existing stack prefix untouched: %#v", vm.stack)
	}
	if vm.i32RegisterProgram != calleeProgram {
		t.Fatalf("callee register frame not preseeded from caller slot args")
	}
	if got, ok := vm.i32RegisterRaw(0); !ok || got != 55 {
		t.Fatalf("callee raw slot 0 = (%d, %t), want (55, true)", got, ok)
	}
	if got, ok := bytecodeRawI32Value(vm.slots[0]); !ok || got != 55 {
		t.Fatalf("callee boxed slot 0 = (%d, %t), want (55, true)", got, ok)
	}
	if len(vm.callFrames) != 1 {
		t.Fatalf("call frame count = %d, want 1", len(vm.callFrames))
	}
	if len(vm.callFrames[0].i32Registers) != 1 || !vm.callFrames[0].i32RegisterValid[0] || vm.callFrames[0].i32Registers[0] != 55 {
		t.Fatalf("saved caller raw frame = (%v, %v), want slot0=55 valid", vm.callFrames[0].i32Registers, vm.callFrames[0].i32RegisterValid)
	}
}
