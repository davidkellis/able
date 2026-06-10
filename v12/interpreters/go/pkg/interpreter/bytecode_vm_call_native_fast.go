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
		callState = vm.runtimeData()
	}
	var ctx *runtime.NativeCallContext
	if !target.native.SkipContext {
		ctx = vm.interp.acquireNativeCallContext(vm.env, callState)
		defer vm.interp.releaseNativeCallContext(ctx)
	}
	if target.native.RawImpl != nil {
		result, err := vm.execExactNativeRawCall(ctx, target, explicitArgs)
		if err != nil {
			return nil, true, err
		}
		return bytecodeValueFromRuntimeRawValue(result), true, nil
	}
	vm.recordPrimitiveMaterializationValues(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonHostNative, explicitArgs)
	bytecodeMaterializeRawFloatArgs(explicitArgs)
	if !target.hasReceiver {
		args := explicitArgs
		if !target.native.BorrowArgs {
			args = copyCallArgs(explicitArgs)
		}
		result, err := target.native.Impl(ctx, args)
		return result, true, err
	}
	var scratch *nativeBorrowCallArgScratch
	if target.native.BorrowArgs {
		scratch = vm.interp.acquireNativeBorrowCallArgScratch()
		defer vm.interp.releaseNativeBorrowCallArgScratch(scratch)
		vm.interp.recordBytecodePrimitiveMaterialization(vm, bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonHostNative, target.injectedReceiver)
	}
	result, err := bytecodeExecExactNativeBoundCall(ctx, scratch, target.native, target.injectedReceiver, explicitArgs)
	return result, true, err
}

func (vm *bytecodeVM) execExactNativeRawResultCall(target bytecodeExactNativeCallTarget, explicitArgs []runtime.Value, callNode *ast.FunctionCall) (runtime.RawValue, bool, error) {
	if vm == nil || vm.interp == nil || target.native.RawImpl == nil {
		return runtime.RawValue{}, false, nil
	}
	if callNode != nil {
		state := vm.interp.stateFromEnv(vm.env)
		state.pushCallFrame(callNode)
		defer state.popCallFrame()
	}
	var callState any
	if vm.env != nil {
		callState = vm.runtimeData()
	}
	var ctx *runtime.NativeCallContext
	if !target.native.SkipContext {
		ctx = vm.interp.acquireNativeCallContext(vm.env, callState)
		defer vm.interp.releaseNativeCallContext(ctx)
	}
	result, err := vm.execExactNativeRawCall(ctx, target, explicitArgs)
	return result, true, err
}

func (vm *bytecodeVM) execExactNativeRawCall(ctx *runtime.NativeCallContext, target bytecodeExactNativeCallTarget, explicitArgs []runtime.Value) (runtime.RawValue, error) {
	if target.native.RawImpl == nil {
		return runtime.RawValue{}, nil
	}
	argCount := len(explicitArgs)
	if target.hasReceiver {
		argCount++
	}
	args, pooled := vm.acquireNativeRawArgs(argCount)
	bytecodeFillExactNativeRawArgs(args, target, explicitArgs)
	result, err := target.native.RawImpl(ctx, args)
	vm.releaseNativeRawArgs(args, pooled)
	if err != nil {
		return runtime.RawValue{}, err
	}
	return result, nil
}

func bytecodeFillExactNativeRawArgs(args []runtime.RawValue, target bytecodeExactNativeCallTarget, explicitArgs []runtime.Value) {
	offset := 0
	if target.hasReceiver {
		args[0] = bytecodeNativeRawValue(target.injectedReceiver)
		offset = 1
	}
	for idx, arg := range explicitArgs {
		args[offset+idx] = bytecodeNativeRawValue(arg)
	}
}

func (vm *bytecodeVM) acquireNativeRawArgs(count int) ([]runtime.RawValue, bool) {
	if count <= 0 {
		return nil, false
	}
	if vm == nil || vm.nativeRawArgsBusy {
		return make([]runtime.RawValue, count), false
	}
	vm.nativeRawArgsBusy = true
	if cap(vm.nativeRawArgs) < count {
		if count <= len(vm.nativeRawArgsInline) {
			vm.nativeRawArgs = vm.nativeRawArgsInline[:0]
		} else {
			vm.nativeRawArgs = make([]runtime.RawValue, count)
		}
	}
	args := vm.nativeRawArgs[:count]
	return args, true
}

func (vm *bytecodeVM) releaseNativeRawArgs(args []runtime.RawValue, pooled bool) {
	if len(args) == 0 {
		return
	}
	clear(args)
	if !pooled || vm == nil {
		return
	}
	vm.nativeRawArgs = args[:0]
	vm.nativeRawArgsBusy = false
}

func bytecodeNativeRawValue(value runtime.Value) runtime.RawValue {
	if kind, raw, ok := bytecodeRawIntegerValueInfo(value); ok {
		return runtime.NewRawIntegerValue(kind, raw)
	}
	if raw, kind, ok := bytecodeDirectFloatValue(value); ok {
		return runtime.NewRawFloatValue(kind, raw)
	}
	return runtime.NewRawValue(bytecodeMaterializeRawValue(value))
}

func bytecodePrepareExactNativeCallArgs(target bytecodeExactNativeCallTarget, args []runtime.Value) {
	if target.native.RawImpl == nil {
		bytecodeMaterializeRawFloatArgs(args)
	}
}

func bytecodeExecExactNativeBoundCall(ctx *runtime.NativeCallContext, scratch *nativeBorrowCallArgScratch, native runtime.NativeFunctionValue, receiver runtime.Value, explicitArgs []runtime.Value) (runtime.Value, error) {
	if !native.BorrowArgs {
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
	if scratch != nil {
		args := scratch.buffer(len(explicitArgs) + 1)
		args[0] = bytecodeMaterializeRawValue(receiver)
		bytecodeCopyMaterializedCallArgs(args[1:], explicitArgs)
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

func (vm *bytecodeVM) attachOrAppendBytecodeRuntimeContext(err error, callNode *ast.FunctionCall, state *evalState) error {
	if err == nil || vm == nil || vm.interp == nil {
		return err
	}
	if callNode != nil {
		if ctx := runtimeContextFromError(err); ctx != nil {
			ctx.freezeCallStack()
			for _, frame := range ctx.callStack {
				if frame.node == callNode {
					return err
				}
			}
			ctx.callStack = append([]runtimeCallFrame{{node: callNode}}, ctx.callStack...)
			return err
		}
	}
	return vm.attachBytecodeRuntimeContext(err, callNode, state)
}

func (vm *bytecodeVM) finishCompletedCall(result runtime.Value, err error, callNode *ast.FunctionCall, state *evalState) (*bytecodeProgram, error) {
	if err != nil {
		if errors.Is(err, errSerialYield) {
			payload := payloadFromState(vm.runtimeData())
			if !payload.isAwaitBlocked() {
				vm.appendStackValue(runtime.NilValue{})
				vm.ip++
			}
			return nil, err
		}
		err = vm.attachOrAppendBytecodeRuntimeContext(err, callNode, state)
		if vm.handleLoopSignal(err) {
			return nil, nil
		}
		return nil, err
	}
	if result == nil {
		result = runtime.NilValue{}
	}
	result = bytecodeSlotReadValue(result)
	vm.adoptBytecodeArrayOwnershipReturnedValue(result)
	vm.appendStackValue(result)
	vm.ip++
	return nil, nil
}

func (vm *bytecodeVM) finishCompletedRawCall(result runtime.RawValue, err error, callNode *ast.FunctionCall, state *evalState) (*bytecodeProgram, error) {
	if err != nil {
		return vm.finishCompletedCall(nil, err, callNode, state)
	}
	if vm.appendRuntimeRawValue(result) {
		vm.ip++
		return nil, nil
	}
	return vm.finishCompletedCall(bytecodeValueFromRuntimeRawValue(result), nil, callNode, state)
}

func (vm *bytecodeVM) execAndFinishExactNativeCall(target bytecodeExactNativeCallTarget, explicitArgs []runtime.Value, callNode *ast.FunctionCall) (*bytecodeProgram, error) {
	// Native/extern code has no bytecode ownership contract. The observer keeps
	// any locally-created Array argument alive until normal GC cleanup.
	vm.markBytecodeArrayOwnershipValuesEscaped(explicitArgs, bytecodeArrayOwnershipEscapeUnknownCall)
	if target.native.RawImpl != nil {
		result, handled, err := vm.execExactNativeRawResultCall(target, explicitArgs, callNode)
		if handled || err != nil {
			return vm.finishCompletedRawCall(result, err, callNode, nil)
		}
	}
	result, _, err := vm.execExactNativeCall(target, explicitArgs, callNode)
	return vm.finishCompletedCall(result, err, callNode, nil)
}

func (vm *bytecodeVM) finishCompletedVoidCallFast() (*bytecodeProgram, error) {
	vm.appendStackValue(runtime.VoidValue{})
	vm.ip++
	return nil, nil
}
