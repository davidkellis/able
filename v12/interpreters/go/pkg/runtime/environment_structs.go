package runtime

// ForEachCurrentStructDefinition visits struct definitions in the current scope
// without allocating a snapshot map. Iteration stops when visit returns false.
func (e *Environment) ForEachCurrentStructDefinition(visit func(string, *StructDefinitionValue) bool) {
	if e == nil || visit == nil {
		return
	}
	if e.isSingleThread() {
		meta := e.metaNoLock()
		if meta == nil {
			return
		}
		for name, def := range meta.structs {
			if !visit(name, def) {
				return
			}
		}
		return
	}
	mu := e.mutex()
	mu.RLock()
	if meta := e.metaNoLock(); meta != nil {
		for name, def := range meta.structs {
			if !visit(name, def) {
				break
			}
		}
	}
	mu.RUnlock()
}
