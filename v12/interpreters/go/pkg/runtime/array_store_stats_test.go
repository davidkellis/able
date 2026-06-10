package runtime

import (
	"sync"
	"testing"
)

func TestArrayStoreStatsSnapshotTracksDynamicAndMonoBacking(t *testing.T) {
	newArrayStoreTestScope(t)

	baseline := ArrayStoreStatsSnapshot()
	if baseline.HandleCount != 0 || baseline.RevisionCount != 0 || baseline.TotalStateCount != 0 {
		t.Fatalf("isolated ArrayStore snapshot = %#v, want empty", baseline)
	}

	dynamic := ArrayStoreNewReservedCapacity(8)
	if err := ArrayStoreWrite(dynamic, 0, NewSmallInt(7, IntegerI32)); err != nil {
		t.Fatalf("write dynamic array: %v", err)
	}
	monoI32 := ArrayStoreMonoNewWithCapacityI32(4)
	ownedU8 := ArrayStoreMonoValueFromOwnedU8Bytes([]byte{1, 2, 3})

	stats := ArrayStoreStatsSnapshot()
	if stats.HandleCount != 3 || stats.RevisionCount != 3 || stats.TotalStateCount != 3 {
		t.Fatalf("registry counts = handles=%d revisions=%d states=%d, want 3/3/3", stats.HandleCount, stats.RevisionCount, stats.TotalStateCount)
	}
	if stats.Dynamic.StateCount != 1 || stats.Dynamic.ValueCount != 1 ||
		stats.Dynamic.DeclaredCapacity != 8 || stats.Dynamic.BackingCapacity != 8 ||
		stats.Dynamic.BackingBytes <= 0 {
		t.Fatalf("dynamic stats = %#v, want one value with 8 backing slots", stats.Dynamic)
	}
	if stats.I32.StateCount != 1 || stats.I32.ValueCount != 0 ||
		stats.I32.DeclaredCapacity != 4 || stats.I32.BackingCapacity != 4 ||
		stats.I32.BackingBytes != 16 {
		t.Fatalf("i32 stats = %#v, want one reserved four-slot state", stats.I32)
	}
	if stats.U8.StateCount != 1 || stats.U8.ValueCount != 3 ||
		stats.U8.DeclaredCapacity != 3 || stats.U8.BackingCapacity != 3 ||
		stats.U8.BackingBytes != 3 {
		t.Fatalf("owned u8 stats = %#v, want one three-byte state", stats.U8)
	}
	if dynamic == monoI32 || dynamic == ownedU8.Handle || monoI32 == ownedU8.Handle {
		t.Fatalf("expected distinct handles: dynamic=%d i32=%d u8=%d", dynamic, monoI32, ownedU8.Handle)
	}
}

func TestArrayStoreStatsSnapshotTracksMonoDeoptimization(t *testing.T) {
	newArrayStoreTestScope(t)

	handle := ArrayStoreMonoNewWithCapacityChar(2)
	before := ArrayStoreStatsSnapshot()
	if before.HandleCount != 1 || before.TotalStateCount != 1 ||
		before.Char.StateCount != 1 || before.Dynamic.StateCount != 0 {
		t.Fatalf("before deoptimization stats = %#v", before)
	}

	if _, err := ArrayStoreState(handle); err != nil {
		t.Fatalf("deoptimize char handle: %v", err)
	}
	after := ArrayStoreStatsSnapshot()
	if after.HandleCount != 1 || after.RevisionCount != 1 || after.TotalStateCount != 1 ||
		after.Char.StateCount != 0 || after.Dynamic.StateCount != 1 ||
		after.Dynamic.DeclaredCapacity != 2 || after.Dynamic.BackingCapacity != 0 {
		t.Fatalf("after deoptimization stats = %#v", after)
	}
}

func TestArrayStoreStatsSnapshotIsSafeDuringConcurrentWrites(t *testing.T) {
	newArrayStoreTestScope(t)

	const workers = 8
	start := make(chan struct{})
	done := make(chan struct{})
	var writers sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		worker := worker
		writers.Add(1)
		go func() {
			defer writers.Done()
			<-start
			handle := ArrayStoreMonoNewWithCapacityU8(4)
			for index := 0; index < 4; index++ {
				if err := ArrayStoreMonoWriteU8(handle, index, uint8(worker+index)); err != nil {
					t.Errorf("writer %d index %d: %v", worker, index, err)
					return
				}
			}
		}()
	}
	go func() {
		writers.Wait()
		close(done)
	}()
	close(start)
	for {
		select {
		case <-done:
			stats := ArrayStoreStatsSnapshot()
			if stats.HandleCount != workers || stats.TotalStateCount != workers ||
				stats.U8.StateCount != workers || stats.U8.ValueCount != workers*4 {
				t.Fatalf("final concurrent stats = %#v", stats)
			}
			return
		default:
			stats := ArrayStoreStatsSnapshot()
			if stats.TotalStateCount < 0 || stats.HandleCount < stats.TotalStateCount {
				t.Fatalf("invalid concurrent stats = %#v", stats)
			}
		}
	}
}
