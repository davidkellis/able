package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter10-go/pkg/runtime"
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
