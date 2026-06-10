package interpreter

func bytecodeFuseImplicitReturnBinary(instructions []bytecodeInstruction, layout *bytecodeFrameLayout) {
	for idx := 0; idx+1 < len(instructions); idx++ {
		if !bytecodeImplicitReturnBinaryOpcode(instructions[idx].op) {
			continue
		}
		ret := instructions[idx+1]
		if ret.op != bytecodeOpReturn || ret.node != nil {
			continue
		}
		if instructions[idx].op == bytecodeOpBinaryIntAdd && layout != nil && layout.returnSimpleCheck == bytecodeSimpleTypeCheckI32 {
			instructions[idx].op = bytecodeOpReturnBinaryIntAddI32
		} else if instructions[idx].op == bytecodeOpBinaryIntAdd {
			instructions[idx].op = bytecodeOpReturnBinaryIntAdd
		} else {
			instructions[idx].op = bytecodeOpReturnBinary
		}
	}
}

func bytecodeImplicitReturnBinaryOpcode(op bytecodeOp) bool {
	switch op {
	case bytecodeOpBinary,
		bytecodeOpBinaryIntAdd,
		bytecodeOpBinaryIntSub,
		bytecodeOpBinaryIntLessEqual:
		return true
	default:
		return false
	}
}
