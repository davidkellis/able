package interpreter

import (
	"context"
	"fmt"
	"sync"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
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
		return nil, fmt.Errorf("await expressions must run inside a proc")
	}
	if data := env.RuntimeData(); data != nil {
		if payload, ok := data.(*asyncContextPayload); ok && payload != nil {
			return payload, nil
		}
	}
	return nil, fmt.Errorf("await expressions must run inside a proc")
}

func (i *Interpreter) evaluateAwaitExpression(expr *ast.AwaitExpression, env *runtime.Environment) (runtime.Value, error) {
	payload, err := payloadFromEnv(env)
	if err != nil {
		return nil, err
	}
	if payload.kind != asyncContextProc {
		return nil, fmt.Errorf("await expressions must run inside a proc")
	}

	state := payload.getAwaitState(expr)
	if state == nil {
		state, err = i.initializeAwaitState(payload, expr, env)
		if err != nil {
			return nil, err
		}
		payload.setAwaitState(expr, state)
	}

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
		if state.wakePending {
			state.waiting = false
			state.wakePending = false
			continue
		}
		if !state.waiting {
			if err := i.registerAwaitState(state, env); err != nil {
				return nil, err
			}
			state.waiting = true
			state.wakePending = false
		}

		waitCh := state.ensureWaitCh()
		payload.awaitBlocked = true

		if _, ok := i.executor.(*SerialExecutor); ok {
			return nil, errSerialYield
		}

		ctx := payload.handle.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		select {
		case <-waitCh:
		case <-ctx.Done():
			payload.awaitBlocked = false
			i.cleanupAwaitState(payload, expr, state, env)
			return nil, ctx.Err()
		}
		payload.awaitBlocked = false
		state.waiting = false
		state.wakePending = false
	}
}

func (i *Interpreter) initializeAwaitState(payload *asyncContextPayload, expr *ast.AwaitExpression, env *runtime.Environment) (*awaitEvalState, error) {
	iterable, err := i.evaluateExpression(expr.Expression, env)
	if err != nil {
		return nil, err
	}
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

func (i *Interpreter) collectAwaitArms(iterable runtime.Value, env *runtime.Environment) ([]*awaitArmState, error) {
	arr, err := i.toArrayValue(iterable)
	if err != nil {
		return nil, fmt.Errorf("await currently expects an array of Awaitable values")
	}
	state, err := i.ensureArrayState(arr, 0)
	if err != nil {
		return nil, err
	}
	arms := make([]*awaitArmState, 0, len(state.values))
	for _, el := range state.values {
		arms = append(arms, &awaitArmState{
			awaitable: el,
			isDefault: i.awaitArmIsDefault(el, env),
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
	return isTruthy(result)
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
		if isTruthy(result) {
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
		if arm == nil || arm.isDefault || arm.registration != nil {
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
	payload.awaitBlocked = false
	if result == nil {
		return runtime.NilValue{}, nil
	}
	return result, nil
}

func (i *Interpreter) cleanupAwaitState(payload *asyncContextPayload, expr *ast.AwaitExpression, state *awaitEvalState, env *runtime.Environment) {
	for _, arm := range state.arms {
		if arm == nil {
			continue
		}
		i.cancelAwaitRegistration(arm.registration, env)
		arm.registration = nil
	}
	state.waiting = false
	state.wakePending = false
	if payload != nil {
		payload.awaitBlocked = false
		payload.clearAwaitState(expr)
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
			if payload != nil && payload.resume != nil {
				payload.resume()
			}
			return runtime.NilValue{}, nil
		},
	}
	inst.Fields["wake"] = &runtime.NativeBoundMethodValue{Receiver: inst, Method: wakeFn}
	return inst, nil
}
