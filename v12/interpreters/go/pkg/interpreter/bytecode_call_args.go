package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

const bytecodeInlinePreparedCallArgStorage = overloadArgSignatureInlineLimit + 2
const bytecodeResolvedCallArgRetainLimit = 32

func bytecodeResolvedCallArgAt(stack []runtime.Value, argBase int, idx int, injectedReceiver runtime.Value, hasInjectedReceiver bool) runtime.Value {
	if hasInjectedReceiver {
		if idx == 0 {
			return bytecodeSlotReadValue(injectedReceiver)
		}
		idx--
	}
	return bytecodeSlotReadValue(stack[argBase+idx])
}

func (i *Interpreter) populateCallTypeArgumentsFromBytecodeResolvedCallArgs(funcNode ast.Node, call *ast.FunctionCall, stack []runtime.Value, argBase int, argCount int, injectedReceiver runtime.Value, hasInjectedReceiver bool) error {
	if funcNode == nil || call == nil {
		return nil
	}
	plan := i.functionCallGenericPlan(funcNode)
	if plan == nil || plan.expectedCount == 0 {
		return nil
	}
	if len(call.TypeArguments) > 0 {
		if !i.callHasExplicitTypeArguments(call) {
			goto infer
		}
		if len(call.TypeArguments) != plan.expectedCount {
			return fmt.Errorf("Type arguments count mismatch calling %s: expected %d, got %d", plan.functionName, plan.expectedCount, len(call.TypeArguments))
		}
		return nil
	}
infer:
	start := plan.skipLeadingRuntimeArgs
	totalArgs := argCount
	if hasInjectedReceiver {
		totalArgs++
	}
	available := totalArgs - start
	max := 0
	if available > 0 {
		for _, param := range plan.inferenceRelevantParams {
			if param.argIndex >= available {
				break
			}
			max++
		}
	}
	i.inferAndSetCallTypeArgumentsFromValues(plan, funcNode, call, max, func(idx int) runtime.Value {
		return bytecodeResolvedCallArgAt(
			stack,
			argBase,
			start+plan.inferenceRelevantParams[idx].argIndex,
			injectedReceiver,
			hasInjectedReceiver,
		)
	})
	return nil
}

func resolveMethodSetReceiverFromBytecodeResolvedCallArgs(def *ast.FunctionDefinition, stack []runtime.Value, argBase int, argCount int, injectedReceiver runtime.Value, hasInjectedReceiver bool) (runtime.Value, bool) {
	if def == nil || !functionDefinitionExpectsSelf(def) {
		return nil, false
	}
	if argCount == 0 && !hasInjectedReceiver {
		return nil, false
	}
	return bytecodeResolvedCallArgAt(stack, argBase, 0, injectedReceiver, hasInjectedReceiver), true
}

func bytecodePrepareCallArgsWithOptionalReceiverIntoBuffer(buf []runtime.Value, args []runtime.Value, needsStableCopy bool, receiver runtime.Value, hasReceiver bool) ([]runtime.Value, bool) {
	if !hasReceiver && !needsStableCopy {
		return nil, false
	}
	total := len(args)
	if hasReceiver {
		total++
		if !needsStableCopy && cap(args) > len(args) {
			return nil, false
		}
	}
	if total == 0 || total > len(buf) {
		return nil, false
	}
	prepared := buf[:total]
	dst := prepared
	if hasReceiver {
		prepared[0] = bytecodeSlotReadValue(receiver)
		dst = prepared[1:]
	}
	copy(dst, args)
	return prepared, true
}

func bytecodePrepareCallArgsIntoBuffer(buf []runtime.Value, args []runtime.Value, needsStableCopy bool) ([]runtime.Value, bool) {
	if !needsStableCopy {
		return nil, false
	}
	if len(args) == 0 || len(args) > len(buf) {
		return nil, false
	}
	prepared := buf[:len(args)]
	copy(prepared, args)
	return prepared, true
}

func bytecodePrepareCallArgsWithOptionalReceiver(args []runtime.Value, needsStableCopy bool, receiver runtime.Value, hasReceiver bool) []runtime.Value {
	if !hasReceiver {
		if !needsStableCopy {
			return args
		}
		return copyCallArgs(args)
	}
	receiver = bytecodeSlotReadValue(receiver)
	if !needsStableCopy {
		return prependReceiverCallArgs(receiver, args, true)
	}
	merged := make([]runtime.Value, len(args)+1)
	merged[0] = receiver
	copy(merged[1:], args)
	return merged
}

func (vm *bytecodeVM) resolvedCallArgBuffer(size int) []runtime.Value {
	if vm == nil || size <= 0 {
		return nil
	}
	if size <= len(vm.resolvedCallArgsInline) {
		return vm.resolvedCallArgsInline[:size]
	}
	if cap(vm.resolvedCallArgsSpill) < size {
		vm.resolvedCallArgsSpill = make([]runtime.Value, size)
	} else {
		vm.resolvedCallArgsSpill = vm.resolvedCallArgsSpill[:size]
	}
	return vm.resolvedCallArgsSpill
}

func (vm *bytecodeVM) resetResolvedCallArgScratch() {
	if vm == nil {
		return
	}
	clear(vm.resolvedCallArgsInline[:])
	if cap(vm.resolvedCallArgsSpill) > bytecodeResolvedCallArgRetainLimit {
		vm.resolvedCallArgsSpill = nil
		return
	}
	if len(vm.resolvedCallArgsSpill) > 0 {
		clear(vm.resolvedCallArgsSpill)
		vm.resolvedCallArgsSpill = vm.resolvedCallArgsSpill[:0]
	}
}

// Resolved-function paths do not retain argument slices after return:
// partial application copies bound args, and direct invocation only consumes
// the slice for the duration of the call. That lets the VM reuse a small
// scratch buffer here without changing generic callable/native behavior.
func (vm *bytecodeVM) prepareResolvedFunctionCallArgsWithOptionalReceiver(args []runtime.Value, needsStableCopy bool, receiver runtime.Value, hasReceiver bool) []runtime.Value {
	if vm == nil {
		return bytecodePrepareCallArgsWithOptionalReceiver(args, needsStableCopy, receiver, hasReceiver)
	}
	if !hasReceiver {
		if !needsStableCopy {
			return args
		}
		if len(args) == 0 {
			return args
		}
		prepared := vm.resolvedCallArgBuffer(len(args))
		copy(prepared, args)
		return prepared
	}
	receiver = bytecodeSlotReadValue(receiver)
	if !needsStableCopy && cap(args) > len(args) {
		return prependReceiverCallArgs(receiver, args, true)
	}
	prepared := vm.resolvedCallArgBuffer(len(args) + 1)
	prepared[0] = receiver
	copy(prepared[1:], args)
	return prepared
}
