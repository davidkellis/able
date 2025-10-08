package interpreter

import (
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestRescueExpressionPattern(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	expr := ast.Rescue(
		ast.Block(ast.Raise(ast.Int(42))),
		ast.Mc(ast.Wc(), ast.Int(7)),
	)
	val, err := interp.evaluateExpression(expr, env)
	if err != nil {
		t.Fatalf("rescue evaluation failed: %v", err)
	}
	intVal, ok := val.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(7)) != 0 {
		t.Fatalf("expected integer 7, got %#v", val)
	}
}

func TestOrElseExpressionBindsError(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	expr := ast.OrElse(
		ast.Prop(ast.Block(ast.Raise(ast.Str("x")))),
		"e",
		ast.Str("handled"),
	)
	val, err := interp.evaluateExpression(expr, env)
	if err != nil {
		t.Fatalf("or else evaluation failed: %v", err)
	}
	strVal, ok := val.(runtime.StringValue)
	if !ok || strVal.Val != "handled" {
		t.Fatalf("expected string 'handled', got %#v", val)
	}
}

func TestEnsureRunsOnError(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	if _, err := interp.evaluateExpression(ast.Assign(ast.ID("flag"), ast.Str("")), env); err != nil {
		t.Fatalf("failed to initialise flag: %v", err)
	}
	expr := ast.Ensure(
		ast.Rescue(
			ast.Block(ast.Raise(ast.Str("oops"))),
			ast.Mc(ast.Wc(), ast.Str("rescued")),
		),
		ast.AssignOp(ast.AssignmentAssign, ast.ID("flag"), ast.Str("done")),
	)
	val, err := interp.evaluateExpression(expr, env)
	if err != nil {
		t.Fatalf("ensure evaluation failed: %v", err)
	}
	strVal, ok := val.(runtime.StringValue)
	if !ok || strVal.Val != "rescued" {
		t.Fatalf("expected string 'rescued', got %#v", val)
	}
	flagVal, err := env.Get("flag")
	if err != nil {
		t.Fatalf("expected flag binding: %v", err)
	}
	flagStr, ok := flagVal.(runtime.StringValue)
	if !ok || flagStr.Val != "done" {
		t.Fatalf("expected flag 'done', got %#v", flagVal)
	}
}

func TestRethrowPropagatesToOuterRescue(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	inner := ast.Rescue(
		ast.Block(ast.Raise(ast.Str("oops"))),
		ast.Mc(ast.Wc(), ast.Block(ast.Rethrow())),
	)
	outer := ast.Rescue(
		inner,
		ast.Mc(ast.Wc(), ast.Str("handled")),
	)
	val, err := interp.evaluateExpression(outer, env)
	if err != nil {
		t.Fatalf("outer rescue failed: %v", err)
	}
	strVal, ok := val.(runtime.StringValue)
	if !ok || strVal.Val != "handled" {
		t.Fatalf("expected string 'handled', got %#v", val)
	}
}

func TestRaiseConvertsValueToErrorStruct(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	_, err := interp.evaluateStatement(ast.Raise(ast.Int(5)), env)
	if err == nil {
		t.Fatalf("expected raise to escape")
	}
	rs, ok := err.(raiseSignal)
	if !ok {
		t.Fatalf("expected raiseSignal, got %T", err)
	}
	errVal, ok := rs.value.(runtime.ErrorValue)
	if !ok {
		t.Fatalf("expected ErrorValue, got %#v", rs.value)
	}
	if errVal.Message == "" {
		t.Fatalf("expected error message")
	}
	if _, ok := errVal.Payload["value"]; !ok {
		t.Fatalf("expected payload with original value")
	}
}

func TestRescueGuardSkipsClause(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	expr := ast.Rescue(
		ast.Block(ast.Raise(ast.Str("boom"))),
		ast.Mc(ast.ID("msg"), ast.Str("guard"), ast.Bin("==", ast.ID("msg"), ast.Str("skip"))),
		ast.Mc(ast.Wc(), ast.Str("fallback")),
	)
	val, err := interp.evaluateExpression(expr, env)
	if err != nil {
		t.Fatalf("rescue evaluation failed: %v", err)
	}
	strVal, ok := val.(runtime.StringValue)
	if !ok || strVal.Val != "fallback" {
		t.Fatalf("expected fallback string, got %#v", val)
	}
}

func TestRescueNoMatchPropagatesError(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	expr := ast.Rescue(
		ast.Block(ast.Raise(ast.Str("boom"))),
		ast.Mc(ast.LitP(ast.Str("ok")), ast.Str("handled")),
	)
	_, err := interp.evaluateExpression(expr, env)
	if err == nil {
		t.Fatalf("expected error to propagate")
	}
	rs, ok := err.(raiseSignal)
	if !ok {
		t.Fatalf("expected raiseSignal, got %T", err)
	}
	errVal, ok := rs.value.(runtime.ErrorValue)
	if !ok {
		t.Fatalf("expected ErrorValue, got %#v", rs.value)
	}
	if errVal.Message != "boom" {
		t.Fatalf("expected message 'boom', got %q", errVal.Message)
	}
}

func TestPropagationExpressionRaisesOnErrorValue(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	env.Define("err", runtime.ErrorValue{Message: "bad"})
	expr := ast.Prop(ast.ID("err"))
	_, err := interp.evaluateExpression(expr, env)
	if err == nil {
		t.Fatalf("expected propagation to raise")
	}
	rs, ok := err.(raiseSignal)
	if !ok {
		t.Fatalf("expected raiseSignal, got %T", err)
	}
	errVal, ok := rs.value.(runtime.ErrorValue)
	if !ok || errVal.Message != "bad" {
		t.Fatalf("expected ErrorValue 'bad', got %#v", rs.value)
	}
}

func TestPropagationExpressionPassesThroughNonError(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	expr := ast.Prop(ast.Int(9))
	val, err := interp.evaluateExpression(expr, env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := val.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(9)) != 0 {
		t.Fatalf("expected integer 9, got %#v", val)
	}
}

func TestOrElseExpressionHandlesRaise(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.OrElse(
			ast.Prop(ast.Block(ast.Raise(ast.Str("x")))),
			"e",
			ast.Str("handled"),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "handled" {
		t.Fatalf("expected 'handled', got %#v", result)
	}
}

func TestRescueExpressionTypedPattern(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Rescue(
			ast.Block(
				ast.Raise(ast.Str("boom")),
			),
			ast.Mc(
				ast.TypedP(ast.ID("err"), ast.Ty("Error")),
				ast.Str("caught"),
			),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "caught" {
		t.Fatalf("expected 'caught', got %#v", result)
	}
}

func TestEnsureExpressionRunsFinally(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("flag"), ast.Str("")),
		ast.Assign(
			ast.ID("value"),
			ast.Ensure(
				ast.Rescue(
					ast.Block(ast.Raise(ast.Str("oops"))),
					ast.Mc(ast.Wc(), ast.Str("rescued")),
				),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("flag"), ast.Str("done")),
			),
		),
		ast.ID("value"),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if str, ok := result.(runtime.StringValue); !ok || str.Val != "rescued" {
		t.Fatalf("expected ensure to return 'rescued', got %#v", result)
	}
	rescued, err := env.Get("flag")
	if err != nil {
		t.Fatalf("expected binding for flag: %v", err)
	}
	if str, ok := rescued.(runtime.StringValue); !ok || str.Val != "done" {
		t.Fatalf("expected final flag value 'done', got %#v", rescued)
	}
}

func TestRethrowStatementBubblesToOuterRescue(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(
			ast.ID("result"),
			ast.Rescue(
				ast.Rescue(
					ast.Block(ast.Raise(ast.Str("oops"))),
					ast.Mc(ast.Wc(), ast.Block(ast.Rethrow())),
				),
				ast.Mc(ast.Wc(), ast.Str("handled")),
			),
		),
		ast.ID("result"),
	}, nil, nil)

	value, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if str, ok := value.(runtime.StringValue); !ok || str.Val != "handled" {
		t.Fatalf("expected 'handled', got %#v", value)
	}
}

func TestErrorPayloadDestructuring(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Rescue(
			ast.Block(ast.Raise(ast.Str("fail"))),
			ast.Mc(
				ast.StructP([]*ast.StructPatternField{
					ast.FieldP(ast.ID("value"), "value", nil),
				}, false, nil),
				ast.Str("handled"),
			),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if str, ok := result.(runtime.StringValue); !ok || str.Val != "handled" {
		t.Fatalf("expected 'handled', got %#v", result)
	}
}
