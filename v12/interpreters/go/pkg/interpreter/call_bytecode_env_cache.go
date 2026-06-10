package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
	"strings"
)

type reusableBytecodeCallEnvCacheKey struct {
	function            *runtime.FunctionValue
	call                *ast.FunctionCall
	receiverType        string
	explicitBindingsKey string
}

type functionCallTypeBindingSet struct {
	explicit     []runtime.EnvironmentBinding
	callLocal    []runtime.EnvironmentBinding
	receiverType string
}

func reusableBytecodeCallEnvBindingKey(bindings []runtime.EnvironmentBinding) (string, bool) {
	if len(bindings) == 0 {
		return "", true
	}
	var b strings.Builder
	for _, binding := range bindings {
		if binding.Name == "" {
			return "", false
		}
		b.WriteString(binding.Name)
		b.WriteByte('=')
		switch value := binding.Value.(type) {
		case runtime.StringValue:
			b.WriteString("string:")
			b.WriteString(value.Val)
		case *runtime.StringValue:
			if value == nil {
				return "", false
			}
			b.WriteString("string:")
			b.WriteString(value.Val)
		case runtime.TypeRefValue:
			b.WriteString("typeref:")
			b.WriteString(value.TypeName)
			for _, arg := range value.TypeArgs {
				b.WriteByte('<')
				b.WriteString(typeExpressionToString(arg))
				b.WriteByte('>')
			}
		case *runtime.TypeRefValue:
			if value == nil {
				return "", false
			}
			b.WriteString("typeref:")
			b.WriteString(value.TypeName)
			for _, arg := range value.TypeArgs {
				b.WriteByte('<')
				b.WriteString(typeExpressionToString(arg))
				b.WriteByte('>')
			}
		default:
			return "", false
		}
		b.WriteByte(0)
	}
	return b.String(), true
}

func reusableBytecodeCallEnvCacheKeyForResolvedBindings(fn *runtime.FunctionValue, call *ast.FunctionCall, bindings functionCallTypeBindingSet) (reusableBytecodeCallEnvCacheKey, bool) {
	if fn == nil || bindings.empty() {
		return reusableBytecodeCallEnvCacheKey{}, false
	}
	key := reusableBytecodeCallEnvCacheKey{
		function:     fn,
		receiverType: bindings.receiverType,
	}
	// Call-local runtime bindings are already memoized per function/receiver
	// type, so equivalent receiver types can share one seeded env across sites.
	if len(bindings.explicit) == 0 {
		return key, key.receiverType != ""
	}
	if explicitKey, ok := reusableBytecodeCallEnvBindingKey(bindings.explicit); ok && explicitKey != "" {
		key.explicitBindingsKey = explicitKey
		return key, true
	}
	if call == nil {
		return reusableBytecodeCallEnvCacheKey{}, false
	}
	key.call = call
	return key, true
}

func (i *Interpreter) lookupReusableBytecodeCallEnv(key reusableBytecodeCallEnvCacheKey) (*runtime.Environment, bool) {
	if i == nil || key.function == nil || (key.call == nil && key.receiverType == "" && key.explicitBindingsKey == "") {
		return nil, false
	}
	if i.envSingleThread {
		env, ok := i.reusableBytecodeCallEnvCache[key]
		return env, ok && env != nil
	}
	i.reusableBytecodeCallEnvCacheMu.RLock()
	defer i.reusableBytecodeCallEnvCacheMu.RUnlock()
	env, ok := i.reusableBytecodeCallEnvCache[key]
	return env, ok && env != nil
}

func (i *Interpreter) storeReusableBytecodeCallEnv(key reusableBytecodeCallEnvCacheKey, env *runtime.Environment) {
	if i == nil || key.function == nil || (key.call == nil && key.receiverType == "" && key.explicitBindingsKey == "") || env == nil {
		return
	}
	if i.envSingleThread {
		if i.reusableBytecodeCallEnvCache == nil {
			i.reusableBytecodeCallEnvCache = make(map[reusableBytecodeCallEnvCacheKey]*runtime.Environment)
		}
		i.reusableBytecodeCallEnvCache[key] = env
		return
	}
	i.reusableBytecodeCallEnvCacheMu.Lock()
	defer i.reusableBytecodeCallEnvCacheMu.Unlock()
	if i.reusableBytecodeCallEnvCache == nil {
		i.reusableBytecodeCallEnvCache = make(map[reusableBytecodeCallEnvCacheKey]*runtime.Environment)
	}
	i.reusableBytecodeCallEnvCache[key] = env
}

func (bindings functionCallTypeBindingSet) empty() bool {
	return len(bindings.explicit) == 0 && len(bindings.callLocal) == 0
}

func (bindings functionCallTypeBindingSet) totalLen() int {
	return len(bindings.explicit) + len(bindings.callLocal)
}

func (i *Interpreter) functionCallTypeBindingSet(fn *runtime.FunctionValue, decl *ast.FunctionDefinition, call *ast.FunctionCall, receiver runtime.Value, needsCallLocalTypeBindings bool) functionCallTypeBindingSet {
	return i.functionCallTypeBindingSetWithPlanAndEnv(fn, decl, call, receiver, needsCallLocalTypeBindings, i.functionRuntimeGenericBindingPlan(fn), nil)
}

func (i *Interpreter) functionCallTypeBindingSetWithPlan(fn *runtime.FunctionValue, decl *ast.FunctionDefinition, call *ast.FunctionCall, receiver runtime.Value, needsCallLocalTypeBindings bool, plan *functionRuntimeGenericBindingPlan) functionCallTypeBindingSet {
	return i.functionCallTypeBindingSetWithPlanAndEnv(fn, decl, call, receiver, needsCallLocalTypeBindings, plan, nil)
}

func (i *Interpreter) functionCallTypeBindingSetWithPlanAndEnv(fn *runtime.FunctionValue, decl *ast.FunctionDefinition, call *ast.FunctionCall, receiver runtime.Value, needsCallLocalTypeBindings bool, plan *functionRuntimeGenericBindingPlan, env *runtime.Environment) functionCallTypeBindingSet {
	bindings := functionCallTypeBindingSet{
		explicit: nil,
	}
	if plan == nil || plan.explicitUsed {
		bindings.explicit = i.explicitCallTypeBindingValuesIfAny(decl, call)
	}
	if !needsCallLocalTypeBindings || receiver == nil {
		return bindings
	}
	receiverTypeHint := ast.TypeExpression(nil)
	if _, isGenericUnion := i.genericUnionMethodTarget(fn); isGenericUnion {
		receiverTypeHint = i.staticReceiverTypeForCall(call, env)
	}
	bindings.callLocal, bindings.receiverType = i.callLocalTypeBindingValuesAndStaticReceiverTypeIfAny(fn, receiver, receiverTypeHint)
	return bindings
}

func bytecodeCallEnvMutatesBindingSet(decl *ast.FunctionDefinition, bindings functionCallTypeBindingSet) bool {
	if decl == nil || decl.Body == nil || bindings.empty() {
		return false
	}
	seen := make(map[string]struct{}, bindings.totalLen())
	for _, bindingSet := range [][]runtime.EnvironmentBinding{bindings.explicit, bindings.callLocal} {
		for _, binding := range bindingSet {
			if binding.Name == "" {
				continue
			}
			if _, ok := seen[binding.Name]; ok {
				continue
			}
			seen[binding.Name] = struct{}{}
			if blockMutatesIdentifier(decl.Body, binding.Name) {
				return true
			}
		}
	}
	return false
}

func (i *Interpreter) reusableBytecodeCallEnvForResolvedBindings(fn *runtime.FunctionValue, decl *ast.FunctionDefinition, call *ast.FunctionCall, slotProgram *bytecodeProgram, bindings functionCallTypeBindingSet) (*runtime.Environment, bool) {
	if i == nil || fn == nil || decl == nil || fn.Closure == nil || slotProgram == nil || slotProgram.frameLayout == nil {
		return nil, false
	}
	if slotProgram.frameLayout.needsEnvScopes {
		return nil, false
	}
	if bindings.empty() {
		return nil, false
	}
	if len(bindings.callLocal) > 0 && bindings.receiverType == "" {
		return nil, false
	}
	key, ok := reusableBytecodeCallEnvCacheKeyForResolvedBindings(fn, call, bindings)
	if !ok {
		return nil, false
	}
	if cached, ok := i.lookupReusableBytecodeCallEnv(key); ok {
		return cached, true
	}
	if bytecodeCallEnvMutatesBindingSet(decl, bindings) {
		return nil, false
	}
	env := runtime.NewEnvironmentWithBindingSets(fn.Closure, bindings.distinctLen(), bindings.explicit, bindings.callLocal)
	i.storeReusableBytecodeCallEnv(key, env)
	return env, true
}

func (i *Interpreter) reusableBytecodeCallEnvForBindings(fn *runtime.FunctionValue, decl *ast.FunctionDefinition, call *ast.FunctionCall, receiver runtime.Value, needsCallLocalTypeBindings bool, slotProgram *bytecodeProgram) (*runtime.Environment, bool) {
	bindings := i.functionCallTypeBindingSet(fn, decl, call, receiver, needsCallLocalTypeBindings)
	return i.reusableBytecodeCallEnvForResolvedBindings(fn, decl, call, slotProgram, bindings)
}

func (i *Interpreter) reusableBytecodeCallEnvForExplicitBindings(fn *runtime.FunctionValue, decl *ast.FunctionDefinition, call *ast.FunctionCall, slotProgram *bytecodeProgram) (*runtime.Environment, bool) {
	return i.reusableBytecodeCallEnvForResolvedBindings(fn, decl, call, slotProgram, functionCallTypeBindingSet{
		explicit: i.explicitCallTypeBindingValuesIfAny(decl, call),
	})
}
