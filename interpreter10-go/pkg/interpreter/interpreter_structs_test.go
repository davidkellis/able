package interpreter

import (
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestPrivateStaticMethodNotAccessible(t *testing.T) {
	interp := New()
	defModule := ast.Mod([]ast.Statement{
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
		ast.Methods(
			ast.Ty("Point"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"hidden_static",
					nil,
					[]ast.Statement{
						ast.Ret(
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
					},
					nil,
					nil,
					nil,
					false,
					true,
				),
				ast.Fn(
					"origin",
					nil,
					[]ast.Statement{
						ast.Ret(
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
					},
					nil,
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
		),
	}, nil, nil)

	if _, _, err := interp.EvaluateModule(defModule); err != nil {
		t.Fatalf("setup module evaluation failed: %v", err)
	}

	callHidden := ast.Mod([]ast.Statement{
		ast.CallExpr(ast.Member(ast.ID("Point"), "hidden_static")),
	}, nil, nil)

	if _, _, err := interp.EvaluateModule(callHidden); err == nil {
		t.Fatalf("expected private static method call to fail")
	} else {
		expected := "Method 'hidden_static' on Point is private"
		if err.Error() != expected {
			t.Fatalf("expected error %q, got %q", expected, err.Error())
		}
	}

	callPublic := ast.Mod([]ast.Statement{
		ast.CallExpr(ast.Member(ast.ID("Point"), "origin")),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(callPublic)
	if err != nil {
		t.Fatalf("public static method call failed: %v", err)
	}
	if _, ok := result.(*runtime.StructInstanceValue); !ok {
		t.Fatalf("expected struct instance result, got %#v", result)
	}
}

func TestPrivateInstanceMethodNotAccessible(t *testing.T) {
	interp := New()
	defModule := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Counter",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "value"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Counter"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"hidden",
					[]*ast.FunctionParameter{ast.Param("self", nil)},
					[]ast.Statement{ast.Ret(ast.Int(1))},
					nil,
					nil,
					nil,
					false,
					true,
				),
				ast.Fn(
					"get",
					[]*ast.FunctionParameter{ast.Param("self", nil)},
					[]ast.Statement{
						ast.Ret(ast.Member(ast.ID("self"), "value")),
					},
					nil,
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
		),
	}, nil, nil)

	if _, _, err := interp.EvaluateModule(defModule); err != nil {
		t.Fatalf("setup module evaluation failed: %v", err)
	}

	instanceModule := ast.Mod([]ast.Statement{
		ast.Assign(
			ast.ID("counter"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(5), "value")},
				false,
				"Counter",
				nil,
				nil,
			),
		),
	}, nil, nil)

	if _, _, err := interp.EvaluateModule(instanceModule); err != nil {
		t.Fatalf("instance setup failed: %v", err)
	}

	callHidden := ast.Mod([]ast.Statement{
		ast.CallExpr(ast.Member(ast.ID("counter"), "hidden")),
	}, nil, nil)

	if _, _, err := interp.EvaluateModule(callHidden); err == nil {
		t.Fatalf("expected private instance method call to fail")
	} else {
		expected := "Method 'hidden' on Counter is private"
		if err.Error() != expected {
			t.Fatalf("expected error %q, got %q", expected, err.Error())
		}
	}

	callPublic := ast.Mod([]ast.Statement{
		ast.CallExpr(ast.Member(ast.ID("counter"), "get")),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(callPublic)
	if err != nil {
		t.Fatalf("public instance method call failed: %v", err)
	}
	intResult, ok := result.(runtime.IntegerValue)
	if !ok || intResult.Val.Cmp(bigInt(5)) != 0 {
		t.Fatalf("expected 5 from get(), got %#v", result)
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
				ast.ID("base"),
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
		ast.StructLit([]*ast.StructFieldInitializer{}, false, "A", ast.ID("b"), nil),
	}, nil, nil)

	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected functional update type mismatch error")
	}
	if got := err.Error(); got != "Functional update source must be same struct type" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStructStaticMethod(t *testing.T) {
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
		ast.Methods(
			ast.Ty("Point"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"origin",
					[]*ast.FunctionParameter{},
					[]ast.Statement{
						ast.Ret(ast.StructLit(
							[]*ast.StructFieldInitializer{
								ast.FieldInit(ast.Int(0), "x"),
								ast.FieldInit(ast.Int(0), "y"),
							},
							false,
							"Point",
							nil,
							nil,
						)),
					},
					nil,
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
		),
		ast.CallExpr(ast.Member(ast.ID("Point"), "origin")),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	inst, ok := result.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("expected struct instance, got %#v", result)
	}
	if inst.Fields == nil {
		t.Fatalf("expected named struct instance fields")
	}
	if x, ok := inst.Fields["x"].(runtime.IntegerValue); !ok || x.Val.Cmp(bigInt(0)) != 0 {
		t.Fatalf("expected x == 0, got %#v", inst.Fields["x"])
	}
	if y, ok := inst.Fields["y"].(runtime.IntegerValue); !ok || y.Val.Cmp(bigInt(0)) != 0 {
		t.Fatalf("expected y == 0, got %#v", inst.Fields["y"])
	}
}

func TestStructStaticMethodPrivateError(t *testing.T) {
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
		ast.Methods(
			ast.Ty("Point"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"hidden_static",
					[]*ast.FunctionParameter{},
					[]ast.Statement{
						ast.Ret(ast.StructLit(
							[]*ast.StructFieldInitializer{
								ast.FieldInit(ast.Int(0), "x"),
								ast.FieldInit(ast.Int(0), "y"),
							},
							false,
							"Point",
							nil,
							nil,
						)),
					},
					nil,
					nil,
					nil,
					false,
					true,
				),
			},
			nil,
			nil,
		),
		ast.CallExpr(ast.Member(ast.ID("Point"), "hidden_static")),
	}, nil, nil)

	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected private static method error")
	}
	if got := err.Error(); got != "Method 'hidden_static' on Point is private" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStructInstanceMethodCall(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Counter",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "value"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Counter"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"get",
					[]*ast.FunctionParameter{ast.Param("self", nil)},
					[]ast.Statement{
						ast.Ret(ast.Member(ast.ID("self"), "value")),
					},
					nil,
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
		),
		ast.Assign(
			ast.ID("c"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(5), "value"),
				},
				false,
				"Counter",
				nil,
				nil,
			),
		),
		ast.CallExpr(ast.Member(ast.ID("c"), "get")),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok := result.(runtime.IntegerValue)
	if !ok || val.Val.Cmp(bigInt(5)) != 0 {
		t.Fatalf("expected result 5, got %#v", result)
	}
}

func TestStructInstanceMethodPrivateError(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Counter",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "value"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Counter"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"hidden",
					[]*ast.FunctionParameter{ast.Param("self", nil)},
					[]ast.Statement{
						ast.Ret(ast.Int(1)),
					},
					nil,
					nil,
					nil,
					false,
					true,
				),
			},
			nil,
			nil,
		),
		ast.Assign(
			ast.ID("c"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(2), "value"),
				},
				false,
				"Counter",
				nil,
				nil,
			),
		),
		ast.CallExpr(ast.Member(ast.ID("c"), "hidden")),
	}, nil, nil)

	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected private instance method error")
	}
	if got := err.Error(); got != "Method 'hidden' on Counter is private" {
		t.Fatalf("unexpected error: %v", err)
	}
}
