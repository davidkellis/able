package interpreter

import (
	"fmt"
	"math"

	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execStoreSlotIntMulConstAdd(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode int affine slot update missing instruction")
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return fmt.Errorf("bytecode slot out of range")
	}
	mulImmediate, hasImmediate := instr.intImmediate, instr.hasIntImmediate
	if !hasImmediate {
		mulImmediate, hasImmediate = bytecodeImmediateIntegerValue(instr.value)
	}
	if !hasImmediate {
		return fmt.Errorf("bytecode int affine slot update missing integer immediate")
	}
	baseIdx := len(vm.stack) - 2
	var base, addend runtime.Value
	switch instr.op {
	case bytecodeOpStoreSlotIntMulConstAdd:
		if len(vm.stack) < 2 {
			return fmt.Errorf("bytecode stack underflow")
		}
		base = vm.stack[baseIdx]
		addend = vm.stack[baseIdx+1]
	case bytecodeOpStoreSlotIntMulConstAddFromSlot:
		if len(vm.stack) < 1 {
			return fmt.Errorf("bytecode stack underflow")
		}
		baseIdx = len(vm.stack) - 1
		base = vm.slots[instr.target]
		addend = vm.stack[baseIdx]
	default:
		return fmt.Errorf("bytecode int affine slot update opcode %d unsupported", instr.op)
	}
	if instr.discardResult {
		if raw, handled, err := bytecodeIntMulConstAddI32RawFast(base, mulImmediate, addend); handled {
			if err != nil {
				if vm.interp != nil {
					err = vm.interp.wrapStandardRuntimeError(err)
				}
				if instr.node != nil && vm.interp != nil {
					return vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
				}
				return err
			}
			result := bytecodeRawI32SlotValue(raw)
			vm.stack = vm.stack[:baseIdx]
			vm.slots[instr.target] = result
			if instr.target == 0 {
				vm.setSelfFastSlot0I32Value(result)
			}
			vm.ip++
			return nil
		}
	}
	result, err := vm.storeSlotIntMulConstAddResult(base, mulImmediate, addend)
	if err != nil {
		if vm.interp != nil {
			err = vm.interp.wrapStandardRuntimeError(err)
		}
		if instr.node != nil && vm.interp != nil {
			return vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return err
	}
	result = bytecodeStackResultValue(result)
	vm.stack = vm.stack[:baseIdx]
	vm.slots[instr.target] = result
	if instr.target == 0 {
		vm.setSelfFastSlot0I32Value(result)
	}
	if !instr.discardResult {
		vm.stack = append(vm.stack, bytecodeSlotReadValue(result))
	}
	vm.ip++
	return nil
}

func (vm *bytecodeVM) storeSlotIntMulConstAddResult(base runtime.Value, mulImmediate runtime.IntegerValue, addend runtime.Value) (runtime.Value, error) {
	if fast, handled, err := bytecodeIntMulConstAddFast(base, mulImmediate, addend); handled {
		return fast, err
	}
	product, err := applyBinaryOperator(vm.interp, "*", base, mulImmediate)
	if err != nil {
		return nil, err
	}
	return applyBinaryOperator(vm.interp, "+", product, addend)
}

func bytecodeIntMulConstAddFast(base runtime.Value, mulImmediate runtime.IntegerValue, addend runtime.Value) (runtime.Value, bool, error) {
	mulRef := &mulImmediate
	if mulImmediate.TypeSuffix == runtime.IntegerI32 && mulRef.IsSmallRef() {
		baseRaw, baseOK := bytecodeDirectSmallI32Value(base)
		addRaw, addOK := bytecodeDirectSmallI32Value(addend)
		if baseOK && addOK {
			product, overflow := mulInt64Overflow(baseRaw, mulRef.Int64FastRef())
			if overflow {
				return nil, false, nil
			}
			result, overflow := addInt64Overflow(product, addRaw)
			if overflow {
				return nil, false, nil
			}
			if result < math.MinInt32 || result > math.MaxInt32 {
				return nil, true, newOverflowError("integer overflow")
			}
			if _, rawBase := base.(bytecodeRawI32SlotValue); rawBase {
				return bytecodeRawI32SlotValue(int32(result)), true, nil
			}
			return bytecodeBoxedIntegerI32Value(result), true, nil
		}
	}
	kind, baseRaw, addRaw, ok := bytecodeDirectSameTypeSmallIntPair(base, addend)
	if !ok || kind != mulImmediate.TypeSuffix {
		return nil, false, nil
	}
	if !mulRef.IsSmallRef() {
		return nil, false, nil
	}
	product, overflow := mulInt64Overflow(baseRaw, mulRef.Int64FastRef())
	if overflow {
		return nil, false, nil
	}
	result, overflow := addInt64Overflow(product, addRaw)
	if overflow {
		return nil, false, nil
	}
	if err := ensureFitsInt64Type(kind, result); err != nil {
		return nil, true, err
	}
	return boxedOrSmallIntegerValue(kind, result), true, nil
}

func bytecodeIntMulConstAddI32RawFast(base runtime.Value, mulImmediate runtime.IntegerValue, addend runtime.Value) (int32, bool, error) {
	mulRef := &mulImmediate
	if mulImmediate.TypeSuffix != runtime.IntegerI32 || !mulRef.IsSmallRef() {
		return 0, false, nil
	}
	baseRaw, baseOK := bytecodeDirectSmallI32Value(base)
	addRaw, addOK := bytecodeDirectSmallI32Value(addend)
	if !baseOK || !addOK {
		return 0, false, nil
	}
	product, overflow := mulInt64Overflow(baseRaw, mulRef.Int64FastRef())
	if overflow {
		return 0, false, nil
	}
	result, overflow := addInt64Overflow(product, addRaw)
	if overflow {
		return 0, false, nil
	}
	if result < math.MinInt32 || result > math.MaxInt32 {
		return 0, true, newOverflowError("integer overflow")
	}
	return int32(result), true, nil
}
