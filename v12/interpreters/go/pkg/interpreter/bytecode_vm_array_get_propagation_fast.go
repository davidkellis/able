package interpreter

import "able/interpreter-go/pkg/runtime"

func (vm *bytecodeVM) canSkipArrayGetSuccessPropagation(result runtime.Value, elementToken uint16, tokenKnown bool) bool {
	if !vm.hasFollowingSuccessPropagation(result) {
		return false
	}
	if tokenKnown &&
		elementToken != bytecodeIndexTypeUnknown &&
		bytecodeArrayGetResultMatchesToken(result, elementToken) &&
		vm.arrayGetPrimitiveNoErrorToken(elementToken) {
		return true
	}
	return !vm.bytecodePropagationValueMayImplementError(result)
}

func (vm *bytecodeVM) canSkipExactPrimitiveArrayGetSuccessPropagation(result runtime.Value, elementToken uint16, tokenKnown bool) bool {
	if !bytecodeExactPrimitiveArrayGetToken(elementToken, tokenKnown) ||
		result == nil ||
		isNilRuntimeValue(result) ||
		bytecodeArrayGetResultIsRuntimeError(result) ||
		!vm.hasFollowingSuccessPropagationOpcode() {
		return false
	}
	return vm.arrayGetPrimitiveNoErrorToken(elementToken)
}

func bytecodeExactPrimitiveArrayGetToken(elementToken uint16, tokenKnown bool) bool {
	if !tokenKnown {
		return false
	}
	switch elementToken {
	case bytecodeIndexTypeI8, bytecodeIndexTypeI16, bytecodeIndexTypeI32, bytecodeIndexTypeI64,
		bytecodeIndexTypeI128, bytecodeIndexTypeU8, bytecodeIndexTypeU16, bytecodeIndexTypeU32,
		bytecodeIndexTypeU64, bytecodeIndexTypeU128, bytecodeIndexTypeIsize, bytecodeIndexTypeUsize,
		bytecodeIndexTypeF32, bytecodeIndexTypeF64, bytecodeIndexTypeBool, bytecodeIndexTypeChar:
		return true
	default:
		return false
	}
}

func bytecodeArrayGetResultIsRuntimeError(result runtime.Value) bool {
	switch v := result.(type) {
	case runtime.ErrorValue:
		return true
	case *runtime.ErrorValue:
		return v != nil
	default:
		return false
	}
}

func bytecodeArrayGetResultMatchesToken(result runtime.Value, elementToken uint16) bool {
	switch elementToken {
	case bytecodeIndexTypeI32:
		switch v := result.(type) {
		case bytecodeRawI32SlotValue:
			return true
		case *bytecodeRawI32StackCell:
			return v != nil
		case runtime.IntegerValue:
			return v.TypeSuffix == runtime.IntegerI32
		case *runtime.IntegerValue:
			return v != nil && v.TypeSuffix == runtime.IntegerI32
		}
	case bytecodeIndexTypeU32:
		switch result.(type) {
		case bytecodeRawU32ResultValue:
			return true
		}
	case bytecodeIndexTypeChar:
		switch v := result.(type) {
		case runtime.CharValue:
			return true
		case *runtime.CharValue:
			return v != nil
		}
	case bytecodeIndexTypeBool:
		switch v := result.(type) {
		case runtime.BoolValue:
			return true
		case *runtime.BoolValue:
			return v != nil
		}
	case bytecodeIndexTypeString:
		switch v := result.(type) {
		case runtime.StringValue:
			return true
		case *runtime.StringValue:
			return v != nil
		}
	}
	switch v := result.(type) {
	case runtime.InterfaceValue:
		return bytecodeArrayGetResultMatchesToken(unwrapInterfaceValue(&v), elementToken)
	case *runtime.InterfaceValue:
		return bytecodeArrayGetResultMatchesToken(unwrapInterfaceValue(v), elementToken)
	}
	switch elementToken {
	case bytecodeIndexTypeI8, bytecodeIndexTypeI16, bytecodeIndexTypeI32, bytecodeIndexTypeI64,
		bytecodeIndexTypeI128, bytecodeIndexTypeU8, bytecodeIndexTypeU16, bytecodeIndexTypeU32,
		bytecodeIndexTypeU64, bytecodeIndexTypeU128, bytecodeIndexTypeIsize, bytecodeIndexTypeUsize:
		switch v := result.(type) {
		case runtime.IntegerValue:
			return bytecodeIntegerTypeToken(v.TypeSuffix) == elementToken
		case *runtime.IntegerValue:
			return v != nil && bytecodeIntegerTypeToken(v.TypeSuffix) == elementToken
		}
		if kind, _, ok := bytecodeRawIntegerValueInfo(result); ok {
			return bytecodeIntegerTypeToken(kind) == elementToken
		}
	case bytecodeIndexTypeF32, bytecodeIndexTypeF64:
		switch v := result.(type) {
		case runtime.FloatValue:
			return bytecodeFloatTypeToken(v.TypeSuffix) == elementToken
		case *runtime.FloatValue:
			return v != nil && bytecodeFloatTypeToken(v.TypeSuffix) == elementToken
		}
		if _, kind, ok := bytecodeDirectRawFloatValue(result); ok {
			return bytecodeFloatTypeToken(kind) == elementToken
		}
	case bytecodeIndexTypeString:
		switch v := result.(type) {
		case runtime.StringValue:
			return true
		case *runtime.StringValue:
			return v != nil
		}
	case bytecodeIndexTypeBool:
		switch v := result.(type) {
		case runtime.BoolValue:
			return true
		case *runtime.BoolValue:
			return v != nil
		}
	case bytecodeIndexTypeChar:
		switch v := result.(type) {
		case runtime.CharValue:
			return true
		case *runtime.CharValue:
			return v != nil
		}
	case bytecodeIndexTypeVoid:
		switch v := result.(type) {
		case runtime.VoidValue:
			return true
		case *runtime.VoidValue:
			return v != nil
		}
	}
	return false
}

func bytecodeArrayGetResultMatchesFloatToken(result runtime.Value, elementToken uint16) bool {
	switch v := result.(type) {
	case runtime.FloatValue:
		return bytecodeFloatTypeToken(v.TypeSuffix) == elementToken
	case *runtime.FloatValue:
		return v != nil && bytecodeFloatTypeToken(v.TypeSuffix) == elementToken
	case runtime.InterfaceValue, *runtime.InterfaceValue:
		return bytecodeArrayGetResultMatchesFloatToken(unwrapInterfaceValue(result), elementToken)
	default:
	}
	if _, kind, ok := bytecodeDirectRawFloatValue(result); ok {
		return bytecodeFloatTypeToken(kind) == elementToken
	}
	return false
}

func bytecodeIndexTypeTokenName(token uint16) (string, bool) {
	switch token {
	case bytecodeIndexTypeI8:
		return string(runtime.IntegerI8), true
	case bytecodeIndexTypeI16:
		return string(runtime.IntegerI16), true
	case bytecodeIndexTypeI32:
		return string(runtime.IntegerI32), true
	case bytecodeIndexTypeI64:
		return string(runtime.IntegerI64), true
	case bytecodeIndexTypeI128:
		return string(runtime.IntegerI128), true
	case bytecodeIndexTypeU8:
		return string(runtime.IntegerU8), true
	case bytecodeIndexTypeU16:
		return string(runtime.IntegerU16), true
	case bytecodeIndexTypeU32:
		return string(runtime.IntegerU32), true
	case bytecodeIndexTypeU64:
		return string(runtime.IntegerU64), true
	case bytecodeIndexTypeU128:
		return string(runtime.IntegerU128), true
	case bytecodeIndexTypeIsize:
		return string(runtime.IntegerIsize), true
	case bytecodeIndexTypeUsize:
		return string(runtime.IntegerUsize), true
	case bytecodeIndexTypeF32:
		return string(runtime.FloatF32), true
	case bytecodeIndexTypeF64:
		return string(runtime.FloatF64), true
	case bytecodeIndexTypeString:
		return "String", true
	case bytecodeIndexTypeBool:
		return "bool", true
	case bytecodeIndexTypeChar:
		return "char", true
	case bytecodeIndexTypeNil:
		return "nil", true
	case bytecodeIndexTypeVoid:
		return "void", true
	default:
		return "", false
	}
}

func (vm *bytecodeVM) arrayGetPrimitiveNoError(typeName string) bool {
	if vm == nil || vm.interp == nil {
		return false
	}
	interp := vm.interp
	version := interp.methodCacheVersion
	if !interp.envSingleThread {
		version = interp.currentMethodCacheVersion()
	}
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

func (vm *bytecodeVM) arrayGetPrimitiveNoErrorToken(token uint16) bool {
	switch token {
	case bytecodeIndexTypeF32:
		return vm.arrayGetPrimitiveNoError("f32")
	case bytecodeIndexTypeF64:
		return vm.arrayGetPrimitiveNoError("f64")
	}
	if vm == nil || vm.interp == nil {
		return false
	}
	interp := vm.interp
	version := interp.methodCacheVersion
	if !interp.envSingleThread {
		version = interp.currentMethodCacheVersion()
	}
	if vm.arrayGetPrimitiveNoErrorTokenHotKnown &&
		vm.arrayGetPrimitiveNoErrorTokenHotToken == token &&
		vm.arrayGetPrimitiveNoErrorTokenHotVersion == version {
		return vm.arrayGetPrimitiveNoErrorTokenHotNoError
	}
	typeName, ok := bytecodeIndexTypeTokenName(token)
	if !ok {
		return false
	}
	noError := !vm.interp.typeNameMayImplementError(typeName)
	vm.arrayGetPrimitiveNoErrorTokenHotToken = token
	vm.arrayGetPrimitiveNoErrorTokenHotVersion = version
	vm.arrayGetPrimitiveNoErrorTokenHotKnown = true
	vm.arrayGetPrimitiveNoErrorTokenHotNoError = noError
	return noError
}

func (vm *bytecodeVM) arrayValueNoErrorForPropagation() bool {
	if vm == nil || vm.interp == nil {
		return false
	}
	interp := vm.interp
	version := interp.methodCacheVersion
	if !interp.envSingleThread {
		version = interp.currentMethodCacheVersion()
	}
	if vm.arrayValueNoErrorKnown && vm.arrayValueNoErrorVersion == version {
		return vm.arrayValueNoError
	}
	noError := !vm.interp.typeNameMayImplementError("Array")
	vm.arrayValueNoErrorVersion = version
	vm.arrayValueNoErrorKnown = true
	vm.arrayValueNoError = noError
	return noError
}
