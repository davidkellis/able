package interpreter

import (
	"fmt"
	"math"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type bytecodeI32RecurrenceKernel struct {
	baseLimit   int64
	baseReturn  int32
	firstSub    int64
	secondSub   int64
	overflowAST ast.Node
}

func bytecodeDetectI32RecurrenceKernel(program *bytecodeProgram) *bytecodeI32RecurrenceKernel {
	if program == nil || program.frameLayout == nil {
		return nil
	}
	layout := program.frameLayout
	if !layout.selfCallOneArgFast || layout.paramSlots != 1 || layout.slotCount != 2 || layout.selfCallSlot != 1 {
		return nil
	}
	if layout.usesImplicitMember || layout.needsEnvScopes || layout.returnSimpleCheck != bytecodeSimpleTypeCheckI32 {
		return nil
	}
	if len(layout.slotKinds) < 2 || layout.slotKinds[0] != bytecodeCellKindI32 {
		return nil
	}
	instructions := program.instructions
	if len(instructions) != 5 {
		return nil
	}
	base := instructions[0]
	firstCall := instructions[1]
	secondCall := instructions[2]
	retAdd := instructions[3]
	finalRet := instructions[4]
	if base.op != bytecodeOpReturnConstIfIntLessEqualSlotConst ||
		firstCall.op != bytecodeOpCallSelfIntSubSlotConst ||
		secondCall.op != bytecodeOpCallSelfIntSubSlotConst ||
		retAdd.op != bytecodeOpReturnBinaryIntAddI32 ||
		finalRet.op != bytecodeOpReturn {
		return nil
	}
	if base.argCount != 0 || base.target != -1 || !base.hasIntRaw {
		return nil
	}
	if !bytecodeRecurrenceCallShape(firstCall, layout.selfCallSlot) ||
		!bytecodeRecurrenceCallShape(secondCall, layout.selfCallSlot) {
		return nil
	}
	baseReturn, ok := bytecodeRawI32Value(base.value)
	if !ok {
		return nil
	}
	return &bytecodeI32RecurrenceKernel{
		baseLimit:   base.intImmediateRaw,
		baseReturn:  baseReturn,
		firstSub:    firstCall.intImmediateRaw,
		secondSub:   secondCall.intImmediateRaw,
		overflowAST: retAdd.node,
	}
}

func bytecodeRecurrenceCallShape(instr bytecodeInstruction, selfSlot int) bool {
	return instr.target == selfSlot &&
		instr.argCount == 0 &&
		instr.hasIntRaw &&
		instr.intImmediateRaw > 0 &&
		instr.intImmediateRaw <= math.MaxInt32
}

func (k *bytecodeI32RecurrenceKernel) eval(n int32) (int32, bool) {
	if k == nil {
		return 0, true
	}
	if int64(n) <= k.baseLimit {
		return k.baseReturn, false
	}
	firstArg := int64(n) - k.firstSub
	if firstArg < math.MinInt32 || firstArg > math.MaxInt32 {
		return 0, true
	}
	left, overflow := k.eval(int32(firstArg))
	if overflow {
		return 0, true
	}
	secondArg := int64(n) - k.secondSub
	if secondArg < math.MinInt32 || secondArg > math.MaxInt32 {
		return 0, true
	}
	right, overflow := k.eval(int32(secondArg))
	if overflow {
		return 0, true
	}
	sum := int64(left) + int64(right)
	if sum < math.MinInt32 || sum > math.MaxInt32 {
		return 0, true
	}
	return int32(sum), false
}

func (vm *bytecodeVM) tryExecI32RecurrenceProgram(program **bytecodeProgram, instructions *[]bytecodeInstruction, validatedIntConsts *[]bool, slotConstIntImmTable **bytecodeSlotConstIntImmediateTable, resume bool) (bool, runtime.Value, error) {
	if resume || vm == nil || vm.ip != 0 || vm.interp == nil || vm.interp.bytecodeStatsEnabled || program == nil || *program == nil {
		return false, nil, nil
	}
	activeProgram := *program
	kernel := activeProgram.i32RecurrenceKernel
	if kernel == nil {
		return false, nil, nil
	}
	if len(vm.slots) == 0 {
		return true, nil, fmt.Errorf("bytecode slot out of range")
	}
	raw, ok := bytecodeRawI32Value(vm.slots[0])
	if !ok {
		return false, nil, nil
	}
	result, overflow := kernel.eval(raw)
	if overflow {
		err := vm.interp.wrapStandardRuntimeError(newOverflowError("integer overflow"))
		if kernel.overflowAST != nil {
			err = vm.interp.attachRuntimeContext(err, kernel.overflowAST, vm.interp.stateFromEnv(vm.env))
		}
		return true, nil, err
	}
	value := bytecodeBoxedIntegerI32Value(int64(result))
	if vm.hasCallFrames() {
		err := vm.finishInlineReturn(program, instructions, validatedIntConsts, slotConstIntImmTable, nil, value, bytecodeSimpleTypeCheckI32)
		return true, nil, err
	}
	return true, value, nil
}
