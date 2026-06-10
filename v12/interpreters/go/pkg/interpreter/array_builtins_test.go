package interpreter

import (
	"fmt"
	"math/big"
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestArrayBuiltins(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	val, err := global.Get("Array")
	if err != nil {
		t.Fatalf("Array package not registered: %v", err)
	}

	var pkg *runtime.PackageValue
	switch v := val.(type) {
	case *runtime.PackageValue:
		pkg = v
	case runtime.PackageValue:
		pkg = &v
	default:
		t.Fatalf("unexpected Array binding type %T", val)
	}

	newSym, ok := pkg.Public["new"]
	if !ok {
		t.Fatalf("Array.new missing from package")
	}

	var newFn runtime.NativeFunctionValue
	switch fn := newSym.(type) {
	case runtime.NativeFunctionValue:
		newFn = fn
	case *runtime.NativeFunctionValue:
		newFn = *fn
	default:
		t.Fatalf("Array.new unexpected type %T", newSym)
	}

	ctx := &runtime.NativeCallContext{Env: global}
	arrayVal, err := newFn.Impl(ctx, nil)
	if err != nil {
		t.Fatalf("Array.new call failed: %v", err)
	}

	arr, ok := arrayVal.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("Array.new did not return array, got %T", arrayVal)
	}

	if len(arr.Elements) != 0 {
		t.Fatalf("expected new array to be empty")
	}

	capVal, err := newFn.Impl(ctx, []runtime.Value{runtime.IntegerValue{Val: big.NewInt(8), TypeSuffix: runtime.IntegerI32}})
	if err != nil {
		t.Fatalf("Array.new(8) failed: %v", err)
	}
	arrCap, ok := capVal.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("Array.new(8) did not return array, got %T", capVal)
	}
	if cap(arrCap.Elements) < 8 {
		t.Fatalf("Array.new should respect capacity, got %d", cap(arrCap.Elements))
	}
}

func TestArrayWithCapacityUsesReservedBacking(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()
	sym, err := global.Get("__able_array_with_capacity")
	if err != nil {
		t.Fatalf("__able_array_with_capacity missing: %v", err)
	}
	fn, ok := sym.(runtime.NativeFunctionValue)
	if !ok {
		t.Fatalf("__able_array_with_capacity unexpected type %T", sym)
	}
	handleValue, err := fn.Impl(&runtime.NativeCallContext{Env: global}, []runtime.Value{
		runtime.IntegerValue{Val: big.NewInt(8), TypeSuffix: runtime.IntegerI32},
	})
	if err != nil {
		t.Fatalf("__able_array_with_capacity call failed: %v", err)
	}
	handleInt, ok := handleValue.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("__able_array_with_capacity returned %T", handleValue)
	}
	handle, ok := handleInt.ToInt64()
	if !ok {
		t.Fatalf("__able_array_with_capacity handle out of range: %#v", handleInt)
	}
	capacity, err := runtime.ArrayStoreCapacity(handle)
	if err != nil {
		t.Fatalf("ArrayStoreCapacity: %v", err)
	}
	if capacity != 8 {
		t.Fatalf("reserved capacity = %d, want 8", capacity)
	}
	state, err := runtime.ArrayStoreState(handle)
	if err != nil {
		t.Fatalf("ArrayStoreState: %v", err)
	}
	if len(state.Values) != 0 || cap(state.Values) != 0 {
		t.Fatalf("reserved backing len=%d cap=%d, want len 0 cap 0", len(state.Values), cap(state.Values))
	}
}

func TestArrayBuiltinsUseMonoCharHandlesWhenGenericTypeIsBound(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()
	env := runtime.NewEnvironmentWithValueCapacity(global, 1)
	env.Define("T", runtime.TypeRefValue{TypeName: "char"})

	withCapacitySym, err := global.Get("__able_array_with_capacity")
	if err != nil {
		t.Fatalf("__able_array_with_capacity missing: %v", err)
	}
	withCapacityFn, ok := withCapacitySym.(runtime.NativeFunctionValue)
	if !ok {
		t.Fatalf("__able_array_with_capacity unexpected type %T", withCapacitySym)
	}
	handleValue, err := withCapacityFn.Impl(&runtime.NativeCallContext{Env: env}, []runtime.Value{
		runtime.IntegerValue{Val: big.NewInt(4), TypeSuffix: runtime.IntegerI32},
	})
	if err != nil {
		t.Fatalf("__able_array_with_capacity(char) failed: %v", err)
	}
	handleInt, ok := handleValue.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("__able_array_with_capacity(char) returned %T", handleValue)
	}
	handle, ok := handleInt.ToInt64()
	if !ok {
		t.Fatalf("__able_array_with_capacity(char) handle out of range: %#v", handleInt)
	}
	if err := runtime.ArrayStoreWrite(handle, 0, runtime.CharValue{Val: 'z'}); err != nil {
		t.Fatalf("ArrayStoreWrite mono char handle: %v", err)
	}
	raw, ok, err := runtime.ArrayStoreMonoReadCharIfAvailable(handle, 0)
	if err != nil {
		t.Fatalf("ArrayStoreMonoReadCharIfAvailable: %v", err)
	}
	if !ok || raw != 'z' {
		t.Fatalf("mono char handle read = (%q, %v), want ('z', true)", raw, ok)
	}
	arr, err := interp.arrayValueFromHandle(handle, 0, 4)
	if err != nil {
		t.Fatalf("arrayValueFromHandle mono char: %v", err)
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("mono char handle view should not materialize boxed state")
	}

	arraySym, err := global.Get("Array")
	if err != nil {
		t.Fatalf("Array package not registered: %v", err)
	}
	var pkg *runtime.PackageValue
	switch typed := arraySym.(type) {
	case *runtime.PackageValue:
		pkg = typed
	case runtime.PackageValue:
		pkg = &typed
	default:
		t.Fatalf("Array binding type = %T, want runtime.PackageValue", arraySym)
	}
	newSym, ok := pkg.Public["new"]
	if !ok {
		t.Fatalf("Array.new missing from package")
	}
	newFn, ok := newSym.(runtime.NativeFunctionValue)
	if !ok {
		t.Fatalf("Array.new unexpected type %T", newSym)
	}
	arrayValue, err := newFn.Impl(&runtime.NativeCallContext{Env: env}, nil)
	if err != nil {
		t.Fatalf("Array.new(char) failed: %v", err)
	}
	monoArr, ok := arrayValue.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("Array.new(char) returned %T", arrayValue)
	}
	if monoArr.State != nil || monoArr.Elements != nil {
		t.Fatalf("Array.new(char) should preserve mono handle view")
	}
	if err := runtime.ArrayStoreWrite(monoArr.Handle, 0, runtime.CharValue{Val: 'q'}); err != nil {
		t.Fatalf("Array.new(char) write: %v", err)
	}
	raw, ok, err = runtime.ArrayStoreMonoReadCharIfAvailable(monoArr.Handle, 0)
	if err != nil {
		t.Fatalf("Array.new(char) mono read: %v", err)
	}
	if !ok || raw != 'q' {
		t.Fatalf("Array.new(char) mono read = (%q, %v), want ('q', true)", raw, ok)
	}
}

func TestArrayBuiltinsUseMonoUnsignedHandlesWhenGenericTypeIsBound(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		write    runtime.Value
		verify   func(int64) error
	}{
		{
			name:     "u32",
			typeName: "u32",
			write:    runtime.NewSmallInt(7, runtime.IntegerU32),
			verify: func(handle int64) error {
				raw, ok, err := runtime.ArrayStoreMonoReadU32IfAvailable(handle, 0)
				if err != nil {
					return err
				}
				if !ok || raw != 7 {
					return fmt.Errorf("mono u32 read = (%d, %v), want (7, true)", raw, ok)
				}
				return nil
			},
		},
		{
			name:     "u64",
			typeName: "u64",
			write:    runtime.NewBigIntValue(new(big.Int).SetUint64(uint64(1)<<63+9), runtime.IntegerU64),
			verify: func(handle int64) error {
				raw, ok, err := runtime.ArrayStoreMonoReadU64IfAvailable(handle, 0)
				if err != nil {
					return err
				}
				if !ok || raw != uint64(1)<<63+9 {
					return fmt.Errorf("mono u64 read = (%d, %v), want (%d, true)", raw, ok, uint64(1)<<63+9)
				}
				return nil
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			interp := New()
			global := interp.GlobalEnvironment()
			env := runtime.NewEnvironmentWithValueCapacity(global, 1)
			env.Define("T", runtime.TypeRefValue{TypeName: tc.typeName})

			withCapacitySym, err := global.Get("__able_array_with_capacity")
			if err != nil {
				t.Fatalf("__able_array_with_capacity missing: %v", err)
			}
			withCapacityFn, ok := withCapacitySym.(runtime.NativeFunctionValue)
			if !ok {
				t.Fatalf("__able_array_with_capacity unexpected type %T", withCapacitySym)
			}
			handleValue, err := withCapacityFn.Impl(&runtime.NativeCallContext{Env: env}, []runtime.Value{
				runtime.IntegerValue{Val: big.NewInt(2), TypeSuffix: runtime.IntegerI32},
			})
			if err != nil {
				t.Fatalf("__able_array_with_capacity(%s) failed: %v", tc.typeName, err)
			}
			handleInt, ok := handleValue.(runtime.IntegerValue)
			if !ok {
				t.Fatalf("__able_array_with_capacity(%s) returned %T", tc.typeName, handleValue)
			}
			handle, ok := handleInt.ToInt64()
			if !ok {
				t.Fatalf("__able_array_with_capacity(%s) handle out of range: %#v", tc.typeName, handleInt)
			}
			if err := runtime.ArrayStoreWrite(handle, 0, tc.write); err != nil {
				t.Fatalf("ArrayStoreWrite mono %s handle: %v", tc.typeName, err)
			}
			if err := tc.verify(handle); err != nil {
				t.Fatalf("mono %s handle verification failed: %v", tc.typeName, err)
			}

			arraySym, err := global.Get("Array")
			if err != nil {
				t.Fatalf("Array package not registered: %v", err)
			}
			var pkg *runtime.PackageValue
			switch typed := arraySym.(type) {
			case *runtime.PackageValue:
				pkg = typed
			case runtime.PackageValue:
				pkg = &typed
			default:
				t.Fatalf("Array binding type = %T, want runtime.PackageValue", arraySym)
			}
			newSym, ok := pkg.Public["new"]
			if !ok {
				t.Fatalf("Array.new missing from package")
			}
			newFn, ok := newSym.(runtime.NativeFunctionValue)
			if !ok {
				t.Fatalf("Array.new unexpected type %T", newSym)
			}
			arrayValue, err := newFn.Impl(&runtime.NativeCallContext{Env: env}, nil)
			if err != nil {
				t.Fatalf("Array.new(%s) failed: %v", tc.typeName, err)
			}
			arr, ok := arrayValue.(*runtime.ArrayValue)
			if !ok {
				t.Fatalf("Array.new(%s) returned %T", tc.typeName, arrayValue)
			}
			if arr.State != nil || arr.Elements != nil {
				t.Fatalf("Array.new(%s) should preserve mono handle view", tc.typeName)
			}
		})
	}
}

func TestArrayWithCapacityFromSourcePreservesMonoU32Handle(t *testing.T) {
	program := mustLoadAbleProgramFromSource(t, `
import able.kernel.{Array}

fn main() -> Array u32 {
  Array.with_capacity(4)
}

main()
`)

	assertMonoU32Array := func(t *testing.T, value runtime.Value, mode string) {
		t.Helper()

		arr, ok := value.(*runtime.ArrayValue)
		if !ok {
			t.Fatalf("%s returned %T, want *runtime.ArrayValue", mode, value)
		}
		if arr.State != nil || arr.Elements != nil {
			t.Fatalf("%s should preserve mono handle view", mode)
		}
		size, err := runtime.ArrayStoreSize(arr.Handle)
		if err != nil {
			t.Fatalf("%s ArrayStoreSize: %v", mode, err)
		}
		if size != 0 {
			t.Fatalf("%s size = %d, want 0", mode, size)
		}
		capacity, err := runtime.ArrayStoreCapacity(arr.Handle)
		if err != nil {
			t.Fatalf("%s ArrayStoreCapacity: %v", mode, err)
		}
		if capacity != 4 {
			t.Fatalf("%s capacity = %d, want 4", mode, capacity)
		}
		typeName, ok, err := runtime.ArrayStoreMonoElementTypeNameIfKnown(arr.Handle)
		if err != nil {
			t.Fatalf("%s ArrayStoreMonoElementTypeNameIfKnown: %v", mode, err)
		}
		if !ok || typeName != string(runtime.IntegerU32) {
			t.Fatalf("%s mono element type = (%q, %v), want (u32, true)", mode, typeName, ok)
		}
	}

	treeValue, _, _, err := New().EvaluateProgram(program, ProgramEvaluationOptions{})
	if err != nil {
		t.Fatalf("tree evaluation failed: %v", err)
	}
	assertMonoU32Array(t, treeValue, "tree")

	bytecodeValue, _, _, err := NewBytecode().EvaluateProgram(program, ProgramEvaluationOptions{})
	if err != nil {
		t.Fatalf("bytecode evaluation failed: %v", err)
	}
	assertMonoU32Array(t, bytecodeValue, "bytecode")
}

func TestArrayWithCapacityFromSourcePreservesMonoPrimitiveHandles(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
	}{
		{name: "i32", typeName: "i32"},
		{name: "i64", typeName: "i64"},
		{name: "bool", typeName: "bool"},
		{name: "char", typeName: "char"},
		{name: "u8", typeName: "u8"},
		{name: "u64", typeName: "u64"},
		{name: "f64", typeName: "f64"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			program := mustLoadAbleProgramFromSource(t, fmt.Sprintf(`
import able.kernel.{Array}

fn main() -> Array %s {
  Array.with_capacity(4)
}

main()
`, tc.typeName))

			assertMonoArray := func(t *testing.T, value runtime.Value, mode string) {
				t.Helper()

				arr, ok := value.(*runtime.ArrayValue)
				if !ok {
					t.Fatalf("%s returned %T, want *runtime.ArrayValue", mode, value)
				}
				if arr.State != nil || arr.Elements != nil {
					t.Fatalf("%s should preserve mono handle view", mode)
				}
				size, err := runtime.ArrayStoreSize(arr.Handle)
				if err != nil {
					t.Fatalf("%s ArrayStoreSize: %v", mode, err)
				}
				if size != 0 {
					t.Fatalf("%s size = %d, want 0", mode, size)
				}
				capacity, err := runtime.ArrayStoreCapacity(arr.Handle)
				if err != nil {
					t.Fatalf("%s ArrayStoreCapacity: %v", mode, err)
				}
				if capacity != 4 {
					t.Fatalf("%s capacity = %d, want 4", mode, capacity)
				}
				typeName, ok, err := runtime.ArrayStoreMonoElementTypeNameIfKnown(arr.Handle)
				if err != nil {
					t.Fatalf("%s ArrayStoreMonoElementTypeNameIfKnown: %v", mode, err)
				}
				if !ok || typeName != tc.typeName {
					t.Fatalf("%s mono element type = (%q, %v), want (%s, true)", mode, typeName, ok, tc.typeName)
				}
			}

			treeValue, _, _, err := New().EvaluateProgram(program, ProgramEvaluationOptions{})
			if err != nil {
				t.Fatalf("tree evaluation failed: %v", err)
			}
			assertMonoArray(t, treeValue, "tree")

			bytecodeValue, _, _, err := NewBytecode().EvaluateProgram(program, ProgramEvaluationOptions{})
			if err != nil {
				t.Fatalf("bytecode evaluation failed: %v", err)
			}
			assertMonoArray(t, bytecodeValue, "bytecode")
		})
	}
}

func TestArrayWithCapacityFromSourceKeepsDynamicHandlesForNonPrimitiveArrays(t *testing.T) {
	tests := []struct {
		name       string
		returnType string
		body       string
	}{
		{
			name:       "string",
			returnType: "Array String",
			body: `values: Array String = Array.with_capacity(4)
  values.push("able")
  values`,
		},
		{
			name:       "nested_array",
			returnType: "Array (Array i32)",
			body: `values: Array (Array i32) = Array.with_capacity(4)
  values.push([1, 2, 3])
  values`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			program := mustLoadAbleProgramFromSource(t, fmt.Sprintf(`
import able.kernel.{Array}

fn main() -> %s {
  %s
}

main()
`, tc.returnType, tc.body))

			assertDynamicArray := func(t *testing.T, value runtime.Value, mode string) {
				t.Helper()

				arr, ok := value.(*runtime.ArrayValue)
				if !ok {
					t.Fatalf("%s returned %T, want *runtime.ArrayValue", mode, value)
				}
				typeName, ok, err := runtime.ArrayStoreMonoElementTypeNameIfKnown(arr.Handle)
				if err != nil {
					t.Fatalf("%s ArrayStoreMonoElementTypeNameIfKnown: %v", mode, err)
				}
				if ok {
					t.Fatalf("%s mono element type = %q, want dynamic handle", mode, typeName)
				}
				size, err := runtime.ArrayStoreSize(arr.Handle)
				if err != nil {
					t.Fatalf("%s ArrayStoreSize: %v", mode, err)
				}
				if size != 1 {
					t.Fatalf("%s size = %d, want 1", mode, size)
				}
				capacity, err := runtime.ArrayStoreCapacity(arr.Handle)
				if err != nil {
					t.Fatalf("%s ArrayStoreCapacity: %v", mode, err)
				}
				if capacity != 4 {
					t.Fatalf("%s capacity = %d, want 4", mode, capacity)
				}
			}

			treeValue, _, _, err := New().EvaluateProgram(program, ProgramEvaluationOptions{})
			if err != nil {
				t.Fatalf("tree evaluation failed: %v", err)
			}
			assertDynamicArray(t, treeValue, "tree")

			bytecodeValue, _, _, err := NewBytecode().EvaluateProgram(program, ProgramEvaluationOptions{})
			if err != nil {
				t.Fatalf("bytecode evaluation failed: %v", err)
			}
			assertDynamicArray(t, bytecodeValue, "bytecode")
		})
	}
}
