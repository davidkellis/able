package interpreter

import (
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestImplicitMemberInFreeFunction(t *testing.T) {
	interp := New()
	structDef := ast.StructDef(
		"Counter",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "value"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	valueOfFn := ast.Fn(
		"value_of",
		[]*ast.FunctionParameter{
			ast.Param("counter", ast.Ty("Counter")),
		},
		[]ast.Statement{
			ast.Ret(ast.ImplicitMember("value")),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	module := ast.Mod([]ast.Statement{
		structDef,
		valueOfFn,
		ast.Assign(
			ast.ID("counter"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(42), "value"),
				},
				false,
				"Counter",
				nil,
				nil,
			),
		),
		ast.CallExpr(ast.ID("value_of"), ast.ID("counter")),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("free function implicit member failed: %v", err)
	}
	intResult, ok := result.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %#v", result)
	}
	if intResult.Val.Cmp(bigInt(42)) != 0 {
		t.Fatalf("expected 42 from value_of, got %#v", intResult.Val)
	}
}

func TestStructLiteralNamed(t *testing.T) {
	interp := New()
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
		ast.Assign(
			ast.ID("p"),
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
		ast.Member(ast.ID("p"), "x"),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok := result.(runtime.IntegerValue)
	if !ok || val.Val.Cmp(bigInt(3)) != 0 {
		t.Fatalf("expected struct field x == 3, got %#v", result)
	}
	structVal, err := env.Get("p")
	if err != nil {
		t.Fatalf("expected binding for p: %v", err)
	}
	instance, ok := structVal.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("expected struct instance, got %#v", structVal)
	}
	if instance.Fields == nil {
		t.Fatalf("expected named struct fields map")
	}
	if field, ok := instance.Fields["y"].(runtime.IntegerValue); !ok || field.Val.Cmp(bigInt(4)) != 0 {
		t.Fatalf("expected struct field y == 4, got %#v", instance.Fields["y"])
	}
}

func TestStructLiteralPositional(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Pair",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), nil),
				ast.FieldDef(ast.Ty("i32"), nil),
			},
			ast.StructKindPositional,
			nil,
			nil,
			false,
		),
		ast.Assign(
			ast.ID("pair"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(7), nil),
					ast.FieldInit(ast.Int(9), nil),
				},
				true,
				"Pair",
				nil,
				nil,
			),
		),
		ast.Member(ast.ID("pair"), ast.Int(1)),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok := result.(runtime.IntegerValue)
	if !ok || val.Val.Cmp(bigInt(9)) != 0 {
		t.Fatalf("expected positional field 1 == 9, got %#v", result)
	}
}

func TestStructMemberAssignmentMutation(t *testing.T) {
	interp := New()
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
		ast.Assign(
			ast.ID("p"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(0), "x"),
					ast.FieldInit(ast.Int(0), "y"),
				},
				false,
				"Point",
				nil,
				nil,
			),
		),
		ast.AssignMember(ast.ID("p"), "x", ast.Int(5)),
		ast.Member(ast.ID("p"), "x"),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok := result.(runtime.IntegerValue)
	if !ok || val.Val.Cmp(bigInt(5)) != 0 {
		t.Fatalf("expected updated field x == 5, got %#v", result)
	}
	structVal, err := env.Get("p")
	if err != nil {
		t.Fatalf("expected struct binding for p: %v", err)
	}
	inst, ok := structVal.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("expected struct instance, got %#v", structVal)
	}
	if field, ok := inst.Fields["y"].(runtime.IntegerValue); !ok || field.Val.Cmp(bigInt(0)) != 0 {
		t.Fatalf("unexpected change to y field: %#v", inst.Fields["y"])
	}
}

func TestArrayIndexRead(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("a"), ast.Arr(ast.Int(10), ast.Int(20), ast.Int(30))),
		ast.Index(ast.ID("a"), ast.Int(1)),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(20)) != 0 {
		t.Fatalf("expected index read 20, got %#v", result)
	}
}

func TestArrayIndexAssignment(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("a"), ast.Arr(ast.Int(1), ast.Int(2))),
		ast.AssignOp(ast.AssignmentAssign, ast.Index(ast.ID("a"), ast.Int(1)), ast.Int(9)),
		ast.Index(ast.ID("a"), ast.Int(1)),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(9)) != 0 {
		t.Fatalf("expected updated index value 9, got %#v", result)
	}
	arrVal, err := env.Get("a")
	if err != nil {
		t.Fatalf("expected array binding for 'a': %v", err)
	}
	arr, ok := arrVal.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected array value, got %#v", arrVal)
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("expected array length 2, got %d", len(arr.Elements))
	}
	if elem, ok := arr.Elements[1].(runtime.IntegerValue); !ok || elem.Val.Cmp(bigInt(9)) != 0 {
		t.Fatalf("expected element 1 == 9, got %#v", arr.Elements[1])
	}
}

func TestCompoundAssignments(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("x"), ast.Int(2)),
		ast.AssignOp(ast.AssignmentAdd, ast.ID("x"), ast.Int(3)),
		ast.AssignOp(ast.AssignmentShiftL, ast.ID("x"), ast.Int(1)),
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
		ast.Assign(
			ast.ID("p"),
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
		),
		ast.AssignOp(ast.AssignmentAdd, ast.Member(ast.ID("p"), "x"), ast.Int(4)),
		ast.Assign(ast.ID("arr"), ast.Arr(ast.Int(3), ast.Int(4))),
		ast.AssignOp(ast.AssignmentMul, ast.Index(ast.ID("arr"), ast.Int(1)), ast.Int(2)),
		ast.ID("x"),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intVal, ok := result.(runtime.IntegerValue); !ok || intVal.Val.Cmp(bigInt(10)) != 0 {
		t.Fatalf("expected x == 10, got %#v", result)
	}
	pVal, err := env.Get("p")
	if err != nil {
		t.Fatalf("expected struct binding for 'p': %v", err)
	}
	inst, ok := pVal.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("expected struct instance, got %#v", pVal)
	}
	if field, ok := inst.Fields["x"].(runtime.IntegerValue); !ok || field.Val.Cmp(bigInt(5)) != 0 {
		t.Fatalf("expected struct field x == 5, got %#v", inst.Fields["x"])
	}
	arrVal, err := env.Get("arr")
	if err != nil {
		t.Fatalf("expected array binding for 'arr': %v", err)
	}
	arr, ok := arrVal.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected array value, got %#v", arrVal)
	}
	if elem, ok := arr.Elements[1].(runtime.IntegerValue); !ok || elem.Val.Cmp(bigInt(8)) != 0 {
		t.Fatalf("expected array element 1 == 8, got %#v", arr.Elements[1])
	}
}

func TestStructLiteralMissingFieldError(t *testing.T) {
	interp := New()
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
		ast.StructLit(
			[]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Int(3), "x"),
			},
			false,
			"Point",
			nil,
			nil,
		),
	}, nil, nil)

	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected missing field error")
	}
	if got := err.Error(); got != "Missing field 'y' for struct 'Point'" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStructLiteralPositionalArityError(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Pair",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), nil),
				ast.FieldDef(ast.Ty("i32"), nil),
			},
			ast.StructKindPositional,
			nil,
			nil,
			false,
		),
		ast.StructLit(
			[]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Int(7), nil),
			},
			true,
			"Pair",
			nil,
			nil,
		),
	}, nil, nil)

	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected arity error")
	}
	if got := err.Error(); got != "Struct 'Pair' expects 2 fields, got 1" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStructFunctionalUpdate(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"User",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "id"),
				ast.FieldDef(ast.Ty("string"), "name"),
				ast.FieldDef(ast.Ty("bool"), "active"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Assign(
			ast.ID("base"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(1), "id"),
					ast.FieldInit(ast.Str("Alice"), "name"),
					ast.FieldInit(ast.Bool(true), "active"),
				},
				false,
				"User",
				nil,
				nil,
			),
		),
		ast.Assign(
			ast.ID("updated"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Str("Bob"), "name"),
				},
				false,
				"User",
				[]ast.Expression{ast.ID("base")},
				nil,
			),
		),
		ast.Member(ast.ID("updated"), "name"),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	nameVal, ok := result.(runtime.StringValue)
	if !ok || nameVal.Val != "Bob" {
		t.Fatalf("expected updated.name == Bob, got %#v", result)
	}
	baseVal, err := env.Get("base")
	if err != nil {
		t.Fatalf("expected binding for base: %v", err)
	}
	baseStruct, ok := baseVal.(*runtime.StructInstanceValue)
	if !ok || baseStruct.Fields == nil {
		t.Fatalf("expected named struct base, got %#v", baseVal)
	}
	if field, ok := baseStruct.Fields["name"].(runtime.StringValue); !ok || field.Val != "Alice" {
		t.Fatalf("base struct mutated unexpectedly: %#v", baseStruct.Fields["name"])
	}
}

func TestStructFunctionalUpdateMultipleSources(t *testing.T) {
	interp := New()
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
		ast.Assign(
			ast.ID("merged"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(99), "x"),
				},
				false,
				"Point",
				[]ast.Expression{
					ast.StructLit(
						[]*ast.StructFieldInitializer{
							ast.FieldInit(ast.Int(1), "x"),
							ast.FieldInit(ast.Int(10), "y"),
						},
						false,
						"Point",
						nil,
						nil,
					),
					ast.StructLit(
						[]*ast.StructFieldInitializer{
							ast.FieldInit(ast.Int(2), "x"),
							ast.FieldInit(ast.Int(20), "y"),
						},
						false,
						"Point",
						nil,
						nil,
					),
				},
				nil,
			),
		),
		ast.Member(ast.ID("merged"), "y"),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	yVal, ok := result.(runtime.IntegerValue)
	if !ok || yVal.Val.Int64() != 20 {
		t.Fatalf("expected merged.y == 20, got %#v", result)
	}
	mergedVal, err := env.Get("merged")
	if err != nil {
		t.Fatalf("expected merged binding: %v", err)
	}
	mergedStruct, ok := mergedVal.(*runtime.StructInstanceValue)
	if !ok || mergedStruct.Fields == nil {
		t.Fatalf("expected named struct merged, got %#v", mergedVal)
	}
	if field, ok := mergedStruct.Fields["x"].(runtime.IntegerValue); !ok || field.Val.Int64() != 99 {
		t.Fatalf("merged.x incorrect, got %#v", mergedStruct.Fields["x"])
	}
}

func TestStructFunctionalUpdateWrongType(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"A",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "x"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"B",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "y"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Assign(
			ast.ID("a"),
			ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(10), "x")}, false, "A", nil, nil),
		),
		ast.Assign(
			ast.ID("b"),
			ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(20), "y")}, false, "B", nil, nil),
		),
		ast.StructLit([]*ast.StructFieldInitializer{}, false, "A", []ast.Expression{ast.ID("b")}, nil),
	}, nil, nil)

	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected functional update type mismatch error")
	}
	if got := err.Error(); got != "Functional update source must be same struct type" {
		t.Fatalf("unexpected error: %v", err)
	}
}
