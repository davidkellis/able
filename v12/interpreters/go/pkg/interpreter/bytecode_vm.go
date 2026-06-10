package interpreter

import (
	"able/interpreter-go/pkg/runtime"
)

type bytecodeOp int

const (
	bytecodeOpConst bytecodeOp = iota
	bytecodeOpLoadName
	bytecodeOpLoadStaticReceiver
	bytecodeOpDeclareName
	bytecodeOpAssignName
	bytecodeOpAssignPattern
	bytecodeOpAssignNameCompound
	bytecodeOpDup
	bytecodeOpPop
	bytecodeOpBinary
	bytecodeOpBinaryIntAdd
	bytecodeOpBinaryIntSub
	bytecodeOpBinaryIntLessEqual
	bytecodeOpBinaryIntDivCast
	bytecodeOpBinaryCastSlotFloatConstDiv
	bytecodeOpBinaryFloatMulSlotConst
	bytecodeOpBinaryIntAddSlotConst
	bytecodeOpBinaryIntSubSlotConst
	bytecodeOpBinaryIntMulSlotConst
	bytecodeOpBinaryIntModSlotConst
	bytecodeOpBinaryIntLessEqualSlotConst
	bytecodeOpBinaryIntCompareSlotConst
	bytecodeOpUnary
	bytecodeOpRange
	bytecodeOpCast
	bytecodeOpStringInterpolation
	bytecodeOpPropagation
	bytecodeOpOrElse
	bytecodeOpSpawn
	bytecodeOpAwait
	bytecodeOpImplicitMember
	bytecodeOpImplicitMemberSet
	bytecodeOpIteratorLiteral
	bytecodeOpBreakpoint
	bytecodeOpPlaceholderLambda
	bytecodeOpPlaceholderValue
	bytecodeOpIterInit
	bytecodeOpIterNext
	bytecodeOpIterClose
	bytecodeOpBindPattern
	bytecodeOpYield
	bytecodeOpMakeFunction
	bytecodeOpDefineFunction
	bytecodeOpDefineStruct
	bytecodeOpDefineUnion
	bytecodeOpDefineTypeAlias
	bytecodeOpDefineMethods
	bytecodeOpDefineInterface
	bytecodeOpDefineImplementation
	bytecodeOpDefineExtern
	bytecodeOpImport
	bytecodeOpDynImport
	bytecodeOpStructLiteral
	bytecodeOpStructLiteralNamedFast
	bytecodeOpMapLiteral
	bytecodeOpArrayLiteral
	bytecodeOpIndexGet
	bytecodeOpIndexSet
	bytecodeOpArrayIndexGetSlot
	bytecodeOpArrayIndexSetSlot
	bytecodeOpArrayIndexSwapSlot
	bytecodeOpArrayReadSlot
	bytecodeOpArrayReadSlotI32
	bytecodeOpArraySlotSwapSlot
	bytecodeOpForLoop
	bytecodeOpCall
	bytecodeOpCallName
	bytecodeOpCallMember
	bytecodeOpCallGenericUnionMember
	bytecodeOpCallStaticMember
	bytecodeOpCallMemberArrayGet
	bytecodeOpCallMemberNext
	bytecodeOpCallMemberArrayNew
	bytecodeOpCallMemberArraySlot
	bytecodeOpTryArrayPushF64AffineProduct
	bytecodeOpTryArrayPushF64NestedGet
	bytecodeOpTryFloatUpdatePair
	bytecodeOpMemberAccess
	bytecodeOpMemberSet
	bytecodeOpMatch
	bytecodeOpRescue
	bytecodeOpRaise
	bytecodeOpEnsure
	bytecodeOpEnsureEnd
	bytecodeOpRethrow
	bytecodeOpPipe
	bytecodeOpBreakLabel
	bytecodeOpBreakSignal
	bytecodeOpContinueSignal
	bytecodeOpJump
	bytecodeOpJumpIfFalse
	bytecodeOpJumpIfBoolSlotFalse
	bytecodeOpJumpIfIntLessEqualSlotConstFalse
	bytecodeOpJumpIfIntCompareSlotConstFalse
	bytecodeOpJumpIfArrayReadSlotCompareSlotFalse
	bytecodeOpJumpIfArrayIndexSlotCompareSlotFalse
	bytecodeOpJumpIfFloatMulAddMulCompareConstFalse
	bytecodeOpJumpIfFloatAddCompareConstFalse
	bytecodeOpJumpIfIntCompareSlotFalse
	bytecodeOpReturnIfIntLessEqualSlotConst
	bytecodeOpReturnConstIfIntLessEqualSlotConst
	bytecodeOpJumpIfNil
	bytecodeOpJumpIfNotNil
	bytecodeOpJumpIfNotTypedPattern
	bytecodeOpMatchNoClause
	bytecodeOpLoopEnter
	bytecodeOpLoopExit
	bytecodeOpEnterScope
	bytecodeOpExitScope
	bytecodeOpConstI32
	bytecodeOpBinaryI32Add
	bytecodeOpBinaryI32Sub
	bytecodeOpUnboxI32
	bytecodeOpBoxI32
	bytecodeOpReturnBinary
	bytecodeOpReturnBinaryIntAddI32
	bytecodeOpReturnBinaryIntAdd
	bytecodeOpReturn
	bytecodeOpLoadSlot
	bytecodeOpLoadImplicitSlot
	bytecodeOpLoadSlotI32
	bytecodeOpLoadSlotStructField
	bytecodeOpStoreSlot
	bytecodeOpStoreSlotNew
	bytecodeOpStoreImplicitSlot
	bytecodeOpStoreSlotI32
	bytecodeOpStoreSlotCastSlotFloatConstDiv
	bytecodeOpStoreSlotFloatAffine
	bytecodeOpStoreSlotFloatRegion
	bytecodeOpCompoundAssignSlotI32
	bytecodeOpStoreSlotBinaryIntSlotConst
	bytecodeOpStoreSlotIntMulConstAdd
	bytecodeOpStoreSlotIntMulConstModConst
	bytecodeOpStoreSlotIntMulConstAddFromSlot
	bytecodeOpStoreSlotFloatBinary
	bytecodeOpStoreSlotFloatAddSub
	bytecodeOpStoreSlotFloatAddMul
	bytecodeOpStoreSlotFloatAddMulSlot
	bytecodeOpStoreSlotFloatAddMulArrayGet
	bytecodeOpCompoundAssignSlot
	bytecodeOpCompoundAssignImplicitSlot
	bytecodeOpCallSelf
	bytecodeOpCallSelfIntSubSlotConst
	bytecodeOpJumpIfBinaryCompareFalse
)

const bytecodeOpCount = int(bytecodeOpJumpIfBinaryCompareFalse) + 1

func (vm *bytecodeVM) run(program *bytecodeProgram) (runtime.Value, error) {
	if vm != nil && vm.interp != nil && vm.interp.bytecodeArrayOwnershipProfile != nil {
		vm.ensureBytecodeArrayOwnershipForProgram(program)
	}
	result, err := vm.runResumable(program, false)
	if err != nil {
		switch signal := err.(type) {
		case returnSignal:
			signal.value = vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonErrorControl, signal.value)
			vm.finishBytecodeArrayOwnershipPublicReturn(signal.value, false)
			return result, signal
		case breakSignal:
			signal.value = vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonErrorControl, signal.value)
			vm.finishBytecodeArrayOwnershipPublicReturn(result, true)
			return result, signal
		case raiseSignal:
			signal.value = vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonErrorControl, signal.value)
			vm.finishBytecodeArrayOwnershipPublicReturn(result, true)
			return result, signal
		}
		vm.finishBytecodeArrayOwnershipPublicReturn(result, true)
		return result, err
	}
	vm.finishBytecodeArrayOwnershipPublicReturn(result, false)
	// Raw scalar carriers are an implementation detail of one continuous VM
	// run. A completed run can return into the tree-walker, compiler bridge, or
	// another VM, each of which requires an ordinary runtime value.
	return vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonPublicEscape, result), nil
}

// runDetached executes a bytecode function reached through the generic call
// bridge. Its Array observer treats returned wrappers as pending caller values
// rather than public results, allowing the receiving VM to adopt provenance.
// With observation disabled, it has the same result/error behavior as run.
func (vm *bytecodeVM) runDetached(program *bytecodeProgram) (runtime.Value, error) {
	if vm != nil && vm.interp != nil && vm.interp.bytecodeArrayOwnershipProfile != nil {
		vm.ensureBytecodeArrayOwnershipForProgram(program)
	}
	result, err := vm.runResumable(program, false)
	if err != nil {
		switch signal := err.(type) {
		case returnSignal:
			signal.value = vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonErrorControl, signal.value)
			vm.finishBytecodeArrayOwnershipDetachedReturn(signal.value, false)
			return result, signal
		case breakSignal:
			signal.value = vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonErrorControl, signal.value)
			vm.finishBytecodeArrayOwnershipDetachedReturn(result, true)
			return result, signal
		case raiseSignal:
			signal.value = vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonErrorControl, signal.value)
			vm.finishBytecodeArrayOwnershipDetachedReturn(result, true)
			return result, signal
		}
		vm.finishBytecodeArrayOwnershipDetachedReturn(result, true)
		return result, err
	}
	vm.finishBytecodeArrayOwnershipDetachedReturn(result, false)
	return vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonPublicEscape, result), nil
}
