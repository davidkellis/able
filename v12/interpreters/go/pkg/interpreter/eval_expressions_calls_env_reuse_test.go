package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestCanReuseFunctionClosureEnvForBytecode(t *testing.T) {
	closure := runtime.NewEnvironment(nil)
	decl := ast.Fn(
		"f",
		nil,
		[]ast.Statement{ast.Int(1)},
		nil,
		nil,
		nil,
		false,
		false,
	)

	if canReuseFunctionClosureEnvForBytecode(nil, decl, nil, closure) {
		t.Fatalf("expected false for nil slot program")
	}
	if canReuseFunctionClosureEnvForBytecode(&bytecodeProgram{}, decl, nil, closure) {
		t.Fatalf("expected false for missing frame layout")
	}
	if canReuseFunctionClosureEnvForBytecode(&bytecodeProgram{frameLayout: &bytecodeFrameLayout{needsEnvScopes: true}}, decl, nil, closure) {
		t.Fatalf("expected false when frame layout requires env scopes")
	}
	if canReuseFunctionClosureEnvForBytecode(&bytecodeProgram{frameLayout: &bytecodeFrameLayout{}}, decl, nil, nil) {
		t.Fatalf("expected false for nil closure")
	}

	genericDecl := ast.Fn(
		"g",
		nil,
		[]ast.Statement{ast.Int(1)},
		nil,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	if canReuseFunctionClosureEnvForBytecode(&bytecodeProgram{frameLayout: &bytecodeFrameLayout{}}, genericDecl, nil, closure) {
		t.Fatalf("expected false for generic declarations")
	}

	callWithTypeArgs := ast.Call("f", ast.Int(1))
	callWithTypeArgs.TypeArguments = []ast.TypeExpression{ast.Ty("i32")}
	if canReuseFunctionClosureEnvForBytecode(&bytecodeProgram{frameLayout: &bytecodeFrameLayout{}}, decl, callWithTypeArgs, closure) {
		t.Fatalf("expected false for call sites with explicit type args")
	}

	if !canReuseFunctionClosureEnvForBytecode(&bytecodeProgram{frameLayout: &bytecodeFrameLayout{}}, decl, nil, closure) {
		t.Fatalf("expected true for non-generic slot bytecode without env scopes")
	}
}
