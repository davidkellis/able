package interpreter

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type implMethodContext struct {
	implName      string
	interfaceName string
	target        ast.TypeExpression
	methods       map[string]runtime.Value
}

func (i *Interpreter) invalidateMethodCache() {
	i.methodCacheMu.Lock()
	defer i.methodCacheMu.Unlock()
	i.methodCacheVersion++
	if len(i.methodCache) > 0 {
		i.methodCache = make(map[methodCacheKey]methodCacheEntry)
	}
	if len(i.boundMethodCache) > 0 {
		i.boundMethodCache = make(map[boundMethodCacheKey]runtime.Value)
	}
}

func (i *Interpreter) currentMethodCacheVersion() uint64 {
	if i == nil {
		return 0
	}
	if i.envSingleThread {
		return i.methodCacheVersion
	}
	i.methodCacheMu.RLock()
	defer i.methodCacheMu.RUnlock()
	return i.methodCacheVersion
}

type methodCacheKey struct {
	typeName    string
	methodName  string
	ifaceFilter string
}

type methodCacheEntry struct {
	method runtime.Value
	err    error
}

type boundMethodCacheKey struct {
	receiver      any
	methodName    string
	ifaceFilter   string
	allowInherent bool
}

const boundMethodCacheMaxEntries = 2048

func boundMethodReceiverKey(receiver runtime.Value) (any, bool) {
	switch r := receiver.(type) {
	case *runtime.ArrayValue:
		if r == nil {
			return nil, false
		}
		return r, true
	case *runtime.StructInstanceValue:
		if r == nil {
			return nil, false
		}
		return r, true
	case *runtime.InterfaceValue:
		if r == nil {
			return nil, false
		}
		return r, true
	case *runtime.StringValue:
		if r == nil {
			return nil, false
		}
		return r, true
	case *runtime.IteratorValue:
		if r == nil {
			return nil, false
		}
		return r, true
	case *runtime.FutureValue:
		if r == nil {
			return nil, false
		}
		return r, true
	case *runtime.HasherValue:
		if r == nil {
			return nil, false
		}
		return r, true
	case *runtime.ErrorValue:
		if r == nil {
			return nil, false
		}
		return r, true
	default:
		return nil, false
	}
}

func (i *Interpreter) lookupBoundMethodCache(key boundMethodCacheKey) (runtime.Value, bool) {
	if i == nil {
		return nil, false
	}
	if i.envSingleThread {
		method, ok := i.boundMethodCache[key]
		return method, ok
	}
	i.methodCacheMu.RLock()
	defer i.methodCacheMu.RUnlock()
	method, ok := i.boundMethodCache[key]
	return method, ok
}

func (i *Interpreter) storeBoundMethodCache(key boundMethodCacheKey, method runtime.Value) {
	if i == nil {
		return
	}
	if method == nil {
		return
	}
	if i.envSingleThread {
		if i.boundMethodCache == nil {
			i.boundMethodCache = make(map[boundMethodCacheKey]runtime.Value)
		}
		if len(i.boundMethodCache) >= boundMethodCacheMaxEntries {
			i.boundMethodCache = make(map[boundMethodCacheKey]runtime.Value, boundMethodCacheMaxEntries/2)
		}
		i.boundMethodCache[key] = method
		return
	}
	i.methodCacheMu.Lock()
	defer i.methodCacheMu.Unlock()
	if i.boundMethodCache == nil {
		i.boundMethodCache = make(map[boundMethodCacheKey]runtime.Value)
	}
	if len(i.boundMethodCache) >= boundMethodCacheMaxEntries {
		i.boundMethodCache = make(map[boundMethodCacheKey]runtime.Value, boundMethodCacheMaxEntries/2)
	}
	i.boundMethodCache[key] = method
}

type methodResolutionAccumulator struct {
	singleFunction     *runtime.FunctionValue
	functionCandidates []*runtime.FunctionValue
	singleNative       runtime.Value
	singleNativeKey    string
	nativeCandidates   []runtime.Value
	nativeKeys         []string
}

func (a *methodResolutionAccumulator) count() int {
	if a == nil {
		return 0
	}
	return a.functionCount() + a.nativeCount()
}

func (a *methodResolutionAccumulator) functionCount() int {
	if a == nil {
		return 0
	}
	if len(a.functionCandidates) > 0 {
		return len(a.functionCandidates)
	}
	if a.singleFunction != nil {
		return 1
	}
	return 0
}

func (a *methodResolutionAccumulator) nativeCount() int {
	if a == nil {
		return 0
	}
	if len(a.nativeCandidates) > 0 {
		return len(a.nativeCandidates)
	}
	if a.singleNative != nil {
		return 1
	}
	return 0
}

func (a *methodResolutionAccumulator) hasFunction(fn *runtime.FunctionValue) bool {
	if fn == nil {
		return false
	}
	if a.singleFunction == fn {
		return true
	}
	for _, existing := range a.functionCandidates {
		if existing == fn {
			return true
		}
	}
	return false
}

func (a *methodResolutionAccumulator) hasNativeKey(key string) bool {
	if key == "" {
		return false
	}
	if a.singleNativeKey == key {
		return true
	}
	for _, existing := range a.nativeKeys {
		if existing == key {
			return true
		}
	}
	return false
}

func (a *methodResolutionAccumulator) addNativeCandidate(key string, candidate runtime.Value) {
	if key != "" {
		if a.hasNativeKey(key) {
			return
		}
		if len(a.nativeCandidates) > 0 {
			a.nativeKeys = append(a.nativeKeys, key)
		}
	}
	if len(a.nativeCandidates) > 0 {
		a.nativeCandidates = append(a.nativeCandidates, candidate)
		return
	}
	if a.singleNative == nil {
		a.singleNative = candidate
		a.singleNativeKey = key
		return
	}
	a.nativeCandidates = append(a.nativeCandidates, a.singleNative, candidate)
	if a.singleNativeKey != "" {
		a.nativeKeys = append(a.nativeKeys, a.singleNativeKey)
	}
	if key != "" {
		a.nativeKeys = append(a.nativeKeys, key)
	}
	a.singleNative = nil
	a.singleNativeKey = ""
}

func (a *methodResolutionAccumulator) addFunctionCandidate(fn *runtime.FunctionValue) {
	if fn == nil {
		return
	}
	if len(a.functionCandidates) > 0 {
		if a.hasFunction(fn) {
			return
		}
		a.functionCandidates = append(a.functionCandidates, fn)
		return
	}
	if a.singleFunction == nil {
		a.singleFunction = fn
		return
	}
	if a.singleFunction == fn {
		return
	}
	a.functionCandidates = append(a.functionCandidates, a.singleFunction, fn)
	a.singleFunction = nil
}

func (a *methodResolutionAccumulator) addCallable(funcName string, receiver runtime.Value, method runtime.Value, privacyContext string) error {
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
		a.addFunctionCandidate(fn)
	case *runtime.FunctionOverloadValue:
		if fn == nil {
			return nil
		}
		for _, entry := range fn.Overloads {
			if entry == nil {
				continue
			}
			if err := a.addCallable(funcName, receiver, entry, privacyContext); err != nil {
				return err
			}
		}
	case runtime.NativeFunctionValue:
		key := fn.Name
		if key == "" {
			key = fmt.Sprintf("%p", &fn)
		}
		a.addNativeCandidate(key, runtime.NativeBoundMethodValue{Receiver: receiver, Method: fn})
	case *runtime.NativeFunctionValue:
		if fn == nil {
			return nil
		}
		key := fn.Name
		if key == "" {
			key = fmt.Sprintf("%p", fn)
		}
		a.addNativeCandidate(key, runtime.NativeBoundMethodValue{Receiver: receiver, Method: *fn})
	case runtime.NativeBoundMethodValue:
		key := fn.Method.Name
		if key == "" {
			key = fmt.Sprintf("%p", &fn)
		}
		a.addNativeCandidate(key, fn)
	case *runtime.NativeBoundMethodValue:
		if fn == nil {
			return nil
		}
		key := fn.Method.Name
		if key == "" {
			key = fmt.Sprintf("%p", fn)
		}
		a.addNativeCandidate(key, *fn)
	}
	return nil
}

func (i *Interpreter) resolveMethodFromPool(env *runtime.Environment, funcName string, receiver runtime.Value, ifaceFilter string) (runtime.Value, error) {
	acc := methodResolutionAccumulator{}
	cacheReceiver, cacheableReceiver := boundMethodReceiverKey(receiver)
	var implCtx *implMethodContext
	if env != nil {
		if data := env.RuntimeData(); data != nil {
			if ctx, ok := data.(*implMethodContext); ok && ctx != nil && ctx.target != nil && receiver != nil {
				implCtx = ctx
			}
		}
	}
	hasImplMethodContext := implCtx != nil
	if cacheableReceiver && !hasImplMethodContext && isPrimitiveReceiver(receiver) {
		earlyCacheKey := boundMethodCacheKey{
			receiver:      cacheReceiver,
			methodName:    funcName,
			ifaceFilter:   ifaceFilter,
			allowInherent: true,
		}
		if cached, ok := i.lookupBoundMethodCache(earlyCacheKey); ok {
			return cached, nil
		}
	}
	var scopeCallable runtime.Value
	var scopeFilter functionScopeFilter
	nameInScope := false
	if env != nil {
		if val, ok := env.Lookup(funcName); ok && isCallableRuntimeValue(val) {
			nameInScope = true
			scopeCallable = val
			scopeFilter = functionScopeFilterFromValue(val)
		}
	}
	var info typeInfo
	var hasInfo bool
	if receiver != nil {
		info, hasInfo = i.getTypeInfoForValue(receiver)
	}
	typeNameInScope := false
	if hasInfo && env != nil {
		for _, name := range i.canonicalTypeNames(info.name) {
			if env.Has(name) {
				typeNameInScope = true
				break
			}
		}
	}
	allowInherent := nameInScope || typeNameInScope || isPrimitiveReceiver(receiver)
	cacheKey := boundMethodCacheKey{
		receiver:      cacheReceiver,
		methodName:    funcName,
		ifaceFilter:   ifaceFilter,
		allowInherent: allowInherent,
	}

	if implCtx != nil {
		if ifaceFilter == "" || ifaceFilter == implCtx.interfaceName {
			if i.matchesType(implCtx.target, receiver) {
				if method := implCtx.methods[funcName]; method != nil {
					if callable, ok := i.selectUfcsCallable(method, receiver, true, functionScopeFilter{}); ok {
						if err := checkPrivateMethod(funcName, implCtx.implName, callable); err != nil {
							return nil, err
						}
						return runtime.BoundMethodValue{Receiver: receiver, Method: callable}, nil
					}
				}
			}
		}
	}
	if cacheableReceiver && !hasImplMethodContext {
		if cached, ok := i.lookupBoundMethodCache(cacheKey); ok {
			return cached, nil
		}
	}

	if hasInfo {
		if allowInherent {
			for _, name := range i.canonicalTypeNames(info.name) {
				if bucket, ok := i.inherentMethods[name]; ok {
					if method := bucket[funcName]; method != nil {
						if callable, ok := i.selectUfcsCallable(method, receiver, true, functionScopeFilter{}); ok {
							if err := acc.addCallable(funcName, receiver, callable, name); err != nil {
								return nil, err
							}
						}
					}
				}
			}
		}
		existing := acc.count()
		if method, err := i.findMethodCached(info, funcName, ifaceFilter); err == nil && method != nil {
			if callable, ok := i.selectUfcsCallable(method, receiver, true, functionScopeFilter{}); ok {
				if err := acc.addCallable(funcName, receiver, callable, info.name); err != nil {
					return nil, err
				}
			}
		} else if err != nil {
			if existing == 0 {
				return nil, err
			}
		}
		if acc.count() == 0 && i.compiledInstanceMethodFn != nil {
			if method, found := i.compiledInstanceMethodFn(info.name, funcName); found && method != nil {
				if err := acc.addCallable(funcName, receiver, method, info.name); err != nil {
					return nil, err
				}
			}
		}
		if acc.count() == 0 && i.interfaceMethodResolver != nil {
			if ifaceFilter != "" {
				if resolved, found := i.interfaceMethodResolver(receiver, ifaceFilter, funcName); found && resolved != nil {
					if err := acc.addCallable(funcName, receiver, resolved, info.name); err != nil {
						return nil, err
					}
				}
			}
		}
		if acc.count() == 0 && i.compiledInterfaceMemberFn != nil {
			if resolved, found := i.compiledInterfaceMemberFn(receiver, funcName); found && resolved != nil {
				if err := acc.addCallable(funcName, receiver, resolved, info.name); err != nil {
					return nil, err
				}
			}
		}
	}

	hasMethodCandidate := acc.count() > 0

	if env != nil && nameInScope && !hasMethodCandidate {
		if scopeCallable != nil {
			if callable, ok := i.selectUfcsCallable(scopeCallable, receiver, false, scopeFilter); ok {
				if err := acc.addCallable(funcName, receiver, callable, ""); err != nil {
					return nil, err
				}
			}
		}
	}

	if acc.functionCount() > 0 && acc.nativeCount() > 0 {
		return nil, fmt.Errorf("Ambiguous overload for %s", funcName)
	}
	if acc.functionCount() > 0 {
		var callable runtime.Value
		switch {
		case len(acc.functionCandidates) > 1:
			callable = &runtime.FunctionOverloadValue{Overloads: acc.functionCandidates}
		case len(acc.functionCandidates) == 1:
			callable = acc.functionCandidates[0]
		default:
			callable = acc.singleFunction
		}
		if callable == nil {
			return nil, nil
		} else {
			bound := runtime.BoundMethodValue{Receiver: receiver, Method: callable}
			if cacheableReceiver && !hasImplMethodContext {
				i.storeBoundMethodCache(cacheKey, bound)
			}
			return bound, nil
		}
	}
	if acc.nativeCount() > 0 {
		if acc.nativeCount() > 1 {
			return nil, fmt.Errorf("Ambiguous overload for %s", funcName)
		}
		if len(acc.nativeCandidates) == 1 {
			return acc.nativeCandidates[0], nil
		}
		return acc.singleNative, nil
	}
	return nil, nil
}

type functionScopeFilter struct {
	enabled   bool
	single    *runtime.FunctionValue
	overloads []*runtime.FunctionValue
}

func functionScopeFilterFromValue(scope runtime.Value) functionScopeFilter {
	switch v := scope.(type) {
	case *runtime.FunctionValue:
		if v != nil {
			return functionScopeFilter{enabled: true, single: v}
		}
	case *runtime.FunctionOverloadValue:
		if v != nil && len(v.Overloads) > 0 {
			return functionScopeFilter{enabled: true, overloads: v.Overloads}
		}
	}
	return functionScopeFilter{}
}

func (f functionScopeFilter) contains(fn *runtime.FunctionValue) bool {
	if !f.enabled {
		return true
	}
	if fn == nil {
		return false
	}
	if f.single != nil {
		return f.single == fn
	}
	for _, candidate := range f.overloads {
		if candidate == fn {
			return true
		}
	}
	return false
}

func (i *Interpreter) methodTargetMatchesReceiver(fn *runtime.FunctionValue, receiver runtime.Value) bool {
	if fn == nil || fn.MethodSet == nil || fn.MethodSet.TargetType == nil {
		return true
	}
	inst, ok := receiver.(*runtime.StructInstanceValue)
	if !ok || inst == nil || inst.Definition == nil {
		return true
	}
	target := fn.MethodSet.TargetType
	var targetName string
	switch t := target.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name != nil {
			targetName = t.Name.Name
		}
	case *ast.GenericTypeExpression:
		if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base.Name != nil {
			targetName = base.Name.Name
		}
	}
	if targetName == "" || fn.Closure == nil {
		return true
	}
	def, ok := fn.Closure.StructDefinition(targetName)
	if !ok || def == nil {
		return true
	}
	return def == inst.Definition
}

func (i *Interpreter) selectUfcsCallable(method runtime.Value, receiver runtime.Value, requireSelf bool, scopeFilter functionScopeFilter) (runtime.Value, bool) {
	switch fn := method.(type) {
	case *runtime.FunctionValue:
		if !scopeFilter.contains(fn) {
			return nil, false
		}
		if (!requireSelf || functionExpectsSelf(fn)) && i.functionFirstParamMatches(fn, receiver) {
			if !requireSelf && fn.TypeQualified {
				return nil, false
			}
			if !i.methodTargetMatchesReceiver(fn, receiver) {
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
			if !scopeFilter.contains(entry) {
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
			if !i.methodTargetMatchesReceiver(entry, receiver) {
				continue
			}
			filtered = append(filtered, entry)
		}
		if len(filtered) > 1 {
			targetFiltered := make([]*runtime.FunctionValue, 0, len(filtered))
			for _, entry := range filtered {
				if i.methodTargetMatchesReceiver(entry, receiver) {
					targetFiltered = append(targetFiltered, entry)
				}
			}
			if len(targetFiltered) == 1 {
				return targetFiltered[0], true
			}
			if len(targetFiltered) > 0 {
				filtered = targetFiltered
			}
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
	case runtime.NativeFunctionValue:
		if requireSelf {
			return fn, true
		}
		if fn.Arity <= 0 {
			return nil, false
		}
		adjusted := fn
		adjusted.Arity--
		return runtime.NativeBoundMethodValue{Receiver: receiver, Method: adjusted}, true
	case *runtime.NativeFunctionValue:
		if fn == nil {
			return nil, false
		}
		if requireSelf {
			return fn, true
		}
		if fn.Arity <= 0 {
			return nil, false
		}
		adjusted := *fn
		adjusted.Arity--
		return runtime.NativeBoundMethodValue{Receiver: receiver, Method: adjusted}, true
	default:
		return nil, false
	}
}

func checkPrivateMethod(funcName, context string, callable runtime.Value) error {
	name := context
	if name == "" {
		name = "<unknown>"
	}
	switch fn := callable.(type) {
	case *runtime.FunctionValue:
		if fn == nil {
			return nil
		}
		if def, ok := fn.Declaration.(*ast.FunctionDefinition); ok && def != nil && def.IsPrivate {
			return fmt.Errorf("Method '%s' on %s is private", funcName, name)
		}
	case *runtime.FunctionOverloadValue:
		if fn == nil {
			return nil
		}
		for _, entry := range fn.Overloads {
			if entry == nil {
				continue
			}
			if def, ok := entry.Declaration.(*ast.FunctionDefinition); ok && def != nil && def.IsPrivate {
				return fmt.Errorf("Method '%s' on %s is private", funcName, name)
			}
		}
	}
	return nil
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
