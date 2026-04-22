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
	if _, err := i.ensureArrayState(arr, 0); err != nil {
		return nil, err
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
		var handle int64
		if v.Fields != nil {
			if raw, ok := v.Fields["storage_handle"]; ok {
				if intVal, ok := raw.(runtime.IntegerValue); ok {
					if h, ok := intVal.ToInt64(); ok {
						handle = h
					}
				}
			}
		}
		if handle != 0 {
			return i.arrayValueFromHandle(handle, 0, 0)
		}
		return i.newArrayValue(nil, 0), nil
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
			i.syncArrayValues(inst.Handle, state)
			return value, nil
		case *ast.Identifier:
			state, err := i.ensureArrayState(inst, 0)
			if err != nil {
				return nil, err
			}
			switch member.Name {
			case "storage_handle":
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
				newState, err := runtime.ArrayStoreEnsureHandle(handle, 0, 0)
				if err != nil {
					return nil, err
				}
				i.trackArrayValue(handle, inst)
				inst.Elements = newState.Values
				i.syncArrayValues(handle, newState)
				return value, nil
			case "length":
				newLen, err := arrayIndexFromValue(value)
				if err != nil {
					return nil, fmt.Errorf("array length must be a non-negative integer")
				}
				setArrayLength(state, newLen)
				i.syncArrayValues(inst.Handle, state)
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
				} else if newCap > state.Capacity {
					state.Capacity = newCap
				}
				i.syncArrayValues(inst.Handle, state)
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
		if inst.Fields == nil {
			return nil, fmt.Errorf("Expected named struct instance")
		}
		if preferMethods {
			if val, ok := inst.Fields[ident.Name]; ok {
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
			if val, ok := inst.Fields[ident.Name]; ok {
				return val, nil
			}
		} else {
			if val, ok := inst.Fields[ident.Name]; ok {
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
	default:
		if ifaceVal, err := i.coerceToInterfaceValue("Iterator", iter, nil); err == nil {
			if iface, ok := ifaceVal.(*runtime.InterfaceValue); ok {
				return i.interfaceMember(iface, member)
			}
			if iface, ok := ifaceVal.(runtime.InterfaceValue); ok {
				return i.interfaceMember(&iface, member)
			}
		}
		return nil, fmt.Errorf("iterator has no member '%s'", ident.Name)
	}
}

func iteratorNextNativeMethod() runtime.NativeFunctionValue {
	return runtime.NativeFunctionValue{
		Name:  "iterator.next",
		Arity: 0,
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
	}
}

func (i *Interpreter) structDefinitionMember(def *runtime.StructDefinitionValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Static access expects identifier member")
	}
	if def == nil || def.Node == nil || def.Node.ID == nil {
		return nil, fmt.Errorf("struct definition missing identifier")
	}
	typeName := def.Node.ID.Name
	bucket := i.inherentMethods[typeName]
	var method runtime.Value
	var found bool
	if bucket != nil {
		method, found = bucket[ident.Name]
	}
	if !found {
		candidate, err := i.findMethodCached(typeInfo{name: typeName}, ident.Name, "")
		if err != nil {
			return nil, err
		}
		method = candidate
	}
	if method == nil {
		return nil, fmt.Errorf("No static method '%s' for %s", ident.Name, typeName)
	}
	if fn := firstFunction(method); fn != nil {
		if fnDef, ok := fn.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
			return nil, fmt.Errorf("Method '%s' on %s is private", ident.Name, typeName)
		}
	}
	return method, nil
}

func (i *Interpreter) interfaceDefinitionMember(def *runtime.InterfaceDefinitionValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Interface access expects identifier member")
	}
	if def == nil || def.Node == nil || def.Node.ID == nil {
		return nil, fmt.Errorf("interface definition missing identifier")
	}
	ifaceName := def.Node.ID.Name
	var sig *ast.FunctionSignature
	for _, candidate := range def.Node.Signatures {
		if candidate == nil || candidate.Name == nil || candidate.Name.Name != ident.Name {
			continue
		}
		sig = candidate
		break
	}
	if sig == nil {
		return nil, fmt.Errorf("No method '%s' for interface %s", ident.Name, ifaceName)
	}
	arity := len(sig.Params)
	fn := runtime.NativeFunctionValue{
		Name:  fmt.Sprintf("%s.%s", ifaceName, ident.Name),
		Arity: arity,
		Impl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("%s.%s requires a receiver", ifaceName, ident.Name)
			}
			receiver := unwrapInterfaceValue(args[0])
			method, err := i.resolveInterfaceMethod(receiver, ifaceName, ident.Name)
			if err != nil {
				return nil, err
			}
			if method == nil {
				return nil, fmt.Errorf("No method '%s' for interface %s", ident.Name, ifaceName)
			}
			callArgs := append([]runtime.Value{receiver}, args[1:]...)
			return i.callCallableValue(method, callArgs, ctx.Env, nil)
		},
	}
	return fn, nil
}

func (i *Interpreter) typeRefMember(ref runtime.TypeRefValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Static access expects identifier member")
	}
	typeName := ref.TypeName
	if typeName == "" {
		return nil, fmt.Errorf("type reference missing name")
	}
	bucket := i.inherentMethods[typeName]
	var method runtime.Value
	var found bool
	if bucket != nil {
		method, found = bucket[ident.Name]
	}
	if !found {
		candidate, err := i.findMethod(typeInfo{name: typeName, typeArgs: ref.TypeArgs}, ident.Name, "", nil)
		if err != nil {
			return nil, err
		}
		method = candidate
	}
	if method == nil {
		return nil, fmt.Errorf("No static method '%s' for %s", ident.Name, typeName)
	}
	if fn := firstFunction(method); fn != nil {
		if fnDef, ok := fn.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
			return nil, fmt.Errorf("Method '%s' on %s is private", ident.Name, typeName)
		}
	}
	return method, nil
}
