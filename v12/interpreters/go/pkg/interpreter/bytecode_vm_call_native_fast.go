package interpreter

import (
	"errors"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type bytecodeExactNativeCallTarget struct {
	native           runtime.NativeFunctionValue
	injectedReceiver runtime.Value
	hasReceiver      bool
}

func bytecodeResolveExactNativeCallTarget(callee runtime.Value, explicitArgCount int) (bytecodeExactNativeCallTarget, bool) {
	switch fn := callee.(type) {
	case runtime.NativeFunctionValue:
		if fn.Arity >= 0 && explicitArgCount != fn.Arity {
			return bytecodeExactNativeCallTarget{}, false
		}
		return bytecodeExactNativeCallTarget{native: fn}, true
	case *runtime.NativeFunctionValue:
		if fn == nil {
			return bytecodeExactNativeCallTarget{}, false
		}
		if fn.Arity >= 0 && explicitArgCount != fn.Arity {
			return bytecodeExactNativeCallTarget{}, false
		}
		return bytecodeExactNativeCallTarget{native: *fn}, true
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

func (vm *bytecodeVM) tryExecExactNativeCall(callee runtime.Value, explicitArgs []runtime.Value, callNode *ast.FunctionCall) (runtime.Value, bool, error) {
	target, ok := bytecodeResolveExactNativeCallTarget(callee, len(explicitArgs))
	if !ok {
		return nil, false, nil
	}
	return vm.execExactNativeCall(target, explicitArgs, callNode)
}

func (vm *bytecodeVM) execExactNativeCall(target bytecodeExactNativeCallTarget, explicitArgs []runtime.Value, callNode *ast.FunctionCall) (runtime.Value, bool, error) {
	if vm == nil || vm.interp == nil {
		return nil, false, nil
	}
	if callNode != nil {
		state := vm.interp.stateFromEnv(vm.env)
		state.pushCallFrame(callNode)
		defer state.popCallFrame()
	}
	var callState any
	if vm.env != nil {
		callState = vm.env.RuntimeData()
	}
	ctx := vm.interp.acquireNativeCallContext(vm.env, callState)
	defer vm.interp.releaseNativeCallContext(ctx)
	if !target.hasReceiver {
		args := explicitArgs
		if !target.native.BorrowArgs {
			args = copyCallArgs(explicitArgs)
		}
		result, err := target.native.Impl(ctx, args)
		return result, true, err
	}
	result, err := bytecodeExecExactNativeBoundCall(ctx, target.native, target.injectedReceiver, explicitArgs)
	return result, true, err
}

func bytecodeExecExactNativeBoundCall(ctx *runtime.NativeCallContext, native runtime.NativeFunctionValue, receiver runtime.Value, explicitArgs []runtime.Value) (runtime.Value, error) {
	if !native.BorrowArgs {
		args := make([]runtime.Value, len(explicitArgs)+1)
		args[0] = receiver
		copy(args[1:], explicitArgs)
		return native.Impl(ctx, args)
	}
	switch len(explicitArgs) {
	case 0:
		args := [1]runtime.Value{receiver}
		return native.Impl(ctx, args[:])
	case 1:
		args := [2]runtime.Value{receiver, explicitArgs[0]}
		return native.Impl(ctx, args[:])
	case 2:
		args := [3]runtime.Value{receiver, explicitArgs[0], explicitArgs[1]}
		return native.Impl(ctx, args[:])
	case 3:
		args := [4]runtime.Value{receiver, explicitArgs[0], explicitArgs[1], explicitArgs[2]}
		return native.Impl(ctx, args[:])
	default:
		args := make([]runtime.Value, len(explicitArgs)+1)
		args[0] = receiver
		copy(args[1:], explicitArgs)
		return native.Impl(ctx, args)
	}
}

func (vm *bytecodeVM) bytecodeEvalState() *evalState {
	if vm == nil || vm.interp == nil {
		return nil
	}
	return vm.interp.stateFromEnv(vm.env)
}

func (vm *bytecodeVM) attachBytecodeRuntimeContext(err error, node ast.Node, state *evalState) error {
	if err == nil || vm == nil || vm.interp == nil {
		return err
	}
	if state == nil {
		state = vm.bytecodeEvalState()
	}
	return vm.interp.attachRuntimeContext(err, node, state)
}

func (vm *bytecodeVM) finishCompletedCall(result runtime.Value, err error, callNode *ast.FunctionCall, state *evalState) (*bytecodeProgram, error) {
	if err != nil {
		if errors.Is(err, errSerialYield) {
			payload := payloadFromState(vm.env.RuntimeData())
			if payload == nil || !payload.awaitBlocked {
				vm.stack = append(vm.stack, runtime.NilValue{})
				vm.ip++
			}
			return nil, err
		}
		err = vm.attachBytecodeRuntimeContext(err, callNode, state)
		if vm.handleLoopSignal(err) {
			return nil, nil
		}
		return nil, err
	}
	if result == nil {
		result = runtime.NilValue{}
	}
	vm.stack = append(vm.stack, result)
	vm.ip++
	return nil, nil
}
