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
	vm.ip = 0
	vm.currentProgram = nil
	vm.clearSelfFastSlot0I32()

	if len(vm.stack) > 0 {
		clear(vm.stack)
		vm.stack = vm.stack[:0]
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
	if vm.slots != nil {
		clear(vm.slots)
		vm.slots = nil
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
			frame.returnGenericNames = nil
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
			frame.returnGenericNames = nil
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
			frame.slot0I32Raw = 0
			frame.slot0I32Valid = false
			frame.reusesSlots = false
		}
		vm.selfFastMinimal = vm.selfFastMinimal[:0]
	}
	vm.selfFastMinimalSuffix = 0
	if len(vm.callFrameKinds) > 0 {
		clear(vm.callFrameKinds)
		vm.callFrameKinds = vm.callFrameKinds[:0]
	}
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
	vm.resetForRun(vm.interp, nil)
	vm.interp = nil
}
