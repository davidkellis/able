package compiler

import (
	"bytes"
	"fmt"
)

func (g *generator) renderRuntimeChannelAwaitableSurface(buf *bytes.Buffer) {
	fmt.Fprintf(buf, "type __able_channel_awaitable struct {\n")
	fmt.Fprintf(buf, "\thandle       int64\n")
	fmt.Fprintf(buf, "\top           __able_channel_await_kind\n")
	fmt.Fprintf(buf, "\tpayload      runtime.Value\n")
	fmt.Fprintf(buf, "\tcallback     runtime.Value\n")
	fmt.Fprintf(buf, "\tregistration *__able_channel_await_registration\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func (a *__able_channel_awaitable) Kind() runtime.Kind { return runtime.KindStructInstance }\n\n")
	fmt.Fprintf(buf, "func (a *__able_channel_awaitable) NativeAwaitableIsReady(_ *runtime.NativeCallContext) (bool, error) {\n")
	fmt.Fprintf(buf, "\treturn a.isReady()\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func (a *__able_channel_awaitable) NativeAwaitableRegister(ctx *runtime.NativeCallContext, waker runtime.Value) (runtime.Value, error) {\n")
	if g.callableExecutionContextsEnabled() {
		fmt.Fprintf(buf, "\treturn a.register(waker, ctx)\n")
	} else {
		fmt.Fprintf(buf, "\treturn a.register(waker)\n")
	}
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func (a *__able_channel_awaitable) NativeAwaitableCommit(ctx *runtime.NativeCallContext) (runtime.Value, error) {\n")
	if g.callableExecutionContextsEnabled() {
		fmt.Fprintf(buf, "\treturn a.commit(ctx)\n")
	} else {
		fmt.Fprintf(buf, "\treturn a.commit()\n")
	}
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func (a *__able_channel_awaitable) NativeAwaitableIsDefault() bool { return false }\n\n")
	fmt.Fprintf(buf, "func (a *__able_channel_awaitable) MaterializeRuntimeValue() runtime.Value { return a.toStruct() }\n\n")
	fmt.Fprintf(buf, "func (a *__able_channel_awaitable) isReady() (bool, error) {\n")
	fmt.Fprintf(buf, "\tif a == nil {\n")
	fmt.Fprintf(buf, "\t\treturn false, fmt.Errorf(\"awaitable not initialized\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif a.handle == 0 {\n")
	fmt.Fprintf(buf, "\t\treturn false, nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tstate, err := runtime.ChannelStoreState(a.handle)\n")
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn false, err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tclosed, err := runtime.ChannelStoreIsClosed(a.handle)\n")
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn false, err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tlength := len(state.Ch)\n")
	fmt.Fprintf(buf, "\tcapacity := state.Capacity\n")
	fmt.Fprintf(buf, "\tsendWaiters, recvWaiters, sendAwaiters, recvAwaiters := __able_channel_await_snapshot(a.handle)\n")
	fmt.Fprintf(buf, "\tswitch a.op {\n")
	fmt.Fprintf(buf, "\tcase __able_channel_await_recv:\n")
	fmt.Fprintf(buf, "\t\tif length > 0 {\n")
	fmt.Fprintf(buf, "\t\t\treturn true, nil\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tif capacity == 0 && (sendWaiters > 0 || sendAwaiters > 0) {\n")
	fmt.Fprintf(buf, "\t\t\treturn true, nil\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tif closed {\n")
	fmt.Fprintf(buf, "\t\t\treturn true, nil\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\treturn false, nil\n")
	fmt.Fprintf(buf, "\tcase __able_channel_await_send:\n")
	fmt.Fprintf(buf, "\t\tif closed {\n")
	fmt.Fprintf(buf, "\t\t\tbridge.Raise(__able_concurrency_error_value(\"ChannelSendOnClosed\", \"send on closed channel\"))\n")
	fmt.Fprintf(buf, "\t\t\treturn false, nil\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tif capacity == 0 {\n")
	fmt.Fprintf(buf, "\t\t\treturn recvWaiters > 0 || recvAwaiters > 0, nil\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\treturn length < capacity, nil\n")
	fmt.Fprintf(buf, "\tdefault:\n")
	fmt.Fprintf(buf, "\t\treturn false, fmt.Errorf(\"unknown awaitable operation\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "}\n\n")
	if g.callableExecutionContextsEnabled() {
		fmt.Fprintf(buf, "func (a *__able_channel_awaitable) register(waker runtime.Value, ctx *runtime.NativeCallContext) (runtime.Value, error) {\n")
	} else {
		fmt.Fprintf(buf, "func (a *__able_channel_awaitable) register(waker runtime.Value) (runtime.Value, error) {\n")
	}
	fmt.Fprintf(buf, "\tif a == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"awaitable not initialized\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif a.handle == 0 {\n")
	if g.callableExecutionContextsEnabled() {
		fmt.Fprintf(buf, "\t\treturn __able_make_await_registration_value_ctx(nil, __able_context_from_native(ctx)), nil\n")
	} else {
		fmt.Fprintf(buf, "\t\treturn __able_make_await_registration_value(nil), nil\n")
	}
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif a.registration != nil {\n")
	if g.callableExecutionContextsEnabled() {
		fmt.Fprintf(buf, "\t\treturn __able_make_await_registration_value_ctx(a.registration.cancel, __able_context_from_native(ctx)), nil\n")
	} else {
		fmt.Fprintf(buf, "\t\treturn __able_make_await_registration_value(a.registration.cancel), nil\n")
	}
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treg := __able_add_channel_awaiter(a.handle, a.op, waker)\n")
	fmt.Fprintf(buf, "\ta.registration = reg\n")
	fmt.Fprintf(buf, "\tcancelFn := func() {\n")
	fmt.Fprintf(buf, "\t\tif a.registration != nil {\n")
	fmt.Fprintf(buf, "\t\t\ta.registration.cancel()\n")
	fmt.Fprintf(buf, "\t\t\ta.registration = nil\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t}\n")
	if g.callableExecutionContextsEnabled() {
		fmt.Fprintf(buf, "\treturn __able_make_await_registration_value_ctx(cancelFn, __able_context_from_native(ctx)), nil\n")
	} else {
		fmt.Fprintf(buf, "\treturn __able_make_await_registration_value(cancelFn), nil\n")
	}
	fmt.Fprintf(buf, "}\n\n")
	if g.callableExecutionContextsEnabled() {
		fmt.Fprintf(buf, "func (a *__able_channel_awaitable) commit(ctx *runtime.NativeCallContext) (runtime.Value, error) {\n")
	} else {
		fmt.Fprintf(buf, "func (a *__able_channel_awaitable) commit() (runtime.Value, error) {\n")
	}
	fmt.Fprintf(buf, "\tif a == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"awaitable not initialized\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\thandleVal := bridge.ToInt(a.handle, runtime.IntegerI64)\n")
	fmt.Fprintf(buf, "\tswitch a.op {\n")
	fmt.Fprintf(buf, "\tcase __able_channel_await_recv:\n")
	if g.callableExecutionContextsEnabled() {
		fmt.Fprintf(buf, "\t\tvalue, err := __able_channel_receive_impl([]runtime.Value{handleVal}, __able_context_from_native(ctx))\n")
	} else {
		fmt.Fprintf(buf, "\t\tvalue, err := __able_channel_receive_impl([]runtime.Value{handleVal})\n")
	}
	fmt.Fprintf(buf, "\t\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\t\treturn nil, err\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tif a.callback == nil {\n")
	fmt.Fprintf(buf, "\t\t\treturn value, nil\n")
	fmt.Fprintf(buf, "\t\t}\n")
	if g.callableExecutionContextsEnabled() {
		fmt.Fprintf(buf, "\t\treturn __able_call_value_fast_ctx(a.callback, []runtime.Value{value}, __able_context_from_native(ctx))\n")
	} else {
		fmt.Fprintf(buf, "\t\tresult, control := __able_call_value(a.callback, []runtime.Value{value}, nil)\n")
		fmt.Fprintf(buf, "\t\treturn result, __able_control_to_error(__able_runtime, nil, control)\n")
	}
	fmt.Fprintf(buf, "\tcase __able_channel_await_send:\n")
	if g.callableExecutionContextsEnabled() {
		fmt.Fprintf(buf, "\t\t_, err := __able_channel_send_impl([]runtime.Value{handleVal, a.payload}, __able_context_from_native(ctx))\n")
	} else {
		fmt.Fprintf(buf, "\t\t_, err := __able_channel_send_impl([]runtime.Value{handleVal, a.payload})\n")
	}
	fmt.Fprintf(buf, "\t\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\t\treturn nil, err\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tif a.callback == nil {\n")
	fmt.Fprintf(buf, "\t\t\treturn runtime.NilValue{}, nil\n")
	fmt.Fprintf(buf, "\t\t}\n")
	if g.callableExecutionContextsEnabled() {
		fmt.Fprintf(buf, "\t\treturn __able_call_value_fast_ctx(a.callback, nil, __able_context_from_native(ctx))\n")
	} else {
		fmt.Fprintf(buf, "\t\tresult, control := __able_call_value(a.callback, nil, nil)\n")
		fmt.Fprintf(buf, "\t\treturn result, __able_control_to_error(__able_runtime, nil, control)\n")
	}
	fmt.Fprintf(buf, "\tdefault:\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"unknown awaitable operation\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func (a *__able_channel_awaitable) toStruct() *runtime.StructInstanceValue {\n")
	fmt.Fprintf(buf, "\tinst := &runtime.StructInstanceValue{Fields: make(map[string]runtime.Value)}\n")
	fmt.Fprintf(buf, "\tisReady := runtime.NativeFunctionValue{\n")
	fmt.Fprintf(buf, "\t\tName:  \"Awaitable.is_ready\",\n")
	fmt.Fprintf(buf, "\t\tArity: 0,\n")
	fmt.Fprintf(buf, "\t\tImpl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {\n")
	fmt.Fprintf(buf, "\t\t\tready, err := a.isReady()\n")
	fmt.Fprintf(buf, "\t\t\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\t\t\treturn nil, err\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\treturn runtime.BoolValue{Val: ready}, nil\n")
	fmt.Fprintf(buf, "\t\t},\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tregister := runtime.NativeFunctionValue{\n")
	fmt.Fprintf(buf, "\t\tName:  \"Awaitable.register\",\n")
	fmt.Fprintf(buf, "\t\tArity: 1,\n")
	if g.callableExecutionContextsEnabled() {
		fmt.Fprintf(buf, "\t\tImpl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n")
	} else {
		fmt.Fprintf(buf, "\t\tImpl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n")
	}
	fmt.Fprintf(buf, "\t\t\tif len(args) == 0 {\n")
	fmt.Fprintf(buf, "\t\t\t\treturn nil, fmt.Errorf(\"register expects waker argument\")\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\twaker := args[len(args)-1]\n")
	if g.callableExecutionContextsEnabled() {
		fmt.Fprintf(buf, "\t\t\treturn a.register(waker, ctx)\n")
	} else {
		fmt.Fprintf(buf, "\t\t\treturn a.register(waker)\n")
	}
	fmt.Fprintf(buf, "\t\t},\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tcommit := runtime.NativeFunctionValue{\n")
	fmt.Fprintf(buf, "\t\tName:  \"Awaitable.commit\",\n")
	fmt.Fprintf(buf, "\t\tArity: 0,\n")
	if g.callableExecutionContextsEnabled() {
		fmt.Fprintf(buf, "\t\tImpl: func(ctx *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {\n")
		fmt.Fprintf(buf, "\t\t\treturn a.commit(ctx)\n")
	} else {
		fmt.Fprintf(buf, "\t\tImpl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {\n")
		fmt.Fprintf(buf, "\t\t\treturn a.commit()\n")
	}
	fmt.Fprintf(buf, "\t\t},\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tisDefault := runtime.NativeFunctionValue{\n")
	fmt.Fprintf(buf, "\t\tName:  \"Awaitable.is_default\",\n")
	fmt.Fprintf(buf, "\t\tArity: 0,\n")
	fmt.Fprintf(buf, "\t\tImpl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {\n")
	fmt.Fprintf(buf, "\t\t\treturn runtime.BoolValue{Val: false}, nil\n")
	fmt.Fprintf(buf, "\t\t},\n")
	fmt.Fprintf(buf, "\t}\n")
	g.renderClosureOwnedNativeMethodField(buf, "is_ready", "isReady")
	g.renderClosureOwnedNativeMethodField(buf, "register", "register")
	g.renderClosureOwnedNativeMethodField(buf, "commit", "commit")
	g.renderClosureOwnedNativeMethodField(buf, "is_default", "isDefault")
	fmt.Fprintf(buf, "\treturn inst\n")
	fmt.Fprintf(buf, "}\n\n")
}
