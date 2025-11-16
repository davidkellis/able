package interpreter

import (
	"math"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestForLoopSumArray(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("sum"), ast.Int(0)),
		ast.ForLoopPattern(
			ast.ID("x"),
			ast.Arr(ast.Int(1), ast.Int(2), ast.Int(3)),
			ast.Block(
				ast.AssignOp(
					ast.AssignmentAssign,
					ast.ID("sum"),
					ast.Bin("+", ast.ID("sum"), ast.ID("x")),
				),
			),
		),
		ast.ID("sum"),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(6)) != 0 {
		t.Fatalf("expected sum 6, got %#v", result)
	}
}

func TestWhileLoopIncrementsCounter(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("i"), ast.Int(0)),
		ast.While(
			ast.Bin("<", ast.ID("i"), ast.Int(3)),
			ast.Block(
				ast.AssignOp(
					ast.AssignmentAssign,
					ast.ID("i"),
					ast.Bin("+", ast.ID("i"), ast.Int(1)),
				),
			),
		),
		ast.ID("i"),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(3)) != 0 {
		t.Fatalf("expected i == 3, got %#v", result)
	}
}

func TestForLoopRangeCountdownInclusive(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("last"), ast.Int(0)),
		ast.ForLoopPattern(
			ast.ID("i"),
			ast.Range(ast.Int(3), ast.Int(1), true),
			ast.Block(
				ast.AssignOp(
					ast.AssignmentAssign,
					ast.ID("last"),
					ast.ID("i"),
				),
			),
		),
		ast.ID("last"),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(1)) != 0 {
		t.Fatalf("expected last == 1, got %#v", result)
	}
}

func TestForLoopRangeExclusiveUpperBound(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("count"), ast.Int(0)),
		ast.Assign(ast.ID("last"), ast.Int(0)),
		ast.ForLoopPattern(
			ast.ID("n"),
			ast.Range(ast.Int(1), ast.Int(4), false),
			ast.Block(
				ast.AssignOp(
					ast.AssignmentAssign,
					ast.ID("count"),
					ast.Bin("+", ast.ID("count"), ast.Int(1)),
				),
				ast.AssignOp(
					ast.AssignmentAssign,
					ast.ID("last"),
					ast.ID("n"),
				),
			),
		),
		ast.ID("last"),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(3)) != 0 {
		t.Fatalf("expected last == 3, got %#v", result)
	}
	countVal, err := env.Get("count")
	if err != nil {
		t.Fatalf("expected count binding: %v", err)
	}
	countInt, ok := countVal.(runtime.IntegerValue)
	if !ok || countInt.Val.Cmp(bigInt(3)) != 0 {
		t.Fatalf("expected count == 3, got %#v", countVal)
	}
}

func TestForLoopRangeEndpointMustBeFinite(t *testing.T) {
	cases := []struct {
		name  string
		start ast.Expression
		end   ast.Expression
	}{
		{"StartInfinite", ast.Flt(math.Inf(1)), ast.Int(5)},
		{"EndInfinite", ast.Int(1), ast.Flt(math.Inf(-1))},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			interp := New()
			module := ast.Mod([]ast.Statement{
				ast.ForLoopPattern(
					ast.ID("n"),
					ast.Range(tc.start, tc.end, true),
					ast.Block(),
				),
			}, nil, nil)

			if _, _, err := interp.EvaluateModule(module); err == nil {
				t.Fatalf("expected error for %s", tc.name)
			} else if err.Error() != "Range endpoint must be finite" {
				t.Fatalf("unexpected error for %s: %v", tc.name, err)
			}
		})
	}
}

func TestLoopExpressionReturnsBreakValue(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("counter"), ast.Int(0)),
		ast.Assign(
			ast.ID("result"),
			ast.Loop(
				ast.AssignOp(
					ast.AssignmentAssign,
					ast.ID("counter"),
					ast.Bin("+", ast.ID("counter"), ast.Int(1)),
				),
				ast.Iff(
					ast.Bin(">=", ast.ID("counter"), ast.Int(3)),
					ast.Brk(nil, ast.ID("counter")),
				),
			),
		),
		ast.ID("result"),
	}, nil, nil)

	value, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := value.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(3)) != 0 {
		t.Fatalf("expected loop result 3, got %#v", value)
	}
}

func TestIfOrSelectsFirstBranch(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.IfExpr(
			ast.Bool(true),
			ast.Block(ast.Int(1)),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(1)) != 0 {
		t.Fatalf("expected result 1, got %#v", result)
	}
}

func TestIfOrFallsThroughToElse(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.IfExpr(
			ast.Bool(false),
			ast.Block(ast.Int(1)),
			ast.OrC(ast.Block(ast.Int(2)), ast.Bool(false)),
			ast.OrC(ast.Block(ast.Int(3)), nil),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(3)) != 0 {
		t.Fatalf("expected else result 3, got %#v", result)
	}
}

func TestIfOrConditionClause(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.IfExpr(
			ast.Bool(false),
			ast.Block(ast.Int(1)),
			ast.OrC(ast.Block(ast.Int(2)), ast.Bool(true)),
			ast.OrC(ast.Block(ast.Int(3)), nil),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(2)) != 0 {
		t.Fatalf("expected or clause result 2, got %#v", result)
	}
}

func TestBreakpointExpressionReturnsLastValue(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("x"), ast.Int(1)),
		ast.Breakpoint(
			"dbg",
			ast.Block(
				ast.AssignOp(ast.AssignmentAssign, ast.ID("x"), ast.Int(2)),
				ast.Bin("+", ast.ID("x"), ast.Int(3)),
			),
		),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("breakpoint module failed: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(5)) != 0 {
		t.Fatalf("expected breakpoint to return 5, got %#v", result)
	}
	xVal, err := env.Get("x")
	if err != nil {
		t.Fatalf("expected binding for x: %v", err)
	}
	if intX, ok := xVal.(runtime.IntegerValue); !ok || intX.Val.Cmp(bigInt(2)) != 0 {
		t.Fatalf("expected x == 2 after breakpoint body, got %#v", xVal)
	}
}
