package interpreter

// bytecodeDynamicIntBoxCacheReuse is an opt-in diagnostic snapshot for the
// bounded dynamic integer-box cache. Small and extended-i32 static-cache hits
// are intentionally excluded because they do not exercise this cache policy.
type bytecodeDynamicIntBoxCacheReuse struct {
	Lookups        uint64 `json:"lookups"`
	Hits           uint64 `json:"hits"`
	Inserts        uint64 `json:"inserts"`
	CapacityMisses uint64 `json:"capacity_misses"`
	I64Bypasses    uint64 `json:"i64_bypasses"`
}

type bytecodeDynamicIntBoxCacheEvent uint8

const (
	bytecodeDynamicIntBoxCacheLookup bytecodeDynamicIntBoxCacheEvent = iota
	bytecodeDynamicIntBoxCacheHit
	bytecodeDynamicIntBoxCacheInsert
	bytecodeDynamicIntBoxCacheCapacityMiss
	bytecodeDynamicIntBoxCacheI64Bypass
)
