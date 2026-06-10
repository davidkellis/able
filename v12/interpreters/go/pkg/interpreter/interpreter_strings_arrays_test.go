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

func TestToArrayValueReadsPositionalArrayStructMetadata(t *testing.T) {
	interp := New()
	backing := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(65, runtime.IntegerU8),
		runtime.NewSmallInt(98, runtime.IntegerU8),
		runtime.NewSmallInt(108, runtime.IntegerU8),
		runtime.NewSmallInt(101, runtime.IntegerU8),
	}, 4)
	definition := &runtime.StructDefinitionValue{Node: ast.StructDef("Array", []*ast.StructFieldDefinition{
		ast.FieldDef(ast.Ty("i32"), "length"),
		ast.FieldDef(ast.Ty("i32"), "capacity"),
		ast.FieldDef(ast.Ty("ArrayHandle"), "storage_handle"),
	}, ast.StructKindNamed, nil, nil, false)}
	value := &runtime.StructInstanceValue{
		Definition: definition,
		Positional: []runtime.Value{
			runtime.NewSmallInt(4, runtime.IntegerI32),
			runtime.NewSmallInt(4, runtime.IntegerI32),
			runtime.NewSmallInt(backing.Handle, runtime.IntegerI64),
		},
	}

	array, err := interp.toArrayValue(value)
	if err != nil {
		t.Fatalf("toArrayValue: %v", err)
	}
	elements, err := interp.ArrayElements(array)
	if err != nil {
		t.Fatalf("ArrayElements: %v", err)
	}
	if len(elements) != 4 {
		t.Fatalf("array length = %d, want 4", len(elements))
	}
	for idx, want := range []int64{65, 98, 108, 101} {
		got, ok := elements[idx].(runtime.IntegerValue)
		if !ok || got.Int64Fast() != want {
			t.Fatalf("array element %d = %#v, want %d_u8", idx, elements[idx], want)
		}
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

func TestArrayMemberWithOverrides_DirectMetadataKeepsMonoHandleBackedArray(t *testing.T) {
	interp := New()
	interp.ensureArrayBuiltins()
	handle := runtime.ArrayStoreMonoNewWithCapacityI32(8)
	if err := runtime.ArrayStoreMonoWriteI32(handle, 0, 11); err != nil {
		t.Fatalf("seed mono i32 array: %v", err)
	}
	arr, _, err := runtime.ArrayStoreValueViewFromHandle(handle, 1, 8)
	if err != nil {
		t.Fatalf("create handle-backed array view: %v", err)
	}
	if arr.State != nil {
		t.Fatalf("expected mono handle-backed array to start without tracked state")
	}

	lengthVal, err := interp.arrayMemberWithOverrides(arr, ast.NewIdentifier("length"), interp.GlobalEnvironment(), false)
	if err != nil {
		t.Fatalf("resolve direct mono-backed array length: %v", err)
	}
	length, err := arrayIndexFromValue(lengthVal)
	if err != nil || length != 1 {
		t.Fatalf("mono-backed array length = %#v (%v), want 1", lengthVal, err)
	}

	capacityVal, err := interp.arrayMemberWithOverrides(arr, ast.NewIdentifier("capacity"), interp.GlobalEnvironment(), false)
	if err != nil {
		t.Fatalf("resolve direct mono-backed array capacity: %v", err)
	}
	capacity, err := arrayIndexFromValue(capacityVal)
	if err != nil || capacity != 8 {
		t.Fatalf("mono-backed array capacity = %#v (%v), want 8", capacityVal, err)
	}

	if arr.State != nil {
		t.Fatalf("direct metadata access should not materialize dynamic array state")
	}
	got, err := runtime.ArrayStoreMonoReadI32(handle, 0)
	if err != nil {
		t.Fatalf("mono i32 array should remain typed after direct metadata access: %v", err)
	}
	if got != 11 {
		t.Fatalf("mono i32 array value = %d, want 11", got)
	}
}

func TestArrayMemberWithOverrides_MethodLookupKeepsMonoHandleBackedArray(t *testing.T) {
	interp := New()
	interp.ensureArrayBuiltins()
	env := interp.GlobalEnvironment()
	handle := runtime.ArrayStoreMonoNewWithCapacityI32(4)
	arr, _, err := runtime.ArrayStoreValueViewFromHandle(handle, 0, 4)
	if err != nil {
		t.Fatalf("create handle-backed array view: %v", err)
	}
	if arr.State != nil {
		t.Fatalf("expected mono handle-backed array to start without tracked state")
	}

	bucket := interp.inherentMethods["Array"]
	if bucket == nil {
		bucket = make(map[string]runtime.Value)
		interp.inherentMethods["Array"] = bucket
	}
	expected := cacheProbeFunction("cache_probe_impl", env)
	bucket["cache_probe"] = expected

	resolved, err := interp.arrayMemberWithOverrides(arr, ast.NewIdentifier("cache_probe"), env, false)
	if err != nil {
		t.Fatalf("resolve array helper method on mono-backed array: %v", err)
	}
	bound, ok := resolved.(runtime.BoundMethodValue)
	if !ok {
		t.Fatalf("expected bound method value, got %T (%#v)", resolved, resolved)
	}
	methodFn, ok := bound.Method.(*runtime.FunctionValue)
	if !ok || methodFn != expected {
		t.Fatalf("expected helper method binding, got %T (%#v)", bound.Method, bound.Method)
	}
	if arr.State != nil {
		t.Fatalf("method lookup should not materialize dynamic array state")
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
