package interpreter

import (
	"testing"
	"time"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
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

	awaitProc := ast.Proc(ast.Block(
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("result"),
			ast.Await(ast.Arr(ast.ID("arm"))),
		),
	))
	mustEvalExpr(ast.Assign(ast.ID("handle"), awaitProc))

	// Drive the proc to the await point.
	mustEvalExpr(ast.Call("proc_flush"))

	handleVal, ok := mustEvalExpr(ast.ID("handle")).(*runtime.ProcHandleValue)
	if !ok {
		t.Fatalf("expected handle to be a proc handle")
	}
	if status := handleVal.Status(); status != runtime.ProcPending {
		t.Fatalf("expected pending handle after await, got %v (value=%#v)", status, interp.procHandleValue(handleVal))
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
		t.Fatalf("expected last_waker to be struct instance, got %#v (status=%v armWaker=%#v)", wakerVal, handleVal.Status(), armWaker)
	}

	mustEvalExpr(ast.AssignMember(ast.ID("arm"), "ready", ast.Bool(true)))
	mustEvalExpr(ast.CallExpr(ast.Member(ast.ID("last_waker"), "wake")))
	mustEvalExpr(ast.Call("proc_flush"))

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

	consumer := mustEval(ast.Proc(ast.Block(
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("result"),
			ast.Await(ast.Arr(recvArm)),
		),
	)))
	consumerHandle, ok := consumer.(*runtime.ProcHandleValue)
	if !ok {
		t.Fatalf("expected consumer proc handle, got %#v", consumer)
	}

	producer := mustEval(ast.Proc(ast.Block(
		ast.Call("proc_yield"),
		ast.Call("__able_channel_send", ast.ID("ch"), ast.Str("ping")),
		ast.Nil(),
	)))
	producerHandle, ok := producer.(*runtime.ProcHandleValue)
	if !ok {
		t.Fatalf("expected producer proc handle, got %#v", producer)
	}

	if !waitForStatus(consumerHandle, runtime.ProcResolved, 500*time.Millisecond) {
		t.Fatalf("consumer did not resolve: %v", consumerHandle.Status())
	}
	if !waitForStatus(producerHandle, runtime.ProcResolved, 500*time.Millisecond) {
		t.Fatalf("producer did not resolve: %v", producerHandle.Status())
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
	receiver := mustEval(ast.Proc(ast.Block(
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("receiver_result"),
			ast.Await(ast.Arr(recvArm)),
		),
	)))
	receiverHandle, ok := receiver.(*runtime.ProcHandleValue)
	if !ok {
		t.Fatalf("expected receiver proc handle, got %#v", receiver)
	}

	sendArm := ast.Call(
		"__able_channel_await_try_send",
		ast.ID("ch"),
		ast.Str("payload"),
		ast.Lam(nil, ast.Str("sent")),
	)
	sender := mustEval(ast.Proc(ast.Block(
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("sender_result"),
			ast.Await(ast.Arr(sendArm)),
		),
	)))
	senderHandle, ok := sender.(*runtime.ProcHandleValue)
	if !ok {
		t.Fatalf("expected sender proc handle, got %#v", sender)
	}

	if !waitForStatus(receiverHandle, runtime.ProcResolved, 500*time.Millisecond) {
		t.Fatalf("receiver did not resolve, status=%v", receiverHandle.Status())
	}
	if !waitForStatus(senderHandle, runtime.ProcResolved, 500*time.Millisecond) {
		t.Fatalf("sender did not resolve, status=%v", senderHandle.Status())
	}

	if got := mustGetString(t, global, "receiver_result"); got != "payload" {
		t.Fatalf("expected receiver_result to be payload, got %q", got)
	}
	if got := mustGetString(t, global, "sender_result"); got != "sent" {
		t.Fatalf("expected sender_result to be sent, got %q", got)
	}
}
