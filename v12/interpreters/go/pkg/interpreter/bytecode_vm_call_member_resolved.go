package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func bytecodeResolveExactInjectedNativeCallTarget(callable runtime.Value, receiver runtime.Value, explicitArgCount int) (bytecodeExactNativeCallTarget, bool) {
	switch fn := callable.(type) {
	case runtime.NativeFunctionValue:
		if fn.Arity >= 0 && explicitArgCount != fn.Arity {
			return bytecodeExactNativeCallTarget{}, false
		}
		return bytecodeExactNativeCallTarget{
			native:           fn,
			injectedReceiver: receiver,
			hasReceiver:      true,
		}, true
	case *runtime.NativeFunctionValue:
		if fn == nil {
			return bytecodeExactNativeCallTarget{}, false
		}
		if fn.Arity >= 0 && explicitArgCount != fn.Arity {
			return bytecodeExactNativeCallTarget{}, false
		}
		return bytecodeExactNativeCallTarget{
			native:           *fn,
			injectedReceiver: receiver,
			hasReceiver:      true,
		}, true
	case runtime.NativeBoundMethodValue:
		if fn.Method.Arity >= 0 && explicitArgCount != fn.Method.Arity {
			return bytecodeExactNativeCallTarget{}, false
		}
		return bytecodeExactNativeCallTarget{
			native:           fn.Method,
			injectedReceiver: fn.Receiver,
			hasReceiver:      true,
		}, true
	case *runtime.NativeBoundMethodValue:
		if fn == nil {
			return bytecodeExactNativeCallTarget{}, false
		}
		if fn.Method.Arity >= 0 && explicitArgCount != fn.Method.Arity {
			return bytecodeExactNativeCallTarget{}, false
		}
		return bytecodeExactNativeCallTarget{
			native:           fn.Method,
			injectedReceiver: fn.Receiver,
			hasReceiver:      true,
		}, true
	default:
		return bytecodeExactNativeCallTarget{}, false
	}
}

func bytecodeFunctionAcceptsStaticExplicitArgs(fn *runtime.FunctionValue, explicitArgCount int) bool {
	if fn == nil || explicitArgCount < 0 {
		return false
	}
	switch decl := fn.Declaration.(type) {
	case *ast.FunctionDefinition:
		if functionDefinitionExpectsSelf(decl) {
			return false
		}
		paramCount := len(decl.Params)
		return arityMatchesRuntime(paramCount, explicitArgCount, paramCount > 0 && isNullableParam(decl.Params[paramCount-1]))
	case *ast.LambdaExpression:
		paramCount := len(decl.Params)
		return arityMatchesRuntime(paramCount, explicitArgCount, paramCount > 0 && isNullableParam(decl.Params[paramCount-1]))
	default:
		return false
	}
}

func bytecodeCallableAcceptsStaticExplicitArgs(callable runtime.Value, explicitArgCount int) bool {
	switch fn := callable.(type) {
	case runtime.NativeFunctionValue:
		return fn.Arity < 0 || fn.Arity == explicitArgCount
	case *runtime.NativeFunctionValue:
		return fn != nil && (fn.Arity < 0 || fn.Arity == explicitArgCount)
	}
	for _, fn := range functionOverloadsView(callable) {
		if bytecodeFunctionAcceptsStaticExplicitArgs(fn, explicitArgCount) {
			return true
		}
	}
	return false
}

func (vm *bytecodeVM) resolveDirectStaticMemberCallable(receiver runtime.Value, memberName string, explicitArgCount int) (runtime.Value, bool, error) {
	if vm == nil || vm.interp == nil || memberName == "" {
		return nil, false, nil
	}
	singletonReceiver := false
	switch def := receiver.(type) {
	case *runtime.StructDefinitionValue:
		if def != nil && isSingletonStructDef(def.Node) {
			singletonReceiver = true
		}
	case runtime.StructDefinitionValue:
		if isSingletonStructDef(def.Node) {
			singletonReceiver = true
		}
	}
	callable, found, err := vm.interp.resolveStaticMemberCallable(receiver, memberName)
	if !singletonReceiver {
		return callable, found, err
	}
	if err != nil || !found || !bytecodeCallableAcceptsStaticExplicitArgs(callable, explicitArgCount) {
		return nil, false, nil
	}
	return callable, true, nil
}

func (vm *bytecodeVM) callResolvedCallableWithInjectedReceiver(callable runtime.Value, receiver runtime.Value, explicitArgs []runtime.Value, callNode *ast.FunctionCall) (runtime.Value, error) {
	if vm == nil || vm.interp == nil {
		return nil, fmt.Errorf("bytecode VM is nil")
	}
	return vm.callCallableValueWithInjectedReceiver(callable, receiver, explicitArgs, callNode)
}

func (vm *bytecodeVM) tryCallResolvedCallableFromMemberStack(callable runtime.Value, receiver runtime.Value, receiverIndex int, argBase int, argCount int, callNode *ast.FunctionCall) (runtime.Value, bool, error) {
	if vm == nil || vm.interp == nil {
		return nil, false, fmt.Errorf("bytecode VM is nil")
	}
	if receiverIndex < 0 || argBase != receiverIndex+1 || argCount < 0 || argBase+argCount > vm.stackDepth() {
		return nil, false, fmt.Errorf("bytecode stack underflow")
	}
	if fn, ok := callable.(*runtime.FunctionValue); ok {
		if fn == nil {
			return nil, false, nil
		}
		evalArgs := vm.stackValues(receiverIndex, argBase+argCount)
		vm.truncateStack(receiverIndex)
		result, err := vm.interp.callResolvedFunctionValue(fn, fn, evalArgs, vm.env, callNode, true)
		return result, true, err
	}
	fn, partialTarget, injectedReceiver, ok := bytecodeResolvedDirectFunctionCallTarget(callable, receiver)
	if !ok || fn == nil || !bytecodeCanReuseResolvedMemberStackReceiver(receiver, injectedReceiver) {
		return nil, false, nil
	}
	evalArgs := vm.stackValues(receiverIndex, argBase+argCount)
	vm.truncateStack(receiverIndex)
	result, err := vm.interp.callResolvedFunctionValue(fn, partialTarget, evalArgs, vm.env, callNode, true)
	return result, true, err
}

func bytecodeResolvedDirectFunctionCallTarget(callable runtime.Value, receiver runtime.Value) (*runtime.FunctionValue, runtime.Value, runtime.Value, bool) {
	switch fn := callable.(type) {
	case *runtime.FunctionValue:
		if fn == nil {
			return nil, nil, nil, false
		}
		return fn, fn, receiver, true
	case runtime.BoundMethodValue:
		methodFn, ok := fn.Method.(*runtime.FunctionValue)
		if !ok || methodFn == nil {
			return nil, nil, nil, false
		}
		return methodFn, fn.Method, fn.Receiver, true
	case *runtime.BoundMethodValue:
		if fn == nil {
			return nil, nil, nil, false
		}
		methodFn, ok := fn.Method.(*runtime.FunctionValue)
		if !ok || methodFn == nil {
			return nil, nil, nil, false
		}
		return methodFn, fn.Method, fn.Receiver, true
	default:
		return nil, nil, nil, false
	}
}
