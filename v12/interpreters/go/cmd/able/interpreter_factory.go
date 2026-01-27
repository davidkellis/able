package main

import "able/interpreter-go/pkg/interpreter"

func newInterpreter(mode interpreterMode) *interpreter.Interpreter {
	switch mode {
	case interpreterTreewalker:
		return interpreter.New()
	case interpreterBytecode:
		return interpreter.NewBytecode()
	default:
		return interpreter.New()
	}
}
