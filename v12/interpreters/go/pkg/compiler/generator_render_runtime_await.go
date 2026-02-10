package compiler

import "bytes"

func (g *generator) renderRuntimeAwaitHelpers(buf *bytes.Buffer) {
	buf.WriteString(`
const __able_max_sleep_ms = int64(2_147_483_647)

func __able_duration_from_value(val runtime.Value) (time.Duration, error) {
	switch v := __able_unwrap_interface(val).(type) {
	case runtime.IntegerValue:
		limit := big.NewInt(__able_max_sleep_ms)
		raw := new(big.Int).Set(v.Val)
		if raw.Sign() < 0 {
			raw = big.NewInt(0)
		}
		if raw.Cmp(limit) > 0 {
			raw = limit
		}
		return time.Duration(raw.Int64()) * time.Millisecond, nil
	case *runtime.IntegerValue:
		if v == nil || v.Val == nil {
			return 0, fmt.Errorf("sleep_ms expects a numeric duration")
		}
		limit := big.NewInt(__able_max_sleep_ms)
		raw := new(big.Int).Set(v.Val)
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

func __able_make_default_awaitable(callback runtime.Value) runtime.Value {
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
			if callback == nil {
				return runtime.NilValue{}, nil
			}
			return __able_call_value(callback, nil, nil), nil
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
	return __able_call_value(callback, nil, nil), nil
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
	return awaitable.toStruct(), nil
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
}

var __able_await_round_robin atomic.Int64

func (s *__able_await_state) ensureWaitCh() chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.waitCh == nil {
		s.waitCh = make(chan struct{}, 1)
	}
	return s.waitCh
}

func (s *__able_await_state) signal() {
	ch := s.ensureWaitCh()
	select {
	case ch <- struct{}{}:
	default:
	}
}

func (p *__able_async_payload) getAwaitState(expr *ast.AwaitExpression) *__able_await_state {
	if p == nil || expr == nil {
		return nil
	}
	if p.awaitStates == nil {
		return nil
	}
	return p.awaitStates[expr]
}

func (p *__able_async_payload) setAwaitState(expr *ast.AwaitExpression, state *__able_await_state) {
	if p == nil || expr == nil || state == nil {
		return
	}
	if p.awaitStates == nil {
		p.awaitStates = make(map[*ast.AwaitExpression]*__able_await_state)
	}
	p.awaitStates[expr] = state
}

func (p *__able_async_payload) clearAwaitState(expr *ast.AwaitExpression) {
	if p == nil || expr == nil {
		return
	}
	if p.awaitStates == nil {
		return
	}
	delete(p.awaitStates, expr)
}

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
	for _, arm := range state.arms {
		if arm == nil {
			continue
		}
		__able_cancel_await_registration(arm.registration)
		arm.registration = nil
	}
	state.waiting = false
	state.wakePending = false
	if payload != nil {
		payload.awaitBlocked = false
		payload.clearAwaitState(expr)
	}
}

func __able_cancel_await_registration(reg runtime.Value) {
	if reg == nil {
		return
	}
	member := __able_member(reg, runtime.StringValue{Val: "cancel"})
	_, _ = __able_call_value_fast(member, nil)
}

func __able_invoke_awaitable_method(awaitable runtime.Value, method string, args []runtime.Value) (runtime.Value, error) {
	member := __able_member(awaitable, runtime.StringValue{Val: method})
	return __able_call_value_fast(member, args)
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
	}
	triggered := false
	wakeFn := runtime.NativeFunctionValue{
		Name:  "AwaitWaker.wake",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			if triggered {
				return runtime.NilValue{}, nil
			}
			triggered = true
			state.wakePending = true
			if payload != nil {
				payload.awaitBlocked = false
			}
			state.signal()
			if payload != nil && payload.resumeTask != nil {
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
		if state.wakePending {
			state.waiting = false
			state.wakePending = false
			continue
		}
		if !state.waiting {
			if err := __able_register_await_state(state); err != nil {
				return nil, err
			}
			state.waiting = true
			state.wakePending = false
		}

		waitCh := state.ensureWaitCh()
		payload.awaitBlocked = true

		if payload == nil || payload.yield == nil || payload.resume == nil {
			return nil, fmt.Errorf("await expressions must run inside an asynchronous task")
		}

		payload.yield <- __able_compiled_yield{}
		<-payload.resume

		payload.awaitBlocked = false
		state.waiting = false
		state.wakePending = false

		select {
		case <-waitCh:
		default:
		}
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
`)
}
