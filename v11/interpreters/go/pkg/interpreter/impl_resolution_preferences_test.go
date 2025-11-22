package interpreter

import (
	"strings"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestImplResolutionPrefersStricterConstraints(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Display",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("string"),
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
		ast.Iface(
			"Copyable",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"duplicate",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("Self"),
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
		ast.StructDef(
			"Item",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "value"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"Wrapper",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("T"), "value"),
			},
			ast.StructKindNamed,
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			false,
		),
		ast.Impl(
			"Display",
			ast.Ty("Item"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Item"))},
					[]ast.Statement{ast.Ret(ast.Str("Item"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Copyable",
			ast.Ty("Item"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"duplicate",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Item"))},
					[]ast.Statement{ast.Ret(ast.ID("self"))},
					ast.Ty("Item"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Display",
			ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")))},
					[]ast.Statement{ast.Ret(ast.Str("generic"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			[]*ast.GenericParameter{ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Display")))},
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Display",
			ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")))},
					[]ast.Statement{ast.Ret(ast.Str("copyable"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			[]*ast.GenericParameter{ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Display")), ast.InterfaceConstr(ast.Ty("Copyable")))},
			nil,
			nil,
			false,
		),
		ast.Assign(
			ast.ID("item"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(3), "value")},
				false,
				"Item",
				nil,
				nil,
			),
		),
		ast.Assign(
			ast.ID("wrapper"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{ast.FieldInit(ast.ID("item"), "value")},
				false,
				"Wrapper",
				nil,
				[]ast.TypeExpression{ast.Ty("Item")},
			),
		),
		ast.CallExpr(ast.Member(ast.ID("wrapper"), "to_string")),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "copyable" {
		t.Fatalf("expected copyable, got %#v", result)
	}
}

func TestImplResolutionAmbiguousMultiTraitConstraints(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Show",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("string"),
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
		ast.Iface(
			"Copyable",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"duplicate",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("Self"),
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
		ast.StructDef(
			"Point",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "x"),
				ast.FieldDef(ast.Ty("i32"), "y"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"Wrapper",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("T"), "value"),
			},
			ast.StructKindNamed,
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			false,
		),
		ast.Impl(
			"Show",
			ast.Ty("Point"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Point"))},
					[]ast.Statement{ast.Ret(ast.Str("Point"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Copyable",
			ast.Ty("Point"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"duplicate",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Point"))},
					[]ast.Statement{ast.Ret(ast.ID("self"))},
					ast.Ty("Point"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Show",
			ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")))},
					[]ast.Statement{ast.Ret(ast.Str("show-constrained"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			[]*ast.GenericParameter{ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Show")))},
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Show",
			ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")))},
					[]ast.Statement{ast.Ret(ast.Str("copy-constrained"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			[]*ast.GenericParameter{ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Copyable")))},
			nil,
			nil,
			false,
		),
		ast.Assign(
			ast.ID("point"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(3), "x"),
					ast.FieldInit(ast.Int(4), "y"),
				},
				false,
				"Point",
				nil,
				nil,
			),
		),
		ast.Assign(
			ast.ID("wrapper"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{ast.FieldInit(ast.ID("point"), "value")},
				false,
				"Wrapper",
				nil,
				[]ast.TypeExpression{ast.Ty("Point")},
			),
		),
		ast.CallExpr(ast.Member(ast.ID("wrapper"), "to_string")),
	}, nil, nil)

	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected ambiguity error, got nil")
	}
	if !strings.Contains(err.Error(), "Ambiguous method 'to_string'") {
		t.Fatalf("expected ambiguous method error, got %v", err)
	}
}

func TestImplResolutionInherentMethodsTakePrecedence(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Speakable",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"speak",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("string"),
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
		ast.StructDef(
			"Bot",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("i32"), "id")},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Bot"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"speak",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Bot"))},
					[]ast.Statement{ast.Ret(ast.Str("beep inherent"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
		),
		ast.Impl(
			"Speakable",
			ast.Ty("Bot"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"speak",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Bot"))},
					[]ast.Statement{ast.Ret(ast.Str("beep impl"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.CallExpr(
			ast.Member(
				ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(42), "id")}, false, "Bot", nil, nil),
				"speak",
			),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "beep inherent" {
		t.Fatalf("expected beep inherent, got %#v", result)
	}
}

func TestImplResolutionMoreSpecificImplWinsOverGeneric(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Show",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("string"),
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
		ast.StructDef(
			"Point",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "x"),
				ast.FieldDef(ast.Ty("i32"), "y"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Point"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Point"))},
					[]ast.Statement{ast.Ret(ast.Str("Point inherent"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
		),
		ast.StructDef(
			"Wrapper",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("T"), "value")},
			ast.StructKindNamed,
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			false,
		),
		ast.Impl(
			"Show",
			ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")))},
					[]ast.Statement{ast.Ret(ast.Str("generic"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			[]*ast.GenericParameter{ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Show")))},
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Show",
			ast.Gen(ast.Ty("Wrapper"), ast.Ty("Point")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Wrapper"), ast.Ty("Point")))},
					[]ast.Statement{ast.Ret(ast.Str("specific"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.CallExpr(
			ast.Member(
				ast.StructLit(
					[]*ast.StructFieldInitializer{
						ast.FieldInit(
							ast.StructLit(
								[]*ast.StructFieldInitializer{
									ast.FieldInit(ast.Int(1), "x"),
									ast.FieldInit(ast.Int(2), "y"),
								},
								false,
								"Point",
								nil,
								nil,
							),
							"value",
						),
					},
					false,
					"Wrapper",
					nil,
					[]ast.TypeExpression{ast.Ty("Point")},
				),
				"to_string",
			),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "specific" {
		t.Fatalf("expected specific, got %#v", result)
	}
}

func TestMethodLookupSkipsConstraintsWhenMethodMissing(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Writable",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"write",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("string"),
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
		ast.Iface(
			"Show",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("string"),
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
		ast.StructDef(
			"Wrap",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("T"), "value")},
			ast.StructKindNamed,
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			false,
		),
		ast.Impl(
			"Show",
			ast.Gen(ast.Ty("Wrap"), ast.Ty("T")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Wrap"), ast.Ty("T")))},
					[]ast.Statement{ast.Ret(ast.Str("wrapped"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			[]*ast.GenericParameter{ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Writable")))},
			nil,
			nil,
			false,
		),
		ast.Assign(
			ast.ID("w"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(1), "value")},
				false,
				"Wrap",
				nil,
				[]ast.TypeExpression{ast.Ty("i32")},
			),
		),
		ast.CallExpr(ast.Member(ast.ID("w"), "missing")),
	}, nil, nil)

	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected missing method error")
	}
	if strings.Contains(err.Error(), "does not satisfy interface") {
		t.Fatalf("constraint error should be ignored for missing method, got %v", err)
	}
	if !strings.Contains(err.Error(), "No field or method named 'missing'") {
		t.Fatalf("expected missing method error, got %v", err)
	}
}
