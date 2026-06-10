package typechecker

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestFunctionCallAllowsIntegerWidening(t *testing.T) {
	checker := New()
	fn := ast.Fn(
		"takes_i64",
		[]*ast.FunctionParameter{
			ast.Param("value", ast.Ty("i64")),
		},
		[]ast.Statement{
			ast.Block(ast.Nil()),
		},
		ast.Ty("void"),
		nil,
		nil,
		false,
		false,
	)
	call := ast.Call("takes_i64", ast.Int(1))
	module := ast.NewModule([]ast.Statement{fn, call}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for widening call, got %v", diags)
	}
}

func TestFunctionReturnAllowsIntegerWidening(t *testing.T) {
	checker := New()
	fn := ast.Fn(
		"make_i64",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("value"), ast.Int(1)),
			ast.Ret(ast.ID("value")),
		},
		ast.Ty("i64"),
		nil,
		nil,
		false,
		false,
	)
	module := ast.NewModule([]ast.Statement{fn}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for widening return, got %v", diags)
	}
}

func TestExactIsizeRecurrenceSourceTypechecks(t *testing.T) {
	checker := New()
	fn := ast.Fn(
		"fib",
		[]*ast.FunctionParameter{
			ast.Param("n", ast.Ty("isize")),
		},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin("<=", ast.ID("n"), ast.Int(2)),
				ast.Block(ast.Ret(ast.Int(1))),
			),
			ast.Bin(
				"+",
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(1))),
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(2))),
			),
		},
		ast.Ty("isize"),
		nil,
		nil,
		false,
		false,
	)
	call := ast.Call("fib", ast.Int(30))
	module := ast.NewModule([]ast.Statement{fn, call}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for exact isize recurrence source, got %v", diags)
	}
}

func TestExactU64RecurrenceSourceTypechecks(t *testing.T) {
	checker := New()
	fn := ast.Fn(
		"fib",
		[]*ast.FunctionParameter{
			ast.Param("n", ast.Ty("u64")),
		},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin("<=", ast.ID("n"), ast.Int(2)),
				ast.Block(ast.Ret(ast.Int(1))),
			),
			ast.Bin(
				"+",
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(1))),
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(2))),
			),
		},
		ast.Ty("u64"),
		nil,
		nil,
		false,
		false,
	)
	call := ast.Call("fib", ast.Int(30))
	module := ast.NewModule([]ast.Statement{fn, call}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for exact u64 recurrence source, got %v", diags)
	}
}

func TestExactUsizeRecurrenceSourceTypechecks(t *testing.T) {
	checker := New()
	fn := ast.Fn(
		"fib",
		[]*ast.FunctionParameter{
			ast.Param("n", ast.Ty("usize")),
		},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin("<=", ast.ID("n"), ast.Int(2)),
				ast.Block(ast.Ret(ast.Int(1))),
			),
			ast.Bin(
				"+",
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(1))),
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(2))),
			),
		},
		ast.Ty("usize"),
		nil,
		nil,
		false,
		false,
	)
	call := ast.Call("fib", ast.Int(30))
	module := ast.NewModule([]ast.Statement{fn, call}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for exact usize recurrence source, got %v", diags)
	}
}

func TestAdditionalExactIntegerRecurrenceSourcesTypecheck(t *testing.T) {
	cases := []struct {
		name     string
		typeName string
		callArg  int64
	}{
		{name: "i8", typeName: "i8", callArg: 10},
		{name: "i16", typeName: "i16", callArg: 20},
		{name: "u8", typeName: "u8", callArg: 10},
		{name: "u16", typeName: "u16", callArg: 20},
		{name: "u32", typeName: "u32", callArg: 30},
		{name: "i128", typeName: "i128", callArg: 30},
		{name: "u128", typeName: "u128", callArg: 30},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checker := New()
			fn := ast.Fn(
				"fib",
				[]*ast.FunctionParameter{
					ast.Param("n", ast.Ty(tc.typeName)),
				},
				[]ast.Statement{
					ast.IfExpr(
						ast.Bin("<=", ast.ID("n"), ast.Int(2)),
						ast.Block(ast.Ret(ast.Int(1))),
					),
					ast.Bin(
						"+",
						ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(1))),
						ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(2))),
					),
				},
				ast.Ty(tc.typeName),
				nil,
				nil,
				false,
				false,
			)
			call := ast.Call("fib", ast.Int(tc.callArg))
			module := ast.NewModule([]ast.Statement{fn, call}, nil, nil)
			diags, err := checker.CheckModule(module)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(diags) != 0 {
				t.Fatalf("expected no diagnostics for exact %s recurrence source, got %v", tc.typeName, diags)
			}
		})
	}
}
