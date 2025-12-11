package interpreter

import (
	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
	"strings"
	"testing"
)

func TestUfcsPrefersMatchingReceiverType(t *testing.T) {
	interp := New()
	module := ast.Mod(
		[]ast.Statement{
			ast.StructDef(
				"Point",
				[]*ast.StructFieldDefinition{
					ast.FieldDef(ast.Ty("i32"), "x"),
				},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
			ast.Fn(
				"tag",
				[]*ast.FunctionParameter{
					ast.Param("p", ast.Ty("Point")),
				},
				[]ast.Statement{
					ast.Ret(ast.Str("point")),
				},
				ast.Ty("String"),
				nil,
				nil,
				false,
				false,
			),
			ast.Fn(
				"tag",
				[]*ast.FunctionParameter{
					ast.Param("s", ast.Ty("String")),
				},
				[]ast.Statement{
					ast.Ret(ast.Str("string")),
				},
				ast.Ty("String"),
				nil,
				nil,
				false,
				false,
			),
			ast.Assign(
				ast.ID("p"),
				ast.StructLit([]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(1), "x"),
				}, false, "Point", nil, nil),
			),
			ast.Assign(ast.ID("point_tag"), ast.CallExpr(ast.Member(ast.ID("p"), "tag"))),
			ast.Assign(ast.ID("string_tag"), ast.CallExpr(ast.Member(ast.Str("hi"), "tag"))),
		},
		nil,
		nil,
	)
	_, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("module evaluation failed: %v", err)
	}
	pointTag, err := env.Get("point_tag")
	if err != nil {
		t.Fatalf("env lookup failed: %v", err)
	}
	if sv, ok := pointTag.(runtime.StringValue); !ok || sv.Val != "point" {
		t.Fatalf("expected point_tag to be 'point', got %T (%v)", pointTag, pointTag)
	}
	stringTag, err := env.Get("string_tag")
	if err != nil {
		t.Fatalf("env lookup failed: %v", err)
	}
	if sv, ok := stringTag.(runtime.StringValue); !ok || sv.Val != "string" {
		t.Fatalf("expected string_tag to be 'string', got %T (%v)", stringTag, stringTag)
	}
}

func TestUfcsRejectsMismatchedReceiver(t *testing.T) {
	interp := New()
	module := ast.Mod(
		[]ast.Statement{
			ast.StructDef(
				"Point",
				[]*ast.StructFieldDefinition{
					ast.FieldDef(ast.Ty("i32"), "x"),
				},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
			ast.Fn(
				"label",
				[]*ast.FunctionParameter{
					ast.Param("s", ast.Ty("String")),
				},
				[]ast.Statement{
					ast.Ret(ast.Str("nope")),
				},
				ast.Ty("String"),
				nil,
				nil,
				false,
				false,
			),
			ast.CallExpr(ast.Member(ast.StructLit([]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Int(1), "x"),
			}, false, "Point", nil, nil), "label")),
		},
		nil,
		nil,
	)
	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected evaluation to fail for mismatched UFCS receiver")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "field or method") {
		t.Fatalf("expected member access error, got %v", err)
	}
}

func TestCallableFieldPrecedenceOverMethods(t *testing.T) {
	interp := New()
	module := ast.Mod(
		[]ast.Statement{
			ast.StructDef(
				"Box",
				[]*ast.StructFieldDefinition{
					ast.FieldDef(ast.Ty("String"), "name"),
					ast.FieldDef(nil, "action"),
				},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
			ast.Methods(
				ast.Ty("Box"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"action",
						nil,
						[]ast.Statement{ast.Ret(ast.Str("method"))},
						ast.Ty("String"),
						nil,
						nil,
						true,
						false,
					),
				},
				nil,
				nil,
			),
			ast.Assign(
				ast.ID("b"),
				ast.StructLit(
					[]*ast.StructFieldInitializer{
						ast.FieldInit(ast.Str("ok"), "name"),
						ast.FieldInit(ast.Lam(nil, ast.Str("field")), "action"),
					},
					false,
					"Box",
					nil,
					nil,
				),
			),
			ast.Assign(ast.ID("result"), ast.CallExpr(ast.Member(ast.ID("b"), "action"))),
		},
		nil,
		nil,
	)
	_, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("module evaluation failed: %v", err)
	}
	value, err := env.Get("result")
	if err != nil {
		t.Fatalf("env lookup failed: %v", err)
	}
	str, ok := value.(runtime.StringValue)
	if !ok || str.Val != "field" {
		t.Fatalf("expected callable field to win, got %T (%v)", value, value)
	}
}

func TestAmbiguousCallablePoolWithInherentAndFreeFunction(t *testing.T) {
	interp := New()
	module := ast.Mod(
		[]ast.Statement{
			ast.StructDef(
				"Point",
				[]*ast.StructFieldDefinition{
					ast.FieldDef(ast.Ty("i32"), "x"),
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
						"describe",
						[]*ast.FunctionParameter{
							ast.Param("self", ast.Ty("Point")),
						},
						[]ast.Statement{ast.Ret(ast.Str("method"))},
						ast.Ty("String"),
						nil,
						nil,
						false,
						false,
					),
				},
				nil,
				nil,
			),
			ast.Fn(
				"describe",
				[]*ast.FunctionParameter{
					ast.Param("p", ast.Ty("Point")),
				},
				[]ast.Statement{ast.Ret(ast.Str("free"))},
				ast.Ty("String"),
				nil,
				nil,
				false,
				false,
			),
			ast.Assign(ast.ID("p"), ast.StructLit([]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Int(1), "x"),
			}, false, "Point", nil, nil)),
			ast.CallExpr(ast.Member(ast.ID("p"), "describe")),
		},
		nil,
		nil,
	)
	value := mustEvalModule(t, interp, module)
	str, ok := value.(runtime.StringValue)
	if !ok {
		t.Fatalf("expected string result, got %#v", value)
	}
	if str.Val != "method" {
		t.Fatalf("expected inherent method to win, got %q", str.Val)
	}
}
