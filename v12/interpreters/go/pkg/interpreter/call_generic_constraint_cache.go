package interpreter

import "able/interpreter-go/pkg/ast"

type functionCallConstraintResultCacheKey struct {
	function ast.Node
	call     *ast.FunctionCall
	version  uint64
}

type functionCallConstraintResultCacheEntry struct {
	err error
}

func (i *Interpreter) lookupFunctionCallConstraintResultCache(key functionCallConstraintResultCacheKey) (functionCallConstraintResultCacheEntry, bool) {
	if i == nil || key.function == nil {
		return functionCallConstraintResultCacheEntry{}, false
	}
	if i.envSingleThread {
		entry, ok := i.functionCallConstraintResultCache[key]
		return entry, ok
	}
	i.methodCacheMu.RLock()
	defer i.methodCacheMu.RUnlock()
	entry, ok := i.functionCallConstraintResultCache[key]
	return entry, ok
}

func (i *Interpreter) storeFunctionCallConstraintResultCache(key functionCallConstraintResultCacheKey, err error) {
	if i == nil || key.function == nil {
		return
	}
	entry := functionCallConstraintResultCacheEntry{err: err}
	if i.envSingleThread {
		if i.functionCallConstraintResultCache == nil {
			i.functionCallConstraintResultCache = make(map[functionCallConstraintResultCacheKey]functionCallConstraintResultCacheEntry)
		}
		i.functionCallConstraintResultCache[key] = entry
		return
	}
	i.methodCacheMu.Lock()
	defer i.methodCacheMu.Unlock()
	if i.functionCallConstraintResultCache == nil {
		i.functionCallConstraintResultCache = make(map[functionCallConstraintResultCacheKey]functionCallConstraintResultCacheEntry)
	}
	i.functionCallConstraintResultCache[key] = entry
}
