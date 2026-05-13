package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execArrayIndexSlotMutation(instr *bytecodeInstruction) error {
	if instr != nil && instr.op == bytecodeOpArrayIndexSwapSlot {
		return vm.execArrayIndexSwapSlot(instr)
	}
	return vm.execArrayIndexSetSlot(instr)
}

func (vm *bytecodeVM) execArrayIndexSwapSlot(instr *bytecodeInstruction) error {
	if vm == nil || vm.interp == nil || instr == nil {
		return fmt.Errorf("bytecode array index swap slot missing VM or instruction")
	}
	objSlot, firstSlot, secondSlot := instr.argCount, instr.loopBreak, instr.loopContinue
	if objSlot < 0 || objSlot >= len(vm.slots) || firstSlot < 0 || firstSlot >= len(vm.slots) || secondSlot < 0 || secondSlot >= len(vm.slots) {
		return fmt.Errorf("bytecode array index swap slot out of range")
	}
	obj := vm.slots[objSlot]
	firstIdx := vm.slots[firstSlot]
	secondIdx := vm.slots[secondSlot]
	result, err := vm.resolveArrayIndexSwapSlot(instr, obj, firstIdx, secondIdx)
	if err != nil {
		return vm.attachArrayIndexSwapSlotError(err, instr)
	}
	vm.stack = append(vm.stack, bytecodeStackResultValue(result))
	vm.ip++
	return nil
}

func (vm *bytecodeVM) resolveArrayIndexSwapSlot(instr *bytecodeInstruction, obj runtime.Value, firstIdx runtime.Value, secondIdx runtime.Value) (runtime.Value, error) {
	if vm.interp.canUseDirectArrayIndexGetFastPath() && vm.interp.canUseDirectArrayIndexSetFastPath() {
		if arr, ok := obj.(*runtime.ArrayValue); ok && arr != nil {
			if result, handled, err := vm.resolveDirectArrayIndexSwapSlot(instr, arr, firstIdx, secondIdx); handled || err != nil {
				return result, err
			}
		}
	}
	return vm.resolveGenericArrayIndexSwapSlot(instr, obj, firstIdx, secondIdx)
}

func (vm *bytecodeVM) resolveDirectArrayIndexSwapSlot(instr *bytecodeInstruction, arr *runtime.ArrayValue, firstIdx runtime.Value, secondIdx runtime.Value) (runtime.Value, bool, error) {
	if first, ok := bytecodeDirectSmallArrayIndex(firstIdx); ok {
		if second, ok := bytecodeDirectSmallArrayIndex(secondIdx); ok {
			if state, tracked := bytecodeTrackedArrayState(arr); tracked {
				result, err := vm.resolveTrackedSmallArrayIndexSwapSlot(instr, arr, state, first, second)
				return result, true, err
			}
		}
	}
	first, ok, err := bytecodeDirectArrayIndex(firstIdx)
	if err != nil || !ok {
		return nil, ok, err
	}
	second, ok, err := bytecodeDirectArrayIndex(secondIdx)
	if err != nil || !ok {
		return nil, ok, err
	}
	left, err := vm.resolveDirectArrayIndexGetAt(arr, first)
	if err != nil {
		return nil, true, err
	}
	left, err = vm.castArrayIndexSwapSlotValue(instr, left)
	if err != nil {
		return nil, true, err
	}
	right, err := vm.resolveDirectArrayIndexGetAt(arr, second)
	if err != nil {
		return nil, true, err
	}
	right, err = vm.castArrayIndexSwapSlotValue(instr, right)
	if err != nil {
		return nil, true, err
	}
	if _, err := vm.resolveDirectArrayIndexSetAt(arr, first, right); err != nil {
		return nil, true, err
	}
	if _, err := vm.resolveDirectArrayIndexSetAt(arr, second, left); err != nil {
		return nil, true, err
	}
	return left, true, nil
}

func (vm *bytecodeVM) resolveTrackedSmallArrayIndexSwapSlot(instr *bytecodeInstruction, arr *runtime.ArrayValue, state *runtime.ArrayState, first int, second int) (runtime.Value, error) {
	left := vm.trackedArrayIndexSwapSlotValue(state, first)
	left, err := vm.castArrayIndexSwapSlotValue(instr, left)
	if err != nil {
		return nil, err
	}
	right := vm.trackedArrayIndexSwapSlotValue(state, second)
	right, err = vm.castArrayIndexSwapSlotValue(instr, right)
	if err != nil {
		return nil, err
	}
	if first < 0 || first >= len(state.Values) || second < 0 || second >= len(state.Values) {
		return nil, fmt.Errorf("Array index out of bounds")
	}
	state.Values[first] = right
	state.Values[second] = left
	vm.syncTrackedArrayIndexSwapSlot(arr, state, first, right, second, left)
	return left, nil
}

func (vm *bytecodeVM) trackedArrayIndexSwapSlotValue(state *runtime.ArrayState, idx int) runtime.Value {
	if state == nil || vm == nil || vm.interp == nil {
		return runtime.NilValue{}
	}
	if idx < 0 || idx >= len(state.Values) {
		return vm.interp.makeIndexErrorValue(idx, len(state.Values))
	}
	value := state.Values[idx]
	if value == nil {
		return vm.interp.makeIndexErrorValue(idx, len(state.Values))
	}
	return value
}

func (vm *bytecodeVM) syncTrackedArrayIndexSwapSlot(arr *runtime.ArrayValue, state *runtime.ArrayState, first int, firstValue runtime.Value, second int, secondValue runtime.Value) {
	if vm == nil || vm.interp == nil || arr == nil || state == nil {
		return
	}
	if bytecodeSyncUnaliasedTrackedArrayWrite(arr, state, first, firstValue) {
		bytecodeSyncUnaliasedTrackedArrayWrite(arr, state, second, secondValue)
		return
	}
	vm.interp.syncTrackedArrayWrite(arr, state, first, firstValue)
	vm.interp.syncTrackedArrayWrite(arr, state, second, secondValue)
}

func (vm *bytecodeVM) resolveGenericArrayIndexSwapSlot(instr *bytecodeInstruction, obj runtime.Value, firstIdx runtime.Value, secondIdx runtime.Value) (runtime.Value, error) {
	left, err := vm.resolveIndexGet(obj, firstIdx)
	if err != nil {
		return nil, err
	}
	left, err = vm.castArrayIndexSwapSlotValue(instr, left)
	if err != nil {
		return nil, err
	}
	right, err := vm.resolveIndexGet(obj, secondIdx)
	if err != nil {
		return nil, err
	}
	right, err = vm.castArrayIndexSwapSlotValue(instr, right)
	if err != nil {
		return nil, err
	}
	if _, err := vm.resolveIndexSet(obj, firstIdx, right, ast.AssignmentAssign, "", false); err != nil {
		return nil, err
	}
	result, err := vm.resolveIndexSet(obj, secondIdx, left, ast.AssignmentAssign, "", false)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (vm *bytecodeVM) castArrayIndexSwapSlotValue(instr *bytecodeInstruction, value runtime.Value) (runtime.Value, error) {
	if instr == nil || instr.typeExpr == nil {
		return value, nil
	}
	return vm.interp.castValueToType(instr.typeExpr, value)
}

func (vm *bytecodeVM) attachArrayIndexSwapSlotError(err error, instr *bytecodeInstruction) error {
	if err == nil {
		return nil
	}
	err = vm.interp.wrapStandardRuntimeError(err)
	if instr != nil && instr.node != nil {
		err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
	}
	return err
}
