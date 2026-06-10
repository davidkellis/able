package interpreter

import "able/interpreter-go/pkg/runtime"

func (vm *bytecodeVM) bytecodeSingleThread() bool {
	return vm != nil && vm.interp != nil && vm.interp.envSingleThread
}

func (vm *bytecodeVM) bytecodeEnvRevision(env *runtime.Environment) uint64 {
	return bytecodeEnvironmentRevision(env, vm.bytecodeSingleThread())
}

func bytecodeEnvironmentRevision(env *runtime.Environment, singleThread bool) uint64 {
	if env == nil {
		return 0
	}
	if singleThread {
		return env.RevisionSingleThread()
	}
	return env.RevisionWithHint(false)
}

func (vm *bytecodeVM) bytecodeLookupWithOwner(name string) (runtime.Value, *runtime.Environment, uint64, bool) {
	if vm == nil || vm.env == nil {
		return nil, nil, 0, false
	}
	return vm.env.LookupWithOwnerAndRevisionHint(name, vm.bytecodeSingleThread())
}

func (vm *bytecodeVM) bytecodeGlobalRevision() uint64 {
	if vm == nil || vm.interp == nil || vm.interp.global == nil {
		return 0
	}
	if vm.interp.envSingleThread {
		return vm.interp.global.RevisionSingleThread()
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

func (vm *bytecodeVM) bytecodeGlobalAndMethodVersions() (uint64, uint64) {
	if vm == nil || vm.interp == nil {
		return 0, 0
	}
	interp := vm.interp
	globalRevision := uint64(0)
	if interp.global != nil {
		if interp.envSingleThread {
			globalRevision = interp.global.RevisionSingleThread()
		} else {
			globalRevision = interp.global.RevisionWithHint(false)
		}
	}
	if interp.envSingleThread {
		return globalRevision, interp.methodCacheVersion
	}
	return globalRevision, interp.currentMethodCacheVersion()
}

func (vm *bytecodeVM) bytecodeEnvGlobalAndMethodVersions(env *runtime.Environment) (uint64, uint64, uint64) {
	if vm == nil || vm.interp == nil {
		return 0, 0, 0
	}
	interp := vm.interp
	singleThread := interp.envSingleThread
	envRevision := uint64(0)
	if env != nil {
		if singleThread {
			envRevision = env.RevisionSingleThread()
		} else {
			envRevision = env.RevisionWithHint(false)
		}
	}
	globalRevision := uint64(0)
	if interp.global != nil {
		if singleThread {
			globalRevision = interp.global.RevisionSingleThread()
		} else {
			globalRevision = interp.global.RevisionWithHint(false)
		}
	}
	if singleThread {
		return envRevision, globalRevision, interp.methodCacheVersion
	}
	return envRevision, globalRevision, interp.currentMethodCacheVersion()
}
