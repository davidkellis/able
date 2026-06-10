package runtime

import "testing"

func TestEnvironmentBindingShapeRevisionTracksTopologyOnly(t *testing.T) {
	root := NewEnvironment(nil)
	root.SetSingleThread()
	initialState := root.BindingShapeStateID()
	initialShape := root.BindingShapeRevision()

	root.DefineWithoutMerge("value", NilValue{})
	afterAdd := root.BindingShapeRevision()
	if initialState == 0 {
		t.Fatalf("binding shape state id should be nonzero")
	}
	if afterAdd <= initialShape {
		t.Fatalf("binding shape revision after add = %d, want > %d", afterAdd, initialShape)
	}

	root.DefineWithoutMerge("value", StringValue{Val: "next"})
	if got := root.BindingShapeRevision(); got != afterAdd {
		t.Fatalf("binding shape revision after same-name replace = %d, want %d", got, afterAdd)
	}
	if err := root.Assign("value", StringValue{Val: "assigned"}); err != nil {
		t.Fatalf("Assign(value): %v", err)
	}
	if got := root.BindingShapeRevision(); got != afterAdd {
		t.Fatalf("binding shape revision after assign = %d, want %d", got, afterAdd)
	}

	child := NewEnvironment(root)
	child.DefineWithoutMerge("child", NilValue{})
	afterChildAdd := root.BindingShapeRevision()
	if afterChildAdd <= afterAdd {
		t.Fatalf("binding shape revision after child add = %d, want > %d", afterChildAdd, afterAdd)
	}

	child.ResetForSingleBindingReuse(root, 1, "next", NilValue{})
	if got := root.BindingShapeRevision(); got <= afterChildAdd {
		t.Fatalf("binding shape revision after child reuse = %d, want > %d", got, afterChildAdd)
	}
}

func TestEnvironmentBindingNameRevisionTracksSpecificNames(t *testing.T) {
	root := NewEnvironment(nil)
	root.SetSingleThread()
	initialValue := root.BindingNameRevision("value")
	initialOther := root.BindingNameRevision("other")

	root.DefineWithoutMerge("other", NilValue{})
	if got := root.BindingNameRevision("value"); got != initialValue {
		t.Fatalf("value name revision after unrelated add = %d, want %d", got, initialValue)
	}
	if got := root.BindingNameRevision("other"); got <= initialOther {
		t.Fatalf("other name revision after add = %d, want > %d", got, initialOther)
	}

	root.DefineWithoutMerge("value", NilValue{})
	afterValueAdd := root.BindingNameRevision("value")
	if afterValueAdd <= initialValue {
		t.Fatalf("value name revision after add = %d, want > %d", afterValueAdd, initialValue)
	}
	root.DefineWithoutMerge("value", StringValue{Val: "next"})
	if got := root.BindingNameRevision("value"); got != afterValueAdd {
		t.Fatalf("value name revision after same-name replace = %d, want %d", got, afterValueAdd)
	}

	child := NewEnvironment(root)
	child.DefineWithoutMerge("child", NilValue{})
	if got := root.BindingNameRevision("value"); got != afterValueAdd {
		t.Fatalf("value name revision after unrelated child add = %d, want %d", got, afterValueAdd)
	}
	child.ResetForSingleBindingReuse(root, 1, "value", NilValue{})
	if got := root.BindingNameRevision("value"); got <= afterValueAdd {
		t.Fatalf("value name revision after child reuse = %d, want > %d", got, afterValueAdd)
	}
}
