package interpreter

import "able/interpreter-go/pkg/runtime"

type bytecodeSlotConstIntImmediateTable struct {
	instructionCount int
	hasSingle        bool
	singleIP         int
	singleValue      runtime.IntegerValue
	index            map[int]runtime.IntegerValue
}

func newBytecodeSlotConstIntImmediateTable(instructionCount int) *bytecodeSlotConstIntImmediateTable {
	return &bytecodeSlotConstIntImmediateTable{
		instructionCount: instructionCount,
		singleIP:         -1,
	}
}

func (table *bytecodeSlotConstIntImmediateTable) add(ip int, value runtime.IntegerValue) {
	if table == nil {
		return
	}
	if !table.hasSingle && table.index == nil {
		table.hasSingle = true
		table.singleIP = ip
		table.singleValue = value
		return
	}
	if table.index == nil {
		table.index = make(map[int]runtime.IntegerValue, 4)
		if table.hasSingle && table.singleIP >= 0 {
			table.index[table.singleIP] = table.singleValue
		}
	}
	table.index[ip] = value
}

func (vm *bytecodeVM) slotConstImmediateTable(program *bytecodeProgram) *bytecodeSlotConstIntImmediateTable {
	if vm == nil || program == nil {
		return nil
	}
	instructionCount := len(program.instructions)
	if program.slotConstIntImmTableKnown && program.slotConstIntImmInstructionCount == instructionCount {
		return program.slotConstIntImmTable
	}
	for _, cached := range vm.slotConstIntImmDirect {
		if cached.program != program || cached.table == nil || cached.table.instructionCount != instructionCount {
			continue
		}
		return cached.table
	}
	if vm.slotConstIntImm == nil {
		vm.slotConstIntImm = make(map[*bytecodeProgram]*bytecodeSlotConstIntImmediateTable)
	}
	table, ok := vm.slotConstIntImm[program]
	if ok && table != nil && table.instructionCount == instructionCount {
		vm.cacheSlotConstImmediateTable(program, table)
		return table
	}
	table = newBytecodeSlotConstIntImmediateTable(instructionCount)
	for idx, instr := range program.instructions {
		if imm, ok := bytecodeInstructionImmediateInteger(instr); ok && bytecodeInstructionUsesSlotConstIntegerImmediate(instr) {
			table.add(idx, imm)
		}
	}
	vm.slotConstIntImm[program] = table
	vm.cacheSlotConstImmediateTable(program, table)
	return table
}

func (vm *bytecodeVM) cacheSlotConstImmediateTable(program *bytecodeProgram, table *bytecodeSlotConstIntImmediateTable) {
	if vm == nil || program == nil || table == nil {
		return
	}
	entry := bytecodeSlotConstIntImmediateDirectCacheEntry{program: program, table: table}
	entries := vm.slotConstIntImmDirect[:]
	for idx := range entries {
		if entries[idx].program != program {
			continue
		}
		entries[idx] = entry
		return
	}
	insert := int(vm.slotConstIntImmDirectNext % bytecodeProgramMetadataDirectCacheSize)
	entries[insert] = entry
	vm.slotConstIntImmDirectNext = (vm.slotConstIntImmDirectNext + 1) % bytecodeProgramMetadataDirectCacheSize
}

func bytecodeInstructionImmediateInteger(instr bytecodeInstruction) (runtime.IntegerValue, bool) {
	if instr.hasIntImmediate {
		return instr.intImmediate, true
	}
	return bytecodeImmediateIntegerValue(instr.value)
}

func bytecodeInstructionUsesSlotConstIntegerImmediate(instr bytecodeInstruction) bool {
	switch instr.op {
	case bytecodeOpBinaryIntAddSlotConst, bytecodeOpBinaryIntSubSlotConst, bytecodeOpBinaryIntMulSlotConst, bytecodeOpBinaryIntModSlotConst, bytecodeOpBinaryIntLessEqualSlotConst, bytecodeOpBinaryIntCompareSlotConst, bytecodeOpStoreSlotBinaryIntSlotConst, bytecodeOpCallSelfIntSubSlotConst, bytecodeOpJumpIfIntLessEqualSlotConstFalse, bytecodeOpJumpIfIntCompareSlotConstFalse, bytecodeOpReturnIfIntLessEqualSlotConst, bytecodeOpReturnConstIfIntLessEqualSlotConst:
		return true
	default:
		return false
	}
}

func bytecodeSlotConstImmediateAt(instr bytecodeInstruction, ip int, table *bytecodeSlotConstIntImmediateTable) (runtime.IntegerValue, bool) {
	if !bytecodeInstructionUsesSlotConstIntegerImmediate(instr) {
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
