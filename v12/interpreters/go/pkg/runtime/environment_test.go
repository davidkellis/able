package runtime

import (
	"math/big"
	"testing"
)

func TestEnvironmentDefineAndGet(t *testing.T) {
	env := NewEnvironment(nil)
	value := StringValue{Val: "hello"}
	env.Define("greeting", value)

	got, err := env.Get("greeting")
	if err != nil {
		t.Fatalf("expected to retrieve binding: %v", err)
	}

	if gv, ok := got.(StringValue); !ok || gv.Val != "hello" {
		t.Fatalf("unexpected value returned: %#v", got)
	}
}

func TestEnvironmentAssignRespectsLexicalParent(t *testing.T) {
	env := NewEnvironment(nil)
	env.Define("counter", IntegerValue{Val: bigInt(1), TypeSuffix: IntegerI32})

	child := NewEnvironment(env)
	if err := child.Assign("counter", IntegerValue{Val: bigInt(2), TypeSuffix: IntegerI32}); err != nil {
		t.Fatalf("assign into parent failed: %v", err)
	}

	got, err := env.Get("counter")
	if err != nil {
		t.Fatalf("parent lookup failed: %v", err)
	}
	if iv, ok := got.(IntegerValue); !ok || iv.Val.Cmp(bigInt(2)) != 0 {
		t.Fatalf("unexpected counter value: %#v", got)
	}
}

func TestEnvironmentAssignUnknownFails(t *testing.T) {
	env := NewEnvironment(nil)
	err := env.Assign("missing", NilValue{})
	if err == nil {
		t.Fatalf("expected error when assigning undefined variable")
	}
	if err.Error() != "Undefined variable 'missing'" {
		t.Fatalf("unexpected error message: %q", err.Error())
	}
}

func TestEnvironmentLookupRespectsLexicalScope(t *testing.T) {
	parent := NewEnvironment(nil)
	parent.Define("outer", StringValue{Val: "p"})
	child := NewEnvironment(parent)
	child.Define("inner", StringValue{Val: "c"})

	if got, ok := child.Lookup("inner"); !ok {
		t.Fatalf("expected inner lookup to succeed")
	} else if sv, ok := got.(StringValue); !ok || sv.Val != "c" {
		t.Fatalf("unexpected inner value: %#v", got)
	}
	if got, ok := child.Lookup("outer"); !ok {
		t.Fatalf("expected outer lookup via parent to succeed")
	} else if sv, ok := got.(StringValue); !ok || sv.Val != "p" {
		t.Fatalf("unexpected outer value: %#v", got)
	}
	if got, ok := child.Lookup("missing"); ok || got != nil {
		t.Fatalf("expected missing lookup to fail, got (%#v, %t)", got, ok)
	}
}

func TestEnvironmentLookupWithOwnerRespectsLexicalScope(t *testing.T) {
	parent := NewEnvironment(nil)
	parent.Define("outer", StringValue{Val: "p"})
	child := NewEnvironment(parent)
	child.Define("inner", StringValue{Val: "c"})

	if got, owner, ok := child.LookupWithOwner("inner"); !ok {
		t.Fatalf("expected inner lookup with owner to succeed")
	} else if owner != child {
		t.Fatalf("expected inner owner to be child env")
	} else if sv, ok := got.(StringValue); !ok || sv.Val != "c" {
		t.Fatalf("unexpected inner value: %#v", got)
	}
	if got, owner, ok := child.LookupWithOwner("outer"); !ok {
		t.Fatalf("expected outer lookup with owner to succeed")
	} else if owner != parent {
		t.Fatalf("expected outer owner to be parent env")
	} else if sv, ok := got.(StringValue); !ok || sv.Val != "p" {
		t.Fatalf("unexpected outer value: %#v", got)
	}
	if got, owner, ok := child.LookupWithOwner("missing"); ok || got != nil || owner != nil {
		t.Fatalf("expected missing lookup with owner to fail, got (%#v, %p, %t)", got, owner, ok)
	}
}

func TestEnvironmentLookupInCurrentScopeDoesNotWalkParent(t *testing.T) {
	parent := NewEnvironment(nil)
	parent.Define("outer", StringValue{Val: "p"})
	child := NewEnvironment(parent)
	child.Define("inner", StringValue{Val: "c"})

	if got, ok := child.LookupInCurrentScope("inner"); !ok {
		t.Fatalf("expected inner lookup in current scope to succeed")
	} else if sv, ok := got.(StringValue); !ok || sv.Val != "c" {
		t.Fatalf("unexpected inner value: %#v", got)
	}
	if got, ok := child.LookupInCurrentScope("outer"); ok || got != nil {
		t.Fatalf("expected outer lookup in current scope to fail, got (%#v, %t)", got, ok)
	}
}

func TestEnvironmentStructSnapshotCopiesCurrentStructBindings(t *testing.T) {
	env := NewEnvironment(nil)
	def := &StructDefinitionValue{}
	env.DefineStruct("Example", def)

	snapshot := env.StructSnapshot()
	if got, ok := snapshot["Example"]; !ok || got != def {
		t.Fatalf("StructSnapshot[Example] = (%v, %t), want (%v, true)", got, ok, def)
	}
	delete(snapshot, "Example")
	if got, ok := env.StructDefinition("Example"); !ok || got != def {
		t.Fatalf("mutating snapshot should not affect environment; got (%v, %t)", got, ok)
	}
}

func TestEnvironmentRuntimeDataFallsBackToParent(t *testing.T) {
	parent := NewEnvironment(nil)
	child := NewEnvironment(parent)

	parent.SetRuntimeData("root-data")

	if got := child.RuntimeData(); got != "root-data" {
		t.Fatalf("RuntimeData() = %#v, want root-data", got)
	}

	child.SetRuntimeData("child-data")
	if got := child.RuntimeData(); got != "child-data" {
		t.Fatalf("child RuntimeData() = %#v, want child-data", got)
	}
	if got := parent.RuntimeData(); got != "root-data" {
		t.Fatalf("parent RuntimeData() = %#v, want root-data", got)
	}
}

func TestEnvironmentRevisionIncrementsOnMutation(t *testing.T) {
	env := NewEnvironment(nil)
	if got := env.Revision(); got != 0 {
		t.Fatalf("initial revision = %d, want 0", got)
	}
	env.Define("x", IntegerValue{Val: bigInt(1), TypeSuffix: IntegerI32})
	if got := env.Revision(); got != 1 {
		t.Fatalf("revision after define = %d, want 1", got)
	}
	if err := env.Assign("x", IntegerValue{Val: bigInt(2), TypeSuffix: IntegerI32}); err != nil {
		t.Fatalf("assign failed: %v", err)
	}
	if got := env.Revision(); got != 2 {
		t.Fatalf("revision after assign = %d, want 2", got)
	}
	child := NewEnvironment(env)
	if !child.AssignExisting("x", IntegerValue{Val: bigInt(3), TypeSuffix: IntegerI32}) {
		t.Fatalf("assign existing in parent failed")
	}
	if got := env.Revision(); got != 3 {
		t.Fatalf("revision after assign existing = %d, want 3", got)
	}
	if err := child.Assign("missing", NilValue{}); err == nil {
		t.Fatalf("expected assign missing to fail")
	}
	if got := env.Revision(); got != 3 {
		t.Fatalf("failed assign should not change revision, got %d", got)
	}
}

func TestEnvironmentRevisionWithHintMatchesRevision(t *testing.T) {
	env := NewEnvironment(nil)
	env.Define("x", IntegerValue{Val: bigInt(1), TypeSuffix: IntegerI32})

	if got, want := env.RevisionWithHint(false), env.Revision(); got != want {
		t.Fatalf("RevisionWithHint(false) = %d, want %d", got, want)
	}

	env.SetSingleThread()
	if got, want := env.RevisionWithHint(true), env.Revision(); got != want {
		t.Fatalf("RevisionWithHint(true) = %d, want %d", got, want)
	}

	if err := env.Assign("x", IntegerValue{Val: bigInt(2), TypeSuffix: IntegerI32}); err != nil {
		t.Fatalf("assign failed: %v", err)
	}
	if got, want := env.RevisionWithHint(true), env.Revision(); got != want {
		t.Fatalf("RevisionWithHint(true) after assign = %d, want %d", got, want)
	}
}

func TestEnvironmentThreadModePropagatesToChildren(t *testing.T) {
	parent := NewEnvironment(nil)
	if parent.isSingleThread() {
		t.Fatalf("new environment should default to multi-thread mode")
	}

	parent.SetSingleThread()
	child := NewEnvironment(parent)
	if !child.isSingleThread() {
		t.Fatalf("child should inherit single-thread mode from parent")
	}

	parent.SetMultiThread()
	if child.isSingleThread() {
		t.Fatalf("child should observe parent switch to multi-thread mode")
	}
}

func TestEnvironmentChildReusesParentThreadModePointer(t *testing.T) {
	parent := NewEnvironment(nil)
	child := NewEnvironment(parent)
	if child.threadMode != parent.threadMode {
		t.Fatalf("child thread mode pointer should reuse parent mode")
	}
}

func TestEnvironmentMutexAllocatesLazilyInMultiThreadMode(t *testing.T) {
	env := NewEnvironment(nil)
	if env.mu.Load() != nil {
		t.Fatalf("new environment should not allocate mutex eagerly")
	}

	env.DefineWithoutMerge("value", NilValue{})

	if env.mu.Load() == nil {
		t.Fatalf("slow-path mutation should allocate mutex lazily")
	}
}

func TestEnvironmentSingleThreadMutationKeepsMutexNil(t *testing.T) {
	parent := NewEnvironment(nil)
	parent.SetSingleThread()
	child := NewEnvironment(parent)
	if child.mu.Load() != nil {
		t.Fatalf("single-thread child should start without mutex allocation")
	}

	child.DefineWithoutMerge("value", NilValue{})

	if child.mu.Load() != nil {
		t.Fatalf("single-thread mutation should not allocate mutex")
	}
}

func TestEnvironmentNewChildAllocationCount(t *testing.T) {
	parent := NewEnvironment(nil)
	allocs := testing.AllocsPerRun(1000, func() {
		_ = NewEnvironment(parent)
	})
	if allocs > 1.1 {
		t.Fatalf("unexpected child environment allocations: got %.2f want <= 1.1", allocs)
	}
}

func TestEnvironmentDefineWithoutMergeReplacesBinding(t *testing.T) {
	env := NewEnvironment(nil)
	first := StringValue{Val: "first"}
	second := StringValue{Val: "second"}

	env.Define("value", first)
	env.DefineWithoutMerge("value", second)

	got, err := env.Get("value")
	if err != nil {
		t.Fatalf("Get(value): %v", err)
	}
	if got != second {
		t.Fatalf("DefineWithoutMerge should replace binding directly, got %#v want %#v", got, second)
	}
	if gotRevision := env.Revision(); gotRevision != 2 {
		t.Fatalf("revision after Define + DefineWithoutMerge = %d, want 2", gotRevision)
	}
}

func TestEnvironmentSingleBindingUsesInlineSlot(t *testing.T) {
	env := NewEnvironment(nil)
	value := StringValue{Val: "inline"}

	env.DefineWithoutMerge("value", value)

	if env.values != nil {
		t.Fatalf("single binding should not allocate value map")
	}
	if !env.hasSingle || env.singleName != "value" || env.singleValue != value {
		t.Fatalf("unexpected inline binding state: hasSingle=%t name=%q value=%#v", env.hasSingle, env.singleName, env.singleValue)
	}
	if got, ok := env.LookupInCurrentScope("value"); !ok || got != value {
		t.Fatalf("LookupInCurrentScope(value) = (%#v, %t), want (%#v, true)", got, ok, value)
	}
	if keys := env.Keys(); len(keys) != 1 || keys[0] != "value" {
		t.Fatalf("Keys() = %#v, want [value]", keys)
	}
	if snapshot := env.Snapshot(); len(snapshot) != 1 || snapshot["value"] != value {
		t.Fatalf("Snapshot() = %#v, want single inline binding", snapshot)
	}
}

func TestEnvironmentSecondBindingPromotesInlineSlotToMap(t *testing.T) {
	env := NewEnvironment(nil)
	first := StringValue{Val: "first"}
	second := StringValue{Val: "second"}

	env.DefineWithoutMerge("first", first)
	env.DefineWithoutMerge("second", second)

	if env.hasSingle {
		t.Fatalf("second distinct binding should promote inline slot to map")
	}
	if env.values == nil {
		t.Fatalf("promoted environment should have a value map")
	}
	if got, ok := env.LookupInCurrentScope("first"); !ok || got != first {
		t.Fatalf("LookupInCurrentScope(first) = (%#v, %t), want (%#v, true)", got, ok, first)
	}
	if got, ok := env.LookupInCurrentScope("second"); !ok || got != second {
		t.Fatalf("LookupInCurrentScope(second) = (%#v, %t), want (%#v, true)", got, ok, second)
	}
}

func TestEnvironmentAssignExistingUpdatesInlineBinding(t *testing.T) {
	parent := NewEnvironment(nil)
	first := StringValue{Val: "first"}
	second := StringValue{Val: "second"}
	parent.DefineWithoutMerge("value", first)

	child := NewEnvironment(parent)
	if !child.AssignExisting("value", second) {
		t.Fatalf("AssignExisting(value) should succeed")
	}
	if parent.values != nil {
		t.Fatalf("AssignExisting on inline binding should not force map allocation")
	}
	if !parent.hasSingle || parent.singleValue != second {
		t.Fatalf("parent inline binding not updated, got hasSingle=%t value=%#v", parent.hasSingle, parent.singleValue)
	}
}

func TestEnvironmentSingleBindingChildAllocationCount(t *testing.T) {
	parent := NewEnvironment(nil)
	parent.SetSingleThread()
	allocs := testing.AllocsPerRun(1000, func() {
		child := NewEnvironment(parent)
		child.DefineWithoutMerge("value", NilValue{})
	})
	if allocs > 1.1 {
		t.Fatalf("unexpected child+single-binding allocations: got %.2f want <= 1.1", allocs)
	}
}

func bigInt(v int64) *big.Int {
	return big.NewInt(v)
}
