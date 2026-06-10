package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func methodScopeCacheProbeFunction(name string, closure *runtime.Environment) *runtime.FunctionValue {
	return &runtime.FunctionValue{
		Declaration: ast.Fn(
			name,
			[]*ast.FunctionParameter{ast.Param("receiver", nil)},
			[]ast.Statement{ast.ID("receiver")},
			nil,
			nil,
			nil,
			false,
			false,
		),
		Closure: closure,
	}
}

func TestResolveMethodFromPool_MethodScopeCallableCacheInvalidatesOnOwnerAssign(t *testing.T) {
	interp := New()
	root := interp.GlobalEnvironment()
	first := methodScopeCacheProbeFunction("first_probe", root)
	second := methodScopeCacheProbeFunction("second_probe", root)
	root.DefineWithoutMerge("scope_probe", first)
	child := runtime.NewEnvironment(root)
	receiver := runtime.NewSmallInt(1, runtime.IntegerI32)

	callable, found, err := interp.resolveMethodCallableFromPool(child, "scope_probe", receiver, "")
	if err != nil {
		t.Fatalf("resolve first scope callable: %v", err)
	}
	if !found || callable != first {
		t.Fatalf("expected first scope callable, found=%v callable=%T (%#v)", found, callable, callable)
	}
	if got := len(interp.methodScopeCallableCache); got != 1 {
		t.Fatalf("expected one method scope callable cache entry, got %d", got)
	}

	if err := root.Assign("scope_probe", second); err != nil {
		t.Fatalf("assign second scope callable: %v", err)
	}
	callable, found, err = interp.resolveMethodCallableFromPool(child, "scope_probe", receiver, "")
	if err != nil {
		t.Fatalf("resolve second scope callable: %v", err)
	}
	if !found || callable != second {
		t.Fatalf("expected owner-version invalidation to resolve second callable, found=%v callable=%T (%#v)", found, callable, callable)
	}
}

func TestResolveMethodFromPool_MethodScopeCallableCacheInvalidatesOnIntermediateShadow(t *testing.T) {
	interp := New()
	root := interp.GlobalEnvironment()
	parent := runtime.NewEnvironment(root)
	child := runtime.NewEnvironment(parent)
	first := methodScopeCacheProbeFunction("outer_probe", root)
	second := methodScopeCacheProbeFunction("shadow_probe", parent)
	root.DefineWithoutMerge("scope_probe", first)
	receiver := runtime.NewSmallInt(1, runtime.IntegerI32)

	callable, found, err := interp.resolveMethodCallableFromPool(child, "scope_probe", receiver, "")
	if err != nil {
		t.Fatalf("resolve outer scope callable: %v", err)
	}
	if !found || callable != first {
		t.Fatalf("expected outer scope callable, found=%v callable=%T (%#v)", found, callable, callable)
	}

	parent.DefineWithoutMerge("scope_probe", second)
	callable, found, err = interp.resolveMethodCallableFromPool(child, "scope_probe", receiver, "")
	if err != nil {
		t.Fatalf("resolve shadowed scope callable: %v", err)
	}
	if !found || callable != second {
		t.Fatalf("expected binding-shape invalidation to resolve shadow callable, found=%v callable=%T (%#v)", found, callable, callable)
	}
}

func TestMethodTypeNameScopeCacheInvalidatesOnBindingShapeChange(t *testing.T) {
	interp := New()
	root := interp.GlobalEnvironment()
	parent := runtime.NewEnvironment(root)
	child := runtime.NewEnvironment(parent)
	const typeName = "ScopeCacheType"

	if interp.methodTypeNameInScope(child, typeName) {
		t.Fatalf("unexpected initial type-name visibility for %s", typeName)
	}
	if got := len(interp.methodScopeHasCache); got != 1 {
		t.Fatalf("expected one method type-name scope cache entry, got %d", got)
	}

	parent.DefineWithoutMerge(typeName, runtime.StringValue{Val: "type-marker"})
	if !interp.methodTypeNameInScope(child, typeName) {
		t.Fatalf("expected binding-shape invalidation to observe %s", typeName)
	}
}
