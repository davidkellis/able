package interpreter

import (
	"context"
	"errors"
	"fmt"
	goRuntime "runtime"
	"sync"
	"sync/atomic"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

// ProcTask represents a unit of asynchronous Able work executed by an Executor.
type ProcTask func(ctx context.Context) (runtime.Value, error)

// Executor abstracts the underlying scheduling strategy used for spawn tasks.
type Executor interface {
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
	asyncContextFuture
)

type asyncContextPayload struct {
	kind            asyncContextKind
	handle          *runtime.FutureValue
	state           *evalState
	compiled        bool
	compiledYield   chan compiledYield
	compiledResume  chan struct{}
	bytecodeVM      *bytecodeVM
	bytecodeProgram *bytecodeProgram
	bytecodeEnv     *runtime.Environment
	// awaitBlocked is set when an await expression is pending and should
	// prevent the serial executor from automatically rescheduling the task.
	awaitBlocked atomic.Bool
	// awaitStates memoizes await expression state for the current async task.
	awaitStates map[*ast.AwaitExpression]*awaitEvalState
	// resume requeues the current task in the serial executor; populated only
	// when running under the serial scheduler.
	resume func()
}

func (p *asyncContextPayload) setAwaitBlocked(blocked bool) {
	if p != nil {
		p.awaitBlocked.Store(blocked)
	}
}

func (p *asyncContextPayload) isAwaitBlocked() bool {
	return p != nil && p.awaitBlocked.Load()
}

type asyncContextKey struct{}

type compiledYield struct {
	result runtime.Value
	err    error
	done   bool
}

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

func (b *executorBase) applyOutcome(handle *runtime.FutureValue, result runtime.Value, err error) {
	switch {
	case err == nil:
		handle.Resolve(result)
	case errors.Is(err, context.Canceled):
		handle.Cancel(nil)
	default:
		if taskErr, ok := err.(taskStatusError); ok {
			switch taskErr.Status() {
			case runtime.FutureCancelled:
				handle.Cancel(taskErr.FailureValue())
			case runtime.FutureFailed:
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

func (e *GoroutineExecutor) RunFuture(task ProcTask) *runtime.FutureValue {
	ctx, cancel := context.WithCancel(context.Background())
	handle := runtime.NewFutureWithContext(ctx, cancel)
	e.registerHandle(handle)
	e.pending.Add(1)
	go e.runTask(handle, task, asyncContextFuture)
	return handle
}

func (e *GoroutineExecutor) Flush() {
	for {
		pending := e.pending.Load()
		if pending <= 0 {
			return
		}
		blocked := e.blocked.Load()
		if blocked >= pending && pending > 0 && !e.hasCancellingBlockedTask() {
			return
		}
		goRuntime.Gosched()
	}
}

func (e *GoroutineExecutor) hasCancellingBlockedTask() bool {
	if e == nil {
		return false
	}
	found := false
	e.handles.Range(func(handleAny, stateAny any) bool {
		handle, handleOK := handleAny.(*runtime.FutureValue)
		state, stateOK := stateAny.(*goroutineHandleState)
		if handleOK && stateOK && handle != nil && state != nil &&
			state.blocked.Load() && handle.Status() == runtime.FuturePending && handle.CancelRequested() {
			found = true
			return false
		}
		return true
	})
	return found
}

func (e *GoroutineExecutor) PendingTasks() int {
	pending := e.pending.Load()
	if pending < 0 {
		return 0
	}
	return int(pending)
}

func (e *GoroutineExecutor) runTask(handle *runtime.FutureValue, task ProcTask, kind asyncContextKind) {
	defer e.unregisterHandle(handle)
	ctx := handle.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	payload := &asyncContextPayload{kind: kind, handle: handle}
	ctx = contextWithPayload(ctx, payload)
	handle.MarkStarted()
	result, err := e.safeInvoke(ctx, task)
	e.applyOutcome(handle, result, err)
	e.pending.Add(-1)
}

func (e *GoroutineExecutor) registerHandle(handle *runtime.FutureValue) {
	if handle == nil {
		return
	}
	e.handles.Store(handle, &goroutineHandleState{})
}

func (e *GoroutineExecutor) unregisterHandle(handle *runtime.FutureValue) {
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

func (e *GoroutineExecutor) MarkBlocked(handle *runtime.FutureValue) {
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

func (e *GoroutineExecutor) MarkUnblocked(handle *runtime.FutureValue) {
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
	handle  *runtime.FutureValue
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
	blocked   map[*runtime.FutureValue]serialTask
	closed    bool
	started   bool
	active    bool
	current   *runtime.FutureValue
	paused    bool
	syncDepth int
	forceAuto int
	// workerInFlight closes the interval between the background worker
	// dequeuing a task and runSerialTask publishing it as active. Flush must
	// wait across that interval or it can return before the task resumes.
	workerInFlight int
}

func (e *SerialExecutor) beginSynchronousSection() {
	if e == nil {
		return
	}
	e.mu.Lock()
	e.syncDepth++
	e.mu.Unlock()
}

func (e *SerialExecutor) beginSynchronousSectionIfNeeded() bool {
	if e == nil {
		return false
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.syncDepth > 0 {
		return false
	}
	e.syncDepth++
	return true
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
		blocked:      make(map[*runtime.FutureValue]serialTask),
	}
	exec.cond = sync.NewCond(&exec.mu)
	return exec
}

func (e *SerialExecutor) RunFuture(task ProcTask) *runtime.FutureValue {
	ctx, cancel := context.WithCancel(context.Background())
	handle := runtime.NewFutureWithContext(ctx, cancel)
	e.enqueue(serialTask{
		handle:  handle,
		kind:    asyncContextFuture,
		task:    task,
		payload: &asyncContextPayload{kind: asyncContextFuture, handle: handle},
	})
	return handle
}

func (e *SerialExecutor) enqueue(task serialTask) {
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return
	}
	if !e.started {
		e.started = true
		go e.loop()
	}
	if task.handle != nil && !task.handle.Started() {
		inserted := false
		for idx, queued := range e.queue {
			if queued.handle != nil && queued.handle.Started() {
				e.queue = append(e.queue, serialTask{})
				copy(e.queue[idx+1:], e.queue[idx:])
				e.queue[idx] = task
				inserted = true
				break
			}
		}
		if !inserted {
			e.queue = append(e.queue, task)
		}
	} else {
		e.queue = append(e.queue, task)
	}
	e.cond.Signal()
	e.mu.Unlock()
}

func (e *SerialExecutor) loop() {
	for {
		task, ok := e.nextTask()
		if !ok {
			return
		}
		if errors.Is(e.runSerialTask(task, true), errSerialYield) {
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
	for (len(e.queue) > 0 || e.workerInFlight > 0 || (e.active && !e.paused)) && !e.closed {
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
// scheduler semantics expected by the Able runtime.
func (e *SerialExecutor) Drive(handle *runtime.FutureValue) {
	if handle == nil {
		return
	}
	for handle.Status() == runtime.FuturePending {
		task, ok := e.stealTask(handle)
		if !ok {
			// Nothing queued for this handle; assume it is already running or resolved.
			return
		}
		if !errors.Is(e.runSerialTask(task, false), errSerialYield) {
			return
		}
	}
}

// YieldCurrent runs one queued task (if any) while the current task is paused.
// Used to emulate cooperative yielding for compiled tasks without resumable frames.
func (e *SerialExecutor) YieldCurrent() bool {
	if e == nil {
		return false
	}
	e.mu.Lock()
	current := e.current
	if current == nil || e.closed {
		e.mu.Unlock()
		return false
	}
	idx := -1
	var task serialTask
	for i, queued := range e.queue {
		if queued.handle != nil && queued.handle != current {
			idx = i
			task = queued
			break
		}
	}
	if idx == -1 {
		e.mu.Unlock()
		return false
	}
	e.queue = append(e.queue[:idx], e.queue[idx+1:]...)
	e.mu.Unlock()
	_ = e.runSerialTask(task, false)
	return true
}

func (e *SerialExecutor) suspendCurrent(handle *runtime.FutureValue) {
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

func (e *SerialExecutor) resumeCurrent(handle *runtime.FutureValue) {
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

func (e *SerialExecutor) ResumeHandle(handle *runtime.FutureValue) {
	if e == nil || handle == nil {
		return
	}
	var task serialTask
	var ok bool
	e.mu.Lock()
	if e.blocked != nil {
		task, ok = e.blocked[handle]
		if ok {
			delete(e.blocked, handle)
		}
	}
	e.mu.Unlock()
	if ok {
		e.enqueue(task)
	}
}

// taskStatusError carries the task status and associated failure value.
type taskStatusError interface {
	error
	Status() runtime.FutureStatus
	FailureValue() runtime.Value
}

type taskError struct {
	status  runtime.FutureStatus
	failure runtime.Value
	message string
}

func (e taskError) Error() string {
	if e.message != "" {
		return e.message
	}
	switch e.status {
	case runtime.FutureCancelled:
		return "task cancelled"
	case runtime.FutureFailed:
		return "task failed"
	default:
		return "task error"
	}
}

func (e taskError) Status() runtime.FutureStatus {
	return e.status
}

func (e taskError) FailureValue() runtime.Value {
	return e.failure
}

func newTaskFailure(value runtime.Value, message string) error {
	return taskError{status: runtime.FutureFailed, failure: value, message: message}
}

func newTaskCancellation(value runtime.Value, message string) error {
	return taskError{status: runtime.FutureCancelled, failure: value, message: message}
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
	e.workerInFlight++
	return task, true
}

func (e *SerialExecutor) finishWorkerTask() {
	e.mu.Lock()
	if e.workerInFlight > 0 {
		e.workerInFlight--
	}
	e.cond.Broadcast()
	e.mu.Unlock()
}

func (e *SerialExecutor) stealTask(handle *runtime.FutureValue) (serialTask, bool) {
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

func (e *SerialExecutor) runSerialTask(task serialTask, workerTask bool) error {
	if task.handle == nil {
		if workerTask {
			e.finishWorkerTask()
		}
		return nil
	}
	if task.handle.Status() != runtime.FuturePending {
		if workerTask {
			e.finishWorkerTask()
		}
		e.mu.Lock()
		e.cond.Broadcast()
		e.mu.Unlock()
		return nil
	}
	e.mu.Lock()
	if e.blocked != nil {
		delete(e.blocked, task.handle)
	}
	e.mu.Unlock()
	ctx := task.handle.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	payload := task.payload
	if payload == nil {
		payload = &asyncContextPayload{kind: task.kind, handle: task.handle}
		task.payload = payload
	}
	payload.kind = task.kind
	payload.handle = task.handle
	payload.setAwaitBlocked(false)
	payload.resume = func() {
		e.mu.Lock()
		if e.blocked != nil {
			delete(e.blocked, task.handle)
		}
		e.mu.Unlock()
		e.enqueue(task)
	}
	ctx = contextWithPayload(ctx, payload)
	task.handle.MarkStarted()

	prevCurrent, prevActive, prevPaused := e.swapCurrent(task.handle, workerTask)
	result, err := e.safeInvoke(ctx, task.task)
	e.restoreCurrent(prevCurrent, prevActive, prevPaused)

	if errors.Is(err, errSerialYield) {
		if payload.isAwaitBlocked() {
			// Awaiting an external wake; the waker will reschedule via resume().
			e.mu.Lock()
			if e.blocked != nil {
				e.blocked[task.handle] = task
			}
			e.mu.Unlock()
			return err
		}
		e.enqueue(task)
		return err
	}
	e.applyOutcome(task.handle, result, err)
	return err
}

func (e *SerialExecutor) swapCurrent(handle *runtime.FutureValue, workerTask bool) (*runtime.FutureValue, bool, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	prev := e.current
	prevActive := e.active
	prevPaused := e.paused
	e.current = handle
	e.active = true
	e.paused = false
	if workerTask && e.workerInFlight > 0 {
		e.workerInFlight--
		e.cond.Broadcast()
	}
	return prev, prevActive, prevPaused
}

func (e *SerialExecutor) restoreCurrent(prev *runtime.FutureValue, prevActive bool, prevPaused bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.current = prev
	e.active = prevActive
	e.paused = prevPaused
	e.cond.Broadcast()
}
