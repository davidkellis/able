package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) evaluateAssignment(assign *ast.AssignmentExpression, env *runtime.Environment) (runtime.Value, error) {
	value, err := i.evaluateExpression(assign.Right, env)
	if err != nil {
		return nil, err
	}
	binaryOp, isCompound := binaryOpForAssignment(assign.Operator)

	switch lhs := assign.Left.(type) {
	case *ast.Identifier:
		switch assign.Operator {
		case ast.AssignmentDeclare:
			if env.HasInCurrentScope(lhs.Name) {
				return nil, fmt.Errorf(":= requires at least one new binding")
			}
			env.Define(lhs.Name, value)
			if i.currentPackage != "" && env.Parent() == i.global {
				i.registerSymbol(lhs.Name, value)
			}
		case ast.AssignmentAssign:
			if !env.AssignExisting(lhs.Name, value) {
				env.Define(lhs.Name, value)
				if i.currentPackage != "" && env.Parent() == i.global {
					i.registerSymbol(lhs.Name, value)
				}
			}
		default:
			if !isCompound {
				return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
			}
			current, err := env.Get(lhs.Name)
			if err != nil {
				return nil, err
			}
			computed, err := applyBinaryOperator(i, binaryOp, current, value)
			if err != nil {
				return nil, err
			}
			if err := env.Assign(lhs.Name, computed); err != nil {
				return nil, err
			}
			return computed, nil
		}
	case *ast.MemberAccessExpression:
		if lhs.Safe {
			return nil, fmt.Errorf("Cannot assign through safe navigation")
		}
		if assign.Operator == ast.AssignmentDeclare {
			return nil, fmt.Errorf("Cannot use := on member access")
		}
		target, err := i.evaluateExpression(lhs.Object, env)
		if err != nil {
			return nil, err
		}
		switch inst := target.(type) {
		case *runtime.StructInstanceValue:
			return assignStructMember(i, inst, lhs.Member, value, assign.Operator, binaryOp, isCompound)
		case *runtime.ArrayValue:
			arrayVal := inst
			switch member := lhs.Member.(type) {
			case *ast.IntegerLiteral:
				if member.Value == nil {
					return nil, fmt.Errorf("Array index out of bounds")
				}
				idx := int(member.Value.Int64())
				state, err := i.ensureArrayState(arrayVal, 0)
				if err != nil {
					return nil, err
				}
				if idx < 0 || idx >= len(state.values) {
					return nil, fmt.Errorf("Array index out of bounds")
				}
				if assign.Operator == ast.AssignmentAssign {
					state.values[idx] = value
					i.syncArrayValues(arrayVal.Handle, state)
					return value, nil
				}
				if !isCompound {
					return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
				}
				current := state.values[idx]
				computed, err := applyBinaryOperator(i, binaryOp, current, value)
				if err != nil {
					return nil, err
				}
				state.values[idx] = computed
				i.syncArrayValues(arrayVal.Handle, state)
				return computed, nil
			case *ast.Identifier:
				if assign.Operator != ast.AssignmentAssign {
					return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
				}
				state, err := i.ensureArrayState(arrayVal, 0)
				if err != nil {
					return nil, err
				}
				switch member.Name {
				case "storage_handle":
					intVal, ok := value.(runtime.IntegerValue)
					if !ok || intVal.Val == nil || !intVal.Val.IsInt64() {
						return nil, fmt.Errorf("array storage_handle must be an integer")
					}
					handle := intVal.Val.Int64()
					if handle <= 0 {
						return nil, fmt.Errorf("array storage_handle must be positive")
					}
					newState, ok := i.arrayStates[handle]
					if !ok {
						newState = &arrayState{values: make([]runtime.Value, 0), capacity: 0}
						i.arrayStates[handle] = newState
					}
					i.trackArrayValue(handle, arrayVal)
					arrayVal.Elements = newState.values
					i.syncArrayValues(handle, newState)
					return value, nil
				case "length":
					newLen, err := arrayIndexFromValue(value)
					if err != nil {
						return nil, fmt.Errorf("array length must be a non-negative integer")
					}
					setArrayLength(state, newLen)
					i.syncArrayValues(arrayVal.Handle, state)
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
						// ensureArrayCapacity already sets capacity and sync handles reallocations
					} else if newCap > state.capacity {
						state.capacity = newCap
					}
					i.syncArrayValues(arrayVal.Handle, state)
					return value, nil
				default:
					return nil, fmt.Errorf("Array has no member '%s'", member.Name)
				}
			default:
				return nil, fmt.Errorf("Array member assignment requires integer member")
			}
		default:
			return nil, fmt.Errorf("Member assignment requires struct or array")
		}
	case *ast.ImplicitMemberExpression:
		if assign.Operator == ast.AssignmentDeclare {
			if lhs.Member != nil {
				return nil, fmt.Errorf("Cannot use := on implicit member '#%s'", lhs.Member.Name)
			}
			return nil, fmt.Errorf("Cannot use := on implicit member")
		}
		state := i.stateFromEnv(env)
		receiver, ok := state.currentImplicitReceiver()
		if !ok || receiver == nil {
			if lhs.Member != nil {
				return nil, fmt.Errorf("Implicit member '#%s' used outside of function with implicit receiver", lhs.Member.Name)
			}
			return nil, fmt.Errorf("Implicit member used outside of function with implicit receiver")
		}
		switch inst := receiver.(type) {
		case *runtime.StructInstanceValue:
			return assignStructMember(i, inst, lhs.Member, value, assign.Operator, binaryOp, isCompound)
		default:
			return nil, fmt.Errorf("Implicit member assignments supported only on struct instances")
		}
	case *ast.IndexExpression:
		if assign.Operator == ast.AssignmentDeclare {
			return nil, fmt.Errorf("Cannot use := on index assignment")
		}
		arrObj, err := i.evaluateExpression(lhs.Object, env)
		if err != nil {
			return nil, err
		}
		idxVal, err := i.evaluateExpression(lhs.Index, env)
		if err != nil {
			return nil, err
		}
		if setMethod, err := i.findIndexMethod(arrObj, "set", "IndexMut"); err != nil {
			return nil, err
		} else if setMethod != nil {
			if assign.Operator == ast.AssignmentAssign {
				setResult, err := i.CallFunction(setMethod, []runtime.Value{arrObj, idxVal, value})
				if err != nil {
					return nil, err
				}
				if isErrorResult(i, setResult) {
					return setResult, nil
				}
				return value, nil
			}
			if !isCompound {
				return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
			}
			getMethod, err := i.findIndexMethod(arrObj, "get", "Index")
			if err != nil {
				return nil, err
			}
			if getMethod == nil {
				return nil, fmt.Errorf("Compound index assignment requires readable Index implementation")
			}
			current, err := i.CallFunction(getMethod, []runtime.Value{arrObj, idxVal})
			if err != nil {
				return nil, err
			}
			computed, err := applyBinaryOperator(i, binaryOp, current, value)
			if err != nil {
				return nil, err
			}
			setResult, err := i.CallFunction(setMethod, []runtime.Value{arrObj, idxVal, computed})
			if err != nil {
				return nil, err
			}
			if isErrorResult(i, setResult) {
				return setResult, nil
			}
			return computed, nil
		}
		arr, err := i.toArrayValue(arrObj)
		if err != nil {
			return nil, err
		}
		idx, err := indexFromValue(idxVal)
		if err != nil {
			return nil, err
		}
		state, err := i.ensureArrayState(arr, 0)
		if err != nil {
			return nil, err
		}
		if idx < 0 || idx >= len(state.values) {
			return nil, fmt.Errorf("Array index out of bounds")
		}
		if assign.Operator == ast.AssignmentAssign {
			state.values[idx] = value
			i.syncArrayValues(arr.Handle, state)
			return value, nil
		}
		if !isCompound {
			return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
		}
		current := state.values[idx]
		computed, err := applyBinaryOperator(i, binaryOp, current, value)
		if err != nil {
			return nil, err
		}
		state.values[idx] = computed
		i.syncArrayValues(arr.Handle, state)
		return computed, nil
	case ast.Pattern:
		if isCompound {
			return nil, fmt.Errorf("compound assignment not supported with patterns")
		}
		switch assign.Operator {
		case ast.AssignmentDeclare:
			newNames, hasAny := analyzePatternDeclarationNames(env, lhs)
			if !hasAny || len(newNames) == 0 {
				return nil, fmt.Errorf(":= requires at least one new binding")
			}
			intent := &bindingIntent{declarationNames: newNames}
			if err := i.assignPattern(lhs, value, env, true, intent); err != nil {
				return nil, err
			}
		case ast.AssignmentAssign:
			intent := &bindingIntent{allowFallback: true}
			if err := i.assignPattern(lhs, value, env, false, intent); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
		}
	default:
		return nil, fmt.Errorf("unsupported assignment target %s", lhs.NodeType())
	}

	return value, nil
}

func isErrorResult(i *Interpreter, value runtime.Value) bool {
	if value == nil {
		return false
	}
	if _, ok := asErrorValue(value); ok {
		return true
	}
	return i.matchesType(ast.Ty("Error"), value)
}
