package interpreter

import "able/interpreter-go/pkg/runtime"

func finalizeBytecodeProgramMetadata(program *bytecodeProgram) *bytecodeProgram {
	if program == nil {
		return nil
	}
	bytecodeQuickenBinaryCompareJumps(program.instructions)
	instructions := program.instructions
	var followedByPropagation []bool
	hasIntegerConstValidation := false
	var slotConstIntImmTable *bytecodeSlotConstIntImmediateTable
	for idx := range instructions {
		if idx+1 < len(instructions) && instructions[idx+1].op == bytecodeOpPropagation {
			if followedByPropagation == nil {
				followedByPropagation = make([]bool, len(instructions))
			}
			followedByPropagation[idx] = true
		}
		instr := instructions[idx]
		if instr.op == bytecodeOpConst {
			if _, ok := instr.value.(runtime.IntegerValue); ok {
				hasIntegerConstValidation = true
			}
		}
		if imm, ok := bytecodeInstructionImmediateInteger(instr); ok && bytecodeInstructionUsesSlotConstIntegerImmediate(instr) {
			if slotConstIntImmTable == nil {
				slotConstIntImmTable = newBytecodeSlotConstIntImmediateTable(len(instructions))
			}
			slotConstIntImmTable.add(idx, imm)
		}
	}
	program.followedByPropagation = followedByPropagation
	program.integerConstValidationKnown = true
	program.hasIntegerConstValidation = hasIntegerConstValidation
	program.integerConstInstructionCount = len(instructions)
	program.slotConstIntImmTable = slotConstIntImmTable
	program.slotConstIntImmTableKnown = true
	program.slotConstIntImmInstructionCount = len(instructions)
	program.arrayOwnershipMetadata = bytecodeArrayOwnershipMetadataForInstructions(instructions)
	return program
}
