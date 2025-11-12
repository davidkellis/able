package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter10-go/pkg/ast"
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

	// push elements
	pushVal, err := interp.arrayMember(arr, ast.NewIdentifier("push"))
	if err != nil {
		t.Fatalf("array push lookup failed: %v", err)
	}
	push := pushVal.(*runtime.NativeBoundMethodValue)
	intOne := runtime.IntegerValue{Val: big.NewInt(1), TypeSuffix: runtime.IntegerI32}
	intTwo := runtime.IntegerValue{Val: big.NewInt(2), TypeSuffix: runtime.IntegerI32}

	if _, err := push.Method.Impl(ctx, []runtime.Value{push.Receiver, intOne}); err != nil {
		t.Fatalf("push(1) failed: %v", err)
	}
	if _, err := push.Method.Impl(ctx, []runtime.Value{push.Receiver, intTwo}); err != nil {
		t.Fatalf("push(2) failed: %v", err)
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("push did not append elements, length=%d", len(arr.Elements))
	}

	// size check
	sizeVal, err := interp.arrayMember(arr, ast.NewIdentifier("size"))
	if err != nil {
		t.Fatalf("array size lookup failed: %v", err)
	}
	size := sizeVal.(*runtime.NativeBoundMethodValue)
	sizeResult, err := size.Method.Impl(ctx, []runtime.Value{size.Receiver})
	if err != nil {
		t.Fatalf("size() failed: %v", err)
	}
	sizeInt, ok := sizeResult.(runtime.IntegerValue)
	if !ok || sizeInt.TypeSuffix != runtime.IntegerU64 || sizeInt.Val.Int64() != 2 {
		t.Fatalf("size() returned unexpected value %#v", sizeResult)
	}

	// get checks
	getVal, err := interp.arrayMember(arr, ast.NewIdentifier("get"))
	if err != nil {
		t.Fatalf("array get lookup failed: %v", err)
	}
	get := getVal.(*runtime.NativeBoundMethodValue)
	first, err := get.Method.Impl(ctx, []runtime.Value{get.Receiver, runtime.IntegerValue{Val: big.NewInt(0), TypeSuffix: runtime.IntegerI32}})
	if err != nil {
		t.Fatalf("get(0) failed: %v", err)
	}
	if v, ok := first.(runtime.IntegerValue); !ok || v.Val.Int64() != 1 {
		t.Fatalf("get(0) returned unexpected value %#v", first)
	}
	outOfRange, err := get.Method.Impl(ctx, []runtime.Value{get.Receiver, runtime.IntegerValue{Val: big.NewInt(5), TypeSuffix: runtime.IntegerI32}})
	if err != nil {
		t.Fatalf("get(5) failed: %v", err)
	}
	if _, ok := outOfRange.(runtime.NilValue); !ok {
		t.Fatalf("get(5) should return nil, got %#v", outOfRange)
	}

	// clone
	cloneVal, err := interp.arrayMember(arr, ast.NewIdentifier("clone"))
	if err != nil {
		t.Fatalf("array clone lookup failed: %v", err)
	}
	clone := cloneVal.(*runtime.NativeBoundMethodValue)
	cloneRes, err := clone.Method.Impl(ctx, []runtime.Value{clone.Receiver})
	if err != nil {
		t.Fatalf("clone() failed: %v", err)
	}
	cloneArr, ok := cloneRes.(*runtime.ArrayValue)
	if !ok || cloneArr == arr {
		t.Fatalf("clone() returned unexpected value %#v", cloneRes)
	}
	if len(cloneArr.Elements) != len(arr.Elements) {
		t.Fatalf("clone length mismatch")
	}
	cloneArr.Elements[0] = runtime.IntegerValue{Val: big.NewInt(42), TypeSuffix: runtime.IntegerI32}
	if v := arr.Elements[0].(runtime.IntegerValue); v.Val.Int64() != 1 {
		t.Fatalf("clone should copy elements, but mutation affected original")
	}

	// set
	setVal, err := interp.arrayMember(arr, ast.NewIdentifier("set"))
	if err != nil {
		t.Fatalf("array set lookup failed: %v", err)
	}
	set := setVal.(*runtime.NativeBoundMethodValue)
	if _, err := set.Method.Impl(ctx, []runtime.Value{set.Receiver, runtime.IntegerValue{Val: big.NewInt(1), TypeSuffix: runtime.IntegerI32}, runtime.IntegerValue{Val: big.NewInt(5), TypeSuffix: runtime.IntegerI32}}); err != nil {
		t.Fatalf("set(1,5) failed: %v", err)
	}
	if v := arr.Elements[1].(runtime.IntegerValue); v.Val.Int64() != 5 {
		t.Fatalf("set did not update element")
	}
	setErr, err := set.Method.Impl(ctx, []runtime.Value{set.Receiver, runtime.IntegerValue{Val: big.NewInt(10), TypeSuffix: runtime.IntegerI32}, runtime.IntegerValue{Val: big.NewInt(7), TypeSuffix: runtime.IntegerI32}})
	if err != nil {
		t.Fatalf("set(10,7) invocation failed: %v", err)
	}
	if _, ok := setErr.(runtime.ErrorValue); !ok {
		t.Fatalf("set out of range should return Error, got %#v", setErr)
	}

	// pop
	popVal, err := interp.arrayMember(arr, ast.NewIdentifier("pop"))
	if err != nil {
		t.Fatalf("array pop lookup failed: %v", err)
	}
	pop := popVal.(*runtime.NativeBoundMethodValue)
	popRes, err := pop.Method.Impl(ctx, []runtime.Value{pop.Receiver})
	if err != nil {
		t.Fatalf("pop() failed: %v", err)
	}
	if v, ok := popRes.(runtime.IntegerValue); !ok || v.Val.Int64() != 5 {
		t.Fatalf("pop returned unexpected value %#v", popRes)
	}
	popRes, err = pop.Method.Impl(ctx, []runtime.Value{pop.Receiver})
	if err != nil {
		t.Fatalf("second pop() failed: %v", err)
	}
	if v, ok := popRes.(runtime.IntegerValue); !ok || v.Val.Int64() != 1 {
		t.Fatalf("second pop returned unexpected value %#v", popRes)
	}
	popRes, err = pop.Method.Impl(ctx, []runtime.Value{pop.Receiver})
	if err != nil {
		t.Fatalf("third pop() failed: %v", err)
	}
	if _, ok := popRes.(runtime.NilValue); !ok {
		t.Fatalf("pop on empty should return nil, got %#v", popRes)
	}
}
