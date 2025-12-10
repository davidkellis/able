package interpreter

import (
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestImplResolutionInterfaceInheritancePrefersDeeper(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Show",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("String"),
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
					ast.Ty("String"),
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
					ast.Ty("String"),
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
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("String"), "label")},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"FancySpecial",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("String"), "label")},
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
					ast.Ty("String"),
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
					ast.Ty("String"),
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
					ast.Ty("String"),
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
					ast.Ty("String"),
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
					ast.Ty("String"),
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
					ast.Ty("String"),
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
					ast.Ty("String"),
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
					ast.Ty("String"),
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
					ast.Ty("String"),
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
					ast.Ty("String"),
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
					ast.Ty("String"),
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
					ast.Ty("String"),
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
						ast.Ty("String"),
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
						ast.Ty("String"),
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
						ast.Ty("String"),
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
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("String"), "label")},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
			ast.StructDef(
				"Basic",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("String"), "label")},
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
						ast.Ty("String"),
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
						ast.Ty("String"),
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
						ast.Ty("String"),
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
						ast.Ty("String"),
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
						ast.Ty("String"),
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
