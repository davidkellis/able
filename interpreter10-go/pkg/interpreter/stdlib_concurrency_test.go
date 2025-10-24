package interpreter

import (
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestStdlibChannelMutexPreludeSmoke(t *testing.T) {
	interp := New()
	interp.EnableTypechecker(TypecheckConfig{FailFast: true})

	channelStruct := ast.StructDef(
		"Channel",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "capacity"),
			ast.FieldDef(ast.Ty("i64"), "handle"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)

	channelNew := ast.Fn(
		"new",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("handle"), ast.Call("__able_channel_new", ast.Int(0))),
			ast.Ret(ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(0), "capacity"),
					ast.FieldInit(ast.ID("handle"), "handle"),
				},
				false,
				"Channel",
				nil,
				nil,
			)),
		},
		ast.Ty("Channel"),
		nil,
		nil,
		false,
		false,
	)

	channelMethods := ast.Methods(
		ast.Ty("Channel"),
		[]*ast.FunctionDefinition{channelNew},
		nil,
		nil,
	)

	channelHandleFn := ast.Fn(
		"channel_handle",
		[]*ast.FunctionParameter{ast.Param("capacity", ast.Ty("i32"))},
		[]ast.Statement{
			ast.Ret(ast.Call("__able_channel_new", ast.ID("capacity"))),
		},
		ast.Ty("i64"),
		nil,
		nil,
		false,
		false,
	)

	mutexStruct := ast.StructDef(
		"Mutex",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i64"), "handle"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)

	mutexNew := ast.Fn(
		"new",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("handle"), ast.Call("__able_mutex_new")),
			ast.Ret(ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.ID("handle"), "handle"),
				},
				false,
				"Mutex",
				nil,
				nil,
			)),
		},
		ast.Ty("Mutex"),
		nil,
		nil,
		false,
		false,
	)

	mutexMethods := ast.Methods(
		ast.Ty("Mutex"),
		[]*ast.FunctionDefinition{mutexNew},
		nil,
		nil,
	)

	mutexHandleFn := ast.Fn(
		"mutex_handle",
		nil,
		[]ast.Statement{
			ast.Ret(ast.Call("__able_mutex_new")),
		},
		ast.Ty("i64"),
		nil,
		nil,
		false,
		false,
	)

	module := ast.Mod(
		[]ast.Statement{
			channelStruct,
			channelMethods,
			channelHandleFn,
			mutexStruct,
			mutexMethods,
			mutexHandleFn,
			ast.Assign(
				ast.ID("channel_instance"),
				ast.CallExpr(
					ast.Member(ast.ID("Channel"), ast.ID("new")),
				),
			),
			ast.Assign(
				ast.ID("channel_handle_value"),
				ast.Call("channel_handle", ast.Int(0)),
			),
			ast.Assign(
				ast.ID("mutex_instance"),
				ast.CallExpr(
					ast.Member(ast.ID("Mutex"), ast.ID("new")),
				),
			),
			ast.Assign(
				ast.ID("mutex_handle_value"),
				ast.Call("mutex_handle"),
			),
			ast.ID("mutex_instance"),
		},
		nil,
		nil,
	)

	value, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("module evaluation failed: %v", err)
	}
	if value == nil {
		t.Fatalf("expected non-nil final value from module")
	}
	if diags := interp.TypecheckDiagnostics(); len(diags) != 0 {
		t.Fatalf("unexpected typecheck diagnostics: %v", diags)
	}

	scope := env
	if scope == nil {
		scope = interp.GlobalEnvironment()
	}

	chanVal, err := scope.Get("channel_instance")
	if err != nil {
		t.Fatalf("missing channel_instance binding: %v", err)
	}
	chanStruct, ok := chanVal.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("channel_instance type = %T, want *runtime.StructInstanceValue", chanVal)
	}

	capacityVal, ok := chanStruct.Fields["capacity"]
	if !ok {
		t.Fatalf("channel capacity field missing")
	}
	capacityInt, ok := capacityVal.(runtime.IntegerValue)
	if !ok {
		if ptr, ok := capacityVal.(*runtime.IntegerValue); ok && ptr != nil {
			capacityInt = *ptr
		} else {
			t.Fatalf("channel capacity type = %T, want runtime.IntegerValue", capacityVal)
		}
	}
	if capacityInt.Val == nil || capacityInt.Val.Sign() != 0 {
		t.Fatalf("channel capacity = %v, want 0", capacityInt.Val)
	}

	handleVal, ok := chanStruct.Fields["handle"]
	if !ok {
		t.Fatalf("channel handle field missing")
	}
	handleInt, ok := handleVal.(runtime.IntegerValue)
	if !ok {
		if ptr, ok := handleVal.(*runtime.IntegerValue); ok && ptr != nil {
			handleInt = *ptr
		} else {
			t.Fatalf("channel handle type = %T, want runtime.IntegerValue", handleVal)
		}
	}
	if handleInt.Val == nil || handleInt.Val.Sign() <= 0 {
		t.Fatalf("channel handle not positive: %v", handleInt.Val)
	}
	if handleInt.TypeSuffix != runtime.IntegerI64 {
		t.Fatalf("channel handle suffix = %q, want %q", handleInt.TypeSuffix, runtime.IntegerI64)
	}

	mutexVal, err := scope.Get("mutex_instance")
	if err != nil {
		t.Fatalf("missing mutex_instance binding: %v", err)
	}
	mutexStructVal, ok := mutexVal.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("mutex_instance type = %T, want *runtime.StructInstanceValue", mutexVal)
	}

	mutexHandle, ok := mutexStructVal.Fields["handle"]
	if !ok {
		t.Fatalf("mutex handle field missing")
	}
	mutexHandleInt, ok := mutexHandle.(runtime.IntegerValue)
	if !ok {
		if ptr, ok := mutexHandle.(*runtime.IntegerValue); ok && ptr != nil {
			mutexHandleInt = *ptr
		} else {
			t.Fatalf("mutex handle type = %T, want runtime.IntegerValue", mutexHandle)
		}
	}
	if mutexHandleInt.Val == nil || mutexHandleInt.Val.Sign() <= 0 {
		t.Fatalf("mutex handle not positive: %v", mutexHandleInt.Val)
	}
	if mutexHandleInt.TypeSuffix != runtime.IntegerI64 {
		t.Fatalf("mutex handle suffix = %q, want %q", mutexHandleInt.TypeSuffix, runtime.IntegerI64)
	}

	callHandle, err := interp.evaluateExpression(ast.Call("channel_handle", ast.Int(5)), scope)
	if err != nil {
		t.Fatalf("channel_handle call failed: %v", err)
	}
	callHandleInt, ok := callHandle.(runtime.IntegerValue)
	if !ok {
		if ptr, ok := callHandle.(*runtime.IntegerValue); ok && ptr != nil {
			callHandleInt = *ptr
		} else {
			t.Fatalf("channel_handle returned %T, want runtime.IntegerValue", callHandle)
		}
	}
	if callHandleInt.Val == nil || callHandleInt.Val.Sign() <= 0 {
		t.Fatalf("channel_handle result not positive: %v", callHandleInt.Val)
	}

	mutexCall, err := interp.evaluateExpression(ast.Call("mutex_handle"), scope)
	if err != nil {
		t.Fatalf("mutex_handle call failed: %v", err)
	}
	mutexCallInt, ok := mutexCall.(runtime.IntegerValue)
	if !ok {
		if ptr, ok := mutexCall.(*runtime.IntegerValue); ok && ptr != nil {
			mutexCallInt = *ptr
		} else {
			t.Fatalf("mutex_handle returned %T, want runtime.IntegerValue", mutexCall)
		}
	}
	if mutexCallInt.Val == nil || mutexCallInt.Val.Sign() <= 0 {
		t.Fatalf("mutex_handle result not positive: %v", mutexCallInt.Val)
	}

	if _, err := scope.Get("channel_handle_value"); err != nil {
		t.Fatalf("channel_handle_value missing: %v", err)
	}
	if _, err := scope.Get("mutex_handle_value"); err != nil {
		t.Fatalf("mutex_handle_value missing: %v", err)
	}

	cond1 := ast.Bin(
		"!=",
		ast.Member(ast.ID("channel_instance"), ast.ID("handle")),
		ast.Int(0),
	)
	cond1Val, err := interp.evaluateExpression(cond1, scope)
	if err != nil {
		t.Fatalf("cond1 evaluation failed: %v", err)
	}
	if cBool, ok := cond1Val.(runtime.BoolValue); !ok || !cBool.Val {
		t.Fatalf("cond1 expected true, got %#v", cond1Val)
	}

	scoreExpr := ast.Block(
		ast.Assign(ast.ID("score"), ast.Int(0)),
		ast.Iff(
			ast.Bin(
				"!=",
				ast.Member(ast.ID("channel_instance"), ast.ID("handle")),
				ast.Int(0),
			),
			ast.Block(
				ast.AssignOp(
					ast.AssignmentAdd,
					ast.ID("score"),
					ast.Int(1),
				),
			),
		),
		ast.Iff(
			ast.Bin(
				"==",
				ast.Member(ast.ID("channel_instance"), ast.ID("capacity")),
				ast.Int(0),
			),
			ast.Block(
				ast.AssignOp(
					ast.AssignmentAdd,
					ast.ID("score"),
					ast.Int(1),
				),
			),
		),
		ast.Iff(
			ast.Bin("!=", ast.ID("channel_handle_value"), ast.Int(0)),
			ast.Block(
				ast.AssignOp(
					ast.AssignmentAdd,
					ast.ID("score"),
					ast.Int(1),
				),
			),
		),
		ast.Iff(
			ast.Bin(
				"!=",
				ast.Member(ast.ID("mutex_instance"), ast.ID("handle")),
				ast.Int(0),
			),
			ast.Block(
				ast.AssignOp(
					ast.AssignmentAdd,
					ast.ID("score"),
					ast.Int(1),
				),
			),
		),
		ast.Iff(
			ast.Bin("!=", ast.ID("mutex_handle_value"), ast.Int(0)),
			ast.Block(
				ast.AssignOp(
					ast.AssignmentAdd,
					ast.ID("score"),
					ast.Int(1),
				),
			),
		),
		ast.Bin("==", ast.ID("score"), ast.Int(5)),
	)

	scoreValue, err := interp.evaluateExpression(scoreExpr, scope)
	if err != nil {
		t.Fatalf("score expression failed: %v", err)
	}
	scoreBool, ok := scoreValue.(runtime.BoolValue)
	if !ok || !scoreBool.Val {
		t.Fatalf("score expression expected true, got %#v", scoreValue)
	}
}
