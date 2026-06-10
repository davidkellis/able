package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type bytecodeCallNameDispatchKind uint8

const (
	bytecodeCallNameDispatchGeneric bytecodeCallNameDispatchKind = iota
	bytecodeCallNameDispatchExactNative
	bytecodeCallNameDispatchInline
)

type bytecodeCallNameCacheEntry struct {
	name                 string
	env                  *runtime.Environment
	envVersion           uint64
	nameShapeStateID     uint64
	bindingShapeVersion  uint64
	nameShapeVersion     uint64
	owner                *runtime.Environment
	ownerVersion         uint64
	callee               runtime.Value
	dispatch             bytecodeCallNameDispatchKind
	exactTarget          bytecodeExactNativeCallTarget
	inlineFn             *runtime.FunctionValue
	injectedReceiver     runtime.Value
	hasInjectedReceiver  bool
	inlineProgram        *bytecodeProgram
	inlineLayout         *bytecodeFrameLayout
	inlineReturnGenerics map[string]struct{}
	inlineDirect         bool
	inlineI32ParamMask   uint8
	inlineKeepNilI32Mask uint8
	inlineCoercionMask   uint8
	inlineRuntimeMatch   bytecodeCallNameRuntimeMatchPlan
	inlineRuntimeValid   bool
	needsStableArgsCopy  bool
}

type bytecodeInlineCallNameCacheEntry struct {
	valid   bool
	program *bytecodeProgram
	ip      int
	entry   *bytecodeCallNameCacheEntry
}

const bytecodeCallNameRuntimeMatchParamLimit = 8

type bytecodeCallNameRuntimeMatchPlan struct {
	minArgs    int
	maxArgs    int
	paramStart int
	paramCount int
	checks     [bytecodeCallNameRuntimeMatchParamLimit]bytecodeSimpleTypeCheck
	arityOnly  bool
}

func bytecodeResolveCachedCallNameFunction(callee runtime.Value) (*runtime.FunctionValue, runtime.Value, bool, bool) {
	if fn, injectedReceiver, hasInjectedReceiver, ok := inlineCallFunctionValue(callee); ok {
		return fn, injectedReceiver, hasInjectedReceiver, true
	}
	switch fn := callee.(type) {
	case *runtime.FunctionOverloadValue:
		if fn == nil || len(fn.Overloads) != 1 || fn.Overloads[0] == nil {
			return nil, nil, false, false
		}
		return fn.Overloads[0], nil, false, true
	case runtime.BoundMethodValue:
		overloads := functionOverloadsView(fn.Method)
		if len(overloads) != 1 || overloads[0] == nil {
			return nil, nil, false, false
		}
		return overloads[0], fn.Receiver, true, true
	case *runtime.BoundMethodValue:
		if fn == nil {
			return nil, nil, false, false
		}
		overloads := functionOverloadsView(fn.Method)
		if len(overloads) != 1 || overloads[0] == nil {
			return nil, nil, false, false
		}
		return overloads[0], fn.Receiver, true, true
	default:
		return nil, nil, false, false
	}
}

func bytecodeBuildCallNameCacheEntry(name string, lookup bytecodeResolvedIdentifierLookup, callee runtime.Value, argCount int, callNode *ast.FunctionCall) bytecodeCallNameCacheEntry {
	entry := bytecodeCallNameCacheEntry{
		name:                name,
		env:                 lookup.env,
		envVersion:          lookup.envVersion,
		owner:               lookup.owner,
		ownerVersion:        lookup.ownerVersion,
		callee:              callee,
		dispatch:            bytecodeCallNameDispatchGeneric,
		needsStableArgsCopy: bytecodeCallTargetNeedsStableArgs(callee),
	}
	if lookup.env != nil && lookup.owner != lookup.env {
		entry.nameShapeStateID = lookup.env.BindingShapeStateID()
		entry.bindingShapeVersion = lookup.env.BindingShapeRevision()
		entry.nameShapeVersion = lookup.env.BindingNameRevision(name)
	}
	if target, ok := bytecodeResolveExactNativeCallTarget(callee, argCount); ok {
		entry.dispatch = bytecodeCallNameDispatchExactNative
		entry.exactTarget = target
		return entry
	}
	if fn, injectedReceiver, hasInjectedReceiver, ok := bytecodeResolveCachedCallNameFunction(callee); ok {
		entry.dispatch = bytecodeCallNameDispatchInline
		entry.inlineFn = fn
		entry.injectedReceiver = injectedReceiver
		entry.hasInjectedReceiver = hasInjectedReceiver
		if plan, ok := bytecodeBuildCallNameRuntimeMatchPlan(fn); ok {
			entry.inlineRuntimeMatch = plan
			entry.inlineRuntimeValid = true
		}
		if !hasInjectedReceiver {
			entry.inlineProgram, entry.inlineLayout, entry.inlineReturnGenerics, entry.inlineDirect = bytecodeDirectCallNameInlineShape(fn, argCount, callNode)
			if entry.inlineDirect {
				entry.inlineI32ParamMask, entry.inlineKeepNilI32Mask, entry.inlineCoercionMask = bytecodeDirectCallNameInlineMasks(entry.inlineLayout, argCount)
			}
		}
	}
	return entry
}

func (entry *bytecodeCallNameCacheEntry) lexicalShapeValid(env *runtime.Environment) bool {
	if entry == nil || env == nil {
		return false
	}
	if entry.nameShapeStateID == 0 && entry.env == env && entry.owner == env {
		return true
	}
	if entry.nameShapeStateID != env.BindingShapeStateID() {
		return false
	}
	bindingShapeVersion := env.BindingShapeRevision()
	if entry.bindingShapeVersion == bindingShapeVersion {
		return true
	}
	if entry.nameShapeVersion != env.BindingNameRevision(entry.name) {
		return false
	}
	entry.bindingShapeVersion = bindingShapeVersion
	return true
}

func bytecodeBuildCallNameRuntimeMatchPlan(fn *runtime.FunctionValue) (bytecodeCallNameRuntimeMatchPlan, bool) {
	if fn == nil || fn.Declaration == nil {
		return bytecodeCallNameRuntimeMatchPlan{}, false
	}
	switch decl := fn.Declaration.(type) {
	case *ast.FunctionDefinition:
		paramCount := len(decl.Params)
		if paramCount > bytecodeCallNameRuntimeMatchParamLimit {
			return bytecodeCallNameRuntimeMatchPlan{}, false
		}
		expectedArgs := paramCount
		paramStart := 0
		if decl.IsMethodShorthand {
			expectedArgs++
			paramStart = 1
		}
		minArgs := expectedArgs
		if paramCount > 0 && isNullableParam(decl.Params[paramCount-1]) {
			minArgs--
		}
		var checks [bytecodeCallNameRuntimeMatchParamLimit]bytecodeSimpleTypeCheck
		arityOnly := true
		generics := fn.GenericNameSet(decl)
		for idx, param := range decl.Params {
			if param == nil {
				return bytecodeCallNameRuntimeMatchPlan{}, false
			}
			if param.ParamType == nil || typeExpressionUsesGenerics(param.ParamType, generics) {
				continue
			}
			if _, ok := param.ParamType.(*ast.WildcardTypeExpression); ok {
				continue
			}
			simple, ok := param.ParamType.(*ast.SimpleTypeExpression)
			if !ok || simple == nil || simple.Name == nil || simple.Name.Name == "Self" {
				return bytecodeCallNameRuntimeMatchPlan{}, false
			}
			check := bytecodeSimpleTypeCheckForName(simple.Name.Name)
			if check == bytecodeSimpleTypeCheckUnknown {
				return bytecodeCallNameRuntimeMatchPlan{}, false
			}
			checks[idx] = check
			arityOnly = false
		}
		return bytecodeCallNameRuntimeMatchPlan{
			minArgs:    minArgs,
			maxArgs:    expectedArgs,
			paramStart: paramStart,
			paramCount: paramCount,
			checks:     checks,
			arityOnly:  arityOnly,
		}, true
	case *ast.LambdaExpression:
		return bytecodeCallNameRuntimeMatchPlan{
			minArgs:    len(decl.Params),
			maxArgs:    len(decl.Params),
			paramStart: 0,
			paramCount: len(decl.Params),
			arityOnly:  true,
		}, true
	default:
		return bytecodeCallNameRuntimeMatchPlan{}, false
	}
}

func bytecodeCallNameRuntimeSimpleMatch(check bytecodeSimpleTypeCheck, value runtime.Value) bool {
	if check == bytecodeSimpleTypeCheckUnknown {
		return true
	}
	value = bytecodeSlotReadValue(value)
	switch check {
	case bytecodeSimpleTypeCheckString:
		_, ok := value.(runtime.StringValue)
		return ok
	case bytecodeSimpleTypeCheckBool:
		_, ok := value.(runtime.BoolValue)
		return ok
	default:
		return inlineCoercionUnnecessaryBySimpleCheck(check, value)
	}
}

func (plan *bytecodeCallNameRuntimeMatchPlan) matches(args []runtime.Value) (bool, bool) {
	if plan == nil {
		return false, false
	}
	if len(args) < plan.minArgs || len(args) > plan.maxArgs {
		return false, true
	}
	if plan.arityOnly {
		return true, true
	}
	paramArgCount := len(args) - plan.paramStart
	if paramArgCount < 0 || paramArgCount > plan.paramCount {
		return false, true
	}
	for idx := 0; idx < paramArgCount; idx++ {
		check := plan.checks[idx]
		if check == bytecodeSimpleTypeCheckUnknown {
			continue
		}
		argIndex := plan.paramStart + idx
		if argIndex < 0 || argIndex >= len(args) {
			return false, true
		}
		if !bytecodeCallNameRuntimeSimpleMatch(check, args[argIndex]) {
			return false, false
		}
	}
	return true, true
}

func bytecodeDirectCallNameInlineShape(fn *runtime.FunctionValue, argCount int, callNode *ast.FunctionCall) (*bytecodeProgram, *bytecodeFrameLayout, map[string]struct{}, bool) {
	if fn == nil {
		return nil, nil, nil, false
	}
	if callNode != nil && len(callNode.TypeArguments) > 0 {
		return nil, nil, nil, false
	}
	if fn.MethodSet != nil && len(fn.MethodSet.GenericParams) > 0 {
		return nil, nil, nil, false
	}
	if bytecodeInlineSkipsGenericLambda(fn) {
		return nil, nil, nil, false
	}
	prog, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || prog == nil || prog.frameLayout == nil {
		return nil, nil, nil, false
	}
	layout := prog.frameLayout
	if layout.methodShorthand || layout.paramSlots != argCount {
		return nil, nil, nil, false
	}
	return prog, layout, bytecodeInlineReturnGenericNames(fn, prog), true
}

func bytecodeDirectCallNameInlineMasks(layout *bytecodeFrameLayout, argCount int) (i32Mask uint8, keepNilI32Mask uint8, coercionMask uint8) {
	if layout == nil || argCount <= 0 {
		return 0, 0, 0
	}
	if argCount > 3 {
		argCount = 3
	}
	for idx := 0; idx < argCount; idx++ {
		bit := uint8(1 << uint(idx))
		if idx < len(layout.paramNeedsCoercion) && layout.paramNeedsCoercion[idx] {
			coercionMask |= bit
		}
		if idx >= len(layout.slotKinds) || layout.slotKinds[idx] != bytecodeCellKindI32 {
			continue
		}
		i32Mask |= bit
		if idx < len(layout.paramSimpleChecks) && layout.paramSimpleChecks[idx] == bytecodeSimpleTypeCheckI32 {
			keepNilI32Mask |= bit
			continue
		}
		if idx < len(layout.paramSimpleTypes) && layout.paramSimpleTypes[idx] == "i32" {
			keepNilI32Mask |= bit
		}
	}
	return i32Mask, keepNilI32Mask, coercionMask
}

func (vm *bytecodeVM) callNameCacheEntries(program *bytecodeProgram, create bool) ([]*bytecodeCallNameCacheEntry, bool) {
	if vm == nil || program == nil {
		return nil, false
	}
	if len(program.instructions) == 0 {
		return nil, false
	}
	if vm.callNameHotProgram == program && vm.callNameHotEntries != nil {
		return vm.callNameHotEntries, true
	}
	entries, ok := vm.callNameCache[program]
	if !ok {
		if !create {
			return nil, false
		}
		entries = make([]*bytecodeCallNameCacheEntry, len(program.instructions))
		if vm.callNameCache == nil {
			vm.callNameCache = make(map[*bytecodeProgram][]*bytecodeCallNameCacheEntry, 8)
		}
		vm.callNameCache[program] = entries
	}
	vm.callNameHotProgram = program
	vm.callNameHotEntries = entries
	return entries, true
}

func (vm *bytecodeVM) activeCallNameCacheEntries(program *bytecodeProgram, create bool) ([]*bytecodeCallNameCacheEntry, bool) {
	if vm == nil || program == nil {
		return nil, false
	}
	if vm.activeLookup.program == program && vm.activeLookup.callNameEntries != nil {
		return vm.activeLookup.callNameEntries, true
	}
	entries, ok := vm.callNameCacheEntries(program, create)
	if !ok {
		return nil, false
	}
	if vm.activeLookup.program == program {
		vm.activeLookup.callNameEntries = entries
	}
	return entries, true
}

func (vm *bytecodeVM) lookupCachedCallName(program *bytecodeProgram, ip int, name string) (*bytecodeCallNameCacheEntry, bool) {
	if vm == nil || vm.env == nil {
		return nil, false
	}
	currentEnv := vm.env
	if hot := vm.callNameHot; hot.valid &&
		hot.program == program &&
		hot.ip == ip &&
		hot.entry != nil &&
		hot.entry.name == name &&
		vm.callNameEntryCurrentEnvValid(hot.entry, currentEnv) {
		entry := hot.entry
		if !entry.lexicalShapeValid(currentEnv) {
			return nil, false
		}
		if entry.owner == nil {
			return nil, false
		}
		if entry.ownerVersion != vm.bytecodeEnvRevision(entry.owner) {
			return nil, false
		}
		return entry, true
	}
	entries, ok := vm.activeCallNameCacheEntries(program, false)
	if !ok {
		return nil, false
	}
	entry := entries[ip]
	if entry == nil || entry.name != name || !vm.callNameEntryCurrentEnvValid(entry, currentEnv) {
		return nil, false
	}
	if !entry.lexicalShapeValid(currentEnv) {
		return nil, false
	}
	if entry.owner == nil {
		return nil, false
	}
	if entry.ownerVersion != vm.bytecodeEnvRevision(entry.owner) {
		return nil, false
	}
	vm.callNameHot = bytecodeInlineCallNameCacheEntry{
		valid:   true,
		program: program,
		ip:      ip,
		entry:   entry,
	}
	return entry, true
}

func (vm *bytecodeVM) callNameEntryCurrentEnvValid(entry *bytecodeCallNameCacheEntry, currentEnv *runtime.Environment) bool {
	if vm == nil || entry == nil || currentEnv == nil {
		return false
	}
	if entry.env == currentEnv {
		return true
	}
	return vm.interp != nil && vm.interp.global != nil && entry.owner == vm.interp.global
}

func (vm *bytecodeVM) storeCachedCallName(program *bytecodeProgram, ip int, entry bytecodeCallNameCacheEntry) *bytecodeCallNameCacheEntry {
	if vm == nil || program == nil || entry.name == "" || entry.env == nil || entry.owner == nil {
		return nil
	}
	if entry.nameShapeStateID == 0 && entry.owner != entry.env {
		entry.nameShapeStateID = entry.env.BindingShapeStateID()
		entry.nameShapeVersion = entry.env.BindingNameRevision(entry.name)
	}
	if entry.nameShapeStateID != 0 && entry.bindingShapeVersion == 0 {
		entry.bindingShapeVersion = entry.env.BindingShapeRevision()
	}
	entries, ok := vm.activeCallNameCacheEntries(program, true)
	if !ok {
		return nil
	}
	cached := entries[ip]
	if cached == nil {
		cached = new(bytecodeCallNameCacheEntry)
		entries[ip] = cached
	}
	*cached = entry
	vm.callNameHot = bytecodeInlineCallNameCacheEntry{
		valid:   true,
		program: program,
		ip:      ip,
		entry:   cached,
	}
	return cached
}

func (vm *bytecodeVM) tryInvokeCachedCallNameResolvedFunctionFromStack(entry *bytecodeCallNameCacheEntry, argBase int, argCount int, truncateTo int, callNode *ast.FunctionCall) (runtime.Value, bool, error) {
	if vm == nil || vm.interp == nil || entry == nil || entry.inlineFn == nil {
		return nil, false, nil
	}
	if argBase < 0 || argCount < 0 || argBase+argCount > vm.stackDepth() {
		return nil, false, fmt.Errorf("bytecode stack underflow")
	}
	if truncateTo < 0 || truncateTo > argBase {
		return nil, false, fmt.Errorf("bytecode stack underflow")
	}
	args := vm.stackValues(argBase, argBase+argCount)
	needsStableCopy := entry.needsStableArgsCopy || entry.hasInjectedReceiver
	args = vm.prepareResolvedFunctionCallArgsWithOptionalReceiver(args, needsStableCopy, entry.injectedReceiver, entry.hasInjectedReceiver)
	if entry.inlineRuntimeValid {
		matched, decided := entry.inlineRuntimeMatch.matches(args)
		if decided {
			if !matched {
				return nil, false, nil
			}
		} else if !vm.interp.matchesSingleRuntimeOverload(entry.inlineFn, args) {
			return nil, false, nil
		}
	} else {
		if len(args) < minArgsForFunctionValue(entry.inlineFn) {
			return nil, false, nil
		}
		if !vm.interp.matchesSingleRuntimeOverload(entry.inlineFn, args) {
			return nil, false, nil
		}
	}
	vm.interp.recordBytecodeDirectFunctionStackHit()
	vm.truncateStack(truncateTo)
	result, err := vm.interp.callResolvedFunctionValue(entry.inlineFn, entry.inlineFn, args, vm.env, callNode, true)
	return result, true, err
}

func (vm *bytecodeVM) execCachedCallName(entry *bytecodeCallNameCacheEntry, argBase int, argCount int, callNode *ast.FunctionCall, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if entry == nil {
		return nil, fmt.Errorf("bytecode cached call entry missing")
	}
	traceEnabled := vm.interp != nil && vm.interp.bytecodeTraceEnabled
	statsEnabled := vm.interp != nil && vm.interp.bytecodeStatsEnabled
	var traceNode ast.Node
	if traceEnabled && callNode != nil {
		traceNode = callNode
	}
	switch entry.dispatch {
	case bytecodeCallNameDispatchExactNative:
		if traceEnabled {
			vm.interp.recordBytecodeCallTrace("call_name", entry.name, "name", "exact_native", traceNode)
		}
		if statsEnabled {
			vm.interp.recordBytecodeCallNameDispatch(bytecodeCallNameStatsExactNative)
		}
		args := vm.stackValuesFrom(argBase)
		bytecodePrepareExactNativeCallArgs(entry.exactTarget, args)
		vm.truncateStack(argBase)
		if result, handled, err := vm.tryExecExactNativeArrayReadFast(entry.name, args); handled || err != nil {
			return vm.finishCompletedCall(result, err, callNode, nil)
		}
		return vm.execAndFinishExactNativeCall(entry.exactTarget, args, callNode)
	case bytecodeCallNameDispatchInline:
		if newProg, handled, err := vm.tryInlineCachedCallNameDirectFromStack(entry, argBase, argCount, callNode, currentProgram); handled || err != nil {
			if err != nil {
				return nil, err
			}
			if traceEnabled {
				vm.interp.recordBytecodeCallTrace("call_name", entry.name, "name", "inline", traceNode)
			}
			if statsEnabled {
				vm.interp.recordBytecodeCallNameDispatch(bytecodeCallNameStatsInlineDirectStack)
				vm.interp.recordBytecodeInlineCallHit()
			}
			return newProg, nil
		}
		if newProg, err := vm.tryInlineResolvedCallFromStack(entry.inlineFn, entry.injectedReceiver, entry.hasInjectedReceiver, argBase, argCount, argBase, callNode, currentProgram); err != nil {
			return nil, err
		} else if newProg != nil {
			if traceEnabled {
				vm.interp.recordBytecodeCallTrace("call_name", entry.name, "name", "inline", traceNode)
			}
			if statsEnabled {
				vm.interp.recordBytecodeCallNameDispatch(bytecodeCallNameStatsInlineResolved)
				vm.interp.recordBytecodeInlineCallHit()
			}
			return newProg, nil
		}
		if statsEnabled {
			vm.interp.recordBytecodeInlineCallMiss()
		}
		if traceEnabled {
			vm.interp.recordBytecodeCallTrace("call_name", entry.name, "name", "generic", traceNode)
		}
		if result, handled, err := vm.tryInvokeCachedCallNameResolvedFunctionFromStack(entry, argBase, argCount, argBase, callNode); handled || err != nil {
			if handled && statsEnabled {
				vm.interp.recordBytecodeCallNameDispatch(bytecodeCallNameStatsResolvedFunction)
			}
			return vm.finishCompletedCall(result, err, callNode, nil)
		}
		args := vm.stackValuesFrom(argBase)
		vm.truncateStack(argBase)
		if entry.needsStableArgsCopy {
			var inline [bytecodeInlinePreparedCallArgStorage]runtime.Value
			if prepared, ok := bytecodePrepareCallArgsIntoBuffer(inline[:], args, true); ok {
				args = prepared
			} else {
				args = vm.prepareMaterializedCallArgs(args, true, bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticCall)
			}
		} else {
			args = vm.prepareMaterializedCallArgs(args, false, bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticCall)
		}
		if statsEnabled {
			vm.interp.recordBytecodeCallNameDispatch(bytecodeCallNameStatsGenericFallback)
		}
		result, err := vm.callCallableValueMutable(entry.callee, args, callNode)
		return vm.finishCompletedCall(result, err, callNode, nil)
	}
	args := vm.stackValuesFrom(argBase)
	vm.truncateStack(argBase)
	if entry.needsStableArgsCopy {
		var inline [bytecodeInlinePreparedCallArgStorage]runtime.Value
		if prepared, ok := bytecodePrepareCallArgsIntoBuffer(inline[:], args, true); ok {
			args = prepared
		} else {
			args = vm.prepareMaterializedCallArgs(args, true, bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticCall)
		}
	} else {
		args = vm.prepareMaterializedCallArgs(args, false, bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticCall)
	}
	if traceEnabled {
		vm.interp.recordBytecodeCallTrace("call_name", entry.name, "name", "generic", traceNode)
	}
	if statsEnabled {
		vm.interp.recordBytecodeCallNameDispatch(bytecodeCallNameStatsGenericFallback)
	}
	result, err := vm.callCallableValueMutable(entry.callee, args, callNode)
	return vm.finishCompletedCall(result, err, callNode, nil)
}

func bytecodeCallNameArgSlots(instr bytecodeInstruction) [3]int {
	return [3]int{instr.target, instr.loopBreak, instr.loopContinue}
}

func (vm *bytecodeVM) tryInlineCachedCallNameDirectFromSlots(entry *bytecodeCallNameCacheEntry, instr bytecodeInstruction, callNode *ast.FunctionCall, currentProgram *bytecodeProgram) (*bytecodeProgram, bool, error) {
	if entry == nil || !entry.inlineDirect || !instr.slotArgs {
		return nil, false, nil
	}
	if instr.argCount < 0 || instr.argCount > 3 {
		return nil, true, fmt.Errorf("bytecode slot-arg call count invalid")
	}
	if callNode != nil && len(callNode.TypeArguments) > 0 {
		return nil, false, nil
	}
	fn := entry.inlineFn
	prog := entry.inlineProgram
	layout := entry.inlineLayout
	if fn == nil || prog == nil || layout == nil || instr.argCount != layout.paramSlots {
		return nil, false, nil
	}
	argSlot0, argSlot1, argSlot2 := instr.target, instr.loopBreak, instr.loopContinue
	i32ParamMask := entry.inlineI32ParamMask
	keepNilI32Mask := entry.inlineKeepNilI32Mask
	coercionMask := entry.inlineCoercionMask
	slots := vm.acquireSlotFrame(layout.slotCount)
	calleeI32Values, calleeI32Valid := vm.acquireInlineCalleeI32RegisterFrame(layout)
	if !layout.anyParamCoercion {
		for idx := 0; idx < instr.argCount; idx++ {
			callerSlot := argSlot0
			switch idx {
			case 1:
				callerSlot = argSlot1
			case 2:
				callerSlot = argSlot2
			}
			if callerSlot < 0 || callerSlot >= len(vm.slots) {
				vm.releaseSlotFrame(slots)
				vm.releaseI32RegisterFrame(calleeI32Values, calleeI32Valid)
				return nil, true, fmt.Errorf("bytecode slot-arg call slot out of range")
			}
			if i32ParamMask&(1<<uint(idx)) != 0 {
				if seedInlineCalleeI32RegisterSlotFromValidatedCallerSlot(vm, calleeI32Values, calleeI32Valid, idx, callerSlot) {
					if keepNilI32Mask&(1<<uint(idx)) != 0 {
						slots[idx] = nil
						continue
					}
				}
			}
			if vm.copySlotRawIntegerCellInto(slots, idx, callerSlot) {
				if i32ParamMask&(1<<uint(idx)) != 0 {
					seedInlineCalleeI32RegisterSlotUnchecked(calleeI32Values, calleeI32Valid, idx, slots[idx])
				}
				continue
			}
			arg := vm.slotMaterializedValue(callerSlot)
			vm.copyInlineCallArgToSlot(slots, idx, arg)
			if i32ParamMask&(1<<uint(idx)) != 0 {
				seedInlineCalleeI32RegisterSlotUnchecked(calleeI32Values, calleeI32Valid, idx, arg)
			}
		}
	} else {
		for idx := 0; idx < layout.paramSlots; idx++ {
			callerSlot := argSlot0
			switch idx {
			case 1:
				callerSlot = argSlot1
			case 2:
				callerSlot = argSlot2
			}
			if callerSlot < 0 || callerSlot >= len(vm.slots) {
				vm.releaseSlotFrame(slots)
				vm.releaseI32RegisterFrame(calleeI32Values, calleeI32Valid)
				return nil, true, fmt.Errorf("bytecode slot-arg call slot out of range")
			}
			if i32ParamMask&(1<<uint(idx)) != 0 {
				if seedInlineCalleeI32RegisterSlotFromValidatedCallerSlot(vm, calleeI32Values, calleeI32Valid, idx, callerSlot) {
					if keepNilI32Mask&(1<<uint(idx)) != 0 {
						slots[idx] = nil
						continue
					}
				}
			}
			arg := vm.slotMaterializedValue(callerSlot)
			if coercionMask&(1<<uint(idx)) != 0 {
				paramType := inlineParamType(layout, idx)
				if !inlineParamCoercionUnnecessary(vm.interp, layout, idx, paramType, arg) {
					if coerced, ok, err := inlineCoerceValueBySimpleType(inlineParamSimpleType(layout, idx), arg); err != nil {
						vm.releaseSlotFrame(slots)
						vm.releaseI32RegisterFrame(calleeI32Values, calleeI32Valid)
						return nil, true, err
					} else if ok {
						arg = coerced
					} else {
						coerced, err := vm.interp.coerceValueToType(paramType, arg)
						if err != nil {
							vm.releaseSlotFrame(slots)
							vm.releaseI32RegisterFrame(calleeI32Values, calleeI32Valid)
							return nil, true, err
						}
						arg = coerced
					}
				}
			}
			vm.copyInlineCallArgToSlot(slots, idx, arg)
			if i32ParamMask&(1<<uint(idx)) != 0 {
				seedInlineCalleeI32RegisterSlotUnchecked(calleeI32Values, calleeI32Valid, idx, arg)
			}
		}
	}
	if layout.selfCallSlot >= 0 && layout.selfCallSlot < len(slots) {
		slots[layout.selfCallSlot] = fn
	}
	calleeEnv := vm.bytecodeCalleeEnv(fn.Closure)

	hasImplicit := layout.paramSlots > 0 && layout.usesImplicitMember
	if hasImplicit {
		state := vm.interp.stateFromEnv(calleeEnv)
		state.pushImplicitReceiver(vm.slotMaterializedValue(argSlot0))
	}

	selfFast := bytecodeCanUseSelfFastFrame(currentProgram, prog, vm.env, calleeEnv)
	vm.pushCallFrame(vm.ip+1, currentProgram, vm.slots, vm.env, entry.inlineReturnGenerics, len(vm.iterStack), len(vm.loopStack), hasImplicit, selfFast)
	vm.slots = slots
	vm.prepareValueSlotI32Frame(prog)
	vm.env = calleeEnv
	vm.ip = 0
	vm.installInlineCalleeI32RegisterFrame(prog, calleeI32Values, calleeI32Valid)
	return prog, true, nil
}

func (vm *bytecodeVM) tryInlineCachedCallNameDirectFromStack(entry *bytecodeCallNameCacheEntry, argBase int, argCount int, callNode *ast.FunctionCall, currentProgram *bytecodeProgram) (*bytecodeProgram, bool, error) {
	if entry == nil || !entry.inlineDirect {
		return nil, false, nil
	}
	if argBase < 0 || argCount < 0 || argBase+argCount > vm.stackDepth() {
		return nil, true, fmt.Errorf("bytecode stack underflow")
	}
	if callNode != nil && len(callNode.TypeArguments) > 0 {
		return nil, false, nil
	}
	fn := entry.inlineFn
	prog := entry.inlineProgram
	layout := entry.inlineLayout
	if fn == nil || prog == nil || layout == nil || argCount != layout.paramSlots {
		return nil, false, nil
	}
	slots := vm.acquireSlotFrame(layout.slotCount)
	calleeI32Values, calleeI32Valid := vm.acquireInlineCalleeI32RegisterFrame(layout)
	if !layout.anyParamCoercion {
		vm.inlineCopyArgsToSlots(slots, vm.stackValues(argBase, argBase+argCount), argCount)
		for idx := 0; idx < argCount; idx++ {
			seedInlineCalleeI32RegisterSlot(layout, calleeI32Values, calleeI32Valid, idx, vm.stackValue(argBase+idx))
		}
	} else {
		for idx := 0; idx < layout.paramSlots; idx++ {
			arg := vm.stackValue(argBase + idx)
			paramType := inlineParamType(layout, idx)
			if inlineParamNeedsRuntimeCoercion(layout, idx, fn) && !inlineParamCoercionUnnecessary(vm.interp, layout, idx, paramType, arg) {
				if coerced, ok, err := inlineCoerceValueBySimpleType(inlineParamSimpleType(layout, idx), arg); err != nil {
					vm.releaseSlotFrame(slots)
					vm.releaseI32RegisterFrame(calleeI32Values, calleeI32Valid)
					return nil, true, err
				} else if ok {
					arg = coerced
				} else {
					coerced, err := vm.interp.coerceValueToType(paramType, arg)
					if err != nil {
						vm.releaseSlotFrame(slots)
						vm.releaseI32RegisterFrame(calleeI32Values, calleeI32Valid)
						return nil, true, err
					}
					arg = coerced
				}
			}
			vm.copyInlineCallArgToSlot(slots, idx, arg)
			seedInlineCalleeI32RegisterSlot(layout, calleeI32Values, calleeI32Valid, idx, arg)
		}
	}
	if layout.selfCallSlot >= 0 && layout.selfCallSlot < len(slots) {
		slots[layout.selfCallSlot] = fn
	}
	calleeEnv := vm.bytecodeCalleeEnv(fn.Closure)

	hasImplicit := layout.paramSlots > 0 && layout.usesImplicitMember
	if hasImplicit {
		state := vm.interp.stateFromEnv(calleeEnv)
		state.pushImplicitReceiver(bytecodeStackSnapshotValue(vm.stackValue(argBase)))
	}

	vm.truncateStack(argBase)
	selfFast := bytecodeCanUseSelfFastFrame(currentProgram, prog, vm.env, calleeEnv)
	vm.pushCallFrame(vm.ip+1, currentProgram, vm.slots, vm.env, entry.inlineReturnGenerics, len(vm.iterStack), len(vm.loopStack), hasImplicit, selfFast)
	vm.slots = slots
	vm.prepareValueSlotI32Frame(prog)
	vm.env = calleeEnv
	vm.ip = 0
	vm.installInlineCalleeI32RegisterFrame(prog, calleeI32Values, calleeI32Valid)
	return prog, true, nil
}
