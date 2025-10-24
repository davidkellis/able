package interpreter

import (
	"testing"
	"time"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func newAsyncInterpreter(t *testing.T) *Interpreter {
	t.Helper()
	interp := New()
	if serial, ok := interp.executor.(*SerialExecutor); ok {
		serial.Close()
	}
	interp.executor = NewGoroutineExecutor(nil)
	return interp
}

func TestChannelSendReceiveBetweenProcs(t *testing.T) {
	interp := newAsyncInterpreter(t)
	global := interp.GlobalEnvironment()

	mustEval := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, global)
		if err != nil {
			t.Fatalf("evaluation failed: %v", err)
		}
		return val
	}

	mustEval(ast.Assign(ast.ID("ch"), ast.Call("__able_channel_new", ast.Int(0))))
	mustEval(ast.Assign(ast.ID("received"), ast.Nil()))

	consumer := mustEval(ast.Proc(ast.Block(
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("received"),
			ast.Call("__able_channel_receive", ast.ID("ch")),
		),
		ast.Nil(),
	)))
	consumerHandle, ok := consumer.(*runtime.ProcHandleValue)
	if !ok {
		t.Fatalf("expected proc handle for consumer, got %#v", consumer)
	}

	producer := mustEval(ast.Proc(ast.Block(
		ast.Call("__able_channel_send", ast.ID("ch"), ast.Str("hello")),
		ast.Nil(),
	)))
	producerHandle, ok := producer.(*runtime.ProcHandleValue)
	if !ok {
		t.Fatalf("expected proc handle for producer, got %#v", producer)
	}

	if !waitForStatus(consumerHandle, runtime.ProcResolved, 500*time.Millisecond) {
		t.Fatalf("consumer did not resolve: %v", consumerHandle.Status())
	}
	if !waitForStatus(producerHandle, runtime.ProcResolved, 500*time.Millisecond) {
		t.Fatalf("producer did not resolve: %v", producerHandle.Status())
	}

	if got := mustGetString(t, global, "received"); got != "hello" {
		t.Fatalf("expected received = \"hello\", got %q", got)
	}
}

func TestChannelTrySendReceive(t *testing.T) {
	interp := newAsyncInterpreter(t)
	global := interp.GlobalEnvironment()

	mustEval := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, global)
		if err != nil {
			t.Fatalf("evaluation failed: %v", err)
		}
		return val
	}

	mustEval(ast.Assign(ast.ID("ch"), ast.Call("__able_channel_new", ast.Int(1))))

	trySend := mustEval(ast.Call("__able_channel_try_send", ast.ID("ch"), ast.Int(5)))
	trySendBool, ok := trySend.(runtime.BoolValue)
	if !ok {
		t.Fatalf("expected bool from try_send, got %#v", trySend)
	}
	if !trySendBool.Val {
		t.Fatalf("expected first try_send to succeed")
	}

	trySend2 := mustEval(ast.Call("__able_channel_try_send", ast.ID("ch"), ast.Int(9)))
	trySend2Bool, ok := trySend2.(runtime.BoolValue)
	if !ok {
		t.Fatalf("expected bool from second try_send, got %#v", trySend2)
	}
	if trySend2Bool.Val {
		t.Fatalf("expected second try_send to fail due to full buffer")
	}

	recv1 := mustEval(ast.Call("__able_channel_try_receive", ast.ID("ch")))
	intVal, ok := recv1.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(5)) != 0 {
		t.Fatalf("expected first try_receive = 5, got %#v", recv1)
	}

	recv2 := mustEval(ast.Call("__able_channel_try_receive", ast.ID("ch")))
	if _, ok := recv2.(runtime.NilValue); !ok {
		t.Fatalf("expected second try_receive to return nil, got %#v", recv2)
	}
}

func TestChannelCloseWakesReceivers(t *testing.T) {
	interp := newAsyncInterpreter(t)
	global := interp.GlobalEnvironment()

	mustEval := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, global)
		if err != nil {
			t.Fatalf("evaluation failed: %v", err)
		}
		return val
	}

	mustEval(ast.Assign(ast.ID("ch"), ast.Call("__able_channel_new", ast.Int(0))))
	mustEval(ast.Assign(ast.ID("result"), ast.Str("unset")))

	receiver := mustEval(ast.Proc(ast.Block(
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("result"),
			ast.Call("__able_channel_receive", ast.ID("ch")),
		),
		ast.Nil(),
	)))
	receiverHandle, ok := receiver.(*runtime.ProcHandleValue)
	if !ok {
		t.Fatalf("expected proc handle for receiver, got %#v", receiver)
	}

	if _, err := interp.evaluateExpression(ast.Call("__able_channel_close", ast.ID("ch")), global); err != nil {
		t.Fatalf("channel close failed: %v", err)
	}

	if !waitForStatus(receiverHandle, runtime.ProcResolved, 500*time.Millisecond) {
		t.Fatalf("receiver did not resolve after close: %v", receiverHandle.Status())
	}

	val, err := global.Get("result")
	if err != nil {
		t.Fatalf("failed to read result: %v", err)
	}
	if _, ok := val.(runtime.NilValue); !ok {
		t.Fatalf("expected result to be nil after close, got %#v", val)
	}
}

func TestMutexLocksCoordinateProcs(t *testing.T) {
	interp := newAsyncInterpreter(t)
	global := interp.GlobalEnvironment()

	mustEval := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, global)
		if err != nil {
			t.Fatalf("evaluation failed: %v", err)
		}
		return val
	}

	mustEval(ast.Assign(ast.ID("m"), ast.Call("__able_mutex_new")))
	mustEval(ast.Assign(ast.ID("trace"), ast.Str("")))
	mustEval(ast.Assign(ast.ID("stage"), ast.Str("")))

	appendTrace := func(letter string) ast.Expression {
		return ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("trace"),
			ast.Bin("+", ast.ID("trace"), ast.Str(letter)),
		)
	}

	first := mustEval(ast.Proc(ast.Block(
		ast.Call("__able_mutex_lock", ast.ID("m")),
		appendTrace("A"),
		ast.AssignOp(ast.AssignmentAssign, ast.ID("stage"), ast.Str("locked")),
		ast.Call("proc_yield"),
		appendTrace("B"),
		ast.Call("__able_mutex_unlock", ast.ID("m")),
		ast.Nil(),
	)))
	firstHandle, ok := first.(*runtime.ProcHandleValue)
	if !ok {
		t.Fatalf("expected proc handle for first worker, got %#v", first)
	}

	if !waitForEnvString(t, global, "stage", "locked", 200*time.Millisecond) {
		t.Fatalf("expected stage to reach 'locked'")
	}

	second := mustEval(ast.Proc(ast.Block(
		ast.Call("__able_mutex_lock", ast.ID("m")),
		appendTrace("C"),
		ast.Call("__able_mutex_unlock", ast.ID("m")),
		ast.Nil(),
	)))
	secondHandle, ok := second.(*runtime.ProcHandleValue)
	if !ok {
		t.Fatalf("expected proc handle for second worker, got %#v", second)
	}

	if !waitForStatus(firstHandle, runtime.ProcResolved, 500*time.Millisecond) {
		t.Fatalf("first worker did not resolve: %v", firstHandle.Status())
	}
	if !waitForStatus(secondHandle, runtime.ProcResolved, 500*time.Millisecond) {
		t.Fatalf("second worker did not resolve: %v", secondHandle.Status())
	}

	if got := mustGetString(t, global, "trace"); got != "ABC" {
		t.Fatalf("expected trace to be ABC, got %q", got)
	}
}

func TestNilChannelSendBlocksUntilCancelled(t *testing.T) {
	interp := newAsyncInterpreter(t)
	global := interp.GlobalEnvironment()

	mustEval := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, global)
		if err != nil {
			t.Fatalf("evaluation failed: %v", err)
		}
		return val
	}

	val := mustEval(ast.Proc(ast.Block(
		ast.Call("__able_channel_send", ast.Int(0), ast.Str("value")),
		ast.Nil(),
	)))
	handle, ok := val.(*runtime.ProcHandleValue)
	if !ok {
		t.Fatalf("expected proc handle, got %#v", val)
	}

	time.Sleep(10 * time.Millisecond)
	if status := handle.Status(); status != runtime.ProcPending {
		t.Fatalf("expected pending status before cancel, got %v", status)
	}

	handle.RequestCancel()
	if !waitForStatus(handle, runtime.ProcCancelled, 200*time.Millisecond) {
		t.Fatalf("expected handle to cancel after request, got %v", handle.Status())
	}
}

func TestNilChannelReceiveBlocksUntilCancelled(t *testing.T) {
	interp := newAsyncInterpreter(t)
	global := interp.GlobalEnvironment()

	mustEval := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, global)
		if err != nil {
			t.Fatalf("evaluation failed: %v", err)
		}
		return val
	}

	val := mustEval(ast.Proc(ast.Block(
		ast.Call("__able_channel_receive", ast.Int(0)),
		ast.Nil(),
	)))
	handle, ok := val.(*runtime.ProcHandleValue)
	if !ok {
		t.Fatalf("expected proc handle, got %#v", val)
	}

	time.Sleep(10 * time.Millisecond)
	if status := handle.Status(); status != runtime.ProcPending {
		t.Fatalf("expected pending status before cancel, got %v", status)
	}

	handle.RequestCancel()
	if !waitForStatus(handle, runtime.ProcCancelled, 200*time.Millisecond) {
		t.Fatalf("expected handle to cancel after request, got %v", handle.Status())
	}
}
