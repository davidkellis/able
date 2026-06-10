package runtime

import "testing"

func TestArrayStoreRevisionIfAvailableTracksDynamicAndMonoHandles(t *testing.T) {
	dynamic := ArrayStoreNew()
	initialDynamicRevision, ok, err := ArrayStoreRevisionIfAvailable(dynamic)
	if err != nil || !ok {
		t.Fatalf("dynamic initial revision = (%d, %v, %v), want revision/true/nil", initialDynamicRevision, ok, err)
	}
	if err := ArrayStoreWrite(dynamic, 0, NewSmallInt(1, IntegerI32)); err != nil {
		t.Fatalf("ArrayStoreWrite dynamic: %v", err)
	}
	nextDynamicRevision, ok, err := ArrayStoreRevisionIfAvailable(dynamic)
	if err != nil || !ok {
		t.Fatalf("dynamic next revision = (%d, %v, %v), want revision/true/nil", nextDynamicRevision, ok, err)
	}
	if nextDynamicRevision <= initialDynamicRevision {
		t.Fatalf("dynamic revision after write = %d, want > %d", nextDynamicRevision, initialDynamicRevision)
	}

	mono := ArrayStoreMonoNewWithCapacityChar(1)
	initialMonoRevision, ok, err := ArrayStoreRevisionIfAvailable(mono)
	if err != nil || !ok {
		t.Fatalf("mono initial revision = (%d, %v, %v), want revision/true/nil", initialMonoRevision, ok, err)
	}
	if err := ArrayStoreMonoWriteChar(mono, 0, 'x'); err != nil {
		t.Fatalf("ArrayStoreMonoWriteChar: %v", err)
	}
	nextMonoRevision, ok, err := ArrayStoreRevisionIfAvailable(mono)
	if err != nil || !ok {
		t.Fatalf("mono next revision = (%d, %v, %v), want revision/true/nil", nextMonoRevision, ok, err)
	}
	if nextMonoRevision <= initialMonoRevision {
		t.Fatalf("mono revision after write = %d, want > %d", nextMonoRevision, initialMonoRevision)
	}
}

func TestArrayStoreRevisionMatchesIfAvailableTracksWrites(t *testing.T) {
	handle := ArrayStoreNew()
	initialRevision, ok, err := ArrayStoreRevisionIfAvailable(handle)
	if err != nil || !ok {
		t.Fatalf("initial revision = (%d, %v, %v), want revision/true/nil", initialRevision, ok, err)
	}
	matches, ok, err := ArrayStoreRevisionMatchesIfAvailable(handle, initialRevision)
	if err != nil || !ok || !matches {
		t.Fatalf("initial match = (%v, %v, %v), want true/true/nil", matches, ok, err)
	}

	if err := ArrayStoreWrite(handle, 0, NewSmallInt(1, IntegerI32)); err != nil {
		t.Fatalf("ArrayStoreWrite: %v", err)
	}
	matches, ok, err = ArrayStoreRevisionMatchesIfAvailable(handle, initialRevision)
	if err != nil || !ok {
		t.Fatalf("stale match metadata = (%v, %v, %v), want */true/nil", matches, ok, err)
	}
	if matches {
		t.Fatalf("stale revision still matched after write")
	}
	nextRevision, ok, err := ArrayStoreRevisionIfAvailable(handle)
	if err != nil || !ok {
		t.Fatalf("next revision = (%d, %v, %v), want revision/true/nil", nextRevision, ok, err)
	}
	matches, ok, err = ArrayStoreRevisionMatchesIfAvailable(handle, nextRevision)
	if err != nil || !ok || !matches {
		t.Fatalf("next match = (%v, %v, %v), want true/true/nil", matches, ok, err)
	}
}

func TestArrayStoreRevisionCursorMatchesAndInvalidates(t *testing.T) {
	handle := ArrayStoreNew()
	cursor, initialRevision, ok, err := ArrayStoreRevisionCursorIfAvailable(handle)
	if err != nil || !ok {
		t.Fatalf("initial cursor = (%d, %v, %v), want revision/true/nil", initialRevision, ok, err)
	}
	if !cursor.Matches(handle, initialRevision) {
		t.Fatalf("cursor did not match initial revision")
	}
	if !cursor.MatchesKnownHandle(handle, initialRevision) {
		t.Fatalf("cursor did not match known handle initial revision")
	}
	if cursor.Matches(handle+1, initialRevision) {
		t.Fatalf("cursor matched a different handle")
	}
	if cursor.MatchesKnownHandle(handle+1, initialRevision) {
		t.Fatalf("cursor matched a different known handle")
	}
	if cursor.MatchesKnownHandle(0, initialRevision) {
		t.Fatalf("cursor matched zero known handle")
	}

	if err := ArrayStoreWrite(handle, 0, NewSmallInt(1, IntegerI32)); err != nil {
		t.Fatalf("ArrayStoreWrite: %v", err)
	}
	if cursor.Matches(handle, initialRevision) {
		t.Fatalf("cursor matched stale revision after write")
	}
	if cursor.MatchesKnownHandle(handle, initialRevision) {
		t.Fatalf("cursor matched stale known-handle revision after write")
	}
	nextCursor, nextRevision, ok, err := ArrayStoreRevisionCursorIfAvailable(handle)
	if err != nil || !ok {
		t.Fatalf("next cursor = (%d, %v, %v), want revision/true/nil", nextRevision, ok, err)
	}
	if !nextCursor.Matches(handle, nextRevision) {
		t.Fatalf("next cursor did not match updated revision")
	}
	if !nextCursor.MatchesKnownHandle(handle, nextRevision) {
		t.Fatalf("next cursor did not match updated known-handle revision")
	}
}

func TestArrayStoreRevisionMatchesIfAvailableTracksPromotion(t *testing.T) {
	handle := ArrayStoreNew()
	initialRevision, ok, err := ArrayStoreRevisionIfAvailable(handle)
	if err != nil || !ok {
		t.Fatalf("initial revision = (%d, %v, %v), want revision/true/nil", initialRevision, ok, err)
	}

	promoted, err := ArrayStoreAppendCharPromote(handle, 'x')
	if err != nil {
		t.Fatalf("ArrayStoreAppendCharPromote: %v", err)
	}
	if !promoted {
		t.Fatalf("ArrayStoreAppendCharPromote returned promoted=false, want true")
	}
	matches, ok, err := ArrayStoreRevisionMatchesIfAvailable(handle, initialRevision)
	if err != nil || !ok {
		t.Fatalf("stale promoted match metadata = (%v, %v, %v), want */true/nil", matches, ok, err)
	}
	if matches {
		t.Fatalf("stale revision still matched after dynamic->mono promotion")
	}
	nextRevision, ok, err := ArrayStoreRevisionIfAvailable(handle)
	if err != nil || !ok {
		t.Fatalf("next revision = (%d, %v, %v), want revision/true/nil", nextRevision, ok, err)
	}
	matches, ok, err = ArrayStoreRevisionMatchesIfAvailable(handle, nextRevision)
	if err != nil || !ok || !matches {
		t.Fatalf("promoted match = (%v, %v, %v), want true/true/nil", matches, ok, err)
	}
}

func TestArrayStoreRevisionIfAvailableTracksDynamicToMonoPromotion(t *testing.T) {
	handle := ArrayStoreNew()
	initialRevision, ok, err := ArrayStoreRevisionIfAvailable(handle)
	if err != nil || !ok {
		t.Fatalf("initial revision = (%d, %v, %v), want revision/true/nil", initialRevision, ok, err)
	}

	promoted, err := ArrayStoreAppendCharPromote(handle, 'x')
	if err != nil {
		t.Fatalf("ArrayStoreAppendCharPromote: %v", err)
	}
	if !promoted {
		t.Fatalf("ArrayStoreAppendCharPromote returned promoted=false, want true")
	}

	nextRevision, ok, err := ArrayStoreRevisionIfAvailable(handle)
	if err != nil || !ok {
		t.Fatalf("next revision = (%d, %v, %v), want revision/true/nil", nextRevision, ok, err)
	}
	if nextRevision <= initialRevision {
		t.Fatalf("revision after dynamic->mono promotion = %d, want > %d", nextRevision, initialRevision)
	}
}

func TestArrayStoreRevisionIfAvailableTracksGenericDynamicToMonoPromotion(t *testing.T) {
	handle := ArrayStoreNew()
	initialRevision, ok, err := ArrayStoreRevisionIfAvailable(handle)
	if err != nil || !ok {
		t.Fatalf("initial revision = (%d, %v, %v), want revision/true/nil", initialRevision, ok, err)
	}

	promoted, err := ArrayStorePromoteHandleToMonoTypeIfPossible(handle, "bool")
	if err != nil {
		t.Fatalf("ArrayStorePromoteHandleToMonoTypeIfPossible: %v", err)
	}
	if !promoted {
		t.Fatalf("ArrayStorePromoteHandleToMonoTypeIfPossible returned promoted=false, want true")
	}

	nextRevision, ok, err := ArrayStoreRevisionIfAvailable(handle)
	if err != nil || !ok {
		t.Fatalf("next revision = (%d, %v, %v), want revision/true/nil", nextRevision, ok, err)
	}
	if nextRevision <= initialRevision {
		t.Fatalf("revision after generic dynamic->mono promotion = %d, want > %d", nextRevision, initialRevision)
	}
}

func TestArrayStoreRevisionIfAvailableTracksMonoToDynamicDeopt(t *testing.T) {
	handle := ArrayStoreMonoNewWithCapacityChar(1)
	if err := ArrayStoreMonoWriteChar(handle, 0, 'x'); err != nil {
		t.Fatalf("ArrayStoreMonoWriteChar: %v", err)
	}
	initialRevision, ok, err := ArrayStoreRevisionIfAvailable(handle)
	if err != nil || !ok {
		t.Fatalf("initial revision = (%d, %v, %v), want revision/true/nil", initialRevision, ok, err)
	}
	if initialRevision == 0 {
		t.Fatalf("initial mono revision = %d, want non-zero before deopt", initialRevision)
	}

	state, err := ArrayStoreState(handle)
	if err != nil {
		t.Fatalf("ArrayStoreState: %v", err)
	}

	nextRevision, ok, err := ArrayStoreRevisionIfAvailable(handle)
	if err != nil || !ok {
		t.Fatalf("next revision = (%d, %v, %v), want revision/true/nil", nextRevision, ok, err)
	}
	if nextRevision != state.Revision {
		t.Fatalf("revision after mono->dynamic deopt = %d, want state revision %d", nextRevision, state.Revision)
	}
	if nextRevision == initialRevision {
		t.Fatalf("revision after mono->dynamic deopt stayed on stale mono revision %d", nextRevision)
	}
}

func TestArrayStoreRevisionIfAvailableTracksEmptyMonoToDynamicDeopt(t *testing.T) {
	handle := ArrayStoreMonoNewWithCapacityChar(0)
	initialRevision, ok, err := ArrayStoreRevisionIfAvailable(handle)
	if err != nil || !ok {
		t.Fatalf("initial revision = (%d, %v, %v), want revision/true/nil", initialRevision, ok, err)
	}

	state, err := ArrayStoreState(handle)
	if err != nil {
		t.Fatalf("ArrayStoreState: %v", err)
	}

	nextRevision, ok, err := ArrayStoreRevisionIfAvailable(handle)
	if err != nil || !ok {
		t.Fatalf("next revision = (%d, %v, %v), want revision/true/nil", nextRevision, ok, err)
	}
	if nextRevision != state.Revision {
		t.Fatalf("revision after empty mono->dynamic deopt = %d, want state revision %d", nextRevision, state.Revision)
	}
	if nextRevision <= initialRevision {
		t.Fatalf("revision after empty mono->dynamic deopt = %d, want > %d", nextRevision, initialRevision)
	}
}
