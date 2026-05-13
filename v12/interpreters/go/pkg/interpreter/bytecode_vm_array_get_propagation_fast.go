package interpreter

import "able/interpreter-go/pkg/runtime"

func (vm *bytecodeVM) canSkipArrayGetSuccessPropagation(result runtime.Value, elementToken uint16, tokenKnown bool) bool {
	if vm == nil || vm.interp == nil || vm.currentProgram == nil || !tokenKnown || result == nil || isNilRuntimeValue(result) {
		return false
	}
	nextIP := vm.ip + 1
	if nextIP < 0 || nextIP >= len(vm.currentProgram.instructions) || vm.currentProgram.instructions[nextIP].op != bytecodeOpPropagation {
		return false
	}
	if !bytecodeArrayGetResultMatchesFloatToken(result, elementToken) {
		return false
	}
	switch elementToken {
	case bytecodeIndexTypeF32:
		return vm.arrayGetPrimitiveNoError("f32")
	case bytecodeIndexTypeF64:
		return vm.arrayGetPrimitiveNoError("f64")
	default:
		return false
	}
}

func bytecodeArrayGetResultMatchesFloatToken(result runtime.Value, elementToken uint16) bool {
	switch v := result.(type) {
	case runtime.FloatValue:
		return bytecodeFloatTypeToken(v.TypeSuffix) == elementToken
	case *runtime.FloatValue:
		return v != nil && bytecodeFloatTypeToken(v.TypeSuffix) == elementToken
	default:
		return false
	}
}

func (vm *bytecodeVM) arrayGetPrimitiveNoError(typeName string) bool {
	if vm == nil || vm.interp == nil {
		return false
	}
	version := vm.interp.currentMethodCacheVersion()
	switch typeName {
	case "f32":
		if vm.arrayGetF32NoErrorKnown && vm.arrayGetF32NoErrorVersion == version {
			return vm.arrayGetF32NoError
		}
		noError := !vm.interp.typeNameMayImplementError(typeName)
		vm.arrayGetF32NoErrorVersion = version
		vm.arrayGetF32NoErrorKnown = true
		vm.arrayGetF32NoError = noError
		return noError
	case "f64":
		if vm.arrayGetF64NoErrorKnown && vm.arrayGetF64NoErrorVersion == version {
			return vm.arrayGetF64NoError
		}
		noError := !vm.interp.typeNameMayImplementError(typeName)
		vm.arrayGetF64NoErrorVersion = version
		vm.arrayGetF64NoErrorKnown = true
		vm.arrayGetF64NoError = noError
		return noError
	default:
		return false
	}
}

func (vm *bytecodeVM) arrayValueNoErrorForPropagation() bool {
	if vm == nil || vm.interp == nil {
		return false
	}
	version := vm.interp.currentMethodCacheVersion()
	if vm.arrayValueNoErrorKnown && vm.arrayValueNoErrorVersion == version {
		return vm.arrayValueNoError
	}
	noError := !vm.interp.typeNameMayImplementError("Array")
	vm.arrayValueNoErrorVersion = version
	vm.arrayValueNoErrorKnown = true
	vm.arrayValueNoError = noError
	return noError
}
