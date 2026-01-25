package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestExternHandlersRegisterNativeFunctions(t *testing.T) {
	interp := New()
	sig := ast.Fn("now_nanos", nil, nil, ast.Ty("i64"), nil, nil, false, false)
	mod := ast.Mod([]ast.Statement{ast.Extern(ast.HostTargetGo, sig, "return 0")}, nil, ast.Pkg([]interface{}{"host"}, false))

	_, env, err := interp.EvaluateModule(mod)
	if err != nil {
		t.Fatalf("evaluate module: %v", err)
	}
	val, err := env.Get("now_nanos")
	if err != nil {
		t.Fatalf("now_nanos missing: %v", err)
	}
	if _, ok := val.(*runtime.NativeFunctionValue); !ok {
		t.Fatalf("expected native function, got %#v", val)
	}
	bucket := interp.packageRegistry["host"]
	if bucket == nil || bucket["now_nanos"] == nil {
		t.Fatalf("package registry missing now_nanos")
	}
}

func TestExternIgnoresNonGoTargets(t *testing.T) {
	interp := New()
	sig := ast.Fn("ts_only", nil, nil, ast.Ty("i64"), nil, nil, false, false)
	mod := ast.Mod([]ast.Statement{ast.Extern(ast.HostTargetTypeScript, sig, "")}, nil, nil)

	_, env, err := interp.EvaluateModule(mod)
	if err != nil {
		t.Fatalf("evaluate module: %v", err)
	}
	if _, err := env.Get("ts_only"); err == nil {
		t.Fatalf("expected ts_only to be undefined for go target")
	}
}

func TestExternPreservesExistingBinding(t *testing.T) {
	interp := New()
	existing := &runtime.NativeFunctionValue{Name: "now_nanos", Arity: 0}
	interp.global.Define("now_nanos", existing)

	sig := ast.Fn("now_nanos", nil, nil, ast.Ty("i64"), nil, nil, false, false)
	mod := ast.Mod([]ast.Statement{ast.Extern(ast.HostTargetGo, sig, "return 0")}, nil, nil)

	_, env, err := interp.EvaluateModule(mod)
	if err != nil {
		t.Fatalf("evaluate module: %v", err)
	}
	val, err := env.Get("now_nanos")
	if err != nil {
		t.Fatalf("now_nanos missing: %v", err)
	}
	if val != existing {
		t.Fatalf("expected existing binding to remain")
	}
}
