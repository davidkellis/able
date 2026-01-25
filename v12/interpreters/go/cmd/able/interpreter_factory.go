package main

import "able/interpreter-go/pkg/interpreter"

func newInterpreter(mode interpreterMode) *interpreter.Interpreter {
	switch mode {
	case interpreterTreewalker:
		return interpreter.New()
	case interpreterBytecode:
		// TODO: swap to bytecode backend once available.
		return interpreter.New()
	default:
		return interpreter.New()
	}
}
