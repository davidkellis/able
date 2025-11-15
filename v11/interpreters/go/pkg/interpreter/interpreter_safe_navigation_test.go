package interpreter

import (
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestSafeMemberAccessOnNilReturnsNil(t *testing.T) {
	interp := New()
	safeMember := ast.Member(ast.ID("user"), "profile")
	safeMember.Safe = true
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("user"), ast.Nil()),
		safeMember,
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("safe navigation on nil failed: %v", err)
	}
	if _, ok := result.(runtime.NilValue); !ok {
		t.Fatalf("expected nil result from safe navigation, got %#v", result)
	}
}

func TestSafeCallSkipsArgumentEvaluation(t *testing.T) {
	interp := New()
	safeMember := ast.Member(ast.ID("wrapper"), "call")
	safeMember.Safe = true
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("calls"), ast.Int(0)),
		ast.Fn(
			"trigger",
			nil,
			[]ast.Statement{
				ast.AssignOp(ast.AssignmentAssign, ast.ID("calls"), ast.Bin("+", ast.ID("calls"), ast.Int(1))),
				ast.Ret(ast.ID("calls")),
			},
			ast.Ty("i32"),
			nil,
			nil,
			false,
			false,
		),
		ast.Assign(ast.ID("wrapper"), ast.Nil()),
		ast.CallExpr(safeMember, ast.CallExpr(ast.ID("trigger"))),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("safe method call failed: %v", err)
	}
	if _, ok := result.(runtime.NilValue); !ok {
		t.Fatalf("expected nil result from safe call, got %#v", result)
	}
	callsVal, err := env.Get("calls")
	if err != nil {
		t.Fatalf("expected calls binding: %v", err)
	}
	intCalls, ok := callsVal.(runtime.IntegerValue)
	if !ok || intCalls.Val.Cmp(bigInt(0)) != 0 {
		t.Fatalf("expected calls to remain 0, got %#v", callsVal)
	}
}
