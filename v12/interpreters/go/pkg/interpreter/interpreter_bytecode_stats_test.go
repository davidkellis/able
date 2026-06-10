package interpreter

import "testing"

func TestBytecodeStatsLoadNameCountsByName(t *testing.T) {
	interp := NewBytecode()
	interp.bytecodeStatsEnabled = true

	interp.recordBytecodeLoadNameLookupForName("beta")
	interp.recordBytecodeLoadNameLookupForName("alpha")
	interp.recordBytecodeLoadNameLookupForName("alpha")

	stats := interp.BytecodeStats()
	if stats.LoadNameLookups != 3 {
		t.Fatalf("LoadNameLookups = %d, want 3", stats.LoadNameLookups)
	}
	if stats.LoadNameLookupsByName["alpha"] != 2 || stats.LoadNameLookupsByName["beta"] != 1 {
		t.Fatalf("unexpected per-name lookup counts: %#v", stats.LoadNameLookupsByName)
	}
	if len(stats.TopLoadNames) != 2 {
		t.Fatalf("TopLoadNames len = %d, want 2", len(stats.TopLoadNames))
	}
	if stats.TopLoadNames[0].Name != "alpha" || stats.TopLoadNames[0].Count != 2 {
		t.Fatalf("unexpected first top load-name entry: %#v", stats.TopLoadNames[0])
	}
	if stats.TopLoadNames[1].Name != "beta" || stats.TopLoadNames[1].Count != 1 {
		t.Fatalf("unexpected second top load-name entry: %#v", stats.TopLoadNames[1])
	}

	interp.ResetBytecodeStats()
	stats = interp.BytecodeStats()
	if stats.LoadNameLookups != 0 || len(stats.LoadNameLookupsByName) != 0 || len(stats.TopLoadNames) != 0 {
		t.Fatalf("expected reset load-name counts, got %#v", stats)
	}
}

func TestBytecodeStatsArrayMemberSlotCountsByKind(t *testing.T) {
	interp := NewBytecode()
	interp.bytecodeStatsEnabled = true

	interp.recordBytecodeArrayMemberSlotLookup(bytecodeMemberMethodFastPathArrayLen)
	interp.recordBytecodeArrayMemberSlotLookup(bytecodeMemberMethodFastPathArrayReadSlot)
	interp.recordBytecodeArrayMemberSlotLookup(bytecodeMemberMethodFastPathArrayWriteSlot)
	interp.recordBytecodeArrayMemberSlotLookup(bytecodeMemberMethodFastPathArrayPush)
	interp.recordBytecodeArrayMemberSlotCacheHit()
	interp.recordBytecodeArrayMemberSlotFastHit()
	interp.recordBytecodeArrayMemberSlotFallback(bytecodeArrayMemberSlotFallbackFastPathMiss)

	stats := interp.BytecodeStats()
	if stats.ArrayMemberSlotLookups != 4 ||
		stats.ArrayMemberSlotLenLookups != 1 ||
		stats.ArrayMemberSlotReadLookups != 1 ||
		stats.ArrayMemberSlotWriteLookups != 1 ||
		stats.ArrayMemberSlotPushLookups != 1 {
		t.Fatalf("unexpected Array member-kind counts: %#v", stats)
	}
	if stats.ArrayMemberSlotCacheHits != 1 || stats.ArrayMemberSlotFastHits != 1 ||
		stats.ArrayMemberSlotFallbacks != 1 || stats.ArrayMemberSlotFastPathMiss != 1 {
		t.Fatalf("unexpected Array member dispatch counts: %#v", stats)
	}

	interp.ResetBytecodeStats()
	stats = interp.BytecodeStats()
	if stats.ArrayMemberSlotLookups != 0 || stats.ArrayMemberSlotLenLookups != 0 ||
		stats.ArrayMemberSlotReadLookups != 0 || stats.ArrayMemberSlotWriteLookups != 0 ||
		stats.ArrayMemberSlotPushLookups != 0 || stats.ArrayMemberSlotCacheHits != 0 ||
		stats.ArrayMemberSlotFastHits != 0 || stats.ArrayMemberSlotFallbacks != 0 ||
		stats.ArrayMemberSlotFastPathMiss != 0 {
		t.Fatalf("expected reset Array member counts, got %#v", stats)
	}
}
