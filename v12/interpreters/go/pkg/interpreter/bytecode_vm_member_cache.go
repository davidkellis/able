package interpreter

import (
	"reflect"

	"able/interpreter-go/pkg/runtime"
)

type bytecodeMemberReceiverKind uint8

const (
	bytecodeMemberReceiverUnknown bytecodeMemberReceiverKind = iota
	bytecodeMemberReceiverArray
	bytecodeMemberReceiverBool
	bytecodeMemberReceiverChar
	bytecodeMemberReceiverFloat
	bytecodeMemberReceiverInteger
	bytecodeMemberReceiverNil
	bytecodeMemberReceiverString
	bytecodeMemberReceiverStruct
	bytecodeMemberReceiverInterface
)

type bytecodeMemberReceiverIdentity struct {
	kind                    bytecodeMemberReceiverKind
	structDef               *runtime.StructDefinitionValue
	primitiveKey            bytecodeMemberPrimitiveCacheKey
	interfaceDef            *runtime.InterfaceDefinitionValue
	interfaceArgs           typeExpressionSliceKey
	interfaceUnderlyingKind bytecodeMemberReceiverKind
}

type bytecodeMemberMethodCacheKey struct {
	program                 *bytecodeProgram
	ip                      int
	env                     *runtime.Environment
	member                  string
	preferMethods           bool
	implContextActive       bool
	implContextStateID      uint64
	implContextRevision     uint64
	receiverKind            bytecodeMemberReceiverKind
	structDef               *runtime.StructDefinitionValue
	primitiveKey            bytecodeMemberPrimitiveCacheKey
	interfaceDef            *runtime.InterfaceDefinitionValue
	interfaceArgs           typeExpressionSliceKey
	interfaceUnderlyingKind bytecodeMemberReceiverKind
}

type bytecodeMemberMethodCacheEntry struct {
	methodCacheVersion            uint64
	scopeStateID                  uint64
	bindingShapeVersion           uint64
	memberNameShapeVersion        uint64
	receiverTypeShapeVersion      uint64
	receiverAliasTypeShapeVersion uint64
	implContextActive             bool
	implContextStateID            uint64
	implContextRevision           uint64
	memberOwner                   *runtime.Environment
	memberOwnerVersion            uint64
	receiverTypeOwner             *runtime.Environment
	receiverAliasTypeOwner        *runtime.Environment
	methodTemplate                runtime.Value
	fastPath                      bytecodeMemberMethodFastPathKind
	dispatch                      bytecodeMemberMethodDispatchKind
	inlineFn                      *runtime.FunctionValue
}

type bytecodeInlineMemberMethodCacheEntry struct {
	valid                         bool
	program                       *bytecodeProgram
	ip                            int
	env                           *runtime.Environment
	member                        string
	preferMethods                 bool
	receiverKind                  bytecodeMemberReceiverKind
	structDef                     *runtime.StructDefinitionValue
	primitiveKey                  bytecodeMemberPrimitiveCacheKey
	interfaceDef                  *runtime.InterfaceDefinitionValue
	interfaceArgs                 typeExpressionSliceKey
	interfaceUnderlyingKind       bytecodeMemberReceiverKind
	methodCacheVersion            uint64
	scopeStateID                  uint64
	bindingShapeVersion           uint64
	memberNameShapeVersion        uint64
	receiverTypeShapeVersion      uint64
	receiverAliasTypeShapeVersion uint64
	implContextActive             bool
	implContextStateID            uint64
	implContextRevision           uint64
	memberOwner                   *runtime.Environment
	memberOwnerVersion            uint64
	receiverTypeOwner             *runtime.Environment
	receiverAliasTypeOwner        *runtime.Environment
	methodTemplate                runtime.Value
	fastPath                      bytecodeMemberMethodFastPathKind
	dispatch                      bytecodeMemberMethodDispatchKind
	inlineFn                      *runtime.FunctionValue
}

type bytecodeCachedMemberMethod struct {
	template runtime.Value
	fastPath bytecodeMemberMethodFastPathKind
	dispatch bytecodeMemberMethodDispatchKind
	inlineFn *runtime.FunctionValue
}

type bytecodeMemberMethodLexicalState struct {
	methodCacheVersion            uint64
	scopeStateID                  uint64
	bindingShapeVersion           uint64
	memberNameShapeVersion        uint64
	receiverTypeShapeVersion      uint64
	receiverAliasTypeShapeVersion uint64
	implContextActive             bool
	implContextStateID            uint64
	implContextRevision           uint64
	memberOwner                   *runtime.Environment
	memberOwnerVersion            uint64
	receiverTypeOwner             *runtime.Environment
	receiverAliasTypeOwner        *runtime.Environment
}

const (
	bytecodeMemberMethodNameRevisionCacheSize = 4
	bytecodeMemberMethodDirectEntries         = 16
	// Call-local environments are safe cache keys but can be one-shot.
	// Bound their retention; hot/direct entries still serve current sites.
	bytecodeMemberMethodCacheMaxEntries = 4096
)

type bytecodeMemberMethodNameRevisionCacheEntry struct {
	env           *runtime.Environment
	name          string
	shapeRevision uint64
	revision      uint64
	valid         bool
}

type bytecodeMemberMethodDispatchKind uint8

const (
	bytecodeMemberMethodDispatchGeneric bytecodeMemberMethodDispatchKind = iota
	bytecodeMemberMethodDispatchExactNative
	bytecodeMemberMethodDispatchInline
)

func bytecodeIsExactInjectedNativeTemplate(template runtime.Value) bool {
	switch fn := template.(type) {
	case runtime.NativeFunctionValue:
		return true
	case *runtime.NativeFunctionValue:
		return fn != nil
	case runtime.NativeBoundMethodValue:
		return true
	case *runtime.NativeBoundMethodValue:
		return fn != nil
	default:
		return false
	}
}

func bytecodeMemberMethodDispatchForTemplate(template runtime.Value) (bytecodeMemberMethodDispatchKind, *runtime.FunctionValue) {
	if bytecodeIsExactInjectedNativeTemplate(template) {
		return bytecodeMemberMethodDispatchExactNative, nil
	}
	if fn, ok := bytecodeResolvedMemberFastPathFunction(template); ok && fn != nil {
		return bytecodeMemberMethodDispatchInline, fn
	}
	return bytecodeMemberMethodDispatchGeneric, nil
}

func (vm *bytecodeVM) canUseMemberMethodCache(memberName string, preferMethods bool) bool {
	if vm == nil || vm.interp == nil || vm.interp.global == nil || vm.env == nil {
		return false
	}
	if !preferMethods || memberName == "" {
		return false
	}
	if vm.env == vm.interp.global {
		return true
	}
	return true
}

func (vm *bytecodeVM) hasImplMethodContextRuntimeData() bool {
	if vm == nil {
		return false
	}
	ctx, ok := vm.runtimeData().(*implMethodContext)
	return ok && ctx != nil
}

func (vm *bytecodeVM) memberMethodImplContextState() (bool, uint64, uint64, bool) {
	if vm == nil || vm.env == nil {
		return false, 0, 0, false
	}
	ctx, ok := vm.runtimeData().(*implMethodContext)
	if !ok || ctx == nil {
		return false, 0, 0, true
	}
	return true, vm.env.RuntimeDataStateID(), vm.env.RuntimeDataRevision(), true
}

func bytecodeMemberReceiverIdentityForValue(receiver runtime.Value) (bytecodeMemberReceiverIdentity, bool) {
	switch v := receiver.(type) {
	case *runtime.ArrayValue:
		if v == nil {
			return bytecodeMemberReceiverIdentity{}, false
		}
		return bytecodeMemberReceiverIdentity{kind: bytecodeMemberReceiverArray}, true
	case runtime.BoolValue:
		return bytecodeMemberReceiverIdentity{
			kind:         bytecodeMemberReceiverBool,
			primitiveKey: primitiveMethodCacheKeyBool,
		}, true
	case *runtime.BoolValue:
		if v == nil {
			return bytecodeMemberReceiverIdentity{}, false
		}
		return bytecodeMemberReceiverIdentity{
			kind:         bytecodeMemberReceiverBool,
			primitiveKey: primitiveMethodCacheKeyBool,
		}, true
	case runtime.CharValue:
		return bytecodeMemberReceiverIdentity{
			kind:         bytecodeMemberReceiverChar,
			primitiveKey: primitiveMethodCacheKeyChar,
		}, true
	case *runtime.CharValue:
		if v == nil {
			return bytecodeMemberReceiverIdentity{}, false
		}
		return bytecodeMemberReceiverIdentity{
			kind:         bytecodeMemberReceiverChar,
			primitiveKey: primitiveMethodCacheKeyChar,
		}, true
	case runtime.FloatValue:
		key, ok := primitiveMethodCacheKeyForFloatSuffix(v.TypeSuffix)
		if !ok {
			return bytecodeMemberReceiverIdentity{}, false
		}
		return bytecodeMemberReceiverIdentity{
			kind:         bytecodeMemberReceiverFloat,
			primitiveKey: key,
		}, true
	case *runtime.FloatValue:
		if v == nil {
			return bytecodeMemberReceiverIdentity{}, false
		}
		key, ok := primitiveMethodCacheKeyForFloatSuffix(v.TypeSuffix)
		if !ok {
			return bytecodeMemberReceiverIdentity{}, false
		}
		return bytecodeMemberReceiverIdentity{
			kind:         bytecodeMemberReceiverFloat,
			primitiveKey: key,
		}, true
	case runtime.IntegerValue:
		key, ok := primitiveMethodCacheKeyForIntegerSuffix(v.TypeSuffix)
		if !ok {
			return bytecodeMemberReceiverIdentity{}, false
		}
		return bytecodeMemberReceiverIdentity{
			kind:         bytecodeMemberReceiverInteger,
			primitiveKey: key,
		}, true
	case *runtime.IntegerValue:
		if v == nil {
			return bytecodeMemberReceiverIdentity{}, false
		}
		key, ok := primitiveMethodCacheKeyForIntegerSuffix(v.TypeSuffix)
		if !ok {
			return bytecodeMemberReceiverIdentity{}, false
		}
		return bytecodeMemberReceiverIdentity{
			kind:         bytecodeMemberReceiverInteger,
			primitiveKey: key,
		}, true
	case runtime.NilValue:
		return bytecodeMemberReceiverIdentity{
			kind:         bytecodeMemberReceiverNil,
			primitiveKey: primitiveMethodCacheKeyNil,
		}, true
	case *runtime.NilValue:
		if v == nil {
			return bytecodeMemberReceiverIdentity{}, false
		}
		return bytecodeMemberReceiverIdentity{
			kind:         bytecodeMemberReceiverNil,
			primitiveKey: primitiveMethodCacheKeyNil,
		}, true
	case runtime.StringValue:
		return bytecodeMemberReceiverIdentity{kind: bytecodeMemberReceiverString}, true
	case *runtime.StringValue:
		if v == nil {
			return bytecodeMemberReceiverIdentity{}, false
		}
		return bytecodeMemberReceiverIdentity{kind: bytecodeMemberReceiverString}, true
	case *runtime.StructInstanceValue:
		if v == nil || v.Definition == nil {
			return bytecodeMemberReceiverIdentity{}, false
		}
		return bytecodeMemberReceiverIdentity{
			kind:      bytecodeMemberReceiverStruct,
			structDef: v.Definition,
		}, true
	case *runtime.InterfaceValue:
		if v == nil || v.Interface == nil {
			return bytecodeMemberReceiverIdentity{}, false
		}
		underlying, ok := bytecodeMemberReceiverIdentityForValue(v.Underlying)
		if !ok {
			return bytecodeMemberReceiverIdentity{}, false
		}
		return bytecodeMemberReceiverIdentity{
			kind:                    bytecodeMemberReceiverInterface,
			structDef:               underlying.structDef,
			primitiveKey:            underlying.primitiveKey,
			interfaceDef:            v.Interface,
			interfaceArgs:           makeTypeExpressionSliceKey(v.InterfaceArgs),
			interfaceUnderlyingKind: underlying.kind,
		}, true
	default:
		return bytecodeMemberReceiverIdentity{}, false
	}
}

func (vm *bytecodeVM) memberMethodCacheIdentity(memberName string, preferMethods bool, receiver runtime.Value) (bytecodeMemberReceiverIdentity, bool) {
	if _, ok := receiver.(*runtime.InterfaceValue); ok {
		if !vm.canUseInterfaceMemberMethodCache(memberName, preferMethods, receiver) {
			return bytecodeMemberReceiverIdentity{}, false
		}
		return bytecodeMemberReceiverIdentityForValue(receiver)
	}
	if vm == nil || vm.interp == nil || vm.interp.global == nil || vm.env == nil {
		return bytecodeMemberReceiverIdentity{}, false
	}
	if !preferMethods || memberName == "" {
		return bytecodeMemberReceiverIdentity{}, false
	}
	return bytecodeMemberReceiverIdentityForValue(receiver)
}

func (vm *bytecodeVM) canUseMemberMethodCacheForReceiver(memberName string, preferMethods bool, receiver runtime.Value) bool {
	if _, ok := receiver.(*runtime.InterfaceValue); ok {
		return vm.canUseInterfaceMemberMethodCache(memberName, preferMethods, receiver)
	}
	return vm.canUseMemberMethodCache(memberName, preferMethods)
}

func (vm *bytecodeVM) canUseInterfaceMemberMethodCache(memberName string, preferMethods bool, receiver runtime.Value) bool {
	if vm == nil || vm.interp == nil || vm.interp.global == nil || vm.env == nil {
		return false
	}
	if !preferMethods || memberName == "" || vm.hasImplMethodContextRuntimeData() {
		return false
	}
	iface, ok := receiver.(*runtime.InterfaceValue)
	if !ok || iface == nil {
		return false
	}
	method, ok := interfaceValueLookupMethod(iface, memberName)
	return ok && method != nil && !interfaceValueMethodIsBound(method)
}

func (vm *bytecodeVM) memberMethodLexicalStateHeader() bytecodeMemberMethodLexicalState {
	if vm == nil {
		return bytecodeMemberMethodLexicalState{}
	}
	state := bytecodeMemberMethodLexicalState{
		methodCacheVersion: vm.bytecodeMethodCacheVersion(),
	}
	if vm.env == nil {
		return state
	}
	env := vm.env
	if active, stateID, revision, ok := vm.memberMethodImplContextState(); ok {
		state.implContextActive = active
		state.implContextStateID = stateID
		state.implContextRevision = revision
	}
	state.scopeStateID = env.BindingShapeStateID()
	state.bindingShapeVersion = env.BindingShapeRevision()
	return state
}

func (vm *bytecodeVM) memberMethodLexicalState(memberName string, receiver runtime.Value, captureOwners bool) bytecodeMemberMethodLexicalState {
	state := vm.memberMethodLexicalStateHeader()
	if vm == nil || vm.env == nil {
		return state
	}
	env := vm.env
	state.memberNameShapeVersion = vm.cachedMemberMethodBindingNameRevisionWithShape(env, memberName, state.bindingShapeVersion)
	if typeName, ok := methodReceiverNominalTypeName(receiver); ok && vm.interp != nil {
		canonicalName, aliasName := vm.interp.canonicalTypeNamePair(typeName)
		if canonicalName != "" {
			state.receiverTypeShapeVersion = vm.cachedMemberMethodBindingNameRevisionWithShape(env, canonicalName, state.bindingShapeVersion)
			if captureOwners {
				_, state.receiverTypeOwner, _, _ = env.LookupWithOwnerAndRevisionHint(canonicalName, vm.bytecodeSingleThread())
			}
		}
		if aliasName != "" {
			state.receiverAliasTypeShapeVersion = vm.cachedMemberMethodBindingNameRevisionWithShape(env, aliasName, state.bindingShapeVersion)
			if captureOwners {
				_, state.receiverAliasTypeOwner, _, _ = env.LookupWithOwnerAndRevisionHint(aliasName, vm.bytecodeSingleThread())
			}
		}
	}
	if captureOwners && memberName != "" {
		_, owner, ownerVersion, found := env.LookupWithOwnerAndRevisionHint(memberName, vm.bytecodeSingleThread())
		if found {
			state.memberOwner = owner
			state.memberOwnerVersion = ownerVersion
		}
	}
	return state
}

func (vm *bytecodeVM) memberMethodLexicalOwnerValid(owner *runtime.Environment, ownerVersion uint64) bool {
	if owner == nil {
		return true
	}
	if vm == nil {
		return false
	}
	return owner.RevisionWithHint(vm.bytecodeSingleThread()) == ownerVersion
}

func (vm *bytecodeVM) cachedMemberMethodBindingNameRevision(env *runtime.Environment, name string) uint64 {
	if env == nil || name == "" {
		return 0
	}
	shapeRevision := env.BindingShapeRevision()
	return vm.cachedMemberMethodBindingNameRevisionWithShape(env, name, shapeRevision)
}

func (vm *bytecodeVM) cachedMemberMethodBindingNameRevisionWithShape(env *runtime.Environment, name string, shapeRevision uint64) uint64 {
	if env == nil || name == "" {
		return 0
	}
	for idx := range vm.memberMethodNameRevisions {
		entry := &vm.memberMethodNameRevisions[idx]
		if entry.valid && entry.env == env && entry.name == name && entry.shapeRevision == shapeRevision {
			return entry.revision
		}
	}
	revision := env.BindingNameRevision(name)
	slot := vm.memberMethodNameRevisionNext % bytecodeMemberMethodNameRevisionCacheSize
	vm.memberMethodNameRevisionNext++
	vm.memberMethodNameRevisions[slot] = bytecodeMemberMethodNameRevisionCacheEntry{
		env:           env,
		name:          name,
		shapeRevision: shapeRevision,
		revision:      revision,
		valid:         true,
	}
	return revision
}

func bytecodeMemberMethodDirectIndex(ip int) int {
	return int(uint(ip) & uint(bytecodeMemberMethodDirectEntries-1))
}

func bytecodeCachedMemberMethodFromInline(entry bytecodeInlineMemberMethodCacheEntry) bytecodeCachedMemberMethod {
	return bytecodeCachedMemberMethod{
		template: entry.methodTemplate,
		fastPath: entry.fastPath,
		dispatch: entry.dispatch,
		inlineFn: entry.inlineFn,
	}
}

func (vm *bytecodeVM) inlineMemberMethodEntryMatches(
	entry *bytecodeInlineMemberMethodCacheEntry,
	program *bytecodeProgram,
	ip int,
	env *runtime.Environment,
	memberName string,
	preferMethods bool,
	identity bytecodeMemberReceiverIdentity,
	lexicalState bytecodeMemberMethodLexicalState,
) bool {
	return entry != nil &&
		entry.valid &&
		entry.program == program &&
		entry.ip == ip &&
		entry.member == memberName &&
		entry.preferMethods == preferMethods &&
		entry.implContextActive == lexicalState.implContextActive &&
		entry.implContextStateID == lexicalState.implContextStateID &&
		entry.implContextRevision == lexicalState.implContextRevision &&
		entry.receiverKind == identity.kind &&
		entry.structDef == identity.structDef &&
		entry.primitiveKey == identity.primitiveKey &&
		entry.interfaceDef == identity.interfaceDef &&
		entry.interfaceArgs == identity.interfaceArgs &&
		entry.interfaceUnderlyingKind == identity.interfaceUnderlyingKind &&
		entry.methodCacheVersion == lexicalState.methodCacheVersion &&
		entry.scopeStateID == lexicalState.scopeStateID &&
		entry.memberNameShapeVersion == lexicalState.memberNameShapeVersion &&
		entry.receiverTypeShapeVersion == lexicalState.receiverTypeShapeVersion &&
		entry.receiverAliasTypeShapeVersion == lexicalState.receiverAliasTypeShapeVersion &&
		entry.memberOwner == lexicalState.memberOwner &&
		entry.memberOwnerVersion == lexicalState.memberOwnerVersion &&
		entry.receiverTypeOwner == lexicalState.receiverTypeOwner &&
		entry.receiverAliasTypeOwner == lexicalState.receiverAliasTypeOwner &&
		vm.memberMethodLexicalOwnerValid(entry.memberOwner, entry.memberOwnerVersion)
}

func (vm *bytecodeVM) inlineMemberMethodEntryStableShapeMatches(
	entry *bytecodeInlineMemberMethodCacheEntry,
	program *bytecodeProgram,
	ip int,
	env *runtime.Environment,
	memberName string,
	preferMethods bool,
	identity bytecodeMemberReceiverIdentity,
	lexicalState bytecodeMemberMethodLexicalState,
) bool {
	return entry != nil &&
		entry.valid &&
		entry.program == program &&
		entry.ip == ip &&
		entry.env == env &&
		entry.member == memberName &&
		entry.preferMethods == preferMethods &&
		entry.implContextActive == lexicalState.implContextActive &&
		entry.implContextStateID == lexicalState.implContextStateID &&
		entry.implContextRevision == lexicalState.implContextRevision &&
		entry.receiverKind == identity.kind &&
		entry.structDef == identity.structDef &&
		entry.primitiveKey == identity.primitiveKey &&
		entry.interfaceDef == identity.interfaceDef &&
		entry.interfaceArgs == identity.interfaceArgs &&
		entry.interfaceUnderlyingKind == identity.interfaceUnderlyingKind &&
		entry.methodCacheVersion == lexicalState.methodCacheVersion &&
		entry.scopeStateID == lexicalState.scopeStateID &&
		entry.bindingShapeVersion == lexicalState.bindingShapeVersion &&
		vm.memberMethodLexicalOwnerValid(entry.memberOwner, entry.memberOwnerVersion)
}

func (vm *bytecodeVM) finishInlineCachedMemberMethodEntry(
	entry bytecodeInlineMemberMethodCacheEntry,
	receiver runtime.Value,
	memberName string,
	identity bytecodeMemberReceiverIdentity,
) (bytecodeCachedMemberMethod, bool) {
	if identity.kind == bytecodeMemberReceiverInterface &&
		!bytecodeCachedMemberMethodValidForReceiver(receiver, memberName, entry.methodTemplate, identity) {
		return bytecodeCachedMemberMethod{}, false
	}
	if entry.fastPath != bytecodeMemberMethodFastPathNone {
		return bytecodeCachedMemberMethodFromInline(entry), true
	}
	if _, ok := bindMemberMethodTemplate(receiver, entry.methodTemplate); ok {
		return bytecodeCachedMemberMethodFromInline(entry), true
	}
	return bytecodeCachedMemberMethod{}, false
}

func (vm *bytecodeVM) storeMemberMethodDirect(entry bytecodeInlineMemberMethodCacheEntry) {
	if vm == nil || !entry.valid {
		return
	}
	vm.memberMethodDirect[bytecodeMemberMethodDirectIndex(entry.ip)] = entry
}

func (vm *bytecodeVM) memberMethodCacheKey(program *bytecodeProgram, ip int, memberName string, preferMethods bool, receiver runtime.Value) (bytecodeMemberMethodCacheKey, bool) {
	identity, ok := vm.memberMethodCacheIdentity(memberName, preferMethods, receiver)
	if !ok {
		return bytecodeMemberMethodCacheKey{}, false
	}
	implContextActive, implContextStateID, implContextRevision, ok := vm.memberMethodImplContextState()
	if !ok {
		return bytecodeMemberMethodCacheKey{}, false
	}
	return bytecodeMemberMethodCacheKey{
		program:                 program,
		ip:                      ip,
		env:                     vm.env,
		member:                  memberName,
		preferMethods:           preferMethods,
		implContextActive:       implContextActive,
		implContextStateID:      implContextStateID,
		implContextRevision:     implContextRevision,
		receiverKind:            identity.kind,
		structDef:               identity.structDef,
		primitiveKey:            identity.primitiveKey,
		interfaceDef:            identity.interfaceDef,
		interfaceArgs:           identity.interfaceArgs,
		interfaceUnderlyingKind: identity.interfaceUnderlyingKind,
	}, true
}

func extractMemberMethodTemplate(resolved runtime.Value) (runtime.Value, bool) {
	switch method := resolved.(type) {
	case runtime.NativeFunctionValue:
		return method, true
	case *runtime.NativeFunctionValue:
		if method == nil {
			return nil, false
		}
		return method, true
	case *runtime.FunctionValue, *runtime.FunctionOverloadValue:
		return method, true
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

func bytecodeCachedMemberMethodValidForReceiver(receiver runtime.Value, memberName string, template runtime.Value, identity bytecodeMemberReceiverIdentity) bool {
	if identity.kind != bytecodeMemberReceiverInterface {
		return true
	}
	iface, ok := receiver.(*runtime.InterfaceValue)
	if !ok || iface == nil {
		return false
	}
	method, ok := interfaceValueLookupMethod(iface, memberName)
	if !ok || method == nil || interfaceValueMethodIsBound(method) {
		return false
	}
	return bytecodeMemberMethodTemplateMatches(method, template)
}

func bytecodeMemberMethodTemplateMatches(current runtime.Value, cached runtime.Value) bool {
	if template, ok := extractMemberMethodTemplate(current); ok {
		current = template
	}
	if template, ok := extractMemberMethodTemplate(cached); ok {
		cached = template
	}
	switch cur := current.(type) {
	case *runtime.FunctionValue:
		cachedFn, ok := cached.(*runtime.FunctionValue)
		return ok && cur == cachedFn
	case *runtime.FunctionOverloadValue:
		cachedOverload, ok := cached.(*runtime.FunctionOverloadValue)
		return ok && cur == cachedOverload
	case runtime.NativeFunctionValue:
		return false
	case *runtime.NativeFunctionValue:
		cachedNative, ok := cached.(*runtime.NativeFunctionValue)
		return ok && cur == cachedNative
	default:
		currentType := reflect.TypeOf(current)
		if currentType == nil || currentType != reflect.TypeOf(cached) || !currentType.Comparable() {
			return false
		}
		return current == cached
	}
}

func (vm *bytecodeVM) lookupCachedMemberMethod(program *bytecodeProgram, ip int, memberName string, preferMethods bool, receiver runtime.Value) (runtime.Value, bool) {
	cached, ok := vm.lookupCachedMemberMethodEntry(program, ip, memberName, preferMethods, receiver)
	if !ok {
		return nil, false
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
	identity, ok := vm.memberMethodCacheIdentity(memberName, preferMethods, receiver)
	if !ok {
		return bytecodeCachedMemberMethod{}, false
	}
	env := vm.env
	lexicalHeader := vm.memberMethodLexicalStateHeader()
	if hot := vm.memberMethodHot; hot.valid &&
		vm.inlineMemberMethodEntryStableShapeMatches(&hot, program, ip, env, memberName, preferMethods, identity, lexicalHeader) {
		if cached, ok := vm.finishInlineCachedMemberMethodEntry(hot, receiver, memberName, identity); ok {
			vm.interp.recordBytecodeMemberMethodCacheHit()
			return cached, true
		}
		vm.interp.recordBytecodeMemberMethodCacheMiss()
		return bytecodeCachedMemberMethod{}, false
	}
	direct := &vm.memberMethodDirect[bytecodeMemberMethodDirectIndex(ip)]
	if vm.inlineMemberMethodEntryStableShapeMatches(direct, program, ip, env, memberName, preferMethods, identity, lexicalHeader) {
		if cached, ok := vm.finishInlineCachedMemberMethodEntry(*direct, receiver, memberName, identity); ok {
			vm.memberMethodHot = *direct
			vm.interp.recordBytecodeMemberMethodCacheHit()
			return cached, true
		}
		vm.interp.recordBytecodeMemberMethodCacheMiss()
		return bytecodeCachedMemberMethod{}, false
	}
	lexicalState := vm.memberMethodLexicalState(memberName, receiver, true)
	if hot := vm.memberMethodHot; hot.valid &&
		vm.inlineMemberMethodEntryMatches(&hot, program, ip, env, memberName, preferMethods, identity, lexicalState) {
		hot.env = env
		hot.bindingShapeVersion = lexicalState.bindingShapeVersion
		if cached, ok := vm.finishInlineCachedMemberMethodEntry(hot, receiver, memberName, identity); ok {
			vm.memberMethodHot = hot
			vm.storeMemberMethodDirect(vm.memberMethodHot)
			vm.interp.recordBytecodeMemberMethodCacheHit()
			return cached, true
		}
		vm.interp.recordBytecodeMemberMethodCacheMiss()
		return bytecodeCachedMemberMethod{}, false
	}
	if vm.inlineMemberMethodEntryMatches(direct, program, ip, env, memberName, preferMethods, identity, lexicalState) {
		direct.env = env
		direct.bindingShapeVersion = lexicalState.bindingShapeVersion
		if cached, ok := vm.finishInlineCachedMemberMethodEntry(*direct, receiver, memberName, identity); ok {
			vm.memberMethodHot = *direct
			vm.interp.recordBytecodeMemberMethodCacheHit()
			return cached, true
		}
		vm.interp.recordBytecodeMemberMethodCacheMiss()
		return bytecodeCachedMemberMethod{}, false
	}
	if vm.memberMethodCache == nil {
		vm.interp.recordBytecodeMemberMethodCacheMiss()
		return bytecodeCachedMemberMethod{}, false
	}
	key := bytecodeMemberMethodCacheKey{
		program:                 program,
		ip:                      ip,
		env:                     env,
		member:                  memberName,
		preferMethods:           preferMethods,
		implContextActive:       lexicalState.implContextActive,
		implContextStateID:      lexicalState.implContextStateID,
		implContextRevision:     lexicalState.implContextRevision,
		receiverKind:            identity.kind,
		structDef:               identity.structDef,
		primitiveKey:            identity.primitiveKey,
		interfaceDef:            identity.interfaceDef,
		interfaceArgs:           identity.interfaceArgs,
		interfaceUnderlyingKind: identity.interfaceUnderlyingKind,
	}
	entry, ok := vm.memberMethodCache[key]
	if !ok {
		vm.interp.recordBytecodeMemberMethodCacheMiss()
		return bytecodeCachedMemberMethod{}, false
	}
	if entry.methodCacheVersion != lexicalState.methodCacheVersion {
		vm.interp.recordBytecodeMemberMethodCacheMiss()
		return bytecodeCachedMemberMethod{}, false
	}
	if entry.scopeStateID != lexicalState.scopeStateID ||
		entry.memberNameShapeVersion != lexicalState.memberNameShapeVersion ||
		entry.receiverTypeShapeVersion != lexicalState.receiverTypeShapeVersion ||
		entry.receiverAliasTypeShapeVersion != lexicalState.receiverAliasTypeShapeVersion ||
		entry.receiverTypeOwner != lexicalState.receiverTypeOwner ||
		entry.receiverAliasTypeOwner != lexicalState.receiverAliasTypeOwner ||
		entry.implContextActive != lexicalState.implContextActive ||
		entry.implContextStateID != lexicalState.implContextStateID ||
		entry.implContextRevision != lexicalState.implContextRevision {
		vm.interp.recordBytecodeMemberMethodCacheMiss()
		return bytecodeCachedMemberMethod{}, false
	}
	if !vm.memberMethodLexicalOwnerValid(entry.memberOwner, entry.memberOwnerVersion) {
		vm.interp.recordBytecodeMemberMethodCacheMiss()
		return bytecodeCachedMemberMethod{}, false
	}
	if identity.kind == bytecodeMemberReceiverInterface &&
		!bytecodeCachedMemberMethodValidForReceiver(receiver, memberName, entry.methodTemplate, identity) {
		vm.interp.recordBytecodeMemberMethodCacheMiss()
		return bytecodeCachedMemberMethod{}, false
	}
	entry.bindingShapeVersion = lexicalState.bindingShapeVersion
	vm.memberMethodCache[key] = entry
	vm.memberMethodHot = bytecodeInlineMemberMethodCacheEntry{
		valid:                         true,
		program:                       program,
		ip:                            ip,
		env:                           env,
		member:                        memberName,
		preferMethods:                 preferMethods,
		implContextActive:             entry.implContextActive,
		implContextStateID:            entry.implContextStateID,
		implContextRevision:           entry.implContextRevision,
		receiverKind:                  identity.kind,
		structDef:                     identity.structDef,
		primitiveKey:                  identity.primitiveKey,
		interfaceDef:                  identity.interfaceDef,
		interfaceArgs:                 identity.interfaceArgs,
		interfaceUnderlyingKind:       identity.interfaceUnderlyingKind,
		methodCacheVersion:            entry.methodCacheVersion,
		scopeStateID:                  entry.scopeStateID,
		bindingShapeVersion:           entry.bindingShapeVersion,
		memberNameShapeVersion:        entry.memberNameShapeVersion,
		receiverTypeShapeVersion:      entry.receiverTypeShapeVersion,
		receiverAliasTypeShapeVersion: entry.receiverAliasTypeShapeVersion,
		memberOwner:                   entry.memberOwner,
		memberOwnerVersion:            entry.memberOwnerVersion,
		receiverTypeOwner:             entry.receiverTypeOwner,
		receiverAliasTypeOwner:        entry.receiverAliasTypeOwner,
		methodTemplate:                entry.methodTemplate,
		fastPath:                      entry.fastPath,
		dispatch:                      entry.dispatch,
		inlineFn:                      entry.inlineFn,
	}
	vm.storeMemberMethodDirect(vm.memberMethodHot)
	if entry.fastPath != bytecodeMemberMethodFastPathNone {
		vm.interp.recordBytecodeMemberMethodCacheHit()
		return bytecodeCachedMemberMethod{
			template: entry.methodTemplate,
			fastPath: entry.fastPath,
			dispatch: entry.dispatch,
			inlineFn: entry.inlineFn,
		}, true
	}
	if _, ok := bindMemberMethodTemplate(receiver, entry.methodTemplate); !ok {
		vm.interp.recordBytecodeMemberMethodCacheMiss()
		return bytecodeCachedMemberMethod{}, false
	}
	vm.interp.recordBytecodeMemberMethodCacheHit()
	return bytecodeCachedMemberMethod{
		template: entry.methodTemplate,
		fastPath: entry.fastPath,
		dispatch: entry.dispatch,
		inlineFn: entry.inlineFn,
	}, true
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
	lexicalState := vm.memberMethodLexicalState(memberName, receiver, true)
	fastPath := vm.memberMethodFastPathFor(key, template)
	if fn, ok := template.(*runtime.FunctionValue); ok {
		fastPath = vm.memberMethodFastPathForFunction(key, fn)
	}
	dispatch, inlineFn := bytecodeMemberMethodDispatchForTemplate(template)
	entry := bytecodeMemberMethodCacheEntry{
		methodCacheVersion:            lexicalState.methodCacheVersion,
		scopeStateID:                  lexicalState.scopeStateID,
		bindingShapeVersion:           lexicalState.bindingShapeVersion,
		memberNameShapeVersion:        lexicalState.memberNameShapeVersion,
		receiverTypeShapeVersion:      lexicalState.receiverTypeShapeVersion,
		receiverAliasTypeShapeVersion: lexicalState.receiverAliasTypeShapeVersion,
		implContextActive:             lexicalState.implContextActive,
		implContextStateID:            lexicalState.implContextStateID,
		implContextRevision:           lexicalState.implContextRevision,
		memberOwner:                   lexicalState.memberOwner,
		memberOwnerVersion:            lexicalState.memberOwnerVersion,
		receiverTypeOwner:             lexicalState.receiverTypeOwner,
		receiverAliasTypeOwner:        lexicalState.receiverAliasTypeOwner,
		methodTemplate:                template,
		fastPath:                      fastPath,
		dispatch:                      dispatch,
		inlineFn:                      inlineFn,
	}
	if len(vm.memberMethodCache) < bytecodeMemberMethodCacheMaxEntries {
		vm.memberMethodCache[key] = entry
	} else if _, exists := vm.memberMethodCache[key]; exists {
		vm.memberMethodCache[key] = entry
	}
	vm.memberMethodHot = bytecodeInlineMemberMethodCacheEntry{
		valid:                         true,
		program:                       program,
		ip:                            ip,
		env:                           key.env,
		member:                        memberName,
		preferMethods:                 preferMethods,
		implContextActive:             entry.implContextActive,
		implContextStateID:            entry.implContextStateID,
		implContextRevision:           entry.implContextRevision,
		receiverKind:                  key.receiverKind,
		structDef:                     key.structDef,
		primitiveKey:                  key.primitiveKey,
		interfaceDef:                  key.interfaceDef,
		interfaceArgs:                 key.interfaceArgs,
		interfaceUnderlyingKind:       key.interfaceUnderlyingKind,
		methodCacheVersion:            entry.methodCacheVersion,
		scopeStateID:                  entry.scopeStateID,
		bindingShapeVersion:           entry.bindingShapeVersion,
		memberNameShapeVersion:        entry.memberNameShapeVersion,
		receiverTypeShapeVersion:      entry.receiverTypeShapeVersion,
		receiverAliasTypeShapeVersion: entry.receiverAliasTypeShapeVersion,
		memberOwner:                   entry.memberOwner,
		memberOwnerVersion:            entry.memberOwnerVersion,
		receiverTypeOwner:             entry.receiverTypeOwner,
		receiverAliasTypeOwner:        entry.receiverAliasTypeOwner,
		methodTemplate:                entry.methodTemplate,
		fastPath:                      entry.fastPath,
		dispatch:                      entry.dispatch,
		inlineFn:                      entry.inlineFn,
	}
	vm.storeMemberMethodDirect(vm.memberMethodHot)
	return entry, true
}
