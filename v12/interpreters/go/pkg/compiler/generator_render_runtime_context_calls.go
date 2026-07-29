package compiler

import "bytes"

func (g *generator) renderRuntimeExecutionContextCallHelpers(buf *bytes.Buffer) {
	if g == nil || buf == nil || !g.callableExecutionContextsEnabled() {
		return
	}
	buf.WriteString(`
func __able_native_call_context(__able_exec_ctx *__able_execution_context) *runtime.NativeCallContext {
	__able_exec_ctx = __able_context_from_args(__able_exec_ctx)
	return &__able_exec_ctx.native
}

func __able_call_value_fast_ctx(fn runtime.Value, args []runtime.Value, __able_exec_ctx *__able_execution_context) (runtime.Value, error) {
	if __able_runtime == nil {
		return nil, fmt.Errorf("compiler: missing runtime")
	}
	ctx := __able_native_call_context(__able_exec_ctx)
	switch v := fn.(type) {
	case runtime.NativeFunctionValue:
		return v.Impl(ctx, args)
	case *runtime.NativeFunctionValue:
		if v == nil {
			return nil, fmt.Errorf("native function is nil")
		}
		return v.Impl(ctx, args)
	case runtime.NativeBoundMethodValue:
		injected := append([]runtime.Value{v.Receiver}, args...)
		return v.Method.Impl(ctx, injected)
	case *runtime.NativeBoundMethodValue:
		if v == nil {
			return nil, fmt.Errorf("native bound method is nil")
		}
		injected := append([]runtime.Value{v.Receiver}, args...)
		return v.Method.Impl(ctx, injected)
	default:
		if ctx.Env != nil {
			if previous, swapped := bridge.SwapEnvIfNeeded(__able_runtime, ctx.Env); swapped {
				defer bridge.RestoreEnvIfNeeded(__able_runtime, previous, swapped)
			}
		}
		return bridge.CallValueWithNode(__able_runtime, fn, args, nil)
	}
}

func __able_method_call_ctx(obj runtime.Value, methodName string, args []runtime.Value, __able_exec_ctx *__able_execution_context) (runtime.Value, *__ableControl) {
	method, err := __able_try_member_get_method(obj, runtime.StringValue{Val: methodName})
	if err != nil {
		return runtime.NilValue{}, __able_control_from_error(err)
	}
	val, err := __able_call_value_fast_ctx(method, args, __able_exec_ctx)
	if err != nil {
		return runtime.NilValue{}, __able_control_from_error(err)
	}
	if val == nil {
		return runtime.NilValue{}, nil
	}
	return val, nil
}

func __able_method_call_node_ctx(obj runtime.Value, methodName string, args []runtime.Value, call *ast.FunctionCall, __able_exec_ctx *__able_execution_context) (runtime.Value, *__ableControl) {
	__able_pushed_call_frame := false
	if call != nil {
		bridge.PushCallFrame(__able_runtime, call)
		__able_pushed_call_frame = true
		defer func() {
			if __able_pushed_call_frame {
				bridge.PopCallFrame(__able_runtime)
			}
		}()
	}
	method, err := __able_try_member_get_method(obj, runtime.StringValue{Val: methodName})
	if err != nil {
		if __able_pushed_call_frame && __able_is_caller_frame_error(err) {
			bridge.PopCallFrame(__able_runtime)
			__able_pushed_call_frame = false
			return runtime.NilValue{}, __able_control_from_error(err)
		}
		return runtime.NilValue{}, __able_control_from_error_with_node(call, err)
	}
	val, err := __able_call_value_fast_ctx(method, args, __able_exec_ctx)
	if err != nil {
		if __able_pushed_call_frame && __able_is_caller_frame_error(err) {
			bridge.PopCallFrame(__able_runtime)
			__able_pushed_call_frame = false
			return runtime.NilValue{}, __able_control_from_error(err)
		}
		return runtime.NilValue{}, __able_control_from_error_with_node(call, err)
	}
	if val == nil {
		return runtime.NilValue{}, nil
	}
	return val, nil
}
`)
}

func (g *generator) renderRuntimeExecutionContextAwaitHelpers(buf *bytes.Buffer) {
	if g == nil || buf == nil || !g.executionContextsEnabled() ||
		(len(g.awaitExprs) == 0 && !g.callableExecutionContextsEnabled()) {
		return
	}
	buf.WriteString(`
func __able_await_value_ctx(expr *ast.AwaitExpression, iterable runtime.Value, __able_exec_ctx *__able_execution_context) (runtime.Value, error) {
	payload := __able_context_payload(__able_exec_ctx)
	if payload == nil || payload.handle == nil {
		return nil, fmt.Errorf("await expressions must run inside an asynchronous task")
	}
	state := payload.getAwaitState(expr)
	if state == nil {
		var err error
		state, err = __able_initialize_await_state_ctx(payload, iterable, __able_exec_ctx)
		if err != nil {
			return nil, err
		}
		payload.setAwaitState(expr, state)
	}
	return __able_await_with_state_ctx(payload, expr, state, __able_exec_ctx)
}

func __able_initialize_await_state_ctx(payload *__able_async_payload, iterable runtime.Value, __able_exec_ctx *__able_execution_context) (*__able_await_state, error) {
	state := __able_acquire_await_state(payload)
	if err := __able_collect_await_arms_ctx(state, iterable, __able_exec_ctx); err != nil {
		state.releaseReusable()
		return nil, err
	}
	if len(state.arms) == 0 {
		state.releaseReusable()
		return nil, fmt.Errorf("await requires at least one arm")
	}
	var defaultArm *__able_await_arm_state
	for _, arm := range state.arms {
		if arm != nil && arm.isDefault {
			if defaultArm != nil {
				state.releaseReusable()
				return nil, fmt.Errorf("await accepts at most one default arm")
			}
			defaultArm = arm
		}
	}
	state.defaultArm = defaultArm
	state.ensureWaitCh()
	waker, err := __able_make_await_waker_ctx(payload, state, __able_exec_ctx)
	if err != nil {
		state.releaseReusable()
		return nil, err
	}
	state.setWaker(waker)
	return state, nil
}

type __able_native_await_waker struct {
	payload      *__able_async_payload
	state        *__able_await_state
	context      *__able_execution_context
	definition   *runtime.StructDefinitionValue
	materialized atomic.Pointer[runtime.StructInstanceValue]
}

func (w *__able_native_await_waker) Kind() runtime.Kind {
	return runtime.KindStructInstance
}

func (w *__able_native_await_waker) wake() {
	if w == nil || w.state == nil {
		return
	}
	if !w.state.signalWaker(w) {
		return
	}
	if w.payload != nil {
		w.payload.awaitBlocked.Store(false)
	}
	if w.payload != nil && w.payload.resumeTask != nil {
		w.payload.resumeTask()
	}
}

func (w *__able_native_await_waker) MaterializeRuntimeValue() runtime.Value {
	if w == nil {
		return (*runtime.StructInstanceValue)(nil)
	}
	if materialized := w.materialized.Load(); materialized != nil {
		return materialized
	}
	inst := &runtime.StructInstanceValue{
		Definition: w.definition,
		Fields:     make(map[string]runtime.Value),
	}
	inst.Fields["wake"] = runtime.NativeFunctionValue{
		Name:  "AwaitWaker.wake",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			w.wake()
			return runtime.NilValue{}, nil
		},
	}
	if w.materialized.CompareAndSwap(nil, inst) {
		return inst
	}
	return w.materialized.Load()
}

type __able_native_await_registration struct {
	cancelFn     func()
	definition   *runtime.StructDefinitionValue
	materialized atomic.Pointer[runtime.StructInstanceValue]
}

func (r *__able_native_await_registration) Kind() runtime.Kind {
	return runtime.KindStructInstance
}

func (r *__able_native_await_registration) cancel() {
	if r != nil && r.cancelFn != nil {
		r.cancelFn()
	}
}

func (r *__able_native_await_registration) MaterializeRuntimeValue() runtime.Value {
	if r == nil {
		return (*runtime.StructInstanceValue)(nil)
	}
	if materialized := r.materialized.Load(); materialized != nil {
		return materialized
	}
	inst := &runtime.StructInstanceValue{
		Definition: r.definition,
		Fields:     make(map[string]runtime.Value),
	}
	inst.Fields["cancel"] = runtime.NativeFunctionValue{
		Name:  "AwaitRegistration.cancel",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			r.cancel()
			return runtime.NilValue{}, nil
		},
	}
	if r.materialized.CompareAndSwap(nil, inst) {
		return inst
	}
	return r.materialized.Load()
}

func __able_make_await_waker_ctx(payload *__able_async_payload, state *__able_await_state, __able_exec_ctx *__able_execution_context) (runtime.Value, error) {
	if __able_runtime == nil {
		return nil, fmt.Errorf("Await waker builtins are not initialized")
	}
	__able_exec_ctx = __able_context_from_args(__able_exec_ctx)
	def, err := __able_runtime.StructDefinitionIn(__able_exec_ctx.env, "AwaitWaker")
	if err != nil || def == nil {
		return nil, fmt.Errorf("Await waker builtins are not initialized")
	}
	return &__able_native_await_waker{
		payload:    payload,
		state:      state,
		context:    __able_exec_ctx,
		definition: def,
	}, nil
}

func __able_make_await_registration_value_ctx(cancelFn func(), __able_exec_ctx *__able_execution_context) runtime.Value {
	var def *runtime.StructDefinitionValue
	if __able_runtime != nil {
		__able_exec_ctx = __able_context_from_args(__able_exec_ctx)
		if found, err := __able_runtime.StructDefinitionIn(__able_exec_ctx.env, "AwaitRegistration"); err == nil {
			def = found
		}
	}
	return &__able_native_await_registration{
		cancelFn:   cancelFn,
		definition: def,
	}
}

func __able_collect_await_arms_ctx(state *__able_await_state, iterable runtime.Value, __able_exec_ctx *__able_execution_context) error {
	if values, ok := __able_array_values(iterable); ok {
		state.prepareArmScratch(len(values))
		for _, el := range values {
			state.appendArm(el, __able_await_arm_is_default_ctx(el, __able_exec_ctx))
		}
		return nil
	}
	iter := __able_resolve_iterator(iterable)
	if iter == nil {
		return fmt.Errorf("await requires an Iterable of Awaitable values")
	}
	defer iter.Close()
	state.prepareArmScratch(0)
	for {
		val, done, err := iter.Next()
		if err != nil {
			return err
		}
		if done {
			break
		}
		state.appendArm(val, __able_await_arm_is_default_ctx(val, __able_exec_ctx))
	}
	return nil
}

func __able_await_arm_is_default_ctx(awaitable runtime.Value, __able_exec_ctx *__able_execution_context) bool {
	result, err := __able_invoke_awaitable_method_ctx(awaitable, "is_default", nil, __able_exec_ctx)
	return err == nil && __able_truthy(result)
}

func __able_select_ready_await_arm_ctx(state *__able_await_state, __able_exec_ctx *__able_execution_context) (*__able_await_arm_state, error) {
	ready := make([]*__able_await_arm_state, 0)
	for _, arm := range state.arms {
		if arm == nil || arm.isDefault {
			continue
		}
		result, err := __able_invoke_awaitable_method_ctx(arm.awaitable, "is_ready", nil, __able_exec_ctx)
		if err != nil {
			return nil, err
		}
		if __able_truthy(result) {
			ready = append(ready, arm)
		}
	}
	if len(ready) == 0 {
		return nil, nil
	}
	idx := int(__able_await_round_robin.Add(1) - 1)
	if idx < 0 {
		idx = 0
	}
	return ready[idx%len(ready)], nil
}

func __able_register_await_state_ctx(state *__able_await_state, __able_exec_ctx *__able_execution_context) error {
	if state.waker == nil {
		return fmt.Errorf("Await waker not initialised")
	}
	for _, arm := range state.arms {
		if arm == nil || arm.isDefault || arm.registration != nil {
			continue
		}
		reg, err := __able_invoke_awaitable_method_ctx(arm.awaitable, "register", []runtime.Value{state.waker}, __able_exec_ctx)
		if err != nil {
			return err
		}
		arm.registration = reg
	}
	return nil
}

func __able_cancel_await_registration_ctx(reg runtime.Value, __able_exec_ctx *__able_execution_context) {
	if reg == nil {
		return
	}
	if native, ok := reg.(*__able_native_await_registration); ok {
		native.cancel()
		return
	}
	_, control := __able_method_call_ctx(reg, "cancel", nil, __able_exec_ctx)
	_ = __able_control_to_error(__able_runtime, __able_native_call_context(__able_exec_ctx), control)
}

func __able_clear_await_registrations_ctx(state *__able_await_state, __able_exec_ctx *__able_execution_context) {
	if state == nil {
		return
	}
	for _, arm := range state.arms {
		if arm == nil {
			continue
		}
		__able_cancel_await_registration_ctx(arm.registration, __able_exec_ctx)
		arm.registration = nil
	}
}

func __able_cleanup_await_state_ctx(payload *__able_async_payload, expr *ast.AwaitExpression, state *__able_await_state, __able_exec_ctx *__able_execution_context) {
	__able_clear_await_registrations_ctx(state, __able_exec_ctx)
	state.clearWaiting()
	if payload != nil {
		payload.awaitBlocked.Store(false)
		payload.clearAwaitState(expr)
	}
	state.releaseReusable()
}

func __able_complete_await_ctx(payload *__able_async_payload, expr *ast.AwaitExpression, state *__able_await_state, winner *__able_await_arm_state, __able_exec_ctx *__able_execution_context) (runtime.Value, error) {
	for _, arm := range state.arms {
		if arm == nil || arm == winner {
			continue
		}
		__able_cancel_await_registration_ctx(arm.registration, __able_exec_ctx)
		arm.registration = nil
	}
	result, err := __able_invoke_awaitable_method_ctx(winner.awaitable, "commit", nil, __able_exec_ctx)
	if err != nil {
		return nil, err
	}
	__able_cleanup_await_state_ctx(payload, expr, state, __able_exec_ctx)
	if result == nil {
		return runtime.NilValue{}, nil
	}
	return result, nil
}

func __able_invoke_awaitable_method_ctx(awaitable runtime.Value, method string, args []runtime.Value, __able_exec_ctx *__able_execution_context) (runtime.Value, error) {
	if native, ok := awaitable.(runtime.NativeAwaitableValue); ok {
		ctx := __able_native_call_context(__able_exec_ctx)
		switch method {
		case "is_ready":
			ready, err := native.NativeAwaitableIsReady(ctx)
			return runtime.BoolValue{Val: ready}, err
		case "register":
			if len(args) == 0 {
				return nil, fmt.Errorf("register expects waker argument")
			}
			return native.NativeAwaitableRegister(ctx, args[len(args)-1])
		case "commit":
			return native.NativeAwaitableCommit(ctx)
		case "is_default":
			return runtime.BoolValue{Val: native.NativeAwaitableIsDefault()}, nil
		}
	}
	val, control := __able_method_call_ctx(awaitable, method, args, __able_exec_ctx)
	if control != nil {
		return nil, __able_control_to_error(__able_runtime, __able_native_call_context(__able_exec_ctx), control)
	}
	return val, nil
}

func __able_await_with_state_ctx(payload *__able_async_payload, expr *ast.AwaitExpression, state *__able_await_state, __able_exec_ctx *__able_execution_context) (runtime.Value, error) {
	for {
		winner, err := __able_select_ready_await_arm_ctx(state, __able_exec_ctx)
		if err != nil {
			return nil, err
		}
		if winner != nil {
			return __able_complete_await_ctx(payload, expr, state, winner, __able_exec_ctx)
		}
		if state.defaultArm != nil {
			return __able_complete_await_ctx(payload, expr, state, state.defaultArm, __able_exec_ctx)
		}
		if payload != nil && payload.handle != nil && payload.handle.CancelRequested() {
			__able_cleanup_await_state_ctx(payload, expr, state, __able_exec_ctx)
			return nil, context.Canceled
		}
		if state.consumeWakePending() {
			__able_clear_await_registrations_ctx(state, __able_exec_ctx)
			continue
		}
		if state.beginWaiting() {
			if err := __able_register_await_state_ctx(state, __able_exec_ctx); err != nil {
				state.clearWaiting()
				return nil, err
			}
		}
		waitCh := state.ensureWaitCh()
		if payload == nil || payload.handle == nil {
			return nil, fmt.Errorf("await expressions must run inside an asynchronous task")
		}
		payload.awaitBlocked.Store(true)
		if payload.yield != nil && payload.resume != nil {
			payload.yield <- __able_compiled_yield{}
			<-payload.resume
			select {
			case <-waitCh:
			default:
			}
		} else {
			if exec := __able_future_executor(); exec != nil {
				exec.MarkBlocked(payload.handle)
			}
			waitContext := payload.handle.Context()
			if waitContext == nil {
				waitContext = context.Background()
			}
			select {
			case <-waitCh:
			case <-waitContext.Done():
				if exec := __able_future_executor(); exec != nil {
					exec.MarkUnblocked(payload.handle)
				}
				__able_cleanup_await_state_ctx(payload, expr, state, __able_exec_ctx)
				return nil, context.Canceled
			}
			if exec := __able_future_executor(); exec != nil {
				exec.MarkUnblocked(payload.handle)
			}
		}
		payload.awaitBlocked.Store(false)
		__able_clear_await_registrations_ctx(state, __able_exec_ctx)
		state.clearWaiting()
	}
}

func __able_await_ctx(expr *ast.AwaitExpression, iterable runtime.Value, __able_exec_ctx *__able_execution_context) runtime.Value {
	if __able_runtime == nil {
		panic(fmt.Errorf("compiler: missing runtime"))
	}
	val, err := __able_await_value_ctx(expr, iterable, __able_exec_ctx)
	__able_panic_on_error(err)
	if val == nil {
		return runtime.NilValue{}
	}
	return val
}
`)
}
