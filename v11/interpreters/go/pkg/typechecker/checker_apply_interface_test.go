package typechecker

import (
	"strings"
	"testing"

	"able/interpreter10-go/pkg/ast"
)

func makeApplyInterface() *ast.InterfaceDefinition {
	args := ast.GenericParam("Args", nil)
	result := ast.GenericParam("Result", nil)
	return ast.Iface(
		"Apply",
		[]*ast.FunctionSignature{
			ast.FnSig(
				"apply",
				[]*ast.FunctionParameter{
					ast.Param("self", ast.Ty("Self")),
					ast.Param("args", ast.Ty("Args")),
				},
				ast.Ty("Result"),
				nil,
				nil,
				nil,
			),
		},
		[]*ast.GenericParameter{args, result},
		nil,
		nil,
		nil,
		false,
	)
}

func TestApplyImplementationCallable(t *testing.T) {
	checker := New()
	applyIface := makeApplyInterface()
	multStruct := ast.StructDef(
		"Multiplier",
		[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("i32"), "factor")},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	applyFn := ast.Fn(
		"apply",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("input", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Ret(ast.Bin("*", ast.ImplicitMember("factor"), ast.ID("input"))),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	applyImpl := ast.Impl(
		"Apply",
		ast.Ty("Multiplier"),
		[]*ast.FunctionDefinition{applyFn},
		nil,
		nil,
		[]ast.TypeExpression{ast.Ty("i32"), ast.Ty("i32")},
		nil,
		false,
	)
	assign := ast.Assign(
		ast.ID("m"),
		ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(2), "factor")}, false, "Multiplier", nil, nil),
	)
	call := ast.CallExpr(ast.ID("m"), ast.Int(3))
	module := ast.NewModule([]ast.Statement{applyIface, multStruct, applyImpl, assign, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestApplyMissingImplementation(t *testing.T) {
	checker := New()
	applyIface := makeApplyInterface()
	boxStruct := ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("i32"), "value")},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	assign := ast.Assign(ast.ID("b"), ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(1), "value")}, false, "Box", nil, nil))
	call := ast.CallExpr(ast.ID("b"), nil)
	module := ast.NewModule([]ast.Statement{applyIface, boxStruct, assign, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostics for missing Apply implementation, got none")
	}
	found := false
	for _, diag := range diags {
		if strings.Contains(diag.Message, "Apply") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected Apply diagnostic, got %v", diags)
	}
}

func TestApplyImplementationArgumentMismatch(t *testing.T) {
	checker := New()
	applyIface := makeApplyInterface()
	appenderStruct := ast.StructDef(
		"Appender",
		[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("String"), "prefix")},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	applyFn := ast.Fn(
		"apply",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("suffix", ast.Ty("String")),
		},
		[]ast.Statement{
			ast.Ret(ast.Bin("+", ast.ImplicitMember("prefix"), ast.ID("suffix"))),
		},
		ast.Ty("String"),
		nil,
		nil,
		false,
		false,
	)
	applyImpl := ast.Impl(
		"Apply",
		ast.Ty("Appender"),
		[]*ast.FunctionDefinition{applyFn},
		nil,
		nil,
		[]ast.TypeExpression{ast.Ty("String"), ast.Ty("String")},
		nil,
		false,
	)
	assign := ast.Assign(
		ast.ID("a"),
		ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("hi "), "prefix")}, false, "Appender", nil, nil),
	)
	call := ast.CallExpr(ast.ID("a"), ast.Int(1))
	module := ast.NewModule([]ast.Statement{applyIface, appenderStruct, applyImpl, assign, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostics for argument mismatch, got none")
	}
	found := false
	for _, diag := range diags {
		if strings.Contains(diag.Message, "argument 1") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected argument mismatch diagnostic, got %v", diags)
	}
}
