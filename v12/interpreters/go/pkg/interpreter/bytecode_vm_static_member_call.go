package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execLoadStaticReceiver(instr bytecodeInstruction, currentProgram *bytecodeProgram) error {
	if instr.name == "" {
		return fmt.Errorf("bytecode static receiver missing name")
	}
	var (
		val runtime.Value
		ok  bool
	)
	if instr.nameSimple {
		lookup, found := vm.lookupIdentifierNameForCallCache(currentProgram, vm.ip, instr.name)
		if found {
			val = lookup.value
			ok = true
		}
	} else {
		val, ok = vm.lookupCachedName(currentProgram, vm.ip, instr.name)
	}
	if !ok {
		err := fmt.Errorf("Undefined variable '%s'", instr.name)
		if instr.node != nil {
			err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return err
	}
	vm.appendStackValue(val)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execCallStaticMember(instr bytecodeInstruction, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if instr.argCount < 0 {
		return nil, fmt.Errorf("bytecode static member arg count invalid")
	}
	if vm.stackDepth() < instr.argCount+1 {
		return nil, fmt.Errorf("bytecode stack underflow")
	}
	if instr.name == "" {
		return nil, fmt.Errorf("bytecode static member missing member name")
	}
	receiverIndex := vm.stackDepth() - instr.argCount - 1
	argBase := receiverIndex + 1
	receiver := vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion, vm.stackValue(receiverIndex))
	vm.setStackValue(receiverIndex, receiver)

	var callNode *ast.FunctionCall
	if instr.node != nil {
		if call, ok := instr.node.(*ast.FunctionCall); ok {
			callNode = call
		}
	}
	traceNode := instr.node
	if callNode != nil {
		traceNode = callNode
	}
	statsEnabled := vm.interp != nil && vm.interp.bytecodeStatsEnabled

	if _, ok := bytecodeStaticMemberReceiverIdentityForValue(receiver); ok {
		if cached, ok := vm.lookupCachedStaticMemberCall(currentProgram, vm.ip, instr.name, instr.argCount, receiver); ok {
			return vm.execStaticMemberCallable(cached.callable, instr, receiverIndex, argBase, callNode, traceNode, currentProgram, statsEnabled)
		}
		if callee, found, err := vm.resolveDirectStaticMemberCallable(receiver, instr.name, instr.argCount); err != nil {
			return nil, vm.attachBytecodeRuntimeContext(err, callNode, nil)
		} else if found {
			vm.storeCachedStaticMemberCall(currentProgram, vm.ip, instr.name, instr.argCount, receiver, callee)
			return vm.execStaticMemberCallable(callee, instr, receiverIndex, argBase, callNode, traceNode, currentProgram, statsEnabled)
		}
	}

	fallback := bytecodeCallMemberInstructionForName(instr.name, instr.argCount, instr.node)
	return vm.execCallMember(fallback, currentProgram)
}
