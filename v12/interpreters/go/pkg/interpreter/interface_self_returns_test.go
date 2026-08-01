package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestPreserveInterfaceSelfReturnKeepsCapturedDictionaryIsolated(t *testing.T) {
	signature := ast.FnSig(
		"duplicate",
		[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
		ast.Ty("Self"),
		nil,
		nil,
		nil,
	)
	node := ast.Iface("Duplicable", []*ast.FunctionSignature{signature}, nil, nil, nil, nil, false)
	definition := &runtime.InterfaceDefinitionValue{Node: node, QualifiedName: "Duplicable"}
	shared := &runtime.NativeFunctionValue{Name: "shared", Arity: 0}
	overlay := &runtime.NativeFunctionValue{Name: "overlay", Arity: 0}
	original := &runtime.InterfaceValue{
		Interface:     definition,
		Underlying:    runtime.StringValue{Val: "before"},
		SharedMethods: map[string]runtime.Value{"shared": shared},
		Methods:       map[string]runtime.Value{"overlay": overlay},
	}
	interp := New()
	interp.interfaces["Duplicable"] = definition
	if !interp.interfaceMethodReturnsSelf(original, "duplicate") {
		t.Fatal("expected duplicate signature to be recognized as an exact Self return")
	}

	preserved, ok := preserveInterfaceSelfReturn(original, runtime.StringValue{Val: "after"}).(*runtime.InterfaceValue)
	if !ok || preserved == nil {
		t.Fatalf("preserved value = %T, want *runtime.InterfaceValue", preserved)
	}
	if preserved.Interface != original.Interface || preserved.SharedMethods["shared"] != shared {
		t.Fatal("expected the returned interface value to keep the captured interface and shared dictionary")
	}
	if preserved.Methods["overlay"] != overlay {
		t.Fatal("expected the returned interface value to keep its captured overlay entries")
	}
	preserved.Methods["new"] = shared
	if _, leaked := original.Methods["new"]; leaked {
		t.Fatal("expected the returned interface value to own an isolated overlay map")
	}
	if got, ok := preserved.Underlying.(runtime.StringValue); !ok || got.Val != "after" {
		t.Fatalf("underlying result = %#v, want String(\"after\")", preserved.Underlying)
	}
}
