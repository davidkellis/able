package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestGenericStructLiteralDefersAndMemoizesTypeArgumentsAcrossModes(t *testing.T) {
	boxDef := ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("T"), "value"),
		},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	module := ast.Mod([]ast.Statement{
		boxDef,
		ast.Assign(
			ast.ID("box"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(7), "value"),
				},
				false,
				"Box",
				nil,
				nil,
			),
		),
		ast.ID("box"),
	}, nil, nil)

	for _, tc := range []struct {
		name   string
		create func() *Interpreter
	}{
		{name: "treewalker", create: New},
		{name: "bytecode", create: NewBytecode},
	} {
		t.Run(tc.name, func(t *testing.T) {
			interp := tc.create()
			result, _, err := interp.EvaluateModule(module)
			if err != nil {
				t.Fatalf("EvaluateModule failed: %v", err)
			}
			inst, ok := result.(*runtime.StructInstanceValue)
			if !ok || inst == nil {
				t.Fatalf("expected struct instance, got %#v", result)
			}
			if len(inst.TypeArguments) != 0 {
				t.Fatalf("expected unconstrained generic literal to defer type args, got %#v", inst.TypeArguments)
			}
			typeExpr := interp.typeExpressionForValue(inst)
			if got := typeExpressionToString(typeExpr); got != "Box<i32>" {
				t.Fatalf("typeExpressionForValue = %q, want %q", got, "Box<i32>")
			}
			if len(inst.TypeArguments) != 1 || typeExpressionToString(inst.TypeArguments[0]) != "i32" {
				t.Fatalf("expected typeExpressionForValue to memoize inferred args, got %#v", inst.TypeArguments)
			}
			typeArg := inst.TypeArguments[0]
			allocs := testing.AllocsPerRun(1000, func() {
				info, ok := interp.typeInfoFromStructInstance(inst)
				if !ok || len(info.typeArgs) != 1 || info.typeArgs[0] != typeArg {
					panic("unexpected struct type info")
				}
			})
			if allocs != 0 {
				t.Fatalf("expected seeded generic struct type-info hot path to allocate zero, got %.2f", allocs)
			}
		})
	}
}

func TestGenericStructLiteralInfersThroughNullableFieldAcrossModes(t *testing.T) {
	boxDef := ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("T"), "value"),
		},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	holderDef := ast.StructDef(
		"Holder",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Nullable(ast.Gen(ast.Ty("Box"), ast.Ty("T"))), "maybe"),
		},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	module := ast.Mod([]ast.Statement{
		boxDef,
		holderDef,
		ast.Assign(
			ast.ID("box"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(7), "value"),
				},
				false,
				"Box",
				nil,
				nil,
			),
		),
		ast.Assign(
			ast.ID("holder"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.ID("box"), "maybe"),
				},
				false,
				"Holder",
				nil,
				nil,
			),
		),
		ast.ID("holder"),
	}, nil, nil)

	for _, tc := range []struct {
		name   string
		create func() *Interpreter
	}{
		{name: "treewalker", create: New},
		{name: "bytecode", create: NewBytecode},
	} {
		t.Run(tc.name, func(t *testing.T) {
			interp := tc.create()
			result, _, err := interp.EvaluateModule(module)
			if err != nil {
				t.Fatalf("EvaluateModule failed: %v", err)
			}
			inst, ok := result.(*runtime.StructInstanceValue)
			if !ok || inst == nil {
				t.Fatalf("expected struct instance, got %#v", result)
			}
			if len(inst.TypeArguments) != 0 {
				t.Fatalf("expected unconstrained generic literal to defer type args, got %#v", inst.TypeArguments)
			}
			if got := typeExpressionToString(interp.typeExpressionForValue(inst)); got != "Holder<i32>" {
				t.Fatalf("typeExpressionForValue = %q, want %q", got, "Holder<i32>")
			}
			if len(inst.TypeArguments) != 1 || typeExpressionToString(inst.TypeArguments[0]) != "i32" {
				t.Fatalf("expected nullable field inference to memoize Holder<i32>, got %#v", inst.TypeArguments)
			}
			maybe, ok := structNamedFieldValue(inst, "maybe")
			if !ok {
				t.Fatalf("expected maybe field")
			}
			box, ok := maybe.(*runtime.StructInstanceValue)
			if !ok || box == nil {
				t.Fatalf("expected maybe field to hold Box instance, got %#v", maybe)
			}
			if len(box.TypeArguments) != 0 {
				t.Fatalf("expected nested evidence value to stay lazily typed, got %#v", box.TypeArguments)
			}
		})
	}
}

func TestGenericStructFunctionalUpdateUsesDeferredSourceTypeArgumentsAcrossModes(t *testing.T) {
	boxDef := ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("T"), "value"),
		},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	module := ast.Mod([]ast.Statement{
		boxDef,
		ast.Assign(
			ast.ID("box"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(1), "value"),
				},
				false,
				"Box",
				nil,
				nil,
			),
		),
		ast.Assign(
			ast.ID("box2"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(2), "value"),
				},
				false,
				"Box",
				[]ast.Expression{ast.ID("box")},
				nil,
			),
		),
		ast.ID("box2"),
	}, nil, nil)

	for _, tc := range []struct {
		name   string
		create func() *Interpreter
	}{
		{name: "treewalker", create: New},
		{name: "bytecode", create: NewBytecode},
	} {
		t.Run(tc.name, func(t *testing.T) {
			interp := tc.create()
			result, _, err := interp.EvaluateModule(module)
			if err != nil {
				t.Fatalf("EvaluateModule failed: %v", err)
			}
			inst, ok := result.(*runtime.StructInstanceValue)
			if !ok || inst == nil {
				t.Fatalf("expected struct instance, got %#v", result)
			}
			if got := typeExpressionToString(interp.typeExpressionForValue(inst)); got != "Box<i32>" {
				t.Fatalf("updated box type = %q, want %q", got, "Box<i32>")
			}
			raw, ok := structNamedFieldValue(inst, "value")
			if !ok {
				t.Fatalf("expected updated value field")
			}
			intVal, ok := raw.(runtime.IntegerValue)
			if !ok || intVal.BigInt().Cmp(bigInt(2)) != 0 {
				t.Fatalf("updated value field = %#v, want integer 2", raw)
			}
		})
	}
}
