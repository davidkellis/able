package interpreter

import "fmt"

func (vm *bytecodeVM) execJumpOpcode(instr *bytecodeInstruction, slotConstIntImmTable *bytecodeSlotConstIntImmediateTable, program *bytecodeProgram) error {
	if instr == nil {
		return fmt.Errorf("bytecode jump missing instruction")
	}
	switch instr.op {
	case bytecodeOpJumpIfIntLessEqualSlotConstFalse:
		return vm.execJumpIfIntLessEqualSlotConstFalse(instr, slotConstIntImmTable)
	case bytecodeOpJumpIfIntCompareSlotConstFalse:
		return vm.execJumpIfIntCompareSlotConstFalse(instr, slotConstIntImmTable)
	case bytecodeOpJumpIfArrayReadSlotCompareSlotFalse:
		return vm.execJumpIfArrayReadSlotCompareSlotFalse(instr, program)
	case bytecodeOpJumpIfIntCompareSlotFalse:
		return vm.execJumpIfIntCompareSlotFalse(instr)
	default:
		return fmt.Errorf("unsupported bytecode jump opcode %d", instr.op)
	}
}
