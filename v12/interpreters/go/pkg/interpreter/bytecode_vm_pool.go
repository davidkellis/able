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

	if len(vm.stack) > 0 {
		clear(vm.stack)
		vm.stack = vm.stack[:0]
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
			frame.iterBase = 0
			frame.loopBase = 0
			frame.hasImplicitReceiver = false
			frame.selfFast = false
		}
		vm.callFrames = vm.callFrames[:0]
	}

	vm.globalLookupCache = nil
	vm.scopeLookupCache = nil
	vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{}
	vm.memberMethodCache = nil
	vm.memberMethodHot = bytecodeInlineMemberMethodCacheEntry{}
	vm.indexMethodCache = nil
	vm.indexMethodHot = bytecodeInlineIndexMethodCacheEntry{}
	// Keep immutable per-program decode/validation caches across pooled runs.
	// They are keyed by bytecodeProgram pointers and refreshed on length mismatch.
}

func (vm *bytecodeVM) resetForPool() {
	if vm == nil {
		return
	}
	vm.resetForRun(vm.interp, nil)
	vm.interp = nil
}
