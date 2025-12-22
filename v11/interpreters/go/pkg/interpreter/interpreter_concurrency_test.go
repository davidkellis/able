package interpreter

import (
	"strings"
	"testing"
	"time"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestProcHandleResolvesValue(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	handleVal, err := interp.evaluateExpression(
		ast.Proc(ast.Block(ast.Int(5))),
		global,
	)
	if err != nil {
		t.Fatalf("proc expression evaluation failed: %v", err)
	}
	handle, ok := handleVal.(*runtime.ProcHandleValue)
	if !ok {
		t.Fatalf("expected proc handle, got %#v", handleVal)
	}

	valueVal := interp.procHandleValue(handle)
	intVal, ok := valueVal.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %#v", valueVal)
	}
	if intVal.Val.Cmp(bigInt(5)) != 0 {
		t.Fatalf("expected value 5, got %v", intVal.Val)
	}

	statusVal := interp.procHandleStatus(handle)
	statusInst, ok := statusVal.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("expected struct status value, got %#v", statusVal)
	}
	name := ""
	if statusInst.Definition != nil && statusInst.Definition.Node != nil && statusInst.Definition.Node.ID != nil {
		name = statusInst.Definition.Node.ID.Name
	}
	if name != "Resolved" {
		t.Fatalf("expected Resolved status, got %q", name)
	}
}

func TestProcHandleFailureStatusAndValue(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	handleVal, err := interp.evaluateExpression(
		ast.Proc(ast.Block(ast.Raise(ast.Str("boom")))),
		global,
	)
	if err != nil {
		t.Fatalf("proc evaluation failed: %v", err)
	}
	handle, ok := handleVal.(*runtime.ProcHandleValue)
	if !ok {
		t.Fatalf("expected proc handle, got %#v", handleVal)
	}

	valueVal := interp.procHandleValue(handle)
	errValue, ok := valueVal.(runtime.ErrorValue)
	if !ok {
		t.Fatalf("expected runtime error value, got %#v", valueVal)
	}
	if errValue.Message != "Proc failed: boom" {
		t.Fatalf("unexpected error message %q", errValue.Message)
	}

	statusVal := interp.procHandleStatus(handle)
	statusInst, ok := statusVal.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("expected struct status value, got %#v", statusVal)
	}
	name := ""
	if statusInst.Definition != nil && statusInst.Definition.Node != nil && statusInst.Definition.Node.ID != nil {
		name = statusInst.Definition.Node.ID.Name
	}
	if name != "Failed" {
		t.Fatalf("expected Failed status, got %q", name)
	}
	errField, ok := statusInst.Fields["error"]
	if !ok {
		t.Fatalf("expected error field on Failed status")
	}
	if details := interp.procErrorDetails(errField); details != "boom" {
		t.Fatalf("expected proc error details 'boom', got %q", details)
	}
}

func TestSpawnFutureValue(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	futureVal, err := interp.evaluateExpression(
		ast.Spawn(ast.Block(ast.Int(7))),
		global,
	)
	if err != nil {
		t.Fatalf("spawn expression failed: %v", err)
	}
	future, ok := futureVal.(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected future handle, got %#v", futureVal)
	}

	valueVal := interp.futureValue(future)
	intVal, ok := valueVal.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %#v", valueVal)
	}
	if intVal.Val.Cmp(bigInt(7)) != 0 {
		t.Fatalf("expected value 7, got %v", intVal.Val)
	}

	statusVal := interp.futureStatus(future)
	statusInst, ok := statusVal.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("expected struct status value, got %#v", statusVal)
	}
	name := ""
	if statusInst.Definition != nil && statusInst.Definition.Node != nil && statusInst.Definition.Node.ID != nil {
		name = statusInst.Definition.Node.ID.Name
	}
	if name != "Resolved" {
		t.Fatalf("expected Resolved status, got %q", name)
	}
}

func TestProcCancelBeforeStart(t *testing.T) {
	interp := New()
	if serial, ok := interp.executor.(*SerialExecutor); ok {
		serial.Close()
	}
	interp.executor = NewGoroutineExecutor(nil)
	global := interp.GlobalEnvironment()

	handleVal, err := interp.evaluateExpression(
		ast.Proc(ast.Block(
			ast.Call("proc_yield"),
			ast.Int(42),
		)),
		global,
	)
	if err != nil {
		t.Fatalf("proc evaluation failed: %v", err)
	}
	handle, ok := handleVal.(*runtime.ProcHandleValue)
	if !ok {
		t.Fatalf("expected proc handle, got %#v", handleVal)
	}

	handle.RequestCancel()
	if !waitForStatus(handle, runtime.ProcCancelled, 100*time.Millisecond) {
		t.Fatalf("expected handle to enter cancelled state, got %v", handle.Status())
	}

	valueVal := interp.procHandleValue(handle)
	errValue, ok := valueVal.(runtime.ErrorValue)
	if !ok {
		t.Fatalf("expected runtime error value, got %#v", valueVal)
	}
	if errValue.Message != "Proc cancelled" {
		t.Fatalf("unexpected error message %q", errValue.Message)
	}

	statusVal := interp.procHandleStatus(handle)
	statusInst, ok := statusVal.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("expected struct status value, got %#v", statusVal)
	}
	name := ""
	if statusInst.Definition != nil && statusInst.Definition.Node != nil && statusInst.Definition.Node.ID != nil {
		name = statusInst.Definition.Node.ID.Name
	}
	if name != "Cancelled" {
		t.Fatalf("expected Cancelled status, got %q", name)
	}
}

func TestProcTaskObservesCancellation(t *testing.T) {
	interp := New()
	if serial, ok := interp.executor.(*SerialExecutor); ok {
		serial.Close()
	}
	interp.executor = NewGoroutineExecutor(nil)
	global := interp.GlobalEnvironment()

	mustEval := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, global)
		if err != nil {
			t.Fatalf("expression evaluation failed: %v", err)
		}
		return val
	}

	mustEval(ast.Assign(ast.ID("trace"), ast.Str("")))
	mustEval(ast.Assign(ast.ID("saw_cancel"), ast.Bool(false)))
	mustEval(ast.Assign(ast.ID("stage"), ast.Int(0)))

	handleVal := mustEval(ast.Proc(ast.Block(
		ast.AssignOp(ast.AssignmentAssign, ast.ID("stage"), ast.Bin("+", ast.ID("stage"), ast.Int(1))),
		ast.AssignOp(ast.AssignmentAssign, ast.ID("trace"), ast.Bin("+", ast.ID("trace"), ast.Str("w"))),
		ast.While(
			ast.Un("!", ast.Call("proc_cancelled")),
			ast.Block(
				ast.Call("proc_yield"),
			),
		),
		ast.AssignOp(ast.AssignmentAssign, ast.ID("trace"), ast.Bin("+", ast.ID("trace"), ast.Str("x"))),
		ast.AssignOp(ast.AssignmentAssign, ast.ID("saw_cancel"), ast.Call("proc_cancelled")),
		ast.Int(0),
	)))
	handle, ok := handleVal.(*runtime.ProcHandleValue)
	if !ok {
		t.Fatalf("expected proc handle, got %#v", handleVal)
	}

	if !waitForEnvString(t, global, "trace", "w", 200*time.Millisecond) {
		t.Fatalf("expected trace to be \"w\" before cancellation, got %q", mustGetString(t, global, "trace"))
	}

	handle.RequestCancel()
	if !waitForStatus(handle, runtime.ProcCancelled, 200*time.Millisecond) {
		t.Fatalf("expected cancelled status, got %v", handle.Status())
	}

	valueVal := interp.procHandleValue(handle)
	errValue, ok := valueVal.(runtime.ErrorValue)
	if !ok {
		t.Fatalf("expected runtime error value, got %#v", valueVal)
	}
	if errValue.Message != "Proc cancelled" {
		t.Fatalf("unexpected error message %q", errValue.Message)
	}

	if got := mustGetString(t, global, "trace"); got != "wx" {
		t.Fatalf("expected trace \"wx\", got %q", got)
	}
	if got := mustGetBool(t, global, "saw_cancel"); !got {
		t.Fatalf("expected saw_cancel to be true")
	}
}

func TestFutureMemoizesResult(t *testing.T) {
	interp := New()
	if serial, ok := interp.executor.(*SerialExecutor); ok {
		serial.Close()
	}
	interp.executor = NewGoroutineExecutor(nil)
	global := interp.GlobalEnvironment()

	mustEval := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, global)
		if err != nil {
			t.Fatalf("expression evaluation failed: %v", err)
		}
		return val
	}

	mustEval(ast.Assign(ast.ID("count"), ast.Int(0)))

	futureVal := mustEval(ast.Spawn(ast.Block(
		ast.AssignOp(ast.AssignmentAdd, ast.ID("count"), ast.Int(1)),
		ast.Int(1),
	)))
	future, ok := futureVal.(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected future value, got %#v", futureVal)
	}

	first := interp.futureValue(future)
	intVal, ok := first.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(1)) != 0 {
		t.Fatalf("expected future value 1, got %#v", first)
	}

	second := interp.futureValue(future)
	intVal, ok = second.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(1)) != 0 {
		t.Fatalf("expected memoized future value 1, got %#v", second)
	}

	countVal, err := global.Get("count")
	if err != nil {
		t.Fatalf("failed to read count: %v", err)
	}
	countInt, ok := countVal.(runtime.IntegerValue)
	if !ok || countInt.Val.Cmp(bigInt(1)) != 0 {
		t.Fatalf("expected count to be 1, got %#v", countVal)
	}
}

func TestFutureFailurePropagates(t *testing.T) {
	interp := New()
	if serial, ok := interp.executor.(*SerialExecutor); ok {
		serial.Close()
	}
	interp.executor = NewGoroutineExecutor(nil)
	global := interp.GlobalEnvironment()

	futureVal, err := interp.evaluateExpression(
		ast.Spawn(ast.Block(ast.Raise(ast.Str("boom")))),
		global,
	)
	if err != nil {
		t.Fatalf("spawn evaluation failed: %v", err)
	}
	future, ok := futureVal.(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected future value, got %#v", futureVal)
	}

	valueVal := interp.futureValue(future)
	errValue, ok := valueVal.(runtime.ErrorValue)
	if !ok {
		t.Fatalf("expected runtime error value, got %#v", valueVal)
	}
	if errValue.Message != "Future failed: boom" {
		t.Fatalf("unexpected error message %q", errValue.Message)
	}

	statusVal := interp.futureStatus(future)
	statusInst, ok := statusVal.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("expected struct status value, got %#v", statusVal)
	}
	name := ""
	if statusInst.Definition != nil && statusInst.Definition.Node != nil && statusInst.Definition.Node.ID != nil {
		name = statusInst.Definition.Node.ID.Name
	}
	if name != "Failed" {
		t.Fatalf("expected Failed status, got %q", name)
	}
}

func TestProcCancelledOutsideProc(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	if _, err := interp.evaluateExpression(ast.Call("proc_cancelled"), global); err == nil {
		t.Fatalf("expected proc_cancelled outside async context to error")
	} else if !strings.Contains(err.Error(), "proc_cancelled must be called inside an asynchronous task") {
		t.Fatalf("unexpected error message %q", err.Error())
	}
}

func TestProcFlushDelegatesToExecutor(t *testing.T) {
	interp := New()
	stub := &stubExecutor{}
	interp.executor = stub
	global := interp.GlobalEnvironment()

	val, err := interp.evaluateExpression(ast.Call("proc_flush"), global)
	if err != nil {
		t.Fatalf("proc_flush evaluation failed: %v", err)
	}
	if _, ok := val.(runtime.NilValue); !ok {
		t.Fatalf("expected proc_flush to return nil, got %#v", val)
	}
	if stub.flushCalls != 1 {
		t.Fatalf("expected executor flush to be called exactly once, got %d", stub.flushCalls)
	}
}

func TestProcPendingTasksSerialExecutor(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	initialVal, err := interp.evaluateExpression(ast.Call("proc_pending_tasks"), global)
	if err != nil {
		t.Fatalf("proc_pending_tasks failed: %v", err)
	}
	if got := intFromValue(t, initialVal); got != 0 {
		t.Fatalf("expected empty queue, got %d", got)
	}

	task := ast.Spawn(ast.Block(
		ast.Int(1),
	))
	if _, err := interp.evaluateExpression(task, global); err != nil {
		t.Fatalf("spawn failed: %v", err)
	}
	if _, err := interp.evaluateExpression(task, global); err != nil {
		t.Fatalf("spawn failed: %v", err)
	}

	pendingMid, err := interp.evaluateExpression(ast.Call("proc_pending_tasks"), global)
	if err != nil {
		t.Fatalf("proc_pending_tasks failed: %v", err)
	}
	if got := intFromValue(t, pendingMid); got <= 0 {
		t.Fatalf("expected pending tasks after spawn, got %d", got)
	}

	if _, err := interp.evaluateExpression(ast.Call("proc_flush"), global); err != nil {
		t.Fatalf("proc_flush failed: %v", err)
	}

	pendingEnd, err := interp.evaluateExpression(ast.Call("proc_pending_tasks"), global)
	if err != nil {
		t.Fatalf("proc_pending_tasks failed: %v", err)
	}
	if got := intFromValue(t, pendingEnd); got != 0 {
		t.Fatalf("expected queue to drain after flush, got %d", got)
	}
}

func TestProcPendingTasksGoroutineExecutor(t *testing.T) {
	interp := New()
	if serial, ok := interp.executor.(*SerialExecutor); ok {
		serial.Close()
	}
	interp.executor = NewGoroutineExecutor(nil)
	global := interp.GlobalEnvironment()

	if _, err := interp.evaluateExpression(ast.Spawn(ast.Block(
		ast.Int(1),
	)), global); err != nil {
		t.Fatalf("spawn failed: %v", err)
	}

	value, err := interp.evaluateExpression(ast.Call("proc_pending_tasks"), global)
	if err != nil {
		t.Fatalf("proc_pending_tasks failed: %v", err)
	}
	if got := intFromValue(t, value); got < 0 {
		t.Fatalf("goroutine executor pending tasks must be non-negative, got %d", got)
	}
}
