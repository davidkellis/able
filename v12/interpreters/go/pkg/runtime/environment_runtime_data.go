package runtime

// SetRuntimeData attaches interpreter-specific metadata to the environment.
func (e *Environment) SetRuntimeData(data any) {
	if e.isSingleThread() {
		e.ensureMetaNoLock().data = data
		e.bumpRuntimeDataVersion()
		return
	}
	mu := e.mutex()
	mu.Lock()
	e.ensureMetaNoLock().data = data
	mu.Unlock()
	e.bumpRuntimeDataVersion()
}

// RuntimeData returns the metadata associated with this environment, falling
// back to parents.
func (e *Environment) RuntimeData() any {
	if e == nil {
		return nil
	}
	singleThread := e.isSingleThread()
	for cur := e; cur != nil; {
		if singleThread {
			if meta := cur.metaNoLock(); meta != nil && meta.data != nil {
				return meta.data
			}
			cur = cur.parent
			continue
		}
		mu := cur.mutex()
		mu.RLock()
		var data any
		if meta := cur.metaNoLock(); meta != nil {
			data = meta.data
		}
		parent := cur.parent
		mu.RUnlock()
		if data != nil {
			return data
		}
		cur = parent
	}
	return nil
}

func (e *Environment) bumpRuntimeDataVersion() {
	if e == nil || e.shared == nil {
		return
	}
	e.shared.runtimeDataVersion.Add(1)
}

// RuntimeDataRevision returns the mutation revision for runtime metadata across
// the full lexical chain rooted at this environment.
func (e *Environment) RuntimeDataRevision() uint64 {
	if e == nil || e.shared == nil {
		return 0
	}
	return e.shared.runtimeDataVersion.Load()
}
