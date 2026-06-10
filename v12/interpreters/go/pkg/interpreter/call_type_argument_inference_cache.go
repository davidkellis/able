package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type inferredCallTypeArgumentCacheKey struct {
	function ast.Node
	args     typeExpressionSliceKey
}

type inferredCallTypeArgumentRuntimeArgKey struct {
	kind         runtime.Kind
	typeName     string
	typeArgCount uint8
	typeArg0     ast.TypeExpression
	typeArg1     ast.TypeExpression
}

type inferredCallTypeArgumentRuntimeArgSignature struct {
	count  uint8
	inline [overloadArgSignatureInlineLimit]inferredCallTypeArgumentRuntimeArgKey
}

type inferredCallTypeArgumentRuntimeCacheKey1 struct {
	function ast.Node
	arg0     inferredCallTypeArgumentRuntimeArgKey
}

type inferredCallTypeArgumentRuntimeCacheKey2 struct {
	function ast.Node
	arg0     inferredCallTypeArgumentRuntimeArgKey
	arg1     inferredCallTypeArgumentRuntimeArgKey
}

type inferredCallTypeArgumentRuntimeCacheKey struct {
	function ast.Node
	args     inferredCallTypeArgumentRuntimeArgSignature
}

func (i *Interpreter) lookupInferredCallTypeArgumentCache(key inferredCallTypeArgumentCacheKey) ([]ast.TypeExpression, bool) {
	if i == nil || key.function == nil {
		return nil, false
	}
	if i.envSingleThread {
		typeArgs, ok := i.inferredCallTypeArgumentCache[key]
		return typeArgs, ok
	}
	i.inferredCallTypeArgumentCacheMu.RLock()
	defer i.inferredCallTypeArgumentCacheMu.RUnlock()
	typeArgs, ok := i.inferredCallTypeArgumentCache[key]
	return typeArgs, ok
}

func (i *Interpreter) storeInferredCallTypeArgumentCache(key inferredCallTypeArgumentCacheKey, typeArgs []ast.TypeExpression) {
	if i == nil || key.function == nil {
		return
	}
	cached := append([]ast.TypeExpression(nil), typeArgs...)
	if i.envSingleThread {
		if i.inferredCallTypeArgumentCache == nil {
			i.inferredCallTypeArgumentCache = make(map[inferredCallTypeArgumentCacheKey][]ast.TypeExpression)
		}
		i.inferredCallTypeArgumentCache[key] = cached
		return
	}
	i.inferredCallTypeArgumentCacheMu.Lock()
	defer i.inferredCallTypeArgumentCacheMu.Unlock()
	if i.inferredCallTypeArgumentCache == nil {
		i.inferredCallTypeArgumentCache = make(map[inferredCallTypeArgumentCacheKey][]ast.TypeExpression)
	}
	i.inferredCallTypeArgumentCache[key] = cached
}

func (i *Interpreter) inferredCallTypeArgumentRuntimeCacheEntryCount() int {
	if i == nil {
		return 0
	}
	if i.envSingleThread {
		return len(i.inferredCallTypeArgumentRuntimeCache1) +
			len(i.inferredCallTypeArgumentRuntimeCache2) +
			len(i.inferredCallTypeArgumentRuntimeCache)
	}
	i.inferredCallTypeArgumentCacheMu.RLock()
	defer i.inferredCallTypeArgumentCacheMu.RUnlock()
	return len(i.inferredCallTypeArgumentRuntimeCache1) +
		len(i.inferredCallTypeArgumentRuntimeCache2) +
		len(i.inferredCallTypeArgumentRuntimeCache)
}

func (i *Interpreter) lookupInferredCallTypeArgumentRuntimeCache(function ast.Node, sig inferredCallTypeArgumentRuntimeArgSignature) ([]ast.TypeExpression, bool) {
	if i == nil || function == nil {
		return nil, false
	}
	if i.envSingleThread {
		switch sig.count {
		case 1:
			typeArgs, ok := i.inferredCallTypeArgumentRuntimeCache1[inferredCallTypeArgumentRuntimeCacheKey1{
				function: function,
				arg0:     sig.inline[0],
			}]
			return typeArgs, ok
		case 2:
			typeArgs, ok := i.inferredCallTypeArgumentRuntimeCache2[inferredCallTypeArgumentRuntimeCacheKey2{
				function: function,
				arg0:     sig.inline[0],
				arg1:     sig.inline[1],
			}]
			return typeArgs, ok
		default:
			typeArgs, ok := i.inferredCallTypeArgumentRuntimeCache[inferredCallTypeArgumentRuntimeCacheKey{
				function: function,
				args:     sig,
			}]
			return typeArgs, ok
		}
	}
	i.inferredCallTypeArgumentCacheMu.RLock()
	defer i.inferredCallTypeArgumentCacheMu.RUnlock()
	switch sig.count {
	case 1:
		typeArgs, ok := i.inferredCallTypeArgumentRuntimeCache1[inferredCallTypeArgumentRuntimeCacheKey1{
			function: function,
			arg0:     sig.inline[0],
		}]
		return typeArgs, ok
	case 2:
		typeArgs, ok := i.inferredCallTypeArgumentRuntimeCache2[inferredCallTypeArgumentRuntimeCacheKey2{
			function: function,
			arg0:     sig.inline[0],
			arg1:     sig.inline[1],
		}]
		return typeArgs, ok
	default:
		typeArgs, ok := i.inferredCallTypeArgumentRuntimeCache[inferredCallTypeArgumentRuntimeCacheKey{
			function: function,
			args:     sig,
		}]
		return typeArgs, ok
	}
}

func (i *Interpreter) storeInferredCallTypeArgumentRuntimeCache(function ast.Node, sig inferredCallTypeArgumentRuntimeArgSignature, typeArgs []ast.TypeExpression) {
	if i == nil || function == nil {
		return
	}
	cached := append([]ast.TypeExpression(nil), typeArgs...)
	if i.envSingleThread {
		switch sig.count {
		case 1:
			if i.inferredCallTypeArgumentRuntimeCache1 == nil {
				i.inferredCallTypeArgumentRuntimeCache1 = make(map[inferredCallTypeArgumentRuntimeCacheKey1][]ast.TypeExpression)
			}
			i.inferredCallTypeArgumentRuntimeCache1[inferredCallTypeArgumentRuntimeCacheKey1{
				function: function,
				arg0:     sig.inline[0],
			}] = cached
		case 2:
			if i.inferredCallTypeArgumentRuntimeCache2 == nil {
				i.inferredCallTypeArgumentRuntimeCache2 = make(map[inferredCallTypeArgumentRuntimeCacheKey2][]ast.TypeExpression)
			}
			i.inferredCallTypeArgumentRuntimeCache2[inferredCallTypeArgumentRuntimeCacheKey2{
				function: function,
				arg0:     sig.inline[0],
				arg1:     sig.inline[1],
			}] = cached
		default:
			if i.inferredCallTypeArgumentRuntimeCache == nil {
				i.inferredCallTypeArgumentRuntimeCache = make(map[inferredCallTypeArgumentRuntimeCacheKey][]ast.TypeExpression)
			}
			i.inferredCallTypeArgumentRuntimeCache[inferredCallTypeArgumentRuntimeCacheKey{
				function: function,
				args:     sig,
			}] = cached
		}
		return
	}
	i.inferredCallTypeArgumentCacheMu.Lock()
	defer i.inferredCallTypeArgumentCacheMu.Unlock()
	switch sig.count {
	case 1:
		if i.inferredCallTypeArgumentRuntimeCache1 == nil {
			i.inferredCallTypeArgumentRuntimeCache1 = make(map[inferredCallTypeArgumentRuntimeCacheKey1][]ast.TypeExpression)
		}
		i.inferredCallTypeArgumentRuntimeCache1[inferredCallTypeArgumentRuntimeCacheKey1{
			function: function,
			arg0:     sig.inline[0],
		}] = cached
	case 2:
		if i.inferredCallTypeArgumentRuntimeCache2 == nil {
			i.inferredCallTypeArgumentRuntimeCache2 = make(map[inferredCallTypeArgumentRuntimeCacheKey2][]ast.TypeExpression)
		}
		i.inferredCallTypeArgumentRuntimeCache2[inferredCallTypeArgumentRuntimeCacheKey2{
			function: function,
			arg0:     sig.inline[0],
			arg1:     sig.inline[1],
		}] = cached
	default:
		if i.inferredCallTypeArgumentRuntimeCache == nil {
			i.inferredCallTypeArgumentRuntimeCache = make(map[inferredCallTypeArgumentRuntimeCacheKey][]ast.TypeExpression)
		}
		i.inferredCallTypeArgumentRuntimeCache[inferredCallTypeArgumentRuntimeCacheKey{
			function: function,
			args:     sig,
		}] = cached
	}
}

func shallowArrayElementTypeExpressionForCallInference(arr *runtime.ArrayValue) (ast.TypeExpression, bool) {
	if arr == nil {
		return nil, false
	}
	if arr.State != nil && arr.State.ElementTypeTokenKnown {
		if expr, ok := arrayElementTypeExpressionFromIndexToken(arr.State.ElementTypeToken); ok {
			return expr, true
		}
	}
	for _, handle := range []int64{arr.Handle, arr.TrackedHandle} {
		if handle == 0 {
			continue
		}
		typeName, ok, err := runtime.ArrayStoreMonoElementTypeNameIfKnown(handle)
		if err == nil && ok {
			return cachedSimpleTypeExpression(typeName), true
		}
	}
	return nil, false
}

func (i *Interpreter) shallowCallInferenceRuntimeArgKey(arg runtime.Value) (inferredCallTypeArgumentRuntimeArgKey, bool) {
	arg = bytecodeSlotReadValue(arg)
	if kind, _, ok := bytecodeRawIntegerValueInfo(arg); ok {
		typeName := string(kind)
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindInteger,
			typeName: typeName,
		}, true
	}
	if _, kind, ok := bytecodeDirectRawFloatValue(arg); ok {
		typeName := string(kind)
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindFloat,
			typeName: typeName,
		}, true
	}
	switch v := arg.(type) {
	case nil:
		return inferredCallTypeArgumentRuntimeArgKey{typeName: "nil"}, false
	case runtime.StringValue:
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindString,
			typeName: "String",
		}, true
	case *runtime.StringValue:
		if v == nil {
			return inferredCallTypeArgumentRuntimeArgKey{typeName: "nil"}, false
		}
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindString,
			typeName: "String",
		}, true
	case runtime.BoolValue:
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindBool,
			typeName: "bool",
		}, true
	case *runtime.BoolValue:
		if v == nil {
			return inferredCallTypeArgumentRuntimeArgKey{typeName: "nil"}, false
		}
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindBool,
			typeName: "bool",
		}, true
	case runtime.CharValue:
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindChar,
			typeName: "char",
		}, true
	case *runtime.CharValue:
		if v == nil {
			return inferredCallTypeArgumentRuntimeArgKey{typeName: "nil"}, false
		}
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindChar,
			typeName: "char",
		}, true
	case runtime.NilValue:
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindNil,
			typeName: "nil",
		}, true
	case runtime.VoidValue:
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindVoid,
			typeName: "void",
		}, true
	case *runtime.VoidValue:
		if v == nil {
			return inferredCallTypeArgumentRuntimeArgKey{typeName: "nil"}, false
		}
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindVoid,
			typeName: "void",
		}, true
	case runtime.IteratorEndValue:
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindIteratorEnd,
			typeName: "IteratorEnd",
		}, true
	case *runtime.IteratorEndValue:
		if v == nil {
			return inferredCallTypeArgumentRuntimeArgKey{typeName: "nil"}, false
		}
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindIteratorEnd,
			typeName: "IteratorEnd",
		}, true
	case runtime.IntegerValue:
		typeName := string(v.TypeSuffix)
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindInteger,
			typeName: typeName,
		}, true
	case *runtime.IntegerValue:
		if v == nil {
			return inferredCallTypeArgumentRuntimeArgKey{typeName: "nil"}, false
		}
		typeName := string(v.TypeSuffix)
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindInteger,
			typeName: typeName,
		}, true
	case runtime.FloatValue:
		typeName := string(v.TypeSuffix)
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindFloat,
			typeName: typeName,
		}, true
	case *runtime.FloatValue:
		if v == nil {
			return inferredCallTypeArgumentRuntimeArgKey{typeName: "nil"}, false
		}
		typeName := string(v.TypeSuffix)
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindFloat,
			typeName: typeName,
		}, true
	case *runtime.StructDefinitionValue:
		if v == nil || v.Node == nil || v.Node.ID == nil || !isSingletonStructDef(v.Node) {
			return inferredCallTypeArgumentRuntimeArgKey{}, false
		}
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindStructDefinition,
			typeName: v.Node.ID.Name,
		}, true
	case runtime.StructDefinitionValue:
		return i.shallowCallInferenceRuntimeArgKey(&v)
	case *runtime.StructInstanceValue:
		if v == nil || v.Definition == nil || v.Definition.Node == nil || v.Definition.Node.ID == nil {
			return inferredCallTypeArgumentRuntimeArgKey{}, false
		}
		def := v.Definition.Node
		if def.ID.Name == "Array" {
			return inferredCallTypeArgumentRuntimeArgKey{}, false
		}
		key := inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindStructInstance,
			typeName: def.ID.Name,
		}
		if len(def.GenericParams) == 0 {
			return key, true
		}
		if structTypeArgsNeedInference(i.structGenericInferencePlan(def), v.TypeArguments) || len(v.TypeArguments) > 2 {
			return inferredCallTypeArgumentRuntimeArgKey{}, false
		}
		key.typeArgCount = uint8(len(v.TypeArguments))
		if key.typeArgCount > 0 {
			key.typeArg0 = v.TypeArguments[0]
		}
		if key.typeArgCount > 1 {
			key.typeArg1 = v.TypeArguments[1]
		}
		return key, true
	case *runtime.InterfaceValue:
		if v == nil || v.Interface == nil || v.Interface.Node == nil || v.Interface.Node.ID == nil {
			return inferredCallTypeArgumentRuntimeArgKey{}, false
		}
		typeName := v.Interface.Node.ID.Name
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindInterfaceValue,
			typeName: typeName,
		}, true
	case runtime.InterfaceValue:
		return i.shallowCallInferenceRuntimeArgKey(&v)
	case *runtime.ArrayValue:
		if v == nil {
			return inferredCallTypeArgumentRuntimeArgKey{}, false
		}
		elemType, ok := shallowArrayElementTypeExpressionForCallInference(v)
		if !ok {
			return inferredCallTypeArgumentRuntimeArgKey{}, false
		}
		key := inferredCallTypeArgumentRuntimeArgKey{
			kind:         runtime.KindArray,
			typeName:     "Array",
			typeArgCount: 1,
			typeArg0:     elemType,
		}
		return key, true
	case *runtime.IteratorValue:
		if v == nil {
			return inferredCallTypeArgumentRuntimeArgKey{}, false
		}
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:         runtime.KindIterator,
			typeName:     "Iterator",
			typeArgCount: 1,
			typeArg0:     cachedWildcardTypeExpression,
		}, true
	case *runtime.FutureValue:
		if v == nil {
			return inferredCallTypeArgumentRuntimeArgKey{}, false
		}
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:         runtime.KindFuture,
			typeName:     "Future",
			typeArgCount: 1,
			typeArg0:     cachedWildcardTypeExpression,
		}, true
	case runtime.ErrorValue:
		typeName := "Error"
		if v.TypeName != nil && v.TypeName.Name != "" {
			typeName = v.TypeName.Name
		}
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindError,
			typeName: typeName,
		}, true
	case *runtime.ErrorValue:
		if v == nil {
			return inferredCallTypeArgumentRuntimeArgKey{}, false
		}
		typeName := "Error"
		if v.TypeName != nil && v.TypeName.Name != "" {
			typeName = v.TypeName.Name
		}
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindError,
			typeName: typeName,
		}, true
	case *runtime.HostHandleValue:
		if v == nil || v.HandleType == "" {
			return inferredCallTypeArgumentRuntimeArgKey{}, false
		}
		return inferredCallTypeArgumentRuntimeArgKey{
			kind:     runtime.KindHostHandle,
			typeName: v.HandleType,
		}, true
	default:
		return inferredCallTypeArgumentRuntimeArgKey{}, false
	}
}

func (i *Interpreter) inferredCallTypeArgumentRuntimeSignature(max int, valueAt func(idx int) runtime.Value) (inferredCallTypeArgumentRuntimeArgSignature, bool) {
	if max < 0 || max > overloadArgSignatureInlineLimit {
		return inferredCallTypeArgumentRuntimeArgSignature{}, false
	}
	sig := inferredCallTypeArgumentRuntimeArgSignature{count: uint8(max)}
	if max == 0 {
		return sig, true
	}
	for idx := 0; idx < max; idx++ {
		key, ok := i.shallowCallInferenceRuntimeArgKey(valueAt(idx))
		if !ok {
			return inferredCallTypeArgumentRuntimeArgSignature{}, false
		}
		sig.inline[idx] = key
	}
	return sig, true
}

func (i *Interpreter) typeExpressionFromShallowCallInferenceRuntimeArgKey(key inferredCallTypeArgumentRuntimeArgKey) ast.TypeExpression {
	if key.typeName == "" {
		return nil
	}
	if key.typeArgCount == 0 {
		switch key.kind {
		case runtime.KindInteger:
			return cachedIntegerTypeExpression(runtime.IntegerType(key.typeName))
		case runtime.KindFloat:
			return cachedFloatTypeExpression(runtime.FloatType(key.typeName))
		default:
			return cachedSimpleTypeExpression(key.typeName)
		}
	}
	switch key.kind {
	case runtime.KindArray:
		return cachedArrayTypeExpression(key.typeArg0)
	case runtime.KindIterator:
		return cachedIteratorTypeExpression
	case runtime.KindFuture:
		return cachedFutureTypeExpression
	default:
		switch key.typeArgCount {
		case 1:
			return i.cachedTypeExpressionFromInfo(typeInfo{
				name:     key.typeName,
				typeArgs: i.cachedTypeExpressionTuple1(key.typeArg0),
			})
		case 2:
			return i.cachedTypeExpressionFromInfo(typeInfo{
				name:     key.typeName,
				typeArgs: i.cachedTypeExpressionTuple2(key.typeArg0, key.typeArg1),
			})
		default:
			return nil
		}
	}
}

func (i *Interpreter) inferredCallTypeArgumentActualTypesFromRuntimeSignature(sig inferredCallTypeArgumentRuntimeArgSignature) []ast.TypeExpression {
	switch sig.count {
	case 0:
		return nil
	case 1:
		return i.cachedTypeExpressionTuple1(i.typeExpressionFromShallowCallInferenceRuntimeArgKey(sig.inline[0]))
	case 2:
		return i.cachedTypeExpressionTuple2(
			i.typeExpressionFromShallowCallInferenceRuntimeArgKey(sig.inline[0]),
			i.typeExpressionFromShallowCallInferenceRuntimeArgKey(sig.inline[1]),
		)
	case 3:
		return i.cachedTypeExpressionTuple3(
			i.typeExpressionFromShallowCallInferenceRuntimeArgKey(sig.inline[0]),
			i.typeExpressionFromShallowCallInferenceRuntimeArgKey(sig.inline[1]),
			i.typeExpressionFromShallowCallInferenceRuntimeArgKey(sig.inline[2]),
		)
	default:
		actualTypes := make([]ast.TypeExpression, int(sig.count))
		for idx := 0; idx < int(sig.count); idx++ {
			actualTypes[idx] = i.typeExpressionFromShallowCallInferenceRuntimeArgKey(sig.inline[idx])
		}
		return actualTypes
	}
}

func (i *Interpreter) inferCallTypeArgumentsFromResolvedActualTypes(plan *functionCallGenericPlan, actualTypes []ast.TypeExpression) []ast.TypeExpression {
	if plan == nil || plan.expectedCount == 0 {
		return nil
	}
	var inline [3]ast.TypeExpression
	bindings := inlineBindingsSlice(inline[:], plan.expectedCount)
	max := len(plan.inferenceRelevantParams)
	if len(actualTypes) < max {
		max = len(actualTypes)
	}
	for idx := 0; idx < max; idx++ {
		param := plan.inferenceRelevantParams[idx]
		actual := actualTypes[idx]
		if param.paramType == nil || actual == nil {
			continue
		}
		matchTypeExpressionTemplateIndexed(param.paramType, actual, plan.genericIndex, bindings)
	}
	return i.cachedTypeArgumentsFromIndexedBindings(bindings)
}

func (i *Interpreter) inferAndSetCallTypeArgumentsFromValues(plan *functionCallGenericPlan, funcNode ast.Node, call *ast.FunctionCall, max int, valueAt func(idx int) runtime.Value) {
	if runtimeSig, ok := i.inferredCallTypeArgumentRuntimeSignature(max, valueAt); ok {
		if cached, ok := i.lookupInferredCallTypeArgumentRuntimeCache(funcNode, runtimeSig); ok {
			i.setInferredCallTypeArguments(call, cached)
			return
		}
		typeArgs := i.inferCallTypeArgumentsFromResolvedActualTypes(
			plan,
			i.inferredCallTypeArgumentActualTypesFromRuntimeSignature(runtimeSig),
		)
		i.storeInferredCallTypeArgumentRuntimeCache(funcNode, runtimeSig, typeArgs)
		i.setInferredCallTypeArguments(call, typeArgs)
		return
	}
	var inline [overloadArgSignatureInlineLimit]ast.TypeExpression
	actualTypes := inlineBindingsSlice(inline[:], max)
	for idx := 0; idx < max; idx++ {
		actualTypes[idx] = i.typeExpressionFromValue(valueAt(idx))
	}
	typeArgs := i.inferCallTypeArgumentsFromActualTypes(plan, funcNode, actualTypes)
	i.setInferredCallTypeArguments(call, typeArgs)
}

func (i *Interpreter) inferCallTypeArgumentsFromActualTypes(plan *functionCallGenericPlan, funcNode ast.Node, actualTypes []ast.TypeExpression) []ast.TypeExpression {
	if plan == nil || plan.expectedCount == 0 {
		return nil
	}
	cacheKey := inferredCallTypeArgumentCacheKey{
		function: funcNode,
		args:     makeTypeExpressionSliceKey(actualTypes),
	}
	if cached, ok := i.lookupInferredCallTypeArgumentCache(cacheKey); ok {
		return cached
	}
	typeArgs := i.inferCallTypeArgumentsFromResolvedActualTypes(plan, actualTypes)
	i.storeInferredCallTypeArgumentCache(cacheKey, typeArgs)
	return typeArgs
}
