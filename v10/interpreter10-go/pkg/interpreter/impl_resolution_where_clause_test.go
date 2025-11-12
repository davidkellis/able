package interpreter

import (
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

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
