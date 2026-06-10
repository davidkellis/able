package interpreter

import (
	"math/big"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

const (
	bytecodeMemberMethodFastPathRandomNextI64 bytecodeMemberMethodFastPathKind = bytecodeMemberMethodFastPathStringBuilderFinish + 1 + iota
	bytecodeMemberMethodFastPathRandomNextI32
	bytecodeMemberMethodFastPathRandomNextF64
	bytecodeMemberMethodFastPathInt128ToI128
	bytecodeMemberMethodFastPathInt128ToI64
	bytecodeMemberMethodFastPathInt128ToU64
	bytecodeMemberMethodFastPathInt128IsZero
	bytecodeMemberMethodFastPathInt128IsNegative
	bytecodeMemberMethodFastPathInt128Add
	bytecodeMemberMethodFastPathInt128Sub
	bytecodeMemberMethodFastPathInt128Mul
	bytecodeMemberMethodFastPathInt128Div
	bytecodeMemberMethodFastPathInt128Rem
	bytecodeMemberMethodFastPathUInt128ToU128
	bytecodeMemberMethodFastPathUInt128ToU64
	bytecodeMemberMethodFastPathUInt128ToI64
	bytecodeMemberMethodFastPathUInt128IsZero
	bytecodeMemberMethodFastPathUInt128IsPositive
	bytecodeMemberMethodFastPathUInt128Add
	bytecodeMemberMethodFastPathUInt128Sub
	bytecodeMemberMethodFastPathUInt128Mul
	bytecodeMemberMethodFastPathUInt128Div
	bytecodeMemberMethodFastPathUInt128Rem
)

var (
	bytecodeU64MaskBig   = new(big.Int).SetUint64(^uint64(0))
	bytecodeTwoPow128Big = new(big.Int).Lsh(big.NewInt(1), 128)
)

func isCanonicalResultReturnType(expr ast.TypeExpression, inner string) bool {
	generic, ok := expr.(*ast.GenericTypeExpression)
	if !ok || generic == nil || len(generic.Arguments) != 1 {
		return false
	}
	return typeExpressionToString(generic.Base) == "Result" &&
		typeExpressionToString(generic.Arguments[0]) == inner
}

func isCanonicalSelfOnlyMethod(def *ast.FunctionDefinition, returnType string) bool {
	return def != nil &&
		len(def.Params) == 1 &&
		typeExpressionToString(def.ReturnType) == returnType
}

func isCanonicalUnaryStructMethod(def *ast.FunctionDefinition, argType string, returnType string) bool {
	return def != nil &&
		len(def.Params) == 2 &&
		typeExpressionToString(def.Params[1].ParamType) == argType &&
		typeExpressionToString(def.ReturnType) == returnType
}

func (vm *bytecodeVM) memberMethodStructCanonicalStdlibFastPath(key bytecodeMemberMethodCacheKey, def *ast.FunctionDefinition, origin string) bytecodeMemberMethodFastPathKind {
	if vm == nil || key.structDef == nil || def == nil {
		return bytecodeMemberMethodFastPathNone
	}
	switch {
	case vm.isCanonicalStructDefinition(key.structDef, "Random", "random.able") &&
		isCanonicalAbleStdlibOrigin(origin, "random.able"):
		switch key.member {
		case "next_i64":
			if isCanonicalSelfOnlyMethod(def, "i64") {
				return bytecodeMemberMethodFastPathRandomNextI64
			}
		case "next_i32":
			if isCanonicalSelfOnlyMethod(def, "i32") {
				return bytecodeMemberMethodFastPathRandomNextI32
			}
		case "next_f64":
			if isCanonicalSelfOnlyMethod(def, "f64") {
				return bytecodeMemberMethodFastPathRandomNextF64
			}
		}
	case vm.isCanonicalStructDefinition(key.structDef, "Int128", "numbers/int128.able") &&
		isCanonicalAbleStdlibOrigin(origin, "numbers/int128.able"):
		switch key.member {
		case "to_i128":
			if isCanonicalSelfOnlyMethod(def, "i128") {
				return bytecodeMemberMethodFastPathInt128ToI128
			}
		case "to_i64":
			if len(def.Params) == 1 && isCanonicalResultReturnType(def.ReturnType, "i64") {
				return bytecodeMemberMethodFastPathInt128ToI64
			}
		case "to_u64":
			if len(def.Params) == 1 && isCanonicalResultReturnType(def.ReturnType, "u64") {
				return bytecodeMemberMethodFastPathInt128ToU64
			}
		case "is_zero":
			if isCanonicalSelfOnlyMethod(def, "bool") {
				return bytecodeMemberMethodFastPathInt128IsZero
			}
		case "is_negative":
			if isCanonicalSelfOnlyMethod(def, "bool") {
				return bytecodeMemberMethodFastPathInt128IsNegative
			}
		case "add":
			if isCanonicalUnaryStructMethod(def, "Int128", "Int128") {
				return bytecodeMemberMethodFastPathInt128Add
			}
		case "sub":
			if isCanonicalUnaryStructMethod(def, "Int128", "Int128") {
				return bytecodeMemberMethodFastPathInt128Sub
			}
		case "mul":
			if isCanonicalUnaryStructMethod(def, "Int128", "Int128") {
				return bytecodeMemberMethodFastPathInt128Mul
			}
		case "div":
			if isCanonicalUnaryStructMethod(def, "Int128", "Int128") {
				return bytecodeMemberMethodFastPathInt128Div
			}
		case "rem":
			if isCanonicalUnaryStructMethod(def, "Int128", "Int128") {
				return bytecodeMemberMethodFastPathInt128Rem
			}
		}
	case vm.isCanonicalStructDefinition(key.structDef, "UInt128", "numbers/uint128.able") &&
		isCanonicalAbleStdlibOrigin(origin, "numbers/uint128.able"):
		switch key.member {
		case "to_u128":
			if isCanonicalSelfOnlyMethod(def, "u128") {
				return bytecodeMemberMethodFastPathUInt128ToU128
			}
		case "to_u64":
			if len(def.Params) == 1 && isCanonicalResultReturnType(def.ReturnType, "u64") {
				return bytecodeMemberMethodFastPathUInt128ToU64
			}
		case "to_i64":
			if len(def.Params) == 1 && isCanonicalResultReturnType(def.ReturnType, "i64") {
				return bytecodeMemberMethodFastPathUInt128ToI64
			}
		case "is_zero":
			if isCanonicalSelfOnlyMethod(def, "bool") {
				return bytecodeMemberMethodFastPathUInt128IsZero
			}
		case "is_positive":
			if isCanonicalSelfOnlyMethod(def, "bool") {
				return bytecodeMemberMethodFastPathUInt128IsPositive
			}
		case "add":
			if isCanonicalUnaryStructMethod(def, "UInt128", "UInt128") {
				return bytecodeMemberMethodFastPathUInt128Add
			}
		case "sub":
			if isCanonicalUnaryStructMethod(def, "UInt128", "UInt128") {
				return bytecodeMemberMethodFastPathUInt128Sub
			}
		case "mul":
			if isCanonicalUnaryStructMethod(def, "UInt128", "UInt128") {
				return bytecodeMemberMethodFastPathUInt128Mul
			}
		case "div":
			if isCanonicalUnaryStructMethod(def, "UInt128", "UInt128") {
				return bytecodeMemberMethodFastPathUInt128Div
			}
		case "rem":
			if isCanonicalUnaryStructMethod(def, "UInt128", "UInt128") {
				return bytecodeMemberMethodFastPathUInt128Rem
			}
		}
	}
	return bytecodeMemberMethodFastPathNone
}

func (vm *bytecodeVM) execStaticCanonicalStructMemberFast(instr bytecodeInstruction, receiver runtime.Value, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	def, ok := receiver.(*runtime.StructDefinitionValue)
	if !ok || def == nil || vm == nil || vm.interp == nil {
		return nil, false, nil
	}
	switch {
	case vm.isCanonicalStructDefinition(def, "Random", "random.able"):
		switch instr.name {
		case "seeded":
			return vm.execRandomSeededStaticFast(instr, def, receiverIndex, argBase, callNode)
		case "default":
			return vm.execRandomDefaultStaticFast(instr, def, receiverIndex, callNode)
		}
	case vm.isCanonicalStructDefinition(def, "Int128", "numbers/int128.able"):
		switch instr.name {
		case "zero":
			return vm.execInt128LiteralStaticFast(instr, def, receiverIndex, 0, 0, "int128_zero_fast", callNode)
		case "one":
			return vm.execInt128LiteralStaticFast(instr, def, receiverIndex, 0, 1, "int128_one_fast", callNode)
		case "new":
			return vm.execInt128NewStaticFast(instr, def, receiverIndex, argBase, callNode)
		case "from_i64":
			return vm.execInt128FromI64StaticFast(instr, def, receiverIndex, argBase, callNode)
		case "from_u64":
			return vm.execInt128FromU64StaticFast(instr, def, receiverIndex, argBase, callNode)
		case "from_i128":
			return vm.execInt128FromI128StaticFast(instr, def, receiverIndex, argBase, callNode)
		}
	case vm.isCanonicalStructDefinition(def, "UInt128", "numbers/uint128.able"):
		switch instr.name {
		case "zero":
			return vm.execUInt128LiteralStaticFast(instr, def, receiverIndex, 0, 0, "uint128_zero_fast", callNode)
		case "one":
			return vm.execUInt128LiteralStaticFast(instr, def, receiverIndex, 0, 1, "uint128_one_fast", callNode)
		case "new":
			return vm.execUInt128NewStaticFast(instr, def, receiverIndex, argBase, callNode)
		case "from_u64":
			return vm.execUInt128FromU64StaticFast(instr, def, receiverIndex, argBase, callNode)
		case "from_i64":
			return vm.execUInt128FromI64StaticFast(instr, def, receiverIndex, argBase, callNode)
		case "from_u128":
			return vm.execUInt128FromU128StaticFast(instr, def, receiverIndex, argBase, callNode)
		}
	}
	return nil, false, nil
}

func (vm *bytecodeVM) finishCanonicalStructFastCall(instr bytecodeInstruction, lookup string, dispatch string, result runtime.Value, err error, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if vm == nil {
		return nil, false, nil
	}
	if vm.interp != nil {
		err = vm.interp.wrapStandardRuntimeError(err)
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, lookup, dispatch, instr.node)
	}
	newProg, finishErr := vm.finishCompletedCall(result, err, callNode, nil)
	return newProg, true, finishErr
}

func (vm *bytecodeVM) bytecodeCanonicalStructInstance(value runtime.Value, name string, originRelative string) (*runtime.StructInstanceValue, bool) {
	for {
		switch typed := value.(type) {
		case runtime.InterfaceValue:
			value = typed.Underlying
			continue
		case *runtime.InterfaceValue:
			if typed == nil {
				return nil, false
			}
			value = typed.Underlying
			continue
		case *runtime.StructInstanceValue:
			if typed == nil || typed.Definition == nil || !vm.isCanonicalStructDefinition(typed.Definition, name, originRelative) {
				return nil, false
			}
			return typed, true
		default:
			return nil, false
		}
	}
}

func bytecodeStructIntegerField(inst *runtime.StructInstanceValue, name string) (runtime.IntegerValue, bool) {
	raw, ok := structNamedFieldValue(inst, name)
	if !ok {
		return runtime.IntegerValue{}, false
	}
	return bytecodeIntegerValue(raw)
}

func bytecodeStructU64Field(inst *runtime.StructInstanceValue, name string) (uint64, bool) {
	intVal, ok := bytecodeStructIntegerField(inst, name)
	if !ok {
		return 0, false
	}
	if raw, fits := intVal.ToInt64(); fits {
		if raw < 0 {
			return 0, false
		}
		return uint64(raw), true
	}
	if intVal.Sign() < 0 || !intVal.BigInt().IsUint64() {
		return 0, false
	}
	return intVal.BigInt().Uint64(), true
}

func bytecodeStructI64Field(inst *runtime.StructInstanceValue, name string) (int64, bool) {
	intVal, ok := bytecodeStructIntegerField(inst, name)
	if !ok {
		return 0, false
	}
	raw, fits := intVal.ToInt64()
	if !fits {
		return 0, false
	}
	return raw, true
}

func bytecodeUnsignedIntegerValue(kind runtime.IntegerType, value uint64) runtime.Value {
	if value <= uint64(^uint64(0)>>1) {
		return boxedOrSmallIntegerValue(kind, int64(value))
	}
	return runtime.NewBigIntValue(new(big.Int).SetUint64(value), kind)
}

func bytecodeIntegerValueFromBig(kind runtime.IntegerType, value *big.Int) runtime.IntegerValue {
	if value.IsInt64() {
		return runtime.NewSmallInt(value.Int64(), kind)
	}
	return runtime.NewBigIntValue(runtime.CloneBigInt(value), kind)
}

func bytecodeStructInstanceU64Pair(def *runtime.StructDefinitionValue, high uint64, low uint64) *runtime.StructInstanceValue {
	inst, fields := runtime.NewStructInstancePositionalSized(def, 2, nil)
	fields[0] = bytecodeUnsignedIntegerValue(runtime.IntegerU64, high)
	fields[1] = bytecodeUnsignedIntegerValue(runtime.IntegerU64, low)
	return inst
}

func bytecodeStructInstanceI64(def *runtime.StructDefinitionValue, fieldName string, value int64) *runtime.StructInstanceValue {
	inst, fields := runtime.NewStructInstancePositionalSized(def, 1, nil)
	if fieldName == "state" && len(fields) > 0 {
		fields[0] = boxedOrSmallIntegerValue(runtime.IntegerI64, value)
	}
	return inst
}

func bytecodeInt128ValueFromInstance(inst *runtime.StructInstanceValue) (runtime.IntegerValue, bool) {
	high, ok := bytecodeStructU64Field(inst, "high")
	if !ok {
		return runtime.IntegerValue{}, false
	}
	low, ok := bytecodeStructU64Field(inst, "low")
	if !ok {
		return runtime.IntegerValue{}, false
	}
	pattern := new(big.Int).SetUint64(high)
	pattern.Lsh(pattern, 64)
	pattern.Or(pattern, new(big.Int).SetUint64(low))
	if high&(uint64(1)<<63) != 0 {
		pattern.Sub(pattern, bytecodeTwoPow128Big)
	}
	return bytecodeIntegerValueFromBig(runtime.IntegerI128, pattern), true
}

func bytecodeInt128InstanceFromValue(def *runtime.StructDefinitionValue, value runtime.IntegerValue) (*runtime.StructInstanceValue, bool) {
	info, ok := lookupIntegerInfo(runtime.IntegerI128)
	if !ok {
		return nil, false
	}
	pattern := bitPattern(value.BigInt(), info)
	lowPart := new(big.Int).And(runtime.CloneBigInt(pattern), bytecodeU64MaskBig)
	if !lowPart.IsUint64() {
		return nil, false
	}
	highPart := new(big.Int).Rsh(runtime.CloneBigInt(pattern), 64)
	if !highPart.IsUint64() {
		return nil, false
	}
	return bytecodeStructInstanceU64Pair(def, highPart.Uint64(), lowPart.Uint64()), true
}

func bytecodeIntegerResultOfKind(value runtime.Value, kind runtime.IntegerType) (runtime.IntegerValue, bool) {
	intVal, ok := bytecodeIntegerValue(value)
	if !ok || intVal.TypeSuffix != kind {
		return runtime.IntegerValue{}, false
	}
	return intVal, true
}

func bytecodeApplyIntegerBinaryFast(op string, left runtime.IntegerValue, right runtime.IntegerValue, kind runtime.IntegerType) (runtime.IntegerValue, error, bool) {
	result, handled, err := ApplyBinaryOperatorFast(op, left, right)
	if !handled || err != nil {
		return runtime.IntegerValue{}, err, handled
	}
	intResult, ok := bytecodeIntegerResultOfKind(result, kind)
	if !ok {
		return runtime.IntegerValue{}, nil, false
	}
	return intResult, nil, true
}

func (vm *bytecodeVM) execRandomSeededStaticFast(instr bytecodeInstruction, def *runtime.StructDefinitionValue, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 1 || argBase >= vm.stackDepth() {
		return nil, false, nil
	}
	seed, ok := bytecodeIntegerValue(vm.stackValue(argBase))
	if !ok || seed.TypeSuffix != runtime.IntegerI64 {
		return nil, false, nil
	}
	rawSeed, fits := seed.ToInt64()
	if !fits {
		return nil, false, nil
	}
	state := rawSeed % 2147483647
	if state <= 0 {
		state = 1
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "member_access", "random_seeded_fast", bytecodeStructInstanceI64(def, "state", state), nil, callNode)
}

func (vm *bytecodeVM) execRandomDefaultStaticFast(instr bytecodeInstruction, def *runtime.StructDefinitionValue, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 0 {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "member_access", "random_default_fast", bytecodeStructInstanceI64(def, "state", 1), nil, callNode)
}

func (vm *bytecodeVM) execCanonicalRandomNextState(receiver runtime.Value) (*runtime.StructInstanceValue, int64, bool) {
	inst, ok := vm.bytecodeCanonicalStructInstance(receiver, "Random", "random.able")
	if !ok {
		return nil, 0, false
	}
	state, ok := bytecodeStructI64Field(inst, "state")
	if !ok {
		return nil, 0, false
	}
	next := (state * 48271) % 2147483647
	if !structSetNamedFieldValue(inst, "state", boxedOrSmallIntegerValue(runtime.IntegerI64, next)) {
		return nil, 0, false
	}
	return inst, next, true
}

func (vm *bytecodeVM) execRandomNextI64MemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 0 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() {
		return nil, false, nil
	}
	_, next, ok := vm.execCanonicalRandomNextState(vm.stackValue(receiverIndex))
	if !ok {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "resolved_method", "random_next_i64_fast", boxedOrSmallIntegerValue(runtime.IntegerI64, next), nil, callNode)
}

func (vm *bytecodeVM) execRandomNextI32MemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 0 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() {
		return nil, false, nil
	}
	_, next, ok := vm.execCanonicalRandomNextState(vm.stackValue(receiverIndex))
	if !ok {
		return nil, false, nil
	}
	value, handled, err := castValueToCanonicalSimpleTypeFast("i32", runtime.NewSmallInt(next, runtime.IntegerI64))
	if err != nil || !handled {
		return nil, false, err
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "resolved_method", "random_next_i32_fast", value, nil, callNode)
}

func (vm *bytecodeVM) execRandomNextF64MemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 0 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() {
		return nil, false, nil
	}
	_, next, ok := vm.execCanonicalRandomNextState(vm.stackValue(receiverIndex))
	if !ok {
		return nil, false, nil
	}
	result := runtime.FloatValue{Val: float64(next) / 2147483647.0, TypeSuffix: runtime.FloatF64}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "resolved_method", "random_next_f64_fast", result, nil, callNode)
}

func (vm *bytecodeVM) execInt128LiteralStaticFast(instr bytecodeInstruction, def *runtime.StructDefinitionValue, receiverIndex int, high uint64, low uint64, dispatch string, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 0 {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "member_access", dispatch, bytecodeStructInstanceU64Pair(def, high, low), nil, callNode)
}

func (vm *bytecodeVM) execInt128NewStaticFast(instr bytecodeInstruction, def *runtime.StructDefinitionValue, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 2 || argBase+1 >= vm.stackDepth() {
		return nil, false, nil
	}
	high, ok := bytecodeIntegerValue(vm.stackValue(argBase))
	if !ok || high.TypeSuffix != runtime.IntegerU64 {
		return nil, false, nil
	}
	low, ok := bytecodeIntegerValue(vm.stackValue(argBase + 1))
	if !ok || low.TypeSuffix != runtime.IntegerU64 {
		return nil, false, nil
	}
	highRaw, ok := bytecodeStructU64Field(&runtime.StructInstanceValue{Fields: map[string]runtime.Value{"x": high}}, "x")
	if !ok {
		return nil, false, nil
	}
	lowRaw, ok := bytecodeStructU64Field(&runtime.StructInstanceValue{Fields: map[string]runtime.Value{"x": low}}, "x")
	if !ok {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "member_access", "int128_new_fast", bytecodeStructInstanceU64Pair(def, highRaw, lowRaw), nil, callNode)
}

func (vm *bytecodeVM) execUInt128NewStaticFast(instr bytecodeInstruction, def *runtime.StructDefinitionValue, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 2 || argBase+1 >= vm.stackDepth() {
		return nil, false, nil
	}
	high, ok := bytecodeIntegerValue(vm.stackValue(argBase))
	if !ok || high.TypeSuffix != runtime.IntegerU64 {
		return nil, false, nil
	}
	low, ok := bytecodeIntegerValue(vm.stackValue(argBase + 1))
	if !ok || low.TypeSuffix != runtime.IntegerU64 {
		return nil, false, nil
	}
	highRaw, ok := bytecodeStructU64Field(&runtime.StructInstanceValue{Fields: map[string]runtime.Value{"x": high}}, "x")
	if !ok {
		return nil, false, nil
	}
	lowRaw, ok := bytecodeStructU64Field(&runtime.StructInstanceValue{Fields: map[string]runtime.Value{"x": low}}, "x")
	if !ok {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "member_access", "uint128_new_fast", bytecodeStructInstanceU64Pair(def, highRaw, lowRaw), nil, callNode)
}

func (vm *bytecodeVM) execInt128FromI64StaticFast(instr bytecodeInstruction, def *runtime.StructDefinitionValue, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 1 || argBase >= vm.stackDepth() {
		return nil, false, nil
	}
	value, ok := bytecodeIntegerValue(vm.stackValue(argBase))
	if !ok || value.TypeSuffix != runtime.IntegerI64 {
		return nil, false, nil
	}
	inst, ok := bytecodeInt128InstanceFromValue(def, runtime.NewSmallInt(value.Int64FastRef(), runtime.IntegerI128))
	if !ok {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "member_access", "int128_from_i64_fast", inst, nil, callNode)
}

func (vm *bytecodeVM) execInt128FromU64StaticFast(instr bytecodeInstruction, def *runtime.StructDefinitionValue, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
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
	inst, ok := bytecodeInt128InstanceFromValue(def, bytecodeIntegerValueFromBig(runtime.IntegerI128, new(big.Int).SetUint64(raw)))
	if !ok {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "member_access", "int128_from_u64_fast", inst, nil, callNode)
}

func (vm *bytecodeVM) execInt128FromI128StaticFast(instr bytecodeInstruction, def *runtime.StructDefinitionValue, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 1 || argBase >= vm.stackDepth() {
		return nil, false, nil
	}
	value, ok := bytecodeIntegerValue(vm.stackValue(argBase))
	if !ok || value.TypeSuffix != runtime.IntegerI128 {
		return nil, false, nil
	}
	inst, ok := bytecodeInt128InstanceFromValue(def, value)
	if !ok {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "member_access", "int128_from_i128_fast", inst, nil, callNode)
}

func (vm *bytecodeVM) execInt128ToI128MemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 0 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() {
		return nil, false, nil
	}
	inst, ok := vm.bytecodeCanonicalStructInstance(vm.stackValue(receiverIndex), "Int128", "numbers/int128.able")
	if !ok {
		return nil, false, nil
	}
	value, ok := bytecodeInt128ValueFromInstance(inst)
	if !ok {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "resolved_method", "int128_to_i128_fast", value, nil, callNode)
}

func (vm *bytecodeVM) execInt128ToI64MemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	inst, value, ok := vm.bytecodeCanonicalInt128ValueAt(receiverIndex, instr.argCount)
	if !ok {
		return nil, false, nil
	}
	result, ok := coerceIntegerValueToTargetKindIfInRange(value, runtime.IntegerI64)
	if !ok {
		return nil, false, nil
	}
	_ = inst
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "resolved_method", "int128_to_i64_fast", result, nil, callNode)
}

func (vm *bytecodeVM) execInt128ToU64MemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	_, value, ok := vm.bytecodeCanonicalInt128ValueAt(receiverIndex, instr.argCount)
	if !ok {
		return nil, false, nil
	}
	result, ok := coerceIntegerValueToTargetKindIfInRange(value, runtime.IntegerU64)
	if !ok {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "resolved_method", "int128_to_u64_fast", result, nil, callNode)
}

func (vm *bytecodeVM) bytecodeCanonicalInt128ValueAt(receiverIndex int, argCount int) (*runtime.StructInstanceValue, runtime.IntegerValue, bool) {
	if receiverIndex < 0 || receiverIndex >= vm.stackDepth() || argCount != 0 && receiverIndex+argCount >= vm.stackDepth() {
		return nil, runtime.IntegerValue{}, false
	}
	inst, ok := vm.bytecodeCanonicalStructInstance(vm.stackValue(receiverIndex), "Int128", "numbers/int128.able")
	if !ok {
		return nil, runtime.IntegerValue{}, false
	}
	value, ok := bytecodeInt128ValueFromInstance(inst)
	if !ok {
		return nil, runtime.IntegerValue{}, false
	}
	return inst, value, true
}

func (vm *bytecodeVM) execInt128IsZeroMemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	inst, ok := vm.bytecodeCanonicalStructInstance(vm.stackValue(receiverIndex), "Int128", "numbers/int128.able")
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
	return vm.finishCanonicalStructFastCall(instr, "resolved_method", "int128_is_zero_fast", runtime.BoolValue{Val: high == 0 && low == 0}, nil, callNode)
}

func (vm *bytecodeVM) execInt128IsNegativeMemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	inst, ok := vm.bytecodeCanonicalStructInstance(vm.stackValue(receiverIndex), "Int128", "numbers/int128.able")
	if !ok {
		return nil, false, nil
	}
	high, ok := bytecodeStructU64Field(inst, "high")
	if !ok {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "resolved_method", "int128_is_negative_fast", runtime.BoolValue{Val: (high & (uint64(1) << 63)) != 0}, nil, callNode)
}

func (vm *bytecodeVM) execInt128BinaryMemberFast(instr bytecodeInstruction, receiverIndex int, argBase int, op string, dispatch string, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 1 || receiverIndex < 0 || argBase >= vm.stackDepth() {
		return nil, false, nil
	}
	leftInst, ok := vm.bytecodeCanonicalStructInstance(vm.stackValue(receiverIndex), "Int128", "numbers/int128.able")
	if !ok {
		return nil, false, nil
	}
	rightInst, ok := vm.bytecodeCanonicalStructInstance(vm.stackValue(argBase), "Int128", "numbers/int128.able")
	if !ok {
		return nil, false, nil
	}
	left, ok := bytecodeInt128ValueFromInstance(leftInst)
	if !ok {
		return nil, false, nil
	}
	right, ok := bytecodeInt128ValueFromInstance(rightInst)
	if !ok {
		return nil, false, nil
	}
	result, err, handled := bytecodeApplyIntegerBinaryFast(op, left, right, runtime.IntegerI128)
	if !handled {
		return nil, false, err
	}
	if err != nil {
		vm.truncateStack(receiverIndex)
		return vm.finishCanonicalStructFastCall(instr, "resolved_method", dispatch, nil, err, callNode)
	}
	inst, ok := bytecodeInt128InstanceFromValue(leftInst.Definition, result)
	if !ok {
		return nil, false, nil
	}
	vm.truncateStack(receiverIndex)
	return vm.finishCanonicalStructFastCall(instr, "resolved_method", dispatch, inst, nil, callNode)
}
