package runtime

import "testing"

func TestArrayStoreAppendCharIfMonoAppendsToMonoChar(t *testing.T) {
	handle := ArrayStoreMonoNewWithCapacityChar(1)

	ok, err := ArrayStoreAppendCharIfMono(handle, 'a')
	if err != nil || !ok {
		t.Fatalf("ArrayStoreAppendCharIfMono = (%v, %v), want true/nil", ok, err)
	}
	raw, ok, err := ArrayStoreMonoReadCharIfAvailable(handle, 0)
	if err != nil || !ok || raw != 'a' {
		t.Fatalf("ArrayStoreMonoReadCharIfAvailable = (%q, %v, %v), want 'a'/true/nil", raw, ok, err)
	}
	if idx, cacheOK := arrayHandleKindCacheIndex(handle); !cacheOK || idx >= len(monoArrayCharAppendCache) || monoArrayCharAppendCache[idx] == nil {
		t.Fatalf("append direct cache missing for handle %d", handle)
	}
}

func TestArrayStoreAppendCharIfMonoUsesDirectCacheForMultipleHandles(t *testing.T) {
	first := ArrayStoreMonoNewWithCapacityChar(1)
	second := ArrayStoreMonoNewWithCapacityChar(1)
	if ok, err := ArrayStoreAppendCharIfMono(first, 'a'); err != nil || !ok {
		t.Fatalf("ArrayStoreAppendCharIfMono first seed = (%v, %v), want true/nil", ok, err)
	}
	if ok, err := ArrayStoreAppendCharIfMono(second, 'b'); err != nil || !ok {
		t.Fatalf("ArrayStoreAppendCharIfMono second seed = (%v, %v), want true/nil", ok, err)
	}
	monoArrayCharAppendHotHandle, monoArrayCharAppendHot = 0, nil
	if ok, err := ArrayStoreAppendCharIfMono(first, 'c'); err != nil || !ok {
		t.Fatalf("ArrayStoreAppendCharIfMono first cached = (%v, %v), want true/nil", ok, err)
	}
	if monoArrayCharAppendHotHandle != first || monoArrayCharAppendHot == nil {
		t.Fatalf("append hot cache after direct hit = handle %d state %p, want handle %d non-nil", monoArrayCharAppendHotHandle, monoArrayCharAppendHot, first)
	}
	if size, err := ArrayStoreSize(first); err != nil || size != 2 {
		t.Fatalf("ArrayStoreSize first after direct hit = (%d, %v), want 2/nil", size, err)
	}
}

func TestArrayStoreAppendCharIfMonoReturnsFalseForDynamic(t *testing.T) {
	handle := ArrayStoreNew()
	if err := ArrayStoreWrite(handle, 0, CharValue{Val: 'a'}); err != nil {
		t.Fatalf("ArrayStoreWrite: %v", err)
	}

	ok, err := ArrayStoreAppendCharIfMono(handle, 'z')
	if err != nil || ok {
		t.Fatalf("ArrayStoreAppendCharIfMono dynamic = (%v, %v), want false/nil", ok, err)
	}
	if size, err := ArrayStoreSize(handle); err != nil || size != 1 {
		t.Fatalf("ArrayStoreSize after skipped mono append = (%d, %v), want 1/nil", size, err)
	}
}

func TestArrayStoreAppendCharIfMonoDoesNotUseStaleStateAfterDeopt(t *testing.T) {
	handle := ArrayStoreMonoNewWithCapacityChar(1)
	ok, err := ArrayStoreAppendCharIfMono(handle, 'a')
	if err != nil || !ok {
		t.Fatalf("ArrayStoreAppendCharIfMono seed = (%v, %v), want true/nil", ok, err)
	}
	if monoArrayCharAppendHotHandle != handle || monoArrayCharAppendHot == nil {
		t.Fatalf("append hot cache after seed = handle %d state %p, want handle %d non-nil", monoArrayCharAppendHotHandle, monoArrayCharAppendHot, handle)
	}
	if _, err := ArrayStoreState(handle); err != nil {
		t.Fatalf("ArrayStoreState deopt: %v", err)
	}
	if monoArrayCharAppendHotHandle == handle || monoArrayCharAppendHot != nil {
		t.Fatalf("append hot cache after deopt = handle %d state %p, want cleared", monoArrayCharAppendHotHandle, monoArrayCharAppendHot)
	}
	if idx, cacheOK := arrayHandleKindCacheIndex(handle); cacheOK && idx < len(monoArrayCharAppendCache) && monoArrayCharAppendCache[idx] != nil {
		t.Fatalf("append direct cache after deopt = %p, want nil", monoArrayCharAppendCache[idx])
	}

	ok, err = ArrayStoreAppendCharIfMono(handle, 'z')
	if err != nil || ok {
		t.Fatalf("ArrayStoreAppendCharIfMono after deopt = (%v, %v), want false/nil", ok, err)
	}
	if size, err := ArrayStoreSize(handle); err != nil || size != 1 {
		t.Fatalf("ArrayStoreSize after skipped stale append = (%d, %v), want 1/nil", size, err)
	}
}
