package interpreter

import (
	"context"
	"errors"
	"fmt"
	goRuntime "runtime"
	"sync"
	"sync/atomic"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

// ProcTask represents a unit of asynchronous Able work executed by an Executor.
type ProcTask func(ctx context.Context) (runtime.Value, error)

// Executor abstracts the underlying scheduling strategy used for proc/spawn.
type Executor interface {
	RunProc(task ProcTask) *runtime.ProcHandleValue
	RunFuture(task ProcTask) *runtime.FutureValue
	Flush()
	PendingTasks() int
}

type panicValueFunc func(any) runtime.Value

type executorBase struct {
	panicValue panicValueFunc
}

var errSerialYield = errors.New("serial executor yield requested")

type asyncContextKind int

const (
	asyncContextNone asyncContextKind = iota
	asyncContextProc
	asyncContextFuture
)

type asyncContextPayload struct {
	kind   asyncContextKind
	handle *runtime.ProcHandleValue
	future *runtime.FutureValue
	state  *evalState
	// awaitBlocked is set when an await expression is pending and should
	// prevent the serial executor from automatically rescheduling the task.
	awaitBlocked bool
	// awaitStates memoizes await expression state for the current async task.
	awaitStates map[*ast.AwaitExpression]*awaitEvalState
	// resume requeues the current task in the serial executor; populated only
	// when running under the serial scheduler.
	resume func()
}

type asyncContextKey struct{}

func contextWithPayload(ctx context.Context, payload *asyncContextPayload) context.Context {
	return context.WithValue(ctx, asyncContextKey{}, payload)
}

func payloadFromContext(ctx context.Context) *asyncContextPayload {
	if ctx == nil {
		return nil
	}
	if payload, ok := ctx.Value(asyncContextKey{}).(*asyncContextPayload); ok {
		return payload
	}
	return nil
}

func payloadFromState(state any) *asyncContextPayload {
	if payload, ok := state.(*asyncContextPayload); ok {
		return payload
	}
	return nil
}

func (b *executorBase) safeInvoke(ctx context.Context, task ProcTask) (runtime.Value, error) {
	var (
		result runtime.Value
		err    error
	)
	defer func() {
		if r := recover(); r != nil {
			if b.panicValue != nil {
				err = newTaskFailure(b.panicValue(r), fmt.Sprintf("panic: %v", r))
				return
			}
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	result, err = task(ctx)
	return result, err
}

func (b *executorBase) applyOutcome(handle *runtime.ProcHandleValue, result runtime.Value, err error) {
	switch {
	case err == nil:
		handle.Resolve(result)
	case errors.Is(err, context.Canceled):
		handle.Cancel(nil)
	default:
		if taskErr, ok := err.(procTaskError); ok {
			switch taskErr.Status() {
			case runtime.ProcCancelled:
				handle.Cancel(taskErr.FailureValue())
			case runtime.ProcFailed:
				handle.Fail(taskErr.FailureValue())
			default:
				handle.Fail(taskErr.FailureValue())
			}
			return
		}
		handle.Fail(runtime.ErrorValue{Message: err.Error()})
	}
}

// GoroutineExecutor runs tasks using Go goroutines and contexts.
type GoroutineExecutor struct {
	executorBase
	pending atomic.Int64
	blocked atomic.Int64
	handles sync.Map
}

type goroutineHandleState struct {
	blocked atomic.Bool
}

func NewGoroutineExecutor(panicHandler panicValueFunc) *GoroutineExecutor {
	return &GoroutineExecutor{
		executorBase: executorBase{panicValue: panicHandler},
	}
}

func (e *GoroutineExecutor) RunProc(task ProcTask) *runtime.ProcHandleValue {
	ctx, cancel := context.WithCancel(context.Background())
	handle := runtime.NewProcHandleWithContext(ctx, cancel)
	e.registerHandle(handle)
	e.pending.Add(1)
	go e.runTask(handle, nil, task, asyncContextProc)
	return handle
}

func (e *GoroutineExecutor) RunFuture(task ProcTask) *runtime.FutureValue {
	ctx, cancel := context.WithCancel(context.Background())
	handle := runtime.NewProcHandleWithContext(ctx, cancel)
	future := runtime.NewFutureFromHandle(handle)
	e.registerHandle(handle)
	e.pending.Add(1)
	go e.runTask(handle, future, task, asyncContextFuture)
	return future
}

func (e *GoroutineExecutor) Flush() {
	for {
		pending := e.pending.Load()
		if pending <= 0 {
			return
		}
		blocked := e.blocked.Load()
		if blocked >= pending && pending > 0 {
			return
		}
		goRuntime.Gosched()
	}
}

func (e *GoroutineExecutor) PendingTasks() int {
	pending := e.pending.Load()
	if pending < 0 {
		return 0
	}
	return int(pending)
}

func (e *GoroutineExecutor) runTask(handle *runtime.ProcHandleValue, future *runtime.FutureValue, task ProcTask, kind asyncContextKind) {
	defer e.unregisterHandle(handle)
	ctx := handle.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	payload := &asyncContextPayload{kind: kind, handle: handle, future: future}
	ctx = contextWithPayload(ctx, payload)
	handle.MarkStarted()
	result, err := e.safeInvoke(ctx, task)
	e.applyOutcome(handle, result, err)
	e.pending.Add(-1)
}

func (e *GoroutineExecutor) registerHandle(handle *runtime.ProcHandleValue) {
	if handle == nil {
		return
	}
	e.handles.Store(handle, &goroutineHandleState{})
}

func (e *GoroutineExecutor) unregisterHandle(handle *runtime.ProcHandleValue) {
	if handle == nil {
		return
	}
	if stateAny, ok := e.handles.LoadAndDelete(handle); ok {
		if state, ok := stateAny.(*goroutineHandleState); ok {
			if state.blocked.Load() {
				e.blocked.Add(-1)
			}
		}
	}
}

func (e *GoroutineExecutor) MarkBlocked(handle *runtime.ProcHandleValue) {
	if handle == nil {
		return
	}
	if stateAny, ok := e.handles.Load(handle); ok {
		if state, ok := stateAny.(*goroutineHandleState); ok {
			if state.blocked.CompareAndSwap(false, true) {
				e.blocked.Add(1)
			}
		}
	}
}

func (e *GoroutineExecutor) MarkUnblocked(handle *runtime.ProcHandleValue) {
	if handle == nil {
		return
	}
	if stateAny, ok := e.handles.Load(handle); ok {
		if state, ok := stateAny.(*goroutineHandleState); ok {
			if state.blocked.CompareAndSwap(true, false) {
				e.blocked.Add(-1)
			}
		}
	}
}

type serialTask struct {
	handle  *runtime.ProcHandleValue
	future  *runtime.FutureValue
	kind    asyncContextKind
	task    ProcTask
	payload *asyncContextPayload
}

// SerialExecutor executes tasks on a single worker goroutine to provide deterministic scheduling for tests.
type SerialExecutor struct {
	executorBase

	mu        sync.Mutex
	cond      *sync.Cond
	queue     []serialTask
	closed    bool
	active    bool
	current   *runtime.ProcHandleValue
	paused    bool
	syncDepth int
	forceAuto int
}

func (e *SerialExecutor) beginSynchronousSection() {
	if e == nil {
		return
	}
	e.mu.Lock()
	e.syncDepth++
	e.mu.Unlock()
}

func (e *SerialExecutor) endSynchronousSection() {
	if e == nil {
		return
	}
	e.mu.Lock()
	if e.syncDepth > 0 {
		e.syncDepth--
		if e.syncDepth == 0 {
			e.cond.Broadcast()
		}
	}
	e.mu.Unlock()
}

func NewSerialExecutor(panicHandler panicValueFunc) *SerialExecutor {
	exec := &SerialExecutor{
		executorBase: executorBase{panicValue: panicHandler},
	}
	exec.cond = sync.NewCond(&exec.mu)
	go exec.loop()
	return exec
}

func (e *SerialExecutor) RunProc(task ProcTask) *runtime.ProcHandleValue {
	ctx, cancel := context.WithCancel(context.Background())
	handle := runtime.NewProcHandleWithContext(ctx, cancel)
	e.enqueue(serialTask{
		handle:  handle,
		kind:    asyncContextProc,
		task:    task,
		payload: &asyncContextPayload{kind: asyncContextProc, handle: handle},
	})
	return handle
}

func (e *SerialExecutor) RunFuture(task ProcTask) *runtime.FutureValue {
	ctx, cancel := context.WithCancel(context.Background())
	handle := runtime.NewProcHandleWithContext(ctx, cancel)
	future := runtime.NewFutureFromHandle(handle)
	e.enqueue(serialTask{
		handle:  handle,
		future:  future,
		kind:    asyncContextFuture,
		task:    task,
		payload: &asyncContextPayload{kind: asyncContextFuture, handle: handle, future: future},
	})
	return future
}

func (e *SerialExecutor) enqueue(task serialTask) {
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return
	}
	e.queue = append(e.queue, task)
	e.cond.Signal()
	e.mu.Unlock()
}

func (e *SerialExecutor) loop() {
	for {
		task, ok := e.nextTask()
		if !ok {
			return
		}
		if errors.Is(e.runSerialTask(task), errSerialYield) {
			continue
		}
	}
}

func (e *SerialExecutor) Close() {
	e.mu.Lock()
	e.closed = true
	e.cond.Broadcast()
	e.mu.Unlock()
}

func (e *SerialExecutor) Flush() {
	e.mu.Lock()
	e.forceAuto++
	e.cond.Broadcast()
	for (len(e.queue) > 0 || (e.active && !e.paused)) && !e.closed {
		e.cond.Wait()
	}
	e.forceAuto--
	if e.forceAuto < 0 {
		e.forceAuto = 0
	}
	e.mu.Unlock()
}

func (e *SerialExecutor) PendingTasks() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.queue)
}

// Drive executes the task associated with the provided handle on the current goroutine
// until the handle transitions out of the pending state, mirroring the cooperative
// scheduler semantics used by the TypeScript interpreter.
func (e *SerialExecutor) Drive(handle *runtime.ProcHandleValue) {
	if handle == nil {
		return
	}
	for handle.Status() == runtime.ProcPending {
		task, ok := e.stealTask(handle)
		if !ok {
			// Nothing queued for this handle; assume it is already running or resolved.
			return
		}
		if !errors.Is(e.runSerialTask(task), errSerialYield) {
			return
		}
	}
}

func (e *SerialExecutor) suspendCurrent(handle *runtime.ProcHandleValue) {
	if handle == nil {
		return
	}
	e.mu.Lock()
	if e.current == handle && !e.paused {
		e.paused = true
		e.active = false
		e.cond.Broadcast()
	}
	e.mu.Unlock()
}

func (e *SerialExecutor) resumeCurrent(handle *runtime.ProcHandleValue) {
	if handle == nil {
		return
	}
	e.mu.Lock()
	if e.current == handle && e.paused {
		e.paused = false
		e.active = true
		e.cond.Broadcast()
	}
	e.mu.Unlock()
}

// procTaskError carries the Proc status and associated failure value.
type procTaskError interface {
	error
	Status() runtime.ProcStatus
	FailureValue() runtime.Value
}

type taskError struct {
	status  runtime.ProcStatus
	failure runtime.Value
	message string
}

func (e taskError) Error() string {
	if e.message != "" {
		return e.message
	}
	switch e.status {
	case runtime.ProcCancelled:
		return "task cancelled"
	case runtime.ProcFailed:
		return "task failed"
	default:
		return "task error"
	}
}

func (e taskError) Status() runtime.ProcStatus {
	return e.status
}

func (e taskError) FailureValue() runtime.Value {
	return e.failure
}

func newTaskFailure(value runtime.Value, message string) error {
	return taskError{status: runtime.ProcFailed, failure: value, message: message}
}

func newTaskCancellation(value runtime.Value, message string) error {
	return taskError{status: runtime.ProcCancelled, failure: value, message: message}
}

func (e *SerialExecutor) nextTask() (serialTask, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for (len(e.queue) == 0 || (e.syncDepth > 0 && e.forceAuto == 0)) && !e.closed {
		e.cond.Wait()
	}
	if e.closed && len(e.queue) == 0 {
		return serialTask{}, false
	}
	task := e.queue[0]
	e.queue = e.queue[1:]
	return task, true
}

func (e *SerialExecutor) stealTask(handle *runtime.ProcHandleValue) (serialTask, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for idx, task := range e.queue {
		if task.handle == handle {
			e.queue = append(e.queue[:idx], e.queue[idx+1:]...)
			return task, true
		}
	}
	return serialTask{}, false
}

func (e *SerialExecutor) runSerialTask(task serialTask) error {
	if task.handle == nil {
		return nil
	}
	ctx := task.handle.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	payload := task.payload
	if payload == nil {
		payload = &asyncContextPayload{kind: task.kind, handle: task.handle, future: task.future}
		task.payload = payload
	}
	payload.kind = task.kind
	payload.handle = task.handle
	payload.future = task.future
	payload.awaitBlocked = false
	payload.resume = func() {
		e.enqueue(task)
	}
	ctx = contextWithPayload(ctx, payload)
	task.handle.MarkStarted()

	prevCurrent, prevActive, prevPaused := e.swapCurrent(task.handle)
	result, err := e.safeInvoke(ctx, task.task)
	e.restoreCurrent(prevCurrent, prevActive, prevPaused)

	if errors.Is(err, errSerialYield) {
		if payload != nil && payload.awaitBlocked {
			// Awaiting an external wake; the waker will reschedule via resume().
			return err
		}
		e.enqueue(task)
		return err
	}
	e.applyOutcome(task.handle, result, err)
	return err
}

func (e *SerialExecutor) swapCurrent(handle *runtime.ProcHandleValue) (*runtime.ProcHandleValue, bool, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	prev := e.current
	prevActive := e.active
	prevPaused := e.paused
	e.current = handle
	e.active = true
	e.paused = false
	return prev, prevActive, prevPaused
}

func (e *SerialExecutor) restoreCurrent(prev *runtime.ProcHandleValue, prevActive bool, prevPaused bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.current = prev
	e.active = prevActive
	e.paused = prevPaused
	e.cond.Broadcast()
}
