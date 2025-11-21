package interpreter

import (
	"strings"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestInterfaceDynamicDispatch(t *testing.T) {
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
		ast.Impl(
			"Display",
			ast.Ty("Point"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Point"))},
					[]ast.Statement{ast.Ret(ast.Str("point"))},
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
		ast.Assign(
			ast.TypedP(ast.PatternFrom("value"), ast.Ty("Display")),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(2), "x"),
					ast.FieldInit(ast.Int(3), "y"),
				},
				false,
				"Point",
				nil,
				nil,
			),
		),
		ast.CallExpr(ast.Member(ast.ID("value"), "to_string")),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "point" {
		t.Fatalf("expected interface dispatch to return 'point', got %#v", result)
	}
}

func TestInterfaceAssignmentMissingImplementation(t *testing.T) {
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
		ast.Assign(
			ast.TypedP(ast.PatternFrom("value"), ast.Ty("Display")),
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
		),
	}, nil, nil)

	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected failure when assigning struct without impl to interface")
	}
	if got := err.Error(); !strings.Contains(got, "Typed pattern mismatch in assignment") || !strings.Contains(got, "expected Display") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestInterfaceUnionDispatchPrefersSpecific(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Show",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"describe",
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
			"Fancy",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("string"), "label"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"Basic",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("string"), "label"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Show",
			ast.Ty("Fancy"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"describe",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Fancy"))},
					[]ast.Statement{ast.Ret(ast.Str("fancy"))},
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
			"Show",
			ast.Ty("Basic"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"describe",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Basic"))},
					[]ast.Statement{ast.Ret(ast.Str("basic"))},
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
			"Show",
			ast.UnionT(ast.Ty("Fancy"), ast.Ty("Basic")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"describe",
					[]*ast.FunctionParameter{ast.Param("self", nil)},
					[]ast.Statement{ast.Ret(ast.Str("union"))},
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
		ast.Assign(
			ast.ID("items"),
			ast.Arr(
				ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("f"), "label")}, false, "Fancy", nil, nil),
				ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("b"), "label")}, false, "Basic", nil, nil),
			),
		),
		ast.Assign(ast.ID("buffer"), ast.Str("")),
		ast.ForLoopPattern(
			ast.TypedP(ast.PatternFrom("item"), ast.Ty("Show")),
			ast.ID("items"),
			ast.Block(
				ast.AssignOp(
					ast.AssignmentAssign,
					ast.ID("buffer"),
					ast.Bin(
						"+",
						ast.ID("buffer"),
						ast.CallExpr(ast.Member(ast.ID("item"), "describe")),
					),
				),
			),
		),
		ast.ID("buffer"),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "fancybasic" {
		t.Fatalf("expected fancybasic, got %#v", result)
	}
}

func TestInterfaceDefaultMethodFallback(t *testing.T) {
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
					ast.Block(ast.Ret(ast.Str("default"))),
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
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("string"), "name"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Speakable",
			ast.Ty("Bot"),
			[]*ast.FunctionDefinition{},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.CallExpr(
			ast.Member(
				ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("Beep"), "name")}, false, "Bot", nil, nil),
				"speak",
			),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "default" {
		t.Fatalf("expected default, got %#v", result)
	}
}

func TestInterfaceDefaultMethodOverride(t *testing.T) {
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
					ast.Block(ast.Ret(ast.Str("default"))),
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
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("string"), "name"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Speakable",
			ast.Ty("Bot"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"speak",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Bot"))},
					[]ast.Statement{ast.Ret(ast.Str("custom"))},
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
				ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("Beep"), "name")}, false, "Bot", nil, nil),
				"speak",
			),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "custom" {
		t.Fatalf("expected custom, got %#v", result)
	}
}

func TestInterfaceDynamicDispatchUsesUnderlyingImpl(t *testing.T) {
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
		ast.Impl(
			"Display",
			ast.Ty("Point"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Point"))},
					[]ast.Statement{
						ast.Ret(
							ast.Interp(
								ast.Str("Point("),
								ast.Member(ast.ID("self"), "x"),
								ast.Str(", "),
								ast.Member(ast.ID("self"), "y"),
								ast.Str(")"),
							),
						),
					},
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
		ast.Assign(
			ast.TypedP(ast.PatternFrom("value"), ast.Ty("Display")),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(2), "x"),
					ast.FieldInit(ast.Int(3), "y"),
				},
				false,
				"Point",
				nil,
				nil,
			),
		),
		ast.CallExpr(ast.Member(ast.ID("value"), "to_string")),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "Point(2, 3)" {
		t.Fatalf("expected Point(2, 3), got %#v", result)
	}
}

func TestInterfaceDynamicDispatchPrefersMostSpecificImpl(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Show",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"describe",
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
			"Fancy",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("string"), "label"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"Basic",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("string"), "label"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Show",
			ast.Ty("Fancy"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"describe",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Fancy"))},
					[]ast.Statement{ast.Ret(ast.Str("fancy"))},
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
			"Show",
			ast.Ty("Basic"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"describe",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Basic"))},
					[]ast.Statement{ast.Ret(ast.Str("basic"))},
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
			"Show",
			ast.UnionT(ast.Ty("Fancy"), ast.Ty("Basic")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"describe",
					[]*ast.FunctionParameter{ast.Param("self", nil)},
					[]ast.Statement{ast.Ret(ast.Str("union"))},
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
		ast.Assign(
			ast.ID("items"),
			ast.Arr(
				ast.StructLit(
					[]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("f"), "label")},
					false,
					"Fancy",
					nil,
					nil,
				),
				ast.StructLit(
					[]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("b"), "label")},
					false,
					"Basic",
					nil,
					nil,
				),
			),
		),
		ast.Assign(ast.ID("buffer"), ast.Str("")),
		ast.ForLoopPattern(
			ast.TypedP(ast.PatternFrom("item"), ast.Ty("Show")),
			ast.ID("items"),
			ast.Block(
				ast.AssignOp(
					ast.AssignmentAssign,
					ast.ID("buffer"),
					ast.Bin(
						"+",
						ast.ID("buffer"),
						ast.CallExpr(ast.Member(ast.ID("item"), "describe")),
					),
				),
			),
		),
		ast.ID("buffer"),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "fancybasic" {
		t.Fatalf("expected fancybasic, got %#v", result)
	}
}
