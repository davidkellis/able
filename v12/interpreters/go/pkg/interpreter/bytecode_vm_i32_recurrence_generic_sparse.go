package interpreter

import "unsafe"

const bytecodeRecurrenceNearDensePagedDPDensityNumerator = 1
const bytecodeRecurrenceNearDensePagedDPDensityDenominator = 2

type bytecodeRecurrenceMemoLayout struct {
	baseRaw  int64
	step     int64
	rowCount int
}

type bytecodeRecurrenceMemoRows struct {
	rowWidth          int
	kindSet           bytecodeRecurrenceKindSet
	indexBase         int64
	indexStep         int64
	indexRows         int
	indexedKnownMasks []uint16
	indexedValues     []bytecodeRecurrenceStateResult
	indexedPages      *bytecodeRecurrencePagedValues
	spillRowIndex     map[int64]int
	spillKnownMasks   []uint16
	spillValues       []bytecodeRecurrenceStateResult
}

type bytecodeRecurrenceValuePage struct {
	knownMasks []uint16
	values     []bytecodeRecurrenceStateResult
}

type bytecodeRecurrencePagedValues struct {
	rowWidth int
	pageSize int
	kindSet  bytecodeRecurrenceKindSet
	pages    []*bytecodeRecurrenceValuePage
}

func bytecodeRecurrenceGCD64(a int64, b int64) int64 {
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func bytecodeRecurrenceGenericIndexedBudgetOK(rowCount int64, rowWidth int) bool {
	if rowCount <= 0 || rowWidth <= 0 {
		return false
	}
	valueEntries := rowCount * int64(rowWidth)
	if valueEntries < rowCount {
		return false
	}
	valueBytes := valueEntries * int64(unsafe.Sizeof(bytecodeRecurrenceStateResult{}))
	knownBytes := rowCount * int64(unsafe.Sizeof(uint16(0)))
	totalBytes := valueBytes + knownBytes
	return totalBytes > 0 && totalBytes <= bytecodeRecurrenceGenericDenseDPBudgetBytes
}

func bytecodeRecurrenceGenericPagedDPBudgetOK(n int64, rowWidth int) bool {
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
	rowKnownBytes := stateCount * int64(unsafe.Sizeof(uint16(0)))
	reachableBytes := stateCount * int64(unsafe.Sizeof(uint16(0)))
	totalBytes := valueBytes + rowKnownBytes + reachableBytes
	return totalBytes > 0 && totalBytes <= bytecodeRecurrenceGenericPagedDPBudgetBytes
}

func bytecodeRecurrenceFirstCongruentRaw(minRaw int64, target int64, step int64) (int64, bool) {
	if minRaw < 0 || target < minRaw || step <= 0 {
		return 0, false
	}
	residue := target % step
	delta := residue - (minRaw % step)
	if delta < 0 {
		delta += step
	}
	start, overflow := addInt64Overflow(minRaw, delta)
	if overflow || start > target {
		return 0, false
	}
	return start, true
}

func (k *bytecodeI32RecurrenceKernel) genericMemoIndexedLayout(n int64) (bytecodeRecurrenceMemoLayout, bool) {
	if k == nil || n < 0 || k.baseLimit < 0 {
		return bytecodeRecurrenceMemoLayout{}, false
	}
	step := bytecodeRecurrenceGCD64(k.firstSub, k.secondSub)
	if step <= 0 {
		return bytecodeRecurrenceMemoLayout{}, false
	}
	minRaw, overflow := addInt64Overflow(k.baseLimit, 1)
	if overflow || minRaw < 0 {
		return bytecodeRecurrenceMemoLayout{}, false
	}
	start, ok := bytecodeRecurrenceFirstCongruentRaw(minRaw, n, step)
	if !ok {
		return bytecodeRecurrenceMemoLayout{}, false
	}
	rowCount := ((n - start) / step) + 1
	if rowCount <= 0 || rowCount > int64(^uint(0)>>1) {
		return bytecodeRecurrenceMemoLayout{}, false
	}
	return bytecodeRecurrenceMemoLayout{
		baseRaw:  start,
		step:     step,
		rowCount: int(rowCount),
	}, true
}

func (k *bytecodeI32RecurrenceKernel) genericMemoStorageLayout(n int64, rowWidth int) (bytecodeRecurrenceMemoLayout, bool, bool) {
	layout, ok := k.genericMemoIndexedLayout(n)
	if !ok || rowWidth <= 0 {
		return bytecodeRecurrenceMemoLayout{}, false, false
	}
	return layout, true, !bytecodeRecurrenceGenericIndexedBudgetOK(int64(layout.rowCount), rowWidth)
}

func bytecodeRecurrenceRepresentableCount(limit int64, first int64, second int64) (int64, bool) {
	if limit < 0 || first <= 0 || second <= 0 {
		return 0, false
	}
	if first > second {
		first, second = second, first
	}
	if first == 1 {
		return limit + 1, true
	}
	if first > 4096 {
		return 0, false
	}
	var count int64
	for multiplier := int64(0); multiplier < first; multiplier++ {
		base, overflow := mulInt64Overflow(second, multiplier)
		if overflow || base > limit {
			continue
		}
		count += ((limit - base) / first) + 1
	}
	return count, true
}

func bytecodeRecurrenceDensityAtLeast(reachableRows int64, totalRows int64, numerator int64, denominator int64) bool {
	if reachableRows < 0 || totalRows <= 0 || numerator <= 0 || denominator <= 0 {
		return false
	}
	left, overflow := mulInt64Overflow(reachableRows, denominator)
	if overflow {
		return true
	}
	right, overflow := mulInt64Overflow(totalRows, numerator)
	if overflow {
		return false
	}
	return left >= right
}

func (k *bytecodeI32RecurrenceKernel) genericNearDensePagedDPPreferred(n int64, rowWidth int) bool {
	layout, indexed, paged := k.genericMemoStorageLayout(n, rowWidth)
	if !indexed || !paged || layout.rowCount <= 0 || layout.step <= 0 {
		return false
	}
	firstReduced := k.firstSub / layout.step
	secondReduced := k.secondSub / layout.step
	if firstReduced <= 0 || secondReduced <= 0 {
		return false
	}
	reachableRows, ok := bytecodeRecurrenceRepresentableCount(int64(layout.rowCount-1), firstReduced, secondReduced)
	if !ok {
		return false
	}
	totalRows := int64(layout.rowCount)
	return bytecodeRecurrenceDensityAtLeast(
		reachableRows,
		totalRows,
		bytecodeRecurrenceNearDensePagedDPDensityNumerator,
		bytecodeRecurrenceNearDensePagedDPDensityDenominator,
	)
}

func newBytecodeRecurrenceMemoRows(kindSet bytecodeRecurrenceKindSet, rowHint int, layout bytecodeRecurrenceMemoLayout, pagedIndexed bool) *bytecodeRecurrenceMemoRows {
	if rowHint < 0 {
		rowHint = 0
	}
	indexRows := layout.rowCount
	if indexRows < 0 {
		indexRows = 0
	}
	spillHint := rowHint
	if spillHint > indexRows {
		spillHint -= indexRows
	} else {
		spillHint = 0
	}
	memo := &bytecodeRecurrenceMemoRows{
		rowWidth:        len(kindSet.kinds),
		kindSet:         kindSet,
		indexBase:       layout.baseRaw,
		indexStep:       layout.step,
		indexRows:       indexRows,
		spillRowIndex:   make(map[int64]int, spillHint),
		spillKnownMasks: make([]uint16, 0, spillHint),
		spillValues:     make([]bytecodeRecurrenceStateResult, 0, spillHint*len(kindSet.kinds)),
	}
	if indexRows == 0 {
		return memo
	}
	if pagedIndexed {
		memo.indexedPages = newBytecodeRecurrencePagedValues(kindSet, indexRows)
		return memo
	}
	valuesLen := indexRows * len(kindSet.kinds)
	memo.indexedKnownMasks = make([]uint16, indexRows)
	memo.indexedValues = make([]bytecodeRecurrenceStateResult, valuesLen)
	return memo
}

func (m *bytecodeRecurrenceMemoRows) indexedRow(raw int64) (int, bool) {
	if m == nil || m.indexRows == 0 || m.indexStep <= 0 || raw < m.indexBase {
		return 0, false
	}
	delta := raw - m.indexBase
	if delta < 0 || delta%m.indexStep != 0 {
		return 0, false
	}
	idx := delta / m.indexStep
	if idx < 0 || idx >= int64(m.indexRows) {
		return 0, false
	}
	return int(idx), true
}

func newBytecodeRecurrencePagedValues(kindSet bytecodeRecurrenceKindSet, rowCount int) *bytecodeRecurrencePagedValues {
	if rowCount < 0 {
		rowCount = 0
	}
	pageCount := 0
	if rowCount > 0 {
		pageCount = (rowCount + bytecodeRecurrencePagedDPPageSize - 1) / bytecodeRecurrencePagedDPPageSize
	}
	return &bytecodeRecurrencePagedValues{
		rowWidth: len(kindSet.kinds),
		pageSize: bytecodeRecurrencePagedDPPageSize,
		kindSet:  kindSet,
		pages:    make([]*bytecodeRecurrenceValuePage, pageCount),
	}
}

func (p *bytecodeRecurrencePagedValues) pageOffset(raw int64) (int, int, bool) {
	if p == nil || raw < 0 {
		return 0, 0, false
	}
	rawInt := int(raw)
	pageIdx := rawInt / p.pageSize
	if pageIdx < 0 || pageIdx >= len(p.pages) {
		return 0, 0, false
	}
	return pageIdx, rawInt % p.pageSize, true
}

func (p *bytecodeRecurrencePagedValues) ensurePage(pageIdx int) *bytecodeRecurrenceValuePage {
	if p == nil || pageIdx < 0 || pageIdx >= len(p.pages) {
		return nil
	}
	if p.pages[pageIdx] == nil {
		p.pages[pageIdx] = &bytecodeRecurrenceValuePage{
			knownMasks: make([]uint16, p.pageSize),
			values:     make([]bytecodeRecurrenceStateResult, p.pageSize*p.rowWidth),
		}
	}
	return p.pages[pageIdx]
}

func (p *bytecodeRecurrencePagedValues) get(state bytecodeRecurrenceState) (bytecodeRecurrenceStateResult, bool) {
	if p == nil {
		return bytecodeRecurrenceStateResult{}, false
	}
	pageIdx, offset, ok := p.pageOffset(state.raw)
	if !ok {
		return bytecodeRecurrenceStateResult{}, false
	}
	page := p.pages[pageIdx]
	if page == nil {
		return bytecodeRecurrenceStateResult{}, false
	}
	denseIdx, ok := p.kindSet.index(state.kind)
	if !ok {
		return bytecodeRecurrenceStateResult{}, false
	}
	mask := uint16(1) << denseIdx
	if page.knownMasks[offset]&mask == 0 {
		return bytecodeRecurrenceStateResult{}, false
	}
	flat := offset*p.rowWidth + denseIdx
	return page.values[flat], true
}

func (p *bytecodeRecurrencePagedValues) set(state bytecodeRecurrenceState, result bytecodeRecurrenceStateResult) bool {
	if p == nil {
		return false
	}
	pageIdx, offset, ok := p.pageOffset(state.raw)
	if !ok {
		return false
	}
	page := p.ensurePage(pageIdx)
	if page == nil {
		return false
	}
	denseIdx, ok := p.kindSet.index(state.kind)
	if !ok {
		return false
	}
	flat := offset*p.rowWidth + denseIdx
	page.values[flat] = result
	page.knownMasks[offset] |= uint16(1) << denseIdx
	return true
}

func (m *bytecodeRecurrenceMemoRows) get(state bytecodeRecurrenceState) (bytecodeRecurrenceStateResult, bool) {
	if m == nil {
		return bytecodeRecurrenceStateResult{}, false
	}
	denseIdx, ok := m.kindSet.index(state.kind)
	if !ok {
		return bytecodeRecurrenceStateResult{}, false
	}
	mask := uint16(1) << denseIdx
	if rowIdx, ok := m.indexedRow(state.raw); ok {
		if len(m.indexedKnownMasks) > 0 {
			if m.indexedKnownMasks[rowIdx]&mask == 0 {
				return bytecodeRecurrenceStateResult{}, false
			}
			flat := rowIdx*m.rowWidth + denseIdx
			return m.indexedValues[flat], true
		}
		if m.indexedPages != nil {
			indexedState := bytecodeRecurrenceState{raw: int64(rowIdx), kind: state.kind}
			return m.indexedPages.get(indexedState)
		}
	}
	rowIdx, ok := m.spillRowIndex[state.raw]
	if !ok {
		return bytecodeRecurrenceStateResult{}, false
	}
	if m.spillKnownMasks[rowIdx]&mask == 0 {
		return bytecodeRecurrenceStateResult{}, false
	}
	flat := rowIdx*m.rowWidth + denseIdx
	return m.spillValues[flat], true
}

func (m *bytecodeRecurrenceMemoRows) ensureSpillRow(raw int64) int {
	if m == nil {
		return -1
	}
	if idx, ok := m.spillRowIndex[raw]; ok {
		return idx
	}
	idx := len(m.spillKnownMasks)
	m.spillRowIndex[raw] = idx
	m.spillKnownMasks = append(m.spillKnownMasks, 0)
	for count := 0; count < m.rowWidth; count++ {
		m.spillValues = append(m.spillValues, bytecodeRecurrenceStateResult{})
	}
	return idx
}

func (m *bytecodeRecurrenceMemoRows) set(state bytecodeRecurrenceState, result bytecodeRecurrenceStateResult) bool {
	if m == nil {
		return false
	}
	denseIdx, ok := m.kindSet.index(state.kind)
	if !ok {
		return false
	}
	if rowIdx, ok := m.indexedRow(state.raw); ok {
		if len(m.indexedKnownMasks) > 0 {
			flat := rowIdx*m.rowWidth + denseIdx
			m.indexedValues[flat] = result
			m.indexedKnownMasks[rowIdx] |= uint16(1) << denseIdx
			return true
		}
		if m.indexedPages != nil {
			indexedState := bytecodeRecurrenceState{raw: int64(rowIdx), kind: state.kind}
			return m.indexedPages.set(indexedState, result)
		}
	}
	rowIdx := m.ensureSpillRow(state.raw)
	if rowIdx < 0 {
		return false
	}
	flat := rowIdx*m.rowWidth + denseIdx
	m.spillValues[flat] = result
	m.spillKnownMasks[rowIdx] |= uint16(1) << denseIdx
	return true
}
