package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execIteratorNextCallMemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if vm == nil || vm.interp == nil || instr.argCount != 0 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() {
		return nil, false, nil
	}
	iter, ok := bytecodeIteratorNextFastReceiver(vm, vm.interp, vm.stackValue(receiverIndex))
	if !ok {
		return nil, false, nil
	}
	value, done, err := iter.NextRaw()
	if err != nil {
		vm.truncateStack(receiverIndex)
		newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
		return newProg, true, finishErr
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "member_access", "iterator_next_fast", instr.node)
	}
	vm.truncateStack(receiverIndex)
	if done {
		newProg, finishErr := vm.finishCompletedCall(runtime.IteratorEnd, nil, callNode, nil)
		return newProg, true, finishErr
	}
	if vm.appendRuntimeRawValue(value) {
		vm.ip++
		return nil, true, nil
	}
	result := bytecodeValueFromRuntimeRawValue(value)
	newProg, finishErr := vm.finishCompletedCall(result, nil, callNode, nil)
	return newProg, true, finishErr
}

func bytecodeIteratorNextFastReceiver(vm *bytecodeVM, interp *Interpreter, receiver runtime.Value) (*runtime.IteratorValue, bool) {
	if iter, ok := receiver.(*runtime.IteratorValue); ok && iter != nil {
		return iter, true
	}
	iface, ok := vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion, receiver).(*runtime.InterfaceValue)
	if !ok || iface == nil {
		return nil, false
	}
	method, ok := interfaceValueLookupMethod(iface, "next")
	if !ok || method == nil {
		return nil, false
	}
	target, ok := bytecodeResolveExactInjectedNativeCallTarget(method, interfaceMethodReceiver(interp, iface, method), 0)
	if !ok || target.native.Name != "iterator.next" {
		return nil, false
	}
	iter, ok := vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion, target.injectedReceiver).(*runtime.IteratorValue)
	if !ok || iter == nil {
		return nil, false
	}
	return iter, true
}

func (vm *bytecodeVM) execCachedNextCallMemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall, currentProgram *bytecodeProgram) (*bytecodeProgram, bool, error) {
	if vm == nil || receiverIndex < 0 || receiverIndex >= vm.stackDepth() || instr.name != "next" || instr.argCount != 0 || instr.safe {
		return nil, false, nil
	}
	receiver := vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion, vm.stackValue(receiverIndex))
	vm.setStackValue(receiverIndex, receiver)
	if !bytecodeCanDirectMemberCall(receiver, instr.name) {
		return nil, false, nil
	}
	if !vm.canUseMemberMethodCacheForReceiver(instr.name, true, receiver) {
		return nil, false, nil
	}
	cached, ok := vm.lookupCachedMemberMethodEntry(currentProgram, vm.ip, instr.name, true, receiver)
	if !ok {
		return nil, false, nil
	}
	argBase := receiverIndex + 1
	if newProg, handled, err := vm.execCallMemberFastPath(cached.fastPath, instr, receiverIndex, argBase, callNode, currentProgram, receiver); handled {
		return newProg, true, err
	}
	if cached.template == nil {
		return nil, false, nil
	}
	newProg, err := vm.execCachedResolvedMemberCall(cached, instr.name, receiverIndex, argBase, instr.argCount, callNode, currentProgram)
	return newProg, true, err
}

func bytecodeIteratorNextCallMemberValue(iter *runtime.IteratorValue) (runtime.Value, error) {
	value, done, err := iter.NextRaw()
	if err != nil {
		return nil, err
	}
	if done {
		return runtime.IteratorEnd, nil
	}
	result := bytecodeValueFromRuntimeRawValue(value)
	if result == nil {
		return runtime.NilValue{}, nil
	}
	return result, nil
}

func bytecodeValueFromRuntimeRawValue(value runtime.RawValue) runtime.Value {
	if kind, raw, ok := value.Integer(); ok {
		return bytecodeRawIntegerResultValue(kind, raw)
	}
	if kind, raw, ok := value.Float(); ok {
		return bytecodeRawFloatSlotValue(raw, kind)
	}
	return value.Materialize()
}

func (vm *bytecodeVM) appendRuntimeRawValue(value runtime.RawValue) bool {
	if vm == nil {
		return false
	}
	if kind, raw, ok := value.Integer(); ok {
		if kind == runtime.IntegerI64 {
			vm.appendRawI64Stack(raw)
			return true
		}
		vm.appendRawIntegerStack(kind, raw)
		return true
	}
	if kind, raw, ok := value.Float(); ok {
		vm.appendStackValue(bytecodeRawFloatSlotValue(raw, kind))
		return true
	}
	return false
}
