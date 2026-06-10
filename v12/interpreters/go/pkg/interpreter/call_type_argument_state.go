package interpreter

import "able/interpreter-go/pkg/ast"

type callTypeArgumentState struct {
	inferred bool
	version  uint64
}

func typeExpressionSlicesEqual(left []ast.TypeExpression, right []ast.TypeExpression) bool {
	if len(left) != len(right) {
		return false
	}
	for idx := range left {
		if !typeExpressionsEqual(left[idx], right[idx]) {
			return false
		}
	}
	return true
}

func (i *Interpreter) lookupCallTypeArgumentState(call *ast.FunctionCall) (callTypeArgumentState, bool) {
	if i == nil || call == nil {
		return callTypeArgumentState{}, false
	}
	if i.envSingleThread {
		state, ok := i.callTypeArgumentState[call]
		return state, ok
	}
	i.callTypeArgumentStateMu.RLock()
	defer i.callTypeArgumentStateMu.RUnlock()
	state, ok := i.callTypeArgumentState[call]
	return state, ok
}

func (i *Interpreter) storeCallTypeArgumentState(call *ast.FunctionCall, state callTypeArgumentState) {
	if i == nil || call == nil {
		return
	}
	if i.envSingleThread {
		if i.callTypeArgumentState == nil {
			i.callTypeArgumentState = make(map[*ast.FunctionCall]callTypeArgumentState)
		}
		i.callTypeArgumentState[call] = state
		return
	}
	i.callTypeArgumentStateMu.Lock()
	defer i.callTypeArgumentStateMu.Unlock()
	if i.callTypeArgumentState == nil {
		i.callTypeArgumentState = make(map[*ast.FunctionCall]callTypeArgumentState)
	}
	i.callTypeArgumentState[call] = state
}

func (i *Interpreter) inferredCallTypeArgumentVersion(call *ast.FunctionCall) uint64 {
	if state, ok := i.lookupCallTypeArgumentState(call); ok && state.inferred {
		return state.version
	}
	return 0
}

func (i *Interpreter) callHasExplicitTypeArguments(call *ast.FunctionCall) bool {
	if call == nil || len(call.TypeArguments) == 0 {
		return false
	}
	state, ok := i.lookupCallTypeArgumentState(call)
	return !ok || !state.inferred
}

func (i *Interpreter) setInferredCallTypeArguments(call *ast.FunctionCall, typeArgs []ast.TypeExpression) {
	if call == nil {
		return
	}
	if i == nil {
		call.TypeArguments = typeArgs
		return
	}
	state, _ := i.lookupCallTypeArgumentState(call)
	if state.inferred && typeExpressionSlicesEqual(call.TypeArguments, typeArgs) {
		return
	}
	call.TypeArguments = typeArgs
	state.inferred = true
	state.version++
	if state.version == 0 {
		state.version = 1
	}
	i.storeCallTypeArgumentState(call, state)
}
