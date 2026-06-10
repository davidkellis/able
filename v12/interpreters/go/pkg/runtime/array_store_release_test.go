package runtime

import "testing"

func TestArrayStoreFinalLeaseReleaseRemovesEveryStorageKind(t *testing.T) {
	for _, tc := range []struct {
		name string
		new  func() int64
	}{
		{name: "dynamic", new: ArrayStoreNew},
		{name: "i32", new: ArrayStoreMonoNewI32},
		{name: "i64", new: ArrayStoreMonoNewI64},
		{name: "bool", new: ArrayStoreMonoNewBool},
		{name: "char", new: ArrayStoreMonoNewChar},
		{name: "u8", new: ArrayStoreMonoNewU8},
		{name: "u32", new: ArrayStoreMonoNewU32},
		{name: "u64", new: ArrayStoreMonoNewU64},
		{name: "f64", new: ArrayStoreMonoNewF64},
	} {
		t.Run(tc.name, func(t *testing.T) {
			newArrayStoreTestScope(t)

			handle := tc.new()
			owner := &arrayStoreLeaseTestOwner{}
			if err := ArrayStoreTrackLeaseOwner(owner, &owner.lease, handle); err != nil {
				t.Fatalf("track %s owner: %v", tc.name, err)
			}
			if err := ArrayStoreUpdateLease(&owner.lease, 0); err != nil {
				t.Fatalf("release %s owner: %v", tc.name, err)
			}
			if stats := ArrayStoreStatsSnapshot(); stats.TotalStateCount != 0 || stats.HandleCount != 0 || stats.RevisionCount != 0 {
				t.Fatalf("%s release stats = %#v, want empty registry", tc.name, stats)
			}
			if _, err := ArrayStoreSize(handle); err == nil {
				t.Fatalf("%s released handle remained readable", tc.name)
			}
			if _, err := ArrayStoreEnsureHandle(handle, 0, 0); err == nil {
				t.Fatalf("%s released handle was silently recreated", tc.name)
			}
		})
	}
}

func TestArrayStoreFinalReleaseClearsKindAndRevisionMetadata(t *testing.T) {
	newArrayStoreTestScope(t)

	handle := ArrayStoreMonoNewWithCapacityU8(1)
	if _, ok, err := ArrayStoreMonoElementTypeNameIfKnown(handle); err != nil || !ok {
		t.Fatalf("prime u8 type metadata = (%v, %v), want known", ok, err)
	}
	if _, ok, err := ArrayStoreRevisionIfAvailable(handle); err != nil || !ok {
		t.Fatalf("prime revision metadata = (%v, %v), want known", ok, err)
	}
	if _, ok, err := ArrayStoreMonoReadU8IfAvailable(handle, 0); err != nil || ok {
		t.Fatalf("prime u8 read metadata = (%v, %v), want unavailable in-bounds value", ok, err)
	}

	owner := &arrayStoreLeaseTestOwner{}
	if err := ArrayStoreTrackLeaseOwner(owner, &owner.lease, handle); err != nil {
		t.Fatalf("track u8 owner: %v", err)
	}
	if err := ArrayStoreUpdateLease(&owner.lease, 0); err != nil {
		t.Fatalf("release u8 owner: %v", err)
	}
	if _, ok, err := ArrayStoreRevisionIfAvailable(handle); err == nil || ok {
		t.Fatalf("released revision metadata = (%v, %v), want unknown-handle error", ok, err)
	}
	if _, ok, err := ArrayStoreMonoElementTypeNameIfKnown(handle); err == nil || ok {
		t.Fatalf("released type metadata = (%v, %v), want unknown-handle error", ok, err)
	}
	if _, ok, err := ArrayStoreMonoReadU8IfAvailable(handle, 0); err == nil || ok {
		t.Fatalf("released u8 read = (%v, %v), want unknown-handle error", ok, err)
	}
}

func TestArrayStoreAdoptHandleIsExplicitAndStaleEnsureFails(t *testing.T) {
	newArrayStoreTestScope(t)

	const handle int64 = 91
	if _, err := ArrayStoreEnsureHandle(handle, 2, 3); err == nil {
		t.Fatal("normal ensure created an unknown handle")
	}
	state, err := ArrayStoreAdoptHandle(handle, 2, 3)
	if err != nil {
		t.Fatalf("adopt external handle: %v", err)
	}
	if len(state.Values) != 2 || state.Capacity != 3 {
		t.Fatalf("adopted state = len %d cap %d, want 2/3", len(state.Values), state.Capacity)
	}

	owner := &arrayStoreLeaseTestOwner{}
	if err := ArrayStoreTrackLeaseOwner(owner, &owner.lease, handle); err != nil {
		t.Fatalf("track adopted owner: %v", err)
	}
	if err := ArrayStoreUpdateLease(&owner.lease, 0); err != nil {
		t.Fatalf("release adopted owner: %v", err)
	}
	if _, err := ArrayStoreEnsureHandle(handle, 0, 0); err == nil {
		t.Fatal("released adopted handle was silently recreated")
	}
}

func TestArrayStoreEnsureRejectsReleasedArrayValue(t *testing.T) {
	newArrayStoreTestScope(t)

	owner := &ArrayValue{}
	_, handle, err := ArrayStoreEnsure(owner, 0)
	if err != nil {
		t.Fatalf("create owner: %v", err)
	}
	if err := ArrayStoreReleaseArrayValueLease(owner); err != nil {
		t.Fatalf("release owner: %v", err)
	}
	if _, _, err := ArrayStoreEnsure(owner, 0); err == nil {
		t.Fatalf("released ArrayValue handle %d was silently recreated", handle)
	}
}

func TestArrayStoreMoveLeaseTransfersFinalOwnerWithoutStateGap(t *testing.T) {
	newArrayStoreTestScope(t)

	handle := ArrayStoreNew()
	source := &arrayStoreLeaseTestOwner{}
	target := &arrayStoreLeaseTestOwner{}
	if err := ArrayStoreTrackLeaseOwner(source, &source.lease, handle); err != nil {
		t.Fatalf("track source: %v", err)
	}
	arrayStoreMu.Lock()
	oldSourceKey := arrayStoreLeaseKeyForLocked(&source.lease)
	arrayStoreMu.Unlock()

	if err := ArrayStoreMoveLease(&target.lease, &source.lease); err != nil {
		t.Fatalf("move final lease: %v", err)
	}
	if source.lease.tracked || !target.lease.tracked || target.lease.handle != handle {
		t.Fatalf("moved leases = source %#v target %#v, want source released and target on %d", source.lease, target.lease, handle)
	}
	if stats := ArrayStoreLeaseStatsSnapshot(); stats.OwnerCount != 1 || stats.OwnersByHandle[handle] != 1 {
		t.Fatalf("moved lease stats = %#v, want one target owner", stats)
	}
	if _, err := ArrayStoreSize(handle); err != nil {
		t.Fatalf("move removed live backing state: %v", err)
	}

	other := ArrayStoreNew()
	if err := ArrayStoreTrackLeaseOwner(source, &source.lease, other); err != nil {
		t.Fatalf("reuse source lease: %v", err)
	}
	arrayStoreReleaseLeaseByCleanup(oldSourceKey)
	if stats := ArrayStoreLeaseStatsSnapshot(); stats.OwnersByHandle[other] != 1 {
		t.Fatalf("delayed source cleanup changed reused lease: %#v", stats)
	}

	if err := ArrayStoreUpdateLease(&target.lease, 0); err != nil {
		t.Fatalf("release moved target: %v", err)
	}
	if err := ArrayStoreUpdateLease(&source.lease, 0); err != nil {
		t.Fatalf("release reused source: %v", err)
	}
	if stats := ArrayStoreStatsSnapshot(); stats.TotalStateCount != 0 {
		t.Fatalf("final moved lease release left backing state: %#v", stats)
	}
}
