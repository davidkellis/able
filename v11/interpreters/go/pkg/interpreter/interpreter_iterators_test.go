package interpreter

import (
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestIteratorLiteralIsLazy(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()

	mustEval := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, env)
		if err != nil {
			t.Fatalf("evaluation failed: %v", err)
		}
		return val
	}

	mustEval(ast.Assign(ast.ID("count"), ast.Int(0)))
	mustEval(ast.Assign(ast.ID("iter"), ast.IteratorLit(
		ast.AssignOp(ast.AssignmentAssign, ast.ID("count"), ast.Bin("+", ast.ID("count"), ast.Int(1))),
		ast.Yield(ast.ID("count")),
		ast.AssignOp(ast.AssignmentAssign, ast.ID("count"), ast.Bin("+", ast.ID("count"), ast.Int(1))),
		ast.Yield(ast.ID("count")),
	)))

	if got := mustGetInt(t, env, "count"); got != 0 {
		t.Fatalf("expected count to remain 0 before iteration, got %d", got)
	}

	first := mustEval(ast.CallExpr(ast.Member(ast.ID("iter"), ast.ID("next"))))
	firstInt, ok := first.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected first next() to yield integer, got %#v", first)
	}
	if firstInt.Val.Int64() != 1 {
		t.Fatalf("expected first next() = 1, got %s", firstInt.Val.String())
	}
	if got := mustGetInt(t, env, "count"); got != 1 {
		t.Fatalf("expected count to be 1 after first yield, got %d", got)
	}

	second := mustEval(ast.CallExpr(ast.Member(ast.ID("iter"), ast.ID("next"))))
	secondInt, ok := second.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected second next() to yield integer, got %#v", second)
	}
	if secondInt.Val.Int64() != 2 {
		t.Fatalf("expected second next() = 2, got %s", secondInt.Val.String())
	}
	if got := mustGetInt(t, env, "count"); got != 2 {
		t.Fatalf("expected count to be 2 after second yield, got %d", got)
	}

	third := mustEval(ast.CallExpr(ast.Member(ast.ID("iter"), ast.ID("next"))))
	if _, ok := third.(runtime.IteratorEndValue); !ok {
		t.Fatalf("expected third next() to return IteratorEnd, got %#v", third)
	}
	if got := mustGetInt(t, env, "count"); got != 2 {
		t.Fatalf("expected count to remain 2 after exhaustion, got %d", got)
	}
}

func TestIteratorLiteralPropagatesRaise(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()

	if _, err := interp.evaluateExpression(ast.Assign(ast.ID("iter"), ast.IteratorLit(
		ast.Raise(ast.Str("boom")),
	)), env); err != nil {
		t.Fatalf("iterator literal construction failed: %v", err)
	}

	_, err := interp.evaluateExpression(ast.CallExpr(ast.Member(ast.ID("iter"), ast.ID("next"))), env)
	if err == nil {
		t.Fatalf("expected next() to raise error")
	}
	if _, ok := err.(raiseSignal); !ok {
		t.Fatalf("expected raiseSignal, got %T: %v", err, err)
	}
}

func TestForLoopConsumesIterableIterator(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()

	mustEval := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, env)
		if err != nil {
			t.Fatalf("evaluation failed: %v", err)
		}
		return val
	}

	mustEval(ast.Assign(ast.ID("iter"), ast.IteratorLit(
		ast.Yield(ast.Int(1)),
		ast.Yield(ast.Int(2)),
	)))
	mustEval(ast.Assign(ast.ID("sum"), ast.Int(0)))

	forLoop := ast.ForIn("item", ast.ID("iter"),
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("sum"),
			ast.Bin("+", ast.ID("sum"), ast.ID("item")),
		),
	)
	if _, err := interp.evaluateStatement(forLoop, env); err != nil {
		t.Fatalf("for loop evaluation failed: %v", err)
	}

	if got := mustGetInt(t, env, "sum"); got != 3 {
		t.Fatalf("expected sum = 3, got %d", got)
	}
}

func TestIteratorForLoopBodyWithinGenerator(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()

	if _, err := interp.evaluateExpression(ast.Assign(ast.ID("iter"), ast.IteratorLit(
		ast.ForIn(ast.ID("item"), ast.Arr(ast.Int(1), ast.Int(2), ast.Int(3)),
			ast.Yield(ast.ID("item")),
		),
	)), env); err != nil {
		t.Fatalf("iterator literal construction failed: %v", err)
	}

	nextCall := ast.CallExpr(ast.Member(ast.ID("iter"), ast.ID("next")))

	expectValue := func(expected int64) {
		val, err := interp.evaluateExpression(nextCall, env)
		if err != nil {
			t.Fatalf("next() failed: %v", err)
		}
		intVal, ok := val.(runtime.IntegerValue)
		if !ok || intVal.Val.Int64() != expected {
			t.Fatalf("expected %d, got %#v", expected, val)
		}
	}

	expectValue(1)
	expectValue(2)
	expectValue(3)
	endVal, err := interp.evaluateExpression(nextCall, env)
	if err != nil {
		t.Fatalf("next() failed: %v", err)
	}
	if _, ok := endVal.(runtime.IteratorEndValue); !ok {
		t.Fatalf("expected IteratorEnd, got %#v", endVal)
	}
}

func TestIteratorWhileLoopBodyWithinGenerator(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()

	mustEval := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, env)
		if err != nil {
			t.Fatalf("evaluation failed: %v", err)
		}
		return val
	}

	mustEval(ast.Assign(ast.ID("count"), ast.Int(0)))

	if _, err := interp.evaluateExpression(ast.Assign(ast.ID("iter"), ast.IteratorLit(
		ast.Wloop(
			ast.Bin("<", ast.ID("count"), ast.Int(3)),
			ast.Yield(ast.ID("count")),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("count"), ast.Bin("+", ast.ID("count"), ast.Int(1))),
		),
	)), env); err != nil {
		t.Fatalf("iterator literal construction failed: %v", err)
	}

	nextCall := ast.CallExpr(ast.Member(ast.ID("iter"), ast.ID("next")))

	expectValue := func(expected int64) {
		val, err := interp.evaluateExpression(nextCall, env)
		if err != nil {
			t.Fatalf("next() failed: %v", err)
		}
		intVal, ok := val.(runtime.IntegerValue)
		if !ok || intVal.Val.Int64() != expected {
			t.Fatalf("expected %d, got %#v", expected, val)
		}
	}

	expectValue(0)
	expectValue(1)
	expectValue(2)
	endVal, err := interp.evaluateExpression(nextCall, env)
	if err != nil {
		t.Fatalf("next() failed: %v", err)
	}
	if _, ok := endVal.(runtime.IteratorEndValue); !ok {
		t.Fatalf("expected IteratorEnd, got %#v", endVal)
	}
	if got := mustGetInt(t, env, "count"); got != 3 {
		t.Fatalf("expected count = 3, got %d", got)
	}
}

func TestIteratorIfExpressionWithinGenerator(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()

	mustEval := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, env)
		if err != nil {
			t.Fatalf("evaluation failed: %v", err)
		}
		return val
	}

	mustEval(ast.Assign(ast.ID("calls"), ast.Int(0)))

	fn := ast.Fn(
		"tick",
		nil,
		[]ast.Statement{
			ast.AssignOp(ast.AssignmentAssign, ast.ID("calls"), ast.Bin("+", ast.ID("calls"), ast.Int(1))),
			ast.Ret(ast.Bool(true)),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)
	if _, err := interp.evaluateStatement(fn, env); err != nil {
		t.Fatalf("failed to define tick: %v", err)
	}

	if _, err := interp.evaluateExpression(ast.Assign(ast.ID("iter"), ast.IteratorLit(
		ast.Iff(
			ast.Call("tick"),
			ast.Yield(ast.Int(1)),
			ast.Yield(ast.Int(2)),
		),
	)), env); err != nil {
		t.Fatalf("iterator literal construction failed: %v", err)
	}

	nextCall := ast.CallExpr(ast.Member(ast.ID("iter"), ast.ID("next")))

	expectValue := func(expected int64) {
		val, err := interp.evaluateExpression(nextCall, env)
		if err != nil {
			t.Fatalf("next() failed: %v", err)
		}
		intVal, ok := val.(runtime.IntegerValue)
		if !ok || intVal.Val.Int64() != expected {
			t.Fatalf("expected %d, got %#v", expected, val)
		}
	}

	expectValue(1)
	expectValue(2)
	endVal, err := interp.evaluateExpression(nextCall, env)
	if err != nil {
		t.Fatalf("next() failed: %v", err)
	}
	if _, ok := endVal.(runtime.IteratorEndValue); !ok {
		t.Fatalf("expected IteratorEnd, got %#v", endVal)
	}
	if got := mustGetInt(t, env, "calls"); got != 1 {
		t.Fatalf("expected tick evaluated once, got %d", got)
	}
}

func TestIteratorMatchExpressionWithinGenerator(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()

	if _, err := interp.evaluateExpression(ast.Assign(ast.ID("subject_calls"), ast.Int(0)), env); err != nil {
		t.Fatalf("failed to assign subject_calls: %v", err)
	}
	if _, err := interp.evaluateExpression(ast.Assign(ast.ID("guard_calls"), ast.Int(0)), env); err != nil {
		t.Fatalf("failed to assign guard_calls: %v", err)
	}

	getSubject := ast.Fn(
		"getSubject",
		nil,
		[]ast.Statement{
			ast.AssignOp(ast.AssignmentAssign, ast.ID("subject_calls"), ast.Bin("+", ast.ID("subject_calls"), ast.Int(1))),
			ast.Ret(ast.Int(1)),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)
	if _, err := interp.evaluateStatement(getSubject, env); err != nil {
		t.Fatalf("failed to define getSubject: %v", err)
	}

	guardCheck := ast.Fn(
		"guardCheck",
		[]*ast.FunctionParameter{ast.Param("value", nil)},
		[]ast.Statement{
			ast.AssignOp(ast.AssignmentAssign, ast.ID("guard_calls"), ast.Bin("+", ast.ID("guard_calls"), ast.Int(1))),
			ast.Ret(ast.Bool(true)),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)
	if _, err := interp.evaluateStatement(guardCheck, env); err != nil {
		t.Fatalf("failed to define guardCheck: %v", err)
	}

	if _, err := interp.evaluateExpression(ast.Assign(ast.ID("iter"), ast.IteratorLit(
		ast.Match(
			ast.Call("getSubject"),
			ast.Mc(ast.LitP(ast.Int(1)),
				ast.Block(
					ast.Yield(ast.Int(10)),
					ast.Yield(ast.Int(11)),
				),
				ast.Call("guardCheck", ast.Int(1)),
			),
		),
	)), env); err != nil {
		t.Fatalf("iterator literal construction failed: %v", err)
	}

	nextCall := ast.CallExpr(ast.Member(ast.ID("iter"), ast.ID("next")))

	expectValue := func(expected int64) {
		val, err := interp.evaluateExpression(nextCall, env)
		if err != nil {
			t.Fatalf("next() failed: %v", err)
		}
		intVal, ok := val.(runtime.IntegerValue)
		if !ok || intVal.Val.Int64() != expected {
			t.Fatalf("expected %d, got %#v", expected, val)
		}
	}

	expectValue(10)
	expectValue(11)
	endVal, err := interp.evaluateExpression(nextCall, env)
	if err != nil {
		t.Fatalf("next() failed: %v", err)
	}
	if _, ok := endVal.(runtime.IteratorEndValue); !ok {
		t.Fatalf("expected IteratorEnd, got %#v", endVal)
	}

	if got := mustGetInt(t, env, "subject_calls"); got != 1 {
		t.Fatalf("expected subject evaluated once, got %d", got)
	}
	if got := mustGetInt(t, env, "guard_calls"); got != 1 {
		t.Fatalf("expected guard evaluated once, got %d", got)
	}
}

func mustGetInt(t *testing.T, env *runtime.Environment, name string) int {
	t.Helper()
	val, err := env.Get(name)
	if err != nil {
		t.Fatalf("failed to read %s: %v", name, err)
	}
	intVal, ok := val.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected %s to be integer, got %#v", name, val)
	}
	if !intVal.Val.IsInt64() {
		t.Fatalf("expected %s to fit in int64, got %s", name, intVal.Val.String())
	}
	return int(intVal.Val.Int64())
}
