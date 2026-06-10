//go:build able_bytecode_box_reuse

package interpreter

import (
	"sync"

	"able/interpreter-go/pkg/runtime"
)

const bytecodeDynamicIntBoxReuseEnabled = true

var bytecodeDynamicIntBoxCacheReuseState struct {
	sync.Mutex
	byKind map[runtime.IntegerType]bytecodeDynamicIntBoxCacheReuse
}

func bytecodeRecordDynamicIntBoxCacheEvent(kind runtime.IntegerType, event bytecodeDynamicIntBoxCacheEvent) {
	bytecodeDynamicIntBoxCacheReuseState.Lock()
	defer bytecodeDynamicIntBoxCacheReuseState.Unlock()
	if bytecodeDynamicIntBoxCacheReuseState.byKind == nil {
		bytecodeDynamicIntBoxCacheReuseState.byKind = make(map[runtime.IntegerType]bytecodeDynamicIntBoxCacheReuse)
	}
	stats := bytecodeDynamicIntBoxCacheReuseState.byKind[kind]
	switch event {
	case bytecodeDynamicIntBoxCacheLookup:
		stats.Lookups++
	case bytecodeDynamicIntBoxCacheHit:
		stats.Hits++
	case bytecodeDynamicIntBoxCacheInsert:
		stats.Inserts++
	case bytecodeDynamicIntBoxCacheCapacityMiss:
		stats.CapacityMisses++
	case bytecodeDynamicIntBoxCacheI64Bypass:
		stats.I64Bypasses++
	}
	bytecodeDynamicIntBoxCacheReuseState.byKind[kind] = stats
}

func bytecodeResetDynamicIntBoxCacheReuseForTest() {
	bytecodeDynamicIntBoxCacheReuseState.Lock()
	defer bytecodeDynamicIntBoxCacheReuseState.Unlock()
	bytecodeDynamicIntBoxCacheReuseState.byKind = nil
}

func bytecodeDynamicIntBoxCacheReuseForTest() map[string]bytecodeDynamicIntBoxCacheReuse {
	bytecodeDynamicIntBoxCacheReuseState.Lock()
	defer bytecodeDynamicIntBoxCacheReuseState.Unlock()
	if len(bytecodeDynamicIntBoxCacheReuseState.byKind) == 0 {
		return nil
	}
	snapshot := make(map[string]bytecodeDynamicIntBoxCacheReuse, len(bytecodeDynamicIntBoxCacheReuseState.byKind))
	for kind, stats := range bytecodeDynamicIntBoxCacheReuseState.byKind {
		snapshot[string(kind)] = stats
	}
	return snapshot
}
