package semanticabi

import "fmt"

type instructionRange struct {
	start int
	end   int
}

func validateFlowFunction(image *Image, functionIndex int, function FunctionRecord) error {
	ranges, err := flowBlockRanges(image, functionIndex, function)
	if err != nil {
		return err
	}
	successors := make([][]uint32, function.BlockCount)
	predecessors := make([][]uint32, function.BlockCount)
	for block, current := range ranges {
		last := image.Instructions[current.end-1]
		successors[block], err = instructionSuccessors(last)
		if err != nil {
			return fmt.Errorf("semanticabi: function %d block %d: %w", functionIndex, block, err)
		}
		for _, successor := range successors[block] {
			predecessors[successor] = append(predecessors[successor], uint32(block))
		}
	}
	if err := validateReachableBlocks(functionIndex, successors); err != nil {
		return err
	}
	for block, current := range ranges {
		for instructionIndex := current.start; instructionIndex < current.end; instructionIndex++ {
			if err := validateFlowInstructionTypes(image, functionIndex, function, instructionIndex); err != nil {
				return err
			}
		}
		if block != 0 && len(predecessors[block]) == 0 {
			return fmt.Errorf("semanticabi: function %d block %d has no predecessor", functionIndex, block)
		}
	}
	return validateDefiniteAssignment(image, functionIndex, function, ranges, predecessors)
}

func flowBlockRanges(image *Image, functionIndex int, function FunctionRecord) ([]instructionRange, error) {
	if function.InstructionCount == 0 {
		return nil, fmt.Errorf("semanticabi: flow function %d has no instructions", functionIndex)
	}
	ranges := make([]instructionRange, function.BlockCount)
	start := int(function.InstructionStart)
	end := start + int(function.InstructionCount)
	expectedBlock := uint32(0)
	blockStart := start
	terminated := false
	for index := start; index < end; index++ {
		instruction := image.Instructions[index]
		if instruction.Block != expectedBlock {
			if instruction.Block != expectedBlock+1 {
				return nil, fmt.Errorf("semanticabi: flow function %d instruction %d skips from block %d to %d", functionIndex, index, expectedBlock, instruction.Block)
			}
			if !terminated {
				return nil, fmt.Errorf("semanticabi: flow function %d block %d lacks terminator", functionIndex, expectedBlock)
			}
			ranges[expectedBlock] = instructionRange{start: blockStart, end: index}
			expectedBlock = instruction.Block
			blockStart = index
			terminated = false
		} else if terminated {
			return nil, fmt.Errorf("semanticabi: flow function %d block %d has instruction after terminator", functionIndex, expectedBlock)
		}
		descriptor, _ := OpByCode(instruction.Opcode)
		terminated = descriptor.Terminator
	}
	if !terminated {
		return nil, fmt.Errorf("semanticabi: flow function %d block %d lacks terminator", functionIndex, expectedBlock)
	}
	ranges[expectedBlock] = instructionRange{start: blockStart, end: end}
	if expectedBlock+1 != function.BlockCount {
		return nil, fmt.Errorf("semanticabi: flow function %d declares %d blocks but encodes %d", functionIndex, function.BlockCount, expectedBlock+1)
	}
	return ranges, nil
}

func instructionSuccessors(instruction Instruction) ([]uint32, error) {
	switch instruction.Opcode {
	case OpJump:
		return []uint32{instruction.Operands[0]}, nil
	case OpBranch:
		return []uint32{instruction.Operands[1], instruction.Operands[2]}, nil
	case OpHostEffectResume:
		return []uint32{instruction.Operands[3]}, nil
	case OpReturnValue, OpRaiseValue, OpMatchFail:
		return nil, nil
	default:
		return nil, fmt.Errorf("terminator opcode %d has no successor contract", instruction.Opcode)
	}
}

func validateReachableBlocks(functionIndex int, successors [][]uint32) error {
	reachable := make([]bool, len(successors))
	work := []uint32{0}
	for len(work) != 0 {
		block := work[len(work)-1]
		work = work[:len(work)-1]
		if reachable[block] {
			continue
		}
		reachable[block] = true
		work = append(work, successors[block]...)
	}
	for block, ok := range reachable {
		if !ok {
			return fmt.Errorf("semanticabi: flow function %d block %d is unreachable", functionIndex, block)
		}
	}
	return nil
}

func validateFlowInstructionTypes(image *Image, functionIndex int, function FunctionRecord, instructionIndex int) error {
	instruction := image.Instructions[instructionIndex]
	typeOf := func(register uint32) uint32 {
		return image.RegisterTypes[function.RegisterStart+register]
	}
	checkResultType := func(registerOperand, typeOperand int) error {
		registerType := typeOf(instruction.Operands[registerOperand])
		declaredType := instruction.Operands[typeOperand]
		if !compatibleTypes(image, registerType, declaredType) {
			return fmt.Errorf("semanticabi: function %d instruction %d result register type %d conflicts with declared type %d", functionIndex, instructionIndex, registerType, declaredType)
		}
		return nil
	}
	switch instruction.Opcode {
	case OpLoadConst:
		registerType := typeOf(instruction.Operands[0])
		constant := image.Constants[instruction.Operands[1]]
		if !constantCompatible(image, registerType, constant) {
			return fmt.Errorf("semanticabi: function %d instruction %d constant tag %d conflicts with register type %d", functionIndex, instructionIndex, constant.Tag, registerType)
		}
	case OpLoadGlobal, OpUnaryValue, OpBinaryValue, OpCastValue, OpGetMemberValue:
		if err := checkResultType(0, 1); err != nil {
			return err
		}
	case OpMoveValue:
		if !compatibleTypes(image, typeOf(instruction.Operands[0]), typeOf(instruction.Operands[1])) {
			return fmt.Errorf("semanticabi: function %d instruction %d moves incompatible register types", functionIndex, instructionIndex)
		}
	case OpInvoke:
		if err := checkResultType(0, 2); err != nil {
			return err
		}
		target := image.CallTargets[instruction.Operands[1]]
		if len(instruction.Operands)-3 != int(target.Arity) {
			return fmt.Errorf("semanticabi: function %d instruction %d call arity %d, target requires %d", functionIndex, instructionIndex, len(instruction.Operands)-3, target.Arity)
		}
		if target.ReturnType != NoIndex && !compatibleTypes(image, instruction.Operands[2], target.ReturnType) {
			return fmt.Errorf("semanticabi: function %d instruction %d call return type mismatch", functionIndex, instructionIndex)
		}
	case OpTypeTest:
		if typeNameAt(image, typeOf(instruction.Operands[0])) != "bool" {
			return fmt.Errorf("semanticabi: function %d instruction %d type-test destination is not bool", functionIndex, instructionIndex)
		}
	case OpBranch:
		conditionType := typeNameAt(image, typeOf(instruction.Operands[0]))
		if conditionType != "bool" && conditionType != "dynamic" {
			return fmt.Errorf("semanticabi: function %d instruction %d branch condition type is %s", functionIndex, instructionIndex, conditionType)
		}
	case OpHostEffectResume:
		if err := checkResultType(0, 2); err != nil {
			return err
		}
	}
	return nil
}

func validateDefiniteAssignment(image *Image, functionIndex int, function FunctionRecord, ranges []instructionRange, predecessors [][]uint32) error {
	registerCount := int(function.RegisterCount)
	inSets := make([][]bool, function.BlockCount)
	outSets := make([][]bool, function.BlockCount)
	writes := make([][]bool, function.BlockCount)
	for block, current := range ranges {
		writes[block] = make([]bool, registerCount)
		for index := current.start; index < current.end; index++ {
			instruction := image.Instructions[index]
			descriptor, _ := OpByCode(instruction.Opcode)
			for _, operandIndex := range descriptor.Writes {
				writes[block][instruction.Operands[operandIndex]] = true
			}
		}
		inSets[block] = make([]bool, registerCount)
		outSets[block] = make([]bool, registerCount)
		if block == 0 {
			for register := 0; register < int(function.ParameterCount); register++ {
				inSets[block][register] = true
			}
		} else {
			for register := range inSets[block] {
				inSets[block][register] = true
				outSets[block][register] = true
			}
		}
	}
	changed := true
	for changed {
		changed = false
		for block := range ranges {
			if block != 0 {
				next := make([]bool, registerCount)
				copy(next, outSets[predecessors[block][0]])
				for _, predecessor := range predecessors[block][1:] {
					for register := range next {
						next[register] = next[register] && outSets[predecessor][register]
					}
				}
				if !equalBoolSets(next, inSets[block]) {
					inSets[block] = next
					changed = true
				}
			}
			nextOut := append([]bool(nil), inSets[block]...)
			for register, written := range writes[block] {
				nextOut[register] = nextOut[register] || written
			}
			if !equalBoolSets(nextOut, outSets[block]) {
				outSets[block] = nextOut
				changed = true
			}
		}
	}
	for block, current := range ranges {
		defined := append([]bool(nil), inSets[block]...)
		for index := current.start; index < current.end; index++ {
			instruction := image.Instructions[index]
			descriptor, _ := OpByCode(instruction.Opcode)
			for operandIndex, operand := range instruction.Operands {
				kind := descriptor.Variadic
				if operandIndex < len(descriptor.Operands) {
					kind = descriptor.Operands[operandIndex]
				}
				if kind != OperandRegister || isWriteOperand(descriptor, operandIndex) {
					continue
				}
				if !defined[operand] {
					return fmt.Errorf("semanticabi: function %d instruction %d reads undefined register %d", functionIndex, index, operand)
				}
			}
			for _, operandIndex := range descriptor.Writes {
				defined[instruction.Operands[operandIndex]] = true
			}
		}
	}
	return nil
}

func isWriteOperand(descriptor OpDescriptor, operandIndex int) bool {
	for _, current := range descriptor.Writes {
		if int(current) == operandIndex {
			return true
		}
	}
	return false
}

func compatibleTypes(image *Image, left, right uint32) bool {
	if left == right {
		return true
	}
	return typeNameAt(image, left) == "dynamic" || typeNameAt(image, right) == "dynamic"
}

func typeNameAt(image *Image, index uint32) string {
	return image.Strings[image.Types[index].Name]
}

func constantCompatible(image *Image, registerType uint32, constant ConstantRecord) bool {
	name := typeNameAt(image, registerType)
	if name == "dynamic" {
		return true
	}
	switch constant.Tag {
	case TagKindString:
		return name == "String"
	case TagKindBool:
		return name == "bool"
	case TagKindChar:
		return name == "char"
	case TagKindNil:
		return name == "nil" || name == "dynamic"
	case TagKindVoid:
		return name == "void"
	case TagKindInteger:
		return name == integerTypeName(constant.Aux)
	case TagKindFloat:
		return (constant.Aux == 1 && name == "f32") || (constant.Aux != 1 && name == "f64")
	default:
		return true
	}
}

func integerTypeName(aux uint32) string {
	names := []string{"i32", "i8", "i16", "i32", "i64", "i128", "u8", "u16", "u32", "u64", "u128"}
	if int(aux) >= len(names) {
		return "dynamic"
	}
	return names[aux]
}

func equalBoolSets(left, right []bool) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
