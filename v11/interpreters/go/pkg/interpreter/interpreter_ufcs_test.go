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
				ast.Ty("string"),
				nil,
				nil,
				false,
				false,
			),
			ast.Fn(
				"tag",
				[]*ast.FunctionParameter{
					ast.Param("s", ast.Ty("string")),
				},
				[]ast.Statement{
					ast.Ret(ast.Str("string")),
				},
				ast.Ty("string"),
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
					ast.Param("s", ast.Ty("string")),
				},
				[]ast.Statement{
					ast.Ret(ast.Str("nope")),
				},
				ast.Ty("string"),
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
