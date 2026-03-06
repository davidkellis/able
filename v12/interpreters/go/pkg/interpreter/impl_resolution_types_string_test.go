package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestTypeExpressionToStringStableForms(t *testing.T) {
	tests := []struct {
		name string
		expr ast.TypeExpression
		want string
	}{
		{name: "simple", expr: ast.Ty("i32"), want: "i32"},
		{name: "generic", expr: ast.Gen(ast.Ty("Array"), ast.Ty("i64")), want: "Array<i64>"},
		{name: "nullable", expr: ast.Nullable(ast.Ty("String")), want: "String?"},
		{name: "union", expr: ast.UnionT(ast.Ty("i32"), ast.Ty("String")), want: "i32 | String"},
		{name: "function", expr: ast.FnType([]ast.TypeExpression{ast.Ty("i32"), ast.Ty("i32")}, ast.Ty("i32")), want: "fn(i32, i32) -> i32"},
		{name: "nil", expr: nil, want: "<?>"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := typeExpressionToString(tc.expr); got != tc.want {
				t.Fatalf("typeExpressionToString() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestTypeInfoToStringStableForms(t *testing.T) {
	if got := typeInfoToString(typeInfo{}); got != "<unknown>" {
		t.Fatalf("typeInfoToString(empty) = %q, want <unknown>", got)
	}
	info := typeInfo{
		name: "Array",
		typeArgs: []ast.TypeExpression{
			ast.Gen(ast.Ty("Result"), ast.Ty("String")),
		},
	}
	if got := typeInfoToString(info); got != "Array<Result<String>>" {
		t.Fatalf("typeInfoToString() = %q, want %q", got, "Array<Result<String>>")
	}
}
