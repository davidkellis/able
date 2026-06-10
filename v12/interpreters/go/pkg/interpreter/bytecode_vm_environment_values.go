package interpreter

import "able/interpreter-go/pkg/runtime"

// bytecodeEnvironmentValue materializes VM-only raw carriers before they are
// persisted in an Environment, which can be observed by native and compiled
// runtime boundaries after the current bytecode step has completed.
func bytecodeEnvironmentValue(value runtime.Value) runtime.Value {
	return bytecodeMaterializeRawValue(bytecodeSlotReadValue(value))
}

func (vm *bytecodeVM) environmentValue(value runtime.Value) runtime.Value {
	return vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonEnvironment, bytecodeSlotReadValue(value))
}
