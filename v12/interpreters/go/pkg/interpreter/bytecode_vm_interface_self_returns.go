package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execInterfaceSelfReturnMember(
	instr bytecodeInstruction,
	receiver *runtime.InterfaceValue,
	receiverIndex int,
	argBase int,
	callNode *ast.FunctionCall,
) (*bytecodeProgram, error) {
	if vm == nil || vm.interp == nil || receiver == nil {
		return nil, fmt.Errorf("bytecode interface Self-return call is unavailable")
	}
	receiverTypeHint := vm.interp.staticReceiverTypeForCall(callNode, vm.env)
	callable, injectedReceiver, hasInjectedReceiver, found, err := vm.interp.resolveDirectCallMemberCallable(vm.env, receiver, instr.name, receiverTypeHint)
	if err != nil {
		return nil, vm.attachBytecodeRuntimeContext(err, callNode, nil)
	}
	if !found {
		callable, err = vm.interp.memberAccessOnValueWithOptions(receiver, ast.NewIdentifier(instr.name), vm.env, true)
		if err != nil {
			return nil, vm.attachBytecodeRuntimeContext(err, callNode, nil)
		}
	}
	args := append([]runtime.Value(nil), vm.stackValuesFrom(argBase)...)
	vm.truncateStack(receiverIndex)
	args = vm.prepareMaterializedCallArgs(args, true, bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion)
	var result runtime.Value
	if hasInjectedReceiver {
		result, err = vm.callCallableValueWithInjectedReceiver(callable, injectedReceiver, args, callNode)
	} else {
		result, err = vm.callCallableValueMutable(callable, args, callNode)
	}
	if err == nil {
		result = preserveInterfaceSelfReturn(receiver, result)
	}
	return vm.finishCompletedCall(result, err, callNode, nil)
}
