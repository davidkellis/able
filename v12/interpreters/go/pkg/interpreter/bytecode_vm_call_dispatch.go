package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func bytecodeResolveDirectFunctionCallTarget(callee runtime.Value) (*runtime.FunctionValue, runtime.Value, runtime.Value, bool, bool) {
	switch fn := callee.(type) {
	case *runtime.FunctionValue:
		if fn == nil {
			return nil, nil, nil, false, false
		}
		return fn, fn, nil, false, true
	case *runtime.FunctionOverloadValue:
		if fn == nil || len(fn.Overloads) != 1 || fn.Overloads[0] == nil {
			return nil, nil, nil, false, false
		}
		return fn.Overloads[0], fn, nil, false, true
	case runtime.BoundMethodValue:
		if methodFn, ok := fn.Method.(*runtime.FunctionValue); ok && methodFn != nil {
			return methodFn, fn.Method, fn.Receiver, true, true
		}
		overloads := functionOverloadsView(fn.Method)
		if len(overloads) != 1 || overloads[0] == nil {
			return nil, nil, nil, false, false
		}
		return overloads[0], fn.Method, fn.Receiver, true, true
	case *runtime.BoundMethodValue:
		if fn == nil {
			return nil, nil, nil, false, false
		}
		if methodFn, ok := fn.Method.(*runtime.FunctionValue); ok && methodFn != nil {
			return methodFn, fn.Method, fn.Receiver, true, true
		}
		overloads := functionOverloadsView(fn.Method)
		if len(overloads) != 1 || overloads[0] == nil {
			return nil, nil, nil, false, false
		}
		return overloads[0], fn.Method, fn.Receiver, true, true
	default:
		return nil, nil, nil, false, false
	}
}

func (vm *bytecodeVM) tryCallDirectFunctionValue(callee runtime.Value, args []runtime.Value, callNode *ast.FunctionCall) (runtime.Value, bool, error) {
	if vm == nil || vm.interp == nil {
		return nil, false, fmt.Errorf("bytecode VM is nil")
	}
	fn, partialTarget, injectedReceiver, hasInjectedReceiver, ok := bytecodeResolveDirectFunctionCallTarget(callee)
	if !ok || fn == nil {
		return nil, false, nil
	}
	evalArgs := vm.prepareResolvedFunctionCallArgsWithOptionalReceiver(
		args,
		bytecodeCallTargetNeedsStableArgs(callee),
		injectedReceiver,
		hasInjectedReceiver,
	)
	if len(evalArgs) < minArgsForFunctionValue(fn) {
		return nil, false, nil
	}
	if !vm.interp.matchesSingleRuntimeOverload(fn, evalArgs) {
		return nil, false, nil
	}
	vm.markBytecodeArrayOwnershipValuesEscaped(evalArgs, bytecodeArrayOwnershipEscapeUnknownCall)
	result, err := vm.interp.callResolvedFunctionValue(fn, partialTarget, evalArgs, vm.env, callNode, true)
	return result, true, err
}

func (vm *bytecodeVM) tryCallDirectFunctionValueFromStack(callee runtime.Value, argBase int, argCount int, truncateTo int, callNode *ast.FunctionCall) (runtime.Value, bool, error) {
	if vm == nil || vm.interp == nil {
		return nil, false, fmt.Errorf("bytecode VM is nil")
	}
	if argBase < 0 || argCount < 0 || argBase+argCount > vm.stackDepth() {
		return nil, false, fmt.Errorf("bytecode stack underflow")
	}
	if truncateTo < 0 || truncateTo > argBase {
		return nil, false, fmt.Errorf("bytecode stack underflow")
	}
	if fn, ok := callee.(*runtime.FunctionValue); ok {
		if fn == nil {
			return nil, false, nil
		}
		evalArgs := vm.stackValues(argBase, argBase+argCount)
		vm.markBytecodeArrayOwnershipValuesEscaped(evalArgs, bytecodeArrayOwnershipEscapeUnknownCall)
		vm.interp.recordBytecodeDirectFunctionStackHit()
		vm.truncateStack(truncateTo)
		result, err := vm.interp.callResolvedFunctionValue(fn, fn, evalArgs, vm.env, callNode, true)
		return result, true, err
	}
	fn, partialTarget, injectedReceiver, hasInjectedReceiver, ok := bytecodeResolveDirectFunctionCallTarget(callee)
	if !ok || fn == nil {
		return nil, false, nil
	}
	args := vm.stackValues(argBase, argBase+argCount)
	evalArgs := vm.prepareResolvedFunctionCallArgsWithOptionalReceiver(
		args,
		bytecodeCallTargetNeedsStableArgs(callee),
		injectedReceiver,
		hasInjectedReceiver,
	)
	if len(evalArgs) < minArgsForFunctionValue(fn) {
		return nil, false, nil
	}
	if !vm.interp.matchesSingleRuntimeOverload(fn, evalArgs) {
		return nil, false, nil
	}
	vm.markBytecodeArrayOwnershipValuesEscaped(evalArgs, bytecodeArrayOwnershipEscapeUnknownCall)
	vm.interp.recordBytecodeDirectFunctionStackHit()
	vm.truncateStack(truncateTo)
	result, err := vm.interp.callResolvedFunctionValue(fn, partialTarget, evalArgs, vm.env, callNode, true)
	return result, true, err
}

func (vm *bytecodeVM) callCallableValueMutable(callee runtime.Value, args []runtime.Value, callNode *ast.FunctionCall) (runtime.Value, error) {
	if vm == nil || vm.interp == nil {
		return nil, fmt.Errorf("bytecode VM is nil")
	}
	if callee == nil {
		return nil, fmt.Errorf("call target missing function value")
	}
	vm.markBytecodeArrayOwnershipValuesEscaped(args, bytecodeArrayOwnershipEscapeUnknownCall)
	switch fn := callee.(type) {
	case runtime.PartialFunctionValue:
		return vm.callCallableValueMutable(fn.Target, mergePartialCallArgs(fn.BoundArgs, args), callNode)
	case *runtime.PartialFunctionValue:
		if fn == nil {
			return nil, fmt.Errorf("partial function is nil")
		}
		return vm.callCallableValueMutable(fn.Target, mergePartialCallArgs(fn.BoundArgs, args), callNode)
	case runtime.DynRefValue:
		resolved, err := vm.interp.resolveDynRef(fn)
		if err != nil {
			return nil, err
		}
		return vm.callCallableValueMutable(resolved, args, callNode)
	case *runtime.DynRefValue:
		if fn == nil {
			return nil, fmt.Errorf("dyn ref is nil")
		}
		resolved, err := vm.interp.resolveDynRef(*fn)
		if err != nil {
			return nil, err
		}
		return vm.callCallableValueMutable(resolved, args, callNode)
	}
	if result, handled, err := vm.tryCallDirectFunctionValue(callee, args, callNode); handled || err != nil {
		return result, err
	}
	return vm.interp.callCallableValueMutable(callee, args, vm.env, callNode)
}

func bytecodeResolveInjectedNativeCallTarget(callee runtime.Value, receiver runtime.Value) (bytecodeExactNativeCallTarget, runtime.Value, bool) {
	switch fn := callee.(type) {
	case runtime.NativeFunctionValue:
		return bytecodeExactNativeCallTarget{
			native:           fn,
			injectedReceiver: receiver,
			hasReceiver:      true,
		}, callee, true
	case *runtime.NativeFunctionValue:
		if fn == nil {
			return bytecodeExactNativeCallTarget{}, nil, false
		}
		return bytecodeExactNativeCallTarget{
			native:           *fn,
			injectedReceiver: receiver,
			hasReceiver:      true,
		}, callee, true
	case runtime.NativeBoundMethodValue:
		return bytecodeExactNativeCallTarget{
			native:           fn.Method,
			injectedReceiver: fn.Receiver,
			hasReceiver:      true,
		}, fn.Method, true
	case *runtime.NativeBoundMethodValue:
		if fn == nil {
			return bytecodeExactNativeCallTarget{}, nil, false
		}
		return bytecodeExactNativeCallTarget{
			native:           fn.Method,
			injectedReceiver: fn.Receiver,
			hasReceiver:      true,
		}, fn.Method, true
	case runtime.BoundMethodValue:
		switch method := fn.Method.(type) {
		case runtime.NativeFunctionValue:
			return bytecodeExactNativeCallTarget{
				native:           method,
				injectedReceiver: fn.Receiver,
				hasReceiver:      true,
			}, fn.Method, true
		case *runtime.NativeFunctionValue:
			if method == nil {
				return bytecodeExactNativeCallTarget{}, nil, false
			}
			return bytecodeExactNativeCallTarget{
				native:           *method,
				injectedReceiver: fn.Receiver,
				hasReceiver:      true,
			}, fn.Method, true
		}
	case *runtime.BoundMethodValue:
		if fn == nil {
			return bytecodeExactNativeCallTarget{}, nil, false
		}
		switch method := fn.Method.(type) {
		case runtime.NativeFunctionValue:
			return bytecodeExactNativeCallTarget{
				native:           method,
				injectedReceiver: fn.Receiver,
				hasReceiver:      true,
			}, fn.Method, true
		case *runtime.NativeFunctionValue:
			if method == nil {
				return bytecodeExactNativeCallTarget{}, nil, false
			}
			return bytecodeExactNativeCallTarget{
				native:           *method,
				injectedReceiver: fn.Receiver,
				hasReceiver:      true,
			}, fn.Method, true
		}
	}
	return bytecodeExactNativeCallTarget{}, nil, false
}

func bytecodeResolveInjectedOverloads(callee runtime.Value, receiver runtime.Value) ([]*runtime.FunctionValue, runtime.Value, runtime.Value, bool) {
	switch fn := callee.(type) {
	case *runtime.FunctionOverloadValue:
		if fn == nil || len(fn.Overloads) == 0 {
			return nil, nil, nil, false
		}
		return functionOverloadsView(fn), fn, receiver, true
	case runtime.BoundMethodValue:
		overloads := functionOverloadsView(fn.Method)
		if len(overloads) == 0 {
			return nil, nil, nil, false
		}
		return overloads, fn.Method, fn.Receiver, true
	case *runtime.BoundMethodValue:
		if fn == nil {
			return nil, nil, nil, false
		}
		overloads := functionOverloadsView(fn.Method)
		if len(overloads) == 0 {
			return nil, nil, nil, false
		}
		return overloads, fn.Method, fn.Receiver, true
	default:
		return nil, nil, nil, false
	}
}

func (vm *bytecodeVM) callCallableValueWithInjectedReceiver(callee runtime.Value, receiver runtime.Value, args []runtime.Value, callNode *ast.FunctionCall) (runtime.Value, error) {
	if vm == nil || vm.interp == nil {
		return nil, fmt.Errorf("bytecode VM is nil")
	}
	if callee == nil {
		return nil, fmt.Errorf("call target missing function value")
	}
	switch fn := callee.(type) {
	case runtime.PartialFunctionValue:
		return vm.callCallableValueWithInjectedReceiver(fn.Target, receiver, mergePartialCallArgs(fn.BoundArgs, args), callNode)
	case *runtime.PartialFunctionValue:
		if fn == nil {
			return nil, fmt.Errorf("partial function is nil")
		}
		return vm.callCallableValueWithInjectedReceiver(fn.Target, receiver, mergePartialCallArgs(fn.BoundArgs, args), callNode)
	case runtime.DynRefValue:
		resolved, err := vm.interp.resolveDynRef(fn)
		if err != nil {
			return nil, err
		}
		return vm.callCallableValueWithInjectedReceiver(resolved, receiver, args, callNode)
	case *runtime.DynRefValue:
		if fn == nil {
			return nil, fmt.Errorf("dyn ref is nil")
		}
		resolved, err := vm.interp.resolveDynRef(*fn)
		if err != nil {
			return nil, err
		}
		return vm.callCallableValueWithInjectedReceiver(resolved, receiver, args, callNode)
	}
	if target, partialTarget, ok := bytecodeResolveInjectedNativeCallTarget(callee, receiver); ok {
		if target.native.Arity >= 0 {
			provided := len(args)
			if provided > target.native.Arity {
				name := target.native.Name
				if name == "" {
					name = "(native)"
				}
				return nil, fmt.Errorf("Arity mismatch calling %s: expected %d, got %d", name, target.native.Arity, provided)
			}
			if provided < target.native.Arity {
				evalArgs := vm.prepareResolvedFunctionCallArgsWithOptionalReceiver(
					args,
					bytecodeCallTargetNeedsStableArgs(callee),
					target.injectedReceiver,
					true,
				)
				return makePartialFunctionValue(partialTarget, evalArgs, callNode), nil
			}
		}
		result, _, err := vm.execExactNativeCall(target, args, callNode)
		return result, err
	}
	if fn, partialTarget, injectedReceiver, hasInjectedReceiver, ok := bytecodeResolveDirectFunctionCallTarget(callee); ok {
		if !hasInjectedReceiver {
			injectedReceiver = receiver
			hasInjectedReceiver = true
		}
		evalArgs := vm.prepareResolvedFunctionCallArgsWithOptionalReceiver(
			args,
			bytecodeCallTargetNeedsStableArgs(callee),
			injectedReceiver,
			hasInjectedReceiver,
		)
		return vm.interp.callResolvedFunctionValue(fn, partialTarget, evalArgs, vm.env, callNode, true)
	}
	if overloads, partialTarget, injectedReceiver, ok := bytecodeResolveInjectedOverloads(callee, receiver); ok {
		evalArgs := vm.prepareResolvedFunctionCallArgsWithOptionalReceiver(
			args,
			bytecodeCallTargetNeedsStableArgs(callee),
			injectedReceiver,
			true,
		)
		if len(evalArgs) < minArgsForOverloads(overloads) {
			return makePartialFunctionValue(partialTarget, evalArgs, callNode), nil
		}
		selected, err := vm.interp.selectRuntimeOverload(overloads, evalArgs, callNode)
		if err != nil {
			return nil, err
		}
		if selected == nil {
			return nil, fmt.Errorf("No overloads of %s match provided arguments", overloadName(callNode))
		}
		return vm.interp.callResolvedFunctionValue(selected, selected, evalArgs, vm.env, callNode, true)
	}
	preparedArgs := vm.prepareMaterializedCallArgs(args, false, bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion)
	return vm.interp.callCallableValueWithInjectedReceiver(
		callee,
		vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion, receiver),
		preparedArgs,
		vm.env,
		callNode,
		true,
	)
}
