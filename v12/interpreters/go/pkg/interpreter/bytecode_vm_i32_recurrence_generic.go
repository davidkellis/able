package interpreter

import (
	"unsafe"

	"able/interpreter-go/pkg/runtime"
)

const bytecodeRecurrenceIntegerKindCount = 12
const bytecodeRecurrenceGenericDenseDPBudgetBytes = 64 << 20
const bytecodeRecurrenceGenericPagedDPBudgetBytes = 96 << 20
const bytecodeRecurrencePagedDPPageSize = 4096

type bytecodeRecurrenceBaseValue struct {
	raw         int64
	current     bool
	literalKind runtime.IntegerType
}

type bytecodeRecurrenceState struct {
	raw  int64
	kind runtime.IntegerType
}

type bytecodeRecurrenceStateResult struct {
	raw  int64
	kind runtime.IntegerType
}

type bytecodeRecurrenceKindSet struct {
	kinds          []runtime.IntegerType
	denseByOrdinal [bytecodeRecurrenceIntegerKindCount]int
}

type bytecodeRecurrenceEvalFrame struct {
	state      bytecodeRecurrenceState
	leftState  bytecodeRecurrenceState
	rightState bytecodeRecurrenceState
	phase      uint8
}

func (g bytecodeRecurrenceBaseGuard) baseValue(n int64) bytecodeRecurrenceBaseValue {
	return bytecodeRecurrenceBaseValue{
		raw:         g.value(n),
		current:     g.returnCurrent,
		literalKind: g.returnKind,
	}
}

func (v bytecodeRecurrenceBaseValue) resolve(state bytecodeRecurrenceState) bytecodeRecurrenceStateResult {
	if v.current {
		return bytecodeRecurrenceStateResult{raw: state.raw, kind: state.kind}
	}
	return bytecodeRecurrenceStateResult{raw: v.raw, kind: v.literalKind}
}

func bytecodeRecurrenceSubtractState(state bytecodeRecurrenceState, subRaw int64, subKind runtime.IntegerType) (bytecodeRecurrenceState, bool) {
	nextKind, err := promoteIntegerTypes(state.kind, subKind)
	if err != nil {
		return bytecodeRecurrenceState{}, false
	}
	nextRaw, overflow := subInt64Overflow(state.raw, subRaw)
	if overflow {
		return bytecodeRecurrenceState{}, false
	}
	if err := ensureFitsInt64Type(nextKind, nextRaw); err != nil {
		return bytecodeRecurrenceState{}, false
	}
	return bytecodeRecurrenceState{raw: nextRaw, kind: nextKind}, true
}

func bytecodeRecurrenceAddStateResults(left bytecodeRecurrenceStateResult, right bytecodeRecurrenceStateResult) (bytecodeRecurrenceStateResult, bool, bool) {
	resultKind, err := promoteIntegerTypes(left.kind, right.kind)
	if err != nil {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	sum, overflow := addInt64Overflow(left.raw, right.raw)
	if overflow {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	if err := ensureFitsInt64Type(resultKind, sum); err != nil {
		return bytecodeRecurrenceStateResult{}, true, true
	}
	return bytecodeRecurrenceStateResult{raw: sum, kind: resultKind}, false, true
}

func (k *bytecodeI32RecurrenceKernel) genericBaseValue(state bytecodeRecurrenceState) (bytecodeRecurrenceStateResult, bool) {
	if k == nil {
		return bytecodeRecurrenceStateResult{}, false
	}
	if len(k.basePrefix) > 0 && state.raw >= 0 && state.raw < int64(len(k.basePrefix)) {
		return k.basePrefix[int(state.raw)].resolve(state), true
	}
	if !k.baseRangeEnabled && len(k.basePrefix) > 0 {
		return bytecodeRecurrenceStateResult{}, false
	}
	if state.raw > k.baseLimit {
		return bytecodeRecurrenceStateResult{}, false
	}
	return k.baseRangeValue.resolve(state), true
}

func (k *bytecodeI32RecurrenceKernel) evalGeneric(kind runtime.IntegerType, n int64) (bytecodeRecurrenceStateResult, bool, bool) {
	if k == nil {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	root := bytecodeRecurrenceState{raw: n, kind: kind}
	if result, ok := k.genericBaseValue(root); ok {
		return result, false, true
	}
	if stableKind, ok := k.genericStableRecursiveKind(kind); ok {
		if result, overflow, ok := k.evalGenericStableDP(stableKind, n); ok {
			return result, overflow, true
		}
	}
	if result, overflow, ok := k.evalGenericDenseDP(kind, n); ok {
		return result, overflow, true
	}
	if result, overflow, ok := k.evalGenericPagedDP(kind, n); ok {
		return result, overflow, true
	}
	if result, overflow, ok := k.evalGenericMemoRows(kind, n); ok {
		return result, overflow, true
	}
	return k.evalGenericState(
		root,
		make(map[bytecodeRecurrenceState]bytecodeRecurrenceStateResult),
	)
}

func (k *bytecodeI32RecurrenceKernel) evalGenericState(state bytecodeRecurrenceState, memo map[bytecodeRecurrenceState]bytecodeRecurrenceStateResult) (bytecodeRecurrenceStateResult, bool, bool) {
	if k == nil {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	if result, ok := memo[state]; ok {
		return result, false, true
	}
	if result, ok := k.genericBaseValue(state); ok {
		memo[state] = result
		return result, false, true
	}
	leftState, ok := bytecodeRecurrenceSubtractState(state, k.firstSub, k.firstSubKind)
	if !ok {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	left, overflow, ok := k.evalGenericState(leftState, memo)
	if !ok || overflow {
		return bytecodeRecurrenceStateResult{}, overflow, ok
	}
	rightState, ok := bytecodeRecurrenceSubtractState(state, k.secondSub, k.secondSubKind)
	if !ok {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	right, overflow, ok := k.evalGenericState(rightState, memo)
	if !ok || overflow {
		return bytecodeRecurrenceStateResult{}, overflow, ok
	}
	result, overflow, ok := bytecodeRecurrenceAddStateResults(left, right)
	if !ok || overflow {
		return bytecodeRecurrenceStateResult{}, overflow, ok
	}
	memo[state] = result
	return result, false, true
}

func (k *bytecodeI32RecurrenceKernel) genericStableRecursiveKind(entryKind runtime.IntegerType) (runtime.IntegerType, bool) {
	if k == nil {
		return "", false
	}
	firstKind, err := promoteIntegerTypes(entryKind, k.firstSubKind)
	if err != nil {
		return "", false
	}
	secondKind, err := promoteIntegerTypes(entryKind, k.secondSubKind)
	if err != nil {
		return "", false
	}
	if firstKind != secondKind {
		return "", false
	}
	stableKind := firstKind
	nextKind, err := promoteIntegerTypes(stableKind, k.firstSubKind)
	if err != nil || nextKind != stableKind {
		return "", false
	}
	nextKind, err = promoteIntegerTypes(stableKind, k.secondSubKind)
	if err != nil || nextKind != stableKind {
		return "", false
	}
	return stableKind, true
}

func (k *bytecodeI32RecurrenceKernel) evalGenericStableDP(stableKind runtime.IntegerType, n int64) (bytecodeRecurrenceStateResult, bool, bool) {
	if k == nil || k.baseLimit < 0 || n < 0 || n > bytecodeI32RecurrenceDPMaxInput {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	values := make([]bytecodeRecurrenceStateResult, int(n)+1)
	for idx := 0; idx <= int(n); idx++ {
		cur := int64(idx)
		state := bytecodeRecurrenceState{raw: cur, kind: stableKind}
		if result, ok := k.genericBaseValue(state); ok {
			values[idx] = result
			continue
		}
		left, ok, overflow := k.genericDPRecurrenceTerm(stableKind, values, cur, k.firstSub, k.firstSubKind)
		if overflow {
			return bytecodeRecurrenceStateResult{}, true, true
		}
		if !ok {
			return bytecodeRecurrenceStateResult{}, false, false
		}
		right, ok, overflow := k.genericDPRecurrenceTerm(stableKind, values, cur, k.secondSub, k.secondSubKind)
		if overflow {
			return bytecodeRecurrenceStateResult{}, true, true
		}
		if !ok {
			return bytecodeRecurrenceStateResult{}, false, false
		}
		result, overflow, ok := bytecodeRecurrenceAddStateResults(left, right)
		if overflow {
			return bytecodeRecurrenceStateResult{}, true, true
		}
		if !ok {
			return bytecodeRecurrenceStateResult{}, false, false
		}
		values[idx] = result
	}
	return values[int(n)], false, true
}

func (k *bytecodeI32RecurrenceKernel) genericDPRecurrenceTerm(stableKind runtime.IntegerType, values []bytecodeRecurrenceStateResult, cur int64, subRaw int64, subKind runtime.IntegerType) (bytecodeRecurrenceStateResult, bool, bool) {
	if k == nil {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	prevState, ok := bytecodeRecurrenceSubtractState(bytecodeRecurrenceState{raw: cur, kind: stableKind}, subRaw, subKind)
	if !ok {
		return bytecodeRecurrenceStateResult{}, true, true
	}
	if prevState.kind != stableKind {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	if result, ok := k.genericBaseValue(prevState); ok {
		return result, true, false
	}
	if prevState.raw < 0 || prevState.raw >= int64(len(values)) {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	return values[int(prevState.raw)], true, false
}

func bytecodeRecurrenceKindOrdinal(kind runtime.IntegerType) int {
	switch kind {
	case runtime.IntegerI8:
		return 0
	case runtime.IntegerI16:
		return 1
	case runtime.IntegerI32:
		return 2
	case runtime.IntegerI64:
		return 3
	case runtime.IntegerI128:
		return 4
	case runtime.IntegerU8:
		return 5
	case runtime.IntegerU16:
		return 6
	case runtime.IntegerU32:
		return 7
	case runtime.IntegerU64:
		return 8
	case runtime.IntegerU128:
		return 9
	case runtime.IntegerIsize:
		return 10
	case runtime.IntegerUsize:
		return 11
	default:
		return -1
	}
}

func (s *bytecodeRecurrenceKindSet) add(kind runtime.IntegerType) bool {
	if s == nil {
		return false
	}
	ord := bytecodeRecurrenceKindOrdinal(kind)
	if ord < 0 {
		return false
	}
	if s.denseByOrdinal[ord] != 0 {
		return true
	}
	s.kinds = append(s.kinds, kind)
	s.denseByOrdinal[ord] = len(s.kinds)
	return true
}

func (s bytecodeRecurrenceKindSet) index(kind runtime.IntegerType) (int, bool) {
	ord := bytecodeRecurrenceKindOrdinal(kind)
	if ord < 0 {
		return 0, false
	}
	dense := s.denseByOrdinal[ord]
	if dense == 0 {
		return 0, false
	}
	return dense - 1, true
}

func bytecodeBuildRecurrenceKindSet(entryKind runtime.IntegerType, firstSubKind runtime.IntegerType, secondSubKind runtime.IntegerType) (bytecodeRecurrenceKindSet, bool) {
	var set bytecodeRecurrenceKindSet
	if !set.add(entryKind) {
		return bytecodeRecurrenceKindSet{}, false
	}
	for idx := 0; idx < len(set.kinds); idx++ {
		cur := set.kinds[idx]
		firstKind, err := promoteIntegerTypes(cur, firstSubKind)
		if err != nil {
			return bytecodeRecurrenceKindSet{}, false
		}
		if !set.add(firstKind) {
			return bytecodeRecurrenceKindSet{}, false
		}
		secondKind, err := promoteIntegerTypes(cur, secondSubKind)
		if err != nil {
			return bytecodeRecurrenceKindSet{}, false
		}
		if !set.add(secondKind) {
			return bytecodeRecurrenceKindSet{}, false
		}
	}
	return set, true
}

func (k *bytecodeI32RecurrenceKernel) genericEstimatedRowCount(n int64) int64 {
	if k == nil {
		return 64
	}
	minSub := k.firstSub
	if k.secondSub > 0 && (minSub <= 0 || k.secondSub < minSub) {
		minSub = k.secondSub
	}
	if minSub <= 0 {
		return 64
	}
	span := int64(64)
	switch {
	case n >= 0:
		span = n + 1
	case k.baseLimit < n:
		span = n - k.baseLimit + 1
	}
	return span/minSub + 2
}

func (k *bytecodeI32RecurrenceKernel) genericMemoRowHint(n int64) int {
	rows := k.genericEstimatedRowCount(n)
	if rows < 64 {
		return 64
	}
	if rows > 1<<20 {
		return 1 << 20
	}
	return int(rows)
}

func bytecodeRecurrenceGenericDenseDPBudgetOK(n int64, rowWidth int) bool {
	if n < 0 || rowWidth <= 0 {
		return false
	}
	stateCount := n + 1
	if stateCount <= 0 {
		return false
	}
	valueEntries := stateCount * int64(rowWidth)
	if valueEntries < stateCount {
		return false
	}
	valueBytes := valueEntries * int64(unsafe.Sizeof(bytecodeRecurrenceStateResult{}))
	knownBytes := valueEntries
	reachableBytes := stateCount * int64(unsafe.Sizeof(uint16(0)))
	totalBytes := valueBytes + knownBytes + reachableBytes
	return totalBytes > 0 && totalBytes <= bytecodeRecurrenceGenericDenseDPBudgetBytes
}

func (k *bytecodeI32RecurrenceKernel) genericPagedDPPreferred(n int64, rowWidth int) bool {
	if k == nil || n < 0 || rowWidth <= 0 || !bytecodeRecurrenceGenericPagedDPBudgetOK(n, rowWidth) {
		return false
	}
	rows := k.genericEstimatedRowCount(n)
	if rows <= 0 {
		return false
	}
	if rows*4 >= n+1 {
		return true
	}
	return k.genericNearDensePagedDPPreferred(n, rowWidth)
}

func (k *bytecodeI32RecurrenceKernel) evalGenericMemoRows(entryKind runtime.IntegerType, n int64) (bytecodeRecurrenceStateResult, bool, bool) {
	if k == nil {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	kindSet, ok := bytecodeBuildRecurrenceKindSet(entryKind, k.firstSubKind, k.secondSubKind)
	if !ok || len(kindSet.kinds) == 0 || len(kindSet.kinds) > 16 {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	root := bytecodeRecurrenceState{raw: n, kind: entryKind}
	layout, indexed, paged := k.genericMemoStorageLayout(n, len(kindSet.kinds))
	if !indexed {
		layout = bytecodeRecurrenceMemoLayout{}
	}
	memo := newBytecodeRecurrenceMemoRows(kindSet, k.genericMemoRowHint(n), layout, paged)
	stack := []bytecodeRecurrenceEvalFrame{{state: root}}
	for len(stack) > 0 {
		frameIdx := len(stack) - 1
		frame := &stack[frameIdx]
		if result, ok := memo.get(frame.state); ok {
			_ = result
			stack = stack[:frameIdx]
			continue
		}
		if result, ok := k.genericBaseValue(frame.state); ok {
			if !memo.set(frame.state, result) {
				return bytecodeRecurrenceStateResult{}, false, false
			}
			stack = stack[:frameIdx]
			continue
		}
		switch frame.phase {
		case 0:
			leftState, ok := bytecodeRecurrenceSubtractState(frame.state, k.firstSub, k.firstSubKind)
			if !ok {
				return bytecodeRecurrenceStateResult{}, false, false
			}
			frame.leftState = leftState
			frame.phase = 1
			if _, ok := memo.get(leftState); ok {
				continue
			}
			if result, ok := k.genericBaseValue(leftState); ok {
				if !memo.set(leftState, result) {
					return bytecodeRecurrenceStateResult{}, false, false
				}
				continue
			}
			stack = append(stack, bytecodeRecurrenceEvalFrame{state: leftState})
		case 1:
			if _, ok := memo.get(frame.leftState); !ok {
				return bytecodeRecurrenceStateResult{}, false, false
			}
			rightState, ok := bytecodeRecurrenceSubtractState(frame.state, k.secondSub, k.secondSubKind)
			if !ok {
				return bytecodeRecurrenceStateResult{}, false, false
			}
			frame.rightState = rightState
			frame.phase = 2
			if _, ok := memo.get(rightState); ok {
				continue
			}
			if result, ok := k.genericBaseValue(rightState); ok {
				if !memo.set(rightState, result) {
					return bytecodeRecurrenceStateResult{}, false, false
				}
				continue
			}
			stack = append(stack, bytecodeRecurrenceEvalFrame{state: rightState})
		default:
			left, ok := memo.get(frame.leftState)
			if !ok {
				return bytecodeRecurrenceStateResult{}, false, false
			}
			right, ok := memo.get(frame.rightState)
			if !ok {
				return bytecodeRecurrenceStateResult{}, false, false
			}
			result, overflow, ok := bytecodeRecurrenceAddStateResults(left, right)
			if overflow {
				return bytecodeRecurrenceStateResult{}, true, true
			}
			if !ok || !memo.set(frame.state, result) {
				return bytecodeRecurrenceStateResult{}, false, false
			}
			stack = stack[:frameIdx]
		}
	}
	result, ok := memo.get(root)
	return result, false, ok
}

func (k *bytecodeI32RecurrenceKernel) evalGenericPagedDP(entryKind runtime.IntegerType, n int64) (bytecodeRecurrenceStateResult, bool, bool) {
	if k == nil || k.baseLimit < 0 || n < 0 {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	kindSet, ok := bytecodeBuildRecurrenceKindSet(entryKind, k.firstSubKind, k.secondSubKind)
	if !ok || len(kindSet.kinds) == 0 {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	rowWidth := len(kindSet.kinds)
	if rowWidth > 16 || !k.genericPagedDPPreferred(n, rowWidth) {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	rootIdx, ok := kindSet.index(entryKind)
	if !ok {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	stateCount := int(n) + 1
	if stateCount <= 0 {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	reachable := make([]uint16, stateCount)
	reachable[int(n)] = uint16(1) << rootIdx
	for raw := int(n); raw >= 0; raw-- {
		mask := reachable[raw]
		if mask == 0 {
			continue
		}
		curRaw := int64(raw)
		for denseIdx, kind := range kindSet.kinds {
			if mask&(uint16(1)<<denseIdx) == 0 {
				continue
			}
			state := bytecodeRecurrenceState{raw: curRaw, kind: kind}
			if _, ok := k.genericBaseValue(state); ok {
				continue
			}
			leftState, ok := bytecodeRecurrenceSubtractState(state, k.firstSub, k.firstSubKind)
			if !ok {
				continue
			}
			k.genericDenseDPTrackState(kindSet, reachable, n, leftState)
			rightState, ok := bytecodeRecurrenceSubtractState(state, k.secondSub, k.secondSubKind)
			if !ok {
				continue
			}
			k.genericDenseDPTrackState(kindSet, reachable, n, rightState)
		}
	}
	values := newBytecodeRecurrencePagedValues(kindSet, stateCount)
	for raw := 0; raw <= int(n); raw++ {
		mask := reachable[raw]
		if mask == 0 {
			continue
		}
		curRaw := int64(raw)
		for denseIdx, kind := range kindSet.kinds {
			if mask&(uint16(1)<<denseIdx) == 0 {
				continue
			}
			state := bytecodeRecurrenceState{raw: curRaw, kind: kind}
			if result, ok := k.genericBaseValue(state); ok {
				if !values.set(state, result) {
					return bytecodeRecurrenceStateResult{}, false, false
				}
				continue
			}
			leftState, ok := bytecodeRecurrenceSubtractState(state, k.firstSub, k.firstSubKind)
			if !ok {
				continue
			}
			left, ok, overflow := k.genericPagedDPValue(values, n, leftState)
			if overflow {
				return bytecodeRecurrenceStateResult{}, true, true
			}
			if !ok {
				continue
			}
			rightState, ok := bytecodeRecurrenceSubtractState(state, k.secondSub, k.secondSubKind)
			if !ok {
				continue
			}
			right, ok, overflow := k.genericPagedDPValue(values, n, rightState)
			if overflow {
				return bytecodeRecurrenceStateResult{}, true, true
			}
			if !ok {
				continue
			}
			result, overflow, ok := bytecodeRecurrenceAddStateResults(left, right)
			if overflow {
				return bytecodeRecurrenceStateResult{}, true, true
			}
			if !ok || !values.set(state, result) {
				continue
			}
		}
	}
	result, ok, overflow := k.genericPagedDPValue(values, n, bytecodeRecurrenceState{raw: n, kind: entryKind})
	return result, overflow, ok
}

func (k *bytecodeI32RecurrenceKernel) evalGenericDenseDP(entryKind runtime.IntegerType, n int64) (bytecodeRecurrenceStateResult, bool, bool) {
	if k == nil || k.baseLimit < 0 || n < 0 {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	kindSet, ok := bytecodeBuildRecurrenceKindSet(entryKind, k.firstSubKind, k.secondSubKind)
	if !ok || len(kindSet.kinds) == 0 {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	rowWidth := len(kindSet.kinds)
	if rowWidth > 16 || !bytecodeRecurrenceGenericDenseDPBudgetOK(n, rowWidth) {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	rootIdx, ok := kindSet.index(entryKind)
	if !ok {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	reachable := make([]uint16, int(n)+1)
	reachable[int(n)] = uint16(1) << rootIdx
	for raw := int(n); raw >= 0; raw-- {
		mask := reachable[raw]
		if mask == 0 {
			continue
		}
		curRaw := int64(raw)
		for denseIdx, kind := range kindSet.kinds {
			if mask&(uint16(1)<<denseIdx) == 0 {
				continue
			}
			state := bytecodeRecurrenceState{raw: curRaw, kind: kind}
			if _, ok := k.genericBaseValue(state); ok {
				continue
			}
			leftState, ok := bytecodeRecurrenceSubtractState(state, k.firstSub, k.firstSubKind)
			if !ok {
				continue
			}
			k.genericDenseDPTrackState(kindSet, reachable, n, leftState)
			rightState, ok := bytecodeRecurrenceSubtractState(state, k.secondSub, k.secondSubKind)
			if !ok {
				continue
			}
			k.genericDenseDPTrackState(kindSet, reachable, n, rightState)
		}
	}
	values := make([]bytecodeRecurrenceStateResult, (int(n)+1)*rowWidth)
	known := make([]bool, len(values))
	for raw := 0; raw <= int(n); raw++ {
		mask := reachable[raw]
		if mask == 0 {
			continue
		}
		curRaw := int64(raw)
		rowBase := raw * rowWidth
		for denseIdx, kind := range kindSet.kinds {
			if mask&(uint16(1)<<denseIdx) == 0 {
				continue
			}
			idx := rowBase + denseIdx
			state := bytecodeRecurrenceState{raw: curRaw, kind: kind}
			if result, ok := k.genericBaseValue(state); ok {
				values[idx] = result
				known[idx] = true
				continue
			}
			leftState, ok := bytecodeRecurrenceSubtractState(state, k.firstSub, k.firstSubKind)
			if !ok {
				continue
			}
			left, ok, overflow := k.genericDenseDPValue(kindSet, rowWidth, values, known, n, leftState)
			if overflow {
				return bytecodeRecurrenceStateResult{}, true, true
			}
			if !ok {
				continue
			}
			rightState, ok := bytecodeRecurrenceSubtractState(state, k.secondSub, k.secondSubKind)
			if !ok {
				continue
			}
			right, ok, overflow := k.genericDenseDPValue(kindSet, rowWidth, values, known, n, rightState)
			if overflow {
				return bytecodeRecurrenceStateResult{}, true, true
			}
			if !ok {
				continue
			}
			result, overflow, ok := bytecodeRecurrenceAddStateResults(left, right)
			if overflow {
				return bytecodeRecurrenceStateResult{}, true, true
			}
			if !ok {
				continue
			}
			values[idx] = result
			known[idx] = true
		}
	}
	flat := int(n)*rowWidth + rootIdx
	if !known[flat] {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	return values[flat], false, true
}

func (k *bytecodeI32RecurrenceKernel) genericPagedDPValue(values *bytecodeRecurrencePagedValues, limit int64, state bytecodeRecurrenceState) (bytecodeRecurrenceStateResult, bool, bool) {
	if k == nil {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	if result, ok := k.genericBaseValue(state); ok {
		return result, true, false
	}
	if state.raw < 0 || state.raw > limit {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	if values == nil {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	result, ok := values.get(state)
	return result, ok, false
}

func (k *bytecodeI32RecurrenceKernel) genericDenseDPValue(kindSet bytecodeRecurrenceKindSet, rowWidth int, values []bytecodeRecurrenceStateResult, known []bool, limit int64, state bytecodeRecurrenceState) (bytecodeRecurrenceStateResult, bool, bool) {
	if k == nil {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	if result, ok := k.genericBaseValue(state); ok {
		return result, true, false
	}
	if state.raw < 0 || state.raw > limit {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	denseIdx, ok := kindSet.index(state.kind)
	if !ok {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	flat := int(state.raw)*rowWidth + denseIdx
	if flat < 0 || flat >= len(values) || !known[flat] {
		return bytecodeRecurrenceStateResult{}, false, false
	}
	return values[flat], true, false
}

func (k *bytecodeI32RecurrenceKernel) genericDenseDPTrackState(kindSet bytecodeRecurrenceKindSet, reachable []uint16, limit int64, state bytecodeRecurrenceState) bool {
	if k == nil {
		return false
	}
	if _, ok := k.genericBaseValue(state); ok {
		return true
	}
	if state.raw < 0 || state.raw > limit {
		return false
	}
	denseIdx, ok := kindSet.index(state.kind)
	if !ok {
		return false
	}
	reachable[int(state.raw)] |= uint16(1) << denseIdx
	return true
}
