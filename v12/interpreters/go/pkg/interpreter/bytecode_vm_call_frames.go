package interpreter

import "able/interpreter-go/pkg/runtime"

func (vm *bytecodeVM) hasCallFrames() bool {
	return vm != nil && (len(vm.callFrameKinds) > 0 || vm.selfFastMinimalSuffix > 0)
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

func (vm *bytecodeVM) pushSelfFastMinimalCallFrame(returnIP int, slots []runtime.Value) {
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
	frame.slots = slots
	vm.selfFastMinimalSuffix++
}

func (vm *bytecodeVM) pushCallFrame(returnIP int, program *bytecodeProgram, slots []runtime.Value, env *runtime.Environment, returnGenericNames map[string]struct{}, iterBase int, loopBase int, hasImplicitReceiver bool, selfFast bool) {
	if vm == nil {
		return
	}
	if selfFast {
		if returnGenericNames == nil && iterBase == 0 && loopBase == 0 && !hasImplicitReceiver {
			vm.pushSelfFastMinimalCallFrame(returnIP, slots)
			return
		}
		vm.materializeSelfFastMinimalSuffixKinds()
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
		frame.slots = slots
		frame.returnGenericNames = returnGenericNames
		frame.iterBase = iterBase
		frame.loopBase = loopBase
		frame.hasImplicitReceiver = hasImplicitReceiver
		return
	}
	vm.materializeSelfFastMinimalSuffixKinds()
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
	frame.program = program
	frame.slots = slots
	frame.env = env
	frame.returnGenericNames = returnGenericNames
	frame.iterBase = iterBase
	frame.loopBase = loopBase
	frame.hasImplicitReceiver = hasImplicitReceiver
	frame.selfFast = selfFast
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

func (vm *bytecodeVM) popCallFrameFields() (returnIP int, program *bytecodeProgram, slots []runtime.Value, env *runtime.Environment, iterBase int, loopBase int, hasImplicitReceiver bool, selfFast bool, ok bool) {
	if vm == nil {
		return 0, nil, nil, nil, 0, 0, false, false, false
	}
	if vm.selfFastMinimalSuffix > 0 {
		idx := len(vm.selfFastMinimal) - 1
		frame := &vm.selfFastMinimal[idx]
		returnIP = frame.returnIP
		slots = frame.slots
		vm.selfFastMinimal = vm.selfFastMinimal[:idx]
		vm.selfFastMinimalSuffix--
		return returnIP, nil, slots, nil, 0, 0, false, true, true
	}
	if len(vm.callFrameKinds) == 0 {
		return 0, nil, nil, nil, 0, 0, false, false, false
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
		vm.selfFastMinimal = vm.selfFastMinimal[:idx]
		return returnIP, nil, slots, nil, 0, 0, false, true, true
	case bytecodeCallFrameKindSelfFast:
		idx := len(vm.selfFastCallFrames) - 1
		frame := &vm.selfFastCallFrames[idx]
		returnIP = frame.returnIP
		program = vm.currentProgram
		slots = frame.slots
		env = vm.env
		frame.returnGenericNames = nil
		iterBase = frame.iterBase
		loopBase = frame.loopBase
		hasImplicitReceiver = frame.hasImplicitReceiver
		selfFast = true
		vm.selfFastCallFrames = vm.selfFastCallFrames[:idx]
		return returnIP, nil, slots, nil, iterBase, loopBase, hasImplicitReceiver, selfFast, true
	default:
		idx := len(vm.callFrames) - 1
		frame := &vm.callFrames[idx]
		returnIP = frame.returnIP
		program = frame.program
		slots = frame.slots
		env = frame.env
		frame.returnGenericNames = nil
		iterBase = frame.iterBase
		loopBase = frame.loopBase
		hasImplicitReceiver = frame.hasImplicitReceiver
		selfFast = frame.selfFast
		vm.callFrames = vm.callFrames[:idx]
		return returnIP, program, slots, env, iterBase, loopBase, hasImplicitReceiver, selfFast, true
	}
}
