package interpreter

import "able/interpreter-go/pkg/runtime"

func (vm *bytecodeVM) bytecodeSingleThread() bool {
	return vm != nil && vm.interp != nil && vm.interp.envSingleThread
}

func (vm *bytecodeVM) bytecodeEnvRevision(env *runtime.Environment) uint64 {
	if env == nil {
		return 0
	}
	return env.RevisionWithHint(vm.bytecodeSingleThread())
}

func (vm *bytecodeVM) bytecodeGlobalRevision() uint64 {
	if vm == nil || vm.interp == nil || vm.interp.global == nil {
		return 0
	}
	return vm.interp.global.RevisionWithHint(vm.interp.envSingleThread)
}

func (vm *bytecodeVM) bytecodeMethodCacheVersion() uint64 {
	if vm == nil || vm.interp == nil {
		return 0
	}
	if vm.interp.envSingleThread {
		return vm.interp.methodCacheVersion
	}
	return vm.interp.currentMethodCacheVersion()
}
