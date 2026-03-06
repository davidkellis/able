package interpreter

func bytecodeBinaryOpcodeForOperator(op string) bytecodeOp {
	switch op {
	case "+":
		return bytecodeOpBinaryIntAdd
	case "-":
		return bytecodeOpBinaryIntSub
	case "<=":
		return bytecodeOpBinaryIntLessEqual
	default:
		return bytecodeOpBinary
	}
}
