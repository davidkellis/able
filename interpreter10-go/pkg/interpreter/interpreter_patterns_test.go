package interpreter

import (
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestMatchExpressionWithIdentifierAndLiteral(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Match(
			ast.Int(2),
			ast.Mc(ast.LitP(ast.Int(1)), ast.Int(10)),
			ast.Mc(ast.ID("x"), ast.Bin("+", ast.ID("x"), ast.Int(5))),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(7)) != 0 {
		t.Fatalf("expected integer 7, got %#v", result)
	}
}

func TestMatchExpressionStructGuard(t *testing.T) {
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
		ast.Match(
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
			ast.Mc(
				ast.StructP([]*ast.StructPatternField{
					ast.FieldP(ast.ID("a"), "x", nil),
					ast.FieldP(ast.ID("b"), "y", nil),
				}, false, "Point"),
				ast.Bin("+", ast.ID("a"), ast.ID("b")),
				ast.Bin(">", ast.ID("b"), ast.ID("a")),
			),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(3)) != 0 {
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

func TestForLoopContinueSkipsElements(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("sum"), ast.Int(0)),
		ast.ForLoopPattern(
			ast.ID("x"),
			ast.Arr(ast.Int(1), ast.Int(2), ast.Int(3)),
			ast.Block(
				ast.Iff(
					ast.Bin("==", ast.ID("x"), ast.Int(2)),
					ast.Block(ast.Cont(nil)),
				),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("sum"), ast.Bin("+", ast.ID("sum"), ast.ID("x"))),
			),
		),
		ast.ID("sum"),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(4)) != 0 {
		t.Fatalf("expected 4 from continue loop, got %#v", result)
	}
}

func TestBreakpointLabeledBreak(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("sum"), ast.Int(0)),
		ast.Breakpoint(
			"exit",
			ast.Block(
				ast.ForLoopPattern(
					ast.ID("n"),
					ast.Range(ast.Int(1), ast.Int(5), true),
					ast.Block(
						ast.AssignOp(ast.AssignmentAssign, ast.ID("sum"), ast.Bin("+", ast.ID("sum"), ast.ID("n"))),
						ast.Iff(
							ast.Bin("==", ast.ID("n"), ast.Int(3)),
							ast.Block(ast.Brk("exit", ast.Str("done"))),
						),
					),
				),
				ast.Str("fallthrough"),
			),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "done" {
		t.Fatalf("expected 'done', got %#v", result)
	}
}
