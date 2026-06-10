package runtime

import "sync/atomic"

var environmentSharedStateCounter atomic.Uint64

func newEnvironmentSharedState() *environmentSharedState {
	return &environmentSharedState{
		id: environmentSharedStateCounter.Add(1),
	}
}

// RuntimeDataStateID returns the identity of the shared runtime-data state
// attached to the lexical chain rooted at this environment.
func (e *Environment) RuntimeDataStateID() uint64 {
	if e == nil || e.shared == nil {
		return 0
	}
	return e.shared.id
}

// BindingShapeStateID returns the identity of the shared binding-shape state
// attached to the lexical chain rooted at this environment.
func (e *Environment) BindingShapeStateID() uint64 {
	if e == nil || e.shared == nil {
		return 0
	}
	return e.shared.id
}

// BindingShapeRevision returns a revision that changes when binding existence
// or lexical parentage can affect name lookup results for this environment
// family. Plain assignment to an existing binding is tracked by the owning
// scope's value revision instead.
func (e *Environment) BindingShapeRevision() uint64 {
	if e == nil || e.shared == nil {
		return 0
	}
	return e.shared.bindingShapeVersion.Load()
}

// BindingNameRevision returns a revision that changes when binding existence
// for a specific name changes within this environment family.
func (e *Environment) BindingNameRevision(name string) uint64 {
	if e == nil || e.shared == nil || name == "" {
		return 0
	}
	e.shared.bindingNameMu.Lock()
	version := e.shared.bindingNameVersions[name]
	e.shared.bindingNameMu.Unlock()
	return version
}

func (e *Environment) bumpBindingShapeVersion() {
	if e == nil || e.shared == nil {
		return
	}
	e.shared.bindingShapeVersion.Add(1)
}

func (e *Environment) bumpBindingNameVersion(name string) {
	if e == nil || e.shared == nil || name == "" {
		return
	}
	e.shared.bindingNameMu.Lock()
	if e.shared.bindingNameVersions == nil {
		e.shared.bindingNameVersions = make(map[string]uint64, 8)
	}
	e.shared.bindingNameVersions[name]++
	e.shared.bindingNameMu.Unlock()
}

func (e *Environment) bumpBindingNameVersionsForBindings(bindings []EnvironmentBinding) {
	for _, binding := range bindings {
		e.bumpBindingNameVersion(binding.Name)
	}
}

func (e *Environment) bumpCurrentBindingNameVersionsNoLock() {
	if e == nil {
		return
	}
	if e.values != nil {
		for name := range e.values {
			e.bumpBindingNameVersion(name)
		}
		return
	}
	if e.spill != nil {
		for idx := 0; idx < int(e.spill.count); idx++ {
			e.bumpBindingNameVersion(e.spill.bindings[idx].Name)
		}
		return
	}
	for idx := 0; idx < int(e.inlineCount); idx++ {
		e.bumpBindingNameVersion(e.inlineNames[idx])
	}
}
