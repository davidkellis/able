package interpreter

// bytecodeDynamicIntBoxCacheEntriesForTest snapshots the test-process-only
// dynamic integer-box cache. Retention probes run one program per process, so
// the counts identify that program's cache contribution without adding any
// instrumentation to the VM hot path.
func bytecodeDynamicIntBoxCacheEntriesForTest() map[string]int {
	bytecodeIntBoxDynamicMu.RLock()
	defer bytecodeIntBoxDynamicMu.RUnlock()
	return map[string]int{
		"i8":    len(bytecodeDynamicBoxedI8),
		"i16":   len(bytecodeDynamicBoxedI16),
		"i32":   len(bytecodeDynamicBoxedI32),
		"i64":   len(bytecodeDynamicBoxedI64),
		"i128":  len(bytecodeDynamicBoxedI128),
		"u8":    len(bytecodeDynamicBoxedU8),
		"u16":   len(bytecodeDynamicBoxedU16),
		"u32":   len(bytecodeDynamicBoxedU32),
		"u64":   len(bytecodeDynamicBoxedU64),
		"u128":  len(bytecodeDynamicBoxedU128),
		"isize": len(bytecodeDynamicBoxedIsize),
		"usize": len(bytecodeDynamicBoxedUsize),
	}
}
