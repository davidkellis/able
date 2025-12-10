package interpreter

import (
	"fmt"
	"math/big"
	"sync"

	"able/interpreter10-go/pkg/runtime"
)

type channelAwaitKind int

const (
	channelAwaitSend channelAwaitKind = iota
	channelAwaitRecv
)

type channelAwaitRegistration struct {
	state  *channelState
	kind   channelAwaitKind
	waker  runtime.Value
	env    *runtime.Environment
	interp *Interpreter

	mu        sync.Mutex
	cancelled bool
}

type channelAwaitOperation int

const (
	channelAwaitOpReceive channelAwaitOperation = iota
	channelAwaitOpSend
)

type channelAwaitable struct {
	interp       *Interpreter
	handle       int64
	op           channelAwaitOperation
	payload      runtime.Value
	callback     runtime.Value
	registration *channelAwaitRegistration
}

func (r *channelAwaitRegistration) cancel() {
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
	switch r.kind {
	case channelAwaitSend:
		delete(r.state.sendAwaiters, r)
	case channelAwaitRecv:
		delete(r.state.recvAwaiters, r)
	}
	r.state.mu.Unlock()
}

func (r *channelAwaitRegistration) trigger() {
	if r == nil || r.interp == nil || r.waker == nil {
		return
	}
	r.mu.Lock()
	if r.cancelled {
		r.mu.Unlock()
		return
	}
	r.mu.Unlock()
	r.interp.invokeAwaitWaker(r.waker, r.env)
}

func (i *Interpreter) addChannelAwaiter(state *channelState, kind channelAwaitKind, waker runtime.Value, env *runtime.Environment) *channelAwaitRegistration {
	if state == nil {
		return nil
	}
	reg := &channelAwaitRegistration{
		state:  state,
		kind:   kind,
		waker:  waker,
		env:    env,
		interp: i,
	}
	state.mu.Lock()
	switch kind {
	case channelAwaitSend:
		if state.sendAwaiters == nil {
			state.sendAwaiters = make(map[*channelAwaitRegistration]struct{})
		}
		state.sendAwaiters[reg] = struct{}{}
	case channelAwaitRecv:
		if state.recvAwaiters == nil {
			state.recvAwaiters = make(map[*channelAwaitRegistration]struct{})
		}
		state.recvAwaiters[reg] = struct{}{}
	}
	state.mu.Unlock()
	return reg
}

func (i *Interpreter) notifyChannelAwaiters(state *channelState, kind channelAwaitKind) {
	if state == nil {
		return
	}
	state.mu.Lock()
	var regs []*channelAwaitRegistration
	switch kind {
	case channelAwaitSend:
		for reg := range state.sendAwaiters {
			regs = append(regs, reg)
		}
	case channelAwaitRecv:
		for reg := range state.recvAwaiters {
			regs = append(regs, reg)
		}
	}
	state.mu.Unlock()
	for _, reg := range regs {
		reg.trigger()
	}
}

func (a *channelAwaitable) isReady() (bool, error) {
	if a == nil || a.interp == nil {
		return false, fmt.Errorf("awaitable not initialized")
	}
	if a.handle == 0 {
		return false, nil
	}
	state, err := a.interp.channelStateFromHandle(a.handle)
	if err != nil {
		return false, err
	}
	state.mu.Lock()
	closed := state.closed
	capacity := state.capacity
	sendWaiters := state.sendWaiters + len(state.serialSendWaiters)
	recvWaiters := state.recvWaiters + len(state.serialRecvWaiters)
	sendAwaiters := len(state.sendAwaiters)
	recvAwaiters := len(state.recvAwaiters)
	length := len(state.ch)
	if serialLen := len(state.serialQueue); serialLen > length {
		length = serialLen
	}
	state.mu.Unlock()

	switch a.op {
	case channelAwaitOpReceive:
		if length > 0 {
			return true, nil
		}
		if capacity == 0 && (sendWaiters > 0 || sendAwaiters > 0) {
			return true, nil
		}
		if closed {
			return true, nil
		}
		return false, nil
	case channelAwaitOpSend:
		if closed {
			return false, a.interp.concurrencyError("ChannelSendOnClosed", "send on closed channel")
		}
		if capacity == 0 {
			return recvWaiters > 0 || recvAwaiters > 0, nil
		}
		return length < capacity, nil
	default:
		return false, fmt.Errorf("unknown awaitable operation")
	}
}

func (a *channelAwaitable) register(callCtx *runtime.NativeCallContext, waker runtime.Value) (runtime.Value, error) {
	if a == nil || a.interp == nil {
		return nil, fmt.Errorf("awaitable not initialized")
	}
	if a.handle == 0 {
		return a.interp.makeAwaitRegistrationValue(nil), nil
	}
	state, err := a.interp.channelStateFromHandle(a.handle)
	if err != nil {
		return nil, err
	}
	if a.registration != nil {
		return a.interp.makeAwaitRegistrationValue(a.registration.cancel), nil
	}
	var reg *channelAwaitRegistration
	if a.op == channelAwaitOpReceive {
		reg = a.interp.addChannelAwaiter(state, channelAwaitRecv, waker, callCtx.Env)
	} else {
		reg = a.interp.addChannelAwaiter(state, channelAwaitSend, waker, callCtx.Env)
	}
	a.registration = reg
	cancelFn := func() {
		if a.registration != nil {
			a.registration.cancel()
			a.registration = nil
		}
	}
	return a.interp.makeAwaitRegistrationValue(cancelFn), nil
}

func (a *channelAwaitable) commit(callCtx *runtime.NativeCallContext) (runtime.Value, error) {
	if a == nil || a.interp == nil {
		return nil, fmt.Errorf("awaitable not initialized")
	}
	switch a.op {
	case channelAwaitOpReceive:
		value, err := a.interp.channelReceiveOp(callCtx, runtime.IntegerValue{
			Val:        big.NewInt(a.handle),
			TypeSuffix: runtime.IntegerI64,
		})
		if err != nil {
			return nil, err
		}
		if a.callback == nil {
			return value, nil
		}
		return a.interp.callCallableValue(a.callback, []runtime.Value{value}, callCtx.Env, nil)
	case channelAwaitOpSend:
		if _, err := a.interp.channelSendOp(callCtx, runtime.IntegerValue{
			Val:        big.NewInt(a.handle),
			TypeSuffix: runtime.IntegerI64,
		}, a.payload); err != nil {
			return nil, err
		}
		if a.callback == nil {
			return runtime.NilValue{}, nil
		}
		return a.interp.callCallableValue(a.callback, nil, callCtx.Env, nil)
	default:
		return nil, fmt.Errorf("unknown awaitable operation")
	}
}

func (a *channelAwaitable) toStruct() *runtime.StructInstanceValue {
	inst := &runtime.StructInstanceValue{
		Fields: make(map[string]runtime.Value),
	}
	isReady := runtime.NativeFunctionValue{
		Name:  "Awaitable.is_ready",
		Arity: 0,
		Impl: func(callCtx *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			if callCtx != nil {
				if payload := payloadFromState(callCtx.State); payload != nil {
					if pendingSend := a.interp.pendingSendWaiter(payload.handle); pendingSend != nil && pendingSend.delivered {
						return runtime.BoolValue{Val: true}, nil
					}
					if pendingRecv := a.interp.pendingReceiveWaiter(payload.handle); pendingRecv != nil && pendingRecv.ready {
						return runtime.BoolValue{Val: true}, nil
					}
				}
			}
			ready, err := a.isReady()
			if err != nil {
				return nil, err
			}
			return runtime.BoolValue{Val: ready}, nil
		},
	}
	register := runtime.NativeFunctionValue{
		Name:  "Awaitable.register",
		Arity: 1,
		Impl: func(callCtx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("register expects waker argument")
			}
			waker := args[len(args)-1]
			return a.register(callCtx, waker)
		},
	}
	commit := runtime.NativeFunctionValue{
		Name:  "Awaitable.commit",
		Arity: 0,
		Impl: func(callCtx *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			return a.commit(callCtx)
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
