package interpreter

import (
	"math"
	"path/filepath"
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
	bytecodeMemberMethodFastPathStringLenBytes
	bytecodeMemberMethodFastPathStringContains
	bytecodeMemberMethodFastPathStringReplace
	bytecodeMemberMethodFastPathStringBytes
	bytecodeMemberMethodFastPathStringByteIteratorNext
)

type bytecodeMemberMethodFastPathCacheKey struct {
	fn           *runtime.FunctionValue
	member       string
	receiverKind bytecodeMemberReceiverKind
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
		}
	case bytecodeMemberReceiverStruct:
		if key.member == "next" &&
			isCanonicalAbleStdlibOrigin(origin, "text/string.able") &&
			isStringByteIteratorNextReturnType(def.ReturnType) {
			return bytecodeMemberMethodFastPathStringByteIteratorNext
		}
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
	receiverKind, ok := bytecodeMemberFastPathReceiverKind(receiver)
	if !ok {
		return bytecodeMemberMethodFastPathNone
	}
	return vm.memberMethodFastPathForFunction(bytecodeMemberMethodCacheKey{
		member:        memberName,
		preferMethods: true,
		receiverKind:  receiverKind,
	}, fn)
}

func bytecodeMemberFastPathReceiverKind(receiver runtime.Value) (bytecodeMemberReceiverKind, bool) {
	switch v := receiver.(type) {
	case *runtime.ArrayValue:
		if v == nil {
			return bytecodeMemberReceiverUnknown, false
		}
		return bytecodeMemberReceiverArray, true
	case runtime.StringValue:
		return bytecodeMemberReceiverString, true
	case *runtime.StringValue:
		if v == nil {
			return bytecodeMemberReceiverUnknown, false
		}
		return bytecodeMemberReceiverString, true
	case *runtime.StructInstanceValue:
		if v == nil || v.Definition == nil {
			return bytecodeMemberReceiverUnknown, false
		}
		return bytecodeMemberReceiverStruct, true
	default:
		return bytecodeMemberReceiverUnknown, false
	}
}

func bytecodeResolvedMemberFastPathFunction(callable runtime.Value) (*runtime.FunctionValue, bool) {
	fn, _, _, ok := inlineCallFunctionValue(callable)
	if !ok || fn == nil {
		return nil, false
	}
	return fn, true
}

func isCanonicalAbleStdlibOrigin(origin string, relative string) bool {
	if origin == "" || relative == "" {
		return false
	}
	origin = filepath.ToSlash(origin)
	relative = strings.TrimPrefix(filepath.ToSlash(relative), "/")
	return hasCanonicalPathSuffix(origin, "/able-stdlib/src/", relative) ||
		hasCanonicalPathSuffix(origin, "/pkg/src/", relative)
}

func hasCanonicalPathSuffix(origin string, base string, relative string) bool {
	if relative == "" || !strings.HasSuffix(origin, relative) {
		return false
	}
	prefixLen := len(origin) - len(relative)
	return prefixLen >= len(base) && strings.HasSuffix(origin[:prefixLen], base)
}

func isCanonicalAbleKernelOrigin(origin string) bool {
	if origin == "" {
		return false
	}
	origin = filepath.ToSlash(origin)
	return strings.HasSuffix(origin, "/v12/kernel/src/kernel.able") ||
		strings.HasSuffix(origin, "/kernel/src/kernel.able")
}

func (vm *bytecodeVM) execCachedMemberMethodFastPath(kind bytecodeMemberMethodFastPathKind, instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	switch kind {
	case bytecodeMemberMethodFastPathArrayLen:
		return vm.execArrayLenMemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathArrayGet:
		return vm.execArrayGetMemberFast(instr, receiverIndex, argBase, callNode)
	case bytecodeMemberMethodFastPathArrayPush:
		return vm.execArrayPushMemberFast(instr, receiverIndex, argBase, callNode)
	case bytecodeMemberMethodFastPathStringLenBytes:
		return vm.execStringLenBytesMemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathStringContains:
		return vm.execStringContainsMemberFast(instr, receiverIndex, argBase, callNode)
	case bytecodeMemberMethodFastPathStringReplace:
		return vm.execStringReplaceMemberFast(instr, receiverIndex, argBase, callNode)
	case bytecodeMemberMethodFastPathStringBytes:
		return vm.execStringBytesMemberFast(instr, receiverIndex, callNode)
	case bytecodeMemberMethodFastPathStringByteIteratorNext:
		return vm.execStringByteIteratorNextMemberFast(instr, receiverIndex, callNode)
	default:
		return nil, false, nil
	}
}

func (vm *bytecodeVM) execArrayLenMemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 0 || receiverIndex < 0 || receiverIndex >= len(vm.stack) {
		return nil, false, nil
	}
	arr, ok := vm.stack[receiverIndex].(*runtime.ArrayValue)
	if !ok || arr == nil {
		return nil, false, nil
	}
	size, ok, err := vm.arraySizeI32Fast(arr)
	if err != nil {
		vm.stack = vm.stack[:receiverIndex]
		newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
		return newProg, true, finishErr
	}
	if !ok {
		return nil, false, nil
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "array_len_fast", instr.node)
	}
	vm.stack = vm.stack[:receiverIndex]
	newProg, finishErr := vm.finishCompletedCall(boxedOrSmallIntegerValue(runtime.IntegerI32, int64(size)), nil, callNode, nil)
	return newProg, true, finishErr
}

func (vm *bytecodeVM) execArrayGetMemberFast(instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 1 || receiverIndex < 0 || receiverIndex >= len(vm.stack) || argBase < 0 || argBase >= len(vm.stack) {
		return nil, false, nil
	}
	arr, ok := vm.stack[receiverIndex].(*runtime.ArrayValue)
	if !ok || arr == nil {
		return nil, false, nil
	}
	idx, ok := bytecodeArrayGetIndexI32(vm.stack[argBase])
	if !ok {
		return nil, false, nil
	}
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		size := len(state.Values)
		if size > 1<<31-1 {
			return nil, false, nil
		}
		var result runtime.Value
		if idx < 0 || idx >= int64(size) {
			result = runtime.NilValue{}
		} else {
			result = state.Values[int(idx)]
		}
		if vm.interp != nil {
			vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "array_get_tracked_fast", instr.node)
		}
		vm.stack = vm.stack[:receiverIndex]
		newProg, finishErr := vm.finishCompletedCall(result, nil, callNode, nil)
		return newProg, true, finishErr
	}
	handle, ok, err := vm.arrayHandleFast(arr)
	if err != nil {
		vm.stack = vm.stack[:receiverIndex]
		newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
		return newProg, true, finishErr
	}
	if !ok {
		return nil, false, nil
	}
	size, err := runtime.ArrayStoreSize(handle)
	if err != nil {
		vm.stack = vm.stack[:receiverIndex]
		newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
		return newProg, true, finishErr
	}
	if size < 0 || size > 1<<31-1 {
		return nil, false, nil
	}
	var result runtime.Value
	if idx < 0 || idx >= int64(size) {
		result = runtime.NilValue{}
	} else {
		result, err = runtime.ArrayStoreRead(handle, int(idx))
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "array_get_fast", instr.node)
	}
	vm.stack = vm.stack[:receiverIndex]
	newProg, finishErr := vm.finishCompletedCall(result, err, callNode, nil)
	return newProg, true, finishErr
}

func (vm *bytecodeVM) execArrayPushMemberFast(instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 1 || receiverIndex < 0 || receiverIndex >= len(vm.stack) || argBase < 0 || argBase >= len(vm.stack) {
		return nil, false, nil
	}
	arr, ok := vm.stack[receiverIndex].(*runtime.ArrayValue)
	if !ok || arr == nil || vm == nil || vm.interp == nil {
		return nil, false, nil
	}
	value := vm.stack[argBase]
	state, tracked := bytecodeTrackedArrayState(arr)
	if !tracked {
		var err error
		state, err = vm.interp.ensureArrayState(arr, 0)
		if err != nil {
			vm.stack = vm.stack[:receiverIndex]
			newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
			return newProg, true, finishErr
		}
	}
	idx := len(state.Values)
	runtime.ArrayEnsureCapacity(state, idx+1)
	state.Values = append(state.Values, value)
	if state.Capacity < cap(state.Values) {
		state.Capacity = cap(state.Values)
	}
	vm.interp.syncTrackedArrayWrite(arr, state, idx, value)
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "array_push_fast", instr.node)
	}
	vm.stack = vm.stack[:receiverIndex]
	newProg, finishErr := vm.finishCompletedCall(runtime.VoidValue{}, nil, callNode, nil)
	return newProg, true, finishErr
}

func (vm *bytecodeVM) execStaticArrayNewMemberFast(instr bytecodeInstruction, receiver runtime.Value, callee runtime.Value, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 0 || instr.name != "new" || receiverIndex < 0 || receiverIndex >= len(vm.stack) || vm == nil || vm.interp == nil {
		return nil, false, nil
	}
	if !bytecodeCanonicalArrayDefinitionReceiver(vm.interp, receiver) {
		return nil, false, nil
	}
	fn, ok := bytecodeSingleFunction(callee)
	if !ok || !vm.isCanonicalArrayNewFunction(fn) {
		return nil, false, nil
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "member_access", "array_new_fast", instr.node)
	}
	vm.stack = vm.stack[:receiverIndex]
	newProg, finishErr := vm.finishCompletedCall(vm.interp.newArrayValue(nil, 0), nil, callNode, nil)
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
	intVal, ok := val.(runtime.IntegerValue)
	if !ok {
		return 0, false
	}
	idx, ok := intVal.ToInt64()
	if !ok {
		return 0, false
	}
	if err := ensureFitsInt64Type(runtime.IntegerI32, idx); err != nil {
		return 0, false
	}
	return idx, true
}

func (vm *bytecodeVM) execStringLenBytesMemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 0 || receiverIndex < 0 || receiverIndex >= len(vm.stack) {
		return nil, false, nil
	}
	text, ok := bytecodeStringValueFast(vm.stack[receiverIndex])
	if !ok {
		return nil, false, nil
	}
	length := len(text)
	if length > math.MaxInt32 {
		return nil, false, nil
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "string_len_bytes_fast", instr.node)
	}
	vm.stack = vm.stack[:receiverIndex]
	newProg, finishErr := vm.finishCompletedCall(boxedOrSmallIntegerValue(runtime.IntegerU64, int64(length)), nil, callNode, nil)
	return newProg, true, finishErr
}

func (vm *bytecodeVM) execStringContainsMemberFast(instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 1 || receiverIndex < 0 || receiverIndex >= len(vm.stack) || argBase < 0 || argBase >= len(vm.stack) {
		return nil, false, nil
	}
	haystack, ok := bytecodeStringValueFast(vm.stack[receiverIndex])
	if !ok {
		return nil, false, nil
	}
	needle, ok := bytecodeStringValueFast(vm.stack[argBase])
	if !ok {
		return nil, false, nil
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "string_contains_fast", instr.node)
	}
	vm.stack = vm.stack[:receiverIndex]
	newProg, finishErr := vm.finishCompletedCall(runtime.BoolValue{Val: strings.Contains(haystack, needle)}, nil, callNode, nil)
	return newProg, true, finishErr
}

func (vm *bytecodeVM) execStringReplaceMemberFast(instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 2 || receiverIndex < 0 || receiverIndex >= len(vm.stack) || argBase < 0 || argBase+1 >= len(vm.stack) {
		return nil, false, nil
	}
	haystack, ok := bytecodeStringValueFast(vm.stack[receiverIndex])
	if !ok {
		return nil, false, nil
	}
	old, ok := bytecodeStringValueFast(vm.stack[argBase])
	if !ok {
		return nil, false, nil
	}
	replacement, ok := bytecodeStringValueFast(vm.stack[argBase+1])
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
	vm.stack = vm.stack[:receiverIndex]
	newProg, finishErr := vm.finishCompletedCall(runtime.StringValue{Val: result}, nil, callNode, nil)
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

func (vm *bytecodeVM) execStringBytesMemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 0 || receiverIndex < 0 || receiverIndex >= len(vm.stack) || vm == nil || vm.interp == nil {
		return nil, false, nil
	}
	text, ok := bytecodeStringValueFast(vm.stack[receiverIndex])
	if !ok {
		return nil, false, nil
	}
	if !utf8.ValidString(text) {
		return nil, false, nil
	}
	if len(text) > math.MaxInt32 {
		return nil, false, nil
	}
	iterDef, ok := vm.canonicalStringBytesIteratorDefinition()
	if !ok {
		return nil, false, nil
	}
	byteLen := len(text)
	bytes := vm.interp.newU8ArrayValueFromString(text)
	iter := &runtime.StructInstanceValue{
		Definition: iterDef,
		Native:     bytecodeStringBytesIteratorNative{text: text},
		Fields: map[string]runtime.Value{
			"bytes":     bytes,
			"offset":    boxedOrSmallIntegerValue(runtime.IntegerI32, 0),
			"len_bytes": boxedOrSmallIntegerValue(runtime.IntegerI32, int64(byteLen)),
		},
	}
	result, ok, err := vm.canonicalStringBytesIteratorInterfaceValue(iter)
	if err == nil && !ok {
		result, err = vm.interp.coerceToInterfaceValue("Iterator", iter, bytecodeStringBytesIteratorTypeArgs)
	}
	if err != nil {
		vm.stack = vm.stack[:receiverIndex]
		newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
		return newProg, true, finishErr
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "string_bytes_fast", instr.node)
	}
	vm.stack = vm.stack[:receiverIndex]
	newProg, finishErr := vm.finishCompletedCall(result, nil, callNode, nil)
	return newProg, true, finishErr
}

func (vm *bytecodeVM) canonicalStringBytesIteratorInterfaceValue(iter *runtime.StructInstanceValue) (runtime.Value, bool, error) {
	if iter == nil {
		return nil, false, nil
	}
	iterDef, ok := vm.canonicalStringBytesIteratorDefinition()
	if !ok || iter.Definition != iterDef {
		return nil, false, nil
	}
	ifaceDef, ok := vm.canonicalIteratorInterfaceDefinition()
	if !ok {
		return nil, false, nil
	}
	nextMethod, ok, err := vm.canonicalStringBytesIteratorNextMethod()
	if err != nil || !ok {
		return nil, ok, err
	}
	return &runtime.InterfaceValue{
		Interface:     ifaceDef,
		Underlying:    iter,
		Methods:       map[string]runtime.Value{"next": nextMethod},
		InterfaceArgs: bytecodeStringBytesIteratorTypeArgs,
	}, true, nil
}

func (vm *bytecodeVM) canonicalIteratorInterfaceDefinition() (*runtime.InterfaceDefinitionValue, bool) {
	if vm == nil || vm.interp == nil {
		return nil, false
	}
	if vm.stringBytesIteratorInterfaceDefSet {
		return vm.stringBytesIteratorInterfaceDef, vm.stringBytesIteratorInterfaceDef != nil
	}
	vm.stringBytesIteratorInterfaceDefSet = true
	def, ok := vm.interp.interfaces["Iterator"]
	if !ok || def == nil || def.Node == nil || def.Node.ID == nil || def.Node.ID.Name != "Iterator" {
		return nil, false
	}
	if !isCanonicalAbleStdlibOrigin(vm.interp.nodeOrigins[def.Node], "core/iteration.able") {
		return nil, false
	}
	vm.stringBytesIteratorInterfaceDef = def
	return def, true
}

func (vm *bytecodeVM) canonicalStringBytesIteratorNextMethod() (runtime.Value, bool, error) {
	if vm == nil || vm.interp == nil {
		return nil, false, nil
	}
	version := vm.bytecodeMethodCacheVersion()
	globalRev := vm.bytecodeGlobalRevision()
	if vm.stringBytesIteratorNextSet &&
		vm.stringBytesIteratorNextVersion == version &&
		vm.stringBytesIteratorNextGlobalRev == globalRev {
		return vm.stringBytesIteratorNextMethod, vm.stringBytesIteratorNextMethod != nil, nil
	}
	method, err := vm.interp.findMethod(
		typeInfo{name: "RawStringBytesIter"},
		"next",
		"Iterator",
		bytecodeStringBytesIteratorTypeArgs,
	)
	if err != nil {
		return nil, false, err
	}
	if method == nil {
		vm.stringBytesIteratorNextMethod = nil
		vm.stringBytesIteratorNextVersion = version
		vm.stringBytesIteratorNextGlobalRev = globalRev
		vm.stringBytesIteratorNextSet = true
		return nil, false, nil
	}
	fn := firstFunction(method)
	if fn == nil {
		return nil, false, nil
	}
	def, ok := fn.Declaration.(*ast.FunctionDefinition)
	if !ok || def == nil || def.ID == nil || def.ID.Name != "next" ||
		!isStringByteIteratorNextReturnType(def.ReturnType) ||
		!isCanonicalAbleStdlibOrigin(vm.interp.nodeOrigins[def], "text/string.able") {
		return nil, false, nil
	}
	vm.stringBytesIteratorNextMethod = method
	vm.stringBytesIteratorNextVersion = version
	vm.stringBytesIteratorNextGlobalRev = globalRev
	vm.stringBytesIteratorNextSet = true
	return method, true, nil
}

func (vm *bytecodeVM) canonicalStringBytesIteratorDefinition() (*runtime.StructDefinitionValue, bool) {
	if vm == nil || vm.interp == nil {
		return nil, false
	}
	if vm.stringBytesIterDefSet {
		return vm.stringBytesIterDef, vm.stringBytesIterDef != nil
	}
	vm.stringBytesIterDefSet = true
	def, ok := vm.lookupCanonicalStructDefinition("RawStringBytesIter", "text/string.able")
	if ok {
		vm.stringBytesIterDef = def
	}
	return def, ok
}

func (vm *bytecodeVM) lookupCanonicalStructDefinition(name string, originRelative string) (*runtime.StructDefinitionValue, bool) {
	if vm == nil || vm.interp == nil || name == "" || originRelative == "" {
		return nil, false
	}
	if def, ok := vm.interp.lookupStructDefinition(name); ok && vm.isCanonicalStructDefinition(def, name, originRelative) {
		return def, true
	}
	seen := make(map[string]struct{}, len(vm.interp.packageEnvs)+len(vm.interp.dynamicPackageEnvs)+len(vm.interp.packageRegistry))
	for pkgName := range vm.interp.packageEnvs {
		seen[pkgName] = struct{}{}
		if def, ok := vm.interp.lookupStructDefinitionInPackage(pkgName, name); ok && vm.isCanonicalStructDefinition(def, name, originRelative) {
			return def, true
		}
	}
	for pkgName := range vm.interp.dynamicPackageEnvs {
		if _, ok := seen[pkgName]; ok {
			continue
		}
		seen[pkgName] = struct{}{}
		if def, ok := vm.interp.lookupStructDefinitionInPackage(pkgName, name); ok && vm.isCanonicalStructDefinition(def, name, originRelative) {
			return def, true
		}
	}
	for pkgName := range vm.interp.packageRegistry {
		if _, ok := seen[pkgName]; ok {
			continue
		}
		if def, ok := vm.interp.lookupStructDefinitionInPackage(pkgName, name); ok && vm.isCanonicalStructDefinition(def, name, originRelative) {
			return def, true
		}
	}
	return nil, false
}

func (vm *bytecodeVM) isCanonicalStructDefinition(def *runtime.StructDefinitionValue, name string, originRelative string) bool {
	if vm == nil || vm.interp == nil || def == nil || def.Node == nil || def.Node.ID == nil || def.Node.ID.Name != name {
		return false
	}
	return isCanonicalAbleStdlibOrigin(vm.interp.nodeOrigins[def.Node], originRelative)
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
	if instr.argCount != 0 || receiverIndex < 0 || receiverIndex >= len(vm.stack) || vm == nil || vm.interp == nil {
		return nil, false, nil
	}
	inst, ok := bytecodeStringByteIteratorInstance(vm.stack[receiverIndex])
	if !ok {
		return nil, false, nil
	}
	bytes, ok := inst.Fields["bytes"].(*runtime.ArrayValue)
	if !ok || bytes == nil {
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
			inst.Fields["offset"] = boxedOrSmallIntegerValue(runtime.IntegerI32, offset+1)
		} else {
			rawByte, ok, err := bytecodeReadMonoU8Array(bytes, offset)
			if err != nil {
				vm.stack = vm.stack[:receiverIndex]
				newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
				return newProg, true, finishErr
			}
			if ok {
				result = boxedOrSmallIntegerValue(runtime.IntegerU8, int64(rawByte))
				inst.Fields["offset"] = boxedOrSmallIntegerValue(runtime.IntegerI32, offset+1)
			} else {
				handle, ok, err := vm.arrayHandleFast(bytes)
				if err != nil {
					vm.stack = vm.stack[:receiverIndex]
					newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
					return newProg, true, finishErr
				}
				if !ok {
					return nil, false, nil
				}
				size, err := runtime.ArrayStoreSize(handle)
				if err != nil {
					vm.stack = vm.stack[:receiverIndex]
					newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
					return newProg, true, finishErr
				}
				if offset < 0 || offset >= int64(size) {
					result = runtime.IteratorEnd
				} else {
					result, err = runtime.ArrayStoreRead(handle, int(offset))
					if err != nil {
						vm.stack = vm.stack[:receiverIndex]
						newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
						return newProg, true, finishErr
					}
					if isNilRuntimeValue(result) {
						result = runtime.IteratorEnd
					} else if !bytecodeIsU8Value(result) {
						return nil, false, nil
					} else {
						inst.Fields["offset"] = boxedOrSmallIntegerValue(runtime.IntegerI32, offset+1)
					}
				}
			}
		}
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "member_access", "string_byte_iter_next_fast", instr.node)
	}
	vm.stack = vm.stack[:receiverIndex]
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
			if v == nil || v.Definition == nil || v.Definition.Node == nil || v.Definition.Node.ID == nil || v.Fields == nil {
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
	if inst == nil || inst.Fields == nil {
		return 0, false
	}
	intVal, ok := inst.Fields[name].(runtime.IntegerValue)
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
