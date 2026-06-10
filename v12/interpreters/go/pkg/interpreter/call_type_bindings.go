package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type callLocalTypeBindingCacheKey struct {
	function     *runtime.FunctionValue
	receiverType string
}

func (i *Interpreter) receiverTypeName(receiver runtime.Value) string {
	if i == nil || receiver == nil {
		return ""
	}
	if info, ok := i.getTypeInfoForValue(receiver); ok {
		return i.cachedTypeInfoName(info)
	}
	if actual := i.typeExpressionFromValue(receiver); actual != nil {
		return typeExpressionToString(actual)
	}
	return ""
}

func (i *Interpreter) implMethodContextFromEnv(env *runtime.Environment) *implMethodContext {
	if env == nil {
		return nil
	}
	data := i.runtimeDataFromEnv(env)
	ctx, ok := data.(*implMethodContext)
	if !ok || ctx == nil {
		return nil
	}
	return ctx
}

func (i *Interpreter) functionNeedsCallLocalTypeBindings(fn *runtime.FunctionValue) bool {
	if fn == nil {
		return false
	}
	plan := i.functionRuntimeGenericBindingPlan(fn)
	if plan == nil {
		return false
	}
	return plan.callLocalUsed
}

func (i *Interpreter) callLocalTypeBindingValuesIfAny(fn *runtime.FunctionValue, receiver runtime.Value) []runtime.EnvironmentBinding {
	runtimeBindings, _ := i.callLocalTypeBindingValuesAndReceiverTypeIfAny(fn, receiver)
	return runtimeBindings
}

func (i *Interpreter) callLocalTypeBindingValuesAndReceiverTypeIfAny(fn *runtime.FunctionValue, receiver runtime.Value) ([]runtime.EnvironmentBinding, string) {
	return i.callLocalTypeBindingValuesAndStaticReceiverTypeIfAny(fn, receiver, nil)
}

func (i *Interpreter) callLocalTypeBindingValuesAndStaticReceiverTypeIfAny(fn *runtime.FunctionValue, receiver runtime.Value, receiverTypeHint ast.TypeExpression) ([]runtime.EnvironmentBinding, string) {
	if i == nil || fn == nil || receiver == nil {
		return nil, ""
	}
	var (
		actual       ast.TypeExpression
		cacheKey     callLocalTypeBindingCacheKey
		cacheOK      bool
		receiverType string
	)
	if receiverTypeHint != nil {
		actual = receiverTypeHint
		receiverType = typeExpressionToString(actual)
		cacheKey = callLocalTypeBindingCacheKey{function: fn, receiverType: receiverType}
		cacheOK = cacheKey.function != nil && cacheKey.receiverType != ""
		if cacheOK {
			if cached, ok := i.lookupCallLocalTypeBindingCache(cacheKey); ok {
				return cached, receiverType
			}
		}
	} else if info, ok := i.getTypeInfoForValue(receiver); ok {
		receiverType = i.cachedTypeInfoName(info)
		cacheKey = callLocalTypeBindingCacheKey{function: fn, receiverType: receiverType}
		if cached, ok := i.lookupCallLocalTypeBindingCache(cacheKey); ok {
			return cached, receiverType
		}
		actual = i.cachedTypeExpressionFromInfo(info)
		cacheOK = cacheKey.function != nil && cacheKey.receiverType != ""
	}
	if actual == nil {
		actual = i.typeExpressionFromValue(receiver)
	}
	if actual == nil {
		return nil, receiverType
	}
	if receiverType == "" {
		receiverType = typeExpressionToString(actual)
		cacheKey = callLocalTypeBindingCacheKey{function: fn, receiverType: receiverType}
		cacheOK = cacheKey.function != nil && cacheKey.receiverType != ""
		if cacheOK {
			if cached, ok := i.lookupCallLocalTypeBindingCache(cacheKey); ok {
				return cached, receiverType
			}
		}
	}
	actualExpanded := i.expandTypeAliasesCached(actual)
	if actualExpanded == nil {
		actualExpanded = actual
	}
	bindings := make(map[string]ast.TypeExpression)
	if fn.MethodSet != nil {
		if len(fn.MethodSet.GenericParams) > 0 && fn.MethodSet.TargetType != nil {
			target, isGenericUnion := i.genericUnionMethodTarget(fn)
			if !isGenericUnion {
				target = i.expandTypeAliasesCached(fn.MethodSet.TargetType)
				if target == nil {
					target = fn.MethodSet.TargetType
				}
			} else {
				actualExpanded = genericUnionStructuralTypeExpression(actualExpanded)
			}
			matchTypeExpressionTemplate(target, actualExpanded, genericNameSet(fn.MethodSet.GenericParams), bindings)
		}
		bindings["Self"] = actualExpanded
		bindings["SelfType"] = actualExpanded
		i.addStructTypeArgBindings(bindings, receiver)
	}
	if ctx := i.implMethodContextFromEnv(fn.Closure); ctx != nil {
		i.bindInterfaceSelfPatternBindings(bindings, ctx.interfaceName, actualExpanded)
	}
	runtimeBindings := buildCallLocalTypeBindingRuntimeValues(bindings)
	if cacheOK && len(runtimeBindings) > 0 {
		i.storeCallLocalTypeBindingCache(cacheKey, runtimeBindings)
	}
	return runtimeBindings, receiverType
}

func (i *Interpreter) bindCallLocalTypeBindings(fn *runtime.FunctionValue, receiver runtime.Value, env *runtime.Environment) {
	if env == nil {
		return
	}
	runtimeBindings := i.callLocalTypeBindingValuesIfAny(fn, receiver)
	env.DefineWithoutMergeBindings(runtimeBindings)
}

func (i *Interpreter) bindInterfaceSelfPatternBindings(bindings map[string]ast.TypeExpression, interfaceName string, actual ast.TypeExpression) {
	if i == nil || actual == nil || interfaceName == "" {
		return
	}
	iface := i.interfaces[interfaceName]
	if iface == nil || iface.Node == nil || iface.Node.SelfTypePattern == nil {
		return
	}
	pattern := i.expandTypeAliasesCached(iface.Node.SelfTypePattern)
	if pattern == nil {
		pattern = iface.Node.SelfTypePattern
	}
	genericNames := make(map[string]struct{}, len(bindings)+4)
	for name := range bindings {
		if name != "" {
			genericNames[name] = struct{}{}
		}
	}
	for name := range i.typePatternBindingNames(pattern, iface.Env, make(map[string]struct{})) {
		genericNames[name] = struct{}{}
	}
	if len(genericNames) == 0 {
		return
	}
	matchTypeExpressionTemplate(pattern, actual, genericNames, bindings)
	promoteHigherKindedSelfBindings(pattern, actual, bindings)
}

func (i *Interpreter) typePatternBindingNames(expr ast.TypeExpression, env *runtime.Environment, names map[string]struct{}) map[string]struct{} {
	if names == nil {
		names = make(map[string]struct{})
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return names
		}
		name := t.Name.Name
		if i.shouldTreatTypeNameAsBinding(name, env) {
			names[name] = struct{}{}
		}
	case *ast.GenericTypeExpression:
		if t == nil {
			return names
		}
		i.typePatternBindingNames(t.Base, env, names)
		for _, arg := range t.Arguments {
			i.typePatternBindingNames(arg, env, names)
		}
	case *ast.NullableTypeExpression:
		if t != nil {
			i.typePatternBindingNames(t.InnerType, env, names)
		}
	case *ast.ResultTypeExpression:
		if t != nil {
			i.typePatternBindingNames(t.InnerType, env, names)
		}
	case *ast.UnionTypeExpression:
		if t != nil {
			for _, member := range t.Members {
				i.typePatternBindingNames(member, env, names)
			}
		}
	case *ast.FunctionTypeExpression:
		if t != nil {
			for _, param := range t.ParamTypes {
				i.typePatternBindingNames(param, env, names)
			}
			i.typePatternBindingNames(t.ReturnType, env, names)
		}
	}
	return names
}

func (i *Interpreter) shouldTreatTypeNameAsBinding(name string, env *runtime.Environment) bool {
	if name == "" {
		return false
	}
	switch name {
	case "Self", "SelfType":
		return true
	case "void", "nil":
		return false
	}
	if isPrimitiveTypeName(name) {
		return false
	}
	if _, ok := i.interfaces[name]; ok {
		return false
	}
	if env != nil {
		if val, ok := env.Lookup(name); ok {
			switch val.(type) {
			case *runtime.StructDefinitionValue,
				runtime.StructDefinitionValue,
				*runtime.InterfaceDefinitionValue,
				runtime.InterfaceDefinitionValue,
				*runtime.UnionDefinitionValue,
				runtime.UnionDefinitionValue:
				return false
			}
		}
	}
	return true
}

func (i *Interpreter) defineTypeBindingValues(env *runtime.Environment, bindings map[string]ast.TypeExpression) {
	if env == nil || len(bindings) == 0 {
		return
	}
	env.DefineWithoutMergeBindings(buildCallLocalTypeBindingRuntimeValues(bindings))
}

func buildCallLocalTypeBindingRuntimeValues(bindings map[string]ast.TypeExpression) []runtime.EnvironmentBinding {
	if len(bindings) == 0 {
		return nil
	}
	values := make([]runtime.EnvironmentBinding, 0, len(bindings)*2)
	for name, expr := range bindings {
		values = appendCallLocalTypeBindingRuntimeValues(values, name, expr)
	}
	if len(values) == 0 {
		return nil
	}
	return values
}

func appendCallLocalTypeBindingRuntimeValues(values []runtime.EnvironmentBinding, name string, expr ast.TypeExpression) []runtime.EnvironmentBinding {
	if name == "" || expr == nil {
		return values
	}
	values = append(values, runtime.EnvironmentBinding{
		Name:  name + "_type",
		Value: runtime.StringValue{Val: typeExpressionToString(expr)},
	})
	if info, ok := parseTypeExpression(expr); ok && info.name != "" {
		values = append(values, runtime.EnvironmentBinding{
			Name:  name,
			Value: runtime.TypeRefValue{TypeName: info.name, TypeArgs: info.typeArgs},
		})
	}
	return values
}

func (i *Interpreter) lookupCallLocalTypeBindingCache(key callLocalTypeBindingCacheKey) ([]runtime.EnvironmentBinding, bool) {
	if i == nil || key.function == nil || key.receiverType == "" {
		return nil, false
	}
	if i.envSingleThread {
		values, ok := i.callLocalTypeBindingCache[key]
		return values, ok
	}
	i.callLocalTypeBindingCacheMu.RLock()
	defer i.callLocalTypeBindingCacheMu.RUnlock()
	values, ok := i.callLocalTypeBindingCache[key]
	return values, ok
}

func (i *Interpreter) storeCallLocalTypeBindingCache(key callLocalTypeBindingCacheKey, values []runtime.EnvironmentBinding) {
	if i == nil || key.function == nil || key.receiverType == "" || len(values) == 0 {
		return
	}
	if i.envSingleThread {
		if i.callLocalTypeBindingCache == nil {
			i.callLocalTypeBindingCache = make(map[callLocalTypeBindingCacheKey][]runtime.EnvironmentBinding)
		}
		i.callLocalTypeBindingCache[key] = values
		return
	}
	i.callLocalTypeBindingCacheMu.Lock()
	defer i.callLocalTypeBindingCacheMu.Unlock()
	if i.callLocalTypeBindingCache == nil {
		i.callLocalTypeBindingCache = make(map[callLocalTypeBindingCacheKey][]runtime.EnvironmentBinding)
	}
	i.callLocalTypeBindingCache[key] = values
}

func promoteHigherKindedSelfBindings(pattern ast.TypeExpression, actual ast.TypeExpression, bindings map[string]ast.TypeExpression) {
	if pattern == nil || actual == nil || len(bindings) == 0 {
		return
	}
	switch p := pattern.(type) {
	case *ast.GenericTypeExpression:
		switch a := actual.(type) {
		case *ast.GenericTypeExpression:
			if base, ok := p.Base.(*ast.SimpleTypeExpression); ok && base != nil && base.Name != nil {
				name := base.Name.Name
				if bound, ok := bindings[name].(*ast.SimpleTypeExpression); ok && bound != nil && bound.Name != nil && len(a.Arguments) > 0 {
					wildcards := make([]ast.TypeExpression, len(a.Arguments))
					for idx := range wildcards {
						wildcards[idx] = ast.NewWildcardTypeExpression()
					}
					bindings[name] = ast.Gen(ast.Ty(bound.Name.Name), wildcards...)
				}
			}
			promoteHigherKindedSelfBindings(p.Base, a.Base, bindings)
			limit := len(p.Arguments)
			if len(a.Arguments) < limit {
				limit = len(a.Arguments)
			}
			for idx := 0; idx < limit; idx++ {
				promoteHigherKindedSelfBindings(p.Arguments[idx], a.Arguments[idx], bindings)
			}
		case *ast.SimpleTypeExpression:
			promoteHigherKindedSelfBindings(p.Base, a, bindings)
		}
	case *ast.NullableTypeExpression:
		if a, ok := actual.(*ast.NullableTypeExpression); ok && p != nil {
			promoteHigherKindedSelfBindings(p.InnerType, a.InnerType, bindings)
		}
	case *ast.ResultTypeExpression:
		if a, ok := actual.(*ast.ResultTypeExpression); ok && p != nil {
			promoteHigherKindedSelfBindings(p.InnerType, a.InnerType, bindings)
		}
	case *ast.UnionTypeExpression:
		a, ok := actual.(*ast.UnionTypeExpression)
		if !ok || p == nil {
			return
		}
		limit := len(p.Members)
		if len(a.Members) < limit {
			limit = len(a.Members)
		}
		for idx := 0; idx < limit; idx++ {
			promoteHigherKindedSelfBindings(p.Members[idx], a.Members[idx], bindings)
		}
	}
}
