package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
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

func TestMethodsExportedAsFunctions(t *testing.T) {
	interp := New()
	defModule := ast.Mod([]ast.Statement{
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
					"norm",
					nil,
					[]ast.Statement{ast.Ret(ast.Int(1))},
					ast.Ty("i32"),
					nil,
					nil,
					true,
					false,
				),
				ast.Fn(
					"origin",
					nil,
					[]ast.Statement{
						ast.Ret(
							ast.StructLit(
								[]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(0), "x")},
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

	mustEvalModule(t, interp, defModule)
	mustEvalModule(t, interp, ast.Mod([]ast.Statement{
		ast.Assign(
			ast.ID("p"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(3), "x")},
				false,
				"Point",
				nil,
				nil,
			),
		),
	}, nil, nil))

	methodResult := mustEvalModule(t, interp, ast.Mod([]ast.Statement{
		ast.CallExpr(ast.Member(ast.ID("p"), "norm")),
	}, nil, nil))
	if v, ok := methodResult.(runtime.IntegerValue); !ok || v.Val.Cmp(bigInt(1)) != 0 {
		t.Fatalf("expected norm() via method to return 1, got %#v", methodResult)
	}

	fnResult := mustEvalModule(t, interp, ast.Mod([]ast.Statement{
		ast.CallExpr(ast.ID("norm"), ast.ID("p")),
	}, nil, nil))
	if v, ok := fnResult.(runtime.IntegerValue); !ok || v.Val.Cmp(bigInt(1)) != 0 {
		t.Fatalf("expected norm(Point) free call to return 1, got %#v", fnResult)
	}

	if _, _, err := interp.EvaluateModule(ast.Mod([]ast.Statement{
		ast.CallExpr(ast.ID("origin")),
	}, nil, nil)); err == nil {
		t.Fatalf("expected unqualified type-qualified function call to fail")
	} else if err.Error() != "Undefined variable 'origin'" {
		t.Fatalf("unexpected error: %v", err)
	}

	origin := mustEvalModule(t, interp, ast.Mod([]ast.Statement{
		ast.CallExpr(ast.ID("Point.origin")),
	}, nil, nil))
	if origin.Kind() != runtime.KindStructInstance {
		t.Fatalf("expected origin() to return struct instance, got %#v", origin)
	}
}

func TestMethodShorthandImplicitMember(t *testing.T) {
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
	incrementBody := []ast.Statement{
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ImplicitMember("value"),
			ast.Bin(
				"+",
				ast.ImplicitMember("value"),
				ast.Int(1),
			),
		),
	}
	addBody := []ast.Statement{
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ImplicitMember("value"),
			ast.Bin(
				"+",
				ast.ImplicitMember("value"),
				ast.ID("amount"),
			),
		),
	}
	methods := ast.Methods(
		ast.Ty("Counter"),
		[]*ast.FunctionDefinition{
			ast.Fn(
				"increment",
				nil,
				incrementBody,
				nil,
				nil,
				nil,
				true,
				false,
			),
			ast.Fn(
				"add",
				[]*ast.FunctionParameter{
					ast.Param("amount", nil),
				},
				addBody,
				nil,
				nil,
				nil,
				true,
				false,
			),
		},
		nil,
		nil,
	)
	setupModule := ast.Mod([]ast.Statement{
		structDef,
		methods,
		ast.Assign(
			ast.ID("counter"),
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
		ast.CallExpr(ast.Member(ast.ID("counter"), "increment")),
		ast.CallExpr(ast.Member(ast.ID("counter"), "add"), ast.Int(3)),
		ast.Member(ast.ID("counter"), "value"),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(setupModule)
	if err != nil {
		t.Fatalf("method shorthand evaluation failed: %v", err)
	}
	intResult, ok := result.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %#v", result)
	}
	if intResult.Val.Cmp(bigInt(9)) != 0 {
		t.Fatalf("expected counter.value == 9, got %#v", intResult.Val)
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
