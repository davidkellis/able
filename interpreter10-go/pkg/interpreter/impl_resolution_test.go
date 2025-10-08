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

func TestImplResolutionWhereClauseSupersetPreferred(t *testing.T) {
	buildModule := func(wrapper ast.Expression) *ast.Module {
		return ast.Mod([]ast.Statement{
			ast.Iface(
				"Readable",
				[]*ast.FunctionSignature{
					ast.FnSig(
						"read",
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
				"Fancy",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
			ast.StructDef(
				"Basic",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
				ast.StructKindNamed,
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
				"Readable",
				ast.Ty("Fancy"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"read",
						[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Fancy"))},
						[]ast.Statement{ast.Ret(ast.Str("read-fancy"))},
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
				"Writable",
				ast.Ty("Fancy"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"write",
						[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Fancy"))},
						[]ast.Statement{ast.Ret(ast.Str("write-fancy"))},
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
				"Readable",
				ast.Ty("Basic"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"read",
						[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Basic"))},
						[]ast.Statement{ast.Ret(ast.Str("read-basic"))},
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
				ast.Gen(ast.Ty("Wrap"), ast.Ty("T")),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"to_string",
						[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Wrap"), ast.Ty("T")))},
						[]ast.Statement{
							ast.Ret(
								ast.CallExpr(ast.Member(ast.Member(ast.ID("self"), "value"), "read")),
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
				[]*ast.GenericParameter{ast.GenericParam("T")},
				nil,
				[]*ast.WhereClauseConstraint{
					ast.WhereConstraint("T", ast.InterfaceConstr(ast.Ty("Readable"))),
				},
				false,
			),
			ast.Impl(
				"Show",
				ast.Gen(ast.Ty("Wrap"), ast.Ty("T")),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"to_string",
						[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Wrap"), ast.Ty("T")))},
						[]ast.Statement{
							ast.Ret(
								ast.CallExpr(ast.Member(ast.Member(ast.ID("self"), "value"), "write")),
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
				[]*ast.GenericParameter{ast.GenericParam("T")},
				nil,
				[]*ast.WhereClauseConstraint{
					ast.WhereConstraint("T", ast.InterfaceConstr(ast.Ty("Readable")), ast.InterfaceConstr(ast.Ty("Writable"))),
				},
				false,
			),
			ast.Assign(ast.ID("wrapper"), wrapper),
			ast.CallExpr(ast.Member(ast.ID("wrapper"), "to_string")),
		}, nil, nil)
	}

	fancyInner := ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("f"), "label")}, false, "Fancy", nil, nil)
	fancyWrapper := ast.StructLit(
		[]*ast.StructFieldInitializer{ast.FieldInit(fancyInner, "value")},
		false,
		"Wrap",
		nil,
		[]ast.TypeExpression{ast.Ty("Fancy")},
	)
	moduleFancy := buildModule(fancyWrapper)
	interpFancy := New()
	resultFancy, _, err := interpFancy.EvaluateModule(moduleFancy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	strFancy, ok := resultFancy.(runtime.StringValue)
	if !ok || strFancy.Val != "write-fancy" {
		t.Fatalf("expected write-fancy, got %#v", resultFancy)
	}

	basicInner := ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("b"), "label")}, false, "Basic", nil, nil)
	basicWrapper := ast.StructLit(
		[]*ast.StructFieldInitializer{ast.FieldInit(basicInner, "value")},
		false,
		"Wrap",
		nil,
		[]ast.TypeExpression{ast.Ty("Basic")},
	)
	moduleBasic := buildModule(basicWrapper)
	interpBasic := New()
	resultBasic, _, err := interpBasic.EvaluateModule(moduleBasic)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	strBasic, ok := resultBasic.(runtime.StringValue)
	if !ok || strBasic.Val != "read-basic" {
		t.Fatalf("expected read-basic, got %#v", resultBasic)
	}
}

func TestImplResolutionWhereClauseMultiParamPreferred(t *testing.T) {
	buildModule := func(pair ast.Expression) *ast.Module {
		return ast.Mod([]ast.Statement{
			ast.Iface(
				"Readable",
				[]*ast.FunctionSignature{
					ast.FnSig(
						"read",
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
				"Combine",
				[]*ast.FunctionSignature{
					ast.FnSig(
						"combine",
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
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
			ast.StructDef(
				"Basic",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
			ast.StructDef(
				"Pair",
				[]*ast.StructFieldDefinition{
					ast.FieldDef(ast.Ty("A"), "left"),
					ast.FieldDef(ast.Ty("B"), "right"),
				},
				ast.StructKindNamed,
				[]*ast.GenericParameter{ast.GenericParam("A"), ast.GenericParam("B")},
				nil,
				false,
			),
			ast.Impl(
				"Readable",
				ast.Ty("Fancy"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"read",
						[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Fancy"))},
						[]ast.Statement{ast.Ret(ast.Str("read-fancy"))},
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
				"Writable",
				ast.Ty("Fancy"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"write",
						[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Fancy"))},
						[]ast.Statement{ast.Ret(ast.Str("write-fancy"))},
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
				"Readable",
				ast.Ty("Basic"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"read",
						[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Basic"))},
						[]ast.Statement{ast.Ret(ast.Str("read-basic"))},
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
				"Combine",
				ast.Gen(ast.Ty("Pair"), ast.Ty("A"), ast.Ty("B")),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"combine",
						[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Pair"), ast.Ty("A"), ast.Ty("B")))},
						[]ast.Statement{
							ast.Ret(
								ast.Bin(
									"+",
									ast.CallExpr(ast.Member(ast.Member(ast.ID("self"), "left"), "read")),
									ast.CallExpr(ast.Member(ast.Member(ast.ID("self"), "right"), "read")),
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
				[]*ast.GenericParameter{ast.GenericParam("A"), ast.GenericParam("B")},
				nil,
				[]*ast.WhereClauseConstraint{
					ast.WhereConstraint("A", ast.InterfaceConstr(ast.Ty("Readable"))),
					ast.WhereConstraint("B", ast.InterfaceConstr(ast.Ty("Readable"))),
				},
				false,
			),
			ast.Impl(
				"Combine",
				ast.Gen(ast.Ty("Pair"), ast.Ty("A"), ast.Ty("B")),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"combine",
						[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Pair"), ast.Ty("A"), ast.Ty("B")))},
						[]ast.Statement{
							ast.Ret(
								ast.CallExpr(ast.Member(ast.Member(ast.ID("self"), "right"), "write")),
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
				[]*ast.GenericParameter{ast.GenericParam("A"), ast.GenericParam("B")},
				nil,
				[]*ast.WhereClauseConstraint{
					ast.WhereConstraint("A", ast.InterfaceConstr(ast.Ty("Readable"))),
					ast.WhereConstraint("B", ast.InterfaceConstr(ast.Ty("Readable")), ast.InterfaceConstr(ast.Ty("Writable"))),
				},
				false,
			),
			ast.Assign(ast.ID("pair"), pair),
			ast.CallExpr(ast.Member(ast.ID("pair"), "combine")),
		}, nil, nil)
	}

	fancy := ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("f"), "label")}, false, "Fancy", nil, nil)
	basic := ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("b"), "label")}, false, "Basic", nil, nil)
	fancyPair := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.FieldInit(fancy, "left"),
			ast.FieldInit(fancy, "right"),
		},
		false,
		"Pair",
		nil,
		[]ast.TypeExpression{ast.Ty("Fancy"), ast.Ty("Fancy")},
	)
	moduleFancy := buildModule(fancyPair)
	interpFancy := New()
	resultFancy, _, err := interpFancy.EvaluateModule(moduleFancy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	strFancy, ok := resultFancy.(runtime.StringValue)
	if !ok || strFancy.Val != "write-fancy" {
		t.Fatalf("expected write-fancy, got %#v", resultFancy)
	}

	mixedPair := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.FieldInit(fancy, "left"),
			ast.FieldInit(basic, "right"),
		},
		false,
		"Pair",
		nil,
		[]ast.TypeExpression{ast.Ty("Fancy"), ast.Ty("Basic")},
	)
	moduleMixed := buildModule(mixedPair)
	interpMixed := New()
	resultMixed, _, err := interpMixed.EvaluateModule(moduleMixed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	strMixed, ok := resultMixed.(runtime.StringValue)
	if !ok || strMixed.Val != "read-fancyread-basic" {
		t.Fatalf("expected read-fancyread-basic, got %#v", resultMixed)
	}
}

func TestImplResolutionUnionSpecificityPrefersSubset(t *testing.T) {
	buildModule := func(lit ast.Expression) *ast.Module {
		return ast.Mod([]ast.Statement{
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
				"Fancy",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
			ast.StructDef(
				"Basic",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
			ast.StructDef(
				"Extra",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
			ast.Impl(
				"Show",
				ast.UnionT(ast.Ty("Fancy"), ast.Ty("Basic")),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"to_string",
						[]*ast.FunctionParameter{ast.Param("self", nil)},
						[]ast.Statement{ast.Ret(ast.Str("pair"))},
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
				ast.UnionT(ast.Ty("Fancy"), ast.Ty("Basic"), ast.Ty("Extra")),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"to_string",
						[]*ast.FunctionParameter{ast.Param("self", nil)},
						[]ast.Statement{ast.Ret(ast.Str("triple"))},
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
			ast.CallExpr(ast.Member(lit, "to_string")),
		}, nil, nil)
	}

	fancyLiteral := ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("f"), "label")}, false, "Fancy", nil, nil)
	moduleFancy := buildModule(fancyLiteral)
	interpFancy := New()
	resultFancy, _, err := interpFancy.EvaluateModule(moduleFancy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	strFancy, ok := resultFancy.(runtime.StringValue)
	if !ok || strFancy.Val != "pair" {
		t.Fatalf("expected pair, got %#v", resultFancy)
	}

	extraLiteral := ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("e"), "label")}, false, "Extra", nil, nil)
	moduleExtra := buildModule(extraLiteral)
	interpExtra := New()
	resultExtra, _, err := interpExtra.EvaluateModule(moduleExtra)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	strExtra, ok := resultExtra.(runtime.StringValue)
	if !ok || strExtra.Val != "triple" {
		t.Fatalf("expected triple, got %#v", resultExtra)
	}
}

func TestImplResolutionUnionAmbiguousWithoutSubset(t *testing.T) {
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
			"Fancy",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"Basic",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"Extra",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Show",
			ast.UnionT(ast.Ty("Fancy"), ast.Ty("Basic")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", nil)},
					[]ast.Statement{ast.Ret(ast.Str("pair"))},
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
			ast.UnionT(ast.Ty("Fancy"), ast.Ty("Extra")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", nil)},
					[]ast.Statement{ast.Ret(ast.Str("other"))},
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
		ast.CallExpr(ast.Member(
			ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("f"), "label")}, false, "Fancy", nil, nil),
			"to_string",
		)),
	}, nil, nil)

	interp := New()
	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected ambiguity error")
	}
	if !strings.Contains(err.Error(), "Ambiguous method 'to_string'") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInterfaceDynamicValueUsesUnionImpl(t *testing.T) {
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
			"Fancy",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"Basic",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
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
					"to_string",
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
			ast.TypedP(ast.PatternFrom("item"), ast.Ty("Show")),
			ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("f"), "label")}, false, "Fancy", nil, nil),
		),
		ast.CallExpr(ast.Member(ast.ID("item"), "describe")),
	}, nil, nil)

	interp := New()
	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "union" {
		t.Fatalf("expected union, got %#v", result)
	}
}

func TestImplResolutionInterfaceInheritancePrefersDeeper(t *testing.T) {
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
			"FancyShow",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"fancy",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("string"),
					nil,
					nil,
					nil,
				),
			},
			nil,
			ast.Ty("Show"),
			nil,
			[]ast.TypeExpression{ast.Ty("Show")},
			false,
		),
		ast.Iface(
			"ShinyShow",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"shine",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("string"),
					nil,
					nil,
					nil,
				),
			},
			nil,
			ast.Ty("Show"),
			nil,
			[]ast.TypeExpression{ast.Ty("FancyShow")},
			false,
		),
		ast.StructDef(
			"FancyBase",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"FancySpecial",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
			ast.StructKindNamed,
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
			ast.Ty("FancyBase"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("FancyBase"))},
					[]ast.Statement{ast.Ret(ast.Str("base"))},
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
			"FancyShow",
			ast.Ty("FancyBase"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"fancy",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("FancyBase"))},
					[]ast.Statement{ast.Ret(ast.Str("fancy-base"))},
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
			ast.Ty("FancySpecial"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("FancySpecial"))},
					[]ast.Statement{ast.Ret(ast.Str("special"))},
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
			"FancyShow",
			ast.Ty("FancySpecial"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"fancy",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("FancySpecial"))},
					[]ast.Statement{ast.Ret(ast.Str("fancy-special"))},
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
			"ShinyShow",
			ast.Ty("FancySpecial"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"shine",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("FancySpecial"))},
					[]ast.Statement{ast.Ret(ast.Str("shine"))},
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
			ast.Gen(ast.Ty("Wrap"), ast.Ty("T")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Wrap"))},
					[]ast.Statement{
						ast.Ret(
							ast.CallExpr(ast.Member(ast.Member(ast.ID("self"), "value"), "fancy")),
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
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			[]*ast.WhereClauseConstraint{
				ast.WhereConstraint("T", ast.InterfaceConstr(ast.Ty("FancyShow"))),
			},
			false,
		),
		ast.Impl(
			"Show",
			ast.Gen(ast.Ty("Wrap"), ast.Ty("T")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Wrap"))},
					[]ast.Statement{
						ast.Ret(
							ast.CallExpr(ast.Member(ast.Member(ast.ID("self"), "value"), "shine")),
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
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			[]*ast.WhereClauseConstraint{
				ast.WhereConstraint("T", ast.InterfaceConstr(ast.Ty("ShinyShow"))),
			},
			false,
		),
		ast.CallExpr(ast.Member(
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(
						ast.StructLit(
							[]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("s"), "label")},
							false,
							"FancySpecial",
							nil,
							nil,
						),
						"value",
					),
				},
				false,
				"Wrap",
				nil,
				[]ast.TypeExpression{ast.Ty("FancySpecial")},
			),
			"to_string",
		)),
	}, nil, nil)

	interp := New()
	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "shine" {
		t.Fatalf("expected shine, got %#v", result)
	}
}

func TestImplResolutionNestedGenericConstraintsPreferred(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Readable",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"read",
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
			"Comparable",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"cmp",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self")), ast.Param("other", ast.Ty("Self"))},
					ast.Ty("i32"),
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
					"show",
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
			"Container",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("T"), "value")},
			ast.StructKindNamed,
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			false,
		),
		ast.StructDef(
			"FancyNum",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("i32"), "value")},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Readable",
			ast.Ty("FancyNum"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"read",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("FancyNum"))},
					[]ast.Statement{
						ast.Ret(
							ast.Interp(
								ast.Str("#"),
								ast.Member(ast.ID("self"), "value"),
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
		ast.Impl(
			"Comparable",
			ast.Ty("FancyNum"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"cmp",
					[]*ast.FunctionParameter{
						ast.Param("self", ast.Ty("FancyNum")),
						ast.Param("other", ast.Ty("FancyNum")),
					},
					[]ast.Statement{ast.Ret(ast.Int(0))},
					ast.Ty("i32"),
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
			ast.Gen(ast.Ty("Container"), ast.Ty("T")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"show",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Container"))},
					[]ast.Statement{
						ast.Ret(
							ast.CallExpr(ast.Member(ast.Member(ast.ID("self"), "value"), "read")),
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
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			[]*ast.WhereClauseConstraint{
				ast.WhereConstraint("T", ast.InterfaceConstr(ast.Ty("Readable"))),
			},
			false,
		),
		ast.Impl(
			"Show",
			ast.Gen(ast.Ty("Container"), ast.Ty("T")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"show",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Container"))},
					[]ast.Statement{ast.Ret(ast.Str("comparable"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			[]*ast.WhereClauseConstraint{
				ast.WhereConstraint("T", ast.InterfaceConstr(ast.Ty("Readable")), ast.InterfaceConstr(ast.Ty("Comparable"))),
			},
			false,
		),
		ast.CallExpr(ast.Member(
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(
						ast.StructLit(
							[]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(42), "value")},
							false,
							"FancyNum",
							nil,
							nil,
						),
						"value",
					),
				},
				false,
				"Container",
				nil,
				[]ast.TypeExpression{ast.Ty("FancyNum")},
			),
			"show",
		)),
	}, nil, nil)

	interp := New()
	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "comparable" {
		t.Fatalf("expected comparable, got %#v", result)
	}
}

func TestImplResolutionSupersetConstraintsPreferred(t *testing.T) {
	buildModule := func(wrapper ast.Expression) *ast.Module {
		return ast.Mod([]ast.Statement{
			ast.Iface(
				"TraitA",
				[]*ast.FunctionSignature{
					ast.FnSig(
						"trait_a",
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
				"TraitB",
				[]*ast.FunctionSignature{
					ast.FnSig(
						"trait_b",
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
				"Fancy",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
			ast.StructDef(
				"Basic",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("string"), "label")},
				ast.StructKindNamed,
				nil,
				nil,
				false,
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
				"TraitA",
				ast.Ty("Fancy"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"trait_a",
						[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Fancy"))},
						[]ast.Statement{ast.Ret(ast.Str("A:Fancy"))},
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
				"TraitB",
				ast.Ty("Fancy"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"trait_b",
						[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Fancy"))},
						[]ast.Statement{ast.Ret(ast.Str("B:Fancy"))},
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
				"TraitA",
				ast.Ty("Basic"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"trait_a",
						[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Basic"))},
						[]ast.Statement{ast.Ret(ast.Str("A:Basic"))},
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
				ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"to_string",
						[]*ast.FunctionParameter{ast.Param("self", ast.Gen(ast.Ty("Wrapper"), ast.Ty("T")))},
						[]ast.Statement{
							ast.Ret(
								ast.CallExpr(ast.Member(ast.Member(ast.ID("self"), "value"), "trait_a")),
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
				[]*ast.GenericParameter{ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("TraitA")))},
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
						[]ast.Statement{
							ast.Ret(
								ast.CallExpr(ast.Member(ast.Member(ast.ID("self"), "value"), "trait_b")),
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
				[]*ast.GenericParameter{ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("TraitA")), ast.InterfaceConstr(ast.Ty("TraitB")))},
				nil,
				nil,
				false,
			),
			ast.Assign(ast.ID("wrapper"), wrapper),
			ast.CallExpr(ast.Member(ast.ID("wrapper"), "to_string")),
		}, nil, nil)
	}

	fancyInner := ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("f"), "label")}, false, "Fancy", nil, nil)
	fancyWrapper := ast.StructLit(
		[]*ast.StructFieldInitializer{ast.FieldInit(fancyInner, "value")},
		false,
		"Wrapper",
		nil,
		[]ast.TypeExpression{ast.Ty("Fancy")},
	)
	moduleFancy := buildModule(fancyWrapper)
	interpFancy := New()
	resultFancy, _, err := interpFancy.EvaluateModule(moduleFancy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	strFancy, ok := resultFancy.(runtime.StringValue)
	if !ok || strFancy.Val != "B:Fancy" {
		t.Fatalf("expected B:Fancy, got %#v", resultFancy)
	}

	basicInner := ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("b"), "label")}, false, "Basic", nil, nil)
	basicWrapper := ast.StructLit(
		[]*ast.StructFieldInitializer{ast.FieldInit(basicInner, "value")},
		false,
		"Wrapper",
		nil,
		[]ast.TypeExpression{ast.Ty("Basic")},
	)
	moduleBasic := buildModule(basicWrapper)
	interpBasic := New()
	resultBasic, _, err := interpBasic.EvaluateModule(moduleBasic)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	strBasic, ok := resultBasic.(runtime.StringValue)
	if !ok || strBasic.Val != "A:Basic" {
		t.Fatalf("expected A:Basic, got %#v", resultBasic)
	}
}
