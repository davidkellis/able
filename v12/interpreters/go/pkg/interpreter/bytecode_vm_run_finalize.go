package interpreter

import (
	"errors"

	"able/interpreter-go/pkg/ast"
)

func (vm *bytecodeVM) finishRunResumable(runErr *error) {
	if runErr == nil {
		return
	}
	if *runErr == nil || !errors.Is(*runErr, errSerialYield) {
		vm.closeAllIterators()
		vm.loopStack = vm.loopStack[:0]
	}
	// Unwind inline call frames on error, adding "called from here"
	// context and cleaning up implicit receivers.
	if *runErr != nil && !errors.Is(*runErr, errSerialYield) && len(vm.callFrames) > 0 {
		if _, ok := (*runErr).(returnSignal); !ok {
			if _, ok := (*runErr).(breakSignal); !ok {
				if _, ok := (*runErr).(continueSignal); !ok {
					for len(vm.callFrames) > 0 {
						frame := vm.callFrames[len(vm.callFrames)-1]
						vm.callFrames = vm.callFrames[:len(vm.callFrames)-1]
						if frame.hasImplicitReceiver {
							state := vm.interp.stateFromEnv(vm.env)
							state.popImplicitReceiver()
						}
						callIP := frame.returnIP - 1
						if callIP >= 0 && callIP < len(frame.program.instructions) {
							callInstr := frame.program.instructions[callIP]
							if callInstr.node != nil {
								if callNode, ok := callInstr.node.(*ast.FunctionCall); ok {
									// Append to the existing diagnostic context's call stack
									// so BuildRuntimeDiagnostic produces "called from here" notes.
									// We cannot use attachRuntimeContext here because it returns
									// early when the error already has a context.
									ctx := runtimeContextFromError(*runErr)
									if ctx != nil {
										ctx.callStack = append(ctx.callStack, runtimeCallFrame{node: callNode})
									} else {
										*runErr = vm.interp.attachRuntimeContext(*runErr, callInstr.node, vm.interp.stateFromEnv(frame.env))
									}
								}
							}
						}
						vm.env = frame.env
						vm.slots = frame.slots
					}
				}
			}
		}
	}
}
