package interpreter

import (
	"sort"
	"sync/atomic"

	"able/interpreter-go/pkg/ast"
)

// BytecodeStatsSnapshot captures optional bytecode runtime counters.
// OpCounts is indexed by bytecode opcode numeric value.
type BytecodeStatsSnapshot struct {
	Enabled                           bool
	PrimitiveMaterializationsEnabled  bool
	PrimitiveMaterializations         []BytecodePrimitiveMaterializationSnapshot
	PrimitiveMaterializationsDropped  uint64
	ProgramReach                      []BytecodeProgramReachSnapshot
	ProgramReachDropped               uint64
	OpCounts                          []uint64
	OpCountsByName                    map[string]uint64
	TopOps                            []BytecodeOpCountSnapshot
	ValueStackDeltas                  []int64
	ValueStackDeltasByName            map[string]int64
	TopValueStackDeltas               []BytecodeOpStackDeltaSnapshot
	TopValueStackPeakSites            []BytecodeStackPeakSiteSnapshot
	TopValueStackDeltaSites           []BytecodeStackDeltaSiteSnapshot
	TopCallOperandBalances            []BytecodeCallOperandBalanceSnapshot
	TopLoopBackedgeBalances           []BytecodeLoopBackedgeBalanceSnapshot
	TopInlineFrameBalances            []BytecodeInlineFrameBalanceSnapshot
	ValueStackMaxDepth                uint64
	ValueStackMaxCapacity             uint64
	ValueStackCapacityGrowths         uint64
	CallFrameMaxDepth                 uint64
	LoadNameLookups                   uint64
	LoadNameLookupsByName             map[string]uint64
	TopLoadNames                      []BytecodeNameCountSnapshot
	LoadNameHotHits                   uint64
	LoadNameScopeHits                 uint64
	LoadNameGlobalHits                uint64
	LoadNameDirectCurrent             uint64
	LoadNameDirectOuter               uint64
	LoadNameScopeStores               uint64
	LoadNameGlobalStores              uint64
	CallNameLookups                   uint64
	CallNameDotFallback               uint64
	CallNameExactNativeHits           uint64
	CallNameInlineDirectSlotHits      uint64
	CallNameInlineDirectStackHits     uint64
	CallNameInlineResolvedHits        uint64
	CallNameInlineGenericHits         uint64
	CallNameResolvedFunctionHits      uint64
	CallNameGenericFallbacks          uint64
	InlineCallHits                    uint64
	InlineCallMisses                  uint64
	DirectFunctionStackHits           uint64
	InlineResolvedMissNoBytecode      uint64
	InlineResolvedMissArity           uint64
	InlineResolvedMissTypeArgs        uint64
	InlineResolvedMissGenericLambda   uint64
	MemberMethodCacheHits             uint64
	MemberMethodCacheMiss             uint64
	CallMemberResolvedExactNativeHits uint64
	CallMemberResolvedInlineHits      uint64
	CallMemberResolvedGenericHits     uint64
	CallMemberResolvedFallbacks       uint64
	GenericUnionCallCacheHits         uint64
	GenericUnionCallCacheMisses       uint64
	CallMemberStaticCacheHits         uint64
	CallMemberStaticCacheMisses       uint64
	CallMemberStaticExactNativeHits   uint64
	CallMemberStaticInlineHits        uint64
	CallMemberStaticGenericHits       uint64
	ExprCacheHits                     uint64
	ExprCacheMisses                   uint64
	ArrayIndexSlotLookups             uint64
	ArrayIndexSlotTrackedHits         uint64
	ArrayIndexSlotMonoUnsignedHits    uint64
	ArrayIndexSlotDirectHits          uint64
	ArrayIndexSlotFallbacks           uint64
	ArrayIndexSlotFastDisabledMiss    uint64
	ArrayIndexSlotReceiverMiss        uint64
	ArrayIndexSlotIndexMiss           uint64
	ArrayIndexSlotHandleMiss          uint64
	ArrayIndexSlotDirectMiss          uint64
	ArrayMemberSlotLookups            uint64
	ArrayMemberSlotLenLookups         uint64
	ArrayMemberSlotReadLookups        uint64
	ArrayMemberSlotWriteLookups       uint64
	ArrayMemberSlotPushLookups        uint64
	ArrayMemberSlotCacheHits          uint64
	ArrayMemberSlotFastHits           uint64
	ArrayMemberSlotFallbacks          uint64
	ArrayMemberSlotReceiverMiss       uint64
	ArrayMemberSlotCacheMiss          uint64
	ArrayMemberSlotFastPathMiss       uint64
}

type BytecodeOpCountSnapshot struct {
	Op    int    `json:"op"`
	Name  string `json:"name"`
	Count uint64 `json:"count"`
}

// BytecodeOpStackDeltaSnapshot reports an opcode's net contribution to the
// value-stack depth in an opt-in diagnostic run.
type BytecodeOpStackDeltaSnapshot struct {
	Op    int    `json:"op"`
	Name  string `json:"name"`
	Delta int64  `json:"delta"`
}

const bytecodeStackPeakSiteLimit = 256
const bytecodeStackDeltaSiteLimit = 512

type bytecodeStackPeakSiteKey struct {
	Op     int
	IP     int
	Name   string
	Origin string
	Line   int
	Column int
}

// BytecodeStackPeakSiteSnapshot identifies an instruction that extended a
// VM's high-watermark value-stack depth during an opt-in diagnostic run.
type BytecodeStackPeakSiteSnapshot struct {
	Op     int    `json:"op"`
	IP     int    `json:"ip"`
	Name   string `json:"name"`
	Origin string `json:"origin,omitempty"`
	Line   int    `json:"line,omitempty"`
	Column int    `json:"column,omitempty"`
	Growth uint64 `json:"growth"`
}

// BytecodeStackDeltaSiteSnapshot reports an instruction site's signed net
// contribution to the value-stack depth during an opt-in diagnostic run.
type BytecodeStackDeltaSiteSnapshot struct {
	Op     int    `json:"op"`
	IP     int    `json:"ip"`
	Name   string `json:"name"`
	Origin string `json:"origin,omitempty"`
	Line   int    `json:"line,omitempty"`
	Column int    `json:"column,omitempty"`
	Delta  int64  `json:"delta"`
}

type bytecodeInlineFrameBalanceKey struct {
	Origin string
	Line   int
	Column int
}

type bytecodeInlineFrameBalance struct {
	Returns uint64
	Excess  uint64
	Max     uint64
}

// BytecodeInlineFrameBalanceSnapshot reports residual value-stack entries
// immediately before an inline callee's return value is appended.
type BytecodeInlineFrameBalanceSnapshot struct {
	Origin  string `json:"origin,omitempty"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
	Returns uint64 `json:"returns"`
	Excess  uint64 `json:"excess"`
	Max     uint64 `json:"max"`
}

type BytecodeNameCountSnapshot struct {
	Name  string `json:"name"`
	Count uint64 `json:"count"`
}

type bytecodeArrayIndexSlotFallbackReason uint8

const (
	bytecodeArrayIndexSlotFallbackFastDisabled bytecodeArrayIndexSlotFallbackReason = iota + 1
	bytecodeArrayIndexSlotFallbackReceiverMiss
	bytecodeArrayIndexSlotFallbackIndexMiss
	bytecodeArrayIndexSlotFallbackHandleMiss
	bytecodeArrayIndexSlotFallbackDirectMiss
)

type bytecodeArrayMemberSlotFallbackReason uint8

const (
	bytecodeArrayMemberSlotFallbackReceiverMiss bytecodeArrayMemberSlotFallbackReason = iota + 1
	bytecodeArrayMemberSlotFallbackCacheMiss
	bytecodeArrayMemberSlotFallbackFastPathMiss
)

type bytecodeCallNameDispatchStats uint8

const (
	bytecodeCallNameStatsExactNative bytecodeCallNameDispatchStats = iota + 1
	bytecodeCallNameStatsInlineDirectSlot
	bytecodeCallNameStatsInlineDirectStack
	bytecodeCallNameStatsInlineResolved
	bytecodeCallNameStatsInlineGeneric
	bytecodeCallNameStatsResolvedFunction
	bytecodeCallNameStatsGenericFallback
)

type bytecodeInlineResolvedMissReason uint8

const (
	bytecodeInlineResolvedMissNoBytecode bytecodeInlineResolvedMissReason = iota + 1
	bytecodeInlineResolvedMissArity
	bytecodeInlineResolvedMissTypeArgs
	bytecodeInlineResolvedMissGenericLambda
)

type bytecodeCallMemberDispatchStats uint8

const (
	bytecodeCallMemberStatsResolvedExactNative bytecodeCallMemberDispatchStats = iota + 1
	bytecodeCallMemberStatsResolvedInline
	bytecodeCallMemberStatsResolvedGeneric
	bytecodeCallMemberStatsResolvedFallback
	bytecodeCallMemberStatsStaticExactNative
	bytecodeCallMemberStatsStaticInline
	bytecodeCallMemberStatsStaticGeneric
)

func (i *Interpreter) recordBytecodeOp(op bytecodeOp) {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	idx := int(op)
	if idx < 0 || idx >= len(i.bytecodeOpCounts) {
		return
	}
	atomic.AddUint64(&i.bytecodeOpCounts[idx], 1)
}

func (i *Interpreter) recordBytecodeValueStackDelta(op bytecodeOp, before int, after int) {
	if i == nil || !i.bytecodeStatsEnabled || before == after {
		return
	}
	idx := int(op)
	if idx < 0 || idx >= len(i.bytecodeValueStackDeltas) {
		return
	}
	atomic.AddInt64(&i.bytecodeValueStackDeltas[idx], int64(after-before))
}

func (i *Interpreter) bytecodeStackPeakSite(op bytecodeOp, ip int, instr *bytecodeInstruction) bytecodeStackPeakSiteKey {
	key := bytecodeStackPeakSiteKey{Op: int(op), IP: ip}
	if instr == nil {
		return key
	}
	key.Name = instr.name
	if instr.node == nil {
		return key
	}
	if span := instr.node.Span(); span != (ast.Span{}) {
		key.Line = span.Start.Line
		key.Column = span.Start.Column
	}
	if i != nil && i.nodeOrigins != nil {
		key.Origin = i.nodeOrigins[instr.node]
	}
	return key
}

func (i *Interpreter) recordBytecodeValueStackPeakSite(site bytecodeStackPeakSiteKey, growth int) {
	if i == nil || !i.bytecodeStatsEnabled || growth <= 0 {
		return
	}
	i.bytecodeStackPeakSitesMu.Lock()
	if i.bytecodeStackPeakSites == nil {
		i.bytecodeStackPeakSites = make(map[bytecodeStackPeakSiteKey]uint64, 16)
	}
	if current, ok := i.bytecodeStackPeakSites[site]; ok {
		i.bytecodeStackPeakSites[site] = current + uint64(growth)
		i.bytecodeStackPeakSitesMu.Unlock()
		return
	}
	if len(i.bytecodeStackPeakSites) < bytecodeStackPeakSiteLimit {
		i.bytecodeStackPeakSites[site] = uint64(growth)
	}
	i.bytecodeStackPeakSitesMu.Unlock()
}

func (i *Interpreter) recordBytecodeValueStackDeltaSite(site bytecodeStackPeakSiteKey, delta int) {
	if i == nil || !i.bytecodeStatsEnabled || delta == 0 {
		return
	}
	i.bytecodeStackDeltaSitesMu.Lock()
	if i.bytecodeStackDeltaSites == nil {
		i.bytecodeStackDeltaSites = make(map[bytecodeStackPeakSiteKey]int64, 16)
	}
	if current, ok := i.bytecodeStackDeltaSites[site]; ok {
		i.bytecodeStackDeltaSites[site] = current + int64(delta)
	} else if len(i.bytecodeStackDeltaSites) < bytecodeStackDeltaSiteLimit {
		i.bytecodeStackDeltaSites[site] = int64(delta)
	} else if delta > 0 {
		var (
			leastSite  bytecodeStackPeakSiteKey
			leastDelta int64
			haveLeast  bool
		)
		for candidate, candidateDelta := range i.bytecodeStackDeltaSites {
			if !haveLeast || candidateDelta < leastDelta {
				leastSite = candidate
				leastDelta = candidateDelta
				haveLeast = true
			}
		}
		if haveLeast && int64(delta) >= leastDelta {
			delete(i.bytecodeStackDeltaSites, leastSite)
			i.bytecodeStackDeltaSites[site] = int64(delta)
		}
	}
	i.bytecodeStackDeltaSitesMu.Unlock()
}

func (i *Interpreter) recordBytecodeInlineFrameBalance(instr *bytecodeInstruction, stackBase int, stackDepth int) {
	if i == nil || !i.bytecodeStatsEnabled || stackBase < 0 || stackDepth <= stackBase {
		return
	}
	key := bytecodeInlineFrameBalanceKey{}
	if instr != nil && instr.node != nil {
		if span := instr.node.Span(); span != (ast.Span{}) {
			key.Line = span.Start.Line
			key.Column = span.Start.Column
		}
		if i.nodeOrigins != nil {
			key.Origin = i.nodeOrigins[instr.node]
		}
	}
	excess := uint64(stackDepth - stackBase)
	i.bytecodeInlineFrameBalancesMu.Lock()
	if i.bytecodeInlineFrameBalances == nil {
		i.bytecodeInlineFrameBalances = make(map[bytecodeInlineFrameBalanceKey]bytecodeInlineFrameBalance, 8)
	}
	balance := i.bytecodeInlineFrameBalances[key]
	balance.Returns++
	balance.Excess += excess
	if excess > balance.Max {
		balance.Max = excess
	}
	i.bytecodeInlineFrameBalances[key] = balance
	i.bytecodeInlineFrameBalancesMu.Unlock()
}

func (i *Interpreter) recordBytecodeVMDepths(valueStackDepth int, valueStackCapacity int, callFrameDepth int, valueStackCapacityGrew bool) {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	recordBytecodeMax(&i.bytecodeValueStackMaxDepth, valueStackDepth)
	recordBytecodeMax(&i.bytecodeValueStackMaxCapacity, valueStackCapacity)
	recordBytecodeMax(&i.bytecodeCallFrameMaxDepth, callFrameDepth)
	if valueStackCapacityGrew {
		atomic.AddUint64(&i.bytecodeValueStackCapacityGrowths, 1)
	}
}

func recordBytecodeMax(counter *uint64, value int) {
	if value <= 0 {
		return
	}
	want := uint64(value)
	for previous := atomic.LoadUint64(counter); want > previous; previous = atomic.LoadUint64(counter) {
		if atomic.CompareAndSwapUint64(counter, previous, want) {
			return
		}
	}
}

func (i *Interpreter) recordBytecodeLoadNameLookup() {
	i.recordBytecodeLoadNameLookupForName("")
}

func (i *Interpreter) recordBytecodeLoadNameLookupForName(name string) {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeLoadNameLookups, 1)
	if name == "" {
		return
	}
	i.bytecodeLoadNameCountsMu.Lock()
	if i.bytecodeLoadNameCounts == nil {
		i.bytecodeLoadNameCounts = make(map[string]uint64, 8)
	}
	i.bytecodeLoadNameCounts[name]++
	i.bytecodeLoadNameCountsMu.Unlock()
}

func (i *Interpreter) recordBytecodeLoadNameHotHit() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeLoadNameHotHits, 1)
}

func (i *Interpreter) recordBytecodeLoadNameScopeCacheHit() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeLoadNameScopeCacheHits, 1)
}

func (i *Interpreter) recordBytecodeLoadNameGlobalCacheHit() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeLoadNameGlobalCacheHits, 1)
}

func (i *Interpreter) recordBytecodeLoadNameDirectResolve(currentOwner bool) {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	if currentOwner {
		atomic.AddUint64(&i.bytecodeLoadNameDirectCurrent, 1)
		return
	}
	atomic.AddUint64(&i.bytecodeLoadNameDirectOuter, 1)
}

func (i *Interpreter) recordBytecodeLoadNameScopeStore() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeLoadNameScopeStores, 1)
}

func (i *Interpreter) recordBytecodeLoadNameGlobalStore() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeLoadNameGlobalStores, 1)
}

func (i *Interpreter) recordBytecodeCallNameLookup() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeCallNameLookups, 1)
}

func (i *Interpreter) recordBytecodeCallNameDotFallback() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeCallNameDottedFallbacks, 1)
}

func (i *Interpreter) recordBytecodeCallNameDispatch(kind bytecodeCallNameDispatchStats) {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	switch kind {
	case bytecodeCallNameStatsExactNative:
		atomic.AddUint64(&i.bytecodeCallNameExactNativeHits, 1)
	case bytecodeCallNameStatsInlineDirectSlot:
		atomic.AddUint64(&i.bytecodeCallNameInlineDirectSlotHits, 1)
	case bytecodeCallNameStatsInlineDirectStack:
		atomic.AddUint64(&i.bytecodeCallNameInlineDirectStackHits, 1)
	case bytecodeCallNameStatsInlineResolved:
		atomic.AddUint64(&i.bytecodeCallNameInlineResolvedHits, 1)
	case bytecodeCallNameStatsInlineGeneric:
		atomic.AddUint64(&i.bytecodeCallNameInlineGenericHits, 1)
	case bytecodeCallNameStatsResolvedFunction:
		atomic.AddUint64(&i.bytecodeCallNameResolvedFunctionHits, 1)
	case bytecodeCallNameStatsGenericFallback:
		atomic.AddUint64(&i.bytecodeCallNameGenericFallbacks, 1)
	}
}

func (i *Interpreter) recordBytecodeInlineCallHit() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeInlineCallHits, 1)
}

func (i *Interpreter) recordBytecodeInlineCallMiss() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeInlineCallMisses, 1)
}

func (i *Interpreter) recordBytecodeDirectFunctionStackHit() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeDirectFunctionStackHits, 1)
}

func (i *Interpreter) recordBytecodeInlineResolvedMiss(reason bytecodeInlineResolvedMissReason) {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	switch reason {
	case bytecodeInlineResolvedMissNoBytecode:
		atomic.AddUint64(&i.bytecodeInlineResolvedMissNoBytecode, 1)
	case bytecodeInlineResolvedMissArity:
		atomic.AddUint64(&i.bytecodeInlineResolvedMissArity, 1)
	case bytecodeInlineResolvedMissTypeArgs:
		atomic.AddUint64(&i.bytecodeInlineResolvedMissTypeArgs, 1)
	case bytecodeInlineResolvedMissGenericLambda:
		atomic.AddUint64(&i.bytecodeInlineResolvedMissGenericLambda, 1)
	}
}

func (i *Interpreter) recordBytecodeMemberMethodCacheHit() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeMemberMethodCacheHits, 1)
}

func (i *Interpreter) recordBytecodeMemberMethodCacheMiss() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeMemberMethodCacheMisses, 1)
}

func (i *Interpreter) recordBytecodeCallMemberDispatch(kind bytecodeCallMemberDispatchStats) {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	switch kind {
	case bytecodeCallMemberStatsResolvedExactNative:
		atomic.AddUint64(&i.bytecodeCallMemberResolvedExactNative, 1)
	case bytecodeCallMemberStatsResolvedInline:
		atomic.AddUint64(&i.bytecodeCallMemberResolvedInline, 1)
	case bytecodeCallMemberStatsResolvedGeneric:
		atomic.AddUint64(&i.bytecodeCallMemberResolvedGeneric, 1)
	case bytecodeCallMemberStatsResolvedFallback:
		atomic.AddUint64(&i.bytecodeCallMemberResolvedFallback, 1)
	case bytecodeCallMemberStatsStaticExactNative:
		atomic.AddUint64(&i.bytecodeCallMemberStaticExactNative, 1)
	case bytecodeCallMemberStatsStaticInline:
		atomic.AddUint64(&i.bytecodeCallMemberStaticInline, 1)
	case bytecodeCallMemberStatsStaticGeneric:
		atomic.AddUint64(&i.bytecodeCallMemberStaticGeneric, 1)
	}
}

func (i *Interpreter) recordBytecodeCallMemberStaticCacheHit() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeCallMemberStaticCacheHits, 1)
}

func (i *Interpreter) recordBytecodeCallMemberStaticCacheMiss() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeCallMemberStaticCacheMisses, 1)
}

func (i *Interpreter) recordBytecodeExpressionCacheHit() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeExprCacheHits, 1)
}

func (i *Interpreter) recordBytecodeExpressionCacheMiss() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeExprCacheMisses, 1)
}

func (i *Interpreter) recordBytecodeArrayIndexSlotLookup() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeArrayIndexSlotLookups, 1)
}

func (i *Interpreter) recordBytecodeArrayIndexSlotTrackedHit() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeArrayIndexSlotTrackedHits, 1)
}

func (i *Interpreter) recordBytecodeArrayIndexSlotMonoUnsignedHit() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeArrayIndexSlotMonoUnsignedHits, 1)
}

func (i *Interpreter) recordBytecodeArrayIndexSlotDirectHit() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeArrayIndexSlotDirectHits, 1)
}

func (i *Interpreter) recordBytecodeArrayIndexSlotFallback(reason bytecodeArrayIndexSlotFallbackReason) {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeArrayIndexSlotFallbacks, 1)
	switch reason {
	case bytecodeArrayIndexSlotFallbackFastDisabled:
		atomic.AddUint64(&i.bytecodeArrayIndexSlotFastDisabledMiss, 1)
	case bytecodeArrayIndexSlotFallbackReceiverMiss:
		atomic.AddUint64(&i.bytecodeArrayIndexSlotReceiverMiss, 1)
	case bytecodeArrayIndexSlotFallbackIndexMiss:
		atomic.AddUint64(&i.bytecodeArrayIndexSlotIndexMiss, 1)
	case bytecodeArrayIndexSlotFallbackHandleMiss:
		atomic.AddUint64(&i.bytecodeArrayIndexSlotHandleMiss, 1)
	case bytecodeArrayIndexSlotFallbackDirectMiss:
		atomic.AddUint64(&i.bytecodeArrayIndexSlotDirectMiss, 1)
	}
}

// BytecodeStats returns a snapshot of bytecode counters.
func (i *Interpreter) BytecodeStats() BytecodeStatsSnapshot {
	snapshot := BytecodeStatsSnapshot{}
	if i == nil {
		return snapshot
	}
	snapshot.Enabled = i.bytecodeStatsEnabled
	snapshot.PrimitiveMaterializationsEnabled = i.bytecodePrimitiveMaterializationStatsEnabled
	snapshot.PrimitiveMaterializations, snapshot.PrimitiveMaterializationsDropped = i.bytecodePrimitiveMaterializationSnapshot()
	snapshot.ProgramReach, snapshot.ProgramReachDropped = i.bytecodeProgramReachSnapshot()
	snapshot.ValueStackMaxDepth = atomic.LoadUint64(&i.bytecodeValueStackMaxDepth)
	snapshot.ValueStackMaxCapacity = atomic.LoadUint64(&i.bytecodeValueStackMaxCapacity)
	snapshot.ValueStackCapacityGrowths = atomic.LoadUint64(&i.bytecodeValueStackCapacityGrowths)
	snapshot.CallFrameMaxDepth = atomic.LoadUint64(&i.bytecodeCallFrameMaxDepth)
	snapshot.OpCounts = make([]uint64, bytecodeOpCount)
	snapshot.OpCountsByName = make(map[string]uint64, bytecodeOpCount)
	snapshot.TopOps = make([]BytecodeOpCountSnapshot, 0, bytecodeOpCount)
	snapshot.ValueStackDeltas = make([]int64, bytecodeOpCount)
	snapshot.ValueStackDeltasByName = make(map[string]int64, bytecodeOpCount)
	snapshot.TopValueStackDeltas = make([]BytecodeOpStackDeltaSnapshot, 0, bytecodeOpCount)
	for idx := 0; idx < bytecodeOpCount; idx++ {
		count := atomic.LoadUint64(&i.bytecodeOpCounts[idx])
		snapshot.OpCounts[idx] = count
		name := bytecodeOpName(bytecodeOp(idx))
		snapshot.OpCountsByName[name] = count
		if count > 0 {
			snapshot.TopOps = append(snapshot.TopOps, BytecodeOpCountSnapshot{
				Op:    idx,
				Name:  name,
				Count: count,
			})
		}
		delta := atomic.LoadInt64(&i.bytecodeValueStackDeltas[idx])
		snapshot.ValueStackDeltas[idx] = delta
		snapshot.ValueStackDeltasByName[name] = delta
		if delta != 0 {
			snapshot.TopValueStackDeltas = append(snapshot.TopValueStackDeltas, BytecodeOpStackDeltaSnapshot{Op: idx, Name: name, Delta: delta})
		}
	}
	sort.Slice(snapshot.TopOps, func(left, right int) bool {
		if snapshot.TopOps[left].Count != snapshot.TopOps[right].Count {
			return snapshot.TopOps[left].Count > snapshot.TopOps[right].Count
		}
		return snapshot.TopOps[left].Name < snapshot.TopOps[right].Name
	})
	sort.Slice(snapshot.TopValueStackDeltas, func(left, right int) bool {
		if snapshot.TopValueStackDeltas[left].Delta != snapshot.TopValueStackDeltas[right].Delta {
			return snapshot.TopValueStackDeltas[left].Delta > snapshot.TopValueStackDeltas[right].Delta
		}
		return snapshot.TopValueStackDeltas[left].Name < snapshot.TopValueStackDeltas[right].Name
	})
	snapshot.TopValueStackPeakSites = i.bytecodeStackPeakSiteSnapshot()
	snapshot.TopValueStackDeltaSites = i.bytecodeStackDeltaSiteSnapshot()
	snapshot.TopCallOperandBalances = i.bytecodeCallOperandBalanceSnapshot()
	snapshot.TopLoopBackedgeBalances = i.bytecodeLoopBackedgeBalanceSnapshot()
	snapshot.TopInlineFrameBalances = i.bytecodeInlineFrameBalanceSnapshot()
	snapshot.LoadNameLookups = atomic.LoadUint64(&i.bytecodeLoadNameLookups)
	snapshot.LoadNameLookupsByName, snapshot.TopLoadNames = i.bytecodeLoadNameCountSnapshot()
	snapshot.LoadNameHotHits = atomic.LoadUint64(&i.bytecodeLoadNameHotHits)
	snapshot.LoadNameScopeHits = atomic.LoadUint64(&i.bytecodeLoadNameScopeCacheHits)
	snapshot.LoadNameGlobalHits = atomic.LoadUint64(&i.bytecodeLoadNameGlobalCacheHits)
	snapshot.LoadNameDirectCurrent = atomic.LoadUint64(&i.bytecodeLoadNameDirectCurrent)
	snapshot.LoadNameDirectOuter = atomic.LoadUint64(&i.bytecodeLoadNameDirectOuter)
	snapshot.LoadNameScopeStores = atomic.LoadUint64(&i.bytecodeLoadNameScopeStores)
	snapshot.LoadNameGlobalStores = atomic.LoadUint64(&i.bytecodeLoadNameGlobalStores)
	snapshot.CallNameLookups = atomic.LoadUint64(&i.bytecodeCallNameLookups)
	snapshot.CallNameDotFallback = atomic.LoadUint64(&i.bytecodeCallNameDottedFallbacks)
	snapshot.CallNameExactNativeHits = atomic.LoadUint64(&i.bytecodeCallNameExactNativeHits)
	snapshot.CallNameInlineDirectSlotHits = atomic.LoadUint64(&i.bytecodeCallNameInlineDirectSlotHits)
	snapshot.CallNameInlineDirectStackHits = atomic.LoadUint64(&i.bytecodeCallNameInlineDirectStackHits)
	snapshot.CallNameInlineResolvedHits = atomic.LoadUint64(&i.bytecodeCallNameInlineResolvedHits)
	snapshot.CallNameInlineGenericHits = atomic.LoadUint64(&i.bytecodeCallNameInlineGenericHits)
	snapshot.CallNameResolvedFunctionHits = atomic.LoadUint64(&i.bytecodeCallNameResolvedFunctionHits)
	snapshot.CallNameGenericFallbacks = atomic.LoadUint64(&i.bytecodeCallNameGenericFallbacks)
	snapshot.InlineCallHits = atomic.LoadUint64(&i.bytecodeInlineCallHits)
	snapshot.InlineCallMisses = atomic.LoadUint64(&i.bytecodeInlineCallMisses)
	snapshot.DirectFunctionStackHits = atomic.LoadUint64(&i.bytecodeDirectFunctionStackHits)
	snapshot.InlineResolvedMissNoBytecode = atomic.LoadUint64(&i.bytecodeInlineResolvedMissNoBytecode)
	snapshot.InlineResolvedMissArity = atomic.LoadUint64(&i.bytecodeInlineResolvedMissArity)
	snapshot.InlineResolvedMissTypeArgs = atomic.LoadUint64(&i.bytecodeInlineResolvedMissTypeArgs)
	snapshot.InlineResolvedMissGenericLambda = atomic.LoadUint64(&i.bytecodeInlineResolvedMissGenericLambda)
	snapshot.MemberMethodCacheHits = atomic.LoadUint64(&i.bytecodeMemberMethodCacheHits)
	snapshot.MemberMethodCacheMiss = atomic.LoadUint64(&i.bytecodeMemberMethodCacheMisses)
	snapshot.CallMemberResolvedExactNativeHits = atomic.LoadUint64(&i.bytecodeCallMemberResolvedExactNative)
	snapshot.CallMemberResolvedInlineHits = atomic.LoadUint64(&i.bytecodeCallMemberResolvedInline)
	snapshot.CallMemberResolvedGenericHits = atomic.LoadUint64(&i.bytecodeCallMemberResolvedGeneric)
	snapshot.CallMemberResolvedFallbacks = atomic.LoadUint64(&i.bytecodeCallMemberResolvedFallback)
	snapshot.GenericUnionCallCacheHits = atomic.LoadUint64(&i.bytecodeGenericUnionCallCacheHits)
	snapshot.GenericUnionCallCacheMisses = atomic.LoadUint64(&i.bytecodeGenericUnionCallCacheMisses)
	snapshot.CallMemberStaticCacheHits = atomic.LoadUint64(&i.bytecodeCallMemberStaticCacheHits)
	snapshot.CallMemberStaticCacheMisses = atomic.LoadUint64(&i.bytecodeCallMemberStaticCacheMisses)
	snapshot.CallMemberStaticExactNativeHits = atomic.LoadUint64(&i.bytecodeCallMemberStaticExactNative)
	snapshot.CallMemberStaticInlineHits = atomic.LoadUint64(&i.bytecodeCallMemberStaticInline)
	snapshot.CallMemberStaticGenericHits = atomic.LoadUint64(&i.bytecodeCallMemberStaticGeneric)
	snapshot.ExprCacheHits = atomic.LoadUint64(&i.bytecodeExprCacheHits)
	snapshot.ExprCacheMisses = atomic.LoadUint64(&i.bytecodeExprCacheMisses)
	snapshot.ArrayIndexSlotLookups = atomic.LoadUint64(&i.bytecodeArrayIndexSlotLookups)
	snapshot.ArrayIndexSlotTrackedHits = atomic.LoadUint64(&i.bytecodeArrayIndexSlotTrackedHits)
	snapshot.ArrayIndexSlotMonoUnsignedHits = atomic.LoadUint64(&i.bytecodeArrayIndexSlotMonoUnsignedHits)
	snapshot.ArrayIndexSlotDirectHits = atomic.LoadUint64(&i.bytecodeArrayIndexSlotDirectHits)
	snapshot.ArrayIndexSlotFallbacks = atomic.LoadUint64(&i.bytecodeArrayIndexSlotFallbacks)
	snapshot.ArrayIndexSlotFastDisabledMiss = atomic.LoadUint64(&i.bytecodeArrayIndexSlotFastDisabledMiss)
	snapshot.ArrayIndexSlotReceiverMiss = atomic.LoadUint64(&i.bytecodeArrayIndexSlotReceiverMiss)
	snapshot.ArrayIndexSlotIndexMiss = atomic.LoadUint64(&i.bytecodeArrayIndexSlotIndexMiss)
	snapshot.ArrayIndexSlotHandleMiss = atomic.LoadUint64(&i.bytecodeArrayIndexSlotHandleMiss)
	snapshot.ArrayIndexSlotDirectMiss = atomic.LoadUint64(&i.bytecodeArrayIndexSlotDirectMiss)
	snapshot.ArrayMemberSlotLookups = atomic.LoadUint64(&i.bytecodeArrayMemberSlotLookups)
	snapshot.ArrayMemberSlotLenLookups = atomic.LoadUint64(&i.bytecodeArrayMemberSlotLenLookups)
	snapshot.ArrayMemberSlotReadLookups = atomic.LoadUint64(&i.bytecodeArrayMemberSlotReadLookups)
	snapshot.ArrayMemberSlotWriteLookups = atomic.LoadUint64(&i.bytecodeArrayMemberSlotWriteLookups)
	snapshot.ArrayMemberSlotPushLookups = atomic.LoadUint64(&i.bytecodeArrayMemberSlotPushLookups)
	snapshot.ArrayMemberSlotCacheHits = atomic.LoadUint64(&i.bytecodeArrayMemberSlotCacheHits)
	snapshot.ArrayMemberSlotFastHits = atomic.LoadUint64(&i.bytecodeArrayMemberSlotFastHits)
	snapshot.ArrayMemberSlotFallbacks = atomic.LoadUint64(&i.bytecodeArrayMemberSlotFallbacks)
	snapshot.ArrayMemberSlotReceiverMiss = atomic.LoadUint64(&i.bytecodeArrayMemberSlotReceiverMiss)
	snapshot.ArrayMemberSlotCacheMiss = atomic.LoadUint64(&i.bytecodeArrayMemberSlotCacheMiss)
	snapshot.ArrayMemberSlotFastPathMiss = atomic.LoadUint64(&i.bytecodeArrayMemberSlotFastPathMiss)
	return snapshot
}

func (i *Interpreter) bytecodeStackPeakSiteSnapshot() []BytecodeStackPeakSiteSnapshot {
	if i == nil {
		return nil
	}
	i.bytecodeStackPeakSitesMu.Lock()
	defer i.bytecodeStackPeakSitesMu.Unlock()
	if len(i.bytecodeStackPeakSites) == 0 {
		return nil
	}
	sites := make([]BytecodeStackPeakSiteSnapshot, 0, len(i.bytecodeStackPeakSites))
	for site, growth := range i.bytecodeStackPeakSites {
		sites = append(sites, BytecodeStackPeakSiteSnapshot{
			Op:     site.Op,
			IP:     site.IP,
			Name:   site.Name,
			Origin: site.Origin,
			Line:   site.Line,
			Column: site.Column,
			Growth: growth,
		})
	}
	sort.Slice(sites, func(left, right int) bool {
		if sites[left].Growth != sites[right].Growth {
			return sites[left].Growth > sites[right].Growth
		}
		if sites[left].Origin != sites[right].Origin {
			return sites[left].Origin < sites[right].Origin
		}
		if sites[left].Line != sites[right].Line {
			return sites[left].Line < sites[right].Line
		}
		if sites[left].Column != sites[right].Column {
			return sites[left].Column < sites[right].Column
		}
		if sites[left].Op != sites[right].Op {
			return sites[left].Op < sites[right].Op
		}
		return sites[left].IP < sites[right].IP
	})
	return sites
}

func (i *Interpreter) bytecodeStackDeltaSiteSnapshot() []BytecodeStackDeltaSiteSnapshot {
	if i == nil {
		return nil
	}
	i.bytecodeStackDeltaSitesMu.Lock()
	defer i.bytecodeStackDeltaSitesMu.Unlock()
	if len(i.bytecodeStackDeltaSites) == 0 {
		return nil
	}
	sites := make([]BytecodeStackDeltaSiteSnapshot, 0, len(i.bytecodeStackDeltaSites))
	for site, delta := range i.bytecodeStackDeltaSites {
		if delta <= 0 {
			continue
		}
		sites = append(sites, BytecodeStackDeltaSiteSnapshot{
			Op:     site.Op,
			IP:     site.IP,
			Name:   site.Name,
			Origin: site.Origin,
			Line:   site.Line,
			Column: site.Column,
			Delta:  delta,
		})
	}
	sort.Slice(sites, func(left, right int) bool {
		if sites[left].Delta != sites[right].Delta {
			return sites[left].Delta > sites[right].Delta
		}
		if sites[left].Origin != sites[right].Origin {
			return sites[left].Origin < sites[right].Origin
		}
		if sites[left].Line != sites[right].Line {
			return sites[left].Line < sites[right].Line
		}
		if sites[left].Column != sites[right].Column {
			return sites[left].Column < sites[right].Column
		}
		if sites[left].Op != sites[right].Op {
			return sites[left].Op < sites[right].Op
		}
		return sites[left].IP < sites[right].IP
	})
	return sites
}

func (i *Interpreter) bytecodeInlineFrameBalanceSnapshot() []BytecodeInlineFrameBalanceSnapshot {
	if i == nil {
		return nil
	}
	i.bytecodeInlineFrameBalancesMu.Lock()
	defer i.bytecodeInlineFrameBalancesMu.Unlock()
	if len(i.bytecodeInlineFrameBalances) == 0 {
		return nil
	}
	balances := make([]BytecodeInlineFrameBalanceSnapshot, 0, len(i.bytecodeInlineFrameBalances))
	for key, balance := range i.bytecodeInlineFrameBalances {
		balances = append(balances, BytecodeInlineFrameBalanceSnapshot{
			Origin:  key.Origin,
			Line:    key.Line,
			Column:  key.Column,
			Returns: balance.Returns,
			Excess:  balance.Excess,
			Max:     balance.Max,
		})
	}
	sort.Slice(balances, func(left, right int) bool {
		if balances[left].Excess != balances[right].Excess {
			return balances[left].Excess > balances[right].Excess
		}
		if balances[left].Max != balances[right].Max {
			return balances[left].Max > balances[right].Max
		}
		if balances[left].Origin != balances[right].Origin {
			return balances[left].Origin < balances[right].Origin
		}
		if balances[left].Line != balances[right].Line {
			return balances[left].Line < balances[right].Line
		}
		return balances[left].Column < balances[right].Column
	})
	return balances
}

func (i *Interpreter) bytecodeLoadNameCountSnapshot() (map[string]uint64, []BytecodeNameCountSnapshot) {
	if i == nil {
		return nil, nil
	}
	i.bytecodeLoadNameCountsMu.Lock()
	defer i.bytecodeLoadNameCountsMu.Unlock()
	if len(i.bytecodeLoadNameCounts) == 0 {
		return nil, nil
	}
	counts := make(map[string]uint64, len(i.bytecodeLoadNameCounts))
	top := make([]BytecodeNameCountSnapshot, 0, len(i.bytecodeLoadNameCounts))
	for name, count := range i.bytecodeLoadNameCounts {
		counts[name] = count
		top = append(top, BytecodeNameCountSnapshot{Name: name, Count: count})
	}
	sort.Slice(top, func(left, right int) bool {
		if top[left].Count != top[right].Count {
			return top[left].Count > top[right].Count
		}
		return top[left].Name < top[right].Name
	})
	return counts, top
}

// ResetBytecodeStats clears bytecode counters.
func (i *Interpreter) ResetBytecodeStats() {
	if i == nil {
		return
	}
	i.resetBytecodeProgramReach()
	i.resetBytecodePrimitiveMaterializations()
	for idx := 0; idx < bytecodeOpCount; idx++ {
		atomic.StoreUint64(&i.bytecodeOpCounts[idx], 0)
		atomic.StoreInt64(&i.bytecodeValueStackDeltas[idx], 0)
	}
	atomic.StoreUint64(&i.bytecodeValueStackMaxDepth, 0)
	atomic.StoreUint64(&i.bytecodeValueStackMaxCapacity, 0)
	atomic.StoreUint64(&i.bytecodeValueStackCapacityGrowths, 0)
	atomic.StoreUint64(&i.bytecodeCallFrameMaxDepth, 0)
	atomic.StoreUint64(&i.bytecodeLoadNameLookups, 0)
	i.bytecodeLoadNameCountsMu.Lock()
	clear(i.bytecodeLoadNameCounts)
	i.bytecodeLoadNameCountsMu.Unlock()
	i.bytecodeStackPeakSitesMu.Lock()
	clear(i.bytecodeStackPeakSites)
	i.bytecodeStackPeakSitesMu.Unlock()
	i.bytecodeStackDeltaSitesMu.Lock()
	clear(i.bytecodeStackDeltaSites)
	i.bytecodeStackDeltaSitesMu.Unlock()
	i.bytecodeCallOperandBalancesMu.Lock()
	clear(i.bytecodeCallOperandBalances)
	i.bytecodeCallOperandBalancesMu.Unlock()
	i.bytecodeLoopBackedgeBalancesMu.Lock()
	clear(i.bytecodeLoopBackedgeBalances)
	i.bytecodeLoopBackedgeBalancesMu.Unlock()
	i.bytecodeInlineFrameBalancesMu.Lock()
	clear(i.bytecodeInlineFrameBalances)
	i.bytecodeInlineFrameBalancesMu.Unlock()
	atomic.StoreUint64(&i.bytecodeLoadNameHotHits, 0)
	atomic.StoreUint64(&i.bytecodeLoadNameScopeCacheHits, 0)
	atomic.StoreUint64(&i.bytecodeLoadNameGlobalCacheHits, 0)
	atomic.StoreUint64(&i.bytecodeLoadNameDirectCurrent, 0)
	atomic.StoreUint64(&i.bytecodeLoadNameDirectOuter, 0)
	atomic.StoreUint64(&i.bytecodeLoadNameScopeStores, 0)
	atomic.StoreUint64(&i.bytecodeLoadNameGlobalStores, 0)
	atomic.StoreUint64(&i.bytecodeCallNameLookups, 0)
	atomic.StoreUint64(&i.bytecodeCallNameDottedFallbacks, 0)
	atomic.StoreUint64(&i.bytecodeCallNameExactNativeHits, 0)
	atomic.StoreUint64(&i.bytecodeCallNameInlineDirectSlotHits, 0)
	atomic.StoreUint64(&i.bytecodeCallNameInlineDirectStackHits, 0)
	atomic.StoreUint64(&i.bytecodeCallNameInlineResolvedHits, 0)
	atomic.StoreUint64(&i.bytecodeCallNameInlineGenericHits, 0)
	atomic.StoreUint64(&i.bytecodeCallNameResolvedFunctionHits, 0)
	atomic.StoreUint64(&i.bytecodeCallNameGenericFallbacks, 0)
	atomic.StoreUint64(&i.bytecodeInlineCallHits, 0)
	atomic.StoreUint64(&i.bytecodeInlineCallMisses, 0)
	atomic.StoreUint64(&i.bytecodeDirectFunctionStackHits, 0)
	atomic.StoreUint64(&i.bytecodeInlineResolvedMissNoBytecode, 0)
	atomic.StoreUint64(&i.bytecodeInlineResolvedMissArity, 0)
	atomic.StoreUint64(&i.bytecodeInlineResolvedMissTypeArgs, 0)
	atomic.StoreUint64(&i.bytecodeInlineResolvedMissGenericLambda, 0)
	atomic.StoreUint64(&i.bytecodeMemberMethodCacheHits, 0)
	atomic.StoreUint64(&i.bytecodeMemberMethodCacheMisses, 0)
	atomic.StoreUint64(&i.bytecodeCallMemberResolvedExactNative, 0)
	atomic.StoreUint64(&i.bytecodeCallMemberResolvedInline, 0)
	atomic.StoreUint64(&i.bytecodeCallMemberResolvedGeneric, 0)
	atomic.StoreUint64(&i.bytecodeCallMemberResolvedFallback, 0)
	atomic.StoreUint64(&i.bytecodeGenericUnionCallCacheHits, 0)
	atomic.StoreUint64(&i.bytecodeGenericUnionCallCacheMisses, 0)
	atomic.StoreUint64(&i.bytecodeCallMemberStaticCacheHits, 0)
	atomic.StoreUint64(&i.bytecodeCallMemberStaticCacheMisses, 0)
	atomic.StoreUint64(&i.bytecodeCallMemberStaticExactNative, 0)
	atomic.StoreUint64(&i.bytecodeCallMemberStaticInline, 0)
	atomic.StoreUint64(&i.bytecodeCallMemberStaticGeneric, 0)
	atomic.StoreUint64(&i.bytecodeExprCacheHits, 0)
	atomic.StoreUint64(&i.bytecodeExprCacheMisses, 0)
	atomic.StoreUint64(&i.bytecodeArrayIndexSlotLookups, 0)
	atomic.StoreUint64(&i.bytecodeArrayIndexSlotTrackedHits, 0)
	atomic.StoreUint64(&i.bytecodeArrayIndexSlotMonoUnsignedHits, 0)
	atomic.StoreUint64(&i.bytecodeArrayIndexSlotDirectHits, 0)
	atomic.StoreUint64(&i.bytecodeArrayIndexSlotFallbacks, 0)
	atomic.StoreUint64(&i.bytecodeArrayIndexSlotFastDisabledMiss, 0)
	atomic.StoreUint64(&i.bytecodeArrayIndexSlotReceiverMiss, 0)
	atomic.StoreUint64(&i.bytecodeArrayIndexSlotIndexMiss, 0)
	atomic.StoreUint64(&i.bytecodeArrayIndexSlotHandleMiss, 0)
	atomic.StoreUint64(&i.bytecodeArrayIndexSlotDirectMiss, 0)
	atomic.StoreUint64(&i.bytecodeArrayMemberSlotLookups, 0)
	atomic.StoreUint64(&i.bytecodeArrayMemberSlotLenLookups, 0)
	atomic.StoreUint64(&i.bytecodeArrayMemberSlotReadLookups, 0)
	atomic.StoreUint64(&i.bytecodeArrayMemberSlotWriteLookups, 0)
	atomic.StoreUint64(&i.bytecodeArrayMemberSlotPushLookups, 0)
	atomic.StoreUint64(&i.bytecodeArrayMemberSlotCacheHits, 0)
	atomic.StoreUint64(&i.bytecodeArrayMemberSlotFastHits, 0)
	atomic.StoreUint64(&i.bytecodeArrayMemberSlotFallbacks, 0)
	atomic.StoreUint64(&i.bytecodeArrayMemberSlotReceiverMiss, 0)
	atomic.StoreUint64(&i.bytecodeArrayMemberSlotCacheMiss, 0)
	atomic.StoreUint64(&i.bytecodeArrayMemberSlotFastPathMiss, 0)
}
