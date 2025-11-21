package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestArrayHelpersBuiltins(t *testing.T) {
	interp := New()
	interp.ensureArrayBuiltins()
	ctx := &runtime.NativeCallContext{Env: interp.global}
	arr := &runtime.ArrayValue{
		Elements: []runtime.Value{
			runtime.IntegerValue{Val: big.NewInt(1), TypeSuffix: runtime.IntegerI32},
			runtime.IntegerValue{Val: big.NewInt(2), TypeSuffix: runtime.IntegerI32},
		},
	}

	sizeBound, err := interp.arrayMember(arr, ast.NewIdentifier("size"))
	if err != nil {
		t.Fatalf("size bind failed: %v", err)
	}
	sizeRes, err := sizeBound.(*runtime.NativeBoundMethodValue).Method.Impl(ctx, []runtime.Value{sizeBound.(*runtime.NativeBoundMethodValue).Receiver})
	if err != nil {
		t.Fatalf("size call failed: %v", err)
	}
	if v, ok := sizeRes.(runtime.IntegerValue); !ok || v.Val.Int64() != 2 {
		t.Fatalf("unexpected size result %#v", sizeRes)
	}

	pushBound, err := interp.arrayMember(arr, ast.NewIdentifier("push"))
	if err != nil {
		t.Fatalf("push bind failed: %v", err)
	}
	if _, err := pushBound.(*runtime.NativeBoundMethodValue).Method.Impl(ctx, []runtime.Value{pushBound.(*runtime.NativeBoundMethodValue).Receiver, runtime.IntegerValue{Val: big.NewInt(3), TypeSuffix: runtime.IntegerI32}}); err != nil {
		t.Fatalf("push call failed: %v", err)
	}

	popBound, err := interp.arrayMember(arr, ast.NewIdentifier("pop"))
	if err != nil {
		t.Fatalf("pop bind failed: %v", err)
	}
	popRes, err := popBound.(*runtime.NativeBoundMethodValue).Method.Impl(ctx, []runtime.Value{popBound.(*runtime.NativeBoundMethodValue).Receiver})
	if err != nil {
		t.Fatalf("pop call failed: %v", err)
	}
	if v, ok := popRes.(runtime.IntegerValue); !ok || v.Val.Int64() != 3 {
		t.Fatalf("unexpected pop result %#v", popRes)
	}

	getBound, err := interp.arrayMember(arr, ast.NewIdentifier("get"))
	if err != nil {
		t.Fatalf("get bind failed: %v", err)
	}
	getRes, err := getBound.(*runtime.NativeBoundMethodValue).Method.Impl(ctx, []runtime.Value{getBound.(*runtime.NativeBoundMethodValue).Receiver, runtime.IntegerValue{Val: big.NewInt(0), TypeSuffix: runtime.IntegerI32}})
	if err != nil {
		t.Fatalf("get call failed: %v", err)
	}
	if v, ok := getRes.(runtime.IntegerValue); !ok || v.Val.Int64() != 1 {
		t.Fatalf("unexpected get result %#v", getRes)
	}

	setBound, err := interp.arrayMember(arr, ast.NewIdentifier("set"))
	if err != nil {
		t.Fatalf("set bind failed: %v", err)
	}
	setRes, err := setBound.(*runtime.NativeBoundMethodValue).Method.Impl(ctx, []runtime.Value{setBound.(*runtime.NativeBoundMethodValue).Receiver, runtime.IntegerValue{Val: big.NewInt(0), TypeSuffix: runtime.IntegerI32}, runtime.IntegerValue{Val: big.NewInt(9), TypeSuffix: runtime.IntegerI32}})
	if err != nil {
		t.Fatalf("set call failed: %v", err)
	}
	if _, ok := setRes.(runtime.NilValue); !ok {
		t.Fatalf("expected nil from set, got %#v", setRes)
	}
	setErr, err := setBound.(*runtime.NativeBoundMethodValue).Method.Impl(ctx, []runtime.Value{setBound.(*runtime.NativeBoundMethodValue).Receiver, runtime.IntegerValue{Val: big.NewInt(5), TypeSuffix: runtime.IntegerI32}, runtime.IntegerValue{Val: big.NewInt(1), TypeSuffix: runtime.IntegerI32}})
	if err != nil {
		t.Fatalf("set out-of-range errored: %v", err)
	}
	if _, ok := setErr.(runtime.ErrorValue); !ok {
		t.Fatalf("expected error value from out-of-range set, got %#v", setErr)
	}

	clearBound, err := interp.arrayMember(arr, ast.NewIdentifier("clear"))
	if err != nil {
		t.Fatalf("clear bind failed: %v", err)
	}
	if _, err := clearBound.(*runtime.NativeBoundMethodValue).Method.Impl(ctx, []runtime.Value{clearBound.(*runtime.NativeBoundMethodValue).Receiver}); err != nil {
		t.Fatalf("clear call failed: %v", err)
	}
	if len(arr.Elements) != 0 {
		t.Fatalf("expected array to be cleared, has %d elements", len(arr.Elements))
	}
}
