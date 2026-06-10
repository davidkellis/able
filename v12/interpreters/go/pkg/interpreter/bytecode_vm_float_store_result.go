package interpreter

import "able/interpreter-go/pkg/runtime"

func (vm *bytecodeVM) finishStoreSlotFloatResult(instr *bytecodeInstruction, result runtime.Value, err error) error {
	if err != nil {
		if vm.interp != nil {
			err = vm.interp.wrapStandardRuntimeError(err)
		}
		if instr.node != nil && vm.interp != nil {
			return vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return err
	}
	storedValue := vm.storeFloatSlotValue(instr.target, result)
	if !instr.discardResult {
		vm.appendStackValue(bytecodeStackSnapshotValue(storedValue))
	}
	vm.ip++
	return nil
}

func (vm *bytecodeVM) finishStoreSlotFloatRawResult(instr *bytecodeInstruction, raw float64, kind runtime.FloatType) error {
	if instr.discardResult {
		vm.storeReusableNormalizedFloatSlotRawDiscard(instr.target, raw, kind)
		vm.ip++
		return nil
	}
	vm.storeReusableNormalizedFloatSlotRaw(instr.target, raw, kind)
	vm.appendStackValue(bytecodeNormalizedRawFloatSlotValue(raw, kind))
	vm.ip++
	return nil
}

func (vm *bytecodeVM) storeReusableFloatSlotRaw(target int, raw float64, kind runtime.FloatType) runtime.Value {
	if target < 0 || target >= len(vm.slots) {
		return bytecodeRawFloatSlotValue(raw, kind)
	}
	return vm.storeReusableNormalizedFloatSlotRaw(target, normalizeFloat(kind, raw), kind)
}

// storeReusableNormalizedFloatSlotRaw assumes target is in range and raw is already normalized for kind.
func (vm *bytecodeVM) storeReusableNormalizedFloatSlotRaw(target int, raw float64, kind runtime.FloatType) runtime.Value {
	slot := &vm.slots[target]
	vm.clearActiveValueSlotFloat(target)

	var storedValue runtime.Value
	switch current := (*slot).(type) {
	case bytecodeRawF32SlotValue, bytecodeRawF64SlotValue:
		bytecodeSetNormalizedRawFloatValue(slot, raw, kind)
		storedValue = *slot
	case *runtime.FloatValue:
		if current == nil {
			goto fallback
		}
		current.Val = raw
		current.TypeSuffix = kind
		storedValue = current
	default:
		goto fallback
	}
	goto finalize

fallback:
	if vm.ownedFloatSlots != nil {
		if cell := vm.ownedFloatSlots[slot]; cell != nil {
			cell.Val = raw
			cell.TypeSuffix = kind
			storedValue = cell
			*slot = cell
		} else {
			bytecodeSetNormalizedRawFloatValue(slot, raw, kind)
			storedValue = *slot
		}
	} else {
		bytecodeSetNormalizedRawFloatValue(slot, raw, kind)
		storedValue = *slot
	}

finalize:
	vm.clearActiveValueSlotI32(target)
	if vm.hasI32RegisterFrame() {
		vm.setI32RegisterValue(target, storedValue)
	}
	if target == 0 {
		vm.setSelfFastSlot0I32Value(storedValue)
	}
	return storedValue
}

func (vm *bytecodeVM) storeReusableNormalizedFloatSlotRawDiscard(target int, raw float64, kind runtime.FloatType) {
	slot := &vm.slots[target]
	vm.clearActiveValueSlotFloat(target)

	switch current := (*slot).(type) {
	case bytecodeRawF32SlotValue, bytecodeRawF64SlotValue:
		bytecodeSetNormalizedRawFloatValue(slot, raw, kind)
	case *runtime.FloatValue:
		if current == nil {
			goto fallback
		}
		current.Val = raw
		current.TypeSuffix = kind
	default:
		goto fallback
	}
	goto finalize

fallback:
	if vm.ownedFloatSlots != nil {
		if cell := vm.ownedFloatSlots[slot]; cell != nil {
			cell.Val = raw
			cell.TypeSuffix = kind
			*slot = cell
		} else {
			bytecodeSetNormalizedRawFloatValue(slot, raw, kind)
		}
	} else {
		bytecodeSetNormalizedRawFloatValue(slot, raw, kind)
	}

finalize:
	vm.clearActiveValueSlotI32(target)
}

func (vm *bytecodeVM) reusableOwnedFloatSlot(target int) (*runtime.FloatValue, bool) {
	if vm == nil || target < 0 || target >= len(vm.slots) {
		return nil, false
	}
	if cell, ok := vm.slots[target].(*runtime.FloatValue); ok && cell != nil {
		return cell, true
	}
	if vm.ownedFloatSlots != nil {
		if cell := vm.ownedFloatSlots[&vm.slots[target]]; cell != nil {
			return cell, true
		}
	}
	return nil, false
}
