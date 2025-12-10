package interpreter

import (
	"strings"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestImplResolutionUnionSpecificityPrefersSubset(t *testing.T) {
	buildModule := func(lit ast.Expression) *ast.Module {
		return ast.Mod([]ast.Statement{
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
				"Extra",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("String"), "label")},
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
				ast.UnionT(ast.Ty("Fancy"), ast.Ty("Basic"), ast.Ty("Extra")),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"to_string",
						[]*ast.FunctionParameter{ast.Param("self", nil)},
						[]ast.Statement{ast.Ret(ast.Str("triple"))},
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
			"Extra",
			[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("String"), "label")},
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
			ast.UnionT(ast.Ty("Fancy"), ast.Ty("Extra")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", nil)},
					[]ast.Statement{ast.Ret(ast.Str("other"))},
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
	if !strings.Contains(err.Error(), "ambiguous implementations of Show") {
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
		ast.Impl(
			"Show",
			ast.Ty("Fancy"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Fancy"))},
					[]ast.Statement{ast.Ret(ast.Str("fancy"))},
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
			ast.UnionT(ast.Ty("Fancy"), ast.Ty("Basic")),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"describe",
					[]*ast.FunctionParameter{ast.Param("self", nil)},
					[]ast.Statement{ast.Ret(ast.Str("union"))},
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
