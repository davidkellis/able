package interpreter

import (
	"fmt"
	"math/big"
	"sync"

	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) ensureChannelMutexBuiltins() {
	if i.channelMutexReady {
		return
	}
	i.initChannelMutexBuiltins()
}

func (i *Interpreter) initChannelMutexBuiltins() {
	if i.channelMutexReady {
		return
	}
	if i.channels == nil {
		i.channels = make(map[int64]*channelState)
	}
	if i.mutexes == nil {
		i.mutexes = make(map[int64]*mutexState)
	}
	if i.nextChannelHandle == 0 {
		i.nextChannelHandle = 1
	}
	if i.nextMutexHandle == 0 {
		i.nextMutexHandle = 1
	}

	makeHandleValue := func(handle int64) runtime.IntegerValue {
		return runtime.IntegerValue{
			Val:        big.NewInt(handle),
			TypeSuffix: runtime.IntegerI64,
		}
	}

	int64FromValue := i.int64FromValue
	contextFromCall := i.contextFromCall
	markBlocked := i.markBlocked
	markUnblocked := i.markUnblocked
	channelSendOp := i.channelSendOp
	channelReceiveOp := i.channelReceiveOp

	makeBool := func(value bool) runtime.BoolValue {
		return runtime.BoolValue{Val: value}
	}

	channelNew := runtime.NativeFunctionValue{
		Name:  "__able_channel_new",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_channel_new expects capacity argument")
			}
			capacity, err := int64FromValue(args[0], "channel capacity")
			if err != nil {
				return nil, err
			}
			if capacity < 0 {
				return nil, fmt.Errorf("channel capacity must be non-negative")
			}
			state := &channelState{
				capacity: int(capacity),
				ch:       make(chan runtime.Value, int(capacity)),
			}
			i.channelMu.Lock()
			handle := i.nextChannelHandle
			i.nextChannelHandle++
			i.channels[handle] = state
			i.channelMu.Unlock()
			return makeHandleValue(handle), nil
		},
	}

	channelSend := runtime.NativeFunctionValue{
		Name:  "__able_channel_send",
		Arity: 2,
		Impl: func(callCtx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__able_channel_send expects handle and value")
			}
			return channelSendOp(callCtx, args[0], args[1])
		},
	}

	channelReceive := runtime.NativeFunctionValue{
		Name:  "__able_channel_receive",
		Arity: 1,
		Impl: func(callCtx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_channel_receive expects handle argument")
			}
			return channelReceiveOp(callCtx, args[0])
		},
	}

	channelTrySend := runtime.NativeFunctionValue{
		Name:  "__able_channel_try_send",
		Arity: 2,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__able_channel_try_send expects handle and value")
			}
			handle, err := int64FromValue(args[0], "channel handle")
			if err != nil {
				return nil, err
			}
			if handle == 0 {
				return makeBool(false), nil
			}
			if handle < 0 {
				return nil, fmt.Errorf("channel handle must be non-negative")
			}
			state, err := i.channelStateFromHandle(handle)
			if err != nil {
				return nil, err
			}
			if _, serial := i.executor.(*SerialExecutor); serial {
				state.mu.Lock()
				if state.closed {
					state.mu.Unlock()
					return nil, i.concurrencyError("ChannelSendOnClosed", "send on closed channel")
				}
				if len(state.serialRecvWaiters) > 0 {
					receiver := state.serialRecvWaiters[0]
					state.serialRecvWaiters = state.serialRecvWaiters[1:]
					if receiver != nil {
						receiver.ready = true
						receiver.closed = false
						receiver.value = args[1]
						i.setPendingReceiveWaiter(receiver)
						resumePayload(receiver.payload)
					}
					state.mu.Unlock()
					i.notifyChannelAwaiters(state, channelAwaitSend)
					return makeBool(true), nil
				}
				if state.capacity > 0 && len(state.serialQueue) < state.capacity {
					state.serialQueue = append(state.serialQueue, args[1])
					state.mu.Unlock()
					i.notifyChannelAwaiters(state, channelAwaitRecv)
					return makeBool(true), nil
				}
				state.mu.Unlock()
				return makeBool(false), nil
			}
			state.mu.Lock()
			if state.closed {
				state.mu.Unlock()
				return nil, i.concurrencyError("ChannelSendOnClosed", "send on closed channel")
			}
			ch := state.ch
			state.mu.Unlock()

			select {
			case ch <- args[1]:
				return makeBool(true), nil
			default:
				return makeBool(false), nil
			}
		},
	}

	channelTryReceive := runtime.NativeFunctionValue{
		Name:  "__able_channel_try_receive",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_channel_try_receive expects handle argument")
			}
			handle, err := int64FromValue(args[0], "channel handle")
			if err != nil {
				return nil, err
			}
			if handle == 0 {
				return runtime.NilValue{}, nil
			}
			if handle < 0 {
				return nil, fmt.Errorf("channel handle must be non-negative")
			}
			state, err := i.channelStateFromHandle(handle)
			if err != nil {
				return nil, err
			}
			if _, serial := i.executor.(*SerialExecutor); serial {
				state.mu.Lock()
				if len(state.serialQueue) > 0 {
					value := state.serialQueue[0]
					state.serialQueue = state.serialQueue[1:]
					if state.capacity > 0 && len(state.serialSendWaiters) > 0 {
						sender := state.serialSendWaiters[0]
						state.serialSendWaiters = state.serialSendWaiters[1:]
						if sender != nil {
							sender.delivered = true
							state.serialQueue = append(state.serialQueue, sender.value)
							i.setPendingSendWaiter(sender)
							resumePayload(sender.payload)
						}
					}
					state.mu.Unlock()
					i.notifyChannelAwaiters(state, channelAwaitSend)
					if len(state.serialQueue) > 0 {
						i.notifyChannelAwaiters(state, channelAwaitRecv)
					}
					if value == nil {
						return runtime.NilValue{}, nil
					}
					return value, nil
				}
				if len(state.serialSendWaiters) > 0 {
					sender := state.serialSendWaiters[0]
					state.serialSendWaiters = state.serialSendWaiters[1:]
					if sender != nil {
						sender.delivered = true
						i.setPendingSendWaiter(sender)
						resumePayload(sender.payload)
						val := sender.value
						state.mu.Unlock()
						i.notifyChannelAwaiters(state, channelAwaitSend)
						if val == nil {
							return runtime.NilValue{}, nil
						}
						return val, nil
					}
				}
				if state.closed {
					state.mu.Unlock()
					return runtime.NilValue{}, nil
				}
				state.mu.Unlock()
				return runtime.NilValue{}, nil
			}
			state.mu.Lock()
			ch := state.ch
			state.mu.Unlock()

			select {
			case value, ok := <-ch:
				if !ok {
					return runtime.NilValue{}, nil
				}
				if value == nil {
					return runtime.NilValue{}, nil
				}
				return value, nil
			default:
				return runtime.NilValue{}, nil
			}
		},
	}

	channelAwaitTryRecv := runtime.NativeFunctionValue{
		Name:  "__able_channel_await_try_recv",
		Arity: 2,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__able_channel_await_try_recv expects handle and callback")
			}
			handle, err := int64FromValue(args[0], "channel handle")
			if err != nil {
				return nil, err
			}
			awaitable := &channelAwaitable{
				interp:   i,
				handle:   handle,
				op:       channelAwaitOpReceive,
				callback: args[1],
			}
			return awaitable.toStruct(), nil
		},
	}

	channelAwaitTrySend := runtime.NativeFunctionValue{
		Name:  "__able_channel_await_try_send",
		Arity: 3,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("__able_channel_await_try_send expects handle, value, and callback")
			}
			handle, err := int64FromValue(args[0], "channel handle")
			if err != nil {
				return nil, err
			}
			awaitable := &channelAwaitable{
				interp:   i,
				handle:   handle,
				op:       channelAwaitOpSend,
				payload:  args[1],
				callback: args[2],
			}
			return awaitable.toStruct(), nil
		},
	}

	channelClose := runtime.NativeFunctionValue{
		Name:  "__able_channel_close",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_channel_close expects handle argument")
			}
			handle, err := int64FromValue(args[0], "channel handle")
			if err != nil {
				return nil, err
			}
			if handle == 0 {
				return nil, i.concurrencyError("ChannelNil", "close of nil channel")
			}
			if handle < 0 {
				return nil, fmt.Errorf("channel handle must be non-negative")
			}
			state, err := i.channelStateFromHandle(handle)
			if err != nil {
				return nil, err
			}
			state.mu.Lock()
			if state.closed {
				state.mu.Unlock()
				return nil, i.concurrencyError("ChannelClosed", "close of closed channel")
			}
			state.closed = true
			close(state.ch)
			serialRecv := append([]*channelReceiveWaiter(nil), state.serialRecvWaiters...)
			serialSend := append([]*channelSendWaiter(nil), state.serialSendWaiters...)
			state.serialRecvWaiters = nil
			state.serialSendWaiters = nil
			state.mu.Unlock()

			for _, recv := range serialRecv {
				if recv == nil {
					continue
				}
				recv.ready = true
				recv.closed = true
				i.setPendingReceiveWaiter(recv)
				resumePayload(recv.payload)
			}
			for _, send := range serialSend {
				if send == nil {
					continue
				}
				send.err = i.concurrencyError("ChannelSendOnClosed", "send on closed channel")
				i.setPendingSendWaiter(send)
				resumePayload(send.payload)
			}
			i.notifyChannelAwaiters(state, channelAwaitRecv)
			i.notifyChannelAwaiters(state, channelAwaitSend)
			return runtime.NilValue{}, nil
		},
	}

	channelIsClosed := runtime.NativeFunctionValue{
		Name:  "__able_channel_is_closed",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_channel_is_closed expects handle argument")
			}
			handle, err := int64FromValue(args[0], "channel handle")
			if err != nil {
				return nil, err
			}
			if handle == 0 {
				return makeBool(false), nil
			}
			if handle < 0 {
				return nil, fmt.Errorf("channel handle must be non-negative")
			}
			state, err := i.channelStateFromHandle(handle)
			if err != nil {
				return nil, err
			}
			state.mu.Lock()
			isClosed := state.closed
			state.mu.Unlock()
			return makeBool(isClosed), nil
		},
	}

	mutexNew := runtime.NativeFunctionValue{
		Name:  "__able_mutex_new",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			state := &mutexState{}
			state.cond = sync.NewCond(&state.mu)
			i.mutexMu.Lock()
			handle := i.nextMutexHandle
			i.nextMutexHandle++
			i.mutexes[handle] = state
			i.mutexMu.Unlock()
			return makeHandleValue(handle), nil
		},
	}

	mutexLock := runtime.NativeFunctionValue{
		Name:  "__able_mutex_lock",
		Arity: 1,
		Impl: func(callCtx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_mutex_lock expects handle argument")
			}
			handle, err := int64FromValue(args[0], "mutex handle")
			if err != nil {
				return nil, err
			}
			state, err := i.mutexStateFromHandle(handle)
			if err != nil {
				return nil, err
			}
			ctx := contextFromCall(callCtx)
			var procHandle *runtime.ProcHandleValue
			if callCtx != nil {
				if payload := payloadFromState(callCtx.State); payload != nil && payload.handle != nil {
					procHandle = payload.handle
				}
			}
			var serialExec *SerialExecutor
			if procHandle != nil {
				if exec, ok := i.executor.(*SerialExecutor); ok {
					serialExec = exec
				}
			}

			state.mu.Lock()
			defer state.mu.Unlock()

			if !state.locked {
				state.locked = true
				state.owner = procHandle
				return runtime.NilValue{}, nil
			}

			registered := false
			defer func() {
				if registered {
					state.waiters--
				}
			}()
			waiting := false
			markWaiting := func() {
				if waiting {
					return
				}
				markBlocked(procHandle)
				waiting = true
			}
			clearWaiting := func() {
				if !waiting {
					return
				}
				markUnblocked(procHandle)
				waiting = false
			}
			for {
				if !state.locked {
					state.locked = true
					state.owner = procHandle
					if registered {
						state.waiters--
						registered = false
					}
					clearWaiting()
					state.cond.Signal()
					return runtime.NilValue{}, nil
				}
				if !registered {
					state.waiters++
					registered = true
				}
				markWaiting()
				if serialExec != nil {
					serialExec.suspendCurrent(procHandle)
				}
				if ctx != nil {
					select {
					case <-ctx.Done():
						clearWaiting()
						if serialExec != nil {
							serialExec.resumeCurrent(procHandle)
						}
						if registered {
							state.waiters--
							registered = false
							state.cond.Signal()
						}
						return nil, ctx.Err()
					default:
					}
				}
				state.cond.Wait()
				clearWaiting()
				if serialExec != nil {
					serialExec.resumeCurrent(procHandle)
				}
				if ctx != nil {
					select {
					case <-ctx.Done():
						clearWaiting()
						if registered {
							state.waiters--
							registered = false
							state.cond.Signal()
						}
						return nil, ctx.Err()
					default:
					}
				}
			}
		},
	}

	mutexUnlock := runtime.NativeFunctionValue{
		Name:  "__able_mutex_unlock",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_mutex_unlock expects handle argument")
			}
			handle, err := int64FromValue(args[0], "mutex handle")
			if err != nil {
				return nil, err
			}
			state, err := i.mutexStateFromHandle(handle)
			if err != nil {
				return nil, err
			}
			state.mu.Lock()
			if !state.locked {
				state.mu.Unlock()
				return nil, i.concurrencyError("MutexUnlocked", "unlock of unlocked mutex")
			}
			state.locked = false
			state.owner = nil
			i.notifyMutexAwaiters(state)
			if state.waiters > 0 {
				state.cond.Signal()
				for state.waiters > 0 && !state.locked {
					state.cond.Wait()
				}
			}
			state.mu.Unlock()
			return runtime.NilValue{}, nil
		},
	}

	i.global.Define("__able_channel_new", channelNew)
	i.global.Define("__able_channel_send", channelSend)
	i.global.Define("__able_channel_receive", channelReceive)
	i.global.Define("__able_channel_try_send", channelTrySend)
	i.global.Define("__able_channel_try_receive", channelTryReceive)
	i.global.Define("__able_channel_await_try_recv", channelAwaitTryRecv)
	i.global.Define("__able_channel_await_try_send", channelAwaitTrySend)
	i.global.Define("__able_channel_close", channelClose)
	i.global.Define("__able_channel_is_closed", channelIsClosed)
	i.global.Define("__able_mutex_new", mutexNew)
	i.global.Define("__able_mutex_lock", mutexLock)
	i.global.Define("__able_mutex_unlock", mutexUnlock)
	i.global.Define("__able_mutex_await_lock", runtime.NativeFunctionValue{
		Name:  "__able_mutex_await_lock",
		Arity: 2,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("__able_mutex_await_lock expects handle and callback")
			}
			handle, err := int64FromValue(args[0], "mutex handle")
			if err != nil {
				return nil, err
			}
			var callback runtime.Value
			if len(args) > 1 {
				callback = args[1]
			}
			return i.makeMutexAwaitable(handle, callback), nil
		},
	})

	i.channelMutexReady = true
}
