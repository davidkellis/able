package interpreter

import (
	"context"
	"fmt"
	"sync"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type awaitArmState struct {
	awaitable    runtime.Value
	isDefault    bool
	registration runtime.Value
}

type awaitEvalState struct {
	mu          sync.Mutex
	env         *runtime.Environment
	arms        []*awaitArmState
	defaultArm  *awaitArmState
	waiting     bool
	wakePending bool
	waitCh      chan struct{}
	payload     *asyncContextPayload
	waker       runtime.Value
}

func (s *awaitEvalState) ensureWaitCh() chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.waitCh == nil {
		s.waitCh = make(chan struct{}, 1)
	}
	return s.waitCh
}

func (s *awaitEvalState) signal() {
	ch := s.ensureWaitCh()
	select {
	case ch <- struct{}{}:
	default:
	}
}

func (s *awaitEvalState) consumeWakePending() bool {
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
	return true
}

// beginWaiting marks the state before registering its arms. An awaitable is
// allowed to wake synchronously during registration, so publishing this state
// afterward would erase that wake and leave the task parked forever.
func (s *awaitEvalState) beginWaiting() bool {
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

func (s *awaitEvalState) clearWaiting() {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.waiting = false
	s.wakePending = false
	s.mu.Unlock()
}

func (s *awaitEvalState) markWakePending() {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.wakePending = true
	s.mu.Unlock()
}

func (p *asyncContextPayload) getAwaitState(expr *ast.AwaitExpression) *awaitEvalState {
	if p == nil || expr == nil {
		return nil
	}
	if p.awaitStates == nil {
		return nil
	}
	return p.awaitStates[expr]
}

func (p *asyncContextPayload) setAwaitState(expr *ast.AwaitExpression, state *awaitEvalState) {
	if p == nil || expr == nil || state == nil {
		return
	}
	if p.awaitStates == nil {
		p.awaitStates = make(map[*ast.AwaitExpression]*awaitEvalState)
	}
	p.awaitStates[expr] = state
}

func (p *asyncContextPayload) clearAwaitState(expr *ast.AwaitExpression) {
	if p == nil || expr == nil {
		return
	}
	if p.awaitStates == nil {
		return
	}
	delete(p.awaitStates, expr)
}

func payloadFromEnv(env *runtime.Environment) (*asyncContextPayload, error) {
	if env == nil {
		return nil, fmt.Errorf("await expressions must run inside an asynchronous task")
	}
	if data := env.RuntimeData(); data != nil {
		if payload, ok := data.(*asyncContextPayload); ok && payload != nil {
			return payload, nil
		}
	}
	return nil, fmt.Errorf("await expressions must run inside an asynchronous task")
}

func (i *Interpreter) evaluateAwaitExpression(expr *ast.AwaitExpression, env *runtime.Environment) (runtime.Value, error) {
	payload, err := payloadFromEnv(env)
	if err != nil {
		return nil, err
	}
	if payload.kind != asyncContextFuture {
		return nil, fmt.Errorf("await expressions must run inside an asynchronous task")
	}

	state := payload.getAwaitState(expr)
	if state == nil {
		state, err = i.initializeAwaitState(payload, expr, env)
		if err != nil {
			return nil, err
		}
		payload.setAwaitState(expr, state)
	}
	return i.awaitWithState(payload, expr, state, env)
}

func (i *Interpreter) initializeAwaitState(payload *asyncContextPayload, expr *ast.AwaitExpression, env *runtime.Environment) (*awaitEvalState, error) {
	iterable, err := i.evaluateExpression(expr.Expression, env)
	if err != nil {
		return nil, err
	}
	return i.initializeAwaitStateWithIterable(payload, expr, iterable, env)
}

// AwaitIterable evaluates an await expression against a precomputed iterable.
func (i *Interpreter) AwaitIterable(expr *ast.AwaitExpression, iterable runtime.Value, env *runtime.Environment) (runtime.Value, error) {
	payload, err := payloadFromEnv(env)
	if err != nil {
		return nil, err
	}
	if payload.kind != asyncContextFuture {
		return nil, fmt.Errorf("await expressions must run inside an asynchronous task")
	}
	state := payload.getAwaitState(expr)
	if state == nil {
		state, err = i.initializeAwaitStateWithIterable(payload, expr, iterable, env)
		if err != nil {
			return nil, err
		}
		payload.setAwaitState(expr, state)
	}
	return i.awaitWithState(payload, expr, state, env)
}

func (i *Interpreter) initializeAwaitStateWithIterable(payload *asyncContextPayload, expr *ast.AwaitExpression, iterable runtime.Value, env *runtime.Environment) (*awaitEvalState, error) {
	arms, err := i.collectAwaitArms(iterable, env)
	if err != nil {
		return nil, err
	}
	if len(arms) == 0 {
		return nil, fmt.Errorf("await requires at least one arm")
	}
	var defaultArm *awaitArmState
	for _, arm := range arms {
		if arm != nil && arm.isDefault {
			if defaultArm != nil {
				return nil, fmt.Errorf("await accepts at most one default arm")
			}
			defaultArm = arm
		}
	}
	state := &awaitEvalState{
		env:        env,
		arms:       arms,
		defaultArm: defaultArm,
		payload:    payload,
	}
	state.ensureWaitCh()
	i.ensureConcurrencyBuiltins()
	if i.awaitWakerStruct == nil {
		return nil, fmt.Errorf("Await waker builtins are not initialized")
	}
	waker, err := i.makeAwaitWaker(payload, state)
	if err != nil {
		return nil, err
	}
	state.waker = waker
	return state, nil
}

func (i *Interpreter) awaitWithState(payload *asyncContextPayload, expr *ast.AwaitExpression, state *awaitEvalState, env *runtime.Environment) (runtime.Value, error) {
	for {
		winner, err := i.selectReadyAwaitArm(state, env)
		if err != nil {
			return nil, err
		}
		if winner != nil {
			return i.completeAwait(payload, expr, state, winner, env)
		}
		if state.defaultArm != nil {
			return i.completeAwait(payload, expr, state, state.defaultArm, env)
		}
		if payload.handle != nil && payload.handle.CancelRequested() {
			i.cleanupAwaitState(payload, expr, state, env)
			return nil, context.Canceled
		}
		if state.consumeWakePending() {
			i.clearAwaitRegistrations(state, env)
			continue
		}
		if state.beginWaiting() {
			if err := i.registerAwaitState(state, env); err != nil {
				state.clearWaiting()
				return nil, err
			}
		}

		waitCh := state.ensureWaitCh()
		payload.setAwaitBlocked(true)

		if _, ok := i.executor.(*SerialExecutor); ok {
			if payload != nil && payload.compiled && payload.compiledYield != nil && payload.compiledResume != nil {
				payload.compiledYield <- compiledYield{}
				<-payload.compiledResume
				continue
			}
			return nil, errSerialYield
		}

		var handle *runtime.FutureValue
		if payload != nil {
			handle = payload.handle
		}
		i.markBlocked(handle)
		ctx := payload.handle.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		select {
		case <-waitCh:
		case <-ctx.Done():
			i.markUnblocked(handle)
			payload.setAwaitBlocked(false)
			i.cleanupAwaitState(payload, expr, state, env)
			return nil, ctx.Err()
		}
		i.markUnblocked(handle)
		payload.setAwaitBlocked(false)
		i.clearAwaitRegistrations(state, env)
		state.clearWaiting()
	}
}

func (i *Interpreter) collectAwaitArms(iterable runtime.Value, env *runtime.Environment) ([]*awaitArmState, error) {
	if arr, err := i.toArrayValue(iterable); err == nil {
		state, err := i.ensureArrayState(arr, 0)
		if err != nil {
			return nil, err
		}
		arms := make([]*awaitArmState, 0, len(state.Values))
		for _, el := range state.Values {
			arms = append(arms, &awaitArmState{
				awaitable: el,
				isDefault: i.awaitArmIsDefault(el, env),
			})
		}
		return arms, nil
	}
	iter, err := i.resolveIteratorValue(iterable, env)
	if err != nil {
		return nil, fmt.Errorf("await requires an Iterable of Awaitable values: %w", err)
	}
	defer iter.Close()
	arms := make([]*awaitArmState, 0)
	for {
		val, done, stepErr := iter.Next()
		if stepErr != nil {
			return nil, stepErr
		}
		if done {
			break
		}
		arms = append(arms, &awaitArmState{
			awaitable: val,
			isDefault: i.awaitArmIsDefault(val, env),
		})
	}
	return arms, nil
}

func (i *Interpreter) awaitArmIsDefault(awaitable runtime.Value, env *runtime.Environment) bool {
	member, err := i.memberAccessOnValue(awaitable, ast.NewIdentifier("is_default"), env)
	if err != nil {
		return false
	}
	result, err := i.callCallableValue(member, nil, env, nil)
	if err != nil {
		return false
	}
	return i.isTruthy(result)
}

func (i *Interpreter) selectReadyAwaitArm(state *awaitEvalState, env *runtime.Environment) (*awaitArmState, error) {
	ready := make([]*awaitArmState, 0)
	for _, arm := range state.arms {
		if arm == nil || arm.isDefault {
			continue
		}
		result, err := i.invokeAwaitableMethod(arm.awaitable, "is_ready", nil, env)
		if err != nil {
			return nil, err
		}
		if i.isTruthy(result) {
			ready = append(ready, arm)
		}
	}
	if len(ready) == 0 {
		return nil, nil
	}
	start := 0
	if len(ready) > 0 {
		start = i.awaitRoundRobinIndex % len(ready)
		i.awaitRoundRobinIndex = (i.awaitRoundRobinIndex + 1) % len(ready)
	}
	return ready[start], nil
}

func (i *Interpreter) registerAwaitState(state *awaitEvalState, env *runtime.Environment) error {
	if state.waker == nil {
		return fmt.Errorf("Await waker not initialised")
	}
	for _, arm := range state.arms {
		if arm == nil || arm.isDefault {
			continue
		}
		if arm.registration != nil {
			continue
		}
		reg, err := i.invokeAwaitableMethod(arm.awaitable, "register", []runtime.Value{state.waker}, env)
		if err != nil {
			return err
		}
		arm.registration = reg
	}
	return nil
}

func (i *Interpreter) completeAwait(payload *asyncContextPayload, expr *ast.AwaitExpression, state *awaitEvalState, winner *awaitArmState, env *runtime.Environment) (runtime.Value, error) {
	for _, arm := range state.arms {
		if arm == nil || arm == winner {
			continue
		}
		i.cancelAwaitRegistration(arm.registration, env)
		arm.registration = nil
	}
	result, err := i.invokeAwaitableMethod(winner.awaitable, "commit", nil, env)
	if err != nil {
		return nil, err
	}
	i.cleanupAwaitState(payload, expr, state, env)
	payload.setAwaitBlocked(false)
	if result == nil {
		return runtime.NilValue{}, nil
	}
	return result, nil
}

func (i *Interpreter) cleanupAwaitState(payload *asyncContextPayload, expr *ast.AwaitExpression, state *awaitEvalState, env *runtime.Environment) {
	i.clearAwaitRegistrations(state, env)
	state.clearWaiting()
	if payload != nil {
		payload.setAwaitBlocked(false)
		payload.clearAwaitState(expr)
	}
}

// clearAwaitRegistrations discards registrations after a wake as well as at
// final cleanup. A wake only says that the awaitable should be checked again;
// another task can win it before this task commits, in which case its old
// registration cannot be reused for the next wait cycle.
func (i *Interpreter) clearAwaitRegistrations(state *awaitEvalState, env *runtime.Environment) {
	if state == nil {
		return
	}
	for _, arm := range state.arms {
		if arm == nil {
			continue
		}
		i.cancelAwaitRegistration(arm.registration, env)
		arm.registration = nil
	}
}

func (i *Interpreter) cancelAwaitRegistration(reg runtime.Value, env *runtime.Environment) {
	if reg == nil {
		return
	}
	member, err := i.memberAccessOnValue(reg, ast.NewIdentifier("cancel"), env)
	if err != nil {
		return
	}
	if _, err := i.callCallableValue(member, nil, env, nil); err != nil {
		return
	}
}

func (i *Interpreter) invokeAwaitableMethod(awaitable runtime.Value, method string, args []runtime.Value, env *runtime.Environment) (runtime.Value, error) {
	member, err := i.memberAccessOnValue(awaitable, ast.NewIdentifier(method), env)
	if err != nil {
		return nil, err
	}
	return i.callCallableValue(member, args, env, nil)
}

func (i *Interpreter) makeAwaitWaker(payload *asyncContextPayload, state *awaitEvalState) (*runtime.StructInstanceValue, error) {
	if i.awaitWakerStruct == nil {
		return nil, fmt.Errorf("Await waker builtins are not initialized")
	}
	inst := &runtime.StructInstanceValue{
		Definition: i.awaitWakerStruct,
		Fields:     make(map[string]runtime.Value),
	}
	wakeFn := runtime.NativeFunctionValue{
		Name:  "AwaitWaker.wake",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			state.markWakePending()
			if payload != nil {
				payload.setAwaitBlocked(false)
			}
			state.signal()
			if payload != nil && payload.resume != nil {
				payload.resume()
			}
			return runtime.NilValue{}, nil
		},
	}
	inst.Fields["wake"] = &runtime.NativeBoundMethodValue{Receiver: inst, Method: wakeFn}
	return inst, nil
}
