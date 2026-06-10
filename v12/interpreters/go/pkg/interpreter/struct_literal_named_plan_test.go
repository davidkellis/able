package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestNamedStructLiteralFieldOrderRejectsDuplicateField(t *testing.T) {
	def := ast.StructDef(
		"Point",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "x"),
			ast.FieldDef(ast.Ty("i32"), "y"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	lit := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Int(3), "x"),
			ast.FieldInit(ast.Int(5), "x"),
		},
		false,
		"Point",
		nil,
		nil,
	)

	_, err := namedStructLiteralFieldOrder(lit, def)
	if err == nil {
		t.Fatalf("expected duplicate field error")
	}
	if got := err.Error(); got != "Duplicate field 'x' for struct 'Point'" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNamedStructLiteralPlanCachedReusesPlanForSameDefinition(t *testing.T) {
	interp := New()
	def := ast.StructDef(
		"Point",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "x"),
			ast.FieldDef(ast.Ty("i32"), "y"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	lit := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Int(3), "x"),
			ast.FieldInit(ast.Int(5), "y"),
		},
		false,
		"Point",
		nil,
		nil,
	)

	first, err := interp.namedStructLiteralPlanCached(lit, def)
	if err != nil {
		t.Fatalf("first plan build: %v", err)
	}
	second, err := interp.namedStructLiteralPlanCached(lit, def)
	if err != nil {
		t.Fatalf("second plan build: %v", err)
	}
	if len(first.fieldOrder) != 2 || len(second.fieldOrder) != 2 {
		t.Fatalf("unexpected field orders: %#v %#v", first.fieldOrder, second.fieldOrder)
	}
	if &first.fieldOrder[0] != &second.fieldOrder[0] {
		t.Fatalf("expected cached plan field order backing to be reused")
	}
}

func TestBytecodeNamedStructLiteralDuplicateFieldError(t *testing.T) {
	interp := NewBytecode()
	module := ast.Mod([]ast.Statement{
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
		ast.Fn(
			"make_point",
			nil,
			[]ast.Statement{
				ast.Ret(ast.StructLit(
					[]*ast.StructFieldInitializer{
						ast.FieldInit(ast.Int(3), "x"),
						ast.FieldInit(ast.Int(5), "x"),
					},
					false,
					"Point",
					nil,
					nil,
				)),
			},
			ast.Ty("Point"),
			nil,
			nil,
			false,
			false,
		),
	}, nil, nil)

	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected duplicate field error")
	}
	if got := err.Error(); got != "Duplicate field 'x' for struct 'Point'" {
		t.Fatalf("unexpected error: %v", err)
	}
}
