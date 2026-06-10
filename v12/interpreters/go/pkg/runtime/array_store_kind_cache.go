package runtime

var arrayHandleKindCache []monoArrayKind
var arrayHandleKindCacheValid []bool

const arrayHandleKindCacheMaxHandle int64 = 1 << 20

func arrayHandleKindCacheIndex(handle int64) (int, bool) {
	if handle <= 0 || handle > arrayHandleKindCacheMaxHandle {
		return 0, false
	}
	idx := int(handle)
	if int64(idx) != handle {
		return 0, false
	}
	return idx, true
}

func recordArrayHandleKind(handle int64, kind monoArrayKind) {
	if handle == 0 {
		return
	}
	clearMonoArrayPrimitiveReadHot(handle, kind)
	clearMonoArrayElementTypeNameHot(handle, kind)
	if arrayHandleKinds == nil {
		arrayHandleKinds = make(map[int64]monoArrayKind)
	}
	arrayHandleKinds[handle] = kind
	idx, ok := arrayHandleKindCacheIndex(handle)
	if !ok {
		return
	}
	if idx >= len(arrayHandleKindCache) {
		growArrayHandleKindCache(idx + 1)
	}
	arrayHandleKindCache[idx] = kind
	arrayHandleKindCacheValid[idx] = true
}

// removeArrayHandleKind clears every kind-specific metadata cache for a
// released handle. Callers hold arrayStoreMu exclusively.
func removeArrayHandleKind(handle int64) {
	if handle == 0 {
		return
	}
	delete(arrayHandleKinds, handle)
	clearMonoArrayPrimitiveReadHot(handle, monoArrayKindDynamic)
	clearMonoArrayElementTypeNameHot(handle, monoArrayKindDynamic)
	if idx, ok := arrayHandleKindCacheIndex(handle); ok && idx < len(arrayHandleKindCacheValid) {
		arrayHandleKindCache[idx] = monoArrayKindDynamic
		arrayHandleKindCacheValid[idx] = false
	}
}

func growArrayHandleKindCache(newLen int) {
	if newLen <= len(arrayHandleKindCache) {
		return
	}
	if newLen <= cap(arrayHandleKindCache) && newLen <= cap(arrayHandleKindCacheValid) {
		arrayHandleKindCache = arrayHandleKindCache[:newLen]
		arrayHandleKindCacheValid = arrayHandleKindCacheValid[:newLen]
		return
	}
	newCap := grownCapacity(cap(arrayHandleKindCache), newLen)
	if otherCap := grownCapacity(cap(arrayHandleKindCacheValid), newLen); otherCap > newCap {
		newCap = otherCap
	}
	kindCache := make([]monoArrayKind, newLen, newCap)
	validCache := make([]bool, newLen, newCap)
	copy(kindCache, arrayHandleKindCache)
	copy(validCache, arrayHandleKindCacheValid)
	arrayHandleKindCache = kindCache
	arrayHandleKindCacheValid = validCache
}

func cachedArrayHandleKind(handle int64) (monoArrayKind, bool) {
	idx, ok := arrayHandleKindCacheIndex(handle)
	if !ok || idx >= len(arrayHandleKindCacheValid) || !arrayHandleKindCacheValid[idx] {
		return monoArrayKindDynamic, false
	}
	return arrayHandleKindCache[idx], true
}
