package interpreter

import "able/interpreter-go/pkg/runtime"

func (vm *bytecodeVM) execPropagation(program **bytecodeProgram, instructions *[]bytecodeInstruction, validatedIntConsts *[]bool, slotConstIntImmTable **bytecodeSlotConstIntImmediateTable, instr *bytecodeInstruction) (bool, error) {
	val, err := vm.pop()
	if err != nil {
		return false, err
	}
	if isNilRuntimeValue(val) {
		val = runtime.NilValue{}
		if vm.hasCallFrames() {
			if err := vm.finishInlineReturn(program, instructions, validatedIntConsts, slotConstIntImmTable, instr, val, bytecodeSimpleTypeCheckUnknown); err != nil {
				return false, err
			}
			return true, nil
		}
		return false, returnSignal{value: val, node: instr.node}
	}
	if errVal, ok := vm.interp.propagationErrorValue(val, vm.env); ok {
		return false, raiseSignal{value: errVal}
	}
	vm.stack = append(vm.stack, val)
	vm.ip++
	return false, nil
}
