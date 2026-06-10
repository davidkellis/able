package interpreter

import "able/interpreter-go/pkg/runtime"

func (i *Interpreter) acquireBytecodeVM(env *runtime.Environment) *bytecodeVM {
	if i != nil {
		if pooled := i.bytecodeVMPool.Get(); pooled != nil {
			if vm, ok := pooled.(*bytecodeVM); ok && vm != nil {
				vm.resetForRun(i, env)
				return vm
			}
		}
	}
	return newBytecodeVM(i, env)
}

func (i *Interpreter) releaseBytecodeVM(vm *bytecodeVM) {
	if i == nil || vm == nil {
		return
	}
	vm.resetForPool()
	i.bytecodeVMPool.Put(vm)
}

func (vm *bytecodeVM) resetForRun(interp *Interpreter, env *runtime.Environment) {
	if vm == nil {
		return
	}
	vm.interp = interp
	vm.env = env
	vm.runtimeDataCacheEnv = nil
	vm.runtimeDataCacheState = 0
	vm.runtimeDataCacheValue = nil
	vm.runtimeDataCacheRev = 0
	vm.runtimeDataCacheEnvRev = 0
	vm.runtimeDataCacheKnown = false
	vm.ip = 0
	vm.currentProgram = nil
	vm.bytecodeProgramEntryPending = false
	vm.activeLookup = bytecodeActiveLookupProgramState{}
	vm.i32RegisterProgram = nil
	vm.i32UnboxFallbackValue = nil
	vm.i32UnboxFallbackSet = false
	vm.clearSelfFastSlot0I32()
	vm.arrayOwnershipObserver = nil

	if vm.stackDepth() > 0 {
		vm.clearStackFrom(0)
		vm.truncateStack(0)
	}
	if len(vm.i32Stack) > 0 {
		clear(vm.i32Stack)
		vm.i32Stack = vm.i32Stack[:0]
	}
	if len(vm.iterStack) > 0 {
		clear(vm.iterStack)
		vm.iterStack = vm.iterStack[:0]
	}
	if len(vm.loopStack) > 0 {
		clear(vm.loopStack)
		vm.loopStack = vm.loopStack[:0]
	}
	if len(vm.ensureStack) > 0 {
		clear(vm.ensureStack)
		vm.ensureStack = vm.ensureStack[:0]
	}
	if len(vm.activeTransientScopeEnvs) > 0 {
		clear(vm.activeTransientScopeEnvs)
		vm.activeTransientScopeEnvs = vm.activeTransientScopeEnvs[:0]
	}
	if vm.slots != nil {
		clear(vm.slots)
		vm.slots = nil
	}
	if len(vm.implicitSlotActive) > 0 {
		clear(vm.implicitSlotActive)
		vm.implicitSlotActive = nil
	}
	vm.releaseActiveValueSlotI32Frame()
	vm.releaseActiveValueSlotFloatFrame()
	vm.releaseActiveI32RegisterFrame()
	if len(vm.ownedI32Slots) > 0 {
		clear(vm.ownedI32Slots)
	}
	if len(vm.ownedFloatSlots) > 0 {
		clear(vm.ownedFloatSlots)
	}
	if len(vm.f64ArrayCache) > 0 {
		clear(vm.f64ArrayCache)
	}
	if len(vm.f64MatrixRowsCache) > 0 {
		clear(vm.f64MatrixRowsCache)
	}
	if len(vm.callFrames) > 0 {
		for idx := range vm.callFrames {
			frame := &vm.callFrames[idx]
			frame.returnIP = 0
			frame.program = nil
			if frame.slots != nil {
				clear(frame.slots)
				frame.slots = nil
			}
			frame.env = nil
			frame.transientScopeBase = 0
			frame.returnGenericNames = nil
			frame.returnCoercionFn = nil
			frame.arrayOwnershipParent = nil
			vm.releaseI32RegisterFrame(frame.i32Registers, frame.i32RegisterValid)
			frame.i32RegisterProgram = nil
			frame.i32Registers = nil
			frame.i32RegisterValid = nil
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
			frame.returnIP = 0
			if frame.slots != nil {
				clear(frame.slots)
				frame.slots = nil
			}
			frame.env = nil
			frame.transientScopeBase = 0
			frame.returnGenericNames = nil
			frame.returnCoercionFn = nil
			frame.arrayOwnershipParent = nil
			vm.releaseI32RegisterFrame(frame.i32Registers, frame.i32RegisterValid)
			frame.i32RegisterProgram = nil
			frame.i32Registers = nil
			frame.i32RegisterValid = nil
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
			frame.returnIP = 0
			if frame.slots != nil {
				clear(frame.slots)
				frame.slots = nil
			}
			frame.slot0 = nil
			frame.env = nil
			frame.transientScopeBase = 0
			frame.arrayOwnershipParent = nil
			vm.releaseI32RegisterFrame(frame.i32Registers, frame.i32RegisterValid)
			frame.i32RegisterProgram = nil
			frame.i32Registers = nil
			frame.i32RegisterValid = nil
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
	if len(vm.bytecodeStatsInlineCallOperands) > 0 {
		clear(vm.bytecodeStatsInlineCallOperands)
		vm.bytecodeStatsInlineCallOperands = vm.bytecodeStatsInlineCallOperands[:0]
	}
	if len(vm.callFrameKinds) > 0 {
		clear(vm.callFrameKinds)
		vm.callFrameKinds = vm.callFrameKinds[:0]
	}
	vm.resetResolvedCallArgScratch()
	// Keep validated per-program lookup/method caches across pooled runs. They
	// are keyed by bytecodeProgram pointers and revalidated against current
	// environments, revisions, and receiver identities before every hit, so
	// preserving them avoids rebuilding the same steady-state caches on every
	// repeated main() call.
}

func (vm *bytecodeVM) resetForPool() {
	if vm == nil {
		return
	}
	vm.releaseAllActiveTransientRuntimeScopeEnvs()
	vm.resetForRun(vm.interp, nil)
	vm.interp = nil
}
