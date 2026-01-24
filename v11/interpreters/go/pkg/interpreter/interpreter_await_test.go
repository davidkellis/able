package interpreter

import (
	"strings"
	"testing"
	"time"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestAwaitExpressionManualWaker(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	mustEvalStmt := func(stmt ast.Statement) {
		if _, err := interp.evaluateStatement(stmt, global); err != nil {
			t.Fatalf("statement evaluation failed: %v", err)
		}
	}
	mustEvalExpr := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, global)
		if err != nil {
			t.Fatalf("expression evaluation failed: %v", err)
		}
		return val
	}

	mustEvalExpr(ast.Assign(ast.ID("last_waker"), ast.Nil()))

	manualAwaitableDef := ast.StructDef("ManualAwaitable", []*ast.StructFieldDefinition{
		ast.FieldDef(ast.Ty("bool"), "ready"),
		ast.FieldDef(ast.Ty("i32"), "value"),
		ast.FieldDef(ast.Nullable(ast.Ty("AwaitWaker")), "waker"),
	}, ast.StructKindNamed, nil, nil, false)
	mustEvalStmt(manualAwaitableDef)

	manualRegistrationDef := ast.StructDef("ManualRegistration", nil, ast.StructKindNamed, nil, nil, false)
	mustEvalStmt(manualRegistrationDef)

	isReadyFn := ast.Fn(
		"is_ready",
		nil,
		[]ast.Statement{
			ast.Ret(ast.ImplicitMember(ast.ID("ready"))),
		},
		ast.Ty("bool"),
		nil,
		nil,
		true,
		false,
	)

	registerFn := ast.Fn(
		"register",
		[]*ast.FunctionParameter{
			ast.Param("waker", ast.Ty("AwaitWaker")),
		},
		[]ast.Statement{
			ast.AssignOp(ast.AssignmentAssign, ast.ImplicitMember(ast.ID("waker")), ast.ID("waker")),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("last_waker"), ast.ID("waker")),
			ast.Ret(ast.StructLit(nil, false, "ManualRegistration", nil, nil)),
		},
		ast.Ty("ManualRegistration"),
		nil,
		nil,
		true,
		false,
	)

	commitFn := ast.Fn(
		"commit",
		nil,
		[]ast.Statement{
			ast.Ret(ast.ImplicitMember(ast.ID("value"))),
		},
		ast.Ty("i32"),
		nil,
		nil,
		true,
		false,
	)

	mustEvalStmt(ast.Methods(ast.Ty("ManualAwaitable"), []*ast.FunctionDefinition{
		isReadyFn,
		registerFn,
		commitFn,
	}, nil, nil))

	cancelFn := ast.Fn(
		"cancel",
		nil,
		[]ast.Statement{
			ast.AssignOp(ast.AssignmentAssign, ast.ImplicitMember(ast.ID("waker")), ast.Nil()),
			ast.Ret(ast.Nil()),
		},
		ast.Ty("void"),
		nil,
		nil,
		true,
		false,
	)
	mustEvalStmt(ast.Methods(ast.Ty("ManualRegistration"), []*ast.FunctionDefinition{cancelFn}, nil, nil))

	manualArm := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Bool(false), "ready"),
			ast.FieldInit(ast.Int(42), "value"),
			ast.FieldInit(ast.Nil(), "waker"),
		},
		false,
		"ManualAwaitable",
		nil,
		nil,
	)
	mustEvalExpr(ast.Assign(ast.ID("arm"), manualArm))
	mustEvalExpr(ast.Assign(ast.ID("result"), ast.Int(0)))

	awaitFuture := ast.Spawn(ast.Block(
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("result"),
			ast.Await(ast.Arr(ast.ID("arm"))),
		),
	))
	mustEvalExpr(ast.Assign(ast.ID("handle"), awaitFuture))

	// Drive the proc to the await point.
	mustEvalExpr(ast.Call("future_flush"))

	handleVal, ok := mustEvalExpr(ast.ID("handle")).(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected handle to be a proc handle")
	}
	if status := futureStatus(handleVal); status != runtime.FuturePending {
		t.Fatalf("expected pending handle after await, got %v (value=%#v)", status, interp.futureValue(handleVal))
	}

	if got := intFromValue(t, mustEvalExpr(ast.ID("result"))); got != 0 {
		t.Fatalf("expected initial result to remain 0, got %d", got)
	}

	wakerVal := mustEvalExpr(ast.ID("last_waker"))
	if _, ok := wakerVal.(*runtime.StructInstanceValue); !ok {
		var armWaker runtime.Value
		if armVal, ok := mustEvalExpr(ast.ID("arm")).(*runtime.StructInstanceValue); ok && armVal != nil {
			armWaker = armVal.Fields["waker"]
		}
		t.Fatalf("expected last_waker to be struct instance, got %#v (status=%v armWaker=%#v)", wakerVal, futureStatus(handleVal), armWaker)
	}

	mustEvalExpr(ast.AssignMember(ast.ID("arm"), "ready", ast.Bool(true)))
	mustEvalExpr(ast.CallExpr(ast.Member(ast.ID("last_waker"), "wake")))
	mustEvalExpr(ast.Call("future_flush"))

	if got := intFromValue(t, mustEvalExpr(ast.ID("result"))); got != 42 {
		t.Fatalf("expected await to produce 42, got %d", got)
	}
}

func TestAwaitExpressionChannelReceive(t *testing.T) {
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
	mustEval(ast.Assign(ast.ID("result"), ast.Str("pending")))

	recvArm := ast.Call(
		"__able_channel_await_try_recv",
		ast.ID("ch"),
		ast.Lam([]*ast.FunctionParameter{ast.Param("v", nil)}, ast.ID("v")),
	)

	consumer := mustEval(ast.Spawn(ast.Block(
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("result"),
			ast.Await(ast.Arr(recvArm)),
		),
	)))
	consumerHandle, ok := consumer.(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected consumer proc handle, got %#v", consumer)
	}

	producer := mustEval(ast.Spawn(ast.Block(
		ast.Call("future_yield"),
		ast.Call("__able_channel_send", ast.ID("ch"), ast.Str("ping")),
		ast.Nil(),
	)))
	producerHandle, ok := producer.(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected producer proc handle, got %#v", producer)
	}

	if !waitForStatus(consumerHandle, runtime.FutureResolved, 500*time.Millisecond) {
		t.Fatalf("consumer did not resolve: %v", futureStatus(consumerHandle))
	}
	if !waitForStatus(producerHandle, runtime.FutureResolved, 500*time.Millisecond) {
		t.Fatalf("producer did not resolve: %v", futureStatus(producerHandle))
	}

	if got := mustGetString(t, global, "result"); got != "ping" {
		t.Fatalf("expected await result to be ping, got %q", got)
	}
}

func TestAwaitChannelArmsSerialExecutor(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	mustEval := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, global)
		if err != nil {
			t.Fatalf("evaluation failed: %v", err)
		}
		return val
	}

	mustEval(ast.Assign(ast.ID("ch"), ast.Call("__able_channel_new", ast.Int(0))))
	mustEval(ast.Assign(ast.ID("receiver_result"), ast.Str("pending")))
	mustEval(ast.Assign(ast.ID("sender_result"), ast.Str("pending")))

	recvArm := ast.Call(
		"__able_channel_await_try_recv",
		ast.ID("ch"),
		ast.Lam([]*ast.FunctionParameter{ast.Param("v", nil)}, ast.ID("v")),
	)
	receiver := mustEval(ast.Spawn(ast.Block(
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("receiver_result"),
			ast.Await(ast.Arr(recvArm)),
		),
	)))
	receiverHandle, ok := receiver.(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected receiver proc handle, got %#v", receiver)
	}

	sendArm := ast.Call(
		"__able_channel_await_try_send",
		ast.ID("ch"),
		ast.Str("payload"),
		ast.Lam(nil, ast.Str("sent")),
	)
	sender := mustEval(ast.Spawn(ast.Block(
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("sender_result"),
			ast.Await(ast.Arr(sendArm)),
		),
	)))
	senderHandle, ok := sender.(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected sender proc handle, got %#v", sender)
	}

	if !waitForStatus(receiverHandle, runtime.FutureResolved, 500*time.Millisecond) {
		t.Fatalf("receiver did not resolve, status=%v", futureStatus(receiverHandle))
	}
	if !waitForStatus(senderHandle, runtime.FutureResolved, 500*time.Millisecond) {
		t.Fatalf("sender did not resolve, status=%v", futureStatus(senderHandle))
	}

	if got := mustGetString(t, global, "receiver_result"); got != "payload" {
		t.Fatalf("expected receiver_result to be payload, got %q", got)
	}
	if got := mustGetString(t, global, "sender_result"); got != "sent" {
		t.Fatalf("expected sender_result to be sent, got %q", got)
	}
}

func TestAwaitFutureHandle(t *testing.T) {
	interp := newAsyncInterpreter(t)
	global := interp.GlobalEnvironment()

	mustEval := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, global)
		if err != nil {
			t.Fatalf("evaluation failed: %v", err)
		}
		return val
	}

	mustEval(ast.Assign(ast.ID("result"), ast.Int(0)))
	mustEval(ast.Assign(ast.ID("fut"), ast.Spawn(ast.Block(
		ast.Call("future_yield"),
		ast.Int(21),
	))))

	consumer := mustEval(ast.Spawn(ast.Block(
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("result"),
			ast.Await(ast.Arr(ast.ID("fut"))),
		),
	)))
	consumerHandle, ok := consumer.(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected consumer proc handle, got %#v", consumer)
	}

	if !waitForStatus(consumerHandle, runtime.FutureResolved, 500*time.Millisecond) {
		t.Fatalf("consumer did not resolve: %v", futureStatus(consumerHandle))
	}

	if got := intFromValue(t, mustEval(ast.ID("result"))); got != 21 {
		t.Fatalf("expected await result to be 21, got %d", got)
	}
}

func TestAwaitFutureHandleResolvesWorker(t *testing.T) {
	interp := newAsyncInterpreter(t)
	global := interp.GlobalEnvironment()

	mustEval := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, global)
		if err != nil {
			t.Fatalf("evaluation failed: %v", err)
		}
		return val
	}

	mustEval(ast.Assign(ast.ID("result"), ast.Int(0)))
	mustEval(ast.Assign(ast.ID("worker"), ast.Spawn(ast.Block(
		ast.Call("future_yield"),
		ast.Int(12),
	))))

	consumer := mustEval(ast.Spawn(ast.Block(
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("result"),
			ast.Await(ast.Arr(ast.ID("worker"))),
		),
	)))
	consumerHandle, ok := consumer.(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected consumer proc handle, got %#v", consumer)
	}

	workerVal := mustEval(ast.ID("worker"))
	workerHandle, ok := workerVal.(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected worker proc handle, got %#v", workerVal)
	}

	if !waitForStatus(workerHandle, runtime.FutureResolved, 500*time.Millisecond) {
		t.Fatalf("worker did not resolve: %v", futureStatus(workerHandle))
	}
	if !waitForStatus(consumerHandle, runtime.FutureResolved, 500*time.Millisecond) {
		t.Fatalf("consumer did not resolve: %v", futureStatus(consumerHandle))
	}

	if got := intFromValue(t, mustEval(ast.ID("result"))); got != 12 {
		t.Fatalf("expected await result to be 12, got %d", got)
	}
}

func TestAwaitDefaultHelper(t *testing.T) {
	interp := newAsyncInterpreter(t)
	global := interp.GlobalEnvironment()

	mustEval := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, global)
		if err != nil {
			t.Fatalf("evaluation failed: %v", err)
		}
		return val
	}

	mustEval(ast.Assign(ast.ID("result"), ast.Str("pending")))
	consumer := mustEval(ast.Spawn(ast.Block(
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("result"),
			ast.Await(ast.Arr(ast.Call("__able_await_default", ast.Lam(nil, ast.Str("fallback"))))),
		),
	)))
	handle, ok := consumer.(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected consumer proc handle, got %#v", consumer)
	}

	if !waitForStatus(handle, runtime.FutureResolved, 500*time.Millisecond) {
		t.Fatalf("consumer did not resolve: %v", futureStatus(handle))
	}
	if got := mustGetString(t, global, "result"); got != "fallback" {
		t.Fatalf("expected await result to be fallback, got %q", got)
	}
}

func TestAwaitSleepMsHelper(t *testing.T) {
	interp := newAsyncInterpreter(t)
	global := interp.GlobalEnvironment()

	mustEval := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, global)
		if err != nil {
			t.Fatalf("evaluation failed: %v", err)
		}
		return val
	}

	mustEval(ast.Assign(ast.ID("result"), ast.Str("pending")))
	timerArm := ast.Call("__able_await_sleep_ms", ast.Int(5), ast.Lam(nil, ast.Str("timer")))
	consumer := mustEval(ast.Spawn(ast.Block(
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("result"),
			ast.Await(ast.Arr(timerArm)),
		),
	)))
	handle, ok := consumer.(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected consumer proc handle, got %#v", consumer)
	}

	if !waitForStatus(handle, runtime.FutureResolved, 500*time.Millisecond) {
		t.Fatalf("consumer did not resolve: %v", futureStatus(handle))
	}
	if got := mustGetString(t, global, "result"); got != "timer" {
		t.Fatalf("expected await result to be timer, got %q", got)
	}
}

func TestAwaitReadyArmsRoundRobin(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	mustEvalStmt := func(stmt ast.Statement) {
		if _, err := interp.evaluateStatement(stmt, global); err != nil {
			t.Fatalf("statement evaluation failed: %v", err)
		}
	}
	mustEvalExpr := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, global)
		if err != nil {
			t.Fatalf("expression evaluation failed: %v", err)
		}
		return val
	}

	mustEvalStmt(ast.StructDef("ManualRegistration", nil, ast.StructKindNamed, nil, nil, false))
	cancelFn := ast.Fn(
		"cancel",
		nil,
		[]ast.Statement{
			ast.Ret(ast.Nil()),
		},
		ast.Ty("void"),
		nil,
		nil,
		true,
		false,
	)
	mustEvalStmt(ast.Methods(ast.Ty("ManualRegistration"), []*ast.FunctionDefinition{cancelFn}, nil, nil))

	mustEvalStmt(ast.StructDef("ReadyAwaitable", []*ast.StructFieldDefinition{
		ast.FieldDef(ast.Ty("i32"), "value"),
	}, ast.StructKindNamed, nil, nil, false))

	readyMethods := ast.Methods(ast.Ty("ReadyAwaitable"), []*ast.FunctionDefinition{
		ast.Fn(
			"is_ready",
			nil,
			[]ast.Statement{
				ast.Ret(ast.Bool(true)),
			},
			ast.Ty("bool"),
			nil,
			nil,
			true,
			false,
		),
		ast.Fn(
			"register",
			[]*ast.FunctionParameter{ast.Param("waker", ast.Ty("AwaitWaker"))},
			[]ast.Statement{
				ast.Ret(ast.StructLit(nil, false, "ManualRegistration", nil, nil)),
			},
			ast.Ty("ManualRegistration"),
			nil,
			nil,
			true,
			false,
		),
		ast.Fn(
			"commit",
			nil,
			[]ast.Statement{
				ast.Ret(ast.ImplicitMember(ast.ID("value"))),
			},
			ast.Ty("i32"),
			nil,
			nil,
			true,
			false,
		),
		ast.Fn(
			"is_default",
			nil,
			[]ast.Statement{
				ast.Ret(ast.Bool(false)),
			},
			ast.Ty("bool"),
			nil,
			nil,
			true,
			false,
		),
	}, nil, nil)
	mustEvalStmt(readyMethods)

	mustEvalExpr(ast.Assign(ast.ID("arm1"), ast.StructLit([]*ast.StructFieldInitializer{
		ast.FieldInit(ast.Int(1), "value"),
	}, false, "ReadyAwaitable", nil, nil)))
	mustEvalExpr(ast.Assign(ast.ID("arm2"), ast.StructLit([]*ast.StructFieldInitializer{
		ast.FieldInit(ast.Int(2), "value"),
	}, false, "ReadyAwaitable", nil, nil)))
	mustEvalExpr(ast.Assign(ast.ID("arms"), ast.Arr(ast.ID("arm1"), ast.ID("arm2"))))

	mustEvalExpr(ast.Assign(ast.ID("first"), ast.Int(0)))
	mustEvalExpr(ast.Assign(ast.ID("second"), ast.Int(0)))
	mustEvalExpr(ast.Assign(ast.ID("third"), ast.Int(0)))

	procHandleVal := mustEvalExpr(ast.Spawn(ast.Block(
		ast.AssignOp(ast.AssignmentAssign, ast.ID("first"), ast.Await(ast.ID("arms"))),
		ast.AssignOp(ast.AssignmentAssign, ast.ID("second"), ast.Await(ast.ID("arms"))),
		ast.AssignOp(ast.AssignmentAssign, ast.ID("third"), ast.Await(ast.ID("arms"))),
	)))
	handle, ok := procHandleVal.(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected proc handle, got %#v", procHandleVal)
	}

	if !waitForStatus(handle, runtime.FutureResolved, 200*time.Millisecond) {
		t.Fatalf("proc did not resolve, status=%v", futureStatus(handle))
	}

	if got := intFromValue(t, mustEvalExpr(ast.ID("first"))); got != 1 {
		t.Fatalf("expected first await to pick arm1, got %d", got)
	}
	if got := intFromValue(t, mustEvalExpr(ast.ID("second"))); got != 2 {
		t.Fatalf("expected second await to pick arm2, got %d", got)
	}
	if got := intFromValue(t, mustEvalExpr(ast.ID("third"))); got != 1 {
		t.Fatalf("expected third await to rotate back to arm1, got %d", got)
	}
}

func TestAwaitCancellationStopsPendingArms(t *testing.T) {
	interp := newAsyncInterpreter(t)
	global := interp.GlobalEnvironment()

	mustEval := func(expr ast.Expression) runtime.Value {
		val, err := interp.evaluateExpression(expr, global)
		if err != nil {
			t.Fatalf("evaluation failed: %v", err)
		}
		return val
	}

	mustEval(ast.Assign(ast.ID("hits"), ast.Int(0)))
	mustEval(ast.Assign(ast.ID("result"), ast.Str("pending")))

	timerArm := ast.Call(
		"__able_await_sleep_ms",
		ast.Int(15),
		ast.LamBlock(nil, ast.Block(
			ast.AssignOp(ast.AssignmentAssign, ast.ID("hits"), ast.Bin("+", ast.ID("hits"), ast.Int(1))),
			ast.Ret(ast.Str("timer")),
		)),
	)

	consumer := mustEval(ast.Spawn(ast.Block(
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("result"),
			ast.Await(ast.Arr(timerArm)),
		),
	)))
	handle, ok := consumer.(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected proc handle, got %#v", consumer)
	}

	time.Sleep(5 * time.Millisecond)
	if handle != nil {
		handle.RequestCancel()
	}

	if !waitForStatus(handle, runtime.FutureCancelled, 200*time.Millisecond) {
		t.Fatalf("expected cancelled status, got %v", futureStatus(handle))
	}

	time.Sleep(25 * time.Millisecond)

	if got := intFromValue(t, mustEval(ast.ID("hits"))); got != 0 {
		t.Fatalf("expected timer callback to be cancelled, hits=%d", got)
	}
	if got := mustGetString(t, global, "result"); got != "pending" {
		t.Fatalf("expected result to remain pending, got %q", got)
	}

	valueVal := interp.futureValue(handle)
	errValue, ok := valueVal.(runtime.ErrorValue)
	if !ok {
		t.Fatalf("expected runtime error from cancelled proc, got %#v", valueVal)
	}
	if !strings.Contains(errValue.Message, "cancelled") {
		t.Fatalf("expected cancellation message, got %q", errValue.Message)
	}
}

func TestAwaitChannelSendCancellation(t *testing.T) {
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
	mustEval(ast.Assign(ast.ID("hits"), ast.Int(0)))
	mustEval(ast.Assign(ast.ID("result"), ast.Str("pending")))

	sendArm := ast.Call(
		"__able_channel_await_try_send",
		ast.ID("ch"),
		ast.Str("payload"),
		ast.LamBlock(nil, ast.Block(
			ast.AssignOp(ast.AssignmentAssign, ast.ID("hits"), ast.Bin("+", ast.ID("hits"), ast.Int(1))),
			ast.Ret(ast.Str("sent")),
		)),
	)

	consumer := mustEval(ast.Spawn(ast.Block(
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("result"),
			ast.Await(ast.Arr(sendArm)),
		),
	)))
	handle, ok := consumer.(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected proc handle, got %#v", consumer)
	}

	if handle != nil {
		handle.RequestCancel()
	}
	if !waitForStatus(handle, runtime.FutureCancelled, 200*time.Millisecond) {
		t.Fatalf("expected cancelled status, got %v", futureStatus(handle))
	}

	recvVal := mustEval(ast.Call("__able_channel_try_receive", ast.ID("ch")))
	if _, ok := recvVal.(runtime.NilValue); !ok {
		t.Fatalf("expected try_receive to return nil after cancellation, got %#v", recvVal)
	}
	if got := intFromValue(t, mustEval(ast.ID("hits"))); got != 0 {
		t.Fatalf("expected hits to remain 0, got %d", got)
	}
	if got := mustGetString(t, global, "result"); got != "pending" {
		t.Fatalf("expected result to remain pending, got %q", got)
	}
	valueVal := interp.futureValue(handle)
	errValue, ok := valueVal.(runtime.ErrorValue)
	if !ok {
		t.Fatalf("expected runtime error from cancelled proc, got %#v", valueVal)
	}
	if !strings.Contains(errValue.Message, "cancelled") {
		t.Fatalf("expected cancellation message, got %q", errValue.Message)
	}
}

func TestAwaitChannelCancellationStopsRegistrations(t *testing.T) {
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
	mustEval(ast.Assign(ast.ID("hits"), ast.Int(0)))
	mustEval(ast.Assign(ast.ID("result"), ast.Str("pending")))

	recvArm := ast.Call(
		"__able_channel_await_try_recv",
		ast.ID("ch"),
		ast.Lam([]*ast.FunctionParameter{ast.Param("v", nil)}, ast.Block(
			ast.AssignOp(ast.AssignmentAssign, ast.ID("hits"), ast.Bin("+", ast.ID("hits"), ast.Int(1))),
			ast.Ret(ast.ID("v")),
		)),
	)

	consumer := mustEval(ast.Spawn(ast.Block(
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("result"),
			ast.Await(ast.Arr(recvArm)),
		),
	)))
	handle, ok := consumer.(*runtime.FutureValue)
	if !ok {
		t.Fatalf("expected proc handle, got %#v", consumer)
	}

	if futureStatus(handle) != runtime.FuturePending {
		t.Fatalf("expected pending handle before cancellation, got %v", futureStatus(handle))
	}

	if handle != nil {
		handle.RequestCancel()
	}
	if !waitForStatus(handle, runtime.FutureCancelled, 200*time.Millisecond) {
		t.Fatalf("expected cancelled status, got %v", futureStatus(handle))
	}

	// Send after cancellation to ensure the await registration was cancelled and the callback is not invoked.
	mustEval(ast.Call("__able_channel_send", ast.ID("ch"), ast.Str("payload")))
	time.Sleep(10 * time.Millisecond)

	if got := intFromValue(t, mustEval(ast.ID("hits"))); got != 0 {
		t.Fatalf("expected hits to remain 0, got %d", got)
	}
	if got := mustGetString(t, global, "result"); got != "pending" {
		t.Fatalf("expected result to remain pending, got %q", got)
	}

	valueVal := interp.futureValue(handle)
	errValue, ok := valueVal.(runtime.ErrorValue)
	if !ok {
		t.Fatalf("expected runtime error from cancelled proc, got %#v", valueVal)
	}
	if !strings.Contains(errValue.Message, "cancelled") {
		t.Fatalf("expected cancellation message, got %q", errValue.Message)
	}
}
