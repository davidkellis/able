package interpreter

import (
	"strings"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestFunctionDefinitionUnknownInterfaceConstraint(t *testing.T) {
	interp := New()
	fn := ast.Fn(
		"identity",
		[]*ast.FunctionParameter{ast.Param("value", nil)},
		[]ast.Statement{ast.ID("value")},
		nil,
		[]*ast.GenericParameter{
			ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Error"))),
		},
		nil,
		false,
		false,
	)
	module := ast.Mod([]ast.Statement{fn}, nil, nil)
	if _, _, err := interp.EvaluateModule(module); err == nil {
		t.Fatalf("expected unknown interface error")
	} else if err.Error() != "Unknown interface 'Error' in constraint on 'T'" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFunctionCallGenericConstraintSatisfied(t *testing.T) {
	interp := New()
	setupShowPoint(t, interp)
	module := ast.Mod(
		[]ast.Statement{
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
			ast.CallT(
				ast.Member(ast.ID("p"), "accept_show"),
				[]ast.TypeExpression{ast.Ty("Point")},
				ast.ID("p"),
			),
		},
		nil,
		nil,
	)
	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("call evaluation failed: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok {
		t.Fatalf("expected string result, got %T", result)
	}
	if str.Val != "Point(1, 2)" {
		t.Fatalf("expected Point(1, 2), got %s", str.Val)
	}
}

func TestFunctionCallGenericConstraintViolation(t *testing.T) {
	interp := New()
	setupShowPoint(t, interp)
	module := ast.Mod(
		[]ast.Statement{
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
			ast.CallT(
				ast.Member(ast.ID("p"), "accept_show"),
				[]ast.TypeExpression{ast.Ty("i32")},
				ast.Int(3),
			),
		},
		nil,
		nil,
	)
	if _, _, err := interp.EvaluateModule(module); err == nil {
		t.Fatalf("expected constraint violation")
	} else if err.Error() != "Type 'i32' does not satisfy interface 'Show': missing method 'to_string'" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUfcsOnPrimitiveValue(t *testing.T) {
	interp := New()
	module := ast.Mod(
		[]ast.Statement{
			ast.Fn(
				"add",
				[]*ast.FunctionParameter{
					ast.Param("a", nil),
					ast.Param("b", nil),
				},
				[]ast.Statement{
					ast.Ret(ast.Bin("+", ast.ID("a"), ast.ID("b"))),
				},
				nil,
				nil,
				nil,
				false,
				false,
			),
			ast.CallExpr(ast.Member(ast.Int(4), "add"), ast.Int(5)),
		},
		nil,
		nil,
	)
	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("module evaluation failed: %v", err)
	}
	iv, ok := result.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T", result)
	}
	if iv.Val.Cmp(bigInt(9)) != 0 {
		t.Fatalf("expected 9, got %v", iv.Val)
	}
}

func TestUfcsOnStructInstance(t *testing.T) {
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
				"move",
				[]*ast.FunctionParameter{
					ast.Param("p", nil),
					ast.Param("dx", nil),
				},
				[]ast.Statement{
					ast.AssignMember(ast.ID("p"), "x", ast.Bin("+", ast.Member(ast.ID("p"), "x"), ast.ID("dx"))),
					ast.Ret(ast.ID("p")),
				},
				nil,
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
			ast.CallExpr(ast.Member(ast.ID("p"), "move"), ast.Int(3)),
		},
		nil,
		nil,
	)
	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("module evaluation failed: %v", err)
	}
	if _, ok := result.(*runtime.StructInstanceValue); !ok {
		t.Fatalf("expected struct instance result, got %T", result)
	}
	val, err := env.Get("p")
	if err != nil {
		t.Fatalf("env lookup failed: %v", err)
	}
	inst, ok := val.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("expected struct instance for p, got %T", val)
	}
	field, ok := inst.Fields["x"]
	if !ok {
		t.Fatalf("field x missing on struct instance")
	}
	iv, ok := field.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer field, got %T", field)
	}
	if iv.Val.Cmp(bigInt(4)) != 0 {
		t.Fatalf("expected updated x=4, got %v", iv.Val)
	}
}

func TestNamedImplDisambiguation(t *testing.T) {
	interp := New()
	defs := []ast.Statement{
		ast.StructDef("Service", nil, ast.StructKindNamed, nil, nil, false),
		ast.Iface(
			"A",
			[]*ast.FunctionSignature{
				ast.FnSig("act", []*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))}, ast.Ty("string"), nil, nil, nil),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.Iface(
			"B",
			[]*ast.FunctionSignature{
				ast.FnSig("act", []*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))}, ast.Ty("string"), nil, nil, nil),
			},
			nil,
			nil,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"A",
			ast.Ty("Service"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"act",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Service"))},
					[]ast.Statement{ast.Ret(ast.Str("A"))},
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
			"B",
			ast.Ty("Service"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"act",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Service"))},
					[]ast.Statement{ast.Ret(ast.Str("B"))},
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
			"A",
			ast.Ty("Service"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"act",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Service"))},
					[]ast.Statement{ast.Ret(ast.Str("A.named"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			"ActA",
			nil,
			nil,
			nil,
			false,
		),
		ast.Impl(
			"B",
			ast.Ty("Service"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"act",
					[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Service"))},
					[]ast.Statement{ast.Ret(ast.Str("B.named"))},
					ast.Ty("string"),
					nil,
					nil,
					false,
					false,
				),
			},
			"ActB",
			nil,
			nil,
			nil,
			false,
		),
	}
	if _, _, err := interp.EvaluateModule(ast.Mod(defs, nil, nil)); err != nil {
		t.Fatalf("definitions failed: %v", err)
	}
	serviceLiteral := ast.StructLit(nil, false, "Service", nil, nil)
	ambiguous := ast.Mod([]ast.Statement{
		ast.CallExpr(ast.Member(serviceLiteral, "act")),
	}, nil, nil)
	if _, _, err := interp.EvaluateModule(ambiguous); err == nil {
		t.Fatalf("expected ambiguity error")
	} else if !strings.Contains(err.Error(), "Ambiguous method 'act' for type 'Service'") {
		t.Fatalf("unexpected error: %v", err)
	}
	callA := ast.Mod([]ast.Statement{
		ast.CallExpr(ast.Member(ast.ID("ActA"), "act"), ast.StructLit(nil, false, "Service", nil, nil)),
	}, nil, nil)
	valA, _, err := interp.EvaluateModule(callA)
	if err != nil {
		t.Fatalf("ActA call failed: %v", err)
	}
	strA, ok := valA.(runtime.StringValue)
	if !ok || strA.Val != "A.named" {
		t.Fatalf("expected A.named, got %#v", valA)
	}
	callB := ast.Mod([]ast.Statement{
		ast.CallExpr(ast.Member(ast.ID("ActB"), "act"), ast.StructLit(nil, false, "Service", nil, nil)),
	}, nil, nil)
	valB, _, err := interp.EvaluateModule(callB)
	if err != nil {
		t.Fatalf("ActB call failed: %v", err)
	}
	strB, ok := valB.(runtime.StringValue)
	if !ok || strB.Val != "B.named" {
		t.Fatalf("expected B.named, got %#v", valB)
	}
}

func TestUnnamedImplDuplicateRejected(t *testing.T) {
	interp := New()
	module := ast.Mod(
		[]ast.Statement{
			ast.Iface(
				"M",
				[]*ast.FunctionSignature{
					ast.FnSig("id", nil, ast.Ty("Self"), nil, nil, nil),
				},
				nil,
				nil,
				nil,
				nil,
				false,
			),
			ast.Impl(
				"M",
				ast.Ty("i32"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"id",
						nil,
						[]ast.Statement{
							ast.Ret(ast.Int(0)),
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
				nil,
				nil,
				false,
			),
			ast.Impl(
				"M",
				ast.Ty("i32"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"id",
						nil,
						[]ast.Statement{
							ast.Ret(ast.Int(1)),
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
				nil,
				nil,
				false,
			),
		},
		nil,
		nil,
	)
	if _, _, err := interp.EvaluateModule(module); err == nil {
		t.Fatalf("expected duplicate unnamed impl error")
	} else if err.Error() != "Unnamed impl for (M, i32) already exists" {
		t.Fatalf("unexpected error: %v", err)
	}
}
