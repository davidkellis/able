package compiler

import "testing"

func TestRuntimeValueExprUsesDynamicI64BoundaryCache(t *testing.T) {
	gen := newGenerator(Options{PackageName: "main"})

	got, ok := gen.runtimeValueExpr("value", "int64")
	if !ok {
		t.Fatal("runtimeValueExpr(value, int64) was not supported")
	}
	if want := "bridge.ToDynamicI64(int64(value))"; got != want {
		t.Fatalf("runtimeValueExpr(value, int64) = %q, want %q", got, want)
	}

	got, ok = gen.runtimeValueExpr("value", "int32")
	if !ok {
		t.Fatal("runtimeValueExpr(value, int32) was not supported")
	}
	if want := "bridge.ToInt(int64(value), runtime.IntegerType(\"i32\"))"; got != want {
		t.Fatalf("runtimeValueExpr(value, int32) = %q, want %q", got, want)
	}
}
