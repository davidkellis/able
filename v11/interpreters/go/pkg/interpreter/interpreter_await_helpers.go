package interpreter

import (
	"fmt"
	"math"
	"math/big"
	"sync"
	"time"

	"able/interpreter10-go/pkg/runtime"
)

const maxSleepMilliseconds = int64(2_147_483_647)

func durationFromValue(val runtime.Value) (time.Duration, error) {
	switch v := val.(type) {
	case runtime.IntegerValue:
		limit := big.NewInt(maxSleepMilliseconds)
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
		if ms > float64(maxSleepMilliseconds) {
			ms = float64(maxSleepMilliseconds)
		}
		return time.Duration(ms) * time.Millisecond, nil
	default:
		return 0, fmt.Errorf("sleep_ms expects a numeric duration")
	}
}

func (i *Interpreter) makeDefaultAwaitable(callback runtime.Value) runtime.Value {
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
			return i.makeAwaitRegistrationValue(nil), nil
		},
	}
	commit := runtime.NativeFunctionValue{
		Name:  "Awaitable.commit",
		Arity: 0,
		Impl: func(callCtx *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			if callback == nil {
				return runtime.NilValue{}, nil
			}
			return i.callCallableValue(callback, nil, callCtx.Env, nil)
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

type timerAwaitable struct {
	interp    *Interpreter
	deadline  time.Time
	callback  runtime.Value
	mu        sync.Mutex
	ready     bool
	cancelled bool
	timer     *time.Timer
}

func newTimerAwaitable(interp *Interpreter, duration time.Duration, callback runtime.Value) *timerAwaitable {
	return &timerAwaitable{
		interp:   interp,
		deadline: time.Now().Add(duration),
		callback: callback,
	}
}

func (a *timerAwaitable) markReadyLocked() {
	a.ready = true
	if a.timer != nil {
		a.timer.Stop()
		a.timer = nil
	}
}

func (a *timerAwaitable) isReady() bool {
	if a == nil || a.interp == nil {
		return false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.ready {
		return true
	}
	if time.Now().After(a.deadline) || time.Now().Equal(a.deadline) {
		a.markReadyLocked()
		return true
	}
	return false
}

func (a *timerAwaitable) register(callCtx *runtime.NativeCallContext, waker runtime.Value) (runtime.Value, error) {
	if a == nil || a.interp == nil {
		return nil, fmt.Errorf("awaitable not initialized")
	}
	if callCtx == nil || callCtx.Env == nil {
		return nil, fmt.Errorf("register expects an environment")
	}
	env := callCtx.Env
	if a.isReady() {
		a.interp.invokeAwaitWaker(waker, env)
		return a.interp.makeAwaitRegistrationValue(nil), nil
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
		a.interp.invokeAwaitWaker(waker, env)
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
	return a.interp.makeAwaitRegistrationValue(cancelFn), nil
}

func (a *timerAwaitable) commit(callCtx *runtime.NativeCallContext) (runtime.Value, error) {
	if a == nil || a.interp == nil {
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
	return a.interp.callCallableValue(callback, nil, callCtx.Env, nil)
}

func (a *timerAwaitable) toStruct() *runtime.StructInstanceValue {
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
