package interpreter

import (
	"able/interpreter-go/pkg/runtime"
)

type bytecodeOp int

const (
	bytecodeOpConst bytecodeOp = iota
	bytecodeOpLoadName
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
	bytecodeOpBinaryIntAddSlotConst
	bytecodeOpBinaryIntSubSlotConst
	bytecodeOpBinaryIntLessEqualSlotConst
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
	bytecodeOpMapLiteral
	bytecodeOpArrayLiteral
	bytecodeOpIndexGet
	bytecodeOpIndexSet
	bytecodeOpForLoop
	bytecodeOpCall
	bytecodeOpCallName
	bytecodeOpCallMember
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
	bytecodeOpReturnIfIntLessEqualSlotConst
	bytecodeOpReturnConstIfIntLessEqualSlotConst
	bytecodeOpJumpIfNil
	bytecodeOpLoopEnter
	bytecodeOpLoopExit
	bytecodeOpEnterScope
	bytecodeOpExitScope
	bytecodeOpConstI32
	bytecodeOpBinaryI32Add
	bytecodeOpBinaryI32Sub
	bytecodeOpBoxI32
	bytecodeOpReturnBinaryIntAddI32
	bytecodeOpReturnBinaryIntAdd
	bytecodeOpReturn
	bytecodeOpLoadSlot
	bytecodeOpLoadSlotI32
	bytecodeOpStoreSlot
	bytecodeOpStoreSlotNew
	bytecodeOpStoreSlotI32
	bytecodeOpCompoundAssignSlot
	bytecodeOpCallSelf
	bytecodeOpCallSelfIntSubSlotConst
)

const bytecodeOpCount = int(bytecodeOpCallSelfIntSubSlotConst) + 1

func (vm *bytecodeVM) run(program *bytecodeProgram) (runtime.Value, error) {
	return vm.runResumable(program, false)
}
