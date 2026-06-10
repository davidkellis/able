package runtime

func (e *Environment) resetForReuseNoLock(parent *Environment, valueCapacity int) bool {
	if e == nil {
		return false
	}
	e.bumpCurrentBindingNameVersionsNoLock()
	if e.values != nil {
		clear(e.values)
		e.values = nil
	}
	if e.spill != nil {
		clear(e.spill.bindings[:])
		e.spill.count = 0
		e.spill = nil
	}
	for idx := range e.inlineNames {
		e.inlineNames[idx] = ""
		e.inlineValues[idx] = nil
	}
	hadRuntimeData := false
	if meta := e.metaNoLock(); meta != nil {
		hadRuntimeData = meta.data != nil
		if meta.structs != nil {
			clear(meta.structs)
			meta.structs = nil
		}
		meta.data = nil
	}
	e.parent = parent
	if parent != nil && parent.shared != nil {
		e.shared = parent.shared
	} else {
		e.shared = newEnvironmentSharedState()
	}
	e.valueCap = 0
	if valueCapacity > environmentInlineBindingCapacity {
		e.valueCap = uint32(valueCapacity)
	}
	e.inlineCount = 0
	e.version++
	e.bumpBindingShapeVersion()
	if hadRuntimeData {
		e.bumpRuntimeDataVersion()
	}
	return true
}

func (e *Environment) resetForSingleBindingReuseNoLock(parent *Environment, valueCapacity int, name string, value Value) {
	if !e.resetForReuseNoLock(parent, valueCapacity) {
		return
	}
	if name == "" {
		return
	}
	e.inlineNames[0] = name
	e.inlineValues[0] = value
	e.inlineCount = 1
	e.version++
	e.bumpBindingNameVersion(name)
}

// ResetForSingleBindingReuse clears a transient environment so it can be reused
// as a lexical child with one initial binding while preserving a monotonically
// increasing revision for cache invalidation.
func (e *Environment) ResetForSingleBindingReuse(parent *Environment, valueCapacity int, name string, value Value) {
	if e == nil {
		return
	}
	if e.isSingleThread() {
		e.resetForSingleBindingReuseNoLock(parent, valueCapacity, name, value)
		return
	}
	mu := e.mutex()
	mu.Lock()
	e.resetForSingleBindingReuseNoLock(parent, valueCapacity, name, value)
	mu.Unlock()
}

func (e *Environment) resetForBindingSetsReuseNoLock(parent *Environment, valueCapacity int, first []EnvironmentBinding, second []EnvironmentBinding) {
	if !e.resetForReuseNoLock(parent, valueCapacity) {
		return
	}
	e.defineWithoutMergeBindingSetsNoLock(first, second)
	e.bumpBindingNameVersionsForBindings(first)
	e.bumpBindingNameVersionsForBindings(second)
}

// ResetForBindingSetsReuse clears a transient environment so it can be reused
// as a lexical child with up to two initial binding sets while preserving a
// monotonically increasing revision for cache invalidation.
func (e *Environment) ResetForBindingSetsReuse(parent *Environment, valueCapacity int, first []EnvironmentBinding, second []EnvironmentBinding) {
	if e == nil {
		return
	}
	if e.isSingleThread() {
		e.resetForBindingSetsReuseNoLock(parent, valueCapacity, first, second)
		return
	}
	mu := e.mutex()
	mu.Lock()
	e.resetForBindingSetsReuseNoLock(parent, valueCapacity, first, second)
	mu.Unlock()
}
