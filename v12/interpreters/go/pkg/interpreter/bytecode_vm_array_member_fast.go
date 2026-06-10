package interpreter

import (
	"math"
	"strings"
	"unicode/utf8"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type bytecodeMemberMethodFastPathKind uint8

const (
	bytecodeMemberMethodFastPathNone bytecodeMemberMethodFastPathKind = iota
	bytecodeMemberMethodFastPathArrayLen
	bytecodeMemberMethodFastPathArrayGet
	bytecodeMemberMethodFastPathArrayPush
	bytecodeMemberMethodFastPathArrayReadSlot
	bytecodeMemberMethodFastPathArrayWriteSlot
	bytecodeMemberMethodFastPathArrayReadWriteSlot
	bytecodeMemberMethodFastPathStringLenBytes
	bytecodeMemberMethodFastPathStringContains
	bytecodeMemberMethodFastPathStringReplace
	bytecodeMemberMethodFastPathStringBytes
	bytecodeMemberMethodFastPathStringChars
	bytecodeMemberMethodFastPathStringByteIteratorNext
	bytecodeMemberMethodFastPathStringCharIteratorNext
	bytecodeMemberMethodFastPathStringBuilderPushChar
	bytecodeMemberMethodFastPathStringBuilderPushByte
	bytecodeMemberMethodFastPathStringBuilderPushBytes
	bytecodeMemberMethodFastPathStringBuilderPushString
	bytecodeMemberMethodFastPathStringBuilderAppendBuilder
	bytecodeMemberMethodFastPathStringBuilderFinish
)

type bytecodeMemberMethodFastPathCacheKey struct {
	fn           *runtime.FunctionValue
	member       string
	receiverKind bytecodeMemberReceiverKind
	structDef    *runtime.StructDefinitionValue
	primitiveKey bytecodeMemberPrimitiveCacheKey
}

type bytecodeStringBytesIteratorNative struct {
	text string
}

var bytecodeStringBytesIteratorTypeArgs = []ast.TypeExpression{cachedIntegerTypeExpression(runtime.IntegerU8)}

func (vm *bytecodeVM) memberMethodFastPathFor(key bytecodeMemberMethodCacheKey, template runtime.Value) bytecodeMemberMethodFastPathKind {
	if !key.preferMethods || vm == nil || vm.interp == nil {
		return bytecodeMemberMethodFastPathNone
	}
	fn, ok := template.(*runtime.FunctionValue)
	if !ok || fn == nil {
		return bytecodeMemberMethodFastPathNone
	}
	def, ok := fn.Declaration.(*ast.FunctionDefinition)
	if !ok || def == nil || def.ID == nil || def.ID.Name != key.member {
		return bytecodeMemberMethodFastPathNone
	}
	origin := vm.interp.nodeOrigins[def]
	switch key.receiverKind {
	case bytecodeMemberReceiverArray:
		switch key.member {
		case "len":
			if isCanonicalAbleKernelOrigin(origin) && typeExpressionToString(def.ReturnType) == "i32" {
				return bytecodeMemberMethodFastPathArrayLen
			}
		case "get":
			if isCanonicalAbleStdlibOrigin(origin, "collections/array.able") {
				if _, ok := def.ReturnType.(*ast.NullableTypeExpression); ok {
					return bytecodeMemberMethodFastPathArrayGet
				}
			}
		case "push":
			if isCanonicalAbleKernelOrigin(origin) && typeExpressionToString(def.ReturnType) == "void" {
				return bytecodeMemberMethodFastPathArrayPush
			}
		case "read_slot":
			if isCanonicalAbleKernelOrigin(origin) && isCanonicalArrayReadSlotFunction(def) {
				return bytecodeMemberMethodFastPathArrayReadSlot
			}
		case "write_slot":
			if isCanonicalAbleKernelOrigin(origin) && isCanonicalArrayWriteSlotFunction(def) {
				return bytecodeMemberMethodFastPathArrayWriteSlot
			}
		}
	case bytecodeMemberReceiverString:
		if !isCanonicalAbleStdlibOrigin(origin, "text/string.able") {
			return bytecodeMemberMethodFastPathNone
		}
		switch key.member {
		case "len_bytes":
			if typeExpressionToString(def.ReturnType) == "u64" {
				return bytecodeMemberMethodFastPathStringLenBytes
			}
		case "contains":
			if typeExpressionToString(def.ReturnType) == "bool" {
				return bytecodeMemberMethodFastPathStringContains
			}
		case "replace":
			if typeExpressionToString(def.ReturnType) == "String" {
				return bytecodeMemberMethodFastPathStringReplace
			}
		case "bytes":
			if isStringBytesReturnType(def.ReturnType) {
				return bytecodeMemberMethodFastPathStringBytes
			}
		case "chars":
			if isStringCharsReturnType(def.ReturnType) {
				return bytecodeMemberMethodFastPathStringChars
			}
		}
	case bytecodeMemberReceiverStruct:
		if kind := vm.memberMethodStructStringFastPath(key, def, origin); kind != bytecodeMemberMethodFastPathNone {
			return kind
		}
		return vm.memberMethodStructCanonicalStdlibFastPath(key, def, origin)
	}
	return bytecodeMemberMethodFastPathNone
}

func (vm *bytecodeVM) memberMethodFastPathForFunction(key bytecodeMemberMethodCacheKey, fn *runtime.FunctionValue) bytecodeMemberMethodFastPathKind {
	if fn == nil {
		return bytecodeMemberMethodFastPathNone
	}
	cacheKey := bytecodeMemberMethodFastPathCacheKey{
		fn:           fn,
		member:       key.member,
		receiverKind: key.receiverKind,
		structDef:    key.structDef,
		primitiveKey: key.primitiveKey,
	}
	if vm != nil && vm.memberMethodFastPaths != nil {
		if cached, ok := vm.memberMethodFastPaths[cacheKey]; ok {
			return cached
		}
	}
	kind := vm.memberMethodFastPathFor(key, fn)
	if vm != nil {
		if vm.memberMethodFastPaths == nil {
			vm.memberMethodFastPaths = make(map[bytecodeMemberMethodFastPathCacheKey]bytecodeMemberMethodFastPathKind, 8)
		}
		vm.memberMethodFastPaths[cacheKey] = kind
	}
	return kind
}

func (vm *bytecodeVM) resolvedMemberMethodFastPath(memberName string, receiver runtime.Value, fn *runtime.FunctionValue) bytecodeMemberMethodFastPathKind {
	identity, ok := bytecodeMemberReceiverIdentityForValue(receiver)
	if !ok {
		return bytecodeMemberMethodFastPathNone
	}
	return vm.memberMethodFastPathForFunction(bytecodeMemberMethodCacheKey{
		member:        memberName,
		preferMethods: true,
		receiverKind:  identity.kind,
		structDef:     identity.structDef,
		primitiveKey:  identity.primitiveKey,
	}, fn)
}

func bytecodeMemberFastPathReceiverKind(receiver runtime.Value) (bytecodeMemberReceiverKind, bool) {
	identity, ok := bytecodeMemberReceiverIdentityForValue(receiver)
	if !ok {
		return bytecodeMemberReceiverUnknown, false
	}
	return identity.kind, true
}

func bytecodeResolvedMemberFastPathFunction(callable runtime.Value) (*runtime.FunctionValue, bool) {
	fn, _, _, ok := inlineCallFunctionValue(callable)
	if !ok || fn == nil {
		return nil, false
	}
	return fn, true
}

func (vm *bytecodeVM) execCachedMemberMethodFastPath(kind bytecodeMemberMethodFastPathKind, instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	switch kind {
	case bytecodeMemberMethodFastPathArrayLen:
		return vm.execArrayLenMemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathArrayGet:
		return vm.execArrayGetMemberFast(instr, receiverIndex, argBase, callNode)
	case bytecodeMemberMethodFastPathArrayPush:
		return vm.execArrayPushMemberFast(instr.name, instr.argCount, instr.node, receiverIndex, argBase, callNode)
	case bytecodeMemberMethodFastPathArrayReadSlot:
		return vm.execArrayReadSlotMemberFast(instr.name, instr.argCount, instr.node, receiverIndex, argBase, callNode)
	case bytecodeMemberMethodFastPathArrayWriteSlot:
		return vm.execArrayWriteSlotMemberFast(instr.name, instr.argCount, instr.node, receiverIndex, argBase, callNode)
	case bytecodeMemberMethodFastPathStringLenBytes:
		return vm.execStringLenBytesMemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathStringContains:
		return vm.execStringContainsMemberFast(instr, receiverIndex, argBase, callNode)
	case bytecodeMemberMethodFastPathStringReplace:
		return vm.execStringReplaceMemberFast(instr, receiverIndex, argBase, callNode)
	case bytecodeMemberMethodFastPathStringBytes:
		return vm.execStringBytesMemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathStringChars:
		return vm.execStringCharsMemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathStringByteIteratorNext:
		return vm.execStringByteIteratorNextMemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathStringCharIteratorNext:
		return vm.execStringCharIteratorNextMemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathStringBuilderPushChar:
		return vm.execStringBuilderPushCharMemberFast(instr, receiverIndex, argBase, callNode)
	case bytecodeMemberMethodFastPathStringBuilderPushByte:
		return vm.execStringBuilderPushByteMemberFast(instr, receiverIndex, argBase, callNode)
	case bytecodeMemberMethodFastPathStringBuilderPushBytes:
		return vm.execStringBuilderPushBytesMemberFast(instr, receiverIndex, argBase, callNode)
	case bytecodeMemberMethodFastPathStringBuilderPushString:
		return vm.execStringBuilderPushStringMemberFast(instr, receiverIndex, argBase, callNode)
	case bytecodeMemberMethodFastPathStringBuilderAppendBuilder:
		return vm.execStringBuilderAppendBuilderMemberFast(instr, receiverIndex, argBase, callNode)
	case bytecodeMemberMethodFastPathStringBuilderFinish:
		return vm.execStringBuilderFinishMemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathRandomNextI64:
		return vm.execRandomNextI64MemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathRandomNextI32:
		return vm.execRandomNextI32MemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathRandomNextF64:
		return vm.execRandomNextF64MemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathInt128ToI128:
		return vm.execInt128ToI128MemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathInt128ToI64:
		return vm.execInt128ToI64MemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathInt128ToU64:
		return vm.execInt128ToU64MemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathInt128IsZero:
		return vm.execInt128IsZeroMemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathInt128IsNegative:
		return vm.execInt128IsNegativeMemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathInt128Add:
		return vm.execInt128BinaryMemberFast(instr, receiverIndex, argBase, "+", "int128_add_fast", callNode)
	case bytecodeMemberMethodFastPathInt128Sub:
		return vm.execInt128BinaryMemberFast(instr, receiverIndex, argBase, "-", "int128_sub_fast", callNode)
	case bytecodeMemberMethodFastPathInt128Mul:
		return vm.execInt128BinaryMemberFast(instr, receiverIndex, argBase, "*", "int128_mul_fast", callNode)
	case bytecodeMemberMethodFastPathInt128Div:
		return vm.execInt128BinaryMemberFast(instr, receiverIndex, argBase, "//", "int128_div_fast", callNode)
	case bytecodeMemberMethodFastPathInt128Rem:
		return vm.execInt128BinaryMemberFast(instr, receiverIndex, argBase, "%", "int128_rem_fast", callNode)
	case bytecodeMemberMethodFastPathUInt128ToU128:
		return vm.execUInt128ToU128MemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathUInt128ToU64:
		return vm.execUInt128ToU64MemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathUInt128ToI64:
		return vm.execUInt128ToI64MemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathUInt128IsZero:
		return vm.execUInt128IsZeroMemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathUInt128IsPositive:
		return vm.execUInt128IsPositiveMemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathUInt128Add:
		return vm.execUInt128AddMemberFast(instr, receiverIndex, argBase, callNode)
	case bytecodeMemberMethodFastPathUInt128Sub:
		return vm.execUInt128SubMemberFast(instr, receiverIndex, argBase, callNode)
	case bytecodeMemberMethodFastPathUInt128Mul:
		return vm.execUInt128MulMemberFast(instr, receiverIndex, argBase, callNode)
	case bytecodeMemberMethodFastPathUInt128Div:
		return vm.execUInt128BinaryMemberFast(instr, receiverIndex, argBase, "//", "uint128_div_fast", callNode)
	case bytecodeMemberMethodFastPathUInt128Rem:
		return vm.execUInt128BinaryMemberFast(instr, receiverIndex, argBase, "%", "uint128_rem_fast", callNode)
	default:
		return nil, false, nil
	}
}

func (vm *bytecodeVM) execStaticArrayNewMemberFast(instr bytecodeInstruction, receiver runtime.Value, callee runtime.Value, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 0 || instr.name != "new" || receiverIndex < 0 || receiverIndex >= vm.stackDepth() || vm == nil || vm.interp == nil {
		return nil, false, nil
	}
	if !bytecodeCanonicalArrayDefinitionReceiver(vm.interp, receiver) {
		return nil, false, nil
	}
	fn, ok := bytecodeSingleFunction(callee)
	if !ok || !vm.isCanonicalArrayNewFunction(fn) {
		return nil, false, nil
	}
	return vm.finishStaticArrayNewMemberFast(instr, receiverIndex, callNode)
}

func (vm *bytecodeVM) finishStaticArrayNewMemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if receiverIndex < 0 || receiverIndex >= vm.stackDepth() || vm == nil || vm.interp == nil {
		return nil, false, nil
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "member_access", "array_new_fast", instr.node)
	}
	vm.truncateStack(receiverIndex)
	arr := vm.interp.newArrayValue(nil, 0)
	vm.trackBytecodeArrayOwnershipCreation(arr)
	newProg, finishErr := vm.finishCompletedCall(arr, nil, callNode, nil)
	return newProg, true, finishErr
}

func bytecodeSingleFunction(callee runtime.Value) (*runtime.FunctionValue, bool) {
	switch fn := callee.(type) {
	case *runtime.FunctionValue:
		return fn, fn != nil
	case *runtime.FunctionOverloadValue:
		if fn == nil || len(fn.Overloads) != 1 || fn.Overloads[0] == nil {
			return nil, false
		}
		return fn.Overloads[0], true
	default:
		return nil, false
	}
}

func bytecodeCanonicalArrayDefinitionReceiver(interp *Interpreter, receiver runtime.Value) bool {
	var def *runtime.StructDefinitionValue
	switch v := receiver.(type) {
	case *runtime.StructDefinitionValue:
		def = v
	case runtime.StructDefinitionValue:
		def = &v
	default:
		return false
	}
	if def == nil || def.Node == nil || def.Node.ID == nil || def.Node.ID.Name != "Array" || interp == nil {
		return false
	}
	return isCanonicalAbleKernelOrigin(interp.nodeOrigins[def.Node])
}

func (vm *bytecodeVM) isCanonicalArrayNewFunction(fn *runtime.FunctionValue) bool {
	if vm == nil || vm.interp == nil || fn == nil {
		return false
	}
	def, ok := fn.Declaration.(*ast.FunctionDefinition)
	if !ok || def == nil || def.ID == nil || def.ID.Name != "new" {
		return false
	}
	if len(def.Params) != 0 || !isArrayReturnType(def.ReturnType) {
		return false
	}
	return isCanonicalAbleKernelOrigin(vm.interp.nodeOrigins[def])
}

func isArrayReturnType(expr ast.TypeExpression) bool {
	switch t := expr.(type) {
	case *ast.GenericTypeExpression:
		return typeExpressionToString(t.Base) == "Array"
	case *ast.SimpleTypeExpression:
		return t != nil && t.Name != nil && t.Name.Name == "Array"
	default:
		return false
	}
}

func (vm *bytecodeVM) arraySizeI32Fast(arr *runtime.ArrayValue) (int, bool, error) {
	handle, ok, err := vm.arrayHandleFast(arr)
	if !ok || err != nil {
		return 0, ok, err
	}
	size, err := runtime.ArrayStoreSize(handle)
	if err != nil {
		return 0, true, err
	}
	if size < 0 || size > 1<<31-1 {
		return 0, false, nil
	}
	return size, true, nil
}

func (vm *bytecodeVM) arrayHandleFast(arr *runtime.ArrayValue) (int64, bool, error) {
	if vm == nil || vm.interp == nil || arr == nil {
		return 0, false, nil
	}
	if arr.Handle != 0 {
		return arr.Handle, true, nil
	}
	if arr.TrackedHandle != 0 {
		return arr.TrackedHandle, true, nil
	}
	if _, err := vm.interp.ensureArrayState(arr, 0); err != nil {
		return 0, true, err
	}
	if arr.Handle == 0 {
		return 0, false, nil
	}
	return arr.Handle, true, nil
}

func bytecodeArrayGetIndexI32(val runtime.Value) (int64, bool) {
	switch raw := val.(type) {
	case bytecodeRawI32SlotValue:
		return int64(raw), true
	case *bytecodeRawI32StackCell:
		if raw == nil {
			return 0, false
		}
		return int64(raw.Val), true
	case runtime.IntegerValue:
		if raw.IsSmall() {
			idx := raw.Int64Fast()
			if err := ensureFitsInt64Type(runtime.IntegerI32, idx); err != nil {
				return 0, false
			}
			return idx, true
		}
	case *runtime.IntegerValue:
		if raw != nil && raw.IsSmallRef() {
			idx := raw.Int64FastRef()
			if err := ensureFitsInt64Type(runtime.IntegerI32, idx); err != nil {
				return 0, false
			}
			return idx, true
		}
	}
	_, idx, ok := bytecodeRawIntegerValueInfo(val)
	if !ok {
		return 0, false
	}
	if err := ensureFitsInt64Type(runtime.IntegerI32, idx); err != nil {
		return 0, false
	}
	return idx, true
}

func (vm *bytecodeVM) execStringLenBytesMemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 0 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() {
		return nil, false, nil
	}
	length, ok := vm.bytecodeValidatedStringLenBytesFast(vm.stackValue(receiverIndex))
	if !ok {
		return nil, false, nil
	}
	if length > math.MaxInt32 {
		return nil, false, nil
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "string_len_bytes_fast", instr.node)
	}
	vm.truncateStack(receiverIndex)
	newProg, finishErr := vm.finishCompletedCall(boxedOrSmallIntegerValue(runtime.IntegerU64, int64(length)), nil, callNode, nil)
	return newProg, true, finishErr
}

func (vm *bytecodeVM) execStringContainsMemberFast(instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 1 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() || argBase < 0 || argBase >= vm.stackDepth() {
		return nil, false, nil
	}
	haystack, ok := vm.bytecodeValidatedStringTextFast(vm.stackValue(receiverIndex))
	if !ok {
		return nil, false, nil
	}
	needle, ok := vm.bytecodeValidatedStringTextFast(vm.stackValue(argBase))
	if !ok {
		return nil, false, nil
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "string_contains_fast", instr.node)
	}
	vm.truncateStack(receiverIndex)
	newProg, finishErr := vm.finishCompletedCall(runtime.BoolValue{Val: strings.Contains(haystack, needle)}, nil, callNode, nil)
	return newProg, true, finishErr
}

func (vm *bytecodeVM) execStringReplaceMemberFast(instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 2 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() || argBase < 0 || argBase+1 >= vm.stackDepth() {
		return nil, false, nil
	}
	receiver := vm.stackValue(receiverIndex)
	haystack, ok := vm.bytecodeValidatedStringTextFast(receiver)
	if !ok {
		return nil, false, nil
	}
	old, ok := vm.bytecodeValidatedStringTextFast(vm.stackValue(argBase))
	if !ok {
		return nil, false, nil
	}
	replacement, ok := vm.bytecodeValidatedStringTextFast(vm.stackValue(argBase + 1))
	if !ok {
		return nil, false, nil
	}
	result := haystack
	if old != "" {
		result = strings.ReplaceAll(haystack, old, replacement)
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "string_replace_fast", instr.node)
	}
	vm.truncateStack(receiverIndex)
	var resultValue runtime.Value = runtime.StringValue{Val: result}
	if _, ok := vm.bytecodeCanonicalStringInstance(receiver); ok {
		bytes := vm.interp.newU8ArrayValueFromString(result)
		canonical, ok := vm.canonicalStringStructValue(bytes, len(result))
		if !ok {
			return nil, false, nil
		}
		resultValue = canonical
	}
	newProg, finishErr := vm.finishCompletedCall(resultValue, nil, callNode, nil)
	return newProg, true, finishErr
}

func bytecodeStringValueFast(value runtime.Value) (string, bool) {
	switch v := value.(type) {
	case runtime.StringValue:
		return v.Val, true
	case *runtime.StringValue:
		if v == nil {
			return "", false
		}
		return v.Val, true
	default:
		return "", false
	}
}

func isStringBytesReturnType(expr ast.TypeExpression) bool {
	generic, ok := expr.(*ast.GenericTypeExpression)
	if !ok || generic == nil || len(generic.Arguments) != 1 {
		return false
	}
	return typeExpressionToString(generic.Base) == "Iterator" &&
		typeExpressionToString(generic.Arguments[0]) == "u8"
}

func isCanonicalStringBuilderPushMethod(def *ast.FunctionDefinition, argType string) bool {
	if def == nil || def.ID == nil || len(def.Params) != 2 || typeExpressionToString(def.ReturnType) != "void" {
		return false
	}
	return typeExpressionToString(def.Params[0].ParamType) == "Self" &&
		typeExpressionToString(def.Params[1].ParamType) == argType
}

func isCanonicalStringBuilderFinishMethod(def *ast.FunctionDefinition) bool {
	if def == nil || def.ID == nil || len(def.Params) != 1 {
		return false
	}
	if typeExpressionToString(def.Params[0].ParamType) != "Self" {
		return false
	}
	generic, ok := def.ReturnType.(*ast.GenericTypeExpression)
	return ok && generic != nil &&
		typeExpressionToString(generic.Base) == "Result" &&
		len(generic.Arguments) == 1 &&
		typeExpressionToString(generic.Arguments[0]) == "String"
}

func (vm *bytecodeVM) execStringBytesMemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 0 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() || vm == nil || vm.interp == nil {
		return nil, false, nil
	}
	iterDef, ok := vm.canonicalStringBytesIteratorDefinition()
	if !ok {
		return nil, false, nil
	}
	bytes, byteLen, ok := vm.bytecodeValidatedStringIteratorBytesFast(vm.stackValue(receiverIndex))
	if !ok || byteLen > math.MaxInt32 {
		return nil, false, nil
	}
	iter := bytecodeNewStringIteratorStruct(iterDef, bytes, byteLen)
	result, ok, err := vm.canonicalStringBytesIteratorInterfaceValue(iter)
	if err == nil && !ok {
		result, err = vm.interp.coerceToInterfaceValue("Iterator", iter, bytecodeStringBytesIteratorTypeArgs)
	}
	if err != nil {
		vm.truncateStack(receiverIndex)
		newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
		return newProg, true, finishErr
	}
	bytecodeEnsureIteratorSelfMethod(result)
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "string_bytes_fast", instr.node)
	}
	vm.truncateStack(receiverIndex)
	newProg, finishErr := vm.finishCompletedCall(result, nil, callNode, nil)
	return newProg, true, finishErr
}

func (vm *bytecodeVM) execStringBuilderPushCharMemberFast(instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 1 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() || argBase < 0 || argBase >= vm.stackDepth() {
		return nil, false, nil
	}
	buffer, ok := vm.bytecodeCanonicalStringBuilderBuffer(vm.stackValue(receiverIndex))
	if !ok {
		return nil, false, nil
	}
	charVal, ok := bytecodeCharValueFast(vm.stackValue(argBase))
	if !ok || !utf8.ValidRune(charVal) {
		return nil, false, nil
	}
	var encoded [utf8.UTFMax]byte
	size := 1
	if charVal < utf8.RuneSelf {
		encoded[0] = byte(charVal)
	} else {
		size = utf8.EncodeRune(encoded[:], charVal)
	}
	if !vm.appendArrayU8BytesFast(buffer, encoded[:size]) {
		return nil, false, nil
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "string_builder_push_char_fast", instr.node)
	}
	vm.truncateStack(receiverIndex)
	if vm.canSkipFollowingPop(nil) {
		vm.ip += 2
		return nil, true, nil
	}
	newProg, finishErr := vm.finishCompletedCall(runtime.VoidValue{}, nil, callNode, nil)
	return newProg, true, finishErr
}

func (vm *bytecodeVM) execStringBuilderPushByteMemberFast(instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 1 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() || argBase < 0 || argBase >= vm.stackDepth() {
		return nil, false, nil
	}
	buffer, ok := vm.bytecodeCanonicalStringBuilderBuffer(vm.stackValue(receiverIndex))
	if !ok {
		return nil, false, nil
	}
	byteVal, ok := bytecodeU8ValueFast(vm.stackValue(argBase))
	if !ok || !vm.appendArrayU8ValueFast(buffer, byteVal) {
		return nil, false, nil
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "string_builder_push_byte_fast", instr.node)
	}
	vm.truncateStack(receiverIndex)
	if vm.canSkipFollowingPop(nil) {
		vm.ip += 2
		return nil, true, nil
	}
	newProg, finishErr := vm.finishCompletedCall(runtime.VoidValue{}, nil, callNode, nil)
	return newProg, true, finishErr
}

func (vm *bytecodeVM) execStringBuilderPushBytesMemberFast(instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 1 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() || argBase < 0 || argBase >= vm.stackDepth() {
		return nil, false, nil
	}
	buffer, ok := vm.bytecodeCanonicalStringBuilderBuffer(vm.stackValue(receiverIndex))
	if !ok {
		return nil, false, nil
	}
	source, ok := vm.stackValue(argBase).(*runtime.ArrayValue)
	if !ok || source == nil {
		return nil, false, nil
	}
	bytes, sourceHandle, ok := bytecodeBorrowArrayU8BytesFast(source)
	if !ok {
		return nil, false, nil
	}
	if buffer.Handle == sourceHandle && sourceHandle != 0 {
		bytes = append([]byte(nil), bytes...)
	}
	if !vm.appendArrayU8BytesFast(buffer, bytes) {
		return nil, false, nil
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "string_builder_push_bytes_fast", instr.node)
	}
	vm.truncateStack(receiverIndex)
	if vm.canSkipFollowingPop(nil) {
		vm.ip += 2
		return nil, true, nil
	}
	newProg, finishErr := vm.finishCompletedCall(runtime.VoidValue{}, nil, callNode, nil)
	return newProg, true, finishErr
}

func (vm *bytecodeVM) execStringBuilderPushStringMemberFast(instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 1 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() || argBase < 0 || argBase >= vm.stackDepth() {
		return nil, false, nil
	}
	buffer, ok := vm.bytecodeCanonicalStringBuilderBuffer(vm.stackValue(receiverIndex))
	if !ok {
		return nil, false, nil
	}
	text, ok := bytecodeStringValueFast(vm.stackValue(argBase))
	if !ok || !utf8.ValidString(text) || !vm.appendArrayU8StringFast(buffer, text) {
		return nil, false, nil
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "string_builder_push_string_fast", instr.node)
	}
	vm.truncateStack(receiverIndex)
	if vm.canSkipFollowingPop(nil) {
		vm.ip += 2
		return nil, true, nil
	}
	newProg, finishErr := vm.finishCompletedCall(runtime.VoidValue{}, nil, callNode, nil)
	return newProg, true, finishErr
}

func (vm *bytecodeVM) execStringBuilderAppendBuilderMemberFast(instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 1 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() || argBase < 0 || argBase >= vm.stackDepth() {
		return nil, false, nil
	}
	buffer, ok := vm.bytecodeCanonicalStringBuilderBuffer(vm.stackValue(receiverIndex))
	if !ok {
		return nil, false, nil
	}
	sourceBuffer, ok := vm.bytecodeCanonicalStringBuilderBuffer(vm.stackValue(argBase))
	if !ok {
		return nil, false, nil
	}
	bytes, sourceHandle, ok := bytecodeBorrowArrayU8BytesFast(sourceBuffer)
	if !ok {
		return nil, false, nil
	}
	if buffer.Handle == sourceHandle && sourceHandle != 0 {
		bytes = append([]byte(nil), bytes...)
	}
	if !vm.appendArrayU8BytesFast(buffer, bytes) {
		return nil, false, nil
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "string_builder_append_builder_fast", instr.node)
	}
	vm.truncateStack(receiverIndex)
	if vm.canSkipFollowingPop(nil) {
		vm.ip += 2
		return nil, true, nil
	}
	newProg, finishErr := vm.finishCompletedCall(runtime.VoidValue{}, nil, callNode, nil)
	return newProg, true, finishErr
}

func (vm *bytecodeVM) execStringBuilderFinishMemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 0 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() {
		return nil, false, nil
	}
	buffer, ok := vm.bytecodeCanonicalStringBuilderBuffer(vm.stackValue(receiverIndex))
	if !ok {
		return nil, false, nil
	}
	bytes, _, ok := bytecodeBorrowArrayU8BytesFast(buffer)
	if !ok || !utf8.Valid(bytes) {
		return nil, false, nil
	}
	result, ok := vm.canonicalStringStructValue(buffer, len(bytes))
	if !ok {
		return nil, false, nil
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "string_builder_finish_fast", instr.node)
	}
	vm.truncateStack(receiverIndex)
	newProg, finishErr := vm.finishCompletedCall(result, nil, callNode, nil)
	return newProg, true, finishErr
}

func bytecodeCharValueFast(value runtime.Value) (rune, bool) {
	switch v := value.(type) {
	case runtime.CharValue:
		return v.Val, true
	case *runtime.CharValue:
		if v == nil {
			return 0, false
		}
		return v.Val, true
	default:
		return 0, false
	}
}

func bytecodeU8ValueFast(value runtime.Value) (uint8, bool) {
	kind, raw, ok := bytecodeRawIntegerValueInfo(value)
	if !ok || kind != runtime.IntegerU8 || raw < 0 || raw > 255 {
		return 0, false
	}
	return uint8(raw), true
}

func bytecodeBorrowArrayU8BytesFast(arr *runtime.ArrayValue) ([]byte, int64, bool) {
	if arr == nil {
		return nil, 0, false
	}
	handle := arr.Handle
	if handle == 0 {
		handle = arr.TrackedHandle
	}
	if handle == 0 {
		return nil, 0, false
	}
	bytes, ok, err := runtime.ArrayStoreMonoBorrowedU8BytesIfAvailable(handle)
	if err != nil || !ok {
		return nil, 0, false
	}
	return bytes, handle, true
}

func isStringByteIteratorNextReturnType(expr ast.TypeExpression) bool {
	union, ok := expr.(*ast.UnionTypeExpression)
	if !ok || union == nil || len(union.Members) != 2 {
		return false
	}
	hasU8 := false
	hasIteratorEnd := false
	for _, member := range union.Members {
		switch typeExpressionToString(member) {
		case "u8":
			hasU8 = true
		case "IteratorEnd":
			hasIteratorEnd = true
		}
	}
	return hasU8 && hasIteratorEnd
}

func (vm *bytecodeVM) execStringByteIteratorNextMemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 0 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() || vm == nil || vm.interp == nil {
		return nil, false, nil
	}
	inst, ok := bytecodeStringByteIteratorInstance(vm.stackValue(receiverIndex))
	if !ok {
		return nil, false, nil
	}
	bytes, ok := bytecodeStructArrayField(inst, "bytes")
	if !ok {
		return nil, false, nil
	}
	offset, ok := bytecodeI32StructField(inst, "offset")
	if !ok {
		return nil, false, nil
	}
	length, ok := bytecodeI32StructField(inst, "len_bytes")
	if !ok {
		return nil, false, nil
	}
	var result runtime.Value
	if offset >= length {
		result = runtime.IteratorEnd
	} else {
		rawByte, ok := bytecodeReadNativeStringByteIterator(inst, offset)
		if ok {
			result = boxedOrSmallIntegerValue(runtime.IntegerU8, int64(rawByte))
			if !structSetNamedFieldValue(inst, "offset", boxedOrSmallIntegerValue(runtime.IntegerI32, offset+1)) {
				return nil, false, nil
			}
		} else {
			rawByte, ok, err := bytecodeReadMonoU8Array(bytes, offset)
			if err != nil {
				vm.truncateStack(receiverIndex)
				newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
				return newProg, true, finishErr
			}
			if ok {
				result = boxedOrSmallIntegerValue(runtime.IntegerU8, int64(rawByte))
				if !structSetNamedFieldValue(inst, "offset", boxedOrSmallIntegerValue(runtime.IntegerI32, offset+1)) {
					return nil, false, nil
				}
			} else {
				handle, ok, err := vm.arrayHandleFast(bytes)
				if err != nil {
					vm.truncateStack(receiverIndex)
					newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
					return newProg, true, finishErr
				}
				if !ok {
					return nil, false, nil
				}
				size, err := runtime.ArrayStoreSize(handle)
				if err != nil {
					vm.truncateStack(receiverIndex)
					newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
					return newProg, true, finishErr
				}
				if offset < 0 || offset >= int64(size) {
					result = runtime.IteratorEnd
				} else {
					result, err = runtime.ArrayStoreRead(handle, int(offset))
					if err != nil {
						vm.truncateStack(receiverIndex)
						newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
						return newProg, true, finishErr
					}
					if isNilRuntimeValue(result) {
						result = runtime.IteratorEnd
					} else if !bytecodeIsU8Value(result) {
						return nil, false, nil
					} else {
						if !structSetNamedFieldValue(inst, "offset", boxedOrSmallIntegerValue(runtime.IntegerI32, offset+1)) {
							return nil, false, nil
						}
					}
				}
			}
		}
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "member_access", "string_byte_iter_next_fast", instr.node)
	}
	vm.truncateStack(receiverIndex)
	newProg, finishErr := vm.finishCompletedCall(result, nil, callNode, nil)
	return newProg, true, finishErr
}

func bytecodeReadNativeStringByteIterator(inst *runtime.StructInstanceValue, offset int64) (byte, bool) {
	if inst == nil || offset < 0 || offset > math.MaxInt32 {
		return 0, false
	}
	native, ok := inst.Native.(bytecodeStringBytesIteratorNative)
	if !ok || int(offset) >= len(native.text) {
		return 0, false
	}
	return native.text[int(offset)], true
}

func bytecodeReadMonoU8Array(arr *runtime.ArrayValue, offset int64) (uint8, bool, error) {
	if arr == nil || arr.Handle == 0 || offset < 0 || offset > math.MaxInt32 {
		return 0, false, nil
	}
	return runtime.ArrayStoreMonoReadU8IfAvailable(arr.Handle, int(offset))
}

func bytecodeStringByteIteratorInstance(value runtime.Value) (*runtime.StructInstanceValue, bool) {
	for {
		switch v := value.(type) {
		case *runtime.InterfaceValue:
			if v == nil {
				return nil, false
			}
			value = v.Underlying
			continue
		case runtime.InterfaceValue:
			value = v.Underlying
			continue
		case *runtime.StructInstanceValue:
			if v == nil || v.Definition == nil || v.Definition.Node == nil || v.Definition.Node.ID == nil || !structUsesNamedFieldStorage(v) {
				return nil, false
			}
			switch v.Definition.Node.ID.Name {
			case "RawStringBytesIter", "StringBytesIter":
				return v, true
			default:
				return nil, false
			}
		default:
			return nil, false
		}
	}
}

func bytecodeI32StructField(inst *runtime.StructInstanceValue, name string) (int64, bool) {
	if inst == nil {
		return 0, false
	}
	raw, ok := structNamedFieldValue(inst, name)
	if !ok {
		return 0, false
	}
	intVal, ok := raw.(runtime.IntegerValue)
	if !ok {
		return 0, false
	}
	value, ok := intVal.ToInt64()
	if !ok {
		return 0, false
	}
	if err := ensureFitsInt64Type(runtime.IntegerI32, value); err != nil {
		return 0, false
	}
	return value, true
}

func bytecodeIsU8Value(value runtime.Value) bool {
	intVal, ok := value.(runtime.IntegerValue)
	return ok && intVal.TypeSuffix == runtime.IntegerU8
}
