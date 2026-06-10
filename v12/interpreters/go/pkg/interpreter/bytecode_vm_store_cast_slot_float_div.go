package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execStoreSlotCastSlotFloatConstDiv(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode cast-slot-float-const div store missing instruction")
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return fmt.Errorf("bytecode slot out of range")
	}
	sourceSlot := instr.argCount
	if sourceSlot < 0 || sourceSlot >= len(vm.slots) {
		return fmt.Errorf("bytecode source slot out of range")
	}
	targetKind, ok := bytecodeFloatCastTargetKind(instr.typeExpr)
	if !ok {
		return fmt.Errorf("bytecode cast-slot-float-const div store missing float target type")
	}
	right, ok := bytecodeFloatImmediateValue(instr.value)
	if !ok {
		return fmt.Errorf("bytecode cast-slot-float-const div store missing float immediate")
	}
	if instr.discardResult {
		if err, handled := vm.execStoreSlotCastSlotFloatConstDivDiscardFast(instr, sourceSlot, targetKind, right); handled {
			return err
		}
	}
	if resultVal, resultKind, ok, err := vm.bytecodeCastSlotFloatConstDivRawFast(sourceSlot, targetKind, right); ok || err != nil {
		if err != nil {
			return vm.finishStoreSlotFloatResult(instr, nil, err)
		}
		return vm.finishStoreSlotFloatRawResult(instr, resultVal, resultKind)
	}
	rawLeft := vm.slotRuntimeValue(sourceSlot)
	casted, err := vm.interp.castValueToType(ast.Ty(string(targetKind)), rawLeft)
	if err != nil {
		return vm.finishStoreSlotFloatResult(instr, nil, err)
	}
	result, err := applyBinaryOperator(vm.interp, "/", casted, right)
	return vm.finishStoreSlotFloatResult(instr, result, err)
}

func (vm *bytecodeVM) execStoreSlotCastSlotFloatConstDivDiscardFast(instr *bytecodeInstruction, sourceSlot int, targetKind runtime.FloatType, right runtime.FloatValue) (error, bool) {
	if targetKind == runtime.FloatF64 && right.TypeSuffix == runtime.FloatF64 {
		if sourceCell, ok := vm.slots[sourceSlot].(*bytecodeRawI64SlotCell); ok && sourceCell != nil {
			if targetCell, ok := vm.slots[instr.target].(*runtime.FloatValue); ok && targetCell != nil {
				targetCell.Val = normalizeFloat(runtime.FloatF64, float64(sourceCell.Val)/right.Val)
				targetCell.TypeSuffix = runtime.FloatF64
				vm.clearActiveValueSlotI32(instr.target)
				vm.clearActiveValueSlotFloat(instr.target)
				if vm.hasI32RegisterFrame() {
					vm.setI32RegisterValue(instr.target, targetCell)
				}
				if instr.target == 0 {
					vm.setSelfFastSlot0I32Value(targetCell)
				}
				vm.ip++
				return nil, true
			}
		}
	}
	resultVal, resultKind, ok, err := vm.bytecodeCastSlotFloatConstDivRawFast(sourceSlot, targetKind, right)
	if !ok || err != nil {
		return err, ok
	}
	vm.storeReusableFloatSlotRaw(instr.target, resultVal, resultKind)
	vm.ip++
	return nil, true
}
