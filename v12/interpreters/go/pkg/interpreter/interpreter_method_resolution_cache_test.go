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
	ok, err := interp.typeImplementsInterface(info, "Cacheable", nil, make(map[string]struct{}))
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

	cachedOK, err := interp.typeImplementsInterface(info, "Cacheable", nil, make(map[string]struct{}))
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

	afterInvalidate, err := interp.typeImplementsInterface(info, "Cacheable", nil, make(map[string]struct{}))
	if err != nil {
		t.Fatalf("typeImplementsInterface after invalidation: %v", err)
	}
	if afterInvalidate {
		t.Fatalf("expected String -> Cacheable to be recomputed as false after invalidation")
	}
}
