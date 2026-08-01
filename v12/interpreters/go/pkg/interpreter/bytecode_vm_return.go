package interpreter

import (
	"fmt"
	"math"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) restoreCallFrameControlStacks(iterBase int, loopBase int) {
	if vm == nil {
		return
	}
	if len(vm.iterStack) > iterBase {
		for idx := len(vm.iterStack) - 1; idx >= iterBase; idx-- {
			vm.closeForLoopIterator(&vm.iterStack[idx])
		}
		vm.iterStack = vm.iterStack[:iterBase]
	}
	if len(vm.loopStack) > loopBase {
		vm.loopStack = vm.loopStack[:loopBase]
	}
}

func bytecodeTryMaterializedProgramReturnNoCoercion(vm *bytecodeVM, i *Interpreter, program *bytecodeProgram, instr *bytecodeInstruction, val runtime.Value, knownReturnSimple bytecodeSimpleTypeCheck) (runtime.Value, bool) {
	if program == nil || program.frameLayout == nil || program.frameLayout.returnType == nil {
		return val, false
	}
	layout := program.frameLayout
	if layout.returnSimpleType == "void" {
		return runtime.VoidValue{}, true
	}
	if rawVal, ok := bytecodeTryRawProgramReturnNoCoercion(program, instr, val, knownReturnSimple); ok {
		// Return scratch cells are owned by the VM's inline-return path and are
		// immediately copied into the caller's raw stack lane. Keeping that one
		// carrier avoids allocating on nested i64 arithmetic returns. Every
		// other raw value is materialized at this boundary.
		if _, isReturnScratch := rawVal.(*bytecodeRawIntegerReturnScratch); isReturnScratch {
			return rawVal, true
		}
		return vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticReturn, rawVal), true
	}
	if knownReturnSimple != bytecodeSimpleTypeCheckUnknown && knownReturnSimple == layout.returnSimpleCheck {
		return vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticReturn, val), true
	}
	if instr != nil && instr.op == bytecodeOpReturnConstIfIntLessEqualSlotConst && layout.returnSimpleCheck == bytecodeSimpleTypeCheckI32 {
		return vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticReturn, val), true
	}
	if layout.returnSimpleCheck != bytecodeSimpleTypeCheckUnknown {
		val = vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticReturn, val)
		ok := inlineCoercionUnnecessaryBySimpleCheck(layout.returnSimpleCheck, val)
		return val, ok
	}
	if layout.returnExactStructDef != nil {
		val = vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticReturn, val)
		ok := inlineExactNamedStructNoCoercionBytecodeExactDef(layout.returnExactStructDef, val)
		return val, ok
	}
	if i != nil && layout.returnSimpleType != "" && !layout.returnTypeUsesGenerics {
		val = vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticReturn, val)
		ok := inlineCoercionUnnecessaryBySimpleTypeWithInterpreter(i, layout.returnSimpleType, val)
		return val, ok
	}
	return val, false
}

func (vm *bytecodeVM) coerceProgramReturnValue(program *bytecodeProgram, instr *bytecodeInstruction, val runtime.Value, knownReturnSimple bytecodeSimpleTypeCheck, genericNames map[string]struct{}) (runtime.Value, error) {
	if rawVal, ok := bytecodeTryRawProgramReturnNoCoercion(program, instr, val, knownReturnSimple); ok {
		return rawVal, nil
	}
	val = vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticReturn, val)
	if vm == nil || vm.interp == nil || program == nil || program.frameLayout == nil || program.frameLayout.returnType == nil {
		return val, nil
	}
	layout := program.frameLayout
	if layout.returnSimpleType == "void" {
		return runtime.VoidValue{}, nil
	}
	if layout.returnTypeUsesGenerics && len(genericNames) != 0 {
		return val, nil
	}
	noCoercion := false
	if knownReturnSimple != bytecodeSimpleTypeCheckUnknown && knownReturnSimple == layout.returnSimpleCheck {
		noCoercion = true
	} else if instr != nil && instr.op == bytecodeOpReturnConstIfIntLessEqualSlotConst && layout.returnSimpleCheck == bytecodeSimpleTypeCheckI32 {
		noCoercion = true
	} else if layout.returnSimpleCheck != bytecodeSimpleTypeCheckUnknown {
		noCoercion = inlineCoercionUnnecessaryBySimpleCheck(layout.returnSimpleCheck, val)
	} else if layout.returnExactStructDef != nil {
		noCoercion = inlineExactNamedStructNoCoercionBytecodeExactDef(layout.returnExactStructDef, val)
	} else if layout.returnSimpleType != "" {
		noCoercion = inlineCoercionUnnecessaryBySimpleTypeWithInterpreter(vm.interp, layout.returnSimpleType, val)
	} else {
		noCoercion = inlineCoercionUnnecessaryWithInterpreter(vm.interp, layout.returnType, val)
	}
	if noCoercion {
		return val, nil
	}
	if layout.returnNullableSimple != "" {
		if _, isGenericSimple := genericNames[layout.returnNullableSimple]; !isGenericSimple {
			if coerced, ok, err := vm.coerceNullableSimpleProgramReturn(layout.returnNullableSimple, val); ok {
				if err != nil {
					if instr != nil && instr.node != nil {
						err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
					}
					return nil, err
				}
				return coerced, nil
			}
		}
	}
	if layout.returnSimpleType != "" {
		if _, isGenericSimple := genericNames[layout.returnSimpleType]; !isGenericSimple {
			if coerced, ok, err := tryFastSimpleTypeCoercionByName(vm.interp, layout.returnSimpleType, val); ok {
				if err != nil {
					if instr != nil && instr.node != nil {
						err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
					}
					return nil, err
				}
				return coerced, nil
			}
		}
	}
	if !layout.returnTypeUsesGenerics && layout.returnCanonicalType != nil {
		return vm.coerceCachedCanonicalProgramReturnValue(layout.returnCanonicalType, instr, val)
	}
	coerced, err := vm.interp.coerceReturnValue(layout.returnType, val, genericNames, vm.env)
	if err != nil {
		if instr != nil && instr.node != nil {
			err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return nil, err
	}
	return coerced, nil
}

func (vm *bytecodeVM) coerceCachedCanonicalProgramReturnValue(canonical ast.TypeExpression, instr *bytecodeInstruction, val runtime.Value) (runtime.Value, error) {
	if vm == nil || vm.interp == nil || canonical == nil {
		return val, nil
	}
	if isVoidTypeExpr(canonical) {
		return runtime.VoidValue{}, nil
	}
	if isVoidValue(val) {
		if vm.interp.matchesType(canonical, val) || isResultVoidType(canonical) {
			return runtime.VoidValue{}, nil
		}
		expected := typeExpressionToString(canonical)
		err := fmt.Errorf("Return type mismatch: expected %s, got void", expected)
		if instr != nil && instr.node != nil {
			err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return nil, err
	}
	if simple, ok := canonical.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil {
		name := normalizeKernelAliasName(simple.Name.Name)
		if !fastNamedStructTypeNameIsNonNominal(vm.interp, name) {
			if coerced, ok := exactNamedStructCoercionValueForName(val, name); ok {
				return coerced, nil
			}
		}
	}
	if info, ok := parseTypeExpression(canonical); ok && info.name != "" {
		if _, isInterface := vm.interp.interfaces[vm.interp.canonicalInterfaceName(info.name)]; isInterface {
			return vm.interp.coerceValueToTypeInEnv(canonical, val, vm.env)
		}
	}
	if !vm.interp.matchesType(canonical, val) {
		expected := typeExpressionToString(canonical)
		actual := val.Kind().String()
		if actualExpr := vm.interp.typeExpressionFromValue(val); actualExpr != nil {
			actual = typeExpressionToString(actualExpr)
		}
		err := fmt.Errorf("Return type mismatch: expected %s, got %s", expected, actual)
		if instr != nil && instr.node != nil {
			err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return nil, err
	}
	coerced, err := vm.interp.coerceValueToTypeInEnv(canonical, val, vm.env)
	if err != nil {
		if instr != nil && instr.node != nil {
			err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return nil, err
	}
	return coerced, nil
}

func (vm *bytecodeVM) coerceInlineProgramReturnValue(program *bytecodeProgram, instr *bytecodeInstruction, val runtime.Value, knownReturnSimple bytecodeSimpleTypeCheck, genericNames map[string]struct{}) (runtime.Value, error) {
	if bytecodeCanSkipInlineReturnCoercion(program, genericNames) {
		return val, nil
	}
	return vm.coerceProgramReturnValue(program, instr, val, knownReturnSimple, genericNames)
}

func bytecodeCanSkipInlineReturnCoercion(program *bytecodeProgram, genericNames map[string]struct{}) bool {
	if program == nil || program.frameLayout == nil {
		return false
	}
	layout := program.frameLayout
	if layout.returnSimpleType == "void" {
		return false
	}
	if layout.returnType == nil {
		return true
	}
	return layout.returnTypeUsesGenerics && len(genericNames) != 0
}

func (vm *bytecodeVM) coerceNullableSimpleProgramReturn(simpleName string, val runtime.Value) (runtime.Value, bool, error) {
	if simpleName == "" {
		return nil, false, nil
	}
	if isNilRuntimeValue(val) {
		return runtime.NilValue{}, true, nil
	}
	if inlineCoercionUnnecessaryBySimpleTypeWithInterpreter(vm.interp, simpleName, val) {
		return val, true, nil
	}
	if coerced, ok, err := tryFastSimpleTypeCoercionByName(vm.interp, simpleName, val); ok {
		return coerced, true, err
	}
	return nil, false, nil
}

func bytecodeTryRawProgramReturnNoCoercion(program *bytecodeProgram, instr *bytecodeInstruction, val runtime.Value, knownReturnSimple bytecodeSimpleTypeCheck) (runtime.Value, bool) {
	if program == nil || program.frameLayout == nil || program.frameLayout.returnType == nil {
		return nil, false
	}
	if !bytecodeIsRawIntegerCarrier(val) {
		return nil, false
	}
	layout := program.frameLayout
	if layout.returnSimpleType == "void" {
		return runtime.VoidValue{}, true
	}
	if knownReturnSimple != bytecodeSimpleTypeCheckUnknown && knownReturnSimple == layout.returnSimpleCheck {
		if bytecodeRawIntegerReturnMatchesSimpleCheck(layout.returnSimpleCheck, val) {
			return val, true
		}
	}
	if instr != nil && instr.op == bytecodeOpReturnConstIfIntLessEqualSlotConst && layout.returnSimpleCheck == bytecodeSimpleTypeCheckI32 {
		if bytecodeRawIntegerReturnMatchesSimpleCheck(layout.returnSimpleCheck, val) {
			return val, true
		}
	}
	if layout.returnSimpleCheck != bytecodeSimpleTypeCheckUnknown {
		if bytecodeRawIntegerReturnMatchesSimpleCheck(layout.returnSimpleCheck, val) {
			return val, true
		}
	}
	return nil, false
}

func bytecodeRawIntegerReturnMatchesSimpleCheck(check bytecodeSimpleTypeCheck, val runtime.Value) bool {
	if !bytecodeIsRawIntegerCarrier(val) {
		return false
	}
	if check == bytecodeSimpleTypeCheckAnyInteger {
		_, _, ok := bytecodeRawIntegerValueInfo(val)
		return ok
	}
	if _, ok := check.integerType(); ok {
		return inlineCoercionUnnecessaryBySimpleCheck(check, val)
	}
	return false
}

func (vm *bytecodeVM) appendReturnValue(value runtime.Value) {
	if bytecodeIsRawIntegerCarrier(value) {
		if kind, raw, ok := bytecodeRawIntegerValueInfo(value); ok {
			vm.appendRawIntegerStack(kind, raw)
			return
		}
	}
	vm.appendStackValue(value)
}

func (vm *bytecodeVM) finishInlineReturn(program **bytecodeProgram, instructions *[]bytecodeInstruction, validatedIntConsts *[]bool, slotConstIntImmTable **bytecodeSlotConstIntImmediateTable, instr *bytecodeInstruction, val runtime.Value, knownReturnSimple bytecodeSimpleTypeCheck) error {
	activeProgram := *program

	if vm != nil && vm.selfFastMinimalSuffix > 0 {
		if fastVal, ok := bytecodeTryMaterializedProgramReturnNoCoercion(vm, vm.interp, activeProgram, instr, val, knownReturnSimple); ok {
			val = fastVal
		} else {
			var err error
			val, err = vm.coerceProgramReturnValue(activeProgram, instr, val, knownReturnSimple, nil)
			if err != nil {
				return err
			}
		}

		if len(vm.selfFastMinimal) == 0 {
			return fmt.Errorf("bytecode call frame underflow")
		}
		idx := len(vm.selfFastMinimal) - 1
		frame := &vm.selfFastMinimal[idx]
		vm.finishBytecodeArrayOwnershipReturn(val, frame.arrayOwnershipParent)
		returnIP := frame.returnIP
		stackBase := frame.stackBase
		returnSlots := frame.slots
		returnEnv := frame.env
		if frame.transientScopeBase < len(vm.activeTransientScopeEnvs) {
			vm.releaseActiveTransientRuntimeScopeEnvsToBase(frame.transientScopeBase)
		}
		reusesSlots := frame.reusesSlots
		iterBase := frame.iterBase
		loopBase := frame.loopBase
		calleeImplicitSlotActive := vm.detachImplicitSlotActiveFrame()
		implicitSlotActive := frame.implicitSlotActive
		i32Program, i32Registers, i32Valid := frame.i32RegisterProgram, frame.i32Registers, frame.i32RegisterValid
		slotI32Values, slotI32Valid := frame.slotI32Values, frame.slotI32Valid
		slotFloatValues, slotFloatKinds, slotFloatValid := frame.slotFloatValues, frame.slotFloatKinds, frame.slotFloatValid
		frame.i32RegisterProgram, frame.i32Registers, frame.i32RegisterValid = nil, nil, nil
		frame.implicitSlotActive = nil
		frame.slotI32Values, frame.slotI32Valid = nil, nil
		frame.slotFloatValues, frame.slotFloatKinds, frame.slotFloatValid = nil, nil, nil
		frame.env = nil
		frame.transientScopeBase = 0
		frame.arrayOwnershipParent = nil
		frame.iterBase = 0
		frame.loopBase = 0
		vm.selfFastMinimal = vm.selfFastMinimal[:idx]
		vm.selfFastMinimalSuffix--
		calleeSlots := vm.slots
		vm.releaseImplicitSlotActiveFrame(calleeImplicitSlotActive)
		vm.restoreImplicitSlotActiveFrame(implicitSlotActive)
		vm.restoreI32RegisterFrame(i32Program, i32Registers, i32Valid)
		if !reusesSlots {
			vm.restoreValueSlotSidecarFrames(returnSlots, slotI32Values, slotI32Valid, slotFloatValues, slotFloatKinds, slotFloatValid)
		}
		if len(vm.iterStack) != iterBase || len(vm.loopStack) != loopBase {
			vm.restoreCallFrameControlStacks(iterBase, loopBase)
		}
		vm.ip = returnIP
		vm.env = returnEnv
		vm.slots = returnSlots
		if reusesSlots {
			vm.restoreSelfFastMinimalFrameSlot0(frame, returnSlots)
		} else {
			switch len(calleeSlots) {
			case 2:
				vm.releaseSlotFrame2(calleeSlots)
			case 4:
				vm.releaseSlotFrame4(calleeSlots)
			default:
				vm.releaseSlotFrame(calleeSlots)
			}
		}
		vm.interp.recordBytecodeInlineFrameBalance(instr, stackBase, vm.stackDepth())
		vm.completeBytecodeInlineCallOperandRegion(vm.stackDepth() + 1)
		vm.appendReturnValue(val)
		return nil
	}

	var err error
	if fastVal, ok := bytecodeTryMaterializedProgramReturnNoCoercion(vm, vm.interp, activeProgram, instr, val, knownReturnSimple); ok {
		val = fastVal
	} else {
		if returnFn := vm.peekReturnCoercionFunction(); returnFn != nil && (activeProgram == nil || activeProgram.frameLayout == nil) {
			val, err = vm.coerceSlotlessInlineReturnValue(returnFn, activeProgram, val)
		} else {
			returnGenericNames := vm.peekReturnGenericNames()
			val, err = vm.coerceInlineProgramReturnValue(activeProgram, instr, val, knownReturnSimple, returnGenericNames)
		}
	}
	if err != nil {
		return err
	}

	stackBase := vm.topInlineFrameStackBase()
	vm.finishBytecodeArrayOwnershipReturn(val, vm.topBytecodeArrayOwnershipParent())
	returnIP, returnProgram, returnSlots, returnEnv, iterBase, loopBase, hasImplicitReceiver, selfFast, activeLookup, ok := vm.popCallFrameFields()
	if !ok {
		return fmt.Errorf("bytecode call frame underflow")
	}
	calleeSlots := vm.slots
	if hasImplicitReceiver {
		state := vm.interp.stateFromEnv(vm.env)
		state.popImplicitReceiver()
	}
	vm.ip = returnIP
	vm.slots = returnSlots
	vm.env = returnEnv
	if !selfFast {
		vm.switchRunProgramWithActiveLookupState(program, instructions, validatedIntConsts, slotConstIntImmTable, returnProgram, activeLookup)
	}
	if len(vm.iterStack) != iterBase || len(vm.loopStack) != loopBase {
		vm.restoreCallFrameControlStacks(iterBase, loopBase)
	}
	if !sameSlotFrame(calleeSlots, returnSlots) {
		switch len(calleeSlots) {
		case 2:
			vm.releaseSlotFrame2(calleeSlots)
		case 4:
			vm.releaseSlotFrame4(calleeSlots)
		default:
			vm.releaseSlotFrame(calleeSlots)
		}
	}
	vm.interp.recordBytecodeInlineFrameBalance(instr, stackBase, vm.stackDepth())
	vm.completeBytecodeInlineCallOperandRegion(vm.stackDepth() + 1)
	vm.appendReturnValue(val)
	return nil
}

func bytecodeCanFinishMinimalReturnNoCoerce(program *bytecodeProgram, instr *bytecodeInstruction, knownReturnSimple bytecodeSimpleTypeCheck) bool {
	if program == nil || instr == nil {
		return false
	}
	layout := program.frameLayout
	if layout == nil || layout.returnSimpleCheck != bytecodeSimpleTypeCheckI32 {
		return false
	}
	if knownReturnSimple == bytecodeSimpleTypeCheckI32 {
		return true
	}
	switch instr.op {
	case bytecodeOpReturnConstIfIntLessEqualSlotConst:
		return true
	case bytecodeOpReturnIfIntLessEqualSlotConst:
		return instr.target >= 0 && instr.target < len(layout.slotKinds) && layout.slotKinds[instr.target] == bytecodeCellKindI32
	default:
		return false
	}
}

func (vm *bytecodeVM) finishMinimalSelfFastReturnNoCoerce(val runtime.Value) bool {
	if vm == nil || vm.selfFastMinimalSuffix <= 0 || len(vm.selfFastMinimal) == 0 {
		return false
	}
	val = vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticReturn, val)
	idx := len(vm.selfFastMinimal) - 1
	frame := &vm.selfFastMinimal[idx]
	if !frame.reusesSlots || len(frame.slots) == 0 {
		return false
	}
	returnIP := frame.returnIP
	returnSlots := frame.slots
	returnEnv := frame.env
	vm.finishBytecodeArrayOwnershipReturn(val, frame.arrayOwnershipParent)
	if frame.transientScopeBase < len(vm.activeTransientScopeEnvs) {
		vm.releaseActiveTransientRuntimeScopeEnvsToBase(frame.transientScopeBase)
	}
	iterBase := frame.iterBase
	loopBase := frame.loopBase
	calleeImplicitSlotActive := vm.detachImplicitSlotActiveFrame()
	implicitSlotActive := frame.implicitSlotActive
	i32Program, i32Registers, i32Valid := frame.i32RegisterProgram, frame.i32Registers, frame.i32RegisterValid
	frame.slotFloatValues, frame.slotFloatKinds, frame.slotFloatValid = nil, nil, nil
	frame.i32RegisterProgram, frame.i32Registers, frame.i32RegisterValid = nil, nil, nil
	frame.implicitSlotActive = nil
	frame.env = nil
	frame.transientScopeBase = 0
	frame.arrayOwnershipParent = nil
	frame.iterBase = 0
	frame.loopBase = 0
	vm.selfFastMinimal = vm.selfFastMinimal[:idx]
	vm.selfFastMinimalSuffix--
	vm.releaseImplicitSlotActiveFrame(calleeImplicitSlotActive)
	vm.restoreImplicitSlotActiveFrame(implicitSlotActive)
	vm.restoreI32RegisterFrame(i32Program, i32Registers, i32Valid)
	if len(vm.iterStack) != iterBase || len(vm.loopStack) != loopBase {
		vm.restoreCallFrameControlStacks(iterBase, loopBase)
	}
	vm.ip = returnIP
	vm.env = returnEnv
	vm.slots = returnSlots
	vm.restoreSelfFastMinimalFrameSlot0(frame, returnSlots)
	vm.appendStackValue(val)
	return true
}

func (vm *bytecodeVM) tryFinishMinimalSelfFastReturnNoCoerce(program *bytecodeProgram, instr *bytecodeInstruction, val runtime.Value, knownReturnSimple bytecodeSimpleTypeCheck) bool {
	if !bytecodeCanFinishMinimalReturnNoCoerce(program, instr, knownReturnSimple) {
		return false
	}
	return vm.finishMinimalSelfFastReturnNoCoerce(val)
}

func (vm *bytecodeVM) execReturnBinaryIntAdd(instr *bytecodeInstruction) (runtime.Value, bytecodeSimpleTypeCheck, error) {
	if vm.stackDepth() < 2 {
		return nil, bytecodeSimpleTypeCheckUnknown, fmt.Errorf("bytecode stack underflow")
	}
	rightIdx := vm.stackDepth() - 1
	leftIdx := rightIdx - 1
	right := vm.stackValue(rightIdx)
	left := vm.stackValue(leftIdx)
	if instr.op == bytecodeOpReturnBinaryIntAddI32 {
		if lv, ok := left.(runtime.IntegerValue); ok && lv.TypeSuffix == runtime.IntegerI32 {
			if rv, ok := right.(runtime.IntegerValue); ok && rv.TypeSuffix == runtime.IntegerI32 {
				lvRef := &lv
				rvRef := &rv
				if lvRef.IsSmallRef() && rvRef.IsSmallRef() {
					l := lvRef.Int64FastRef()
					r := rvRef.Int64FastRef()
					if l >= math.MinInt32 && l <= math.MaxInt32 && r >= math.MinInt32 && r <= math.MaxInt32 {
						sum := l + r
						vm.truncateStack(leftIdx)
						if sum < math.MinInt32 || sum > math.MaxInt32 {
							return nil, bytecodeSimpleTypeCheckI32, newOverflowError("integer overflow")
						}
						return bytecodeRawI32ResultValue(sum), bytecodeSimpleTypeCheckI32, nil
					}
				}
			}
		}
		if val, handled, err := bytecodeAddSmallI32PairFast(left, right); handled {
			vm.truncateStack(leftIdx)
			return val, bytecodeSimpleTypeCheckI32, err
		}
	}
	if val, handled, err := vm.returnBinaryIntegerArithmeticRaw("+", left, right); handled {
		vm.truncateStack(leftIdx)
		return val, bytecodeSimpleTypeCheckUnknown, err
	}
	val, handled, err := vm.execBinarySpecializedOpcode(instr, left, right)
	if !handled && err == nil {
		err = fmt.Errorf("bytecode return-add opcode missing add handler")
	}
	if err != nil {
		return nil, bytecodeSimpleTypeCheckUnknown, err
	}
	vm.truncateStack(leftIdx)
	return val, bytecodeSimpleTypeCheckUnknown, nil
}

func (vm *bytecodeVM) execReturnBinary(instr *bytecodeInstruction) (runtime.Value, bytecodeSimpleTypeCheck, error) {
	if vm.stackDepth() < 2 {
		return nil, bytecodeSimpleTypeCheckUnknown, fmt.Errorf("bytecode stack underflow")
	}
	rightIdx := vm.stackDepth() - 1
	leftIdx := rightIdx - 1
	right := vm.stackValue(rightIdx)
	left := vm.stackValue(leftIdx)
	operator := instr.operator
	if fast, handled := execBinaryDirectIntegerComparisonFast(operator, left, right); handled {
		vm.truncateStack(leftIdx)
		return fast, bytecodeReturnBinaryKnownSimpleCheck(operator), nil
	}
	if fast, handled := bytecodeDirectFloatCompareFast(operator, left, right); handled {
		vm.truncateStack(leftIdx)
		return fast, bytecodeReturnBinaryKnownSimpleCheck(operator), nil
	}
	if raw, kind, handled := bytecodeDirectFloatArithmeticRawValue(operator, left, right); handled {
		vm.truncateStack(leftIdx)
		return bytecodeRawFloatSlotValue(raw, kind), bytecodeSimpleTypeCheckUnknown, nil
	}
	if fast, handled, err := vm.returnBinaryIntegerArithmeticRaw(operator, left, right); handled {
		if err != nil {
			return nil, bytecodeSimpleTypeCheckUnknown, err
		}
		vm.truncateStack(leftIdx)
		return fast, bytecodeReturnBinaryKnownSimpleCheck(operator), nil
	}
	if isBytecodeBinaryFastPathCandidate(operator) {
		if fast, handled, err := ApplyBinaryOperatorFast(operator, left, right); handled {
			if err != nil {
				return nil, bytecodeSimpleTypeCheckUnknown, err
			}
			vm.truncateStack(leftIdx)
			return fast, bytecodeReturnBinaryKnownSimpleCheck(operator), nil
		}
	}
	result, err := applyBinaryOperator(vm.interp, operator, left, right)
	if err != nil {
		return nil, bytecodeSimpleTypeCheckUnknown, err
	}
	vm.truncateStack(leftIdx)
	return result, bytecodeReturnBinaryKnownSimpleCheck(operator), nil
}

func (vm *bytecodeVM) returnBinaryIntegerArithmeticRaw(operator string, left runtime.Value, right runtime.Value) (runtime.Value, bool, error) {
	normalized, dotted := normalizeOperator(operator)
	switch normalized {
	case "+", "-", "*":
	case "^":
		if dotted {
			return nil, false, nil
		}
	default:
		return nil, false, nil
	}
	if leftInt, ok := bytecodeDirectIntegerValue(left); ok {
		if rightInt, ok := bytecodeDirectIntegerValue(right); ok {
			return vm.returnIntegerArithmeticRaw(normalized, leftInt, rightInt)
		}
	}
	if leftInt, ok := bytecodeIntegerValue(left); ok {
		if rightInt, ok := bytecodeIntegerValue(right); ok {
			return vm.returnIntegerArithmeticRaw(normalized, leftInt, rightInt)
		}
	}
	return nil, false, nil
}

func (vm *bytecodeVM) returnIntegerArithmeticRaw(operator string, left runtime.IntegerValue, right runtime.IntegerValue) (runtime.Value, bool, error) {
	kind, raw, ok, err := evaluateIntegerArithmeticRawFast(operator, left, right)
	if err != nil {
		return nil, true, err
	}
	if !ok {
		return nil, false, nil
	}
	return vm.rawIntegerReturnValue(kind, raw), true, nil
}

func bytecodeReturnBinaryKnownSimpleCheck(operator string) bytecodeSimpleTypeCheck {
	normalized, _ := normalizeOperator(operator)
	switch normalized {
	case "<", "<=", ">", ">=", "==", "!=":
		return bytecodeSimpleTypeCheckBool
	default:
		return bytecodeSimpleTypeCheckUnknown
	}
}
