package interpreter

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type typeInfoCacheKey struct {
	name     string
	argCount uint8
	arg0     ast.TypeExpression
	arg1     ast.TypeExpression
	arg2     ast.TypeExpression
}

type typeExpressionSliceKey struct {
	count     uint8
	arg0      ast.TypeExpression
	arg1      ast.TypeExpression
	arg2      ast.TypeExpression
	signature string
}

type typeExpressionCacheKey struct {
	name string
	args typeExpressionSliceKey
}

func (i *Interpreter) makeInterfaceImplCacheKey(info typeInfo, interfaceName string, ifaceArgs []ast.TypeExpression) interfaceImplCacheKey {
	return i.makeInterfaceImplCacheKeyForPackage(info, interfaceName, ifaceArgs, i.currentPackage)
}

func (i *Interpreter) makeInterfaceImplCacheKeyForPackage(info typeInfo, interfaceName string, ifaceArgs []ast.TypeExpression, packageName string) interfaceImplCacheKey {
	return interfaceImplCacheKey{
		typeName:      i.cachedTypeInfoName(info),
		interfaceName: interfaceName,
		ifaceArgs:     makeTypeExpressionSliceKey(ifaceArgs),
		packageName:   packageName,
	}
}

func (i *Interpreter) makeInterfaceMethodDictionaryCacheKey(info typeInfo, interfaceName string, ifaceArgs []ast.TypeExpression) interfaceMethodDictionaryCacheKey {
	return i.makeInterfaceMethodDictionaryCacheKeyForPackage(info, interfaceName, ifaceArgs, i.currentPackage)
}

func (i *Interpreter) makeInterfaceMethodDictionaryCacheKeyForPackage(info typeInfo, interfaceName string, ifaceArgs []ast.TypeExpression, packageName string) interfaceMethodDictionaryCacheKey {
	return interfaceMethodDictionaryCacheKey{
		typeName:      i.cachedTypeInfoName(info),
		interfaceName: interfaceName,
		ifaceArgs:     makeTypeExpressionSliceKey(ifaceArgs),
		packageName:   packageName,
	}
}

func makeTypeExpressionSliceKey(args []ast.TypeExpression) typeExpressionSliceKey {
	key := typeExpressionSliceKey{count: uint8(len(args))}
	switch len(args) {
	case 0:
		return key
	case 1:
		key.arg0 = args[0]
	case 2:
		key.arg0 = args[0]
		key.arg1 = args[1]
	case 3:
		key.arg0 = args[0]
		key.arg1 = args[1]
		key.arg2 = args[2]
	default:
		key.signature = typeExpressionSliceSignature(args)
	}
	return key
}

func typeExpressionSliceSignature(args []ast.TypeExpression) string {
	if len(args) == 0 {
		return ""
	}
	if len(args) == 1 {
		return typeExpressionToString(args[0])
	}
	if len(args) == 2 {
		return typeExpressionToString(args[0]) + "|" + typeExpressionToString(args[1])
	}
	var b strings.Builder
	for idx, arg := range args {
		if idx > 0 {
			b.WriteByte('|')
		}
		appendTypeExpressionString(&b, arg)
	}
	return b.String()
}

func cloneRuntimeValueMap(src map[string]runtime.Value) map[string]runtime.Value {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]runtime.Value, len(src))
	for name, value := range src {
		dst[name] = value
	}
	return dst
}

func (i *Interpreter) lookupInterfaceImplCache(key interfaceImplCacheKey) (interfaceImplCacheEntry, bool) {
	if i == nil {
		return interfaceImplCacheEntry{}, false
	}
	if i.envSingleThread {
		entry, ok := i.interfaceImplCache[key]
		return entry, ok
	}
	i.methodCacheMu.RLock()
	defer i.methodCacheMu.RUnlock()
	entry, ok := i.interfaceImplCache[key]
	return entry, ok
}

func (i *Interpreter) storeInterfaceImplCache(key interfaceImplCacheKey, okImpl bool, err error) {
	if i == nil {
		return
	}
	entry := interfaceImplCacheEntry{ok: okImpl, err: err}
	if i.envSingleThread {
		if i.interfaceImplCache == nil {
			i.interfaceImplCache = make(map[interfaceImplCacheKey]interfaceImplCacheEntry)
		}
		i.interfaceImplCache[key] = entry
		return
	}
	i.methodCacheMu.Lock()
	defer i.methodCacheMu.Unlock()
	if i.interfaceImplCache == nil {
		i.interfaceImplCache = make(map[interfaceImplCacheKey]interfaceImplCacheEntry)
	}
	i.interfaceImplCache[key] = entry
}

func (i *Interpreter) lookupSelectedInterfaceImplCache(key interfaceImplCacheKey) (*selectedInterfaceImplCacheEntry, bool) {
	if i == nil {
		return nil, false
	}
	if i.envSingleThread {
		entry, ok := i.selectedInterfaceImplCache[key]
		return entry, ok
	}
	i.methodCacheMu.RLock()
	defer i.methodCacheMu.RUnlock()
	entry, ok := i.selectedInterfaceImplCache[key]
	return entry, ok
}

func (i *Interpreter) storeSelectedInterfaceImplCache(key interfaceImplCacheKey, candidate *implCandidate, err error) {
	if i == nil {
		return
	}
	entry := &selectedInterfaceImplCacheEntry{
		found:     candidate != nil,
		candidate: implCandidate{},
		err:       err,
	}
	if candidate != nil {
		entry.candidate = *candidate
	}
	if i.envSingleThread {
		if i.selectedInterfaceImplCache == nil {
			i.selectedInterfaceImplCache = make(map[interfaceImplCacheKey]*selectedInterfaceImplCacheEntry)
		}
		i.selectedInterfaceImplCache[key] = entry
		return
	}
	i.methodCacheMu.Lock()
	defer i.methodCacheMu.Unlock()
	if i.selectedInterfaceImplCache == nil {
		i.selectedInterfaceImplCache = make(map[interfaceImplCacheKey]*selectedInterfaceImplCacheEntry)
	}
	i.selectedInterfaceImplCache[key] = entry
}

func (i *Interpreter) lookupInterfaceMethodDictionaryCache(key interfaceMethodDictionaryCacheKey) (interfaceMethodDictionaryCacheEntry, bool) {
	if i == nil {
		return interfaceMethodDictionaryCacheEntry{}, false
	}
	if i.envSingleThread {
		entry, ok := i.interfaceMethodDictionaryCache[key]
		return entry, ok
	}
	i.methodCacheMu.RLock()
	defer i.methodCacheMu.RUnlock()
	entry, ok := i.interfaceMethodDictionaryCache[key]
	return entry, ok
}

func (i *Interpreter) storeInterfaceMethodDictionaryCache(key interfaceMethodDictionaryCacheKey, methods map[string]runtime.Value, err error) {
	if i == nil {
		return
	}
	entry := interfaceMethodDictionaryCacheEntry{methods: cloneRuntimeValueMap(methods), err: err}
	if i.envSingleThread {
		if i.interfaceMethodDictionaryCache == nil {
			i.interfaceMethodDictionaryCache = make(map[interfaceMethodDictionaryCacheKey]interfaceMethodDictionaryCacheEntry)
		}
		i.interfaceMethodDictionaryCache[key] = entry
		return
	}
	i.methodCacheMu.Lock()
	defer i.methodCacheMu.Unlock()
	if i.interfaceMethodDictionaryCache == nil {
		i.interfaceMethodDictionaryCache = make(map[interfaceMethodDictionaryCacheKey]interfaceMethodDictionaryCacheEntry)
	}
	i.interfaceMethodDictionaryCache[key] = entry
}

func (i *Interpreter) cachedTypeInfoName(info typeInfo) string {
	if len(info.typeArgs) == 0 {
		return info.name
	}
	if len(info.typeArgs) > 3 {
		return typeInfoToString(info)
	}
	key := typeInfoCacheKey{
		name:     info.name,
		argCount: uint8(len(info.typeArgs)),
	}
	if len(info.typeArgs) > 0 {
		key.arg0 = info.typeArgs[0]
	}
	if len(info.typeArgs) > 1 {
		key.arg1 = info.typeArgs[1]
	}
	if len(info.typeArgs) > 2 {
		key.arg2 = info.typeArgs[2]
	}
	i.typeInfoCacheMu.RLock()
	if cached, ok := i.typeInfoNameCache[key]; ok {
		i.typeInfoCacheMu.RUnlock()
		return cached
	}
	i.typeInfoCacheMu.RUnlock()
	typeName := typeInfoToString(info)
	i.typeInfoCacheMu.Lock()
	if i.typeInfoNameCache == nil {
		i.typeInfoNameCache = make(map[typeInfoCacheKey]string)
	}
	if existing, ok := i.typeInfoNameCache[key]; ok {
		i.typeInfoCacheMu.Unlock()
		return existing
	}
	i.typeInfoNameCache[key] = typeName
	i.typeInfoCacheMu.Unlock()
	return typeName
}

func (i *Interpreter) cachedTypeExpressionFromInfo(info typeInfo) ast.TypeExpression {
	if info.name == "" {
		return nil
	}
	if len(info.typeArgs) == 0 {
		return cachedSimpleTypeExpression(info.name)
	}
	if info.name == "Array" && len(info.typeArgs) == 1 {
		return cachedArrayTypeExpression(info.typeArgs[0])
	}
	if len(info.typeArgs) == 1 {
		if _, ok := info.typeArgs[0].(*ast.WildcardTypeExpression); ok {
			switch info.name {
			case "Iterator":
				return cachedIteratorTypeExpression
			case "Future":
				return cachedFutureTypeExpression
			}
		}
	}
	if i == nil {
		return typeExpressionFromInfo(info)
	}
	key := typeExpressionCacheKey{
		name: info.name,
		args: makeTypeExpressionSliceKey(info.typeArgs),
	}
	i.typeInfoCacheMu.RLock()
	if cached, ok := i.typeInfoExpressionCache[key]; ok {
		i.typeInfoCacheMu.RUnlock()
		return cached
	}
	i.typeInfoCacheMu.RUnlock()
	base := cachedSimpleTypeExpression(info.name)
	args := append([]ast.TypeExpression(nil), info.typeArgs...)
	created := ast.NewGenericTypeExpression(base, args)
	i.typeInfoCacheMu.Lock()
	if i.typeInfoExpressionCache == nil {
		i.typeInfoExpressionCache = make(map[typeExpressionCacheKey]ast.TypeExpression)
	}
	if existing, ok := i.typeInfoExpressionCache[key]; ok {
		i.typeInfoCacheMu.Unlock()
		return existing
	}
	i.typeInfoExpressionCache[key] = created
	i.typeInfoCacheMu.Unlock()
	return created
}

func (i *Interpreter) lookupImplEntry(info typeInfo, interfaceName string, ifaceArgs []ast.TypeExpression) (*implCandidate, error) {
	return i.lookupImplEntryForPackage(info, interfaceName, ifaceArgs, i.currentPackage)
}

func (i *Interpreter) lookupImplEntryForPackage(info typeInfo, interfaceName string, ifaceArgs []ast.TypeExpression, packageName string) (*implCandidate, error) {
	interfaceName = i.canonicalInterfaceName(interfaceName)
	cacheKey := i.makeInterfaceImplCacheKeyForPackage(info, interfaceName, ifaceArgs, packageName)
	if cached, ok := i.lookupSelectedInterfaceImplCache(cacheKey); ok {
		if cached.err != nil {
			return nil, cached.err
		}
		if !cached.found {
			return nil, nil
		}
		return &cached.candidate, nil
	}
	best, err := i.lookupImplEntryUncachedForPackage(info, interfaceName, ifaceArgs, packageName)
	i.storeSelectedInterfaceImplCache(cacheKey, best, err)
	return best, err
}

func (i *Interpreter) lookupImplEntryUncached(info typeInfo, interfaceName string, ifaceArgs []ast.TypeExpression) (*implCandidate, error) {
	return i.lookupImplEntryUncachedForPackage(info, interfaceName, ifaceArgs, i.currentPackage)
}

func (i *Interpreter) lookupImplEntryUncachedForPackage(info typeInfo, interfaceName string, ifaceArgs []ast.TypeExpression, packageName string) (*implCandidate, error) {
	matches, err := i.collectImplCandidatesForPackage(info, interfaceName, "", ifaceArgs, packageName)
	if len(matches) == 0 {
		return nil, err
	}
	if interfaceName != "" {
		direct := make([]implCandidate, 0, len(matches))
		for _, cand := range matches {
			if cand.entry != nil && cand.entry.interfaceName == interfaceName {
				direct = append(direct, cand)
			}
		}
		if len(direct) > 0 {
			matches = direct
		}
	}
	best, ambiguous := i.selectBestCandidate(matches)
	if ambiguous != nil {
		detail := descriptionsFromCandidates(ambiguous)
		typeDesc := typeInfoToString(info)
		if typeDesc == "<unknown>" {
			typeDesc = info.name
		}
		return nil, fmt.Errorf("ambiguous implementations of %s for %s: %s", interfaceName, typeDesc, strings.Join(detail, ", "))
	}
	if best == nil {
		return nil, nil
	}
	return best, nil
}

func (i *Interpreter) findMethodCached(info typeInfo, methodName string, interfaceFilter string) (runtime.Value, error) {
	return i.findMethodCachedForPackage(info, methodName, interfaceFilter, i.currentPackage)
}

func (i *Interpreter) findMethodCachedForPackage(info typeInfo, methodName string, interfaceFilter string, packageName string) (runtime.Value, error) {
	typeName := i.cachedTypeInfoName(info)
	key := methodCacheKey{typeName: typeName, methodName: methodName, ifaceFilter: interfaceFilter, packageName: packageName}
	i.methodCacheMu.RLock()
	entry, ok := i.methodCache[key]
	i.methodCacheMu.RUnlock()
	if ok {
		return entry.method, entry.err
	}
	method, err := i.findMethodForPackage(info, methodName, interfaceFilter, nil, packageName)
	i.methodCacheMu.Lock()
	if existing, exists := i.methodCache[key]; exists {
		i.methodCacheMu.Unlock()
		return existing.method, existing.err
	}
	i.methodCache[key] = methodCacheEntry{method: method, err: err}
	i.methodCacheMu.Unlock()
	return method, err
}

func (i *Interpreter) findMethod(info typeInfo, methodName string, interfaceFilter string, ifaceArgs []ast.TypeExpression) (runtime.Value, error) {
	return i.findMethodForPackage(info, methodName, interfaceFilter, ifaceArgs, i.currentPackage)
}

func (i *Interpreter) findMethodForPackage(info typeInfo, methodName string, interfaceFilter string, ifaceArgs []ast.TypeExpression, packageName string) (runtime.Value, error) {
	interfaceFilter = i.canonicalInterfaceName(interfaceFilter)
	var matches []implCandidate
	var err error
	if interfaceFilter == "" {
		matches, err = i.collectImplCandidatesForPackage(info, "", methodName, nil, packageName)
	} else {
		names := i.interfaceSearchNames(interfaceFilter, make(map[string]struct{}))
		if len(names) == 0 {
			names = []string{interfaceFilter}
		}
		var constraintErr error
		for _, name := range names {
			candidates, candErr := i.collectImplCandidatesForPackage(info, name, methodName, ifaceArgs, packageName)
			if candErr != nil && constraintErr == nil {
				constraintErr = candErr
			}
			if len(candidates) > 0 {
				matches = append(matches, candidates...)
			}
		}
		if len(matches) == 0 {
			return nil, constraintErr
		}
	}
	if len(matches) == 0 {
		return nil, err
	}
	if interfaceFilter != "" {
		direct := make([]implCandidate, 0, len(matches))
		for _, cand := range matches {
			if cand.entry != nil && cand.entry.interfaceName == interfaceFilter {
				direct = append(direct, cand)
			}
		}
		if len(direct) > 0 {
			matches = direct
		}
	}
	matches = dedupeImplCandidates(matches)
	methodMatches := make([]methodMatch, 0, len(matches))
	for _, cand := range matches {
		method := cand.entry.methods[methodName]
		if method == nil {
			if ifaceDef, ok := i.interfaces[cand.entry.interfaceName]; ok && ifaceDef.Node != nil {
				for _, sig := range ifaceDef.Node.Signatures {
					if sig == nil || sig.Name == nil || sig.Name.Name != methodName || sig.DefaultImpl == nil {
						continue
					}
					defaultVal, ok, err := i.interfaceDefaultMethodValue(ifaceDef, methodName)
					if err != nil {
						return nil, err
					}
					if !ok {
						break
					}
					method = defaultVal
					if cand.entry.methods == nil {
						cand.entry.methods = make(map[string]runtime.Value)
					}
					mergeFunctionLike(cand.entry.methods, methodName, method)
					break
				}
			}
		}
		if method == nil {
			continue
		}
		methodMatches = append(methodMatches, methodMatch{candidate: cand, method: method})
	}
	if len(methodMatches) == 0 {
		return nil, err
	}
	methodMatches = dedupeMethodMatches(methodMatches)
	if len(methodMatches) > 1 {
		explicit := make([]methodMatch, 0, len(methodMatches))
		for _, match := range methodMatches {
			if implDefinesMethod(match.candidate.entry, methodName) {
				explicit = append(explicit, match)
			}
		}
		if len(explicit) > 0 {
			methodMatches = explicit
		}
	}
	best, ambiguous := i.selectBestMethodCandidate(methodMatches)
	if ambiguous != nil {
		detail := descriptionsFromMethodMatches(ambiguous)
		typeDesc := typeInfoToString(info)
		if typeDesc == "<unknown>" {
			typeDesc = info.name
		}
		ifaceName := methodName
		if len(ambiguous) > 0 && ambiguous[0].candidate.entry != nil && ambiguous[0].candidate.entry.interfaceName != "" {
			ifaceName = ambiguous[0].candidate.entry.interfaceName
		}
		if len(detail) == 0 {
			detail = []string{"<unknown>"}
		}
		return nil, fmt.Errorf("ambiguous implementations of %s for %s: %s", ifaceName, typeDesc, strings.Join(detail, ", "))
	}
	if best == nil {
		return nil, nil
	}
	if fnVal := firstFunction(best.method); fnVal != nil {
		if fnDef, ok := fnVal.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
			return nil, fmt.Errorf("Method '%s' on %s is private", methodName, info.name)
		}
	}
	return best.method, nil
}

func implDefinesMethod(entry *implEntry, methodName string) bool {
	if entry == nil || entry.definition == nil || methodName == "" {
		return false
	}
	for _, fn := range entry.definition.Definitions {
		if fn == nil || fn.ID == nil {
			continue
		}
		if fn.ID.Name == methodName {
			return true
		}
	}
	return false
}

func (i *Interpreter) interfaceSearchNames(interfaceName string, visited map[string]struct{}) []string {
	interfaceName = i.canonicalInterfaceName(interfaceName)
	if interfaceName == "" {
		return nil
	}
	if _, seen := visited[interfaceName]; seen {
		return nil
	}
	visited[interfaceName] = struct{}{}
	names := []string{interfaceName}
	ifaceDef, ok := i.interfaces[interfaceName]
	if !ok || ifaceDef == nil || ifaceDef.Node == nil {
		return names
	}
	for _, base := range ifaceDef.Node.BaseInterfaces {
		info, ok := parseTypeExpression(base)
		if !ok || info.name == "" {
			continue
		}
		names = append(names, i.interfaceSearchNames(info.name, visited)...)
	}
	return names
}

func (i *Interpreter) interfaceExtendsInterface(candidate string, target string, visited map[string]struct{}) bool {
	candidate = i.canonicalInterfaceName(candidate)
	target = i.canonicalInterfaceName(target)
	if candidate == "" || target == "" {
		return false
	}
	if candidate == target {
		return true
	}
	if _, seen := visited[candidate]; seen {
		return false
	}
	visited[candidate] = struct{}{}
	ifaceDef, ok := i.interfaces[candidate]
	if !ok || ifaceDef == nil || ifaceDef.Node == nil {
		return false
	}
	for _, base := range ifaceDef.Node.BaseInterfaces {
		info, ok := parseTypeExpression(base)
		if !ok || info.name == "" {
			continue
		}
		if i.interfaceExtendsInterface(info.name, target, visited) {
			return true
		}
	}
	return false
}

func (i *Interpreter) typeImplementsInterface(info typeInfo, interfaceName string, ifaceArgs []ast.TypeExpression, visited map[interfaceImplCacheKey]struct{}) (bool, error) {
	return i.typeImplementsInterfaceForPackage(info, interfaceName, ifaceArgs, visited, i.currentPackage)
}

func (i *Interpreter) typeImplementsInterfaceForPackage(info typeInfo, interfaceName string, ifaceArgs []ast.TypeExpression, visited map[interfaceImplCacheKey]struct{}, packageName string) (bool, error) {
	interfaceName = i.canonicalInterfaceName(interfaceName)
	if info.name == "" || interfaceName == "" {
		return false, nil
	}
	if interfaceName == "Error" && info.name == "Error" {
		return true, nil
	}
	cacheKey := i.makeInterfaceImplCacheKeyForPackage(info, interfaceName, ifaceArgs, packageName)
	if cached, ok := i.lookupInterfaceImplCache(cacheKey); ok {
		return cached.ok, cached.err
	}
	if _, seen := visited[cacheKey]; seen {
		return true, nil
	}
	visited[cacheKey] = struct{}{}
	ifaceDef, ok := i.interfaces[interfaceName]
	if ok && ifaceDef != nil && ifaceDef.Node != nil && len(ifaceDef.Node.BaseInterfaces) > 0 {
		for _, base := range ifaceDef.Node.BaseInterfaces {
			baseInfo, ok := parseTypeExpression(base)
			if !ok || baseInfo.name == "" {
				i.storeInterfaceImplCache(cacheKey, false, nil)
				return false, nil
			}
			okImpl, err := i.typeImplementsInterfaceForPackage(info, baseInfo.name, baseInfo.typeArgs, visited, packageName)
			if err != nil || !okImpl {
				i.storeInterfaceImplCache(cacheKey, okImpl, err)
				return okImpl, err
			}
		}
		if len(ifaceDef.Node.Signatures) == 0 {
			i.storeSelectedInterfaceImplCache(cacheKey, nil, nil)
			i.storeInterfaceImplCache(cacheKey, true, nil)
			return true, nil
		}
	}
	entry, err := i.lookupImplEntryForPackage(info, interfaceName, ifaceArgs, packageName)
	if err != nil {
		// In compiled no-bootstrap mode, trust the compiled dispatch table.
		if i.compiledImplChecker != nil && i.compiledImplChecker(info.name, interfaceName) {
			i.storeSelectedInterfaceImplCache(cacheKey, nil, nil)
			i.storeInterfaceImplCache(cacheKey, true, nil)
			return true, nil
		}
		i.storeInterfaceImplCache(cacheKey, false, err)
		return false, err
	}
	if entry == nil && i.compiledImplChecker != nil && i.compiledImplChecker(info.name, interfaceName) {
		i.storeInterfaceImplCache(cacheKey, true, nil)
		return true, nil
	}
	okImpl := entry != nil
	i.storeInterfaceImplCache(cacheKey, okImpl, nil)
	return okImpl, nil
}

func (i *Interpreter) interfaceMatches(val *runtime.InterfaceValue, interfaceName string, ifaceArgs []ast.TypeExpression) bool {
	interfaceName = i.canonicalInterfaceName(interfaceName)
	if val == nil {
		return false
	}
	if val.Interface != nil {
		if interfaceDefinitionIdentity(val.Interface) == interfaceName && interfaceArgsEqual(i, val.InterfaceArgs, ifaceArgs) {
			return true
		}
	}
	info, ok := i.getTypeInfoForValue(val.Underlying)
	if !ok {
		return false
	}
	okImpl, err := i.typeImplementsInterface(info, interfaceName, ifaceArgs, make(map[interfaceImplCacheKey]struct{}))
	return err == nil && okImpl
}

func interfaceArgsEqual(i *Interpreter, a, b []ast.TypeExpression) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for idx := range a {
		left := a[idx]
		right := b[idx]
		if left == nil || right == nil {
			return false
		}
		left = expandTypeAliases(left, i.typeAliases, nil)
		right = expandTypeAliases(right, i.typeAliases, nil)
		if !typeExpressionsEqual(left, right) {
			return false
		}
	}
	return true
}

func (i *Interpreter) selectStructMethod(inst *runtime.StructInstanceValue, methodName string) (runtime.Value, error) {
	if inst == nil {
		return nil, nil
	}
	info, ok := i.typeInfoFromStructInstance(inst)
	if !ok {
		return nil, nil
	}
	return i.findMethodCached(info, methodName, "")
}
