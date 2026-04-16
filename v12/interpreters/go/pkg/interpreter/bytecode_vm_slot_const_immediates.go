package interpreter

import "able/interpreter-go/pkg/runtime"

type bytecodeSlotConstIntImmediateTable struct {
	instructionCount int
	hasSingle        bool
	singleIP         int
	singleValue      runtime.IntegerValue
	index            map[int]runtime.IntegerValue
}

func (vm *bytecodeVM) slotConstImmediateTable(program *bytecodeProgram) *bytecodeSlotConstIntImmediateTable {
	if vm == nil || program == nil {
		return nil
	}
	if vm.slotConstIntImm == nil {
		vm.slotConstIntImm = make(map[*bytecodeProgram]*bytecodeSlotConstIntImmediateTable)
	}
	table, ok := vm.slotConstIntImm[program]
	if ok && table != nil && table.instructionCount == len(program.instructions) {
		return table
	}
	table = &bytecodeSlotConstIntImmediateTable{
		instructionCount: len(program.instructions),
		singleIP:         -1,
	}
	for idx, instr := range program.instructions {
		switch instr.op {
		case bytecodeOpBinaryIntAddSlotConst, bytecodeOpBinaryIntSubSlotConst, bytecodeOpBinaryIntLessEqualSlotConst, bytecodeOpCallSelfIntSubSlotConst:
			if imm, ok := bytecodeInstructionImmediateInteger(instr); ok {
				if !table.hasSingle && table.index == nil {
					table.hasSingle = true
					table.singleIP = idx
					table.singleValue = imm
					continue
				}
				if table.index == nil {
					table.index = make(map[int]runtime.IntegerValue, 4)
					if table.hasSingle && table.singleIP >= 0 {
						table.index[table.singleIP] = table.singleValue
					}
				}
				table.index[idx] = imm
			}
		}
	}
	vm.slotConstIntImm[program] = table
	return table
}

func bytecodeInstructionImmediateInteger(instr bytecodeInstruction) (runtime.IntegerValue, bool) {
	if instr.hasIntImmediate {
		return instr.intImmediate, true
	}
	return bytecodeImmediateIntegerValue(instr.value)
}

func bytecodeSlotConstImmediateAt(instr bytecodeInstruction, ip int, table *bytecodeSlotConstIntImmediateTable) (runtime.IntegerValue, bool) {
	switch instr.op {
	case bytecodeOpBinaryIntAddSlotConst, bytecodeOpBinaryIntSubSlotConst, bytecodeOpBinaryIntLessEqualSlotConst, bytecodeOpCallSelfIntSubSlotConst:
	default:
		return runtime.IntegerValue{}, false
	}
	if value, ok := bytecodeSlotConstImmediateAtIP(ip, table); ok {
		return value, true
	}
	imm, ok := bytecodeInstructionImmediateInteger(instr)
	return imm, ok
}

func bytecodeSlotConstImmediateAtIP(ip int, table *bytecodeSlotConstIntImmediateTable) (runtime.IntegerValue, bool) {
	if table != nil {
		if table.index != nil {
			if value, ok := table.index[ip]; ok {
				return value, true
			}
		} else if table.hasSingle && table.singleIP == ip {
			return table.singleValue, true
		}
	}
	return runtime.IntegerValue{}, false
}
