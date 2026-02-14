package compiler

import (
	"bytes"
)

func (g *generator) renderRuntimeFutureHelpers(buf *bytes.Buffer) {
	buf.WriteString(`
var __able_future_error_def *runtime.StructDefinitionValue
var __able_future_status_defs map[string]*runtime.StructDefinitionValue
var __able_future_status_pending runtime.Value
var __able_future_status_resolved runtime.Value
var __able_future_status_cancelled runtime.Value
var __able_future_status_once sync.Once
var __able_future_defs_mu sync.Mutex

type __able_compiled_yield struct {
	result runtime.Value
	err    error
	done   bool
}

type __able_async_payload struct {
	handle       *runtime.FutureValue
	started      atomic.Bool
	yield        chan __able_compiled_yield
	resume       chan struct{}
	awaitBlocked bool
	resumeTask   func()
	awaitStates  map[*ast.AwaitExpression]*__able_await_state
}

type __able_serial_task struct {
	handle  *runtime.FutureValue
	task    func(*runtime.Environment) (runtime.Value, error)
	env     *runtime.Environment
	payload *__able_async_payload
}

type __able_serial_executor struct {
	mu        sync.Mutex
	cond      *sync.Cond
	queue     []__able_serial_task
	blocked   map[*runtime.FutureValue]__able_serial_task
	closed    bool
	active    bool
	current   *runtime.FutureValue
	paused    bool
	syncDepth int
	forceAuto int
}

type __able_future_executor_iface interface {
	RunFuture(env *runtime.Environment, task func(*runtime.Environment) (runtime.Value, error)) *runtime.FutureValue
	Flush()
	PendingTasks() int
	Drive(handle *runtime.FutureValue)
	ResumeHandle(handle *runtime.FutureValue)
	MarkBlocked(handle *runtime.FutureValue)
	MarkUnblocked(handle *runtime.FutureValue)
}

type __able_goroutine_executor struct {
	pending atomic.Int64
	blocked atomic.Int64
	handles sync.Map
}

type __able_goroutine_handle_state struct {
	blocked atomic.Bool
}

var __able_future_executor_once sync.Once
var __able_future_executor_instance __able_future_executor_iface

func __able_future_executor_mode() string {
	if __able_runtime == nil {
		return "serial"
	}
	if strings.EqualFold(bridge.ExecutorKind(__able_runtime), "goroutine") {
		return "goroutine"
	}
	return "serial"
}

func __able_future_executor() __able_future_executor_iface {
	__able_future_executor_once.Do(func() {
		if __able_future_executor_mode() == "goroutine" {
			__able_future_executor_instance = __able_new_goroutine_executor()
		} else {
			__able_future_executor_instance = __able_new_serial_executor()
		}
	})
	return __able_future_executor_instance
}

func __able_spawn_future(task func(*runtime.Environment) (runtime.Value, error)) *runtime.FutureValue {
	if task == nil {
		return nil
	}
	exec := __able_future_executor()
	if exec == nil {
		return nil
	}
	env := (*runtime.Environment)(nil)
	if __able_runtime != nil {
		env = __able_runtime.Env()
	}
	return exec.RunFuture(env, task)
}

func __able_current_payload() *__able_async_payload {
	if __able_runtime == nil {
		return nil
	}
	env := __able_runtime.Env()
	if env == nil {
		return nil
	}
	if payload, ok := env.RuntimeData().(*__able_async_payload); ok {
		return payload
	}
	return nil
}

func __able_payload_from_ctx(ctx *runtime.NativeCallContext) *__able_async_payload {
	if ctx == nil {
		return nil
	}
	if payload, ok := ctx.State.(*__able_async_payload); ok {
		return payload
	}
	return nil
}

func __able_future_error_struct() *runtime.StructDefinitionValue {
	__able_future_defs_mu.Lock()
	if __able_future_error_def != nil {
		def := __able_future_error_def
		__able_future_defs_mu.Unlock()
		return def
	}
	__able_future_defs_mu.Unlock()
	var def *runtime.StructDefinitionValue
	if __able_runtime != nil {
		if found, err := __able_runtime.StructDefinition("FutureError"); err == nil && found != nil {
			def = found
		}
	}
	if def == nil {
		stringType := ast.NewSimpleTypeExpression(ast.NewIdentifier("String"))
		detailsField := ast.NewStructFieldDefinition(stringType, ast.NewIdentifier("details"))
		placeholder := ast.NewStructDefinition(ast.NewIdentifier("FutureError"), []*ast.StructFieldDefinition{detailsField}, ast.StructKindNamed, nil, nil, false)
		def = &runtime.StructDefinitionValue{Node: placeholder}
	}
	__able_future_defs_mu.Lock()
	__able_future_error_def = def
	__able_future_defs_mu.Unlock()
	return def
}

func __able_future_status_def(name string) *runtime.StructDefinitionValue {
	if name == "" {
		return nil
	}
	__able_future_defs_mu.Lock()
	if __able_future_status_defs == nil {
		__able_future_status_defs = make(map[string]*runtime.StructDefinitionValue)
	}
	if def, ok := __able_future_status_defs[name]; ok {
		__able_future_defs_mu.Unlock()
		return def
	}
	__able_future_defs_mu.Unlock()
	var def *runtime.StructDefinitionValue
	if __able_runtime != nil {
		if found, err := __able_runtime.StructDefinition(name); err == nil && found != nil {
			def = found
		}
	}
	if def == nil {
		var fields []*ast.StructFieldDefinition
		if name == "Failed" {
			errorField := ast.NewStructFieldDefinition(ast.NewSimpleTypeExpression(ast.NewIdentifier("FutureError")), ast.NewIdentifier("error"))
			fields = []*ast.StructFieldDefinition{errorField}
		}
		placeholder := ast.NewStructDefinition(ast.NewIdentifier(name), fields, ast.StructKindNamed, nil, nil, false)
		def = &runtime.StructDefinitionValue{Node: placeholder}
	}
	__able_future_defs_mu.Lock()
	if __able_future_status_defs == nil {
		__able_future_status_defs = make(map[string]*runtime.StructDefinitionValue)
	}
	__able_future_status_defs[name] = def
	__able_future_defs_mu.Unlock()
	return def
}

func __able_future_init_status_instances() {
	__able_future_status_once.Do(func() {
		if def := __able_future_status_def("Pending"); def != nil {
			__able_future_status_pending = &runtime.StructInstanceValue{Definition: def}
		} else {
			__able_future_status_pending = runtime.NilValue{}
		}
		if def := __able_future_status_def("Resolved"); def != nil {
			__able_future_status_resolved = &runtime.StructInstanceValue{Definition: def}
		} else {
			__able_future_status_resolved = runtime.NilValue{}
		}
		if def := __able_future_status_def("Cancelled"); def != nil {
			__able_future_status_cancelled = &runtime.StructInstanceValue{Definition: def}
		} else {
			__able_future_status_cancelled = runtime.NilValue{}
		}
	})
}

func __able_future_make_error(details string) runtime.Value {
	def := __able_future_error_struct()
	if def == nil {
		return runtime.ErrorValue{Message: details}
	}
	return &runtime.StructInstanceValue{
		Definition: def,
		Fields: map[string]runtime.Value{
			"details": runtime.StringValue{Val: details},
		},
	}
}

func __able_future_error_details(val runtime.Value) string {
	if v, ok := val.(*runtime.StructInstanceValue); ok {
		if v != nil && __able_future_error_struct() != nil && v.Definition == __able_future_error_struct() {
			if detail, ok := v.Fields["details"]; ok {
				if str, ok := detail.(runtime.StringValue); ok {
					return str.Val
				}
			}
		}
	}
	if errVal, ok := val.(runtime.ErrorValue); ok && errVal.Message != "" {
		return errVal.Message
	}
	if __able_runtime != nil {
		return __able_stringify(val)
	}
	return fmt.Sprintf("%v", val)
}

func __able_future_make_runtime_error(message string, futureErr runtime.Value) runtime.ErrorValue {
	payload := map[string]runtime.Value{
		"future_error": futureErr,
	}
	if futureErr != nil {
		payload["value"] = futureErr
		payload["cause"] = futureErr
	}
	return runtime.ErrorValue{Message: message, Payload: payload}
}

func __able_future_to_error(val runtime.Value, fallback string) runtime.Value {
	if val == nil {
		return __able_future_make_error(fallback)
	}
	switch v := val.(type) {
	case *runtime.StructInstanceValue:
		if v != nil && __able_future_error_struct() != nil && v.Definition == __able_future_error_struct() {
			return v
		}
	case runtime.ErrorValue:
		if v.Payload != nil {
			if futureVal, ok := v.Payload["future_error"]; ok {
				return __able_future_to_error(futureVal, fallback)
			}
		}
		if v.Message != "" {
			return __able_future_make_error(v.Message)
		}
	default:
		if __able_runtime != nil {
			return __able_future_make_error(__able_stringify(val))
		}
	}
	return __able_future_make_error(fallback)
}

func __able_future_status_failed_value(failure runtime.Value) runtime.Value {
	def := __able_future_status_def("Failed")
	if def == nil {
		return runtime.ErrorValue{Message: "Future failed", Payload: map[string]runtime.Value{
			"future_error": __able_future_to_error(failure, "Future failed"),
		}}
	}
	futureErr := __able_future_to_error(failure, "Future failed")
	return &runtime.StructInstanceValue{
		Definition: def,
		Fields: map[string]runtime.Value{
			"error": futureErr,
		},
	}
}

func __able_future_status_value(future *runtime.FutureValue) runtime.Value {
	__able_future_init_status_instances()
	_, failure, status := future.Snapshot()
	switch status {
	case runtime.FuturePending:
		return __able_future_status_pending
	case runtime.FutureResolved:
		return __able_future_status_resolved
	case runtime.FutureCancelled:
		return __able_future_status_cancelled
	case runtime.FutureFailed:
		return __able_future_status_failed_value(failure)
	default:
		return __able_future_status_pending
	}
}

func __able_future_value_with_payload(future *runtime.FutureValue, payload *__able_async_payload) runtime.Value {
	if exec := __able_future_executor(); exec != nil {
		if payload == nil {
			exec.Flush()
		}
		exec.Drive(future)
	}
	value, failure, status := future.Await()
	switch status {
	case runtime.FutureResolved:
		if value == nil {
			return runtime.NilValue{}
		}
		return value
	case runtime.FutureCancelled:
		if failure == nil {
			failure = __able_future_make_runtime_error("Future cancelled", __able_future_make_error("Future cancelled"))
		}
		return failure
	case runtime.FutureFailed:
		if failure == nil {
			failure = __able_future_make_runtime_error("Future failed", __able_future_make_error("Future failed"))
		}
		return failure
	default:
		return __able_future_make_runtime_error("Future pending", __able_future_make_error("Future pending"))
	}
}

func __able_future_member_value(future *runtime.FutureValue, name string) (runtime.Value, bool) {
	switch name {
	case "status":
		fn := runtime.NativeFunctionValue{
			Name:  "future.status",
			Arity: 0,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) == 0 {
					return nil, fmt.Errorf("status requires receiver")
				}
				recv, ok := args[0].(*runtime.FutureValue)
				if !ok {
					return nil, fmt.Errorf("status receiver must be a future")
				}
				return __able_future_status_value(recv), nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: future, Method: fn}, true
	case "value":
		fn := runtime.NativeFunctionValue{
			Name:  "future.value",
			Arity: 0,
			Impl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) == 0 {
					return nil, fmt.Errorf("value requires receiver")
				}
				recv, ok := args[0].(*runtime.FutureValue)
				if !ok {
					return nil, fmt.Errorf("value receiver must be a future")
				}
				payload := __able_payload_from_ctx(ctx)
				return __able_future_value_with_payload(recv, payload), nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: future, Method: fn}, true
	case "cancel":
		fn := runtime.NativeFunctionValue{
			Name:  "future.cancel",
			Arity: 0,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) == 0 {
					return nil, fmt.Errorf("cancel requires receiver")
				}
				recv, ok := args[0].(*runtime.FutureValue)
				if !ok {
					return nil, fmt.Errorf("cancel receiver must be a future")
				}
				return __able_future_cancel(recv), nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: future, Method: fn}, true
	case "is_ready":
		fn := runtime.NativeFunctionValue{
			Name:  "future.is_ready",
			Arity: 0,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) == 0 {
					return nil, fmt.Errorf("is_ready requires receiver")
				}
				recv, ok := args[0].(*runtime.FutureValue)
				if !ok {
					return nil, fmt.Errorf("is_ready receiver must be a future")
				}
				return runtime.BoolValue{Val: recv.Status() != runtime.FuturePending}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: future, Method: fn}, true
	case "register":
		fn := runtime.NativeFunctionValue{
			Name:  "future.register",
			Arity: 1,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) < 2 {
					return nil, fmt.Errorf("register requires receiver and waker")
				}
				recv, ok := args[0].(*runtime.FutureValue)
				if !ok {
					return nil, fmt.Errorf("register receiver must be a future")
				}
				return __able_future_register_awaiter(recv, args[1])
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: future, Method: fn}, true
	case "commit":
		fn := runtime.NativeFunctionValue{
			Name:  "future.commit",
			Arity: 0,
			Impl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) == 0 {
					return nil, fmt.Errorf("commit requires receiver")
				}
				recv, ok := args[0].(*runtime.FutureValue)
				if !ok {
					return nil, fmt.Errorf("commit receiver must be a future")
				}
				payload := __able_payload_from_ctx(ctx)
				return __able_future_value_with_payload(recv, payload), nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: future, Method: fn}, true
	case "is_default":
		fn := runtime.NativeFunctionValue{
			Name:  "future.is_default",
			Arity: 0,
			Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
				return runtime.BoolValue{Val: false}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: future, Method: fn}, true
	default:
		return nil, false
	}
}

func __able_future_register_awaiter(handle *runtime.FutureValue, waker runtime.Value) (runtime.Value, error) {
	if handle == nil {
		return nil, fmt.Errorf("register requires future handle")
	}
	structWaker, ok := waker.(*runtime.StructInstanceValue)
	if !ok || structWaker == nil {
		return nil, fmt.Errorf("register expects AwaitWaker")
	}
	var cancelled atomic.Bool
	handle.AddAwaiter(func() {
		if cancelled.Load() {
			return
		}
		__able_invoke_await_waker(structWaker)
	})
	return __able_make_await_registration_value(func() { cancelled.Store(true) }), nil
}

func __able_future_cancel(handle *runtime.FutureValue) runtime.Value {
	if handle == nil {
		return runtime.NilValue{}
	}
	if handle.Status() != runtime.FuturePending {
		return runtime.NilValue{}
	}
	handle.RequestCancel()
	if !handle.Started() {
		handle.Cancel(nil)
	}
	if exec := __able_future_executor(); exec != nil {
		exec.ResumeHandle(handle)
	}
	return runtime.NilValue{}
}

func __able_future_yield() (runtime.Value, error) {
	payload := __able_current_payload()
	if payload == nil || payload.handle == nil {
		return nil, fmt.Errorf("future_yield must be called inside an asynchronous task")
	}
	if payload.yield == nil || payload.resume == nil {
		time.Sleep(0)
		return runtime.NilValue{}, nil
	}
	payload.yield <- __able_compiled_yield{}
	<-payload.resume
	return runtime.NilValue{}, nil
}

func __able_future_cancelled() (runtime.Value, error) {
	payload := __able_current_payload()
	if payload == nil || payload.handle == nil {
		return nil, fmt.Errorf("future_cancelled must be called inside an asynchronous task")
	}
	return runtime.BoolValue{Val: payload.handle.CancelRequested()}, nil
}

func __able_future_flush() runtime.Value {
	if exec := __able_future_executor(); exec != nil {
		exec.Flush()
	}
	return runtime.NilValue{}
}

func __able_future_pending_tasks() runtime.Value {
	pending := 0
	if exec := __able_future_executor(); exec != nil {
		pending = exec.PendingTasks()
	}
	if pending < 0 {
		pending = 0
	}
	return runtime.IntegerValue{
		Val:        big.NewInt(int64(pending)),
		TypeSuffix: runtime.IntegerI32,
	}
}

func __able_mark_current_task_blocked() {
	payload := __able_current_payload()
	if payload == nil || payload.handle == nil {
		return
	}
	if exec := __able_future_executor(); exec != nil {
		exec.MarkBlocked(payload.handle)
	}
}

func __able_mark_current_task_unblocked() {
	payload := __able_current_payload()
	if payload == nil || payload.handle == nil {
		return
	}
	if exec := __able_future_executor(); exec != nil {
		exec.MarkUnblocked(payload.handle)
	}
}

type __able_task_error struct {
	status  runtime.FutureStatus
	failure runtime.Value
	message string
}

func (e __able_task_error) Error() string {
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

func (e __able_task_error) Status() runtime.FutureStatus {
	return e.status
}

func (e __able_task_error) FailureValue() runtime.Value {
	return e.failure
}

func __able_task_failure(value runtime.Value, message string) error {
	return __able_task_error{status: runtime.FutureFailed, failure: value, message: message}
}

func __able_task_cancellation(value runtime.Value, message string) error {
	return __able_task_error{status: runtime.FutureCancelled, failure: value, message: message}
}

func __able_async_failure(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}
	switch err.(type) {
	case __able_task_error, *__able_task_error:
		return err
	}
	failure := runtime.ErrorValue{Message: err.Error()}
	return __able_future_failure(failure, "task failed")
}

func __able_async_cancelled() error {
	message := "Future cancelled"
	futureErr := __able_future_make_error(message)
	runtimeErr := __able_future_make_runtime_error(message, futureErr)
	return __able_task_cancellation(runtimeErr, runtimeErr.Message)
}

func __able_future_failure(value runtime.Value, fallback string) error {
	futureErr := __able_future_to_error(value, fallback)
	details := __able_future_error_details(futureErr)
	message := fmt.Sprintf("Future failed: %s", details)
	runtimeErr := __able_future_make_runtime_error(message, futureErr)
	return __able_task_failure(runtimeErr, runtimeErr.Message)
}

func __able_run_compiled_task(payload *__able_async_payload, env *runtime.Environment, task func(*runtime.Environment) (runtime.Value, error)) (result runtime.Value, err error) {
	if payload == nil {
		payload = &__able_async_payload{}
	}
	if env != nil {
		env.SetRuntimeData(payload)
		if __able_runtime != nil {
			prev := __able_runtime.SwapEnv(env)
			defer __able_runtime.SwapEnv(prev)
		}
		defer env.SetRuntimeData(nil)
	}
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case runtime.Value:
				err = __able_future_failure(v, "task failed")
			case __able_value_error:
				err = __able_future_failure(v.value, "task failed")
			case *__able_value_error:
				if v != nil {
					err = __able_future_failure(v.value, "task failed")
				} else {
					err = fmt.Errorf("panic: %v", r)
				}
			case error:
				err = __able_async_failure(v)
			default:
				err = __able_async_failure(fmt.Errorf("panic: %v", r))
			}
			result = nil
		}
	}()
	result, err = task(env)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, context.Canceled
		}
		switch v := err.(type) {
		case __able_value_error:
			return nil, __able_future_failure(v.value, "task failed")
		case *__able_value_error:
			if v != nil {
				return nil, __able_future_failure(v.value, "task failed")
			}
		case __able_task_error, *__able_task_error:
			return nil, err
		}
		return nil, __able_async_failure(err)
	}
	if payload.handle != nil && payload.handle.CancelRequested() {
		return nil, __able_async_cancelled()
	}
	return result, nil
}

func __able_apply_task_outcome(handle *runtime.FutureValue, result runtime.Value, err error) {
	if handle == nil {
		return
	}
	switch {
	case err == nil:
		handle.Resolve(result)
	case errors.Is(err, context.Canceled):
		handle.Cancel(nil)
	default:
		switch taskErr := err.(type) {
		case __able_task_error:
			switch taskErr.Status() {
			case runtime.FutureCancelled:
				handle.Cancel(taskErr.FailureValue())
			case runtime.FutureFailed:
				handle.Fail(taskErr.FailureValue())
			default:
				handle.Fail(taskErr.FailureValue())
			}
		case *__able_task_error:
			if taskErr != nil {
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
		default:
			handle.Fail(runtime.ErrorValue{Message: err.Error()})
		}
	}
}

func __able_new_goroutine_executor() *__able_goroutine_executor {
	return &__able_goroutine_executor{}
}

func (e *__able_goroutine_executor) RunFuture(env *runtime.Environment, task func(*runtime.Environment) (runtime.Value, error)) *runtime.FutureValue {
	if task == nil {
		return nil
	}
	if env == nil && __able_runtime != nil {
		env = __able_runtime.Env()
	}
	if env != nil {
		env = runtime.NewEnvironment(env)
	}
	ctx, cancel := context.WithCancel(context.Background())
	handle := runtime.NewFutureWithContext(ctx, cancel)
	payload := &__able_async_payload{handle: handle}
	e.registerHandle(handle)
	e.pending.Add(1)
	go func() {
		defer e.unregisterHandle(handle)
		defer e.pending.Add(-1)
		handle.MarkStarted()
		result, err := __able_run_compiled_task(payload, env, task)
		__able_apply_task_outcome(handle, result, err)
	}()
	return handle
}

func (e *__able_goroutine_executor) Flush() {
	for {
		pending := e.pending.Load()
		if pending <= 0 {
			return
		}
		blocked := e.blocked.Load()
		if blocked >= pending && pending > 0 {
			return
		}
		time.Sleep(0)
	}
}

func (e *__able_goroutine_executor) PendingTasks() int {
	pending := e.pending.Load()
	if pending < 0 {
		return 0
	}
	return int(pending)
}

func (e *__able_goroutine_executor) Drive(_ *runtime.FutureValue) {}

func (e *__able_goroutine_executor) ResumeHandle(_ *runtime.FutureValue) {}

func (e *__able_goroutine_executor) registerHandle(handle *runtime.FutureValue) {
	if handle == nil {
		return
	}
	e.handles.Store(handle, &__able_goroutine_handle_state{})
}

func (e *__able_goroutine_executor) unregisterHandle(handle *runtime.FutureValue) {
	if handle == nil {
		return
	}
	if stateAny, ok := e.handles.LoadAndDelete(handle); ok {
		if state, ok := stateAny.(*__able_goroutine_handle_state); ok {
			if state.blocked.Load() {
				e.blocked.Add(-1)
			}
		}
	}
}

func (e *__able_goroutine_executor) MarkBlocked(handle *runtime.FutureValue) {
	if handle == nil {
		return
	}
	if stateAny, ok := e.handles.Load(handle); ok {
		if state, ok := stateAny.(*__able_goroutine_handle_state); ok {
			if state.blocked.CompareAndSwap(false, true) {
				e.blocked.Add(1)
			}
		}
	}
}

func (e *__able_goroutine_executor) MarkUnblocked(handle *runtime.FutureValue) {
	if handle == nil {
		return
	}
	if stateAny, ok := e.handles.Load(handle); ok {
		if state, ok := stateAny.(*__able_goroutine_handle_state); ok {
			if state.blocked.CompareAndSwap(true, false) {
				e.blocked.Add(-1)
			}
		}
	}
}

var __able_err_serial_yield = errors.New("serial executor yield requested")

func __able_new_serial_executor() *__able_serial_executor {
	exec := &__able_serial_executor{
		blocked: make(map[*runtime.FutureValue]__able_serial_task),
	}
	exec.cond = sync.NewCond(&exec.mu)
	exec.syncDepth = 1
	go exec.loop()
	return exec
}

func (e *__able_serial_executor) RunFuture(env *runtime.Environment, task func(*runtime.Environment) (runtime.Value, error)) *runtime.FutureValue {
	if task == nil {
		return nil
	}
	if env == nil && __able_runtime != nil {
		env = __able_runtime.Env()
	}
	if env != nil {
		env = runtime.NewEnvironment(env)
	}
	ctx, cancel := context.WithCancel(context.Background())
	handle := runtime.NewFutureWithContext(ctx, cancel)
	payload := &__able_async_payload{
		handle: handle,
		yield:  make(chan __able_compiled_yield),
		resume: make(chan struct{}),
	}
	e.enqueue(__able_serial_task{
		handle:  handle,
		task:    task,
		env:     env,
		payload: payload,
	})
	return handle
}

func (e *__able_serial_executor) enqueue(task __able_serial_task) {
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return
	}
	if task.handle != nil && !task.handle.Started() {
		inserted := false
		for idx, queued := range e.queue {
			if queued.handle != nil && queued.handle.Started() {
				e.queue = append(e.queue[:idx], append([]__able_serial_task{task}, e.queue[idx:]...)...)
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

func (e *__able_serial_executor) loop() {
	for {
		task, ok := e.nextTask()
		if !ok {
			return
		}
		if errors.Is(e.runSerialTask(task), __able_err_serial_yield) {
			continue
		}
	}
}

func (e *__able_serial_executor) nextTask() (__able_serial_task, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for (len(e.queue) == 0 || (e.syncDepth > 0 && e.forceAuto == 0)) && !e.closed {
		e.cond.Wait()
	}
	if e.closed && len(e.queue) == 0 {
		return __able_serial_task{}, false
	}
	task := e.queue[0]
	e.queue = e.queue[1:]
	return task, true
}

func (e *__able_serial_executor) PendingTasks() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.queue)
}

func (e *__able_serial_executor) Flush() {
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

func (e *__able_serial_executor) Drive(handle *runtime.FutureValue) {
	if handle == nil {
		return
	}
	for handle.Status() == runtime.FuturePending {
		task, ok := e.stealTask(handle)
		if !ok {
			return
		}
		if !errors.Is(e.runSerialTask(task), __able_err_serial_yield) {
			return
		}
	}
}

func (e *__able_serial_executor) ResumeHandle(handle *runtime.FutureValue) {
	if e == nil || handle == nil {
		return
	}
	var task __able_serial_task
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

func (e *__able_serial_executor) MarkBlocked(_ *runtime.FutureValue) {}

func (e *__able_serial_executor) MarkUnblocked(_ *runtime.FutureValue) {}

func (e *__able_serial_executor) stealTask(handle *runtime.FutureValue) (__able_serial_task, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for idx, task := range e.queue {
		if task.handle == handle {
			e.queue = append(e.queue[:idx], e.queue[idx+1:]...)
			return task, true
		}
	}
	return __able_serial_task{}, false
}

func (e *__able_serial_executor) runSerialTask(task __able_serial_task) error {
	if task.handle == nil {
		return nil
	}
	if task.handle.Status() != runtime.FuturePending {
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

	payload := task.payload
	if payload == nil {
		payload = &__able_async_payload{
			handle: task.handle,
			yield:  make(chan __able_compiled_yield),
			resume: make(chan struct{}),
		}
		task.payload = payload
	}
	payload.handle = task.handle
	payload.awaitBlocked = false
	payload.resumeTask = func() {
		e.mu.Lock()
		if e.blocked != nil {
			delete(e.blocked, task.handle)
		}
		e.mu.Unlock()
		e.enqueue(task)
	}

	if payload.started.CompareAndSwap(false, true) {
		go func(task __able_serial_task, payload *__able_async_payload) {
			<-payload.resume
			result, err := __able_run_compiled_task(payload, task.env, task.task)
			payload.yield <- __able_compiled_yield{result: result, err: err, done: true}
		}(task, payload)
	}

	task.handle.MarkStarted()
	prevCurrent, prevActive, prevPaused := e.swapCurrent(task.handle)
	payload.resume <- struct{}{}
	event := <-payload.yield
	e.restoreCurrent(prevCurrent, prevActive, prevPaused)

	if !event.done {
		if payload.awaitBlocked {
			e.mu.Lock()
			if e.blocked != nil {
				e.blocked[task.handle] = task
			}
			e.mu.Unlock()
			return __able_err_serial_yield
		}
		e.enqueue(task)
		return __able_err_serial_yield
	}

	__able_apply_task_outcome(task.handle, event.result, event.err)
	return event.err
}

func (e *__able_serial_executor) swapCurrent(handle *runtime.FutureValue) (*runtime.FutureValue, bool, bool) {
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

func (e *__able_serial_executor) restoreCurrent(prev *runtime.FutureValue, prevActive bool, prevPaused bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.current = prev
	e.active = prevActive
	e.paused = prevPaused
	e.cond.Broadcast()
}
`)
}
