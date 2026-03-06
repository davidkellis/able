package interpreter

import "sync/atomic"

// BytecodeStatsSnapshot captures optional bytecode runtime counters.
// OpCounts is indexed by bytecode opcode numeric value.
type BytecodeStatsSnapshot struct {
	Enabled               bool
	OpCounts              []uint64
	LoadNameLookups       uint64
	CallNameLookups       uint64
	CallNameDotFallback   uint64
	InlineCallHits        uint64
	InlineCallMisses      uint64
	MemberMethodCacheHits uint64
	MemberMethodCacheMiss uint64
}

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

func (i *Interpreter) recordBytecodeLoadNameLookup() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeLoadNameLookups, 1)
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

// BytecodeStats returns a snapshot of bytecode counters.
func (i *Interpreter) BytecodeStats() BytecodeStatsSnapshot {
	snapshot := BytecodeStatsSnapshot{}
	if i == nil {
		return snapshot
	}
	snapshot.Enabled = i.bytecodeStatsEnabled
	snapshot.OpCounts = make([]uint64, bytecodeOpCount)
	for idx := 0; idx < bytecodeOpCount; idx++ {
		snapshot.OpCounts[idx] = atomic.LoadUint64(&i.bytecodeOpCounts[idx])
	}
	snapshot.LoadNameLookups = atomic.LoadUint64(&i.bytecodeLoadNameLookups)
	snapshot.CallNameLookups = atomic.LoadUint64(&i.bytecodeCallNameLookups)
	snapshot.CallNameDotFallback = atomic.LoadUint64(&i.bytecodeCallNameDottedFallbacks)
	snapshot.InlineCallHits = atomic.LoadUint64(&i.bytecodeInlineCallHits)
	snapshot.InlineCallMisses = atomic.LoadUint64(&i.bytecodeInlineCallMisses)
	snapshot.MemberMethodCacheHits = atomic.LoadUint64(&i.bytecodeMemberMethodCacheHits)
	snapshot.MemberMethodCacheMiss = atomic.LoadUint64(&i.bytecodeMemberMethodCacheMisses)
	return snapshot
}

// ResetBytecodeStats clears bytecode counters.
func (i *Interpreter) ResetBytecodeStats() {
	if i == nil {
		return
	}
	for idx := 0; idx < bytecodeOpCount; idx++ {
		atomic.StoreUint64(&i.bytecodeOpCounts[idx], 0)
	}
	atomic.StoreUint64(&i.bytecodeLoadNameLookups, 0)
	atomic.StoreUint64(&i.bytecodeCallNameLookups, 0)
	atomic.StoreUint64(&i.bytecodeCallNameDottedFallbacks, 0)
	atomic.StoreUint64(&i.bytecodeInlineCallHits, 0)
	atomic.StoreUint64(&i.bytecodeInlineCallMisses, 0)
	atomic.StoreUint64(&i.bytecodeMemberMethodCacheHits, 0)
	atomic.StoreUint64(&i.bytecodeMemberMethodCacheMisses, 0)
}
