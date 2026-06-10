package interpreter

import "able/interpreter-go/pkg/runtime"

// asyncCallableEnv preserves a task-local async payload when a call enters a
// reusable function or lambda closure. Closures are definition-owned and can
// run in several futures at once, so the payload must live in a child
// environment instead of being written to the shared closure.
func (i *Interpreter) asyncCallableEnv(localEnv, callerEnv *runtime.Environment) *runtime.Environment {
	if i == nil || localEnv == nil {
		return localEnv
	}
	payload := payloadFromState(i.runtimeDataFromEnv(callerEnv))
	if payload == nil || payloadFromState(i.runtimeDataFromEnv(localEnv)) == payload {
		return localEnv
	}
	callEnv := runtime.NewEnvironment(localEnv)
	callEnv.SetRuntimeData(payload)
	return callEnv
}
