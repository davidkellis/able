package compiler

import "bytes"

func (g *generator) renderRuntimeAwaitHelpers(buf *bytes.Buffer) {
	buf.WriteString(g.receiverFreeClosureOwnedNativeMethods(`
const __able_max_sleep_ms = int64(2_147_483_647)

func __able_duration_from_value(val runtime.Value) (time.Duration, error) {
	switch v := __able_unwrap_interface(val).(type) {
	case runtime.IntegerValue:
		if n, ok := v.ToInt64(); ok {
			if n < 0 {
				n = 0
			}
			if n > __able_max_sleep_ms {
				n = __able_max_sleep_ms
			}
			return time.Duration(n) * time.Millisecond, nil
		}
		limit := big.NewInt(__able_max_sleep_ms)
		raw := new(big.Int).Set(v.BigInt())
		if raw.Sign() < 0 {
			raw = big.NewInt(0)
		}
		if raw.Cmp(limit) > 0 {
			raw = limit
		}
		return time.Duration(raw.Int64()) * time.Millisecond, nil
	case *runtime.IntegerValue:
		if v == nil {
			return 0, fmt.Errorf("sleep_ms expects a numeric duration")
		}
		if n, ok := v.ToInt64(); ok {
			if n < 0 {
				n = 0
			}
			if n > __able_max_sleep_ms {
				n = __able_max_sleep_ms
			}
			return time.Duration(n) * time.Millisecond, nil
		}
		limit := big.NewInt(__able_max_sleep_ms)
		raw := new(big.Int).Set(v.BigInt())
		if raw.Sign() < 0 {
			raw = big.NewInt(0)
		}
		if raw.Cmp(limit) > 0 {
			raw = limit
		}
		return time.Duration(raw.Int64()) * time.Millisecond, nil
	case runtime.FloatValue:
		if math.IsNaN(v.Val) || math.IsInf(v.Val, 0) {
			return 0, fmt.Errorf("sleep_ms expects a finite duration")
		}
		ms := math.Trunc(v.Val)
		if ms < 0 {
			ms = 0
		}
		if ms > float64(__able_max_sleep_ms) {
			ms = float64(__able_max_sleep_ms)
		}
		return time.Duration(int64(ms)) * time.Millisecond, nil
	case *runtime.FloatValue:
		if v == nil {
			return 0, fmt.Errorf("sleep_ms expects a numeric duration")
		}
		if math.IsNaN(v.Val) || math.IsInf(v.Val, 0) {
			return 0, fmt.Errorf("sleep_ms expects a finite duration")
		}
		ms := math.Trunc(v.Val)
		if ms < 0 {
			ms = 0
		}
		if ms > float64(__able_max_sleep_ms) {
			ms = float64(__able_max_sleep_ms)
		}
		return time.Duration(int64(ms)) * time.Millisecond, nil
	default:
		return 0, fmt.Errorf("sleep_ms expects a numeric duration")
	}
}

type __able_default_awaitable struct {
	callback runtime.Value
}

func (a *__able_default_awaitable) Kind() runtime.Kind {
	return runtime.KindStructInstance
}

func (a *__able_default_awaitable) NativeAwaitableIsReady(_ *runtime.NativeCallContext) (bool, error) {
	return true, nil
}

func (a *__able_default_awaitable) NativeAwaitableRegister(_ *runtime.NativeCallContext, _ runtime.Value) (runtime.Value, error) {
	return __able_make_await_registration_value(nil), nil
}

func (a *__able_default_awaitable) NativeAwaitableCommit(_ *runtime.NativeCallContext) (runtime.Value, error) {
	if a == nil || a.callback == nil {
		return runtime.NilValue{}, nil
	}
	value, control := __able_call_value(a.callback, nil, nil)
	return value, __able_control_to_error(__able_runtime, nil, control)
}

func (a *__able_default_awaitable) NativeAwaitableIsDefault() bool {
	return true
}

func (a *__able_default_awaitable) MaterializeRuntimeValue() runtime.Value {
	inst := &runtime.StructInstanceValue{
		Fields: make(map[string]runtime.Value),
	}
	isReady := runtime.NativeFunctionValue{
		Name:  "Awaitable.is_ready",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			return runtime.BoolValue{Val: true}, nil
		},
	}
	register := runtime.NativeFunctionValue{
		Name:  "Awaitable.register",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			return __able_make_await_registration_value(nil), nil
		},
	}
	commit := runtime.NativeFunctionValue{
		Name:  "Awaitable.commit",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			return a.NativeAwaitableCommit(nil)
		},
	}
	isDefault := runtime.NativeFunctionValue{
		Name:  "Awaitable.is_default",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			return runtime.BoolValue{Val: true}, nil
		},
	}
	inst.Fields["is_ready"] = &runtime.NativeBoundMethodValue{Receiver: inst, Method: isReady}
	inst.Fields["register"] = &runtime.NativeBoundMethodValue{Receiver: inst, Method: register}
	inst.Fields["commit"] = &runtime.NativeBoundMethodValue{Receiver: inst, Method: commit}
	inst.Fields["is_default"] = &runtime.NativeBoundMethodValue{Receiver: inst, Method: isDefault}
	return inst
}

func __able_make_default_awaitable(callback runtime.Value) runtime.Value {
	return &__able_default_awaitable{callback: callback}
}

type __able_timer_awaitable struct {
	deadline  time.Time
	callback  runtime.Value
	mu        sync.Mutex
	ready     bool
	cancelled bool
	timer     *time.Timer
}

func __able_new_timer_awaitable(duration time.Duration, callback runtime.Value) *__able_timer_awaitable {
	return &__able_timer_awaitable{
		deadline: time.Now().Add(duration),
		callback: callback,
	}
}

func (a *__able_timer_awaitable) Kind() runtime.Kind {
	return runtime.KindStructInstance
}

func (a *__able_timer_awaitable) NativeAwaitableIsReady(_ *runtime.NativeCallContext) (bool, error) {
	return a.isReady(), nil
}

func (a *__able_timer_awaitable) NativeAwaitableRegister(_ *runtime.NativeCallContext, waker runtime.Value) (runtime.Value, error) {
	return a.register(waker)
}

func (a *__able_timer_awaitable) NativeAwaitableCommit(_ *runtime.NativeCallContext) (runtime.Value, error) {
	return a.commit()
}

func (a *__able_timer_awaitable) NativeAwaitableIsDefault() bool {
	return false
}

func (a *__able_timer_awaitable) MaterializeRuntimeValue() runtime.Value {
	return a.toStruct()
}

func (a *__able_timer_awaitable) markReadyLocked() {
	a.ready = true
	if a.timer != nil {
		a.timer.Stop()
		a.timer = nil
	}
}

func (a *__able_timer_awaitable) isReady() bool {
	if a == nil {
		return false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.ready {
		return true
	}
	now := time.Now()
	if now.After(a.deadline) || now.Equal(a.deadline) {
		a.markReadyLocked()
		return true
	}
	return false
}

func (a *__able_timer_awaitable) register(waker runtime.Value) (runtime.Value, error) {
	if a == nil {
		return nil, fmt.Errorf("awaitable not initialized")
	}
	if a.isReady() {
		__able_invoke_await_waker(waker)
		return __able_make_await_registration_value(nil), nil
	}

	a.mu.Lock()
	a.cancelled = false
	if a.timer != nil {
		a.timer.Stop()
		a.timer = nil
	}
	remaining := time.Until(a.deadline)
	if remaining < 0 {
		remaining = 0
	}
	a.timer = time.AfterFunc(remaining, func() {
		a.mu.Lock()
		if a.cancelled {
			a.mu.Unlock()
			return
		}
		a.markReadyLocked()
		a.mu.Unlock()
		__able_invoke_await_waker(waker)
	})
	a.mu.Unlock()

	cancelFn := func() {
		a.mu.Lock()
		a.cancelled = true
		if a.timer != nil {
			a.timer.Stop()
			a.timer = nil
		}
		a.mu.Unlock()
	}
	return __able_make_await_registration_value(cancelFn), nil
}

func (a *__able_timer_awaitable) commit() (runtime.Value, error) {
	if a == nil {
		return nil, fmt.Errorf("awaitable not initialized")
	}
	a.mu.Lock()
	a.markReadyLocked()
	a.cancelled = false
	callback := a.callback
	a.mu.Unlock()
	if callback == nil {
		return runtime.NilValue{}, nil
	}
	value, control := __able_call_value(callback, nil, nil)
	return value, __able_control_to_error(__able_runtime, nil, control)
}

func (a *__able_timer_awaitable) toStruct() *runtime.StructInstanceValue {
	inst := &runtime.StructInstanceValue{
		Fields: make(map[string]runtime.Value),
	}
	isReady := runtime.NativeFunctionValue{
		Name:  "Awaitable.is_ready",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			return runtime.BoolValue{Val: a.isReady()}, nil
		},
	}
	register := runtime.NativeFunctionValue{
		Name:  "Awaitable.register",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("register expects waker argument")
			}
			waker := args[len(args)-1]
			return a.register(waker)
		},
	}
	commit := runtime.NativeFunctionValue{
		Name:  "Awaitable.commit",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			return a.commit()
		},
	}
	isDefault := runtime.NativeFunctionValue{
		Name:  "Awaitable.is_default",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			return runtime.BoolValue{Val: false}, nil
		},
	}
	inst.Fields["is_ready"] = &runtime.NativeBoundMethodValue{Receiver: inst, Method: isReady}
	inst.Fields["register"] = &runtime.NativeBoundMethodValue{Receiver: inst, Method: register}
	inst.Fields["commit"] = &runtime.NativeBoundMethodValue{Receiver: inst, Method: commit}
	inst.Fields["is_default"] = &runtime.NativeBoundMethodValue{Receiver: inst, Method: isDefault}
	return inst
}

func __able_await_default_impl(args []runtime.Value) (runtime.Value, error) {
	var callback runtime.Value
	if len(args) > 0 {
		callback = args[len(args)-1]
	}
	return __able_make_default_awaitable(callback), nil
}

func __able_await_sleep_ms_impl(args []runtime.Value) (runtime.Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("__able_await_sleep_ms expects duration")
	}
	duration, err := __able_duration_from_value(args[0])
	if err != nil {
		return nil, err
	}
	var callback runtime.Value
	if len(args) > 1 {
		callback = args[len(args)-1]
	}
	awaitable := __able_new_timer_awaitable(duration, callback)
	return awaitable, nil
}

type __able_await_arm_state struct {
	awaitable    runtime.Value
	isDefault    bool
	registration runtime.Value
}

type __able_await_state struct {
	mu          sync.Mutex
	arms        []*__able_await_arm_state
	defaultArm  *__able_await_arm_state
	waiting     bool
	wakePending bool
	waitCh      chan struct{}
	payload     *__able_async_payload
	waker       runtime.Value
` + g.awaitStateContextFields() + `
}

var __able_await_round_robin atomic.Int64

` + g.awaitSignalMethods() + `

func (s *__able_await_state) consumeWakePending() bool {
	if s == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.wakePending {
		return false
	}
	s.waiting = false
	s.wakePending = false
` + g.awaitSignalDrain() + `	return true
}

// beginWaiting publishes waiting state before registration. An awaitable may
// wake synchronously while register is running, so publishing it afterward
// could erase the wake and leave the task parked indefinitely.
func (s *__able_await_state) beginWaiting() bool {
	if s == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.waiting {
		return false
	}
	s.waiting = true
	s.wakePending = false
	return true
}

func (s *__able_await_state) clearWaiting() {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.waiting = false
	s.wakePending = false
` + g.awaitSignalDrain() + `	s.mu.Unlock()
}

` + g.awaitWakePendingMethod() + `

` + g.awaitStateCacheMethods() + `

func __able_await_value(expr *ast.AwaitExpression, iterable runtime.Value) (runtime.Value, error) {
	payload := __able_current_payload()
	if payload == nil || payload.handle == nil {
		return nil, fmt.Errorf("await expressions must run inside an asynchronous task")
	}
	state := payload.getAwaitState(expr)
	if state == nil {
		var err error
		state, err = __able_initialize_await_state(payload, iterable)
		if err != nil {
			return nil, err
		}
		payload.setAwaitState(expr, state)
	}
	return __able_await_with_state(payload, expr, state)
}

func __able_initialize_await_state(payload *__able_async_payload, iterable runtime.Value) (*__able_await_state, error) {
	arms, err := __able_collect_await_arms(iterable)
	if err != nil {
		return nil, err
	}
	if len(arms) == 0 {
		return nil, fmt.Errorf("await requires at least one arm")
	}
	var defaultArm *__able_await_arm_state
	for _, arm := range arms {
		if arm != nil && arm.isDefault {
			if defaultArm != nil {
				return nil, fmt.Errorf("await accepts at most one default arm")
			}
			defaultArm = arm
		}
	}
	state := &__able_await_state{
		arms:       arms,
		defaultArm: defaultArm,
		payload:    payload,
	}
	state.ensureWaitCh()
	waker, err := __able_make_await_waker(payload, state)
	if err != nil {
		return nil, err
	}
	state.waker = waker
	return state, nil
}

func __able_collect_await_arms(iterable runtime.Value) ([]*__able_await_arm_state, error) {
	if values, ok := __able_array_values(iterable); ok {
		arms := make([]*__able_await_arm_state, 0, len(values))
		for _, el := range values {
			arms = append(arms, &__able_await_arm_state{
				awaitable: el,
				isDefault: __able_await_arm_is_default(el),
			})
		}
		return arms, nil
	}
	iter := __able_resolve_iterator(iterable)
	if iter == nil {
		return nil, fmt.Errorf("await requires an Iterable of Awaitable values")
	}
	defer iter.Close()
	arms := make([]*__able_await_arm_state, 0)
	for {
		val, done, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if done {
			break
		}
		arms = append(arms, &__able_await_arm_state{
			awaitable: val,
			isDefault: __able_await_arm_is_default(val),
		})
	}
	return arms, nil
}

func __able_await_arm_is_default(awaitable runtime.Value) bool {
	result, err := __able_invoke_awaitable_method(awaitable, "is_default", nil)
	if err != nil {
		return false
	}
	return __able_truthy(result)
}

func __able_select_ready_await_arm(state *__able_await_state) (*__able_await_arm_state, error) {
	ready := make([]*__able_await_arm_state, 0)
	for _, arm := range state.arms {
		if arm == nil || arm.isDefault {
			continue
		}
		result, err := __able_invoke_awaitable_method(arm.awaitable, "is_ready", nil)
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
	start := idx % len(ready)
	return ready[start], nil
}

func __able_register_await_state(state *__able_await_state) error {
	if state.waker == nil {
		return fmt.Errorf("Await waker not initialised")
	}
	for _, arm := range state.arms {
		if arm == nil || arm.isDefault || arm.registration != nil {
			continue
		}
		reg, err := __able_invoke_awaitable_method(arm.awaitable, "register", []runtime.Value{state.waker})
		if err != nil {
			return err
		}
		arm.registration = reg
	}
	return nil
}

func __able_complete_await(payload *__able_async_payload, expr *ast.AwaitExpression, state *__able_await_state, winner *__able_await_arm_state) (runtime.Value, error) {
	for _, arm := range state.arms {
		if arm == nil || arm == winner {
			continue
		}
		__able_cancel_await_registration(arm.registration)
		arm.registration = nil
	}
	result, err := __able_invoke_awaitable_method(winner.awaitable, "commit", nil)
	if err != nil {
		return nil, err
	}
	__able_cleanup_await_state(payload, expr, state)
	if result == nil {
		return runtime.NilValue{}, nil
	}
	return result, nil
}

func __able_cleanup_await_state(payload *__able_async_payload, expr *ast.AwaitExpression, state *__able_await_state) {
	__able_clear_await_registrations(state)
	state.clearWaiting()
	if payload != nil {
		payload.awaitBlocked.Store(false)
		payload.clearAwaitState(expr)
	}
}

// __able_clear_await_registrations runs after every wake as well as during
// final cleanup. A wake only asks the task to recheck readiness; another task
// can acquire the resource before commit, so old one-shot registrations cannot
// be retained for the next wait cycle.
func __able_clear_await_registrations(state *__able_await_state) {
	if state == nil {
		return
	}
	for _, arm := range state.arms {
		if arm == nil {
			continue
		}
		__able_cancel_await_registration(arm.registration)
		arm.registration = nil
	}
}

func __able_cancel_await_registration(reg runtime.Value) {
	if reg == nil {
		return
	}
	_, control := __able_method_call(reg, "cancel", nil)
	_ = __able_control_to_error(__able_runtime, nil, control)
}

func __able_invoke_awaitable_method(awaitable runtime.Value, method string, args []runtime.Value) (runtime.Value, error) {
	if native, ok := awaitable.(runtime.NativeAwaitableValue); ok {
		switch method {
		case "is_ready":
			ready, err := native.NativeAwaitableIsReady(nil)
			return runtime.BoolValue{Val: ready}, err
		case "register":
			if len(args) == 0 {
				return nil, fmt.Errorf("register expects waker argument")
			}
			return native.NativeAwaitableRegister(nil, args[len(args)-1])
		case "commit":
			return native.NativeAwaitableCommit(nil)
		case "is_default":
			return runtime.BoolValue{Val: native.NativeAwaitableIsDefault()}, nil
		}
	}
	val, control := __able_method_call(awaitable, method, args)
	if control != nil {
		return nil, __able_control_to_error(__able_runtime, nil, control)
	}
	return val, nil
}

func __able_make_await_waker(payload *__able_async_payload, state *__able_await_state) (runtime.Value, error) {
	if __able_runtime == nil {
		return nil, fmt.Errorf("Await waker builtins are not initialized")
	}
	def, err := __able_runtime.StructDefinition("AwaitWaker")
	if err != nil || def == nil {
		return nil, fmt.Errorf("Await waker builtins are not initialized")
	}
	inst := &runtime.StructInstanceValue{
		Definition: def,
		Fields:     make(map[string]runtime.Value),
	}` + g.awaitWakerContextSpacing() + `
	wakeFn := runtime.NativeFunctionValue{
		Name:  "AwaitWaker.wake",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
` + g.legacyAwaitWakeBeforePayload() + `			if payload != nil {
				payload.awaitBlocked.Store(false)
			}
` + g.legacyAwaitWakeAfterPayload() + `			if payload != nil && payload.resumeTask != nil {
				payload.resumeTask()
			}
			return runtime.NilValue{}, nil
		},
	}
	inst.Fields["wake"] = &runtime.NativeBoundMethodValue{Receiver: inst, Method: wakeFn}
	return inst, nil
}

func __able_await_with_state(payload *__able_async_payload, expr *ast.AwaitExpression, state *__able_await_state) (runtime.Value, error) {
	for {
		winner, err := __able_select_ready_await_arm(state)
		if err != nil {
			return nil, err
		}
		if winner != nil {
			return __able_complete_await(payload, expr, state, winner)
		}
		if state.defaultArm != nil {
			return __able_complete_await(payload, expr, state, state.defaultArm)
		}
		if payload != nil && payload.handle != nil && payload.handle.CancelRequested() {
			__able_cleanup_await_state(payload, expr, state)
			return nil, context.Canceled
		}
		if state.consumeWakePending() {
			__able_clear_await_registrations(state)
			continue
		}
		if state.beginWaiting() {
			if err := __able_register_await_state(state); err != nil {
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
			// Goroutine-backed tasks have no cooperative resume channel. Their
			// waker signals waitCh directly, while task cancellation wakes the
			// wait through the Future context.
			if exec := __able_future_executor(); exec != nil {
				exec.MarkBlocked(payload.handle)
			}
			ctx := payload.handle.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			select {
			case <-waitCh:
			case <-ctx.Done():
				if exec := __able_future_executor(); exec != nil {
					exec.MarkUnblocked(payload.handle)
				}
				__able_cleanup_await_state(payload, expr, state)
				return nil, context.Canceled
			}
			if exec := __able_future_executor(); exec != nil {
				exec.MarkUnblocked(payload.handle)
			}
		}

		payload.awaitBlocked.Store(false)
		__able_clear_await_registrations(state)
		state.clearWaiting()
	}
}

func __able_call_value_fast(fn runtime.Value, args []runtime.Value) (runtime.Value, error) {
	if __able_runtime == nil {
		return nil, fmt.Errorf("compiler: missing runtime")
	}
	env := __able_runtime.Env()
	var state any
	if env != nil {
		state = env.RuntimeData()
	}
	ctx := &runtime.NativeCallContext{Env: env, State: state}
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
		return bridge.CallValueWithNode(__able_runtime, fn, args, nil)
	}
}

func __able_call_known_native_method_fast(receiver runtime.Value, entry *__able_compiled_method_entry, args []runtime.Value) (runtime.Value, error) {
	if __able_runtime == nil {
		return nil, fmt.Errorf("compiler: missing runtime")
	}
	if entry == nil || entry.fn == nil {
		return nil, fmt.Errorf("compiler: missing compiled method")
	}
	env := __able_runtime.Env()
	if entry.direct != nil {
		return entry.direct(__able_runtime, env, receiver, args)
	}
	var state any
	if env != nil {
		state = env.RuntimeData()
	}
	ctx := &runtime.NativeCallContext{Env: env, State: state}
	injected := append([]runtime.Value{receiver}, args...)
	return entry.fn.Impl(ctx, injected)
}
`))
	g.renderRuntimeExecutionContextCallHelpers(buf)
	g.renderRuntimeExecutionContextAwaitHelpers(buf)
}
