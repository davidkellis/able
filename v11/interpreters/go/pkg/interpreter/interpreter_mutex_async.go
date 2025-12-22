package interpreter

import (
	"fmt"
	"math/big"
	"sync"

	"able/interpreter-go/pkg/runtime"
)

type mutexAwaitRegistration struct {
	state  *mutexState
	waker  runtime.Value
	env    *runtime.Environment
	interp *Interpreter

	mu        sync.Mutex
	cancelled bool
}

func (r *mutexAwaitRegistration) cancel() {
	if r == nil || r.state == nil {
		return
	}
	r.mu.Lock()
	if r.cancelled {
		r.mu.Unlock()
		return
	}
	r.cancelled = true
	r.mu.Unlock()

	r.state.mu.Lock()
	if r.state.awaitWaiters != nil {
		delete(r.state.awaitWaiters, r)
	}
	r.state.mu.Unlock()
}

func (r *mutexAwaitRegistration) trigger() {
	if r == nil || r.interp == nil || r.waker == nil {
		return
	}
	r.mu.Lock()
	if r.cancelled {
		r.mu.Unlock()
		return
	}
	r.cancelled = true
	r.mu.Unlock()
	r.interp.invokeAwaitWaker(r.waker, r.env)
}

func (i *Interpreter) addMutexAwaiter(state *mutexState, waker runtime.Value, env *runtime.Environment) *mutexAwaitRegistration {
	if state == nil {
		return nil
	}
	reg := &mutexAwaitRegistration{
		state:  state,
		waker:  waker,
		env:    env,
		interp: i,
	}
	state.mu.Lock()
	if state.awaitWaiters == nil {
		state.awaitWaiters = make(map[*mutexAwaitRegistration]struct{})
	}
	state.awaitWaiters[reg] = struct{}{}
	state.mu.Unlock()
	return reg
}

func (i *Interpreter) notifyMutexAwaiters(state *mutexState) {
	if state == nil {
		return
	}
	regs := make([]*mutexAwaitRegistration, 0, len(state.awaitWaiters))
	for reg := range state.awaitWaiters {
		regs = append(regs, reg)
	}
	state.awaitWaiters = make(map[*mutexAwaitRegistration]struct{})
	for _, reg := range regs {
		reg.trigger()
	}
}

type mutexAwaitable struct {
	interp       *Interpreter
	handle       int64
	callback     runtime.Value
	registration *mutexAwaitRegistration
}

func (a *mutexAwaitable) toStruct() *runtime.StructInstanceValue {
	inst := &runtime.StructInstanceValue{
		Fields: make(map[string]runtime.Value),
	}
	isReady := runtime.NativeFunctionValue{
		Name:  "Awaitable.is_ready",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			if a == nil || a.interp == nil {
				return runtime.BoolValue{Val: false}, nil
			}
			state, err := a.interp.mutexStateFromHandle(a.handle)
			if err != nil {
				return nil, err
			}
			state.mu.Lock()
			ready := !state.locked
			state.mu.Unlock()
			return runtime.BoolValue{Val: ready}, nil
		},
	}
	register := runtime.NativeFunctionValue{
		Name:  "Awaitable.register",
		Arity: 1,
		Impl: func(callCtx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if a == nil || a.interp == nil {
				return nil, fmt.Errorf("awaitable not initialized")
			}
			state, err := a.interp.mutexStateFromHandle(a.handle)
			if err != nil {
				return nil, err
			}
			waker := args[len(args)-1]
			state.mu.Lock()
			if !state.locked {
				state.mu.Unlock()
				a.interp.invokeAwaitWaker(waker, callCtx.Env)
				a.registration = nil
				return a.interp.makeAwaitRegistrationValue(nil), nil
			}
			state.mu.Unlock()
			reg := a.interp.addMutexAwaiter(state, waker, callCtx.Env)
			a.registration = reg
			cancelFn := func() {
				if reg != nil {
					reg.cancel()
				}
				a.registration = nil
			}
			return a.interp.makeAwaitRegistrationValue(cancelFn), nil
		},
	}
	commit := runtime.NativeFunctionValue{
		Name:  "Awaitable.commit",
		Arity: 0,
		Impl: func(callCtx *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			if a == nil || a.interp == nil {
				return nil, fmt.Errorf("awaitable not initialized")
			}
			if a.registration != nil {
				a.registration.cancel()
				a.registration = nil
			}
			lockVal := runtime.IntegerValue{Val: big.NewInt(a.handle), TypeSuffix: runtime.IntegerI64}
			lockFn, err := a.interp.global.Get("__able_mutex_lock")
			if err != nil {
				return nil, err
			}
			if _, err := a.interp.callCallableValue(lockFn, []runtime.Value{lockVal}, callCtx.Env, nil); err != nil {
				return nil, err
			}
			if a.callback == nil {
				return runtime.NilValue{}, nil
			}
			return a.interp.callCallableValue(a.callback, nil, callCtx.Env, nil)
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

func (i *Interpreter) makeMutexAwaitable(handle int64, callback runtime.Value) runtime.Value {
	awaitable := &mutexAwaitable{
		interp:   i,
		handle:   handle,
		callback: callback,
	}
	return awaitable.toStruct()
}
