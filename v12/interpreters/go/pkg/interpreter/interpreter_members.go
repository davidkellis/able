package interpreter

import (
	"fmt"
	"math"
	"math/big"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) evaluateMemberAccess(expr *ast.MemberAccessExpression, env *runtime.Environment) (runtime.Value, error) {
	obj, err := i.evaluateExpression(expr.Object, env)
	if err != nil {
		return nil, err
	}
	if expr.Safe && isNilRuntimeValue(obj) {
		return runtime.NilValue{}, nil
	}
	return i.memberAccessOnValue(obj, expr.Member, env)
}

func (i *Interpreter) memberAccessOnValue(obj runtime.Value, member ast.Expression, env *runtime.Environment) (runtime.Value, error) {
	return i.memberAccessOnValueWithOptions(obj, member, env, false)
}

func (i *Interpreter) memberAccessOnValueWithOptions(obj runtime.Value, member ast.Expression, env *runtime.Environment, preferMethods bool) (runtime.Value, error) {
	switch v := obj.(type) {
	case *runtime.StructDefinitionValue:
		if v != nil && isSingletonStructDef(v.Node) {
			inst := &runtime.StructInstanceValue{Definition: v, Fields: map[string]runtime.Value{}}
			if val, err := i.structInstanceMember(inst, member, env, preferMethods); err == nil {
				return val, nil
			}
		}
		return i.structDefinitionMember(v, member)
	case runtime.StructDefinitionValue:
		if isSingletonStructDef(v.Node) {
			inst := &runtime.StructInstanceValue{Definition: &v, Fields: map[string]runtime.Value{}}
			if val, err := i.structInstanceMember(inst, member, env, preferMethods); err == nil {
				return val, nil
			}
		}
		return i.structDefinitionMember(&v, member)
	case runtime.InterfaceDefinitionValue:
		return i.interfaceDefinitionMember(&v, member)
	case *runtime.InterfaceDefinitionValue:
		return i.interfaceDefinitionMember(v, member)
	case runtime.TypeRefValue:
		return i.typeRefMember(v, member)
	case *runtime.TypeRefValue:
		if v == nil {
			return nil, fmt.Errorf("Type reference member access on nil value")
		}
		return i.typeRefMember(*v, member)
	case runtime.PackageValue:
		return i.packageMemberAccess(v, member)
	case *runtime.PackageValue:
		return i.packageMemberAccess(*v, member)
	case runtime.ImplementationNamespaceValue:
		return i.implNamespaceMember(v, member)
	case *runtime.ImplementationNamespaceValue:
		return i.implNamespaceMember(*v, member)
	case runtime.DynPackageValue:
		return i.dynPackageMemberAccess(v, member)
	case *runtime.DynPackageValue:
		return i.dynPackageMemberAccess(*v, member)
	case *runtime.StructInstanceValue:
		return i.structInstanceMember(v, member, env, preferMethods)
	case *runtime.InterfaceValue:
		return i.interfaceMember(v, member)
	case *runtime.ArrayValue:
		i.ensureArrayBuiltins()
		return i.arrayMemberWithOverrides(v, member, env, preferMethods)
	case *runtime.HasherValue:
		return i.hasherMember(v, member)
	case *runtime.FutureValue:
		return i.futureMember(v, member)
	case *runtime.IteratorValue:
		if val, err := i.iteratorMember(v, member); err == nil {
			return val, nil
		} else if ident, ok := member.(*ast.Identifier); ok {
			if bound, err := i.resolveMethodFromPool(env, ident.Name, v, ""); err != nil {
				return nil, err
			} else if bound != nil {
				return bound, nil
			}
			return nil, err
		} else {
			return nil, err
		}
	case runtime.ErrorValue:
		return i.errorMember(v, member, env)
	case *runtime.ErrorValue:
		if v == nil {
			return nil, fmt.Errorf("Error member access on nil value")
		}
		return i.errorMember(*v, member, env)
	case runtime.StringValue:
		return i.stringMemberWithOverrides(v, member, env)
	case *runtime.StringValue:
		if v == nil {
			return nil, fmt.Errorf("String member access on nil value")
		}
		return i.stringMemberWithOverrides(*v, member, env)
	default:
		if resolved, ok, err := i.resolveZeroArgBoundMethodForMemberAccess(obj, env); err != nil {
			return nil, err
		} else if ok {
			return i.memberAccessOnValueWithOptions(resolved, member, env, preferMethods)
		}
		if ident, ok := member.(*ast.Identifier); ok {
			if bound, err := i.resolveMethodFromPool(env, ident.Name, obj, ""); err != nil {
				return nil, err
			} else if bound != nil {
				return bound, nil
			}
		}
		return nil, fmt.Errorf("Member access only supported on structs/arrays in this milestone (got %s)", obj.Kind())
	}
}

func (i *Interpreter) resolveZeroArgBoundMethodForMemberAccess(obj runtime.Value, env *runtime.Environment) (runtime.Value, bool, error) {
	call := func(target runtime.Value) (runtime.Value, bool, error) {
		result, err := i.CallFunctionIn(target, nil, env)
		if err != nil {
			return nil, false, err
		}
		switch result.(type) {
		case runtime.PartialFunctionValue, *runtime.PartialFunctionValue:
			return nil, false, nil
		}
		return result, true, nil
	}
	switch method := obj.(type) {
	case runtime.NativeBoundMethodValue:
		if method.Method.Arity != 0 {
			return nil, false, nil
		}
		return call(method)
	case *runtime.NativeBoundMethodValue:
		if method == nil || method.Method.Arity != 0 {
			return nil, false, nil
		}
		return call(method)
	case runtime.BoundMethodValue:
		overloads := functionOverloads(method.Method)
		if len(overloads) == 0 || minArgsForOverloads(overloads) > 1 {
			return nil, false, nil
		}
		return call(method)
	case *runtime.BoundMethodValue:
		if method == nil {
			return nil, false, nil
		}
		overloads := functionOverloads(method.Method)
		if len(overloads) == 0 || minArgsForOverloads(overloads) > 1 {
			return nil, false, nil
		}
		return call(method)
	default:
		return nil, false, nil
	}
}

func (i *Interpreter) evaluateImplicitMemberExpression(expr *ast.ImplicitMemberExpression, env *runtime.Environment) (runtime.Value, error) {
	if expr == nil || expr.Member == nil {
		return nil, fmt.Errorf("Implicit member requires identifier")
	}
	state := i.stateFromEnv(env)
	receiver, ok := state.currentImplicitReceiver()
	if !ok {
		return nil, fmt.Errorf("Implicit member '#%s' requires enclosing function with a first parameter", expr.Member.Name)
	}
	return i.memberAccessOnValue(receiver, expr.Member, env)
}

func (i *Interpreter) stringMemberWithOverrides(str runtime.StringValue, member ast.Expression, env *runtime.Environment) (runtime.Value, error) {
	if ident, ok := member.(*ast.Identifier); ok {
		if bound, err := i.resolveMethodFromPool(env, ident.Name, str, ""); err != nil {
			return nil, err
		} else if bound != nil {
			return bound, nil
		}
	}
	return i.stringMember(str, member)
}

func (i *Interpreter) arrayMemberWithOverrides(arr *runtime.ArrayValue, member ast.Expression, env *runtime.Environment, preferMethods bool) (runtime.Value, error) {
	if arr == nil {
		return nil, fmt.Errorf("array receiver is nil")
	}
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("array member access expects identifier")
	}
	if preferMethods {
		if bound, err := i.resolveMethodFromPool(env, ident.Name, arr, ""); err != nil {
			return nil, err
		} else if bound != nil {
			return bound, nil
		}
		return i.arrayMember(arr, member)
	}
	if isDirectArrayMemberName(ident.Name) {
		return i.arrayMember(arr, member)
	}
	if bound, err := i.resolveMethodFromPool(env, ident.Name, arr, ""); err != nil {
		return nil, err
	} else if bound != nil {
		return bound, nil
	}
	return i.arrayMember(arr, member)
}

func (i *Interpreter) evaluateIndexExpression(expr *ast.IndexExpression, env *runtime.Environment) (runtime.Value, error) {
	obj, err := i.evaluateExpression(expr.Object, env)
	if err != nil {
		return nil, err
	}
	idxVal, err := i.evaluateExpression(expr.Index, env)
	if err != nil {
		return nil, err
	}
	return i.indexGet(obj, idxVal)
}

func (i *Interpreter) toArrayValue(val runtime.Value) (*runtime.ArrayValue, error) {
	switch v := val.(type) {
	case *runtime.ArrayValue:
		if _, err := i.ensureArrayState(v, 0); err != nil {
			return nil, err
		}
		return v, nil
	case *runtime.StructInstanceValue:
		if v == nil || v.Definition == nil || v.Definition.Node == nil || v.Definition.Node.ID == nil {
			return nil, fmt.Errorf("Indexing is only supported on arrays")
		}
		if v.Definition.Node.ID.Name != "Array" {
			return nil, fmt.Errorf("Indexing is only supported on arrays")
		}
		return i.arrayValueFromStructInstance(v)
	default:
		return nil, fmt.Errorf("Indexing is only supported on arrays")
	}
}

func (i *Interpreter) findIndexMethod(val runtime.Value, methodName string, iface string) (runtime.Value, error) {
	if ifaceVal, ok := val.(*runtime.InterfaceValue); ok && ifaceVal != nil {
		if method, err := i.findIndexMethod(ifaceVal.Underlying, methodName, iface); err == nil && method != nil {
			return method, nil
		} else if err != nil {
			return nil, err
		}
	}
	info, ok := i.getTypeInfoForValue(val)
	if !ok {
		return nil, nil
	}
	method, err := i.findMethodCached(info, methodName, iface)
	if method != nil || err != nil {
		return method, err
	}
	// In compiled no-bootstrap mode, fall back to the compiled interface dispatch.
	if i.interfaceMethodResolver != nil && iface != "" {
		if resolved, found := i.interfaceMethodResolver(val, iface, methodName); found && resolved != nil {
			return resolved, nil
		}
	}
	return nil, nil
}

// IndexGet is an exported wrapper for index access to support compiled interop.
func (i *Interpreter) IndexGet(obj runtime.Value, idx runtime.Value, _ *runtime.Environment) (runtime.Value, error) {
	return i.indexGet(obj, idx)
}

// IndexAssign is an exported wrapper for index assignment to support compiled interop.
func (i *Interpreter) IndexAssign(obj runtime.Value, idx runtime.Value, value runtime.Value, _ *runtime.Environment) (runtime.Value, error) {
	return i.assignIndex(obj, idx, value, ast.AssignmentAssign, "", false)
}

// MemberAssign is an exported wrapper for member assignment to support compiled interop.
func (i *Interpreter) MemberAssign(obj runtime.Value, member runtime.Value, value runtime.Value, _ *runtime.Environment) (runtime.Value, error) {
	if i == nil {
		return nil, fmt.Errorf("interpreter: nil interpreter")
	}
	memberExpr, err := memberExpressionFromValue(member)
	if err != nil {
		return nil, err
	}
	switch inst := obj.(type) {
	case *runtime.StructInstanceValue:
		if inst == nil {
			return nil, fmt.Errorf("member assignment expects struct instance")
		}
		return assignStructMember(i, inst, memberExpr, value, ast.AssignmentAssign, "", false)
	case *runtime.ArrayValue:
		if inst == nil {
			return nil, fmt.Errorf("array receiver is nil")
		}
		switch member := memberExpr.(type) {
		case *ast.IntegerLiteral:
			if member.Value == nil {
				return nil, fmt.Errorf("Array index out of bounds")
			}
			idx := int(member.Value.Int64())
			state, err := i.ensureArrayState(inst, 0)
			if err != nil {
				return nil, err
			}
			if idx < 0 || idx >= len(state.Values) {
				return nil, fmt.Errorf("Array index out of bounds")
			}
			state.Values[idx] = value
			i.syncTrackedArrayWrite(inst, state, idx, value)
			return value, nil
		case *ast.Identifier:
			switch member.Name {
			case "storage_handle":
				if _, err := i.ensureArrayStateForMetadata(inst, 0); err != nil {
					return nil, err
				}
				intVal, ok := value.(runtime.IntegerValue)
				if !ok {
					return nil, fmt.Errorf("array storage_handle must be an integer")
				}
				handle, ok := intVal.ToInt64()
				if !ok {
					return nil, fmt.Errorf("array storage_handle must be an integer")
				}
				if handle <= 0 {
					return nil, fmt.Errorf("array storage_handle must be positive")
				}
				prevHandle := inst.Handle
				if prevHandle == 0 {
					prevHandle = inst.TrackedHandle
				}
				newState, err := runtime.ArrayStoreEnsureHandle(handle, 0, 0)
				if err != nil {
					return nil, err
				}
				i.trackArrayValue(handle, inst)
				inst.Elements = newState.Values
				if handle == prevHandle {
					i.syncTrackedArrayState(inst, newState)
				} else {
					i.syncArrayValues(handle, newState)
				}
				return value, nil
			case "length":
				state, err := i.ensureArrayStateForMetadata(inst, 0)
				if err != nil {
					return nil, err
				}
				newLen, err := arrayIndexFromValue(value)
				if err != nil {
					return nil, fmt.Errorf("array length must be a non-negative integer")
				}
				setArrayLength(state, newLen)
				i.syncArrayHandleLength(inst.Handle, state)
				return value, nil
			case "capacity":
				state, err := i.ensureArrayStateForMetadata(inst, 0)
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
				} else if newCap > state.Capacity {
					state.Capacity = newCap
				}
				i.syncArrayHandleMetadata(inst.Handle, state)
				return value, nil
			default:
				return nil, fmt.Errorf("Array has no member '%s'", member.Name)
			}
		default:
			return nil, fmt.Errorf("Array member assignment requires integer member")
		}
	default:
		return nil, fmt.Errorf("member assignment expects struct instance")
	}
}

// MemberGet is an exported wrapper for member access to support compiled interop.
func (i *Interpreter) MemberGet(obj runtime.Value, member runtime.Value, env *runtime.Environment) (runtime.Value, error) {
	if i == nil {
		return nil, fmt.Errorf("interpreter: nil interpreter")
	}
	memberExpr, err := memberExpressionFromValue(member)
	if err != nil {
		return nil, err
	}
	return i.memberAccessOnValue(obj, memberExpr, env)
}

// MemberGetPreferMethods is an exported wrapper for member access when methods should take priority.
func (i *Interpreter) MemberGetPreferMethods(obj runtime.Value, member runtime.Value, env *runtime.Environment) (runtime.Value, error) {
	if i == nil {
		return nil, fmt.Errorf("interpreter: nil interpreter")
	}
	memberExpr, err := memberExpressionFromValue(member)
	if err != nil {
		return nil, err
	}
	return i.memberAccessOnValueWithOptions(obj, memberExpr, env, true)
}

func memberExpressionFromValue(member runtime.Value) (ast.Expression, error) {
	switch m := member.(type) {
	case runtime.StringValue:
		return ast.NewIdentifier(m.Val), nil
	case *runtime.StringValue:
		if m == nil {
			return nil, fmt.Errorf("member access expects string member")
		}
		return ast.NewIdentifier(m.Val), nil
	case runtime.IntegerValue:
		idx, ok := m.ToInt64()
		if !ok {
			return nil, fmt.Errorf("member access expects integer index")
		}
		return ast.NewIntegerLiteral(big.NewInt(idx), nil), nil
	case *runtime.IntegerValue:
		if m == nil {
			return nil, fmt.Errorf("member access expects integer index")
		}
		idx, ok := m.ToInt64()
		if !ok {
			return nil, fmt.Errorf("member access expects integer index")
		}
		return ast.NewIntegerLiteral(big.NewInt(idx), nil), nil
	default:
		return nil, fmt.Errorf("member access expects string or integer member")
	}
}

func (i *Interpreter) findApplyMethod(val runtime.Value) (runtime.Value, error) {
	if ifaceVal, ok := val.(*runtime.InterfaceValue); ok && ifaceVal != nil {
		if method, err := i.findApplyMethod(ifaceVal.Underlying); err == nil && method != nil {
			return method, nil
		} else if err != nil {
			return nil, err
		}
	}
	info, ok := i.getTypeInfoForValue(val)
	if !ok {
		return nil, nil
	}
	method, err := i.findMethodCached(info, "apply", "Apply")
	if method != nil || err != nil {
		return method, err
	}
	// In compiled no-bootstrap mode, fall back to the compiled interface dispatch.
	if i.interfaceMethodResolver != nil {
		if resolved, found := i.interfaceMethodResolver(val, "Apply", "apply"); found && resolved != nil {
			// Adjust arity: interfaceMethodResolver returns arity+1 (includes self),
			// but the caller wraps in BoundMethodValue which injects receiver separately.
			if native, ok := resolved.(*runtime.NativeFunctionValue); ok && native.Arity > 0 {
				adjusted := *native
				adjusted.Arity = native.Arity - 1
				return &adjusted, nil
			}
			return resolved, nil
		}
	}
	return nil, nil
}

func indexFromValue(val runtime.Value) (int, error) {
	switch v := val.(type) {
	case runtime.IntegerValue:
		n, ok := v.ToInt64()
		if !ok {
			return 0, fmt.Errorf("Array index must be within int range")
		}
		return int(n), nil
	case runtime.FloatValue:
		if math.IsNaN(v.Val) || math.IsInf(v.Val, 0) {
			return 0, fmt.Errorf("Array index must be a number")
		}
		idx := int(math.Trunc(v.Val))
		return idx, nil
	default:
		return 0, fmt.Errorf("Array index must be a number")
	}
}

func (i *Interpreter) structInstanceMember(inst *runtime.StructInstanceValue, member ast.Expression, env *runtime.Environment, preferMethods bool) (runtime.Value, error) {
	if inst == nil {
		return nil, fmt.Errorf("Member access only supported on structs/arrays in this milestone")
	}
	switch ident := member.(type) {
	case *ast.Identifier:
		if !structUsesNamedFieldStorage(inst) {
			return nil, fmt.Errorf("Expected named struct instance")
		}
		if preferMethods {
			if val, ok := structNamedFieldValue(inst, ident.Name); ok {
				if isCallableRuntimeValue(val) {
					return val, nil
				}
				// Fall back to methods when the field exists but is not callable.
			}
			if bound, err := i.resolveMethodFromPool(env, ident.Name, inst, ""); err != nil {
				return nil, err
			} else if bound != nil {
				return bound, nil
			}
			if val, ok := structNamedFieldValue(inst, ident.Name); ok {
				return val, nil
			}
		} else {
			if val, ok := structNamedFieldValue(inst, ident.Name); ok {
				return val, nil
			}
			if bound, err := i.resolveMethodFromPool(env, ident.Name, inst, ""); err != nil {
				return nil, err
			} else if bound != nil {
				return bound, nil
			}
		}
		return nil, fmt.Errorf("No field or method named '%s'", ident.Name)
	case *ast.IntegerLiteral:
		if inst.Positional == nil {
			return nil, fmt.Errorf("Expected positional struct instance")
		}
		if ident.Value == nil {
			return nil, fmt.Errorf("Struct field index out of bounds")
		}
		idx := int(ident.Value.Int64())
		if idx < 0 || idx >= len(inst.Positional) {
			return nil, fmt.Errorf("Struct field index out of bounds")
		}
		return inst.Positional[idx], nil
	default:
		return nil, fmt.Errorf("Member access only supported on structs/arrays in this milestone")
	}
}

func isNilRuntimeValue(val runtime.Value) bool {
	if val == nil {
		return true
	}
	switch val.(type) {
	case runtime.NilValue:
		return true
	case *runtime.NilValue:
		return true
	default:
		return false
	}
}

func isCallableRuntimeValue(val runtime.Value) bool {
	switch val.(type) {
	case *runtime.FunctionValue,
		*runtime.FunctionOverloadValue,
		runtime.NativeFunctionValue, *runtime.NativeFunctionValue,
		runtime.BoundMethodValue, *runtime.BoundMethodValue,
		runtime.NativeBoundMethodValue, *runtime.NativeBoundMethodValue,
		runtime.PartialFunctionValue, *runtime.PartialFunctionValue:
		return true
	default:
		return false
	}
}

func isPrimitiveReceiver(val runtime.Value) bool {
	switch v := val.(type) {
	case runtime.StringValue, *runtime.StringValue,
		runtime.BoolValue, runtime.CharValue, runtime.NilValue, *runtime.NilValue,
		runtime.IntegerValue, *runtime.IntegerValue,
		runtime.FloatValue, *runtime.FloatValue,
		*runtime.ArrayValue:
		return true
	case *runtime.InterfaceValue:
		if v != nil {
			return isPrimitiveReceiver(v.Underlying)
		}
	}
	return false
}

func (i *Interpreter) iteratorMember(iter *runtime.IteratorValue, member ast.Expression) (runtime.Value, error) {
	if iter == nil {
		return nil, fmt.Errorf("iterator receiver is nil")
	}
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("iterator member access expects identifier")
	}
	switch ident.Name {
	case "next":
		fn := iteratorNextNativeMethod()
		return &runtime.NativeBoundMethodValue{Receiver: iter, Method: fn}, nil
	case "close":
		fn := iteratorCloseNativeMethod()
		return &runtime.NativeBoundMethodValue{Receiver: iter, Method: fn}, nil
	default:
		ifaceDef := i.interfaces["Iterator"]
		if method, ok, err := i.iteratorInterfaceMethodValue(ifaceDef, ident.Name); err != nil {
			return nil, err
		} else if ok {
			return bindResolvedMethodValue(iter, method, ident.Name)
		}
		return nil, fmt.Errorf("iterator has no member '%s'", ident.Name)
	}
}

func iteratorNextNativeMethod() runtime.NativeFunctionValue {
	return runtime.NativeFunctionValue{
		Name:       "iterator.next",
		Arity:      0,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("next expects only a receiver")
			}
			receiver, ok := args[0].(*runtime.IteratorValue)
			if !ok {
				return nil, fmt.Errorf("next receiver must be an iterator")
			}
			value, done, err := receiver.Next()
			if err != nil {
				return nil, err
			}
			if done {
				return runtime.IteratorEnd, nil
			}
			if value == nil {
				return runtime.NilValue{}, nil
			}
			return value, nil
		},
		RawImpl: func(_ *runtime.NativeCallContext, args []runtime.RawValue) (runtime.RawValue, error) {
			if len(args) != 1 {
				return runtime.RawValue{}, fmt.Errorf("next expects only a receiver")
			}
			receiverValue := args[0].Materialize()
			receiver, ok := receiverValue.(*runtime.IteratorValue)
			if !ok {
				return runtime.RawValue{}, fmt.Errorf("next receiver must be an iterator")
			}
			value, done, err := receiver.NextRaw()
			if err != nil {
				return runtime.RawValue{}, err
			}
			if done {
				return runtime.NewRawValue(runtime.IteratorEnd), nil
			}
			if value.Kind() == runtime.RawValueMaterialized && value.Value() == nil {
				return runtime.NewRawValue(runtime.NilValue{}), nil
			}
			return value, nil
		},
	}
}

func iteratorCloseNativeMethod() runtime.NativeFunctionValue {
	return runtime.NativeFunctionValue{
		Name:       "iterator.close",
		Arity:      0,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("close expects only a receiver")
			}
			receiver, ok := args[0].(*runtime.IteratorValue)
			if !ok {
				return nil, fmt.Errorf("close receiver must be an iterator")
			}
			receiver.Close()
			return runtime.NilValue{}, nil
		},
		RawImpl: func(_ *runtime.NativeCallContext, args []runtime.RawValue) (runtime.RawValue, error) {
			if len(args) != 1 {
				return runtime.RawValue{}, fmt.Errorf("close expects only a receiver")
			}
			receiverValue := args[0].Materialize()
			receiver, ok := receiverValue.(*runtime.IteratorValue)
			if !ok {
				return runtime.RawValue{}, fmt.Errorf("close receiver must be an iterator")
			}
			receiver.Close()
			return runtime.NewRawValue(runtime.NilValue{}), nil
		},
	}
}

func (i *Interpreter) structDefinitionMember(def *runtime.StructDefinitionValue, member ast.Expression) (runtime.Value, error) {
	memberName, err := memberIdentifierName(member, "Static access expects identifier member")
	if err != nil {
		return nil, err
	}
	return i.structDefinitionMemberName(def, memberName)
}

func (i *Interpreter) interfaceDefinitionMember(def *runtime.InterfaceDefinitionValue, member ast.Expression) (runtime.Value, error) {
	memberName, err := memberIdentifierName(member, "Interface access expects identifier member")
	if err != nil {
		return nil, err
	}
	return i.interfaceDefinitionMemberName(def, memberName)
}

func (i *Interpreter) typeRefMember(ref runtime.TypeRefValue, member ast.Expression) (runtime.Value, error) {
	memberName, err := memberIdentifierName(member, "Static access expects identifier member")
	if err != nil {
		return nil, err
	}
	return i.typeRefMemberName(ref, memberName)
}
