package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execIndexGet(instr bytecodeInstruction) error {
	if vm.stackDepth() < 2 {
		return fmt.Errorf("bytecode stack underflow")
	}
	idxVal := vm.stackValue(vm.stackDepth() - 1)
	obj := vm.stackValue(vm.stackDepth() - 2)
	var err error
	var (
		result            runtime.Value
		elementToken      uint16
		elementTokenKnown bool
	)
	result, elementToken, elementTokenKnown, err = vm.resolveIndexGetWithToken(obj, idxVal)
	if err != nil {
		err = vm.interp.wrapStandardRuntimeError(err)
		if instr.node != nil {
			err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return err
	}
	vm.replaceTop2Unchecked(result)
	if elementTokenKnown && vm.canSkipArrayGetSuccessPropagation(result, elementToken, true) {
		vm.ip += 2
		return nil
	}
	if !elementTokenKnown && vm.canSkipSuccessPropagation(result) {
		vm.ip += 2
		return nil
	}
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execIndexSet(instr bytecodeInstruction) error {
	if vm.stackDepth() < 3 {
		return fmt.Errorf("bytecode stack underflow")
	}
	idxVal := vm.stackValue(vm.stackDepth() - 1)
	obj := vm.stackValue(vm.stackDepth() - 2)
	val := vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion, vm.stackValue(vm.stackDepth()-3))
	if arr, ok := obj.(*runtime.ArrayValue); ok {
		vm.observeBytecodeArrayOwnershipArrayWrite(arr, val)
	} else {
		vm.markBytecodeArrayOwnershipValueEscaped(val, bytecodeArrayOwnershipEscapeAggregate)
	}
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

func bytecodeDirectStructMemberValue(receiver runtime.Value, memberName string, preferMethods bool) (runtime.Value, bool) {
	if memberName == "" {
		return nil, false
	}
	inst, ok := receiver.(*runtime.StructInstanceValue)
	if !ok || inst == nil {
		return nil, false
	}
	if !preferMethods {
		return structNamedFieldValue(inst, memberName)
	}
	val, ok := structNamedFieldValue(inst, memberName)
	if !ok || !isCallableRuntimeValue(val) {
		return nil, false
	}
	return val, true
}

func bytecodeNamedStructMemberPlanAt(program *bytecodeProgram, ip int) (bytecodeNamedStructMemberPlan, bool) {
	if program == nil || ip < 0 || program.namedStructMembers == nil {
		return bytecodeNamedStructMemberPlan{}, false
	}
	plan, ok := program.namedStructMembers[ip]
	return plan, ok
}

func bytecodeNamedStructMemberPlanForInstruction(program *bytecodeProgram, ip int, instr *bytecodeInstruction) (bytecodeNamedStructMemberPlan, bool) {
	plan, ok := bytecodeNamedStructMemberPlanAt(program, ip)
	if !ok || instr == nil || instr.name == "" {
		return bytecodeNamedStructMemberPlan{}, false
	}
	if plan.definition == nil || plan.definition.Node == nil || plan.fieldIndex < 0 || plan.fieldIndex >= len(plan.definition.Node.Fields) {
		return bytecodeNamedStructMemberPlan{}, false
	}
	field := plan.definition.Node.Fields[plan.fieldIndex]
	if field == nil || field.Name == nil || field.Name.Name != instr.name {
		return bytecodeNamedStructMemberPlan{}, false
	}
	return plan, true
}

func bytecodeStructInstanceMatchesMemberPlan(inst *runtime.StructInstanceValue, def *runtime.StructDefinitionValue) bool {
	if inst == nil || inst.Definition == nil || def == nil {
		return false
	}
	return inst.Definition == def || (inst.Definition.Node != nil && def.Node != nil && inst.Definition.Node == def.Node)
}

func bytecodeDirectPlannedStructMemberValue(receiver runtime.Value, plan bytecodeNamedStructMemberPlan, preferMethods bool) (runtime.Value, bool) {
	if plan.definition == nil {
		return nil, false
	}
	inst, ok := receiver.(*runtime.StructInstanceValue)
	if !ok || inst == nil || inst.Definition == nil || inst.Positional == nil {
		return nil, false
	}
	if !bytecodeStructInstanceMatchesMemberPlan(inst, plan.definition) {
		return nil, false
	}
	if plan.fieldIndex < 0 || plan.fieldIndex >= len(inst.Positional) {
		return nil, false
	}
	val := inst.Positional[plan.fieldIndex]
	if preferMethods && !isCallableRuntimeValue(val) {
		return nil, false
	}
	return val, true
}

func bytecodeDirectPlannedStructMemberSet(interp *Interpreter, receiver runtime.Value, plan bytecodeNamedStructMemberPlan, value runtime.Value, op ast.AssignmentOperator, binaryOp string, isCompound bool) (runtime.Value, bool, error) {
	if plan.definition == nil {
		return nil, false, nil
	}
	inst, ok := receiver.(*runtime.StructInstanceValue)
	if !ok || inst == nil || inst.Definition == nil || inst.Positional == nil {
		return nil, false, nil
	}
	if !bytecodeStructInstanceMatchesMemberPlan(inst, plan.definition) {
		return nil, false, nil
	}
	if plan.fieldIndex < 0 || plan.fieldIndex >= len(inst.Positional) {
		return nil, false, nil
	}
	if op == ast.AssignmentAssign {
		inst.Positional[plan.fieldIndex] = value
		return value, true, nil
	}
	if !isCompound {
		return nil, true, fmt.Errorf("unsupported assignment operator %s", op)
	}
	current := inst.Positional[plan.fieldIndex]
	computed, err := applyBinaryOperator(interp, binaryOp, current, value)
	if err != nil {
		return nil, true, err
	}
	inst.Positional[plan.fieldIndex] = computed
	return computed, true, nil
}

func (vm *bytecodeVM) execMemberAccess(instr bytecodeInstruction) error {
	if vm.stackDepth() < 1 {
		return fmt.Errorf("bytecode stack underflow")
	}
	obj := vm.stackValue(vm.stackDepth() - 1)
	if instr.safe && isNilRuntimeValue(obj) {
		vm.replaceTop1Unchecked(runtime.NilValue{})
		vm.ip++
		return nil
	}
	obj = vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion, obj)
	memberName := instr.name
	if plan, ok := bytecodeNamedStructMemberPlanForInstruction(vm.currentProgram, vm.ip, &instr); ok {
		if val, ok := bytecodeDirectPlannedStructMemberValue(obj, plan, instr.preferMethods); ok {
			vm.replaceTop1Unchecked(val)
			vm.ip++
			return nil
		}
	}
	if val, ok := bytecodeDirectStructMemberValue(obj, memberName, instr.preferMethods); ok {
		vm.replaceTop1Unchecked(val)
		vm.ip++
		return nil
	}
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
	if val, ok := bytecodeDirectStructMemberValue(obj, memberName, instr.preferMethods); ok {
		vm.replaceTop1Unchecked(val)
		vm.ip++
		return nil
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
	if instr.safe {
		return fmt.Errorf("Cannot assign through safe navigation")
	}
	if vm.stackDepth() < 2 {
		return fmt.Errorf("bytecode stack underflow")
	}
	obj := vm.stackValue(vm.stackDepth() - 1)
	// Raw scalar stack cells are reused by later expressions. A member write is
	// an aggregate escape, so it must retain an ordinary independent value.
	val := vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion, vm.stackValue(vm.stackDepth()-2))
	if instr.operator == "" {
		return fmt.Errorf("bytecode member set missing operator")
	}
	op := ast.AssignmentOperator(instr.operator)
	if op == ast.AssignmentDeclare {
		return fmt.Errorf("Cannot use := on member access")
	}
	binaryOp, isCompound := binaryOpForAssignment(op)
	if plan, ok := bytecodeNamedStructMemberPlanForInstruction(vm.currentProgram, vm.ip, &instr); ok {
		if result, handled, err := bytecodeDirectPlannedStructMemberSet(vm.interp, obj, plan, val, op, binaryOp, isCompound); handled {
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
	}
	memberExpr, ok := instr.node.(*ast.MemberAccessExpression)
	if !ok || memberExpr == nil {
		return fmt.Errorf("bytecode member set expects member access node")
	}
	if memberExpr.Safe {
		return fmt.Errorf("Cannot assign through safe navigation")
	}
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
	if vm.stackDepth() < 1 {
		return fmt.Errorf("bytecode stack underflow")
	}
	// Like an explicit member write, an implicit-receiver field write retains
	// the value as aggregate state beyond the reusable raw stack position.
	val := vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion, vm.stackValue(vm.stackDepth()-1))
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
				vm.interp.syncTrackedArrayWrite(arrayVal, state, idx, value)
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
			vm.interp.syncTrackedArrayWrite(arrayVal, state, idx, computed)
			return computed, nil
		case *ast.Identifier:
			if op != ast.AssignmentAssign {
				return nil, fmt.Errorf("unsupported assignment operator %s", op)
			}
			switch mem.Name {
			case "storage_handle":
				if _, err := vm.interp.ensureArrayStateForMetadata(arrayVal, 0); err != nil {
					return nil, err
				}
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
				prevHandle := arrayVal.Handle
				if prevHandle == 0 {
					prevHandle = arrayVal.TrackedHandle
				}
				newState, err := runtime.ArrayStoreEnsureHandle(handle, 0, 0)
				if err != nil {
					return nil, err
				}
				vm.interp.trackArrayValue(handle, arrayVal)
				arrayVal.Elements = newState.Values
				if handle == prevHandle {
					vm.interp.syncTrackedArrayState(arrayVal, newState)
				} else {
					vm.interp.syncArrayValues(handle, newState)
				}
				return value, nil
			case "length":
				state, err := vm.interp.ensureArrayStateForMetadata(arrayVal, 0)
				if err != nil {
					return nil, err
				}
				newLen, err := arrayIndexFromValue(value)
				if err != nil {
					return nil, fmt.Errorf("array length must be a non-negative integer")
				}
				setArrayLength(state, newLen)
				vm.interp.syncArrayHandleLength(arrayVal.Handle, state)
				return value, nil
			case "capacity":
				state, err := vm.interp.ensureArrayStateForMetadata(arrayVal, 0)
				if err != nil {
					return nil, err
				}
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
				vm.interp.syncArrayHandleMetadata(arrayVal.Handle, state)
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
