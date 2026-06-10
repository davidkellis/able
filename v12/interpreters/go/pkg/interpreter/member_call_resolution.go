package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func directCallMemberNeedsMemberAccessFallback(receiver runtime.Value, memberName string) bool {
	if memberName == "" {
		return false
	}
	inst, ok := receiver.(*runtime.StructInstanceValue)
	if !ok || inst == nil {
		return false
	}
	val, ok := structNamedFieldValue(inst, memberName)
	return ok && isCallableRuntimeValue(val)
}

func (i *Interpreter) resolveStaticMemberCallable(receiver runtime.Value, memberName string) (runtime.Value, bool, error) {
	if i == nil || memberName == "" {
		return nil, false, nil
	}
	switch v := receiver.(type) {
	case runtime.TypeRefValue:
		callable, err := i.typeRefMemberName(v, memberName)
		return callable, err == nil, err
	case *runtime.TypeRefValue:
		if v == nil {
			return nil, false, nil
		}
		callable, err := i.typeRefMemberName(*v, memberName)
		return callable, err == nil, err
	case runtime.StructDefinitionValue:
		callable, err := i.structDefinitionMemberName(&v, memberName)
		return callable, err == nil, err
	case *runtime.StructDefinitionValue:
		if v == nil {
			return nil, false, nil
		}
		callable, err := i.structDefinitionMemberName(v, memberName)
		return callable, err == nil, err
	case runtime.InterfaceDefinitionValue:
		callable, err := i.interfaceDefinitionMemberName(&v, memberName)
		return callable, err == nil, err
	case *runtime.InterfaceDefinitionValue:
		if v == nil {
			return nil, false, nil
		}
		callable, err := i.interfaceDefinitionMemberName(v, memberName)
		return callable, err == nil, err
	case runtime.PackageValue:
		callable, err := i.packageMemberAccessName(v, memberName)
		return callable, err == nil, err
	case *runtime.PackageValue:
		if v == nil {
			return nil, false, nil
		}
		callable, err := i.packageMemberAccessName(*v, memberName)
		return callable, err == nil, err
	case runtime.ImplementationNamespaceValue:
		callable, err := i.implNamespaceMemberName(v, memberName)
		return callable, err == nil, err
	case *runtime.ImplementationNamespaceValue:
		if v == nil {
			return nil, false, nil
		}
		callable, err := i.implNamespaceMemberName(*v, memberName)
		return callable, err == nil, err
	case runtime.DynPackageValue:
		callable, err := i.dynPackageMemberAccessName(v, memberName)
		return callable, err == nil, err
	case *runtime.DynPackageValue:
		if v == nil {
			return nil, false, nil
		}
		callable, err := i.dynPackageMemberAccessName(*v, memberName)
		return callable, err == nil, err
	default:
		return nil, false, nil
	}
}

func (i *Interpreter) resolveDirectCallMemberCallable(env *runtime.Environment, receiver runtime.Value, memberName string, receiverTypeHint ast.TypeExpression) (runtime.Value, runtime.Value, bool, bool, error) {
	if i == nil || memberName == "" {
		return nil, nil, false, false, nil
	}
	if directCallMemberNeedsMemberAccessFallback(receiver, memberName) {
		return nil, nil, false, false, nil
	}
	if callable, found := i.resolveStaticGenericUnionMethodCallable(env, memberName, receiverTypeHint); found {
		return callable, receiver, true, true, nil
	}
	if iface, ok := receiver.(*runtime.InterfaceValue); ok && iface != nil {
		callable, found, err := i.resolveInterfaceMethodCallable(iface, memberName)
		if err != nil || found {
			return callable, interfaceMethodReceiver(i, iface, callable), true, found, err
		}
	}
	if callable, found, err := i.resolveMethodCallableFromPool(env, memberName, receiver, ""); err != nil || found {
		callable = i.narrowGenericUnionMethodCallable(callable, receiverTypeHint)
		return callable, receiver, true, found, err
	}
	callable, found, err := i.resolveStaticMemberCallable(receiver, memberName)
	return callable, nil, false, found, err
}

func (i *Interpreter) resolveStaticGenericUnionMethodCallable(env *runtime.Environment, memberName string, receiverTypeHint ast.TypeExpression) (runtime.Value, bool) {
	if i == nil || env == nil || memberName == "" || receiverTypeHint == nil {
		return nil, false
	}
	scoped, filter, found := i.lookupMethodScopeCallable(env, memberName)
	if !found {
		return nil, false
	}
	matched := i.staticGenericUnionMethodMatches(functionOverloadsView(scoped), receiverTypeHint, filter)
	switch len(matched) {
	case 0:
		return nil, false
	case 1:
		return matched[0], true
	default:
		return &runtime.FunctionOverloadValue{Overloads: matched}, true
	}
}

func (i *Interpreter) staticGenericUnionMethodMatches(candidates []*runtime.FunctionValue, receiverTypeHint ast.TypeExpression, filter functionScopeFilter) []*runtime.FunctionValue {
	matched := make([]*runtime.FunctionValue, 0, 2)
	for _, fn := range candidates {
		if !filter.contains(fn) || !functionExpectsSelf(fn) || !i.genericUnionMethodMatchesStaticReceiver(fn, receiverTypeHint) {
			continue
		}
		matched = append(matched, fn)
	}
	return matched
}

func (i *Interpreter) narrowGenericUnionMethodCallable(callable runtime.Value, receiverTypeHint ast.TypeExpression) runtime.Value {
	if receiverTypeHint == nil {
		return callable
	}
	overloads, ok := callable.(*runtime.FunctionOverloadValue)
	if !ok || overloads == nil || len(overloads.Overloads) < 2 {
		return callable
	}
	matched := make([]*runtime.FunctionValue, 0, len(overloads.Overloads))
	for _, fn := range overloads.Overloads {
		if i.genericUnionMethodMatchesStaticReceiver(fn, receiverTypeHint) {
			matched = append(matched, fn)
		}
	}
	if len(matched) == 0 || len(matched) == len(overloads.Overloads) {
		return callable
	}
	if len(matched) == 1 {
		return matched[0]
	}
	return &runtime.FunctionOverloadValue{Overloads: matched}
}

// CallStaticGenericUnionMember executes only a generic named-union method
// whose receiver type was preserved by checked source or compiled lowering.
// The bool reports whether this narrow path applied; callers retain their
// existing member lookup for every other method call.
func (i *Interpreter) CallStaticGenericUnionMember(obj runtime.Value, memberName string, args []runtime.Value, call *ast.FunctionCall, env *runtime.Environment) (runtime.Value, bool, error) {
	if i == nil || call == nil || env == nil || memberName == "" {
		return nil, false, nil
	}
	receiverType := i.staticReceiverTypeForCall(call, env)
	if receiverType == nil {
		return nil, false, nil
	}
	callable, found := i.resolveStaticGenericUnionMethodCallable(env, memberName, receiverType)
	if !found {
		return nil, false, nil
	}
	return i.callStaticGenericUnionMethodCallable(callable, obj, args, call, env)
}

// CallStaticGenericUnionMemberFromCandidates is the compiled-boundary form of
// CallStaticGenericUnionMember. Compiled package registration retains the
// original method values alongside native wrappers; selecting from those
// values keeps MethodSet generic metadata available without a native/original
// overload tie.
func (i *Interpreter) CallStaticGenericUnionMemberFromCandidates(candidates []runtime.Value, obj runtime.Value, memberName string, args []runtime.Value, call *ast.FunctionCall, env *runtime.Environment) (runtime.Value, bool, error) {
	if i == nil || call == nil || env == nil || memberName == "" || len(candidates) == 0 {
		return nil, false, nil
	}
	receiverType := i.staticReceiverTypeForCall(call, env)
	if receiverType == nil {
		return nil, false, nil
	}
	matched := make([]*runtime.FunctionValue, 0, 2)
	for _, candidate := range candidates {
		matched = append(matched, i.staticGenericUnionMethodMatches(functionOverloadsView(candidate), receiverType, functionScopeFilter{})...)
	}
	if len(matched) == 0 {
		return nil, false, nil
	}
	var callable runtime.Value = matched[0]
	if len(matched) > 1 {
		callable = &runtime.FunctionOverloadValue{Overloads: matched}
	}
	return i.callStaticGenericUnionMethodCallable(callable, obj, args, call, env)
}

func (i *Interpreter) callStaticGenericUnionMethodCallable(callable runtime.Value, obj runtime.Value, args []runtime.Value, call *ast.FunctionCall, env *runtime.Environment) (runtime.Value, bool, error) {
	value, err := i.callCallableValueWithInjectedReceiver(callable, obj, args, env, call, false)
	if err != nil {
		return nil, true, err
	}
	return value, true, err
}
