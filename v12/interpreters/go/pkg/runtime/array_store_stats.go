package runtime

import "unsafe"

// ArrayStoreStorageStats is a point-in-time summary for one ArrayStore
// representation. BackingBytes includes only direct slice backing owned by the
// registry; it deliberately excludes values referenced through dynamic
// runtime.Value entries.
type ArrayStoreStorageStats struct {
	StateCount       int
	ValueCount       int
	DeclaredCapacity int
	BackingCapacity  int
	BackingBytes     int64
}

// ArrayStoreStats is a read-only, lock-consistent snapshot of the
// process-wide ArrayStore registry. It is intended for runtime diagnostics and
// deterministic lifecycle tests; it does not retain Array values or force GC.
type ArrayStoreStats struct {
	HandleCount     int
	RevisionCount   int
	TotalStateCount int

	Dynamic ArrayStoreStorageStats
	I32     ArrayStoreStorageStats
	I64     ArrayStoreStorageStats
	Bool    ArrayStoreStorageStats
	Char    ArrayStoreStorageStats
	U8      ArrayStoreStorageStats
	U32     ArrayStoreStorageStats
	U64     ArrayStoreStorageStats
	F64     ArrayStoreStorageStats
}

// ArrayStoreStatsSnapshot returns a point-in-time view of every live Array
// backing state. It does not initialize, promote, deoptimize, or otherwise
// mutate the registry.
func ArrayStoreStatsSnapshot() ArrayStoreStats {
	arrayStoreMu.RLock()
	defer arrayStoreMu.RUnlock()

	stats := ArrayStoreStats{
		HandleCount:   len(arrayHandleKinds),
		RevisionCount: len(arrayHandleRevisions),
		Dynamic:       arrayStoreDynamicStats(arrayStates),
		I32:           arrayStoreMonoStats(monoArrayI32States),
		I64:           arrayStoreMonoStats(monoArrayI64States),
		Bool:          arrayStoreMonoStats(monoArrayBoolStates),
		Char:          arrayStoreMonoStats(monoArrayCharStates),
		U8:            arrayStoreMonoStats(monoArrayU8States),
		U32:           arrayStoreMonoStats(monoArrayU32States),
		U64:           arrayStoreMonoStats(monoArrayU64States),
		F64:           arrayStoreMonoStats(monoArrayF64States),
	}
	stats.TotalStateCount = stats.Dynamic.StateCount +
		stats.I32.StateCount +
		stats.I64.StateCount +
		stats.Bool.StateCount +
		stats.Char.StateCount +
		stats.U8.StateCount +
		stats.U32.StateCount +
		stats.U64.StateCount +
		stats.F64.StateCount
	return stats
}

func arrayStoreDynamicStats(states map[int64]*ArrayState) ArrayStoreStorageStats {
	var stats ArrayStoreStorageStats
	var zero Value
	valueBytes := int64(unsafe.Sizeof(zero))
	for _, state := range states {
		if state == nil {
			continue
		}
		stats.StateCount++
		stats.ValueCount += len(state.Values)
		stats.DeclaredCapacity += state.Capacity
		stats.BackingCapacity += cap(state.Values)
		stats.BackingBytes += int64(cap(state.Values)) * valueBytes
		stats.BackingBytes += int64(cap(state.CachedI32Values)) * int64(unsafe.Sizeof(int32(0)))
		stats.BackingBytes += int64(cap(state.CachedI32ValuesValid)) * int64(unsafe.Sizeof(bool(false)))
	}
	return stats
}

func arrayStoreMonoStats[T any](states map[int64]*monoArrayState[T]) ArrayStoreStorageStats {
	var stats ArrayStoreStorageStats
	var zero T
	elementBytes := int64(unsafe.Sizeof(zero))
	for _, state := range states {
		if state == nil {
			continue
		}
		stats.StateCount++
		stats.ValueCount += len(state.Values)
		stats.DeclaredCapacity += state.Capacity
		stats.BackingCapacity += cap(state.Values)
		stats.BackingBytes += int64(cap(state.Values)) * elementBytes
	}
	return stats
}
