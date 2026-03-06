package interpreter

import (
	"reflect"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_SpawnOrderingBeforeFlush(t *testing.T) {
	recordFn := ast.Fn(
		"record",
		[]*ast.FunctionParameter{
			ast.Param("msg", ast.Ty("String")),
		},
		[]ast.Statement{
			ast.AssignOp(ast.AssignmentAssign, ast.Index(ast.ID("events"), ast.ID("idx")), ast.ID("msg")),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("idx"), ast.Bin("+", ast.ID("idx"), ast.Int(1))),
			ast.Ret(ast.Nil()),
		},
		ast.Ty("void"),
		nil,
		nil,
		false,
		false,
	)

	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("events"), ast.Arr(
			ast.Str(""),
			ast.Str(""),
			ast.Str(""),
			ast.Str(""),
			ast.Str(""),
			ast.Str(""),
		)),
		ast.Assign(ast.ID("idx"), ast.Int(0)),
		recordFn,
		ast.Spawn(ast.Block(
			ast.CallExpr(ast.ID("record"), ast.Str("spawn1:start")),
			ast.Call("future_yield"),
			ast.CallExpr(ast.ID("record"), ast.Str("spawn1:end")),
			ast.Nil(),
		)),
		ast.Assign(ast.ID("future"), ast.Spawn(ast.Block(
			ast.CallExpr(ast.ID("record"), ast.Str("spawn2:start")),
			ast.Call("future_yield"),
			ast.CallExpr(ast.ID("record"), ast.Str("spawn2:end")),
			ast.Str("done"),
		))),
		ast.CallExpr(ast.ID("record"), ast.Str("main:before_flush")),
		ast.Call("future_flush"),
		ast.CallExpr(ast.ID("record"), ast.Str("main:after_flush")),
		ast.CallExpr(ast.Member(ast.ID("future"), "value")),
		ast.ID("events"),
	}, nil, nil)

	expected := []string{
		"main:before_flush",
		"spawn1:start",
		"spawn2:start",
		"spawn1:end",
		"spawn2:end",
		"main:after_flush",
	}

	for run := 0; run < 50; run++ {
		interp := NewBytecode()
		got := runBytecodeModuleWithInterpreter(t, interp, module)
		events := bytecodeEventStrings(t, interp, got)
		if !reflect.DeepEqual(events, expected) {
			t.Fatalf("run %d ordering mismatch: got=%v want=%v", run, events, expected)
		}
	}
}

func bytecodeEventStrings(t *testing.T, interp *Interpreter, value runtime.Value) []string {
	t.Helper()
	arr, ok := value.(*runtime.ArrayValue)
	if !ok || arr == nil {
		t.Fatalf("expected ArrayValue result, got %#v", value)
	}
	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("array state: %v", err)
	}
	events := make([]string, len(state.Values))
	for idx, entry := range state.Values {
		switch v := entry.(type) {
		case runtime.StringValue:
			events[idx] = v.Val
		case *runtime.StringValue:
			if v == nil {
				t.Fatalf("event %d is nil string pointer", idx)
			}
			events[idx] = v.Val
		default:
			t.Fatalf("event %d expected string, got %#v", idx, entry)
		}
	}
	return events
}
