package interpreter

func bytecodeQuickenBinaryCompareJumps(instructions []bytecodeInstruction) {
	for idx := 0; idx+1 < len(instructions); idx++ {
		current := &instructions[idx]
		next := instructions[idx+1]
		if current.op != bytecodeOpBinary || next.op != bytecodeOpJumpIfFalse {
			continue
		}
		switch current.operator {
		case "<", "<=", ">", ">=", "==", "!=":
			current.op = bytecodeOpJumpIfBinaryCompareFalse
			current.target = next.target
		}
	}
}
