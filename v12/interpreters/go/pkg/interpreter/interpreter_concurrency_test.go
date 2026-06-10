package interpreter

import (
	"strings"
	"testing"
	"time"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestFutureHandleResolvesValue(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	handleVal, err := interp.evaluateExpression(
		ast.Spawn(ast.Block(ast.Int(5))),
		global,
	)
	if err != nil {
		t.Fatalf("spawn expression evaluation failed: %v", err)
	}
	handle, ok := handleVal.(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected future handle, got %#v", handleVal)
	}

	valueVal := interp.futureValue(handle)
	intVal, ok := valueVal.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %#v", valueVal)
	}
	if intVal.BigInt().Cmp(bigInt(5)) != 0 {
		t.Fatalf("expected value 5, got %v", intVal.Val)
	}

	statusVal := interp.futureStatus(handle)
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

func TestFutureHandleFailureStatusAndValue(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	handleVal, err := interp.evaluateExpression(
		ast.Spawn(ast.Block(ast.Raise(ast.Str("boom")))),
		global,
	)
	if err != nil {
		t.Fatalf("spawn evaluation failed: %v", err)
	}
	handle, ok := handleVal.(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected future handle, got %#v", handleVal)
	}

	valueVal := interp.futureValue(handle)
	errValue, ok := valueVal.(runtime.ErrorValue)
	if !ok {
		t.Fatalf("expected runtime error value, got %#v", valueVal)
	}
	if errValue.Message != "Future failed: boom" {
		t.Fatalf("unexpected error message %q", errValue.Message)
	}

	statusVal := interp.futureStatus(handle)
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
	if details := interp.futureErrorDetails(errField); details != "boom" {
		t.Fatalf("expected future error details 'boom', got %q", details)
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
	if intVal.BigInt().Cmp(bigInt(7)) != 0 {
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

func TestTreewalkerGoroutineExecutorConcurrentArrayTasks(t *testing.T) {
	interp := NewWithExecutor(NewGoroutineExecutor(nil))
	global := interp.GlobalEnvironment()

	shared, err := interp.evaluateExpression(ast.Arr(ast.Int(7), ast.Int(11), ast.Int(13)), global)
	if err != nil {
		t.Fatalf("create shared array: %v", err)
	}
	global.Define("shared", shared)

	const workers = 16
	handles := make([]*runtime.FutureValue, 0, workers)
	for worker := 0; worker < workers; worker++ {
		value, err := interp.evaluateExpression(
			ast.Spawn(ast.Block(
				ast.Arr(ast.Int(int64(worker)), ast.Int(int64(worker+1))),
				ast.Index(ast.ID("shared"), ast.Int(int64(worker%3))),
			)),
			global,
		)
		if err != nil {
			t.Fatalf("spawn worker %d: %v", worker, err)
		}
		handle, ok := value.(*runtime.FutureValue)
		if !ok {
			t.Fatalf("spawn worker %d returned %T, want *runtime.FutureValue", worker, value)
		}
		handles = append(handles, handle)
	}

	interp.executor.Flush()
	for worker, handle := range handles {
		value, failure, status := handle.Snapshot()
		if status != runtime.FutureResolved || failure != nil {
			t.Fatalf("worker %d status = %v, failure = %#v", worker, status, failure)
		}
		integer, ok := value.(runtime.IntegerValue)
		if !ok {
			t.Fatalf("worker %d value = %T, want runtime.IntegerValue", worker, value)
		}
		want := int64([]int{7, 11, 13}[worker%3])
		if integer.BigInt().Int64() != want {
			t.Fatalf("worker %d value = %d, want %d", worker, integer.BigInt().Int64(), want)
		}
	}
}

func TestFutureCancelBeforeStart(t *testing.T) {
	interp := New()
	if serial, ok := interp.executor.(*SerialExecutor); ok {
		serial.Close()
	}
	interp.executor = NewGoroutineExecutor(nil)
	global := interp.GlobalEnvironment()

	handleVal, err := interp.evaluateExpression(
		ast.Spawn(ast.Block(
			ast.Call("future_yield"),
			ast.Int(42),
		)),
		global,
	)
	if err != nil {
		t.Fatalf("future evaluation failed: %v", err)
	}
	handle, ok := handleVal.(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected future handle, got %#v", handleVal)
	}

	if handle != nil {
		handle.RequestCancel()
	}
	if !waitForStatus(handle, runtime.FutureCancelled, 100*time.Millisecond) {
		t.Fatalf("expected handle to enter cancelled state, got %v", futureStatus(handle))
	}

	valueVal := interp.futureValue(handle)
	errValue, ok := valueVal.(runtime.ErrorValue)
	if !ok {
		t.Fatalf("expected runtime error value, got %#v", valueVal)
	}
	if errValue.Message != "Future cancelled" {
		t.Fatalf("unexpected error message %q", errValue.Message)
	}

	statusVal := interp.futureStatus(handle)
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

func TestFutureTaskObservesCancellation(t *testing.T) {
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

	handleVal := mustEval(ast.Spawn(ast.Block(
		ast.AssignOp(ast.AssignmentAssign, ast.ID("stage"), ast.Bin("+", ast.ID("stage"), ast.Int(1))),
		ast.AssignOp(ast.AssignmentAssign, ast.ID("trace"), ast.Bin("+", ast.ID("trace"), ast.Str("w"))),
		ast.While(
			ast.Un("!", ast.Call("future_cancelled")),
			ast.Block(
				ast.Call("future_yield"),
			),
		),
		ast.AssignOp(ast.AssignmentAssign, ast.ID("trace"), ast.Bin("+", ast.ID("trace"), ast.Str("x"))),
		ast.AssignOp(ast.AssignmentAssign, ast.ID("saw_cancel"), ast.Call("future_cancelled")),
		ast.Int(0),
	)))
	handle, ok := handleVal.(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected future handle, got %#v", handleVal)
	}

	if !waitForEnvString(t, global, "trace", "w", 200*time.Millisecond) {
		t.Fatalf("expected trace to be \"w\" before cancellation, got %q", mustGetString(t, global, "trace"))
	}

	if handle != nil {
		handle.RequestCancel()
	}
	if !waitForStatus(handle, runtime.FutureCancelled, 200*time.Millisecond) {
		t.Fatalf("expected cancelled status, got %v", futureStatus(handle))
	}

	valueVal := interp.futureValue(handle)
	errValue, ok := valueVal.(runtime.ErrorValue)
	if !ok {
		t.Fatalf("expected runtime error value, got %#v", valueVal)
	}
	if errValue.Message != "Future cancelled" {
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
	if !ok || intVal.BigInt().Cmp(bigInt(1)) != 0 {
		t.Fatalf("expected future value 1, got %#v", first)
	}

	second := interp.futureValue(future)
	intVal, ok = second.(runtime.IntegerValue)
	if !ok || intVal.BigInt().Cmp(bigInt(1)) != 0 {
		t.Fatalf("expected memoized future value 1, got %#v", second)
	}

	countVal, err := global.Get("count")
	if err != nil {
		t.Fatalf("failed to read count: %v", err)
	}
	countInt, ok := countVal.(runtime.IntegerValue)
	if !ok || countInt.BigInt().Cmp(bigInt(1)) != 0 {
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

func TestFutureCancelledOutsideTask(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	if _, err := interp.evaluateExpression(ast.Call("future_cancelled"), global); err == nil {
		t.Fatalf("expected future_cancelled outside async context to error")
	} else if !strings.Contains(err.Error(), "future_cancelled must be called inside an asynchronous task") {
		t.Fatalf("unexpected error message %q", err.Error())
	}
}

func TestFutureFlushDelegatesToExecutor(t *testing.T) {
	interp := New()
	stub := &stubExecutor{}
	interp.executor = stub
	global := interp.GlobalEnvironment()

	val, err := interp.evaluateExpression(ast.Call("future_flush"), global)
	if err != nil {
		t.Fatalf("future_flush evaluation failed: %v", err)
	}
	if _, ok := val.(runtime.NilValue); !ok {
		t.Fatalf("expected future_flush to return nil, got %#v", val)
	}
	if stub.flushCalls != 1 {
		t.Fatalf("expected executor flush to be called exactly once, got %d", stub.flushCalls)
	}
}

func TestFuturePendingTasksSerialExecutor(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	initialVal, err := interp.evaluateExpression(ast.Call("future_pending_tasks"), global)
	if err != nil {
		t.Fatalf("future_pending_tasks failed: %v", err)
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

	pendingMid, err := interp.evaluateExpression(ast.Call("future_pending_tasks"), global)
	if err != nil {
		t.Fatalf("future_pending_tasks failed: %v", err)
	}
	if got := intFromValue(t, pendingMid); got <= 0 {
		t.Fatalf("expected pending tasks after spawn, got %d", got)
	}

	if _, err := interp.evaluateExpression(ast.Call("future_flush"), global); err != nil {
		t.Fatalf("future_flush failed: %v", err)
	}

	pendingEnd, err := interp.evaluateExpression(ast.Call("future_pending_tasks"), global)
	if err != nil {
		t.Fatalf("future_pending_tasks failed: %v", err)
	}
	if got := intFromValue(t, pendingEnd); got != 0 {
		t.Fatalf("expected queue to drain after flush, got %d", got)
	}
}

func TestFuturePendingTasksGoroutineExecutor(t *testing.T) {
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

	value, err := interp.evaluateExpression(ast.Call("future_pending_tasks"), global)
	if err != nil {
		t.Fatalf("future_pending_tasks failed: %v", err)
	}
	if got := intFromValue(t, value); got < 0 {
		t.Fatalf("goroutine executor pending tasks must be non-negative, got %d", got)
	}
}
