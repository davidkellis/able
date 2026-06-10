package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type iteratorInterfaceMethodDictionaryCacheEntry struct {
	methods map[string]runtime.Value
	err     error
}

func (i *Interpreter) iteratorInterfaceMethodDictionary(ifaceDef *runtime.InterfaceDefinitionValue) (map[string]runtime.Value, error) {
	if i == nil {
		return nil, nil
	}
	i.methodCacheMu.RLock()
	if entry, ok := i.iteratorInterfaceMethodDictionaryCache[ifaceDef]; ok {
		i.methodCacheMu.RUnlock()
		return entry.methods, entry.err
	}
	i.methodCacheMu.RUnlock()

	methods, err := i.buildIteratorInterfaceMethodDictionary(ifaceDef)
	entry := iteratorInterfaceMethodDictionaryCacheEntry{methods: methods, err: err}

	i.methodCacheMu.Lock()
	if i.iteratorInterfaceMethodDictionaryCache == nil {
		i.iteratorInterfaceMethodDictionaryCache = make(map[*runtime.InterfaceDefinitionValue]iteratorInterfaceMethodDictionaryCacheEntry)
	}
	if existing, ok := i.iteratorInterfaceMethodDictionaryCache[ifaceDef]; ok {
		i.methodCacheMu.Unlock()
		return existing.methods, existing.err
	}
	i.iteratorInterfaceMethodDictionaryCache[ifaceDef] = entry
	i.methodCacheMu.Unlock()
	return methods, err
}

type interfaceDefaultMethodCacheKey struct {
	iface *runtime.InterfaceDefinitionValue
	name  string
}

type interfaceDefaultMethodCacheEntry struct {
	method runtime.Value
	found  bool
	err    error
}

func (i *Interpreter) interfaceDefaultMethodValue(ifaceDef *runtime.InterfaceDefinitionValue, name string) (runtime.Value, bool, error) {
	if i == nil || ifaceDef == nil || ifaceDef.Node == nil || name == "" {
		return nil, false, nil
	}
	key := interfaceDefaultMethodCacheKey{iface: ifaceDef, name: name}
	i.methodCacheMu.RLock()
	if entry, ok := i.interfaceDefaultMethodCache[key]; ok {
		i.methodCacheMu.RUnlock()
		return entry.method, entry.found, entry.err
	}
	i.methodCacheMu.RUnlock()

	method, found, err := i.buildInterfaceDefaultMethodValue(ifaceDef, name)
	entry := interfaceDefaultMethodCacheEntry{method: method, found: found, err: err}

	i.methodCacheMu.Lock()
	if i.interfaceDefaultMethodCache == nil {
		i.interfaceDefaultMethodCache = make(map[interfaceDefaultMethodCacheKey]interfaceDefaultMethodCacheEntry)
	}
	if existing, ok := i.interfaceDefaultMethodCache[key]; ok {
		i.methodCacheMu.Unlock()
		return existing.method, existing.found, existing.err
	}
	i.interfaceDefaultMethodCache[key] = entry
	i.methodCacheMu.Unlock()
	return method, found, err
}

func (i *Interpreter) buildInterfaceDefaultMethodValue(ifaceDef *runtime.InterfaceDefinitionValue, name string) (runtime.Value, bool, error) {
	for _, sig := range ifaceDef.Node.Signatures {
		if sig == nil || sig.Name == nil || sig.Name.Name != name || sig.DefaultImpl == nil {
			continue
		}
		defaultDef := ast.NewFunctionDefinition(sig.Name, sig.Params, sig.DefaultImpl, sig.ReturnType, sig.GenericParams, sig.WhereClause, false, false)
		// A dynamic interface default has both its method-level generic
		// parameters and the enclosing interface parameters. Preserve the latter
		// in its method set so bytecode return and parameter coercion keep T (or
		// an equivalent interface generic) abstract at runtime.
		methodSet := &runtime.MethodSet{
			TargetType:    ifaceDef.Node.SelfTypePattern,
			GenericParams: ifaceDef.Node.GenericParams,
			WhereClause:   ifaceDef.Node.WhereClause,
		}
		defaultVal := &runtime.FunctionValue{Declaration: defaultDef, Closure: ifaceDef.Env, MethodPriority: -1, MethodSet: methodSet}
		if program, err := i.lowerFunctionDefinitionBytecodeWithMethodSetEnv(defaultDef, ifaceDef.Env, methodSet); err != nil {
			if i.execMode == execModeBytecode {
				return nil, false, err
			}
		} else {
			setFunctionBytecodeProgram(defaultVal, program)
		}
		return defaultVal, true, nil
	}
	return nil, false, nil
}
