package interpreter

import (
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type bytecodeStaticMemberReceiverKind uint8

const (
	bytecodeStaticMemberReceiverUnknown bytecodeStaticMemberReceiverKind = iota
	bytecodeStaticMemberReceiverTypeRef
	bytecodeStaticMemberReceiverStructDefinition
	bytecodeStaticMemberReceiverInterfaceDefinition
	bytecodeStaticMemberReceiverPackage
	bytecodeStaticMemberReceiverImplementationNamespace
	bytecodeStaticMemberReceiverDynPackage
)

type bytecodeStaticMemberReceiverIdentity struct {
	kind          bytecodeStaticMemberReceiverKind
	typeName      string
	typeArgs      typeExpressionSliceKey
	structNode    *ast.StructDefinition
	interfaceNode *ast.InterfaceDefinition
	packageName   string
	implName      string
	implInterface string
	implTarget    typeExpressionSliceKey
}

type bytecodeStaticMemberCallCacheKey struct {
	program             *bytecodeProgram
	ip                  int
	env                 *runtime.Environment
	member              string
	argCount            int
	implContextActive   bool
	implContextStateID  uint64
	implContextRevision uint64
	receiver            bytecodeStaticMemberReceiverIdentity
}

type bytecodeStaticMemberCallCacheEntry struct {
	methodCacheVersion            uint64
	scopeStateID                  uint64
	memberNameShapeVersion        uint64
	receiverTypeShapeVersion      uint64
	receiverAliasTypeShapeVersion uint64
	memberOwner                   *runtime.Environment
	memberOwnerVersion            uint64
	callable                      runtime.Value
	dispatch                      bytecodeMemberMethodDispatchKind
	inlineFn                      *runtime.FunctionValue
}

type bytecodeInlineStaticMemberCallCacheEntry struct {
	valid                         bool
	key                           bytecodeStaticMemberCallCacheKey
	methodCacheVersion            uint64
	scopeStateID                  uint64
	memberNameShapeVersion        uint64
	receiverTypeShapeVersion      uint64
	receiverAliasTypeShapeVersion uint64
	memberOwner                   *runtime.Environment
	memberOwnerVersion            uint64
	callable                      runtime.Value
	dispatch                      bytecodeMemberMethodDispatchKind
	inlineFn                      *runtime.FunctionValue
}

type bytecodeCachedStaticMemberCall struct {
	callable runtime.Value
	dispatch bytecodeMemberMethodDispatchKind
	inlineFn *runtime.FunctionValue
}

func bytecodeStaticMemberReceiverIdentityForValue(receiver runtime.Value) (bytecodeStaticMemberReceiverIdentity, bool) {
	receiver = bytecodeMaterializeRawValue(bytecodeSlotReadValue(receiver))
	switch value := receiver.(type) {
	case runtime.TypeRefValue:
		return bytecodeStaticMemberTypeRefIdentity(value)
	case *runtime.TypeRefValue:
		if value == nil {
			return bytecodeStaticMemberReceiverIdentity{}, false
		}
		return bytecodeStaticMemberTypeRefIdentity(*value)
	case runtime.StructDefinitionValue:
		return bytecodeStaticMemberStructDefinitionIdentity(value.Node)
	case *runtime.StructDefinitionValue:
		if value == nil {
			return bytecodeStaticMemberReceiverIdentity{}, false
		}
		return bytecodeStaticMemberStructDefinitionIdentity(value.Node)
	case runtime.InterfaceDefinitionValue:
		return bytecodeStaticMemberInterfaceDefinitionIdentity(value.Node)
	case *runtime.InterfaceDefinitionValue:
		if value == nil {
			return bytecodeStaticMemberReceiverIdentity{}, false
		}
		return bytecodeStaticMemberInterfaceDefinitionIdentity(value.Node)
	case runtime.PackageValue:
		return bytecodeStaticMemberPackageIdentity(value.Name, value.NamePath, value.IdentityKey, bytecodeStaticMemberReceiverPackage)
	case *runtime.PackageValue:
		if value == nil {
			return bytecodeStaticMemberReceiverIdentity{}, false
		}
		return bytecodeStaticMemberPackageIdentity(value.Name, value.NamePath, value.IdentityKey, bytecodeStaticMemberReceiverPackage)
	case runtime.DynPackageValue:
		return bytecodeStaticMemberPackageIdentity(value.Name, value.NamePath, value.IdentityKey, bytecodeStaticMemberReceiverDynPackage)
	case *runtime.DynPackageValue:
		if value == nil {
			return bytecodeStaticMemberReceiverIdentity{}, false
		}
		return bytecodeStaticMemberPackageIdentity(value.Name, value.NamePath, value.IdentityKey, bytecodeStaticMemberReceiverDynPackage)
	case runtime.ImplementationNamespaceValue:
		return bytecodeStaticMemberImplementationNamespaceIdentity(value)
	case *runtime.ImplementationNamespaceValue:
		if value == nil {
			return bytecodeStaticMemberReceiverIdentity{}, false
		}
		return bytecodeStaticMemberImplementationNamespaceIdentity(*value)
	default:
		return bytecodeStaticMemberReceiverIdentity{}, false
	}
}

func bytecodeStaticMemberTypeRefIdentity(value runtime.TypeRefValue) (bytecodeStaticMemberReceiverIdentity, bool) {
	if value.TypeName == "" {
		return bytecodeStaticMemberReceiverIdentity{}, false
	}
	return bytecodeStaticMemberReceiverIdentity{
		kind:     bytecodeStaticMemberReceiverTypeRef,
		typeName: value.TypeName,
		typeArgs: makeTypeExpressionSliceKey(value.TypeArgs),
	}, true
}

func bytecodeStaticMemberStructDefinitionIdentity(node *ast.StructDefinition) (bytecodeStaticMemberReceiverIdentity, bool) {
	if node == nil || node.ID == nil || node.ID.Name == "" {
		return bytecodeStaticMemberReceiverIdentity{}, false
	}
	return bytecodeStaticMemberReceiverIdentity{
		kind:       bytecodeStaticMemberReceiverStructDefinition,
		typeName:   node.ID.Name,
		structNode: node,
	}, true
}

func bytecodeStaticMemberInterfaceDefinitionIdentity(node *ast.InterfaceDefinition) (bytecodeStaticMemberReceiverIdentity, bool) {
	if node == nil || node.ID == nil || node.ID.Name == "" {
		return bytecodeStaticMemberReceiverIdentity{}, false
	}
	return bytecodeStaticMemberReceiverIdentity{
		kind:          bytecodeStaticMemberReceiverInterfaceDefinition,
		typeName:      node.ID.Name,
		interfaceNode: node,
	}, true
}

func bytecodePackageIdentityKey(name string, namePath []string) string {
	if len(namePath) > 0 {
		return strings.Join(namePath, "\x00")
	}
	return name
}

func bytecodeStaticMemberPackageIdentity(name string, namePath []string, identityKey string, kind bytecodeStaticMemberReceiverKind) (bytecodeStaticMemberReceiverIdentity, bool) {
	key := identityKey
	if key == "" {
		key = bytecodePackageIdentityKey(name, namePath)
	}
	if key == "" {
		return bytecodeStaticMemberReceiverIdentity{}, false
	}
	return bytecodeStaticMemberReceiverIdentity{
		kind:        kind,
		packageName: key,
	}, true
}

func bytecodeStaticMemberImplementationNamespaceIdentity(value runtime.ImplementationNamespaceValue) (bytecodeStaticMemberReceiverIdentity, bool) {
	if value.Name == nil || value.Name.Name == "" {
		return bytecodeStaticMemberReceiverIdentity{}, false
	}
	var interfaceName string
	if value.InterfaceName != nil {
		interfaceName = value.InterfaceName.Name
	}
	return bytecodeStaticMemberReceiverIdentity{
		kind:          bytecodeStaticMemberReceiverImplementationNamespace,
		implName:      value.Name.Name,
		implInterface: interfaceName,
		implTarget:    makeTypeExpressionSliceKey([]ast.TypeExpression{value.TargetType}),
	}, true
}

func bytecodeStaticMemberReceiverTypeName(identity bytecodeStaticMemberReceiverIdentity) string {
	switch identity.kind {
	case bytecodeStaticMemberReceiverTypeRef,
		bytecodeStaticMemberReceiverStructDefinition,
		bytecodeStaticMemberReceiverInterfaceDefinition:
		return identity.typeName
	default:
		return ""
	}
}

func bytecodeStaticMemberCallableCacheable(callable runtime.Value) (runtime.Value, bool) {
	switch fn := callable.(type) {
	case runtime.NativeFunctionValue:
		return fn, true
	case *runtime.NativeFunctionValue:
		if fn == nil {
			return nil, false
		}
		return fn, true
	case *runtime.FunctionValue:
		if fn == nil {
			return nil, false
		}
		return fn, true
	case *runtime.FunctionOverloadValue:
		if fn == nil || len(fn.Overloads) == 0 {
			return nil, false
		}
		return fn, true
	default:
		return nil, false
	}
}

func (vm *bytecodeVM) staticMemberCallLexicalState(memberName string, identity bytecodeStaticMemberReceiverIdentity, captureMemberOwner bool) bytecodeMemberMethodLexicalState {
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
	state.memberNameShapeVersion = env.BindingNameRevision(memberName)
	if typeName := bytecodeStaticMemberReceiverTypeName(identity); typeName != "" && vm.interp != nil {
		canonicalName, aliasName := vm.interp.canonicalTypeNamePair(typeName)
		if canonicalName != "" {
			state.receiverTypeShapeVersion = env.BindingNameRevision(canonicalName)
		}
		if aliasName != "" {
			state.receiverAliasTypeShapeVersion = env.BindingNameRevision(aliasName)
		}
	}
	if captureMemberOwner && memberName != "" {
		_, owner, ownerVersion, found := env.LookupWithOwnerAndRevisionHint(memberName, vm.bytecodeSingleThread())
		if found {
			state.memberOwner = owner
			state.memberOwnerVersion = ownerVersion
		}
	}
	return state
}

func (vm *bytecodeVM) staticMemberCallCacheKey(program *bytecodeProgram, ip int, memberName string, argCount int, receiver runtime.Value) (bytecodeStaticMemberCallCacheKey, bytecodeStaticMemberReceiverIdentity, bool) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil || vm.env == nil || memberName == "" || argCount < 0 {
		return bytecodeStaticMemberCallCacheKey{}, bytecodeStaticMemberReceiverIdentity{}, false
	}
	identity, ok := bytecodeStaticMemberReceiverIdentityForValue(receiver)
	if !ok {
		return bytecodeStaticMemberCallCacheKey{}, bytecodeStaticMemberReceiverIdentity{}, false
	}
	implActive, implStateID, implRevision, ok := vm.memberMethodImplContextState()
	if !ok {
		return bytecodeStaticMemberCallCacheKey{}, bytecodeStaticMemberReceiverIdentity{}, false
	}
	return bytecodeStaticMemberCallCacheKey{
		program:             program,
		ip:                  ip,
		env:                 vm.env,
		member:              memberName,
		argCount:            argCount,
		implContextActive:   implActive,
		implContextStateID:  implStateID,
		implContextRevision: implRevision,
		receiver:            identity,
	}, identity, true
}

func (vm *bytecodeVM) lookupCachedStaticMemberCall(program *bytecodeProgram, ip int, memberName string, argCount int, receiver runtime.Value) (bytecodeCachedStaticMemberCall, bool) {
	key, identity, ok := vm.staticMemberCallCacheKey(program, ip, memberName, argCount, receiver)
	if !ok {
		return bytecodeCachedStaticMemberCall{}, false
	}
	lexicalState := vm.staticMemberCallLexicalState(memberName, identity, false)
	if hot := vm.staticMemberCallHot; hot.valid &&
		hot.key == key &&
		hot.methodCacheVersion == lexicalState.methodCacheVersion &&
		hot.scopeStateID == lexicalState.scopeStateID &&
		hot.memberNameShapeVersion == lexicalState.memberNameShapeVersion &&
		hot.receiverTypeShapeVersion == lexicalState.receiverTypeShapeVersion &&
		hot.receiverAliasTypeShapeVersion == lexicalState.receiverAliasTypeShapeVersion &&
		vm.memberMethodLexicalOwnerValid(hot.memberOwner, hot.memberOwnerVersion) {
		vm.interp.recordBytecodeCallMemberStaticCacheHit()
		return bytecodeCachedStaticMemberCall{callable: hot.callable, dispatch: hot.dispatch, inlineFn: hot.inlineFn}, true
	}
	if vm.staticMemberCallCache == nil {
		vm.interp.recordBytecodeCallMemberStaticCacheMiss()
		return bytecodeCachedStaticMemberCall{}, false
	}
	entry, ok := vm.staticMemberCallCache[key]
	if !ok {
		vm.interp.recordBytecodeCallMemberStaticCacheMiss()
		return bytecodeCachedStaticMemberCall{}, false
	}
	if entry.methodCacheVersion != lexicalState.methodCacheVersion ||
		entry.scopeStateID != lexicalState.scopeStateID ||
		entry.memberNameShapeVersion != lexicalState.memberNameShapeVersion ||
		entry.receiverTypeShapeVersion != lexicalState.receiverTypeShapeVersion ||
		entry.receiverAliasTypeShapeVersion != lexicalState.receiverAliasTypeShapeVersion ||
		!vm.memberMethodLexicalOwnerValid(entry.memberOwner, entry.memberOwnerVersion) {
		vm.interp.recordBytecodeCallMemberStaticCacheMiss()
		return bytecodeCachedStaticMemberCall{}, false
	}
	vm.staticMemberCallHot = bytecodeInlineStaticMemberCallCacheEntry{
		valid:                         true,
		key:                           key,
		methodCacheVersion:            entry.methodCacheVersion,
		scopeStateID:                  entry.scopeStateID,
		memberNameShapeVersion:        entry.memberNameShapeVersion,
		receiverTypeShapeVersion:      entry.receiverTypeShapeVersion,
		receiverAliasTypeShapeVersion: entry.receiverAliasTypeShapeVersion,
		memberOwner:                   entry.memberOwner,
		memberOwnerVersion:            entry.memberOwnerVersion,
		callable:                      entry.callable,
		dispatch:                      entry.dispatch,
		inlineFn:                      entry.inlineFn,
	}
	vm.interp.recordBytecodeCallMemberStaticCacheHit()
	return bytecodeCachedStaticMemberCall{callable: entry.callable, dispatch: entry.dispatch, inlineFn: entry.inlineFn}, true
}

func (vm *bytecodeVM) storeCachedStaticMemberCall(program *bytecodeProgram, ip int, memberName string, argCount int, receiver runtime.Value, callable runtime.Value) bool {
	key, identity, ok := vm.staticMemberCallCacheKey(program, ip, memberName, argCount, receiver)
	if !ok {
		return false
	}
	template, ok := bytecodeStaticMemberCallableCacheable(callable)
	if !ok {
		return false
	}
	if vm.staticMemberCallCache == nil {
		vm.staticMemberCallCache = make(map[bytecodeStaticMemberCallCacheKey]bytecodeStaticMemberCallCacheEntry, 8)
	}
	lexicalState := vm.staticMemberCallLexicalState(memberName, identity, true)
	dispatch, inlineFn := bytecodeMemberMethodDispatchForTemplate(template)
	entry := bytecodeStaticMemberCallCacheEntry{
		methodCacheVersion:            lexicalState.methodCacheVersion,
		scopeStateID:                  lexicalState.scopeStateID,
		memberNameShapeVersion:        lexicalState.memberNameShapeVersion,
		receiverTypeShapeVersion:      lexicalState.receiverTypeShapeVersion,
		receiverAliasTypeShapeVersion: lexicalState.receiverAliasTypeShapeVersion,
		memberOwner:                   lexicalState.memberOwner,
		memberOwnerVersion:            lexicalState.memberOwnerVersion,
		callable:                      template,
		dispatch:                      dispatch,
		inlineFn:                      inlineFn,
	}
	vm.staticMemberCallCache[key] = entry
	vm.staticMemberCallHot = bytecodeInlineStaticMemberCallCacheEntry{
		valid:                         true,
		key:                           key,
		methodCacheVersion:            entry.methodCacheVersion,
		scopeStateID:                  entry.scopeStateID,
		memberNameShapeVersion:        entry.memberNameShapeVersion,
		receiverTypeShapeVersion:      entry.receiverTypeShapeVersion,
		receiverAliasTypeShapeVersion: entry.receiverAliasTypeShapeVersion,
		memberOwner:                   entry.memberOwner,
		memberOwnerVersion:            entry.memberOwnerVersion,
		callable:                      entry.callable,
		dispatch:                      entry.dispatch,
		inlineFn:                      entry.inlineFn,
	}
	return true
}

func bytecodeStaticMemberCallableInlineTarget(callee runtime.Value) (*runtime.FunctionValue, runtime.Value, bool, bool, bool) {
	fn, injectedReceiver, hasInjectedReceiver, ok := bytecodeResolveCachedCallNameFunction(callee)
	if !ok {
		return nil, nil, false, false, false
	}
	if fn == nil {
		return nil, nil, false, false, true
	}
	prog, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || prog == nil {
		return nil, nil, false, false, true
	}
	return fn, injectedReceiver, hasInjectedReceiver, true, false
}

func (vm *bytecodeVM) execStaticMemberCallable(callee runtime.Value, instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall, traceNode ast.Node, currentProgram *bytecodeProgram, statsEnabled bool) (*bytecodeProgram, error) {
	if target, ok := bytecodeResolveExactNativeCallTarget(callee, instr.argCount); ok {
		if vm.interp != nil {
			vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "exact_native", traceNode)
		}
		if statsEnabled {
			vm.interp.recordBytecodeCallMemberDispatch(bytecodeCallMemberStatsStaticExactNative)
		}
		args := vm.stackValuesFrom(argBase)
		vm.truncateStack(receiverIndex)
		return vm.execAndFinishExactNativeCall(target, args, callNode)
	}
	if inlineFn, injectedReceiver, hasInjectedReceiver, canInline, noBytecode := bytecodeStaticMemberCallableInlineTarget(callee); canInline {
		if newProg, err := vm.tryInlineResolvedCallFromStack(inlineFn, injectedReceiver, hasInjectedReceiver, argBase, instr.argCount, receiverIndex, callNode, currentProgram); err != nil {
			return nil, err
		} else if newProg != nil {
			if vm.interp != nil {
				vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "inline", traceNode)
			}
			if statsEnabled {
				vm.interp.recordBytecodeInlineCallHit()
				vm.interp.recordBytecodeCallMemberDispatch(bytecodeCallMemberStatsStaticInline)
			}
			return newProg, nil
		} else if statsEnabled {
			vm.interp.recordBytecodeInlineCallMiss()
		}
	} else if statsEnabled {
		vm.interp.recordBytecodeInlineCallMiss()
		if noBytecode {
			vm.interp.recordBytecodeInlineResolvedMiss(bytecodeInlineResolvedMissNoBytecode)
		}
	}
	if result, handled, err := vm.tryCallDirectFunctionValueFromStack(callee, argBase, instr.argCount, receiverIndex, callNode); handled || err != nil {
		if vm.interp != nil {
			vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "generic", traceNode)
		}
		if statsEnabled {
			vm.interp.recordBytecodeCallMemberDispatch(bytecodeCallMemberStatsStaticGeneric)
		}
		return vm.finishCompletedCall(result, err, callNode, nil)
	}
	args := vm.stackValuesFrom(argBase)
	vm.truncateStack(receiverIndex)
	needsStableArgsCopy := bytecodeCallTargetNeedsStableArgs(callee)
	if needsStableArgsCopy {
		var inline [bytecodeInlinePreparedCallArgStorage]runtime.Value
		if prepared, ok := bytecodePrepareCallArgsIntoBuffer(inline[:], args, true); ok {
			args = prepared
		} else {
			args = vm.prepareMaterializedCallArgs(args, true, bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticCall)
		}
	} else {
		args = vm.prepareMaterializedCallArgs(args, false, bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticCall)
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "generic", traceNode)
	}
	if statsEnabled {
		vm.interp.recordBytecodeCallMemberDispatch(bytecodeCallMemberStatsStaticGeneric)
	}
	result, err := vm.callCallableValueMutable(callee, args, callNode)
	return vm.finishCompletedCall(result, err, callNode, nil)
}
