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
	fastPath           bytecodeMemberMethodFastPathKind
}

type bytecodeInlineMemberMethodCacheEntry struct {
	valid              bool
	program            *bytecodeProgram
	ip                 int
	member             string
	preferMethods      bool
	receiverKind       bytecodeMemberReceiverKind
	structDef          *runtime.StructDefinitionValue
	globalRevision     uint64
	methodCacheVersion uint64
	methodTemplate     runtime.Value
	fastPath           bytecodeMemberMethodFastPathKind
}

type bytecodeCachedMemberMethod struct {
	callable runtime.Value
	template runtime.Value
	fastPath bytecodeMemberMethodFastPathKind
}

func (vm *bytecodeVM) canUseMemberMethodCache(memberName string, preferMethods bool) bool {
	if vm == nil || vm.interp == nil || vm.interp.global == nil || vm.env != vm.interp.global {
		return false
	}
	return preferMethods && memberName != ""
}

func (vm *bytecodeVM) memberMethodCacheIdentity(memberName string, preferMethods bool, receiver runtime.Value) (bytecodeMemberReceiverKind, *runtime.StructDefinitionValue, bool) {
	if !vm.canUseMemberMethodCache(memberName, preferMethods) {
		return bytecodeMemberReceiverUnknown, nil, false
	}
	switch v := receiver.(type) {
	case *runtime.ArrayValue:
		if v == nil {
			return bytecodeMemberReceiverUnknown, nil, false
		}
		return bytecodeMemberReceiverArray, nil, true
	case runtime.StringValue:
		return bytecodeMemberReceiverString, nil, true
	case *runtime.StringValue:
		if v == nil {
			return bytecodeMemberReceiverUnknown, nil, false
		}
		return bytecodeMemberReceiverString, nil, true
	case *runtime.StructInstanceValue:
		if v == nil || v.Definition == nil {
			return bytecodeMemberReceiverUnknown, nil, false
		}
		return bytecodeMemberReceiverStruct, v.Definition, true
	default:
		return bytecodeMemberReceiverUnknown, nil, false
	}
}

func (vm *bytecodeVM) memberMethodCacheKey(program *bytecodeProgram, ip int, memberName string, preferMethods bool, receiver runtime.Value) (bytecodeMemberMethodCacheKey, bool) {
	receiverKind, structDef, ok := vm.memberMethodCacheIdentity(memberName, preferMethods, receiver)
	if !ok {
		return bytecodeMemberMethodCacheKey{}, false
	}
	return bytecodeMemberMethodCacheKey{
		program:       program,
		ip:            ip,
		member:        memberName,
		preferMethods: preferMethods,
		receiverKind:  receiverKind,
		structDef:     structDef,
	}, true
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
	cached, ok := vm.lookupCachedMemberMethodEntry(program, ip, memberName, preferMethods, receiver)
	if !ok {
		return nil, false
	}
	return cached.boundCallable(receiver)
}

func (cached bytecodeCachedMemberMethod) boundCallable(receiver runtime.Value) (runtime.Value, bool) {
	if cached.callable != nil {
		return cached.callable, true
	}
	if cached.template == nil {
		return nil, false
	}
	return bindMemberMethodTemplate(receiver, cached.template)
}

func (vm *bytecodeVM) lookupCachedMemberMethodEntry(program *bytecodeProgram, ip int, memberName string, preferMethods bool, receiver runtime.Value) (bytecodeCachedMemberMethod, bool) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil {
		return bytecodeCachedMemberMethod{}, false
	}
	receiverKind, structDef, ok := vm.memberMethodCacheIdentity(memberName, preferMethods, receiver)
	if !ok {
		return bytecodeCachedMemberMethod{}, false
	}
	if hot := vm.memberMethodHot; hot.valid &&
		hot.program == program &&
		hot.ip == ip &&
		hot.member == memberName &&
		hot.preferMethods == preferMethods &&
		hot.receiverKind == receiverKind &&
		hot.structDef == structDef &&
		hot.globalRevision == vm.bytecodeGlobalRevision() &&
		hot.methodCacheVersion == vm.bytecodeMethodCacheVersion() {
		if hot.fastPath != bytecodeMemberMethodFastPathNone {
			vm.interp.recordBytecodeMemberMethodCacheHit()
			return bytecodeCachedMemberMethod{template: hot.methodTemplate, fastPath: hot.fastPath}, true
		}
		if bound, ok := bindMemberMethodTemplate(receiver, hot.methodTemplate); ok {
			vm.interp.recordBytecodeMemberMethodCacheHit()
			return bytecodeCachedMemberMethod{callable: bound, template: hot.methodTemplate, fastPath: hot.fastPath}, true
		}
		vm.interp.recordBytecodeMemberMethodCacheMiss()
		return bytecodeCachedMemberMethod{}, false
	}
	if vm.memberMethodCache == nil {
		vm.interp.recordBytecodeMemberMethodCacheMiss()
		return bytecodeCachedMemberMethod{}, false
	}
	key := bytecodeMemberMethodCacheKey{
		program:       program,
		ip:            ip,
		member:        memberName,
		preferMethods: preferMethods,
		receiverKind:  receiverKind,
		structDef:     structDef,
	}
	entry, ok := vm.memberMethodCache[key]
	if !ok {
		vm.interp.recordBytecodeMemberMethodCacheMiss()
		return bytecodeCachedMemberMethod{}, false
	}
	if entry.globalRevision != vm.bytecodeGlobalRevision() {
		vm.interp.recordBytecodeMemberMethodCacheMiss()
		return bytecodeCachedMemberMethod{}, false
	}
	if entry.methodCacheVersion != vm.bytecodeMethodCacheVersion() {
		vm.interp.recordBytecodeMemberMethodCacheMiss()
		return bytecodeCachedMemberMethod{}, false
	}
	vm.memberMethodHot = bytecodeInlineMemberMethodCacheEntry{
		valid:              true,
		program:            program,
		ip:                 ip,
		member:             memberName,
		preferMethods:      preferMethods,
		receiverKind:       receiverKind,
		structDef:          structDef,
		globalRevision:     entry.globalRevision,
		methodCacheVersion: entry.methodCacheVersion,
		methodTemplate:     entry.methodTemplate,
		fastPath:           entry.fastPath,
	}
	if entry.fastPath != bytecodeMemberMethodFastPathNone {
		vm.interp.recordBytecodeMemberMethodCacheHit()
		return bytecodeCachedMemberMethod{template: entry.methodTemplate, fastPath: entry.fastPath}, true
	}
	bound, ok := bindMemberMethodTemplate(receiver, entry.methodTemplate)
	if !ok {
		vm.interp.recordBytecodeMemberMethodCacheMiss()
		return bytecodeCachedMemberMethod{}, false
	}
	vm.interp.recordBytecodeMemberMethodCacheHit()
	return bytecodeCachedMemberMethod{callable: bound, template: entry.methodTemplate, fastPath: entry.fastPath}, true
}

func (vm *bytecodeVM) storeCachedMemberMethod(program *bytecodeProgram, ip int, memberName string, preferMethods bool, receiver runtime.Value, resolved runtime.Value) (bytecodeMemberMethodCacheEntry, bool) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil {
		return bytecodeMemberMethodCacheEntry{}, false
	}
	key, ok := vm.memberMethodCacheKey(program, ip, memberName, preferMethods, receiver)
	if !ok {
		return bytecodeMemberMethodCacheEntry{}, false
	}
	template, ok := extractMemberMethodTemplate(resolved)
	if !ok {
		return bytecodeMemberMethodCacheEntry{}, false
	}
	if vm.memberMethodCache == nil {
		vm.memberMethodCache = make(map[bytecodeMemberMethodCacheKey]bytecodeMemberMethodCacheEntry, 16)
	}
	fastPath := vm.memberMethodFastPathFor(key, template)
	if fn, ok := template.(*runtime.FunctionValue); ok {
		fastPath = vm.memberMethodFastPathForFunction(key, fn)
	}
	entry := bytecodeMemberMethodCacheEntry{
		globalRevision:     vm.bytecodeGlobalRevision(),
		methodCacheVersion: vm.bytecodeMethodCacheVersion(),
		methodTemplate:     template,
		fastPath:           fastPath,
	}
	vm.memberMethodCache[key] = entry
	vm.memberMethodHot = bytecodeInlineMemberMethodCacheEntry{
		valid:              true,
		program:            program,
		ip:                 ip,
		member:             memberName,
		preferMethods:      preferMethods,
		receiverKind:       key.receiverKind,
		structDef:          key.structDef,
		globalRevision:     entry.globalRevision,
		methodCacheVersion: entry.methodCacheVersion,
		methodTemplate:     entry.methodTemplate,
		fastPath:           entry.fastPath,
	}
	return entry, true
}
