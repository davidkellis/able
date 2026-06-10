package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

var iteratorInterfaceCoercionSink runtime.Value

func defineTestIteratorInterface(interp *Interpreter) {
	interp.interfaces["Iterator"] = &runtime.InterfaceDefinitionValue{
		Node: ast.NewInterfaceDefinition(
			ast.NewIdentifier("Iterator"),
			[]*ast.FunctionSignature{
				ast.FnSig(
					"next",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					nil,
					nil,
					nil,
					nil,
				),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		Env: interp.GlobalEnvironment(),
	}
}

func defineTestIteratorInterfaceWithDefaultReady(interp *Interpreter) {
	interp.interfaces["Iterator"] = &runtime.InterfaceDefinitionValue{
		Node: ast.NewInterfaceDefinition(
			ast.NewIdentifier("Iterator"),
			[]*ast.FunctionSignature{
				ast.FnSig(
					"next",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					nil,
					nil,
					nil,
					nil,
				),
				ast.FnSig(
					"ready",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("bool"),
					nil,
					nil,
					ast.Block(ast.Ret(ast.Bool(true))),
				),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		Env: interp.GlobalEnvironment(),
	}
}

func TestCoerceToInterfaceValueIteratorUsesSharedMethodDictionary(t *testing.T) {
	interp := New()
	defineTestIteratorInterface(interp)
	iter := runtime.NewIteratorValue(func() (runtime.Value, bool, error) {
		return runtime.IteratorEnd, true, nil
	}, nil)

	coercedA, err := interp.coerceToInterfaceValue("Iterator", iter, nil)
	if err != nil {
		t.Fatalf("coerce first iterator: %v", err)
	}
	ifaceA, ok := coercedA.(*runtime.InterfaceValue)
	if !ok || ifaceA == nil {
		t.Fatalf("coerced first = %T (%#v), want *runtime.InterfaceValue", coercedA, coercedA)
	}
	if ifaceA.Methods != nil {
		t.Fatalf("expected iterator interface value to avoid owned method map, got %#v", ifaceA.Methods)
	}
	if ifaceA.SharedMethods == nil {
		t.Fatalf("expected iterator interface value to attach shared method dictionary")
	}
	if method, ok := interfaceValueLookupMethod(ifaceA, "next"); !ok || method == nil {
		t.Fatalf("expected shared Iterator.next method, got %T (%#v)", method, method)
	}
	if len(interp.iteratorInterfaceMethodDictionaryCache) != 1 {
		t.Fatalf("expected iterator method dictionary cache to contain one entry, got %d", len(interp.iteratorInterfaceMethodDictionaryCache))
	}

	interfaceValueSetMethod(ifaceA, "local_only", runtime.BoolValue{Val: true})

	coercedB, err := interp.coerceToInterfaceValue("Iterator", iter, nil)
	if err != nil {
		t.Fatalf("coerce second iterator: %v", err)
	}
	ifaceB, ok := coercedB.(*runtime.InterfaceValue)
	if !ok || ifaceB == nil {
		t.Fatalf("coerced second = %T (%#v), want *runtime.InterfaceValue", coercedB, coercedB)
	}
	if ifaceB.Methods != nil {
		t.Fatalf("expected cached iterator interface value to avoid owned method map, got %#v", ifaceB.Methods)
	}
	if ifaceB.SharedMethods == nil {
		t.Fatalf("expected second iterator interface value to attach shared method dictionary")
	}
	if method, ok := interfaceValueLookupMethod(ifaceB, "local_only"); ok || method != nil {
		t.Fatalf("expected per-value method overlay to stay isolated, got %T (%#v)", method, method)
	}

	interp.invalidateMethodCache()
	if len(interp.iteratorInterfaceMethodDictionaryCache) != 0 {
		t.Fatalf("expected iterator method dictionary cache clear on invalidate, got %d entries", len(interp.iteratorInterfaceMethodDictionaryCache))
	}
}

func TestCoerceToInterfaceValueIteratorSharedDictionaryAllocationBudget(t *testing.T) {
	interp := New()
	defineTestIteratorInterface(interp)
	iter := runtime.NewIteratorValue(func() (runtime.Value, bool, error) {
		return runtime.IteratorEnd, true, nil
	}, nil)

	if _, err := interp.coerceToInterfaceValue("Iterator", iter, nil); err != nil {
		t.Fatalf("warm iterator coercion: %v", err)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		coerced, err := interp.coerceToInterfaceValue("Iterator", iter, nil)
		if err != nil {
			panic(err)
		}
		iface, ok := coerced.(*runtime.InterfaceValue)
		if !ok || iface == nil || iface.SharedMethods == nil || iface.Methods != nil {
			panic("iterator coercion did not use shared method dictionary")
		}
		iteratorInterfaceCoercionSink = coerced
	})
	if allocs > 1 {
		t.Fatalf("cached iterator coercion allocs=%g, want at most escaping interface wrapper allocation", allocs)
	}
}

func TestIteratorDefaultInterfaceMethodValueCache(t *testing.T) {
	interp := NewBytecode()
	defineTestIteratorInterfaceWithDefaultReady(interp)
	ifaceDef := interp.interfaces["Iterator"]

	first, ok, err := interp.iteratorInterfaceMethodValue(ifaceDef, "ready")
	if err != nil {
		t.Fatalf("resolve first default method: %v", err)
	}
	if !ok || first == nil {
		t.Fatalf("expected Iterator.ready default method")
	}
	firstFn, ok := first.(*runtime.FunctionValue)
	if !ok || firstFn == nil || firstFn.Bytecode == nil {
		t.Fatalf("expected cached default method to carry bytecode, got %T (%#v)", first, first)
	}
	if len(interp.interfaceDefaultMethodCache) != 1 {
		t.Fatalf("expected one cached default method, got %d", len(interp.interfaceDefaultMethodCache))
	}

	second, ok, err := interp.iteratorInterfaceMethodValue(ifaceDef, "ready")
	if err != nil {
		t.Fatalf("resolve cached default method: %v", err)
	}
	if !ok || second != first {
		t.Fatalf("expected cached default method pointer reuse, got %T (%#v)", second, second)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		method, ok, err := interp.iteratorInterfaceMethodValue(ifaceDef, "ready")
		if err != nil {
			panic(err)
		}
		if !ok || method != first {
			panic("default method cache miss")
		}
		iteratorInterfaceCoercionSink = method
	})
	if allocs != 0 {
		t.Fatalf("cached default method lookup allocs=%g, want zero", allocs)
	}

	interp.invalidateMethodCache()
	if len(interp.interfaceDefaultMethodCache) != 0 {
		t.Fatalf("expected default method cache clear on invalidate, got %d entries", len(interp.interfaceDefaultMethodCache))
	}
	third, ok, err := interp.iteratorInterfaceMethodValue(ifaceDef, "ready")
	if err != nil {
		t.Fatalf("resolve default method after invalidation: %v", err)
	}
	if !ok || third == nil || third == first {
		t.Fatalf("expected rebuilt default method after invalidation, got %T (%#v)", third, third)
	}
}

func TestIteratorDefaultMethodRetainsInterfaceGenericReturn(t *testing.T) {
	interp := NewBytecode()
	ifaceDef := &runtime.InterfaceDefinitionValue{
		Node: ast.NewInterfaceDefinition(
			ast.NewIdentifier("Iterator"),
			[]*ast.FunctionSignature{
				ast.FnSig(
					"filter",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self")), ast.Param("keep", ast.FnType([]ast.TypeExpression{ast.Ty("T")}, ast.Ty("bool")))},
					ast.Gen(ast.Ty("Iterator"), ast.Ty("T")),
					nil,
					nil,
					ast.Block(ast.Ret(ast.Nil())),
				),
			},
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			nil,
			nil,
			false,
		),
		Env: interp.GlobalEnvironment(),
	}
	interp.interfaces["Iterator"] = ifaceDef

	method, ok, err := interp.iteratorInterfaceMethodValue(ifaceDef, "filter")
	if err != nil {
		t.Fatalf("resolve generic default method: %v", err)
	}
	if !ok {
		t.Fatal("expected generic default method")
	}
	fn, ok := method.(*runtime.FunctionValue)
	if !ok || fn == nil || fn.MethodSet == nil {
		t.Fatalf("default method = %T (%#v), want function with method set", method, method)
	}
	if len(fn.MethodSet.GenericParams) != 1 || fn.MethodSet.GenericParams[0] == nil || fn.MethodSet.GenericParams[0].Name == nil || fn.MethodSet.GenericParams[0].Name.Name != "T" {
		t.Fatalf("method-set generic params = %#v, want interface T", fn.MethodSet.GenericParams)
	}
	program, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || program == nil {
		t.Fatalf("default method bytecode = %T (%#v), want *bytecodeProgram", fn.Bytecode, fn.Bytecode)
	}
	if _, ok := program.returnGenericNames["T"]; !ok {
		t.Fatalf("return generic names = %#v, want interface T", program.returnGenericNames)
	}
	if !program.returnTypeUsesGenerics {
		t.Fatal("expected generic interface default return metadata")
	}
}
