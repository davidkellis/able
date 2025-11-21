package interpreter

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

type channelAwaitKind int

const (
	channelAwaitSend channelAwaitKind = iota
	channelAwaitRecv
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
	handle    *runtime.ProcHandleValue
	payload   *asyncContextPayload
	state     *channelState
	value     runtime.Value
	delivered bool
	err       error
}

type channelReceiveWaiter struct {
	handle  *runtime.ProcHandleValue
	payload *asyncContextPayload
	ready   bool
	value   runtime.Value
	closed  bool
	state   *channelState
}

type mutexState struct {
	mu      sync.Mutex
	cond    *sync.Cond
	locked  bool
	owner   *runtime.ProcHandleValue
	waiters int
}

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

func (i *Interpreter) getProcHandle(callCtx *runtime.NativeCallContext) *runtime.ProcHandleValue {
	if callCtx == nil {
		return nil
	}
	if payload := payloadFromState(callCtx.State); payload != nil {
		return payload.handle
	}
	return nil
}

func (i *Interpreter) markBlocked(handle *runtime.ProcHandleValue) {
	if handle == nil {
		return
	}
	if exec, ok := i.executor.(interface {
		MarkBlocked(*runtime.ProcHandleValue)
	}); ok {
		exec.MarkBlocked(handle)
	}
}

func (i *Interpreter) markUnblocked(handle *runtime.ProcHandleValue) {
	if handle == nil {
		return
	}
	if exec, ok := i.executor.(interface {
		MarkUnblocked(*runtime.ProcHandleValue)
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

func (i *Interpreter) blockOnNilChannel(callCtx *runtime.NativeCallContext) (runtime.Value, error) {
	if callCtx == nil {
		return nil, fmt.Errorf("channel operation on nil handle outside proc context")
	}
	handle := i.getProcHandle(callCtx)
	if handle == nil {
		return nil, fmt.Errorf("channel operation on nil handle outside proc context")
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

func resumePayload(payload *asyncContextPayload) {
	if payload == nil {
		return
	}
	payload.awaitBlocked = false
	if payload.resume != nil {
		payload.resume()
	}
}

func (i *Interpreter) pendingSendWaiter(handle *runtime.ProcHandleValue) *channelSendWaiter {
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
		i.pendingChannelSends = make(map[*runtime.ProcHandleValue]*channelSendWaiter)
	}
	i.pendingChannelSends[waiter.handle] = waiter
	i.channelMu.Unlock()
}

func (i *Interpreter) clearPendingSendWaiter(handle *runtime.ProcHandleValue) {
	if handle == nil {
		return
	}
	i.channelMu.Lock()
	delete(i.pendingChannelSends, handle)
	i.channelMu.Unlock()
}

func (i *Interpreter) pendingReceiveWaiter(handle *runtime.ProcHandleValue) *channelReceiveWaiter {
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
		i.pendingChannelReceives = make(map[*runtime.ProcHandleValue]*channelReceiveWaiter)
	}
	i.pendingChannelReceives[waiter.handle] = waiter
	i.channelMu.Unlock()
}

func (i *Interpreter) clearPendingReceiveWaiter(handle *runtime.ProcHandleValue) {
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
	procHandle := i.getProcHandle(callCtx)

	state.mu.Lock()
	pending := i.pendingSendWaiter(procHandle)
	if pending != nil && pending.state != state {
		i.clearPendingSendWaiter(procHandle)
		pending = nil
	}
	if pending != nil {
		payload = pending.value
		if pending.err != nil {
			i.clearPendingSendWaiter(procHandle)
			state.mu.Unlock()
			return nil, pending.err
		}
		if pending.delivered {
			i.clearPendingSendWaiter(procHandle)
			state.mu.Unlock()
			return runtime.NilValue{}, nil
		}
	}

	if procHandle != nil && procHandle.CancelRequested() {
		i.clearPendingSendWaiter(procHandle)
		state.serialSendWaiters = filterChannelSendWaiters(state.serialSendWaiters, procHandle)
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
		i.clearPendingSendWaiter(procHandle)
		state.mu.Unlock()
		return runtime.NilValue{}, nil
	}

	if state.capacity > 0 && len(state.serialQueue) < state.capacity {
		state.serialQueue = append(state.serialQueue, payload)
		i.clearPendingSendWaiter(procHandle)
		state.mu.Unlock()
		i.notifyChannelAwaiters(state, channelAwaitRecv)
		return runtime.NilValue{}, nil
	}

	if procHandle == nil || payloadCtx == nil {
		state.mu.Unlock()
		return nil, fmt.Errorf("channel send would block outside of proc context")
	}

	waiter := pending
	if waiter == nil {
		waiter = &channelSendWaiter{handle: procHandle, state: state}
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
	procHandle := i.getProcHandle(callCtx)

	state.mu.Lock()
	pending := i.pendingReceiveWaiter(procHandle)
	if pending != nil && pending.state != state {
		i.clearPendingReceiveWaiter(procHandle)
		pending = nil
	}
	if pending != nil && pending.ready {
		val := pending.value
		closed := pending.closed
		i.clearPendingReceiveWaiter(procHandle)
		state.mu.Unlock()
		if closed || val == nil {
			return runtime.NilValue{}, nil
		}
		return val, nil
	}

	if procHandle != nil && procHandle.CancelRequested() {
		i.clearPendingReceiveWaiter(procHandle)
		state.serialRecvWaiters = filterChannelReceiveWaiters(state.serialRecvWaiters, procHandle)
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

	if procHandle == nil || payloadCtx == nil {
		state.mu.Unlock()
		return nil, fmt.Errorf("channel receive would block outside of proc context")
	}

	waiter := pending
	if waiter == nil {
		waiter = &channelReceiveWaiter{handle: procHandle, state: state}
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

func filterChannelSendWaiters(waiters []*channelSendWaiter, handle *runtime.ProcHandleValue) []*channelSendWaiter {
	out := waiters[:0]
	for _, w := range waiters {
		if w == nil || w.handle == handle {
			continue
		}
		out = append(out, w)
	}
	return out
}

func filterChannelReceiveWaiters(waiters []*channelReceiveWaiter, handle *runtime.ProcHandleValue) []*channelReceiveWaiter {
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
	handleValProc := i.getProcHandle(callCtx)
	i.markBlocked(handleValProc)
	defer i.markUnblocked(handleValProc)

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
	handleValProc := i.getProcHandle(callCtx)
	i.markBlocked(handleValProc)
	defer i.markUnblocked(handleValProc)

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
