package runtime

import (
	"math/big"
	"testing"
	"unsafe"
)

func TestEnvironmentSizeStaysCompact(t *testing.T) {
	if got := unsafe.Sizeof(Environment{}); got > 192 {
		t.Fatalf("Environment size = %d, want <= 192", got)
	}
}

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
	grandParent := NewEnvironment(nil)
	grandParent.Define("root", StringValue{Val: "g"})
	parent := NewEnvironment(grandParent)
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
	if got, owner, ok := child.LookupWithOwner("root"); !ok {
		t.Fatalf("expected root lookup with owner to succeed")
	} else if owner != grandParent {
		t.Fatalf("expected root owner to be grandparent env")
	} else if sv, ok := got.(StringValue); !ok || sv.Val != "g" {
		t.Fatalf("unexpected root value: %#v", got)
	}
	if got, owner, ok := child.LookupWithOwner("missing"); ok || got != nil || owner != nil {
		t.Fatalf("expected missing lookup with owner to fail, got (%#v, %p, %t)", got, owner, ok)
	}
}

func TestEnvironmentLookupWithOwnerAndRevisionHintRespectsLexicalScope(t *testing.T) {
	t.Run("single-thread", func(t *testing.T) {
		grandParent := NewEnvironment(nil)
		grandParent.SetSingleThread()
		grandParent.Define("root", StringValue{Val: "g"})
		parent := NewEnvironment(grandParent)
		parent.Define("outer", StringValue{Val: "p"})
		child := NewEnvironment(parent)
		child.Define("inner", StringValue{Val: "c"})

		if got, owner, version, ok := child.LookupWithOwnerAndRevisionHint("outer", true); !ok {
			t.Fatalf("expected outer lookup with revision hint to succeed")
		} else if owner != parent {
			t.Fatalf("expected outer owner to be parent env")
		} else if version != parent.Revision() {
			t.Fatalf("expected outer owner revision %d, got %d", parent.Revision(), version)
		} else if sv, ok := got.(StringValue); !ok || sv.Val != "p" {
			t.Fatalf("unexpected outer value: %#v", got)
		}
	})

	t.Run("multi-thread", func(t *testing.T) {
		grandParent := NewEnvironment(nil)
		grandParent.Define("root", StringValue{Val: "g"})
		parent := NewEnvironment(grandParent)
		parent.Define("outer", StringValue{Val: "p"})
		child := NewEnvironment(parent)
		child.Define("inner", StringValue{Val: "c"})

		if got, owner, version, ok := child.LookupWithOwnerAndRevisionHint("root", false); !ok {
			t.Fatalf("expected root lookup with revision hint to succeed")
		} else if owner != grandParent {
			t.Fatalf("expected root owner to be grandparent env")
		} else if version != grandParent.Revision() {
			t.Fatalf("expected root owner revision %d, got %d", grandParent.Revision(), version)
		} else if sv, ok := got.(StringValue); !ok || sv.Val != "g" {
			t.Fatalf("unexpected root value: %#v", got)
		}
	})
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

func TestEnvironmentStructDefinitionInCurrentScopeDoesNotWalkParent(t *testing.T) {
	parent := NewEnvironment(nil)
	parentDef := &StructDefinitionValue{}
	parent.DefineStruct("Span", parentDef)
	child := NewEnvironment(parent)

	if _, ok := child.StructDefinitionInCurrentScope("Span"); ok {
		t.Fatal("child current scope unexpectedly resolved parent struct")
	}
	childDef := &StructDefinitionValue{}
	child.DefineStruct("Span", childDef)
	if got, ok := child.StructDefinitionInCurrentScope("Span"); !ok || got != childDef {
		t.Fatalf("StructDefinitionInCurrentScope(Span) = (%p, %t), want (%p, true)", got, ok, childDef)
	}
	if got, ok := child.StructDefinition("Span"); !ok || got != childDef {
		t.Fatalf("StructDefinition(Span) = (%p, %t), want (%p, true)", got, ok, childDef)
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

func TestEnvironmentRuntimeDataSingleThreadCacheInvalidatesAcrossParentAndChildUpdates(t *testing.T) {
	parent := NewEnvironment(nil)
	parent.SetSingleThread()
	child := NewEnvironment(parent)
	grandchild := NewEnvironment(child)

	if got := grandchild.RuntimeData(); got != nil {
		t.Fatalf("initial grandchild RuntimeData() = %#v, want nil", got)
	}

	parent.SetRuntimeData("root-data")
	if got := grandchild.RuntimeData(); got != "root-data" {
		t.Fatalf("grandchild RuntimeData() after parent set = %#v, want root-data", got)
	}

	child.SetRuntimeData("child-data")
	if got := grandchild.RuntimeData(); got != "child-data" {
		t.Fatalf("grandchild RuntimeData() after child override = %#v, want child-data", got)
	}

	child.SetRuntimeData(nil)
	if got := grandchild.RuntimeData(); got != "root-data" {
		t.Fatalf("grandchild RuntimeData() after child clear = %#v, want root-data", got)
	}

	parent.SetRuntimeData(nil)
	if got := grandchild.RuntimeData(); got != nil {
		t.Fatalf("grandchild RuntimeData() after parent clear = %#v, want nil", got)
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
	if got, want := env.RevisionSingleThread(), env.Revision(); got != want {
		t.Fatalf("RevisionSingleThread() = %d, want %d", got, want)
	}

	if err := env.Assign("x", IntegerValue{Val: bigInt(2), TypeSuffix: IntegerI32}); err != nil {
		t.Fatalf("assign failed: %v", err)
	}
	if got, want := env.RevisionWithHint(true), env.Revision(); got != want {
		t.Fatalf("RevisionWithHint(true) after assign = %d, want %d", got, want)
	}
	if got, want := env.RevisionSingleThread(), env.Revision(); got != want {
		t.Fatalf("RevisionSingleThread() after assign = %d, want %d", got, want)
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
	if child.shared != parent.shared {
		t.Fatalf("child shared state pointer should reuse parent state")
	}
}

func TestEnvironmentMutexAllocatesLazilyInMultiThreadMode(t *testing.T) {
	env := NewEnvironment(nil)
	if env.state.Load() != nil {
		t.Fatalf("new environment should not allocate state eagerly")
	}

	env.DefineWithoutMerge("value", NilValue{})

	if env.state.Load() == nil {
		t.Fatalf("slow-path mutation should allocate state lazily")
	}
}

func TestEnvironmentSingleThreadMutationKeepsMutexNil(t *testing.T) {
	parent := NewEnvironment(nil)
	parent.SetSingleThread()
	child := NewEnvironment(parent)
	if child.state.Load() != nil {
		t.Fatalf("single-thread child should start without state allocation")
	}

	child.DefineWithoutMerge("value", NilValue{})

	if child.state.Load() != nil {
		t.Fatalf("single-thread mutation should not allocate state")
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

func TestEnvironmentDefineWithoutMergeBindingsSeedsMultipleValues(t *testing.T) {
	env := NewEnvironment(nil)
	first := StringValue{Val: "first"}
	second := StringValue{Val: "second"}

	env.DefineWithoutMergeBindings([]EnvironmentBinding{
		{Name: "first", Value: first},
		{Name: "second", Value: second},
	})

	if env.values != nil {
		t.Fatalf("two bindings should stay in inline storage")
	}
	if env.inlineCount != 2 {
		t.Fatalf("inline count = %d, want 2", env.inlineCount)
	}
	if got, ok := env.LookupInCurrentScope("first"); !ok || got != first {
		t.Fatalf("LookupInCurrentScope(first) = (%#v, %t), want (%#v, true)", got, ok, first)
	}
	if got, ok := env.LookupInCurrentScope("second"); !ok || got != second {
		t.Fatalf("LookupInCurrentScope(second) = (%#v, %t), want (%#v, true)", got, ok, second)
	}
	if gotRevision := env.Revision(); gotRevision != 2 {
		t.Fatalf("revision after two batched bindings = %d, want 2", gotRevision)
	}
}

func TestNewEnvironmentWithBindingsSeedsInitialScope(t *testing.T) {
	parent := NewEnvironment(nil)
	parent.SetSingleThread()
	first := StringValue{Val: "first"}
	second := StringValue{Val: "second"}

	child := NewEnvironmentWithBindings(parent, 4, []EnvironmentBinding{
		{Name: "first", Value: first},
		{Name: "second", Value: second},
	})

	if child.Parent() != parent {
		t.Fatalf("child parent mismatch")
	}
	if child.shared != parent.shared {
		t.Fatalf("child shared state pointer should reuse parent state")
	}
	if child.values != nil {
		t.Fatalf("two seeded bindings should stay in inline storage")
	}
	if child.inlineCount != 2 {
		t.Fatalf("inline count = %d, want 2", child.inlineCount)
	}
	if got, ok := child.LookupInCurrentScope("first"); !ok || got != first {
		t.Fatalf("LookupInCurrentScope(first) = (%#v, %t), want (%#v, true)", got, ok, first)
	}
	if got, ok := child.LookupInCurrentScope("second"); !ok || got != second {
		t.Fatalf("LookupInCurrentScope(second) = (%#v, %t), want (%#v, true)", got, ok, second)
	}
	if gotRevision := child.Revision(); gotRevision != 2 {
		t.Fatalf("revision after seeded child bindings = %d, want 2", gotRevision)
	}
}

func TestNewEnvironmentWithSingleBindingSeedsInitialScope(t *testing.T) {
	parent := NewEnvironment(nil)
	parent.SetSingleThread()
	value := StringValue{Val: "value"}

	child := NewEnvironmentWithSingleBinding(parent, 4, "value", value)

	if child.Parent() != parent {
		t.Fatalf("child parent mismatch")
	}
	if child.shared != parent.shared {
		t.Fatalf("child shared state pointer should reuse parent state")
	}
	if child.values != nil || child.spill != nil {
		t.Fatalf("single seeded binding should stay in inline storage")
	}
	if child.inlineCount != 1 {
		t.Fatalf("inline count = %d, want 1", child.inlineCount)
	}
	if got, ok := child.LookupInCurrentScope("value"); !ok || got != value {
		t.Fatalf("LookupInCurrentScope(value) = (%#v, %t), want (%#v, true)", got, ok, value)
	}
	if gotRevision := child.Revision(); gotRevision != 1 {
		t.Fatalf("revision after seeded single binding = %d, want 1", gotRevision)
	}
}

func TestNewEnvironmentWithSingleBindingHintedFifthBindingUsesSpillStorage(t *testing.T) {
	parent := NewEnvironment(nil)
	parent.SetSingleThread()
	first := StringValue{Val: "first"}
	second := StringValue{Val: "second"}
	third := StringValue{Val: "third"}
	fourth := StringValue{Val: "fourth"}
	fifth := StringValue{Val: "fifth"}

	child := NewEnvironmentWithSingleBinding(parent, 6, "first", first)
	child.DefineWithoutMerge("second", second)
	child.DefineWithoutMerge("third", third)
	child.DefineWithoutMerge("fourth", fourth)
	child.DefineWithoutMerge("fifth", fifth)

	if child.values != nil {
		t.Fatalf("hinted single-binding child should avoid map storage")
	}
	if child.inlineCount != 0 {
		t.Fatalf("hinted spill should clear inline bindings, got count=%d", child.inlineCount)
	}
	if child.spill == nil {
		t.Fatalf("expected hinted spill storage")
	}
	if child.spill.count != 5 {
		t.Fatalf("hinted spill count = %d, want 5", child.spill.count)
	}
	if got, ok := child.LookupInCurrentScope("first"); !ok || got != first {
		t.Fatalf("LookupInCurrentScope(first) = (%#v, %t), want (%#v, true)", got, ok, first)
	}
	if got, ok := child.LookupInCurrentScope("second"); !ok || got != second {
		t.Fatalf("LookupInCurrentScope(second) = (%#v, %t), want (%#v, true)", got, ok, second)
	}
	if got, ok := child.LookupInCurrentScope("third"); !ok || got != third {
		t.Fatalf("LookupInCurrentScope(third) = (%#v, %t), want (%#v, true)", got, ok, third)
	}
	if got, ok := child.LookupInCurrentScope("fourth"); !ok || got != fourth {
		t.Fatalf("LookupInCurrentScope(fourth) = (%#v, %t), want (%#v, true)", got, ok, fourth)
	}
	if got, ok := child.LookupInCurrentScope("fifth"); !ok || got != fifth {
		t.Fatalf("LookupInCurrentScope(fifth) = (%#v, %t), want (%#v, true)", got, ok, fifth)
	}
}

func TestNewEnvironmentWithBindingSetsSeedsAndShadowsInOrder(t *testing.T) {
	parent := NewEnvironment(nil)
	parent.SetSingleThread()
	first := StringValue{Val: "first"}
	override := StringValue{Val: "override"}
	third := StringValue{Val: "third"}

	child := NewEnvironmentWithBindingSets(
		parent,
		6,
		[]EnvironmentBinding{
			{Name: "value", Value: first},
			{Name: "other", Value: third},
		},
		[]EnvironmentBinding{
			{Name: "value", Value: override},
		},
	)

	if child.Parent() != parent {
		t.Fatalf("child parent mismatch")
	}
	if child.shared != parent.shared {
		t.Fatalf("child shared state pointer should reuse parent state")
	}
	if child.values != nil {
		t.Fatalf("small binding sets should avoid map storage")
	}
	if child.spill != nil {
		t.Fatalf("duplicate-shadowed binding sets should stay inline")
	}
	if child.inlineCount != 2 {
		t.Fatalf("inline count = %d, want 2 after shadowing duplicate name", child.inlineCount)
	}
	if got, ok := child.LookupInCurrentScope("value"); !ok || got != override {
		t.Fatalf("LookupInCurrentScope(value) = (%#v, %t), want (%#v, true)", got, ok, override)
	}
	if got, ok := child.LookupInCurrentScope("other"); !ok || got != third {
		t.Fatalf("LookupInCurrentScope(other) = (%#v, %t), want (%#v, true)", got, ok, third)
	}
	if gotRevision := child.Revision(); gotRevision != 3 {
		t.Fatalf("revision after seeded binding sets = %d, want 3", gotRevision)
	}
}

func TestEnvironmentSingleBindingUsesInlineSlot(t *testing.T) {
	env := NewEnvironment(nil)
	value := StringValue{Val: "inline"}

	env.DefineWithoutMerge("value", value)

	if env.values != nil {
		t.Fatalf("single binding should not allocate value map")
	}
	if env.inlineCount != 1 || env.inlineNames[0] != "value" || env.inlineValues[0] != value {
		t.Fatalf("unexpected inline binding state: count=%d name=%q value=%#v", env.inlineCount, env.inlineNames[0], env.inlineValues[0])
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

func TestEnvironmentSecondBindingUsesSecondInlineSlot(t *testing.T) {
	env := NewEnvironment(nil)
	first := StringValue{Val: "first"}
	second := StringValue{Val: "second"}

	env.DefineWithoutMerge("first", first)
	env.DefineWithoutMerge("second", second)

	if env.values != nil {
		t.Fatalf("second distinct binding should stay in inline storage")
	}
	if env.inlineCount != 2 {
		t.Fatalf("inline count = %d, want 2", env.inlineCount)
	}
	if got, ok := env.LookupInCurrentScope("first"); !ok || got != first {
		t.Fatalf("LookupInCurrentScope(first) = (%#v, %t), want (%#v, true)", got, ok, first)
	}
	if got, ok := env.LookupInCurrentScope("second"); !ok || got != second {
		t.Fatalf("LookupInCurrentScope(second) = (%#v, %t), want (%#v, true)", got, ok, second)
	}
	if keys := env.Keys(); len(keys) != 2 || keys[0] != "first" || keys[1] != "second" {
		t.Fatalf("Keys() = %#v, want [first second]", keys)
	}
	if snapshot := env.Snapshot(); len(snapshot) != 2 || snapshot["first"] != first || snapshot["second"] != second {
		t.Fatalf("Snapshot() = %#v, want two inline bindings", snapshot)
	}
}

func TestEnvironmentFourthBindingUsesFourthInlineSlot(t *testing.T) {
	env := NewEnvironment(nil)
	first := StringValue{Val: "first"}
	second := StringValue{Val: "second"}
	third := StringValue{Val: "third"}
	fourth := StringValue{Val: "fourth"}

	env.DefineWithoutMerge("first", first)
	env.DefineWithoutMerge("second", second)
	env.DefineWithoutMerge("third", third)
	env.DefineWithoutMerge("fourth", fourth)

	if env.values != nil || env.spill != nil {
		t.Fatalf("four distinct bindings should stay in inline storage")
	}
	if env.inlineCount != 4 {
		t.Fatalf("inline count = %d, want 4", env.inlineCount)
	}
	if got, ok := env.LookupInCurrentScope("fourth"); !ok || got != fourth {
		t.Fatalf("LookupInCurrentScope(fourth) = (%#v, %t), want (%#v, true)", got, ok, fourth)
	}
	if keys := env.Keys(); len(keys) != 4 || keys[0] != "first" || keys[1] != "fourth" || keys[2] != "second" || keys[3] != "third" {
		t.Fatalf("Keys() = %#v, want [first fourth second third]", keys)
	}
	if snapshot := env.Snapshot(); len(snapshot) != 4 || snapshot["first"] != first || snapshot["second"] != second || snapshot["third"] != third || snapshot["fourth"] != fourth {
		t.Fatalf("Snapshot() = %#v, want four inline bindings", snapshot)
	}
}

func TestEnvironmentFifthBindingPromotesInlineBindingsToMap(t *testing.T) {
	env := NewEnvironment(nil)
	first := StringValue{Val: "first"}
	second := StringValue{Val: "second"}
	third := StringValue{Val: "third"}
	fourth := StringValue{Val: "fourth"}
	fifth := StringValue{Val: "fifth"}

	env.DefineWithoutMerge("first", first)
	env.DefineWithoutMerge("second", second)
	env.DefineWithoutMerge("third", third)
	env.DefineWithoutMerge("fourth", fourth)
	env.DefineWithoutMerge("fifth", fifth)

	if env.values == nil {
		t.Fatalf("fifth distinct binding should promote to a value map")
	}
	if env.inlineCount != 0 {
		t.Fatalf("promoted environment should clear inline bindings, got count=%d", env.inlineCount)
	}
	if got, ok := env.LookupInCurrentScope("first"); !ok || got != first {
		t.Fatalf("LookupInCurrentScope(first) = (%#v, %t), want (%#v, true)", got, ok, first)
	}
	if got, ok := env.LookupInCurrentScope("second"); !ok || got != second {
		t.Fatalf("LookupInCurrentScope(second) = (%#v, %t), want (%#v, true)", got, ok, second)
	}
	if got, ok := env.LookupInCurrentScope("third"); !ok || got != third {
		t.Fatalf("LookupInCurrentScope(third) = (%#v, %t), want (%#v, true)", got, ok, third)
	}
	if got, ok := env.LookupInCurrentScope("fourth"); !ok || got != fourth {
		t.Fatalf("LookupInCurrentScope(fourth) = (%#v, %t), want (%#v, true)", got, ok, fourth)
	}
	if got, ok := env.LookupInCurrentScope("fifth"); !ok || got != fifth {
		t.Fatalf("LookupInCurrentScope(fifth) = (%#v, %t), want (%#v, true)", got, ok, fifth)
	}
}

func TestEnvironmentHintedFifthBindingUsesSpillStorage(t *testing.T) {
	parent := NewEnvironment(nil)
	parent.SetSingleThread()
	env := NewEnvironmentWithValueCapacity(parent, 6)
	first := StringValue{Val: "first"}
	second := StringValue{Val: "second"}
	third := StringValue{Val: "third"}
	fourth := StringValue{Val: "fourth"}
	fifth := StringValue{Val: "fifth"}

	env.DefineWithoutMerge("first", first)
	env.DefineWithoutMerge("second", second)
	env.DefineWithoutMerge("third", third)
	env.DefineWithoutMerge("fourth", fourth)
	env.DefineWithoutMerge("fifth", fifth)

	if env.values != nil {
		t.Fatalf("hinted fifth binding should avoid map storage")
	}
	if env.inlineCount != 0 {
		t.Fatalf("hinted spill should clear inline bindings, got count=%d", env.inlineCount)
	}
	if env.spill == nil {
		t.Fatalf("expected hinted spill storage")
	}
	if env.spill.count != 5 {
		t.Fatalf("hinted spill count = %d, want 5", env.spill.count)
	}
	if got, ok := env.LookupInCurrentScope("first"); !ok || got != first {
		t.Fatalf("LookupInCurrentScope(first) = (%#v, %t), want (%#v, true)", got, ok, first)
	}
	if got, ok := env.LookupInCurrentScope("second"); !ok || got != second {
		t.Fatalf("LookupInCurrentScope(second) = (%#v, %t), want (%#v, true)", got, ok, second)
	}
	if got, ok := env.LookupInCurrentScope("third"); !ok || got != third {
		t.Fatalf("LookupInCurrentScope(third) = (%#v, %t), want (%#v, true)", got, ok, third)
	}
	if got, ok := env.LookupInCurrentScope("fourth"); !ok || got != fourth {
		t.Fatalf("LookupInCurrentScope(fourth) = (%#v, %t), want (%#v, true)", got, ok, fourth)
	}
	if got, ok := env.LookupInCurrentScope("fifth"); !ok || got != fifth {
		t.Fatalf("LookupInCurrentScope(fifth) = (%#v, %t), want (%#v, true)", got, ok, fifth)
	}
	if keys := env.Keys(); len(keys) != 5 || keys[0] != "fifth" || keys[1] != "first" || keys[2] != "fourth" || keys[3] != "second" || keys[4] != "third" {
		t.Fatalf("Keys() = %#v, want [fifth first fourth second third]", keys)
	}
	if snapshot := env.Snapshot(); len(snapshot) != 5 || snapshot["first"] != first || snapshot["second"] != second || snapshot["third"] != third || snapshot["fourth"] != fourth || snapshot["fifth"] != fifth {
		t.Fatalf("Snapshot() = %#v, want five spill bindings", snapshot)
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
	if parent.inlineCount != 1 || parent.inlineValues[0] != second {
		t.Fatalf("parent inline binding not updated, got count=%d value=%#v", parent.inlineCount, parent.inlineValues[0])
	}
}

func TestEnvironmentTwoInlineBindingsChildAllocationCount(t *testing.T) {
	parent := NewEnvironment(nil)
	parent.SetSingleThread()
	allocs := testing.AllocsPerRun(1000, func() {
		child := NewEnvironment(parent)
		child.DefineWithoutMerge("first", NilValue{})
		child.DefineWithoutMerge("second", NilValue{})
	})
	if allocs > 1.1 {
		t.Fatalf("unexpected child+two-inline-bindings allocations: got %.2f want <= 1.1", allocs)
	}
}

func TestEnvironmentFourInlineBindingsChildAllocationCount(t *testing.T) {
	parent := NewEnvironment(nil)
	parent.SetSingleThread()
	allocs := testing.AllocsPerRun(1000, func() {
		child := NewEnvironment(parent)
		child.DefineWithoutMerge("first", NilValue{})
		child.DefineWithoutMerge("second", NilValue{})
		child.DefineWithoutMerge("third", NilValue{})
		child.DefineWithoutMerge("fourth", NilValue{})
	})
	if allocs > 1.1 {
		t.Fatalf("unexpected child+four-inline-bindings allocations: got %.2f want <= 1.1", allocs)
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

func TestEnvironmentHintedSingleBindingChildAllocationCount(t *testing.T) {
	parent := NewEnvironment(nil)
	parent.SetSingleThread()
	allocs := testing.AllocsPerRun(1000, func() {
		child := NewEnvironmentWithValueCapacity(parent, 4)
		child.DefineWithoutMerge("first", NilValue{})
	})
	if allocs > 1.1 {
		t.Fatalf("unexpected hinted child+single-binding allocations: got %.2f want <= 1.1", allocs)
	}
}

func TestNewEnvironmentWithSingleBindingChildAllocationCount(t *testing.T) {
	parent := NewEnvironment(nil)
	parent.SetSingleThread()
	allocs := testing.AllocsPerRun(1000, func() {
		_ = NewEnvironmentWithSingleBinding(parent, 4, "value", NilValue{})
	})
	if allocs > 1.1 {
		t.Fatalf("unexpected seeded single-binding child allocations: got %.2f want <= 1.1", allocs)
	}
}

func TestEnvironmentHintedFifthBindingChildAllocationCount(t *testing.T) {
	parent := NewEnvironment(nil)
	parent.SetSingleThread()
	allocs := testing.AllocsPerRun(1000, func() {
		child := NewEnvironmentWithValueCapacity(parent, 6)
		child.DefineWithoutMerge("first", NilValue{})
		child.DefineWithoutMerge("second", NilValue{})
		child.DefineWithoutMerge("third", NilValue{})
		child.DefineWithoutMerge("fourth", NilValue{})
		child.DefineWithoutMerge("fifth", NilValue{})
	})
	if allocs > 2.1 {
		t.Fatalf("unexpected hinted child+fifth-binding allocations: got %.2f want <= 2.1", allocs)
	}
}

func bigInt(v int64) *big.Int {
	return big.NewInt(v)
}
