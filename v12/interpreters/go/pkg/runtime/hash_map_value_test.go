package runtime

import (
	"reflect"
	"testing"
)

func TestHashMapValueHashCandidatesTrackAppendRemoveAndClear(t *testing.T) {
	state := &HashMapValue{}
	for idx := 0; idx < 32; idx++ {
		state.AppendEntry(HashMapEntry{Hash: uint64(idx % 4)})
	}

	candidates, indexed := state.HashCandidates(2)
	if !indexed {
		t.Fatal("expected a large map to use its hash index")
	}
	if want := []int{2, 6, 10, 14, 18, 22, 26, 30}; !reflect.DeepEqual(candidates, want) {
		t.Fatalf("unexpected initial candidates: got %v want %v", candidates, want)
	}

	state.AppendEntry(HashMapEntry{Hash: 2})
	candidates, indexed = state.HashCandidates(2)
	if !indexed || candidates[len(candidates)-1] != 32 {
		t.Fatalf("append did not update active index: indexed=%v candidates=%v", indexed, candidates)
	}

	removed, ok := state.RemoveEntry(2)
	if !ok || removed.Hash != 2 {
		t.Fatalf("unexpected removal result: ok=%v entry=%#v", ok, removed)
	}
	candidates, indexed = state.HashCandidates(2)
	if want := []int{5, 9, 13, 17, 21, 25, 29, 31}; !indexed || !reflect.DeepEqual(candidates, want) {
		t.Fatalf("remove did not rebuild shifted positions: got %v want %v", candidates, want)
	}

	state.ClearEntries()
	if len(state.Entries) != 0 {
		t.Fatalf("clear retained %d logical entries", len(state.Entries))
	}
	if candidates, indexed = state.HashCandidates(2); indexed || candidates != nil {
		t.Fatalf("small cleared map should use linear mode: indexed=%v candidates=%v", indexed, candidates)
	}
}

func TestHashMapValueSmallMapStaysLinear(t *testing.T) {
	state := &HashMapValue{}
	for idx := 0; idx < hashMapLinearSearchLimit-1; idx++ {
		state.AppendEntry(HashMapEntry{Hash: uint64(idx)})
	}
	if candidates, indexed := state.HashCandidates(3); indexed || candidates != nil {
		t.Fatalf("small map unexpectedly built an index: indexed=%v candidates=%v", indexed, candidates)
	}
}
