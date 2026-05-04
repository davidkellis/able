package interpreter

import (
	"fmt"
	"strconv"
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
	if handled, err := vm.execStringInterpolationFast(instr); handled {
		return err
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

func (vm *bytecodeVM) execStringInterpolationFast(instr *bytecodeInstruction) (bool, error) {
	if instr.argCount != 2 {
		return false, nil
	}
	start := len(vm.stack) - 2
	left := vm.stack[start]
	right := vm.stack[start+1]
	if leftString, ok := left.(runtime.StringValue); ok {
		if rightInt, ok := right.(runtime.IntegerValue); ok {
			vm.finishStringIntegerInterpolationFast(start, leftString.Val, rightInt)
			return true, nil
		}
	}
	if _, leftIsString := left.(runtime.StringValue); leftIsString {
		if _, rightIsString := right.(runtime.StringValue); rightIsString {
			return false, nil
		}
	}
	leftStr, leftOK := bytecodePrimitiveInterpolationString(left)
	rightStr, rightOK := bytecodePrimitiveInterpolationString(right)
	if !leftOK || !rightOK {
		return false, nil
	}
	vm.stack[start] = runtime.StringValue{Val: leftStr + rightStr}
	clear(vm.stack[start+1:])
	vm.stack = vm.stack[:start+1]
	vm.ip++
	return true, nil
}

func (vm *bytecodeVM) finishStringIntegerInterpolationFast(start int, prefix string, value runtime.IntegerValue) {
	var builder strings.Builder
	if raw, ok := value.ToInt64(); ok {
		if raw >= 0 && raw <= 9 {
			builder.Grow(len(prefix) + 1)
			builder.WriteString(prefix)
			builder.WriteByte(byte('0' + raw))
		} else {
			var digits [20]byte
			buf := strconv.AppendInt(digits[:0], raw, 10)
			builder.Grow(len(prefix) + len(buf))
			builder.WriteString(prefix)
			builder.Write(buf)
		}
	} else {
		suffix := value.String()
		builder.Grow(len(prefix) + len(suffix))
		builder.WriteString(prefix)
		builder.WriteString(suffix)
	}
	vm.stack[start] = runtime.StringValue{Val: builder.String()}
	clear(vm.stack[start+1:])
	vm.stack = vm.stack[:start+1]
	vm.ip++
}

func bytecodePrimitiveInterpolationString(val runtime.Value) (string, bool) {
	switch v := val.(type) {
	case runtime.StringValue:
		return v.Val, true
	case runtime.IntegerValue:
		if raw, ok := v.ToInt64(); ok {
			return strconv.FormatInt(raw, 10), true
		}
		return v.String(), true
	case runtime.BoolValue:
		if v.Val {
			return "true", true
		}
		return "false", true
	case runtime.CharValue:
		return string(v.Val), true
	case runtime.NilValue:
		return "nil", true
	default:
		return "", false
	}
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
