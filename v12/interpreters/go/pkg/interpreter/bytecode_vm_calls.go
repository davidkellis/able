package interpreter

import (
	"errors"
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

// tryInlineCall attempts to set up an inline call frame for a slot-enabled
// function value. Returns the new program to switch to, or nil if the
// function cannot be inlined (the caller should fall back to callCallableValue).
func (vm *bytecodeVM) tryInlineCall(callee runtime.Value, args []runtime.Value, callNode *ast.FunctionCall, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	fn, ok := callee.(*runtime.FunctionValue)
	if !ok || fn == nil {
		return nil, nil
	}
	prog, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || prog == nil || prog.frameLayout == nil {
		return nil, nil
	}
	decl, ok := fn.Declaration.(*ast.FunctionDefinition)
	if !ok || decl == nil || decl.Body == nil {
		return nil, nil
	}
	// Method shorthand requires implicit receiver adjustment and extra arg.
	if decl.IsMethodShorthand {
		return nil, nil
	}
	// Require exact arity match (skip optional params).
	paramCount := len(decl.Params)
	if len(args) != paramCount {
		return nil, nil
	}
	// Skip if call site has type arguments that need binding.
	if callNode != nil && len(callNode.TypeArguments) > 0 {
		return nil, nil
	}
	// Skip if function belongs to a generic method set.
	if fn.MethodSet != nil && len(fn.MethodSet.GenericParams) > 0 {
		return nil, nil
	}

	layout := prog.frameLayout
	slots := make([]runtime.Value, layout.slotCount)

	// Coerce parameters (skip for params that use generic type names,
	// matching the behavior in invokeFunction). Use fast path when the
	// value already trivially matches the declared type.
	var generics map[string]struct{}
	if len(decl.GenericParams) > 0 {
		generics = functionGenericNameSet(fn, decl)
	}
	for idx, param := range decl.Params {
		arg := args[idx]
		if param.ParamType != nil && !paramUsesGeneric(param.ParamType, generics) && !inlineCoercionUnnecessary(param.ParamType, arg) {
			coerced, err := vm.interp.coerceValueToType(param.ParamType, arg)
			if err != nil {
				return nil, err
			}
			arg = coerced
		}
		slots[idx] = arg
	}

	// Push implicit receiver only when the function body uses #member syntax.
	hasImplicit := paramCount > 0 && layout.usesImplicitMember
	if hasImplicit {
		state := vm.interp.stateFromEnv(fn.Closure)
		state.pushImplicitReceiver(args[0])
	}

	// Push call frame.
	vm.callFrames = append(vm.callFrames, bytecodeCallFrame{
		returnIP:            vm.ip + 1,
		program:             currentProgram,
		slots:               vm.slots,
		env:                 vm.env,
		iterBase:            len(vm.iterStack),
		loopBase:            len(vm.loopStack),
		hasImplicitReceiver: hasImplicit,
	})

	// Set up new frame.
	vm.slots = slots
	vm.env = fn.Closure
	vm.ip = 0

	return prog, nil
}

// execCall handles bytecodeOpCall. It returns a non-nil program when an
// inline call frame was set up (the caller must switch to the new program).
// A nil program with nil error means the call completed normally.
func (vm *bytecodeVM) execCall(instr bytecodeInstruction, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if instr.argCount < 0 {
		return nil, fmt.Errorf("bytecode call arg count invalid")
	}
	args := make([]runtime.Value, instr.argCount)
	for idx := instr.argCount - 1; idx >= 0; idx-- {
		arg, err := vm.pop()
		if err != nil {
			return nil, err
		}
		args[idx] = arg
	}
	callee, err := vm.pop()
	if err != nil {
		return nil, err
	}
	var callNode *ast.FunctionCall
	if instr.node != nil {
		if call, ok := instr.node.(*ast.FunctionCall); ok {
			callNode = call
		}
	}
	// Try inline call.
	if newProg, err := vm.tryInlineCall(callee, args, callNode, currentProgram); err != nil {
		return nil, err
	} else if newProg != nil {
		return newProg, nil
	}
	// Normal call.
	result, err := vm.interp.callCallableValue(callee, args, vm.env, callNode)
	if err != nil {
		if errors.Is(err, errSerialYield) {
			payload := payloadFromState(vm.env.RuntimeData())
			if payload == nil || !payload.awaitBlocked {
				vm.stack = append(vm.stack, runtime.NilValue{})
				vm.ip++
			}
			return nil, err
		}
		err = vm.interp.attachRuntimeContext(err, callNode, vm.interp.stateFromEnv(vm.env))
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

// execCallName handles bytecodeOpCallName. It returns a non-nil program when
// an inline call frame was set up (the caller must switch to the new program).
// A nil program with nil error means the call completed normally.
func (vm *bytecodeVM) execCallName(instr bytecodeInstruction, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if instr.argCount < 0 {
		return nil, fmt.Errorf("bytecode call arg count invalid")
	}
	args := make([]runtime.Value, instr.argCount)
	for idx := instr.argCount - 1; idx >= 0; idx-- {
		arg, err := vm.pop()
		if err != nil {
			return nil, err
		}
		args[idx] = arg
	}
	if instr.name == "" {
		return nil, fmt.Errorf("bytecode call missing target name")
	}
	var callNode *ast.FunctionCall
	if instr.node != nil {
		if call, ok := instr.node.(*ast.FunctionCall); ok {
			callNode = call
		}
	}
	state := vm.interp.stateFromEnv(vm.env)
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
				return nil, vm.interp.attachRuntimeContext(err, callNode, state)
			}
			calleeVal = candidate
		} else {
			return nil, vm.interp.attachRuntimeContext(err, callNode, state)
		}
	}
	// Try inline call.
	if newProg, err := vm.tryInlineCall(calleeVal, args, callNode, currentProgram); err != nil {
		return nil, err
	} else if newProg != nil {
		return newProg, nil
	}
	// Normal call.
	result, err := vm.interp.callCallableValue(calleeVal, args, vm.env, callNode)
	if err != nil {
		if errors.Is(err, errSerialYield) {
			payload := payloadFromState(vm.env.RuntimeData())
			if payload == nil || !payload.awaitBlocked {
				vm.stack = append(vm.stack, runtime.NilValue{})
				vm.ip++
			}
			return nil, err
		}
		err = vm.interp.attachRuntimeContext(err, callNode, state)
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
