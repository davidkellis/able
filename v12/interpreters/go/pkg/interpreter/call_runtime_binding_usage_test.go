package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestFunctionNeedsCallLocalTypeBindingsOnlyWhenBodyReadsRuntimeBindings(t *testing.T) {
	interp := NewBytecode()
	methodSet := &runtime.MethodSet{
		TargetType: ast.Gen(ast.Ty("Box"), ast.Ty("T")),
		GenericParams: []*ast.GenericParameter{
			ast.GenericParam("T"),
		},
	}

	noBindingUse := &runtime.FunctionValue{
		Declaration: ast.Fn(
			"value",
			[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
			[]ast.Statement{ast.ID("self")},
			ast.Ty("Self"),
			nil,
			nil,
			false,
			false,
		),
		MethodSet: methodSet,
		Closure:   interp.GlobalEnvironment(),
	}
	if interp.functionNeedsCallLocalTypeBindings(noBindingUse) {
		t.Fatalf("expected false when method body never reads Self/T runtime bindings")
	}

	withBindingUse := &runtime.FunctionValue{
		Declaration: ast.Fn(
			"type_name",
			[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
			[]ast.Statement{ast.ID("Self_type")},
			ast.Ty("String"),
			nil,
			nil,
			false,
			false,
		),
		MethodSet: methodSet,
		Closure:   interp.GlobalEnvironment(),
	}
	if !interp.functionNeedsCallLocalTypeBindings(withBindingUse) {
		t.Fatalf("expected true when method body reads Self_type")
	}
}

func TestFunctionRuntimeGenericBindingPlanCachesCallableSignatureFacts(t *testing.T) {
	interp := New()
	methodSet := &runtime.MethodSet{
		TargetType: ast.Gen(ast.Ty("Box"), ast.Ty("T")),
		GenericParams: []*ast.GenericParameter{
			ast.GenericParam("T"),
		},
	}
	fn := &runtime.FunctionValue{
		Declaration: ast.Fn(
			"describe",
			[]*ast.FunctionParameter{
				ast.Param("self", ast.Ty("Self")),
				ast.Param("value", ast.Ty("T")),
				ast.Param("count", ast.Ty("i32")),
			},
			[]ast.Statement{ast.ID("Self_type"), ast.ID("T_type"), ast.Ret(ast.ID("value"))},
			ast.Ty("T"),
			[]*ast.GenericParameter{
				ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Show"))),
			},
			nil,
			false,
			false,
		),
		MethodSet: methodSet,
		Closure:   interp.GlobalEnvironment(),
	}

	planA := interp.functionRuntimeGenericBindingPlan(fn)
	planB := interp.functionRuntimeGenericBindingPlan(fn)

	if planA == nil || planB == nil {
		t.Fatalf("expected cached runtime generic binding plan")
	}
	if planA != planB {
		t.Fatalf("expected function runtime generic binding plan cache reuse")
	}
	if !planA.explicitUsed {
		t.Fatalf("expected plan to record explicit runtime binding usage")
	}
	if !planA.callLocalUsed {
		t.Fatalf("expected plan to record call-local runtime binding usage")
	}
	if !planA.hasGenericConstraints {
		t.Fatalf("expected plan to record generic constraints")
	}
	if !planA.returnTypeUsesGenerics {
		t.Fatalf("expected plan to record generic return type usage")
	}
	if planA.paramUsesGeneric(0) {
		t.Fatalf("did not expect Self parameter to be marked as generic")
	}
	if !planA.paramUsesGeneric(1) {
		t.Fatalf("expected T parameter to be marked as generic")
	}
	if planA.paramUsesGeneric(2) {
		t.Fatalf("did not expect concrete parameter to be marked as generic")
	}
	if !interp.functionParamUsesGenerics(fn, 1, ast.Ty("T")) {
		t.Fatalf("expected helper to reuse cached generic parameter usage")
	}
	if interp.functionParamUsesGenerics(fn, 2, ast.Ty("i32")) {
		t.Fatalf("did not expect helper to mark concrete parameter as generic")
	}
	if !interp.functionReturnTypeUsesGenerics(fn, ast.Ty("T")) {
		t.Fatalf("expected helper to reuse cached generic return-type usage")
	}
	if got := len(interp.functionRuntimeGenericBindingPlanCache); got != 1 {
		t.Fatalf("expected one runtime generic binding plan cache entry, got %d", got)
	}
}

func TestBytecodeVM_GenericMemberCallWithoutRuntimeBindingUseSkipsBindingEnv(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	module := mustParseModuleSource(t, `
struct Box T {
  value: T
}

methods Box T {
  fn value(self: Self) -> T {
    self.value
  }
}

box := Box { value: 7 }
result := 0
i := 0
loop {
  if i >= 2 { break }
  result = box.value()
  i = i + 1
}
result
`)

	want := mustEvalModule(t, New(), module)
	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic member call mismatch: got=%#v want=%#v", got, want)
	}
	if got := len(interp.callLocalTypeBindingCache); got != 0 {
		t.Fatalf("expected no call-local type binding cache entries when runtime bindings are unused, got %d", got)
	}
	if got := len(interp.reusableBytecodeCallEnvCache); got != 0 {
		t.Fatalf("expected no reusable binding env cache entries when closure env can be reused directly, got %d", got)
	}
	stats := interp.BytecodeStats()
	if stats.InlineCallHits == 0 {
		t.Fatalf("expected inline call hits for repeated generic member call")
	}
}
