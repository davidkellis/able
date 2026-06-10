package runtime

import "testing"

func TestEnvironmentResetForBindingSetsReuseClearsStateAndSeedsBindings(t *testing.T) {
	firstParent := NewEnvironment(nil)
	firstParent.SetSingleThread()
	secondParent := NewEnvironment(nil)
	secondParent.SetSingleThread()
	secondParent.SetRuntimeData("parent-data")

	env := NewEnvironmentWithBindingSets(
		firstParent,
		6,
		[]EnvironmentBinding{{Name: "old", Value: StringValue{Val: "old"}}},
		nil,
	)
	env.SetSingleThread()
	env.DefineStruct("Old", &StructDefinitionValue{})
	env.SetRuntimeData("old-data")

	beforeRevision := env.Revision()
	beforeRuntimeDataRevision := env.RuntimeDataRevision()

	env.ResetForBindingSetsReuse(
		secondParent,
		6,
		[]EnvironmentBinding{
			{Name: "left", Value: StringValue{Val: "left"}},
			{Name: "shared", Value: StringValue{Val: "first"}},
		},
		[]EnvironmentBinding{
			{Name: "shared", Value: StringValue{Val: "override"}},
		},
	)

	if env.Parent() != secondParent {
		t.Fatalf("reused env parent mismatch")
	}
	if env.shared != secondParent.shared {
		t.Fatalf("reused env shared state pointer should reuse new parent state")
	}
	if got, ok := env.LookupInCurrentScope("old"); ok || got != nil {
		t.Fatalf("expected old binding to be cleared, got (%#v, %t)", got, ok)
	}
	if got, ok := env.LookupInCurrentScope("left"); !ok {
		t.Fatalf("expected left binding after reset")
	} else if str, ok := got.(StringValue); !ok || str.Val != "left" {
		t.Fatalf("left binding = %#v, want left", got)
	}
	if got, ok := env.LookupInCurrentScope("shared"); !ok {
		t.Fatalf("expected shared binding after reset")
	} else if str, ok := got.(StringValue); !ok || str.Val != "override" {
		t.Fatalf("shared binding = %#v, want override", got)
	}
	if _, ok := env.StructDefinition("Old"); ok {
		t.Fatalf("expected struct definitions to clear across reset")
	}
	if got := env.RuntimeData(); got != "parent-data" {
		t.Fatalf("RuntimeData() after reset = %#v, want parent-data", got)
	}
	if got := env.Revision(); got <= beforeRevision {
		t.Fatalf("revision after reset = %d, want > %d", got, beforeRevision)
	}
	if got := env.RuntimeDataRevision(); got <= beforeRuntimeDataRevision {
		t.Fatalf("runtime data revision after reset = %d, want > %d", got, beforeRuntimeDataRevision)
	}
}
