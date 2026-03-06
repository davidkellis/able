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
