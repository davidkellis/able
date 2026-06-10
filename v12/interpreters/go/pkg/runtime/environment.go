package runtime

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
)

// Environment provides lexical scoping for Able runtime values.
type Environment struct {
	values       map[string]Value
	inlineNames  [environmentInlineBindingCapacity]string
	inlineValues [environmentInlineBindingCapacity]Value
	spill        *environmentHintedBindings
	parent       *Environment
	state        atomic.Pointer[environmentState]
	shared       *environmentSharedState
	version      uint64
	valueCap     uint32
	inlineCount  uint8
}

const environmentInlineBindingCapacity = 4
const environmentHintedSpillBindingCapacity = 6

type environmentHintedBindings struct {
	count    uint8
	bindings [environmentHintedSpillBindingCapacity]EnvironmentBinding
}

type environmentSharedState struct {
	id                  uint64
	threadMode          atomic.Bool
	runtimeDataVersion  atomic.Uint64
	bindingShapeVersion atomic.Uint64
	bindingNameMu       sync.Mutex
	bindingNameVersions map[string]uint64
}

// EnvironmentBinding seeds or updates a named value within an environment.
type EnvironmentBinding struct {
	Name  string
	Value Value
}

// NewEnvironment creates a new environment, optionally nested under a parent.
func NewEnvironment(parent *Environment) *Environment {
	return NewEnvironmentWithValueCapacity(parent, 0)
}

func newEnvironmentBase(parent *Environment) *Environment {
	var shared *environmentSharedState
	if parent != nil && parent.shared != nil {
		shared = parent.shared
	}
	if shared == nil {
		shared = newEnvironmentSharedState()
	}
	return &Environment{
		parent: parent,
		shared: shared,
	}
}

func newEnvironmentWithStoredValueCapacity(parent *Environment, valueCapacity int) *Environment {
	env := newEnvironmentBase(parent)
	if valueCapacity > environmentInlineBindingCapacity {
		env.valueCap = uint32(valueCapacity)
	}
	return env
}

// NewEnvironmentWithValueCapacity creates a new environment with an optional
// pre-sized value map for callers that know they will bind a fixed number of
// locals immediately.
func NewEnvironmentWithValueCapacity(parent *Environment, valueCapacity int) *Environment {
	return newEnvironmentWithStoredValueCapacity(parent, valueCapacity)
}

// NewEnvironmentWithSingleBinding creates a new environment and seeds it with
// a single current-scope binding without merge semantics.
func NewEnvironmentWithSingleBinding(parent *Environment, valueCapacity int, name string, value Value) *Environment {
	env := newEnvironmentWithStoredValueCapacity(parent, valueCapacity)
	if name == "" {
		return env
	}
	env.inlineNames[0] = name
	env.inlineValues[0] = value
	env.inlineCount = 1
	env.version = 1
	return env
}

// NewEnvironmentWithBindings creates a new environment and seeds it with the
// provided current-scope bindings without merge semantics.
func NewEnvironmentWithBindings(parent *Environment, valueCapacity int, bindings []EnvironmentBinding) *Environment {
	effectiveCapacity := valueCapacity
	if len(bindings) > effectiveCapacity {
		effectiveCapacity = len(bindings)
	}
	if len(bindings) == 1 && bindings[0].Name != "" {
		return NewEnvironmentWithSingleBinding(parent, effectiveCapacity, bindings[0].Name, bindings[0].Value)
	}
	env := NewEnvironmentWithValueCapacity(parent, effectiveCapacity)
	env.defineWithoutMergeBindingsNoLock(bindings)
	return env
}

// NewEnvironmentWithBindingSets creates a new environment and seeds it with up
// to two ordered binding sets without merge semantics. Later bindings shadow
// earlier ones when names overlap.
func NewEnvironmentWithBindingSets(parent *Environment, valueCapacity int, first []EnvironmentBinding, second []EnvironmentBinding) *Environment {
	effectiveCapacity := valueCapacity
	totalBindings := len(first) + len(second)
	if totalBindings > effectiveCapacity {
		effectiveCapacity = totalBindings
	}
	if totalBindings == 1 {
		if len(first) == 1 && first[0].Name != "" {
			return NewEnvironmentWithSingleBinding(parent, effectiveCapacity, first[0].Name, first[0].Value)
		}
		if len(second) == 1 && second[0].Name != "" {
			return NewEnvironmentWithSingleBinding(parent, effectiveCapacity, second[0].Name, second[0].Value)
		}
	}
	env := NewEnvironmentWithValueCapacity(parent, effectiveCapacity)
	env.defineWithoutMergeBindingSetsNoLock(first, second)
	return env
}

func (e *Environment) currentValueCountNoLock() int {
	if e.values != nil {
		return len(e.values)
	}
	if e.spill != nil {
		return int(e.spill.count)
	}
	return int(e.inlineCount)
}

func (e *Environment) inlineValueIndexNoLock(name string) int {
	switch e.inlineCount {
	case 0:
		return -1
	case 1:
		if e.inlineNames[0] == name {
			return 0
		}
	case 2:
		if e.inlineNames[0] == name {
			return 0
		}
		if e.inlineNames[1] == name {
			return 1
		}
	default:
		for idx := 0; idx < int(e.inlineCount); idx++ {
			if e.inlineNames[idx] == name {
				return idx
			}
		}
	}
	return -1
}

func (e *Environment) lookupSpillValueNoLock(name string) (Value, bool) {
	if e == nil || e.spill == nil {
		return nil, false
	}
	switch e.spill.count {
	case 0:
		return nil, false
	case 1:
		if e.spill.bindings[0].Name == name {
			return e.spill.bindings[0].Value, true
		}
	case 2:
		if e.spill.bindings[0].Name == name {
			return e.spill.bindings[0].Value, true
		}
		if e.spill.bindings[1].Name == name {
			return e.spill.bindings[1].Value, true
		}
	default:
		for idx := 0; idx < int(e.spill.count); idx++ {
			if e.spill.bindings[idx].Name == name {
				return e.spill.bindings[idx].Value, true
			}
		}
	}
	return nil, false
}

func (e *Environment) lookupCurrentValueNoLock(name string) (Value, bool) {
	if e.values != nil {
		v, ok := e.values[name]
		return v, ok
	}
	if e.spill != nil {
		if v, ok := e.lookupSpillValueNoLock(name); ok {
			return v, true
		}
	}
	switch e.inlineCount {
	case 0:
		return nil, false
	case 1:
		if e.inlineNames[0] == name {
			return e.inlineValues[0], true
		}
	case 2:
		if e.inlineNames[0] == name {
			return e.inlineValues[0], true
		}
		if e.inlineNames[1] == name {
			return e.inlineValues[1], true
		}
	default:
		if idx := e.inlineValueIndexNoLock(name); idx >= 0 {
			return e.inlineValues[idx], true
		}
	}
	return nil, false
}

func (e *Environment) lookupValueWithOwnerAndRevisionSingleThread(name string) (Value, *Environment, uint64, bool) {
	if e == nil {
		return nil, nil, 0, false
	}
	if v, ok := e.lookupCurrentValueNoLock(name); ok {
		return v, e, e.version, true
	}
	parent := e.parent
	if parent == nil {
		return nil, nil, 0, false
	}
	if v, ok := parent.lookupCurrentValueNoLock(name); ok {
		return v, parent, parent.version, true
	}
	grandParent := parent.parent
	if grandParent == nil {
		return nil, nil, 0, false
	}
	if v, ok := grandParent.lookupCurrentValueNoLock(name); ok {
		return v, grandParent, grandParent.version, true
	}
	for cur := grandParent.parent; cur != nil; cur = cur.parent {
		if v, ok := cur.lookupCurrentValueNoLock(name); ok {
			return v, cur, cur.version, true
		}
	}
	return nil, nil, 0, false
}

func (e *Environment) lookupValueWithOwnerSingleThread(name string) (Value, *Environment, bool) {
	v, owner, _, ok := e.lookupValueWithOwnerAndRevisionSingleThread(name)
	return v, owner, ok
}

func (e *Environment) canUseHintedSpillNoLock(minCapacity int) bool {
	if e == nil || e.values != nil || e.spill != nil {
		return false
	}
	if int(e.valueCap) <= environmentInlineBindingCapacity {
		return false
	}
	if minCapacity < int(e.valueCap) {
		minCapacity = int(e.valueCap)
	}
	return minCapacity <= environmentHintedSpillBindingCapacity
}

func (e *Environment) promoteInlineBindingsNoLock(minCapacity int) {
	if e.values != nil || e.spill != nil {
		return
	}
	if e.canUseHintedSpillNoLock(minCapacity) {
		spill := &environmentHintedBindings{count: e.inlineCount}
		for idx := 0; idx < int(e.inlineCount); idx++ {
			spill.bindings[idx] = EnvironmentBinding{
				Name:  e.inlineNames[idx],
				Value: e.inlineValues[idx],
			}
			e.inlineNames[idx] = ""
			e.inlineValues[idx] = nil
		}
		e.inlineCount = 0
		e.spill = spill
		return
	}
	if minCapacity < int(e.valueCap) {
		minCapacity = int(e.valueCap)
	}
	if minCapacity < environmentInlineBindingCapacity+1 {
		minCapacity = environmentInlineBindingCapacity + 1
	}
	e.values = make(map[string]Value, minCapacity)
	for idx := 0; idx < int(e.inlineCount); idx++ {
		e.values[e.inlineNames[idx]] = e.inlineValues[idx]
		e.inlineNames[idx] = ""
		e.inlineValues[idx] = nil
	}
	e.inlineCount = 0
	e.valueCap = 0
}

func (e *Environment) promoteSpillBindingsNoLock(minCapacity int) {
	if e == nil || e.values != nil || e.spill == nil {
		return
	}
	if minCapacity < int(e.valueCap) {
		minCapacity = int(e.valueCap)
	}
	if minCapacity < environmentInlineBindingCapacity+1 {
		minCapacity = environmentInlineBindingCapacity + 1
	}
	e.values = make(map[string]Value, minCapacity)
	for idx := 0; idx < int(e.spill.count); idx++ {
		binding := e.spill.bindings[idx]
		if binding.Name != "" {
			e.values[binding.Name] = binding.Value
		}
		e.spill.bindings[idx] = EnvironmentBinding{}
	}
	e.spill.count = 0
	e.spill = nil
	e.valueCap = 0
}

func (e *Environment) ensureBindingCapacityNoLock(target int) {
	if e == nil || e.values != nil {
		return
	}
	if e.spill != nil {
		if target > environmentHintedSpillBindingCapacity {
			e.promoteSpillBindingsNoLock(target)
		}
		return
	}
	if target > environmentInlineBindingCapacity {
		e.promoteInlineBindingsNoLock(target)
	}
}

func (e *Environment) setCurrentValueNoLock(name string, value Value) {
	if e.values != nil {
		e.values[name] = value
		return
	}
	if e.spill != nil {
		for idx := 0; idx < int(e.spill.count); idx++ {
			if e.spill.bindings[idx].Name == name {
				e.spill.bindings[idx].Value = value
				return
			}
		}
		if int(e.spill.count) < len(e.spill.bindings) {
			e.spill.bindings[e.spill.count] = EnvironmentBinding{Name: name, Value: value}
			e.spill.count++
			return
		}
		e.promoteSpillBindingsNoLock(int(e.spill.count) + 1)
		e.values[name] = value
		return
	}
	if idx := e.inlineValueIndexNoLock(name); idx >= 0 {
		e.inlineValues[idx] = value
		return
	}
	if int(e.inlineCount) < environmentInlineBindingCapacity {
		e.inlineNames[e.inlineCount] = name
		e.inlineValues[e.inlineCount] = value
		e.inlineCount++
		return
	}
	e.promoteInlineBindingsNoLock(e.currentValueCountNoLock() + 1)
	e.setCurrentValueNoLock(name, value)
}

func (e *Environment) isSingleThread() bool {
	return e != nil && e.shared != nil && e.shared.threadMode.Load()
}

// SetSingleThread marks the entire scope chain as single-threaded,
// allowing lock-free access. Call this at startup before any goroutines
// are spawned. Call SetMultiThread before the first spawn.
func (e *Environment) SetSingleThread() {
	if e == nil || e.shared == nil {
		return
	}
	e.shared.threadMode.Store(true)
}

// SetMultiThread reverts the entire scope chain to locked access.
func (e *Environment) SetMultiThread() {
	if e == nil || e.shared == nil {
		return
	}
	e.shared.threadMode.Store(false)
}

// Parent exposes the lexical parent (nil when global).
func (e *Environment) Parent() *Environment {
	if e.isSingleThread() {
		return e.parent
	}
	mu := e.mutex()
	mu.RLock()
	parent := e.parent
	mu.RUnlock()
	return parent
}

// Snapshot returns a deterministic copy of the current bindings.
func (e *Environment) Snapshot() map[string]Value {
	if e.isSingleThread() {
		out := make(map[string]Value, e.currentValueCountNoLock())
		for k, v := range e.values {
			out[k] = v
		}
		if e.spill != nil {
			for idx := 0; idx < int(e.spill.count); idx++ {
				out[e.spill.bindings[idx].Name] = e.spill.bindings[idx].Value
			}
		}
		for idx := 0; idx < int(e.inlineCount); idx++ {
			out[e.inlineNames[idx]] = e.inlineValues[idx]
		}
		return out
	}
	mu := e.mutex()
	mu.RLock()
	out := make(map[string]Value, e.currentValueCountNoLock())
	for k, v := range e.values {
		out[k] = v
	}
	if e.spill != nil {
		for idx := 0; idx < int(e.spill.count); idx++ {
			out[e.spill.bindings[idx].Name] = e.spill.bindings[idx].Value
		}
	}
	for idx := 0; idx < int(e.inlineCount); idx++ {
		out[e.inlineNames[idx]] = e.inlineValues[idx]
	}
	mu.RUnlock()
	return out
}

// StructSnapshot returns a deterministic copy of the current struct bindings.
func (e *Environment) StructSnapshot() map[string]*StructDefinitionValue {
	var structs map[string]*StructDefinitionValue
	if e.isSingleThread() {
		if meta := e.metaNoLock(); meta != nil {
			structs = meta.structs
		}
		out := make(map[string]*StructDefinitionValue, len(structs))
		for k, v := range structs {
			out[k] = v
		}
		return out
	}
	mu := e.mutex()
	mu.RLock()
	if meta := e.metaNoLock(); meta != nil {
		structs = meta.structs
	}
	out := make(map[string]*StructDefinitionValue, len(structs))
	for k, v := range structs {
		out[k] = v
	}
	mu.RUnlock()
	return out
}

// Define inserts or shadows a binding in the current scope.
func (e *Environment) Define(name string, value Value) {
	if e.isSingleThread() {
		if existing, ok := e.lookupCurrentValueNoLock(name); ok {
			if merged, ok := MergeFunctionValues(existing, value); ok {
				e.setCurrentValueNoLock(name, merged)
				e.version++
				return
			}
			e.setCurrentValueNoLock(name, value)
			e.version++
			return
		}
		e.setCurrentValueNoLock(name, value)
		e.version++
		e.bumpBindingShapeVersion()
		e.bumpBindingNameVersion(name)
		return
	}
	mu := e.mutex()
	mu.Lock()
	if existing, ok := e.lookupCurrentValueNoLock(name); ok {
		if merged, ok := MergeFunctionValues(existing, value); ok {
			e.setCurrentValueNoLock(name, merged)
			e.version++
			mu.Unlock()
			return
		}
		e.setCurrentValueNoLock(name, value)
		e.version++
		mu.Unlock()
		return
	}
	e.setCurrentValueNoLock(name, value)
	e.version++
	e.bumpBindingShapeVersion()
	e.bumpBindingNameVersion(name)
	mu.Unlock()
}

// DefineWithoutMerge inserts or shadows a binding in the current scope without
// checking function-merge semantics. This is appropriate for plain local
// bindings such as pattern matches where merge behavior is never desired.
func (e *Environment) DefineWithoutMerge(name string, value Value) {
	if e.isSingleThread() {
		_, existed := e.lookupCurrentValueNoLock(name)
		e.setCurrentValueNoLock(name, value)
		e.version++
		if !existed {
			e.bumpBindingShapeVersion()
			e.bumpBindingNameVersion(name)
		}
		return
	}
	mu := e.mutex()
	mu.Lock()
	_, existed := e.lookupCurrentValueNoLock(name)
	e.setCurrentValueNoLock(name, value)
	e.version++
	if !existed {
		e.bumpBindingShapeVersion()
		e.bumpBindingNameVersion(name)
	}
	mu.Unlock()
}

func (e *Environment) applyWithoutMergeBindingsNoLock(bindings []EnvironmentBinding) (int, int) {
	if e == nil || len(bindings) == 0 {
		return 0, 0
	}
	applied := 0
	added := 0
	for _, binding := range bindings {
		if binding.Name == "" {
			continue
		}
		if _, existed := e.lookupCurrentValueNoLock(binding.Name); !existed {
			added++
		}
		e.setCurrentValueNoLock(binding.Name, binding.Value)
		applied++
	}
	return applied, added
}

func (e *Environment) defineWithoutMergeBindingsNoLock(bindings []EnvironmentBinding) (int, int) {
	if e == nil || len(bindings) == 0 {
		return 0, 0
	}
	e.ensureBindingCapacityNoLock(e.currentValueCountNoLock() + len(bindings))
	applied, added := e.applyWithoutMergeBindingsNoLock(bindings)
	if applied == 0 {
		return 0, 0
	}
	e.version += uint64(applied)
	return applied, added
}

func (e *Environment) defineWithoutMergeBindingSetsNoLock(first []EnvironmentBinding, second []EnvironmentBinding) (int, int) {
	totalBindings := len(first) + len(second)
	if e == nil || totalBindings == 0 {
		return 0, 0
	}
	e.ensureBindingCapacityNoLock(e.currentValueCountNoLock() + totalBindings)
	applied, added := e.applyWithoutMergeBindingsNoLock(first)
	secondApplied, secondAdded := e.applyWithoutMergeBindingsNoLock(second)
	applied += secondApplied
	added += secondAdded
	if applied == 0 {
		return 0, 0
	}
	e.version += uint64(applied)
	return applied, added
}

// DefineWithoutMergeBindings inserts or shadows multiple bindings in the
// current scope without checking function-merge semantics.
func (e *Environment) DefineWithoutMergeBindings(bindings []EnvironmentBinding) {
	if e == nil || len(bindings) == 0 {
		return
	}
	if e.isSingleThread() {
		_, added := e.defineWithoutMergeBindingsNoLock(bindings)
		if added > 0 {
			e.bumpBindingShapeVersion()
			e.bumpBindingNameVersionsForBindings(bindings)
		}
		return
	}
	mu := e.mutex()
	mu.Lock()
	_, added := e.defineWithoutMergeBindingsNoLock(bindings)
	if added > 0 {
		e.bumpBindingShapeVersion()
		e.bumpBindingNameVersionsForBindings(bindings)
	}
	mu.Unlock()
}

// DefineStruct records a struct definition in the current scope.
func (e *Environment) DefineStruct(name string, def *StructDefinitionValue) {
	if def == nil {
		return
	}
	if e.isSingleThread() {
		meta := e.ensureMetaNoLock()
		if meta.structs == nil {
			meta.structs = make(map[string]*StructDefinitionValue)
		}
		meta.structs[name] = def
		return
	}
	mu := e.mutex()
	mu.Lock()
	meta := e.ensureMetaNoLock()
	if meta.structs == nil {
		meta.structs = make(map[string]*StructDefinitionValue)
	}
	meta.structs[name] = def
	mu.Unlock()
}

// StructDefinition retrieves a struct definition, searching outward through the scope chain.
func (e *Environment) StructDefinition(name string) (*StructDefinitionValue, bool) {
	singleThread := e.isSingleThread()
	for cur := e; cur != nil; {
		if singleThread {
			if meta := cur.metaNoLock(); meta != nil {
				if v, ok := meta.structs[name]; ok {
					return v, true
				}
			}
			cur = cur.parent
		} else {
			mu := cur.mutex()
			mu.RLock()
			var (
				v      *StructDefinitionValue
				ok     bool
				parent *Environment
			)
			if meta := cur.metaNoLock(); meta != nil {
				v, ok = meta.structs[name]
			}
			parent = cur.parent
			mu.RUnlock()
			if ok {
				return v, true
			}
			cur = parent
		}
	}
	return nil, false
}

// StructDefinitionInCurrentScope retrieves a struct definition without walking
// parent environments. Package registrars use this to avoid treating a
// same-named definition from another package as a local declaration.
func (e *Environment) StructDefinitionInCurrentScope(name string) (*StructDefinitionValue, bool) {
	if e == nil || name == "" {
		return nil, false
	}
	if e.isSingleThread() {
		meta := e.metaNoLock()
		if meta == nil {
			return nil, false
		}
		def, ok := meta.structs[name]
		return def, ok
	}
	mu := e.mutex()
	mu.RLock()
	defer mu.RUnlock()
	meta := e.metaNoLock()
	if meta == nil {
		return nil, false
	}
	def, ok := meta.structs[name]
	return def, ok
}

// Assign updates an existing binding in the first scope where it appears.
func (e *Environment) Assign(name string, value Value) error {
	singleThread := e.isSingleThread()
	if singleThread {
		if _, owner, ok := e.lookupValueWithOwnerSingleThread(name); ok {
			owner.setCurrentValueNoLock(name, value)
			owner.version++
			return nil
		}
		return fmt.Errorf("Undefined variable '%s'", name)
	}
	for cur := e; cur != nil; {
		mu := cur.mutex()
		mu.Lock()
		if _, ok := cur.lookupCurrentValueNoLock(name); ok {
			cur.setCurrentValueNoLock(name, value)
			cur.version++
			mu.Unlock()
			return nil
		}
		parent := cur.parent
		mu.Unlock()
		cur = parent
	}
	return fmt.Errorf("Undefined variable '%s'", name)
}

// Get retrieves a binding, searching outward through the scope chain.
func (e *Environment) Get(name string) (Value, error) {
	singleThread := e.isSingleThread()
	if singleThread {
		if v, _, ok := e.lookupValueWithOwnerSingleThread(name); ok {
			return v, nil
		}
		return nil, fmt.Errorf("Undefined variable '%s'", name)
	}
	for cur := e; cur != nil; {
		mu := cur.mutex()
		mu.RLock()
		v, ok := cur.lookupCurrentValueNoLock(name)
		parent := cur.parent
		mu.RUnlock()
		if ok {
			return v, nil
		}
		cur = parent
	}
	return nil, fmt.Errorf("Undefined variable '%s'", name)
}

// Lookup retrieves a binding, searching outward through the scope chain.
// It avoids constructing an error on misses and is preferred in hot paths
// where absence is expected.
func (e *Environment) Lookup(name string) (Value, bool) {
	singleThread := e.isSingleThread()
	if singleThread {
		if v, _, ok := e.lookupValueWithOwnerSingleThread(name); ok {
			return v, true
		}
		return nil, false
	}
	for cur := e; cur != nil; {
		mu := cur.mutex()
		mu.RLock()
		v, ok := cur.lookupCurrentValueNoLock(name)
		parent := cur.parent
		mu.RUnlock()
		if ok {
			return v, true
		}
		cur = parent
	}
	return nil, false
}

// LookupWithOwner retrieves a binding plus the lexical scope that currently
// owns it. It avoids constructing an error on misses and is preferred in hot
// paths that want to cache parent/global hits without rewalking the chain.
func (e *Environment) LookupWithOwner(name string) (Value, *Environment, bool) {
	singleThread := e.isSingleThread()
	v, owner, _, ok := e.LookupWithOwnerAndRevisionHint(name, singleThread)
	return v, owner, ok
}

// LookupWithOwnerAndRevisionHint retrieves a binding, its lexical owner, and
// the owner's mutation revision while letting the caller reuse already-known
// thread mode.
func (e *Environment) LookupWithOwnerAndRevisionHint(name string, singleThread bool) (Value, *Environment, uint64, bool) {
	if e == nil {
		return nil, nil, 0, false
	}
	if singleThread {
		return e.lookupValueWithOwnerAndRevisionSingleThread(name)
	}
	for cur := e; cur != nil; {
		mu := cur.mutex()
		mu.RLock()
		v, ok := cur.lookupCurrentValueNoLock(name)
		version := cur.version
		parent := cur.parent
		mu.RUnlock()
		if ok {
			return v, cur, version, true
		}
		cur = parent
	}
	return nil, nil, 0, false
}

// LookupInCurrentScope retrieves a binding only from the current scope.
// It avoids constructing an error on misses and does not walk lexical parents.
func (e *Environment) LookupInCurrentScope(name string) (Value, bool) {
	if e.isSingleThread() {
		return e.lookupCurrentValueNoLock(name)
	}
	mu := e.mutex()
	mu.RLock()
	v, ok := e.lookupCurrentValueNoLock(name)
	mu.RUnlock()
	return v, ok
}

// Keys returns the bindings in sorted order (useful for determinism in tests).
func (e *Environment) Keys() []string {
	if e.isSingleThread() {
		keys := make([]string, 0, e.currentValueCountNoLock())
		for k := range e.values {
			keys = append(keys, k)
		}
		if e.spill != nil {
			for idx := 0; idx < int(e.spill.count); idx++ {
				keys = append(keys, e.spill.bindings[idx].Name)
			}
		}
		for idx := 0; idx < int(e.inlineCount); idx++ {
			keys = append(keys, e.inlineNames[idx])
		}
		sort.Strings(keys)
		return keys
	}
	mu := e.mutex()
	mu.RLock()
	keys := make([]string, 0, e.currentValueCountNoLock())
	for k := range e.values {
		keys = append(keys, k)
	}
	if e.spill != nil {
		for idx := 0; idx < int(e.spill.count); idx++ {
			keys = append(keys, e.spill.bindings[idx].Name)
		}
	}
	for idx := 0; idx < int(e.inlineCount); idx++ {
		keys = append(keys, e.inlineNames[idx])
	}
	mu.RUnlock()
	sort.Strings(keys)
	return keys
}

// Extend clones the current environment into a new child scope.
func (e *Environment) Extend() *Environment {
	return NewEnvironment(e)
}

// Has reports whether the binding exists anywhere in the scope chain.
func (e *Environment) Has(name string) bool {
	singleThread := e.isSingleThread()
	if singleThread {
		_, _, ok := e.lookupValueWithOwnerSingleThread(name)
		return ok
	}
	for cur := e; cur != nil; {
		mu := cur.mutex()
		mu.RLock()
		_, ok := cur.lookupCurrentValueNoLock(name)
		parent := cur.parent
		mu.RUnlock()
		if ok {
			return true
		}
		cur = parent
	}
	return false
}

// HasInCurrentScope reports whether the binding exists in the current scope.
func (e *Environment) HasInCurrentScope(name string) bool {
	if e.isSingleThread() {
		_, ok := e.lookupCurrentValueNoLock(name)
		return ok
	}
	mu := e.mutex()
	mu.RLock()
	_, ok := e.lookupCurrentValueNoLock(name)
	mu.RUnlock()
	return ok
}

// AssignExisting assigns a name if it exists anywhere in the scope chain.
// Returns true when the assignment succeeded.
func (e *Environment) AssignExisting(name string, value Value) bool {
	singleThread := e.isSingleThread()
	if singleThread {
		if _, owner, ok := e.lookupValueWithOwnerSingleThread(name); ok {
			owner.setCurrentValueNoLock(name, value)
			owner.version++
			return true
		}
		return false
	}
	for cur := e; cur != nil; {
		mu := cur.mutex()
		mu.Lock()
		if _, ok := cur.lookupCurrentValueNoLock(name); ok {
			cur.setCurrentValueNoLock(name, value)
			cur.version++
			mu.Unlock()
			return true
		}
		parent := cur.parent
		mu.Unlock()
		cur = parent
	}
	return false
}

// Revision returns the mutation revision for this scope.
func (e *Environment) Revision() uint64 {
	if e.isSingleThread() {
		return e.version
	}
	mu := e.mutex()
	mu.RLock()
	version := e.version
	mu.RUnlock()
	return version
}

// RevisionWithHint returns the mutation revision for this scope while letting
// the caller supply the already-known thread mode. This avoids repeating the
// shared thread-mode load on hot paths that already know execution is
// single-threaded.
func (e *Environment) RevisionWithHint(singleThread bool) uint64 {
	if e == nil {
		return 0
	}
	if singleThread {
		return e.version
	}
	mu := e.mutex()
	mu.RLock()
	version := e.version
	mu.RUnlock()
	return version
}

// RevisionSingleThread returns the mutation revision when the caller already
// owns the single-threaded execution invariant.
func (e *Environment) RevisionSingleThread() uint64 {
	if e == nil {
		return 0
	}
	return e.version
}
