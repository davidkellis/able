package interpreter

import "able/interpreter-go/pkg/runtime"

type methodScopeCallableCacheKey struct {
	env  *runtime.Environment
	name string
}

type methodScopeCallableCacheEntry struct {
	stateID      uint64
	shapeVersion uint64
	owner        *runtime.Environment
	ownerVersion uint64
	value        runtime.Value
	filter       functionScopeFilter
	kind         methodScopeCallableCacheKind
}

type methodScopeCallableCacheKind uint8

const (
	methodScopeCallableCacheUnknown methodScopeCallableCacheKind = iota
	methodScopeCallableCacheMiss
	methodScopeCallableCacheNonCallable
	methodScopeCallableCacheCallable
)

type methodScopeHasCacheKey struct {
	env  *runtime.Environment
	name string
}

type methodScopeHasCacheEntry struct {
	stateID      uint64
	shapeVersion uint64
	exists       bool
}

const methodScopeLookupCacheMaxEntries = 2048
const methodScopeLookupCacheInitialEntries = methodScopeLookupCacheMaxEntries / 2

func (i *Interpreter) lookupMethodScopeCallable(env *runtime.Environment, name string) (runtime.Value, functionScopeFilter, bool) {
	if i == nil || env == nil || name == "" {
		return nil, functionScopeFilter{}, false
	}
	if !i.envSingleThread {
		val, ok := env.Lookup(name)
		if !ok || !isCallableRuntimeValue(val) {
			return nil, functionScopeFilter{}, false
		}
		return val, functionScopeFilterFromValue(val), true
	}
	stateID := env.BindingShapeStateID()
	shapeVersion := env.BindingShapeRevision()
	key := methodScopeCallableCacheKey{env: env, name: name}
	singleThread := i.envSingleThread
	if entry, ok := i.cachedMethodScopeCallable(key, stateID, shapeVersion, singleThread); ok {
		if entry.kind == methodScopeCallableCacheCallable {
			return entry.value, entry.filter, true
		}
		return nil, functionScopeFilter{}, false
	}
	val, owner, ownerVersion, found := env.LookupWithOwnerAndRevisionHint(name, singleThread)
	if !found {
		i.storeMethodScopeCallableCache(key, methodScopeCallableCacheEntry{
			stateID:      stateID,
			shapeVersion: shapeVersion,
			kind:         methodScopeCallableCacheMiss,
		})
		return nil, functionScopeFilter{}, false
	}
	if !isCallableRuntimeValue(val) {
		i.storeMethodScopeCallableCache(key, methodScopeCallableCacheEntry{
			stateID:      stateID,
			shapeVersion: shapeVersion,
			owner:        owner,
			ownerVersion: ownerVersion,
			kind:         methodScopeCallableCacheNonCallable,
		})
		return nil, functionScopeFilter{}, false
	}
	filter := functionScopeFilterFromValue(val)
	i.storeMethodScopeCallableCache(key, methodScopeCallableCacheEntry{
		stateID:      stateID,
		shapeVersion: shapeVersion,
		owner:        owner,
		ownerVersion: ownerVersion,
		value:        val,
		filter:       filter,
		kind:         methodScopeCallableCacheCallable,
	})
	return val, filter, true
}

func (i *Interpreter) cachedMethodScopeCallable(key methodScopeCallableCacheKey, stateID uint64, shapeVersion uint64, singleThread bool) (methodScopeCallableCacheEntry, bool) {
	if i.envSingleThread {
		entry, ok := i.methodScopeCallableCache[key]
		if !ok || !entry.valid(stateID, shapeVersion, singleThread) {
			return methodScopeCallableCacheEntry{}, false
		}
		return entry, true
	}
	i.methodCacheMu.RLock()
	entry, ok := i.methodScopeCallableCache[key]
	i.methodCacheMu.RUnlock()
	if !ok || !entry.valid(stateID, shapeVersion, singleThread) {
		return methodScopeCallableCacheEntry{}, false
	}
	return entry, true
}

func (entry methodScopeCallableCacheEntry) valid(stateID uint64, shapeVersion uint64, singleThread bool) bool {
	if entry.kind == methodScopeCallableCacheUnknown {
		return false
	}
	if entry.stateID != stateID || entry.shapeVersion != shapeVersion {
		return false
	}
	if entry.kind == methodScopeCallableCacheMiss {
		return true
	}
	if entry.owner == nil {
		return false
	}
	return entry.ownerVersion == entry.owner.RevisionWithHint(singleThread)
}

func (i *Interpreter) storeMethodScopeCallableCache(key methodScopeCallableCacheKey, entry methodScopeCallableCacheEntry) {
	if i == nil || key.env == nil || key.name == "" || entry.kind == methodScopeCallableCacheUnknown {
		return
	}
	if i.envSingleThread {
		if i.methodScopeCallableCache == nil {
			i.methodScopeCallableCache = make(map[methodScopeCallableCacheKey]methodScopeCallableCacheEntry, methodScopeLookupCacheInitialEntries)
		}
		if len(i.methodScopeCallableCache) >= methodScopeLookupCacheMaxEntries {
			clear(i.methodScopeCallableCache)
		}
		i.methodScopeCallableCache[key] = entry
		return
	}
	i.methodCacheMu.Lock()
	defer i.methodCacheMu.Unlock()
	if i.methodScopeCallableCache == nil {
		i.methodScopeCallableCache = make(map[methodScopeCallableCacheKey]methodScopeCallableCacheEntry, methodScopeLookupCacheInitialEntries)
	}
	if len(i.methodScopeCallableCache) >= methodScopeLookupCacheMaxEntries {
		clear(i.methodScopeCallableCache)
	}
	i.methodScopeCallableCache[key] = entry
}

func (i *Interpreter) methodTypeNameInScope(env *runtime.Environment, name string) bool {
	if i == nil || env == nil || name == "" {
		return false
	}
	if !i.envSingleThread {
		return env.Has(name)
	}
	stateID := env.BindingShapeStateID()
	shapeVersion := env.BindingShapeRevision()
	key := methodScopeHasCacheKey{env: env, name: name}
	if exists, ok := i.cachedMethodScopeHas(key, stateID, shapeVersion); ok {
		return exists
	}
	exists := env.Has(name)
	i.storeMethodScopeHasCache(key, methodScopeHasCacheEntry{
		stateID:      stateID,
		shapeVersion: shapeVersion,
		exists:       exists,
	})
	return exists
}

func (i *Interpreter) cachedMethodScopeHas(key methodScopeHasCacheKey, stateID uint64, shapeVersion uint64) (bool, bool) {
	if i.envSingleThread {
		entry, ok := i.methodScopeHasCache[key]
		if !ok || entry.stateID != stateID || entry.shapeVersion != shapeVersion {
			return false, false
		}
		return entry.exists, true
	}
	i.methodCacheMu.RLock()
	entry, ok := i.methodScopeHasCache[key]
	i.methodCacheMu.RUnlock()
	if !ok || entry.stateID != stateID || entry.shapeVersion != shapeVersion {
		return false, false
	}
	return entry.exists, true
}

func (i *Interpreter) storeMethodScopeHasCache(key methodScopeHasCacheKey, entry methodScopeHasCacheEntry) {
	if i == nil || key.env == nil || key.name == "" {
		return
	}
	if i.envSingleThread {
		if i.methodScopeHasCache == nil {
			i.methodScopeHasCache = make(map[methodScopeHasCacheKey]methodScopeHasCacheEntry, methodScopeLookupCacheInitialEntries)
		}
		if len(i.methodScopeHasCache) >= methodScopeLookupCacheMaxEntries {
			clear(i.methodScopeHasCache)
		}
		i.methodScopeHasCache[key] = entry
		return
	}
	i.methodCacheMu.Lock()
	defer i.methodCacheMu.Unlock()
	if i.methodScopeHasCache == nil {
		i.methodScopeHasCache = make(map[methodScopeHasCacheKey]methodScopeHasCacheEntry, methodScopeLookupCacheInitialEntries)
	}
	if len(i.methodScopeHasCache) >= methodScopeLookupCacheMaxEntries {
		clear(i.methodScopeHasCache)
	}
	i.methodScopeHasCache[key] = entry
}
