package interpreter

import "able/interpreter-go/pkg/runtime"

type equalityDispatchCacheKey struct {
	typeName string
}

type equalityDispatchCacheKind uint8

const (
	equalityDispatchCacheUnknown equalityDispatchCacheKind = iota
	equalityDispatchCacheNoMethod
	equalityDispatchCacheMethod
	equalityDispatchCacheError
)

type equalityDispatchCacheEntry struct {
	kind      equalityDispatchCacheKind
	dispatch  operatorDispatch
	method    runtime.Value
	primitive bool
	err       error
}

const equalityDispatchCacheMaxEntries = 512
const equalityDispatchCacheInitialEntries = equalityDispatchCacheMaxEntries / 2

func (i *Interpreter) lookupEqualityDispatchCache(typeName string) (equalityDispatchCacheEntry, bool) {
	if i == nil || typeName == "" || i.interfaceMethodResolver != nil {
		return equalityDispatchCacheEntry{}, false
	}
	key := equalityDispatchCacheKey{typeName: typeName}
	if i.envSingleThread {
		entry, ok := i.equalityDispatchCache[key]
		if !ok || entry.kind == equalityDispatchCacheUnknown {
			return equalityDispatchCacheEntry{}, false
		}
		return entry, true
	}
	i.methodCacheMu.RLock()
	entry, ok := i.equalityDispatchCache[key]
	i.methodCacheMu.RUnlock()
	if !ok || entry.kind == equalityDispatchCacheUnknown {
		return equalityDispatchCacheEntry{}, false
	}
	return entry, true
}

func (i *Interpreter) storeEqualityDispatchCache(typeName string, entry equalityDispatchCacheEntry) {
	if i == nil || typeName == "" || entry.kind == equalityDispatchCacheUnknown || i.interfaceMethodResolver != nil {
		return
	}
	key := equalityDispatchCacheKey{typeName: typeName}
	if i.envSingleThread {
		if i.equalityDispatchCache == nil {
			i.equalityDispatchCache = make(map[equalityDispatchCacheKey]equalityDispatchCacheEntry, equalityDispatchCacheInitialEntries)
		}
		if len(i.equalityDispatchCache) >= equalityDispatchCacheMaxEntries {
			clear(i.equalityDispatchCache)
		}
		i.equalityDispatchCache[key] = entry
		return
	}
	i.methodCacheMu.Lock()
	defer i.methodCacheMu.Unlock()
	if i.equalityDispatchCache == nil {
		i.equalityDispatchCache = make(map[equalityDispatchCacheKey]equalityDispatchCacheEntry, equalityDispatchCacheInitialEntries)
	}
	if len(i.equalityDispatchCache) >= equalityDispatchCacheMaxEntries {
		clear(i.equalityDispatchCache)
	}
	i.equalityDispatchCache[key] = entry
}

func (i *Interpreter) clearEqualityDispatchCache() {
	if i == nil {
		return
	}
	if i.envSingleThread {
		if len(i.equalityDispatchCache) > 0 {
			clear(i.equalityDispatchCache)
		}
		return
	}
	i.methodCacheMu.Lock()
	if len(i.equalityDispatchCache) > 0 {
		clear(i.equalityDispatchCache)
	}
	i.methodCacheMu.Unlock()
}
