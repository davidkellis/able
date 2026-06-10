package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_ArrayIndexMethodReceiverTypeKeyCachesHandleRevisionWithoutMaterializingAliasState(t *testing.T) {
	interp := NewBytecode()
	innerChar := monoCharArrayValueForTest(t, 'a', 'b')
	outer := interp.newArrayValue([]runtime.Value{innerChar}, 1)
	alias := &runtime.ArrayValue{Handle: outer.Handle, TrackedHandle: outer.Handle}
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())

	receiverKind, elemType, firstIdentityKey, _, _, ok := vm.indexMethodCacheIdentityKey(alias)
	if !ok {
		t.Fatalf("expected alias-backed nominal array receiver identity")
	}
	if receiverKind != bytecodeMemberReceiverArray {
		t.Fatalf("receiver kind = %v, want array", receiverKind)
	}
	if elemType != bytecodeIndexTypeUnknown {
		t.Fatalf("first element token = %v, want unknown for nominal nested array", elemType)
	}
	if firstIdentityKey != "Array<char>" {
		t.Fatalf("first identity receiver type key = %q, want Array<char>", firstIdentityKey)
	}
	firstKey, ok := vm.arrayIndexMethodReceiverTypeKey(alias)
	if !ok {
		t.Fatalf("expected alias-backed nominal array receiver key")
	}
	if firstKey != "Array<char>" {
		t.Fatalf("first receiver type key = %q, want Array<char>", firstKey)
	}
	if alias.State != nil || alias.Elements != nil {
		t.Fatalf("alias receiver key lookup should not materialize boxed state")
	}

	firstRevision, ok, err := runtime.ArrayStoreRevisionIfAvailable(outer.Handle)
	if err != nil || !ok {
		t.Fatalf("ArrayStoreRevisionIfAvailable first = (%d, %v, %v), want revision/true/nil", firstRevision, ok, err)
	}
	entry, ok := vm.arrayIndexReceiverIdentityCache[outer.Handle]
	if !ok {
		t.Fatalf("expected receiver key cache entry for handle %d", outer.Handle)
	}
	if entry.revision != firstRevision || entry.elemType != bytecodeIndexTypeUnknown || entry.key != "Array<char>" || !entry.ok {
		t.Fatalf("first cache entry = %#v, want revision=%d elemType=unknown key Array<char>", entry, firstRevision)
	}

	innerU32 := monoU32ArrayValueForTest(t, 7, 11)
	if err := runtime.ArrayStoreWrite(outer.Handle, 0, innerU32); err != nil {
		t.Fatalf("ArrayStoreWrite update outer[0]: %v", err)
	}

	receiverKind, elemType, secondIdentityKey, _, _, ok := vm.indexMethodCacheIdentityKey(alias)
	if !ok {
		t.Fatalf("expected updated alias-backed nominal array receiver identity")
	}
	if receiverKind != bytecodeMemberReceiverArray {
		t.Fatalf("updated receiver kind = %v, want array", receiverKind)
	}
	if elemType != bytecodeIndexTypeUnknown {
		t.Fatalf("second element token = %v, want unknown for nominal nested array", elemType)
	}
	if secondIdentityKey != "Array<u32>" {
		t.Fatalf("second identity receiver type key = %q, want Array<u32>", secondIdentityKey)
	}
	secondKey, ok := vm.arrayIndexMethodReceiverTypeKey(alias)
	if !ok {
		t.Fatalf("expected updated alias-backed nominal array receiver key")
	}
	if secondKey != "Array<u32>" {
		t.Fatalf("second receiver type key = %q, want Array<u32>", secondKey)
	}
	if alias.State != nil || alias.Elements != nil {
		t.Fatalf("alias receiver key lookup after revision change should not materialize boxed state")
	}

	secondRevision, ok, err := runtime.ArrayStoreRevisionIfAvailable(outer.Handle)
	if err != nil || !ok {
		t.Fatalf("ArrayStoreRevisionIfAvailable second = (%d, %v, %v), want revision/true/nil", secondRevision, ok, err)
	}
	if secondRevision <= firstRevision {
		t.Fatalf("second revision = %d, want > %d", secondRevision, firstRevision)
	}
	entry = vm.arrayIndexReceiverIdentityCache[outer.Handle]
	if entry.revision != secondRevision || entry.elemType != bytecodeIndexTypeUnknown || entry.key != "Array<u32>" || !entry.ok {
		t.Fatalf("second cache entry = %#v, want revision=%d elemType=unknown key Array<u32>", entry, secondRevision)
	}
}

func TestBytecodeVM_ArrayIndexReceiverIdentityUsesTrackedPrimitiveTokenWithoutCache(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	state := &runtime.ArrayState{
		Values:                []runtime.Value{boxedOrSmallIntegerValue(runtime.IntegerI32, 7)},
		Revision:              11,
		ElementTypeToken:      bytecodeIndexTypeI32,
		ElementTypeTokenKnown: true,
	}
	arr := &runtime.ArrayValue{
		Handle:        9901,
		TrackedHandle: 9901,
		State:         state,
	}

	elemType, key, ok := vm.arrayIndexReceiverIdentity(arr)
	if !ok || elemType != bytecodeIndexTypeI32 || key != "" {
		t.Fatalf("first identity = (%d, %q, %v), want i32/empty/true", elemType, key, ok)
	}
	if vm.arrayIndexReceiverIdentityCache != nil {
		t.Fatalf("primitive tracked identity should not populate cache, got %#v", vm.arrayIndexReceiverIdentityCache)
	}

	state.Values[0] = runtime.BoolValue{Val: true}
	state.Revision = 12
	state.ElementTypeToken = bytecodeIndexTypeBool

	elemType, key, ok = vm.arrayIndexReceiverIdentity(arr)
	if !ok || elemType != bytecodeIndexTypeBool || key != "" {
		t.Fatalf("second identity = (%d, %q, %v), want bool/empty/true", elemType, key, ok)
	}
	if vm.arrayIndexReceiverIdentityCache != nil {
		t.Fatalf("primitive tracked identity after token change should not populate cache, got %#v", vm.arrayIndexReceiverIdentityCache)
	}
}

func TestBytecodeVM_ArrayIndexReceiverIdentityUsesMonoHandleTypeWithoutRevisionCache(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := monoCharArrayValueForTest(t, 'a', 'b')

	elemType, key, ok := vm.arrayIndexReceiverIdentity(arr)
	if !ok || elemType != bytecodeIndexTypeChar || key != "" {
		t.Fatalf("mono identity = (%d, %q, %v), want char/empty/true", elemType, key, ok)
	}
	firstRevision, revOK, err := runtime.ArrayStoreRevisionIfAvailable(arr.Handle)
	if err != nil || !revOK {
		t.Fatalf("ArrayStoreRevisionIfAvailable first = (%d, %v, %v), want revision/true/nil", firstRevision, revOK, err)
	}
	if !vm.arrayIndexReceiverMonoTokenHotOK ||
		vm.arrayIndexReceiverMonoTokenHotHandle != arr.Handle ||
		vm.arrayIndexReceiverMonoTokenHotRevision != firstRevision ||
		vm.arrayIndexReceiverMonoTokenHot != bytecodeIndexTypeChar {
		t.Fatalf("mono token hot cache = handle %d revision %d token %d ok %v, want handle %d revision %d char/true",
			vm.arrayIndexReceiverMonoTokenHotHandle,
			vm.arrayIndexReceiverMonoTokenHotRevision,
			vm.arrayIndexReceiverMonoTokenHot,
			vm.arrayIndexReceiverMonoTokenHotOK,
			arr.Handle,
			firstRevision)
	}
	if vm.arrayIndexReceiverIdentityCache != nil {
		t.Fatalf("mono primitive identity should not populate revision cache, got %#v", vm.arrayIndexReceiverIdentityCache)
	}

	if _, err := runtime.ArrayStoreState(arr.Handle); err != nil {
		t.Fatalf("deopt mono array to dynamic state: %v", err)
	}
	if err := runtime.ArrayStoreWrite(arr.Handle, 0, runtime.BoolValue{Val: true}); err != nil {
		t.Fatalf("write deopted dynamic array: %v", err)
	}

	elemType, key, ok = vm.arrayIndexReceiverIdentity(arr)
	if !ok || elemType != bytecodeIndexTypeBool || key != "" {
		t.Fatalf("deopted identity = (%d, %q, %v), want bool/empty/true", elemType, key, ok)
	}
}

func TestBytecodeVM_ArrayIndexReceiverMonoTokenHotInvalidatesOnEmptyDeopt(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := monoCharArrayValueForTest(t)

	elemType, key, ok := vm.arrayIndexReceiverIdentity(arr)
	if !ok || elemType != bytecodeIndexTypeChar || key != "" {
		t.Fatalf("mono empty identity = (%d, %q, %v), want char/empty/true", elemType, key, ok)
	}

	if _, err := runtime.ArrayStoreState(arr.Handle); err != nil {
		t.Fatalf("deopt mono array to dynamic state: %v", err)
	}

	elemType, key, ok = vm.arrayIndexReceiverIdentity(arr)
	if ok || elemType != bytecodeIndexTypeUnknown || key != "" {
		t.Fatalf("deopted empty identity = (%d, %q, %v), want unknown/empty/false", elemType, key, ok)
	}
}

func TestBytecodeVM_ArrayElementTypeTokenForPropagationUsesMonoTokenHotRevision(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := monoCharArrayValueForTest(t)

	elemType, key, ok := vm.arrayIndexReceiverIdentity(arr)
	if !ok || elemType != bytecodeIndexTypeChar || key != "" {
		t.Fatalf("mono identity = (%d, %q, %v), want char/empty/true", elemType, key, ok)
	}
	token, known := vm.arrayElementTypeTokenForPropagation(arr)
	if !known || token != bytecodeIndexTypeChar {
		t.Fatalf("propagation token = (%d, %v), want char/true", token, known)
	}

	if _, err := runtime.ArrayStoreState(arr.Handle); err != nil {
		t.Fatalf("deopt mono array to dynamic state: %v", err)
	}
	token, known = vm.arrayElementTypeTokenForPropagation(arr)
	if token != bytecodeIndexTypeUnknown {
		t.Fatalf("deopted propagation token = (%d, %v), want unknown token", token, known)
	}
}

func TestBytecodeVM_ArrayIndexReceiverIdentityCachesTrackedNominalRevision(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	state := &runtime.ArrayState{
		Values:   []runtime.Value{&runtime.HostHandleValue{HandleType: "Widget"}},
		Revision: 21,
	}
	arr := &runtime.ArrayValue{
		Handle:        9902,
		TrackedHandle: 9902,
		State:         state,
	}

	elemType, key, ok := vm.arrayIndexReceiverIdentity(arr)
	if !ok || elemType != bytecodeIndexTypeUnknown || key != "Widget" {
		t.Fatalf("first identity = (%d, %q, %v), want unknown/Widget/true", elemType, key, ok)
	}
	entry, ok := vm.arrayIndexReceiverIdentityCache[arr.Handle]
	if !ok {
		t.Fatalf("expected updated tracked-state identity cache entry")
	}
	if entry.revision != 21 || entry.elemType != bytecodeIndexTypeUnknown || entry.key != "Widget" || !entry.ok {
		t.Fatalf("first tracked cache entry = %#v, want revision=21 key=Widget", entry)
	}

	state.Values[0] = &runtime.HostHandleValue{HandleType: "Gadget"}
	state.Revision = 22

	elemType, key, ok = vm.arrayIndexReceiverIdentity(arr)
	if !ok || elemType != bytecodeIndexTypeUnknown || key != "Gadget" {
		t.Fatalf("second identity = (%d, %q, %v), want unknown/Gadget/true", elemType, key, ok)
	}
	entry, ok = vm.arrayIndexReceiverIdentityCache[arr.Handle]
	if !ok {
		t.Fatalf("expected updated tracked-state identity cache entry")
	}
	if entry.revision != 22 || entry.elemType != bytecodeIndexTypeUnknown || entry.key != "Gadget" || !entry.ok {
		t.Fatalf("second tracked cache entry = %#v, want revision=22 key=Gadget", entry)
	}
}

func TestBytecodeVM_ArrayIndexReceiverIdentityHotCacheTracksRevision(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	state := &runtime.ArrayState{
		Values:   []runtime.Value{&runtime.HostHandleValue{HandleType: "Widget"}},
		Revision: 31,
	}
	arr := &runtime.ArrayValue{
		Handle:        9903,
		TrackedHandle: 9903,
		State:         state,
	}

	elemType, key, ok := vm.arrayIndexReceiverIdentity(arr)
	if !ok || elemType != bytecodeIndexTypeUnknown || key != "Widget" {
		t.Fatalf("first identity = (%d, %q, %v), want unknown/Widget/true", elemType, key, ok)
	}
	if vm.arrayIndexReceiverIdentityHotHandle != arr.Handle ||
		vm.arrayIndexReceiverIdentityHot.revision != state.Revision ||
		vm.arrayIndexReceiverIdentityHot.key != "Widget" {
		t.Fatalf("hot identity cache = handle %d entry %#v, want handle %d revision 31 key Widget",
			vm.arrayIndexReceiverIdentityHotHandle, vm.arrayIndexReceiverIdentityHot, arr.Handle)
	}

	vm.arrayIndexReceiverIdentityCache[arr.Handle] = bytecodeArrayIndexReceiverIdentityCacheEntry{
		revision: state.Revision,
		elemType: bytecodeIndexTypeUnknown,
		key:      "Wrong",
		ok:       true,
	}
	elemType, key, ok = vm.arrayIndexReceiverIdentity(arr)
	if !ok || elemType != bytecodeIndexTypeUnknown || key != "Widget" {
		t.Fatalf("hot identity = (%d, %q, %v), want unknown/Widget/true", elemType, key, ok)
	}

	state.Values[0] = &runtime.HostHandleValue{HandleType: "Gadget"}
	state.Revision = 32
	elemType, key, ok = vm.arrayIndexReceiverIdentity(arr)
	if !ok || elemType != bytecodeIndexTypeUnknown || key != "Gadget" {
		t.Fatalf("revision-invalidated identity = (%d, %q, %v), want unknown/Gadget/true", elemType, key, ok)
	}
	if vm.arrayIndexReceiverIdentityHot.revision != 32 || vm.arrayIndexReceiverIdentityHot.key != "Gadget" {
		t.Fatalf("updated hot identity cache = %#v, want revision 32 key Gadget", vm.arrayIndexReceiverIdentityHot)
	}
}
