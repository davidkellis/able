package interpreter

import (
	"errors"
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execCall(instr bytecodeInstruction) error {
	if instr.argCount < 0 {
		return fmt.Errorf("bytecode call arg count invalid")
	}
	args := make([]runtime.Value, instr.argCount)
	for idx := instr.argCount - 1; idx >= 0; idx-- {
		arg, err := vm.pop()
		if err != nil {
			return err
		}
		args[idx] = arg
	}
	callee, err := vm.pop()
	if err != nil {
		return err
	}
	var callNode *ast.FunctionCall
	if instr.node != nil {
		if call, ok := instr.node.(*ast.FunctionCall); ok {
			callNode = call
		}
	}
	result, err := vm.interp.callCallableValue(callee, args, vm.env, callNode)
	if err != nil {
		if errors.Is(err, errSerialYield) {
			payload := payloadFromState(vm.env.RuntimeData())
			if payload == nil || !payload.awaitBlocked {
				vm.stack = append(vm.stack, runtime.NilValue{})
				vm.ip++
			}
			return err
		}
		err = vm.interp.attachRuntimeContext(err, callNode, vm.interp.stateFromEnv(vm.env))
		if vm.handleLoopSignal(err) {
			return nil
		}
		return err
	}
	if result == nil {
		result = runtime.NilValue{}
	}
	vm.stack = append(vm.stack, result)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execCallName(instr bytecodeInstruction) error {
	if instr.argCount < 0 {
		return fmt.Errorf("bytecode call arg count invalid")
	}
	args := make([]runtime.Value, instr.argCount)
	for idx := instr.argCount - 1; idx >= 0; idx-- {
		arg, err := vm.pop()
		if err != nil {
			return err
		}
		args[idx] = arg
	}
	if instr.name == "" {
		return fmt.Errorf("bytecode call missing target name")
	}
	calleeVal, err := vm.env.Get(instr.name)
	if err != nil {
		if dotIdx := strings.Index(instr.name, "."); dotIdx > 0 && dotIdx < len(instr.name)-1 {
			head := instr.name[:dotIdx]
			tail := instr.name[dotIdx+1:]
			receiver, recvErr := vm.env.Get(head)
			if recvErr != nil {
				if def, ok := vm.env.StructDefinition(head); ok {
					receiver = def
				} else {
					receiver = runtime.TypeRefValue{TypeName: head}
				}
			}
			member := ast.ID(tail)
			candidate, err := vm.interp.memberAccessOnValueWithOptions(receiver, member, vm.env, true)
			if err != nil {
				return err
			}
			calleeVal = candidate
		} else {
			return err
		}
	}
	var callNode *ast.FunctionCall
	if instr.node != nil {
		if call, ok := instr.node.(*ast.FunctionCall); ok {
			callNode = call
		}
	}
	result, err := vm.interp.callCallableValue(calleeVal, args, vm.env, callNode)
	if err != nil {
		if errors.Is(err, errSerialYield) {
			payload := payloadFromState(vm.env.RuntimeData())
			if payload == nil || !payload.awaitBlocked {
				vm.stack = append(vm.stack, runtime.NilValue{})
				vm.ip++
			}
			return err
		}
		err = vm.interp.attachRuntimeContext(err, callNode, vm.interp.stateFromEnv(vm.env))
		if vm.handleLoopSignal(err) {
			return nil
		}
		return err
	}
	if result == nil {
		result = runtime.NilValue{}
	}
	vm.stack = append(vm.stack, result)
	vm.ip++
	return nil
}
