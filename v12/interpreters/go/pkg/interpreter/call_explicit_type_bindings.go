package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type explicitCallTypeBindingCacheKey struct {
	function ast.Node
	call     *ast.FunctionCall
	version  uint64
}

func buildExplicitCallTypeBindingRuntimeValues(generics []*ast.GenericParameter, typeArgs []ast.TypeExpression) []runtime.EnvironmentBinding {
	if len(generics) == 0 || len(typeArgs) == 0 {
		return nil
	}
	count := len(generics)
	if len(typeArgs) < count {
		count = len(typeArgs)
	}
	values := make([]runtime.EnvironmentBinding, 0, count*2)
	for idx := 0; idx < count; idx++ {
		gp := generics[idx]
		if gp == nil || gp.Name == nil {
			continue
		}
		ta := typeArgs[idx]
		if ta == nil {
			continue
		}
		values = appendCallLocalTypeBindingRuntimeValues(values, gp.Name.Name, ta)
	}
	if len(values) == 0 {
		return nil
	}
	return values
}

func (i *Interpreter) lookupExplicitCallTypeBindingCache(key explicitCallTypeBindingCacheKey) ([]runtime.EnvironmentBinding, bool) {
	if i == nil || key.function == nil || key.call == nil {
		return nil, false
	}
	if i.envSingleThread {
		values, ok := i.explicitCallTypeBindingCache[key]
		return values, ok
	}
	i.explicitCallTypeBindingCacheMu.RLock()
	defer i.explicitCallTypeBindingCacheMu.RUnlock()
	values, ok := i.explicitCallTypeBindingCache[key]
	return values, ok
}

func (i *Interpreter) storeExplicitCallTypeBindingCache(key explicitCallTypeBindingCacheKey, values []runtime.EnvironmentBinding) {
	if i == nil || key.function == nil || key.call == nil || len(values) == 0 {
		return
	}
	if i.envSingleThread {
		if i.explicitCallTypeBindingCache == nil {
			i.explicitCallTypeBindingCache = make(map[explicitCallTypeBindingCacheKey][]runtime.EnvironmentBinding)
		}
		i.explicitCallTypeBindingCache[key] = values
		return
	}
	i.explicitCallTypeBindingCacheMu.Lock()
	defer i.explicitCallTypeBindingCacheMu.Unlock()
	if i.explicitCallTypeBindingCache == nil {
		i.explicitCallTypeBindingCache = make(map[explicitCallTypeBindingCacheKey][]runtime.EnvironmentBinding)
	}
	i.explicitCallTypeBindingCache[key] = values
}

func (i *Interpreter) explicitCallTypeBindingValuesIfAny(funcNode ast.Node, call *ast.FunctionCall) []runtime.EnvironmentBinding {
	if funcNode == nil || call == nil {
		return nil
	}
	if i != nil && !i.callableNeedsExplicitRuntimeTypeBindings(funcNode) {
		return nil
	}
	if i == nil && !callableNeedsExplicitRuntimeTypeBindings(funcNode) {
		return nil
	}
	cacheKey := explicitCallTypeBindingCacheKey{
		function: funcNode,
		call:     call,
		version:  i.inferredCallTypeArgumentVersion(call),
	}
	if cached, ok := i.lookupExplicitCallTypeBindingCache(cacheKey); ok {
		return cached
	}
	generics, _ := extractFunctionGenerics(funcNode)
	if len(generics) == 0 {
		return nil
	}
	runtimeBindings := buildExplicitCallTypeBindingRuntimeValues(generics, call.TypeArguments)
	if len(runtimeBindings) == 0 {
		return nil
	}
	i.storeExplicitCallTypeBindingCache(cacheKey, runtimeBindings)
	return runtimeBindings
}

func bindSimpleIdentifierPatternIntoEnv(env *runtime.Environment, pattern ast.Pattern, value runtime.Value) bool {
	if env == nil || pattern == nil {
		return false
	}
	switch p := pattern.(type) {
	case *ast.Identifier:
		if p == nil || p.Name == "" || p.Name == "_" {
			return true
		}
		env.DefineWithoutMerge(p.Name, value)
		return true
	case *ast.WildcardPattern:
		return true
	default:
		return false
	}
}
