package interpreter

import (
	"fmt"
	"unicode/utf8"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

var bytecodeStringCharsIteratorTypeArgs = []ast.TypeExpression{cachedSimpleTypeExpression("char")}

func isStringCharsReturnType(expr ast.TypeExpression) bool {
	generic, ok := expr.(*ast.GenericTypeExpression)
	if !ok || generic == nil || len(generic.Arguments) != 1 {
		return false
	}
	return typeExpressionToString(generic.Base) == "Iterator" &&
		typeExpressionToString(generic.Arguments[0]) == "char"
}

func isStringCharIteratorNextReturnType(expr ast.TypeExpression) bool {
	union, ok := expr.(*ast.UnionTypeExpression)
	if !ok || union == nil || len(union.Members) != 2 {
		return false
	}
	hasChar := false
	hasIteratorEnd := false
	for _, member := range union.Members {
		switch typeExpressionToString(member) {
		case "char":
			hasChar = true
		case "IteratorEnd":
			hasIteratorEnd = true
		}
	}
	return hasChar && hasIteratorEnd
}

func (vm *bytecodeVM) memberMethodStructStringFastPath(key bytecodeMemberMethodCacheKey, def *ast.FunctionDefinition, origin string) bytecodeMemberMethodFastPathKind {
	if vm == nil || !isCanonicalAbleStdlibOrigin(origin, "text/string.able") {
		return bytecodeMemberMethodFastPathNone
	}
	if key.structDef == nil {
		if key.member == "next" && isStringByteIteratorNextReturnType(def.ReturnType) {
			return bytecodeMemberMethodFastPathStringByteIteratorNext
		}
		if key.member == "next" && isStringCharIteratorNextReturnType(def.ReturnType) {
			return bytecodeMemberMethodFastPathStringCharIteratorNext
		}
		return bytecodeMemberMethodFastPathNone
	}
	switch {
	case vm.isCanonicalStructDefinition(key.structDef, "String", "text/string.able"):
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
	case vm.isCanonicalStructDefinition(key.structDef, "StringBuilder", "text/string.able"):
		switch key.member {
		case "push_char":
			if isCanonicalStringBuilderPushMethod(def, "char") {
				return bytecodeMemberMethodFastPathStringBuilderPushChar
			}
		case "push_byte":
			if isCanonicalStringBuilderPushMethod(def, "u8") {
				return bytecodeMemberMethodFastPathStringBuilderPushByte
			}
		case "push_bytes":
			if isCanonicalStringBuilderPushMethod(def, "Array u8") {
				return bytecodeMemberMethodFastPathStringBuilderPushBytes
			}
		case "push_string":
			if isCanonicalStringBuilderPushMethod(def, "String") {
				return bytecodeMemberMethodFastPathStringBuilderPushString
			}
		case "append_builder":
			if isCanonicalStringBuilderPushMethod(def, "StringBuilder") {
				return bytecodeMemberMethodFastPathStringBuilderAppendBuilder
			}
		case "finish":
			if isCanonicalStringBuilderFinishMethod(def) {
				return bytecodeMemberMethodFastPathStringBuilderFinish
			}
		}
	case vm.isCanonicalStringBytesIteratorStructDefinition(key.structDef):
		if key.member == "next" && isStringByteIteratorNextReturnType(def.ReturnType) {
			return bytecodeMemberMethodFastPathStringByteIteratorNext
		}
	case vm.isCanonicalStringCharsIteratorStructDefinition(key.structDef):
		if key.member == "next" && isStringCharIteratorNextReturnType(def.ReturnType) {
			return bytecodeMemberMethodFastPathStringCharIteratorNext
		}
	}
	return bytecodeMemberMethodFastPathNone
}

func (vm *bytecodeVM) isCanonicalStringBytesIteratorStructDefinition(def *runtime.StructDefinitionValue) bool {
	if def == nil {
		return false
	}
	return vm.isCanonicalStructDefinition(def, "RawStringBytesIter", "text/string.able") ||
		vm.isCanonicalStructDefinition(def, "StringBytesIter", "text/string.able")
}

func (vm *bytecodeVM) isCanonicalStringCharsIteratorStructDefinition(def *runtime.StructDefinitionValue) bool {
	if def == nil {
		return false
	}
	return vm.isCanonicalStructDefinition(def, "RawStringCharsIter", "text/string.able") ||
		vm.isCanonicalStructDefinition(def, "StringCharsIter", "text/string.able")
}

func bytecodeStructArrayField(inst *runtime.StructInstanceValue, name string) (*runtime.ArrayValue, bool) {
	raw, ok := structNamedFieldValue(inst, name)
	if !ok {
		return nil, false
	}
	arr, ok := raw.(*runtime.ArrayValue)
	if !ok || arr == nil {
		return nil, false
	}
	return arr, true
}

func (vm *bytecodeVM) bytecodeCanonicalStringInstance(value runtime.Value) (*runtime.StructInstanceValue, bool) {
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
			if typed == nil || typed.Definition == nil {
				return nil, false
			}
			if def, ok := vm.canonicalStringDefinition(); ok && typed.Definition == def {
				return typed, true
			}
			if vm.isCanonicalStructDefinition(typed.Definition, "String", "text/string.able") {
				return typed, true
			}
			return nil, false
		default:
			return nil, false
		}
	}
}

func (vm *bytecodeVM) bytecodeCanonicalStringBytesFast(value runtime.Value) (*runtime.ArrayValue, []byte, bool) {
	if vm == nil || vm.interp == nil {
		return nil, nil, false
	}
	inst, ok := vm.bytecodeCanonicalStringInstance(value)
	if !ok {
		return nil, nil, false
	}
	arr, ok := bytecodeStructArrayField(inst, "bytes")
	if !ok {
		return nil, nil, false
	}
	bytes, ok := externU8ArrayBytes(vm.interp, arr, true)
	if !ok {
		return nil, nil, false
	}
	valid := utf8.Valid(bytes)
	vm.interp.recordBytecodeStringCanonicalValidation(len(bytes), valid)
	if !valid {
		return nil, nil, false
	}
	return arr, bytes, true
}

func (vm *bytecodeVM) bytecodeValidatedStringTextFast(value runtime.Value) (string, bool) {
	if text, ok := bytecodeStringValueFast(value); ok {
		return text, true
	}
	_, bytes, ok := vm.bytecodeCanonicalStringBytesFast(value)
	if !ok {
		return "", false
	}
	return string(bytes), true
}

func (vm *bytecodeVM) bytecodeValidatedStringLenBytesFast(value runtime.Value) (int, bool) {
	if text, ok := bytecodeStringValueFast(value); ok {
		return len(text), true
	}
	_, bytes, ok := vm.bytecodeCanonicalStringBytesFast(value)
	if !ok {
		return 0, false
	}
	return len(bytes), true
}

func (vm *bytecodeVM) bytecodeValidatedStringIteratorBytesFast(value runtime.Value) (*runtime.ArrayValue, int, bool) {
	if vm == nil || vm.interp == nil {
		return nil, 0, false
	}
	if text, ok := bytecodeStringValueFast(value); ok {
		valid := utf8.ValidString(text)
		vm.interp.recordBytecodeStringRawValidation(len(text), valid)
		if !valid {
			return nil, 0, false
		}
		return vm.interp.newU8ArrayValueFromString(text), len(text), true
	}
	arr, bytes, ok := vm.bytecodeCanonicalStringBytesFast(value)
	if !ok {
		return nil, 0, false
	}
	return arr, len(bytes), true
}

func (vm *bytecodeVM) execStringCharsMemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 0 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() || vm == nil || vm.interp == nil {
		return nil, false, nil
	}
	iterDef, ok := vm.canonicalStringCharsIteratorDefinition()
	if !ok {
		return nil, false, nil
	}
	bytes, byteLen, ok := vm.bytecodeValidatedStringIteratorBytesFast(vm.stackValue(receiverIndex))
	if !ok {
		return nil, false, nil
	}
	iter := bytecodeNewStringIteratorStruct(iterDef, bytes, byteLen)
	result, ok, err := vm.canonicalStringCharsIteratorInterfaceValue(iter)
	if err == nil && !ok {
		result, err = vm.interp.coerceToInterfaceValue("Iterator", iter, bytecodeStringCharsIteratorTypeArgs)
	}
	if err != nil {
		vm.truncateStack(receiverIndex)
		newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
		return newProg, true, finishErr
	}
	bytecodeEnsureIteratorSelfMethod(result)
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "string_chars_fast", instr.node)
	}
	vm.truncateStack(receiverIndex)
	newProg, finishErr := vm.finishCompletedCall(result, nil, callNode, nil)
	return newProg, true, finishErr
}

func bytecodeNewStringIteratorStruct(iterDef *runtime.StructDefinitionValue, bytes *runtime.ArrayValue, byteLen int) *runtime.StructInstanceValue {
	offset := boxedOrSmallIntegerValue(runtime.IntegerI32, 0)
	length := boxedOrSmallIntegerValue(runtime.IntegerI32, int64(byteLen))
	if iter, ok := newNamedStructInstancePositionalStorage(iterDef, nil); ok &&
		structSetNamedFieldValue(iter, "bytes", bytes) &&
		structSetNamedFieldValue(iter, "offset", offset) &&
		structSetNamedFieldValue(iter, "len_bytes", length) {
		return iter
	}
	return &runtime.StructInstanceValue{
		Definition: iterDef,
		Fields: map[string]runtime.Value{
			"bytes":     bytes,
			"offset":    offset,
			"len_bytes": length,
		},
	}
}

func (vm *bytecodeVM) canonicalStringCharsIteratorInterfaceValue(iter *runtime.StructInstanceValue) (runtime.Value, bool, error) {
	if iter == nil {
		return nil, false, nil
	}
	iterDef, ok := vm.canonicalStringCharsIteratorDefinition()
	if !ok || iter.Definition != iterDef {
		return nil, false, nil
	}
	ifaceDef, ok := vm.canonicalIteratorInterfaceDefinition()
	if !ok {
		return nil, false, nil
	}
	nextMethod, ok, err := vm.canonicalStringCharsIteratorNextMethod()
	if err != nil || !ok {
		return nil, ok, err
	}
	return &runtime.InterfaceValue{
		Interface:  ifaceDef,
		Underlying: iter,
		Methods: map[string]runtime.Value{
			"next":     nextMethod,
			"iterator": bytecodeIteratorSelfNativeMethod(),
		},
		InterfaceArgs: bytecodeStringCharsIteratorTypeArgs,
	}, true, nil
}

func bytecodeEnsureIteratorSelfMethod(value runtime.Value) {
	iface, ok := value.(*runtime.InterfaceValue)
	if !ok {
		return
	}
	if iface == nil {
		return
	}
	if _, ok := interfaceValueLookupMethod(iface, "iterator"); ok {
		return
	}
	interfaceValueSetMethod(iface, "iterator", bytecodeIteratorSelfNativeMethod())
}

func bytecodeIteratorSelfNativeMethod() runtime.NativeFunctionValue {
	return runtime.NativeFunctionValue{
		Name:  "iterator.iterator",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("iterator expects only a receiver")
			}
			return args[0], nil
		},
	}
}

func (vm *bytecodeVM) canonicalStringCharsIteratorNextMethod() (runtime.Value, bool, error) {
	if vm == nil || vm.interp == nil {
		return nil, false, nil
	}
	version := vm.bytecodeMethodCacheVersion()
	globalRev := vm.bytecodeGlobalRevision()
	if vm.stringCharsIteratorNextSet &&
		vm.stringCharsIteratorNextVersion == version &&
		vm.stringCharsIteratorNextGlobalRev == globalRev {
		return vm.stringCharsIteratorNextMethod, vm.stringCharsIteratorNextMethod != nil, nil
	}
	method, err := vm.interp.findMethod(
		typeInfo{name: "RawStringCharsIter"},
		"next",
		"Iterator",
		bytecodeStringCharsIteratorTypeArgs,
	)
	if err != nil {
		return nil, false, err
	}
	if method == nil {
		vm.stringCharsIteratorNextMethod = nil
		vm.stringCharsIteratorNextVersion = version
		vm.stringCharsIteratorNextGlobalRev = globalRev
		vm.stringCharsIteratorNextSet = true
		return nil, false, nil
	}
	fn := firstFunction(method)
	if fn == nil {
		return nil, false, nil
	}
	def, ok := fn.Declaration.(*ast.FunctionDefinition)
	if !ok || def == nil || def.ID == nil || def.ID.Name != "next" ||
		!isStringCharIteratorNextReturnType(def.ReturnType) ||
		!isCanonicalAbleStdlibOrigin(vm.interp.nodeOrigins[def], "text/string.able") {
		return nil, false, nil
	}
	vm.stringCharsIteratorNextMethod = method
	vm.stringCharsIteratorNextVersion = version
	vm.stringCharsIteratorNextGlobalRev = globalRev
	vm.stringCharsIteratorNextSet = true
	return method, true, nil
}

func (vm *bytecodeVM) canonicalStringCharsIteratorDefinition() (*runtime.StructDefinitionValue, bool) {
	if vm == nil || vm.interp == nil {
		return nil, false
	}
	if vm.stringCharsIterDefSet {
		return vm.stringCharsIterDef, vm.stringCharsIterDef != nil
	}
	vm.stringCharsIterDefSet = true
	def, ok := vm.lookupCanonicalStructDefinition("RawStringCharsIter", "text/string.able")
	if ok {
		vm.stringCharsIterDef = def
	}
	return def, ok
}

func (vm *bytecodeVM) execCanonicalStringCharIteratorNextCallMemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if vm == nil || vm.interp == nil || instr.name != "next" || instr.argCount != 0 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() {
		return nil, false, nil
	}
	if !vm.isCanonicalStringCharIteratorInterfaceReceiver(vm.stackValue(receiverIndex)) {
		return nil, false, nil
	}
	if _, ok, err := vm.canonicalStringCharsIteratorNextMethod(); err != nil || !ok {
		return nil, err != nil, err
	}
	return vm.execStringCharIteratorNextMemberFast(instr, receiverIndex, callNode)
}

func (vm *bytecodeVM) isCanonicalStringCharIteratorInterfaceReceiver(value runtime.Value) bool {
	switch iface := value.(type) {
	case *runtime.InterfaceValue:
		return vm.isCanonicalStringCharIteratorInterfaceValue(iface)
	case runtime.InterfaceValue:
		return vm.isCanonicalStringCharIteratorInterfaceValue(&iface)
	default:
		return false
	}
}

func (vm *bytecodeVM) isCanonicalStringCharIteratorInterfaceValue(iface *runtime.InterfaceValue) bool {
	if iface == nil || !vm.isCanonicalIteratorCharInterface(iface) {
		return false
	}
	inst, ok := iface.Underlying.(*runtime.StructInstanceValue)
	if !ok || inst == nil {
		return false
	}
	return vm.isCanonicalRawStringCharIteratorInstance(inst)
}

func (vm *bytecodeVM) isCanonicalIteratorCharInterface(iface *runtime.InterfaceValue) bool {
	if iface == nil {
		return false
	}
	def, ok := vm.canonicalIteratorInterfaceDefinition()
	if !ok || iface.Interface != def {
		return false
	}
	return bytecodeIsCanonicalCharInterfaceArgs(iface.InterfaceArgs)
}

func (vm *bytecodeVM) isCanonicalRawStringCharIteratorInstance(inst *runtime.StructInstanceValue) bool {
	if inst == nil {
		return false
	}
	def, ok := vm.canonicalStringCharsIteratorDefinition()
	return ok && inst.Definition == def
}

func bytecodeIsCanonicalCharInterfaceArgs(args []ast.TypeExpression) bool {
	return len(args) == 1 && args[0] == bytecodeStringCharsIteratorTypeArgs[0]
}

func (vm *bytecodeVM) execStringCharIteratorNextMemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 0 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() || vm == nil || vm.interp == nil {
		return nil, false, nil
	}
	inst, ok := bytecodeStringCharIteratorInstance(vm.stackValue(receiverIndex))
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
		raw, ok := externU8ArrayBytes(vm.interp, bytes, true)
		if !ok || length < 0 || offset < 0 || int(length) > len(raw) || int(offset) >= len(raw) {
			return nil, false, nil
		}
		r, size := utf8.DecodeRune(raw[int(offset):int(length)])
		vm.interp.recordBytecodeStringRuneDecode(size, r != utf8.RuneError || size != 1)
		if r == utf8.RuneError && size == 1 {
			return nil, false, nil
		}
		if !structSetNamedFieldValue(inst, "offset", boxedOrSmallIntegerValue(runtime.IntegerI32, offset+int64(size))) {
			return nil, false, nil
		}
		result = runtime.CharValue{Val: r}
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "member_access", "string_char_iter_next_fast", instr.node)
	}
	vm.truncateStack(receiverIndex)
	newProg, finishErr := vm.finishCompletedCall(result, nil, callNode, nil)
	return newProg, true, finishErr
}

func bytecodeStringCharIteratorInstance(value runtime.Value) (*runtime.StructInstanceValue, bool) {
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
			if typed == nil || typed.Definition == nil || typed.Definition.Node == nil || typed.Definition.Node.ID == nil {
				return nil, false
			}
			switch typed.Definition.Node.ID.Name {
			case "RawStringCharsIter", "StringCharsIter":
				return typed, true
			default:
				return nil, false
			}
		default:
			return nil, false
		}
	}
}
