package interpreter

func bytecodeFuseImplicitReturnBinaryIntAdd(instructions []bytecodeInstruction, layout *bytecodeFrameLayout) {
	for idx := 0; idx+1 < len(instructions); idx++ {
		if instructions[idx].op != bytecodeOpBinaryIntAdd {
			continue
		}
		ret := instructions[idx+1]
		if ret.op != bytecodeOpReturn || ret.node != nil {
			continue
		}
		if layout != nil && layout.returnSimpleCheck == bytecodeSimpleTypeCheckI32 {
			instructions[idx].op = bytecodeOpReturnBinaryIntAddI32
		} else {
			instructions[idx].op = bytecodeOpReturnBinaryIntAdd
		}
	}
}
