package interpreter

import "able/interpreter-go/pkg/runtime"

type bytecodeF64MatrixRowsCacheEntry struct {
	outerRevision uint64
	outerLength   int
	bound         int
	rows          [][]float64
	rowHandles    []int64
	rowRevisions  []uint64
	rowLengths    []int
}

func (vm *bytecodeVM) tryExecF64MatrixRowLoop(program *bytecodeProgram, plan bytecodeF64MatrixRowLoopPlan) (bool, error) {
	if vm == nil || program == nil || !plan.validForSlots(len(vm.slots)) || plan.successTarget <= vm.ip {
		return false, nil
	}
	index, ok := bytecodeF64DotLoopI32Value(vm.slots[plan.indexSlot])
	if !ok {
		return false, nil
	}
	bound, ok := bytecodeF64DotLoopI32Value(vm.slots[plan.boundSlot])
	if !ok {
		return false, nil
	}
	if index >= bound {
		vm.stack = append(vm.stack, runtime.NilValue{})
		vm.ip = plan.successTarget
		return true, nil
	}
	dest, ok := vm.slots[plan.resultReceiverSlot].(*runtime.ArrayValue)
	if !ok || dest == nil {
		return false, nil
	}
	left, ok := vm.slots[plan.leftReceiverSlot].(*runtime.ArrayValue)
	if !ok || left == nil {
		return false, nil
	}
	outer, ok := vm.slots[plan.rightOuterSlot].(*runtime.ArrayValue)
	if !ok || outer == nil {
		return false, nil
	}
	if bytecodeSameArrayStorage(dest, left) || bytecodeSameArrayStorage(dest, outer) {
		return false, nil
	}
	if !vm.arrayValueNoErrorForPropagation() || !vm.canUseValidatedCanonicalArrayGet(program, outer) || !vm.canUseValidatedCanonicalArrayGet(program, left) {
		return false, nil
	}
	leftValues, ok := vm.f64DotLoopFloatValues(left)
	if !ok {
		return false, nil
	}
	start, end, ok := bytecodeF64MatrixRowLoopRange(index, bound, len(leftValues), outer)
	if !ok {
		return false, nil
	}
	results := make([]float64, end-start)
	rows, cachedRows, err := vm.f64MatrixRowLoopCachedRows(program, dest, outer, int(bound), start)
	if err != nil {
		return false, err
	}
	if cachedRows {
		for rowIdx := start; rowIdx < end; rowIdx++ {
			rowValues := rows[rowIdx]
			acc := 0.0
			for colIdx := 0; colIdx < int(bound); colIdx++ {
				acc += leftValues[colIdx] * rowValues[colIdx]
			}
			results[rowIdx-start] = acc
		}
	} else {
		for rowIdx := start; rowIdx < end; rowIdx++ {
			rowValue, _, _, handled, err := vm.readCanonicalArrayGetValue(outer, int64(rowIdx))
			if err != nil {
				return false, err
			}
			if !handled {
				return false, nil
			}
			row, ok := rowValue.(*runtime.ArrayValue)
			if !ok || row == nil || !vm.canUseValidatedCanonicalArrayGet(program, row) {
				return false, nil
			}
			if bytecodeSameArrayStorage(dest, row) {
				return false, nil
			}
			rowValues, ok := vm.f64DotLoopFloatValues(row)
			if !ok || len(rowValues) < int(bound) {
				return false, nil
			}
			acc := 0.0
			for colIdx := 0; colIdx < int(bound); colIdx++ {
				acc += leftValues[colIdx] * rowValues[colIdx]
			}
			results[rowIdx-start] = acc
		}
	}
	if !vm.appendF64MatrixRowLoopResults(program, plan, dest, results) {
		return false, nil
	}
	vm.storeI32Slot(plan.indexSlot, int64(end))
	vm.stack = append(vm.stack, runtime.NilValue{})
	vm.ip = plan.successTarget
	return true, nil
}

func (vm *bytecodeVM) f64MatrixRowLoopCachedRows(program *bytecodeProgram, dest *runtime.ArrayValue, outer *runtime.ArrayValue, bound int, start int) ([][]float64, bool, error) {
	if vm == nil || program == nil || dest == nil || outer == nil || bound <= 0 || start != 0 {
		return nil, false, nil
	}
	outerState, tracked := bytecodeTrackedArrayState(outer)
	if !tracked || outerState == nil || bound > len(outerState.Values) {
		return nil, false, nil
	}
	if cached, ok := vm.f64MatrixRowsCache[outerState]; ok && cached.matchesOuter(outerState, bound) {
		valid, err := cached.validForRows(dest)
		if err != nil {
			return nil, false, err
		}
		if valid {
			return cached.rows, true, nil
		}
	}
	rows := make([][]float64, bound)
	rowHandles := make([]int64, bound)
	rowRevisions := make([]uint64, bound)
	rowLengths := make([]int, bound)
	for rowIdx := 0; rowIdx < bound; rowIdx++ {
		row, ok := outerState.Values[rowIdx].(*runtime.ArrayValue)
		if !ok || row == nil || bytecodeSameArrayStorage(dest, row) || !vm.canUseValidatedCanonicalArrayGet(program, row) {
			return nil, false, nil
		}
		handle, ok, err := vm.arrayHandleFast(row)
		if err != nil || !ok {
			return nil, false, err
		}
		values, revision, mono, err := runtime.ArrayStoreMonoF64ValuesRevisionIfAvailable(handle)
		if err != nil {
			return nil, false, err
		}
		if !mono || len(values) < bound {
			return nil, false, nil
		}
		rows[rowIdx] = values
		rowHandles[rowIdx] = handle
		rowRevisions[rowIdx] = revision
		rowLengths[rowIdx] = len(values)
	}
	if vm.f64MatrixRowsCache == nil {
		vm.f64MatrixRowsCache = make(map[*runtime.ArrayState]bytecodeF64MatrixRowsCacheEntry, 2)
	}
	vm.f64MatrixRowsCache[outerState] = bytecodeF64MatrixRowsCacheEntry{
		outerRevision: outerState.Revision,
		outerLength:   len(outerState.Values),
		bound:         bound,
		rows:          rows,
		rowHandles:    rowHandles,
		rowRevisions:  rowRevisions,
		rowLengths:    rowLengths,
	}
	return rows, true, nil
}

func (entry bytecodeF64MatrixRowsCacheEntry) matchesOuter(state *runtime.ArrayState, bound int) bool {
	return state != nil &&
		entry.outerRevision == state.Revision &&
		entry.outerLength == len(state.Values) &&
		entry.bound == bound &&
		len(entry.rows) == bound &&
		len(entry.rowHandles) == bound &&
		len(entry.rowRevisions) == bound &&
		len(entry.rowLengths) == bound
}

func (entry bytecodeF64MatrixRowsCacheEntry) validForRows(dest *runtime.ArrayValue) (bool, error) {
	destHandle := bytecodeArrayStorageHandle(dest)
	for idx, handle := range entry.rowHandles {
		if handle == 0 || (destHandle != 0 && destHandle == handle) {
			return false, nil
		}
		values, revision, mono, err := runtime.ArrayStoreMonoF64ValuesRevisionIfAvailable(handle)
		if err != nil {
			return false, err
		}
		if !mono || revision != entry.rowRevisions[idx] || len(values) != entry.rowLengths[idx] || len(values) < entry.bound {
			return false, nil
		}
	}
	return true, nil
}

func bytecodeSameArrayStorage(left *runtime.ArrayValue, right *runtime.ArrayValue) bool {
	if left == nil || right == nil {
		return false
	}
	if left == right {
		return true
	}
	leftHandle := left.Handle
	if leftHandle == 0 {
		leftHandle = left.TrackedHandle
	}
	rightHandle := right.Handle
	if rightHandle == 0 {
		rightHandle = right.TrackedHandle
	}
	return leftHandle != 0 && leftHandle == rightHandle
}

func bytecodeArrayStorageHandle(arr *runtime.ArrayValue) int64 {
	if arr == nil {
		return 0
	}
	if arr.Handle != 0 {
		return arr.Handle
	}
	return arr.TrackedHandle
}

func bytecodeF64MatrixRowLoopRange(index, bound int64, leftLen int, outer *runtime.ArrayValue) (int, int, bool) {
	if index < 0 || bound < 0 || bound > int64(leftLen) {
		return 0, 0, false
	}
	start := int(index)
	end := int(bound)
	if start < 0 || end < start {
		return 0, 0, false
	}
	if state, tracked := bytecodeTrackedArrayState(outer); tracked {
		return start, end, state != nil && end <= len(state.Values)
	}
	if outer == nil || outer.Handle == 0 {
		return 0, 0, false
	}
	size, err := runtime.ArrayStoreSize(outer.Handle)
	if err != nil || size < 0 {
		return 0, 0, false
	}
	return start, end, end <= size
}

func (vm *bytecodeVM) appendF64MatrixRowLoopResults(program *bytecodeProgram, plan bytecodeF64MatrixRowLoopPlan, dest *runtime.ArrayValue, values []float64) bool {
	if dest == nil {
		return false
	}
	instr := bytecodeInstruction{op: bytecodeOpCallMemberArraySlot, name: "push", argCount: 1}
	if !vm.canUseCanonicalArrayPushAt(program, plan.resultPushIP, instr, dest) {
		return false
	}
	return vm.appendArrayF64ValuesFast(dest, values)
}

func (vm *bytecodeVM) appendArrayF64ValuesFast(arr *runtime.ArrayValue, values []float64) bool {
	if vm == nil || vm.interp == nil || arr == nil || arr.Handle == 0 || arr.TrackedAliases {
		return false
	}
	if len(values) == 0 {
		return true
	}
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		if state == nil {
			return false
		}
		if state.ElementTypeTokenKnown && state.ElementTypeToken != bytecodeIndexTypeF64 && state.ElementTypeToken != bytecodeIndexTypeUnknown {
			return false
		}
	}
	ok, err := runtime.ArrayStoreAppendF64ValuesPromote(arr.Handle, values)
	if err != nil || !ok {
		return false
	}
	arr.State = nil
	arr.Elements = nil
	arr.TrackedHandle = arr.Handle
	return true
}
