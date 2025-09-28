package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestEvaluateStringLiteral(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{ast.Str("hello")}, nil, nil)
	val, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := val.(runtime.StringValue)
	if !ok || str.Val != "hello" {
		t.Fatalf("unexpected value %#v", val)
	}
}

func TestEvaluateIdentifierLookup(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()
	global.Define("greeting", runtime.StringValue{Val: "hello"})

	val, err := interp.evaluateExpression(ast.ID("greeting"), global)
	if err != nil {
		t.Fatalf("identifier lookup failed: %v", err)
	}
	str, ok := val.(runtime.StringValue)
	if !ok || str.Val != "hello" {
		t.Fatalf("unexpected value %#v", val)
	}
}

func TestEvaluateBlockCreatesScope(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()
	block := ast.Block(
		ast.Assign(ast.ID("x"), ast.Str("inner")),
		ast.ID("x"),
	)

	val, err := interp.evaluateExpression(block, global)
	if err != nil {
		t.Fatalf("block evaluation failed: %v", err)
	}
	str, ok := val.(runtime.StringValue)
	if !ok || str.Val != "inner" {
		t.Fatalf("unexpected block result %#v", val)
	}

	if _, err := global.Get("x"); err == nil {
		t.Fatalf("expected inner binding to stay scoped")
	}
}

func TestEvaluateBinaryAddition(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("a"), ast.Int(1)),
		ast.Assign(ast.ID("b"), ast.Int(2)),
		ast.Bin("+", ast.ID("a"), ast.ID("b")),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("module evaluation failed: %v", err)
	}
	iv, ok := result.(runtime.IntegerValue)
	if !ok || iv.Val.Cmp(bigInt(3)) != 0 {
		t.Fatalf("expected integer 3, got %#v", result)
	}
}

func TestDestructuringAssignmentArrayPattern(t *testing.T) {
	interp := New()
	patternWithRest := ast.ArrP([]ast.Pattern{ast.PatternFrom("first"), ast.PatternFrom("second")}, ast.PatternFrom("rest"))
	patternNoRest := ast.ArrP([]ast.Pattern{ast.PatternFrom("first"), ast.PatternFrom("second")}, nil)
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("arr"), ast.Arr(ast.Int(1), ast.Int(2), ast.Int(3))),
		ast.Assign(patternWithRest, ast.ID("arr")),
		ast.AssignOp(ast.AssignmentAssign, patternNoRest, ast.Arr(ast.Int(4), ast.Int(5))),
		ast.ID("rest"),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	first, err := env.Get("first")
	if err != nil {
		t.Fatalf("expected binding for first: %v", err)
	}
	firstInt, ok := first.(runtime.IntegerValue)
	if !ok || firstInt.Val.Cmp(bigInt(4)) != 0 {
		t.Fatalf("expected first == 4, got %#v", first)
	}
	second, err := env.Get("second")
	if err != nil {
		t.Fatalf("expected binding for second: %v", err)
	}
	secondInt, ok := second.(runtime.IntegerValue)
	if !ok || secondInt.Val.Cmp(bigInt(5)) != 0 {
		t.Fatalf("expected second == 5, got %#v", second)
	}
	if _, err := env.Get("rest"); err != nil {
		t.Fatalf("expected binding for rest: %v", err)
	}
	restVal, ok := result.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected rest array, got %#v", result)
	}
	if len(restVal.Elements) != 1 {
		t.Fatalf("expected rest length 1, got %d", len(restVal.Elements))
	}
	if restElem, ok := restVal.Elements[0].(runtime.IntegerValue); !ok || restElem.Val.Cmp(bigInt(3)) != 0 {
		t.Fatalf("expected rest element 3, got %#v", restVal.Elements[0])
	}
}

func TestForLoopArrayPattern(t *testing.T) {
	interp := New()
	pattern := ast.ArrP([]ast.Pattern{ast.PatternFrom("x"), ast.PatternFrom("y")}, nil)
	pairs := ast.Arr(
		ast.Arr(ast.Int(1), ast.Int(2)),
		ast.Arr(ast.Int(3), ast.Int(4)),
	)
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("pairs"), pairs),
		ast.Assign(ast.ID("sum"), ast.Int(0)),
		ast.ForLoopPattern(pattern, ast.ID("pairs"), ast.Block(
			ast.AssignOp(ast.AssignmentAssign, ast.ID("sum"), ast.Bin("+", ast.ID("sum"), ast.ID("x"))),
		)),
		ast.ID("sum"),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sum, ok := result.(runtime.IntegerValue)
	if !ok || sum.Val.Cmp(bigInt(4)) != 0 {
		t.Fatalf("expected sum 4, got %#v", result)
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

func TestInterfaceDynamicDispatch(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Display",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("string"),
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
		ast.Impl(
			"Display",
			ast.Ty("Point"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Point"))},
					[]ast.Statement{ast.Ret(ast.Str("point"))},
					ast.Ty("string"),
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
			ast.TypedP(ast.PatternFrom("value"), ast.Ty("Display")),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(2), "x"),
					ast.FieldInit(ast.Int(3), "y"),
				},
				false,
				"Point",
				nil,
				nil,
			),
		),
		ast.CallExpr(ast.Member(ast.ID("value"), "to_string")),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "point" {
		t.Fatalf("expected interface dispatch to return 'point', got %#v", result)
	}
}

func TestInterfaceAssignmentMissingImplementation(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Display",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("string"),
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
			ast.TypedP(ast.PatternFrom("value"), ast.Ty("Display")),
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
	}, nil, nil)

	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected failure when assigning struct without impl to interface")
	}
	if got := err.Error(); got != "Typed pattern mismatch in assignment" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestInterfaceUnionDispatchPrefersSpecific(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Show",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"describe",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("string"),
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
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("string"), "label"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"Basic",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("string"), "label"),
			},
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
					"describe",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Fancy"))},
					[]ast.Statement{ast.Ret(ast.Str("fancy"))},
					ast.Ty("string"),
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
			ast.Ty("Basic"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"describe",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Basic"))},
					[]ast.Statement{ast.Ret(ast.Str("basic"))},
					ast.Ty("string"),
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
					ast.Ty("string"),
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
			ast.ID("items"),
			ast.Arr(
				ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("f"), "label")}, false, "Fancy", nil, nil),
				ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("b"), "label")}, false, "Basic", nil, nil),
			),
		),
		ast.Assign(ast.ID("buffer"), ast.Str("")),
		ast.ForLoopPattern(
			ast.TypedP(ast.PatternFrom("item"), ast.Ty("Show")),
			ast.ID("items"),
			ast.Block(
				ast.AssignOp(
					ast.AssignmentAssign,
					ast.ID("buffer"),
					ast.Bin(
						"+",
						ast.ID("buffer"),
						ast.CallExpr(ast.Member(ast.ID("item"), "describe")),
					),
				),
			),
		),
		ast.ID("buffer"),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "fancybasic" {
		t.Fatalf("expected fancybasic, got %#v", result)
	}
}

func TestInterfaceDefaultMethodFallback(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Speakable",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"speak",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("string"),
					nil,
					nil,
					ast.Block(ast.Ret(ast.Str("default"))),
				),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"Bot",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("string"), "name"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Speakable",
			ast.Ty("Bot"),
			[]*ast.FunctionDefinition{},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.CallExpr(
			ast.Member(
				ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("Beep"), "name")}, false, "Bot", nil, nil),
				"speak",
			),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "default" {
		t.Fatalf("expected default, got %#v", result)
	}
}

func TestInterfaceDefaultMethodOverride(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Iface(
			"Speakable",
			[]*ast.FunctionSignature{
				ast.FnSig(
					"speak",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
					ast.Ty("string"),
					nil,
					nil,
					ast.Block(ast.Ret(ast.Str("default"))),
				),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.StructDef(
			"Bot",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("string"), "name"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"Speakable",
			ast.Ty("Bot"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"speak",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Bot"))},
					[]ast.Statement{ast.Ret(ast.Str("custom"))},
					ast.Ty("string"),
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
		ast.CallExpr(
			ast.Member(
				ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("Beep"), "name")}, false, "Bot", nil, nil),
				"speak",
			),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "custom" {
		t.Fatalf("expected custom, got %#v", result)
	}
}

func bigInt(v int64) *big.Int {
	return big.NewInt(v)
}
