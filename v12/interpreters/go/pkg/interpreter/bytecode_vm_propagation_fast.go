package interpreter

import "able/interpreter-go/pkg/runtime"

func (vm *bytecodeVM) hasFollowingSuccessPropagation(result runtime.Value) bool {
	if vm == nil || vm.interp == nil || vm.currentProgram == nil || result == nil || isNilRuntimeValue(result) {
		return false
	}
	return vm.hasFollowingSuccessPropagationOpcode()
}

func (vm *bytecodeVM) hasFollowingSuccessPropagationOpcode() bool {
	if vm == nil || vm.interp == nil || vm.currentProgram == nil {
		return false
	}
	program := vm.currentProgram
	if flags := program.followedByPropagation; flags != nil {
		return uint(vm.ip) < uint(len(flags)) && flags[vm.ip]
	}
	nextIP := vm.ip + 1
	return uint(nextIP) < uint(len(program.instructions)) && program.instructions[nextIP].op == bytecodeOpPropagation
}

func (vm *bytecodeVM) canSkipSuccessPropagation(result runtime.Value) bool {
	if !vm.hasFollowingSuccessPropagation(result) {
		return false
	}
	return !vm.bytecodePropagationValueMayImplementError(result)
}

func (vm *bytecodeVM) bytecodePropagationValueMayImplementError(val runtime.Value) bool {
	if vm == nil || vm.interp == nil || val == nil {
		return false
	}
	switch v := val.(type) {
	case runtime.ErrorValue:
		return true
	case *runtime.ErrorValue:
		return v != nil
	case runtime.InterfaceValue:
		return vm.bytecodeInterfaceValueMayImplementError(&v)
	case *runtime.InterfaceValue:
		return vm.bytecodeInterfaceValueMayImplementError(v)
	case *runtime.StructInstanceValue:
		return v != nil
	case runtime.StructDefinitionValue:
		return true
	case *runtime.StructDefinitionValue:
		return v != nil
	case *runtime.ArrayValue:
		return v != nil && !vm.arrayValueNoErrorForPropagation()
	}
	if token, ok := bytecodeIndexValueTypeToken(val); ok && token != bytecodeIndexTypeUnknown {
		return !vm.arrayGetPrimitiveNoErrorToken(token)
	}
	return vm.interp.propagationValueMayImplementError(val)
}

func (vm *bytecodeVM) bytecodeInterfaceValueMayImplementError(val *runtime.InterfaceValue) bool {
	if vm == nil || vm.interp == nil || val == nil {
		return false
	}
	if val.Interface != nil && val.Interface.Node != nil && val.Interface.Node.ID != nil {
		if val.Interface.Node.ID.Name == "Error" {
			return true
		}
	}
	if val.Underlying != nil {
		return vm.bytecodePropagationValueMayImplementError(val.Underlying)
	}
	return true
}
