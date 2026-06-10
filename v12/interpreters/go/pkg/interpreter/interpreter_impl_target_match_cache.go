package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type implTargetMatchCacheKey struct {
	target       ast.TypeExpression
	receiverType implTargetMatchReceiverTypeKey
}

type implTargetMatchReceiverTypeKey struct {
	name      string
	argCount  uint8
	arg0      string
	arg1      string
	arg2      string
	signature string
}

const implTargetMatchCacheMaxEntries = 2048

func (i *Interpreter) implContextTargetMatchesReceiver(target ast.TypeExpression, receiver runtime.Value, info typeInfo, hasInfo bool) bool {
	key, cacheable := i.implTargetMatchCacheKeyForReceiver(target, receiver, info, hasInfo)
	if cacheable {
		if matched, ok := i.lookupImplTargetMatchCache(key); ok {
			return matched
		}
	}
	matched := i.matchesType(target, receiver)
	if cacheable {
		i.storeImplTargetMatchCache(key, matched)
	}
	return matched
}

func (i *Interpreter) implContextTargetMatchesReceiverNominal(target ast.TypeExpression, receiverTypeName string, hasReceiverTypeName bool) (bool, bool) {
	if target == nil || !hasReceiverTypeName || receiverTypeName == "" {
		return false, false
	}
	if expanded := i.expandTypeAliasesCached(target); expanded != nil {
		target = expanded
	}
	switch t := target.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return false, false
		}
		name := normalizeKernelAliasName(t.Name.Name)
		if name == "Self" || i.implTargetOpenSimpleTypeName(name) {
			return true, true
		}
		if name == receiverTypeName {
			return true, true
		}
		return false, false
	case *ast.GenericTypeExpression:
		base, ok := t.Base.(*ast.SimpleTypeExpression)
		if !ok || base == nil || base.Name == nil {
			return false, false
		}
		baseName := normalizeKernelAliasName(base.Name.Name)
		if baseName == "Self" || implTargetOpenGenericBaseName(baseName) {
			return true, true
		}
		if baseName != receiverTypeName {
			return false, false
		}
		for _, arg := range t.Arguments {
			if !i.implTargetOpenTypeArgument(arg) {
				return false, false
			}
		}
		return true, true
	case *ast.UnionTypeExpression:
		for _, member := range t.Members {
			if matched, decided := i.implContextTargetMatchesReceiverNominal(member, receiverTypeName, hasReceiverTypeName); decided && matched {
				return true, true
			}
		}
		return false, false
	default:
		return false, false
	}
}

func (i *Interpreter) implTargetOpenSimpleTypeName(name string) bool {
	if name == "" {
		return false
	}
	if isPrimitiveName(name) {
		return false
	}
	if i != nil {
		if _, ok := i.interfaces[name]; ok {
			return false
		}
		if i.isKnownTypeName(name) {
			return false
		}
	}
	return true
}

func implTargetOpenGenericBaseName(name string) bool {
	return len(name) == 1 && name[0] >= 'A' && name[0] <= 'Z'
}

func (i *Interpreter) ensureMethodReceiverTypeInfo(receiver runtime.Value, info *typeInfo, hasInfo *bool, loaded *bool) bool {
	if loaded == nil || info == nil || hasInfo == nil {
		return false
	}
	if !*loaded {
		if receiver != nil {
			*info, *hasInfo = i.getTypeInfoForValue(receiver)
		}
		*loaded = true
	}
	return *hasInfo
}

func methodReceiverNominalTypeName(receiver runtime.Value) (string, bool) {
	receiver = bytecodeMaterializeRawValue(bytecodeSlotReadValue(receiver))
	switch v := receiver.(type) {
	case *runtime.StructInstanceValue:
		if v == nil || v.Definition == nil || v.Definition.Node == nil || v.Definition.Node.ID == nil {
			return "", false
		}
		return v.Definition.Node.ID.Name, true
	case *runtime.StructDefinitionValue:
		if v == nil || v.Node == nil || v.Node.ID == nil || !isSingletonStructDef(v.Node) {
			return "", false
		}
		return v.Node.ID.Name, true
	case runtime.StructDefinitionValue:
		if v.Node == nil || v.Node.ID == nil || !isSingletonStructDef(v.Node) {
			return "", false
		}
		return v.Node.ID.Name, true
	case *runtime.InterfaceValue:
		if v == nil {
			return "", false
		}
		return methodReceiverNominalTypeName(v.Underlying)
	case runtime.InterfaceValue:
		return methodReceiverNominalTypeName(v.Underlying)
	case *runtime.ArrayValue:
		if v == nil {
			return "", false
		}
		return "Array", true
	case *runtime.IteratorValue:
		if v == nil {
			return "", false
		}
		return "Iterator", true
	case *runtime.FutureValue:
		if v == nil {
			return "", false
		}
		return "Future", true
	case runtime.IteratorEndValue:
		return "IteratorEnd", true
	case *runtime.IteratorEndValue:
		if v == nil {
			return "", false
		}
		return "IteratorEnd", true
	case runtime.ErrorValue:
		if v.TypeName != nil && v.TypeName.Name != "" {
			return v.TypeName.Name, true
		}
		return "Error", true
	case *runtime.ErrorValue:
		if v == nil {
			return "", false
		}
		if v.TypeName != nil && v.TypeName.Name != "" {
			return v.TypeName.Name, true
		}
		return "Error", true
	default:
		return "", false
	}
}

func (i *Interpreter) implTargetOpenTypeArgument(expr ast.TypeExpression) bool {
	switch t := expr.(type) {
	case *ast.WildcardTypeExpression:
		return true
	case *ast.SimpleTypeExpression:
		return t != nil && t.Name != nil && i.implTargetOpenSimpleTypeName(t.Name.Name)
	default:
		return false
	}
}

func (i *Interpreter) implTargetMatchCacheKeyForReceiver(target ast.TypeExpression, receiver runtime.Value, info typeInfo, hasInfo bool) (implTargetMatchCacheKey, bool) {
	if target == nil || !hasInfo || info.name == "" {
		return implTargetMatchCacheKey{}, false
	}
	if !implTargetMatchReceiverCacheable(receiver) {
		return implTargetMatchCacheKey{}, false
	}
	return implTargetMatchCacheKey{
		target:       target,
		receiverType: implTargetMatchReceiverTypeKeyFromInfo(info),
	}, true
}

func implTargetMatchReceiverTypeKeyFromInfo(info typeInfo) implTargetMatchReceiverTypeKey {
	key := implTargetMatchReceiverTypeKey{
		name:     info.name,
		argCount: uint8(len(info.typeArgs)),
	}
	switch len(info.typeArgs) {
	case 0:
		return key
	case 1:
		key.arg0 = stableImplTargetMatchTypeExpressionKey(info.typeArgs[0])
	case 2:
		key.arg0 = stableImplTargetMatchTypeExpressionKey(info.typeArgs[0])
		key.arg1 = stableImplTargetMatchTypeExpressionKey(info.typeArgs[1])
	case 3:
		key.arg0 = stableImplTargetMatchTypeExpressionKey(info.typeArgs[0])
		key.arg1 = stableImplTargetMatchTypeExpressionKey(info.typeArgs[1])
		key.arg2 = stableImplTargetMatchTypeExpressionKey(info.typeArgs[2])
	default:
		key.signature = typeExpressionSliceSignature(info.typeArgs)
	}
	return key
}

func stableImplTargetMatchTypeExpressionKey(expr ast.TypeExpression) string {
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return "<?>"
		}
		return t.Name.Name
	case *ast.WildcardTypeExpression:
		return "_"
	default:
		return typeExpressionToString(expr)
	}
}

func implTargetMatchReceiverCacheable(receiver runtime.Value) bool {
	receiver = bytecodeMaterializeRawValue(bytecodeSlotReadValue(receiver))
	switch v := receiver.(type) {
	case *runtime.StructInstanceValue:
		if v == nil || v.Definition == nil || v.Definition.Node == nil || v.Definition.Node.ID == nil {
			return false
		}
		// Array element matching can depend on mutable array contents, so leave it
		// on the uncached path even when represented as a stdlib struct value.
		return v.Definition.Node.ID.Name != "Array"
	case *runtime.StructDefinitionValue:
		return v != nil && isSingletonStructDef(v.Node)
	case runtime.StructDefinitionValue:
		return isSingletonStructDef(v.Node)
	default:
		return false
	}
}

func (i *Interpreter) lookupImplTargetMatchCache(key implTargetMatchCacheKey) (bool, bool) {
	if i == nil {
		return false, false
	}
	if i.envSingleThread {
		matched, ok := i.implTargetMatchCache[key]
		return matched, ok
	}
	i.methodCacheMu.RLock()
	defer i.methodCacheMu.RUnlock()
	matched, ok := i.implTargetMatchCache[key]
	return matched, ok
}

func (i *Interpreter) storeImplTargetMatchCache(key implTargetMatchCacheKey, matched bool) {
	if i == nil {
		return
	}
	if i.envSingleThread {
		if i.implTargetMatchCache == nil {
			i.implTargetMatchCache = make(map[implTargetMatchCacheKey]bool)
		}
		if len(i.implTargetMatchCache) >= implTargetMatchCacheMaxEntries {
			clear(i.implTargetMatchCache)
		}
		i.implTargetMatchCache[key] = matched
		return
	}
	i.methodCacheMu.Lock()
	defer i.methodCacheMu.Unlock()
	if i.implTargetMatchCache == nil {
		i.implTargetMatchCache = make(map[implTargetMatchCacheKey]bool)
	}
	if len(i.implTargetMatchCache) >= implTargetMatchCacheMaxEntries {
		clear(i.implTargetMatchCache)
	}
	i.implTargetMatchCache[key] = matched
}
