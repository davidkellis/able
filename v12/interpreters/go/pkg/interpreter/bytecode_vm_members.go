package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execIndexGet(instr bytecodeInstruction) error {
	idxVal, err := vm.pop()
	if err != nil {
		return err
	}
	obj, err := vm.pop()
	if err != nil {
		return err
	}
	result, err := vm.interp.indexGet(obj, idxVal)
	if err != nil {
		err = vm.interp.wrapStandardRuntimeError(err)
		if instr.node != nil {
			err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
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

func (vm *bytecodeVM) execIndexSet(instr bytecodeInstruction) error {
	idxVal, err := vm.pop()
	if err != nil {
		return err
	}
	obj, err := vm.pop()
	if err != nil {
		return err
	}
	val, err := vm.pop()
	if err != nil {
		return err
	}
	if instr.operator == "" {
		return fmt.Errorf("bytecode index set missing operator")
	}
	op := ast.AssignmentOperator(instr.operator)
	binaryOp, isCompound := binaryOpForAssignment(op)
	result, err := vm.interp.assignIndex(obj, idxVal, val, op, binaryOp, isCompound)
	if err != nil {
		err = vm.interp.wrapStandardRuntimeError(err)
		if instr.node != nil {
			err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
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

func (vm *bytecodeVM) execMemberAccess(instr bytecodeInstruction) error {
	obj, err := vm.pop()
	if err != nil {
		return err
	}
	if instr.safe && isNilRuntimeValue(obj) {
		vm.stack = append(vm.stack, runtime.NilValue{})
		vm.ip++
		return nil
	}
	memberExpr := ast.Expression(nil)
	if instr.node != nil {
		if member, ok := instr.node.(*ast.MemberAccessExpression); ok && member != nil {
			memberExpr = member.Member
		}
	}
	if memberExpr == nil {
		return fmt.Errorf("bytecode member access requires member expression")
	}
	val, err := vm.interp.memberAccessOnValueWithOptions(obj, memberExpr, vm.env, instr.preferMethods)
	if err != nil {
		err = vm.interp.wrapStandardRuntimeError(err)
		if instr.node != nil {
			err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return err
	}
	if val == nil {
		val = runtime.NilValue{}
	}
	vm.stack = append(vm.stack, val)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execMemberSet(instr bytecodeInstruction) error {
	memberExpr, ok := instr.node.(*ast.MemberAccessExpression)
	if !ok || memberExpr == nil {
		return fmt.Errorf("bytecode member set expects member access node")
	}
	if memberExpr.Safe {
		return fmt.Errorf("Cannot assign through safe navigation")
	}
	obj, err := vm.pop()
	if err != nil {
		return err
	}
	val, err := vm.pop()
	if err != nil {
		return err
	}
	if instr.operator == "" {
		return fmt.Errorf("bytecode member set missing operator")
	}
	op := ast.AssignmentOperator(instr.operator)
	if op == ast.AssignmentDeclare {
		return fmt.Errorf("Cannot use := on member access")
	}
	binaryOp, isCompound := binaryOpForAssignment(op)
	result, err := vm.assignMemberValue(obj, memberExpr.Member, val, op, binaryOp, isCompound)
	if err != nil {
		err = vm.interp.wrapStandardRuntimeError(err)
		if instr.node != nil {
			err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
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

func (vm *bytecodeVM) execImplicitMemberSet(instr bytecodeInstruction) error {
	implicitExpr, ok := instr.node.(*ast.ImplicitMemberExpression)
	if !ok || implicitExpr == nil {
		return fmt.Errorf("bytecode implicit member set expects node")
	}
	val, err := vm.pop()
	if err != nil {
		return err
	}
	if instr.operator == "" {
		return fmt.Errorf("bytecode implicit member set missing operator")
	}
	op := ast.AssignmentOperator(instr.operator)
	if op == ast.AssignmentDeclare {
		if implicitExpr.Member != nil {
			return fmt.Errorf("Cannot use := on implicit member '#%s'", implicitExpr.Member.Name)
		}
		return fmt.Errorf("Cannot use := on implicit member")
	}
	state := vm.interp.stateFromEnv(vm.env)
	receiver, ok := state.currentImplicitReceiver()
	if !ok || receiver == nil {
		if implicitExpr.Member != nil {
			return fmt.Errorf("Implicit member '#%s' used outside of function with implicit receiver", implicitExpr.Member.Name)
		}
		return fmt.Errorf("Implicit member used outside of function with implicit receiver")
	}
	binaryOp, isCompound := binaryOpForAssignment(op)
	switch inst := receiver.(type) {
	case *runtime.StructInstanceValue:
		result, err := assignStructMember(vm.interp, inst, implicitExpr.Member, val, op, binaryOp, isCompound)
		if err != nil {
			err = vm.interp.wrapStandardRuntimeError(err)
			if instr.node != nil {
				err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
			}
			return err
		}
		if result == nil {
			result = runtime.NilValue{}
		}
		vm.stack = append(vm.stack, result)
		vm.ip++
		return nil
	default:
		return fmt.Errorf("Implicit member assignments supported only on struct instances")
	}
}

func (vm *bytecodeVM) assignMemberValue(target runtime.Value, member ast.Expression, value runtime.Value, op ast.AssignmentOperator, binaryOp string, isCompound bool) (runtime.Value, error) {
	switch inst := target.(type) {
	case *runtime.StructInstanceValue:
		return assignStructMember(vm.interp, inst, member, value, op, binaryOp, isCompound)
	case *runtime.ArrayValue:
		arrayVal := inst
		switch mem := member.(type) {
		case *ast.IntegerLiteral:
			if mem.Value == nil {
				return nil, fmt.Errorf("Array index out of bounds")
			}
			idx := int(mem.Value.Int64())
			state, err := vm.interp.ensureArrayState(arrayVal, 0)
			if err != nil {
				return nil, err
			}
			if idx < 0 || idx >= len(state.values) {
				return nil, fmt.Errorf("Array index out of bounds")
			}
			if op == ast.AssignmentAssign {
				state.values[idx] = value
				vm.interp.syncArrayValues(arrayVal.Handle, state)
				return value, nil
			}
			if !isCompound {
				return nil, fmt.Errorf("unsupported assignment operator %s", op)
			}
			current := state.values[idx]
			computed, err := applyBinaryOperator(vm.interp, binaryOp, current, value)
			if err != nil {
				return nil, err
			}
			state.values[idx] = computed
			vm.interp.syncArrayValues(arrayVal.Handle, state)
			return computed, nil
		case *ast.Identifier:
			if op != ast.AssignmentAssign {
				return nil, fmt.Errorf("unsupported assignment operator %s", op)
			}
			state, err := vm.interp.ensureArrayState(arrayVal, 0)
			if err != nil {
				return nil, err
			}
			switch mem.Name {
			case "storage_handle":
				intVal, ok := value.(runtime.IntegerValue)
				if !ok || intVal.Val == nil || !intVal.Val.IsInt64() {
					return nil, fmt.Errorf("array storage_handle must be an integer")
				}
				handle := intVal.Val.Int64()
				if handle <= 0 {
					return nil, fmt.Errorf("array storage_handle must be positive")
				}
				newState, ok := vm.interp.arrayStates[handle]
				if !ok {
					newState = &arrayState{values: make([]runtime.Value, 0), capacity: 0}
					vm.interp.arrayStates[handle] = newState
				}
				vm.interp.trackArrayValue(handle, arrayVal)
				arrayVal.Elements = newState.values
				vm.interp.syncArrayValues(handle, newState)
				return value, nil
			case "length":
				newLen, err := arrayIndexFromValue(value)
				if err != nil {
					return nil, fmt.Errorf("array length must be a non-negative integer")
				}
				setArrayLength(state, newLen)
				vm.interp.syncArrayValues(arrayVal.Handle, state)
				return value, nil
			case "capacity":
				newCap, err := arrayIndexFromValue(value)
				if err != nil {
					return nil, fmt.Errorf("array capacity must be a non-negative integer")
				}
				if newCap < len(state.values) {
					newCap = len(state.values)
				}
				if ensureArrayCapacity(state, newCap) {
					// ensureArrayCapacity already syncs handle reallocations
				} else if newCap > state.capacity {
					state.capacity = newCap
				}
				vm.interp.syncArrayValues(arrayVal.Handle, state)
				return value, nil
			default:
				return nil, fmt.Errorf("Array has no member '%s'", mem.Name)
			}
		default:
			return nil, fmt.Errorf("Array member assignment requires integer member")
		}
	default:
		return nil, fmt.Errorf("Member assignment requires struct or array")
	}
}
