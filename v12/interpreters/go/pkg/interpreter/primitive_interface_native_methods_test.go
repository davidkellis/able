package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func bootstrapPrimitiveInterfaceNativeTests(t *testing.T, interp *Interpreter) {
	t.Helper()
	program := mustLoadAbleProgramFromSource(t, `
import able.core.interfaces.{Less}

Less
`)
	if _, _, _, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{}); err != nil {
		t.Fatalf("bootstrap program failed: %v", err)
	}
}

func primitiveOrderingTagNameForTest(val runtime.Value) string {
	switch v := val.(type) {
	case *runtime.StructDefinitionValue:
		if v != nil && v.Node != nil && v.Node.ID != nil {
			return v.Node.ID.Name
		}
	case runtime.StructDefinitionValue:
		if v.Node != nil && v.Node.ID != nil {
			return v.Node.ID.Name
		}
	case *runtime.StructInstanceValue:
		if v != nil && v.Definition != nil && v.Definition.Node != nil && v.Definition.Node.ID != nil {
			return v.Definition.Node.ID.Name
		}
	}
	return ""
}

func TestResolveInterfaceMethod_PrimitiveCmpUsesNativeCallable(t *testing.T) {
	interp := NewBytecode()
	bootstrapPrimitiveInterfaceNativeTests(t, interp)

	method, err := interp.resolveInterfaceMethod(runtime.NewSmallInt(1, runtime.IntegerI32), "Ord", "cmp")
	if err != nil {
		t.Fatalf("resolve interface method failed: %v", err)
	}
	if _, ok := method.(runtime.NativeFunctionValue); !ok {
		if _, ok := method.(*runtime.NativeFunctionValue); !ok {
			t.Fatalf("expected native callable, got %T", method)
		}
	}

	result, err := interp.callCallableValue(
		method,
		[]runtime.Value{
			runtime.NewSmallInt(1, runtime.IntegerI32),
			runtime.NewSmallInt(2, runtime.IntegerI32),
		},
		interp.GlobalEnvironment(),
		nil,
	)
	if err != nil {
		t.Fatalf("primitive cmp call failed: %v", err)
	}
	if got := primitiveOrderingTagNameForTest(result); got != "Less" {
		t.Fatalf("primitive cmp result = %#v, want Less ordering", result)
	}
}

func TestResolveInterfaceMethod_PrimitiveHashUsesNativeCallable(t *testing.T) {
	interp := NewBytecode()
	bootstrapPrimitiveInterfaceNativeTests(t, interp)

	method, err := interp.resolveInterfaceMethod(runtime.NewSmallInt(-42, runtime.IntegerI32), "Hash", "hash")
	if err != nil {
		t.Fatalf("resolve primitive hash failed: %v", err)
	}
	if _, ok := method.(runtime.NativeFunctionValue); !ok {
		if _, ok := method.(*runtime.NativeFunctionValue); !ok {
			t.Fatalf("expected native callable, got %T", method)
		}
	}

	hasher, err := interp.newKernelHasher()
	if err != nil {
		t.Fatalf("new kernel hasher failed: %v", err)
	}
	result, err := interp.callCallableValue(
		method,
		[]runtime.Value{
			runtime.NewSmallInt(-42, runtime.IntegerI32),
			hasher,
		},
		interp.GlobalEnvironment(),
		nil,
	)
	if err != nil {
		t.Fatalf("primitive hash call failed: %v", err)
	}
	if !isVoidOrNil(result) {
		t.Fatalf("primitive hash result = %#v, want void/nil", result)
	}
	hash, err := interp.finishKernelHasher(hasher)
	if err != nil {
		t.Fatalf("finish kernel hasher failed: %v", err)
	}
	if hash != 11047133508193088422 {
		t.Fatalf("hash = %d, want 11047133508193088422", hash)
	}
}

func TestResolveInterfaceMethod_PrimitivePartialCmpFloatUsesNativeCallable(t *testing.T) {
	interp := NewBytecode()
	bootstrapPrimitiveInterfaceNativeTests(t, interp)

	method, err := interp.resolveInterfaceMethod(runtime.FloatValue{Val: 2.5, TypeSuffix: runtime.FloatF64}, "PartialOrd", "partial_cmp")
	if err != nil {
		t.Fatalf("resolve float partial_cmp failed: %v", err)
	}
	if _, ok := method.(runtime.NativeFunctionValue); !ok {
		if _, ok := method.(*runtime.NativeFunctionValue); !ok {
			t.Fatalf("expected native callable, got %T", method)
		}
	}

	result, err := interp.callCallableValue(
		method,
		[]runtime.Value{
			runtime.FloatValue{Val: 2.5, TypeSuffix: runtime.FloatF64},
			runtime.FloatValue{Val: 1.5, TypeSuffix: runtime.FloatF64},
		},
		interp.GlobalEnvironment(),
		nil,
	)
	if err != nil {
		t.Fatalf("primitive partial_cmp call failed: %v", err)
	}
	if got := primitiveOrderingTagNameForTest(result); got != "Greater" {
		t.Fatalf("primitive partial_cmp result = %#v, want Greater ordering", result)
	}
}

func TestResolveInterfaceMethod_NilCloneUsesNativeCallable(t *testing.T) {
	interp := NewBytecode()
	interp.initInterfaceBuiltins()

	info, ok := interp.getTypeInfoForValue(runtime.NilValue{})
	if !ok {
		t.Fatalf("expected nil type info")
	}
	okImpl, err := interp.typeImplementsInterface(info, "Clone", nil, make(map[interfaceImplCacheKey]struct{}))
	if err != nil {
		t.Fatalf("nil Clone implementation lookup failed: %v", err)
	}
	if !okImpl {
		t.Fatalf("expected nil to implement Clone")
	}
	if err := interp.ensureTypeSatisfiesInterface(info, ast.Ty("Clone"), "nil", nil); err != nil {
		t.Fatalf("nil Clone constraint failed: %v", err)
	}

	method, err := interp.resolveInterfaceMethod(runtime.NilValue{}, "Clone", "clone")
	if err != nil {
		t.Fatalf("resolve nil clone failed: %v", err)
	}
	if _, ok := method.(runtime.NativeFunctionValue); !ok {
		if _, ok := method.(*runtime.NativeFunctionValue); !ok {
			t.Fatalf("expected native callable, got %T", method)
		}
	}

	result, err := interp.callCallableValue(
		method,
		[]runtime.Value{runtime.NilValue{}},
		interp.GlobalEnvironment(),
		nil,
	)
	if err != nil {
		t.Fatalf("nil clone call failed: %v", err)
	}
	if _, ok := result.(runtime.NilValue); !ok {
		t.Fatalf("nil clone result = %#v, want nil", result)
	}
}

func TestBytecodeTracePrimitiveCmpUsesExactNativeDispatch(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_TRACE", "1")

	program := mustLoadAbleProgramFromSource(t, `
import able.core.interfaces.{Less}

fn is_less(left: i32, right: i32) -> bool {
  left.cmp(right) == Less
}

fn main() -> bool {
  is_less(1, 2) && is_less(1, 2)
}

main()
`)

	interp := NewBytecode()
	got, _, _, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{})
	if err != nil {
		t.Fatalf("bytecode evaluation failed: %v", err)
	}
	if boolVal, ok := got.(runtime.BoolValue); !ok || !boolVal.Val {
		t.Fatalf("expected repeated primitive cmp call to return true, got %#v", got)
	}

	snapshot := interp.BytecodeTrace(0)
	var exactNativeCmp bool
	for _, entry := range snapshot.Entries {
		if entry.Op == "call_member" &&
			entry.Name == "cmp" &&
			entry.Lookup == "resolved_method" &&
			entry.Dispatch == "exact_native" {
			exactNativeCmp = true
			break
		}
	}
	if !exactNativeCmp {
		t.Fatalf("expected exact_native cmp trace entry, got %#v", snapshot.Entries)
	}
}
