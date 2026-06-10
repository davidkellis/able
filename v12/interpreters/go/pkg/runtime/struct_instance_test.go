package runtime

import "testing"

var structInstanceAllocSink *StructInstanceValue

func TestNewStructInstancePositionalSizedUsesInlineStorageForSmallPayloads(t *testing.T) {
	inst, values := NewStructInstancePositionalSized(nil, 2, nil)
	if inst == nil {
		t.Fatalf("expected instance")
	}
	if len(values) != 2 || len(inst.Positional) != 2 {
		t.Fatalf("unexpected positional lengths values=%d inst=%d", len(values), len(inst.Positional))
	}
	values[0] = StringValue{Val: "x"}
	values[1] = BoolValue{Val: true}
	if &inst.Positional[0] != &inst.inlinePositional[0] {
		t.Fatalf("expected small positional payload to reuse inline storage")
	}
	if inst.Positional[0] != values[0] || inst.Positional[1] != values[1] {
		t.Fatalf("inline positional payload mismatch: %#v", inst.Positional)
	}
}

func TestNewStructInstancePositionalSizedUsesInlineStorageAtCapacity(t *testing.T) {
	inst, values := NewStructInstancePositionalSized(nil, structInstanceInlinePositionalCapacity, nil)
	if inst == nil {
		t.Fatalf("expected instance")
	}
	if len(values) != structInstanceInlinePositionalCapacity || len(inst.Positional) != structInstanceInlinePositionalCapacity {
		t.Fatalf("unexpected positional lengths values=%d inst=%d", len(values), len(inst.Positional))
	}
	if len(values) > 0 && &inst.Positional[0] != &inst.inlinePositional[0] {
		t.Fatalf("expected capacity-sized payload to reuse inline storage")
	}
}

func TestNewStructInstancePositionalSizedAllocatesSliceForLargePayloads(t *testing.T) {
	inst, values := NewStructInstancePositionalSized(nil, structInstanceInlinePositionalCapacity+1, nil)
	if inst == nil {
		t.Fatalf("expected instance")
	}
	if len(values) != structInstanceInlinePositionalCapacity+1 || len(inst.Positional) != structInstanceInlinePositionalCapacity+1 {
		t.Fatalf("unexpected positional lengths values=%d inst=%d", len(values), len(inst.Positional))
	}
	if &inst.Positional[0] == &inst.inlinePositional[0] {
		t.Fatalf("expected large positional payload not to reuse inline storage")
	}
}

func TestNewStructInstancePositionalSizedSmallPayloadAllocationBudget(t *testing.T) {
	for _, fieldCount := range []int{0, 1, 2, structInstanceInlinePositionalCapacity} {
		allocs := testing.AllocsPerRun(1000, func() {
			inst, values := NewStructInstancePositionalSized(nil, fieldCount, nil)
			for i := range values {
				values[i] = NilValue{}
			}
			structInstanceAllocSink = inst
		})
		if allocs > 1 {
			t.Fatalf("fieldCount=%d allocs=%g, want at most one escaping instance allocation", fieldCount, allocs)
		}
	}
}

func TestNewStructInstancePositionalSizedLargePayloadAllocationBudget(t *testing.T) {
	fieldCount := structInstanceInlinePositionalCapacity + 1
	allocs := testing.AllocsPerRun(1000, func() {
		inst, values := NewStructInstancePositionalSized(nil, fieldCount, nil)
		for i := range values {
			values[i] = NilValue{}
		}
		structInstanceAllocSink = inst
	})
	if allocs > 2 {
		t.Fatalf("allocs=%g, want at most instance plus positional slice allocations", allocs)
	}
}
