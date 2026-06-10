package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LookupHotCanonicalArrayIndexSiteUsesReceiverIdentityCache(t *testing.T) {
	interp := NewBytecode()
	preloadArrayStdlibForTest(t, interp)
	innerChar := monoCharArrayValueForTest(t, 'a', 'b')
	outer := interp.newArrayValue([]runtime.Value{innerChar}, 1)
	alias := &runtime.ArrayValue{Handle: outer.Handle, TrackedHandle: outer.Handle}
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{instructions: []bytecodeInstruction{{}}}
	vm.currentProgram = program
	vm.ip = 0

	method, fastPath, hasMethod, cacheable, err := vm.resolveCachedIndexMethod(program, 0, outer, "get", "Index")
	if err != nil {
		t.Fatalf("resolveCachedIndexMethod(get): %v", err)
	}
	if !cacheable || !hasMethod || method == nil {
		t.Fatalf("expected canonical Array Index.get method to be cached")
	}
	if fastPath != bytecodeIndexMethodFastPathCanonicalArrayGet {
		t.Fatalf("fast path = %v, want canonical array get", fastPath)
	}

	hotArr, ok := vm.lookupHotCanonicalArrayIndexSite("get", alias, bytecodeIndexMethodFastPathCanonicalArrayGet)
	if !ok {
		t.Fatalf("expected hot canonical array index site lookup to succeed")
	}
	if hotArr != alias {
		t.Fatalf("hot array = %p, want alias receiver %p", hotArr, alias)
	}
	if alias.State != nil || alias.Elements != nil {
		t.Fatalf("hot canonical site lookup should not materialize alias state")
	}

	innerU32 := monoU32ArrayValueForTest(t, 7, 11)
	if err := runtime.ArrayStoreWrite(outer.Handle, 0, innerU32); err != nil {
		t.Fatalf("ArrayStoreWrite update outer[0]: %v", err)
	}
	if hotArr, ok := vm.lookupHotCanonicalArrayIndexSite("get", alias, bytecodeIndexMethodFastPathCanonicalArrayGet); ok || hotArr != nil {
		t.Fatalf("expected stale hot canonical site lookup to miss after receiver identity change, got (%p, %v)", hotArr, ok)
	}
	if alias.State != nil || alias.Elements != nil {
		t.Fatalf("stale hot canonical site lookup should still avoid alias materialization")
	}
}

func TestBytecodeVM_LookupHotCanonicalArrayIndexSiteFallsBackToIdentityForDifferentHandle(t *testing.T) {
	interp := NewBytecode()
	preloadArrayStdlibForTest(t, interp)
	left := monoCharArrayValueForTest(t, 'a', 'b')
	right := monoCharArrayValueForTest(t, 'x', 'y')
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{instructions: []bytecodeInstruction{{}}}
	vm.currentProgram = program
	vm.ip = 0

	method, fastPath, hasMethod, cacheable, err := vm.resolveCachedIndexMethod(program, 0, left, "get", "Index")
	if err != nil {
		t.Fatalf("resolveCachedIndexMethod(get): %v", err)
	}
	if !cacheable || !hasMethod || method == nil {
		t.Fatalf("expected canonical Array Index.get method to be cached")
	}
	if fastPath != bytecodeIndexMethodFastPathCanonicalArrayGet {
		t.Fatalf("fast path = %v, want canonical array get", fastPath)
	}

	hotArr, ok := vm.lookupHotCanonicalArrayIndexSite("get", right, bytecodeIndexMethodFastPathCanonicalArrayGet)
	if !ok {
		t.Fatalf("expected different-handle same-identity hot canonical lookup to succeed")
	}
	if hotArr != right {
		t.Fatalf("hot array = %p, want right receiver %p", hotArr, right)
	}
	if right.State != nil || right.Elements != nil {
		t.Fatalf("different-handle hot canonical site lookup should not materialize boxed state")
	}
}

func TestBytecodeVM_LookupHotCanonicalArrayIndexSiteUsesSameReceiverRevision(t *testing.T) {
	interp := NewBytecode()
	preloadArrayStdlibForTest(t, interp)
	innerChar := monoCharArrayValueForTest(t, 'a', 'b')
	outer := interp.newArrayValue([]runtime.Value{innerChar}, 1)
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{instructions: []bytecodeInstruction{{}}}
	vm.currentProgram = program
	vm.ip = 0

	method, fastPath, hasMethod, cacheable, err := vm.resolveCachedIndexMethod(program, 0, outer, "get", "Index")
	if err != nil {
		t.Fatalf("resolveCachedIndexMethod(get): %v", err)
	}
	if !cacheable || !hasMethod || method == nil {
		t.Fatalf("expected canonical Array Index.get method to be cached")
	}
	if fastPath != bytecodeIndexMethodFastPathCanonicalArrayGet {
		t.Fatalf("fast path = %v, want canonical array get", fastPath)
	}
	revision, ok, err := runtime.ArrayStoreRevisionIfAvailable(outer.Handle)
	if err != nil || !ok {
		t.Fatalf("ArrayStoreRevisionIfAvailable: revision=%d ok=%v err=%v", revision, ok, err)
	}
	vm.arrayIndexReceiverIdentityCache[outer.Handle] = bytecodeArrayIndexReceiverIdentityCacheEntry{
		revision: revision,
		elemType: bytecodeIndexTypeUnknown,
		key:      "WrongType",
		ok:       true,
	}

	hotArr, ok := vm.lookupHotCanonicalArrayIndexSite("get", outer, bytecodeIndexMethodFastPathCanonicalArrayGet)
	if !ok {
		t.Fatalf("expected same-receiver revision shortcut to bypass stale identity cache")
	}
	if hotArr != outer {
		t.Fatalf("hot array = %p, want outer receiver %p", hotArr, outer)
	}
}

func TestBytecodeVM_LookupDirectCompatibleHotArrayIndexSiteAcceptsCanonicalAndNoMethodOnly(t *testing.T) {
	canonicalInterp := NewBytecode()
	preloadArrayStdlibForTest(t, canonicalInterp)
	canonicalArr := monoCharArrayValueForTest(t, 'a', 'b')
	canonicalVM := newBytecodeVM(canonicalInterp, canonicalInterp.GlobalEnvironment())
	program := &bytecodeProgram{instructions: []bytecodeInstruction{{}}}
	canonicalVM.currentProgram = program
	canonicalVM.ip = 0

	if method, fastPath, hasMethod, cacheable, err := canonicalVM.resolveCachedIndexMethod(program, 0, canonicalArr, "get", "Index"); err != nil {
		t.Fatalf("resolveCachedIndexMethod(canonical get): %v", err)
	} else if !cacheable || !hasMethod || method == nil || fastPath != bytecodeIndexMethodFastPathCanonicalArrayGet {
		t.Fatalf("canonical cache = method %#v fastPath %v hasMethod %v cacheable %v, want canonical get", method, fastPath, hasMethod, cacheable)
	}
	globalRevision, methodCacheVersion := canonicalVM.bytecodeGlobalAndMethodVersions()
	if hotArr, handle, ok := canonicalVM.lookupDirectCompatibleHotArrayIndexSiteWithVersions(
		bytecodeIndexMethodCacheGet,
		canonicalArr,
		bytecodeIndexMethodFastPathCanonicalArrayGet,
		globalRevision,
		methodCacheVersion,
	); !ok || hotArr != canonicalArr {
		t.Fatalf("canonical direct-compatible lookup = (%p, %v), want canonical receiver", hotArr, ok)
	} else if handle != canonicalArr.Handle {
		t.Fatalf("canonical direct-compatible handle = %d, want %d", handle, canonicalArr.Handle)
	}

	noMethodInterp := NewBytecode()
	noMethodArr := monoCharArrayValueForTest(t, 'x', 'y')
	noMethodVM := newBytecodeVM(noMethodInterp, noMethodInterp.GlobalEnvironment())
	noMethodVM.currentProgram = program
	noMethodVM.ip = 0
	if method, fastPath, hasMethod, cacheable, err := noMethodVM.resolveCachedIndexMethod(program, 0, noMethodArr, "get", "Index"); err != nil {
		t.Fatalf("resolveCachedIndexMethod(no-method get): %v", err)
	} else if !cacheable || hasMethod || method != nil || fastPath != bytecodeIndexMethodFastPathNone {
		t.Fatalf("no-method cache = method %#v fastPath %v hasMethod %v cacheable %v, want no method", method, fastPath, hasMethod, cacheable)
	}
	globalRevision, methodCacheVersion = noMethodVM.bytecodeGlobalAndMethodVersions()
	if hotArr, handle, ok := noMethodVM.lookupDirectCompatibleHotArrayIndexSiteWithVersions(
		bytecodeIndexMethodCacheGet,
		noMethodArr,
		bytecodeIndexMethodFastPathCanonicalArrayGet,
		globalRevision,
		methodCacheVersion,
	); !ok || hotArr != noMethodArr {
		t.Fatalf("no-method direct-compatible lookup = (%p, %v), want no-method receiver", hotArr, ok)
	} else if handle != noMethodArr.Handle {
		t.Fatalf("no-method direct-compatible handle = %d, want %d", handle, noMethodArr.Handle)
	}

	elemType, typeKey, identityOK := noMethodVM.arrayIndexReceiverIdentity(noMethodArr)
	if !identityOK {
		t.Fatalf("expected no-method receiver identity")
	}
	handle, revision, revisionOK := noMethodVM.indexMethodReceiverRevision(noMethodArr)
	userProgram := &bytecodeProgram{instructions: []bytecodeInstruction{{}}}
	noMethodVM.currentProgram = userProgram
	noMethodVM.ip = 0
	noMethodVM.indexMethodHot = bytecodeInlineIndexMethodCacheEntry{
		valid:               true,
		program:             userProgram,
		ip:                  0,
		methodKind:          bytecodeIndexMethodCacheGet,
		globalRevision:      globalRevision,
		receiverKind:        bytecodeMemberReceiverArray,
		arrayElemType:       elemType,
		receiverTypeKey:     typeKey,
		receiverArrayHandle: handle,
		receiverArrayRev:    revision,
		receiverArrayRevOK:  revisionOK,
		methodCacheVersion:  methodCacheVersion,
		hasMethod:           true,
		fastPath:            bytecodeIndexMethodFastPathNone,
	}
	if hotArr, _, ok := noMethodVM.lookupDirectCompatibleHotArrayIndexSiteWithVersions(
		bytecodeIndexMethodCacheGet,
		noMethodArr,
		bytecodeIndexMethodFastPathCanonicalArrayGet,
		globalRevision,
		methodCacheVersion,
	); ok || hotArr != nil {
		t.Fatalf("non-canonical method lookup = (%p, %v), want miss", hotArr, ok)
	}
}

func TestBytecodeVM_DirectCompatibleHotArrayIndexGetCarriesValidatedHandle(t *testing.T) {
	interp := NewBytecode()
	preloadArrayStdlibForTest(t, interp)
	arr := monoCharArrayValueForTest(t, 'a', 'b', 'c')
	alias := &runtime.ArrayValue{TrackedHandle: arr.Handle}
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{instructions: []bytecodeInstruction{{}}}
	vm.currentProgram = program
	vm.ip = 0

	if method, fastPath, hasMethod, cacheable, err := vm.resolveCachedIndexMethod(program, 0, arr, "get", "Index"); err != nil {
		t.Fatalf("resolveCachedIndexMethod(get): %v", err)
	} else if !cacheable || !hasMethod || method == nil || fastPath != bytecodeIndexMethodFastPathCanonicalArrayGet {
		t.Fatalf("cache = method %#v fastPath %v hasMethod %v cacheable %v, want canonical get", method, fastPath, hasMethod, cacheable)
	}

	globalRevision, methodCacheVersion := vm.bytecodeGlobalAndMethodVersions()
	hotArr, handle, ok := vm.lookupDirectCompatibleHotArrayIndexSiteWithVersions(
		bytecodeIndexMethodCacheGet,
		alias,
		bytecodeIndexMethodFastPathCanonicalArrayGet,
		globalRevision,
		methodCacheVersion,
	)
	if !ok {
		t.Fatalf("expected direct-compatible lookup for alias receiver")
	}
	if hotArr != alias {
		t.Fatalf("hot array = %p, want alias receiver %p", hotArr, alias)
	}
	if handle != arr.Handle {
		t.Fatalf("validated handle = %d, want %d", handle, arr.Handle)
	}
	if alias.State != nil || alias.Elements != nil {
		t.Fatalf("validated handle lookup should not materialize alias state")
	}

	value, handled, err := vm.resolveDirectArrayIndexGetWithHandle(alias, runtime.NewSmallInt(2, runtime.IntegerI32), handle)
	if err != nil || !handled {
		t.Fatalf("resolveDirectArrayIndexGetWithHandle handled=%v err=%v", handled, err)
	}
	char, ok := value.(runtime.CharValue)
	if !ok || char.Val != 'c' {
		t.Fatalf("direct get with validated handle = %#v, want char c", value)
	}
	if alias.State != nil || alias.Elements != nil {
		t.Fatalf("direct get with validated handle should not materialize alias state")
	}
}

func TestBytecodeVM_LookupHotCanonicalArrayIndexSiteUsesDirectPerIPCache(t *testing.T) {
	interp := NewBytecode()
	preloadArrayStdlibForTest(t, interp)
	left := interp.newArrayValue([]runtime.Value{monoCharArrayValueForTest(t, 'a', 'b')}, 1)
	right := interp.newArrayValue([]runtime.Value{monoCharArrayValueForTest(t, 'x', 'y')}, 1)
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{instructions: []bytecodeInstruction{{}, {}}}
	vm.currentProgram = program

	vm.ip = 0
	if _, _, hasMethod, cacheable, err := vm.resolveCachedIndexMethod(program, 0, left, "get", "Index"); err != nil {
		t.Fatalf("resolveCachedIndexMethod(left): %v", err)
	} else if !cacheable || !hasMethod {
		t.Fatalf("expected left index method to be cacheable")
	}
	leftRevision, ok, err := runtime.ArrayStoreRevisionIfAvailable(left.Handle)
	if err != nil || !ok {
		t.Fatalf("ArrayStoreRevisionIfAvailable(left): revision=%d ok=%v err=%v", leftRevision, ok, err)
	}
	vm.arrayIndexReceiverIdentityCache[left.Handle] = bytecodeArrayIndexReceiverIdentityCacheEntry{
		revision: leftRevision,
		elemType: bytecodeIndexTypeUnknown,
		key:      "WrongType",
		ok:       true,
	}

	vm.ip = 1
	if _, _, hasMethod, cacheable, err := vm.resolveCachedIndexMethod(program, 1, right, "get", "Index"); err != nil {
		t.Fatalf("resolveCachedIndexMethod(right): %v", err)
	} else if !cacheable || !hasMethod {
		t.Fatalf("expected right index method to be cacheable")
	}

	vm.currentProgram = program
	vm.activeLookup.program = program
	vm.activeLookup.indexMethodGetEntries = nil
	vm.ip = 0
	hotArr, ok := vm.lookupHotCanonicalArrayIndexSite("get", left, bytecodeIndexMethodFastPathCanonicalArrayGet)
	if !ok {
		t.Fatalf("expected direct per-IP index cache entry to validate left receiver")
	}
	if hotArr != left {
		t.Fatalf("hot array = %p, want left receiver %p", hotArr, left)
	}
	if vm.indexMethodHot.ip != 0 {
		t.Fatalf("single hot cache ip = %d, want direct-promoted ip 0", vm.indexMethodHot.ip)
	}
}

func TestBytecodeVM_LookupHotCanonicalArrayIndexSiteUsesDirectPerIPCacheForSameIdentityReceiver(t *testing.T) {
	interp := NewBytecode()
	preloadArrayStdlibForTest(t, interp)
	left := monoCharArrayValueForTest(t, 'a', 'b')
	right := monoCharArrayValueForTest(t, 'x', 'y')
	alias := monoCharArrayValueForTest(t, 'm', 'n')
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{instructions: []bytecodeInstruction{{}, {}}}
	vm.currentProgram = program

	vm.ip = 0
	if _, fastPath, hasMethod, cacheable, err := vm.resolveCachedIndexMethod(program, 0, left, "get", "Index"); err != nil {
		t.Fatalf("resolveCachedIndexMethod(left): %v", err)
	} else if !cacheable || !hasMethod || fastPath != bytecodeIndexMethodFastPathCanonicalArrayGet {
		t.Fatalf("left cache = fastPath %v hasMethod %v cacheable %v, want canonical get", fastPath, hasMethod, cacheable)
	}

	vm.ip = 1
	if _, fastPath, hasMethod, cacheable, err := vm.resolveCachedIndexMethod(program, 1, right, "get", "Index"); err != nil {
		t.Fatalf("resolveCachedIndexMethod(right): %v", err)
	} else if !cacheable || !hasMethod || fastPath != bytecodeIndexMethodFastPathCanonicalArrayGet {
		t.Fatalf("right cache = fastPath %v hasMethod %v cacheable %v, want canonical get", fastPath, hasMethod, cacheable)
	}

	vm.currentProgram = program
	vm.activeLookup.program = program
	vm.activeLookup.indexMethodGetEntries = nil
	vm.ip = 0
	hotArr, ok := vm.lookupHotCanonicalArrayIndexSite("get", alias, bytecodeIndexMethodFastPathCanonicalArrayGet)
	if !ok {
		t.Fatalf("expected canonical lookup to use direct per-IP cache for same-identity receiver")
	}
	if hotArr != alias {
		t.Fatalf("hot array = %p, want alias receiver %p", hotArr, alias)
	}
	if vm.indexMethodHot.ip != 0 {
		t.Fatalf("single hot cache ip = %d, want direct-promoted ip 0", vm.indexMethodHot.ip)
	}
	if vm.indexMethodHot.receiverArrayHandle != alias.Handle {
		t.Fatalf("single hot cache handle = %d, want alias handle %d", vm.indexMethodHot.receiverArrayHandle, alias.Handle)
	}
}

func TestBytecodeVM_LookupHotCanonicalArrayIndexSetSiteUsesDirectPerIPCache(t *testing.T) {
	interp := NewBytecode()
	preloadArrayStdlibForTest(t, interp)
	left := monoCharArrayValueForTest(t, 'a', 'b')
	right := monoCharArrayValueForTest(t, 'x', 'y')
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{instructions: []bytecodeInstruction{{}, {}}}
	vm.currentProgram = program

	vm.ip = 0
	if _, fastPath, hasMethod, cacheable, err := vm.resolveCachedIndexMethod(program, 0, left, "set", "IndexMut"); err != nil {
		t.Fatalf("resolveCachedIndexMethod(left set): %v", err)
	} else if !cacheable || !hasMethod || fastPath != bytecodeIndexMethodFastPathCanonicalArraySet {
		t.Fatalf("left set cache = fastPath %v hasMethod %v cacheable %v, want canonical set", fastPath, hasMethod, cacheable)
	}

	vm.ip = 1
	if _, fastPath, hasMethod, cacheable, err := vm.resolveCachedIndexMethod(program, 1, right, "set", "IndexMut"); err != nil {
		t.Fatalf("resolveCachedIndexMethod(right set): %v", err)
	} else if !cacheable || !hasMethod || fastPath != bytecodeIndexMethodFastPathCanonicalArraySet {
		t.Fatalf("right set cache = fastPath %v hasMethod %v cacheable %v, want canonical set", fastPath, hasMethod, cacheable)
	}

	vm.currentProgram = program
	vm.activeLookup.program = program
	vm.activeLookup.indexMethodSetEntries = nil
	vm.ip = 0
	hotArr, ok := vm.lookupHotCanonicalArrayIndexSetSite(left)
	if !ok {
		t.Fatalf("expected canonical set lookup to use direct per-IP cache")
	}
	if hotArr != left {
		t.Fatalf("hot array = %p, want left receiver %p", hotArr, left)
	}
	if vm.indexMethodHot.ip != 0 {
		t.Fatalf("single hot cache ip = %d, want direct-promoted ip 0", vm.indexMethodHot.ip)
	}
}

func TestBytecodeVM_LookupDirectCompatibleHotArrayIndexSiteUsesDirectPerIPCache(t *testing.T) {
	interp := NewBytecode()
	preloadArrayStdlibForTest(t, interp)
	left := monoCharArrayValueForTest(t, 'a', 'b')
	right := monoCharArrayValueForTest(t, 'x', 'y')
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{instructions: []bytecodeInstruction{{}, {}}}
	vm.currentProgram = program

	vm.ip = 0
	if _, fastPath, hasMethod, cacheable, err := vm.resolveCachedIndexMethod(program, 0, left, "get", "Index"); err != nil {
		t.Fatalf("resolveCachedIndexMethod(left): %v", err)
	} else if !cacheable || !hasMethod || fastPath != bytecodeIndexMethodFastPathCanonicalArrayGet {
		t.Fatalf("left cache = fastPath %v hasMethod %v cacheable %v, want canonical get", fastPath, hasMethod, cacheable)
	}

	vm.ip = 1
	if _, fastPath, hasMethod, cacheable, err := vm.resolveCachedIndexMethod(program, 1, right, "get", "Index"); err != nil {
		t.Fatalf("resolveCachedIndexMethod(right): %v", err)
	} else if !cacheable || !hasMethod || fastPath != bytecodeIndexMethodFastPathCanonicalArrayGet {
		t.Fatalf("right cache = fastPath %v hasMethod %v cacheable %v, want canonical get", fastPath, hasMethod, cacheable)
	}

	vm.currentProgram = program
	vm.activeLookup.program = program
	vm.activeLookup.indexMethodGetEntries = nil
	vm.ip = 0
	globalRevision, methodCacheVersion := vm.bytecodeGlobalAndMethodVersions()
	hotArr, handle, ok := vm.lookupDirectCompatibleHotArrayIndexSiteForArrayWithVersions(
		bytecodeIndexMethodCacheGet,
		left,
		bytecodeIndexMethodFastPathCanonicalArrayGet,
		globalRevision,
		methodCacheVersion,
	)
	if !ok {
		t.Fatalf("expected direct-compatible lookup to use direct per-IP cache after active entries were cleared")
	}
	if hotArr != left {
		t.Fatalf("hot array = %p, want left receiver %p", hotArr, left)
	}
	if handle != left.Handle {
		t.Fatalf("validated handle = %d, want %d", handle, left.Handle)
	}
	if vm.indexMethodHot.ip != 0 {
		t.Fatalf("single hot cache ip = %d, want direct-promoted ip 0", vm.indexMethodHot.ip)
	}
}

func TestBytecodeVM_LookupDirectCompatibleHotArrayIndexSiteUsesDirectPerIPCacheForSameIdentityReceiver(t *testing.T) {
	interp := NewBytecode()
	preloadArrayStdlibForTest(t, interp)
	left := monoCharArrayValueForTest(t, 'a', 'b')
	right := monoCharArrayValueForTest(t, 'x', 'y')
	alias := monoCharArrayValueForTest(t, 'm', 'n')
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{instructions: []bytecodeInstruction{{}, {}}}
	vm.currentProgram = program

	vm.ip = 0
	if _, fastPath, hasMethod, cacheable, err := vm.resolveCachedIndexMethod(program, 0, left, "get", "Index"); err != nil {
		t.Fatalf("resolveCachedIndexMethod(left): %v", err)
	} else if !cacheable || !hasMethod || fastPath != bytecodeIndexMethodFastPathCanonicalArrayGet {
		t.Fatalf("left cache = fastPath %v hasMethod %v cacheable %v, want canonical get", fastPath, hasMethod, cacheable)
	}

	vm.ip = 1
	if _, fastPath, hasMethod, cacheable, err := vm.resolveCachedIndexMethod(program, 1, right, "get", "Index"); err != nil {
		t.Fatalf("resolveCachedIndexMethod(right): %v", err)
	} else if !cacheable || !hasMethod || fastPath != bytecodeIndexMethodFastPathCanonicalArrayGet {
		t.Fatalf("right cache = fastPath %v hasMethod %v cacheable %v, want canonical get", fastPath, hasMethod, cacheable)
	}

	vm.currentProgram = program
	vm.activeLookup.program = program
	vm.activeLookup.indexMethodGetEntries = nil
	vm.ip = 0
	globalRevision, methodCacheVersion := vm.bytecodeGlobalAndMethodVersions()
	hotArr, handle, ok := vm.lookupDirectCompatibleHotArrayIndexSiteForArrayWithVersions(
		bytecodeIndexMethodCacheGet,
		alias,
		bytecodeIndexMethodFastPathCanonicalArrayGet,
		globalRevision,
		methodCacheVersion,
	)
	if !ok {
		t.Fatalf("expected direct-compatible lookup to use direct per-IP cache for same-identity receiver")
	}
	if hotArr != alias {
		t.Fatalf("hot array = %p, want alias receiver %p", hotArr, alias)
	}
	if handle != alias.Handle {
		t.Fatalf("validated handle = %d, want %d", handle, alias.Handle)
	}
	if vm.indexMethodHot.ip != 0 {
		t.Fatalf("single hot cache ip = %d, want direct-promoted ip 0", vm.indexMethodHot.ip)
	}
	if vm.indexMethodHot.receiverArrayHandle != alias.Handle {
		t.Fatalf("single hot cache handle = %d, want alias handle %d", vm.indexMethodHot.receiverArrayHandle, alias.Handle)
	}
}
