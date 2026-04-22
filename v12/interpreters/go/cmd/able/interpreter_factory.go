package main

import "able/interpreter-go/pkg/interpreter"

func newInterpreter(mode interpreterMode) (*interpreter.Interpreter, error) {
	exec, err := interpreter.NewExecutorFromEnvironment()
	if err != nil {
		return nil, err
	}
	switch mode {
	case interpreterTreewalker:
		return interpreter.NewWithExecutor(exec), nil
	case interpreterBytecode:
		return interpreter.NewBytecodeWithExecutor(exec), nil
	default:
		return interpreter.NewWithExecutor(exec), nil
	}
}
