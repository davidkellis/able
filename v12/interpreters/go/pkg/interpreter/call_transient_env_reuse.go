package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func callableAllowsTransientCallEnvReuse(node ast.Node) bool {
	switch decl := node.(type) {
	case *ast.FunctionDefinition:
		return blockAllowsTransientRuntimeScope(decl.Body)
	case *ast.LambdaExpression:
		return expressionAllowsTransientRuntimeScope(decl.Body)
	default:
		return false
	}
}

func (i *Interpreter) callableAllowsTransientCallEnvReuse(node ast.Node) bool {
	if i == nil || node == nil {
		return false
	}
	if i.envSingleThread {
		if allowed, ok := i.callableTransientCallEnvReuseCache[node]; ok {
			return allowed
		}
		allowed := callableAllowsTransientCallEnvReuse(node)
		i.callableTransientCallEnvReuseCache[node] = allowed
		return allowed
	}
	i.callableTransientCallEnvReuseCacheMu.RLock()
	allowed, ok := i.callableTransientCallEnvReuseCache[node]
	i.callableTransientCallEnvReuseCacheMu.RUnlock()
	if ok {
		return allowed
	}
	allowed = callableAllowsTransientCallEnvReuse(node)
	i.callableTransientCallEnvReuseCacheMu.Lock()
	if existing, ok := i.callableTransientCallEnvReuseCache[node]; ok {
		i.callableTransientCallEnvReuseCacheMu.Unlock()
		return existing
	}
	i.callableTransientCallEnvReuseCache[node] = allowed
	i.callableTransientCallEnvReuseCacheMu.Unlock()
	return allowed
}

func (i *Interpreter) acquireTransientCallEnvForBindingSets(parent *runtime.Environment, valueCapacity int, first []runtime.EnvironmentBinding, second []runtime.EnvironmentBinding) *runtime.Environment {
	if i == nil {
		return runtime.NewEnvironmentWithBindingSets(parent, valueCapacity, first, second)
	}
	if pooled := i.transientCallEnvPool.Get(); pooled != nil {
		if env, ok := pooled.(*runtime.Environment); ok && env != nil {
			env.ResetForBindingSetsReuse(parent, valueCapacity, first, second)
			return env
		}
	}
	return runtime.NewEnvironmentWithBindingSets(parent, valueCapacity, first, second)
}

func (i *Interpreter) releaseTransientCallEnv(env *runtime.Environment) {
	if i == nil || env == nil {
		return
	}
	i.transientCallEnvPool.Put(env)
}
