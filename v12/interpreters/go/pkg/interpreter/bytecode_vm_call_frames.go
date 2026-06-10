package interpreter

import "able/interpreter-go/pkg/runtime"

func (vm *bytecodeVM) hasCallFrames() bool {
	return vm != nil && (len(vm.callFrameKinds) > 0 || vm.selfFastMinimalSuffix > 0)
}

func (vm *bytecodeVM) bytecodeDiagnosticStackBase() int {
	if vm == nil || vm.interp == nil || !vm.interp.bytecodeStatsEnabled {
		return -1
	}
	return vm.stackDepth()
}

func (vm *bytecodeVM) topInlineFrameStackBase() int {
	if vm == nil {
		return -1
	}
	if vm.selfFastMinimalSuffix > 0 && len(vm.selfFastMinimal) > 0 {
		return vm.selfFastMinimal[len(vm.selfFastMinimal)-1].stackBase
	}
	if len(vm.callFrameKinds) == 0 {
		return -1
	}
	switch vm.callFrameKinds[len(vm.callFrameKinds)-1] {
	case bytecodeCallFrameKindSelfFastMinimal:
		if len(vm.selfFastMinimal) > 0 {
			return vm.selfFastMinimal[len(vm.selfFastMinimal)-1].stackBase
		}
	case bytecodeCallFrameKindSelfFast:
		if len(vm.selfFastCallFrames) > 0 {
			return vm.selfFastCallFrames[len(vm.selfFastCallFrames)-1].stackBase
		}
	case bytecodeCallFrameKindFull:
		if len(vm.callFrames) > 0 {
			return vm.callFrames[len(vm.callFrames)-1].stackBase
		}
	}
	return -1
}

func bytecodeCanUseSelfFastMinimalFrame(program *bytecodeProgram, returnGenericNames map[string]struct{}, iterBase int, loopBase int, hasImplicitReceiver bool, selfFast bool) bool {
	if !selfFast || returnGenericNames != nil || hasImplicitReceiver {
		return false
	}
	if iterBase == 0 && loopBase == 0 {
		return true
	}
	return program != nil && program.frameLayout != nil && program.frameLayout.preservesControlFlow
}

func (vm *bytecodeVM) materializeSelfFastMinimalSuffixKinds() {
	if vm == nil || vm.selfFastMinimalSuffix <= 0 {
		return
	}
	if cap(vm.callFrameKinds) == 0 {
		vm.callFrameKinds = make([]bytecodeCallFrameKind, 0, 32)
	}
	for remaining := vm.selfFastMinimalSuffix; remaining > 0; remaining-- {
		vm.callFrameKinds = append(vm.callFrameKinds, bytecodeCallFrameKindSelfFastMinimal)
	}
	vm.selfFastMinimalSuffix = 0
}

func (vm *bytecodeVM) pushSelfFastMinimalCallFrameWithBases(returnIP int, slots []runtime.Value, iterBase int, loopBase int) {
	if vm == nil {
		return
	}
	if cap(vm.selfFastMinimal) == 0 {
		vm.selfFastMinimal = make([]bytecodeSelfFastMinimalCallFrame, 0, 32)
	}
	idx := len(vm.selfFastMinimal)
	if idx < cap(vm.selfFastMinimal) {
		vm.selfFastMinimal = vm.selfFastMinimal[:idx+1]
	} else {
		vm.selfFastMinimal = append(vm.selfFastMinimal, bytecodeSelfFastMinimalCallFrame{})
	}
	frame := &vm.selfFastMinimal[idx]
	frame.returnIP = returnIP
	frame.stackBase = vm.bytecodeDiagnosticStackBase()
	frame.slots = slots
	frame.slot0 = nil
	frame.env = vm.env
	frame.transientScopeBase = len(vm.activeTransientScopeEnvs)
	frame.i32RegisterProgram, frame.i32Registers, frame.i32RegisterValid = vm.detachActiveI32RegisterFrame()
	frame.implicitSlotActive = vm.detachImplicitSlotActiveFrame()
	frame.slotI32Values, frame.slotI32Valid = vm.detachActiveValueSlotI32Frame()
	frame.slotFloatValues, frame.slotFloatKinds, frame.slotFloatValid = vm.detachActiveValueSlotFloatFrame()
	frame.iterBase = iterBase
	frame.loopBase = loopBase
	vm.saveSelfFastSlot0I32(frame)
	frame.slot0FloatRaw = 0
	frame.slot0FloatKind = runtime.FloatF64
	frame.slot0FloatValid = false
	frame.reusesSlots = false
	vm.beginBytecodeArrayOwnershipFrame(&frame.arrayOwnershipParent)
	vm.clearSelfFastSlot0I32()
	vm.selfFastMinimalSuffix++
}

func (vm *bytecodeVM) pushSelfFastMinimalCallFrame(returnIP int, slots []runtime.Value) {
	vm.pushSelfFastMinimalCallFrameWithBases(returnIP, slots, 0, 0)
}

func (vm *bytecodeVM) pushSelfFastSlot0CallFrameWithBases(returnIP int, iterBase int, loopBase int) bool {
	if vm == nil || len(vm.slots) == 0 {
		return false
	}
	if cap(vm.selfFastMinimal) == 0 {
		vm.selfFastMinimal = make([]bytecodeSelfFastMinimalCallFrame, 0, 32)
	}
	idx := len(vm.selfFastMinimal)
	if idx < cap(vm.selfFastMinimal) {
		vm.selfFastMinimal = vm.selfFastMinimal[:idx+1]
	} else {
		vm.selfFastMinimal = append(vm.selfFastMinimal, bytecodeSelfFastMinimalCallFrame{})
	}
	frame := &vm.selfFastMinimal[idx]
	frame.returnIP = returnIP
	frame.stackBase = vm.bytecodeDiagnosticStackBase()
	frame.slots = vm.slots
	frame.slot0 = vm.slots[0]
	frame.env = vm.env
	frame.transientScopeBase = len(vm.activeTransientScopeEnvs)
	frame.i32RegisterProgram, frame.i32Registers, frame.i32RegisterValid = vm.detachActiveI32RegisterFrame()
	frame.implicitSlotActive = vm.detachImplicitSlotActiveFrame()
	frame.slotI32Values, frame.slotI32Valid = nil, nil
	frame.slotFloatValues, frame.slotFloatKinds, frame.slotFloatValid = nil, nil, nil
	frame.iterBase = iterBase
	frame.loopBase = loopBase
	vm.saveSelfFastSlot0I32(frame)
	vm.saveReusedSlot0ValueSlotFloat(frame)
	frame.reusesSlots = true
	vm.beginBytecodeArrayOwnershipFrame(&frame.arrayOwnershipParent)
	vm.clearSelfFastSlot0I32()
	vm.selfFastMinimalSuffix++
	return true
}

func (vm *bytecodeVM) pushSelfFastSlot0CallFrame(returnIP int) bool {
	return vm.pushSelfFastSlot0CallFrameWithBases(returnIP, 0, 0)
}

func (vm *bytecodeVM) pushInlineSelfFastFrame(returnIP int, program *bytecodeProgram, slots []runtime.Value, returnGenericNames map[string]struct{}, iterBase int, loopBase int, hasImplicitReceiver bool) {
	if vm == nil {
		return
	}
	if bytecodeCanUseSelfFastMinimalFrame(program, returnGenericNames, iterBase, loopBase, hasImplicitReceiver, true) {
		vm.pushSelfFastMinimalCallFrameWithBases(returnIP, slots, iterBase, loopBase)
		return
	}
	vm.pushCallFrame(returnIP, program, slots, vm.env, returnGenericNames, iterBase, loopBase, hasImplicitReceiver, true)
}

func (vm *bytecodeVM) restoreSelfFastMinimalFrameSlot0(frame *bytecodeSelfFastMinimalCallFrame, slots []runtime.Value) {
	if frame == nil {
		return
	}
	if frame.reusesSlots && len(slots) > 0 {
		slots[0] = frame.slot0
		vm.restoreReusedSlot0ValueSlotI32(frame.slot0I32Raw, frame.slot0I32Valid)
		vm.restoreReusedSlot0ValueSlotFloat(frame.slot0FloatRaw, frame.slot0FloatKind, frame.slot0FloatValid)
	}
	frame.slot0 = nil
	frame.reusesSlots = false
	frame.slot0FloatRaw = 0
	frame.slot0FloatKind = runtime.FloatF64
	frame.slot0FloatValid = false
	vm.restoreSelfFastSlot0I32(frame)
}

func (vm *bytecodeVM) pushCallFrame(returnIP int, program *bytecodeProgram, slots []runtime.Value, env *runtime.Environment, returnGenericNames map[string]struct{}, iterBase int, loopBase int, hasImplicitReceiver bool, selfFast bool) {
	if vm == nil {
		return
	}
	if vm.interp != nil && vm.interp.bytecodeStatsEnabled {
		vm.bytecodeProgramEntryPending = true
	}
	if selfFast {
		if bytecodeCanUseSelfFastMinimalFrame(program, returnGenericNames, iterBase, loopBase, hasImplicitReceiver, selfFast) {
			vm.pushSelfFastMinimalCallFrameWithBases(returnIP, slots, iterBase, loopBase)
			return
		}
		vm.materializeSelfFastMinimalSuffixKinds()
		vm.clearSelfFastSlot0I32()
		if cap(vm.selfFastCallFrames) == 0 {
			vm.selfFastCallFrames = make([]bytecodeSelfFastCallFrame, 0, 32)
		}
		if cap(vm.callFrameKinds) == 0 {
			vm.callFrameKinds = make([]bytecodeCallFrameKind, 0, 32)
		}
		vm.callFrameKinds = append(vm.callFrameKinds, bytecodeCallFrameKindSelfFast)
		idx := len(vm.selfFastCallFrames)
		if idx < cap(vm.selfFastCallFrames) {
			vm.selfFastCallFrames = vm.selfFastCallFrames[:idx+1]
		} else {
			vm.selfFastCallFrames = append(vm.selfFastCallFrames, bytecodeSelfFastCallFrame{})
		}
		frame := &vm.selfFastCallFrames[idx]
		frame.returnIP = returnIP
		frame.stackBase = vm.bytecodeDiagnosticStackBase()
		frame.slots = slots
		frame.env = vm.env
		frame.transientScopeBase = len(vm.activeTransientScopeEnvs)
		frame.returnGenericNames = returnGenericNames
		frame.i32RegisterProgram, frame.i32Registers, frame.i32RegisterValid = vm.detachActiveI32RegisterFrame()
		frame.implicitSlotActive = vm.detachImplicitSlotActiveFrame()
		frame.slotI32Values, frame.slotI32Valid = vm.detachActiveValueSlotI32Frame()
		frame.slotFloatValues, frame.slotFloatKinds, frame.slotFloatValid = vm.detachActiveValueSlotFloatFrame()
		frame.iterBase = iterBase
		frame.loopBase = loopBase
		frame.hasImplicitReceiver = hasImplicitReceiver
		vm.beginBytecodeArrayOwnershipFrame(&frame.arrayOwnershipParent)
		return
	}
	vm.materializeSelfFastMinimalSuffixKinds()
	vm.clearSelfFastSlot0I32()
	if cap(vm.callFrames) == 0 {
		vm.callFrames = make([]bytecodeCallFrame, 0, 32)
	}
	if cap(vm.callFrameKinds) == 0 {
		vm.callFrameKinds = make([]bytecodeCallFrameKind, 0, 32)
	}
	vm.callFrameKinds = append(vm.callFrameKinds, bytecodeCallFrameKindFull)
	idx := len(vm.callFrames)
	if idx < cap(vm.callFrames) {
		vm.callFrames = vm.callFrames[:idx+1]
	} else {
		vm.callFrames = append(vm.callFrames, bytecodeCallFrame{})
	}
	frame := &vm.callFrames[idx]
	frame.returnIP = returnIP
	frame.stackBase = vm.bytecodeDiagnosticStackBase()
	frame.program = program
	frame.slots = slots
	frame.env = env
	frame.activeLookup = vm.captureActiveLookupProgramState(program)
	frame.transientScopeBase = len(vm.activeTransientScopeEnvs)
	frame.returnGenericNames = returnGenericNames
	frame.i32RegisterProgram, frame.i32Registers, frame.i32RegisterValid = vm.detachActiveI32RegisterFrame()
	frame.implicitSlotActive = vm.detachImplicitSlotActiveFrame()
	frame.slotI32Values, frame.slotI32Valid = vm.detachActiveValueSlotI32Frame()
	frame.slotFloatValues, frame.slotFloatKinds, frame.slotFloatValid = vm.detachActiveValueSlotFloatFrame()
	frame.iterBase = iterBase
	frame.loopBase = loopBase
	frame.hasImplicitReceiver = hasImplicitReceiver
	frame.selfFast = selfFast
	vm.beginBytecodeArrayOwnershipFrame(&frame.arrayOwnershipParent)
}

func (vm *bytecodeVM) peekReturnGenericNames() map[string]struct{} {
	if vm == nil {
		return nil
	}
	if vm.selfFastMinimalSuffix > 0 {
		return nil
	}
	if len(vm.callFrameKinds) == 0 {
		return nil
	}
	switch vm.callFrameKinds[len(vm.callFrameKinds)-1] {
	case bytecodeCallFrameKindSelfFastMinimal:
		return nil
	case bytecodeCallFrameKindSelfFast:
		if len(vm.selfFastCallFrames) == 0 {
			return nil
		}
		return vm.selfFastCallFrames[len(vm.selfFastCallFrames)-1].returnGenericNames
	default:
		if len(vm.callFrames) == 0 {
			return nil
		}
		return vm.callFrames[len(vm.callFrames)-1].returnGenericNames
	}
}

func (vm *bytecodeVM) setTopCallFrameReturnCoercionFunction(fn *runtime.FunctionValue) {
	if vm == nil || fn == nil || vm.selfFastMinimalSuffix > 0 || len(vm.callFrameKinds) == 0 {
		return
	}
	switch vm.callFrameKinds[len(vm.callFrameKinds)-1] {
	case bytecodeCallFrameKindSelfFast:
		if len(vm.selfFastCallFrames) > 0 {
			vm.selfFastCallFrames[len(vm.selfFastCallFrames)-1].returnCoercionFn = fn
		}
	case bytecodeCallFrameKindFull:
		if len(vm.callFrames) > 0 {
			vm.callFrames[len(vm.callFrames)-1].returnCoercionFn = fn
		}
	}
}

func (vm *bytecodeVM) peekReturnCoercionFunction() *runtime.FunctionValue {
	if vm == nil || vm.selfFastMinimalSuffix > 0 || len(vm.callFrameKinds) == 0 {
		return nil
	}
	switch vm.callFrameKinds[len(vm.callFrameKinds)-1] {
	case bytecodeCallFrameKindSelfFast:
		if len(vm.selfFastCallFrames) == 0 {
			return nil
		}
		return vm.selfFastCallFrames[len(vm.selfFastCallFrames)-1].returnCoercionFn
	case bytecodeCallFrameKindFull:
		if len(vm.callFrames) == 0 {
			return nil
		}
		return vm.callFrames[len(vm.callFrames)-1].returnCoercionFn
	default:
		return nil
	}
}

func (vm *bytecodeVM) popCallFrameFields() (returnIP int, program *bytecodeProgram, slots []runtime.Value, env *runtime.Environment, iterBase int, loopBase int, hasImplicitReceiver bool, selfFast bool, activeLookup bytecodeActiveLookupProgramState, ok bool) {
	if vm == nil {
		return 0, nil, nil, nil, 0, 0, false, false, bytecodeActiveLookupProgramState{}, false
	}
	if vm.selfFastMinimalSuffix > 0 {
		idx := len(vm.selfFastMinimal) - 1
		frame := &vm.selfFastMinimal[idx]
		returnIP = frame.returnIP
		slots = frame.slots
		env = frame.env
		if frame.transientScopeBase < len(vm.activeTransientScopeEnvs) {
			vm.releaseActiveTransientRuntimeScopeEnvsToBase(frame.transientScopeBase)
		}
		reusesSlots := frame.reusesSlots
		calleeImplicitSlotActive := vm.detachImplicitSlotActiveFrame()
		implicitSlotActive := frame.implicitSlotActive
		i32Program, i32Registers, i32Valid := frame.i32RegisterProgram, frame.i32Registers, frame.i32RegisterValid
		slotI32Values, slotI32Valid := frame.slotI32Values, frame.slotI32Valid
		slotFloatValues, slotFloatKinds, slotFloatValid := frame.slotFloatValues, frame.slotFloatKinds, frame.slotFloatValid
		frame.i32RegisterProgram, frame.i32Registers, frame.i32RegisterValid = nil, nil, nil
		frame.implicitSlotActive = nil
		frame.slotI32Values, frame.slotI32Valid = nil, nil
		frame.slotFloatValues, frame.slotFloatKinds, frame.slotFloatValid = nil, nil, nil
		frame.env = nil
		frame.transientScopeBase = 0
		frame.arrayOwnershipParent = nil
		vm.restoreSelfFastMinimalFrameSlot0(frame, slots)
		vm.releaseImplicitSlotActiveFrame(calleeImplicitSlotActive)
		vm.restoreImplicitSlotActiveFrame(implicitSlotActive)
		vm.restoreI32RegisterFrame(i32Program, i32Registers, i32Valid)
		if !reusesSlots {
			vm.restoreValueSlotSidecarFrames(slots, slotI32Values, slotI32Valid, slotFloatValues, slotFloatKinds, slotFloatValid)
		}
		vm.selfFastMinimal = vm.selfFastMinimal[:idx]
		vm.selfFastMinimalSuffix--
		return returnIP, nil, slots, env, frame.iterBase, frame.loopBase, false, true, bytecodeActiveLookupProgramState{}, true
	}
	if len(vm.callFrameKinds) == 0 {
		return 0, nil, nil, nil, 0, 0, false, false, bytecodeActiveLookupProgramState{}, false
	}
	lastKindIdx := len(vm.callFrameKinds) - 1
	kind := vm.callFrameKinds[lastKindIdx]
	vm.callFrameKinds = vm.callFrameKinds[:lastKindIdx]
	switch kind {
	case bytecodeCallFrameKindSelfFastMinimal:
		idx := len(vm.selfFastMinimal) - 1
		frame := &vm.selfFastMinimal[idx]
		returnIP = frame.returnIP
		slots = frame.slots
		env = frame.env
		if frame.transientScopeBase < len(vm.activeTransientScopeEnvs) {
			vm.releaseActiveTransientRuntimeScopeEnvsToBase(frame.transientScopeBase)
		}
		reusesSlots := frame.reusesSlots
		calleeImplicitSlotActive := vm.detachImplicitSlotActiveFrame()
		implicitSlotActive := frame.implicitSlotActive
		i32Program, i32Registers, i32Valid := frame.i32RegisterProgram, frame.i32Registers, frame.i32RegisterValid
		slotI32Values, slotI32Valid := frame.slotI32Values, frame.slotI32Valid
		slotFloatValues, slotFloatKinds, slotFloatValid := frame.slotFloatValues, frame.slotFloatKinds, frame.slotFloatValid
		frame.i32RegisterProgram, frame.i32Registers, frame.i32RegisterValid = nil, nil, nil
		frame.implicitSlotActive = nil
		frame.slotI32Values, frame.slotI32Valid = nil, nil
		frame.slotFloatValues, frame.slotFloatKinds, frame.slotFloatValid = nil, nil, nil
		frame.env = nil
		frame.transientScopeBase = 0
		frame.arrayOwnershipParent = nil
		vm.restoreSelfFastMinimalFrameSlot0(frame, slots)
		vm.releaseImplicitSlotActiveFrame(calleeImplicitSlotActive)
		vm.restoreImplicitSlotActiveFrame(implicitSlotActive)
		vm.restoreI32RegisterFrame(i32Program, i32Registers, i32Valid)
		if !reusesSlots {
			vm.restoreValueSlotSidecarFrames(slots, slotI32Values, slotI32Valid, slotFloatValues, slotFloatKinds, slotFloatValid)
		}
		vm.selfFastMinimal = vm.selfFastMinimal[:idx]
		return returnIP, nil, slots, env, frame.iterBase, frame.loopBase, false, true, bytecodeActiveLookupProgramState{}, true
	case bytecodeCallFrameKindSelfFast:
		idx := len(vm.selfFastCallFrames) - 1
		frame := &vm.selfFastCallFrames[idx]
		returnIP = frame.returnIP
		program = vm.currentProgram
		slots = frame.slots
		env = frame.env
		if frame.transientScopeBase < len(vm.activeTransientScopeEnvs) {
			vm.releaseActiveTransientRuntimeScopeEnvsToBase(frame.transientScopeBase)
		}
		frame.returnGenericNames = nil
		frame.returnCoercionFn = nil
		calleeImplicitSlotActive := vm.detachImplicitSlotActiveFrame()
		implicitSlotActive := frame.implicitSlotActive
		i32Program, i32Registers, i32Valid := frame.i32RegisterProgram, frame.i32Registers, frame.i32RegisterValid
		slotI32Values, slotI32Valid := frame.slotI32Values, frame.slotI32Valid
		slotFloatValues, slotFloatKinds, slotFloatValid := frame.slotFloatValues, frame.slotFloatKinds, frame.slotFloatValid
		frame.i32RegisterProgram, frame.i32Registers, frame.i32RegisterValid = nil, nil, nil
		frame.implicitSlotActive = nil
		frame.slotI32Values, frame.slotI32Valid = nil, nil
		frame.slotFloatValues, frame.slotFloatKinds, frame.slotFloatValid = nil, nil, nil
		frame.env = nil
		frame.transientScopeBase = 0
		frame.arrayOwnershipParent = nil
		iterBase = frame.iterBase
		loopBase = frame.loopBase
		hasImplicitReceiver = frame.hasImplicitReceiver
		selfFast = true
		vm.releaseImplicitSlotActiveFrame(calleeImplicitSlotActive)
		vm.restoreImplicitSlotActiveFrame(implicitSlotActive)
		vm.restoreI32RegisterFrame(i32Program, i32Registers, i32Valid)
		vm.restoreValueSlotSidecarFrames(slots, slotI32Values, slotI32Valid, slotFloatValues, slotFloatKinds, slotFloatValid)
		vm.selfFastCallFrames = vm.selfFastCallFrames[:idx]
		return returnIP, nil, slots, env, iterBase, loopBase, hasImplicitReceiver, selfFast, bytecodeActiveLookupProgramState{}, true
	default:
		idx := len(vm.callFrames) - 1
		frame := &vm.callFrames[idx]
		returnIP = frame.returnIP
		program = frame.program
		slots = frame.slots
		env = frame.env
		activeLookup = frame.activeLookup
		if frame.transientScopeBase < len(vm.activeTransientScopeEnvs) {
			vm.releaseActiveTransientRuntimeScopeEnvsToBase(frame.transientScopeBase)
		}
		frame.returnGenericNames = nil
		frame.returnCoercionFn = nil
		frame.activeLookup = bytecodeActiveLookupProgramState{}
		calleeImplicitSlotActive := vm.detachImplicitSlotActiveFrame()
		implicitSlotActive := frame.implicitSlotActive
		i32Program, i32Registers, i32Valid := frame.i32RegisterProgram, frame.i32Registers, frame.i32RegisterValid
		slotI32Values, slotI32Valid := frame.slotI32Values, frame.slotI32Valid
		slotFloatValues, slotFloatKinds, slotFloatValid := frame.slotFloatValues, frame.slotFloatKinds, frame.slotFloatValid
		frame.i32RegisterProgram, frame.i32Registers, frame.i32RegisterValid = nil, nil, nil
		frame.implicitSlotActive = nil
		frame.slotI32Values, frame.slotI32Valid = nil, nil
		frame.slotFloatValues, frame.slotFloatKinds, frame.slotFloatValid = nil, nil, nil
		frame.env = nil
		frame.transientScopeBase = 0
		frame.arrayOwnershipParent = nil
		iterBase = frame.iterBase
		loopBase = frame.loopBase
		hasImplicitReceiver = frame.hasImplicitReceiver
		selfFast = frame.selfFast
		vm.releaseImplicitSlotActiveFrame(calleeImplicitSlotActive)
		vm.restoreImplicitSlotActiveFrame(implicitSlotActive)
		vm.restoreI32RegisterFrame(i32Program, i32Registers, i32Valid)
		vm.restoreValueSlotSidecarFrames(slots, slotI32Values, slotI32Valid, slotFloatValues, slotFloatKinds, slotFloatValid)
		vm.callFrames = vm.callFrames[:idx]
		return returnIP, program, slots, env, iterBase, loopBase, hasImplicitReceiver, selfFast, activeLookup, true
	}
}

func sameSlotFrame(left []runtime.Value, right []runtime.Value) bool {
	if len(left) != len(right) {
		return false
	}
	if len(left) == 0 {
		return left == nil && right == nil
	}
	return &left[0] == &right[0]
}
