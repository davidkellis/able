package interpreter

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

type channelState struct {
	capacity int
	ch       chan runtime.Value

	mu     sync.Mutex
	closed bool
}

type mutexState struct {
	mu      sync.Mutex
	cond    *sync.Cond
	locked  bool
	owner   *runtime.ProcHandleValue
	waiters int
}

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

	int64FromValue := func(val runtime.Value, label string) (int64, error) {
		switch v := val.(type) {
		case runtime.IntegerValue:
			if !v.Val.IsInt64() {
				return 0, fmt.Errorf("%s must fit in 64-bit integer", label)
			}
			return v.Val.Int64(), nil
		case *runtime.IntegerValue:
			if v == nil || v.Val == nil {
				return 0, fmt.Errorf("%s is nil", label)
			}
			if !v.Val.IsInt64() {
				return 0, fmt.Errorf("%s must fit in 64-bit integer", label)
			}
			return v.Val.Int64(), nil
		default:
			return 0, fmt.Errorf("%s must be an integer", label)
		}
	}

	getChannel := func(handle int64) (*channelState, error) {
		if handle <= 0 {
			return nil, fmt.Errorf("channel handle must be positive")
		}
		i.channelMu.Lock()
		state, ok := i.channels[handle]
		i.channelMu.Unlock()
		if !ok {
			return nil, fmt.Errorf("unknown channel handle %d", handle)
		}
		return state, nil
	}

	getMutex := func(handle int64) (*mutexState, error) {
		if handle <= 0 {
			return nil, fmt.Errorf("mutex handle must be positive")
		}
		i.mutexMu.Lock()
		state, ok := i.mutexes[handle]
		i.mutexMu.Unlock()
		if !ok {
			return nil, fmt.Errorf("unknown mutex handle %d", handle)
		}
		return state, nil
	}

	contextFromCall := func(callCtx *runtime.NativeCallContext) context.Context {
		if callCtx == nil {
			return context.Background()
		}
		if payload := payloadFromState(callCtx.State); payload != nil {
			if payload.handle != nil {
				if ctx := payload.handle.Context(); ctx != nil {
					return ctx
				}
			}
		}
		return context.Background()
	}

	getProcHandle := func(callCtx *runtime.NativeCallContext) *runtime.ProcHandleValue {
		if callCtx == nil {
			return nil
		}
		if payload := payloadFromState(callCtx.State); payload != nil {
			return payload.handle
		}
		return nil
	}

	var blockingExec interface {
		MarkBlocked(*runtime.ProcHandleValue)
		MarkUnblocked(*runtime.ProcHandleValue)
	}
	if exec, ok := i.executor.(interface {
		MarkBlocked(*runtime.ProcHandleValue)
		MarkUnblocked(*runtime.ProcHandleValue)
	}); ok {
		blockingExec = exec
	}

	markBlocked := func(handle *runtime.ProcHandleValue) {
		if blockingExec != nil && handle != nil {
			blockingExec.MarkBlocked(handle)
		}
	}
	markUnblocked := func(handle *runtime.ProcHandleValue) {
		if blockingExec != nil && handle != nil {
			blockingExec.MarkUnblocked(handle)
		}
	}

	blockOnNilChannel := func(callCtx *runtime.NativeCallContext) (runtime.Value, error) {
		if callCtx == nil {
			return nil, fmt.Errorf("channel operation on nil handle outside proc context")
		}
		handle := getProcHandle(callCtx)
		if handle == nil {
			return nil, fmt.Errorf("channel operation on nil handle outside proc context")
		}
		ctx := contextFromCall(callCtx)
		markBlocked(handle)
		defer markUnblocked(handle)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	channelErrorStruct := func(name string) (*runtime.StructDefinitionValue, error) {
		if def, ok := i.channelErrorStructs[name]; ok && def != nil {
			return def, nil
		}
		candidates := []string{
			name,
			"concurrency." + name,
			"able." + name,
			"able.concurrency." + name,
			"able.concurrency.channel." + name,
		}
		for _, key := range candidates {
			if key == "" {
				continue
			}
			if val, err := i.global.Get(key); err == nil {
				if def, conv := toStructDefinitionValue(val, key); conv == nil {
					i.channelErrorStructs[name] = def
					return def, nil
				}
			}
		}
		for _, bucket := range i.packageRegistry {
			if val, ok := bucket[name]; ok {
				if def, conv := toStructDefinitionValue(val, name); conv == nil {
					i.channelErrorStructs[name] = def
					return def, nil
				}
			}
		}
		placeholder := ast.NewStructDefinition(ast.NewIdentifier(name), nil, ast.StructKindNamed, nil, nil, false)
		def := &runtime.StructDefinitionValue{Node: placeholder}
		i.channelErrorStructs[name] = def
		return def, nil
	}

	makeChannelErrorValue := func(name, message string) runtime.ErrorValue {
		def, err := channelErrorStruct(name)
		if err != nil {
			return runtime.ErrorValue{Message: message}
		}
		payload := map[string]runtime.Value{
			"value": &runtime.StructInstanceValue{Definition: def},
		}
		return runtime.ErrorValue{Message: message, Payload: payload}
	}

	channelError := func(name, message string) error {
		errVal := makeChannelErrorValue(name, message)
		return raiseSignal{value: errVal}
	}

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
			handle, err := int64FromValue(args[0], "channel handle")
			if err != nil {
				return nil, err
			}
			if handle == 0 {
				return blockOnNilChannel(callCtx)
			}
			if handle < 0 {
				return nil, fmt.Errorf("channel handle must be non-negative")
			}
			state, err := getChannel(handle)
			if err != nil {
				return nil, err
			}
			state.mu.Lock()
			if state.closed {
				state.mu.Unlock()
				return nil, channelError("ChannelSendOnClosed", "send on closed channel")
			}
			ch := state.ch
			state.mu.Unlock()

			ctx := contextFromCall(callCtx)
			select {
			case ch <- args[1]:
				return runtime.NilValue{}, nil
			default:
			}
			handleVal := getProcHandle(callCtx)
			markBlocked(handleVal)
			defer markUnblocked(handleVal)
			select {
			case ch <- args[1]:
				return runtime.NilValue{}, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	}

	channelReceive := runtime.NativeFunctionValue{
		Name:  "__able_channel_receive",
		Arity: 1,
		Impl: func(callCtx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_channel_receive expects handle argument")
			}
			handle, err := int64FromValue(args[0], "channel handle")
			if err != nil {
				return nil, err
			}
			if handle == 0 {
				return blockOnNilChannel(callCtx)
			}
			if handle < 0 {
				return nil, fmt.Errorf("channel handle must be non-negative")
			}
			state, err := getChannel(handle)
			if err != nil {
				return nil, err
			}
			state.mu.Lock()
			ch := state.ch
			state.mu.Unlock()

			ctx := contextFromCall(callCtx)
			select {
			case value, ok := <-ch:
				if !ok || value == nil {
					return runtime.NilValue{}, nil
				}
				return value, nil
			default:
			}
			handleVal := getProcHandle(callCtx)
			markBlocked(handleVal)
			defer markUnblocked(handleVal)
			select {
			case value, ok := <-ch:
				if !ok {
					return runtime.NilValue{}, nil
				}
				if value == nil {
					return runtime.NilValue{}, nil
				}
				return value, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
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
			state, err := getChannel(handle)
			if err != nil {
				return nil, err
			}
			state.mu.Lock()
			if state.closed {
				state.mu.Unlock()
				return nil, channelError("ChannelSendOnClosed", "send on closed channel")
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
			state, err := getChannel(handle)
			if err != nil {
				return nil, err
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
				return nil, channelError("ChannelNil", "close of nil channel")
			}
			if handle < 0 {
				return nil, fmt.Errorf("channel handle must be non-negative")
			}
			state, err := getChannel(handle)
			if err != nil {
				return nil, err
			}
			state.mu.Lock()
			defer state.mu.Unlock()
			if state.closed {
				return nil, channelError("ChannelClosed", "close of closed channel")
			}
			state.closed = true
			close(state.ch)
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
			state, err := getChannel(handle)
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
			state, err := getMutex(handle)
			if err != nil {
				return nil, err
			}
			if state.cond == nil {
				state.cond = sync.NewCond(&state.mu)
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
			state, err := getMutex(handle)
			if err != nil {
				return nil, err
			}
			if state.cond == nil {
				state.cond = sync.NewCond(&state.mu)
			}
			state.mu.Lock()
			if !state.locked {
				state.mu.Unlock()
				return runtime.NilValue{}, nil
			}
			state.locked = false
			state.owner = nil
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
	i.global.Define("__able_channel_close", channelClose)
	i.global.Define("__able_channel_is_closed", channelIsClosed)
	i.global.Define("__able_mutex_new", mutexNew)
	i.global.Define("__able_mutex_lock", mutexLock)
	i.global.Define("__able_mutex_unlock", mutexUnlock)

	i.channelMutexReady = true
}
