package interpreter

import "testing"

func TestBytecodeStringStatsSnapshotAndReset(t *testing.T) {
	interp := &Interpreter{bytecodeStringStats: &bytecodeStringStats{}}
	interp.recordBytecodeStringFromBuiltin(12)
	interp.recordBytecodeStringToBuiltin(9, true, true)
	interp.recordBytecodeStringToBuiltin(4, false, false)
	interp.recordBytecodeStringCanonicalValidation(7, true)
	interp.recordBytecodeStringRawValidation(6, false)
	interp.recordBytecodeStringRuneDecode(3, true)

	got := interp.BytecodeStringStats()
	if !got.Enabled || got.FromBuiltinCalls != 1 || got.FromBuiltinBytes != 12 {
		t.Fatalf("from-builtin stats = %+v", got)
	}
	if got.ToBuiltinCalls != 2 || got.ToBuiltinBytes != 13 || got.ToBuiltinMonoCalls != 1 || got.ToBuiltinFallbackCalls != 1 {
		t.Fatalf("to-builtin stats = %+v", got)
	}
	if got.CanonicalValidations != 1 || got.CanonicalValidationBytes != 7 || got.RawValidations != 1 || got.RawValidationBytes != 6 {
		t.Fatalf("validation stats = %+v", got)
	}
	if got.RuneDecodes != 1 || got.RuneDecodeBytes != 3 || got.InvalidUTF8 != 2 {
		t.Fatalf("decode/error stats = %+v", got)
	}

	interp.ResetBytecodeStringStats()
	reset := interp.BytecodeStringStats()
	if !reset.Enabled || reset.FromBuiltinCalls != 0 || reset.InvalidUTF8 != 0 {
		t.Fatalf("reset stats = %+v", reset)
	}
}

func TestBytecodeStringStatsDisabledIsInert(t *testing.T) {
	interp := &Interpreter{}
	interp.recordBytecodeStringFromBuiltin(10)
	interp.recordBytecodeStringToBuiltin(10, true, false)
	interp.recordBytecodeStringCanonicalValidation(10, false)
	interp.recordBytecodeStringRawValidation(10, false)
	interp.recordBytecodeStringRuneDecode(1, false)
	if got := interp.BytecodeStringStats(); got.Enabled {
		t.Fatalf("disabled stats = %+v", got)
	}
}

func TestBytecodeStringStatsEnvironmentEnablesObserver(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STRING_STATS", "1")
	interp := NewBytecode()
	if got := interp.BytecodeStringStats(); !got.Enabled {
		t.Fatalf("environment-enabled stats = %+v", got)
	}
}
