package interpreter

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync/atomic"

	goRuntime "runtime"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) ensureConcurrencyBuiltins() {
	if i.concurrencyReady {
		return
	}
	i.initConcurrencyBuiltins()
}

func (i *Interpreter) initConcurrencyBuiltins() {
	if i.concurrencyReady {
		return
	}

	stringIdent := ast.NewIdentifier("String")
	stringType := ast.NewSimpleTypeExpression(stringIdent)
	detailsField := ast.NewStructFieldDefinition(stringType, ast.NewIdentifier("details"))
	futureErrorDef := ast.NewStructDefinition(ast.NewIdentifier("FutureError"), []*ast.StructFieldDefinition{detailsField}, ast.StructKindNamed, nil, nil, false)
	_, _ = i.evaluateStructDefinition(futureErrorDef, i.global)

	pendingDef := ast.NewStructDefinition(ast.NewIdentifier("Pending"), nil, ast.StructKindNamed, nil, nil, false)
	resolvedDef := ast.NewStructDefinition(ast.NewIdentifier("Resolved"), nil, ast.StructKindNamed, nil, nil, false)
	cancelledDef := ast.NewStructDefinition(ast.NewIdentifier("Cancelled"), nil, ast.StructKindNamed, nil, nil, false)
	errorField := ast.NewStructFieldDefinition(ast.NewSimpleTypeExpression(ast.NewIdentifier("FutureError")), ast.NewIdentifier("error"))
	failedDef := ast.NewStructDefinition(ast.NewIdentifier("Failed"), []*ast.StructFieldDefinition{errorField}, ast.StructKindNamed, nil, nil, false)
	_, _ = i.evaluateStructDefinition(pendingDef, i.global)
	_, _ = i.evaluateStructDefinition(resolvedDef, i.global)
	_, _ = i.evaluateStructDefinition(cancelledDef, i.global)
	_, _ = i.evaluateStructDefinition(failedDef, i.global)
	futureStatusDef := ast.NewUnionDefinition(
		ast.NewIdentifier("FutureStatus"),
		[]ast.TypeExpression{
			ast.NewSimpleTypeExpression(ast.NewIdentifier("Pending")),
			ast.NewSimpleTypeExpression(ast.NewIdentifier("Resolved")),
			ast.NewSimpleTypeExpression(ast.NewIdentifier("Cancelled")),
			ast.NewSimpleTypeExpression(ast.NewIdentifier("Failed")),
		},
		nil,
		nil,
		false,
	)
	_, _ = i.evaluateUnionDefinition(futureStatusDef, i.global)
	awaitWakerDef := ast.NewStructDefinition(ast.NewIdentifier("AwaitWaker"), nil, ast.StructKindNamed, nil, nil, false)
	awaitRegistrationDef := ast.NewStructDefinition(ast.NewIdentifier("AwaitRegistration"), nil, ast.StructKindNamed, nil, nil, false)
	_, _ = i.evaluateStructDefinition(awaitWakerDef, i.global)
	_, _ = i.evaluateStructDefinition(awaitRegistrationDef, i.global)

	if val, err := i.global.Get("FutureError"); err == nil {
		if def, conv := toStructDefinitionValue(val, "FutureError"); conv == nil {
			i.futureErrorStruct = def
		}
	}

	loadStruct := func(name string) *runtime.StructDefinitionValue {
		val, err := i.global.Get(name)
		if err != nil {
			return nil
		}
		def, conv := toStructDefinitionValue(val, name)
		if conv != nil {
			return nil
		}
		return def
	}

	if i.futureStatusStructs == nil {
		i.futureStatusStructs = make(map[string]*runtime.StructDefinitionValue)
	}
	i.futureStatusStructs["Pending"] = loadStruct("Pending")
	i.futureStatusStructs["Resolved"] = loadStruct("Resolved")
	i.futureStatusStructs["Cancelled"] = loadStruct("Cancelled")
	i.futureStatusStructs["Failed"] = loadStruct("Failed")
	if i.awaitWakerStruct == nil {
		i.awaitWakerStruct = loadStruct("AwaitWaker")
	}

	if def := i.futureStatusStructs["Pending"]; def != nil {
		i.futureStatusPending = &runtime.StructInstanceValue{Definition: def}
	} else {
		i.futureStatusPending = runtime.NilValue{}
	}
	if def := i.futureStatusStructs["Resolved"]; def != nil {
		i.futureStatusResolved = &runtime.StructInstanceValue{Definition: def}
	} else {
		i.futureStatusResolved = runtime.NilValue{}
	}
	if def := i.futureStatusStructs["Cancelled"]; def != nil {
		i.futureStatusCancelled = &runtime.StructInstanceValue{Definition: def}
	} else {
		i.futureStatusCancelled = runtime.NilValue{}
	}

	futureYield := &runtime.NativeFunctionValue{
		Name:  "future_yield",
		Arity: 0,
		Impl: func(callCtx *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			if callCtx == nil {
				return nil, fmt.Errorf("future_yield must be called inside an asynchronous task")
			}
			payload := payloadFromState(callCtx.State)
			if payload == nil || payload.kind != asyncContextFuture || payload.handle == nil {
				return nil, fmt.Errorf("future_yield must be called inside an asynchronous task")
			}
			if _, ok := i.executor.(*SerialExecutor); ok {
				return nil, errSerialYield
			}
			goRuntime.Gosched()
			return runtime.NilValue{}, nil
		},
	}
	i.global.Define("future_yield", futureYield)

	futureCancelled := &runtime.NativeFunctionValue{
		Name:  "future_cancelled",
		Arity: 0,
		Impl: func(ctx *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			payload := payloadFromState(ctx.State)
			if payload == nil || payload.kind != asyncContextFuture || payload.handle == nil {
				return nil, fmt.Errorf("future_cancelled must be called inside an asynchronous task")
			}
			return runtime.BoolValue{Val: payload.handle.CancelRequested()}, nil
		},
	}
	i.global.Define("future_cancelled", futureCancelled)

	futureFlush := &runtime.NativeFunctionValue{
		Name:  "future_flush",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			i.executor.Flush()
			return runtime.NilValue{}, nil
		},
	}
	i.global.Define("future_flush", futureFlush)

	futurePendingTasks := &runtime.NativeFunctionValue{
		Name:  "future_pending_tasks",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			pending := i.executor.PendingTasks()
			if pending < 0 {
				pending = 0
			}
			return runtime.IntegerValue{
				Val:        big.NewInt(int64(pending)),
				TypeSuffix: runtime.IntegerI32,
			}, nil
		},
	}
	i.global.Define("future_pending_tasks", futurePendingTasks)

	awaitDefault := &runtime.NativeFunctionValue{
		Name:  "__able_await_default",
		Arity: 1,
		Impl: func(callCtx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			var callback runtime.Value
			if len(args) > 0 {
				callback = args[len(args)-1]
			}
			return i.makeDefaultAwaitable(callback), nil
		},
	}
	i.global.Define("__able_await_default", awaitDefault)

	awaitSleepMs := &runtime.NativeFunctionValue{
		Name:  "__able_await_sleep_ms",
		Arity: 2,
		Impl: func(callCtx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("__able_await_sleep_ms expects duration")
			}
			duration, err := durationFromValue(args[0])
			if err != nil {
				return nil, err
			}
			var callback runtime.Value
			if len(args) > 1 {
				callback = args[len(args)-1]
			}
			awaitable := newTimerAwaitable(i, duration, callback)
			return awaitable.toStruct(), nil
		},
	}
	i.global.Define("__able_await_sleep_ms", awaitSleepMs)

	i.concurrencyReady = true
}

func (i *Interpreter) makeAsyncTask(node ast.Expression, env *runtime.Environment) ProcTask {
	capturedEnv := runtime.NewEnvironment(env)
	return func(ctx context.Context) (runtime.Value, error) {
		payload := payloadFromContext(ctx)
		if payload == nil {
			payload = &asyncContextPayload{kind: asyncContextFuture}
		} else {
			payload.kind = asyncContextFuture
		}
		return i.runAsyncEvaluation(payload, node, capturedEnv)
	}
}

func (i *Interpreter) runAsyncEvaluation(payload *asyncContextPayload, node ast.Expression, env *runtime.Environment) (runtime.Value, error) {
	if payload == nil {
		payload = &asyncContextPayload{kind: asyncContextNone}
	}
	if payload.state == nil {
		payload.state = newEvalState()
	}
	if env != nil {
		env.SetRuntimeData(payload)
		defer env.SetRuntimeData(nil)
	}
	result, evalErr := i.evaluateExpression(node, env)
	if evalErr != nil {
		if errors.Is(evalErr, context.Canceled) {
			return nil, context.Canceled
		}
		return nil, i.asyncFailure(payload, evalErr)
	}
	if payload != nil && payload.handle != nil && payload.handle.CancelRequested() {
		return nil, i.asyncCancelled(payload)
	}
	return result, nil
}

func (i *Interpreter) asyncFailure(payload *asyncContextPayload, err error) error {
	if errors.Is(err, errSerialYield) {
		return err
	}
	switch sig := err.(type) {
	case raiseSignal:
		return i.futureFailure(payload, sig.value, "task failed")
	default:
		failure := runtime.ErrorValue{Message: err.Error()}
		return newTaskFailure(failure, failure.Message)
	}
}

func (i *Interpreter) asyncCancelled(payload *asyncContextPayload) error {
	message := "Future cancelled"
	futureErr := i.makeFutureError(message)
	runtimeErr := i.makeFutureRuntimeError(message, futureErr)
	return newTaskCancellation(runtimeErr, runtimeErr.Message)
}

func (i *Interpreter) futureFailure(payload *asyncContextPayload, value runtime.Value, fallback string) error {
	futureErr := i.toFutureError(value, fallback)
	details := i.futureErrorDetails(futureErr)
	message := fmt.Sprintf("Future failed: %s", details)
	runtimeErr := i.makeFutureRuntimeError(message, futureErr)
	return newTaskFailure(runtimeErr, runtimeErr.Message)
}

func (i *Interpreter) futureStatus(future *runtime.FutureValue) runtime.Value {
	_, failure, status := future.Snapshot()
	switch status {
	case runtime.FuturePending:
		return i.futureStatusPending
	case runtime.FutureResolved:
		return i.futureStatusResolved
	case runtime.FutureCancelled:
		return i.futureStatusCancelled
	case runtime.FutureFailed:
		return i.makeFutureStatusFailed(failure)
	default:
		return i.futureStatusPending
	}
}

func (i *Interpreter) futureValue(future *runtime.FutureValue) runtime.Value {
	return i.futureValueWithPayload(future, nil)
}

func (i *Interpreter) futureValueWithPayload(future *runtime.FutureValue, payload *asyncContextPayload) runtime.Value {
	if serial, ok := i.executor.(*SerialExecutor); ok {
		if payload == nil {
			// Calls from synchronous contexts should respect queue order before awaiting the target future.
			serial.Flush()
		}
		serial.Drive(future)
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
			failure = i.makeFutureRuntimeError("Future cancelled", i.makeFutureError("Future cancelled"))
		}
		return failure
	case runtime.FutureFailed:
		if failure == nil {
			failure = i.makeFutureRuntimeError("Future failed", i.makeFutureError("Future failed"))
		}
		return failure
	default:
		return i.makeFutureRuntimeError("Future pending", i.makeFutureError("Future pending"))
	}
}

func (i *Interpreter) makeFutureStatusFailed(failure runtime.Value) runtime.Value {
	def := i.futureStatusStructs["Failed"]
	if def == nil {
		return runtime.ErrorValue{Message: "Future failed", Payload: map[string]runtime.Value{
			"future_error": i.toFutureError(failure, "Future failed"),
		}}
	}
	futureErr := i.toFutureError(failure, "Future failed")
	return &runtime.StructInstanceValue{
		Definition: def,
		Fields: map[string]runtime.Value{
			"error": futureErr,
		},
	}
}

func (i *Interpreter) makeFutureError(details string) runtime.Value {
	if i.futureErrorStruct == nil {
		return runtime.ErrorValue{Message: details}
	}
	return &runtime.StructInstanceValue{
		Definition: i.futureErrorStruct,
		Fields: map[string]runtime.Value{
			"details": runtime.StringValue{Val: details},
		},
	}
}

func (i *Interpreter) futureErrorDetails(val runtime.Value) string {
	if v, ok := val.(*runtime.StructInstanceValue); ok {
		if v != nil && i.futureErrorStruct != nil && v.Definition == i.futureErrorStruct {
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
	return valueToString(val)
}

func (i *Interpreter) makeFutureRuntimeError(message string, futureErr runtime.Value) runtime.ErrorValue {
	payload := map[string]runtime.Value{
		"future_error": futureErr,
	}
	if futureErr != nil {
		payload["value"] = futureErr
		payload["cause"] = futureErr
	}
	return runtime.ErrorValue{Message: message, Payload: payload}
}

func (i *Interpreter) toFutureError(val runtime.Value, fallback string) runtime.Value {
	if val == nil {
		return i.makeFutureError(fallback)
	}
	switch v := val.(type) {
	case *runtime.StructInstanceValue:
		if v != nil && i.futureErrorStruct != nil && v.Definition == i.futureErrorStruct {
			return v
		}
	case runtime.ErrorValue:
		if v.Payload != nil {
			if futureVal, ok := v.Payload["future_error"]; ok {
				return i.toFutureError(futureVal, fallback)
			}
		}
		if v.Message != "" {
			return i.makeFutureError(v.Message)
		}
	default:
		return i.makeFutureError(valueToString(val))
	}
	return i.makeFutureError(fallback)
}

func (i *Interpreter) registerHandleAwaiter(handle *runtime.FutureValue, waker runtime.Value) (runtime.Value, error) {
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
		i.invokeAwaitWaker(structWaker, i.global)
	})
	return i.makeAwaitRegistrationValue(func() { cancelled.Store(true) }), nil
}

func (i *Interpreter) futureMember(future *runtime.FutureValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Future member access expects identifier")
	}
	switch ident.Name {
	case "status":
		fn := runtime.NativeFunctionValue{
			Name:  "future.status",
			Arity: 0,
			Impl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) == 0 {
					return nil, fmt.Errorf("status requires receiver")
				}
				recv, ok := args[0].(*runtime.FutureValue)
				if !ok {
					return nil, fmt.Errorf("status receiver must be a future")
				}
				return i.futureStatus(recv), nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: future, Method: fn}, nil
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
				var payload *asyncContextPayload
				if ctx != nil {
					payload = payloadFromState(ctx.State)
				}
				return i.futureValueWithPayload(recv, payload), nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: future, Method: fn}, nil
	case "cancel":
		fn := runtime.NativeFunctionValue{
			Name:  "future.cancel",
			Arity: 0,
			Impl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) == 0 {
					return nil, fmt.Errorf("cancel requires receiver")
				}
				recv, ok := args[0].(*runtime.FutureValue)
				if !ok {
					return nil, fmt.Errorf("cancel receiver must be a future")
				}
				if recv.Status() != runtime.FuturePending {
					return runtime.NilValue{}, nil
				}
				recv.RequestCancel()
				if !recv.Started() {
					recv.Cancel(nil)
				}
				if serial, ok := i.executor.(*SerialExecutor); ok {
					serial.ResumeHandle(recv)
				}
				return runtime.NilValue{}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: future, Method: fn}, nil
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
		return &runtime.NativeBoundMethodValue{Receiver: future, Method: fn}, nil
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
				return i.registerHandleAwaiter(recv, args[1])
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: future, Method: fn}, nil
	case "commit":
		fn := runtime.NativeFunctionValue{
			Name:  "future.commit",
			Arity: 0,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) == 0 {
					return nil, fmt.Errorf("commit requires receiver")
				}
				recv, ok := args[0].(*runtime.FutureValue)
				if !ok {
					return nil, fmt.Errorf("commit receiver must be a future")
				}
				return i.futureValue(recv), nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: future, Method: fn}, nil
	case "is_default":
		fn := runtime.NativeFunctionValue{
			Name:  "future.is_default",
			Arity: 0,
			Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
				return runtime.BoolValue{Val: false}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: future, Method: fn}, nil
	default:
		return nil, fmt.Errorf("Unknown future method '%s'", ident.Name)
	}
}
