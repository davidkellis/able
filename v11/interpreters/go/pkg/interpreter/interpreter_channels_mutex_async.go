package interpreter

import (
	"context"
	"fmt"
	"sync"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type channelState struct {
	capacity int
	ch       chan runtime.Value

	mu     sync.Mutex
	closed bool

	sendWaiters int
	recvWaiters int

	serialQueue       []runtime.Value
	serialSendWaiters []*channelSendWaiter
	serialRecvWaiters []*channelReceiveWaiter

	sendAwaiters map[*channelAwaitRegistration]struct{}
	recvAwaiters map[*channelAwaitRegistration]struct{}
}

type channelSendWaiter struct {
	handle    *runtime.FutureValue
	payload   *asyncContextPayload
	state     *channelState
	value     runtime.Value
	delivered bool
	err       error
}

type channelReceiveWaiter struct {
	handle  *runtime.FutureValue
	payload *asyncContextPayload
	ready   bool
	value   runtime.Value
	closed  bool
	state   *channelState
}

type mutexState struct {
	mu           sync.Mutex
	cond         *sync.Cond
	locked       bool
	owner        *runtime.FutureValue
	waiters      int
	awaitWaiters map[*mutexAwaitRegistration]struct{}
}

func (i *Interpreter) resolveConcurrencyErrorStruct(name string) *runtime.StructDefinitionValue {
	if def, ok := i.concurrencyErrorStructs[name]; ok && def != nil {
		return def
	}
	candidates := []string{
		name,
		"concurrency." + name,
		"able." + name,
		"able.concurrency." + name,
		"able.concurrency.channel." + name,
		"able.concurrency.mutex." + name,
	}
	for _, key := range candidates {
		if key == "" {
			continue
		}
		if val, err := i.global.Get(key); err == nil {
			if def, conv := toStructDefinitionValue(val, key); conv == nil {
				i.concurrencyErrorStructs[name] = def
				return def
			}
		}
	}
	for _, bucket := range i.packageRegistry {
		if val, ok := bucket[name]; ok {
			if def, conv := toStructDefinitionValue(val, name); conv == nil {
				i.concurrencyErrorStructs[name] = def
				return def
			}
		}
	}
	placeholder := ast.NewStructDefinition(ast.NewIdentifier(name), nil, ast.StructKindNamed, nil, nil, false)
	def := &runtime.StructDefinitionValue{Node: placeholder}
	i.concurrencyErrorStructs[name] = def
	return def
}

func (i *Interpreter) makeConcurrencyErrorValue(name, message string) runtime.ErrorValue {
	def := i.resolveConcurrencyErrorStruct(name)
	if def == nil {
		return runtime.ErrorValue{Message: message}
	}
	payload := map[string]runtime.Value{
		"value": &runtime.StructInstanceValue{Definition: def},
	}
	return runtime.ErrorValue{Message: message, Payload: payload}
}

func (i *Interpreter) concurrencyError(name, message string) error {
	errVal := i.makeConcurrencyErrorValue(name, message)
	return raiseSignal{value: errVal}
}

func (i *Interpreter) int64FromValue(val runtime.Value, label string) (int64, error) {
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

func (i *Interpreter) contextFromCall(callCtx *runtime.NativeCallContext) context.Context {
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

func (i *Interpreter) getFutureHandle(callCtx *runtime.NativeCallContext) *runtime.FutureValue {
	if callCtx == nil {
		return nil
	}
	if payload := payloadFromState(callCtx.State); payload != nil {
		return payload.handle
	}
	return nil
}

func (i *Interpreter) markBlocked(handle *runtime.FutureValue) {
	if handle == nil {
		return
	}
	if exec, ok := i.executor.(interface {
		MarkBlocked(*runtime.FutureValue)
	}); ok {
		exec.MarkBlocked(handle)
	}
}

func (i *Interpreter) markUnblocked(handle *runtime.FutureValue) {
	if handle == nil {
		return
	}
	if exec, ok := i.executor.(interface {
		MarkUnblocked(*runtime.FutureValue)
	}); ok {
		exec.MarkUnblocked(handle)
	}
}

func (i *Interpreter) channelStateFromHandle(handle int64) (*channelState, error) {
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

func (i *Interpreter) mutexStateFromHandle(handle int64) (*mutexState, error) {
	if handle <= 0 {
		return nil, fmt.Errorf("mutex handle must be positive")
	}
	i.mutexMu.Lock()
	state, ok := i.mutexes[handle]
	i.mutexMu.Unlock()
	if !ok {
		return nil, fmt.Errorf("unknown mutex handle %d", handle)
	}
	if state.cond == nil {
		state.cond = sync.NewCond(&state.mu)
	}
	return state, nil
}

func (i *Interpreter) blockOnNilChannel(callCtx *runtime.NativeCallContext) (runtime.Value, error) {
	if callCtx == nil {
		return nil, fmt.Errorf("channel operation on nil handle outside async context")
	}
	handle := i.getFutureHandle(callCtx)
	if handle == nil {
		return nil, fmt.Errorf("channel operation on nil handle outside async context")
	}
	ctx := i.contextFromCall(callCtx)
	i.markBlocked(handle)
	defer i.markUnblocked(handle)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (i *Interpreter) invokeAwaitWaker(waker runtime.Value, env *runtime.Environment) {
	if waker == nil {
		return
	}
	member, err := i.memberAccessOnValue(waker, ast.NewIdentifier("wake"), env)
	if err != nil {
		return
	}
	_, _ = i.callCallableValue(member, nil, env, nil)
}

func (i *Interpreter) makeAwaitRegistrationValue(cancelFn func()) runtime.Value {
	inst := &runtime.StructInstanceValue{
		Fields: make(map[string]runtime.Value),
	}
	if val, err := i.global.Get("AwaitRegistration"); err == nil {
		if def, conv := toStructDefinitionValue(val, "AwaitRegistration"); conv == nil {
			inst.Definition = def
		}
	}
	cancelMethod := runtime.NativeFunctionValue{
		Name:  "AwaitRegistration.cancel",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			if cancelFn != nil {
				cancelFn()
			}
			return runtime.NilValue{}, nil
		},
	}
	inst.Fields["cancel"] = &runtime.NativeBoundMethodValue{
		Receiver: inst,
		Method:   cancelMethod,
	}
	return inst
}

func resumePayload(payload *asyncContextPayload) {
	if payload == nil {
		return
	}
	payload.awaitBlocked = false
	if payload.resume != nil {
		payload.resume()
	}
}

func (i *Interpreter) pendingSendWaiter(handle *runtime.FutureValue) *channelSendWaiter {
	if handle == nil {
		return nil
	}
	i.channelMu.Lock()
	defer i.channelMu.Unlock()
	return i.pendingChannelSends[handle]
}

func (i *Interpreter) setPendingSendWaiter(waiter *channelSendWaiter) {
	if waiter == nil || waiter.handle == nil {
		return
	}
	i.channelMu.Lock()
	if i.pendingChannelSends == nil {
		i.pendingChannelSends = make(map[*runtime.FutureValue]*channelSendWaiter)
	}
	i.pendingChannelSends[waiter.handle] = waiter
	i.channelMu.Unlock()
}

func (i *Interpreter) clearPendingSendWaiter(handle *runtime.FutureValue) {
	if handle == nil {
		return
	}
	i.channelMu.Lock()
	delete(i.pendingChannelSends, handle)
	i.channelMu.Unlock()
}

func (i *Interpreter) pendingReceiveWaiter(handle *runtime.FutureValue) *channelReceiveWaiter {
	if handle == nil {
		return nil
	}
	i.channelMu.Lock()
	defer i.channelMu.Unlock()
	return i.pendingChannelReceives[handle]
}

func (i *Interpreter) setPendingReceiveWaiter(waiter *channelReceiveWaiter) {
	if waiter == nil || waiter.handle == nil {
		return
	}
	i.channelMu.Lock()
	if i.pendingChannelReceives == nil {
		i.pendingChannelReceives = make(map[*runtime.FutureValue]*channelReceiveWaiter)
	}
	i.pendingChannelReceives[waiter.handle] = waiter
	i.channelMu.Unlock()
}

func (i *Interpreter) clearPendingReceiveWaiter(handle *runtime.FutureValue) {
	if handle == nil {
		return
	}
	i.channelMu.Lock()
	delete(i.pendingChannelReceives, handle)
	i.channelMu.Unlock()
}

func upsertChannelSendWaiter(waiters []*channelSendWaiter, waiter *channelSendWaiter) ([]*channelSendWaiter, bool) {
	for idx, existing := range waiters {
		if existing != nil && existing.handle == waiter.handle {
			waiters[idx] = waiter
			return waiters, true
		}
	}
	return append(waiters, waiter), false
}

func upsertChannelReceiveWaiter(waiters []*channelReceiveWaiter, waiter *channelReceiveWaiter) ([]*channelReceiveWaiter, bool) {
	for idx, existing := range waiters {
		if existing != nil && existing.handle == waiter.handle {
			waiters[idx] = waiter
			return waiters, true
		}
	}
	return append(waiters, waiter), false
}

func (i *Interpreter) channelSendSerial(callCtx *runtime.NativeCallContext, state *channelState, payload runtime.Value) (runtime.Value, error) {
	payloadCtx := payloadFromState(callCtx.State)
	futureHandle := i.getFutureHandle(callCtx)

	state.mu.Lock()
	pending := i.pendingSendWaiter(futureHandle)
	if pending != nil && pending.state != state {
		i.clearPendingSendWaiter(futureHandle)
		pending = nil
	}
	if pending != nil {
		payload = pending.value
		if pending.err != nil {
			i.clearPendingSendWaiter(futureHandle)
			state.mu.Unlock()
			return nil, pending.err
		}
		if pending.delivered {
			i.clearPendingSendWaiter(futureHandle)
			state.mu.Unlock()
			return runtime.NilValue{}, nil
		}
	}

	if futureHandle != nil && futureHandle.CancelRequested() {
		i.clearPendingSendWaiter(futureHandle)
		state.serialSendWaiters = filterChannelSendWaiters(state.serialSendWaiters, futureHandle)
		state.mu.Unlock()
		return nil, context.Canceled
	}

	if state.closed {
		state.mu.Unlock()
		return nil, i.concurrencyError("ChannelSendOnClosed", "send on closed channel")
	}

	if len(state.serialRecvWaiters) > 0 {
		receiver := state.serialRecvWaiters[0]
		state.serialRecvWaiters = state.serialRecvWaiters[1:]
		if receiver != nil {
			receiver.ready = true
			receiver.value = payload
			receiver.closed = false
			resumePayload(receiver.payload)
		}
		i.clearPendingSendWaiter(futureHandle)
		state.mu.Unlock()
		return runtime.NilValue{}, nil
	}

	if state.capacity > 0 && len(state.serialQueue) < state.capacity {
		state.serialQueue = append(state.serialQueue, payload)
		i.clearPendingSendWaiter(futureHandle)
		state.mu.Unlock()
		i.notifyChannelAwaiters(state, channelAwaitRecv)
		return runtime.NilValue{}, nil
	}

	if futureHandle == nil || payloadCtx == nil {
		state.mu.Unlock()
		return nil, fmt.Errorf("channel send would block outside of async context")
	}

	waiter := pending
	if waiter == nil {
		waiter = &channelSendWaiter{handle: futureHandle, state: state}
	}
	waiter.value = payload
	waiter.payload = payloadCtx
	waiter.delivered = false
	waiter.err = nil
	state.serialSendWaiters, _ = upsertChannelSendWaiter(state.serialSendWaiters, waiter)
	shouldNotify := state.capacity == 0 && len(state.serialSendWaiters) == 1
	state.mu.Unlock()

	i.setPendingSendWaiter(waiter)
	payloadCtx.awaitBlocked = true
	if shouldNotify {
		i.notifyChannelAwaiters(state, channelAwaitRecv)
	}
	return nil, errSerialYield
}

func (i *Interpreter) channelReceiveSerial(callCtx *runtime.NativeCallContext, state *channelState) (runtime.Value, error) {
	payloadCtx := payloadFromState(callCtx.State)
	futureHandle := i.getFutureHandle(callCtx)

	state.mu.Lock()
	pending := i.pendingReceiveWaiter(futureHandle)
	if pending != nil && pending.state != state {
		i.clearPendingReceiveWaiter(futureHandle)
		pending = nil
	}
	if pending != nil && pending.ready {
		val := pending.value
		closed := pending.closed
		i.clearPendingReceiveWaiter(futureHandle)
		state.mu.Unlock()
		if closed || val == nil {
			return runtime.NilValue{}, nil
		}
		return val, nil
	}

	if futureHandle != nil && futureHandle.CancelRequested() {
		i.clearPendingReceiveWaiter(futureHandle)
		state.serialRecvWaiters = filterChannelReceiveWaiters(state.serialRecvWaiters, futureHandle)
		state.mu.Unlock()
		return nil, context.Canceled
	}

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
		var val runtime.Value
		if sender != nil {
			sender.delivered = true
			i.setPendingSendWaiter(sender)
			resumePayload(sender.payload)
			val = sender.value
		}
		state.mu.Unlock()
		i.notifyChannelAwaiters(state, channelAwaitSend)
		if val == nil {
			return runtime.NilValue{}, nil
		}
		return val, nil
	}

	if state.closed {
		state.mu.Unlock()
		return runtime.NilValue{}, nil
	}

	if futureHandle == nil || payloadCtx == nil {
		state.mu.Unlock()
		return nil, fmt.Errorf("channel receive would block outside of async context")
	}

	waiter := pending
	if waiter == nil {
		waiter = &channelReceiveWaiter{handle: futureHandle, state: state}
	}
	waiter.payload = payloadCtx
	waiter.ready = false
	waiter.value = nil
	waiter.closed = false
	state.serialRecvWaiters, _ = upsertChannelReceiveWaiter(state.serialRecvWaiters, waiter)
	shouldNotify := state.capacity == 0 && len(state.serialRecvWaiters) == 1
	state.mu.Unlock()

	i.setPendingReceiveWaiter(waiter)
	payloadCtx.awaitBlocked = true
	if shouldNotify {
		i.notifyChannelAwaiters(state, channelAwaitSend)
	}
	return nil, errSerialYield
}

func filterChannelSendWaiters(waiters []*channelSendWaiter, handle *runtime.FutureValue) []*channelSendWaiter {
	out := waiters[:0]
	for _, w := range waiters {
		if w == nil || w.handle == handle {
			continue
		}
		out = append(out, w)
	}
	return out
}

func filterChannelReceiveWaiters(waiters []*channelReceiveWaiter, handle *runtime.FutureValue) []*channelReceiveWaiter {
	out := waiters[:0]
	for _, w := range waiters {
		if w == nil || w.handle == handle {
			continue
		}
		out = append(out, w)
	}
	return out
}

func (i *Interpreter) channelSendOp(callCtx *runtime.NativeCallContext, handleVal runtime.Value, payload runtime.Value) (runtime.Value, error) {
	handle, err := i.int64FromValue(handleVal, "channel handle")
	if err != nil {
		return nil, err
	}
	if handle == 0 {
		return i.blockOnNilChannel(callCtx)
	}
	if handle < 0 {
		return nil, fmt.Errorf("channel handle must be non-negative")
	}
	state, err := i.channelStateFromHandle(handle)
	if err != nil {
		return nil, err
	}

	if _, ok := i.executor.(*SerialExecutor); ok {
		return i.channelSendSerial(callCtx, state, payload)
	}

	state.mu.Lock()
	if state.closed {
		state.mu.Unlock()
		return nil, i.concurrencyError("ChannelSendOnClosed", "send on closed channel")
	}
	ch := state.ch
	state.mu.Unlock()

	select {
	case ch <- payload:
		i.notifyChannelAwaiters(state, channelAwaitRecv)
		return runtime.NilValue{}, nil
	default:
	}

	state.mu.Lock()
	state.sendWaiters++
	newWaiter := state.capacity == 0 && state.sendWaiters == 1
	state.mu.Unlock()
	if newWaiter {
		i.notifyChannelAwaiters(state, channelAwaitRecv)
	}
	defer func() {
		state.mu.Lock()
		state.sendWaiters--
		state.mu.Unlock()
	}()

	ctx := i.contextFromCall(callCtx)
	handleValFuture := i.getFutureHandle(callCtx)
	i.markBlocked(handleValFuture)
	defer i.markUnblocked(handleValFuture)

	select {
	case ch <- payload:
		i.notifyChannelAwaiters(state, channelAwaitRecv)
		return runtime.NilValue{}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (i *Interpreter) channelReceiveOp(callCtx *runtime.NativeCallContext, handleVal runtime.Value) (runtime.Value, error) {
	handle, err := i.int64FromValue(handleVal, "channel handle")
	if err != nil {
		return nil, err
	}
	if handle == 0 {
		return i.blockOnNilChannel(callCtx)
	}
	if handle < 0 {
		return nil, fmt.Errorf("channel handle must be non-negative")
	}
	state, err := i.channelStateFromHandle(handle)
	if err != nil {
		return nil, err
	}

	if _, ok := i.executor.(*SerialExecutor); ok {
		return i.channelReceiveSerial(callCtx, state)
	}

	state.mu.Lock()
	ch := state.ch
	state.mu.Unlock()

	select {
	case value, ok := <-ch:
		if !ok || value == nil {
			return runtime.NilValue{}, nil
		}
		i.notifyChannelAwaiters(state, channelAwaitSend)
		return value, nil
	default:
	}

	state.mu.Lock()
	state.recvWaiters++
	newWaiter := state.capacity == 0 && state.recvWaiters == 1
	state.mu.Unlock()
	if newWaiter {
		i.notifyChannelAwaiters(state, channelAwaitSend)
	}
	defer func() {
		state.mu.Lock()
		state.recvWaiters--
		state.mu.Unlock()
	}()

	ctx := i.contextFromCall(callCtx)
	handleValFuture := i.getFutureHandle(callCtx)
	i.markBlocked(handleValFuture)
	defer i.markUnblocked(handleValFuture)

	select {
	case value, ok := <-ch:
		if !ok || value == nil {
			return runtime.NilValue{}, nil
		}
		i.notifyChannelAwaiters(state, channelAwaitSend)
		return value, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
