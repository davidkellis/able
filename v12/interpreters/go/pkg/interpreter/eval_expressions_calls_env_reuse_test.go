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
	if !canReuseFunctionClosureEnvForBytecode(&bytecodeProgram{frameLayout: &bytecodeFrameLayout{}}, genericDecl, nil, closure) {
		t.Fatalf("expected true when generic declaration never reads runtime generic bindings")
	}

	genericBindingDecl := ast.Fn(
		"show_t",
		nil,
		[]ast.Statement{ast.ID("T_type")},
		ast.Ty("String"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	if canReuseFunctionClosureEnvForBytecode(&bytecodeProgram{frameLayout: &bytecodeFrameLayout{}}, genericBindingDecl, nil, closure) {
		t.Fatalf("expected false when generic declaration reads runtime generic bindings")
	}

	callWithTypeArgs := ast.Call("g", ast.Int(1))
	callWithTypeArgs.TypeArguments = []ast.TypeExpression{ast.Ty("i32")}
	if !canReuseFunctionClosureEnvForBytecode(&bytecodeProgram{frameLayout: &bytecodeFrameLayout{}}, genericDecl, callWithTypeArgs, closure) {
		t.Fatalf("expected explicit type args to keep closure-env reuse when runtime generic bindings are unused")
	}

	if !canReuseFunctionClosureEnvForBytecode(&bytecodeProgram{frameLayout: &bytecodeFrameLayout{}}, decl, nil, closure) {
		t.Fatalf("expected true for non-generic slot bytecode without env scopes")
	}
}

func TestCanReuseLambdaClosureEnvForBytecode(t *testing.T) {
	closure := runtime.NewEnvironment(nil)
	decl := ast.Lam(
		[]*ast.FunctionParameter{ast.Param("x", nil)},
		ast.ID("x"),
	)

	if canReuseLambdaClosureEnvForBytecode(nil, decl, nil, closure) {
		t.Fatalf("expected false for nil slot program")
	}
	if canReuseLambdaClosureEnvForBytecode(&bytecodeProgram{frameLayout: &bytecodeFrameLayout{needsEnvScopes: true}}, decl, nil, closure) {
		t.Fatalf("expected false when frame layout requires env scopes")
	}
	if canReuseLambdaClosureEnvForBytecode(&bytecodeProgram{frameLayout: &bytecodeFrameLayout{}}, decl, nil, nil) {
		t.Fatalf("expected false for nil closure")
	}

	genericDecl := ast.NewLambdaExpression(
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("T"))},
		ast.ID("x"),
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	if !canReuseLambdaClosureEnvForBytecode(&bytecodeProgram{frameLayout: &bytecodeFrameLayout{}}, genericDecl, nil, closure) {
		t.Fatalf("expected true when generic lambda never reads runtime generic bindings")
	}

	genericBindingDecl := ast.NewLambdaExpression(
		nil,
		ast.ID("T_type"),
		ast.Ty("String"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	if canReuseLambdaClosureEnvForBytecode(&bytecodeProgram{frameLayout: &bytecodeFrameLayout{}}, genericBindingDecl, nil, closure) {
		t.Fatalf("expected false when generic lambda reads runtime generic bindings")
	}

	callWithTypeArgs := ast.NewFunctionCall(ast.ID("id"), []ast.Expression{ast.Int(1)}, []ast.TypeExpression{ast.Ty("i32")}, false)
	if !canReuseLambdaClosureEnvForBytecode(&bytecodeProgram{frameLayout: &bytecodeFrameLayout{}}, genericDecl, callWithTypeArgs, closure) {
		t.Fatalf("expected explicit type args to keep lambda closure-env reuse when runtime generic bindings are unused")
	}

	if !canReuseLambdaClosureEnvForBytecode(&bytecodeProgram{frameLayout: &bytecodeFrameLayout{}}, decl, nil, closure) {
		t.Fatalf("expected true for non-generic slot lambda without env scopes")
	}
}
