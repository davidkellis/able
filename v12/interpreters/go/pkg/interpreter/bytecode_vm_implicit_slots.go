package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) ensureImplicitSlotActiveFrame(slot int) {
	if vm == nil || slot < 0 || slot >= len(vm.slots) {
		return
	}
	if len(vm.implicitSlotActive) >= len(vm.slots) {
		return
	}
	active := make([]bool, len(vm.slots))
	copy(active, vm.implicitSlotActive)
	vm.implicitSlotActive = active
}

func (vm *bytecodeVM) implicitSlotIsActive(slot int) bool {
	return vm != nil && slot >= 0 && slot < len(vm.implicitSlotActive) && vm.implicitSlotActive[slot]
}

func (vm *bytecodeVM) activateImplicitSlot(slot int) {
	vm.ensureImplicitSlotActiveFrame(slot)
	if slot >= 0 && slot < len(vm.implicitSlotActive) {
		vm.implicitSlotActive[slot] = true
	}
}

func (vm *bytecodeVM) detachImplicitSlotActiveFrame() []bool {
	if vm == nil {
		return nil
	}
	active := vm.implicitSlotActive
	vm.implicitSlotActive = nil
	return active
}

func (vm *bytecodeVM) restoreImplicitSlotActiveFrame(active []bool) {
	if vm == nil {
		return
	}
	if vm.implicitSlotActive != nil && len(vm.implicitSlotActive) > 0 {
		clear(vm.implicitSlotActive)
	}
	vm.implicitSlotActive = active
}

func (vm *bytecodeVM) releaseImplicitSlotActiveFrame(active []bool) {
	if len(active) > 0 {
		clear(active)
	}
}

func (vm *bytecodeVM) execLoadImplicitSlot(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode implicit slot load missing instruction")
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return fmt.Errorf("bytecode implicit slot out of range")
	}
	if vm.implicitSlotIsActive(instr.target) {
		vm.appendSlotStackValueChecked(instr.target)
		vm.ip++
		return nil
	}
	if vm.interp != nil && vm.interp.bytecodeStatsEnabled {
		vm.interp.recordBytecodeLoadNameLookupForName(instr.name)
	}
	val, err := vm.resolveCachedIdentifierName(vm.currentProgram, vm.ip, instr.name)
	if err != nil {
		if instr.node != nil {
			return vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return err
	}
	vm.appendStackValue(val)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execStoreImplicitSlot(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode implicit slot store missing instruction")
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return fmt.Errorf("bytecode implicit slot out of range")
	}
	if vm.implicitSlotIsActive(instr.target) {
		return vm.execStoreSlot(instr)
	}
	if vm.stackDepth() == 0 {
		return fmt.Errorf("bytecode stack underflow")
	}
	if vm.env != nil && vm.env.Has(instr.name) {
		return vm.execAssignExistingImplicitSlotName(instr)
	}
	vm.activateImplicitSlot(instr.target)
	return vm.execStoreSlot(instr)
}

func (vm *bytecodeVM) execAssignExistingImplicitSlotName(instr *bytecodeInstruction) error {
	val := vm.stackValue(vm.stackDepth() - 1)
	stackVal := val
	storeVal := val
	shouldAssign := true
	if instr.storeTyped && instr.typeExpr != nil {
		var err error
		storeVal, stackVal, shouldAssign, err = vm.typedSlotAssignmentValues(*instr, val)
		if err != nil {
			if instr.node != nil {
				return vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
			}
			return err
		}
	}
	if shouldAssign && !vm.env.AssignExisting(instr.name, vm.environmentValue(storeVal)) {
		return fmt.Errorf("Undefined variable '%s'", instr.name)
	}
	if instr.discardResult {
		vm.truncateStack(vm.stackDepth() - 1)
		vm.ip++
		return nil
	}
	if stackVal == nil {
		stackVal = runtime.NilValue{}
	}
	vm.setStackValue(vm.stackDepth()-1, stackVal)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execCompoundAssignImplicitSlot(instr bytecodeInstruction) error {
	if vm.implicitSlotIsActive(instr.target) {
		return vm.execCompoundAssignSlot(instr)
	}
	return vm.execAssignNameCompound(instr)
}
