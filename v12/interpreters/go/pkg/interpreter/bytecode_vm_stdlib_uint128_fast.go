package interpreter

import (
	"math/big"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func bytecodeUInt128ValueFromInstance(inst *runtime.StructInstanceValue) (runtime.IntegerValue, bool) {
	high, ok := bytecodeStructU64Field(inst, "high")
	if !ok {
		return runtime.IntegerValue{}, false
	}
	low, ok := bytecodeStructU64Field(inst, "low")
	if !ok {
		return runtime.IntegerValue{}, false
	}
	value := new(big.Int).SetUint64(high)
	value.Lsh(value, 64)
	value.Or(value, new(big.Int).SetUint64(low))
	return bytecodeIntegerValueFromBig(runtime.IntegerU128, value), true
}

func bytecodeUInt128InstanceFromValue(def *runtime.StructDefinitionValue, value runtime.IntegerValue) (*runtime.StructInstanceValue, bool) {
	if value.Sign() < 0 || !integerValueWithinRange(value.BigInt(), runtime.IntegerU128) {
		return nil, false
	}
	lowPart := new(big.Int).And(runtime.CloneBigInt(value.BigInt()), bytecodeU64MaskBig)
	if !lowPart.IsUint64() {
		return nil, false
	}
	highPart := new(big.Int).Rsh(runtime.CloneBigInt(value.BigInt()), 64)
	if !highPart.IsUint64() {
		return nil, false
	}
	return bytecodeStructInstanceU64Pair(def, highPart.Uint64(), lowPart.Uint64()), true
}

func (vm *bytecodeVM) execUInt128LiteralStaticFast(instr bytecodeInstruction, def *runtime.StructDefinitionValue, receiverIndex int, high uint64, low uint64, dispatch string, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 0 {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "member_access", dispatch, bytecodeStructInstanceU64Pair(def, high, low), nil, callNode)
}

func (vm *bytecodeVM) execUInt128FromU64StaticFast(instr bytecodeInstruction, def *runtime.StructDefinitionValue, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 1 || argBase >= vm.stackDepth() {
		return nil, false, nil
	}
	value, ok := bytecodeIntegerValue(vm.stackValue(argBase))
	if !ok || value.TypeSuffix != runtime.IntegerU64 {
		return nil, false, nil
	}
	raw, ok := bytecodeStructU64Field(&runtime.StructInstanceValue{Fields: map[string]runtime.Value{"x": value}}, "x")
	if !ok {
		return nil, false, nil
	}
	inst, ok := bytecodeUInt128InstanceFromValue(def, bytecodeIntegerValueFromBig(runtime.IntegerU128, new(big.Int).SetUint64(raw)))
	if !ok {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "member_access", "uint128_from_u64_fast", inst, nil, callNode)
}

func (vm *bytecodeVM) execUInt128FromI64StaticFast(instr bytecodeInstruction, def *runtime.StructDefinitionValue, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 1 || argBase >= vm.stackDepth() {
		return nil, false, nil
	}
	value, ok := bytecodeIntegerValue(vm.stackValue(argBase))
	if !ok || value.TypeSuffix != runtime.IntegerI64 {
		return nil, false, nil
	}
	raw, fits := value.ToInt64()
	if !fits || raw < 0 {
		return nil, false, nil
	}
	inst, ok := bytecodeUInt128InstanceFromValue(def, runtime.NewSmallInt(raw, runtime.IntegerU128))
	if !ok {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "member_access", "uint128_from_i64_fast", inst, nil, callNode)
}

func (vm *bytecodeVM) execUInt128FromU128StaticFast(instr bytecodeInstruction, def *runtime.StructDefinitionValue, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 1 || argBase >= vm.stackDepth() {
		return nil, false, nil
	}
	value, ok := bytecodeIntegerValue(vm.stackValue(argBase))
	if !ok || value.TypeSuffix != runtime.IntegerU128 {
		return nil, false, nil
	}
	inst, ok := bytecodeUInt128InstanceFromValue(def, value)
	if !ok {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "member_access", "uint128_from_u128_fast", inst, nil, callNode)
}

func (vm *bytecodeVM) execUInt128ToU128MemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 0 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() {
		return nil, false, nil
	}
	inst, ok := vm.bytecodeCanonicalStructInstance(vm.stackValue(receiverIndex), "UInt128", "numbers/uint128.able")
	if !ok {
		return nil, false, nil
	}
	value, ok := bytecodeUInt128ValueFromInstance(inst)
	if !ok {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "resolved_method", "uint128_to_u128_fast", value, nil, callNode)
}

func (vm *bytecodeVM) execUInt128ToU64MemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	_, value, ok := vm.bytecodeCanonicalUInt128ValueAt(receiverIndex, instr.argCount)
	if !ok {
		return nil, false, nil
	}
	result, ok := coerceIntegerValueToTargetKindIfInRange(value, runtime.IntegerU64)
	if !ok {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "resolved_method", "uint128_to_u64_fast", result, nil, callNode)
}

func (vm *bytecodeVM) execUInt128ToI64MemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	_, value, ok := vm.bytecodeCanonicalUInt128ValueAt(receiverIndex, instr.argCount)
	if !ok {
		return nil, false, nil
	}
	result, ok := coerceIntegerValueToTargetKindIfInRange(value, runtime.IntegerI64)
	if !ok {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "resolved_method", "uint128_to_i64_fast", result, nil, callNode)
}

func (vm *bytecodeVM) bytecodeCanonicalUInt128ValueAt(receiverIndex int, argCount int) (*runtime.StructInstanceValue, runtime.IntegerValue, bool) {
	if receiverIndex < 0 || receiverIndex >= vm.stackDepth() || argCount != 0 && receiverIndex+argCount >= vm.stackDepth() {
		return nil, runtime.IntegerValue{}, false
	}
	inst, ok := vm.bytecodeCanonicalStructInstance(vm.stackValue(receiverIndex), "UInt128", "numbers/uint128.able")
	if !ok {
		return nil, runtime.IntegerValue{}, false
	}
	value, ok := bytecodeUInt128ValueFromInstance(inst)
	if !ok {
		return nil, runtime.IntegerValue{}, false
	}
	return inst, value, true
}

func (vm *bytecodeVM) execUInt128IsZeroMemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	inst, ok := vm.bytecodeCanonicalStructInstance(vm.stackValue(receiverIndex), "UInt128", "numbers/uint128.able")
	if !ok {
		return nil, false, nil
	}
	high, ok := bytecodeStructU64Field(inst, "high")
	if !ok {
		return nil, false, nil
	}
	low, ok := bytecodeStructU64Field(inst, "low")
	if !ok {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "resolved_method", "uint128_is_zero_fast", runtime.BoolValue{Val: high == 0 && low == 0}, nil, callNode)
}

func (vm *bytecodeVM) execUInt128IsPositiveMemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	inst, ok := vm.bytecodeCanonicalStructInstance(vm.stackValue(receiverIndex), "UInt128", "numbers/uint128.able")
	if !ok {
		return nil, false, nil
	}
	high, ok := bytecodeStructU64Field(inst, "high")
	if !ok {
		return nil, false, nil
	}
	low, ok := bytecodeStructU64Field(inst, "low")
	if !ok {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "resolved_method", "uint128_is_positive_fast", runtime.BoolValue{Val: high != 0 || low != 0}, nil, callNode)
}

func (vm *bytecodeVM) execUInt128AddMemberFast(instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	inst, _, err, handled := vm.execUInt128BinaryCheckedFast(receiverIndex, argBase, "+", "UInt128.add overflow")
	if !handled {
		return nil, false, err
	}
	if err != nil {
		vm.truncateStack(receiverIndex)
		return vm.finishCanonicalStructFastCall(instr, "resolved_method", "uint128_add_fast", nil, err, callNode)
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "resolved_method", "uint128_add_fast", inst, nil, callNode)
}

func (vm *bytecodeVM) execUInt128BinaryCheckedFast(receiverIndex int, argBase int, op string, overflowMessage string) (*runtime.StructInstanceValue, runtime.IntegerValue, error, bool) {
	if receiverIndex < 0 || receiverIndex >= vm.stackDepth() || argBase < 0 || argBase >= vm.stackDepth() {
		return nil, runtime.IntegerValue{}, nil, false
	}
	leftInst, ok := vm.bytecodeCanonicalStructInstance(vm.stackValue(receiverIndex), "UInt128", "numbers/uint128.able")
	if !ok {
		return nil, runtime.IntegerValue{}, nil, false
	}
	rightInst, ok := vm.bytecodeCanonicalStructInstance(vm.stackValue(argBase), "UInt128", "numbers/uint128.able")
	if !ok {
		return nil, runtime.IntegerValue{}, nil, false
	}
	left, ok := bytecodeUInt128ValueFromInstance(leftInst)
	if !ok {
		return nil, runtime.IntegerValue{}, nil, false
	}
	right, ok := bytecodeUInt128ValueFromInstance(rightInst)
	if !ok {
		return nil, runtime.IntegerValue{}, nil, false
	}
	result, err, handled := bytecodeApplyIntegerBinaryFast(op, left, right, runtime.IntegerU128)
	if !handled {
		return nil, runtime.IntegerValue{}, err, false
	}
	if err != nil {
		return nil, runtime.IntegerValue{}, err, true
	}
	switch op {
	case "+":
		if result.CmpInt(left) < 0 || result.CmpInt(right) < 0 {
			return nil, runtime.IntegerValue{}, newOverflowError(overflowMessage), true
		}
	case "-":
		if left.CmpInt(right) < 0 {
			return nil, runtime.IntegerValue{}, newOverflowError(overflowMessage), true
		}
	}
	inst, ok := bytecodeUInt128InstanceFromValue(leftInst.Definition, result)
	if !ok {
		return nil, runtime.IntegerValue{}, nil, false
	}
	return inst, result, nil, true
}

func (vm *bytecodeVM) execUInt128SubMemberFast(instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	inst, _, err, handled := vm.execUInt128BinaryCheckedFast(receiverIndex, argBase, "-", "UInt128.sub underflow")
	if !handled {
		return nil, false, err
	}
	if err != nil {
		vm.truncateStack(receiverIndex)
		return vm.finishCanonicalStructFastCall(instr, "resolved_method", "uint128_sub_fast", nil, err, callNode)
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "resolved_method", "uint128_sub_fast", inst, nil, callNode)
}

func (vm *bytecodeVM) execUInt128MulMemberFast(instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 1 || receiverIndex < 0 || argBase >= vm.stackDepth() {
		return nil, false, nil
	}
	leftInst, ok := vm.bytecodeCanonicalStructInstance(vm.stackValue(receiverIndex), "UInt128", "numbers/uint128.able")
	if !ok {
		return nil, false, nil
	}
	rightInst, ok := vm.bytecodeCanonicalStructInstance(vm.stackValue(argBase), "UInt128", "numbers/uint128.able")
	if !ok {
		return nil, false, nil
	}
	left, ok := bytecodeUInt128ValueFromInstance(leftInst)
	if !ok {
		return nil, false, nil
	}
	right, ok := bytecodeUInt128ValueFromInstance(rightInst)
	if !ok {
		return nil, false, nil
	}
	product, err, handled := bytecodeApplyIntegerBinaryFast("*", left, right, runtime.IntegerU128)
	if !handled {
		return nil, false, err
	}
	if err != nil {
		vm.truncateStack(receiverIndex)
		return vm.finishCanonicalStructFastCall(instr, "resolved_method", "uint128_mul_fast", nil, err, callNode)
	}
	if right.Sign() != 0 {
		quotient, divErr, divHandled := bytecodeApplyIntegerBinaryFast("//", product, right, runtime.IntegerU128)
		if !divHandled {
			return nil, false, divErr
		}
		if divErr != nil {
			vm.truncateStack(receiverIndex)
			return vm.finishCanonicalStructFastCall(instr, "resolved_method", "uint128_mul_fast", nil, divErr, callNode)
		}
		if quotient.CmpInt(left) != 0 {
			vm.truncateStack(receiverIndex)
			return vm.finishCanonicalStructFastCall(instr, "resolved_method", "uint128_mul_fast", nil, newOverflowError("UInt128.mul overflow"), callNode)
		}
	}
	inst, ok := bytecodeUInt128InstanceFromValue(leftInst.Definition, product)
	if !ok {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "resolved_method", "uint128_mul_fast", inst, nil, callNode)
}

func (vm *bytecodeVM) execUInt128BinaryMemberFast(instr bytecodeInstruction, receiverIndex int, argBase int, op string, dispatch string, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 1 || receiverIndex < 0 || argBase >= vm.stackDepth() {
		return nil, false, nil
	}
	leftInst, ok := vm.bytecodeCanonicalStructInstance(vm.stackValue(receiverIndex), "UInt128", "numbers/uint128.able")
	if !ok {
		return nil, false, nil
	}
	rightInst, ok := vm.bytecodeCanonicalStructInstance(vm.stackValue(argBase), "UInt128", "numbers/uint128.able")
	if !ok {
		return nil, false, nil
	}
	left, ok := bytecodeUInt128ValueFromInstance(leftInst)
	if !ok {
		return nil, false, nil
	}
	right, ok := bytecodeUInt128ValueFromInstance(rightInst)
	if !ok {
		return nil, false, nil
	}
	result, err, handled := bytecodeApplyIntegerBinaryFast(op, left, right, runtime.IntegerU128)
	if !handled {
		return nil, false, err
	}
	if err != nil {
		vm.truncateStack(receiverIndex)
		return vm.finishCanonicalStructFastCall(instr, "resolved_method", dispatch, nil, err, callNode)
	}
	inst, ok := bytecodeUInt128InstanceFromValue(leftInst.Definition, result)
	if !ok {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "resolved_method", dispatch, inst, nil, callNode)
}
