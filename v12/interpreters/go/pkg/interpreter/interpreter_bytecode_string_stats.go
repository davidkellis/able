package interpreter

import "sync/atomic"

// BytecodeStringStatsSnapshot reports narrowly scoped String/UTF-8 semantic
// work. It is populated only when ABLE_BYTECODE_STRING_STATS is set.
type BytecodeStringStatsSnapshot struct {
	Enabled                  bool   `json:"enabled"`
	FromBuiltinCalls         uint64 `json:"from_builtin_calls"`
	FromBuiltinBytes         uint64 `json:"from_builtin_bytes"`
	ToBuiltinCalls           uint64 `json:"to_builtin_calls"`
	ToBuiltinBytes           uint64 `json:"to_builtin_bytes"`
	ToBuiltinMonoCalls       uint64 `json:"to_builtin_mono_calls"`
	ToBuiltinFallbackCalls   uint64 `json:"to_builtin_fallback_calls"`
	CanonicalValidations     uint64 `json:"canonical_validations"`
	CanonicalValidationBytes uint64 `json:"canonical_validation_bytes"`
	RawValidations           uint64 `json:"raw_validations"`
	RawValidationBytes       uint64 `json:"raw_validation_bytes"`
	RuneDecodes              uint64 `json:"rune_decodes"`
	RuneDecodeBytes          uint64 `json:"rune_decode_bytes"`
	InvalidUTF8              uint64 `json:"invalid_utf8"`
}

type bytecodeStringStats struct {
	fromBuiltinCalls         atomic.Uint64
	fromBuiltinBytes         atomic.Uint64
	toBuiltinCalls           atomic.Uint64
	toBuiltinBytes           atomic.Uint64
	toBuiltinMonoCalls       atomic.Uint64
	toBuiltinFallbackCalls   atomic.Uint64
	canonicalValidations     atomic.Uint64
	canonicalValidationBytes atomic.Uint64
	rawValidations           atomic.Uint64
	rawValidationBytes       atomic.Uint64
	runeDecodes              atomic.Uint64
	runeDecodeBytes          atomic.Uint64
	invalidUTF8              atomic.Uint64
}

func (i *Interpreter) recordBytecodeStringFromBuiltin(bytes int) {
	stats := i.bytecodeStringStats
	if stats == nil {
		return
	}
	stats.fromBuiltinCalls.Add(1)
	if bytes > 0 {
		stats.fromBuiltinBytes.Add(uint64(bytes))
	}
}

func (i *Interpreter) recordBytecodeStringToBuiltin(bytes int, mono bool, valid bool) {
	stats := i.bytecodeStringStats
	if stats == nil {
		return
	}
	stats.toBuiltinCalls.Add(1)
	if bytes > 0 {
		stats.toBuiltinBytes.Add(uint64(bytes))
	}
	if mono {
		stats.toBuiltinMonoCalls.Add(1)
	} else {
		stats.toBuiltinFallbackCalls.Add(1)
	}
	if !valid {
		stats.invalidUTF8.Add(1)
	}
}

func (i *Interpreter) recordBytecodeStringCanonicalValidation(bytes int, valid bool) {
	stats := i.bytecodeStringStats
	if stats == nil {
		return
	}
	stats.canonicalValidations.Add(1)
	if bytes > 0 {
		stats.canonicalValidationBytes.Add(uint64(bytes))
	}
	if !valid {
		stats.invalidUTF8.Add(1)
	}
}

func (i *Interpreter) recordBytecodeStringRawValidation(bytes int, valid bool) {
	stats := i.bytecodeStringStats
	if stats == nil {
		return
	}
	stats.rawValidations.Add(1)
	if bytes > 0 {
		stats.rawValidationBytes.Add(uint64(bytes))
	}
	if !valid {
		stats.invalidUTF8.Add(1)
	}
}

func (i *Interpreter) recordBytecodeStringRuneDecode(bytes int, valid bool) {
	stats := i.bytecodeStringStats
	if stats == nil {
		return
	}
	stats.runeDecodes.Add(1)
	if bytes > 0 {
		stats.runeDecodeBytes.Add(uint64(bytes))
	}
	if !valid {
		stats.invalidUTF8.Add(1)
	}
}

func (i *Interpreter) ResetBytecodeStringStats() {
	if i == nil || i.bytecodeStringStats == nil {
		return
	}
	i.bytecodeStringStats = &bytecodeStringStats{}
}

func (i *Interpreter) BytecodeStringStats() BytecodeStringStatsSnapshot {
	if i == nil || i.bytecodeStringStats == nil {
		return BytecodeStringStatsSnapshot{}
	}
	stats := i.bytecodeStringStats
	return BytecodeStringStatsSnapshot{
		Enabled:                  true,
		FromBuiltinCalls:         stats.fromBuiltinCalls.Load(),
		FromBuiltinBytes:         stats.fromBuiltinBytes.Load(),
		ToBuiltinCalls:           stats.toBuiltinCalls.Load(),
		ToBuiltinBytes:           stats.toBuiltinBytes.Load(),
		ToBuiltinMonoCalls:       stats.toBuiltinMonoCalls.Load(),
		ToBuiltinFallbackCalls:   stats.toBuiltinFallbackCalls.Load(),
		CanonicalValidations:     stats.canonicalValidations.Load(),
		CanonicalValidationBytes: stats.canonicalValidationBytes.Load(),
		RawValidations:           stats.rawValidations.Load(),
		RawValidationBytes:       stats.rawValidationBytes.Load(),
		RuneDecodes:              stats.runeDecodes.Load(),
		RuneDecodeBytes:          stats.runeDecodeBytes.Load(),
		InvalidUTF8:              stats.invalidUTF8.Load(),
	}
}
