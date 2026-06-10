package compiler

import (
	"bytes"
	"fmt"
)

func (g *generator) renderRuntimeMutexAwaitableSurface(buf *bytes.Buffer) {
	fmt.Fprintf(buf, "func (a *__able_mutex_awaitable) Kind() runtime.Kind { return runtime.KindStructInstance }\n\n")
	fmt.Fprintf(buf, "func (a *__able_mutex_awaitable) NativeAwaitableIsReady(_ *runtime.NativeCallContext) (bool, error) {\n")
	fmt.Fprintf(buf, "\treturn a.isReady()\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func (a *__able_mutex_awaitable) NativeAwaitableRegister(ctx *runtime.NativeCallContext, waker runtime.Value) (runtime.Value, error) {\n")
	if g.callableExecutionContextsEnabled() {
		fmt.Fprintf(buf, "\treturn a.register(waker, ctx)\n")
	} else {
		fmt.Fprintf(buf, "\treturn a.register(waker)\n")
	}
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func (a *__able_mutex_awaitable) NativeAwaitableCommit(ctx *runtime.NativeCallContext) (runtime.Value, error) {\n")
	if g.callableExecutionContextsEnabled() {
		fmt.Fprintf(buf, "\treturn a.commit(ctx)\n")
	} else {
		fmt.Fprintf(buf, "\treturn a.commit()\n")
	}
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func (a *__able_mutex_awaitable) NativeAwaitableIsDefault() bool { return false }\n\n")
	fmt.Fprintf(buf, "func (a *__able_mutex_awaitable) MaterializeRuntimeValue() runtime.Value { return a.toStruct() }\n\n")
	fmt.Fprintf(buf, "func (a *__able_mutex_awaitable) toStruct() *runtime.StructInstanceValue {\n")
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
	fmt.Fprintf(buf, "func __able_mutex_await_lock_impl(args []runtime.Value%s) (runtime.Value, error) {\n", g.runtimeConcurrencyContextParams())
	fmt.Fprintf(buf, "\tif len(args) < 1 {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"__able_mutex_await_lock expects handle and callback\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\thandle, err := __able_int64_from_value(args[0], \"mutex handle\")\n")
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tvar callback runtime.Value\n")
	fmt.Fprintf(buf, "\tif len(args) > 1 {\n")
	fmt.Fprintf(buf, "\t\tcallback = args[1]\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tawaitable := &__able_mutex_awaitable{handle: handle, callback: callback}\n")
	fmt.Fprintf(buf, "\treturn awaitable, nil\n")
	fmt.Fprintf(buf, "}\n\n")
}
