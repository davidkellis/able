package interpreter

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) stringInterpolationPartsBuffer(size int) []runtime.Value {
	if size <= 0 {
		return nil
	}
	if cap(vm.stringInterpParts) < size {
		vm.stringInterpParts = make([]runtime.Value, size)
	}
	return vm.stringInterpParts[:size]
}

func (vm *bytecodeVM) execStringInterpolation(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode string interpolation instruction missing")
	}
	if instr.argCount < 0 {
		return fmt.Errorf("bytecode string interpolation count invalid")
	}
	if len(vm.stack) < instr.argCount {
		return fmt.Errorf("bytecode stack underflow")
	}
	parts := vm.stringInterpolationPartsBuffer(instr.argCount)
	defer clear(parts)

	start := len(vm.stack) - instr.argCount
	copy(parts, vm.stack[start:])
	clear(vm.stack[start:])
	vm.stack = vm.stack[:start]

	var builder strings.Builder
	for _, part := range parts {
		str, err := vm.interp.stringifyValue(part, vm.env)
		if err != nil {
			return err
		}
		builder.WriteString(str)
	}
	vm.stack = append(vm.stack, runtime.StringValue{Val: builder.String()})
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execArrayLiteral(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode array literal instruction missing")
	}
	if instr.argCount < 0 {
		return fmt.Errorf("bytecode array literal count invalid")
	}
	if len(vm.stack) < instr.argCount {
		return fmt.Errorf("bytecode stack underflow")
	}
	start := len(vm.stack) - instr.argCount
	values := make([]runtime.Value, instr.argCount)
	copy(values, vm.stack[start:])
	clear(vm.stack[start:])
	vm.stack = vm.stack[:start]
	arr := vm.interp.newArrayValue(values, len(values))
	vm.stack = append(vm.stack, arr)
	vm.ip++
	return nil
}
