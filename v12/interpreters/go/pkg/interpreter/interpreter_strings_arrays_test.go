package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestArrayHelpersRequireStdlib(t *testing.T) {
	interp := New()
	interp.ensureArrayBuiltins()
	arr := &runtime.ArrayValue{Elements: []runtime.Value{runtime.NilValue{}}}

	iterVal, err := interp.arrayMemberWithOverrides(arr, ast.NewIdentifier("iterator"), interp.GlobalEnvironment(), false)
	if err != nil {
		t.Fatalf("iterator should remain available without stdlib: %v", err)
	}
	if _, ok := iterVal.(*runtime.NativeBoundMethodValue); !ok {
		t.Fatalf("iterator should bind to native method, got %T", iterVal)
	}
}

func TestArrayMemberWithOverrides_PrefersDirectMembers(t *testing.T) {
	interp := New()
	interp.ensureArrayBuiltins()
	env := interp.GlobalEnvironment()
	arr := interp.newArrayValue([]runtime.Value{runtime.NilValue{}}, 0)

	bucket := interp.inherentMethods["Array"]
	if bucket == nil {
		bucket = make(map[string]runtime.Value)
		interp.inherentMethods["Array"] = bucket
	}
	bucket["length"] = cacheProbeFunction("array_length_override", env)

	resolved, err := interp.arrayMemberWithOverrides(arr, ast.NewIdentifier("length"), env, false)
	if err != nil {
		t.Fatalf("resolve direct array length: %v", err)
	}
	length, err := arrayIndexFromValue(resolved)
	if err != nil {
		t.Fatalf("array length should stay a direct integer member, got %T (%#v)", resolved, resolved)
	}
	if length != 1 {
		t.Fatalf("array length = %d, want 1", length)
	}
}

func TestArrayMemberWithOverrides_UsesMethodLookupForNonDirectNames(t *testing.T) {
	interp := New()
	interp.ensureArrayBuiltins()
	env := interp.GlobalEnvironment()
	arr := interp.newArrayValue(nil, 0)

	bucket := interp.inherentMethods["Array"]
	if bucket == nil {
		bucket = make(map[string]runtime.Value)
		interp.inherentMethods["Array"] = bucket
	}
	expected := cacheProbeFunction("cache_probe_impl", env)
	bucket["cache_probe"] = expected

	resolved, err := interp.arrayMemberWithOverrides(arr, ast.NewIdentifier("cache_probe"), env, false)
	if err != nil {
		t.Fatalf("resolve array helper method: %v", err)
	}
	bound, ok := resolved.(runtime.BoundMethodValue)
	if !ok {
		t.Fatalf("expected bound method value, got %T (%#v)", resolved, resolved)
	}
	methodFn, ok := bound.Method.(*runtime.FunctionValue)
	if !ok || methodFn != expected {
		t.Fatalf("expected helper method binding, got %T (%#v)", bound.Method, bound.Method)
	}
}

func TestArrayMemberCachesLargeLengthMetadataBoxing(t *testing.T) {
	interp := New()
	interp.ensureArrayBuiltins()
	arr := interp.newArrayValue(make([]runtime.Value, 20000), 20000)

	if _, err := interp.arrayMember(arr, ast.NewIdentifier("length")); err != nil {
		t.Fatalf("prime cached length metadata: %v", err)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		val, err := interp.arrayMember(arr, ast.NewIdentifier("length"))
		if err != nil {
			t.Fatalf("resolve cached array length: %v", err)
		}
		if _, err := arrayIndexFromValue(val); err != nil {
			t.Fatalf("cached array length should stay an integer member, got %T (%#v)", val, val)
		}
	})
	if allocs > 0.1 {
		t.Fatalf("unexpected large-length metadata allocations: got %.2f want <= 0.1", allocs)
	}
}

func TestArraySizeBuiltinUsesSharedLargeMetadataBoxing(t *testing.T) {
	interp := New()
	interp.ensureArrayBuiltins()
	global := interp.GlobalEnvironment()

	sizeVal, err := global.Get("__able_array_size")
	if err != nil {
		t.Fatalf("lookup __able_array_size: %v", err)
	}

	var sizeFn runtime.NativeFunctionValue
	switch fn := sizeVal.(type) {
	case runtime.NativeFunctionValue:
		sizeFn = fn
	case *runtime.NativeFunctionValue:
		sizeFn = *fn
	default:
		t.Fatalf("__able_array_size type = %T, want runtime.NativeFunctionValue", sizeVal)
	}

	handle := runtime.ArrayStoreNewWithCapacity(20000)
	if err := runtime.ArrayStoreSetLength(handle, 20000); err != nil {
		t.Fatalf("seed array store length: %v", err)
	}
	ctx := &runtime.NativeCallContext{Env: global}
	args := []runtime.Value{runtime.NewSmallInt(handle, runtime.IntegerI64)}

	allocs := testing.AllocsPerRun(1000, func() {
		got, err := sizeFn.Impl(ctx, args)
		if err != nil {
			t.Fatalf("__able_array_size call failed: %v", err)
		}
		intVal, ok := got.(runtime.IntegerValue)
		if !ok {
			t.Fatalf("__able_array_size type = %T, want runtime.IntegerValue", got)
		}
		if intVal.Int64Fast() != 20000 || intVal.TypeSuffix != runtime.IntegerU64 {
			t.Fatalf("__able_array_size = (%d, %s), want (%d, %s)", intVal.Int64Fast(), intVal.TypeSuffix, 20000, runtime.IntegerU64)
		}
	})
	if allocs > 0.1 {
		t.Fatalf("unexpected large-size metadata allocations: got %.2f want <= 0.1", allocs)
	}
}

func TestStringHelpersRequireStdlib(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()

	if _, err := interp.stringMemberWithOverrides(runtime.StringValue{Val: "hello"}, ast.NewIdentifier("split"), env); err == nil {
		t.Fatalf("expected split to be unavailable without stdlib import")
	}
}
