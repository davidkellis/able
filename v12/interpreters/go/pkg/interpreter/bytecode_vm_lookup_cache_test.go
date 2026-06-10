package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func assertBytecodeLoadNamePathStatsZero(t *testing.T, stats BytecodeStatsSnapshot) {
	t.Helper()
	if stats.LoadNameHotHits != 0 ||
		stats.LoadNameScopeHits != 0 ||
		stats.LoadNameGlobalHits != 0 ||
		stats.LoadNameDirectCurrent != 0 ||
		stats.LoadNameDirectOuter != 0 ||
		stats.LoadNameScopeStores != 0 ||
		stats.LoadNameGlobalStores != 0 {
		t.Fatalf(
			"expected zero load-name path stats, got hot=%d scope_hits=%d global_hits=%d direct_current=%d direct_outer=%d scope_stores=%d global_stores=%d",
			stats.LoadNameHotHits,
			stats.LoadNameScopeHits,
			stats.LoadNameGlobalHits,
			stats.LoadNameDirectCurrent,
			stats.LoadNameDirectOuter,
			stats.LoadNameScopeStores,
			stats.LoadNameGlobalStores,
		)
	}
}

func TestBytecodeVM_LookupCachedIdentifierNameUsesHotValueCache(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	want := runtime.NewSmallInt(7, runtime.IntegerI32)
	version := vm.bytecodeEnvRevision(env)
	entry := &bytecodeScopeLookupCacheEntry{
		env:          env,
		envVersion:   version,
		owner:        env,
		ownerVersion: version,
		value:        want,
	}
	vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{
		valid:   true,
		program: program,
		ip:      1,
		entry:   entry,
	}

	got, ok := vm.lookupCachedIdentifierName(program, 1, "x")
	if !ok {
		t.Fatalf("lookupCachedIdentifierName() cache miss, want hit")
	}
	if got != want {
		t.Fatalf("lookupCachedIdentifierName() = %#v, want %#v", got, want)
	}
}

func TestBytecodeVM_ResolveCachedIdentifierNameUsesScopeCache(t *testing.T) {
	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	env := runtime.NewEnvironment(global)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 3)}
	want := runtime.NewSmallInt(11, runtime.IntegerI32)
	env.Define("x", want)
	vm.storeCachedScopeValue(program, 2, "x", env, want)

	got, err := vm.resolveCachedIdentifierName(program, 2, "x")
	if err != nil {
		t.Fatalf("resolveCachedIdentifierName() error = %v", err)
	}
	if got != want {
		t.Fatalf("resolveCachedIdentifierName() = %#v, want %#v", got, want)
	}
}

func TestBytecodeVM_ResolveCachedIdentifierNameRefreshesReusedTransientScope(t *testing.T) {
	interp := NewBytecode()
	parent := runtime.NewEnvironment(interp.GlobalEnvironment())
	parent.SetSingleThread()
	parent.Define("idx", runtime.NewSmallInt(0, runtime.IntegerI32))
	child := runtime.NewEnvironment(parent)
	child.SetSingleThread()
	vm := newBytecodeVM(interp, child)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 1)}

	got, err := vm.resolveCachedIdentifierName(program, 0, "idx")
	if err != nil || !valuesEqual(got, runtime.NewSmallInt(0, runtime.IntegerI32)) {
		t.Fatalf("initial cached lookup = (%#v, %v), want 0", got, err)
	}
	if !parent.AssignExisting("idx", runtime.NewSmallInt(1, runtime.IntegerI32)) {
		t.Fatal("expected outer idx assignment to succeed")
	}
	child.ResetForSingleBindingReuse(parent, 0, "", nil)

	got, err = vm.resolveCachedIdentifierName(program, 0, "idx")
	if err != nil {
		t.Fatalf("lookup after transient reuse failed: %v", err)
	}
	if want := runtime.NewSmallInt(1, runtime.IntegerI32); !valuesEqual(got, want) {
		t.Fatalf("lookup after transient reuse = %#v, want %#v", got, want)
	}
}

func TestBytecodeVM_LookupCachedIdentifierNameDirectResolveSeedsScopeCacheAndHot(t *testing.T) {
	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	env := runtime.NewEnvironment(global)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	want := runtime.NewSmallInt(17, runtime.IntegerI32)
	env.Define("x", want)

	got, ok := vm.lookupCachedIdentifierName(program, 1, "x")
	if !ok {
		t.Fatalf("lookupCachedIdentifierName() direct resolve miss, want hit")
	}
	if got != want {
		t.Fatalf("lookupCachedIdentifierName() = %#v, want %#v", got, want)
	}
	if vm.scopeLookupCache == nil {
		t.Fatalf("expected scope lookup cache to be initialized")
	}
	entry := &vm.scopeLookupCache[program][1]
	if entry.env == nil || entry.owner == nil {
		t.Fatalf("expected direct value lookup to seed scope cache entry")
	}
	if entry.value != want || entry.env != env || entry.owner != env {
		t.Fatalf("scope cache entry = %#v, want value/env/owner %#v/%p/%p", entry, want, env, env)
	}
	if entry.nameShapeStateID != 0 || entry.bindingShapeVersion != 0 || entry.nameShapeVersion != 0 {
		t.Fatalf("same-env owner should skip shape metadata, got state=%d shape=%d name=%d", entry.nameShapeStateID, entry.bindingShapeVersion, entry.nameShapeVersion)
	}
	if !vm.nameLookupHot.valid || vm.nameLookupHot.entry != entry {
		t.Fatalf("expected direct value lookup to seed hot cache with stored entry")
	}
}

func TestBytecodeVM_LookupCachedIdentifierNameRejectsSameEnvNameMismatch(t *testing.T) {
	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	env := runtime.NewEnvironment(global)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	want := runtime.NewSmallInt(19, runtime.IntegerI32)
	env.Define("x", want)
	vm.storeCachedScopeValue(program, 1, "x", env, want)

	vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{}
	if entry, ok := vm.lookupCachedScopeEntryWithVersion(program, 1, "y", env, vm.bytecodeEnvRevision(env), vm.bytecodeSingleThread()); ok || entry != nil {
		t.Fatalf("lookupCachedScopeEntryWithVersion() wrong name = (%#v, %t), want miss", entry, ok)
	}
	if got, ok := vm.lookupCachedIdentifierName(program, 1, "y"); ok || got != nil {
		t.Fatalf("lookupCachedIdentifierName() wrong name = (%#v, %t), want miss", got, ok)
	}
}

func TestBytecodeVM_StoreCachedScopeEntryClearsShapeMetadataForSameEnvOwner(t *testing.T) {
	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	env := runtime.NewEnvironment(global)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	globalValue := runtime.NewSmallInt(23, runtime.IntegerI32)
	localValue := runtime.NewSmallInt(29, runtime.IntegerI32)
	global.Define("x", globalValue)

	vm.storeCachedScopeValue(program, 1, "x", global, globalValue)
	entry := &vm.scopeLookupCache[program][1]
	if entry.nameShapeStateID == 0 {
		t.Fatalf("expected global-owner scope cache entry to record shape state")
	}

	env.Define("x", localValue)
	vm.storeCachedScopeValue(program, 1, "x", env, localValue)
	if entry.value != localValue || entry.owner != env {
		t.Fatalf("same-env overwrite entry = %#v, want owner=%p value=%#v", entry, env, localValue)
	}
	if entry.nameShapeStateID != 0 || entry.bindingShapeVersion != 0 || entry.nameShapeVersion != 0 {
		t.Fatalf("same-env overwrite should clear shape metadata, got state=%d shape=%d name=%d", entry.nameShapeStateID, entry.bindingShapeVersion, entry.nameShapeVersion)
	}
}

func TestBytecodeVM_PrepareRunProgramLazilyActivatesExistingLookupCaches(t *testing.T) {
	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	env := runtime.NewEnvironment(global)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 3)}
	scopeWant := runtime.NewSmallInt(19, runtime.IntegerI32)
	globalWant := runtime.NewSmallInt(21, runtime.IntegerI32)
	env.Define("x", scopeWant)
	global.Define("g", globalWant)
	vm.storeCachedScopeValue(program, 1, "x", env, scopeWant)
	vm.storeCachedGlobalValue(program, 2, "g", globalWant)

	vm.scopeLookupHotProgram = nil
	vm.scopeLookupHotEntries = nil
	vm.globalLookupHotProgram = nil
	vm.globalLookupHotEntries = nil
	vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{}

	vm.prepareRunProgram(program, false)

	if vm.activeLookup.program != program {
		t.Fatalf("active lookup program = %p, want %p", vm.activeLookup.program, program)
	}
	if vm.activeLookup.scopeLookupEntries != nil {
		t.Fatalf("active scope lookup entries should stay lazy until first scope lookup")
	}
	if vm.activeLookup.globalLookupEntries != nil {
		t.Fatalf("active global lookup entries should stay lazy until first global lookup")
	}

	scopeEntry, ok := vm.lookupCachedScopeEntry(program, 1, "x")
	if !ok || scopeEntry == nil || scopeEntry.value != scopeWant {
		t.Fatalf("lookupCachedScopeEntry() = (%#v, %t), want value %#v", scopeEntry, ok, scopeWant)
	}
	if len(vm.activeLookup.scopeLookupEntries) != len(program.instructions) {
		t.Fatalf("active scope lookup entries length = %d, want %d", len(vm.activeLookup.scopeLookupEntries), len(program.instructions))
	}
	globalValue, ok := vm.lookupCachedGlobalValue(program, 2, "g")
	if !ok || globalValue != globalWant {
		t.Fatalf("lookupCachedGlobalValue() = (%#v, %t), want (%#v, true)", globalValue, ok, globalWant)
	}
	if len(vm.activeLookup.globalLookupEntries) != len(program.instructions) {
		t.Fatalf("active global lookup entries length = %d, want %d", len(vm.activeLookup.globalLookupEntries), len(program.instructions))
	}

	vm.scopeLookupCache = nil
	vm.globalLookupCache = nil

	scopeEntry, ok = vm.lookupCachedScopeEntry(program, 1, "x")
	if !ok || scopeEntry == nil || scopeEntry.value != scopeWant {
		t.Fatalf("lookupCachedScopeEntry() after backing cache clear = (%#v, %t), want value %#v", scopeEntry, ok, scopeWant)
	}
	globalValue, ok = vm.lookupCachedGlobalValue(program, 2, "g")
	if !ok || globalValue != globalWant {
		t.Fatalf("lookupCachedGlobalValue() after backing cache clear = (%#v, %t), want (%#v, true)", globalValue, ok, globalWant)
	}
}

func TestBytecodeVM_LookupCachedIdentifierNameEntryUsesScopeCacheAndSeedsHotCache(t *testing.T) {
	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	env := runtime.NewEnvironment(global)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 4)}
	want := runtime.NewSmallInt(23, runtime.IntegerI32)
	env.Define("x", want)
	vm.storeCachedScopeValue(program, 3, "x", env, want)
	vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{}

	lookup, ok := vm.lookupCachedIdentifierNameEntry(program, 3, "x")
	if !ok {
		t.Fatalf("lookupCachedIdentifierNameEntry() cache miss, want hit")
	}
	if lookup.value != want {
		t.Fatalf("lookupCachedIdentifierNameEntry().value = %#v, want %#v", lookup.value, want)
	}
	if lookup.env != env || lookup.owner != env {
		t.Fatalf("lookupCachedIdentifierNameEntry() env metadata = (%p, %p), want (%p, %p)", lookup.env, lookup.owner, env, env)
	}
	if !vm.nameLookupHot.valid {
		t.Fatalf("expected scope cache hit to reseed hot name cache")
	}
	if vm.nameLookupHot.program != program || vm.nameLookupHot.ip != 3 || vm.nameLookupHot.entry == nil {
		t.Fatalf("hot name cache seeded with wrong site metadata: %#v", vm.nameLookupHot)
	}
	if vm.nameLookupHot.entry.value != want {
		t.Fatalf("hot name cache value = %#v, want %#v", vm.nameLookupHot.entry.value, want)
	}
}

func TestBytecodeVM_StoreCachedScopeValueReusesEntryObject(t *testing.T) {
	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	env := runtime.NewEnvironment(global)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 5)}
	first := runtime.NewSmallInt(3, runtime.IntegerI32)
	second := runtime.NewSmallInt(5, runtime.IntegerI32)

	vm.storeCachedScopeValue(program, 4, "x", env, first)
	entry := &vm.scopeLookupCache[program][4]
	if entry.env == nil || entry.owner == nil {
		t.Fatalf("expected scope cache entry after first store")
	}

	vm.storeCachedScopeValue(program, 4, "x", env, second)
	updated := &vm.scopeLookupCache[program][4]
	if updated.env == nil || updated.owner == nil {
		t.Fatalf("expected scope cache entry after second store")
	}
	if updated != entry {
		t.Fatalf("expected scope cache entry object reuse, got %p then %p", entry, updated)
	}
	if updated.value != second {
		t.Fatalf("updated scope cache value = %#v, want %#v", updated.value, second)
	}
	if vm.nameLookupHot.entry != updated {
		t.Fatalf("expected hot name cache to point at reused scope entry")
	}
	if vm.scopeLookupHotProgram != program || len(vm.scopeLookupHotEntries) != len(program.instructions) {
		t.Fatalf("expected scope lookup hot entries to cache program slice")
	}
}

func TestBytecodeVM_CurrentProgramStoresSeedActiveLookupEntries(t *testing.T) {
	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	env := runtime.NewEnvironment(global)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 3)}
	scopeWant := runtime.NewSmallInt(27, runtime.IntegerI32)
	globalWant := runtime.NewSmallInt(31, runtime.IntegerI32)
	env.Define("x", scopeWant)
	global.Define("g", globalWant)

	vm.prepareRunProgram(program, false)
	vm.storeCachedScopeValue(program, 1, "x", env, scopeWant)
	vm.storeCachedGlobalValue(program, 2, "g", globalWant)

	vm.scopeLookupCache = nil
	vm.globalLookupCache = nil
	vm.scopeLookupHotProgram = nil
	vm.scopeLookupHotEntries = nil
	vm.globalLookupHotProgram = nil
	vm.globalLookupHotEntries = nil
	vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{}

	scopeEntry, ok := vm.lookupCachedScopeEntry(program, 1, "x")
	if !ok || scopeEntry == nil || scopeEntry.value != scopeWant {
		t.Fatalf("lookupCachedScopeEntry() = (%#v, %t), want value %#v", scopeEntry, ok, scopeWant)
	}
	globalValue, ok := vm.lookupCachedGlobalValue(program, 2, "g")
	if !ok || globalValue != globalWant {
		t.Fatalf("lookupCachedGlobalValue() = (%#v, %t), want (%#v, true)", globalValue, ok, globalWant)
	}
}

func TestBytecodeVM_LookupCachedIdentifierNameEntrySupportsOuterOwnerScopeCache(t *testing.T) {
	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	env := runtime.NewEnvironment(global)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 6)}
	want := runtime.NewSmallInt(29, runtime.IntegerI32)
	global.Define("x", want)

	vm.storeCachedScopeValue(program, 5, "x", global, want)

	lookup, ok := vm.lookupCachedIdentifierNameEntry(program, 5, "x")
	if !ok {
		t.Fatalf("lookupCachedIdentifierNameEntry() cache miss for outer-owner entry, want hit")
	}
	if lookup.value != want {
		t.Fatalf("lookupCachedIdentifierNameEntry().value = %#v, want %#v", lookup.value, want)
	}
	if lookup.env != env || lookup.owner != global {
		t.Fatalf("lookupCachedIdentifierNameEntry() env metadata = (%p, %p), want (%p, %p)", lookup.env, lookup.owner, env, global)
	}
	if !vm.nameLookupHot.valid || vm.nameLookupHot.entry == nil {
		t.Fatalf("expected hot name cache reseed for outer-owner entry")
	}
	if vm.nameLookupHot.entry.owner != global {
		t.Fatalf("hot name cache owner = %p, want %p", vm.nameLookupHot.entry.owner, global)
	}
}

func TestBytecodeVM_ScopeLookupCacheReusesGlobalOwnerAcrossSameShapeEnvs(t *testing.T) {
	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	envA := runtime.NewEnvironment(global)
	envB := runtime.NewEnvironment(global)
	vm := newBytecodeVM(interp, envA)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	want := runtime.NewSmallInt(37, runtime.IntegerI32)
	global.Define("x", want)

	vm.storeCachedScopeValue(program, 1, "x", global, want)

	vm.env = envB
	got, ok := vm.lookupCachedIdentifierName(program, 1, "x")
	if !ok || got != want {
		t.Fatalf("lookupCachedIdentifierName() across same-shape envs = (%#v, %t), want (%#v, true)", got, ok, want)
	}

	vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{}
	lookup, ok := vm.lookupCachedIdentifierNameEntry(program, 1, "x")
	if !ok || lookup.value != want || lookup.owner != global {
		t.Fatalf("lookupCachedIdentifierNameEntry() backing cache across envs = (%#v, %t), want owner=%p value=%#v", lookup, ok, global, want)
	}
}

func TestBytecodeVM_ScopeLookupCacheRejectsGlobalOwnerAfterLocalShadow(t *testing.T) {
	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	envA := runtime.NewEnvironment(global)
	envB := runtime.NewEnvironment(global)
	vm := newBytecodeVM(interp, envA)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	want := runtime.NewSmallInt(39, runtime.IntegerI32)
	global.Define("x", want)

	vm.storeCachedScopeValue(program, 1, "x", global, want)
	envB.Define("x", runtime.NewSmallInt(41, runtime.IntegerI32))
	vm.env = envB
	currentVersion := vm.bytecodeEnvRevision(envB)
	singleThread := vm.bytecodeSingleThread()
	if entry, ok := vm.lookupHotNameEntryWithVersion(program, 1, "x", envB, currentVersion, singleThread); ok || entry != nil {
		t.Fatalf("lookupHotNameEntryWithVersion() after local shadow = (%#v, %t), want miss", entry, ok)
	}

	vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{}
	if entry, ok := vm.lookupCachedScopeEntryWithVersion(program, 1, "x", envB, currentVersion, singleThread); ok || entry != nil {
		t.Fatalf("lookupCachedScopeEntryWithVersion() after local shadow = (%#v, %t), want miss", entry, ok)
	}
}

func TestBytecodeVM_ScopeLookupCacheKeepsGlobalOwnerAfterUnrelatedLocalBinding(t *testing.T) {
	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	env := runtime.NewEnvironment(global)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	want := runtime.NewSmallInt(40, runtime.IntegerI32)
	global.Define("x", want)

	vm.storeCachedScopeValue(program, 1, "x", global, want)
	env.Define("other", runtime.NewSmallInt(41, runtime.IntegerI32))
	vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{}
	currentVersion := vm.bytecodeEnvRevision(env)
	singleThread := vm.bytecodeSingleThread()

	entry, ok := vm.lookupCachedScopeEntryWithVersion(program, 1, "x", env, currentVersion, singleThread)
	if !ok || entry == nil || entry.value != want || entry.owner != global {
		t.Fatalf("lookupCachedScopeEntryWithVersion() after unrelated local binding = (%#v, %t), want global owner value %#v", entry, ok, want)
	}
	if entry.envVersion != currentVersion {
		t.Fatalf("entry envVersion = %d, want refreshed current version %d", entry.envVersion, currentVersion)
	}

	env.Define("x", runtime.NewSmallInt(42, runtime.IntegerI32))
	currentVersion = vm.bytecodeEnvRevision(env)
	if entry, ok := vm.lookupCachedScopeEntryWithVersion(program, 1, "x", env, currentVersion, singleThread); ok || entry != nil {
		t.Fatalf("lookupCachedScopeEntryWithVersion() after local shadow = (%#v, %t), want miss", entry, ok)
	}
}

func TestBytecodeVM_ScopeLookupCacheKeepsDirectParentOwnerAfterUnrelatedLocalBinding(t *testing.T) {
	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	parent := runtime.NewEnvironment(global)
	env := runtime.NewEnvironment(parent)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	want := runtime.NewSmallInt(44, runtime.IntegerI32)
	parent.Define("x", want)

	vm.storeCachedScopeValue(program, 1, "x", parent, want)
	entry := &vm.scopeLookupCache[program][1]
	if entry.nameShapeStateID == 0 {
		t.Fatalf("expected direct-parent owner scope cache entry to record shape state")
	}

	env.Define("other", runtime.NewSmallInt(45, runtime.IntegerI32))
	vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{}
	currentVersion := vm.bytecodeEnvRevision(env)
	singleThread := vm.bytecodeSingleThread()

	entry, ok := vm.lookupCachedScopeEntryWithVersion(program, 1, "x", env, currentVersion, singleThread)
	if !ok || entry == nil || entry.value != want || entry.owner != parent {
		t.Fatalf("lookupCachedScopeEntryWithVersion() after unrelated local binding = (%#v, %t), want parent owner value %#v", entry, ok, want)
	}
	if entry.envVersion != currentVersion {
		t.Fatalf("entry envVersion = %d, want refreshed current version %d", entry.envVersion, currentVersion)
	}

	env.Define("x", runtime.NewSmallInt(46, runtime.IntegerI32))
	currentVersion = vm.bytecodeEnvRevision(env)
	if entry, ok := vm.lookupCachedScopeEntryWithVersion(program, 1, "x", env, currentVersion, singleThread); ok || entry != nil {
		t.Fatalf("lookupCachedScopeEntryWithVersion() after direct local shadow = (%#v, %t), want miss", entry, ok)
	}
}

func TestBytecodeVM_ScopeLookupCacheRejectsDirectParentOwnerAfterOwnerMutation(t *testing.T) {
	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	parent := runtime.NewEnvironment(global)
	env := runtime.NewEnvironment(parent)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	first := runtime.NewSmallInt(47, runtime.IntegerI32)
	parent.Define("x", first)

	vm.storeCachedScopeValue(program, 1, "x", parent, first)
	parent.Define("x", runtime.NewSmallInt(48, runtime.IntegerI32))
	vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{}
	currentVersion := vm.bytecodeEnvRevision(env)
	singleThread := vm.bytecodeSingleThread()

	if entry, ok := vm.lookupCachedScopeEntryWithVersion(program, 1, "x", env, currentVersion, singleThread); ok || entry != nil {
		t.Fatalf("lookupCachedScopeEntryWithVersion() after parent owner mutation = (%#v, %t), want miss", entry, ok)
	}
}

func TestBytecodeVM_StoreCachedGlobalValueSeedsProgramHotEntries(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 3)}
	want := runtime.NewSmallInt(31, runtime.IntegerI32)

	vm.storeCachedGlobalValue(program, 1, "x", want)

	if vm.globalLookupHotProgram != program || len(vm.globalLookupHotEntries) != len(program.instructions) {
		t.Fatalf("expected global lookup hot entries to cache program slice")
	}
	if got, ok := vm.lookupCachedGlobalValue(program, 1, "x"); !ok || got != want {
		t.Fatalf("lookupCachedGlobalValue() = (%#v, %t), want (%#v, true)", got, ok, want)
	}
}

func TestBytecodeVM_LookupIdentifierNameForCallCacheSkipsNewScopeCacheEntryOnDirectResolve(t *testing.T) {
	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	env := runtime.NewEnvironment(global)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	want := runtime.NewSmallInt(41, runtime.IntegerI32)
	env.Define("f", want)

	lookup, ok := vm.lookupIdentifierNameForCallCache(program, 1, "f")
	if !ok {
		t.Fatalf("lookupIdentifierNameForCallCache() miss, want hit")
	}
	if lookup.value != want || lookup.env != env || lookup.owner != env {
		t.Fatalf("lookupIdentifierNameForCallCache() = %#v, want value/env/owner %#v/%p/%p", lookup, want, env, env)
	}
	if vm.scopeLookupCache != nil {
		if entries := vm.scopeLookupCache[program]; len(entries) > 1 && entries[1].env != nil {
			t.Fatalf("expected direct call-cache resolve to skip new scope cache entry, got %#v", entries[1])
		}
	}
}

func TestBytecodeVM_LookupIdentifierNameForCallCacheSkipsNewGlobalCacheEntryOnDirectResolve(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	want := runtime.NewSmallInt(43, runtime.IntegerI32)
	env.Define("f", want)

	lookup, ok := vm.lookupIdentifierNameForCallCache(program, 1, "f")
	if !ok {
		t.Fatalf("lookupIdentifierNameForCallCache() miss, want hit")
	}
	if lookup.value != want || lookup.env != env || lookup.owner != env {
		t.Fatalf("lookupIdentifierNameForCallCache() = %#v, want value/env/owner %#v/%p/%p", lookup, want, env, env)
	}
	if vm.globalLookupCache != nil {
		if entries := vm.globalLookupCache[program]; len(entries) > 1 && entries[1].valid {
			t.Fatalf("expected direct call-cache resolve to skip new global cache entry, got %#v", entries[1])
		}
	}
}

func TestBytecodeVM_LoadNameStatsRecordHotHit(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	want := runtime.NewSmallInt(47, runtime.IntegerI32)
	version := vm.bytecodeEnvRevision(env)
	vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{
		valid:   true,
		program: program,
		ip:      1,
		entry: &bytecodeScopeLookupCacheEntry{
			env:          env,
			envVersion:   version,
			owner:        env,
			ownerVersion: version,
			value:        want,
		},
	}

	interp.ResetBytecodeStats()

	lookup, ok := vm.lookupCachedIdentifierNameEntry(program, 1, "x")
	if !ok || lookup.value != want {
		t.Fatalf("lookupCachedIdentifierNameEntry() = (%#v, %t), want (%#v, true)", lookup, ok, want)
	}

	stats := interp.BytecodeStats()
	if stats.LoadNameHotHits != 1 {
		t.Fatalf("expected one hot hit, got %d", stats.LoadNameHotHits)
	}
	if stats.LoadNameScopeHits != 0 || stats.LoadNameGlobalHits != 0 || stats.LoadNameDirectCurrent != 0 || stats.LoadNameDirectOuter != 0 || stats.LoadNameScopeStores != 0 || stats.LoadNameGlobalStores != 0 {
		t.Fatalf("unexpected non-hot load-name stats: %#v", stats)
	}
}

func TestBytecodeVM_LoadNameStatsRecordScopeHit(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	env := runtime.NewEnvironment(global)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	want := runtime.NewSmallInt(53, runtime.IntegerI32)
	env.Define("x", want)
	vm.storeCachedScopeValue(program, 1, "x", env, want)
	vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{}

	interp.ResetBytecodeStats()

	lookup, ok := vm.lookupCachedIdentifierNameEntry(program, 1, "x")
	if !ok || lookup.value != want {
		t.Fatalf("lookupCachedIdentifierNameEntry() = (%#v, %t), want (%#v, true)", lookup, ok, want)
	}

	stats := interp.BytecodeStats()
	if stats.LoadNameScopeHits != 1 {
		t.Fatalf("expected one scope cache hit, got %d", stats.LoadNameScopeHits)
	}
	if stats.LoadNameHotHits != 0 || stats.LoadNameGlobalHits != 0 || stats.LoadNameDirectCurrent != 0 || stats.LoadNameDirectOuter != 0 || stats.LoadNameScopeStores != 0 || stats.LoadNameGlobalStores != 0 {
		t.Fatalf("unexpected non-scope-hit load-name stats: %#v", stats)
	}
}

func TestBytecodeVM_LoadNameRefreshesSameEnvEntryAfterReassignment(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	env := runtime.NewEnvironment(global)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	first := runtime.NewSmallInt(54, runtime.IntegerI32)
	second := runtime.NewSmallInt(55, runtime.IntegerI32)
	env.Define("x", first)
	vm.storeCachedScopeValue(program, 1, "x", env, first)
	env.Define("x", second)
	vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{}

	interp.ResetBytecodeStats()

	lookup, ok := vm.lookupCachedIdentifierNameEntry(program, 1, "x")
	if !ok || lookup.value != second || lookup.owner != env {
		t.Fatalf("lookupCachedIdentifierNameEntry() = (%#v, %t), want owner=%p value=%#v", lookup, ok, env, second)
	}

	stats := interp.BytecodeStats()
	if stats.LoadNameScopeHits != 1 {
		t.Fatalf("expected same-env refresh to count as one scope hit, got %d", stats.LoadNameScopeHits)
	}
	if stats.LoadNameHotHits != 0 || stats.LoadNameGlobalHits != 0 || stats.LoadNameDirectCurrent != 0 || stats.LoadNameDirectOuter != 0 || stats.LoadNameScopeStores != 0 || stats.LoadNameGlobalStores != 0 {
		t.Fatalf("unexpected non-refresh load-name stats: %#v", stats)
	}
	entry := &vm.scopeLookupCache[program][1]
	if entry.value != second || entry.envVersion != vm.bytecodeEnvRevision(env) || entry.ownerVersion != vm.bytecodeEnvRevision(env) {
		t.Fatalf("refreshed entry = %#v, want value=%#v current env revision", entry, second)
	}
}

func TestBytecodeVM_LoadNameStatsRecordGlobalHit(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	want := runtime.NewSmallInt(59, runtime.IntegerI32)
	env.Define("x", want)
	vm.storeCachedGlobalValue(program, 1, "x", want)
	vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{}

	interp.ResetBytecodeStats()

	lookup, ok := vm.lookupCachedIdentifierNameEntry(program, 1, "x")
	if !ok || lookup.value != want {
		t.Fatalf("lookupCachedIdentifierNameEntry() = (%#v, %t), want (%#v, true)", lookup, ok, want)
	}

	stats := interp.BytecodeStats()
	if stats.LoadNameGlobalHits != 1 {
		t.Fatalf("expected one global cache hit, got %d", stats.LoadNameGlobalHits)
	}
	if stats.LoadNameHotHits != 0 || stats.LoadNameScopeHits != 0 || stats.LoadNameDirectCurrent != 0 || stats.LoadNameDirectOuter != 0 || stats.LoadNameScopeStores != 0 || stats.LoadNameGlobalStores != 0 {
		t.Fatalf("unexpected non-global-hit load-name stats: %#v", stats)
	}
}

func TestBytecodeVM_LoadNameStatsRecordScopeDirectCurrentResolveAndReset(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	env := runtime.NewEnvironment(global)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	want := runtime.NewSmallInt(61, runtime.IntegerI32)
	env.Define("x", want)

	interp.ResetBytecodeStats()

	lookup, ok := vm.lookupCachedIdentifierNameEntry(program, 1, "x")
	if !ok || lookup.value != want || lookup.owner != env {
		t.Fatalf("lookupCachedIdentifierNameEntry() = (%#v, %t), want owner=%p value=%#v", lookup, ok, env, want)
	}

	stats := interp.BytecodeStats()
	if stats.LoadNameDirectCurrent != 1 || stats.LoadNameScopeStores != 1 {
		t.Fatalf("expected one current direct resolve and one scope store, got %#v", stats)
	}
	if stats.LoadNameHotHits != 0 || stats.LoadNameScopeHits != 0 || stats.LoadNameGlobalHits != 0 || stats.LoadNameDirectOuter != 0 || stats.LoadNameGlobalStores != 0 {
		t.Fatalf("unexpected extra load-name stats: %#v", stats)
	}

	interp.ResetBytecodeStats()
	assertBytecodeLoadNamePathStatsZero(t, interp.BytecodeStats())
}

func TestBytecodeVM_LoadNameStatsRecordScopeDirectOuterResolve(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	env := runtime.NewEnvironment(global)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	want := runtime.NewSmallInt(67, runtime.IntegerI32)
	global.Define("x", want)

	interp.ResetBytecodeStats()

	lookup, ok := vm.lookupCachedIdentifierNameEntry(program, 1, "x")
	if !ok || lookup.value != want || lookup.owner != global {
		t.Fatalf("lookupCachedIdentifierNameEntry() = (%#v, %t), want owner=%p value=%#v", lookup, ok, global, want)
	}

	stats := interp.BytecodeStats()
	if stats.LoadNameDirectOuter != 1 || stats.LoadNameScopeStores != 1 {
		t.Fatalf("expected one outer direct resolve and one scope store, got %#v", stats)
	}
	if stats.LoadNameHotHits != 0 || stats.LoadNameScopeHits != 0 || stats.LoadNameGlobalHits != 0 || stats.LoadNameDirectCurrent != 0 || stats.LoadNameGlobalStores != 0 {
		t.Fatalf("unexpected extra load-name stats: %#v", stats)
	}
}

func TestBytecodeVM_LoadNameStatsRecordGlobalDirectResolve(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	want := runtime.NewSmallInt(71, runtime.IntegerI32)
	env.Define("x", want)

	interp.ResetBytecodeStats()

	lookup, ok := vm.lookupCachedIdentifierNameEntry(program, 1, "x")
	if !ok || lookup.value != want || lookup.owner != env {
		t.Fatalf("lookupCachedIdentifierNameEntry() = (%#v, %t), want owner=%p value=%#v", lookup, ok, env, want)
	}

	stats := interp.BytecodeStats()
	if stats.LoadNameDirectCurrent != 1 || stats.LoadNameGlobalStores != 1 {
		t.Fatalf("expected one current direct resolve and one global store, got %#v", stats)
	}
	if stats.LoadNameHotHits != 0 || stats.LoadNameScopeHits != 0 || stats.LoadNameGlobalHits != 0 || stats.LoadNameDirectOuter != 0 || stats.LoadNameScopeStores != 0 {
		t.Fatalf("unexpected extra load-name stats: %#v", stats)
	}
}

func TestBytecodeVM_CallLookupSkipsLoadNamePathStats(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	env := runtime.NewEnvironment(global)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	want := runtime.NewSmallInt(73, runtime.IntegerI32)
	env.Define("f", want)

	interp.ResetBytecodeStats()

	lookup, ok := vm.lookupIdentifierNameForCallCache(program, 1, "f")
	if !ok || lookup.value != want {
		t.Fatalf("lookupIdentifierNameForCallCache() = (%#v, %t), want (%#v, true)", lookup, ok, want)
	}

	assertBytecodeLoadNamePathStatsZero(t, interp.BytecodeStats())
}
