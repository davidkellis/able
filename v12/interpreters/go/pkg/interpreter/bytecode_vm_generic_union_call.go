package interpreter

import (
	"fmt"
	"sync/atomic"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

const (
	bytecodeGenericUnionCallDirectSets = 16
	bytecodeGenericUnionCallDirectWays = 2
)

type bytecodeGenericUnionCallCacheEntry struct {
	valid              bool
	program            *bytecodeProgram
	ip                 int
	member             string
	methodCacheVersion uint64
	scopeStateID       uint64
	memberNameRevision uint64
	implContextActive  bool
	implContextStateID uint64
	implContextVersion uint64
	memberOwner        *runtime.Environment
	memberOwnerVersion uint64
	cached             bytecodeCachedMemberMethod
}

func bytecodeGenericUnionCallDirectIndex(ip int) int {
	return int(uint(ip) & uint(bytecodeGenericUnionCallDirectSets-1))
}

func (vm *bytecodeVM) lookupCachedGenericUnionCall(program *bytecodeProgram, ip int, member string) (bytecodeCachedMemberMethod, bool) {
	if vm == nil || vm.interp == nil || vm.env == nil || program == nil || member == "" {
		return bytecodeCachedMemberMethod{}, false
	}
	set := &vm.genericUnionCallDirect[bytecodeGenericUnionCallDirectIndex(ip)]
	for way := range set {
		entry := &set[way]
		if !entry.valid || entry.program != program || entry.ip != ip || entry.member != member {
			continue
		}
		if entry.methodCacheVersion != vm.bytecodeMethodCacheVersion() ||
			entry.scopeStateID != vm.env.BindingShapeStateID() ||
			entry.memberNameRevision != vm.env.BindingNameRevision(member) ||
			!vm.memberMethodLexicalOwnerValid(entry.memberOwner, entry.memberOwnerVersion) {
			return bytecodeCachedMemberMethod{}, false
		}
		active, stateID, version, ok := vm.memberMethodImplContextState()
		if !ok || entry.implContextActive != active || entry.implContextStateID != stateID || entry.implContextVersion != version {
			return bytecodeCachedMemberMethod{}, false
		}
		return entry.cached, entry.cached.template != nil
	}
	return bytecodeCachedMemberMethod{}, false
}

func (vm *bytecodeVM) storeCachedGenericUnionCall(program *bytecodeProgram, ip int, member string, callable runtime.Value) (bytecodeCachedMemberMethod, bool) {
	if vm == nil || vm.interp == nil || vm.env == nil || program == nil || member == "" || callable == nil {
		return bytecodeCachedMemberMethod{}, false
	}
	_, owner, ownerVersion, found := vm.env.LookupWithOwnerAndRevisionHint(member, vm.bytecodeSingleThread())
	if !found || owner == nil {
		return bytecodeCachedMemberMethod{}, false
	}
	active, stateID, version, ok := vm.memberMethodImplContextState()
	if !ok {
		return bytecodeCachedMemberMethod{}, false
	}
	dispatch, inlineFn := bytecodeMemberMethodDispatchForTemplate(callable)
	cached := bytecodeCachedMemberMethod{template: callable, dispatch: dispatch, inlineFn: inlineFn}
	setIndex := bytecodeGenericUnionCallDirectIndex(ip)
	set := &vm.genericUnionCallDirect[setIndex]
	way := -1
	for index := range set {
		if (set[index].valid && set[index].program == program && set[index].ip == ip && set[index].member == member) || !set[index].valid {
			way = index
			break
		}
	}
	if way < 0 {
		way = int(vm.genericUnionCallDirectNext[setIndex] % bytecodeGenericUnionCallDirectWays)
		vm.genericUnionCallDirectNext[setIndex]++
	}
	set[way] = bytecodeGenericUnionCallCacheEntry{
		valid:              true,
		program:            program,
		ip:                 ip,
		member:             member,
		methodCacheVersion: vm.bytecodeMethodCacheVersion(),
		scopeStateID:       vm.env.BindingShapeStateID(),
		memberNameRevision: vm.env.BindingNameRevision(member),
		implContextActive:  active,
		implContextStateID: stateID,
		implContextVersion: version,
		memberOwner:        owner,
		memberOwnerVersion: ownerVersion,
		cached:             cached,
	}
	return cached, true
}

func (vm *bytecodeVM) execCallGenericUnionMember(instr bytecodeInstruction, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if instr.argCount < 0 {
		return nil, fmt.Errorf("bytecode generic-union call-member arg count invalid")
	}
	if vm.stackDepth() < instr.argCount+1 {
		return nil, fmt.Errorf("bytecode stack underflow")
	}
	if instr.name == "" {
		return nil, fmt.Errorf("bytecode generic-union call-member missing member name")
	}
	receiverIndex := vm.stackDepth() - instr.argCount - 1
	argBase := receiverIndex + 1
	receiver := vm.stackValue(receiverIndex)
	if instr.safe && isNilRuntimeValue(receiver) {
		vm.truncateStack(receiverIndex)
		vm.appendStackValue(runtime.NilValue{})
		vm.ip++
		return nil, nil
	}
	receiver = vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion, receiver)
	vm.setStackValue(receiverIndex, receiver)
	callNode, _ := instr.node.(*ast.FunctionCall)
	if directCallMemberNeedsMemberAccessFallback(receiver, instr.name) {
		return vm.execCallMember(instr, currentProgram)
	}
	if cached, ok := vm.lookupCachedGenericUnionCall(currentProgram, vm.ip, instr.name); ok {
		if vm.interp.bytecodeStatsEnabled {
			atomic.AddUint64(&vm.interp.bytecodeGenericUnionCallCacheHits, 1)
		}
		return vm.execCachedResolvedMemberCall(cached, instr.name, receiverIndex, argBase, instr.argCount, callNode, currentProgram)
	}
	if vm.interp.bytecodeStatsEnabled {
		atomic.AddUint64(&vm.interp.bytecodeGenericUnionCallCacheMisses, 1)
	}
	receiverType := vm.interp.staticReceiverTypeForCall(callNode, vm.env)
	callable, found := vm.interp.resolveStaticGenericUnionMethodCallable(vm.env, instr.name, receiverType)
	if !found {
		return vm.execCallMember(instr, currentProgram)
	}
	cached, ok := vm.storeCachedGenericUnionCall(currentProgram, vm.ip, instr.name, callable)
	if !ok {
		dispatch, inlineFn := bytecodeMemberMethodDispatchForTemplate(callable)
		cached = bytecodeCachedMemberMethod{template: callable, dispatch: dispatch, inlineFn: inlineFn}
	}
	return vm.execCachedResolvedMemberCall(cached, instr.name, receiverIndex, argBase, instr.argCount, callNode, currentProgram)
}
