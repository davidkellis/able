package interpreter

import "fmt"

func (vm *bytecodeVM) execJumpIfNotNil(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode jump-if-not-nil missing instruction")
	}
	val, err := vm.pop()
	if err != nil {
		return err
	}
	if !isNilRuntimeValue(val) {
		vm.ip = instr.target
		return nil
	}
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execJumpIfNotTypedPattern(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode jump-if-not-typed-pattern missing instruction")
	}
	val, err := vm.pop()
	if err != nil {
		return err
	}
	coerced, ok := vm.interp.matchTypedPatternValue(instr.typeExpr, val)
	if !ok {
		vm.ip = instr.target
		return nil
	}
	if coerced == nil {
		coerced = val
	}
	vm.stack = append(vm.stack, coerced)
	vm.ip++
	return nil
}
