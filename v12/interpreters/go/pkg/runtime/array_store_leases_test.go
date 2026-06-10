package runtime

import (
	"sync"
	"testing"
)

type arrayStoreLeaseTestOwner struct {
	lease ArrayStoreLease
}

func TestArrayStoreLeaseLedgerTracksArrayValueViewsAndTransfers(t *testing.T) {
	newArrayStoreTestScope(t)

	owner := &ArrayValue{}
	_, first, err := ArrayStoreEnsure(owner, 2)
	if err != nil {
		t.Fatalf("create first owner: %v", err)
	}
	view, _, err := ArrayStoreValueViewFromHandle(first, 0, 0)
	if err != nil {
		t.Fatalf("create array view: %v", err)
	}
	stats := ArrayStoreLeaseStatsSnapshot()
	if stats.OwnerCount != 2 || stats.OwnersByHandle[first] != 2 {
		t.Fatalf("first-handle leases = %#v, want two owners", stats)
	}

	if err := ArrayStoreTrackArrayValueLease(view, first); err != nil {
		t.Fatalf("repeat view retain: %v", err)
	}
	stats = ArrayStoreLeaseStatsSnapshot()
	if stats.OwnerCount != 2 || stats.OwnersByHandle[first] != 2 {
		t.Fatalf("idempotent view retain changed leases: %#v", stats)
	}

	second := ArrayStoreNew()
	if err := ArrayStoreTrackArrayValueLease(view, second); err != nil {
		t.Fatalf("transfer view lease: %v", err)
	}
	stats = ArrayStoreLeaseStatsSnapshot()
	if stats.OwnerCount != 2 || stats.OwnersByHandle[first] != 1 || stats.OwnersByHandle[second] != 1 {
		t.Fatalf("transferred leases = %#v, want one owner on each handle", stats)
	}

	if err := ArrayStoreReleaseArrayValueLease(owner); err != nil {
		t.Fatalf("release first owner: %v", err)
	}
	if err := ArrayStoreReleaseArrayValueLease(view); err != nil {
		t.Fatalf("release transferred view: %v", err)
	}
	stats = ArrayStoreLeaseStatsSnapshot()
	if stats.OwnerCount != 0 || len(stats.OwnersByHandle) != 0 {
		t.Fatalf("released leases = %#v, want empty ledger", stats)
	}
	if state := ArrayStoreStatsSnapshot(); state.TotalStateCount != 0 || state.HandleCount != 0 || state.RevisionCount != 0 {
		t.Fatalf("final lease release left backing state: %#v", state)
	}
}

func TestArrayStoreLeaseLedgerTracksOwnedU8Factory(t *testing.T) {
	newArrayStoreTestScope(t)

	owner := ArrayStoreMonoValueFromOwnedU8Bytes([]byte{1, 2, 3})
	if owner == nil {
		t.Fatal("owned u8 factory returned nil")
	}
	stats := ArrayStoreLeaseStatsSnapshot()
	if stats.OwnerCount != 1 || stats.OwnersByHandle[owner.Handle] != 1 {
		t.Fatalf("owned u8 lease stats = %#v, want one owner", stats)
	}
	if err := ArrayStoreReleaseArrayValueLease(owner); err != nil {
		t.Fatalf("release owned u8 lease: %v", err)
	}
	if state := ArrayStoreStatsSnapshot(); state.U8.StateCount != 0 || state.TotalStateCount != 0 {
		t.Fatalf("owned u8 backing survived final lease release: %#v", state.U8)
	}
}

func TestArrayStoreArrayValueLeaseCleanupReleasesOnlyItsLedgerToken(t *testing.T) {
	newArrayStoreTestScope(t)

	owner := &ArrayValue{}
	_, handle, err := ArrayStoreEnsure(owner, 0)
	if err != nil {
		t.Fatalf("create owner: %v", err)
	}
	view, _, err := ArrayStoreValueViewFromHandle(handle, 0, 0)
	if err != nil {
		t.Fatalf("create view: %v", err)
	}
	if !owner.Lease.cleanupRegistered || !view.Lease.cleanupRegistered {
		t.Fatal("ArrayValue lease tracking did not register cleanup")
	}

	releaseArrayValueLeaseByCleanupForTest(owner)
	stats := ArrayStoreLeaseStatsSnapshot()
	if stats.OwnerCount != 1 || stats.OwnersByHandle[handle] != 1 {
		t.Fatalf("owner cleanup leases = %#v, want one surviving view", stats)
	}
	if state := ArrayStoreStatsSnapshot(); state.TotalStateCount != 1 {
		t.Fatalf("owner cleanup reclaimed backing state: %#v", state)
	}

	releaseArrayValueLeaseByCleanupForTest(view)
	stats = ArrayStoreLeaseStatsSnapshot()
	if stats.OwnerCount != 0 || len(stats.OwnersByHandle) != 0 {
		t.Fatalf("all cleanup leases = %#v, want empty ledger", stats)
	}
	if state := ArrayStoreStatsSnapshot(); state.TotalStateCount != 0 {
		t.Fatalf("final cleanup left backing state: %#v", state)
	}
}

func TestArrayStoreGenericLeaseOwnerCleanupStopsAndRecreates(t *testing.T) {
	newArrayStoreTestScope(t)

	handle := ArrayStoreNew()
	keeper := &arrayStoreLeaseTestOwner{}
	if err := ArrayStoreTrackLeaseOwner(keeper, &keeper.lease, handle); err != nil {
		t.Fatalf("track generic keeper: %v", err)
	}
	owner := &arrayStoreLeaseTestOwner{}
	if err := ArrayStoreTrackLeaseOwner(owner, &owner.lease, handle); err != nil {
		t.Fatalf("track generic owner: %v", err)
	}
	if !owner.lease.cleanupRegistered {
		t.Fatal("generic lease owner did not register cleanup")
	}
	if stats := ArrayStoreLeaseStatsSnapshot(); stats.OwnerCount != 2 || stats.OwnersByHandle[handle] != 2 {
		t.Fatalf("generic owner leases = %#v, want two owners", stats)
	}
	if err := ArrayStoreTrackLeaseOwner(owner, &owner.lease, handle); err != nil {
		t.Fatalf("repeat track generic owner: %v", err)
	}
	if stats := ArrayStoreLeaseStatsSnapshot(); stats.OwnerCount != 2 || stats.OwnersByHandle[handle] != 2 {
		t.Fatalf("repeat generic owner tracking changed leases: %#v", stats)
	}

	if err := ArrayStoreUpdateLease(&owner.lease, 0); err != nil {
		t.Fatalf("release generic owner lease: %v", err)
	}
	if owner.lease.cleanupRegistered {
		t.Fatal("released generic owner retained cleanup registration")
	}
	if stats := ArrayStoreLeaseStatsSnapshot(); stats.OwnerCount != 1 || stats.OwnersByHandle[handle] != 1 {
		t.Fatalf("released generic owner leases = %#v, want keeper", stats)
	}

	if err := ArrayStoreTrackLeaseOwner(owner, &owner.lease, handle); err != nil {
		t.Fatalf("retrack generic owner: %v", err)
	}
	if !owner.lease.cleanupRegistered {
		t.Fatal("retracked generic owner did not recreate cleanup")
	}
	releaseArrayStoreLeaseByCleanupForTest(&owner.lease)
	if stats := ArrayStoreLeaseStatsSnapshot(); stats.OwnerCount != 1 || stats.OwnersByHandle[handle] != 1 {
		t.Fatalf("generic owner cleanup leases = %#v, want keeper", stats)
	}
	if err := ArrayStoreUpdateLease(&keeper.lease, 0); err != nil {
		t.Fatalf("release generic keeper lease: %v", err)
	}
	if state := ArrayStoreStatsSnapshot(); state.TotalStateCount != 0 {
		t.Fatalf("final generic cleanup left backing state: %#v", state)
	}
}

func TestArrayStoreLeaseCleanupRequiresExactGeneration(t *testing.T) {
	newArrayStoreTestScope(t)

	owner := &ArrayValue{}
	_, handle, err := ArrayStoreEnsure(owner, 0)
	if err != nil {
		t.Fatalf("create owner: %v", err)
	}
	arrayStoreMu.Lock()
	wrong := arrayStoreLeaseKeyForLocked(&owner.Lease)
	wrong.generation++
	arrayStoreMu.Unlock()

	arrayStoreReleaseLeaseByCleanup(wrong)
	stats := ArrayStoreLeaseStatsSnapshot()
	if stats.OwnerCount != 1 || stats.OwnersByHandle[handle] != 1 {
		t.Fatalf("wrong-generation cleanup changed leases: %#v", stats)
	}
	if err := ArrayStoreReleaseArrayValueLease(owner); err != nil {
		t.Fatalf("release owner: %v", err)
	}
}

func TestArrayStoreLeaseLedgerRejectsUnknownHandle(t *testing.T) {
	newArrayStoreTestScope(t)

	var lease ArrayStoreLease
	if err := ArrayStoreUpdateLease(&lease, 99); err == nil {
		t.Fatal("expected unknown handle to be rejected")
	}
	if stats := ArrayStoreLeaseStatsSnapshot(); stats.OwnerCount != 0 {
		t.Fatalf("unknown handle changed lease stats: %#v", stats)
	}
}

func TestArrayStoreLeaseLedgerIsSafeDuringConcurrentUpdates(t *testing.T) {
	newArrayStoreTestScope(t)

	const workers = 8
	handle := ArrayStoreNew()
	var root ArrayStoreLease
	if err := ArrayStoreUpdateLease(&root, handle); err != nil {
		t.Fatalf("retain shared root lease: %v", err)
	}
	leases := make([]ArrayStoreLease, workers)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for index := range leases {
		index := index
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if err := ArrayStoreUpdateLease(&leases[index], handle); err != nil {
				t.Errorf("retain lease %d: %v", index, err)
				return
			}
			if err := ArrayStoreUpdateLease(&leases[index], 0); err != nil {
				t.Errorf("release lease %d: %v", index, err)
			}
		}()
	}
	close(start)
	wg.Wait()
	if err := ArrayStoreUpdateLease(&root, 0); err != nil {
		t.Fatalf("release shared root lease: %v", err)
	}
	if stats := ArrayStoreLeaseStatsSnapshot(); stats.OwnerCount != 0 || len(stats.OwnersByHandle) != 0 {
		t.Fatalf("concurrent releases left leases behind: %#v", stats)
	}
	if state := ArrayStoreStatsSnapshot(); state.TotalStateCount != 0 {
		t.Fatalf("concurrent final release left backing state: %#v", state)
	}
}
