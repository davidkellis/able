package interpreter

import "able/interpreter-go/pkg/runtime"

func (i *Interpreter) runtimeDataFromEnv(env *runtime.Environment) any {
	if env == nil {
		return nil
	}
	if i == nil || !i.envSingleThread {
		return env.RuntimeData()
	}
	state := env.RuntimeDataStateID()
	rev := env.RuntimeDataRevision()
	envRev := env.RevisionSingleThread()
	if i.runtimeDataCacheKnown && i.runtimeDataCacheEnv == env && i.runtimeDataCacheState == state && i.runtimeDataCacheRev == rev && i.runtimeDataCacheEnvRev == envRev {
		return i.runtimeDataCacheValue
	}
	data := env.RuntimeData()
	i.runtimeDataCacheEnv = env
	i.runtimeDataCacheState = state
	i.runtimeDataCacheValue = data
	i.runtimeDataCacheRev = rev
	i.runtimeDataCacheEnvRev = envRev
	i.runtimeDataCacheKnown = true
	return data
}

func (vm *bytecodeVM) runtimeData() any {
	if vm == nil || vm.env == nil {
		return nil
	}
	env := vm.env
	singleThread := vm.bytecodeSingleThread()
	state := env.RuntimeDataStateID()
	rev := env.RuntimeDataRevision()
	envRev := env.RevisionWithHint(singleThread)
	if vm.runtimeDataCacheKnown && vm.runtimeDataCacheEnv == env && vm.runtimeDataCacheState == state && vm.runtimeDataCacheRev == rev && vm.runtimeDataCacheEnvRev == envRev {
		return vm.runtimeDataCacheValue
	}
	data := env.RuntimeData()
	vm.runtimeDataCacheEnv = env
	vm.runtimeDataCacheState = state
	vm.runtimeDataCacheValue = data
	vm.runtimeDataCacheRev = rev
	vm.runtimeDataCacheEnvRev = envRev
	vm.runtimeDataCacheKnown = true
	return data
}

func (vm *bytecodeVM) hasRuntimeData() bool {
	return vm.runtimeData() != nil
}

func (vm *bytecodeVM) noRuntimeDataGlobalAndMethodVersions() (uint64, uint64, bool) {
	if vm == nil || vm.env == nil || vm.interp == nil {
		return 0, 0, false
	}
	env := vm.env
	interp := vm.interp
	singleThread := interp.envSingleThread
	state := env.RuntimeDataStateID()
	rev := env.RuntimeDataRevision()
	envRev := env.RevisionWithHint(singleThread)
	dataKnown := vm.runtimeDataCacheKnown &&
		vm.runtimeDataCacheEnv == env &&
		vm.runtimeDataCacheState == state &&
		vm.runtimeDataCacheRev == rev &&
		vm.runtimeDataCacheEnvRev == envRev
	data := vm.runtimeDataCacheValue
	if !dataKnown {
		data = env.RuntimeData()
		vm.runtimeDataCacheEnv = env
		vm.runtimeDataCacheState = state
		vm.runtimeDataCacheValue = data
		vm.runtimeDataCacheRev = rev
		vm.runtimeDataCacheEnvRev = envRev
		vm.runtimeDataCacheKnown = true
	}
	if data != nil {
		return 0, 0, false
	}
	globalRevision := uint64(0)
	if interp.global != nil {
		if interp.global == env {
			globalRevision = envRev
		} else if singleThread {
			globalRevision = interp.global.RevisionSingleThread()
		} else {
			globalRevision = interp.global.RevisionWithHint(false)
		}
	}
	if singleThread {
		return globalRevision, interp.methodCacheVersion, true
	}
	return globalRevision, interp.currentMethodCacheVersion(), true
}
