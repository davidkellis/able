package compiler

func (g *generator) asyncPayloadContextFields() string {
	fields := ""
	if g.callableExecutionContextsEnabled() {
		fields = "\texecutionContext atomic.Pointer[__able_execution_context]\n"
	}
	if g.executionContextsEnabled() {
		fields += "\tawaitState *__able_await_state\n"
	}
	return fields
}

func (g *generator) awaitStateContextFields() string {
	return ""
}

func (g *generator) asyncPayloadAwaitStateField() string {
	if g.executionContextsEnabled() {
		return ""
	}
	return "\tawaitStates  map[*ast.AwaitExpression]*__able_await_state\n"
}

func (g *generator) taskOwnedAwaitSignal() string {
	if !g.executionContextsEnabled() {
		return ""
	}
	return `
	if s.payload != nil && s.payload.awaitState == s {
		if s.waitCh == nil {
			s.waitCh = make(chan struct{}, 1)
		}
		return s.waitCh
	}`
}

func (g *generator) awaitStateCacheMethods() string {
	if g.executionContextsEnabled() {
		return `
func (p *__able_async_payload) getAwaitState(_ *ast.AwaitExpression) *__able_await_state {
	return nil
}

func (p *__able_async_payload) setAwaitState(_ *ast.AwaitExpression, _ *__able_await_state) {}

func (p *__able_async_payload) clearAwaitState(_ *ast.AwaitExpression) {}
`
	}
	return `
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
`
}

func (g *generator) awaitSignalMethods() string {
	if !g.executionContextsEnabled() {
		return `
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
`
	}
	return `
var __able_await_waker_pending runtime.Value = runtime.NilValue{}

func __able_acquire_await_state(payload *__able_async_payload) *__able_await_state {
	if payload == nil {
		return &__able_await_state{waker: __able_await_waker_pending}
	}
	state := payload.awaitState
	if state == nil {
		state = &__able_await_state{waker: __able_await_waker_pending, payload: payload}
		payload.awaitState = state
		return state
	}
	if state.waker != nil {
		return &__able_await_state{waker: __able_await_waker_pending, payload: payload}
	}
	state.setWaker(__able_await_waker_pending)
	return state
}

func (s *__able_await_state) prepareArmScratch(capacity int) {
	s.arms = s.arms[:0]
	if cap(s.arms) < capacity {
		s.arms = make([]*__able_await_arm_state, 0, capacity)
	}
	s.defaultArm = nil
}

func (s *__able_await_state) appendArm(awaitable runtime.Value, isDefault bool) {
	idx := len(s.arms)
	if idx == cap(s.arms) {
		s.arms = append(s.arms, &__able_await_arm_state{})
	} else {
		s.arms = s.arms[:idx+1]
		if s.arms[idx] == nil {
			s.arms[idx] = &__able_await_arm_state{}
		}
	}
	arm := s.arms[idx]
	arm.awaitable = awaitable
	arm.isDefault = isDefault
	arm.registration = nil
}

func (s *__able_await_state) releaseReusable() {
	if s == nil {
		return
	}
	for _, arm := range s.arms {
		if arm == nil {
			continue
		}
		arm.awaitable = nil
		arm.isDefault = false
		arm.registration = nil
	}
	s.defaultArm = nil
	s.clearWaker()
	if s.payload != nil && s.payload.awaitState == s {
		s.arms = s.arms[:0]
		return
	}
	s.arms = nil
}

func (s *__able_await_state) setWaker(waker runtime.Value) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.waker = waker
	s.mu.Unlock()
}

func (s *__able_await_state) clearWaker() {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.waker = nil
	s.mu.Unlock()
}

func (s *__able_await_state) ensureWaitCh() chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.waitChLocked()
}

func (s *__able_await_state) waitChLocked() chan struct{} {
` + g.taskOwnedAwaitSignal() + `
	if s.waitCh == nil {
		s.waitCh = make(chan struct{}, 1)
	}
	return s.waitCh
}

func (s *__able_await_state) signalWaker(waker runtime.Value) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.waker != waker || !s.waiting {
		return false
	}
	s.wakePending = true
	ch := s.waitChLocked()
	select {
	case ch <- struct{}{}:
	default:
	}
	return true
}

func (s *__able_await_state) signal() bool {
	return s.signalWaker(s.waker)
}

func (s *__able_await_state) drainSignalLocked() {
	ch := s.waitChLocked()
	select {
	case <-ch:
	default:
	}
}
`
}

func (g *generator) awaitSignalDrain() string {
	if !g.executionContextsEnabled() {
		return ""
	}
	return "\ts.drainSignalLocked()\n"
}

func (g *generator) legacyAwaitWakeBeforePayload() string {
	if !g.executionContextsEnabled() {
		return "\t\t\tstate.markWakePending()\n"
	}
	return "\t\t\tif !state.signalWaker(inst) {\n\t\t\t\treturn runtime.NilValue{}, nil\n\t\t\t}\n"
}

func (g *generator) awaitWakerContextSpacing() string {
	if g.executionContextsEnabled() {
		return "\n"
	}
	return ""
}

func (g *generator) legacyAwaitWakeAfterPayload() string {
	if g.executionContextsEnabled() {
		return ""
	}
	return "\t\t\tstate.signal()\n"
}

func (g *generator) awaitWakePendingMethod() string {
	if g.executionContextsEnabled() {
		return ""
	}
	return `
func (s *__able_await_state) markWakePending() {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.wakePending = true
	s.mu.Unlock()
}
`
}
