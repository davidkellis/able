package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

var (
	benchmarkBytecodeLookupValueSink   runtime.Value
	benchmarkBytecodeLookupEntrySink   *bytecodeScopeLookupCacheEntry
	benchmarkBytecodeLookupOwnerSink   *runtime.Environment
	benchmarkBytecodeLookupVersionSink uint64
	benchmarkBytecodeLookupOKSink      bool
)

func benchmarkBytecodeLookupProgram(ip int) *bytecodeProgram {
	return &bytecodeProgram{instructions: make([]bytecodeInstruction, ip+1)}
}

func benchmarkClearBytecodeScopeLookupSite(vm *bytecodeVM, program *bytecodeProgram, ip int) {
	vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{}
	entries, ok := vm.activeScopeLookupCacheEntries(program, false)
	if !ok || ip < 0 || ip >= len(entries) {
		panic("missing scope lookup cache entry")
	}
	entries[ip] = bytecodeScopeLookupCacheEntry{}
}

func benchmarkClearBytecodeGlobalLookupSite(vm *bytecodeVM, program *bytecodeProgram, ip int) {
	vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{}
	entries, ok := vm.activeGlobalLookupCacheEntries(program, false)
	if !ok || ip < 0 || ip >= len(entries) {
		panic("missing global lookup cache entry")
	}
	entries[ip] = bytecodeGlobalLookupCacheEntry{}
}

func BenchmarkBytecodeVMLookupCachedIdentifierName(b *testing.B) {
	b.Run("hot_current", func(b *testing.B) {
		interp := NewBytecode()
		env := runtime.NewEnvironment(interp.GlobalEnvironment())
		want := runtime.NewSmallInt(7, runtime.IntegerI32)
		env.DefineWithoutMerge("x", want)

		vm := newBytecodeVM(interp, env)
		program := benchmarkBytecodeLookupProgram(1)
		vm.prepareRunProgram(program, false)

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
		if !ok || got != want {
			b.Fatalf("lookupCachedIdentifierName() = (%#v, %t), want (%#v, true)", got, ok, want)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeLookupValueSink, benchmarkBytecodeLookupOKSink = vm.lookupCachedIdentifierName(program, 1, "x")
		}
	})

	b.Run("hot_outer", func(b *testing.B) {
		interp := NewBytecode()
		global := interp.GlobalEnvironment()
		env := runtime.NewEnvironment(global)
		want := runtime.NewSmallInt(11, runtime.IntegerI32)
		global.DefineWithoutMerge("x", want)

		vm := newBytecodeVM(interp, env)
		program := benchmarkBytecodeLookupProgram(1)
		vm.prepareRunProgram(program, false)

		currentVersion := vm.bytecodeEnvRevision(env)
		ownerVersion := vm.bytecodeEnvRevision(global)
		entry := &bytecodeScopeLookupCacheEntry{
			env:          env,
			envVersion:   currentVersion,
			owner:        global,
			ownerVersion: ownerVersion,
			value:        want,
		}
		vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{
			valid:   true,
			program: program,
			ip:      1,
			entry:   entry,
		}

		got, ok := vm.lookupCachedIdentifierName(program, 1, "x")
		if !ok || got != want {
			b.Fatalf("lookupCachedIdentifierName() = (%#v, %t), want (%#v, true)", got, ok, want)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeLookupValueSink, benchmarkBytecodeLookupOKSink = vm.lookupCachedIdentifierName(program, 1, "x")
		}
	})

	b.Run("direct_current", func(b *testing.B) {
		interp := NewBytecode()
		global := interp.GlobalEnvironment()
		env := runtime.NewEnvironment(global)
		want := runtime.NewSmallInt(13, runtime.IntegerI32)
		env.DefineWithoutMerge("x", want)

		vm := newBytecodeVM(interp, env)
		program := benchmarkBytecodeLookupProgram(1)
		vm.prepareRunProgram(program, false)

		got, ok := vm.lookupCachedIdentifierName(program, 1, "x")
		if !ok || got != want {
			b.Fatalf("lookupCachedIdentifierName() = (%#v, %t), want (%#v, true)", got, ok, want)
		}
		benchmarkClearBytecodeScopeLookupSite(vm, program, 1)
		scopeEntries, ok := vm.activeScopeLookupCacheEntries(program, false)
		if !ok {
			b.Fatalf("missing active scope lookup entries")
		}

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{}
			scopeEntries[1] = bytecodeScopeLookupCacheEntry{}
			benchmarkBytecodeLookupValueSink, benchmarkBytecodeLookupOKSink = vm.lookupCachedIdentifierName(program, 1, "x")
		}
	})

	b.Run("direct_outer", func(b *testing.B) {
		interp := NewBytecode()
		global := interp.GlobalEnvironment()
		env := runtime.NewEnvironment(global)
		want := runtime.NewSmallInt(17, runtime.IntegerI32)
		global.DefineWithoutMerge("x", want)

		vm := newBytecodeVM(interp, env)
		program := benchmarkBytecodeLookupProgram(1)
		vm.prepareRunProgram(program, false)

		got, ok := vm.lookupCachedIdentifierName(program, 1, "x")
		if !ok || got != want {
			b.Fatalf("lookupCachedIdentifierName() = (%#v, %t), want (%#v, true)", got, ok, want)
		}
		benchmarkClearBytecodeScopeLookupSite(vm, program, 1)
		scopeEntries, ok := vm.activeScopeLookupCacheEntries(program, false)
		if !ok {
			b.Fatalf("missing active scope lookup entries")
		}

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{}
			scopeEntries[1] = bytecodeScopeLookupCacheEntry{}
			benchmarkBytecodeLookupValueSink, benchmarkBytecodeLookupOKSink = vm.lookupCachedIdentifierName(program, 1, "x")
		}
	})

	b.Run("global_direct", func(b *testing.B) {
		interp := NewBytecode()
		global := interp.GlobalEnvironment()
		want := runtime.NewSmallInt(19, runtime.IntegerI32)
		global.DefineWithoutMerge("x", want)

		vm := newBytecodeVM(interp, global)
		program := benchmarkBytecodeLookupProgram(1)
		vm.prepareRunProgram(program, false)

		got, ok := vm.lookupCachedIdentifierName(program, 1, "x")
		if !ok || got != want {
			b.Fatalf("lookupCachedIdentifierName() = (%#v, %t), want (%#v, true)", got, ok, want)
		}
		benchmarkClearBytecodeGlobalLookupSite(vm, program, 1)
		globalEntries, ok := vm.activeGlobalLookupCacheEntries(program, false)
		if !ok {
			b.Fatalf("missing active global lookup entries")
		}

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{}
			globalEntries[1] = bytecodeGlobalLookupCacheEntry{}
			benchmarkBytecodeLookupValueSink, benchmarkBytecodeLookupOKSink = vm.lookupCachedIdentifierName(program, 1, "x")
		}
	})
}

func BenchmarkBytecodeVMLookupCachedScopeEntryWithVersion(b *testing.B) {
	b.Run("current_owner", func(b *testing.B) {
		interp := NewBytecode()
		global := interp.GlobalEnvironment()
		env := runtime.NewEnvironment(global)
		want := runtime.NewSmallInt(23, runtime.IntegerI32)
		env.DefineWithoutMerge("x", want)

		vm := newBytecodeVM(interp, env)
		program := benchmarkBytecodeLookupProgram(1)
		vm.prepareRunProgram(program, false)
		vm.storeCachedScopeValue(program, 1, "x", env, want)

		singleThread := vm.bytecodeSingleThread()
		currentVersion := env.RevisionWithHint(singleThread)
		entry, ok := vm.lookupCachedScopeEntryWithVersion(program, 1, "x", env, currentVersion, singleThread)
		if !ok || entry == nil || entry.value != want {
			b.Fatalf("lookupCachedScopeEntryWithVersion() = (%#v, %t), want value %#v", entry, ok, want)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeLookupEntrySink, benchmarkBytecodeLookupOKSink =
				vm.lookupCachedScopeEntryWithVersion(program, 1, "x", env, currentVersion, singleThread)
		}
	})

	b.Run("outer_owner", func(b *testing.B) {
		interp := NewBytecode()
		global := interp.GlobalEnvironment()
		env := runtime.NewEnvironment(global)
		want := runtime.NewSmallInt(29, runtime.IntegerI32)
		global.DefineWithoutMerge("x", want)

		vm := newBytecodeVM(interp, env)
		program := benchmarkBytecodeLookupProgram(1)
		vm.prepareRunProgram(program, false)
		vm.storeCachedScopeValue(program, 1, "x", global, want)

		singleThread := vm.bytecodeSingleThread()
		currentVersion := env.RevisionWithHint(singleThread)
		entry, ok := vm.lookupCachedScopeEntryWithVersion(program, 1, "x", env, currentVersion, singleThread)
		if !ok || entry == nil || entry.value != want {
			b.Fatalf("lookupCachedScopeEntryWithVersion() = (%#v, %t), want value %#v", entry, ok, want)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeLookupEntrySink, benchmarkBytecodeLookupOKSink =
				vm.lookupCachedScopeEntryWithVersion(program, 1, "x", env, currentVersion, singleThread)
		}
	})
}

func BenchmarkBytecodeVMLookupCachedGlobalValue(b *testing.B) {
	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	want := runtime.NewSmallInt(31, runtime.IntegerI32)
	global.DefineWithoutMerge("x", want)

	vm := newBytecodeVM(interp, global)
	program := benchmarkBytecodeLookupProgram(1)
	vm.prepareRunProgram(program, false)
	vm.storeCachedGlobalValue(program, 1, "x", want)

	got, ok := vm.lookupCachedGlobalValue(program, 1, "x")
	if !ok || got != want {
		b.Fatalf("lookupCachedGlobalValue() = (%#v, %t), want (%#v, true)", got, ok, want)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for idx := 0; idx < b.N; idx++ {
		benchmarkBytecodeLookupValueSink, benchmarkBytecodeLookupOKSink = vm.lookupCachedGlobalValue(program, 1, "x")
	}
}

func BenchmarkBytecodeVMStoreCachedScopeEntryWithVersions(b *testing.B) {
	b.Run("current_owner", func(b *testing.B) {
		interp := NewBytecode()
		global := interp.GlobalEnvironment()
		env := runtime.NewEnvironment(global)
		var want runtime.Value = runtime.NewSmallInt(37, runtime.IntegerI32)

		vm := newBytecodeVM(interp, env)
		program := benchmarkBytecodeLookupProgram(1)
		vm.prepareRunProgram(program, false)
		vm.storeCachedScopeValue(program, 1, "x", env, want)

		currentVersion := vm.bytecodeEnvRevision(env)
		entry := vm.storeCachedScopeEntryWithVersions(program, 1, "x", env, currentVersion, env, currentVersion, want)
		if entry == nil || entry.value != want || entry.env != env || entry.owner != env {
			b.Fatalf("storeCachedScopeEntryWithVersions() = %#v, want env/owner=%p/%p value=%#v", entry, env, env, want)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			entry := vm.storeCachedScopeEntryWithVersions(program, 1, "x", env, currentVersion, env, currentVersion, want)
			if entry == nil {
				b.Fatalf("storeCachedScopeEntryWithVersions() returned nil")
			}
			benchmarkBytecodeLookupValueSink = entry.value
		}
	})

	b.Run("outer_owner", func(b *testing.B) {
		interp := NewBytecode()
		global := interp.GlobalEnvironment()
		env := runtime.NewEnvironment(global)
		var want runtime.Value = runtime.NewSmallInt(41, runtime.IntegerI32)

		vm := newBytecodeVM(interp, env)
		program := benchmarkBytecodeLookupProgram(1)
		vm.prepareRunProgram(program, false)
		vm.storeCachedScopeValue(program, 1, "x", global, want)

		currentVersion := vm.bytecodeEnvRevision(env)
		ownerVersion := vm.bytecodeEnvRevision(global)
		entry := vm.storeCachedScopeEntryWithVersions(program, 1, "x", env, currentVersion, global, ownerVersion, want)
		if entry == nil || entry.value != want || entry.env != env || entry.owner != global {
			b.Fatalf("storeCachedScopeEntryWithVersions() = %#v, want env/owner=%p/%p value=%#v", entry, env, global, want)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			entry := vm.storeCachedScopeEntryWithVersions(program, 1, "x", env, currentVersion, global, ownerVersion, want)
			if entry == nil {
				b.Fatalf("storeCachedScopeEntryWithVersions() returned nil")
			}
			benchmarkBytecodeLookupValueSink = entry.value
		}
	})
}

func BenchmarkBytecodeVMStoreHotGlobalNameEntryWithVersions(b *testing.B) {
	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	var want runtime.Value = runtime.NewSmallInt(43, runtime.IntegerI32)

	vm := newBytecodeVM(interp, global)
	program := benchmarkBytecodeLookupProgram(1)
	vm.prepareRunProgram(program, false)

	currentVersion := vm.bytecodeEnvRevision(global)
	entry := vm.storeHotGlobalNameEntryWithVersions(program, 1, "x", global, currentVersion, global, currentVersion, want)
	if entry == nil || entry.value != want || entry.env != global || entry.owner != global {
		b.Fatalf("storeHotGlobalNameEntryWithVersions() = %#v, want env/owner=%p/%p value=%#v", entry, global, global, want)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for idx := 0; idx < b.N; idx++ {
		entry := vm.storeHotGlobalNameEntryWithVersions(program, 1, "x", global, currentVersion, global, currentVersion, want)
		if entry == nil {
			b.Fatalf("storeHotGlobalNameEntryWithVersions() returned nil")
		}
		benchmarkBytecodeLookupValueSink = entry.value
	}
}

func BenchmarkBytecodeVMDirectLookupMissComponents(b *testing.B) {
	b.Run("current_revision", func(b *testing.B) {
		interp := NewBytecode()
		global := interp.GlobalEnvironment()
		env := runtime.NewEnvironment(global)
		vm := newBytecodeVM(interp, env)
		singleThread := vm.bytecodeSingleThread()

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeLookupVersionSink = env.RevisionWithHint(singleThread)
		}
	})

	b.Run("hot_miss_empty", func(b *testing.B) {
		interp := NewBytecode()
		global := interp.GlobalEnvironment()
		env := runtime.NewEnvironment(global)
		vm := newBytecodeVM(interp, env)
		program := benchmarkBytecodeLookupProgram(1)
		vm.prepareRunProgram(program, false)

		singleThread := vm.bytecodeSingleThread()
		currentVersion := env.RevisionWithHint(singleThread)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeLookupEntrySink, benchmarkBytecodeLookupOKSink =
				vm.lookupHotNameEntryWithVersion(program, 1, "x", env, currentVersion, singleThread)
		}
	})

	b.Run("scope_miss_empty_entry", func(b *testing.B) {
		interp := NewBytecode()
		global := interp.GlobalEnvironment()
		env := runtime.NewEnvironment(global)
		var want runtime.Value = runtime.NewSmallInt(47, runtime.IntegerI32)

		vm := newBytecodeVM(interp, env)
		program := benchmarkBytecodeLookupProgram(1)
		vm.prepareRunProgram(program, false)
		vm.storeCachedScopeEntryWithVersions(program, 1, "x", env, vm.bytecodeEnvRevision(env), env, vm.bytecodeEnvRevision(env), want)
		benchmarkClearBytecodeScopeLookupSite(vm, program, 1)

		singleThread := vm.bytecodeSingleThread()
		currentVersion := env.RevisionWithHint(singleThread)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeLookupEntrySink, benchmarkBytecodeLookupOKSink =
				vm.lookupCachedScopeEntryWithVersion(program, 1, "x", env, currentVersion, singleThread)
		}
	})

	b.Run("global_miss_empty_entry", func(b *testing.B) {
		interp := NewBytecode()
		global := interp.GlobalEnvironment()
		var want runtime.Value = runtime.NewSmallInt(53, runtime.IntegerI32)

		vm := newBytecodeVM(interp, global)
		program := benchmarkBytecodeLookupProgram(1)
		vm.prepareRunProgram(program, false)
		vm.storeCachedGlobalValue(program, 1, "x", want)
		benchmarkClearBytecodeGlobalLookupSite(vm, program, 1)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeLookupValueSink, benchmarkBytecodeLookupOKSink = vm.lookupCachedGlobalValue(program, 1, "x")
		}
	})

	b.Run("runtime_lookup_current", func(b *testing.B) {
		interp := NewBytecode()
		global := interp.GlobalEnvironment()
		env := runtime.NewEnvironment(global)
		var want runtime.Value = runtime.NewSmallInt(59, runtime.IntegerI32)
		env.DefineWithoutMerge("x", want)
		singleThread := interp.envSingleThread

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeLookupValueSink, benchmarkBytecodeLookupOwnerSink, benchmarkBytecodeLookupVersionSink, benchmarkBytecodeLookupOKSink =
				env.LookupWithOwnerAndRevisionHint("x", singleThread)
		}
	})

	b.Run("runtime_lookup_outer", func(b *testing.B) {
		interp := NewBytecode()
		global := interp.GlobalEnvironment()
		env := runtime.NewEnvironment(global)
		var want runtime.Value = runtime.NewSmallInt(61, runtime.IntegerI32)
		global.DefineWithoutMerge("x", want)
		singleThread := interp.envSingleThread

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeLookupValueSink, benchmarkBytecodeLookupOwnerSink, benchmarkBytecodeLookupVersionSink, benchmarkBytecodeLookupOKSink =
				env.LookupWithOwnerAndRevisionHint("x", singleThread)
		}
	})

	b.Run("runtime_current_plus_scope_store", func(b *testing.B) {
		interp := NewBytecode()
		global := interp.GlobalEnvironment()
		env := runtime.NewEnvironment(global)
		var want runtime.Value = runtime.NewSmallInt(67, runtime.IntegerI32)
		env.DefineWithoutMerge("x", want)

		vm := newBytecodeVM(interp, env)
		program := benchmarkBytecodeLookupProgram(1)
		vm.prepareRunProgram(program, false)

		singleThread := vm.bytecodeSingleThread()
		currentVersion := env.RevisionWithHint(singleThread)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			value, owner, ownerVersion, ok := env.LookupWithOwnerAndRevisionHint("x", singleThread)
			if !ok {
				b.Fatalf("LookupWithOwnerAndRevisionHint returned miss")
			}
			entry := vm.storeCachedScopeEntryWithVersions(program, 1, "x", env, currentVersion, owner, ownerVersion, value)
			if entry == nil {
				b.Fatalf("storeCachedScopeEntryWithVersions returned nil")
			}
			benchmarkBytecodeLookupValueSink = entry.value
		}
	})

	b.Run("runtime_outer_plus_scope_store", func(b *testing.B) {
		interp := NewBytecode()
		global := interp.GlobalEnvironment()
		env := runtime.NewEnvironment(global)
		var want runtime.Value = runtime.NewSmallInt(71, runtime.IntegerI32)
		global.DefineWithoutMerge("x", want)

		vm := newBytecodeVM(interp, env)
		program := benchmarkBytecodeLookupProgram(1)
		vm.prepareRunProgram(program, false)

		singleThread := vm.bytecodeSingleThread()
		currentVersion := env.RevisionWithHint(singleThread)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			value, owner, ownerVersion, ok := env.LookupWithOwnerAndRevisionHint("x", singleThread)
			if !ok {
				b.Fatalf("LookupWithOwnerAndRevisionHint returned miss")
			}
			entry := vm.storeCachedScopeEntryWithVersions(program, 1, "x", env, currentVersion, owner, ownerVersion, value)
			if entry == nil {
				b.Fatalf("storeCachedScopeEntryWithVersions returned nil")
			}
			benchmarkBytecodeLookupValueSink = entry.value
		}
	})
}
