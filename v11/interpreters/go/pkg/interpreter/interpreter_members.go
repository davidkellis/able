package interpreter

import (
	"fmt"
	"math"
	"strings"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
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
		return i.structDefinitionMember(v, member)
	case runtime.StructDefinitionValue:
		return i.structDefinitionMember(&v, member)
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
		return i.arrayMemberWithOverrides(v, member, env)
	case *runtime.HashMapValue:
		i.ensureHashMapBuiltins()
		return i.hashMapMember(v, member)
	case *runtime.HasherValue:
		return i.hasherMember(v, member)
	case *runtime.ProcHandleValue:
		return i.procHandleMember(v, member)
	case *runtime.FutureValue:
		return i.futureMember(v, member)
	case *runtime.IteratorValue:
		return i.iteratorMember(v, member)
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

func (i *Interpreter) arrayMemberWithOverrides(arr *runtime.ArrayValue, member ast.Expression, env *runtime.Environment) (runtime.Value, error) {
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
	if method, err := i.findIndexMethod(obj, "get", "Index"); err == nil && method != nil {
		return i.CallFunction(method, []runtime.Value{obj, idxVal})
	} else if err != nil {
		return nil, err
	}
	arr, err := i.toArrayValue(obj)
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
	val := state.values[idx]
	if val == nil {
		return nil, fmt.Errorf("Array index out of bounds")
	}
	return val, nil
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
				if intVal, ok := raw.(runtime.IntegerValue); ok && intVal.Val != nil && intVal.Val.IsInt64() {
					handle = intVal.Val.Int64()
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
	return i.findMethod(info, methodName, iface)
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
	return i.findMethod(info, "apply", "Apply")
}

func indexFromValue(val runtime.Value) (int, error) {
	switch v := val.(type) {
	case runtime.IntegerValue:
		if v.Val == nil || !v.Val.IsInt64() {
			return 0, fmt.Errorf("Array index must be within int range")
		}
		return int(v.Val.Int64()), nil
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
		fn := runtime.NativeFunctionValue{
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
		return &runtime.NativeBoundMethodValue{Receiver: iter, Method: fn}, nil
	default:
		return nil, fmt.Errorf("iterator has no member '%s'", ident.Name)
	}
}

func (i *Interpreter) resolveMethodFromPool(env *runtime.Environment, funcName string, receiver runtime.Value, ifaceFilter string) (runtime.Value, error) {
	functionCandidates := make([]*runtime.FunctionValue, 0)
	nativeCandidates := make([]runtime.Value, 0)
	seenFuncs := make(map[*runtime.FunctionValue]struct{})
	seenNatives := make(map[string]struct{})
	nameInScope := env != nil && env.Has(funcName)
	allowInherent := nameInScope || isPrimitiveReceiver(receiver)

	var addCallable func(method runtime.Value, privacyContext string) error
	addCallable = func(method runtime.Value, privacyContext string) error {
		switch fn := method.(type) {
		case *runtime.FunctionValue:
			if fn == nil {
				return nil
			}
			if fnDecl, ok := fn.Declaration.(*ast.FunctionDefinition); ok && fnDecl != nil && fnDecl.IsPrivate {
				name := privacyContext
				if name == "" {
					name = "<unknown>"
				}
				return fmt.Errorf("Method '%s' on %s is private", funcName, name)
			}
			if _, ok := seenFuncs[fn]; ok {
				return nil
			}
			seenFuncs[fn] = struct{}{}
			functionCandidates = append(functionCandidates, fn)
		case *runtime.FunctionOverloadValue:
			if fn == nil {
				return nil
			}
			for _, entry := range fn.Overloads {
				if entry == nil {
					continue
				}
				if err := addCallable(entry, privacyContext); err != nil {
					return err
				}
			}
		case runtime.NativeFunctionValue:
			key := fn.Name
			if key == "" {
				key = fmt.Sprintf("%p", &fn)
			}
			if _, ok := seenNatives[key]; ok {
				return nil
			}
			seenNatives[key] = struct{}{}
			nativeCandidates = append(nativeCandidates, runtime.NativeBoundMethodValue{Receiver: receiver, Method: fn})
		case *runtime.NativeFunctionValue:
			if fn == nil {
				return nil
			}
			key := fn.Name
			if key == "" {
				key = fmt.Sprintf("%p", fn)
			}
			if _, ok := seenNatives[key]; ok {
				return nil
			}
			seenNatives[key] = struct{}{}
			nativeCandidates = append(nativeCandidates, runtime.NativeBoundMethodValue{Receiver: receiver, Method: *fn})
		case runtime.NativeBoundMethodValue:
			key := fn.Method.Name
			if key == "" {
				key = fmt.Sprintf("%p", &fn)
			}
			if _, ok := seenNatives[key]; ok {
				return nil
			}
			seenNatives[key] = struct{}{}
			nativeCandidates = append(nativeCandidates, fn)
		case *runtime.NativeBoundMethodValue:
			if fn == nil {
				return nil
			}
			key := fn.Method.Name
			if key == "" {
				key = fmt.Sprintf("%p", fn)
			}
			if _, ok := seenNatives[key]; ok {
				return nil
			}
			seenNatives[key] = struct{}{}
			nativeCandidates = append(nativeCandidates, *fn)
		}
		return nil
	}

	if info, ok := i.getTypeInfoForValue(receiver); ok {
		if allowInherent {
			for _, name := range i.canonicalTypeNames(info.name) {
				if bucket, ok := i.inherentMethods[name]; ok {
					if method := bucket[funcName]; method != nil {
						if callable, ok := i.selectUfcsCallable(method, receiver, true); ok {
							if err := addCallable(callable, name); err != nil {
								return nil, err
							}
						}
					}
				}
			}
		}
		existing := len(functionCandidates) + len(nativeCandidates)
		if method, err := i.findMethod(info, funcName, ifaceFilter); err == nil && method != nil {
			if err := addCallable(method, info.name); err != nil {
				return nil, err
			}
		} else if err != nil {
			if existing == 0 {
				return nil, err
			}
		}
	}

	hasMethodCandidate := len(functionCandidates)+len(nativeCandidates) > 0

	if env != nil && !hasMethodCandidate && nameInScope {
		if val, err := env.Get(funcName); err == nil {
			if callable, ok := i.selectUfcsCallable(val, receiver, false); ok {
				if err := addCallable(callable, ""); err != nil {
					return nil, err
				}
			}
		}
	}

	if len(functionCandidates) > 0 && len(nativeCandidates) > 0 {
		return nil, fmt.Errorf("Ambiguous overload for %s", funcName)
	}
	if len(functionCandidates) > 0 {
		var callable runtime.Value
		if len(functionCandidates) == 1 {
			callable = functionCandidates[0]
		} else {
			callable = &runtime.FunctionOverloadValue{Overloads: functionCandidates}
		}
		return &runtime.BoundMethodValue{Receiver: receiver, Method: callable}, nil
	}
	if len(nativeCandidates) > 0 {
		if len(nativeCandidates) > 1 {
			return nil, fmt.Errorf("Ambiguous overload for %s", funcName)
		}
		return nativeCandidates[0], nil
	}
	return nil, nil
}

func (i *Interpreter) selectUfcsCallable(method runtime.Value, receiver runtime.Value, requireSelf bool) (runtime.Value, bool) {
	switch fn := method.(type) {
	case *runtime.FunctionValue:
		if (!requireSelf || functionExpectsSelf(fn)) && i.functionFirstParamMatches(fn, receiver) {
			if !requireSelf && fn.TypeQualified {
				return nil, false
			}
			return fn, true
		}
		return nil, false
	case *runtime.FunctionOverloadValue:
		filtered := make([]*runtime.FunctionValue, 0, len(fn.Overloads))
		for _, entry := range fn.Overloads {
			if entry == nil {
				continue
			}
			if !requireSelf && entry.TypeQualified {
				continue
			}
			if requireSelf && !functionExpectsSelf(entry) {
				continue
			}
			if !i.functionFirstParamMatches(entry, receiver) {
				continue
			}
			filtered = append(filtered, entry)
		}
		if len(filtered) > 1 {
			if recvType := i.typeExpressionFromValue(receiver); recvType != nil {
				narrowed := make([]*runtime.FunctionValue, 0, len(filtered))
				for _, entry := range filtered {
					if def, ok := entry.Declaration.(*ast.FunctionDefinition); ok && def != nil {
						if len(def.Params) > 0 && def.Params[0] != nil && def.Params[0].ParamType != nil {
							if typeExpressionsEqual(def.Params[0].ParamType, recvType) {
								narrowed = append(narrowed, entry)
							}
						}
					}
				}
				if len(narrowed) == 1 {
					return narrowed[0], true
				}
				if len(narrowed) > 0 {
					filtered = narrowed
				}
			}
		}
		if len(filtered) == 0 {
			return nil, false
		}
		if len(filtered) == 1 {
			return filtered[0], true
		}
		return &runtime.FunctionOverloadValue{Overloads: filtered}, true
	default:
		return nil, false
	}
}

func functionExpectsSelf(fn *runtime.FunctionValue) bool {
	if fn == nil {
		return false
	}
	decl, ok := fn.Declaration.(*ast.FunctionDefinition)
	if !ok || decl == nil {
		return false
	}
	if decl.IsMethodShorthand {
		return true
	}
	if len(decl.Params) == 0 {
		return false
	}
	first := decl.Params[0]
	if first == nil {
		return false
	}
	if ident, ok := first.Name.(*ast.Identifier); ok && strings.EqualFold(ident.Name, "self") {
		return true
	}
	if simple, ok := first.ParamType.(*ast.SimpleTypeExpression); ok && simple.Name != nil && simple.Name.Name == "Self" {
		return true
	}
	return false
}

func (i *Interpreter) functionFirstParamMatches(fn *runtime.FunctionValue, receiver runtime.Value) bool {
	if fn == nil || fn.Declaration == nil {
		return false
	}
	switch decl := fn.Declaration.(type) {
	case *ast.FunctionDefinition:
		if decl.IsMethodShorthand {
			return true
		}
		if len(decl.Params) == 0 {
			return false
		}
		first := decl.Params[0]
		if first == nil {
			return false
		}
		if first.ParamType == nil {
			return true
		}
		return i.matchesType(first.ParamType, receiver)
	case *ast.LambdaExpression:
		if len(decl.Params) == 0 {
			return false
		}
		first := decl.Params[0]
		if first == nil {
			return false
		}
		if first.ParamType == nil {
			return true
		}
		return i.matchesType(first.ParamType, receiver)
	default:
		return false
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
	if bucket == nil {
		return nil, fmt.Errorf("No static method '%s' for %s", ident.Name, typeName)
	}
	method, ok := bucket[ident.Name]
	if !ok {
		return nil, fmt.Errorf("No static method '%s' for %s", ident.Name, typeName)
	}
	if fn := firstFunction(method); fn != nil {
		if fnDef, ok := fn.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
			return nil, fmt.Errorf("Method '%s' on %s is private", ident.Name, typeName)
		}
	}
	return method, nil
}

func (i *Interpreter) packageMemberAccess(pkg runtime.PackageValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Package member access expects identifier")
	}
	if pkg.Public == nil {
		return nil, fmt.Errorf("Package has no public members")
	}
	val, ok := pkg.Public[ident.Name]
	if !ok {
		pkgName := pkg.Name
		if pkgName == "" {
			pkgName = strings.Join(pkg.NamePath, ".")
		}
		if pkgName == "" {
			pkgName = "<package>"
		}
		return nil, fmt.Errorf("No public member '%s' on package %s", ident.Name, pkgName)
	}
	return val, nil
}

func (i *Interpreter) dynPackageMemberAccess(pkg runtime.DynPackageValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Dyn package member access expects identifier")
	}
	pkgName := pkg.Name
	if pkgName == "" {
		pkgName = strings.Join(pkg.NamePath, ".")
	}
	if pkgName == "" {
		return nil, fmt.Errorf("Dyn package missing name")
	}
	bucket, ok := i.packageRegistry[pkgName]
	if !ok {
		return nil, fmt.Errorf("dyn package '%s' not found", pkgName)
	}
	sym, ok := bucket[ident.Name]
	if !ok {
		return nil, fmt.Errorf("dyn package '%s' has no member '%s'", pkgName, ident.Name)
	}
	if isPrivateSymbol(sym) {
		return nil, fmt.Errorf("dyn package '%s' member '%s' is private", pkgName, ident.Name)
	}
	return runtime.DynRefValue{Package: pkgName, Name: ident.Name}, nil
}

func (i *Interpreter) implNamespaceMember(ns runtime.ImplementationNamespaceValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Impl namespace member access expects identifier")
	}
	if ns.Methods == nil {
		return nil, fmt.Errorf("Impl namespace has no methods")
	}
	method, ok := ns.Methods[ident.Name]
	if !ok {
		name := "<impl>"
		if ns.Name != nil {
			name = ns.Name.Name
		}
		return nil, fmt.Errorf("No method '%s' on impl %s", ident.Name, name)
	}
	return method, nil
}

func (i *Interpreter) interfaceMember(val *runtime.InterfaceValue, member ast.Expression) (runtime.Value, error) {
	if val == nil {
		return nil, fmt.Errorf("Interface value is nil")
	}
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Interface member access expects identifier")
	}
	ifaceName := ""
	if val.Interface != nil && val.Interface.Node != nil && val.Interface.Node.ID != nil {
		ifaceName = val.Interface.Node.ID.Name
	}
	if ifaceName == "" {
		return nil, fmt.Errorf("Unknown interface for member access")
	}
	var method runtime.Value
	if val.Methods != nil {
		method = val.Methods[ident.Name]
	}
	if method == nil {
		if info, ok := i.getTypeInfoForValue(val.Underlying); ok {
			resolved, err := i.findMethod(info, ident.Name, ifaceName)
			if err != nil {
				return nil, err
			}
			method = resolved
			if method != nil {
				if val.Methods == nil {
					val.Methods = make(map[string]runtime.Value)
				}
				val.Methods[ident.Name] = method
			}
		}
	}
	if method == nil {
		return nil, fmt.Errorf("No method '%s' for interface %s", ident.Name, ifaceName)
	}
	if fn := firstFunction(method); fn != nil {
		if fnDef, ok := fn.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
			return nil, fmt.Errorf("Method '%s' on %s is private", ident.Name, ifaceName)
		}
	}
	return &runtime.BoundMethodValue{Receiver: val.Underlying, Method: method}, nil
}

func (i *Interpreter) resolveDynRef(ref runtime.DynRefValue) (runtime.Value, error) {
	bucket, ok := i.packageRegistry[ref.Package]
	if !ok {
		return nil, fmt.Errorf("dyn ref '%s.%s' not found", ref.Package, ref.Name)
	}
	val, ok := bucket[ref.Name]
	if !ok {
		return nil, fmt.Errorf("dyn ref '%s.%s' not found", ref.Package, ref.Name)
	}
	if runtime.IsFunctionLike(val) {
		return val, nil
	}
	return nil, fmt.Errorf("dyn ref '%s.%s' is not callable", ref.Package, ref.Name)
}

func toStructDefinitionValue(val runtime.Value, name string) (*runtime.StructDefinitionValue, error) {
	switch v := val.(type) {
	case *runtime.StructDefinitionValue:
		return v, nil
	case runtime.StructDefinitionValue:
		return &v, nil
	default:
		return nil, fmt.Errorf("'%s' is not a struct type", name)
	}
}
