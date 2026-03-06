package interpreter

import "able/interpreter-go/pkg/runtime"

func (vm *bytecodeVM) pushCallFrame(returnIP int, program *bytecodeProgram, slots []runtime.Value, env *runtime.Environment, iterBase int, loopBase int, hasImplicitReceiver bool, selfFast bool) {
	if vm == nil {
		return
	}
	if cap(vm.callFrames) == 0 {
		vm.callFrames = make([]bytecodeCallFrame, 0, 32)
	}
	vm.callFrames = append(vm.callFrames, bytecodeCallFrame{
		returnIP:            returnIP,
		program:             program,
		slots:               slots,
		env:                 env,
		iterBase:            iterBase,
		loopBase:            loopBase,
		hasImplicitReceiver: hasImplicitReceiver,
		selfFast:            selfFast,
	})
}

func (vm *bytecodeVM) popCallFrameFields() (returnIP int, program *bytecodeProgram, slots []runtime.Value, env *runtime.Environment, iterBase int, loopBase int, hasImplicitReceiver bool, selfFast bool, ok bool) {
	if vm == nil || len(vm.callFrames) == 0 {
		return 0, nil, nil, nil, 0, 0, false, false, false
	}
	idx := len(vm.callFrames) - 1
	frame := &vm.callFrames[idx]
	returnIP = frame.returnIP
	program = frame.program
	slots = frame.slots
	env = frame.env
	iterBase = frame.iterBase
	loopBase = frame.loopBase
	hasImplicitReceiver = frame.hasImplicitReceiver
	selfFast = frame.selfFast
	vm.callFrames = vm.callFrames[:idx]
	return returnIP, program, slots, env, iterBase, loopBase, hasImplicitReceiver, selfFast, true
}
