package interpreter

import "able/interpreter-go/pkg/runtime"

type bytecodeMemberReceiverKind uint8

const (
	bytecodeMemberReceiverUnknown bytecodeMemberReceiverKind = iota
	bytecodeMemberReceiverArray
	bytecodeMemberReceiverString
	bytecodeMemberReceiverStruct
)

type bytecodeMemberMethodCacheKey struct {
	program       *bytecodeProgram
	ip            int
	member        string
	preferMethods bool
	receiverKind  bytecodeMemberReceiverKind
	structDef     *runtime.StructDefinitionValue
}

type bytecodeMemberMethodCacheEntry struct {
	globalRevision     uint64
	methodCacheVersion uint64
	methodTemplate     runtime.Value
}

func (vm *bytecodeVM) canUseMemberMethodCache(memberName string, preferMethods bool) bool {
	if vm == nil || vm.interp == nil || vm.interp.global == nil || vm.env != vm.interp.global {
		return false
	}
	return preferMethods && memberName != ""
}

func (vm *bytecodeVM) memberMethodCacheKey(program *bytecodeProgram, ip int, memberName string, preferMethods bool, receiver runtime.Value) (bytecodeMemberMethodCacheKey, bool) {
	if !vm.canUseMemberMethodCache(memberName, preferMethods) {
		return bytecodeMemberMethodCacheKey{}, false
	}
	key := bytecodeMemberMethodCacheKey{
		program:       program,
		ip:            ip,
		member:        memberName,
		preferMethods: preferMethods,
	}
	switch v := receiver.(type) {
	case *runtime.ArrayValue:
		if v == nil {
			return bytecodeMemberMethodCacheKey{}, false
		}
		key.receiverKind = bytecodeMemberReceiverArray
	case runtime.StringValue:
		key.receiverKind = bytecodeMemberReceiverString
	case *runtime.StringValue:
		if v == nil {
			return bytecodeMemberMethodCacheKey{}, false
		}
		key.receiverKind = bytecodeMemberReceiverString
	case *runtime.StructInstanceValue:
		if v == nil || v.Definition == nil {
			return bytecodeMemberMethodCacheKey{}, false
		}
		key.receiverKind = bytecodeMemberReceiverStruct
		key.structDef = v.Definition
	default:
		return bytecodeMemberMethodCacheKey{}, false
	}
	return key, true
}

func extractMemberMethodTemplate(resolved runtime.Value) (runtime.Value, bool) {
	switch method := resolved.(type) {
	case runtime.NativeBoundMethodValue:
		return method.Method, true
	case *runtime.NativeBoundMethodValue:
		if method == nil {
			return nil, false
		}
		return method.Method, true
	case runtime.BoundMethodValue:
		return method.Method, true
	case *runtime.BoundMethodValue:
		if method == nil {
			return nil, false
		}
		return method.Method, true
	default:
		return nil, false
	}
}

func bindMemberMethodTemplate(receiver runtime.Value, template runtime.Value) (runtime.Value, bool) {
	switch method := template.(type) {
	case runtime.NativeFunctionValue:
		return runtime.NativeBoundMethodValue{Receiver: receiver, Method: method}, true
	case *runtime.NativeFunctionValue:
		if method == nil {
			return nil, false
		}
		return runtime.NativeBoundMethodValue{Receiver: receiver, Method: *method}, true
	case *runtime.FunctionValue, *runtime.FunctionOverloadValue:
		return runtime.BoundMethodValue{Receiver: receiver, Method: method}, true
	default:
		return nil, false
	}
}

func (vm *bytecodeVM) lookupCachedMemberMethod(program *bytecodeProgram, ip int, memberName string, preferMethods bool, receiver runtime.Value) (runtime.Value, bool) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil {
		return nil, false
	}
	key, ok := vm.memberMethodCacheKey(program, ip, memberName, preferMethods, receiver)
	if !ok {
		return nil, false
	}
	if vm.memberMethodCache == nil {
		vm.interp.recordBytecodeMemberMethodCacheMiss()
		return nil, false
	}
	entry, ok := vm.memberMethodCache[key]
	if !ok {
		vm.interp.recordBytecodeMemberMethodCacheMiss()
		return nil, false
	}
	if entry.globalRevision != vm.interp.global.Revision() {
		vm.interp.recordBytecodeMemberMethodCacheMiss()
		return nil, false
	}
	if entry.methodCacheVersion != vm.interp.currentMethodCacheVersion() {
		vm.interp.recordBytecodeMemberMethodCacheMiss()
		return nil, false
	}
	bound, ok := bindMemberMethodTemplate(receiver, entry.methodTemplate)
	if !ok {
		vm.interp.recordBytecodeMemberMethodCacheMiss()
		return nil, false
	}
	vm.interp.recordBytecodeMemberMethodCacheHit()
	return bound, true
}

func (vm *bytecodeVM) storeCachedMemberMethod(program *bytecodeProgram, ip int, memberName string, preferMethods bool, receiver runtime.Value, resolved runtime.Value) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil {
		return
	}
	key, ok := vm.memberMethodCacheKey(program, ip, memberName, preferMethods, receiver)
	if !ok {
		return
	}
	template, ok := extractMemberMethodTemplate(resolved)
	if !ok {
		return
	}
	if vm.memberMethodCache == nil {
		vm.memberMethodCache = make(map[bytecodeMemberMethodCacheKey]bytecodeMemberMethodCacheEntry, 16)
	}
	vm.memberMethodCache[key] = bytecodeMemberMethodCacheEntry{
		globalRevision:     vm.interp.global.Revision(),
		methodCacheVersion: vm.interp.currentMethodCacheVersion(),
		methodTemplate:     template,
	}
}
