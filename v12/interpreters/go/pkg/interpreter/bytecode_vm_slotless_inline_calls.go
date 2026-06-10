package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) coerceSlotlessInlineReturnValue(fn *runtime.FunctionValue, program *bytecodeProgram, val runtime.Value) (runtime.Value, error) {
	val = vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticReturn, val)
	if vm == nil || vm.interp == nil || fn == nil {
		return val, nil
	}
	if program != nil && program.returnTypeMetadataCached {
		if program.returnType == nil || program.returnTypeUsesGenerics {
			return val, nil
		}
		if program.returnSimpleType == "void" {
			return runtime.VoidValue{}, nil
		}
		if program.returnSimpleCheck != bytecodeSimpleTypeCheckUnknown && inlineCoercionUnnecessaryBySimpleCheck(program.returnSimpleCheck, val) {
			return val, nil
		}
		if program.returnNullableSimple != "" {
			if coerced, ok, err := vm.coerceNullableSimpleProgramReturn(program.returnNullableSimple, val); ok {
				return coerced, err
			}
		}
		if program.returnSimpleType != "" {
			if coerced, ok, err := tryFastSimpleTypeCoercionByName(vm.interp, program.returnSimpleType, val); ok {
				return coerced, err
			}
		}
	}
	switch decl := fn.Declaration.(type) {
	case *ast.FunctionDefinition:
		if decl == nil {
			return val, nil
		}
		return vm.interp.coerceCallableReturnValue(fn, decl.ReturnType, val, fn.Closure)
	case *ast.LambdaExpression:
		if decl == nil {
			return val, nil
		}
		return vm.interp.coerceCallableReturnValue(fn, decl.ReturnType, val, fn.Closure)
	default:
		return val, nil
	}
}

func (vm *bytecodeVM) tryInlineSlotlessResolvedCallFromStack(fn *runtime.FunctionValue, prog *bytecodeProgram, injectedReceiver runtime.Value, hasInjectedReceiver bool, argBase int, argCount int, truncateTo int, callNode *ast.FunctionCall, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if vm == nil || vm.interp == nil || fn == nil || prog == nil || prog.frameLayout != nil {
		return nil, nil
	}
	decl, ok := fn.Declaration.(*ast.FunctionDefinition)
	if !ok || decl == nil || decl.Body == nil {
		if vm.interp != nil {
			vm.interp.recordBytecodeInlineResolvedMiss(bytecodeInlineResolvedMissNoBytecode)
		}
		return nil, nil
	}
	if argBase < 0 || argCount < 0 || argBase+argCount > vm.stackDepth() {
		return nil, fmt.Errorf("bytecode stack underflow")
	}
	if truncateTo < 0 || truncateTo > argBase {
		return nil, fmt.Errorf("bytecode stack underflow")
	}

	paramCount := len(decl.Params)
	optionalLast := paramCount > 0 && isNullableParam(decl.Params[paramCount-1])
	expectedArgs := paramCount
	sourceArgBase := argBase
	injectReceiverIntoParam0 := hasInjectedReceiver && !decl.IsMethodShorthand
	var implicitReceiver runtime.Value
	hasImplicit := false
	if decl.IsMethodShorthand {
		hasImplicit = true
		if hasInjectedReceiver {
			implicitReceiver = injectedReceiver
		} else {
			expectedArgs++
			if argCount > 0 {
				implicitReceiver = vm.stackValue(argBase)
			}
			sourceArgBase++
		}
	} else if hasInjectedReceiver {
		expectedArgs--
		if paramCount > 0 {
			implicitReceiver = injectedReceiver
			hasImplicit = true
		}
	} else if paramCount > 0 && argCount > 0 {
		implicitReceiver = vm.stackValue(argBase)
		hasImplicit = true
	}
	if expectedArgs < 0 || !arityMatchesRuntime(expectedArgs, argCount, optionalLast) {
		if vm.interp != nil {
			vm.interp.recordBytecodeInlineResolvedMiss(bytecodeInlineResolvedMissArity)
		}
		return nil, nil
	}
	if bytecodeInlineSkipsGenericLambda(fn) {
		if vm.interp != nil {
			vm.interp.recordBytecodeInlineResolvedMiss(bytecodeInlineResolvedMissGenericLambda)
		}
		return nil, nil
	}

	localEnv, err := vm.inlineResolvedCallEnvForBindings(fn, prog, nil, injectedReceiver, hasInjectedReceiver, argBase, argCount, callNode)
	if err != nil {
		return nil, err
	}
	if localEnv == nil {
		localEnv = fn.Closure
	}
	localEnv = vm.bytecodeCalleeEnv(localEnv)
	callEnv := runtime.NewEnvironmentWithValueCapacity(localEnv, functionLocalBindingCapacityForLayout(decl, callNode, nil))
	callPlan := vm.interp.functionRuntimeGenericBindingPlan(fn)
	missingOptional := optionalLast && argCount == expectedArgs-1
	for idx, param := range decl.Params {
		if param == nil {
			return nil, fmt.Errorf("function parameter %d is nil", idx)
		}
		arg := vm.slotlessInlineCallParamValue(idx, sourceArgBase, paramCount, missingOptional, injectReceiverIntoParam0, injectedReceiver)
		arg = vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticCall, arg)
		paramType := vm.interp.canonicalizeTypeExpressionCached(param.ParamType, fn.Closure, vm.interp.typeExpressionReferencesAliasCached(param.ParamType))
		if paramType != nil && !callPlan.paramUsesGeneric(idx) && !vm.interp.coerceValueToTypeWouldBeNoOp(paramType) && !inlineCoercionUnnecessaryWithInterpreter(vm.interp, paramType, arg) {
			coerced, err := vm.interp.coerceValueToType(paramType, arg)
			if err != nil {
				return nil, err
			}
			arg = coerced
		}
		if bindSimpleIdentifierPatternIntoEnv(callEnv, param.Name, arg) {
			continue
		}
		if err := vm.interp.assignPattern(param.Name, arg, callEnv, true, nil); err != nil {
			return nil, err
		}
	}
	if hasImplicit {
		state := vm.interp.stateFromEnv(callEnv)
		state.pushImplicitReceiver(vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticCall, implicitReceiver))
	}

	vm.truncateStack(truncateTo)
	returnGenericNames := bytecodeInlineReturnGenericNames(fn, prog)
	vm.pushCallFrame(vm.ip+1, currentProgram, vm.slots, vm.env, returnGenericNames, len(vm.iterStack), len(vm.loopStack), hasImplicit, false)
	vm.setTopCallFrameReturnCoercionFunction(fn)
	vm.slots = nil
	vm.prepareValueSlotI32Frame(prog)
	vm.env = callEnv
	vm.ip = 0
	return prog, nil
}

func (vm *bytecodeVM) slotlessInlineCallParamValue(idx int, sourceArgBase int, paramCount int, missingOptional bool, injectReceiverIntoParam0 bool, injectedReceiver runtime.Value) runtime.Value {
	if missingOptional && idx == paramCount-1 {
		return runtime.NilValue{}
	}
	if injectReceiverIntoParam0 {
		if idx == 0 {
			return injectedReceiver
		}
		return vm.stackValue(sourceArgBase + idx - 1)
	}
	return vm.stackValue(sourceArgBase + idx)
}
