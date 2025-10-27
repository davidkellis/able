package interpreter

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"able/interpreter10-go/pkg/runtime"
)

// ProcTask represents a unit of asynchronous Able work executed by an Executor.
type ProcTask func(ctx context.Context) (runtime.Value, error)

// Executor abstracts the underlying scheduling strategy used for proc/spawn.
type Executor interface {
	RunProc(task ProcTask) *runtime.ProcHandleValue
	RunFuture(task ProcTask) *runtime.FutureValue
	Flush()
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
}

func NewGoroutineExecutor(panicHandler panicValueFunc) *GoroutineExecutor {
	return &GoroutineExecutor{
		executorBase: executorBase{panicValue: panicHandler},
	}
}

func (e *GoroutineExecutor) RunProc(task ProcTask) *runtime.ProcHandleValue {
	ctx, cancel := context.WithCancel(context.Background())
	handle := runtime.NewProcHandleWithContext(ctx, cancel)
	go e.runTask(handle, nil, task, asyncContextProc)
	return handle
}

func (e *GoroutineExecutor) RunFuture(task ProcTask) *runtime.FutureValue {
	ctx, cancel := context.WithCancel(context.Background())
	handle := runtime.NewProcHandleWithContext(ctx, cancel)
	future := runtime.NewFutureFromHandle(handle)
	go e.runTask(handle, future, task, asyncContextFuture)
	return future
}

func (e *GoroutineExecutor) Flush() {}

func (e *GoroutineExecutor) runTask(handle *runtime.ProcHandleValue, future *runtime.FutureValue, task ProcTask, kind asyncContextKind) {
	ctx := handle.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	payload := &asyncContextPayload{kind: kind, handle: handle, future: future}
	ctx = contextWithPayload(ctx, payload)
	handle.MarkStarted()
	result, err := e.safeInvoke(ctx, task)
	e.applyOutcome(handle, result, err)
}

type serialTask struct {
	handle *runtime.ProcHandleValue
	future *runtime.FutureValue
	kind   asyncContextKind
	task   ProcTask
}

// SerialExecutor executes tasks on a single worker goroutine to provide deterministic scheduling for tests.
type SerialExecutor struct {
	executorBase

	mu      sync.Mutex
	cond    *sync.Cond
	queue   []serialTask
	closed  bool
	active  bool
	current *runtime.ProcHandleValue
	paused  bool
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
	e.enqueue(serialTask{handle: handle, kind: asyncContextProc, task: task})
	return handle
}

func (e *SerialExecutor) RunFuture(task ProcTask) *runtime.FutureValue {
	ctx, cancel := context.WithCancel(context.Background())
	handle := runtime.NewProcHandleWithContext(ctx, cancel)
	future := runtime.NewFutureFromHandle(handle)
	e.enqueue(serialTask{handle: handle, future: future, kind: asyncContextFuture, task: task})
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
		e.mu.Lock()
		for len(e.queue) == 0 && !e.closed {
			e.cond.Wait()
		}
		if e.closed && len(e.queue) == 0 {
			e.mu.Unlock()
			return
		}
		task := e.queue[0]
		e.queue = e.queue[1:]
		e.active = true
		e.paused = false
		e.current = task.handle
		e.mu.Unlock()

		ctx := task.handle.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		payload := &asyncContextPayload{kind: task.kind, handle: task.handle, future: task.future}
		ctx = contextWithPayload(ctx, payload)
		task.handle.MarkStarted()
		result, err := e.safeInvoke(ctx, task.task)

		e.mu.Lock()
		e.active = false
		e.paused = false
		e.current = nil
		e.cond.Broadcast()
		e.mu.Unlock()

		if errors.Is(err, errSerialYield) {
			e.enqueue(task)
			continue
		}

		e.applyOutcome(task.handle, result, err)
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
	for (len(e.queue) > 0 || (e.active && !e.paused)) && !e.closed {
		e.cond.Wait()
	}
	e.mu.Unlock()
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
