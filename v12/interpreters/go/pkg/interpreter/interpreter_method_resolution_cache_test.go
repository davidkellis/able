package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func cacheProbeFunction(name string, closure *runtime.Environment) *runtime.FunctionValue {
	def := ast.Fn(
		name,
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Array")),
		},
		[]ast.Statement{
			ast.ID("self"),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)
	return &runtime.FunctionValue{Declaration: def, Closure: closure}
}

func TestResolveMethodFromPool_BoundMethodCacheInvalidatesWithMethodCache(t *testing.T) {
	interp := New()
	arr := interp.newArrayValue(nil, 0)
	env := interp.GlobalEnvironment()

	bucket := interp.inherentMethods["Array"]
	if bucket == nil {
		bucket = make(map[string]runtime.Value)
		interp.inherentMethods["Array"] = bucket
	}
	first := cacheProbeFunction("cache_probe_impl_1", env)
	bucket["cache_probe"] = first

	resolved, err := interp.resolveMethodFromPool(env, "cache_probe", arr, "")
	if err != nil {
		t.Fatalf("resolve first method: %v", err)
	}
	bound, ok := resolved.(runtime.BoundMethodValue)
	if !ok {
		t.Fatalf("expected bound method value, got %T (%#v)", resolved, resolved)
	}
	methodFn, ok := bound.Method.(*runtime.FunctionValue)
	if !ok {
		t.Fatalf("expected function method, got %T (%#v)", bound.Method, bound.Method)
	}
	if methodFn != first {
		t.Fatalf("expected first method function pointer, got %#v", methodFn)
	}
	if len(interp.boundMethodCache) == 0 {
		t.Fatalf("expected bound method cache to store first resolution")
	}

	secondResolved, err := interp.resolveMethodFromPool(env, "cache_probe", arr, "")
	if err != nil {
		t.Fatalf("resolve cached method: %v", err)
	}
	secondBound, ok := secondResolved.(runtime.BoundMethodValue)
	if !ok {
		t.Fatalf("expected bound method value on cached resolve, got %T (%#v)", secondResolved, secondResolved)
	}
	secondFn, ok := secondBound.Method.(*runtime.FunctionValue)
	if !ok || secondFn != first {
		t.Fatalf("expected cached bound method to keep first function, got %T (%#v)", secondBound.Method, secondBound.Method)
	}

	second := cacheProbeFunction("cache_probe_impl_2", env)
	bucket["cache_probe"] = second
	interp.invalidateMethodCache()
	if len(interp.boundMethodCache) != 0 {
		t.Fatalf("expected bound method cache clear on invalidate, got size=%d", len(interp.boundMethodCache))
	}

	resolvedAfterInvalidate, err := interp.resolveMethodFromPool(env, "cache_probe", arr, "")
	if err != nil {
		t.Fatalf("resolve method after invalidation: %v", err)
	}
	boundAfterInvalidate, ok := resolvedAfterInvalidate.(runtime.BoundMethodValue)
	if !ok {
		t.Fatalf("expected bound method after invalidation, got %T (%#v)", resolvedAfterInvalidate, resolvedAfterInvalidate)
	}
	afterFn, ok := boundAfterInvalidate.Method.(*runtime.FunctionValue)
	if !ok {
		t.Fatalf("expected function method after invalidation, got %T (%#v)", boundAfterInvalidate.Method, boundAfterInvalidate.Method)
	}
	if afterFn != second {
		t.Fatalf("expected second method function pointer after invalidation, got %#v", afterFn)
	}
}

func TestResolveMethodFromPool_BoundMethodCacheUsesPrimitiveTypeKeyForStrings(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()

	bucket := interp.inherentMethods["String"]
	if bucket == nil {
		bucket = make(map[string]runtime.Value)
		interp.inherentMethods["String"] = bucket
	}
	method := ast.Fn(
		"string_cache_probe",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("String")),
		},
		[]ast.Statement{
			ast.ID("self"),
		},
		ast.Ty("String"),
		nil,
		nil,
		false,
		false,
	)
	first := &runtime.FunctionValue{Declaration: method, Closure: env}
	bucket["cache_probe"] = first

	firstResolved, err := interp.resolveMethodFromPool(env, "cache_probe", runtime.StringValue{Val: "alpha"}, "")
	if err != nil {
		t.Fatalf("resolve first string method: %v", err)
	}
	firstBound, ok := firstResolved.(runtime.BoundMethodValue)
	if !ok {
		t.Fatalf("expected bound method value, got %T (%#v)", firstResolved, firstResolved)
	}
	firstFn, ok := firstBound.Method.(*runtime.FunctionValue)
	if !ok || firstFn != first {
		t.Fatalf("expected first string method pointer, got %T (%#v)", firstBound.Method, firstBound.Method)
	}
	if got := len(interp.boundMethodCache); got != 1 {
		t.Fatalf("expected one primitive string cache entry after first resolve, got %d", got)
	}

	secondResolved, err := interp.resolveMethodFromPool(env, "cache_probe", runtime.StringValue{Val: "beta"}, "")
	if err != nil {
		t.Fatalf("resolve second string method: %v", err)
	}
	secondBound, ok := secondResolved.(runtime.BoundMethodValue)
	if !ok {
		t.Fatalf("expected bound method value on second resolve, got %T (%#v)", secondResolved, secondResolved)
	}
	secondFn, ok := secondBound.Method.(*runtime.FunctionValue)
	if !ok || secondFn != first {
		t.Fatalf("expected primitive cache hit to keep first method pointer, got %T (%#v)", secondBound.Method, secondBound.Method)
	}
	if got := len(interp.boundMethodCache); got != 1 {
		t.Fatalf("expected primitive string resolves to reuse one cache entry, got %d", got)
	}
}

func TestStoreBoundMethodCache_RolloverKeepsNewestEntry(t *testing.T) {
	interp := New()
	for idx := 0; idx < boundMethodCacheMaxEntries; idx++ {
		interp.storeBoundMethodCache(
			boundMethodCacheKey{
				receiver:      idx,
				methodName:    "probe",
				ifaceFilter:   "",
				allowInherent: true,
			},
			runtime.StringValue{Val: "old"},
		)
	}
	if got := len(interp.boundMethodCache); got != boundMethodCacheMaxEntries {
		t.Fatalf("expected full bound method cache before rollover, got %d", got)
	}

	rolloverKey := boundMethodCacheKey{
		receiver:      "next",
		methodName:    "probe",
		ifaceFilter:   "",
		allowInherent: true,
	}
	rolloverValue := runtime.StringValue{Val: "new"}
	interp.storeBoundMethodCache(rolloverKey, rolloverValue)

	if got := len(interp.boundMethodCache); got != 1 {
		t.Fatalf("expected rollover to clear old entries and keep newest one, got size=%d", got)
	}
	stored, ok := interp.lookupBoundMethodCache(rolloverKey)
	if !ok {
		t.Fatalf("expected rollover entry to remain cached")
	}
	if str, ok := stored.(runtime.StringValue); !ok || str.Val != "new" {
		t.Fatalf("expected rollover cache value to survive, got %T (%#v)", stored, stored)
	}
}

func TestBoundMethodReceiverKeyPrimitiveIntegersUseSuffixKey(t *testing.T) {
	keyI32A, ok := boundMethodReceiverKey(runtime.NewSmallInt(7, runtime.IntegerI32))
	if !ok {
		t.Fatalf("expected i32 receiver key")
	}
	keyI32B, ok := boundMethodReceiverKey(runtime.NewSmallInt(11, runtime.IntegerI32))
	if !ok {
		t.Fatalf("expected second i32 receiver key")
	}
	keyI64, ok := boundMethodReceiverKey(runtime.NewSmallInt(11, runtime.IntegerI64))
	if !ok {
		t.Fatalf("expected i64 receiver key")
	}
	if keyI32A != keyI32B {
		t.Fatalf("expected i32 receiver keys to match across values, got %v and %v", keyI32A, keyI32B)
	}
	if keyI32A == keyI64 {
		t.Fatalf("expected integer receiver keys to differ by suffix, got shared key %v", keyI32A)
	}
}

func TestResolveMethodFromPool_DoesNotCachePrimitiveScopeFallbackCallable(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()

	first := &runtime.FunctionValue{
		Declaration: ast.Fn(
			"scope_fallback_a",
			[]*ast.FunctionParameter{
				ast.Param("self", ast.Ty("String")),
			},
			[]ast.Statement{
				ast.Str("a"),
			},
			ast.Ty("String"),
			nil,
			nil,
			false,
			false,
		),
		Closure: env,
	}
	second := &runtime.FunctionValue{
		Declaration: ast.Fn(
			"scope_fallback_b",
			[]*ast.FunctionParameter{
				ast.Param("self", ast.Ty("String")),
			},
			[]ast.Statement{
				ast.Str("b"),
			},
			ast.Ty("String"),
			nil,
			nil,
			false,
			false,
		),
		Closure: env,
	}

	env.Define("scope_fallback", first)
	resolvedA, err := interp.resolveMethodFromPool(env, "scope_fallback", runtime.StringValue{Val: "alpha"}, "")
	if err != nil {
		t.Fatalf("resolve primitive scope fallback A: %v", err)
	}
	boundA, ok := resolvedA.(runtime.BoundMethodValue)
	if !ok {
		t.Fatalf("expected bound method value for first scope fallback, got %T (%#v)", resolvedA, resolvedA)
	}
	methodA, ok := boundA.Method.(*runtime.FunctionValue)
	if !ok || methodA != first {
		t.Fatalf("expected first scope fallback pointer, got %T (%#v)", boundA.Method, boundA.Method)
	}
	if got := len(interp.boundMethodCache); got != 0 {
		t.Fatalf("expected primitive scope fallback to skip cache storage, got %d entries", got)
	}

	if err := env.Assign("scope_fallback", second); err != nil {
		t.Fatalf("assign primitive scope fallback: %v", err)
	}
	resolvedB, err := interp.resolveMethodFromPool(env, "scope_fallback", runtime.StringValue{Val: "beta"}, "")
	if err != nil {
		t.Fatalf("resolve primitive scope fallback B: %v", err)
	}
	boundB, ok := resolvedB.(runtime.BoundMethodValue)
	if !ok {
		t.Fatalf("expected bound method value for second scope fallback, got %T (%#v)", resolvedB, resolvedB)
	}
	methodB, ok := boundB.Method.(*runtime.FunctionValue)
	if !ok || methodB != second {
		t.Fatalf("expected second scope fallback pointer after assign, got %T (%#v)", boundB.Method, boundB.Method)
	}
	if got := len(interp.boundMethodCache); got != 0 {
		t.Fatalf("expected primitive scope fallback to remain uncached after assign, got %d entries", got)
	}
}

func TestTypeImplementsInterface_CachesAndInvalidatesWithMethodCache(t *testing.T) {
	interp := New()
	ifaceNode := ast.NewInterfaceDefinition(ast.NewIdentifier("Cacheable"), nil, nil, nil, nil, nil, false)
	interp.interfaces["Cacheable"] = &runtime.InterfaceDefinitionValue{Node: ifaceNode}
	interp.implMethods["String"] = []implEntry{
		{
			interfaceName: "Cacheable",
			definition: ast.NewImplementationDefinition(
				ast.NewIdentifier("Cacheable"),
				ast.Ty("String"),
				nil,
				nil,
				nil,
				nil,
				nil,
				false,
			),
		},
	}

	info := typeInfo{name: "String"}
	ok, err := interp.typeImplementsInterface(info, "Cacheable", nil, make(map[interfaceImplCacheKey]struct{}))
	if err != nil {
		t.Fatalf("typeImplementsInterface first lookup: %v", err)
	}
	if !ok {
		t.Fatalf("expected String to satisfy Cacheable on first lookup")
	}
	if len(interp.interfaceImplCache) == 0 {
		t.Fatalf("expected interface impl cache to store first lookup")
	}

	delete(interp.implMethods, "String")

	cachedOK, err := interp.typeImplementsInterface(info, "Cacheable", nil, make(map[interfaceImplCacheKey]struct{}))
	if err != nil {
		t.Fatalf("typeImplementsInterface cached lookup: %v", err)
	}
	if !cachedOK {
		t.Fatalf("expected cached String -> Cacheable result before invalidation")
	}

	interp.invalidateMethodCache()
	if len(interp.interfaceImplCache) != 0 {
		t.Fatalf("expected interface impl cache clear on invalidate, got size=%d", len(interp.interfaceImplCache))
	}

	afterInvalidate, err := interp.typeImplementsInterface(info, "Cacheable", nil, make(map[interfaceImplCacheKey]struct{}))
	if err != nil {
		t.Fatalf("typeImplementsInterface after invalidation: %v", err)
	}
	if afterInvalidate {
		t.Fatalf("expected String -> Cacheable to be recomputed as false after invalidation")
	}
}

func TestTypeImplementsInterface_PrimesSelectedImplLookupCache(t *testing.T) {
	interp := New()
	ifaceNode := ast.NewInterfaceDefinition(ast.NewIdentifier("Cacheable"), nil, nil, nil, nil, nil, false)
	interp.interfaces["Cacheable"] = &runtime.InterfaceDefinitionValue{Node: ifaceNode}
	interp.implMethods["String"] = []implEntry{
		{
			interfaceName: "Cacheable",
			definition: ast.NewImplementationDefinition(
				ast.NewIdentifier("Cacheable"),
				ast.Ty("String"),
				nil,
				nil,
				nil,
				nil,
				nil,
				false,
			),
		},
	}

	info := typeInfo{name: "String"}
	ok, err := interp.typeImplementsInterface(info, "Cacheable", nil, make(map[interfaceImplCacheKey]struct{}))
	if err != nil {
		t.Fatalf("typeImplementsInterface first lookup: %v", err)
	}
	if !ok {
		t.Fatalf("expected String to satisfy Cacheable on first lookup")
	}
	if len(interp.selectedInterfaceImplCache) == 0 {
		t.Fatalf("expected selected impl cache to store first lookup")
	}

	delete(interp.implMethods, "String")

	cachedEntry, err := interp.lookupImplEntry(info, "Cacheable", nil)
	if err != nil {
		t.Fatalf("lookupImplEntry cached lookup: %v", err)
	}
	if cachedEntry == nil || cachedEntry.entry == nil || cachedEntry.entry.interfaceName != "Cacheable" {
		t.Fatalf("expected cached selected impl before invalidation, got %#v", cachedEntry)
	}

	interp.invalidateMethodCache()
	if len(interp.selectedInterfaceImplCache) != 0 {
		t.Fatalf("expected selected impl cache clear on invalidate, got size=%d", len(interp.selectedInterfaceImplCache))
	}

	afterInvalidate, err := interp.lookupImplEntry(info, "Cacheable", nil)
	if err != nil {
		t.Fatalf("lookupImplEntry after invalidation: %v", err)
	}
	if afterInvalidate != nil {
		t.Fatalf("expected selected impl lookup to be recomputed as nil after invalidation, got %#v", afterInvalidate)
	}
}

func TestLookupImplEntry_SelectedImplCacheHitIsAllocationFree(t *testing.T) {
	interp := New()
	ifaceNode := ast.NewInterfaceDefinition(ast.NewIdentifier("Cacheable"), nil, nil, nil, nil, nil, false)
	interp.interfaces["Cacheable"] = &runtime.InterfaceDefinitionValue{Node: ifaceNode}
	typeParam := ast.GenericParam("T")
	interp.implMethods["Box"] = []implEntry{
		{
			interfaceName: "Cacheable",
			argTemplates:  []ast.TypeExpression{ast.Ty("T")},
			genericParams: []*ast.GenericParameter{typeParam},
			definition: ast.NewImplementationDefinition(
				ast.NewIdentifier("Cacheable"),
				ast.Gen(ast.Ty("Box"), ast.Ty("T")),
				nil,
				nil,
				[]*ast.GenericParameter{typeParam},
				nil,
				nil,
				false,
			),
		},
	}

	info := typeInfo{name: "Box", typeArgs: []ast.TypeExpression{ast.Ty("i32")}}
	first, err := interp.lookupImplEntry(info, "Cacheable", nil)
	if err != nil {
		t.Fatalf("lookupImplEntry first lookup: %v", err)
	}
	if first == nil {
		t.Fatalf("expected selected impl candidate on first lookup")
	}
	bound, ok := first.bindings["T"].(*ast.SimpleTypeExpression)
	if !ok || bound == nil || bound.Name == nil || bound.Name.Name != "i32" {
		t.Fatalf("expected T binding to resolve to i32, got %#v", first.bindings["T"])
	}

	allocs := testing.AllocsPerRun(1000, func() {
		cached, err := interp.lookupImplEntry(info, "Cacheable", nil)
		if err != nil {
			panic(err)
		}
		if cached == nil || cached.entry == nil || cached.entry.interfaceName != "Cacheable" {
			panic("expected cached selected impl candidate")
		}
		bound, ok := cached.bindings["T"].(*ast.SimpleTypeExpression)
		if !ok || bound == nil || bound.Name == nil || bound.Name.Name != "i32" {
			panic("expected cached T binding to stay at i32")
		}
	})
	if allocs != 0 {
		t.Fatalf("expected selected impl cache hit allocations to be zero, got %.2f", allocs)
	}
}

func TestCoerceToInterfaceValue_MethodDictionaryCacheInvalidatesAndKeepsSharedMethodsIsolated(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	ifaceNode := ast.NewInterfaceDefinition(
		ast.NewIdentifier("Cacheable"),
		[]*ast.FunctionSignature{
			ast.FnSig("probe", []*ast.FunctionParameter{ast.Param("self", ast.Ty("String"))}, ast.Ty("String"), nil, nil, nil),
		},
		nil,
		nil,
		nil,
		nil,
		false,
	)
	interp.interfaces["Cacheable"] = &runtime.InterfaceDefinitionValue{Node: ifaceNode, Env: env}

	first := cacheProbeFunction("cacheable_probe_impl_1", env)
	firstDef, ok := first.Declaration.(*ast.FunctionDefinition)
	if !ok {
		t.Fatalf("expected function definition declaration, got %T", first.Declaration)
	}
	second := cacheProbeFunction("cacheable_probe_impl_2", env)
	secondDef, ok := second.Declaration.(*ast.FunctionDefinition)
	if !ok {
		t.Fatalf("expected function definition declaration, got %T", second.Declaration)
	}

	interp.implMethods["String"] = []implEntry{
		{
			interfaceName: "Cacheable",
			methods: map[string]runtime.Value{
				"probe": first,
			},
			definition: ast.NewImplementationDefinition(
				ast.NewIdentifier("Cacheable"),
				ast.Ty("String"),
				[]*ast.FunctionDefinition{firstDef},
				nil,
				nil,
				nil,
				nil,
				false,
			),
		},
	}

	coercedA, err := interp.coerceToInterfaceValue("Cacheable", runtime.StringValue{Val: "alpha"}, nil)
	if err != nil {
		t.Fatalf("coerce first interface value: %v", err)
	}
	ifaceA, ok := coercedA.(*runtime.InterfaceValue)
	if !ok {
		t.Fatalf("expected interface value, got %T (%#v)", coercedA, coercedA)
	}
	methodA, ok := interfaceValueLookupMethod(ifaceA, "probe")
	if !ok || methodA != first {
		t.Fatalf("expected first cached probe method, got %T (%#v)", methodA, methodA)
	}
	if len(interp.interfaceMethodDictionaryCache) == 0 {
		t.Fatalf("expected interface method dictionary cache to store first coercion")
	}

	interfaceValueSetMethod(ifaceA, "mutated", second)

	interp.implMethods["String"] = []implEntry{
		{
			interfaceName: "Cacheable",
			methods: map[string]runtime.Value{
				"probe": second,
			},
			definition: ast.NewImplementationDefinition(
				ast.NewIdentifier("Cacheable"),
				ast.Ty("String"),
				[]*ast.FunctionDefinition{secondDef},
				nil,
				nil,
				nil,
				nil,
				false,
			),
		},
	}

	coercedB, err := interp.coerceToInterfaceValue("Cacheable", runtime.StringValue{Val: "beta"}, nil)
	if err != nil {
		t.Fatalf("coerce cached interface value: %v", err)
	}
	ifaceB, ok := coercedB.(*runtime.InterfaceValue)
	if !ok {
		t.Fatalf("expected interface value on cached coercion, got %T (%#v)", coercedB, coercedB)
	}
	if ifaceB.Methods != nil {
		t.Fatalf("expected cached interface value to defer owned method map allocation, got %#v", ifaceB.Methods)
	}
	if ifaceB.SharedMethods == nil {
		t.Fatalf("expected cached interface value to reuse shared method dictionary")
	}
	if method, ok := interfaceValueLookupMethod(ifaceB, "mutated"); ok || method != nil {
		t.Fatalf("expected cached interface methods to stay isolated per value")
	}
	if method, ok := interfaceValueLookupMethod(ifaceB, "probe"); !ok || method != first {
		t.Fatalf("expected cached probe method before invalidation, got %T (%#v)", method, method)
	}

	interp.invalidateMethodCache()
	if len(interp.interfaceMethodDictionaryCache) != 0 {
		t.Fatalf("expected interface method dictionary cache clear on invalidate, got size=%d", len(interp.interfaceMethodDictionaryCache))
	}

	coercedAfterInvalidate, err := interp.coerceToInterfaceValue("Cacheable", runtime.StringValue{Val: "gamma"}, nil)
	if err != nil {
		t.Fatalf("coerce interface value after invalidation: %v", err)
	}
	ifaceAfterInvalidate, ok := coercedAfterInvalidate.(*runtime.InterfaceValue)
	if !ok {
		t.Fatalf("expected interface value after invalidation, got %T (%#v)", coercedAfterInvalidate, coercedAfterInvalidate)
	}
	if method, ok := interfaceValueLookupMethod(ifaceAfterInvalidate, "probe"); !ok || method != second {
		t.Fatalf("expected recomputed probe method after invalidation, got %T (%#v)", method, method)
	}
}

func TestInterfaceMember_MemoizesBoundWrappersWithoutCloningSharedMethods(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	ifaceNode := ast.NewInterfaceDefinition(
		ast.NewIdentifier("Cacheable"),
		[]*ast.FunctionSignature{
			ast.FnSig("probe", []*ast.FunctionParameter{ast.Param("self", ast.Ty("String"))}, ast.Ty("String"), nil, nil, nil),
		},
		nil,
		nil,
		nil,
		nil,
		false,
	)
	interp.interfaces["Cacheable"] = &runtime.InterfaceDefinitionValue{Node: ifaceNode, Env: env}

	first := cacheProbeFunction("cacheable_probe_impl_1", env)
	firstDef, ok := first.Declaration.(*ast.FunctionDefinition)
	if !ok {
		t.Fatalf("expected function definition declaration, got %T", first.Declaration)
	}
	interp.implMethods["String"] = []implEntry{
		{
			interfaceName: "Cacheable",
			methods: map[string]runtime.Value{
				"probe": first,
			},
			definition: ast.NewImplementationDefinition(
				ast.NewIdentifier("Cacheable"),
				ast.Ty("String"),
				[]*ast.FunctionDefinition{firstDef},
				nil,
				nil,
				nil,
				nil,
				false,
			),
		},
	}

	if _, err := interp.coerceToInterfaceValue("Cacheable", runtime.StringValue{Val: "seed"}, nil); err != nil {
		t.Fatalf("seed interface cache: %v", err)
	}

	coerced, err := interp.coerceToInterfaceValue("Cacheable", runtime.StringValue{Val: "beta"}, nil)
	if err != nil {
		t.Fatalf("coerce cached interface value: %v", err)
	}
	iface, ok := coerced.(*runtime.InterfaceValue)
	if !ok {
		t.Fatalf("expected interface value, got %T (%#v)", coerced, coerced)
	}
	if iface.Methods != nil {
		t.Fatalf("expected cached interface value to defer local overlay allocation, got %#v", iface.Methods)
	}
	if iface.SharedMethods == nil {
		t.Fatalf("expected cached interface value to reuse shared method dictionary")
	}

	firstResolved, err := interp.interfaceMember(iface, ast.NewIdentifier("probe"))
	if err != nil {
		t.Fatalf("first interface member lookup: %v", err)
	}
	firstBound, ok := firstResolved.(runtime.BoundMethodValue)
	if !ok {
		t.Fatalf("expected bound method value, got %T (%#v)", firstResolved, firstResolved)
	}
	if methodFn, ok := firstBound.Method.(*runtime.FunctionValue); !ok || methodFn != first {
		t.Fatalf("expected first function pointer in bound method, got %T (%#v)", firstBound.Method, firstBound.Method)
	}
	if receiver, ok := firstBound.Receiver.(runtime.StringValue); !ok || receiver.Val != "beta" {
		t.Fatalf("expected beta receiver in bound method, got %T (%#v)", firstBound.Receiver, firstBound.Receiver)
	}
	if iface.Methods != nil {
		t.Fatalf("expected first bound lookup to avoid local method-map allocation, got %#v", iface.Methods)
	}
	if iface.SharedMethods == nil {
		t.Fatalf("expected first bound lookup to keep shared dictionary attached")
	}
	cachedBound, ok := interfaceValueLookupBoundMethod(iface, "probe")
	if !ok || !interfaceValueMethodIsBound(cachedBound) {
		t.Fatalf("expected dedicated bound-method cache to store bound method, got %T (%#v)", cachedBound, cachedBound)
	}
	if sharedMethod, ok := iface.SharedMethods["probe"]; !ok || sharedMethod != first {
		t.Fatalf("expected shared dictionary to keep raw method pointer, got %T (%#v)", sharedMethod, sharedMethod)
	}

	secondResolved, err := interp.interfaceMember(iface, ast.NewIdentifier("probe"))
	if err != nil {
		t.Fatalf("second interface member lookup: %v", err)
	}
	secondBound, ok := secondResolved.(runtime.BoundMethodValue)
	if !ok {
		t.Fatalf("expected bound method value on second lookup, got %T (%#v)", secondResolved, secondResolved)
	}
	if methodFn, ok := secondBound.Method.(*runtime.FunctionValue); !ok || methodFn != first {
		t.Fatalf("expected memoized bound method to keep first function pointer, got %T (%#v)", secondBound.Method, secondBound.Method)
	}
	if receiver, ok := secondBound.Receiver.(runtime.StringValue); !ok || receiver.Val != "beta" {
		t.Fatalf("expected beta receiver in second bound method, got %T (%#v)", secondBound.Receiver, secondBound.Receiver)
	}
}

func TestBindCallLocalTypeBindings_CachesAndInvalidates(t *testing.T) {
	interp := New()
	boxDef := ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("T"), "value"),
		},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	boxValue := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: boxDef},
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(1, runtime.IntegerI32),
		},
		TypeArguments: []ast.TypeExpression{ast.Ty("i32")},
	}
	fn := &runtime.FunctionValue{
		Declaration: ast.Fn(
			"box_value",
			[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Box"), ast.Ty("T")))},
			[]ast.Statement{ast.ID("self")},
			nil,
			nil,
			nil,
			false,
			false,
		),
		MethodSet: &runtime.MethodSet{
			TargetType:    ast.Gen(ast.Ty("Box"), ast.Ty("T")),
			GenericParams: []*ast.GenericParameter{ast.GenericParam("T")},
		},
		Closure: interp.GlobalEnvironment(),
	}

	envA := runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 4)
	interp.bindCallLocalTypeBindings(fn, boxValue, envA)
	if got := len(interp.callLocalTypeBindingCache); got != 1 {
		t.Fatalf("expected one call-local type binding cache entry after first bind, got %d", got)
	}
	if value, ok := envA.Lookup("T_type"); !ok {
		t.Fatalf("expected T_type binding in first env")
	} else if str, ok := value.(runtime.StringValue); !ok || str.Val != "i32" {
		t.Fatalf("expected T_type=i32, got %T (%#v)", value, value)
	}
	if value, ok := envA.Lookup("Self_type"); !ok {
		t.Fatalf("expected Self_type binding in first env")
	} else if str, ok := value.(runtime.StringValue); !ok || str.Val != "Box<i32>" {
		t.Fatalf("expected Self_type=Box<i32>, got %T (%#v)", value, value)
	}

	envB := runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 4)
	interp.bindCallLocalTypeBindings(fn, boxValue, envB)
	if got := len(interp.callLocalTypeBindingCache); got != 1 {
		t.Fatalf("expected call-local type binding cache reuse, got %d entries", got)
	}
	if value, ok := envB.Lookup("T"); !ok {
		t.Fatalf("expected cached T binding in second env")
	} else if ref, ok := value.(runtime.TypeRefValue); !ok || ref.TypeName != "i32" {
		t.Fatalf("expected cached T=i32 type ref, got %T (%#v)", value, value)
	}

	interp.invalidateMethodCache()
	if got := len(interp.callLocalTypeBindingCache); got != 0 {
		t.Fatalf("expected invalidateMethodCache to clear call-local type binding cache, got %d entries", got)
	}

	interp.bindCallLocalTypeBindings(fn, boxValue, runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 4))
	if got := len(interp.callLocalTypeBindingCache); got != 1 {
		t.Fatalf("expected cache to repopulate after invalidation, got %d entries", got)
	}

	interp.RegisterTypeAlias("MyInt", ast.NewTypeAliasDefinition(ast.NewIdentifier("MyInt"), ast.Ty("i32"), nil, nil, false))
	if got := len(interp.callLocalTypeBindingCache); got != 0 {
		t.Fatalf("expected RegisterTypeAlias to clear call-local type binding cache, got %d entries", got)
	}
}

func TestMakeInterfaceImplCacheKeySmallArgsHotPathIsAllocationFree(t *testing.T) {
	interp := New()
	info := typeInfo{name: "Box", typeArgs: []ast.TypeExpression{ast.Ty("i32")}}
	ifaceArgs := []ast.TypeExpression{ast.Ty("String")}

	allocs := testing.AllocsPerRun(1000, func() {
		key := interp.makeInterfaceImplCacheKey(info, "Cacheable", ifaceArgs)
		if key.typeName != "Box<i32>" || key.interfaceName != "Cacheable" || key.ifaceArgs.count != 1 {
			panic("unexpected interface impl cache key")
		}
	})
	if allocs != 0 {
		t.Fatalf("expected interface impl cache key hot path allocations to be zero, got %.2f", allocs)
	}
}

func TestMakeInterfaceMethodDictionaryCacheKeySmallArgsHotPathIsAllocationFree(t *testing.T) {
	interp := New()
	info := typeInfo{name: "Box", typeArgs: []ast.TypeExpression{ast.Ty("i32")}}
	ifaceArgs := []ast.TypeExpression{ast.Ty("String")}

	allocs := testing.AllocsPerRun(1000, func() {
		key := interp.makeInterfaceMethodDictionaryCacheKey(info, "Cacheable", ifaceArgs)
		if key.typeName != "Box<i32>" || key.interfaceName != "Cacheable" || key.ifaceArgs.count != 1 {
			panic("unexpected interface method dictionary cache key")
		}
	})
	if allocs != 0 {
		t.Fatalf("expected interface method dictionary cache key hot path allocations to be zero, got %.2f", allocs)
	}
}

func TestImplContextTargetMatchCacheUsesNominalReceiverTypeKey(t *testing.T) {
	interp := New()
	boxDef := &runtime.StructDefinitionValue{
		Node: ast.StructDef(
			"Box",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("T"), "value")},
			ast.StructKindNamed,
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			false,
		),
	}
	target := ast.Gen(ast.Ty("Box"), ast.Ty("T"))
	receiverA := &runtime.StructInstanceValue{
		Definition:    boxDef,
		Fields:        map[string]runtime.Value{"value": runtime.NewSmallInt(1, runtime.IntegerI32)},
		TypeArguments: []ast.TypeExpression{ast.Ty("i32")},
	}
	receiverB := &runtime.StructInstanceValue{
		Definition:    boxDef,
		Fields:        map[string]runtime.Value{"value": runtime.NewSmallInt(2, runtime.IntegerI32)},
		TypeArguments: []ast.TypeExpression{ast.Ty("i32")},
	}
	info, ok := interp.getTypeInfoForValue(receiverA)
	if !ok {
		t.Fatalf("expected type info for first receiver")
	}
	if !interp.implContextTargetMatchesReceiver(target, receiverA, info, true) {
		t.Fatalf("expected first receiver to match impl target")
	}
	if got := len(interp.implTargetMatchCache); got != 1 {
		t.Fatalf("expected first target match to populate one cache entry, got %d", got)
	}

	info, ok = interp.getTypeInfoForValue(receiverB)
	if !ok {
		t.Fatalf("expected type info for second receiver")
	}
	if !interp.implContextTargetMatchesReceiver(target, receiverB, info, true) {
		t.Fatalf("expected same nominal receiver type to match impl target")
	}
	if got := len(interp.implTargetMatchCache); got != 1 {
		t.Fatalf("expected same nominal receiver type to reuse cache entry, got %d", got)
	}

	interp.invalidateMethodCache()
	if got := len(interp.implTargetMatchCache); got != 0 {
		t.Fatalf("expected method cache invalidation to clear impl target match cache, got %d", got)
	}
}

func TestImplTargetMatchReceiverCacheableSkipsArrayBackedStructs(t *testing.T) {
	arrayDef := &runtime.StructDefinitionValue{
		Node: ast.StructDef("Array", nil, ast.StructKindNamed, nil, nil, false),
	}
	arrayStruct := &runtime.StructInstanceValue{Definition: arrayDef}
	if implTargetMatchReceiverCacheable(arrayStruct) {
		t.Fatalf("expected array-backed struct receiver to stay off impl target match cache")
	}
}

func TestResolveMethodFromPool_ImplContextTargetMatchCache(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	boxDef := &runtime.StructDefinitionValue{
		Node: ast.StructDef(
			"Box",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("T"), "value")},
			ast.StructKindNamed,
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			false,
		),
	}
	target := ast.Gen(ast.Ty("Box"), ast.Ty("i32"))
	method := &runtime.FunctionValue{
		Declaration: ast.Fn(
			"probe",
			[]*ast.FunctionParameter{ast.Param("self", nil)},
			[]ast.Statement{ast.ID("self")},
			target,
			nil,
			nil,
			false,
			false,
		),
		Closure: env,
	}
	ctx := &implMethodContext{
		implName:      "CacheImpl",
		interfaceName: "CacheIface",
		target:        target,
		methods:       map[string]runtime.Value{"probe": method},
	}
	callEnv := runtime.NewEnvironment(env)
	callEnv.SetRuntimeData(ctx)
	receiver := &runtime.StructInstanceValue{
		Definition: boxDef,
		Fields:     map[string]runtime.Value{"value": runtime.NewSmallInt(1, runtime.IntegerI32)},
	}

	callable, found, err := interp.resolveMethodCallableFromPool(callEnv, "probe", receiver, "")
	if err != nil {
		t.Fatalf("resolve impl method: %v", err)
	}
	if !found || callable != method {
		t.Fatalf("expected impl method callable, found=%v callable=%T (%#v)", found, callable, callable)
	}
	if got := len(interp.implTargetMatchCache); got != 1 {
		t.Fatalf("expected impl target match cache entry after resolve, got %d", got)
	}
	if len(receiver.TypeArguments) != 1 || typeExpressionToString(receiver.TypeArguments[0]) != "i32" {
		t.Fatalf("expected concrete impl target match to infer receiver type arguments, got %#v", receiver.TypeArguments)
	}
}

func TestResolveMethodFromPool_ImplContextOpenGenericTargetSkipsReceiverTypeInference(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	boxDef := &runtime.StructDefinitionValue{
		Node: ast.StructDef(
			"Box",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("T"), "value")},
			ast.StructKindNamed,
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			false,
		),
	}
	target := ast.Gen(ast.Ty("Box"), ast.Ty("T"))
	method := &runtime.FunctionValue{
		Declaration: ast.Fn(
			"probe",
			[]*ast.FunctionParameter{ast.Param("self", nil)},
			[]ast.Statement{ast.ID("self")},
			target,
			nil,
			nil,
			false,
			false,
		),
		Closure: env,
	}
	ctx := &implMethodContext{
		implName:      "CacheImpl",
		interfaceName: "CacheIface",
		target:        target,
		methods:       map[string]runtime.Value{"probe": method},
	}
	callEnv := runtime.NewEnvironment(env)
	callEnv.SetRuntimeData(ctx)
	receiver := &runtime.StructInstanceValue{
		Definition: boxDef,
		Fields:     map[string]runtime.Value{"value": runtime.NewSmallInt(1, runtime.IntegerI32)},
	}

	callable, found, err := interp.resolveMethodCallableFromPool(callEnv, "probe", receiver, "")
	if err != nil {
		t.Fatalf("resolve impl method: %v", err)
	}
	if !found || callable != method {
		t.Fatalf("expected impl method callable, found=%v callable=%T (%#v)", found, callable, callable)
	}
	if got := len(interp.implTargetMatchCache); got != 0 {
		t.Fatalf("expected open generic target fast path not to populate target match cache, got %d", got)
	}
	if len(receiver.TypeArguments) != 0 {
		t.Fatalf("expected open generic impl target match to avoid receiver type inference, got %#v", receiver.TypeArguments)
	}
}
