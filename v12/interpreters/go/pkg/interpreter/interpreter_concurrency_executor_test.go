package interpreter

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestSerialExecutorFutureYieldFairness(t *testing.T) {
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

	mustEval(ast.Assign(ast.ID("worker"), ast.Spawn(ast.Block(
		ast.IfExpr(
			ast.Bin("==", ast.ID("stage_a"), ast.Int(0)),
			ast.Block(
				appendTrace("A1"),
				assignStage("stage_a", 1),
				ast.Call("future_yield"),
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

	mustEval(ast.Assign(ast.ID("other"), ast.Spawn(ast.Block(
		ast.IfExpr(
			ast.Bin("==", ast.ID("stage_b"), ast.Int(0)),
			ast.Block(
				appendTrace("B1"),
				assignStage("stage_b", 1),
				ast.Call("future_yield"),
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

	mustEval(ast.Call("future_flush"))

	traceVal := mustEval(ast.ID("trace"))
	traceStr, ok := traceVal.(runtime.StringValue)
	if !ok {
		t.Fatalf("expected trace to be a string, got %#v", traceVal)
	}
	if traceStr.Val != "A1B1A2B2" {
		t.Fatalf("expected trace to be A1B1A2B2, got %q", traceStr.Val)
	}

	getStatusName := func(handle *runtime.FutureValue) string {
		val := interp.futureStatus(handle)
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
	workerHandle, ok := workerHandleVal.(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected worker handle, got %#v", workerHandleVal)
	}
	otherHandleVal, err := global.Get("other")
	if err != nil {
		t.Fatalf("failed to read other handle: %v", err)
	}
	otherHandle, ok := otherHandleVal.(*runtime.FutureValue)
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
	// The SerialExecutor runs the inner future to completion when outer calls value();
	// the observed trace reflects inner first, then the outer append + concatenation.
	if str.Val != "IJOXdone" {
		t.Fatalf("unexpected trace output %q", str.Val)
	}
}

func TestSerialExecutorFutureValueReentrancySynchronousSection(t *testing.T) {
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

	serial.beginSynchronousSection()
	defer serial.endSynchronousSection()

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
	if str.Val != "IJOXdone" {
		t.Fatalf("unexpected trace output %q", str.Val)
	}
}

func TestFutureValueMemoizesResult(t *testing.T) {
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

	handleVal := mustEval(ast.Spawn(ast.Block(
		ast.AssignOp(ast.AssignmentAdd, ast.ID("count"), ast.Int(1)),
		ast.Int(21),
	)))
	handle, ok := handleVal.(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected future handle, got %#v", handleVal)
	}

	first := interp.futureValue(handle)
	second := interp.futureValue(handle)

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

func TestFutureValueCancellationMemoized(t *testing.T) {
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

	if handle != nil {
		handle.RequestCancel()
	}

	first := interp.futureValue(handle)
	second := interp.futureValue(handle)

	if valueToString(first) != valueToString(second) {
		t.Fatalf("expected repeated value() calls to return identical errors, got %q vs %q", valueToString(first), valueToString(second))
	}
	if !strings.Contains(valueToString(first), "Future cancelled") {
		t.Fatalf("expected cancellation error, got %q", valueToString(first))
	}
}

func TestConcurrentFuturesSharedStateWithMutex(t *testing.T) {
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
	handles := make([]*runtime.FutureValue, 0, len(letters))
	for _, letter := range letters {
		handleVal := mustEval(ast.Spawn(ast.Block(
			ast.Call("lock_acquire"),
			ast.AssignOp(
				ast.AssignmentAssign,
				ast.ID("trace"),
				ast.Bin("+", ast.ID("trace"), ast.Str(letter)),
			),
			ast.Call("lock_release"),
			ast.Int(0),
		)))
		handle, ok := handleVal.(*runtime.FutureValue)
		if !ok {
			t.Fatalf("expected future handle, got %#v", handleVal)
		}
		handles = append(handles, handle)
	}

	for _, handle := range handles {
		val := interp.futureValue(handle)
		if _, ok := val.(runtime.IntegerValue); !ok {
			if _, isNil := val.(runtime.NilValue); !isNil {
				t.Fatalf("expected future to resolve with value, got %#v", val)
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

func TestGoroutineExecutorRunsFuturesInParallel(t *testing.T) {
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
	handles := make([]*runtime.FutureValue, 0, taskCount)
	for idx := 0; idx < taskCount; idx++ {
		handleVal, err := interp.evaluateExpression(
			ast.Spawn(ast.Block(
				ast.Call("sleep_ms", ast.Int(sleepDelay)),
				ast.Int(int64(idx)),
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
		handles = append(handles, handle)
	}

	for _, handle := range handles {
		result := interp.futureValue(handle)
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
