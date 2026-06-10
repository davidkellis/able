package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) acquireTransientRuntimeScopeEnv(parent *runtime.Environment) *runtime.Environment {
	return i.acquireTransientRuntimeScopeEnvWithCapacity(parent, 0)
}

func (i *Interpreter) acquireTransientRuntimeScopeEnvWithCapacity(parent *runtime.Environment, valueCapacity int) *runtime.Environment {
	if i == nil {
		return runtime.NewEnvironmentWithValueCapacity(parent, valueCapacity)
	}
	if pooled := i.transientRuntimeScopeEnvPool.Get(); pooled != nil {
		if env, ok := pooled.(*runtime.Environment); ok && env != nil {
			env.ResetForSingleBindingReuse(parent, valueCapacity, "", nil)
			return env
		}
	}
	return runtime.NewEnvironmentWithValueCapacity(parent, valueCapacity)
}

func (i *Interpreter) releaseTransientRuntimeScopeEnv(env *runtime.Environment) {
	if i == nil || env == nil {
		return
	}
	i.transientRuntimeScopeEnvPool.Put(env)
}

func (vm *bytecodeVM) enterRuntimeScope(transient bool) error {
	if vm == nil || vm.interp == nil || vm.env == nil {
		return fmt.Errorf("bytecode vm missing environment")
	}
	if transient {
		env := vm.interp.acquireTransientRuntimeScopeEnv(vm.env)
		vm.env = env
		vm.activeTransientScopeEnvs = append(vm.activeTransientScopeEnvs, env)
		return nil
	}
	vm.env = runtime.NewEnvironment(vm.env)
	return nil
}

func (vm *bytecodeVM) exitRuntimeScopes(count int) error {
	if vm == nil {
		return fmt.Errorf("bytecode vm missing state")
	}
	for idx := 0; idx < count; idx++ {
		if vm.env == nil || vm.env.Parent() == nil {
			return fmt.Errorf("bytecode scope underflow")
		}
		parent := vm.env.Parent()
		if active := len(vm.activeTransientScopeEnvs); active > 0 && vm.activeTransientScopeEnvs[active-1] == vm.env {
			transient := vm.env
			vm.activeTransientScopeEnvs[active-1] = nil
			vm.activeTransientScopeEnvs = vm.activeTransientScopeEnvs[:active-1]
			if vm.interp != nil {
				vm.interp.releaseTransientRuntimeScopeEnv(transient)
			}
		}
		vm.env = parent
	}
	return nil
}

func (vm *bytecodeVM) releaseActiveTransientRuntimeScopeEnvsToBase(base int) {
	if vm == nil {
		return
	}
	if base < 0 {
		base = 0
	}
	if base >= len(vm.activeTransientScopeEnvs) {
		return
	}
	for len(vm.activeTransientScopeEnvs) > base {
		last := len(vm.activeTransientScopeEnvs) - 1
		env := vm.activeTransientScopeEnvs[last]
		vm.activeTransientScopeEnvs[last] = nil
		vm.activeTransientScopeEnvs = vm.activeTransientScopeEnvs[:last]
		if vm.interp != nil && env != nil {
			vm.interp.releaseTransientRuntimeScopeEnv(env)
		}
	}
}

func (vm *bytecodeVM) releaseAllActiveTransientRuntimeScopeEnvs() {
	vm.releaseActiveTransientRuntimeScopeEnvsToBase(0)
}
