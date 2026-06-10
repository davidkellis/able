package interpreter

import "able/interpreter-go/pkg/runtime"

// bytecodeCalleeEnv preserves an async task's runtime payload while bytecode
// executes a reusable function closure. Function closures belong to the
// definition, not to an invocation, so attaching task data to one directly
// would leak state between concurrent calls. A task-local child environment
// keeps lexical lookup intact and gives the callee its own eval/await state.
func (vm *bytecodeVM) bytecodeCalleeEnv(closure *runtime.Environment) *runtime.Environment {
	if vm == nil || closure == nil {
		return closure
	}
	payload := payloadFromState(vm.runtimeData())
	if payload == nil || payloadFromState(closure.RuntimeData()) == payload {
		return closure
	}
	calleeEnv := runtime.NewEnvironment(closure)
	calleeEnv.SetRuntimeData(payload)
	return calleeEnv
}
