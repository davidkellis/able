package bridge

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type testNativeAwaitable struct {
	materialized runtime.Value
}

func (*testNativeAwaitable) Kind() runtime.Kind {
	return runtime.KindStructInstance
}

func (a *testNativeAwaitable) MaterializeRuntimeValue() runtime.Value {
	return a.materialized
}

func (*testNativeAwaitable) NativeAwaitableIsReady(*runtime.NativeCallContext) (bool, error) {
	return true, nil
}

func (*testNativeAwaitable) NativeAwaitableRegister(*runtime.NativeCallContext, runtime.Value) (runtime.Value, error) {
	return runtime.NilValue{}, nil
}

func (*testNativeAwaitable) NativeAwaitableCommit(*runtime.NativeCallContext) (runtime.Value, error) {
	return runtime.NilValue{}, nil
}

func (*testNativeAwaitable) NativeAwaitableIsDefault() bool {
	return false
}

func TestMatchTypeWithoutInterpreterAcceptsGenericRuntimeAsyncValues(t *testing.T) {
	rt := New(nil)
	tests := []struct {
		name     string
		typeExpr ast.TypeExpression
		value    runtime.Value
	}{
		{
			name:     "future",
			typeExpr: ast.Gen(ast.Ty("Future"), ast.Ty("i64")),
			value:    runtime.NewFuture(),
		},
		{
			name:     "iterator",
			typeExpr: ast.Gen(ast.Ty("Iterator"), ast.Ty("String")),
			value: runtime.NewIteratorValue(func() (runtime.Value, bool, error) {
				return runtime.NilValue{}, true, nil
			}, nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok, err := MatchType(rt, tt.typeExpr, tt.value)
			if err != nil {
				t.Fatalf("MatchType: %v", err)
			}
			if !ok {
				t.Fatalf("MatchType(%T) did not match %s", tt.value, tt.name)
			}
			if got != tt.value {
				t.Fatalf("MatchType returned %#v, want original %#v", got, tt.value)
			}
		})
	}
}

func TestMatchTypeWithoutInterpreterRecognizesKernelAwaitableShape(t *testing.T) {
	rt := New(nil)
	awaitable := &runtime.StructInstanceValue{
		Fields: map[string]runtime.Value{
			"is_ready":   runtime.NativeFunctionValue{Name: "Awaitable.is_ready", Arity: 0},
			"register":   runtime.NativeFunctionValue{Name: "Awaitable.register", Arity: 1},
			"commit":     runtime.NativeFunctionValue{Name: "Awaitable.commit", Arity: 0},
			"is_default": runtime.NativeFunctionValue{Name: "Awaitable.is_default", Arity: 0},
		},
	}

	got, ok, err := MatchType(rt, ast.Gen(ast.Ty("Awaitable"), ast.Ty("i64")), awaitable)
	if err != nil {
		t.Fatalf("MatchType: %v", err)
	}
	if !ok || got != awaitable {
		t.Fatalf("MatchType(Awaitable<i64>) = (%#v, %t), want original awaitable and true", got, ok)
	}

	delete(awaitable.Fields, "register")
	if _, ok, err := MatchType(rt, ast.Gen(ast.Ty("Awaitable"), ast.Ty("i64")), awaitable); err != nil || ok {
		t.Fatalf("incomplete Awaitable match = (%t, %v), want false, nil", ok, err)
	}

	awaitable.Fields["register"] = runtime.NewSmallInt(1, runtime.IntegerI32)
	if _, ok, err := MatchType(rt, ast.Gen(ast.Ty("Awaitable"), ast.Ty("i64")), awaitable); err != nil || ok {
		t.Fatalf("non-callable Awaitable field match = (%t, %v), want false, nil", ok, err)
	}
}

func TestMatchTypeWithoutInterpreterPreservesNativeAwaitableCarrier(t *testing.T) {
	rt := New(nil)
	materialized := &runtime.StructInstanceValue{Fields: map[string]runtime.Value{}}
	awaitable := &testNativeAwaitable{materialized: materialized}

	got, ok, err := MatchType(rt, ast.Gen(ast.Ty("Awaitable"), ast.Ty("i64")), awaitable)
	if err != nil {
		t.Fatalf("MatchType: %v", err)
	}
	if !ok || got != awaitable {
		t.Fatalf("MatchType(Awaitable<i64>) = (%#v, %t), want native carrier and true", got, ok)
	}

	cast, err := Cast(rt, ast.Gen(ast.Ty("Awaitable"), ast.Ty("i64")), awaitable)
	if err != nil {
		t.Fatalf("Cast: %v", err)
	}
	if cast != awaitable {
		t.Fatalf("Cast(Awaitable<i64>) = %#v, want native carrier", cast)
	}

	inferred, err := TypeExpressionFromValue(rt, awaitable)
	if err != nil {
		t.Fatalf("TypeExpressionFromValue: %v", err)
	}
	simple, ok := inferred.(*ast.SimpleTypeExpression)
	if !ok || simple == nil || simple.Name == nil || simple.Name.Name != "Awaitable" {
		t.Fatalf("TypeExpressionFromValue(native Awaitable) = %#v", inferred)
	}

	if got := materializeBoundaryValue(awaitable); got != materialized {
		t.Fatalf("materializeBoundaryValue(native Awaitable) = %#v, want %#v", got, materialized)
	}
}

func TestMatchTypeWithoutInterpreterRecognizesHostHandleTypes(t *testing.T) {
	rt := New(nil)
	for _, handleType := range []string{"IoHandle", "ProcHandle"} {
		t.Run(handleType, func(t *testing.T) {
			handle := &runtime.HostHandleValue{HandleType: handleType, Value: struct{}{}}
			got, ok, err := MatchType(rt, ast.Ty(handleType), handle)
			if err != nil {
				t.Fatalf("MatchType: %v", err)
			}
			if !ok || got != handle {
				t.Fatalf("MatchType(%s) = (%#v, %t), want original handle and true", handleType, got, ok)
			}
			other := "IoHandle"
			if handleType == other {
				other = "ProcHandle"
			}
			if _, ok, err := MatchType(rt, ast.Ty(other), handle); err != nil || ok {
				t.Fatalf("MatchType(%s as %s) = (%t, %v), want false, nil", handleType, other, ok, err)
			}
			inferred, err := TypeExpressionFromValue(rt, handle)
			if err != nil {
				t.Fatalf("TypeExpressionFromValue: %v", err)
			}
			simple, ok := inferred.(*ast.SimpleTypeExpression)
			if !ok || simple == nil || simple.Name == nil || simple.Name.Name != handleType {
				t.Fatalf("TypeExpressionFromValue(%s) = %#v", handleType, inferred)
			}
		})
	}
}
