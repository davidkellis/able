package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type methodSetConstraintResultCacheKey struct {
	methodSet    *runtime.MethodSet
	receiverType string
}

type methodSetConstraintResultCacheEntry struct {
	err error
}

func resolveMethodSetReceiver(def *ast.FunctionDefinition, args []runtime.Value) (runtime.Value, bool) {
	if def == nil || !functionDefinitionExpectsSelf(def) {
		return nil, false
	}
	if len(args) == 0 {
		return nil, false
	}
	return args[0], true
}

func (i *Interpreter) lookupMethodSetConstraintResultCache(key methodSetConstraintResultCacheKey) (methodSetConstraintResultCacheEntry, bool) {
	if i == nil || key.methodSet == nil || key.receiverType == "" {
		return methodSetConstraintResultCacheEntry{}, false
	}
	if i.envSingleThread {
		entry, ok := i.methodSetConstraintResultCache[key]
		return entry, ok
	}
	i.methodCacheMu.RLock()
	defer i.methodCacheMu.RUnlock()
	entry, ok := i.methodSetConstraintResultCache[key]
	return entry, ok
}

func (i *Interpreter) storeMethodSetConstraintResultCache(key methodSetConstraintResultCacheKey, err error) {
	if i == nil || key.methodSet == nil || key.receiverType == "" {
		return
	}
	if i.envSingleThread {
		if i.methodSetConstraintResultCache == nil {
			i.methodSetConstraintResultCache = make(map[methodSetConstraintResultCacheKey]methodSetConstraintResultCacheEntry)
		}
		i.methodSetConstraintResultCache[key] = methodSetConstraintResultCacheEntry{err: err}
		return
	}
	i.methodCacheMu.Lock()
	defer i.methodCacheMu.Unlock()
	if i.methodSetConstraintResultCache == nil {
		i.methodSetConstraintResultCache = make(map[methodSetConstraintResultCacheKey]methodSetConstraintResultCacheEntry)
	}
	i.methodSetConstraintResultCache[key] = methodSetConstraintResultCacheEntry{err: err}
}

func (i *Interpreter) enforceMethodSetConstraints(fn *runtime.FunctionValue, receiver runtime.Value) error {
	if fn == nil || fn.MethodSet == nil {
		return nil
	}
	plan := i.methodSetConstraintPlan(fn.MethodSet)
	if plan == nil || len(plan.constraints) == 0 {
		return nil
	}
	cacheKey := methodSetConstraintResultCacheKey{
		methodSet:    fn.MethodSet,
		receiverType: i.receiverTypeName(receiver),
	}
	if cached, ok := i.lookupMethodSetConstraintResultCache(cacheKey); ok {
		return cached.err
	}
	actual := i.typeExpressionFromValue(receiver)
	if actual == nil {
		return nil
	}
	bindings := make(map[string]ast.TypeExpression)
	actualExpanded := i.expandTypeAliasesCached(actual)
	if actualExpanded == nil {
		actualExpanded = actual
	}
	if fn.MethodSet.TargetType != nil && len(plan.genericNames) > 0 {
		target := i.expandTypeAliasesCached(fn.MethodSet.TargetType)
		if target == nil {
			target = fn.MethodSet.TargetType
		}
		matchTypeExpressionTemplate(target, actualExpanded, plan.genericNames, bindings)
	}
	bindings["Self"] = actualExpanded
	i.addStructTypeArgBindings(bindings, receiver)
	err := i.enforceConstraintSpecs(plan.constraints, bindings)
	i.storeMethodSetConstraintResultCache(cacheKey, err)
	return err
}

func (i *Interpreter) addStructTypeArgBindings(bindings map[string]ast.TypeExpression, receiver runtime.Value) {
	inst := structInstanceFromValue(receiver)
	if inst == nil || inst.Definition == nil || inst.Definition.Node == nil {
		return
	}
	generics := inst.Definition.Node.GenericParams
	if len(generics) == 0 {
		return
	}
	typeArgs := i.resolvedStructInstanceTypeArguments(inst)
	if len(typeArgs) != len(generics) {
		return
	}
	mapped, err := mapTypeArguments(generics, typeArgs, "method set")
	if err != nil {
		return
	}
	for name, expr := range mapped {
		if _, ok := bindings[name]; ok {
			continue
		}
		bindings[name] = expr
	}
}

func structInstanceFromValue(value runtime.Value) *runtime.StructInstanceValue {
	switch v := value.(type) {
	case *runtime.StructInstanceValue:
		return v
	case *runtime.InterfaceValue:
		if v == nil {
			return nil
		}
		return structInstanceFromValue(v.Underlying)
	case runtime.InterfaceValue:
		return structInstanceFromValue(v.Underlying)
	default:
		return nil
	}
}
