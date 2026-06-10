package interpreter

import (
	"fmt"
	"math"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type bytecodeI32RecurrenceKernel struct {
	baseLimit        int64
	basePrefix       []bytecodeRecurrenceBaseValue
	baseRangeEnabled bool
	baseRangeValue   bytecodeRecurrenceBaseValue
	firstSub         int64
	secondSub        int64
	firstSubKind     runtime.IntegerType
	secondSubKind    runtime.IntegerType
	overflowAST      ast.Node
}

type bytecodeRecurrenceBaseMatch struct {
	baseLimit        int64
	basePrefix       []bytecodeRecurrenceBaseValue
	baseRangeEnabled bool
	baseRangeValue   bytecodeRecurrenceBaseValue
	firstCallIndex   int
}

const bytecodeI32RecurrenceDPMaxInput = 1 << 20
const bytecodeI32RecurrenceGuardPrefixMaxLimit = 4096

type bytecodeRecurrenceBaseGuard struct {
	operator        string
	raw             int64
	baseLimit       int64
	returnValue     int64
	returnCurrent   bool
	returnKind      runtime.IntegerType
	hasNegativeSpan bool
}

func bytecodeDetectI32RecurrenceKernel(program *bytecodeProgram) *bytecodeI32RecurrenceKernel {
	if program == nil || program.frameLayout == nil {
		return nil
	}
	layout := program.frameLayout
	if !layout.selfCallOneArgFast || layout.paramSlots != 1 || layout.slotCount != 2 || layout.selfCallSlot != 1 {
		return nil
	}
	if layout.usesImplicitMember || layout.needsEnvScopes {
		return nil
	}
	if !bytecodeRecurrenceReturnCheckEligible(layout.returnSimpleCheck) {
		return nil
	}
	if !bytecodeRecurrenceEntryAcceptsSupportedInteger(layout) {
		return nil
	}
	if layout.returnSimpleCheck != bytecodeSimpleTypeCheckAnyInteger &&
		len(layout.paramSimpleChecks) > 0 &&
		layout.paramSimpleChecks[0] != layout.returnSimpleCheck {
		return nil
	}
	instructions := program.instructions
	baseMatch, ok := bytecodeDetectI32RecurrenceBase(instructions, layout)
	if !ok {
		return nil
	}
	if len(instructions) != baseMatch.firstCallIndex+4 {
		return nil
	}
	firstCall := instructions[baseMatch.firstCallIndex]
	secondCall := instructions[baseMatch.firstCallIndex+1]
	retAdd := instructions[baseMatch.firstCallIndex+2]
	finalRet := instructions[baseMatch.firstCallIndex+3]
	if retAdd.op != bytecodeOpReturnBinaryIntAddI32 &&
		retAdd.op != bytecodeOpReturnBinaryIntAdd ||
		firstCall.op != bytecodeOpCallSelfIntSubSlotConst ||
		secondCall.op != bytecodeOpCallSelfIntSubSlotConst ||
		finalRet.op != bytecodeOpReturn {
		return nil
	}
	if !bytecodeRecurrenceCallShape(firstCall, layout.selfCallSlot) ||
		!bytecodeRecurrenceCallShape(secondCall, layout.selfCallSlot) {
		return nil
	}
	if !bytecodeRecurrenceBaseSupportsCalls(baseMatch, firstCall.intImmediateRaw, secondCall.intImmediateRaw) {
		return nil
	}
	return &bytecodeI32RecurrenceKernel{
		baseLimit:        baseMatch.baseLimit,
		basePrefix:       baseMatch.basePrefix,
		baseRangeEnabled: baseMatch.baseRangeEnabled,
		baseRangeValue:   baseMatch.baseRangeValue,
		firstSub:         firstCall.intImmediateRaw,
		secondSub:        secondCall.intImmediateRaw,
		firstSubKind:     bytecodeRecurrenceInstructionLiteralKind(firstCall),
		secondSubKind:    bytecodeRecurrenceInstructionLiteralKind(secondCall),
		overflowAST:      retAdd.node,
	}
}

func bytecodeDetectI32RecurrenceBase(instructions []bytecodeInstruction, layout *bytecodeFrameLayout) (bytecodeRecurrenceBaseMatch, bool) {
	if match, ok := bytecodeDetectI32RecurrenceGuardPrefixBase(instructions, layout); ok {
		return match, true
	}
	return bytecodeDetectI32RecurrenceFusedBase(instructions, layout)
}

func bytecodeDetectI32RecurrenceGuardPrefixBase(instructions []bytecodeInstruction, layout *bytecodeFrameLayout) (bytecodeRecurrenceBaseMatch, bool) {
	if layout == nil || len(instructions) < 6 {
		return bytecodeRecurrenceBaseMatch{}, false
	}
	guards := make([]bytecodeRecurrenceBaseGuard, 0, 2)
	for idx := 0; idx < len(instructions); idx++ {
		guard, ok, isBase := bytecodeParseRecurrenceBaseGuard(instructions[idx], layout)
		if !isBase {
			break
		}
		if !ok {
			return bytecodeRecurrenceBaseMatch{}, false
		}
		guards = append(guards, guard)
	}
	if len(guards) < 2 {
		return bytecodeRecurrenceBaseMatch{}, false
	}
	maxPositiveLimit, ok := bytecodeRecurrenceGuardMaxNonNegativeLimit(guards)
	if !ok || maxPositiveLimit > bytecodeI32RecurrenceGuardPrefixMaxLimit {
		return bytecodeRecurrenceBaseMatch{}, false
	}
	basePrefix, ok := bytecodeRecurrenceBuildGuardPrefix(guards, maxPositiveLimit)
	if !ok {
		return bytecodeRecurrenceBaseMatch{}, false
	}
	match := bytecodeRecurrenceBaseMatch{
		baseLimit:      int64(len(basePrefix) - 1),
		basePrefix:     basePrefix,
		firstCallIndex: len(guards),
	}
	if rangeGuard, ok := bytecodeRecurrenceFirstNegativeSpanningRangeGuard(guards); ok {
		match.baseLimit = rangeGuard.baseLimit
		match.baseRangeEnabled = true
		match.baseRangeValue = rangeGuard.baseValue(rangeGuard.baseLimit)
	}
	return match, true
}

func bytecodeDetectI32RecurrenceFusedBase(instructions []bytecodeInstruction, layout *bytecodeFrameLayout) (bytecodeRecurrenceBaseMatch, bool) {
	if layout == nil || len(instructions) < 5 {
		return bytecodeRecurrenceBaseMatch{}, false
	}
	return bytecodeDetectI32RecurrenceFusedBaseInstruction(instructions[0], layout)
}

func bytecodeDetectI32RecurrenceFusedBaseInstruction(base bytecodeInstruction, layout *bytecodeFrameLayout) (bytecodeRecurrenceBaseMatch, bool) {
	if layout == nil {
		return bytecodeRecurrenceBaseMatch{}, false
	}
	if base.op != bytecodeOpReturnConstIfIntLessEqualSlotConst &&
		base.op != bytecodeOpReturnIfIntLessEqualSlotConst {
		return bytecodeRecurrenceBaseMatch{}, false
	}
	if base.argCount != 0 || !base.hasIntRaw {
		return bytecodeRecurrenceBaseMatch{}, false
	}
	baseLimit, ok := bytecodeRecurrenceBaseLimitForOperator(base.operator, base.intImmediateRaw)
	if !ok {
		return bytecodeRecurrenceBaseMatch{}, false
	}
	match := bytecodeRecurrenceBaseMatch{
		baseLimit:        baseLimit,
		baseRangeEnabled: true,
		firstCallIndex:   1,
	}
	switch base.op {
	case bytecodeOpReturnConstIfIntLessEqualSlotConst:
		if base.target != -1 {
			return bytecodeRecurrenceBaseMatch{}, false
		}
		rawBaseReturn, returnKind, ok := bytecodeRecurrenceBaseReturnValue(base.value, layout.returnSimpleCheck)
		if !ok {
			return bytecodeRecurrenceBaseMatch{}, false
		}
		match.baseRangeValue = bytecodeRecurrenceBaseValue{raw: rawBaseReturn, literalKind: returnKind}
	case bytecodeOpReturnIfIntLessEqualSlotConst:
		if base.target != 0 {
			return bytecodeRecurrenceBaseMatch{}, false
		}
		match.baseRangeValue = bytecodeRecurrenceBaseValue{current: true}
	}
	return match, true
}

func bytecodeParseRecurrenceBaseGuard(base bytecodeInstruction, layout *bytecodeFrameLayout) (bytecodeRecurrenceBaseGuard, bool, bool) {
	if layout == nil {
		return bytecodeRecurrenceBaseGuard{}, false, false
	}
	if base.op != bytecodeOpReturnConstIfIntLessEqualSlotConst &&
		base.op != bytecodeOpReturnIfIntLessEqualSlotConst {
		return bytecodeRecurrenceBaseGuard{}, false, false
	}
	if base.argCount != 0 || !base.hasIntRaw {
		return bytecodeRecurrenceBaseGuard{}, false, true
	}
	guard := bytecodeRecurrenceBaseGuard{
		operator: base.operator,
		raw:      base.intImmediateRaw,
	}
	switch base.operator {
	case "==":
		guard.baseLimit = base.intImmediateRaw
	case "", "<=", "<":
		baseLimit, ok := bytecodeRecurrenceBaseLimitForOperator(base.operator, base.intImmediateRaw)
		if !ok {
			return bytecodeRecurrenceBaseGuard{}, false, true
		}
		guard.baseLimit = baseLimit
		guard.hasNegativeSpan = baseLimit >= -1
	default:
		return bytecodeRecurrenceBaseGuard{}, false, true
	}
	switch base.op {
	case bytecodeOpReturnConstIfIntLessEqualSlotConst:
		if base.target != -1 {
			return bytecodeRecurrenceBaseGuard{}, false, true
		}
		rawBaseReturn, returnKind, ok := bytecodeRecurrenceBaseReturnValue(base.value, layout.returnSimpleCheck)
		if !ok {
			return bytecodeRecurrenceBaseGuard{}, false, true
		}
		guard.returnValue = rawBaseReturn
		guard.returnKind = returnKind
	case bytecodeOpReturnIfIntLessEqualSlotConst:
		if base.target != 0 {
			return bytecodeRecurrenceBaseGuard{}, false, true
		}
		guard.returnCurrent = true
	}
	return guard, true, true
}

func (g bytecodeRecurrenceBaseGuard) matches(n int64) bool {
	switch g.operator {
	case "==":
		return n == g.raw
	default:
		return n <= g.baseLimit
	}
}

func (g bytecodeRecurrenceBaseGuard) value(n int64) int64 {
	if g.returnCurrent {
		return n
	}
	return g.returnValue
}

func (g bytecodeRecurrenceBaseGuard) nonNegativeMax() (int64, bool) {
	switch g.operator {
	case "==":
		if g.raw < 0 {
			return 0, false
		}
		return g.raw, true
	default:
		if g.baseLimit < 0 {
			return 0, false
		}
		return g.baseLimit, true
	}
}

func bytecodeRecurrenceGuardMaxNonNegativeLimit(guards []bytecodeRecurrenceBaseGuard) (int64, bool) {
	maxPositiveLimit := int64(-1)
	for _, guard := range guards {
		limit, ok := guard.nonNegativeMax()
		if ok && limit > maxPositiveLimit {
			maxPositiveLimit = limit
		}
	}
	if maxPositiveLimit < 0 {
		return 0, false
	}
	return maxPositiveLimit, true
}

func bytecodeRecurrenceBuildGuardPrefix(guards []bytecodeRecurrenceBaseGuard, maxPositiveLimit int64) ([]bytecodeRecurrenceBaseValue, bool) {
	if len(guards) == 0 || maxPositiveLimit < 0 {
		return nil, false
	}
	basePrefix := make([]bytecodeRecurrenceBaseValue, 0, maxPositiveLimit+1)
	foundGap := false
	for cur := int64(0); cur <= maxPositiveLimit; cur++ {
		value, ok := bytecodeRecurrenceGuardValueAt(guards, cur)
		if !ok {
			foundGap = true
			continue
		}
		if foundGap {
			return nil, false
		}
		basePrefix = append(basePrefix, value)
	}
	if len(basePrefix) == 0 {
		return nil, false
	}
	return basePrefix, true
}

func bytecodeRecurrenceGuardValueAt(guards []bytecodeRecurrenceBaseGuard, n int64) (bytecodeRecurrenceBaseValue, bool) {
	for _, guard := range guards {
		if guard.matches(n) {
			return guard.baseValue(n), true
		}
	}
	return bytecodeRecurrenceBaseValue{}, false
}

func bytecodeRecurrenceFirstNegativeSpanningRangeGuard(guards []bytecodeRecurrenceBaseGuard) (bytecodeRecurrenceBaseGuard, bool) {
	for _, guard := range guards {
		if guard.operator == "==" {
			if guard.raw < 0 {
				return bytecodeRecurrenceBaseGuard{}, false
			}
			continue
		}
		if !guard.hasNegativeSpan {
			return bytecodeRecurrenceBaseGuard{}, false
		}
		return guard, true
	}
	return bytecodeRecurrenceBaseGuard{}, false
}

func bytecodeRecurrenceBaseSupportsCalls(match bytecodeRecurrenceBaseMatch, firstSub int64, secondSub int64) bool {
	if firstSub <= 0 || secondSub <= 0 {
		return false
	}
	baseMax := match.nonNegativeBaseLimit()
	if baseMax < 0 {
		return false
	}
	maxSub := firstSub
	if secondSub > maxSub {
		maxSub = secondSub
	}
	supportLimit := baseMax + maxSub
	if supportLimit < baseMax {
		return false
	}
	supported := make([]bool, int(supportLimit)+1)
	for cur := int64(0); cur <= supportLimit; cur++ {
		if match.hasBaseValue(cur) {
			supported[int(cur)] = true
			continue
		}
		left := cur - firstSub
		right := cur - secondSub
		if !bytecodeRecurrenceBaseTermSupported(match, supported, left) ||
			!bytecodeRecurrenceBaseTermSupported(match, supported, right) {
			return false
		}
		supported[int(cur)] = true
	}
	return true
}

func bytecodeRecurrenceBaseTermSupported(match bytecodeRecurrenceBaseMatch, supported []bool, n int64) bool {
	if match.hasBaseValue(n) {
		return true
	}
	if n < 0 || n >= int64(len(supported)) {
		return false
	}
	return supported[int(n)]
}

func (m bytecodeRecurrenceBaseMatch) hasBaseValue(n int64) bool {
	if len(m.basePrefix) > 0 {
		if n >= 0 && n < int64(len(m.basePrefix)) {
			return true
		}
		if m.baseRangeEnabled {
			return n <= m.baseLimit
		}
		return n >= 0 && n <= m.baseLimit
	}
	return n <= m.baseLimit
}

func (m bytecodeRecurrenceBaseMatch) nonNegativeBaseLimit() int64 {
	baseMax := int64(len(m.basePrefix) - 1)
	if m.baseRangeEnabled && m.baseLimit > baseMax {
		return m.baseLimit
	}
	return baseMax
}

func (m bytecodeRecurrenceBaseMatch) baseValuesFitKind(kind runtime.IntegerType) bool {
	if kind == "" {
		return true
	}
	for _, base := range m.basePrefix {
		if err := ensureFitsInt64Type(kind, base.raw); err != nil {
			return false
		}
	}
	if (!m.baseRangeValue.current) && (m.baseRangeEnabled || len(m.basePrefix) == 0) {
		if err := ensureFitsInt64Type(kind, m.baseRangeValue.raw); err != nil {
			return false
		}
	}
	return true
}

func bytecodeRecurrenceReturnCheckEligible(check bytecodeSimpleTypeCheck) bool {
	switch check {
	case bytecodeSimpleTypeCheckAnyInteger,
		bytecodeSimpleTypeCheckI8,
		bytecodeSimpleTypeCheckI16,
		bytecodeSimpleTypeCheckI32,
		bytecodeSimpleTypeCheckI64,
		bytecodeSimpleTypeCheckI128,
		bytecodeSimpleTypeCheckIsize,
		bytecodeSimpleTypeCheckU8,
		bytecodeSimpleTypeCheckU16,
		bytecodeSimpleTypeCheckU32,
		bytecodeSimpleTypeCheckU64,
		bytecodeSimpleTypeCheckU128,
		bytecodeSimpleTypeCheckUsize:
		return true
	default:
		return false
	}
}

func bytecodeRecurrenceEntryAcceptsSupportedInteger(layout *bytecodeFrameLayout) bool {
	if layout == nil || len(layout.slotKinds) < 2 || len(layout.paramSimpleChecks) == 0 {
		return false
	}
	if layout.slotKinds[0] == bytecodeCellKindI32 &&
		(layout.paramSimpleChecks[0] == bytecodeSimpleTypeCheckI32 || layout.paramSimpleChecks[0] == bytecodeSimpleTypeCheckAnyInteger) {
		return true
	}
	return layout.slotKinds[0] == bytecodeCellKindValue &&
		bytecodeRecurrenceReturnCheckEligible(layout.paramSimpleChecks[0])
}

func bytecodeRecurrenceBaseReturnValue(value runtime.Value, check bytecodeSimpleTypeCheck) (int64, runtime.IntegerType, bool) {
	intVal, ok := bytecodeDirectIntegerValue(value)
	if !ok {
		return 0, "", false
	}
	ref := &intVal
	if !ref.IsSmallRef() {
		return 0, "", false
	}
	return bytecodeRecurrenceBaseReturnRaw(ref.Int64FastRef(), intVal.TypeSuffix, check)
}

func bytecodeRecurrenceBaseReturnRaw(raw int64, literalKind runtime.IntegerType, check bytecodeSimpleTypeCheck) (int64, runtime.IntegerType, bool) {
	if _, ok := lookupIntegerInfo(literalKind); !ok {
		return 0, "", false
	}
	if err := ensureFitsInt64Type(literalKind, raw); err != nil {
		return 0, "", false
	}
	if check == bytecodeSimpleTypeCheckAnyInteger {
		return raw, literalKind, true
	}
	kind, ok := check.integerType()
	if !ok {
		return 0, "", false
	}
	return raw, kind, ensureFitsInt64Type(kind, raw) == nil
}

func bytecodeRecurrenceBaseLimitForOperator(operator string, raw int64) (int64, bool) {
	switch operator {
	case "", "<=":
		return raw, true
	case "<":
		limit, overflow := subInt64Overflow(raw, 1)
		if overflow {
			return 0, false
		}
		return limit, true
	default:
		return 0, false
	}
}

func bytecodeRecurrenceInstructionLiteralKind(instr bytecodeInstruction) runtime.IntegerType {
	if instr.hasIntImmediate {
		return instr.intImmediate.TypeSuffix
	}
	intVal, ok := bytecodeImmediateIntegerValue(instr.value)
	if !ok {
		return ""
	}
	return intVal.TypeSuffix
}

func bytecodeRecurrenceExecutionValue(layout *bytecodeFrameLayout, value runtime.Value) (runtime.IntegerType, int64, bool) {
	if layout == nil {
		return "", 0, false
	}
	intVal, ok := bytecodeDirectIntegerValue(value)
	if !ok {
		return "", 0, false
	}
	ref := &intVal
	if !ref.IsSmallRef() {
		return "", 0, false
	}
	switch layout.returnSimpleCheck {
	case bytecodeSimpleTypeCheckAnyInteger:
		switch intVal.TypeSuffix {
		case runtime.IntegerI8,
			runtime.IntegerI16,
			runtime.IntegerI32,
			runtime.IntegerI64,
			runtime.IntegerI128,
			runtime.IntegerIsize,
			runtime.IntegerU8,
			runtime.IntegerU16,
			runtime.IntegerU32,
			runtime.IntegerU64,
			runtime.IntegerU128,
			runtime.IntegerUsize:
			return intVal.TypeSuffix, ref.Int64FastRef(), true
		default:
			return "", 0, false
		}
	default:
		kind, ok := layout.returnSimpleCheck.integerType()
		if !ok || intVal.TypeSuffix != kind {
			return "", 0, false
		}
		return kind, ref.Int64FastRef(), true
	}
}

func bytecodeRecurrenceSimpleCheckForKind(kind runtime.IntegerType) bytecodeSimpleTypeCheck {
	switch kind {
	case runtime.IntegerI8:
		return bytecodeSimpleTypeCheckI8
	case runtime.IntegerI16:
		return bytecodeSimpleTypeCheckI16
	case runtime.IntegerI32:
		return bytecodeSimpleTypeCheckI32
	case runtime.IntegerI64:
		return bytecodeSimpleTypeCheckI64
	case runtime.IntegerI128:
		return bytecodeSimpleTypeCheckI128
	case runtime.IntegerIsize:
		return bytecodeSimpleTypeCheckIsize
	case runtime.IntegerU8:
		return bytecodeSimpleTypeCheckU8
	case runtime.IntegerU16:
		return bytecodeSimpleTypeCheckU16
	case runtime.IntegerU32:
		return bytecodeSimpleTypeCheckU32
	case runtime.IntegerU64:
		return bytecodeSimpleTypeCheckU64
	case runtime.IntegerU128:
		return bytecodeSimpleTypeCheckU128
	case runtime.IntegerUsize:
		return bytecodeSimpleTypeCheckUsize
	default:
		return bytecodeSimpleTypeCheckUnknown
	}
}

func bytecodeRecurrenceCallShape(instr bytecodeInstruction, selfSlot int) bool {
	return instr.target == selfSlot &&
		instr.argCount == 0 &&
		instr.hasIntRaw &&
		instr.intImmediateRaw > 0 &&
		instr.intImmediateRaw <= math.MaxInt32
}

func (k *bytecodeI32RecurrenceKernel) eval(kind runtime.IntegerType, n int64) (int64, bool) {
	if k == nil {
		return 0, true
	}
	if k.hasBaseValue(n) {
		return k.baseValue(n), false
	}
	if result, ok, overflow := k.evalNonNegativeDP(kind, n); ok {
		return result, overflow
	}
	firstArg, ok := bytecodeRecurrenceSubtract(kind, n, k.firstSub)
	if !ok {
		return 0, true
	}
	left, overflow := k.eval(kind, firstArg)
	if overflow {
		return 0, true
	}
	secondArg, ok := bytecodeRecurrenceSubtract(kind, n, k.secondSub)
	if !ok {
		return 0, true
	}
	right, overflow := k.eval(kind, secondArg)
	if overflow {
		return 0, true
	}
	sum, ok := bytecodeRecurrenceAdd(kind, left, right)
	if !ok {
		return 0, true
	}
	return sum, false
}

func (k *bytecodeI32RecurrenceKernel) hasExplicitBasePrefix() bool {
	return k != nil && len(k.basePrefix) > 0
}

func (k *bytecodeI32RecurrenceKernel) hasBaseValue(n int64) bool {
	if k == nil {
		return false
	}
	if k.hasExplicitBasePrefix() {
		if n >= 0 && n < int64(len(k.basePrefix)) {
			return true
		}
		if k.baseRangeEnabled {
			return n <= k.baseLimit
		}
		return n >= 0 && n <= k.baseLimit
	}
	return n <= k.baseLimit
}

func (k *bytecodeI32RecurrenceKernel) baseValue(n int64) int64 {
	if k == nil {
		return 0
	}
	if k.hasExplicitBasePrefix() && n >= 0 && n < int64(len(k.basePrefix)) {
		return k.basePrefix[int(n)].raw
	}
	if k.baseRangeEnabled || !k.hasExplicitBasePrefix() {
		if k.baseRangeValue.current {
			return n
		}
		return k.baseRangeValue.raw
	}
	if k.baseRangeValue.current {
		return n
	}
	return k.baseRangeValue.raw
}

func (k *bytecodeI32RecurrenceKernel) baseValuesFitKind(kind runtime.IntegerType) bool {
	if k == nil || kind == "" {
		return true
	}
	for _, base := range k.basePrefix {
		if err := ensureFitsInt64Type(kind, base.raw); err != nil {
			return false
		}
	}
	if (!k.baseRangeValue.current) && (k.baseRangeEnabled || len(k.basePrefix) == 0) {
		if err := ensureFitsInt64Type(kind, k.baseRangeValue.raw); err != nil {
			return false
		}
	}
	return true
}

func bytecodeRecurrenceSubtract(kind runtime.IntegerType, cur int64, sub int64) (int64, bool) {
	prev, overflow := subInt64Overflow(cur, sub)
	if overflow {
		return 0, false
	}
	if err := ensureFitsInt64Type(kind, prev); err != nil {
		return 0, false
	}
	return prev, true
}

func bytecodeRecurrenceAdd(kind runtime.IntegerType, left int64, right int64) (int64, bool) {
	sum, overflow := addInt64Overflow(left, right)
	if overflow {
		return 0, false
	}
	if err := ensureFitsInt64Type(kind, sum); err != nil {
		return 0, false
	}
	return sum, true
}

func (k *bytecodeI32RecurrenceKernel) evalNonNegativeDP(kind runtime.IntegerType, n int64) (int64, bool, bool) {
	if k == nil || k.baseLimit < 0 || n < 0 || n > bytecodeI32RecurrenceDPMaxInput {
		return 0, false, false
	}
	size := int(n) + 1
	values := make([]int64, size)
	for idx := 0; idx < size; idx++ {
		cur := int64(idx)
		if k.hasBaseValue(cur) {
			values[idx] = k.baseValue(cur)
			continue
		}
		left, ok, overflow := k.dpRecurrenceTerm(kind, values, cur, k.firstSub)
		if overflow {
			return 0, true, true
		}
		if !ok {
			return 0, false, false
		}
		right, ok, overflow := k.dpRecurrenceTerm(kind, values, cur, k.secondSub)
		if overflow {
			return 0, true, true
		}
		if !ok {
			return 0, false, false
		}
		sum, ok := bytecodeRecurrenceAdd(kind, left, right)
		if !ok {
			return 0, true, true
		}
		values[idx] = sum
	}
	return values[int(n)], true, false
}

func (k *bytecodeI32RecurrenceKernel) dpRecurrenceTerm(kind runtime.IntegerType, values []int64, cur int64, sub int64) (int64, bool, bool) {
	if k == nil {
		return 0, false, false
	}
	prev, ok := bytecodeRecurrenceSubtract(kind, cur, sub)
	if !ok {
		return 0, true, true
	}
	if k.hasBaseValue(prev) {
		return k.baseValue(prev), true, false
	}
	if prev < 0 || prev >= int64(len(values)) {
		return 0, false, false
	}
	return values[int(prev)], true, false
}

func (vm *bytecodeVM) tryExecI32RecurrenceProgram(program **bytecodeProgram, instructions *[]bytecodeInstruction, validatedIntConsts *[]bool, slotConstIntImmTable **bytecodeSlotConstIntImmediateTable, resume bool) (bool, runtime.Value, error) {
	if resume || vm == nil || vm.ip != 0 || vm.interp == nil || vm.interp.bytecodeStatsEnabled || program == nil || *program == nil {
		return false, nil, nil
	}
	activeProgram := *program
	kernel := activeProgram.i32RecurrenceKernel
	if kernel == nil {
		return false, nil, nil
	}
	if len(vm.slots) == 0 {
		return true, nil, fmt.Errorf("bytecode slot out of range")
	}
	kind, raw, ok := bytecodeRecurrenceExecutionValue(activeProgram.frameLayout, vm.slots[0])
	if !ok {
		return false, nil, nil
	}
	if kernel.hasExplicitBasePrefix() && !kernel.baseRangeEnabled && raw < 0 {
		return false, nil, nil
	}
	if activeProgram.frameLayout != nil && activeProgram.frameLayout.returnSimpleCheck == bytecodeSimpleTypeCheckAnyInteger {
		result, overflow, ok := kernel.evalGeneric(kind, raw)
		if !ok {
			return false, nil, nil
		}
		if overflow {
			err := vm.interp.wrapStandardRuntimeError(newOverflowError("integer overflow"))
			if kernel.overflowAST != nil {
				err = vm.interp.attachRuntimeContext(err, kernel.overflowAST, vm.interp.stateFromEnv(vm.env))
			}
			return true, nil, err
		}
		value := boxedOrSmallIntegerValue(result.kind, result.raw)
		if vm.hasCallFrames() {
			err := vm.finishInlineReturn(program, instructions, validatedIntConsts, slotConstIntImmTable, nil, value, bytecodeRecurrenceSimpleCheckForKind(result.kind))
			return true, nil, err
		}
		return true, value, nil
	}
	if !kernel.baseValuesFitKind(kind) {
		return false, nil, nil
	}
	result, overflow := kernel.eval(kind, raw)
	if overflow {
		err := vm.interp.wrapStandardRuntimeError(newOverflowError("integer overflow"))
		if kernel.overflowAST != nil {
			err = vm.interp.attachRuntimeContext(err, kernel.overflowAST, vm.interp.stateFromEnv(vm.env))
		}
		return true, nil, err
	}
	var value runtime.Value
	if kind == runtime.IntegerI32 {
		value = bytecodeBoxedIntegerI32Value(result)
	} else {
		value = boxedOrSmallIntegerValue(kind, result)
	}
	if vm.hasCallFrames() {
		err := vm.finishInlineReturn(program, instructions, validatedIntConsts, slotConstIntImmTable, nil, value, bytecodeRecurrenceSimpleCheckForKind(kind))
		return true, nil, err
	}
	return true, value, nil
}
