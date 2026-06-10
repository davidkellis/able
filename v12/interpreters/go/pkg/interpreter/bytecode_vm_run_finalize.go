package interpreter

import (
	"errors"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) finishRunResumable(runErr *error) {
	if runErr == nil {
		return
	}
	nonYieldExit := *runErr == nil || !errors.Is(*runErr, errSerialYield)
	if nonYieldExit {
		vm.closeAllIterators()
		vm.loopStack = vm.loopStack[:0]
	}
	// Unwind inline call frames on error, adding "called from here"
	// context and cleaning up implicit receivers.
	if *runErr != nil && nonYieldExit && vm.hasCallFrames() {
		if _, ok := (*runErr).(returnSignal); !ok {
			if _, ok := (*runErr).(breakSignal); !ok {
				if _, ok := (*runErr).(continueSignal); !ok {
					for vm.hasCallFrames() {
						calleeSlots := vm.slots
						vm.finishBytecodeArrayOwnershipError(vm.topBytecodeArrayOwnershipParent())
						returnIP, returnProgram, returnSlots, returnEnv, _, _, hasImplicitReceiver, _, _, ok := vm.popCallFrameFields()
						if !ok {
							break
						}
						if hasImplicitReceiver {
							state := vm.interp.stateFromEnv(vm.env)
							state.popImplicitReceiver()
						}
						callIP := returnIP - 1
						if returnProgram != nil && callIP >= 0 && callIP < len(returnProgram.instructions) {
							callInstr := returnProgram.instructions[callIP]
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
										*runErr = vm.interp.attachRuntimeContext(*runErr, callInstr.node, vm.interp.stateFromEnv(returnEnv))
									}
								}
							}
						}
						vm.env = returnEnv
						vm.slots = returnSlots
						if !sameSlotFrame(calleeSlots, returnSlots) {
							switch len(calleeSlots) {
							case 2:
								vm.releaseSlotFrame2(calleeSlots)
							case 4:
								vm.releaseSlotFrame4(calleeSlots)
							default:
								vm.releaseSlotFrame(calleeSlots)
							}
						}
					}
				}
			}
		}
	}
	if nonYieldExit {
		vm.releaseAllActiveTransientRuntimeScopeEnvs()
		vm.releaseCompletedRunFrames()
	}
}

func (vm *bytecodeVM) releaseCompletedRunFrames() {
	if vm == nil {
		return
	}
	if vm.slots != nil {
		vm.releaseSlotFrame(vm.slots)
		vm.slots = nil
	}
	vm.releaseImplicitSlotActiveFrame(vm.implicitSlotActive)
	vm.implicitSlotActive = nil
	vm.releaseActiveValueSlotI32Frame()
	vm.releaseActiveValueSlotFloatFrame()
	vm.releaseActiveI32RegisterFrame()
	if len(vm.callFrames) > 0 {
		for idx := range vm.callFrames {
			frame := &vm.callFrames[idx]
			vm.releaseSlotFrame(frame.slots)
			frame.returnIP = 0
			frame.program = nil
			frame.slots = nil
			frame.env = nil
			frame.transientScopeBase = 0
			frame.returnGenericNames = nil
			frame.returnCoercionFn = nil
			frame.arrayOwnershipParent = nil
			vm.releaseI32RegisterFrame(frame.i32Registers, frame.i32RegisterValid)
			frame.i32RegisterProgram = nil
			frame.i32Registers = nil
			frame.i32RegisterValid = nil
			vm.releaseImplicitSlotActiveFrame(frame.implicitSlotActive)
			frame.implicitSlotActive = nil
			vm.releaseValueSlotI32Frame(frame.slotI32Values, frame.slotI32Valid)
			frame.slotI32Values = nil
			frame.slotI32Valid = nil
			vm.releaseValueSlotFloatFrame(frame.slotFloatValues, frame.slotFloatKinds, frame.slotFloatValid)
			frame.slotFloatValues = nil
			frame.slotFloatKinds = nil
			frame.slotFloatValid = nil
			frame.iterBase = 0
			frame.loopBase = 0
			frame.hasImplicitReceiver = false
			frame.selfFast = false
		}
		vm.callFrames = vm.callFrames[:0]
	}
	if len(vm.selfFastCallFrames) > 0 {
		for idx := range vm.selfFastCallFrames {
			frame := &vm.selfFastCallFrames[idx]
			vm.releaseSlotFrame(frame.slots)
			frame.returnIP = 0
			frame.slots = nil
			frame.env = nil
			frame.transientScopeBase = 0
			frame.returnGenericNames = nil
			frame.returnCoercionFn = nil
			frame.arrayOwnershipParent = nil
			vm.releaseI32RegisterFrame(frame.i32Registers, frame.i32RegisterValid)
			frame.i32RegisterProgram = nil
			frame.i32Registers = nil
			frame.i32RegisterValid = nil
			vm.releaseImplicitSlotActiveFrame(frame.implicitSlotActive)
			frame.implicitSlotActive = nil
			vm.releaseValueSlotI32Frame(frame.slotI32Values, frame.slotI32Valid)
			frame.slotI32Values = nil
			frame.slotI32Valid = nil
			vm.releaseValueSlotFloatFrame(frame.slotFloatValues, frame.slotFloatKinds, frame.slotFloatValid)
			frame.slotFloatValues = nil
			frame.slotFloatKinds = nil
			frame.slotFloatValid = nil
			frame.iterBase = 0
			frame.loopBase = 0
			frame.hasImplicitReceiver = false
		}
		vm.selfFastCallFrames = vm.selfFastCallFrames[:0]
	}
	if len(vm.selfFastMinimal) > 0 {
		for idx := range vm.selfFastMinimal {
			frame := &vm.selfFastMinimal[idx]
			if !frame.reusesSlots {
				vm.releaseSlotFrame(frame.slots)
			}
			frame.returnIP = 0
			frame.slots = nil
			frame.slot0 = nil
			frame.env = nil
			frame.transientScopeBase = 0
			frame.arrayOwnershipParent = nil
			vm.releaseI32RegisterFrame(frame.i32Registers, frame.i32RegisterValid)
			frame.i32RegisterProgram = nil
			frame.i32Registers = nil
			frame.i32RegisterValid = nil
			vm.releaseImplicitSlotActiveFrame(frame.implicitSlotActive)
			frame.implicitSlotActive = nil
			vm.releaseValueSlotI32Frame(frame.slotI32Values, frame.slotI32Valid)
			frame.slotI32Values = nil
			frame.slotI32Valid = nil
			vm.releaseValueSlotFloatFrame(frame.slotFloatValues, frame.slotFloatKinds, frame.slotFloatValid)
			frame.slotFloatValues = nil
			frame.slotFloatKinds = nil
			frame.slotFloatValid = nil
			frame.iterBase = 0
			frame.loopBase = 0
			frame.slot0I32Raw = 0
			frame.slot0I32Valid = false
			frame.slot0FloatRaw = 0
			frame.slot0FloatKind = runtime.FloatF64
			frame.slot0FloatValid = false
			frame.reusesSlots = false
		}
		vm.selfFastMinimal = vm.selfFastMinimal[:0]
	}
	vm.selfFastMinimalSuffix = 0
	vm.clearSelfFastSlot0I32()
	if len(vm.callFrameKinds) > 0 {
		clear(vm.callFrameKinds)
		vm.callFrameKinds = vm.callFrameKinds[:0]
	}
}
