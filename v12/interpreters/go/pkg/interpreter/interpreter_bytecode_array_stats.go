package interpreter

import "sync/atomic"

func (i *Interpreter) recordBytecodeArrayMemberSlotLookup(kind bytecodeMemberMethodFastPathKind) {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeArrayMemberSlotLookups, 1)
	switch kind {
	case bytecodeMemberMethodFastPathArrayLen:
		atomic.AddUint64(&i.bytecodeArrayMemberSlotLenLookups, 1)
	case bytecodeMemberMethodFastPathArrayReadSlot:
		atomic.AddUint64(&i.bytecodeArrayMemberSlotReadLookups, 1)
	case bytecodeMemberMethodFastPathArrayWriteSlot:
		atomic.AddUint64(&i.bytecodeArrayMemberSlotWriteLookups, 1)
	case bytecodeMemberMethodFastPathArrayPush:
		atomic.AddUint64(&i.bytecodeArrayMemberSlotPushLookups, 1)
	}
}

func (i *Interpreter) recordBytecodeArrayMemberSlotCacheHit() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeArrayMemberSlotCacheHits, 1)
}

func (i *Interpreter) recordBytecodeArrayMemberSlotFastHit() {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeArrayMemberSlotFastHits, 1)
}

func (i *Interpreter) recordBytecodeArrayMemberSlotFallback(reason bytecodeArrayMemberSlotFallbackReason) {
	if i == nil || !i.bytecodeStatsEnabled {
		return
	}
	atomic.AddUint64(&i.bytecodeArrayMemberSlotFallbacks, 1)
	switch reason {
	case bytecodeArrayMemberSlotFallbackReceiverMiss:
		atomic.AddUint64(&i.bytecodeArrayMemberSlotReceiverMiss, 1)
	case bytecodeArrayMemberSlotFallbackCacheMiss:
		atomic.AddUint64(&i.bytecodeArrayMemberSlotCacheMiss, 1)
	case bytecodeArrayMemberSlotFallbackFastPathMiss:
		atomic.AddUint64(&i.bytecodeArrayMemberSlotFastPathMiss, 1)
	}
}
