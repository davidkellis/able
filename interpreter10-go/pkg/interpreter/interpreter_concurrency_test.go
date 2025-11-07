package interpreter

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
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

func TestSerialExecutorProcYieldFairness(t *testing.T) {
	interp := New()
	if _, ok := interp.executor.(*SerialExecutor); !ok {
		t.Fatalf("expected SerialExecutor by default")
	}
	global := interp.GlobalEnvironment()

	mustEval := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, global)
		if err != nil {
			t.Fatalf("expression evaluation failed: %v", err)
		}
		return val
	}

	mustEval(ast.Assign(ast.ID("trace"), ast.Str("")))
	mustEval(ast.Assign(ast.ID("stage_a"), ast.Int(0)))
	mustEval(ast.Assign(ast.ID("stage_b"), ast.Int(0)))

	appendTrace := func(prefix string) ast.Expression {
		return ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("trace"),
			ast.Bin("+", ast.ID("trace"), ast.Str(prefix)),
		)
	}

	assignStage := func(name string, value int64) ast.Expression {
		return ast.AssignOp(ast.AssignmentAssign, ast.ID(name), ast.Int(value))
	}

	mustEval(ast.Assign(ast.ID("worker"), ast.Proc(ast.Block(
		ast.IfExpr(
			ast.Bin("==", ast.ID("stage_a"), ast.Int(0)),
			ast.Block(
				appendTrace("A1"),
				assignStage("stage_a", 1),
				ast.Call("proc_yield"),
			),
		),
		ast.IfExpr(
			ast.Bin("==", ast.ID("stage_a"), ast.Int(1)),
			ast.Block(
				appendTrace("A2"),
				assignStage("stage_a", 2),
			),
		),
		ast.Int(0),
	))))

	mustEval(ast.Assign(ast.ID("other"), ast.Proc(ast.Block(
		ast.IfExpr(
			ast.Bin("==", ast.ID("stage_b"), ast.Int(0)),
			ast.Block(
				appendTrace("B1"),
				assignStage("stage_b", 1),
				ast.Call("proc_yield"),
			),
		),
		ast.IfExpr(
			ast.Bin("==", ast.ID("stage_b"), ast.Int(1)),
			ast.Block(
				appendTrace("B2"),
				assignStage("stage_b", 2),
			),
		),
		ast.Int(0),
	))))

	mustEval(ast.Call("proc_flush"))

	traceVal := mustEval(ast.ID("trace"))
	traceStr, ok := traceVal.(runtime.StringValue)
	if !ok {
		t.Fatalf("expected trace to be a string, got %#v", traceVal)
	}
	if traceStr.Val != "A1B1A2B2" {
		t.Fatalf("expected trace to be A1B1A2B2, got %q", traceStr.Val)
	}

	getStatusName := func(handle *runtime.ProcHandleValue) string {
		val := interp.procHandleStatus(handle)
		inst, ok := val.(*runtime.StructInstanceValue)
		if !ok {
			t.Fatalf("expected struct instance status, got %#v", val)
		}
		if inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil {
			t.Fatalf("status definition not initialised: %#v", inst)
		}
		return inst.Definition.Node.ID.Name
	}

	workerHandleVal, err := global.Get("worker")
	if err != nil {
		t.Fatalf("failed to read worker handle: %v", err)
	}
	workerHandle, ok := workerHandleVal.(*runtime.ProcHandleValue)
	if !ok {
		t.Fatalf("expected worker handle, got %#v", workerHandleVal)
	}
	otherHandleVal, err := global.Get("other")
	if err != nil {
		t.Fatalf("failed to read other handle: %v", err)
	}
	otherHandle, ok := otherHandleVal.(*runtime.ProcHandleValue)
	if !ok {
		t.Fatalf("expected other handle, got %#v", otherHandleVal)
	}

	if status := getStatusName(workerHandle); status != "Resolved" {
		t.Fatalf("expected worker status Resolved, got %s", status)
	}
	if status := getStatusName(otherHandle); status != "Resolved" {
		t.Fatalf("expected other status Resolved, got %s", status)
	}
}

func TestSerialExecutorFutureValueReentrancy(t *testing.T) {
	interp := New()
	serial, ok := interp.executor.(*SerialExecutor)
	if !ok {
		t.Fatalf("expected SerialExecutor by default")
	}
	if serial == nil {
		t.Fatalf("serial executor is nil")
	}
	global := interp.GlobalEnvironment()

	mustEval := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, global)
		if err != nil {
			t.Fatalf("expression evaluation failed: %v", err)
		}
		return val
	}

	appendTrace := func(label string) ast.Expression {
		return ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("trace"),
			ast.Bin("+", ast.ID("trace"), ast.Str(label)),
		)
	}

	mustEval(ast.Assign(ast.ID("trace"), ast.Str("")))

	mustEval(ast.Assign(
		ast.ID("inner"),
		ast.Spawn(ast.Block(
			appendTrace("I"),
			appendTrace("J"),
			ast.Str("X"),
		)),
	))

	mustEval(ast.Assign(
		ast.ID("outer"),
		ast.Spawn(ast.Block(
			appendTrace("O"),
			ast.Assign(
				ast.ID("result"),
				ast.CallExpr(ast.Member(ast.ID("inner"), "value")),
			),
			ast.AssignOp(
				ast.AssignmentAssign,
				ast.ID("trace"),
				ast.Interp(ast.ID("trace"), ast.ID("result")),
			),
			ast.Str("done"),
		)),
	))

	mustEval(ast.Assign(
		ast.ID("final"),
		ast.CallExpr(ast.Member(ast.ID("outer"), "value")),
	))

	val := mustEval(ast.Interp(ast.ID("trace"), ast.ID("final")))
	str, ok := val.(runtime.StringValue)
	if !ok {
		t.Fatalf("expected string value, got %#v", val)
	}
	if str.Val != "OIJXdone" {
		t.Fatalf("unexpected trace output %q", str.Val)
	}
}

func TestSerialExecutorProcValueReentrancy(t *testing.T) {
	interp := New()
	if _, ok := interp.executor.(*SerialExecutor); !ok {
		t.Fatalf("expected SerialExecutor by default")
	}
	global := interp.GlobalEnvironment()

	mustEval := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, global)
		if err != nil {
			t.Fatalf("expression evaluation failed: %v", err)
		}
		return val
	}

	appendTrace := func(label string) ast.Expression {
		return ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("trace"),
			ast.Bin("+", ast.ID("trace"), ast.Str(label)),
		)
	}

	mustEval(ast.Assign(ast.ID("trace"), ast.Str("")))

	mustEval(ast.Assign(
		ast.ID("inner"),
		ast.Proc(ast.Block(
			appendTrace("I"),
			appendTrace("J"),
			ast.Str("X"),
		)),
	))

	mustEval(ast.Assign(
		ast.ID("outer"),
		ast.Proc(ast.Block(
			appendTrace("O"),
			ast.Assign(
				ast.ID("result"),
				ast.CallExpr(ast.Member(ast.ID("inner"), "value")),
			),
			ast.AssignOp(
				ast.AssignmentAssign,
				ast.ID("trace"),
				ast.Interp(ast.ID("trace"), ast.ID("result")),
			),
			ast.Str("done"),
		)),
	))

	mustEval(ast.Assign(
		ast.ID("final"),
		ast.CallExpr(ast.Member(ast.ID("outer"), "value")),
	))

	val := mustEval(ast.Interp(ast.ID("trace"), ast.ID("final")))
	str, ok := val.(runtime.StringValue)
	if !ok {
		t.Fatalf("expected string value, got %#v", val)
	}
	if str.Val != "OIJXdone" {
		t.Fatalf("unexpected trace output %q", str.Val)
	}
}

func TestProcHandleValueMemoizesResult(t *testing.T) {
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

	handleVal := mustEval(ast.Proc(ast.Block(
		ast.AssignOp(ast.AssignmentAdd, ast.ID("count"), ast.Int(1)),
		ast.Int(21),
	)))
	handle, ok := handleVal.(*runtime.ProcHandleValue)
	if !ok {
		t.Fatalf("expected proc handle, got %#v", handleVal)
	}

	first := interp.procHandleValue(handle)
	second := interp.procHandleValue(handle)

	intVal, ok := first.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(21)) != 0 {
		t.Fatalf("expected first value 21, got %#v", first)
	}
	intVal, ok = second.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(21)) != 0 {
		t.Fatalf("expected memoized value 21, got %#v", second)
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

func TestProcHandleValueCancellationMemoized(t *testing.T) {
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

	handle.RequestCancel()

	first := interp.procHandleValue(handle)
	second := interp.procHandleValue(handle)

	if valueToString(first) != valueToString(second) {
		t.Fatalf("expected repeated value() calls to return identical errors, got %q vs %q", valueToString(first), valueToString(second))
	}
	if !strings.Contains(valueToString(first), "Proc cancelled") {
		t.Fatalf("expected cancellation error, got %q", valueToString(first))
	}
}

func TestConcurrentProcsSharedStateWithMutex(t *testing.T) {
	interp := New()
	if serial, ok := interp.executor.(*SerialExecutor); ok {
		serial.Close()
	}
	interp.executor = NewGoroutineExecutor(nil)
	global := interp.GlobalEnvironment()

	var mu sync.Mutex
	acquire := &runtime.NativeFunctionValue{
		Name:  "lock_acquire",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			mu.Lock()
			return runtime.NilValue{}, nil
		},
	}
	release := &runtime.NativeFunctionValue{
		Name:  "lock_release",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			mu.Unlock()
			return runtime.NilValue{}, nil
		},
	}
	global.Define("lock_acquire", acquire)
	global.Define("lock_release", release)

	mustEval := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, global)
		if err != nil {
			t.Fatalf("expression evaluation failed: %v", err)
		}
		return val
	}

	mustEval(ast.Assign(ast.ID("trace"), ast.Str("")))

	letters := []string{"A", "B", "C", "D"}
	handles := make([]*runtime.ProcHandleValue, 0, len(letters))
	for _, letter := range letters {
		handleVal := mustEval(ast.Proc(ast.Block(
			ast.Call("lock_acquire"),
			ast.AssignOp(
				ast.AssignmentAssign,
				ast.ID("trace"),
				ast.Bin("+", ast.ID("trace"), ast.Str(letter)),
			),
			ast.Call("lock_release"),
			ast.Int(0),
		)))
		handle, ok := handleVal.(*runtime.ProcHandleValue)
		if !ok {
			t.Fatalf("expected proc handle, got %#v", handleVal)
		}
		handles = append(handles, handle)
	}

	for _, handle := range handles {
		val := interp.procHandleValue(handle)
		if _, ok := val.(runtime.IntegerValue); !ok {
			if _, isNil := val.(runtime.NilValue); !isNil {
				t.Fatalf("expected proc to resolve with value, got %#v", val)
			}
		}
	}

	traceVal, err := global.Get("trace")
	if err != nil {
		t.Fatalf("failed to read trace: %v", err)
	}
	traceStr, ok := traceVal.(runtime.StringValue)
	if !ok {
		t.Fatalf("expected trace to be string, got %#v", traceVal)
	}
	if len(traceStr.Val) != len(letters) {
		t.Fatalf("expected trace length %d, got %d", len(letters), len(traceStr.Val))
	}
	for _, letter := range letters {
		if strings.Count(traceStr.Val, letter) != 1 {
			t.Fatalf("expected trace to contain %q exactly once, got %q", letter, traceStr.Val)
		}
	}
}

func TestGoroutineExecutorRunsProcsInParallel(t *testing.T) {
	interp := New()
	if serial, ok := interp.executor.(*SerialExecutor); ok {
		serial.Close()
	}
	interp.executor = NewGoroutineExecutor(nil)
	global := interp.GlobalEnvironment()

	sleepFn := &runtime.NativeFunctionValue{
		Name:  "sleep_ms",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("sleep_ms expects 1 argument")
			}
			intVal, ok := args[0].(runtime.IntegerValue)
			if !ok || intVal.Val == nil {
				return nil, fmt.Errorf("sleep_ms expects integer argument")
			}
			ms := intVal.Val.Int64()
			if ms < 0 {
				return nil, fmt.Errorf("sleep_ms expects non-negative duration")
			}
			time.Sleep(time.Duration(ms) * time.Millisecond)
			return runtime.NilValue{}, nil
		},
	}
	global.Define("sleep_ms", sleepFn)

	const (
		taskCount  = 4
		sleepDelay = 30
	)

	start := time.Now()
	handles := make([]*runtime.ProcHandleValue, 0, taskCount)
	for idx := 0; idx < taskCount; idx++ {
		handleVal, err := interp.evaluateExpression(
			ast.Proc(ast.Block(
				ast.Call("sleep_ms", ast.Int(sleepDelay)),
				ast.Int(int64(idx)),
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
		handles = append(handles, handle)
	}

	for _, handle := range handles {
		result := interp.procHandleValue(handle)
		if _, ok := result.(runtime.IntegerValue); !ok {
			t.Fatalf("expected integer result, got %#v", result)
		}
	}

	elapsed := time.Since(start)
	sleepDuration := time.Duration(sleepDelay) * time.Millisecond
	serialDuration := sleepDuration * taskCount
	parallelThreshold := serialDuration/2 + sleepDuration
	if elapsed >= serialDuration {
		t.Fatalf("expected goroutine executor to run tasks concurrently; elapsed %v >= serial duration %v", elapsed, serialDuration)
	}
	if elapsed > parallelThreshold {
		t.Fatalf("expected elapsed time %v to be <= %v when running in parallel", elapsed, parallelThreshold)
	}
}

func waitForStatus(handle *runtime.ProcHandleValue, desired runtime.ProcStatus, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if handle.Status() == desired {
			return true
		}
		time.Sleep(time.Millisecond)
	}
	return handle.Status() == desired
}

func waitForEnvString(t *testing.T, env *runtime.Environment, name string, desired string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		current := mustGetString(t, env, name)
		if current == desired {
			return true
		}
		time.Sleep(time.Millisecond)
	}
	return mustGetString(t, env, name) == desired
}

func mustGetString(t *testing.T, env *runtime.Environment, name string) string {
	val, err := env.Get(name)
	if err != nil {
		t.Fatalf("failed to read %s: %v", name, err)
	}
	strVal, ok := val.(runtime.StringValue)
	if !ok {
		t.Fatalf("expected %s to be string, got %#v", name, val)
	}
	return strVal.Val
}

func mustGetBool(t *testing.T, env *runtime.Environment, name string) bool {
	val, err := env.Get(name)
	if err != nil {
		t.Fatalf("failed to read %s: %v", name, err)
	}
	boolVal, ok := val.(runtime.BoolValue)
	if !ok {
		t.Fatalf("expected %s to be bool, got %#v", name, val)
	}
	return boolVal.Val
}

type stubExecutor struct {
	flushCalls int
}

func (s *stubExecutor) RunProc(task ProcTask) *runtime.ProcHandleValue {
	handle := runtime.NewProcHandle()
	go func() {
		if task != nil {
			if _, err := task(context.Background()); err != nil {
				handle.Fail(runtime.ErrorValue{Message: err.Error()})
				return
			}
		}
		handle.Resolve(runtime.NilValue{})
	}()
	return handle
}

func (s *stubExecutor) RunFuture(task ProcTask) *runtime.FutureValue {
	handle := s.RunProc(task)
	return runtime.NewFutureFromHandle(handle)
}

func (s *stubExecutor) Flush() {
	s.flushCalls++
}
