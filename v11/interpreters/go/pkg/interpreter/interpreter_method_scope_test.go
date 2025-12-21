package interpreter

import (
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func methodScopePackageModule() *ast.Module {
	counterStruct := ast.StructDef(
		"Widget",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "value"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)

	bumpMethod := ast.Fn(
		"bump",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("delta", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Ret(ast.Bin("+", ast.Member(ast.ID("self"), "value"), ast.ID("delta"))),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	augmentFn := ast.Fn(
		"augment",
		[]*ast.FunctionParameter{
			ast.Param("item", ast.Ty("Widget")),
			ast.Param("delta", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Ret(ast.Bin("+", ast.Member(ast.ID("item"), "value"), ast.ID("delta"))),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	makeFn := ast.Fn(
		"make",
		[]*ast.FunctionParameter{
			ast.Param("start", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Ret(ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.ID("start"), "value"),
				},
				false,
				"Widget",
				nil,
				nil,
			)),
		},
		ast.Ty("Widget"),
		nil,
		nil,
		false,
		false,
	)

	methods := ast.Methods(
		ast.Ty("Widget"),
		[]*ast.FunctionDefinition{bumpMethod, augmentFn, makeFn},
		nil,
		nil,
	)

	return ast.Mod([]ast.Statement{counterStruct, methods}, nil, ast.Pkg([]interface{}{"pkgmethods"}, false))
}

func TestMethodCallRequiresImportedName(t *testing.T) {
	i := New()
	if _, _, err := i.EvaluateModule(methodScopePackageModule()); err != nil {
		t.Fatalf("package setup failed: %v", err)
	}

	entry := ast.Mod(
		[]ast.Statement{
			ast.Assign(
				ast.ID("inst"),
				ast.CallExpr(
					ast.Member(
						ast.Member(ast.ID("pkgmethods"), "Widget"),
						"make",
					),
					ast.Int(3),
				),
			),
			ast.CallExpr(ast.Member(ast.ID("inst"), "bump"), ast.Int(2)),
		},
		[]*ast.ImportStatement{
			ast.Imp([]interface{}{"pkgmethods"}, false, nil, nil),
		},
		nil,
	)

	if _, _, err := i.EvaluateModule(entry); err == nil {
		t.Fatalf("expected method lookup to fail without importing name")
	} else if got := err.Error(); got != "No field or method named 'bump'" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTypeQualifiedFunctionsExcludedFromUFCS(t *testing.T) {
	i := New()
	if _, _, err := i.EvaluateModule(methodScopePackageModule()); err != nil {
		t.Fatalf("package setup failed: %v", err)
	}

	entry := ast.Mod(
		[]ast.Statement{
			ast.Assign(
				ast.ID("inst"),
				ast.CallExpr(
					ast.Member(
						ast.Member(ast.ID("pkgmethods"), "Widget"),
						"make",
					),
					ast.Int(4),
				),
			),
			ast.Assign(
				ast.ID("augment"),
				ast.Member(ast.ID("pkgmethods"), "Widget.augment"),
			),
			ast.Assign(
				ast.ID("direct"),
				ast.CallExpr(
					ast.Member(
						ast.Member(ast.ID("pkgmethods"), "Widget"),
						"augment",
					),
					ast.ID("inst"),
					ast.Int(5),
				),
			),
			ast.CallExpr(ast.Member(ast.ID("inst"), "augment"), ast.Int(1)),
		},
		[]*ast.ImportStatement{
			ast.Imp([]interface{}{"pkgmethods"}, false, nil, nil),
		},
		nil,
	)

	if _, _, err := i.EvaluateModule(entry); err == nil {
		t.Fatalf("expected UFCS lookup of type-qualified function to fail")
	} else if got := err.Error(); got != "No field or method named 'augment'" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWildcardImportProvidesTypeQualifiedSymbol(t *testing.T) {
	i := New()
	if _, _, err := i.EvaluateModule(methodScopePackageModule()); err != nil {
		t.Fatalf("package setup failed: %v", err)
	}

	entry := ast.Mod(
		[]ast.Statement{
			ast.Assign(
				ast.ID("inst"),
				ast.CallExpr(ast.Member(ast.ID("Widget"), "make"), ast.Int(11)),
			),
			ast.Member(ast.ID("inst"), "value"),
		},
		[]*ast.ImportStatement{
			ast.Imp([]interface{}{"pkgmethods"}, true, nil, nil),
		},
		nil,
	)

	val, _, err := i.EvaluateModule(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := val.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %#v", val)
	}
	if intVal.Val.Int64() != 11 {
		t.Fatalf("expected value field to be 11, got %d", intVal.Val.Int64())
	}
}
