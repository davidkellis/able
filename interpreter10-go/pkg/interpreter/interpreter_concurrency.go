package interpreter

import (
	"context"
	"errors"
	"fmt"

	goRuntime "runtime"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
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

	stringIdent := ast.NewIdentifier("string")
	stringType := ast.NewSimpleTypeExpression(stringIdent)
	detailsField := ast.NewStructFieldDefinition(stringType, ast.NewIdentifier("details"))
	procErrorDef := ast.NewStructDefinition(ast.NewIdentifier("ProcError"), []*ast.StructFieldDefinition{detailsField}, ast.StructKindNamed, nil, nil, false)
	_, _ = i.evaluateStructDefinition(procErrorDef, i.global)

	pendingDef := ast.NewStructDefinition(ast.NewIdentifier("Pending"), nil, ast.StructKindNamed, nil, nil, false)
	resolvedDef := ast.NewStructDefinition(ast.NewIdentifier("Resolved"), nil, ast.StructKindNamed, nil, nil, false)
	cancelledDef := ast.NewStructDefinition(ast.NewIdentifier("Cancelled"), nil, ast.StructKindNamed, nil, nil, false)
	errorField := ast.NewStructFieldDefinition(ast.NewSimpleTypeExpression(ast.NewIdentifier("ProcError")), ast.NewIdentifier("error"))
	failedDef := ast.NewStructDefinition(ast.NewIdentifier("Failed"), []*ast.StructFieldDefinition{errorField}, ast.StructKindNamed, nil, nil, false)
	_, _ = i.evaluateStructDefinition(pendingDef, i.global)
	_, _ = i.evaluateStructDefinition(resolvedDef, i.global)
	_, _ = i.evaluateStructDefinition(cancelledDef, i.global)
	_, _ = i.evaluateStructDefinition(failedDef, i.global)

	if val, err := i.global.Get("ProcError"); err == nil {
		if def, conv := toStructDefinitionValue(val, "ProcError"); conv == nil {
			i.procErrorStruct = def
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

	if i.procStatusStructs == nil {
		i.procStatusStructs = make(map[string]*runtime.StructDefinitionValue)
	}
	i.procStatusStructs["Pending"] = loadStruct("Pending")
	i.procStatusStructs["Resolved"] = loadStruct("Resolved")
	i.procStatusStructs["Cancelled"] = loadStruct("Cancelled")
	i.procStatusStructs["Failed"] = loadStruct("Failed")

	if def := i.procStatusStructs["Pending"]; def != nil {
		i.procStatusPending = &runtime.StructInstanceValue{Definition: def}
	} else {
		i.procStatusPending = runtime.NilValue{}
	}
	if def := i.procStatusStructs["Resolved"]; def != nil {
		i.procStatusResolved = &runtime.StructInstanceValue{Definition: def}
	} else {
		i.procStatusResolved = runtime.NilValue{}
	}
	if def := i.procStatusStructs["Cancelled"]; def != nil {
		i.procStatusCancelled = &runtime.StructInstanceValue{Definition: def}
	} else {
		i.procStatusCancelled = runtime.NilValue{}
	}

	procYield := &runtime.NativeFunctionValue{
		Name:  "proc_yield",
		Arity: 0,
		Impl: func(callCtx *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			if callCtx == nil {
				return nil, fmt.Errorf("proc_yield must be called inside an asynchronous task")
			}
			payload := payloadFromState(callCtx.State)
			if payload == nil || (payload.kind != asyncContextProc && payload.kind != asyncContextFuture) {
				return nil, fmt.Errorf("proc_yield must be called inside an asynchronous task")
			}
			if _, ok := i.executor.(*SerialExecutor); ok {
				return nil, errSerialYield
			}
			goRuntime.Gosched()
			return runtime.NilValue{}, nil
		},
	}
	i.global.Define("proc_yield", procYield)

	procCancelled := &runtime.NativeFunctionValue{
		Name:  "proc_cancelled",
		Arity: 0,
		Impl: func(ctx *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			payload := payloadFromState(ctx.State)
			if payload == nil || payload.handle == nil || payload.kind != asyncContextProc {
				return nil, fmt.Errorf("proc_cancelled must be called inside an asynchronous task")
			}
			return runtime.BoolValue{Val: payload.handle.CancelRequested()}, nil
		},
	}
	i.global.Define("proc_cancelled", procCancelled)

	procFlush := &runtime.NativeFunctionValue{
		Name:  "proc_flush",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			i.executor.Flush()
			return runtime.NilValue{}, nil
		},
	}
	i.global.Define("proc_flush", procFlush)

	i.concurrencyReady = true
}

func (i *Interpreter) makeAsyncTask(kind asyncContextKind, node ast.Expression, env *runtime.Environment) ProcTask {
	capturedEnv := runtime.NewEnvironment(env)
	return func(ctx context.Context) (runtime.Value, error) {
		payload := payloadFromContext(ctx)
		if payload == nil {
			payload = &asyncContextPayload{kind: kind}
		} else {
			payload.kind = kind
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
		return i.procFailure(payload, sig.value, "task failed")
	default:
		failure := runtime.ErrorValue{Message: err.Error()}
		return newTaskFailure(failure, failure.Message)
	}
}

func (i *Interpreter) asyncCancelled(payload *asyncContextPayload) error {
	label := "Proc"
	if payload != nil && payload.kind == asyncContextFuture {
		label = "Future"
	}
	message := fmt.Sprintf("%s cancelled", label)
	procErr := i.makeProcError(message)
	runtimeErr := i.makeProcRuntimeError(message, procErr)
	return newTaskCancellation(runtimeErr, runtimeErr.Message)
}

func (i *Interpreter) procFailure(payload *asyncContextPayload, value runtime.Value, fallback string) error {
	label := "Proc"
	if payload != nil && payload.kind == asyncContextFuture {
		label = "Future"
	}
	procErr := i.toProcError(value, fallback)
	details := i.procErrorDetails(procErr)
	message := fmt.Sprintf("%s failed: %s", label, details)
	runtimeErr := i.makeProcRuntimeError(message, procErr)
	return newTaskFailure(runtimeErr, runtimeErr.Message)
}

func (i *Interpreter) procHandleStatus(handle *runtime.ProcHandleValue) runtime.Value {
	_, failure, status := handle.Snapshot()
	switch status {
	case runtime.ProcPending:
		return i.procStatusPending
	case runtime.ProcResolved:
		return i.procStatusResolved
	case runtime.ProcCancelled:
		return i.procStatusCancelled
	case runtime.ProcFailed:
		return i.makeProcStatusFailed(failure)
	default:
		return i.procStatusPending
	}
}

func (i *Interpreter) procHandleValue(handle *runtime.ProcHandleValue) runtime.Value {
	result, failure, status := handle.Await()
	switch status {
	case runtime.ProcResolved:
		if result == nil {
			return runtime.NilValue{}
		}
		return result
	case runtime.ProcCancelled:
		if failure == nil {
			failure = i.makeProcRuntimeError("Proc cancelled", i.makeProcError("Proc cancelled"))
		}
		return failure
	case runtime.ProcFailed:
		if failure == nil {
			failure = i.makeProcRuntimeError("Proc failed", i.makeProcError("Proc failed"))
		}
		return failure
	default:
		return i.makeProcRuntimeError("Proc pending", i.makeProcError("Proc pending"))
	}
}

func (i *Interpreter) futureStatus(future *runtime.FutureValue) runtime.Value {
	handle := future.Handle()
	if handle == nil {
		return i.procStatusPending
	}
	_, failure, status := handle.Snapshot()
	switch status {
	case runtime.ProcPending:
		return i.procStatusPending
	case runtime.ProcResolved:
		return i.procStatusResolved
	case runtime.ProcCancelled:
		return i.procStatusCancelled
	case runtime.ProcFailed:
		return i.makeProcStatusFailed(failure)
	default:
		return i.procStatusPending
	}
}

func (i *Interpreter) futureValue(future *runtime.FutureValue) runtime.Value {
	value, failure, status := future.Await()
	switch status {
	case runtime.ProcResolved:
		if value == nil {
			return runtime.NilValue{}
		}
		return value
	case runtime.ProcCancelled:
		if failure == nil {
			failure = i.makeProcRuntimeError("Future cancelled", i.makeProcError("Future cancelled"))
		}
		return failure
	case runtime.ProcFailed:
		if failure == nil {
			failure = i.makeProcRuntimeError("Future failed", i.makeProcError("Future failed"))
		}
		return failure
	default:
		return i.makeProcRuntimeError("Future pending", i.makeProcError("Future pending"))
	}
}

func (i *Interpreter) makeProcStatusFailed(failure runtime.Value) runtime.Value {
	def := i.procStatusStructs["Failed"]
	if def == nil {
		return runtime.ErrorValue{Message: "Proc failed", Payload: map[string]runtime.Value{
			"proc_error": i.toProcError(failure, "Proc failed"),
		}}
	}
	procErr := i.toProcError(failure, "Proc failed")
	return &runtime.StructInstanceValue{
		Definition: def,
		Fields: map[string]runtime.Value{
			"error": procErr,
		},
	}
}

func (i *Interpreter) makeProcError(details string) runtime.Value {
	if i.procErrorStruct == nil {
		return runtime.ErrorValue{Message: details}
	}
	return &runtime.StructInstanceValue{
		Definition: i.procErrorStruct,
		Fields: map[string]runtime.Value{
			"details": runtime.StringValue{Val: details},
		},
	}
}

func (i *Interpreter) procErrorDetails(val runtime.Value) string {
	if v, ok := val.(*runtime.StructInstanceValue); ok {
		if v != nil && i.procErrorStruct != nil && v.Definition == i.procErrorStruct {
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

func (i *Interpreter) makeProcRuntimeError(message string, procErr runtime.Value) runtime.ErrorValue {
	payload := map[string]runtime.Value{
		"proc_error": procErr,
	}
	return runtime.ErrorValue{Message: message, Payload: payload}
}

func (i *Interpreter) toProcError(val runtime.Value, fallback string) runtime.Value {
	if val == nil {
		return i.makeProcError(fallback)
	}
	switch v := val.(type) {
	case *runtime.StructInstanceValue:
		if v != nil && i.procErrorStruct != nil && v.Definition == i.procErrorStruct {
			return v
		}
	case runtime.ErrorValue:
		if v.Payload != nil {
			if procVal, ok := v.Payload["proc_error"]; ok {
				return i.toProcError(procVal, fallback)
			}
		}
		if v.Message != "" {
			return i.makeProcError(v.Message)
		}
	default:
		return i.makeProcError(valueToString(val))
	}
	return i.makeProcError(fallback)
}

func (i *Interpreter) procHandleMember(handle *runtime.ProcHandleValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Proc handle member access expects identifier")
	}
	switch ident.Name {
	case "status":
		fn := runtime.NativeFunctionValue{
			Name:  "proc_handle.status",
			Arity: 0,
			Impl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) == 0 {
					return nil, fmt.Errorf("status requires receiver")
				}
				recv, ok := args[0].(*runtime.ProcHandleValue)
				if !ok {
					return nil, fmt.Errorf("status receiver must be a proc handle")
				}
				return i.procHandleStatus(recv), nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: handle, Method: fn}, nil
	case "value":
		fn := runtime.NativeFunctionValue{
			Name:  "proc_handle.value",
			Arity: 0,
			Impl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) == 0 {
					return nil, fmt.Errorf("value requires receiver")
				}
				recv, ok := args[0].(*runtime.ProcHandleValue)
				if !ok {
					return nil, fmt.Errorf("value receiver must be a proc handle")
				}
				return i.procHandleValue(recv), nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: handle, Method: fn}, nil
	case "cancel":
		fn := runtime.NativeFunctionValue{
			Name:  "proc_handle.cancel",
			Arity: 0,
			Impl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) == 0 {
					return nil, fmt.Errorf("cancel requires receiver")
				}
				recv, ok := args[0].(*runtime.ProcHandleValue)
				if !ok {
					return nil, fmt.Errorf("cancel receiver must be a proc handle")
				}
				recv.RequestCancel()
				return runtime.NilValue{}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: handle, Method: fn}, nil
	default:
		return nil, fmt.Errorf("Unknown proc handle method '%s'", ident.Name)
	}
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
				return i.futureValue(recv), nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: future, Method: fn}, nil
	default:
		return nil, fmt.Errorf("Unknown future method '%s'", ident.Name)
	}
}
