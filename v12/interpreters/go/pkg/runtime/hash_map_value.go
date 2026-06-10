package runtime

const hashMapLinearSearchLimit = 16

// HashCandidates returns entry positions with the requested hash once a map is
// large enough that maintaining an index is cheaper than scanning every entry.
// Key equality remains the caller's responsibility so hash collisions preserve
// Able's Eq semantics.
func (v *HashMapValue) HashCandidates(hash uint64) ([]int, bool) {
	if v == nil || len(v.Entries) < hashMapLinearSearchLimit {
		return nil, false
	}
	if v.hashIndex == nil || v.hashIndexEntries != len(v.Entries) {
		v.rebuildHashIndex()
	}
	return v.hashIndex[hash], true
}

func (v *HashMapValue) AppendEntry(entry HashMapEntry) {
	idx := len(v.Entries)
	v.Entries = append(v.Entries, entry)
	if v.hashIndex != nil && v.hashIndexEntries == idx {
		v.hashIndex[entry.Hash] = append(v.hashIndex[entry.Hash], idx)
		v.hashIndexEntries++
		return
	}
	v.hashIndex = nil
	v.hashIndexEntries = 0
}

func (v *HashMapValue) RemoveEntry(idx int) (HashMapEntry, bool) {
	if v == nil || idx < 0 || idx >= len(v.Entries) {
		return HashMapEntry{}, false
	}
	removed := v.Entries[idx]
	copy(v.Entries[idx:], v.Entries[idx+1:])
	v.Entries[len(v.Entries)-1] = HashMapEntry{}
	v.Entries = v.Entries[:len(v.Entries)-1]
	if v.hashIndex != nil {
		v.rebuildHashIndex()
	}
	return removed, true
}

func (v *HashMapValue) ClearEntries() {
	if v == nil {
		return
	}
	v.Entries = v.Entries[:0]
	v.hashIndex = nil
	v.hashIndexEntries = 0
}

func (v *HashMapValue) rebuildHashIndex() {
	index := make(map[uint64][]int, len(v.Entries))
	for idx, entry := range v.Entries {
		index[entry.Hash] = append(index[entry.Hash], idx)
	}
	v.hashIndex = index
	v.hashIndexEntries = len(v.Entries)
}
