package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execIndexGet(instr bytecodeInstruction) error {
	if len(vm.stack) < 2 {
		return fmt.Errorf("bytecode stack underflow")
	}
	idxVal := vm.stack[len(vm.stack)-1]
	obj := vm.stack[len(vm.stack)-2]
	var err error
	result, err := vm.resolveIndexGet(obj, idxVal)
	if err != nil {
		err = vm.interp.wrapStandardRuntimeError(err)
		if instr.node != nil {
			err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return err
	}
	vm.replaceTop2Unchecked(result)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execIndexSet(instr bytecodeInstruction) error {
	if len(vm.stack) < 3 {
		return fmt.Errorf("bytecode stack underflow")
	}
	idxVal := vm.stack[len(vm.stack)-1]
	obj := vm.stack[len(vm.stack)-2]
	val := vm.stack[len(vm.stack)-3]
	if instr.operator == "" {
		return fmt.Errorf("bytecode index set missing operator")
	}
	op := ast.AssignmentOperator(instr.operator)
	binaryOp, isCompound := binaryOpForAssignment(op)
	var err error
	result, err := vm.resolveIndexSet(obj, idxVal, val, op, binaryOp, isCompound)
	if err != nil {
		err = vm.interp.wrapStandardRuntimeError(err)
		if instr.node != nil {
			err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return err
	}
	vm.replaceTop3Unchecked(result)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execMemberAccess(instr bytecodeInstruction) error {
	if len(vm.stack) < 1 {
		return fmt.Errorf("bytecode stack underflow")
	}
	obj := vm.stack[len(vm.stack)-1]
	if instr.safe && isNilRuntimeValue(obj) {
		vm.replaceTop1Unchecked(runtime.NilValue{})
		vm.ip++
		return nil
	}
	memberName := instr.name
	memberExpr := ast.Expression(nil)
	if instr.node != nil {
		if member, ok := instr.node.(*ast.MemberAccessExpression); ok && member != nil {
			memberExpr = member.Member
			if memberName == "" {
				if ident, ok := memberExpr.(*ast.Identifier); ok && ident != nil {
					memberName = ident.Name
				}
			}
		}
	}
	if memberExpr == nil && memberName != "" {
		ident := ast.Identifier{Name: memberName}
		memberExpr = &ident
	}
	if memberExpr == nil {
		return fmt.Errorf("bytecode member access requires member expression")
	}
	useMethodCache := vm.canUseMemberMethodCache(memberName, instr.preferMethods)
	if useMethodCache {
		if cached, ok := vm.lookupCachedMemberMethod(vm.currentProgram, vm.ip, memberName, instr.preferMethods, obj); ok {
			vm.replaceTop1Unchecked(cached)
			vm.ip++
			return nil
		}
	}
	val, err := vm.interp.memberAccessOnValueWithOptions(obj, memberExpr, vm.env, instr.preferMethods)
	if err != nil {
		err = vm.interp.wrapStandardRuntimeError(err)
		if instr.node != nil {
			err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return err
	}
	if useMethodCache {
		vm.storeCachedMemberMethod(vm.currentProgram, vm.ip, memberName, instr.preferMethods, obj, val)
	}
	vm.replaceTop1Unchecked(val)
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
	if len(vm.stack) < 2 {
		return fmt.Errorf("bytecode stack underflow")
	}
	obj := vm.stack[len(vm.stack)-1]
	val := vm.stack[len(vm.stack)-2]
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
	vm.replaceTop2Unchecked(result)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execImplicitMemberSet(instr bytecodeInstruction) error {
	implicitExpr, ok := instr.node.(*ast.ImplicitMemberExpression)
	if !ok || implicitExpr == nil {
		return fmt.Errorf("bytecode implicit member set expects node")
	}
	if len(vm.stack) < 1 {
		return fmt.Errorf("bytecode stack underflow")
	}
	val := vm.stack[len(vm.stack)-1]
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
		vm.replaceTop1Unchecked(result)
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
			if idx < 0 || idx >= len(state.Values) {
				return nil, fmt.Errorf("Array index out of bounds")
			}
			if op == ast.AssignmentAssign {
				state.Values[idx] = value
				vm.interp.syncArrayValues(arrayVal.Handle, state)
				return value, nil
			}
			if !isCompound {
				return nil, fmt.Errorf("unsupported assignment operator %s", op)
			}
			current := state.Values[idx]
			computed, err := applyBinaryOperator(vm.interp, binaryOp, current, value)
			if err != nil {
				return nil, err
			}
			state.Values[idx] = computed
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
				if !ok {
					return nil, fmt.Errorf("array storage_handle must be an integer")
				}
				handle, fits := intVal.ToInt64()
				if !fits {
					return nil, fmt.Errorf("array storage_handle must be an integer")
				}
				if handle <= 0 {
					return nil, fmt.Errorf("array storage_handle must be positive")
				}
				newState, err := runtime.ArrayStoreEnsureHandle(handle, 0, 0)
				if err != nil {
					return nil, err
				}
				vm.interp.trackArrayValue(handle, arrayVal)
				arrayVal.Elements = newState.Values
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
				if newCap < len(state.Values) {
					newCap = len(state.Values)
				}
				if ensureArrayCapacity(state, newCap) {
					// ensureArrayCapacity already syncs handle reallocations
				} else if newCap > state.Capacity {
					state.Capacity = newCap
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
